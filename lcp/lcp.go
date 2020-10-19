//Package lcp implements PPP, LCP, IPCP and IPv6CP
package lcp

import (
	"context"
	"fmt"
	"math/rand"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"go.uber.org/zap"
)

// LayerNotifyHandler is the handler function to handle Layer event (tlu/tld/tls/tlf as defined in RFC1661)
type LayerNotifyHandler func(ctx context.Context, evt LayerNotifyEvent)

// LCP is the struct for LCP/IPCP/IPv6CP
type LCP struct {
	protoType             PPPProtocolNumber //since lcp could be also used by IPCP
	state                 *uint32
	restartCount          *uint32
	maxRestart            uint32
	restartTimerDuration  time.Duration
	restartTimer          *time.Timer
	keepAliveTimer        *time.Timer
	keepAliveInterval     time.Duration
	cancellRestartTimer   context.CancelFunc
	cancellkeepAliveTimer context.CancelFunc
	sendChan              chan []byte
	recvChan              chan []byte
	requestIDChan         chan uint8
	requestID             uint8
	reqiestIDLock         *sync.RWMutex
	logger                *zap.Logger
	// OwnRule is the OwnOptionRule to handle own options
	OwnRule OwnOptionRule
	// PeerRule is the PeerOptionRule to handle peer's options
	PeerRule    PeerOptionRule
	layerNotify LayerNotifyHandler
}

const (
	// DefaultRestartCounter is the default LCP restart counter value
	DefaultRestartCounter = 3
	// DefaultRestartTimerDuration is the default restart timer
	DefaultRestartTimerDuration = 10 * time.Second
	// DefaultKeepAliveInterval is the default LCP keepalive interval to send
	DefaultKeepAliveInterval = 5 * time.Second
	// DefaultMRU is the default LCP MRU value
	DefaultMRU = 1500
	// DefaultAuthProto is the default auth protocol
	DefaultAuthProto = ProtoCHAP
	// DefaultMagicNum is the default LCP magic number
	DefaultMagicNum LCPOpMagicNum = 0
)

func newOwnDefaultOptions() (r Options) {
	defaultMRUOp := LCPOpMRU(DefaultMRU)
	magicNum := LCPOpMagicNum(rand.Uint32())

	r = Options{
		&defaultMRUOp,
		&magicNum,
		// &defaultAuthProto,
	}
	return
}

// NewLCP creates a new LCP/IPCP/IPv6CP according to the specific proto, runs over specified pppProto, calls h whenever there is layer event.
// optionly, LCPModifier(s) could be specified to change default config
func NewLCP(ctx context.Context, proto PPPProtocolNumber, pppProto *PPP, h LayerNotifyHandler, mods ...Modifier) *LCP {
	lcp := new(LCP)
	lcp.protoType = proto
	lcp.state = new(uint32)
	atomic.StoreUint32(lcp.state, uint32(StateInitial))
	lcp.maxRestart = DefaultRestartCounter
	lcp.restartCount = new(uint32)
	atomic.StoreUint32(lcp.restartCount, lcp.maxRestart)
	// lcp.currentOwnOptions = newDefaultOptions()
	lcp.OwnRule = NewDefaultOwnOptionRule()
	lcp.restartTimerDuration = DefaultRestartTimerDuration
	lcp.keepAliveInterval = DefaultKeepAliveInterval
	lcp.PeerRule, _ = NewDefaultPeerOptionRule(DefaultAuthProto)
	lcp.requestIDChan = make(chan uint8)
	lcp.reqiestIDLock = new(sync.RWMutex)
	lcp.logger = pppProto.GetLogger().Named(lcp.protoType.String())
	lcp.sendChan, lcp.recvChan = pppProto.Register(lcp.protoType)
	lcp.layerNotify = h
	for _, mod := range mods {
		mod(lcp)
	}

	go lcp.issueRequestID(ctx)
	go lcp.recv(ctx)
	return lcp
}

// OwnOptionRule is rule that used to handle own LCP options, user could provide implementation of this interface to get custom behavior
type OwnOptionRule interface {
	// HandlerConfRej is the handler function to handle received Conf-Reject
	HandlerConfRej(rcvd Options)
	// HandlerConfNAK is the handler function to handle received Conf-Nak
	HandlerConfNAK(rcvd Options)
	// GetOptions returns current own options
	GetOptions() Options
	// GetOption returns current option with type o
	GetOption(o uint8) Option
}

