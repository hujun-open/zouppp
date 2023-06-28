package lcp

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// Pkt represents a LCP/IPCP/IPv6CP pkt
type Pkt struct {
	// Proto is one of ProtoLCP, ProtoIPCP, ProtoIPv6CP
	Proto PPPProtocolNumber
	// Msg code
	Code MsgCode
	// Msg Id
	ID uint8
	// Msg length
	Len uint16
	// Magic Number if exists
	MagicNum uint32
	// rejected protocol number if exists
	RejectedProto PPPProtocolNumber
	//LCP allows mulitple instances of same type of option, and require same order between cfg-request/response
	Options []Option
	// pkt payload
	Payload []byte
}

// NewPkt return a new LCP/IPCP/IPv6CP Pkt based on p
func NewPkt(p PPPProtocolNumber) *Pkt {
	return &Pkt{Proto: p}
}

// Serialize into bytes, without copying, and no padding
func (p *Pkt) Serialize() ([]byte, error) {
	header := make([]byte, 4)
	header[0] = uint8(p.Code)
	header[1] = p.ID

	if p.Payload == nil {
		p.Payload = []byte{}
	}
	if p.Code != CodeEchoReply && p.Code != CodeEchoRequest {
		for _, op := range p.Options {
			buf, err := op.Serialize()
			if err != nil {
				return nil, err
			}
			p.Payload = append(p.Payload, buf...)
		}
		binary.BigEndian.PutUint16(header[2:4], uint16(4+len(p.Payload)))
		return append(header, p.Payload...), nil
	}
	//echo pkt
	binary.BigEndian.PutUint16(header[2:4], uint16(8+len(p.Payload)))
	mn := make([]byte, 4)
	binary.BigEndian.PutUint32(mn, p.MagicNum)
	return append(header, append(mn, p.Payload...)...), nil
}

const maxOptions = 32

// Parse buf into LCP
func (p *Pkt) Parse(buf []byte) error {
	if len(buf) < 4 {
		return fmt.Errorf("invalid PPP packet length %d", len(buf))
	}
	p.Code = MsgCode(buf[0])
	p.ID = buf[1]
	p.Len = binary.BigEndian.Uint16(buf[2:4])
	p.Payload = buf[4:p.Len]
	switch p.Code {
	case CodeConfigureRequest, CodeConfigureAck, CodeConfigureNak, CodeConfigureReject:
		p.Options = []Option{}
		newFunc := func(b byte) Option {
			switch p.Proto {
			case ProtoIPCP:
				switch IPCPOptionType(b) {
				case OpIPAddress, OpPrimaryDNSServerAddress, OpSecondaryDNSServerAddress, OpPrimaryNBNSServerAddress, OpSecondaryNBNSServerAddress:
					return new(IPv4AddrOption)
				default:
					return newIPCPGenericOption()
				}
			case ProtoIPv6CP:
				switch IPCP6OptionType(b) {
				case IP6CPOpInterfaceIdentifier:
					return new(InterfaceIDOption)
				default:
					return newIPv6CPGenericOption()
				}
			default:
				switch LCPOptionType(b) {
				case OpTypeAuthenticationProtocol:
					return new(LCPOpAuthProto)
				case OpTypeMagicNumber:
					return new(LCPOpMagicNum)
				case OpTypeMaximumReceiveUnit:
					return new(LCPOpMRU)
				default:
					return newLCPGenericOption()
				}
			}
		}
		if len(p.Payload) > 0 {
			pos := 0
			var i int
			for i = 0; i < maxOptions; i++ {
				op := newFunc(p.Payload[pos])
				if op == nil {
					break
				}
				n, err := op.Parse(p.Payload[pos:])
				if err != nil {
					return fmt.Errorf("failed to parse LCP option #%d %v, %w", i+1, LCPOptionType(p.Payload[pos]), err)
				}
				pos += n
				p.Options = append(p.Options, op)
				if pos >= len(p.Payload) {
					break
				}
			}
			if i == maxOptions {
				return fmt.Errorf("invalid LCP packe, exceed max number of options: %d", maxOptions)
			}
		}
	case CodeEchoRequest, CodeEchoReply, CodeDiscardRequest:
		if len(buf) < 8 {
			return fmt.Errorf("not enough bytes for a LCP echo pkt, %v", buf)
		}
		p.MagicNum = binary.BigEndian.Uint32(buf[4:8])
		p.Payload = buf[8:]
	case CodeTerminateAck, CodeTerminateRequest, CodeCodeReject:
	case CodeProtocolReject:
		if len(buf) < 6 {
			return fmt.Errorf("not enough bytes for a LCP protocol reject pkt, %v", buf)
		}
		p.RejectedProto = PPPProtocolNumber(binary.BigEndian.Uint16(buf[4:6]))
	}
	return nil
}

