package lcp

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"net"
	"sync"
)

// InterfaceIDOption is the IPv6CP interface ID option
type InterfaceIDOption [8]byte

// Serialize implements Option interface
func (ifid *InterfaceIDOption) Serialize() ([]byte, error) {
	buf := make([]byte, 10)
	buf[0] = byte(IP6CPOpInterfaceIdentifier)
	buf[1] = 10
	copy(buf[2:], ifid[:])
	return buf, nil
}

// Parse implements Option interface
func (ifid *InterfaceIDOption) Parse(buf []byte) (int, error) {
	if len(buf) < 10 {
		return 0, fmt.Errorf("not enough bytes")
	}
	if buf[1] != 10 {
		return 0, fmt.Errorf("len field is not 10")
	}
	copy(ifid[:], buf[2:10])
	return 10, nil
}

// Equal implements Option interface
func (ifid *InterfaceIDOption) Equal(b Option) bool {
	return bytes.Equal(b.(*InterfaceIDOption)[:], ifid[:])
}

// Type implements Option interface
func (ifid *InterfaceIDOption) Type() uint8 {
	return uint8(IP6CPOpInterfaceIdentifier)
}

// String implements Option interface
func (ifid InterfaceIDOption) String() string {
	vals := ""
	for i := 0; i < 8; i += 2 {
		vals += fmt.Sprintf("%x%x", ifid[i], ifid[1+i])
		if i < 6 {
			vals += ":"
		}
	}
	return fmt.Sprintf("%v: %v", IP6CPOpInterfaceIdentifier, vals)
}

// GetPayload implements Option interface
func (ifid *InterfaceIDOption) GetPayload() []byte {
	return ifid[:]
}

// DefaultIP6CPRule implements both OwnOptionRule and PeerOptionRule interface;
// only negotiate interface-id option
type DefaultIP6CPRule struct {
	IfID     *InterfaceIDOption
	ifidChan chan *InterfaceIDOption
	mux      *sync.RWMutex
}

// NewDefaultIP6CPRule returns a new DefaultIP6CPRule;
// using a interface-id option that is derived from the mac;
func NewDefaultIP6CPRule(ctx context.Context, mac net.HardwareAddr) *DefaultIP6CPRule {
	r := new(DefaultIP6CPRule)
	r.mux = new(sync.RWMutex)
	r.ifidChan = make(chan *InterfaceIDOption)
	go r.genLCPInterfaceIDOptionByRFC7217(ctx, mac)
	r.IfID = <-r.ifidChan
	return r
}

const rfc7217Key = "mysekey9823718dasdf902klsd"

// IPv6LinkLocalPrefix is the IPv6 Link Local prefix
var IPv6LinkLocalPrefix = []byte{0xfe, 0x80, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}

func (r *DefaultIP6CPRule) genLCPInterfaceIDOptionByRFC7217(ctx context.Context, mac net.HardwareAddr) {
	var dudcount byte
	for {
		h := sha256.New()
		h.Write(IPv6LinkLocalPrefix)
		h.Write([]byte(mac))
		h.Write([]byte{dudcount})
		h.Write([]byte(rfc7217Key))
		newifid := InterfaceIDOption{}
		copy(newifid[:], h.Sum(nil)[:8])
		select {
		case r.ifidChan <- &newifid:
		case <-ctx.Done():
			return
		}
		dudcount++
	}

}

// GetOptions implements OwnOptionRule interface, return own interface id
func (r *DefaultIP6CPRule) GetOptions() Options {
	r.mux.RLock()
	defer r.mux.RUnlock()
	if r.IfID == nil {
		return Options{}
	}
	return Options{r.IfID}
}

// GetOption implements OwnOptionRule interface, return nil if t is not interface-id
func (r *DefaultIP6CPRule) GetOption(t uint8) Option {
	r.mux.RLock()
	defer r.mux.RUnlock()
	switch IPCP6OptionType(t) {
	case IP6CPOpInterfaceIdentifier:
		return r.IfID
	}
	return nil
}

// HandlerConfRej implements OwnOptionRule interface, if interface-id is rejected, then setting own interface-id to nil
func (r *DefaultIP6CPRule) HandlerConfRej(rcvd Options) {
	r.mux.Lock()
	defer r.mux.Unlock()
	if len(rcvd.Get(uint8(IP6CPOpInterfaceIdentifier))) > 0 {
		r.IfID = nil
		return
	}
}

// HandlerConfNAK implements OwnOptionRule interface, generate a new inteface-id if interface-id is naked,
func (r *DefaultIP6CPRule) HandlerConfNAK(rcvd Options) {
	r.mux.Lock()
	defer r.mux.Unlock()
	if len(rcvd.Get(uint8(IP6CPOpInterfaceIdentifier))) > 0 {
		// r.IfID = o.(*LCPInterfaceIDOption) //use peer suggest value
		r.IfID = <-r.ifidChan //generate a new ifid
		return
	}
}

var allZeroIfID = InterfaceIDOption([8]byte{0, 0, 0, 0, 0, 0, 0, 0})

// HandlerConfReq implements PeerOptionRule interface, follow section 4.1 of RFC5072 in terms nak or reject peer's inteface-id
func (r *DefaultIP6CPRule) HandlerConfReq(rcvd Options) (nak, reject Options) {
	r.mux.RLock()
	defer r.mux.RUnlock()
	derivePeerIfIDFunc := func(orig *InterfaceIDOption) *InterfaceIDOption {
		// derive a new ifid based own ifid, by xor 0xff with last byte
		newifidop := new(InterfaceIDOption)
		copy(newifidop[:], orig[:])
		newifidop[7] = newifidop[7] ^ 0xff
		return newifidop
	}
	for _, o := range rcvd {
		if o.Type() == uint8(IP6CPOpInterfaceIdentifier) {
			if o.Equal(&allZeroIfID) {
				if r.IfID.Equal(&allZeroIfID) {
					reject = append(reject, &allZeroIfID)
				} else {
					nak = append(nak, derivePeerIfIDFunc(r.IfID))
				}
			} else {
				if r.IfID.Equal(o) {
					nak = append(nak, derivePeerIfIDFunc(r.IfID))
				}
			}
		}
		return
	}
	return
}
