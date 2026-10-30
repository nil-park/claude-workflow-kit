package main

import (
	"bytes"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// listenInbox accepts one connection on a Unix socket and returns everything
// the client wrote before closing it.
func listenInbox(t *testing.T, socket string) <-chan []byte {
	t.Helper()
	ln, err := net.Listen("unix", socket)
	if err != nil {
		t.Skipf("Unix sockets are unavailable: %v", err)
	}
	t.Cleanup(func() { ln.Close() })
	received := make(chan []byte, 1)
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			received <- nil
			return
		}
		defer conn.Close()
		raw, _ := io.ReadAll(conn)
		received <- raw
	}()
	return received
}

func waitReceived(t *testing.T, received <-chan []byte) []byte {
	t.Helper()
	select {
	case raw := <-received:
		return raw
	case <-time.After(5 * time.Second):
		t.Fatal("the inbox received nothing")
		return nil
	}
}

func TestSendDeliversToUnixSocket(t *testing.T) {
	dir := t.TempDir()
	socket := filepath.Join(dir, "7.sock")
	writeJSON(t, filepath.Join(dir, "7.json"), registration(7, "worker", "/w", socket))
	const token = "0123456789abcdef0123456789abcdef"
	writeKey(t, dir, session{PID: 7, SocketPath: socket}, token)
	stubSessions(t, dir, 7)
	received := listenInbox(t, socket)

	var stdout bytes.Buffer
	if code := run([]string{"send", "--to", "worker", "--from", "agy-peer", "--message", "hello"}, &stdout, io.Discard); code != exitOK {
		t.Fatalf("run() = %d, want %d", code, exitOK)
	}
	lines := decodeLines(t, waitReceived(t, received))
	if len(lines) != 2 || lines[0]["token"] != token {
		t.Fatalf("received %v, want the auth line then the message line", lines)
	}
	if id := strings.TrimSpace(stdout.String()); lines[1]["msg_id"] != id {
		t.Errorf("stdout msg_id %q differs from the delivered %v", id, lines[1]["msg_id"])
	}
}

func clearAgyEnv(t *testing.T) {
	t.Helper()
	t.Setenv("ANTIGRAVITY_CONVERSATION_ID", "")
	t.Setenv("ANTIGRAVITY_LS_ADDRESS", "")
	t.Setenv("ANTIGRAVITY_CSRF_TOKEN", "")
}

func TestSendWithoutKeyFileOmitsAuthLine(t *testing.T) {
	clearAgyEnv(t)
	dir := t.TempDir()
	socket := filepath.Join(dir, "8.sock")
	writeJSON(t, filepath.Join(dir, "8.json"), registration(8, "worker", "/w", socket))
	stubSessions(t, dir, 8)
	received := listenInbox(t, socket)
	body := filepath.Join(dir, "body.md")
	if err := os.WriteFile(body, []byte("파일 본문"), 0o600); err != nil {
		t.Fatal(err)
	}

	if code := run([]string{"send", "--to", "8", "--file", body}, io.Discard, io.Discard); code != exitOK {
		t.Fatalf("run() = %d, want %d", code, exitOK)
	}
	lines := decodeLines(t, waitReceived(t, received))
	if len(lines) != 1 || lines[0]["from"] != defaultFrom {
		t.Errorf("received %v, want one message line from %q", lines, defaultFrom)
	}
}

func TestSendFromAgyEnv(t *testing.T) {
	t.Setenv("ANTIGRAVITY_CONVERSATION_ID", "AbC-123")
	t.Setenv("ANTIGRAVITY_LS_ADDRESS", "")
	t.Setenv("ANTIGRAVITY_CSRF_TOKEN", "")

	for _, tc := range []struct {
		name string
		args []string
		want string
	}{
		{"default_from_env", []string{"send", "--to", "8", "--message", "hi"}, "agy-abc-123"},
		{"explicit_from_wins", []string{"send", "--to", "8", "--from", "explicit-peer", "--message", "hi"}, "explicit-peer"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			socket := filepath.Join(dir, "8.sock")
			writeJSON(t, filepath.Join(dir, "8.json"), registration(8, "worker", "/w", socket))
			stubSessions(t, dir, 8)
			received := listenInbox(t, socket)
			if code := run(tc.args, io.Discard, io.Discard); code != exitOK {
				t.Fatalf("run(%q) = %d, want %d", tc.args, code, exitOK)
			}
			lines := decodeLines(t, waitReceived(t, received))
			if len(lines) != 1 || lines[0]["from"] != tc.want {
				t.Errorf("run(%q) delivered %v, want from %q", tc.args, lines, tc.want)
			}
		})
	}
}

// The worker session accepts every connection, so a case that slipped past
// its check would be delivered and exit 0.
func TestRunRejectsBadArguments(t *testing.T) {
	clearAgyEnv(t)
	dir := t.TempDir()
	socket := filepath.Join(dir, "9.sock")
	writeJSON(t, filepath.Join(dir, "9.json"), registration(9, "worker", "/w", socket))
	stubSessions(t, dir, 9)
	ln, err := net.Listen("unix", socket)
	if err != nil {
		t.Skipf("Unix sockets are unavailable: %v", err)
	}
	t.Cleanup(func() { ln.Close() })
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			io.Copy(io.Discard, conn)
			conn.Close()
		}
	}()
	invalidUTF8 := filepath.Join(dir, "bad.md")
	if err := os.WriteFile(invalidUTF8, []byte{0xff, 0xfe}, 0o600); err != nil {
		t.Fatal(err)
	}
	cases := [][]string{
		{},
		{"unknown"},
		{"list", "extra"},
		{"send", "--message", "hi"},
		{"send", "--to", "worker"},
		{"send", "--to", "worker", "--message", "hi", "--file", invalidUTF8},
		{"send", "--to", "worker", "--message", "hi", "--handshake"},
		{"send", "--to", "worker", "--file", invalidUTF8, "--handshake"},
		{"send", "--to", "worker", "--from", "agy-peer", "--handshake"},
		{"send", "--to", "worker", "--handshake=false"},
		{"send", "--to", "worker", "--message", ""},
		{"send", "--to", "worker", "--message", " \n"},
		{"send", "--to", "worker", "--file", invalidUTF8},
		{"send", "--to", "worker", "--from", "", "--message", "hi"},
		{"send", "--to", "worker", "--from", "Agy", "--message", "hi"},
		{"send", "--to", "worker", "--from", "a b", "--message", "hi"},
		{"send", "--to", "nobody", "--message", "hi"},
	}
	for _, args := range cases {
		if code := run(args, io.Discard, io.Discard); code != exitError {
			t.Errorf("run(%q) = %d, want %d", args, code, exitError)
		}
	}
}
