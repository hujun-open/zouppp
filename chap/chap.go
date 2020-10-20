// Package chap implments CHAPwithMD5 as specified in rfc1994
package chap

import (
	"crypto/md5"
	"fmt"
	"time"

	"github.com/hujun-open/zouppp/lcp"

	"go.uber.org/zap"
)

// CHAP is the CHAP protocol
type CHAP struct {
	peerID   string
	passwd   string
	sendChan chan []byte
	recvChan chan []byte
	logger   *zap.Logger
	timeout  time.Duration
}

// DefaultTimeout is the default timeout for CHAP
const DefaultTimeout = 10 * time.Second

// NewCHAP creates a new CHAP instance with specified uname,passwd; using pppProto as underlying PPP protocol
func NewCHAP(uname, passwd string, pppProto *lcp.PPP) *CHAP {
	r := new(CHAP)
	r.peerID = uname
	r.passwd = passwd
	r.sendChan, r.recvChan = pppProto.Register(lcp.ProtoCHAP)
	r.logger = pppProto.GetLogger()
	r.timeout = DefaultTimeout
	return r
}

func (chap *CHAP) send(p []byte) error {
	t := time.NewTimer(chap.timeout)
	defer t.Stop()
	ppkt := lcp.NewPPPPkt(p, lcp.ProtoCHAP)
	select {
	case <-t.C:
		return fmt.Errorf("send timeout")
	case chap.sendChan <- ppkt.Serialize():
	}
	return nil
}

func (chap *CHAP) getResponse(final bool) (*Pkt, error) {
	var pkt *Pkt
	var err error
	t := time.NewTimer(chap.timeout)
	defer t.Stop()
L1:
	for {
		select {
		case b := <-chap.recvChan:
			pkt = new(Pkt)
			err = pkt.Parse(b)
			if err != nil {
				chap.logger.Sugar().Warnf("got an invalid CHAP pkt,%v", err)
				continue L1
			}
			if !final {
				if pkt.Code == CodeChallenge {
					break L1
				}
			} else {
				if pkt.Code == CodeSuccess || pkt.Code == CodeFailure {
					break L1
				}
			}

		case <-t.C:
			return nil, fmt.Errorf("CHAP authentication failed, timeout")
		}
	}
	return pkt, nil

}

// AUTHSelf auth self to peer, return nil if auth succeeds
func (chap *CHAP) AUTHSelf() error {
	challenge, err := chap.getResponse(false)
	if err != nil {
		return err
	}
	chap.logger.Sugar().Debugf("got CHAP challenge:\n%v", challenge)
	resp := new(Pkt)
	resp.Code = CodeResponse
	resp.ID = challenge.ID

	h := md5.New()
	toBuf := append([]byte{challenge.ID}, []byte(chap.passwd)...)
	toBuf = append(toBuf, challenge.Value...)
	chap.logger.Sugar().Debugf("hashing id %x, passwd %s,challege %x", challenge.ID, chap.passwd, challenge.Value)
	h.Write(toBuf)
	resp.Value = h.Sum(nil)
	chap.logger.Sugar().Debugf("hash value is %x", resp.Value)
	//h.Write(append([]byte{challenge.ID}, []byte(chap.Passwd)...))
	//h.Write(challenge.Value)
	//resp.Value = h.Sum(nil)

	resp.Name = []byte(chap.peerID)
	b, err := resp.Serialize()
	if err != nil {
		return fmt.Errorf("failed to serialize CHAP response,%w", err)
	}
	err = chap.send(b)
	if err != nil {
		return fmt.Errorf("failed to send CHAP response,%w", err)
	}
	chap.logger.Sugar().Debugf("send CHAP response:\n%v", resp)
	finalresp, err := chap.getResponse(true)
	if err != nil {
		return fmt.Errorf("failed to get final CHAP response,%w", err)
	}
	chap.logger.Sugar().Debugf("got CHAP final response:\n%v", finalresp)
	if finalresp.Code == CodeFailure {
		return fmt.Errorf("gateway returned failed")
	}
	return nil
}
