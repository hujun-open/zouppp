// pppoe_test
package pppoe

import (
	"encoding/hex"
	"fmt"
	"testing"
)

func TestEncap(t *testing.T) {
	pktstr := "116507a1002f01010000010300080200000025000000010200076164736c30373101040010b19214f4814f23b53e3691c395a98496"
	pktbytes, err := hex.DecodeString(pktstr)
	if err != nil {
		t.Fatal(err)
	}
	p := new(Pkt)
	err = p.Parse(pktbytes)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("\n%v", p)
	newbuf, _ := p.Serialize()
	if pktstr != fmt.Sprintf("%x", newbuf) {
		t.Fatalf("result of serialization doesn't matched expected result")
	}

}