// DefaultOwnOptionRule is the default OwnOptionRule implementation;
// use NewDefaultOwnOptionRule() to create instance;
// using following options: MRU, AuthProto, MagicNumber with default value;
type DefaultOwnOptionRule struct {
	ownOptions Options
	mux        *sync.RWMutex
}

// NewDefaultOwnOptionRule returns a new DefaultOwnOptionRule
func NewDefaultOwnOptionRule() *DefaultOwnOptionRule {
	return &DefaultOwnOptionRule{
		mux:        new(sync.RWMutex),
		ownOptions: newOwnDefaultOptions(),
	}
}

// GetOptions implements OwnOptionRule
func (own *DefaultOwnOptionRule) GetOptions() Options {
	own.mux.RLock()
	defer own.mux.RUnlock()
	return own.ownOptions
}

// GetOption implements OwnOptionRule
func (own *DefaultOwnOptionRule) GetOption(o uint8) Option {
	own.mux.RLock()
	defer own.mux.RUnlock()
	for _, op := range own.ownOptions {
		if op.Type() == o {
			return op
		}
	}
	return nil
}

// HandlerConfRej implements OwnOptionRule, remove all options listed in conf-rej
func (own *DefaultOwnOptionRule) HandlerConfRej(rcvd Options) {
	own.mux.Lock()
	defer own.mux.Unlock()
	for _, op := range rcvd {
		own.ownOptions.Del(op.Type())
	}
}

// HandlerConfNAK implements OwnOptionRule, accept all options listed in conf-nak
func (own *DefaultOwnOptionRule) HandlerConfNAK(rcvd Options) {
	own.mux.Lock()
	defer own.mux.Unlock()
	own.ownOptions.Replace(rcvd)
}

// PeerOptionRule is rule that use for handle received config-req from peer
type PeerOptionRule interface {
	// HandlerConfReq is the handler function to handle received Conf-Request.
	// if a recived option needs to be naked or rejected, include it in returned nak/reject LCPOptions
	HandlerConfReq(rcvd Options) (nak, reject Options)
	// GetOptions return current peer's options
	GetOptions() Options
}

// DefaultPeerOptionRule is the default PeerOptionRule implementation.
type DefaultPeerOptionRule struct {
	// AuthOp is the required Auth Protocol Option (PAP or CHAP)
	AuthOp         *LCPOpAuthProto
	currentOptions Options
}

// NewDefaultPeerOptionRule create a new DefaultPeerOptionRule instance with specified authp (
func NewDefaultPeerOptionRule(authp PPPProtocolNumber) (*DefaultPeerOptionRule, error) {
	var op *LCPOpAuthProto
	switch authp {
	case ProtoCHAP:
		op = NewCHAPAuthOp()
	case ProtoPAP:
		op = NewPAPAuthOp()
	default:
		return nil, fmt.Errorf("unsupported auth protocol: %v", authp)
	}
	r := new(DefaultPeerOptionRule)
	r.AuthOp = op
	return r, nil
}

// GetOptions implements PeerOptionRule.
func (rule *DefaultPeerOptionRule) GetOptions() Options {
	return rule.currentOptions
}

// HandlerConfReq implements PeerOptionRule, if config-request include an auth-proto option that is different from required one, it will be NAKed;
// Option in conf-req other than auth-proto, magic number and MRU will be rejected.
func (rule *DefaultPeerOptionRule) HandlerConfReq(rcvd Options) (nak, reject Options) {
	rule.currentOptions = rcvd
	for _, o := range rcvd {
		switch LCPOptionType(o.Type()) {
		case OpTypeAuthenticationProtocol:
			if !o.Equal(rule.AuthOp) {
				nak = append(nak, rule.AuthOp)
			}
		case OpTypeMagicNumber, OpTypeMaximumReceiveUnit:
		default:
			reject = append(reject, o)
		}
	}
	return
}

func getCallerName() (fname, callername string, linenum int) {
	fpcs := make([]uintptr, 1)
	// Skip 2 levels to get the caller
	n := runtime.Callers(3, fpcs)
	if n == 0 {
		return
	}

	caller := runtime.FuncForPC(fpcs[0] - 1)
	if caller == nil {
		return
	}
	fname, linenum = caller.FileLine(fpcs[0] - 1)
	callername = caller.Name()
	return

}

func (lcp *LCP) setState(s State) {
	old := State(atomic.LoadUint32(lcp.state))
	atomic.StoreUint32(lcp.state, uint32(s))
	_, callername, linenum := getCallerName()

	lcp.logger.Sugar().Debugf("%v:%v state transit %v -> %v", callername, linenum, old, s)
}

