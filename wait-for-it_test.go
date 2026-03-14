package main

import (
	"bytes"
	"fmt"
	"net"
	"os"
	"testing"
	"time"
)

// startTCPServer starts a local TCP listener on a random port and returns the port.
// The listener is closed when the test ends.
func startTCPServer(t *testing.T) string {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start test TCP server: %v", err)
	}
	t.Cleanup(func() { ln.Close() })
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			conn.Close()
		}
	}()
	return fmt.Sprintf("%d", ln.Addr().(*net.TCPAddr).Port)
}

// captureStderr captures output written to os.Stderr during f().
func captureStderr(f func()) string {
	r, w, _ := os.Pipe()
	old := os.Stderr
	os.Stderr = w
	f()
	w.Close()
	os.Stderr = old
	var buf bytes.Buffer
	buf.ReadFrom(r)
	return buf.String()
}

// --- echoerr ---

func TestEchoerr_Quiet(t *testing.T) {
	quiet = true
	t.Cleanup(func() { quiet = false })
	out := captureStderr(func() {
		echoerr("should not appear")
	})
	if out != "" {
		t.Errorf("expected no output when quiet=true, got %q", out)
	}
}

func TestEchoerr_NotQuiet(t *testing.T) {
	quiet = false
	out := captureStderr(func() {
		echoerr("hello %s", "world")
	})
	if out != "hello world\n" {
		t.Errorf("expected %q, got %q", "hello world\n", out)
	}
}

// --- buildDialer ---

func TestBuildDialer_NoDNS(t *testing.T) {
	d := buildDialer("")
	if d == nil {
		t.Fatal("expected non-nil dialer")
	}
	if d.Timeout != time.Second {
		t.Errorf("expected timeout=1s, got %v", d.Timeout)
	}
	if d.Resolver != nil {
		t.Error("expected nil resolver when no DNS server provided")
	}
}

func TestBuildDialer_WithDNS_NoPort(t *testing.T) {
	d := buildDialer("8.8.8.8")
	if d == nil {
		t.Fatal("expected non-nil dialer")
	}
	if d.Resolver == nil {
		t.Error("expected custom resolver when DNS server is provided")
	}
	if d.Timeout != time.Second {
		t.Errorf("expected timeout=1s, got %v", d.Timeout)
	}
}

func TestBuildDialer_WithDNS_WithPort(t *testing.T) {
	d := buildDialer("8.8.8.8:53")
	if d == nil {
		t.Fatal("expected non-nil dialer")
	}
	if d.Resolver == nil {
		t.Error("expected custom resolver when DNS server with port is provided")
	}
}

// --- waitFor ---

func TestWaitFor_Success(t *testing.T) {
	quiet = true
	t.Cleanup(func() { quiet = false })

	port := startTCPServer(t)
	result := waitFor("127.0.0.1", port, "", 5)
	if result != 0 {
		t.Errorf("expected waitFor to return 0 when server is available, got %d", result)
	}
}

func TestWaitFor_Timeout(t *testing.T) {
	quiet = true
	t.Cleanup(func() { quiet = false })

	// Use a port that nothing is listening on.
	result := waitFor("127.0.0.1", "1", "", 2)
	if result != 1 {
		t.Errorf("expected waitFor to return 1 on timeout, got %d", result)
	}
}

func TestWaitFor_DNSFailure(t *testing.T) {
	quiet = false

	out := captureStderr(func() {
		result := waitFor("this.domain.does.not.exist.invalid", "80", "", 5)
		if result != 2 {
			t.Errorf("expected waitFor to return 2 on DNS failure, got %d", result)
		}
	})

	if !bytes.Contains([]byte(out), []byte("failed to resolve host")) {
		t.Errorf("expected DNS failure message in stderr, got %q", out)
	}
}

func TestWaitFor_OutputMessages(t *testing.T) {
	quiet = false
	port := startTCPServer(t)

	out := captureStderr(func() {
		waitFor("127.0.0.1", port, "", 5)
	})

	if !bytes.Contains([]byte(out), []byte("waiting 5 seconds")) {
		t.Errorf("expected waiting message in stderr, got %q", out)
	}
	if !bytes.Contains([]byte(out), []byte("is available after")) {
		t.Errorf("expected available message in stderr, got %q", out)
	}
}

func TestWaitFor_NoTimeoutMessage(t *testing.T) {
	quiet = false
	port := startTCPServer(t)

	out := captureStderr(func() {
		waitFor("127.0.0.1", port, "", 0)
	})

	if !bytes.Contains([]byte(out), []byte("without a timeout")) {
		t.Errorf("expected 'without a timeout' message in stderr, got %q", out)
	}
}

func TestWaitFor_WithDNSServerMessage(t *testing.T) {
	quiet = false
	port := startTCPServer(t)

	out := captureStderr(func() {
		waitFor("127.0.0.1", port, "8.8.8.8", 5)
	})

	if !bytes.Contains([]byte(out), []byte("using DNS server 8.8.8.8")) {
		t.Errorf("expected DNS server message in stderr, got %q", out)
	}
}
