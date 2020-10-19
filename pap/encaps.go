package pap

import (
	"encoding/binary"
	"fmt"
)

// Pkt represents a PAP pkt
type Pkt struct {
	Code   Code
	ID     uint8
	Len    uint16
	PeerID []byte
	Passwd []byte
	Msg    []byte
}

// Parse buf into pp
func (pp *Pkt) Parse(buf []byte) error {
	if len(buf) < 4 {
		return fmt.Errorf("not enough bytes for a PAP pkt %d", len(buf))
	}
	pp.Code = Code(buf[0])
	pp.ID = uint8(buf[1])
	pp.Len = binary.BigEndian.Uint16(buf[2:4])
	switch pp.Code {
	case CodeAuthRequest:
		if pp.Len < 6 {
			return fmt.Errorf("invalid PAP Auth request length %d", pp.Len)
		}
		if buf[4] == 0 || buf[4] > 249 {
			return fmt.Errorf("invalid Peer ID length %d", buf[4])
		}
		pp.PeerID = buf[5 : 5+buf[4]]
		if buf[5+buf[4]] == 0 || buf[5+buf[4]] > 249 {
			return fmt.Errorf("invalid Passwd length %d", buf[5+buf[4]])
		}
		pp.Passwd = buf[6+buf[4]:]
	default:
		if buf[4] > 250 {
			return fmt.Errorf("invalid msg length %d", buf[4])
		}
		pp.Msg = buf[5:]
	}
	return nil
}

// Serialize pp to byte slice
func (pp *Pkt) Serialize() ([]byte, error) {
	var buf []byte
	header := make([]byte, 4)
	header[0] = uint8(pp.Code)
	header[1] = pp.ID
	var totallen int
	switch pp.Code {
	case CodeAuthRequest:
		if len(pp.PeerID) > 249 || len(pp.PeerID) == 0 {
			return nil, fmt.Errorf("peer ID is either empty or too long")
		}
		if len(pp.Passwd) > 249 {
			return nil, fmt.Errorf("passwd is too long")
		}
		buf = append(header, byte(len(pp.PeerID)))
		buf = append(buf, pp.PeerID...)
		buf = append(buf, byte(len(pp.Passwd)))
		buf = append(buf, pp.Passwd...)
		totallen = 6 + len(pp.PeerID) + len(pp.Passwd)
	default:
		if len(pp.Msg) > 250 {
			return nil, fmt.Errorf("Msg is too long")
		}
		buf = append(header, byte(len(pp.Msg)))
		buf = append(buf, pp.Msg...)
		totallen = 5 + len(pp.Msg)

	}
	if totallen > 65535 {
		return nil, fmt.Errorf("result pkt too big")
	}
	binary.BigEndian.PutUint16(buf[2:4], uint16(totallen))
	return buf, nil
}

// String returns a string representation of pp
func (pp Pkt) String() string {
	s := fmt.Sprintf("Code: %v\n", pp.Code)
	s += fmt.Sprintf("ID: %d\n", pp.ID)
	switch pp.Code {
	case CodeAuthRequest:
		s += fmt.Sprintf("PeerID: %v\n", string(pp.PeerID))
		s += fmt.Sprintf("Passwd: %v\n", string(pp.Passwd))
	default:
		s += fmt.Sprintf("Msg: %v\n", string(pp.Msg))
	}
	return s
}
