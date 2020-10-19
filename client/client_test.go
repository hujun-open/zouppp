package client

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
	"zouppp/lcp"
	"zouppp/pppoe"

	"github.com/hujun-open/etherconn"
)

const (
	svrNS  = "S"
	svrIF  = "S"
	clntIF = "C"
	uName  = "myclient"
	uPass  = "passwd123"
)

const testKillScript = `#!/usr/bin/env bash

if [[ $# -lt 1 ]]; then
  echo "missing program name"
  exit
fi

pids=` + "`" + `ps h -C $1 -o pid` + "`" + `
if [[ -z $pids ]]; then
  echo "can't find process $1"
  exit
fi
while IFS= read -r pid; do
    echo "killing $pid ..."
    kill $pid
done <<< "$pids"
`

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
		testCMD{cmd: fmt.Sprintf("ip netns del %v", svrNS), ignoreFail: true},
		testCMD{cmd: fmt.Sprintf("ip link del %v", clntIF), ignoreFail: true},
		testCMD{cmd: fmt.Sprintf("ip netns add %v", svrNS)},
		testCMD{cmd: fmt.Sprintf("ip link add %v type veth peer name %v", svrIF, clntIF)},
		testCMD{cmd: fmt.Sprintf("ip link set %v netns %v", svrIF, svrNS)},
		testCMD{cmd: fmt.Sprintf("ip netns exec %v ip link set %v up", svrNS, svrIF)},
		testCMD{cmd: "ip netns exec S ip link set lo up"},
		testCMD{cmd: "ip netns exec S ip addr replace 127.0.0.1/32 dev lo"},
		testCMD{cmd: fmt.Sprintf("ip link set %v up", clntIF)},
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
	tmpkillf, err := ioutil.TempFile("", "killbyname")
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(tmpkillf.Name(), []byte(testKillScript), 0750)
	if err != nil {
		return err
	}
	os.Chmod(tmpkillf.Name(), 0750)
	tmpkillf.Close()
	execCMD(fmt.Sprintf("%v pppoe-server", tmpkillf.Name()))
	tmpf, err := ioutil.TempFile("", "pppoesvroptions_*")
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(tmpf.Name(), []byte(cfg), 0644)
	if err != nil {
		return err
	}
	tmpf.Close()
	cmd := fmt.Sprintf("netns exec %v pppoe-server -I %v -F -O %v", svrNS, svrIF, tmpf.Name())
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
		testCase{
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
					Timeout:   3 * time.Second,
					AuthProto: lcp.ProtoPAP,
					IPv4:      true,
					IPv6:      false,
				},
			},
		},

		testCase{
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

		testCase{
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

		testCase{
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

		testCase{
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

		testCase{
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

		testCase{
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

		testCase{
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

		testCase{
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

		testCase{
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

		testCase{
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

		testCase{
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

		testCase{
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

		testCase{
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

	testFunc := func(c testCase, t *testing.T) error {
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
		relay, err := etherconn.NewRawSocketRelay(ctx, clntIF,
			etherconn.WithEtherType([]uint16{pppoe.EtherTypePPPoEDiscovery, pppoe.EtherTypePPPoESession}),
			etherconn.WithDebug(true), etherconn.WithRecvTimeout(3*time.Second))
		if err != nil {
			t.Fatal(err)
		}
		dialwg := new(sync.WaitGroup)
		dialwg.Add(1)
		econn := etherconn.NewEtherConn(c.zouconfig.Mac, relay,
			etherconn.WithVLANs(c.zouconfig.VLANs),
			etherconn.WithRecvMulticast(true))
		if c.svrConfig != dontRunTestSvr {
			err = testRunSvr(ctx, c.svrConfig)
			if err != nil {
				t.Fatal(err)
			}
		}
		z, err := NewZouPPP(econn, c.zouconfig, WithDialWG(dialwg))
		if err != nil {
			return err
		}
		go z.Dial(ctx)
		dialwg.Wait()
		summary := <-summaryCh
		close(c.zouconfig.setup.StopResultCh)

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
		t.Logf("-----> start case %d: %v", i, c.desc)
		err := testFunc(c, t)
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
