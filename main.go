package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
)

func usage() {
	fmt.Fprintf(os.Stderr, `Usage:
    wait-for-it host:port [-s] [-t timeout] [-- command args]
    -h HOST | --host=HOST       Host or IP under test
    -p PORT | --port=PORT       TCP port under test
                                Alternatively, specify host and port as host:port
    -s | --strict               Only execute subcommand if the test succeeds
    -q | --quiet                Don't output any status messages
    -t TIMEOUT | --timeout=TIMEOUT
                                Timeout in seconds, zero for no timeout (default: 15)
    -d DNS | --dns=DNS          Custom DNS server to resolve the host (e.g. 8.8.8.8 or 8.8.8.8:53)
    -- COMMAND ARGS             Execute command with args after the test finishes
`)
	os.Exit(1)
}

func main() {
	cfg := parseArgs()

	// Forward SIGINT/SIGTERM for clean exit.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		os.Exit(130)
	}()

	success := waitFor(cfg.host, cfg.port, cfg.dnsServer, cfg.timeout)

	if len(cfg.cli) > 0 {
		if !success && cfg.strict {
			echoerr("wait-for-it: strict mode, refusing to execute subprocess")
			os.Exit(1)
		}
		cmd := exec.Command(cfg.cli[0], cfg.cli[1:]...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		if err := cmd.Run(); err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				os.Exit(exitErr.ExitCode())
			}
			os.Exit(1)
		}
		os.Exit(0)
	}

	if success {
		os.Exit(0)
	}
	os.Exit(1)
}
