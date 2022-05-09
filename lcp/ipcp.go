package lcp

import (
	"fmt"
	"net"
	"sync"
)

// IPv4AddrOption represents IPCP option contains a single v4 addr like IP-Addr, DNS
type IPv4AddrOption struct {
	Addr     net.IP
	AddrType IPCPOptionType
}

// Serialize implments Option interface
func (addr *IPv4AddrOption) Serialize() ([]byte, error) {
	buf := make([]byte, 6)
	buf[0] = byte(addr.AddrType)
	buf[1] = 6
	copy(buf[2:6], addr.Addr.To4()[:4])
	return buf, nil
}

// Parse implments Option interface
func (addr *IPv4AddrOption) Parse(buf []byte) (int, error) {
	if len(buf) < 6 {
		return 0, fmt.Errorf("not enough bytes")
	}
	if buf[1] != 6 {
		return 0, fmt.Errorf("len field is not 6")
	}
	addr.AddrType = IPCPOptionType(buf[0])
	addr.Addr = net.IP(buf[2:6])
	return 6, nil
}

// Equal implments Option interface
func (addr *IPv4AddrOption) Equal(b Option) bool {
	return addr.Addr.Equal(b.(*IPv4AddrOption).Addr)
}

// Type implments Option interface
func (addr *IPv4AddrOption) Type() uint8 {
	return uint8(addr.AddrType)
}

// String implments Option interface
func (addr IPv4AddrOption) String() string {
	return fmt.Sprintf("%v:%v", addr.AddrType, addr.Addr)
}

// GetPayload implments Option interface
func (addr IPv4AddrOption) GetPayload() []byte {
	return []byte(addr.Addr.To4()[:4])
}

// NewAddrOp returns a new IPCP address option, t specifies the address type
func NewAddrOp(ip net.IP, t IPCPOptionType) Option {
	return &IPv4AddrOption{
		AddrType: t,
		Addr:     ip,
	}
}

// DefaultIPCPOwnRule is the default OwnOptionRule for the IPCP protocol,
// it implements OwnOptionRule interface
type DefaultIPCPOwnRule struct {
	Addr          net.IP
	DNS           net.IP
	SecondaryDNS  net.IP
	NBNS          net.IP
	SecondaryNBNS net.IP
	mux           *sync.RWMutex
}

// GetOptions implements OwnOptionRule interface; a field will not be included as own option if it is nil
func (own *DefaultIPCPOwnRule) GetOptions() Options {
	own.mux.RLock()
	defer own.mux.RUnlock()
	r := Options{}
	if own.Addr != nil {
		r = append(r, NewAddrOp(own.Addr, OpIPAddress))
	}
	if own.DNS != nil {
		r = append(r, NewAddrOp(own.DNS, OpPrimaryDNSServerAddress))
	}
	if own.SecondaryDNS != nil {
		r = append(r, NewAddrOp(own.SecondaryDNS, OpSecondaryDNSServerAddress))
	}
	if own.NBNS != nil {
		r = append(r, NewAddrOp(own.NBNS, OpPrimaryNBNSServerAddress))
	}
	if own.SecondaryNBNS != nil {
		r = append(r, NewAddrOp(own.SecondaryNBNS, OpSecondaryNBNSServerAddress))
	}
	return r
}

// GetOption implements OwnOptionRule interface;
func (own *DefaultIPCPOwnRule) GetOption(o uint8) Option {
	own.mux.RLock()
	defer own.mux.RUnlock()
	switch IPCPOptionType(o) {
	case OpIPAddress:
		return NewAddrOp(own.Addr, OpIPAddress)
	case OpPrimaryDNSServerAddress:
		return NewAddrOp(own.DNS, OpPrimaryDNSServerAddress)
	case OpSecondaryDNSServerAddress:
		return NewAddrOp(own.SecondaryDNS, OpSecondaryDNSServerAddress)
	case OpPrimaryNBNSServerAddress:
		return NewAddrOp(own.NBNS, OpPrimaryNBNSServerAddress)
	case OpSecondaryNBNSServerAddress:
		return NewAddrOp(own.SecondaryNBNS, OpSecondaryNBNSServerAddress)
	}
	return nil
}

// HandlerConfRej implements OwnOptionRule interface;
// option in conf-reject will not be included in next conf-req;
func (own *DefaultIPCPOwnRule) HandlerConfRej(rcvd Options) {
	own.mux.Lock()
	defer own.mux.Unlock()
	for _, o := range rcvd {
		switch IPCPOptionType(o.Type()) {
		case OpIPAddress:
			own.Addr = nil
		case OpPrimaryDNSServerAddress:
			own.DNS = nil
		case OpSecondaryDNSServerAddress:
			own.SecondaryDNS = nil
		case OpPrimaryNBNSServerAddress:
			own.NBNS = nil
		case OpSecondaryNBNSServerAddress:
			own.SecondaryNBNS = nil
		}
	}
}

// HandlerConfNAK implements OwnOptionRule interface;
// option value in received conf-nak will be used as own value in next conf-req;
func (own *DefaultIPCPOwnRule) HandlerConfNAK(rcvd Options) {
	own.mux.Lock()
	defer own.mux.Unlock()
	for _, o := range rcvd {
		switch IPCPOptionType(o.Type()) {
		case OpIPAddress:
			own.Addr = o.(*IPv4AddrOption).Addr
		case OpPrimaryDNSServerAddress:
			own.DNS = o.(*IPv4AddrOption).Addr
		case OpSecondaryDNSServerAddress:
			own.SecondaryDNS = o.(*IPv4AddrOption).Addr
		case OpPrimaryNBNSServerAddress:
			own.NBNS = o.(*IPv4AddrOption).Addr
		case OpSecondaryNBNSServerAddress:
			own.SecondaryNBNS = o.(*IPv4AddrOption).Addr
		}
	}
}

// NewDefaultIPCPOwnRule returns a new DefaultIPCPOwnRule, with all address set to 0.0.0.0
func NewDefaultIPCPOwnRule() *DefaultIPCPOwnRule {
	r := new(DefaultIPCPOwnRule)
	r.Addr = net.ParseIP("0.0.0.0")
	r.DNS = net.ParseIP("0.0.0.0")
	r.SecondaryDNS = net.ParseIP("0.0.0.0")
	r.NBNS = net.ParseIP("0.0.0.0")
	r.SecondaryNBNS = net.ParseIP("0.0.0.0")
	r.mux = new(sync.RWMutex)
	return r
}

// DefaultIPCPPeerRule implments PeerOptionRule interface;
// it ignores all peer options
type DefaultIPCPPeerRule struct{}

// GetOptions implments PeerOptionRule interface;
// always return nil
func (peer *DefaultIPCPPeerRule) GetOptions() Options {
	return nil
}

// HandlerConfReq implments PeerOptionRule interface;
// it will reject any options other than OpIPAddress, and ACK any OpIPAddress value;
func (peer *DefaultIPCPPeerRule) HandlerConfReq(rcvd Options) (nak, reject Options) {
	for _, o := range rcvd {
		switch IPCPOptionType(o.Type()) {
		case OpIPAddress:
		default:
			reject = append(reject, o)
		}
	}
	return
}
