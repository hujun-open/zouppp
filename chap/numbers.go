package chap

import (
	"fmt"
)

// Code is the code of CHAP msg
type Code uint8

// list of CHAPCode
const (
	CodeChallenge Code = 1
	CodeResponse  Code = 2
	CodeSuccess   Code = 3
	CodeFailure   Code = 4
)

// String returns a string representation of c
func (c Code) String() string {
	switch c {
	case CodeChallenge:
		return "Challenge"
	case CodeResponse:
		return "Response"
	case CodeSuccess:
		return "Success"
	case CodeFailure:
		return "Failure"
	}
	return fmt.Sprintf("unknown (%d)", uint8(c))
}
