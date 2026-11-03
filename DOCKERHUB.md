# netutils

**Utils for networking engineers.** A lightweight Alpine-based container with the network
troubleshooting tools you usually need: ping, traceroute, mtr, tcpdump, tcping, paping, curl,
wget, dig, nmap, iperf3 and more. One `docker run` or `kubectl run` and you're debugging.

- **Multi-arch:** `linux/amd64` and `linux/arm64`. Docker and Kubernetes pull the right one automatically.
- **Small:** about 27 MB compressed.
- **No special privileges for everyday use:** `ping`, `tcpdump`, `traceroute`, `arping` and `nmap` work with Docker's default capabilities.
- **Built and tested by CI:** every published image passed the built-in self-test first, and ships with SBOM and provenance attestations.

Source, issues and Dockerfile: **https://github.com/LudwinMarquina/netutils**

---

## Quick start

```bash
# interactive shell
docker run --rm -it ludwin20/netutils

# one-shot commands
docker run --rm ludwin20/netutils tcping -c 4 google.com 443
docker run --rm ludwin20/netutils paping 1.1.1.1 -p 443 -c 4
docker run --rm ludwin20/netutils mtr -rwc 10 8.8.8.8
docker run --rm ludwin20/netutils curl -sI https://example.com

# use the host's network stack (see and capture host interfaces)
docker run --rm -it --net host ludwin20/netutils

# join another container's network namespace
docker run --rm -it --net container:<container-name> ludwin20/netutils
```

### Kubernetes

```bash
# throwaway debug pod
kubectl run netutils --rm -it --image=ludwin20/netutils -- bash

# attach to a running pod's network namespace (ephemeral container)
kubectl debug -it <pod> --image=ludwin20/netutils --target=<container>

# debug a node (host network)
kubectl debug node/<node> -it --image=ludwin20/netutils
```

A long-running debug pod you can `kubectl exec` into:

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: netutils
spec:
  containers:
    - name: netutils
      image: ludwin20/netutils:1
      command: ["sleep", "infinity"]
```

```bash
kubectl exec -it netutils -- bash
```

---

## Tags

| Tag | Meaning | Mutable |
|---|---|---|
| `latest` | newest build from `main` or the newest release | yes |
| `1.0.0` (`X.Y.Z`) | an exact release | **no** |
| `1.0` (`X.Y`), `1` (`X`) | newest patch / minor in that line | yes |
| `sha-<commit>` | the exact git commit the image was built from | **no** |

Versions follow [SemVer](https://semver.org):
- **Major:** a breaking change, such as a tool removed or the entrypoint changed.
- **Minor:** tools added, or the base image upgraded.
- **Patch:** fixes.

Pin `ludwin20/netutils:1` or an exact `X.Y.Z` in anything automated. Use `latest` for ad-hoc debugging.

---

## Included tools

| Category | Tools |
|---|---|
| ICMP / path | `ping`, `ping6`, `tracepath`, `arping`, `fping`, `traceroute` (incl. `-T` TCP / `-U` UDP), `mtr` |
| TCP latency | `tcping`, `paping` |
| DNS | `dig`, `nslookup`, `host`, `drill`, `whois` |
| HTTP / TLS | `curl`, `wget`, `openssl s_client`, `ssh` |
| TCP / UDP | `nc` (OpenBSD), `ncat`, `nmap` (+ NSE scripts), `socat`, `telnet` |
| Capture | `tcpdump`, `ngrep`, `iftop`, `nethogs` |
| Throughput | `iperf3` |
| Interfaces / routing / firewall | `ip`, `ss`, `bridge`, `ifconfig`, `netstat`, `route`, `ethtool`, `iptables`, `ip6tables`, `nft`, `conntrack`, `arp-scan` |
| Misc | `bash`, `jq`, `less`, `nano` |

### tcping / paping

Measures **TCP connect latency**. Unlike ICMP ping, it works through firewalls that drop ICMP
and tells you whether a specific port is reachable. A single static binary answers to both names:

```bash
tcping [-c count] [-t timeout_s] [-i interval_s] [-4|-6] host [port]   # default port 80
tcping host:port
paping host -p port [-c count] [-t timeout_ms]                        # paping-compatible syntax
```

```text
$ tcping -c 3 1.1.1.1 443
Connecting to 1.1.1.1 (1.1.1.1) on TCP 443:

Connected to 1.1.1.1:443 seq=1 time=9.90ms
Connected to 1.1.1.1:443 seq=2 time=8.37ms
Connected to 1.1.1.1:443 seq=3 time=8.33ms

--- 1.1.1.1 tcping statistics ---
3 probes sent, 3 successful, 0 failed (0.0% fail)
rtt min/avg/max = 8.33/8.87/9.90 ms
```

It runs until Ctrl-C when `-c` is omitted. The exit code is `1` if no probe succeeded, so you can use it in scripts and health checks.

---

## Capabilities

No extra flags are needed for normal troubleshooting: Docker's default capabilities include `NET_RAW`.
Add `NET_ADMIN` only to **read or change** network configuration, such as firewall rules, conntrack, or interface/route changes.

| Situation | What to add |
|---|---|
| Docker, normal use | nothing |
| Docker, firewall / interface changes | `--cap-add NET_ADMIN` |
| Kubernetes, default security | nothing |
| Kubernetes with Pod Security `restricted` (drops ALL) | `ping`, `tcpdump` and `arping` won't work; `curl`, `dig`, `tcping` and `nc` still do |
| Kubernetes, firewall / interface changes | `securityContext.capabilities.add: ["NET_ADMIN"]` |

---

## Self-test

```bash
docker run --rm ludwin20/netutils netutils-test
```

It checks that every tool is present, then runs live probes: ping, tcping, paping, dig, HTTPS, traceroute and nmap.
Set `-e SKIP_ICMP=1` where outbound ICMP is blocked.

---

## Image details

- **Base:** `alpine:3.24` (digest-pinned), with `tini` as PID 1. `tcping` is built with Go 1.27. The default command is `bash`, running as `root` so raw sockets and capture work.
- **Labels:** OCI `source`, `revision`, `version`, `created` and `licenses` labels, plus SBOM and provenance attestations.
- **Bigger build:** you can build a larger variant with extra Alpine packages, for example Wireshark's CLI:
  `docker build --build-arg EXTRA_PKGS="tshark" -t netutils https://github.com/LudwinMarquina/netutils.git`

## Security

- **Pinned bases:** the base images are pinned by digest, and Dependabot keeps them current.
- **Scan gate:** every build is scanned with Trivy before it's pushed. A fixable HIGH or CRITICAL vulnerability blocks the release.
- **Weekly scan:** the published `latest` is re-scanned every week, and findings are tracked in [GitHub issues](https://github.com/LudwinMarquina/netutils/issues).
- **`X.Y.Z` tags never change:** stay on the moving `1` / `1.1` tags, or on `latest`, to get security fixes automatically.

## License

[Apache License 2.0](https://github.com/LudwinMarquina/netutils/blob/main/LICENSE)
