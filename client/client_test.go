package client

/*
in order to run this test in ubuntu20.04, do following:
  - apt-get install ppp pppoe
  - sudo rm -rf /etc/ppp/options
  - sudo cp ./testdata/pppsvrconf/* /etc/ppp/
*/
import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/hujun-open/zouppp/lcp"
	"github.com/hujun-open/zouppp/pppoe"

	"github.com/hujun-open/etherconn"
)

const (
	svrNS  = "S"
	svrIF  = "S"
	clntIF = "C"
	uName  = "myclient"
	uPass  = "passwd123"
)

func execCMD(cmdstr string) error {
	// fmt.Printf("executing cmd %v\n", cmdstr)
	clist := strings.Fields(cmdstr)
	cmd := exec.Command(clist[0], clist[1:]...)
	return cmd.Run()
}

type testCMD struct {
	cmd        string
	ignoreFail bool
}

func testCreateVethLink() error {
	cmdList := []testCMD{
		{cmd: fmt.Sprintf("ip netns del %v", svrNS), ignoreFail: true},
		{cmd: fmt.Sprintf("ip link del %v", clntIF), ignoreFail: true},
		{cmd: fmt.Sprintf("ip netns add %v", svrNS)},
		{cmd: fmt.Sprintf("ip link add %v type veth peer name %v", svrIF, clntIF)},
		{cmd: fmt.Sprintf("ip link set %v netns %v", svrIF, svrNS)},
		{cmd: fmt.Sprintf("ip netns exec %v ip link set %v up", svrNS, svrIF)},
		{cmd: "ip netns exec S ip link set lo up"},
		{cmd: "ip netns exec S ip addr replace 127.0.0.1/32 dev lo"},
		{cmd: fmt.Sprintf("ip link set %v up", clntIF)},
	}
	for _, c := range cmdList {
		err := execCMD(c.cmd)
		if err != nil {
			if !c.ignoreFail {
				return err
			}
		}
	}
	return nil
}

const dontRunTestSvr = "no server"

func testRunSvr(ctx context.Context, cfg string) error {
	execCMD("pkill -9 pppoe-server")
	execCMD("pkill -9 pppoe")
	time.Sleep(time.Second)
	tmpf, err := ioutil.TempFile("", "pppoesvroptions_*")
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(tmpf.Name(), []byte(cfg), 0644)
	if err != nil {
		return err
	}
	tmpf.Close()
	cmd := fmt.Sprintf("netns exec %v pppoe-server -q /usr/sbin/pppd -I %v -F -O %v", svrNS, svrIF, tmpf.Name())
	if err := exec.CommandContext(ctx, "ip", strings.Fields(cmd)...).Start(); err != nil {
		return err
	}
	return nil

}

type testCase struct {
	desc       string
	zouconfig  *Config
	svrConfig  string
	shouldFail bool
}