// String return a string representation of p
func (p Pkt) String() string {
	s := fmt.Sprintf("%v Code:%v\n", p.Proto, p.Code)
	s += fmt.Sprintf("ID:%d\n", p.ID)
	// s += fmt.Sprintf("Len:%d\n", lcp.Len)
	s += "Options:\n"
	switch p.Code {
	case CodeEchoReply, CodeEchoRequest, CodeDiscardRequest:
		s += fmt.Sprintf("Magic Number:%x\n", p.MagicNum)
	case CodeProtocolReject:
		s += fmt.Sprintf("Rejected Protocol: %v\n", p.RejectedProto)
	case CodeTerminateAck, CodeTerminateRequest:
		s += fmt.Sprintf("Data: %v\n", string(p.Payload))
	case CodeCodeReject:
	default:
		for _, op := range p.Options {
			s += fmt.Sprintf("    %v\n", op.String())
		}

	}
	return s

}

// GetOption return a slice of options with type as optype
func (p *Pkt) GetOption(optype LCPOptionType) (r []Option) {
	for _, o := range p.Options {
		if o.Type() == uint8(optype) {
			r = append(r, o)
		}
	}
	return
}

// Option is the LCP/IPCP/IPv6 option interface
type Option interface {
	// Serialize option into bytes
	Serialize() ([]byte, error)
	// Parse buf into the option, return length of used bytes
	Parse(buf []byte) (int, error)
	// return option type as uint8
	Type() uint8
	// return payload bytes
	GetPayload() []byte
	// String returns a string representation of the option
	String() string
	// return true if b has same value and type
	Equal(b Option) bool
}

// LCPOpMRU is the LCP MRU option
type LCPOpMRU uint16

// Type implements Option interface
func (mru LCPOpMRU) Type() uint8 {
	return uint8(OpTypeMaximumReceiveUnit)
}

// Serialize implements Option interface
func (mru LCPOpMRU) Serialize() ([]byte, error) {
	buf := make([]byte, 4)
	buf[0] = 1
	buf[1] = 4
	binary.BigEndian.PutUint16(buf[2:4], uint16(mru))
	return buf, nil
}

// GetPayload implements Option interface
func (mru LCPOpMRU) GetPayload() []byte {
	r := make([]byte, 2)
	binary.BigEndian.PutUint16(r, uint16(mru))
	return r
}

// Equal implements Option interface
func (mru LCPOpMRU) Equal(b Option) bool {
	return mru == *(b.(*LCPOpMRU))
}

// Parse implements Option interface
func (mru *LCPOpMRU) Parse(buf []byte) (int, error) {
	if buf[0] != byte(OpTypeMaximumReceiveUnit) || buf[1] != 4 {
		return 0, fmt.Errorf("not a valid %v option", OpTypeMaximumReceiveUnit)
	}
	*mru = LCPOpMRU(binary.BigEndian.Uint16(buf[2:4]))
	return 4, nil
}

