package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"testing"
)

type capturedRequest struct {
	ID     int
	Method string
	Params map[string]any
}

func startFakeServer(t *testing.T, handler func(conn net.Conn, req capturedRequest)) int {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { ln.Close() })

	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		r := bufio.NewReader(conn)
		for {
			payload, err := readFrame(r)
			if err != nil {
				return
			}
			var raw struct {
				ID     int            `json:"id"`
				Method string         `json:"method"`
				Params map[string]any `json:"params"`
			}
			if err := json.Unmarshal(payload, &raw); err != nil {
				return
			}
			handler(conn, capturedRequest{ID: raw.ID, Method: raw.Method, Params: raw.Params})
		}
	}()

	return ln.Addr().(*net.TCPAddr).Port
}

func writeFakeFrame(w io.Writer, msg any) {
	b, _ := json.Marshal(msg)
	fmt.Fprintf(w, "Content-Length: %d\r\n\r\n%s", len(b), b)
}

func writeResult(w io.Writer, id int, result any) {
	writeFakeFrame(w, map[string]any{"jsonrpc": "2.0", "id": id, "result": result})
}

func writeError(w io.Writer, id, code int, message string) {
	writeFakeFrame(w, map[string]any{
		"jsonrpc": "2.0",
		"id":      id,
		"error":   map[string]any{"code": code, "message": message},
	})
}

func TestSendSuccessAndRequestOrder(t *testing.T) {
	seenCh := make(chan capturedRequest, 4)
	port := startFakeServer(t, func(conn net.Conn, req capturedRequest) {
		seenCh <- req
		switch req.Method {
		case "ping":
			// Notification and server-initiated request with colliding id must both be skipped.
			writeFakeFrame(conn, map[string]any{"jsonrpc": "2.0", "method": "session.event", "params": map[string]any{"type": "idle"}})
			writeFakeFrame(conn, map[string]any{"jsonrpc": "2.0", "id": req.ID, "method": "ui.confirm", "params": map[string]any{}})
			writeResult(conn, req.ID, map[string]any{"protocolVersion": 3})
		case "session.getForeground":
			writeResult(conn, req.ID, map[string]any{"sessionId": "sess-42"})
		case "session.send":
			writeResult(conn, req.ID, map[string]any{"messageId": "msg-99"})
		}
	})

	var out, errBuf bytes.Buffer
	code := run([]string{"send", "--port", strconv.Itoa(port), "--from", "main-1", "--handshake", "--message", "hello"}, &out, &errBuf)
	if code != exitOK {
		t.Fatalf("run failed (%d): %s", code, errBuf.String())
	}
	if strings.TrimSpace(out.String()) != "msg-99" {
		t.Fatalf("stdout = %q, want msg-99", out.String())
	}

	close(seenCh)
	var seen []capturedRequest
	for req := range seenCh {
		seen = append(seen, req)
	}
	if len(seen) != 3 {
		t.Fatalf("saw %d requests, want 3", len(seen))
	}
	if seen[0].Method != "ping" || seen[1].Method != "session.getForeground" || seen[2].Method != "session.send" {
		t.Fatalf("unexpected request order: %v, %v, %v", seen[0].Method, seen[1].Method, seen[2].Method)
	}
	if _, hasSource := seen[2].Params["source"]; hasSource {
		t.Fatalf("session.send must not include source: %v", seen[2].Params)
	}
	if seen[2].Params["sessionId"] != "sess-42" {
		t.Fatalf("sessionId = %v, want sess-42", seen[2].Params["sessionId"])
	}
	prompt, _ := seen[2].Params["prompt"].(string)
	if !strings.HasPrefix(prompt, "<cross-session-message from=\"main-1\">\n") || !strings.Contains(prompt, "hello") {
		t.Fatalf("unexpected prompt: %q", prompt)
	}
}

func TestSendWrongProtocolVersion(t *testing.T) {
	port := startFakeServer(t, func(conn net.Conn, req capturedRequest) {
		if req.Method == "ping" {
			writeResult(conn, req.ID, map[string]any{"protocolVersion": 2})
		}
	})

	var out, errBuf bytes.Buffer
	code := run([]string{"send", "--port", strconv.Itoa(port), "--from", "main", "--message", "hi"}, &out, &errBuf)
	if code != exitError {
		t.Fatalf("expected error exit, got %d", code)
	}
	if !strings.Contains(errBuf.String(), "protocolVersion 2") {
		t.Fatalf("error must name version received, got: %q", errBuf.String())
	}
}

func TestSendEmptySessionID(t *testing.T) {
	port := startFakeServer(t, func(conn net.Conn, req capturedRequest) {
		switch req.Method {
		case "ping":
			writeResult(conn, req.ID, map[string]any{"protocolVersion": 3})
		case "session.getForeground":
			writeResult(conn, req.ID, map[string]any{"sessionId": ""})
		}
	})

	var out, errBuf bytes.Buffer
	code := run([]string{"send", "--port", strconv.Itoa(port), "--from", "main", "--message", "hi"}, &out, &errBuf)
	if code != exitError {
		t.Fatalf("expected error exit, got %d", code)
	}
	if !strings.Contains(errBuf.String(), "no foreground session") {
		t.Fatalf("stderr = %q, want no foreground session", errBuf.String())
	}
}

func TestSendJSONRPCError(t *testing.T) {
	port := startFakeServer(t, func(conn net.Conn, req capturedRequest) {
		switch req.Method {
		case "ping":
			writeResult(conn, req.ID, map[string]any{"protocolVersion": 3})
		case "session.getForeground":
			writeResult(conn, req.ID, map[string]any{"sessionId": "sess-1"})
		case "session.send":
			writeError(conn, req.ID, -32000, "session busy")
		}
	})

	var out, errBuf bytes.Buffer
	code := run([]string{"send", "--port", strconv.Itoa(port), "--from", "main", "--message", "hi"}, &out, &errBuf)
	if code != exitError {
		t.Fatalf("expected error exit, got %d", code)
	}
	if !strings.Contains(errBuf.String(), "session busy") {
		t.Fatalf("stderr = %q, want session busy", errBuf.String())
	}
}
