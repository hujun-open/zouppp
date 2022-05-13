package client

import (
	"context"
	"encoding/binary"
	"fmt"
	"net"
	"time"

	"github.com/insomniacslk/dhcp/dhcpv6"
	"github.com/insomniacslk/dhcp/dhcpv6/nclient6"
	"github.com/insomniacslk/dhcp/iana"
)

type DHCP6Cfg struct {
	Mac            net.HardwareAddr
	NeedPD, NeedNA bool
	Debug          bool
}

type DHCP6Clnt struct {
	clnt                              *nclient6.Client
	cfg                               *DHCP6Cfg
	locaddr                           net.IP
	ipHeader, pseudoHeader, udpHeader []byte
	assignedIANAs                     []net.IP
	assignedIAPDs                     []*net.IPNet
}

//mac address is used for client id only, not used for forwarding over PPP
func NewDHCP6Clnt(conn net.PacketConn, cfg *DHCP6Cfg, localLLA net.IP) (*DHCP6Clnt, error) {
	mods := []nclient6.ClientOpt{}
	if cfg.Debug {
		mods = []nclient6.ClientOpt{nclient6.WithDebugLogger(), nclient6.WithLogDroppedPackets()}
	}
	r := new(DHCP6Clnt)
	var err error
	r.clnt, err = nclient6.NewWithConn(conn, cfg.Mac, mods...)
	if err != nil {
		return nil, fmt.Errorf("failed to create DHCPv6 client, %v", err)
	}
	r.cfg = cfg

	//v6
	r.locaddr = localLLA
	r.ipHeader = make([]byte, 40)
	r.ipHeader[0] = 0x60                         //version
	r.ipHeader[6] = 17                           //next header
	r.ipHeader[7] = 32                           //TTL
	copy(r.ipHeader[8:24], localLLA.To16()[:16]) //src addr
	r.pseudoHeader = make([]byte, 40)
	copy(r.pseudoHeader[:16], localLLA.To16()[:16]) //src addr
	r.pseudoHeader[39] = 17                         //next header
	r.udpHeader = make([]byte, 8)
	binary.BigEndian.PutUint16(r.udpHeader[:2], uint16(dhcpv6.DefaultClientPort)) //src port
	r.assignedIANAs = []net.IP{}
	r.assignedIAPDs = []*net.IPNet{}
	return r, nil
}
func getIAIDviaTime(delta int64) (r [4]byte) {
	buf := make([]byte, binary.MaxVarintLen64)
	binary.PutVarint(buf, time.Now().UnixNano()+delta)
	copy(r[:], buf[:4])
	return
}

func (dc *DHCP6Clnt) buildSolicit() (*dhcpv6.Message, error) {
	optModList := []dhcpv6.Modifier{}
	if dc.cfg.NeedNA {
		optModList = append(optModList, dhcpv6.WithIAID(getIAIDviaTime(0)))
	}
	if dc.cfg.NeedPD {
		optModList = append(optModList, dhcpv6.WithIAPD(getIAIDviaTime(1)))
	}
	duid := dhcpv6.Duid{
		Type:          dhcpv6.DUID_LL,
		HwType:        iana.HWTypeEthernet,
		Time:          dhcpv6.GetTime(),
		LinkLayerAddr: dc.cfg.Mac,
	}
	m, err := dhcpv6.NewMessage()
	if err != nil {
		return nil, err
	}
	m.MessageType = dhcpv6.MessageTypeSolicit
	m.AddOption(dhcpv6.OptClientID(duid))
	m.AddOption(dhcpv6.OptRequestedOption(
		dhcpv6.OptionDNSRecursiveNameServer,
		dhcpv6.OptionDomainSearchList,
	))
	m.AddOption(dhcpv6.OptElapsedTime(0))
	for _, mod := range optModList {
		mod(m)
	}
	return m, nil

}

