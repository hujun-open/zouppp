package lcp

import (
	"context"
	"encoding/binary"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/hujun-open/etherconn"
)

type PPPAddr struct {
	proto PPPProtocolNumber
}

// Network implenets net.Addr interface
func (paddr PPPAddr) Network() string {
	return "ppp"
}

// String implenets net.Addr interface, return "ppp:proto"
func (paddr PPPAddr) String() string {
	return fmt.Sprintf("ppp:%v", paddr.proto)
}

type PPPPayloadRcvHandler func([]byte) ([]byte, int, error)
type PPPPayloadSendHandler func([]byte, net.Addr) ([]byte, int, error)

// PPPConn implements etherconn.SharedEconn interface, could be used to fwd network packets over PPP
type PPPConn struct {
	ppp                                 *PPP
	send, recv                          chan []byte
	localAddr                           PPPAddr
	readDeadline, writeDeadline         time.Time
	readDeadlineLock, writeDeadlineLock *sync.RWMutex
	protoPrefix                         [2]byte
	recvList                            *etherconn.ChanMap
	perClntRecvChanDepth                uint
}

const (
	DefaultPerClntRecvChanDepth = 1024
)

// NewPPPConn creates a new PPPConn over ppp, for network protocol proto.
// it implements etherconn.SharedEconn interface
func NewPPPConn(ctx context.Context, ppp *PPP, proto PPPProtocolNumber) *PPPConn {
	r := new(PPPConn)
	r.ppp = ppp
	r.localAddr = PPPAddr{proto: proto}
	r.send, r.recv = ppp.Register(proto)
	r.readDeadlineLock = new(sync.RWMutex)
	r.writeDeadlineLock = new(sync.RWMutex)
	r.recvList = etherconn.NewChanMap()
	r.perClntRecvChanDepth = DefaultPerClntRecvChanDepth
	binary.BigEndian.PutUint16(r.protoPrefix[:], uint16(proto))
	go r.recvHandling(ctx)
	return r
}

// LocalAddr implements net.PacketConn interface.
func (pconn *PPPConn) LocalAddr() net.Addr {
	return pconn.localAddr
}

//pkt is an IPv4 or IPv6 pkt
func parseIPPkt(pkt []byte) (*etherconn.RelayReceival, error) {
	var l4index int
	rcv := new(etherconn.RelayReceival)
	rcv.EtherPayloadBytes = pkt
	if len(pkt) < 20 {
		return nil, fmt.Errorf("pkt smaller than 20B")
	}
	switch pkt[0] & 0b11110000 {
	case 96:
		//ipv6
		rcv.Protocol = rcv.EtherPayloadBytes[6]
		rcv.RemoteIP = rcv.EtherPayloadBytes[8:24]
		rcv.LocalIP = rcv.EtherPayloadBytes[24:40]
		l4index = 40 //NOTE: this means no supporting of any ipv6 options
	case 0b01000000:
		//ipv4
		rcv.RemoteIP = rcv.EtherPayloadBytes[12:16]
		rcv.LocalIP = rcv.EtherPayloadBytes[16:20]
		rcv.Protocol = rcv.EtherPayloadBytes[9]
		l4index = 20 //NOTE: this means no supporting of any ipv4 options
	default:
		return nil, fmt.Errorf("not an IP packet")
	}
	switch rcv.Protocol {
	case 17: //udp
		rcv.RemotePort = binary.BigEndian.Uint16(rcv.EtherPayloadBytes[l4index : l4index+2])
		rcv.LocalPort = binary.BigEndian.Uint16(rcv.EtherPayloadBytes[l4index+2 : l4index+4])
		rcv.TransportPayloadBytes = rcv.EtherPayloadBytes[l4index+8:]
	case 58: //ICMPv6
		rcv.RemotePort = uint16(rcv.EtherPayloadBytes[l4index])
		rcv.LocalPort = rcv.RemotePort
		rcv.TransportPayloadBytes = rcv.EtherPayloadBytes[l4index+4:]
	}
	return rcv, nil

}
func (pconn *PPPConn) recvHandling(ctx context.Context) {
	var buf []byte
	var receival *etherconn.RelayReceival
	var err error
	// runtime.LockOSThread()
	for {
		select {
		case <-ctx.Done():
			return
		case buf = <-pconn.recv:
			receival, err = parseIPPkt(buf)
			if err != nil {
				continue
			}

			if ch := pconn.recvList.Get(receival.GetL4Key()); ch != nil {
				//found registed channel
			L99:
				for {
					select {
					case ch <- receival:
						break L99
					default:
						//channel is full, remove oldest pkt
						<-ch
					}
				}
			}
		}
	}
}

func (pconn *PPPConn) Register(k etherconn.L4RecvKey) (torecvch chan *etherconn.RelayReceival) {
	return pconn.RegisterList([]etherconn.L4RecvKey{k})
}

func (pconn *PPPConn) RegisterList(keys []etherconn.L4RecvKey) (torecvch chan *etherconn.RelayReceival) {
	ch := make(chan *etherconn.RelayReceival, pconn.perClntRecvChanDepth)
	list := make([]interface{}, len(keys))
	for i := range keys {
		list[i] = keys[i]
	}
	pconn.recvList.SetList(list, ch)
	return ch
}

func (pconn *PPPConn) WriteIPPktTo(p []byte, dstmac net.HardwareAddr) (int, error) {
	pconn.writeDeadlineLock.RLock()
	deadline := pconn.writeDeadline
	pconn.writeDeadlineLock.RUnlock()
	delta := time.Until(deadline)
	var buf []byte
	switch p[0] & 0b11110000 {
	case 96:
		//ipv6
		binary.BigEndian.PutUint16(pconn.protoPrefix[:], uint16(ProtoIPv6))
	case 0b01000000:
		//ipv4
		binary.BigEndian.PutUint16(pconn.protoPrefix[:], uint16(ProtoIPv4))
	default:
		return 0, fmt.Errorf("not an IP packet")
	}
	buf = make([]byte, len(p)+2)
	copy(buf[:2], pconn.protoPrefix[:])
	copy(buf[2:], p)
	if delta > 0 {
		select {
		case <-time.After(delta):
			return 0, etherconn.ErrTimeOut
		case pconn.send <- buf:
			return len(p), nil
		}
	} else {
		pconn.send <- buf
		return len(p), nil
	}
}

// Close implements net.PacketConn interface.
func (pconn *PPPConn) Close() error {
	pconn.ppp.UnRegister(pconn.localAddr.proto)
	return nil
}

// SetDeadline implements net.PacketConn interface.
func (pconn *PPPConn) SetDeadline(t time.Time) error {
	pconn.writeDeadlineLock.Lock()
	pconn.writeDeadline = t
	pconn.writeDeadlineLock.Unlock()
	pconn.readDeadlineLock.Lock()
	pconn.readDeadline = t
	pconn.readDeadlineLock.Unlock()
	return nil
}

// SetReadDeadline implements net.PacketConn interface.
func (pconn *PPPConn) SetReadDeadline(t time.Time) error {
	pconn.readDeadlineLock.Lock()
	pconn.readDeadline = t
	pconn.readDeadlineLock.Unlock()
	return nil
}

// SetWriteDeadline implements net.PacketConn interface.
func (pconn *PPPConn) SetWriteDeadline(t time.Time) error {
	pconn.writeDeadlineLock.Lock()
	pconn.writeDeadline = t
	pconn.writeDeadlineLock.Unlock()
	return nil
}
