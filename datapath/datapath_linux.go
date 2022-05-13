// Package datapath implements linux data path for PPPoE/PPP;
// 	TODO: currently datapath does NOT do following:
// 		- create default route with nexthop as the TUN interface
// 		- apply DNS server address
package datapath

import (
	"context"
	"fmt"
	"net"

	"github.com/hujun-open/zouppp/lcp"

	"github.com/songgao/water"
	"github.com/vishvananda/netlink"
	"go.uber.org/zap"
)

// TUNIF is the TUN interface for a opened PPP session
type TUNIF struct {
	intf                   *water.Interface
	nlink                  netlink.Link
	sendChan               chan []byte
	v4recvChan, v6recvChan chan []byte
	maxFrameSize           int
	logger                 *zap.Logger
	ownV4Addr              net.IP
}

// DefaultMaxFrameSize is the default max PPP frame size could be received from the TUN interface
const DefaultMaxFrameSize = 1500

// NewTUNIf creates a new TUN interface the pppproto, using name as interface name, add ifv4addr to the TUN interface;
// also creates an IPv6 link local address via v6ifid, set MTU to peermru;
func NewTUNIf(ctx context.Context, pppproto *lcp.PPP, name string, assignedAddrs []net.IP, v6ifid []byte, peermru uint16) (*TUNIF, error) {
	var err error
	r := new(TUNIF)
	cfg := water.Config{
		DeviceType: water.TUN,
	}
	cfg.Name = name
	r.intf, err = water.New(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create TUN if %v, %w", cfg.Name, err)
	}
	r.nlink, _ = netlink.LinkByName(name)
	err = netlink.LinkSetUp(r.nlink)
	if err != nil {
		return nil, fmt.Errorf("failed to bring the TUN if %v up, %w", cfg.Name, err)
	}
	//add v4 addr
	for _, addr := range assignedAddrs {
		if addr == nil {
			continue
		}
		if !addr.IsUnspecified() {
			plen := "128"
			if addr.To4() != nil {
				r.ownV4Addr = addr
				plen = "32"
				r.sendChan, r.v4recvChan = pppproto.Register(lcp.ProtoIPv4)
			}
			addrstr := fmt.Sprintf("%v/%v", addr, plen)
			naddr, err := netlink.ParseAddr(addrstr)
			if err != nil {
				return nil, fmt.Errorf("failed to parse %v as IP addr, %w", addrstr, err)
			}
			err = netlink.AddrAdd(r.nlink, naddr)
			if err != nil {
				return nil, fmt.Errorf("failed to add addr %v, %w", addrstr, err)
			}
		}
	}
	// if !ifv4addr.IsUnspecified() && ifv4addr != nil {
	// 	r.ownV4Addr = ifv4addr
	// 	addrstr := fmt.Sprintf("%v/32", r.ownV4Addr)
	// 	v4addr, err := netlink.ParseAddr(addrstr)
	// 	if err != nil {
	// 		return nil, fmt.Errorf("failed to parse %v as v4 addr, %w", addrstr, err)
	// 	}
	// 	err = netlink.AddrAdd(r.nlink, v4addr)
	// 	if err != nil {
	// 		return nil, fmt.Errorf("failed to add v4 addr %v, %w", addrstr, err)
	// 	}

	// 	r.sendChan, r.v4recvChan = pppproto.Register(lcp.ProtoIPv4)
	// }
	//add link local
	if v6ifid != nil {
		linklocaladdr := make([]byte, 16)
		copy(linklocaladdr[:8], lcp.IPv6LinkLocalPrefix[:8])
		copy(linklocaladdr[8:16], v6ifid[:8])
		addrstr := fmt.Sprintf("%v/64", net.IP(linklocaladdr).To16().String())
		lla, err := netlink.ParseAddr(addrstr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse %v as v6 addr, %w", addrstr, err)
		}
		lla.Scope = int(netlink.SCOPE_LINK)
		err = netlink.AddrAdd(r.nlink, lla)
		if err != nil {
			return nil, fmt.Errorf("failed to add v6 addr %v, %w", addrstr, err)
		}
		_, r.v6recvChan = pppproto.Register(lcp.ProtoIPv6)
	}

	//adjust mtu based on PPP peer's MRU
	mtu := int(peermru)
	if mtu < 1280 {
		mtu = 1280
	}
	netlink.LinkSetMTU(r.nlink, mtu)

	r.maxFrameSize = DefaultMaxFrameSize
	r.logger = pppproto.GetLogger().Named("datapath")
	go r.send(ctx)
	go r.recv(ctx)
	return r, nil
}

const minimalIPPktSize = 20 //ipv4 header

// send pkt to outside network
func (tif *TUNIF) send(ctx context.Context) {
	for {
		b := make([]byte, tif.maxFrameSize)
		n, err := tif.intf.Read(b)
		if err != nil {
			tif.logger.Sugar().Errorf("failed to read, %v", err)
			return
		}
		select {
		case <-ctx.Done():
			tif.logger.Info("send routine stopped")
			tif.intf.Close()
			return
		default:
		}
		if n < minimalIPPktSize {
			continue
		}
		switch b[0] >> 4 {
		case 4:
			pkt := lcp.NewPPPPkt(b[:n], lcp.ProtoIPv4)
			tif.sendChan <- pkt.Serialize()
		case 6:
			pkt := lcp.NewPPPPkt(b[:n], lcp.ProtoIPv6)
			tif.sendChan <- pkt.Serialize()
		default:
			continue
		}
	}
}

// recv getting pkt from outside network
func (tif *TUNIF) recv(ctx context.Context) {
	for {
		var pktbytes []byte

		select {
		case <-ctx.Done():
			tif.logger.Info("recv routine stopped")
			return
		case pktbytes = <-tif.v4recvChan:
			// gpacket := gopacket.NewPacket(pktbytes, layers.LayerTypeIPv4, gopacket.DecodeOptions{Lazy: true, NoCopy: true})
			// tif.logger.Sugar().Debugf("got a v4 pkt from outside:\n%v", gpacket.String())

		case pktbytes = <-tif.v6recvChan:
			// gpacket := gopacket.NewPacket(pktbytes, layers.LayerTypeIPv6, gopacket.DecodeOptions{Lazy: true, NoCopy: true})
			// tif.logger.Sugar().Debugf("got a v6 pkt from outside:\n%v", gpacket.String())
		}
		_, err := tif.intf.Write(pktbytes)
		if err != nil {
			tif.logger.Sugar().Error("failed to send to TUN interface, %v", err)
			return
		}
	}
}
