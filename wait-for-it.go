package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
	"time"
)

const defaultTimeout = 15

var quiet bool

func echoerr(format string, args ...interface{}) {
	if !quiet {
		fmt.Fprintf(os.Stderr, format+"\n", args...)
	}
}

func buildDialer(dnsServer string) *net.Dialer {
	if dnsServer == "" {
		return &net.Dialer{Timeout: time.Second}
	}
	// Ensure the DNS server address includes a port
	if !strings.Contains(dnsServer, ":") {
		dnsServer = dnsServer + ":53"
	}
	resolver := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			return (&net.Dialer{Timeout: 5 * time.Second}).DialContext(ctx, "udp", dnsServer)
		},
	}
	return &net.Dialer{Timeout: time.Second, Resolver: resolver}
}

func waitFor(host, port, dnsServer string, timeout int) int {
	address := net.JoinHostPort(host, port)
	if timeout > 0 {
		echoerr("wait-for-it: waiting %d seconds for %s", timeout, address)
	} else {
		echoerr("wait-for-it: waiting for %s without a timeout", address)
	}
	if dnsServer != "" {
		echoerr("wait-for-it: using DNS server %s", dnsServer)
	}

	dialer := buildDialer(dnsServer)
	start := time.Now()

	for {
		conn, err := dialer.DialContext(context.Background(), "tcp", address)
		if err == nil {
			conn.Close()
			elapsed := int(time.Since(start).Seconds())
			echoerr("wait-for-it: %s is available after %d seconds", address, elapsed)
			return 0
		}

		var dnsErr *net.DNSError
		if errors.As(err, &dnsErr) {
			echoerr("wait-for-it: failed to resolve host %q: %s", host, dnsErr.Err)
			return 2
		}

		if timeout > 0 && time.Since(start) >= time.Duration(timeout)*time.Second {
			echoerr("wait-for-it: timeout occurred after waiting %d seconds for %s", timeout, address)
			return 1
		}

		time.Sleep(time.Second)
	}
}
