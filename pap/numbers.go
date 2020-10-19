package pap

import (
	"fmt"
)

// Code is the PAP msg code
type Code uint8

// a list of PAP msg code
const (
	CodeAuthRequest Code = 1
	CodeAuthACK     Code = 2
	CodeAuthNAK     Code = 3
)

// String return a string representation of c
func (c Code) String() string {
	switch c {
	case CodeAuthRequest:
		return "Auth-Request"
	case CodeAuthACK:
		return "Auth-ACK"
	case CodeAuthNAK:
		return "Auth-NAK"
	}
	return fmt.Sprintf("unknown (%d)", c)
}
