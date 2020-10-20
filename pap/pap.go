// Package pap implements PAP protocol as specified in RFC1334
package pap

import (
	"fmt"
	"time"

	"github.com/hujun-open/zouppp/lcp"

	"go.uber.org/zap"
)

// PAP is the PAP protocol
type PAP struct {
	peerID   string
	passwd   string
	sendChan chan []byte
	recvChan chan []byte
	logger   *zap.Logger
	timeout  time.Duration
	retry    int
	reqID    uint8
}

const (
	// DefaultTimeout is the default timeout for PAP
	DefaultTimeout = 5 * time.Second
	// DefaultRetry is the default retry for PAP
	DefaultRetry = 3
)

// NewPAP creates a new PAP instance with uname, passwd;
// uses pppProtol as the underlying PPP protocol;
func NewPAP(uname, passwd string, pppProto *lcp.PPP) *PAP {
	r := new(PAP)
	r.peerID = uname
	r.passwd = passwd
	r.sendChan, r.recvChan = pppProto.Register(lcp.ProtoPAP)
	r.logger = pppProto.GetLogger().Named("PAP")
	r.timeout = DefaultTimeout
	r.retry = DefaultRetry
	return r
}

func (pap *PAP) getResponse(req *Pkt) (*Pkt, error) {

	var t *time.Timer
	resp := new(Pkt)
	for i := 0; i < pap.retry; i++ {
		pap.reqID++
		req.ID = pap.reqID
		pktbytes, err := req.Serialize()
		if err != nil {
			return nil, err
		}
		ppkt := lcp.NewPPPPkt(pktbytes, lcp.ProtoPAP)
		pap.sendChan <- ppkt.Serialize()
		pap.logger.Sugar().Debugf("sent PAP auth request:\n%v", req)
		if t == nil {
			t = time.NewTimer(pap.timeout)
		}
		t.Reset(pap.timeout)
		select {
		case <-t.C:
		case rcvdbytes := <-pap.recvChan:
			err := resp.Parse(rcvdbytes)
			if err != nil {
				pap.logger.Sugar().Warnf("got a invalid PAP response, %v", err)
				continue
			}
			if resp.Code != CodeAuthACK && resp.Code != CodeAuthNAK {
				pap.logger.Sugar().Warnf("got a PAP non-response, %v", resp.Code)
				continue
			}
			pap.logger.Sugar().Debugf("got PAP response\n%v", resp)
			return resp, nil
		}
	}
	return nil, fmt.Errorf("timeout")
}

// AuthSelf autnenticate self to the peer
func (pap *PAP) AuthSelf() error {
	req := new(Pkt)
	req.Code = CodeAuthRequest
	req.PeerID = []byte(pap.peerID)
	req.Passwd = []byte(pap.passwd)
	resp, err := pap.getResponse(req)
	if err != nil {
		return err
	}
	if resp.Code == CodeAuthACK {
		return nil
	}
	return fmt.Errorf("auth failed")
}