func (lcp *LCP) getState() State {
	return State(atomic.LoadUint32(lcp.state))
}

func (lcp *LCP) issueRequestID(ctx context.Context) {
	for {
		select {
		case lcp.requestIDChan <- lcp.requestID:
			lcp.reqiestIDLock.Lock()
			lcp.requestID++
			lcp.reqiestIDLock.Unlock()
		case <-ctx.Done():
			lcp.logger.Info("issueRequestID routine stopped")
			return
		}

	}
}

// func (lcp *LCP) layerNotify(notify int) {

// }

func (lcp *LCP) send(p []byte) (err error) {
	ppkt := NewPPPPkt(p, lcp.protoType)
	lcp.sendChan <- ppkt.Serialize()
	// lcp.logger.Sugar().Debugf("send a pkt, current state is %v", lcp.getState())
	return
}

func (lcp *LCP) sendConfReq(ctx context.Context) error {
	lcppkt := NewPkt(lcp.protoType)
	lcppkt.Code = CodeConfigureRequest
	lcppkt.ID = <-lcp.requestIDChan
	lcppkt.Options = lcp.OwnRule.GetOptions()
	lcpbytes, err := lcppkt.Serialize()
	if err != nil {
		return err
	}
	lcp.logger.Info("sending conf-req")
	lcp.logger.Debug("\n" + lcppkt.String())
	defer lcp.resetTimer(ctx)
	return lcp.send(lcpbytes)
}

func (lcp *LCP) sendTermReq(ctx context.Context) error {
	lcppkt := NewPkt(lcp.protoType)
	lcppkt.Code = CodeTerminateRequest
	lcppkt.ID = <-lcp.requestIDChan
	lcpbytes, err := lcppkt.Serialize()
	if err != nil {
		return err
	}
	lcp.logger.Info("sending term-req")
	lcp.logger.Debug("\n" + lcppkt.String())
	defer lcp.resetTimer(ctx)
	return lcp.send(lcpbytes)
}

func (lcp *LCP) sendTermACK(req *Pkt) error {
	lcppkt := NewPkt(lcp.protoType)
	lcppkt.Code = CodeTerminateAck
	lcppkt.ID = req.ID
	lcpbytes, err := lcppkt.Serialize()
	if err != nil {
		return err
	}
	lcp.logger.Info("sending term-ack")
	lcp.logger.Debug("\n" + lcppkt.String())
	return lcp.send(lcpbytes)
}
func (lcp *LCP) sendNAKRejct(req *Pkt, nak, reject Options) (err error) {
	var lcpbytes []byte
	if len(nak) > 0 {
		lcppkt := NewPkt(lcp.protoType)
		lcppkt.Code = CodeConfigureNak
		lcppkt.ID = req.ID
		lcppkt.Options = nak
		lcpbytes, err = lcppkt.Serialize()
		if err != nil {
			return
		}
		err = lcp.send(lcpbytes)
		if err != nil {
			return
		}
		lcp.logger.Info("sending conf-nak")
		lcp.logger.Debug("\n" + lcppkt.String())
	}
	if len(reject) > 0 {
		lcppkt := NewPkt(lcp.protoType)
		lcppkt.Code = CodeConfigureReject
		lcppkt.ID = req.ID
		lcppkt.Options = reject
		lcpbytes, err = lcppkt.Serialize()
		if err != nil {
			return
		}
		err = lcp.send(lcpbytes)
		if err != nil {
			return
		}
		lcp.logger.Info("sending conf-reject")
		lcp.logger.Debug("\n" + lcppkt.String())
	}
	return
}

func (lcp *LCP) sendConfACK(req *Pkt) error {
	lcppkt := NewPkt(lcp.protoType)
	lcppkt.Code = CodeConfigureAck
	lcppkt.ID = req.ID
	// for _, o := range req.Options {
	// 	if o.Type() != OpTypeMagicNumber {
	// 		lcppkt.Options = append(lcppkt.Options, o)
	// 	}
	// }
	lcppkt.Options = req.Options
	lcpbytes, err := lcppkt.Serialize()
	if err != nil {
		return err
	}
	lcp.logger.Info("sending conf-ack")
	lcp.logger.Debug("\n" + lcppkt.String())
	return lcp.send(lcpbytes)
}