// func testNewDefaultConfig() ClientConfig {
// 	return ClientConfig{
// 		Mac:       net.HardwareAddr{0xfa, 0x26, 0x67, 0x79, 0x18, 0x82},
// 		UserName:  uName,
// 		Password:  uPass,
// 		PPPIfName: "testppp0",
// 		setup: &Setup{
// 			rootLogger: rootlog,
// 			Timeout:    3 * time.Second,
// 			AuthProto:  lcp.ProtoPAP,
// 			IPv4:       true,
// 			IPv6:       true,
// 			Apply:      true,
// 		},
// 	}
// }
func TestPPPoE(t *testing.T) {
	rootlog, err := NewDefaultZouPPPLogger(LogLvlDebug)
	if err != nil {
		t.Fatalf("failed to create logger, %v", err)
	}

	caseList := []testCase{
		//case 0
		{
			desc: "pap v4 only",
			svrConfig: `
			name mysystem
			+pap
			-chap
			`,
			zouconfig: &Config{
				//Mac:       net.HardwareAddr{0xaa, 0xbb, 0xcc, 0x11, 0x22, 0x33},
				Mac:       net.HardwareAddr{0xfa, 0x26, 0x67, 0x79, 0x18, 0x82},
				UserName:  uName,
				Password:  uPass,
				PPPIfName: "testppp0",
				setup: &Setup{
					Logger:    rootlog,
					Timeout:   10 * time.Second,
					AuthProto: lcp.ProtoPAP,
					IPv4:      true,
					IPv6:      false,
				},
			},
		},
		//case 1
		{
			desc: "chap v4 only",
			svrConfig: `
			name mysystem
			-pap
			+chap
			`,
			zouconfig: &Config{
				//Mac:       net.HardwareAddr{0xaa, 0xbb, 0xcc, 0x11, 0x22, 0x33},
				Mac:       net.HardwareAddr{0xfa, 0x26, 0x67, 0x79, 0x18, 0x82},
				UserName:  uName,
				Password:  uPass,
				PPPIfName: "testppp0",
				setup: &Setup{
					Logger:    rootlog,
					Timeout:   3 * time.Second,
					AuthProto: lcp.ProtoCHAP,
					IPv4:      true,
					IPv6:      false,
				},
			},
		},
		//case 2
		{
			desc: "pap v6 only",
			svrConfig: `
			name mysystem
			+pap
			-chap
			noip
			+ipv6
			`,
			zouconfig: &Config{
				//Mac:       net.HardwareAddr{0xaa, 0xbb, 0xcc, 0x11, 0x22, 0x33},
				Mac:       net.HardwareAddr{0xfa, 0x26, 0x67, 0x79, 0x18, 0x82},
				UserName:  uName,
				Password:  uPass,
				PPPIfName: "testppp0",
				setup: &Setup{
					Logger: rootlog,

					Timeout:   3 * time.Second,
					AuthProto: lcp.ProtoPAP,
					IPv4:      false,
					IPv6:      true,
				},
			},
		},
		//case 3
		{
			desc: "chap v6 only",
			svrConfig: `
			name mysystem
			-pap
			+chap
			noip
			+ipv6
			`,
			zouconfig: &Config{
				//Mac:       net.HardwareAddr{0xaa, 0xbb, 0xcc, 0x11, 0x22, 0x33},
				Mac:       net.HardwareAddr{0xfa, 0x26, 0x67, 0x79, 0x18, 0x82},
				UserName:  uName,
				Password:  uPass,
				PPPIfName: "testppp0",
				setup: &Setup{
					Logger: rootlog,

					Timeout:   3 * time.Second,
					AuthProto: lcp.ProtoCHAP,
					IPv4:      false,
					IPv6:      true,
				},
			},
		},
		//case 4
		{
			desc: "chap dualstack",
			svrConfig: `
			name mysystem
			-pap
			+chap
			+ipv6
			`,
			zouconfig: &Config{
				//Mac:       net.HardwareAddr{0xaa, 0xbb, 0xcc, 0x11, 0x22, 0x33},
				Mac:       net.HardwareAddr{0xfa, 0x26, 0x67, 0x79, 0x18, 0x82},
				UserName:  uName,
				Password:  uPass,
				PPPIfName: "testppp0",
				setup: &Setup{
					Logger: rootlog,

					Timeout:   3 * time.Second,
					AuthProto: lcp.ProtoCHAP,
					IPv4:      true,
					IPv6:      true,
				},
			},
		},
		//case 5
		{
			desc: "pap dualstack",
			svrConfig: `
			name mysystem
			+pap
			-chap
			+ipv6
			`,
			zouconfig: &Config{
				//Mac:       net.HardwareAddr{0xaa, 0xbb, 0xcc, 0x11, 0x22, 0x33},
				Mac:       net.HardwareAddr{0xfa, 0x26, 0x67, 0x79, 0x18, 0x82},
				UserName:  uName,
				Password:  uPass,
				PPPIfName: "testppp0",
				setup: &Setup{
					Logger: rootlog,

					Timeout:   3 * time.Second,
					AuthProto: lcp.ProtoPAP,
					IPv4:      true,
					IPv6:      true,
				},
			},
		},
		//case 6
		{
			desc: "pap v4 only, wrong passwd, should fail",
			svrConfig: `
			name mysystem
			+pap
			-chap
			`,
			zouconfig: &Config{
				//Mac:       net.HardwareAddr{0xaa, 0xbb, 0xcc, 0x11, 0x22, 0x33},
				Mac:       net.HardwareAddr{0xfa, 0x26, 0x67, 0x79, 0x18, 0x82},
				UserName:  uName,
				Password:  "wrong",
				PPPIfName: "testppp0",
				setup: &Setup{
					Logger: rootlog,

					Timeout:   3 * time.Second,
					AuthProto: lcp.ProtoPAP,
					IPv4:      true,
					IPv6:      false,
				},
			},
			shouldFail: true,
		},
		//case 7
		{
			desc: "chap v4 only, wrong passwd, should fail",
			svrConfig: `
			name mysystem
			-pap
			+chap
			`,
			zouconfig: &Config{
				//Mac:       net.HardwareAddr{0xaa, 0xbb, 0xcc, 0x11, 0x22, 0x33},
				Mac:       net.HardwareAddr{0xfa, 0x26, 0x67, 0x79, 0x18, 0x82},
				UserName:  uName,
				Password:  "wrong",
				PPPIfName: "testppp0",
				setup: &Setup{
					Logger: rootlog,

					Timeout:   3 * time.Second,
					AuthProto: lcp.ProtoCHAP,
					IPv4:      true,
					IPv6:      false,
				},
			},
			shouldFail: true,
		},
		//case 8
		{
			desc:      "no pppoesvr, should fail",
			svrConfig: dontRunTestSvr,
			zouconfig: &Config{
				//Mac:       net.HardwareAddr{0xaa, 0xbb, 0xcc, 0x11, 0x22, 0x33},
				Mac:       net.HardwareAddr{0xfa, 0x26, 0x67, 0x79, 0x18, 0x82},
				UserName:  uName,
				Password:  uPass,
				PPPIfName: "testppp0",
				setup: &Setup{
					Logger: rootlog,

					Timeout:   3 * time.Second,
					AuthProto: lcp.ProtoPAP,
					IPv4:      true,
					IPv6:      false,
				},
			},
			shouldFail: true,
		},
		//case 9
		{
			desc: "no auth configured on svr side, should fail",
			svrConfig: `
			name mysystem
			noauth
			`,
			zouconfig: &Config{
				//Mac:       net.HardwareAddr{0xaa, 0xbb, 0xcc, 0x11, 0x22, 0x33},
				Mac:       net.HardwareAddr{0xfa, 0x26, 0x67, 0x79, 0x18, 0x82},
				UserName:  uName,
				Password:  uPass,
				PPPIfName: "testppp0",
				setup: &Setup{
					Logger: rootlog,

					Timeout:   3 * time.Second,
					AuthProto: lcp.ProtoPAP,
					IPv4:      true,
					IPv6:      false,
				},
			},
			shouldFail: true,
		},
		//case 10
		{
			desc: "pap on client while chap on svr, should fail",
			svrConfig: `
			name mysystem
			refuse-pap
			require-chap
			`,
			zouconfig: &Config{
				//Mac:       net.HardwareAddr{0xaa, 0xbb, 0xcc, 0x11, 0x22, 0x33},
				Mac:       net.HardwareAddr{0xfa, 0x26, 0x67, 0x79, 0x18, 0x82},
				UserName:  uName,
				Password:  uPass,
				PPPIfName: "testppp0",
				setup: &Setup{
					Logger: rootlog,

					Timeout:   3 * time.Second,
					AuthProto: lcp.ProtoPAP,
					IPv4:      true,
					IPv6:      false,
				},
			},
			shouldFail: true,
		},
		//case 11
		{
			desc: "chap on client while pap on svr, should fail",
			svrConfig: `
			name mysystem
			refuse-chap
			require-pap
			`,
			zouconfig: &Config{
				//Mac:       net.HardwareAddr{0xaa, 0xbb, 0xcc, 0x11, 0x22, 0x33},
				Mac:       net.HardwareAddr{0xfa, 0x26, 0x67, 0x79, 0x18, 0x82},
				UserName:  uName,
				Password:  uPass,
				PPPIfName: "testppp0",
				setup: &Setup{
					Logger: rootlog,

					Timeout:   3 * time.Second,
					AuthProto: lcp.ProtoCHAP,
					IPv4:      true,
					IPv6:      false,
				},
			},
			shouldFail: true,
		},
		//case 12
		{
			desc: "chap, client requires dualstack, but svr only configures with v4, should fail",
			svrConfig: `
			name mysystem
			-pap
			+chap
			-ipv6
			`,
			zouconfig: &Config{
				//Mac:       net.HardwareAddr{0xaa, 0xbb, 0xcc, 0x11, 0x22, 0x33},
				Mac:       net.HardwareAddr{0xfa, 0x26, 0x67, 0x79, 0x18, 0x82},
				UserName:  uName,
				Password:  uPass,
				PPPIfName: "testppp0",
				setup: &Setup{
					Logger: rootlog,

					Timeout:   3 * time.Second,
					AuthProto: lcp.ProtoCHAP,
					IPv4:      true,
					IPv6:      true,
				},
			},
			shouldFail: true,
		},
		//case 13
		{
			desc: "chap, client requires dualstack, but svr only configures with v6, should fail",
			svrConfig: `
			name mysystem
			-pap
			+chap
			+ipv6
			noip
			`,
			zouconfig: &Config{
				//Mac:       net.HardwareAddr{0xaa, 0xbb, 0xcc, 0x11, 0x22, 0x33},
				Mac:       net.HardwareAddr{0xfa, 0x26, 0x67, 0x79, 0x18, 0x82},
				UserName:  uName,
				Password:  uPass,
				PPPIfName: "testppp0",
				setup: &Setup{
					Logger:    rootlog,
					Timeout:   3 * time.Second,
					AuthProto: lcp.ProtoCHAP,
					IPv4:      true,
					IPv6:      true,
				},
			},
			shouldFail: true,
		},
	}

	testFunc := func(c testCase, t *testing.T, usingxdp bool) error {
		resultch := make(chan *DialResult)
		stopch := make(chan struct{})
		c.zouconfig.setup.ResultCh = resultch
		c.zouconfig.setup.StopResultCh = stopch
		c.zouconfig.setup.NumOfClients = 1
		summaryCh := make(chan *ResultSummary)
		go CollectResults(c.zouconfig.setup, summaryCh)
		err := testCreateVethLink()
		if err != nil {
			t.Fatal(err)
		}
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		var relay etherconn.PacketRelay
		if !usingxdp {
			relay, err = etherconn.NewRawSocketRelay(ctx, clntIF,
				etherconn.WithDebug(true),
				etherconn.WithRecvTimeout(c.zouconfig.setup.Timeout),
				etherconn.WithBPFFilter(`(ether proto 0x8863 or 0x8864) or (vlan and ether proto 0x8863 or 0x8864)`),
			)
		} else {
			relay, err = etherconn.NewXDPRelay(ctx, clntIF,
				// etherconn.WithQueueID([]int{0}),
				etherconn.WithXDPDebug(true),
				etherconn.WithXDPEtherTypes([]uint16{0x8863, 0x8864}),
				etherconn.WithXDPDefaultReceival(false),
			)

		}
		if err != nil {
			t.Fatal(err)
		}
		defer relay.Stop()
		dialwg := new(sync.WaitGroup)
		dialwg.Add(1)
		econn := etherconn.NewEtherConn(c.zouconfig.Mac, relay,
			etherconn.WithVLANs(c.zouconfig.VLANs),
			etherconn.WithEtherTypes([]uint16{pppoe.EtherTypePPPoEDiscovery, pppoe.EtherTypePPPoESession}),
			etherconn.WithRecvMulticast(true))
		// err = execCMD("ip netns exec S ip link set S xdp object xdpethfilter_kern.o section xdp_pass_sec")
		if err != nil {
			t.Fatal(err)
		}
		if c.svrConfig != dontRunTestSvr {
			err = testRunSvr(ctx, c.svrConfig)
			if err != nil {
				t.Fatal(err)
			}
			defer func() {
				execCMD("pkill -9 pppoe-server")
				execCMD("pkill -9 pppoe")
				time.Sleep(3 * time.Second)
			}()
		}
		z, err := NewZouPPP(econn, c.zouconfig, WithDialWG(dialwg))
		if err != nil {
			return err
		}
		go execCMD("ip netns exec S tcpdump -n -i S -w s.pcap")
		go z.Dial(ctx)
		dialwg.Wait()
		summary := <-summaryCh
		close(c.zouconfig.setup.StopResultCh)
		// ch := make(chan int)
		// <-ch

		if summary.Success != 1 {
			return fmt.Errorf("failed")
		}
		if z.fastpath == nil && z.cfg.setup.Apply {
			return fmt.Errorf("datapath creation failed")
		}
		return nil
	}
	for i, c := range caseList {
		// if i != len(caseList)-1 {
		// 	continue
		// }
		// if c.desc != "no pppoesvr, should fail" {
		// 	continue
		// }
		// if i != 1 {
		// 	continue
		// }
		t.Logf("-----> start case %d using rawrelay: %v", i, c.desc)
		err := testFunc(c, t, false)
		if err != nil {
			if c.shouldFail {
				t.Logf("case %d: %v, failed as expected", i, c.desc)
			} else {
				t.Fatalf("case %d: %v failed, %v", i, c.desc, err)
			}
		} else {
			if c.shouldFail {
				t.Fatalf("case %d: %v should failed but succeed", i, c.desc)
			}
		}
		//using xdp
		t.Logf("-----> start case %d using xdprelay: %v", i, c.desc)
		err = testFunc(c, t, true)
		if err != nil {
			if c.shouldFail {
				t.Logf("case %d: %v, failed as expected", i, c.desc)
			} else {
				t.Fatalf("case %d: %v failed, %v", i, c.desc, err)
			}
		} else {
			if c.shouldFail {
				t.Fatalf("case %d: %v should failed but succeed", i, c.desc)
			}
		}
	}

}
func TestMain(m *testing.M) {
	runtime.SetBlockProfileRate(1000000000)
	go func() {
		log.Println(http.ListenAndServe("0.0.0.0:6060", nil))
	}()
	result := m.Run()
	os.Exit(result)
}
