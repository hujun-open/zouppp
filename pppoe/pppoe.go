// Package pppoe implements pppoe as defined in RFC2516
package pppoe

import (
	"context"
	"encoding/binary"

	"errors"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/hujun-open/etherconn"
	"go.uber.org/zap"
)

// PPPoE is the PPPoE protocol
type PPPoE struct {
	ifName      string
	serviceName string
	sessionID   uint16
	tags        []Tag
	acMAC       net.HardwareAddr
	conn        *etherconn.EtherConn
	wg          *sync.WaitGroup
	cancelFunc  context.CancelFunc
	debug       bool
	recvChan    chan []byte
	state       *uint32
	logger      *zap.Logger
	timeout     time.Duration
	retry       int
}

const (
	// DefaultTimeout is default timeout for PPPoE
	DefaultTimeout = 3 * time.Second
	// DefaultRetry is the default retry for PPPoE
	DefaultRetry = 3
)

// list of PPPoEState
const (
	pppoeStateInitial = iota
	pppoeStateDialing
	pppoeStateOpen
	pppoeStateClosed
)

const (
	// EtherTypePPPoESession is the Ether type for PPPoE session pkt
	EtherTypePPPoESession = 0x8864
	// EtherTypePPPoEDiscovery is the Ether type for PPPoE discovery pkt
	EtherTypePPPoEDiscovery = 0x8863
	recvChanDepth           = 32
	readTimeout             = time.Second
)

// Modifier is a function to provide custom configuration when creating new PPPoE instances
type Modifier func(pppoe *PPPoE)

// WithDebug specifies if debug is enabled
func WithDebug(d bool) Modifier {
	return func(pppoe *PPPoE) {
		pppoe.debug = d
	}
}

// WithTags adds all tags in t in PPPoE request pkt
func WithTags(t []Tag) Modifier {
	return func(pppoe *PPPoE) {
		if t == nil {
			return
		}
		pppoe.tags = t
	}
}

// NewPPPoE return a new PPPoE struct; use conn as underlying transport, logger for logging;
// optionally Modifer could provide custom configurations;
func NewPPPoE(conn *etherconn.EtherConn, logger *zap.Logger, options ...Modifier) *PPPoE {
	r := new(PPPoE)
	r.timeout = DefaultTimeout
	r.retry = DefaultRetry
	r.tags = []Tag{
		&TagString{
			TagType: TagTypeServiceName,
			Value:   r.serviceName,
		},
	}
	for _, option := range options {
		option(r)
	}
	r.state = new(uint32)
	*r.state = pppoeStateInitial
	r.wg = new(sync.WaitGroup)
	r.recvChan = make(chan []byte, recvChanDepth)
	r.conn = conn
	r.logger = logger
	return r
}

// SetReadDeadline implements net.PacketConn interface
func (pppoe *PPPoE) SetReadDeadline(t time.Time) error {
	return pppoe.conn.SetReadDeadline(t)
}

// SetWriteDeadline implements net.PacketConn interface
func (pppoe *PPPoE) SetWriteDeadline(t time.Time) error {
	return pppoe.conn.SetWriteDeadline(t)
}

// SetDeadline implements net.PacketConn interface
func (pppoe *PPPoE) SetDeadline(t time.Time) error {
	pppoe.SetReadDeadline(t)
	pppoe.SetWriteDeadline(t)
	return nil
}

// LocalAddr return local Endpoint, see doc of Endpoint
func (pppoe *PPPoE) LocalAddr() net.Addr {
	return newPPPoEEndpoint(pppoe.conn.LocalAddr(), pppoe.sessionID)
}

// Close implements net.PacketConn interface
func (pppoe *PPPoE) Close() error {
	if atomic.LoadUint32(pppoe.state) == pppoeStateOpen {
		pkt := pppoe.buildPADT()
		pktbytes, err := pkt.Serialize()
		if err != nil {
			return err
		}
		pppoe.conn.WritePktTo(pktbytes, EtherTypePPPoEDiscovery, pppoe.acMAC)
	}
	return nil
}

func (pppoe *PPPoE) buildPADT() *Pkt {
	return &Pkt{
		Code:      CodePADT,
		SessionID: pppoe.sessionID,
	}
}

func (pppoe *PPPoE) buildPADI() *Pkt {
	padi := new(Pkt)
	padi.Code = CodePADI
	padi.SessionID = 0
	padi.Tags = pppoe.tags
	return padi
}

func (pppoe *PPPoE) buildPADRWithPADO(pado *Pkt) *Pkt {
	padr := new(Pkt)
	padr.Code = CodePADR
	padr.SessionID = 0
	padr.Tags = []Tag{
		&TagString{
			TagType: TagTypeServiceName,
			Value:   pppoe.serviceName,
		},
	}
	for _, t := range pppoe.tags {
		if t.Type() != uint16(TagTypeServiceName) {
			padr.Tags = append(padr.Tags, t)
		}
	}
	padr.Tags = append(padr.Tags, pado.GetTag(TagTypeACCookie)...)
	return padr
}

// WriteTo implments net.PacketConn interface, addr is ignored, pkt is always sent to AC's MAC
func (pppoe *PPPoE) WriteTo(p []byte, addr net.Addr) (n int, err error) {
	if atomic.LoadUint32(pppoe.state) != pppoeStateOpen {
		return 0, fmt.Errorf("pppoe is not open")
	}
	pkt := new(Pkt)
	pkt.SessionID = pppoe.sessionID
	pkt.Code = CodeSession
	pkt.Payload = p
	pktbytes, err := pkt.Serialize()
	if err != nil {
		return 0, fmt.Errorf("failed to serialize pppoe pkt,%w", err)
	}
	_, err = pppoe.conn.WritePktTo(pktbytes, EtherTypePPPoESession, pppoe.acMAC)
	if err != nil {
		return 0, fmt.Errorf("failed to send pppoe pkt,%w", err)
	}
	return len(p), nil

}

