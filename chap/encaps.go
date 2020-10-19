package chap

import (
	"encoding/binary"
	"fmt"
)

// Pkt represents a CHAP packet
type Pkt struct {
	Code   Code
	ID     uint8
	Len    uint16
	ValLen uint8
	Value  []byte
	Name   []byte
	Msg    []byte
}

// Parse buf into cp
func (cp *Pkt) Parse(buf []byte) error {
	if len(buf) < 4 {
		return fmt.Errorf("invalid CHAP packet length %d", len(buf))
	}
	cp.Code = Code(buf[0])
	cp.ID = uint8(buf[1])
	cp.Len = binary.BigEndian.Uint16(buf[2:4])
	if cp.Len < 4 {
		return fmt.Errorf("invalid CHAP packet length field %d", cp.Len)
	}
	switch cp.Code {
	case CodeChallenge, CodeResponse:
		if cp.Len < 5 {
			return fmt.Errorf("invalid CHAP challenge/response length")
		}
		cp.ValLen = buf[4]
		if cp.Len < uint16(4+cp.ValLen) {
			return fmt.Errorf("invalid CHAP challenge/response value len")
		}
		cp.Value = buf[5 : 5+cp.ValLen]
		cp.Name = buf[5+cp.ValLen : cp.Len]
	default:
		cp.Msg = buf[4:cp.Len]
	}
	return nil
}

// String returns a string representation of cp
func (cp Pkt) String() string {
	s := fmt.Sprintf("Code:%v\n", cp.Code)
	s += fmt.Sprintf("ID:%d\n", cp.ID)
	s += fmt.Sprintf("Len:%d\n", cp.Len)
	switch cp.Code {
	case CodeChallenge, CodeResponse:
		s += fmt.Sprintf("Val:%x\n", cp.Value)
		s += fmt.Sprintf("Name:%s\n", string(cp.Name))
	default:
		s += fmt.Sprintf("Msg:%s\n", string(cp.Msg))

	}
	return s
}

// Serialize cp into byte slice
func (cp *Pkt) Serialize() ([]byte, error) {
	var buf []byte
	header := make([]byte, 4)
	header[0] = uint8(cp.Code)
	header[1] = cp.ID
	switch cp.Code {
	case CodeChallenge, CodeResponse:
		if len(cp.Value) > 255 {
			return nil, fmt.Errorf("value too long")
		}
		buf = append(header, byte(len(cp.Value)))
		buf = append(buf, cp.Value...)
		buf = append(buf, cp.Name...)

	}
	totalen := 5 + len(cp.Name) + len(cp.Value)
	if totalen > 65535 {
		return nil, fmt.Errorf("result pkt too big")
	}
	binary.BigEndian.PutUint16(buf[2:4], uint16(totalen))
	return buf, nil
}
