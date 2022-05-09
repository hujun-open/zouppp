package client

import (
	"context"
	"encoding/binary"
	"fmt"
	"net"
	"time"

	"github.com/hujun-open/zouppp/lcp"
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

func (dc *DHCP6Clnt) sendHandler(buf []byte, dst net.Addr) ([]byte, int, error) {
	rbuf := dc.buildIPv6Pkt(buf, dst.(*net.UDPAddr))
	return rbuf, len(rbuf), nil
}

func (dc *DHCP6Clnt) rcvHandler(buf []byte) ([]byte, int, error) {
	//get UDP payload, no support for any IPv6 option
	fmt.Println("<<<<<<<<<<<<<<<<<<<<<<<<<<<")
	fmt.Println("recvd handler", buf)
	return buf[48:], len(buf) - 48, nil
}

func (dc *DHCP6Clnt) Dial() error {
	checkResp := func(msg *dhcpv6.Message) error {
		if dc.cfg.NeedNA {

			if len(msg.Options.OneIANA().Options.Addresses()) == 0 {
				return fmt.Errorf("no IANA address is assigned")
			}
		}
		if dc.cfg.NeedPD {
			if len(msg.Options.OneIAPD().Options.Prefixes()) == 0 {
				return fmt.Errorf("no IAPD prefix is assigned")
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
	err = checkResp(adv)
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
	err = checkResp(reply)
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

func (dc *DHCP6Clnt) buildIPv6Pkt(p []byte, dst *net.UDPAddr) []byte {
	src := net.UDPAddr{IP: dc.locaddr, Port: dhcpv6.DefaultClientPort}
	var fullp []byte

	//v6
	fullp = append(make([]byte, 50), p...)
	binary.BigEndian.PutUint16(fullp[:2], uint16(lcp.ProtoIPv6))
	startIndex := 2
	psuHeader := make([]byte, 40)
	copy(psuHeader, dc.pseudoHeader)
	copy(fullp[startIndex:startIndex+40], dc.ipHeader)
	copy(fullp[startIndex+40:startIndex+48], dc.udpHeader)
	//ip header
	binary.BigEndian.PutUint16(fullp[startIndex+4:startIndex+6], uint16(8+len(p))) //payload length
	copy(fullp[startIndex+8:startIndex+24], src.IP.To16()[:16])                    //src addr
	copy(fullp[startIndex+24:startIndex+40], dst.IP.To16()[:16])                   //dst addr
	//psudo header
	copy(psuHeader[:16], src.IP.To16()[:16])                       //src addr
	copy(psuHeader[16:32], dst.IP.To16()[:16])                     //dst addr
	binary.BigEndian.PutUint32(psuHeader[32:36], uint32(8+len(p))) //udp len
	//udp header
	binary.BigEndian.PutUint16(fullp[startIndex+40:startIndex+42], uint16(src.Port))                                //src port
	binary.BigEndian.PutUint16(fullp[startIndex+42:startIndex+44], uint16(dst.Port))                                //dst port
	binary.BigEndian.PutUint16(fullp[startIndex+44:startIndex+46], uint16(8+len(p)))                                //udp len
	binary.BigEndian.PutUint16(fullp[startIndex+46:startIndex+48], v6udpChecksum(fullp[startIndex+40:], psuHeader)) //udp checksum

	return fullp

}

func v6udpChecksum(headerAndPayload, psudoHeader []byte) uint16 {
	length := uint32(len(headerAndPayload))
	csum := v6pseudoheaderChecksum(psudoHeader)
	csum += uint32(17)
	csum += length & 0xffff
	csum += length >> 16
	return tcpipChecksum(headerAndPayload, csum)
}
func v6pseudoheaderChecksum(pHeader []byte) (csum uint32) {
	SrcIP := pHeader[:16]
	DstIP := pHeader[16:32]
	for i := 0; i < 16; i += 2 {
		csum += uint32(SrcIP[i]) << 8
		csum += uint32(SrcIP[i+1])
		csum += uint32(DstIP[i]) << 8
		csum += uint32(DstIP[i+1])
	}
	return csum
}

func tcpipChecksum(data []byte, csum uint32) uint16 {
	// to handle odd lengths, we loop to length - 1, incrementing by 2, then
	// handle the last byte specifically by checking against the original
	// length.
	length := len(data) - 1
	for i := 0; i < length; i += 2 {
		// For our test packet, doing this manually is about 25% faster
		// (740 ns vs. 1000ns) than doing it by calling binary.BigEndian.Uint16.
		csum += uint32(data[i]) << 8
		csum += uint32(data[i+1])
	}
	if len(data)%2 == 1 {
		csum += uint32(data[length]) << 8
	}
	for csum > 0xffff {
		csum = (csum >> 16) + (csum & 0xffff)
	}
	return ^uint16(csum)
}