// ReadFrom implments net.PacketConn interface; only works after pppoe session is open
func (pppoe *PPPoE) ReadFrom(buf []byte) (int, net.Addr, error) {
	if atomic.LoadUint32(pppoe.state) != pppoeStateOpen {
		return 0, nil, fmt.Errorf("pppoe is not open")
	}
	var remotemac net.HardwareAddr
	var err error
	var n int
	for {
		n, remotemac, err = pppoe.conn.ReadPktFrom(buf)
		if err != nil {
			return 0, nil, fmt.Errorf("failed to recv, %w", err)
		}
		if n < 6 {
			continue
		}
		if remotemac.String() != pppoe.acMAC.String() {
			continue
		}
		if Code(buf[1]) != CodeSession {
			continue
		}
		//TODO: change ehtherconn so that L2Endpoint become a interface, and so that pppoe sessionid could be included
		if binary.BigEndian.Uint16(buf[2:4]) != pppoe.sessionID {
			continue
		}
		buf = append(buf[:0], buf[6:]...)
		break
	}
	//return int(binary.BigEndian.Uint16(buf[4:6])), etherconn.NewL2EndpointFromMACVLAN(remotemac, pppoe.vlans), nil
	return n - 6, pppoe.newRemotePPPoEP(remotemac), nil
}

func (pppoe *PPPoE) newRemotePPPoEP(mac net.HardwareAddr) *Endpoint {
	l2ep := etherconn.L2Endpoint{
		HwAddr: mac,
		VLANs:  pppoe.conn.LocalAddr().VLANs,
	}
	return newPPPoEEndpoint(&l2ep, pppoe.sessionID)
}

// getResponse return 1st rcvd PPPoE response as specified by code, along with remote mac
func (pppoe *PPPoE) getResponse(req *Pkt, code Code, dst net.HardwareAddr) (*Pkt, net.HardwareAddr, error) {
	pktbytes, err := req.Serialize()
	if err != nil {
		return nil, nil, err
	}
	for i := 0; i < pppoe.retry; i++ {
		_, err = pppoe.conn.WritePktTo(pktbytes, EtherTypePPPoEDiscovery, dst)
		if err != nil {
			return nil, nil, err
		}
		pppoe.logger.Sugar().Infof("sending %v", req.Code)
		pppoe.logger.Sugar().Debugf("%v:\n%v", req.Code, req)
		resp := new(Pkt)
		pppoe.conn.SetReadDeadline(time.Now().Add(pppoe.timeout))
		rcvpktbuf, remotemac, err := pppoe.conn.ReadPkt()
		if err != nil {
			if !errors.Is(err, etherconn.ErrTimeOut) {
				return nil, nil, fmt.Errorf("failed to recv response, %w", err)
			}
		}
		err = resp.Parse(rcvpktbuf)
		if err != nil {
			continue
		}
		if resp.Code == code {
			return resp, remotemac, nil
		}
	}
	return nil, nil, fmt.Errorf("faile to recv expect response %v", code)
}

// GetLogger returns pppoe's logger
func (pppoe *PPPoE) GetLogger() *zap.Logger {
	return pppoe.logger
}

// Dial complets a full PPPoE discovery exchange (PADI/PADO/PADR/PADS)
func (pppoe *PPPoE) Dial(ctx context.Context) error {
	//build PADI
	atomic.StoreUint32(pppoe.state, pppoeStateDialing)
	defer func() {
		if atomic.LoadUint32(pppoe.state) != pppoeStateOpen {
			atomic.StoreUint32(pppoe.state, pppoeStateClosed)
		}
	}()
	var err error
	padi := pppoe.buildPADI()
	var pado, pads *Pkt
	pado, pppoe.acMAC, err = pppoe.getResponse(padi, CodePADO, etherconn.BroadCastMAC)
	if err != nil {
		return err
	}
	pppoe.logger.Info("Got PADO")
	pppoe.logger.Sugar().Debugf("PADO:\n%v", pado)
	padr := pppoe.buildPADRWithPADO(pado)
	pads, _, err = pppoe.getResponse(padr, CodePADS, pppoe.acMAC)
	if err != nil {
		return err
	}
	pppoe.logger.Info("Got PADS")
	pppoe.logger.Sugar().Debugf("PADS:\n%v", pads)
	if pads.SessionID == 0 {
		return fmt.Errorf("AC rejected,\n %v", pads.String())
	}
	pppoe.sessionID = pads.SessionID
	atomic.StoreUint32(pppoe.state, pppoeStateOpen)
	pppoe.logger = pppoe.logger.Named(fmt.Sprintf("%X", pppoe.sessionID))
	_, pppoe.cancelFunc = context.WithCancel(ctx)
	return nil
}

// Endpoint represents a PPPoE endpont
type Endpoint struct {
	// L2EP is the associated EtherConn's L2Endpoint
	L2EP *etherconn.L2Endpoint
	// SessionId is the PPPoE session ID
	SessionID uint16
}

// Network implenets net.Addr interface, always return "pppoe"
func (pep Endpoint) Network() string {
	return "pppoe"
}

// String implenets net.Addr interface, return "pppoe:<L2EP>:<SessionID>"
func (pep Endpoint) String() string {
	return fmt.Sprintf("pppoe:%v:%x", pep.L2EP.String(), pep.SessionID)
}

func newPPPoEEndpoint(l2ep *etherconn.L2Endpoint, sid uint16) *Endpoint {
	return &Endpoint{
		L2EP:      l2ep,
		SessionID: sid,
	}
}
