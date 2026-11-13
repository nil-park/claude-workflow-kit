package main

import (
	"bytes"
	"fmt"
	"net"
	"strconv"
	"strings"
	"testing"
)

func TestSendOversizedFrame(t *testing.T) {
	port := startFakeServer(t, func(conn net.Conn, req capturedRequest) {
		fmt.Fprintf(conn, "Content-Length: %d\r\n\r\n", maxFrameBytes+1)
	})

	var out, errBuf bytes.Buffer
	code := run([]string{"send", "--port", strconv.Itoa(port), "--from", "main", "--message", "hi"}, &out, &errBuf)
	if code != exitError {
		t.Fatalf("expected error exit, got %d", code)
	}
	wantLen := strconv.Itoa(maxFrameBytes + 1)
	if !strings.Contains(errBuf.String(), wantLen) {
		t.Fatalf("stderr = %q, want length %s", errBuf.String(), wantLen)
	}
}

func TestSendConnectionRefused(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := ln.Addr().(*net.TCPAddr).Port
	ln.Close()

	var out, errBuf bytes.Buffer
	code := run([]string{"send", "--port", strconv.Itoa(port), "--from", "main", "--message", "hi"}, &out, &errBuf)
	if code != exitError {
		t.Fatalf("expected error exit, got %d", code)
	}
	if !strings.Contains(errBuf.String(), "no Copilot UI server is listening on port") {
		t.Fatalf("stderr = %q, want refused message", errBuf.String())
	}
}