// String implements Option interface
func (mru LCPOpMRU) String() string {
	return fmt.Sprintf("%v:%d", OpTypeMaximumReceiveUnit, uint16(mru))
}

// LCPOpAuthProto is the LCP auth protocol option
type LCPOpAuthProto struct {
	Proto   PPPProtocolNumber
	CHAPAlg CHAPAuthAlg
	Payload []byte
}

// NewPAPAuthOp returns a new PAP LCPOpAuthProto
func NewPAPAuthOp() *LCPOpAuthProto {
	return &LCPOpAuthProto{Proto: ProtoPAP}
}

// NewCHAPAuthOp returns a new CHAP LCPOpAuthProto with MD5
func NewCHAPAuthOp() *LCPOpAuthProto {
	return &LCPOpAuthProto{Proto: ProtoCHAP, CHAPAlg: AlgCHAPwithMD5}
}

// Type implements Option interface
func (authp *LCPOpAuthProto) Type() uint8 {
	return uint8(OpTypeAuthenticationProtocol)
}

// Equal implements Option interface
func (authp LCPOpAuthProto) Equal(b Option) bool {
	if authp.Proto == ProtoCHAP {
		return b.(*LCPOpAuthProto).Proto == ProtoCHAP && authp.CHAPAlg == b.(*LCPOpAuthProto).CHAPAlg
	}
	return authp.Proto == b.(*LCPOpAuthProto).Proto
}

// Serialize implements Option interface
func (authp *LCPOpAuthProto) Serialize() ([]byte, error) {
	if authp.Proto == ProtoNone {
		// means no auth
		return []byte{}, nil
	}
	if len(authp.Payload) > 251 {
		return nil, fmt.Errorf("payload of %v is too long", OpTypeAuthenticationProtocol)
	}
	buf := make([]byte, 4)
	buf[0] = 3
	buf[1] = byte(4 + len(authp.Payload))
	binary.BigEndian.PutUint16(buf[2:4], uint16(authp.Proto))
	switch authp.Proto {
	case ProtoCHAP:
		if authp.CHAPAlg != AlgNone {
			buf = append(buf, byte(authp.CHAPAlg))
			buf[1] = 5
			return buf, nil
		}
	case ProtoPAP:
		buf[1] = 4
		return buf, nil
	}

	return append(buf, authp.Payload...), nil
}

// Parse implements Option interface
func (authp *LCPOpAuthProto) Parse(buf []byte) (int, error) {
	if len(buf) < 4 {
		return 0, fmt.Errorf("not enough bytes to parse an Auth-Protocol option")
	}
	if buf[0] != byte(OpTypeAuthenticationProtocol) {
		return 0, fmt.Errorf("not a valid %v option", OpTypeAuthenticationProtocol)
	}
	authp.Proto = PPPProtocolNumber(binary.BigEndian.Uint16(buf[2:4]))
	if buf[1] > 4 && authp.Proto == ProtoCHAP {
		authp.CHAPAlg = CHAPAuthAlg(buf[4])
	} else {
		authp.CHAPAlg = AlgNone
	}
	authp.Payload = buf[4:buf[1]]
	return int(buf[1]), nil
}

// GetPayload implements Option interface
func (authp *LCPOpAuthProto) GetPayload() []byte {
	return authp.Payload
}

// String implements Option interface
func (authp LCPOpAuthProto) String() string {
	switch authp.Proto {
	case ProtoNone:
		return ""
	case ProtoCHAP:
		return fmt.Sprintf("%v:%v alg:%v", OpTypeAuthenticationProtocol, authp.Proto, authp.CHAPAlg)
	}
	return fmt.Sprintf("%v:%v (%d)", OpTypeAuthenticationProtocol, authp.Proto, len(authp.Payload))
}

// LCPOpMagicNum is the LCP magic number option
type LCPOpMagicNum uint32

// Type implements Option interface
func (mn LCPOpMagicNum) Type() uint8 {
	return uint8(OpTypeMagicNumber)
}

