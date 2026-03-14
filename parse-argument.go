package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

type config struct {
	host      string
	port      string
	dnsServer string
	timeout   int
	strict    bool
	cli       []string
}

func parseArgs() config {
	// Split at -- so everything after it is treated as the subcommand,
	// not consumed by the flag parser.
	var mainArgs, cli []string
	for i, a := range os.Args[1:] {
		if a == "--" {
			cli = os.Args[i+2:]
			break
		}
		mainArgs = append(mainArgs, a)
	}

	fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	fs.Usage = usage

	var host, port, dnsServer string
	var timeout int
	var strict bool

	// Register both short and long forms pointing to the same variable.
	fs.StringVar(&host, "h", "", "Host or IP under test")
	fs.StringVar(&host, "host", "", "Host or IP under test")
	fs.StringVar(&port, "p", "", "TCP port under test")
	fs.StringVar(&port, "port", "", "TCP port under test")
	fs.IntVar(&timeout, "t", defaultTimeout, "Timeout in seconds, zero for no timeout")
	fs.IntVar(&timeout, "timeout", defaultTimeout, "Timeout in seconds, zero for no timeout")
	fs.StringVar(&dnsServer, "d", "", "Custom DNS server (e.g. 8.8.8.8 or 8.8.8.8:53)")
	fs.StringVar(&dnsServer, "dns", "", "Custom DNS server (e.g. 8.8.8.8 or 8.8.8.8:53)")
	fs.BoolVar(&strict, "s", false, "Only execute subcommand if the test succeeds")
	fs.BoolVar(&strict, "strict", false, "Only execute subcommand if the test succeeds")
	fs.BoolVar(&quiet, "q", false, "Don't output any status messages")
	fs.BoolVar(&quiet, "quiet", false, "Don't output any status messages")

	if err := fs.Parse(mainArgs); err != nil {
		os.Exit(1)
	}

	// Remaining positional args: accept host:port shorthand.
	for _, arg := range fs.Args() {
		if strings.Contains(arg, ":") {
			parts := strings.SplitN(arg, ":", 2)
			if host == "" {
				host = parts[0]
			}
			if port == "" {
				port = parts[1]
			}
		} else {
			fmt.Fprintf(os.Stderr, "Unknown argument: %s\n", arg)
			usage()
		}
	}

	if host == "" || port == "" {
		fmt.Fprintf(os.Stderr, "Error: you need to provide a host and port to test.\n")
		usage()
	}

	return config{
		host:      host,
		port:      port,
		dnsServer: dnsServer,
		timeout:   timeout,
		strict:    strict,
		cli:       cli,
	}
}
