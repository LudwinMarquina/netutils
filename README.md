# netutils

Utils for networking engineers: a lightweight (Alpine-based) container packed with network troubleshooting tools.

[![docker](https://github.com/LudwinMarquina/netutils/actions/workflows/docker.yml/badge.svg)](https://github.com/LudwinMarquina/netutils/actions/workflows/docker.yml)
[![Docker Hub](https://img.shields.io/docker/v/ludwin20/netutils?sort=semver&label=docker%20hub)](https://hub.docker.com/r/ludwin20/netutils)

Multi-arch image (`linux/amd64`, `linux/arm64`) on Docker Hub: [`ludwin20/netutils`](https://hub.docker.com/r/ludwin20/netutils)

```bash
docker pull ludwin20/netutils
```

## Build locally

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
kubectl run netutils --rm -it --image=ludwin20/netutils -- bash
kubectl debug -it <pod> --image=ludwin20/netutils --target=<container>
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

## Image tags & releases

| Tag | Meaning | Produced by |
|---|---|---|
| `1.2.3` | exact release (immutable) | git tag `v1.2.3` |
| `1.2`, `1` | latest patch / minor of that line | git tag `v1.2.3` |
| `latest` | newest build from `main` or newest release | push to `main` / git tag |
| `sha-<7 chars>` | the exact git commit the image was built from | push to `main` |

Versioning follows [SemVer](https://semver.org): **major** = breaking change (tool removed, entrypoint/
shell changed), **minor** = tools added or base image bumped, **patch** = fixes. Pin a version
(`ludwin20/netutils:1`) in anything automated; use `latest` for ad-hoc debugging.

### CI/CD

[`.github/workflows/docker.yml`](.github/workflows/docker.yml) runs when an image file changes
(`Dockerfile`, `.dockerignore`, `tcping/**`, `test.sh`, the workflow itself), on `v*.*.*` tags,
or manually from the Actions tab:

1. builds an amd64 image and runs `netutils-test` (ICMP checks skipped; GitHub runners block ICMP);
2. on `main`/tags only: builds `linux/amd64` + `linux/arm64` and pushes to Docker Hub with SBOM + provenance.

Pull requests are built and tested but never pushed.

The Docker Hub page (overview and short description) comes from [`DOCKERHUB.md`](DOCKERHUB.md).
[`dockerhub-description.yml`](.github/workflows/dockerhub-description.yml) syncs it whenever that
file changes on `main`. This workflow never rebuilds the image.

**Docker Hub tag immutability.** Lock only the tags that are pushed exactly once, and leave the
moving tags (`latest`, `X.Y`, `X`) mutable. In Docker Hub → repository → *Settings* → *Immutable
tags*, use these rules instead of `.*`:

```
^[0-9]+\.[0-9]+\.[0-9]+$
^sha-[0-9a-f]+$
```

Because of this, re-running a pipeline for a commit or version that was already published fails
by design. Publish a new commit or a new version instead.

Required repository secrets: `DOCKERHUB_USERNAME` (`ludwin20`) and `DOCKERHUB_TOKEN`
(a Docker Hub [access token](https://app.docker.com/settings/personal-access-tokens) with *Read & Write* scope).

To cut a release:

```bash
git checkout main && git pull
git tag -a v1.1.0 -m "v1.1.0" && git push origin v1.1.0
```

## Security

| Layer | What it does |
|---|---|
| **Pinned bases** | `alpine` and `golang` are pinned by tag + multi-arch digest in the `Dockerfile`, so builds are reproducible. |
| **Dependabot** ([`dependabot.yml`](.github/dependabot.yml)) | Weekly PRs bump base-image tags/digests (e.g. Alpine 3.24 → 3.25, Go 1.27 → 1.28) and GitHub Actions. |
| **Scan gate** ([`docker.yml`](.github/workflows/docker.yml)) | Every build is scanned with Trivy before push. A **fixable** HIGH/CRITICAL fails the build, so nothing is published. The full report goes to the repo's *Security → Code scanning* tab. |
| **Weekly scan** ([`security-scan.yml`](.github/workflows/security-scan.yml)) | Mondays 06:00 UTC: scans the published `latest`, updates code scanning, and opens or updates an issue if anything fixable shows up. |

**Responding to a finding:**
1. Merge the Dependabot PR, or bump the base image by hand.
2. Merge it to `main`, then cut a patch release.

Release builds skip the layer cache, so even a release on the same commit installs the newest patched Alpine packages. Published `X.Y.Z` tags are immutable and are never rebuilt. Move to the newest release, or pin the moving `X` / `X.Y` tags.

Some findings have no fix released upstream yet. Grype reports them as NVD/CPE matches; Trivy doesn't flag them, since Alpine's database lists only fixed CVEs. The gate deliberately ignores them, and they're picked up once a fixed package exists.

To scan locally:

```bash
docker run --rm aquasec/trivy:0.76.1 image --scanners vuln --ignore-unfixed ludwin20/netutils:latest
```

## License

Apache License 2.0 — see [LICENSE](LICENSE).
