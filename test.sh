#!/usr/bin/env bash
# Smoke test: docker run --rm netutils netutils-test
fail=0
for t in ping ping6 tracepath arping fping traceroute mtr tcping paping dig nslookup host drill whois \
         curl wget openssl ssh nc ncat nmap socat telnet tcpdump ngrep iftop nethogs iperf3 \
         ip ss bridge ifconfig netstat route ethtool iptables ip6tables nft conntrack arp-scan jq bash; do
  command -v "$t" >/dev/null || { echo "MISSING: $t"; fail=1; }
done
[ $fail = 0 ] && echo "all binaries present"

run() { echo; echo "### $*"; "$@" || { echo "FAILED: $*"; fail=1; }; }
run ping -c 2 1.1.1.1
run tcping -c 3 -i 0.3 1.1.1.1 443
run tcping -c 2 -i 0.3 google.com:443
run paping 1.1.1.1 -p 443 -c 3
run dig +short cloudflare.com
run curl -sS -o /dev/null -w "HTTP %{http_code} in %{time_total}s\n" https://example.com
run traceroute -n -m 5 -q 1 1.1.1.1
run tcpdump --version
run nmap -Pn -p 443 1.1.1.1

echo; echo "### tcping refused port (expect failure exit 1)"
tcping -c 2 -i 0.3 -t 1 127.0.0.1 1; [ $? = 1 ] && echo "ok: exit 1" || { echo "FAILED: expected exit 1"; fail=1; }

echo; [ $fail = 0 ] && echo "=== ALL TESTS PASSED ===" || echo "=== SOME TESTS FAILED ==="
exit $fail
