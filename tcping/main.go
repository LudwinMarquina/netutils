// tcping / paping: measure TCP connect latency to host:port.
//
// Invoked as "tcping" (tcping-style syntax):
//
//	tcping [-c count] [-t timeout] [-i interval] [-4|-6] host [port]
//	tcping host:port
//
// Invoked as "paping" (paping-compatible syntax):
//
//	paping host -p port [-c count] [-t timeout_ms]
package main

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
)

type opts struct {
	host     string
	port     string
	count    int // 0 = infinite
	timeout  time.Duration
	interval time.Duration
	network  string
}

func usage(paping bool) {
	if paping {
		fmt.Fprintln(os.Stderr, "usage: paping <host> -p <port> [-c count] [-t timeout_ms]")
	} else {
		fmt.Fprintln(os.Stderr, `usage: tcping [options] <host> [port]   (or host:port, default port 80)
  -c N    number of probes (default: infinite)
  -t S    timeout per probe in seconds (default 2, decimals ok)
  -i S    interval between probes in seconds (default 1, decimals ok)
  -4/-6   force IPv4 / IPv6`)
	}
	os.Exit(2)
}

func seconds(s string) (time.Duration, error) {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil || f < 0 {
		return 0, fmt.Errorf("invalid duration %q", s)
	}
	return time.Duration(f * float64(time.Second)), nil
}

func parse(args []string, paping bool) opts {
	o := opts{count: 0, timeout: 2 * time.Second, interval: time.Second, network: "tcp"}
	if paping {
		o.timeout = time.Second
	}
	var pos []string
	need := func(i int) string {
		if i+1 >= len(args) {
			usage(paping)
		}
		return args[i+1]
	}
	for i := 0; i < len(args); i++ {
		a := args[i]
		var err error
		switch a {
		case "-h", "--help":
			usage(paping)
		case "-c", "--count":
			o.count, err = strconv.Atoi(need(i))
			i++
		case "-p", "--port":
			o.port = need(i)
			i++
		case "-t", "--timeout":
			v := need(i)
			i++
			if paping { // paping uses milliseconds
				var ms int
				ms, err = strconv.Atoi(v)
				o.timeout = time.Duration(ms) * time.Millisecond
			} else {
				o.timeout, err = seconds(v)
			}
		case "-i", "--interval":
			o.interval, err = seconds(need(i))
			i++
		case "-4":
			o.network = "tcp4"
		case "-6":
			o.network = "tcp6"
		default:
			if strings.HasPrefix(a, "-") && len(a) > 1 {
				fmt.Fprintf(os.Stderr, "unknown option %s\n", a)
				usage(paping)
			}
			pos = append(pos, a)
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "invalid value for %s: %v\n", a, err)
			os.Exit(2)
		}
	}
	switch len(pos) {
	case 1:
		o.host = pos[0]
		if h, p, err := net.SplitHostPort(pos[0]); err == nil {
			o.host, o.port = h, p
		}
	case 2:
		o.host, o.port = pos[0], pos[1]
	default:
		usage(paping)
	}
	if o.port == "" {
		if paping {
			usage(paping)
		}
		o.port = "80"
	}
	return o
}

func main() {
	paping := strings.HasPrefix(filepath.Base(os.Args[0]), "paping")
	o := parse(os.Args[1:], paping)

	ips, err := net.DefaultResolver.LookupIP(context.Background(), map[string]string{"tcp": "ip", "tcp4": "ip4", "tcp6": "ip6"}[o.network], o.host)
	if err != nil || len(ips) == 0 {
		fmt.Fprintf(os.Stderr, "cannot resolve %s: %v\n", o.host, err)
		os.Exit(2)
	}
	addr := net.JoinHostPort(ips[0].String(), o.port)
	fmt.Printf("Connecting to %s (%s) on TCP %s:\n\n", o.host, ips[0], o.port)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	var sent, ok int
	var min, max, sum float64 = math.MaxFloat64, 0, 0
	for o.count == 0 || sent < o.count {
		if sent > 0 {
			select {
			case <-stop:
				goto done
			case <-time.After(o.interval):
			}
		}
		sent++
		start := time.Now()
		conn, err := net.DialTimeout(o.network, addr, o.timeout)
		ms := float64(time.Since(start).Microseconds()) / 1000
		if err != nil {
			reason := "Connection refused"
			var ne net.Error
			if errors.As(err, &ne) && ne.Timeout() {
				reason = "Connection timed out"
			} else if !errors.Is(err, syscall.ECONNREFUSED) {
				reason = err.Error()
			}
			fmt.Printf("%s:%s seq=%d %s\n", ips[0], o.port, sent, reason)
		} else {
			conn.Close()
			ok++
			sum += ms
			min = math.Min(min, ms)
			max = math.Max(max, ms)
			fmt.Printf("Connected to %s:%s seq=%d time=%.2fms\n", ips[0], o.port, sent, ms)
		}
		select {
		case <-stop:
			goto done
		default:
		}
	}
done:
	fmt.Printf("\n--- %s tcping statistics ---\n", o.host)
	fmt.Printf("%d probes sent, %d successful, %d failed (%.1f%% fail)\n",
		sent, ok, sent-ok, float64(sent-ok)/float64(max1(sent))*100)
	if ok > 0 {
		fmt.Printf("rtt min/avg/max = %.2f/%.2f/%.2f ms\n", min, sum/float64(ok), max)
	}
	if ok == 0 {
		os.Exit(1)
	}
}

func max1(n int) int {
	if n < 1 {
		return 1
	}
	return n
}
