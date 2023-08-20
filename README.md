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
 * zouppp/client.DHCP6Clnt: DHCPv6 client

**note: zouppp focus on client side, a PPPoE/PPP server requires addtional logic/code**

## PPPoE Client
The main module implements a PPPoE test client with load testing capability. it could also be used as a starting point to implement your own PPPoE client;

It has following key features:

- Custom VLAN/MAC address without provisioning OS interface (via [etherconn](https://github.com/hujun-open/etherconn))
- Load testing, able to initiate large amount of PPPoE session at the same time
- Option to not creating corresponding PPP TUN interface in OS, e.g. only do control plane processing, this is useful for protocol level only load testing.
- Support BBF PPPoE tag: circuit-id/remote-id
- IPv4, IPv6 and dual-stack
- DHCPv6 over PPP,  IA_NA and/or IA_PD
 

### Example Client Usage

1. on interface eth1, create 100 PPPoE session, CHAP, IPv4 only, enable debug logging

`zouppp -i eth1 -u testuser -p passwd123 -l debug -v6=false -n 100`

2. #1 variant, using PAP

`zouppp -i eth1 -u testuser -p passwd123 -l debug -v6=false -n 100 -authproto PAP`

3. #1 variant, using QinQ 100.200

`zouppp -i eth1 -u testuser -p passwd123 -l debug -v6=false -n 100 -vlan 100.200`

4. #3 variant, using custom mac 

`zouppp -i eth1 -u testuser -p passwd123 -l debug -v6=false -n 100 -vlan 100.200 -mac "aa:bb:cc:11:22:33"`

5. #1 variant, don't create PPP TUN interface

`zouppp -i eth1 -u testuser -p passwd123 -l debug -v6=false -n 100 -apply=false`

6. #1 variant, each session use different username and password, e.g. first one username is "testuser-0", 2nd one is "testuser-1" ..etc; password following same rule

`zouppp -i eth1 -u testuser-@ID -p passwd123-@ID -l debug -v6=false -n 100`

7. #1 variant, each session add BBF remote-id tag, first session remote-id tag is "remote-id-0", 2nd one is "remote-id-1" ..etc;

`zouppp -i eth1 -u testuser -p passwd123 -l debug -v6=false -n 100 -rid remote-id-@id`

8. #1 variant, use XDP socket;

`zouppp -i eth1 -u testuser -p passwd123 -l debug -v6=false -n 100 -xdp`

9. #1 variant, running DHCPv6 over ppp, requesting IA_NA and IA_PD
`zouppp -i eth1 -u testuser -p passwd123 -l debug -n 100 -dhcp6iana -dhcp6iapd`

### CLI

```
Usage:
a pppoe testing tool
  - apply: if Apply is true, then create a PPP interface with assigned addresses; could be set to false if only to test protocol
        default:true
  - authproto: auth protocol, PAP or CHAP
        default:CHAP
  - cid: BBF circuit-id
  - dhcpv6iana: run DHCPv6 over PPP to get an IANA address
        default:false
  - dhcpv6iapd: run DHCPv6 over PPP to get an IAPD prefix
        default:false
  - excludedvlans: a list of excluded VLAN id, apply to all layer of vlans
  - i: listening interface name
  - interval: amount of time to wait between launching each session
        default:0s
  - l: log levl, err|info|debug
        default:err
  - mac: start MAC address
  - macstep: MAC step to increase for each client
        default:0
  - n: number of PPPoE clients
        default:1
  - p: PAP/CHAP password
  - pppifname: name of PPP interface created after successfully dialing, must contain @ID
        default:zouppp@ID
  - profiling: enable profiling, dev use only
        default:false
  - retry: number of setup retry
        default:0
  - rid: BBF remote-id
  - timeout: setup timeout
        default:0s
  - u: PAP/CHAP username
  - v4: run IPCP
        default:true
  - v6: run IPv6CP
        default:false
  - vlan: start VLAN id, could be Dot1q or QinQ
  - vlanstep: VLAN step to increase for each client
        default:0
  - xdp: use XDP to forward packet
        default:false

  -cfgfromfile: load configuration from the specified file
        default:zouppp.conf

```
### Config File
Thanks to [shouchan](https://github.com/hujun-open/shouchan), beside using CLI parameters, a YAML config file could also be used via "-cfgfromfile <conf_file>", the content of YAML is the client.Setup struct 