func (lcp *LCP) resetKeepAliveTimer(ctx context.Context) {
	lcp.logger.Debug("reset keepalive timer")
	if lcp.protoType != ProtoLCP {
		return
	}
	if lcp.keepAliveTimer == nil {
		lcp.keepAliveTimer = time.NewTimer(lcp.keepAliveInterval)
	} else {
		lcp.keepAliveTimer.Stop()
		lcp.keepAliveTimer.Reset(lcp.keepAliveInterval)
		lcp.cancellkeepAliveTimer()
	}
	var childctx context.Context
	childctx, lcp.cancellkeepAliveTimer = context.WithCancel(ctx)
	go func(c context.Context) {
		select {
		case <-lcp.keepAliveTimer.C:
			lcp.keepAliveTimeout(ctx)
		case <-c.Done():
		}
	}(childctx)
}

func (lcp *LCP) resetTimer(ctx context.Context) {
	lcp.logger.Debug("reset timer")
	if lcp.restartTimer == nil {
		lcp.restartTimer = time.NewTimer(lcp.restartTimerDuration)
	} else {
		lcp.restartTimer.Stop()
		lcp.restartTimer.Reset(lcp.restartTimerDuration)
		lcp.cancellRestartTimer()
	}
	var childctx context.Context
	childctx, lcp.cancellRestartTimer = context.WithCancel(ctx)
	go func(c context.Context) {
		select {
		case <-lcp.restartTimer.C:
			lcp.timeout(ctx)
		case <-c.Done():
		}
	}(childctx)
}

// Keepalive Timeout event, called by lcp.resetKeepAliveTimer()
func (lcp *LCP) keepAliveTimeout(ctx context.Context) {
	switch State(atomic.LoadUint32(lcp.state)) {
	case StateOpened:
		err := lcp.sendEchoRequest(ctx)
		if err != nil {
			lcp.logger.Error(err.Error())
			return
		}
		lcp.setState(StateEchoReqSent)
	}

}

// Timeout event, called by lcp.resetTimer()
func (lcp *LCP) timeout(ctx context.Context) {
	defer atomic.AddUint32(lcp.restartCount, ^uint32(0))
	if atomic.LoadUint32(lcp.restartCount) == 0 {
		if atomic.LoadUint32(lcp.state) == uint32(StateEchoReqSent) {
			lcp.logger.Error("keepalive timeout")
		}
		lcp.toMinus(ctx)
	}
	lcp.toPlus(ctx)
}

// toPlus is TO+ event
func (lcp *LCP) toPlus(ctx context.Context) {
	lcp.logger.Debug("timer expired, TO+ event")
	var err error
	switch lcp.getState() {
	case StateClosing, StateStopping:
		//send term req
		err = lcp.sendTermReq(ctx)
	case StateReqSent, StateAckSent:
		//send conf req
		err = lcp.sendConfReq(ctx)
	case StateAckRcvd:
		//send conf req, this is actually send current version of config options
		err = lcp.sendConfReq(ctx)
		if err == nil {
			lcp.setState(StateReqSent)
		}
	case StateEchoReqSent:
		//send echo request
		err = lcp.sendEchoRequest(ctx)
	}
	if err != nil {
		lcp.logger.Sugar().Errorf("failed to process TO+ event,%v", err)
	}
}

// toMinus is TO- event
func (lcp *LCP) toMinus(ctx context.Context) {
	lcp.logger.Debug("timer expired, TO- event")
	switch lcp.getState() {
	case StateClosing:
		lcp.layerNotify(ctx, LCPLayerNotifyFinished)
		lcp.setState(StateClosed)
	case StateStopping:
		lcp.layerNotify(ctx, LCPLayerNotifyFinished)
		lcp.setState(StateStopped)
	case StateReqSent, StateAckSent, StateAckRcvd, StateEchoReqSent:
		lcp.layerNotify(ctx, LCPLayerNotifyFinished)
		lcp.setState(StateStopped)
	}
}

const (
	readTimeout = 3 * time.Second
)

