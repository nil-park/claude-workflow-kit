package main

import (
	"bytes"
	"fmt"
	"net"
	"strconv"
	"strings"
	"testing"
)

func TestSendSuccessAndRequestOrder(t *testing.T) {
	for _, protoVer := range []int{3, 4} {
		t.Run(fmt.Sprintf("v%d", protoVer), func(t *testing.T) {
			seenCh := make(chan capturedRequest, 4)
			port := startFakeServer(t, func(conn net.Conn, req capturedRequest) {
				seenCh <- req
				switch req.Method {
				case "ping":
					// Notification and server-initiated request with colliding id must both be skipped.
					writeFakeFrame(conn, map[string]any{"jsonrpc": "2.0", "method": "session.event", "params": map[string]any{"type": "idle"}})
					writeFakeFrame(conn, map[string]any{"jsonrpc": "2.0", "id": req.ID, "method": "ui.confirm", "params": map[string]any{}})
					writeResult(conn, req.ID, map[string]any{"protocolVersion": protoVer})
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
			wantDisplay := "[main-1] " + handshakeBody("main-1", port) + "\n\nhello"
			if gotDisplay, _ := seen[2].Params["displayPrompt"].(string); gotDisplay != wantDisplay {
				t.Fatalf("displayPrompt:\ngot:  %q\nwant: %q", gotDisplay, wantDisplay)
			}
		})
	}
}

func TestSendWrongProtocolVersion(t *testing.T) {
	for _, protoVer := range []int{2, 5} {
		t.Run(fmt.Sprintf("v%d", protoVer), func(t *testing.T) {
			seenCh := make(chan capturedRequest, 4)
			port := startFakeServer(t, func(conn net.Conn, req capturedRequest) {
				seenCh <- req
				if req.Method == "ping" {
					writeResult(conn, req.ID, map[string]any{"protocolVersion": protoVer})
				}
			})

			var out, errBuf bytes.Buffer
			code := run([]string{"send", "--port", strconv.Itoa(port), "--from", "main", "--message", "hi"}, &out, &errBuf)
			if code != exitError {
				t.Fatalf("expected error exit, got %d", code)
			}
			wantErr := fmt.Sprintf("protocolVersion %d", protoVer)
			if !strings.Contains(errBuf.String(), wantErr) {
				t.Fatalf("error must name version received (%q), got: %q", wantErr, errBuf.String())
			}

			close(seenCh)
			for req := range seenCh {
				if req.Method == "session.getForeground" {
					t.Fatalf("session.getForeground must not be called on rejected protocol version %d", protoVer)
				}
			}
		})
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
