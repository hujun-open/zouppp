# zouppp
![Build Status](https://github.com/hujun-open/zouppp/actions/workflows/main.yml/badge.svg)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/hujun-open/zouppp)](https://pkg.go.dev/github.com/hujun-open/zouppp)

zouppp is a set of GO modules implements PPPoE and related protocols:

 * zouppp/pppoe: PPPoE RFC2516
 * zouppp/lcp: PPP/LCP RFC1661; IPCP RFC1332; IPv6CP RFC5072;
 * zouppp/pap: PAP RFC1334
 * zouppp/chap: CHAP RFC1994
 * zouppp/datapath: linux datapath
 * zouppp/client: PPPoE Client

## PPPoE Client
The main module implements a PPPoE test client with load testing capability. it could also be used as a starting point to implement your own PPPoE client;

It has following key features:

- Custom VLAN/MAC address without provisioning OS interface (via [etherconn](https://github.com/hujun-open/etherconn))
- Load testing, able to initiate large amount of PPPoE session at the same time
- Option to not creating corresponding PPP TUN interface in OS, e.g. only do control plane processing, this is useful for protocol level only load testing.
- Support BBF PPPoE tag: circuit-id/remote-id
- IPv4, IPv6 and dual-stack
 

### Example Client Usage

1. on interface eth1, create 100 PPPoE session, CHAP, IPv4 only

`zouppp -i eth1 -u testuser -p passwd123 -l 1 -v6=false -n 100`

2. #1 variant, using PAP

`zouppp -i eth1 -u testuser -p passwd123 -l 1 -v6=false -n 100 -pap`

3. #1 variant, using vlan 100, svlan 200

`zouppp -i eth1 -u testuser -p passwd123 -l 1 -v6=false -n 100 -vlan 100 -svlan 200`

4. #3 variant, using custom mac 

`zouppp -i eth1 -u testuser -p passwd123 -l 1 -v6=false -n 100 -vlan 100 -svlan 200 -mac "aa:bb:cc:11:22:33"`

5. #1 variant, don't create PPP TUN interface

`zouppp -i eth1 -u testuser -p passwd123 -l 1 -v6=false -n 100 -a=false`

6. #1 variant, each session use different username and password, e.g. first one username is "testuser-0", 2nd one is "testuser-1" ..etc; password following same rule

`zouppp -i eth1 -u testuser-@ID -p passwd123-@ID -l 1 -v6=false -n 100`

7. #1 variant, each session add BBF remote-id tag, first session remote-id tag is "remote-id-0", 2nd one is "remote-id-1" ..etc;

`zouppp -i eth1 -u testuser -p passwd123 -l 1 -v6=false -n 100 -rid remote-id-@id`

8. #1 variant, use XDP socket;

`zouppp -i eth1 -u testuser -p passwd123 -l 1 -v6=false -n 100 -xdp`

### CLI

```
flag provided but not defined: -?
Usage of ./zouppp:
  -a    apply the network config, set false to skip creating the PPP TUN if
  -cid string
        pppoe BBF tag circuit-id
  -excludedvlans string
        excluded vlan IDs
  -i string
        interface name
  -interval duration
        interval between launching client (default 1ms)
  -l uint
        log level: 0,1,2
  -mac string
        mac address
  -macstep uint
        mac address step (default 1)
  -n uint
        number of clients (default 1)
  -p string
        password
  -pap
        use PAP instead of CHAP
  -pppif string
        ppp interface name, must contain @ID (default "zouppp@ID")
  -profiling
        enable profiling, only for dev use
  -retry uint
        number of retry (default 3)
  -rid string
        pppoe BBF tag remote-id
  -svlan int
        svlan tag (default -1)
  -svlanetype uint
        svlan tag EtherType (default 33024)
  -timeout duration
        timeout (default 5s)
  -u string
        user name
  -v4
        enable Ipv4 (default true)
  -v6
        enable Ipv6 (default true)
  -vlan int
        vlan tag (default -1)
  -vlanetype uint
        vlan tag EtherType (default 33024)
  -vlanstep uint
        VLAN Id step
  -xdp
        use XDP
```