func (lcp *LCP) processRecvByte(ctx context.Context, pktbytes []byte) {

	if len(pktbytes) < 4 {
		lcp.logger.Warn("recvd LCP pkt too small")
		return
	}
	pkt := NewPkt(lcp.protoType)
	err := pkt.Parse(pktbytes)
	if err != nil {
		lcp.logger.Sugar().Warnf("invalid LCP pkt,%v", err)
		return
	}
	lcp.logger.Sugar().Infof("got a %v lcp pkt ", pkt.Code.String())
	lcp.logger.Debug("\n" + pkt.String())
	switch pkt.Code {
	case CodeConfigureAck:
		err = lcp.rca(ctx, pkt)
		if err != nil {
			lcp.logger.Sugar().Errorf("failed to process RCA event,%v", err)
		}
	case CodeConfigureRequest:
		nak, reject := lcp.PeerRule.HandlerConfReq(pkt.Options)
		if len(nak) == 0 && len(reject) == 0 {
			err = lcp.rcrPlus(ctx, pkt)
			if err != nil {
				lcp.logger.Sugar().Errorf("failed to process RCR+ event,%v", err)
			}
			return
		}
		err = lcp.rcrMinus(ctx, pkt, nak, reject)
		if err != nil {
			lcp.logger.Sugar().Errorf("failed to process RCR event,%v", err)
		}
	case CodeEchoReply, CodeEchoRequest, CodeDiscardRequest:
		err = lcp.rxr(ctx, pkt)
		if err != nil {
			lcp.logger.Sugar().Errorf("failed to process RXR event,%v", err)
		}
	case CodeTerminateRequest:
		err = lcp.rtr(ctx, pkt)
		if err != nil {
			lcp.logger.Sugar().Errorf("failed to process RTR event,%v", err)
		}
	case CodeTerminateAck:
		err = lcp.rta(ctx)
		if err != nil {
			lcp.logger.Sugar().Errorf("failed to process RTA event,%v", err)
		}
	case CodeConfigureNak, CodeConfigureReject:
		err = lcp.rcn(ctx, pkt)
		if err != nil {
			lcp.logger.Sugar().Errorf("failed to process RCN event,%v", err)
		}
	case CodeCodeReject, CodeProtocolReject:
		err = lcp.rxjMinus(ctx, pkt)
		if err != nil {
			lcp.logger.Sugar().Errorf("failed to process RXJ event,%v", err)
		}
	default:
		err = lcp.ruc(pkt)
		if err != nil {
			lcp.logger.Sugar().Errorf("failed to handle RUC event,%v", err)
		}
	}
}

func (lcp *LCP) recv(ctx context.Context) {
	// var err error
	// var n int
	for {
		select {
		case pktbytes := <-lcp.recvChan:
			lcp.processRecvByte(ctx, pktbytes)
		case <-ctx.Done():
			lcp.logger.Info("recv routine stopped")
			return
		}

	}
}

func (lcp *LCP) rcrPlus(ctx context.Context, req *Pkt) (err error) {
	switch lcp.getState() {
	case StateClosed:
		//send term-ack
		err = lcp.sendTermACK(req)
	case StateStopped:
		// send conf-req
		err = lcp.sendConfReq(ctx)
		if err != nil {
			return
		}
		// send conf-ack
		err = lcp.sendConfACK(req)
		if err != nil {
			return
		}
		atomic.StoreUint32(lcp.restartCount, lcp.maxRestart)
		lcp.setState(StateAckSent)
	case StateReqSent:
		// send conf-ack
		err = lcp.sendConfACK(req)
		if err != nil {
			return
		}
		lcp.setState(StateAckSent)
	case StateAckRcvd:
		//send conf-ack
		err = lcp.sendConfACK(req)
		if err != nil {
			return
		}
		lcp.layerNotify(ctx, LCPLayerNotifyUp)
		lcp.setState(StateOpened)
		lcp.resetKeepAliveTimer(ctx)
	case StateAckSent:
		// send conf-ack
		err = lcp.sendConfACK(req)
	case StateOpened, StateEchoReqSent:
		lcp.layerNotify(ctx, LCPLayerNotifyDown)
		// send conf-req
		err = lcp.sendConfReq(ctx)
		if err != nil {
			return
		}
		// send conf-ack
		err = lcp.sendConfACK(req)
		if err != nil {
			return
		}
		lcp.setState(StateAckSent)

	}
	return
}

