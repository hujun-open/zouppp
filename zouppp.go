/*zouppp is a set of GO modules implements PPPoE and related protocols:

 * zouppp/pppoe: PPPoE RFC2516

 * zouppp/lcp: PPP/LCP RFC1661; IPCP RFC1332; IPv6CP RFC5072;

 * zouppp/pap: PAP RFC1334

 * zouppp/chap: CHAP RFC1994

 * zouppp/datapath: linux datapath

 * zouppp/client: PPPoE Client

The main module implements a simple PPPoE test client with load testing capability. Could also be used as a starting point to implement your own PPPoE client;


*/
package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"

	"flag"
	"log"
	"net/http"
	_ "net/http/pprof"
	"runtime"
	"time"

	"github.com/hujun-open/etherconn"
	"github.com/hujun-open/zouppp/client"
	"github.com/hujun-open/zouppp/lcp"
	"github.com/hujun-open/zouppp/pppoe"
	"go.uber.org/zap"
)

const resultChannelDepth = 128

func newSetupviaFlags(
	Ifname string,
	NumOfClients uint,
	retry uint,
	timeout time.Duration,
	StartMAC string,
	MacStep uint,
	vlan int,
	vlanetype uint,
	svlan int,
	svlanetype uint,
	VLANStep uint,
	excludevlanid string,
	Interval time.Duration,
	logl uint,
	rid, cid string,
	uname, passwd string,
	usepap bool,
	v4, v6 bool,
	pppifname string,
	Apply bool,
	rootlog *zap.Logger,
) (*client.Setup, error) {
	var r client.Setup
	var err error
	r.Logger = rootlog
	if Ifname == "" {
		return nil, fmt.Errorf("interface name can't be empty")
	}
	if NumOfClients == 0 {
		return nil, fmt.Errorf("number of clients can't be zero")
	}
	iff, err := net.InterfaceByName(Ifname)
	if err != nil {
		return nil, fmt.Errorf("can't find interface %v,%w", Ifname, err)
	}
	r.Ifname = Ifname
	if NumOfClients == 0 {
		return nil, fmt.Errorf("number of client is 0")
	}
	r.NumOfClients = NumOfClients
	if StartMAC == "" {
		r.StartMAC = iff.HardwareAddr
	} else {
		r.StartMAC, err = net.ParseMAC(StartMAC)
		if err != nil {
			return nil, err
		}
	}
	r.MacStep = MacStep
	r.Retry = retry
	r.Timeout = timeout
	chkVIDFunc := func(vid int) bool {
		if vid < 0 || vid > 4095 {
			return false
		}
		return true
	}

	newvlanFunc := func(id int, etype uint) *etherconn.VLAN {
		if id >= 0 {
			return &etherconn.VLAN{
				ID:        uint16(id),
				EtherType: uint16(etype),
			}
		}
		return nil
	}
	if chkVIDFunc(vlan) {
		r.StartVLANs = etherconn.VLANs{}
		if v := newvlanFunc(vlan, vlanetype); v != nil {
			r.StartVLANs = append(r.StartVLANs, v)
		}
		if chkVIDFunc(svlan) {
			if v := newvlanFunc(svlan, svlanetype); v != nil {
				r.StartVLANs = append(etherconn.VLANs{v}, r.StartVLANs...)
			}
		}

	} else {
		if chkVIDFunc(svlan) {
			return nil, fmt.Errorf("spcifying svlan also require specifying a valid vlan")
		}
	}

	r.VLANStep = VLANStep
	strToExtListFunc := func(exts string) []uint16 {
		vidstrList := strings.Split(exts, ",")
		r := []uint16{}
		for _, vidstr := range vidstrList {
			vid, err := strconv.Atoi(vidstr)
			if err == nil {
				if vid >= 0 && vid <= 4095 {
					r = append(r, uint16(vid))
				}
			}
		}
		return r
	}
	r.ExcludedVLANs = strToExtListFunc(excludevlanid)
	r.Interval = Interval
	r.LogLevel = client.LoggingLvl(logl)
	r.RID = rid
	r.CID = cid
	r.UserName = uname
	r.Password = passwd
	r.AuthProto = lcp.ProtoCHAP
	if usepap {
		r.AuthProto = lcp.ProtoPAP
	}
	r.IPv4 = v4
	r.IPv6 = v6
	r.Apply = Apply
	if !strings.Contains(pppifname, client.VarName) {
		return nil, fmt.Errorf("ppp interface name must contain %v", client.VarName)
	}
	r.PPPIfName = pppifname
	r.ResultCh = make(chan *client.DialResult, resultChannelDepth)
	r.StopResultCh = make(chan struct{})
	return &r, nil
}

