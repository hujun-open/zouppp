// keepalive_test — LCP keepalive 循环武装 + 3-miss 静默判死 + EchoReply 清零 单元测试
// 补丁来源: brix 项目 (github.com/bpfio/brix) third_party/zouppp 实践,
// 生产实证: PPPoE 会话稳定运行; 本测试补充行为级验证 (原实现无测试覆盖)
package lcp

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"go.uber.org/zap"
)

type mockNotify struct {
	mu     sync.Mutex
	events []LayerNotifyEvent
}

func (m *mockNotify) onEvent(ctx context.Context, evt LayerNotifyEvent) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.events = append(m.events, evt)
}

func (m *mockNotify) count(evt LayerNotifyEvent) int {
	m.mu.Lock()
	defer m.mu.Unlock()
	n := 0
	for _, e := range m.events {
		if e == evt {
			n++
		}
	}
	return n
}

// newTestLCP 构造最小可用 LCP: 不依赖 PPP 网络栈, 直填字段;
// sendChan 缓冲吸收 echo-request; requestID 由 goroutine 持续供给
func newTestLCP(t *testing.T, ctx context.Context, interval time.Duration, notify LayerNotifyHandler) *LCP {
	t.Helper()
	lcp := &LCP{
		protoType:          ProtoLCP,
		state:              new(uint32),
		echoMissed:         new(uint32),
		restartCount:       new(uint32),
		maxRestart:         DefaultRestartCounter,
		requestIDChan:      make(chan uint8, 64),
		reqiestIDLock:      new(sync.RWMutex),
		sendChan:           make(chan []byte, 64),
		recvChan:           make(chan []byte, 64),
		logger:             zap.NewNop(),
		OwnRule:            NewDefaultOwnOptionRule(),
		keepAliveInterval:  interval,
		layerNotify:        notify,
		restartTimerDuration: 10 * time.Second,
	}
	atomic.StoreUint32(lcp.state, uint32(StateOpened))
	// issueRequestID 平替: 持续供给递增 ID
	go func() {
		var id uint8
		for {
			select {
			case lcp.requestIDChan <- id:
				id++
			case <-ctx.Done():
				return
			}
		}
	}()
	// 吸收发出的 echo-request (含 PPP 封装), 供测试读取
	go func() {
		for {
			select {
			case <-lcp.sendChan:
			case <-ctx.Done():
				return
			}
		}
	}()
	lcp.resetKeepAliveTimer(ctx)
	return lcp
}

func waitFor(t *testing.T, timeout time.Duration, cond func() bool) bool {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if cond() {
			return true
		}
		time.Sleep(5 * time.Millisecond)
	}
	return false
}

// 场景 A: 对端静默 — 连续 3 个 keepalive 周期无 EchoReply, 必须触发 Down (LCPLayerNotifyDown)
func TestKeepAliveSilentPeerTriggersDown(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	m := &mockNotify{}
	interval := 20 * time.Millisecond
	lcp := newTestLCP(t, ctx, interval, m.onEvent)

	if !waitFor(t, 5*interval+2*time.Second, func() bool {
		return m.count(LCPLayerNotifyDown) >= 1
	}) {
		t.Fatalf("silent peer did not trigger Down within deadline, events=%v, echoMissed=%d",
			m.events, atomic.LoadUint32(lcp.echoMissed))
	}
	if got := atomic.LoadUint32(lcp.echoMissed); got < 3 {
		t.Fatalf("echoMissed=%d, want >=3 at teardown", got)
	}
	// 判死后 keepalive 循环应退出 (状态离开 Opened/EchoReqSent)
	if st := State(atomic.LoadUint32(lcp.state)); st == StateOpened || st == StateEchoReqSent {
		t.Fatalf("state=%v after teardown, want left EchoReqSent", st)
	}
}

// 场景 B: 对端正常应答 — EchoReply 到达清零静默计数, 永不触发 Down
func TestKeepAliveReplyPreventsDown(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	m := &mockNotify{}
	interval := 20 * time.Millisecond
	lcp := newTestLCP(t, ctx, interval, m.onEvent)

	// 模拟应答方: 每当进入 EchoReqSent, 回一个 EchoReply (rxr 处理)
	deadline := time.Now().Add(300 * time.Millisecond)
	cycles := 0
	for time.Now().Before(deadline) {
		if State(atomic.LoadUint32(lcp.state)) == StateEchoReqSent {
			reply := NewPkt(ProtoLCP)
			reply.Code = CodeEchoReply
			if err := lcp.rxr(ctx, reply); err != nil {
				t.Fatalf("rxr echo-reply: %v", err)
			}
			cycles++
			if State(atomic.LoadUint32(lcp.state)) != StateOpened {
				t.Fatalf("reply did not restore Opened, state=%v", State(atomic.LoadUint32(lcp.state)))
			}
			if got := atomic.LoadUint32(lcp.echoMissed); got != 0 {
				t.Fatalf("echoMissed=%d after reply, want 0", got)
			}
		}
		time.Sleep(5 * time.Millisecond)
	}
	if cycles < 3 {
		t.Fatalf("expected >=3 keepalive cycles, got %d (timer loop not arming?)", cycles)
	}
	if n := m.count(LCPLayerNotifyDown); n != 0 {
		t.Fatalf("Down fired %d times with a live peer, want 0", n)
	}
}

// 场景 C: 判死后恢复 — Down 后再次 Open/Up 进入正常周期, 计数从零开始
func TestKeepAliveRearmAfterTeardown(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	m := &mockNotify{}
	interval := 20 * time.Millisecond
	lcp := newTestLCP(t, ctx, interval, m.onEvent)

	if !waitFor(t, 3*time.Second, func() bool {
		return m.count(LCPLayerNotifyDown) >= 1
	}) {
		t.Fatalf("precondition: silent peer teardown did not fire")
	}
	// 会话重建: 模拟重拨成功后回到 Opened 并重置计数与定时器
	atomic.StoreUint32(lcp.echoMissed, 0)
	lcp.setState(StateOpened)
	lcp.resetKeepAliveTimer(ctx)
	if !waitFor(t, 3*interval+1*time.Second, func() bool {
		return State(atomic.LoadUint32(lcp.state)) == StateEchoReqSent
	}) {
		t.Fatalf("rearmed keepalive did not enter EchoReqSent")
	}
	if got := atomic.LoadUint32(lcp.echoMissed); got != 0 && got >= 3 {
		t.Fatalf("echoMissed=%d right after rearm, want 0..2", got)
	}
}
