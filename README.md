# netutils

Utils for networking engineers: a lightweight (Alpine-based) container packed with network troubleshooting tools.

## Build

```bash
docker build -t netutils .
```

Optional extras (e.g. Wireshark CLI, ~+100 MB):

```bash
docker build -t netutils --build-arg EXTRA_PKGS="tshark" .
```

## Run

```bash
# interactive shell
docker run --rm -it netutils

# one-shot commands
docker run --rm netutils tcping -c 4 google.com 443
docker run --rm netutils paping google.com -p 443 -c 4

# share the host network stack (see/capture host interfaces)
docker run --rm -it --net host netutils

# attach to another container's network namespace
docker run --rm -it --net container:<name> netutils
```

Kubernetes:

```bash
kubectl run netutils --rm -it --image=<registry>/netutils -- bash
kubectl debug -it <pod> --image=<registry>/netutils --target=<container>
```

### Capabilities

No extra flags are needed for normal troubleshooting. Docker's default capabilities
already include `NET_RAW`, so `ping`, `traceroute`, `mtr`, `tcpdump`, `arping`, `nmap`,
`curl`, `dig`, `tcping`, etc. all work out of the box.

Add `--cap-add NET_ADMIN` only when you need to **read or change** network config:
`iptables`/`nft`, `conntrack`, `ip link/addr/route` changes, `ethtool` settings.

| Situation | What to add |
|---|---|
| Docker, normal use | nothing |
| Docker, firewall/interface changes | `--cap-add NET_ADMIN` |
| Kubernetes, default namespace | nothing |
| Kubernetes with Pod Security `restricted` (drops ALL caps) | `ping`/`tcpdump`/`arping` won't work; `curl`, `dig`, `tcping`, `nc` still do |
| Kubernetes, firewall/interface changes | `securityContext.capabilities.add: ["NET_ADMIN"]` |

## Included tools

| Category | Tools |
|---|---|
| ICMP / path | `ping`, `ping6`, `tracepath`, `arping`, `fping`, `traceroute` (incl. `-T` TCP / `-U` UDP), `mtr` |
| TCP latency | `tcping`, `paping` (bundled Go binary, see below) |
| DNS | `dig`, `nslookup`, `host`, `drill`, `whois` |
| HTTP / TLS | `curl`, `wget`, `openssl s_client`, `ssh` |
| TCP/UDP | `nc` (OpenBSD), `ncat`, `nmap` (+ NSE scripts), `socat`, `telnet` |
| Capture | `tcpdump`, `ngrep`, `iftop`, `nethogs` |
| Throughput | `iperf3` |
| Interfaces / routing / firewall | `ip`, `ss`, `bridge`, `ifconfig`, `netstat`, `route`, `ethtool`, `iptables`, `ip6tables`, `nft`, `conntrack`, `arp-scan` |
| Misc | `bash`, `jq`, `less`, `nano` |

### tcping / paping

`paping` and `tcping` aren't packaged for Alpine, so the image ships a small
static Go implementation (`tcping/main.go`) that answers to both names:

```bash
tcping [-c count] [-t timeout_s] [-i interval_s] [-4|-6] host [port]   # default port 80
tcping host:port
paping host -p port [-c count] [-t timeout_ms]                        # paping-compatible syntax
```

Runs until Ctrl-C when `-c` is omitted, then prints min/avg/max RTT and failure rate.
Exit code is 1 if no probe succeeded, which makes it usable in scripts.

## Self-test

The image ships its own smoke test (`test.sh`, installed as `netutils-test`). It checks that every
tool is present and runs live ping / tcping / paping / DNS / HTTPS / traceroute / nmap probes:

```bash
docker run --rm netutils netutils-test
```

## License

Apache License 2.0 — see [LICENSE](LICENSE).
