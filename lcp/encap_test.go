// test_encap
package lcp

import (
	"encoding/hex"
	"testing"
)

func TestLCP(t *testing.T) {
	lcppkt, err := hex.DecodeString("01100012010405d40304c023050642ae33170000000000000000")
	if err != nil {
		t.Fatal(err)
	}
	l := new(Pkt)
	err = l.Parse(lcppkt)
	if err != nil {
		t.Fatal(err)
	}
	if l.Code != CodeConfigureRequest {
		t.Fatal("wrong lcp code")
	}
	if *(l.GetOption(OpTypeMaximumReceiveUnit)[0].(*LCPOpMRU)) != LCPOpMRU(1492) {
		t.Fatal("wrong MRU")
	}
	if l.GetOption(OpTypeAuthenticationProtocol)[0].(*LCPOpAuthProto).Proto != ProtoPAP {
		t.Fatal("wrong auth proto")
	}
	if *(l.GetOption(OpTypeMagicNumber)[0].(*LCPOpMagicNum)) != 0x42ae3317 {
		t.Fatal("wrong magic number")
	}
	lencoded, err := l.Serialize()
	if err != nil {
		t.Fatal(err)
	}
	err = l.Parse(lencoded)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("\n%v", l)
}