func main() {
	ifname := flag.String("i", "", "interface name")
	loglvl := flag.Uint("l", uint(client.LogLvlErr), fmt.Sprintf("log level: %d,%d,%d", client.LogLvlErr, client.LogLvlInfo, client.LogLvlDebug))
	mac := flag.String("mac", "", "mac address")
	macstep := flag.Uint("macstep", 1, "mac address step")
	vlanstep := flag.Uint("vlanstep", 0, "VLAN Id step")
	clntnum := flag.Uint("n", 1, "number of clients")
	uname := flag.String("u", "", "user name")
	passwd := flag.String("p", "", "password")
	usepap := flag.Bool("pap", false, "use PAP instead of CHAP")
	cid := flag.String("cid", "", "pppoe BBF tag circuit-id")
	rid := flag.String("rid", "", "pppoe BBF tag remote-id")
	ipv4 := flag.Bool("v4", true, "enable Ipv4")
	ipv6 := flag.Bool("v6", true, "enable Ipv6")
	vlanid := flag.Int("vlan", -1, "vlan tag")
	vlantype := flag.Uint("vlanetype", 0x8100, "vlan tag EtherType")
	svlanid := flag.Int("svlan", -1, "svlan tag")
	svlantype := flag.Uint("svlanetype", 0x8100, "svlan tag EtherType")
	profiling := flag.Bool("profiling", false, "enable profiling, only for dev use")
	excludevlanid := flag.String("excludedvlans", "", "excluded vlan IDs")
	retry := flag.Uint("retry", 3, "number of retry")
	timeout := flag.Duration("timeout", 5*time.Second, "timeout")
	interval := flag.Duration("interval", time.Millisecond, "interval between launching client")
	pppifname := flag.String("pppif", client.DefaultPPPIfNameTemplate, fmt.Sprintf("ppp interface name, must contain %v", client.VarName))
	usexdp := flag.Bool("xdp", false, "use XDP")
	apply := flag.Bool("a", false, "apply the network config, set false to skip creating the PPP TUN if")
	flag.Parse()
	rootlog, err := client.NewDefaultZouPPPLogger(client.LoggingLvl(*loglvl))
	if err != nil {
		log.Fatalf("failed to create logger, %v", err)
	}
	// getting setup from flags
	setup, err := newSetupviaFlags(
		*ifname,
		*clntnum,
		*retry,
		*timeout,
		*mac,
		*macstep,
		*vlanid,
		*vlantype,
		*svlanid,
		*svlantype,
		*vlanstep,
		*excludevlanid,
		*interval,
		*loglvl,
		// *debug,
		*rid, *cid,
		*uname, *passwd,
		*usepap,
		*ipv4, *ipv6,
		*pppifname,
		*apply,
		rootlog,
	)
	if err != nil {
		log.Fatalf("invalid parameter, %v", err)
	}

	// create a list of client.Config based on the setup
	cfglist, err := client.GenClientConfigurations(setup)
	if err != nil {
		log.Fatalf("failed to generate client config, %v", err)
	}
	// profiling is for dev use only
	if *profiling {
		runtime.SetBlockProfileRate(1000000000)
		go func() {
			log.Println(http.ListenAndServe("0.0.0.0:6060", nil))
		}()
	}
	// create a context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// create a etherconn.RawSoketRelay
	var relay etherconn.PacketRelay
	if !*usexdp {
		relay, err = etherconn.NewRawSocketRelay(ctx, setup.Ifname,
			etherconn.WithDebug(setup.LogLevel == client.LogLvlDebug),
			etherconn.WithBPFFilter(`(ether proto 0x8863 or 0x8864) or (vlan and ether proto 0x8863 or 0x8864)`),
			etherconn.WithRecvTimeout(setup.Timeout))
		if err != nil {
			rootlog.Sugar().Errorf("failed to create raw packet relay, %v", err)
			return
		}
	} else {
		relay, err = etherconn.NewXDPRelay(ctx, setup.Ifname,
			etherconn.WithXDPDebug(setup.LogLevel == client.LogLvlDebug),
			etherconn.WithXDPEtherTypes([]uint16{0x8863, 0x8864}),
			etherconn.WithXDPDefaultReceival(false),
			etherconn.WithXDPSendChanDepth(10240),
			etherconn.WithXDPUMEMNumOfTrunk(65536),
		)
		if err != nil {
			rootlog.Sugar().Errorf("failed to create xdp packet relay, %v", err)
			return
		}
	}
	// create a ResultSummary channel to get a ResultSummary upon dial finishes
	summaryCh := make(chan *client.ResultSummary)
	go client.CollectResults(setup, summaryCh)
	// creates dialwg to get dialing results
	dialwg := new(sync.WaitGroup)
	dialwg.Add(len(cfglist))
	// sessionwg.Wait to wait for all opened sessions
	sessionwg := new(sync.WaitGroup)
	// start dialing
	var clntList []*client.ZouPPP
	for _, cfg := range cfglist {
		econn := etherconn.NewEtherConn(cfg.Mac, relay,
			etherconn.WithEtherTypes([]uint16{pppoe.EtherTypePPPoEDiscovery, pppoe.EtherTypePPPoESession}),
			etherconn.WithVLANs(cfg.VLANs), etherconn.WithRecvMulticast(true))
		z, err := client.NewZouPPP(econn, cfg, client.WithDialWG(dialwg), client.WithSessionWG(sessionwg))
		if err != nil {
			rootlog.Sugar().Errorf("failed to create zouppp,%v", err)
			return
		}
		go z.Dial(ctx)
		clntList = append(clntList, z)
		// sleep for dialing interval
		time.Sleep(setup.Interval)
	}
	// wait for all sessions dialing finish
	dialwg.Wait()
	rootlog.Sugar().Info("all sessions dialing finished")
	// get the dailing result summary
	summary := <-summaryCh
	fmt.Println(summary)
	close(setup.StopResultCh)
	// handle ctrl+c
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("stopping...")
		for _, z := range clntList {
			z.Close()
		}
	}()
	// wait for all opened sessions to close
	if summary.Success > 0 {
		sessionwg.Wait()
	}
	fmt.Println("done")

}