func (dc *DHCP6Clnt) Dial() error {
	checkResp := func(msg *dhcpv6.Message, record bool) error {
		if dc.cfg.NeedNA {

			if len(msg.Options.OneIANA().Options.Addresses()) == 0 {
				return fmt.Errorf("no IANA address is assigned")
			}
			if record {
				dc.assignedIANAs = []net.IP{}
				for _, addr := range msg.Options.OneIANA().Options.Addresses() {
					dc.assignedIANAs = append(dc.assignedIANAs, addr.IPv6Addr)
				}
			}

		}
		if dc.cfg.NeedPD {
			if len(msg.Options.OneIAPD().Options.Prefixes()) == 0 {
				return fmt.Errorf("no IAPD prefix is assigned")
			}
			if record {
				dc.assignedIAPDs = []*net.IPNet{}
				for _, p := range msg.Options.OneIAPD().Options.Prefixes() {
					dc.assignedIAPDs = append(dc.assignedIAPDs, p.Prefix)
				}
			}
		}
		return nil
	}
	solicitMsg, err := dc.buildSolicit()
	if err != nil {
		return fmt.Errorf("failed to create solicit msg for %v, %v", dc.cfg.Mac, err)
	}
	adv, err := dc.clnt.SendAndRead(context.Background(),
		nclient6.AllDHCPRelayAgentsAndServers, solicitMsg,
		nclient6.IsMessageType(dhcpv6.MessageTypeAdvertise))
	if err != nil {
		return fmt.Errorf("failed recv DHCPv6 advertisement for %v, %v", dc.cfg.Mac, err)
	}
	err = checkResp(adv, false)
	if err != nil {
		return fmt.Errorf("got invalid advertise msg for clnt %v, %v", dc.cfg.Mac, err)
	}
	request, err := NewRequestFromAdv(adv)
	if err != nil {
		return fmt.Errorf("failed to build request msg for clnt %v, %v", dc.cfg.Mac, err)
	}
	reply, err := dc.clnt.SendAndRead(context.Background(),
		nclient6.AllDHCPRelayAgentsAndServers,
		request, nclient6.IsMessageType(dhcpv6.MessageTypeReply))
	if err != nil {
		return fmt.Errorf("failed to recv DHCPv6 reply for %v, %v", dc.cfg.Mac, err)
	}
	err = checkResp(reply, true)
	if err != nil {
		return fmt.Errorf("got invalid reply msg for %v, %v", dc.cfg.Mac, err)
	}
	return nil
}
func NewRequestFromAdv(adv *dhcpv6.Message, modifiers ...dhcpv6.Modifier) (*dhcpv6.Message, error) {
	if adv == nil {
		return nil, fmt.Errorf("ADVERTISE cannot be nil")
	}
	if adv.MessageType != dhcpv6.MessageTypeAdvertise {
		return nil, fmt.Errorf("the passed ADVERTISE must have ADVERTISE type set")
	}
	// build REQUEST from ADVERTISE
	req, err := dhcpv6.NewMessage()
	if err != nil {
		return nil, err
	}
	req.MessageType = dhcpv6.MessageTypeRequest
	// add Client ID
	cid := adv.GetOneOption(dhcpv6.OptionClientID)
	if cid == nil {
		return nil, fmt.Errorf("client ID cannot be nil in ADVERTISE when building REQUEST")
	}
	req.AddOption(cid)
	// add Server ID
	sid := adv.GetOneOption(dhcpv6.OptionServerID)
	if sid == nil {
		return nil, fmt.Errorf("server ID cannot be nil in ADVERTISE when building REQUEST")
	}
	req.AddOption(sid)
	// add Elapsed Time
	req.AddOption(dhcpv6.OptElapsedTime(0))
	// add IA_NA
	if iana := adv.Options.OneIANA(); iana != nil {
		req.AddOption(iana)
	}
	// if iana == nil {
	// 	return nil, fmt.Errorf("IA_NA cannot be nil in ADVERTISE when building REQUEST")
	// }
	// req.AddOption(iana)
	// add IA_PD
	if iaPd := adv.GetOneOption(dhcpv6.OptionIAPD); iaPd != nil {
		req.AddOption(iaPd)
	}
	req.AddOption(dhcpv6.OptRequestedOption(
		dhcpv6.OptionDNSRecursiveNameServer,
		dhcpv6.OptionDomainSearchList,
	))
	// add OPTION_VENDOR_CLASS, only if present in the original request
	vClass := adv.GetOneOption(dhcpv6.OptionVendorClass)
	if vClass != nil {
		req.AddOption(vClass)
	}

	// apply modifiers
	for _, mod := range modifiers {
		mod(req)
	}
	return req, nil
}