// Serialize implements Option interface
func (mn LCPOpMagicNum) Serialize() ([]byte, error) {
	buf := make([]byte, 6)
	buf[0] = byte(OpTypeMagicNumber)
	buf[1] = 6
	binary.BigEndian.PutUint32(buf[2:6], uint32(mn))
	return buf, nil
}

// GetPayload implements Option interface
func (mn LCPOpMagicNum) GetPayload() []byte {
	r := make([]byte, 4)
	binary.BigEndian.PutUint32(r, uint32(mn))
	return r
}

// Parse implements Option interface
func (mn *LCPOpMagicNum) Parse(buf []byte) (int, error) {
	if buf[0] != byte(OpTypeMagicNumber) || buf[1] != 6 {
		return 0, fmt.Errorf("not a valid %v option", OpTypeMagicNumber)
	}
	*mn = LCPOpMagicNum(binary.BigEndian.Uint32(buf[2:6]))
	return 6, nil
}

// String implements Option interface
func (mn LCPOpMagicNum) String() string {
	return fmt.Sprintf("%v:%x", OpTypeMagicNumber, uint32(mn))
}

// Equal implements Option interface
func (mn LCPOpMagicNum) Equal(b Option) bool {
	return mn == *(b.(*LCPOpMagicNum))
}

// GenericOption is general LCP/IPCP/IPv6CP option that doesn't have explicit support
type GenericOption struct {
	code    uint8
	payload []byte
	proto   PPPProtocolNumber
}

// NewGenericOption creates a new GenericOption with p as the specified protocol;
// only LCP/IPCP/IPv6CP are supported;
func NewGenericOption(p PPPProtocolNumber) (*GenericOption, error) {
	r := new(GenericOption)
	switch p {
	case ProtoLCP, ProtoIPCP, ProtoIPv6CP:
		r.proto = p
		return r, nil
	}
	return nil, fmt.Errorf("unsupported PPP protocol %v", p)
}

func newLCPGenericOption() *GenericOption {
	r := new(GenericOption)
	r.proto = ProtoLCP
	return r
}

func newIPCPGenericOption() *GenericOption {
	r := new(GenericOption)
	r.proto = ProtoIPCP
	return r
}

func newIPv6CPGenericOption() *GenericOption {
	r := new(GenericOption)
	r.proto = ProtoIPv6CP
	return r
}

// Serialize implements Option interface
func (gop GenericOption) Serialize() ([]byte, error) {
	header := make([]byte, 2)
	header[0] = gop.code
	if 2+len(gop.payload) > 255 {
		return nil, fmt.Errorf("option payload is too big")
	}
	header[1] = byte(2 + len(gop.payload))
	return append(header, gop.payload...), nil
}

// Parse implements Option interface
func (gop *GenericOption) Parse(buf []byte) (int, error) {
	if len(buf) < 2 {
		return 0, fmt.Errorf("not enough bytes")
	}
	if buf[1] < 2 {
		return 0, fmt.Errorf("invalid length field")
	}
	gop.code = buf[0]
	gop.payload = buf[2:buf[1]]
	return int(buf[1]), nil

}

// Type implements Option interface
func (gop GenericOption) Type() uint8 {
	return gop.code
}

// String implements Option interface
func (gop GenericOption) String() string {
	switch gop.proto {
	case ProtoIPCP:
		return fmt.Sprintf("option %v: %v", IPCPOptionType(gop.code), gop.payload)
	case ProtoIPv6CP:
		return fmt.Sprintf("option %v: %v", IPCP6OptionType(gop.code), gop.payload)
	}
	return fmt.Sprintf("option %v: %v", LCPOptionType(gop.code), gop.payload)
}

// GetPayload implements Option interface
func (gop GenericOption) GetPayload() []byte {
	return gop.payload
}

// Equal implements Option interface
func (gop GenericOption) Equal(b Option) bool {
	return bytes.Equal(gop.payload, b.GetPayload())
}