func (lcp *LCP) rcrMinus(ctx context.Context, req *Pkt, nak, reject Options) (err error) {
	switch lcp.getState() {
	case StateClosed:
		//send term-ack
		err = lcp.sendTermACK(req)
	case StateStopped:
		// send conf-req
		err = lcp.sendConfReq(ctx)
		if err != nil {
			return
		}
		// send conf-nak
		err = lcp.sendNAKRejct(req, nak, reject)
		if err != nil {
			return
		}
		atomic.StoreUint32(lcp.restartCount, lcp.maxRestart)
		lcp.setState(StateReqSent)
	case StateReqSent, StateAckRcvd:
		// send conf-nak
		err = lcp.sendNAKRejct(req, nak, reject)
	case StateAckSent:
		// send conf-nak
		err = lcp.sendNAKRejct(req, nak, reject)
		if err != nil {
			return
		}
		lcp.setState(StateReqSent)
	case StateOpened, StateEchoReqSent:
		// send conf-req
		err = lcp.sendConfReq(ctx)
		if err != nil {
			return
		}
		// send conf-nak
		err = lcp.sendNAKRejct(req, nak, reject)
		if err != nil {
			return
		}
		lcp.layerNotify(ctx, LCPLayerNotifyDown)
		lcp.setState(StateReqSent)
	}
	return
}

// RCA event
func (lcp *LCP) rca(ctx context.Context, req *Pkt) (err error) {

	switch lcp.getState() {
	case StateStopped, StateClosed:
		//send term-ack
		err = lcp.sendTermACK(req)
	case StateReqSent:
		atomic.StoreUint32(lcp.restartCount, lcp.maxRestart)
		lcp.setState(StateAckRcvd)
	case StateAckRcvd:
		//send conf req
		err = lcp.sendConfReq(ctx)
		if err != nil {
			return
		}
		lcp.setState(StateReqSent)
	case StateAckSent:
		atomic.StoreUint32(lcp.restartCount, lcp.maxRestart)
		lcp.layerNotify(ctx, LCPLayerNotifyUp)
		lcp.setState(StateOpened)
		lcp.resetKeepAliveTimer(ctx)
	case StateOpened, StateEchoReqSent:
		lcp.layerNotify(ctx, LCPLayerNotifyDown)
		//send conf req
		err = lcp.sendConfReq(ctx)
		if err != nil {
			return
		}
		lcp.setState(StateReqSent)
	}
	return
}

//RCN event
func (lcp *LCP) rcn(ctx context.Context, req *Pkt) error {
	switch req.Code {
	case CodeConfigureNak:
		lcp.OwnRule.HandlerConfNAK(req.Options)
	case CodeConfigureReject:
		lcp.OwnRule.HandlerConfRej(req.Options)
	}
	switch lcp.getState() {
	case StateStopped, StateClosed:
		// send term-ack
		return lcp.sendTermACK(req)
	case StateReqSent:
		//send cfg-req
		err := lcp.sendConfReq(ctx)
		if err != nil {
			return err
		}
		atomic.StoreUint32(lcp.restartCount, lcp.maxRestart)
	case StateAckRcvd:
		//send cfg-req
		err := lcp.sendConfReq(ctx)
		if err != nil {
			return err
		}
		lcp.setState(StateReqSent)
	case StateAckSent:
		//send cfg-req
		err := lcp.sendConfReq(ctx)
		if err != nil {
			return err
		}
		atomic.StoreUint32(lcp.restartCount, lcp.maxRestart)
	case StateOpened, StateEchoReqSent:
		//send cfg-req
		err := lcp.sendConfReq(ctx)
		if err != nil {
			return err
		}
		lcp.layerNotify(ctx, LCPLayerNotifyDown)
		lcp.setState(StateReqSent)
	}
	return nil
}

//PTR event
func (lcp *LCP) rtr(ctx context.Context, req *Pkt) (err error) {
	switch lcp.getState() {
	case StateClosed, StateStopped, StateClosing, StateStopping:
		//send term-ack
		err = lcp.sendTermACK(req)
	case StateReqSent, StateAckRcvd, StateAckSent:

		err = lcp.sendTermACK(req)
		if err != nil {
			return
		}
		lcp.setState(StateReqSent)
	case StateOpened, StateEchoReqSent:
		// send term-ack
		err = lcp.sendTermACK(req)
		if err != nil {
			return
		}
		lcp.layerNotify(ctx, LCPLayerNotifyDown)
		atomic.StoreUint32(lcp.restartCount, 0)
		lcp.resetTimer(ctx)
		lcp.setState(StateStopping)
	}
	return
}

