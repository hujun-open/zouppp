/*
zouppp is a set of GO modules implements PPPoE and related protocols:

  - zouppp/pppoe: PPPoE RFC2516

  - zouppp/lcp: PPP/LCP RFC1661; IPCP RFC1332; IPv6CP RFC5072;

  - zouppp/pap: PAP RFC1334

  - zouppp/chap: CHAP RFC1994

  - zouppp/datapath: linux datapath

  - zouppp/client: PPPoE Client

The main module implements a simple PPPoE test client with load testing capability. Could also be used as a starting point to implement your own PPPoE client;
*/
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"log"
	"net/http"
	_ "net/http/pprof"
	"runtime"
	"time"

	"github.com/hujun-open/etherconn"
	"github.com/hujun-open/shouchan"
	"github.com/hujun-open/zouppp/client"
	"github.com/hujun-open/zouppp/pppoe"
)

func main() {
	setup := client.DefaultSetup()
	cnf, err := shouchan.NewSConf(setup, "zouppp", "a pppoe testing tool", shouchan.WithDefaultConfigFilePath[*client.Setup]("zouppp.conf"))
	if err != nil {
		log.Fatalf("failed to create configuration, %v", err)
	}
	cnf.ReadwithCMDLine()
	setup = cnf.GetConf()
	err = setup.Init()
	if err != nil {
		log.Fatalf("invalid setup, %v", err)
	}
	// create a list of client.Config based on the setup
	cfglist, err := client.GenClientConfigurations(setup)
	if err != nil {
		log.Fatalf("failed to generate client config, %v", err)
	}
	// profiling is for dev use only
	if setup.Profiling {
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
	if !setup.XDP {
		relay, err = etherconn.NewRawSocketRelay(ctx, setup.Ifname,
			etherconn.WithDebug(setup.LogLevel == client.LogLvlDebug),
			etherconn.WithBPFFilter(`(ether proto 0x8863 or 0x8864) or (vlan and ether proto 0x8863 or 0x8864)`),
			etherconn.WithRecvTimeout(setup.Timeout))
		if err != nil {
			setup.Logger().Sugar().Errorf("failed to create raw packet relay, %v", err)
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
			setup.Logger().Sugar().Errorf("failed to create xdp packet relay, %v", err)
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
			setup.Logger().Sugar().Errorf("failed to create zouppp,%v", err)
			return
		}
		go z.Dial(ctx)
		clntList = append(clntList, z)
		// sleep for dialing interval
		time.Sleep(setup.Interval)
	}
	// wait for all sessions dialing finish
	dialwg.Wait()
	setup.Logger().Sugar().Info("all sessions dialing finished")
	// get the dailing result summary
	summary := <-summaryCh
	fmt.Println(summary)
	setup.Close()
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
