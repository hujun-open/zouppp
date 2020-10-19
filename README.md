# zouppp
zouppp is a set of GO modules implements PPPoE and related protocols:

 * zouppp/pppoe: PPPoE RFC2516
 * zouppp/lcp: PPP/LCP RFC1661; IPCP RFC1332; IPv6CP RFC5072;
 * zouppp/pap: PAP RFC1334
 * zouppp/chap: CHAP RFC1994
 * zouppp/datapath: linux datapath
 * zouppp/client: PPPoE Client

## Example PPPoE Client for testing purpose
The main module implements an example PPPoE test client with load testing capability. it could also be used as a starting point to implement your own PPPoE client;

```
~/gowork/src/zouppp# ./zouppp -?
flag provided but not defined: -?
Usage of ./zouppp:
  -a    apply the network config
  -cid string
        pppoe tag circuit-id
  -excludedvlans string
        excluded vlan IDs
  -i string
        interface name
  -interval duration
        interval between launching client (default 1ms)
  -l uint
        log level
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
        pppoe tag remote-id
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
```