// RTA event
func (lcp *LCP) rta(ctx context.Context) error {
	lcp.logger.Debug("RTA (receive term-ack) event")
	switch lcp.getState() {
	case StateClosing:
		lcp.layerNotify(ctx, LCPLayerNotifyFinished)
		lcp.setState(StateClosed)
	case StateStopping:
		lcp.layerNotify(ctx, LCPLayerNotifyFinished)
		lcp.setState(StateStopped)
	case StateAckRcvd:
		lcp.setState(StateReqSent)
	case StateOpened, StateEchoReqSent:
		//send conf req
		err := lcp.sendConfReq(ctx)
		if err != nil {
			return err
		}
		lcp.layerNotify(ctx, LCPLayerNotifyDown)
		lcp.setState(StateReqSent)

	}
	return nil
}

// RUC event
func (lcp *LCP) ruc(req *Pkt) error {
	switch lcp.getState() {
	case StateInitial, StateStarting:
		return nil
	}
	// send code-rej
	return lcp.sendCodeReject(req)
}

func (lcp *LCP) sendCodeReject(req *Pkt) error {
	pkt := NewPkt(lcp.protoType)
	pkt.Code = CodeCodeReject
	pkt.ID = <-lcp.requestIDChan
	pkt.Payload, _ = req.Serialize()
	pktbytes, err := pkt.Serialize()
	if err != nil {
		return err
	}
	lcp.logger.Sugar().Debugf("sending code-reject\n%v", pkt)
	return lcp.send(pktbytes)
}

// rxjPlus is RXJ+ event
func (lcp *LCP) rxjPlus() {
	switch lcp.getState() {
	case StateAckRcvd:
		lcp.setState(StateReqSent)
	}
}

// rxjMnius is RXJ- event
func (lcp *LCP) rxjMinus(ctx context.Context, req *Pkt) error {
	lcp.logger.Sugar().Errorf("Got a %v pkt", req.Code)
	switch lcp.getState() {
	case StateStopped, StateClosed:
		lcp.layerNotify(ctx, LCPLayerNotifyFinished)
	case StateClosing:
		lcp.layerNotify(ctx, LCPLayerNotifyFinished)
		lcp.setState(StateClosed)
	case StateStopping:
		lcp.layerNotify(ctx, LCPLayerNotifyFinished)
		lcp.setState(StateStopped)
	case StateReqSent, StateAckRcvd, StateAckSent:
		lcp.layerNotify(ctx, LCPLayerNotifyFinished)
		lcp.setState(StateStopped)
	case StateOpened, StateEchoReqSent:
		// send term-req
		err := lcp.sendTermReq(ctx)
		if err != nil {
			return err
		}
		lcp.layerNotify(ctx, LCPLayerNotifyDown)
		atomic.StoreUint32(lcp.restartCount, lcp.maxRestart)
	}
	return nil
}

func (lcp *LCP) sendEchoRequest(ctx context.Context) error {
	lcppkt := NewPkt(lcp.protoType)
	lcppkt.Code = CodeEchoRequest
	lcppkt.ID = <-lcp.requestIDChan
	if mn := lcp.OwnRule.GetOption(uint8(OpTypeMagicNumber)); mn != nil {
		lcppkt.MagicNum = uint32(*(mn.(*LCPOpMagicNum)))
	}
	lcpbytes, err := lcppkt.Serialize()
	if err != nil {
		return err
	}
	lcp.logger.Info("sending echo-request")
	lcp.logger.Debug("\n" + lcppkt.String())
	defer lcp.resetTimer(ctx)
	return lcp.send(lcpbytes)

}

func (lcp *LCP) sendEchoReply(req *Pkt) error {
	lcppkt := NewPkt(lcp.protoType)
	lcppkt.Code = CodeEchoReply
	lcppkt.ID = req.ID
	//lcppkt.Payload = make([]byte, 4)
	lcppkt.Options = Options{}
	if mn := lcp.OwnRule.GetOption(uint8(OpTypeMagicNumber)); mn != nil {
		lcppkt.MagicNum = uint32(*(mn.(*LCPOpMagicNum)))
	}
	lcpbytes, err := lcppkt.Serialize()
	if err != nil {
		return err
	}
	lcp.logger.Info("sending echo-reply")
	lcp.logger.Debug("\n" + lcppkt.String())
	return lcp.send(lcpbytes)
}

// RXR event
func (lcp *LCP) rxr(ctx context.Context, req *Pkt) error {
	switch req.Code {
	case CodeEchoRequest:
		switch lcp.getState() {
		case StateOpened, StateEchoReqSent:
			return lcp.sendEchoReply(req)
		}
	case CodeEchoReply:
		switch lcp.getState() {
		case StateEchoReqSent:
			atomic.StoreUint32(lcp.restartCount, lcp.maxRestart)
			lcp.setState(StateOpened)
			lcp.resetKeepAliveTimer(ctx)
		}
	}
	return nil
}

