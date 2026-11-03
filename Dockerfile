# syntax=docker/dockerfile:1
ARG ALPINE_VERSION=3.22
ARG GO_VERSION=1.24

# ---- build tcping/paping (tiny static Go binary) ----
FROM golang:${GO_VERSION}-alpine AS tcping
WORKDIR /src
COPY tcping/ .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/tcping .

# ---- runtime ----
FROM alpine:${ALPINE_VERSION}

# Extra packages can be added at build time, e.g. --build-arg EXTRA_PKGS="tshark"
ARG EXTRA_PKGS=""

RUN apk add --no-cache \
      # shell & utils
      bash bash-completion ca-certificates jq less nano tini tzdata \
      # ICMP / path
      iputils fping traceroute mtr \
      # DNS
      bind-tools drill whois \
      # HTTP / TLS
      curl wget openssl openssh-client \
      # TCP/UDP swiss army knives
      netcat-openbsd nmap nmap-ncat nmap-scripts socat busybox-extras \
      # capture & inspection
      tcpdump ngrep iftop nethogs \
      # throughput
      iperf3 \
      # interfaces, routing, firewall
      iproute2 net-tools ethtool iptables ip6tables nftables conntrack-tools arp-scan \
      ${EXTRA_PKGS} \
 && rm -rf /var/cache/apk/*

COPY --from=tcping /out/tcping /usr/local/bin/tcping
COPY --chmod=755 test.sh /usr/local/bin/netutils-test
RUN ln -s tcping /usr/local/bin/paping \
 && echo 'PS1="\[\e[1;36m\][netutils \h]\[\e[0m\] \w # "' >> /root/.bashrc

LABEL org.opencontainers.image.title="netutils" \
      org.opencontainers.image.description="Lightweight Alpine image with network troubleshooting tools"

WORKDIR /root
ENTRYPOINT ["/sbin/tini", "--"]
CMD ["bash"]