// Up is lower layer up event, as defined in RFC1661
func (lcp *LCP) Up(ctx context.Context) (err error) {
	switch lcp.getState() {
	case StateInitial:
		lcp.setState(StateClosed)
	case StateStarting:
		err = lcp.sendConfReq(ctx)
		if err != nil {
			return
		}
		atomic.StoreUint32(lcp.restartCount, lcp.maxRestart)
		lcp.resetTimer(ctx)
		lcp.setState(StateReqSent)
	}
	return
}

// Down is lower layer down event, as defined in RFC1661
func (lcp *LCP) Down(ctx context.Context) {
	switch lcp.getState() {
	case StateStopped:
		lcp.layerNotify(ctx, LCPLayerNotifyStarted)
		lcp.setState(StateStarting)
	case StateReqSent, StateAckRcvd, StateAckSent:
		lcp.setState(StateStarting)
	case StateOpened, StateEchoReqSent:
		lcp.layerNotify(ctx, LCPLayerNotifyDown)
		lcp.setState(StateStarting)
	}
}

// Open is admin Open event, as defined in RFC1661
func (lcp *LCP) Open(ctx context.Context) error {
	switch lcp.getState() {
	case StateInitial:
		lcp.layerNotify(ctx, LCPLayerNotifyStarted)
		lcp.setState(StateStarting)
	case StateClosed:
		err := lcp.sendConfReq(ctx)
		if err != nil {
			return err
		}
		atomic.StoreUint32(lcp.restartCount, lcp.maxRestart)
		lcp.setState(StateReqSent)
	case StateClosing:
		lcp.setState(StateStopping)
	}
	return nil
}

// Close is admin Close event, as defined in RFC1661
func (lcp *LCP) Close(ctx context.Context) {
	switch lcp.getState() {
	case StateStarting:
		lcp.layerNotify(ctx, LCPLayerNotifyFinished)
		lcp.setState(StateInitial)
	case StateStopped:
		lcp.setState(StateClosed)
	case StateStopping:
		lcp.setState(StateClosing)
	case StateReqSent, StateAckRcvd, StateAckSent:
		// send term req
		err := lcp.sendTermReq(ctx)
		if err != nil {
			lcp.logger.Sugar().Errorf("failed to process TO+ event,err", err)
			return
		}
		atomic.StoreUint32(lcp.restartCount, lcp.maxRestart)

		lcp.setState(StateClosing)
	case StateOpened, StateEchoReqSent:
		//send term req
		err := lcp.sendTermReq(ctx)
		if err != nil {
			lcp.logger.Sugar().Errorf("failed to process TO+ event,err", err)
			return
		}
		lcp.layerNotify(ctx, LCPLayerNotifyDown)
		atomic.StoreUint32(lcp.restartCount, lcp.maxRestart)

		lcp.setState(StateClosing)
	}
}

// Modifier provides custom configuration for NewLCP()
type Modifier func(lcp *LCP)

// WithPeerOptionRule specify r as the PeerOptionRule
func WithPeerOptionRule(r PeerOptionRule) Modifier {
	return func(lcp *LCP) {
		lcp.PeerRule = r
	}
}

// WithOwnOptionRule specify r as the OwnOptionRule
func WithOwnOptionRule(r OwnOptionRule) Modifier {
	return func(lcp *LCP) {
		lcp.OwnRule = r
	}
}

// Options is a slice of LCPOption
type Options []Option

// Get return all options with type t
func (options Options) Get(t uint8) (r Options) {
	for _, o := range options {
		if o.Type() == t {
			r = append(r, o)
		}
	}
	return
}

// GetFirst return 1st option with type t
func (options Options) GetFirst(t uint8) Option {
	for _, o := range options {
		if o.Type() == t {
			return o
		}
	}
	return nil
}

// Del removes all options with type t
func (options *Options) Del(t uint8) {
	for i, o := range *options {
		if o.Type() == t {
			*options = append((*options)[:i], (*options)[i+1:]...)
		}
	}
}

// Append append newoptions
func (options *Options) Append(newoptions Options) {
	*options = append(*options, newoptions...)
}

// Replace removes all options with all options in newoptions, and append newoptions
func (options *Options) Replace(newoptions Options) {
	for _, o := range newoptions {
		options.Del(o.Type())
	}
	options.Append(newoptions)
}
