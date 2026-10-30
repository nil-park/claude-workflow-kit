package main

import (
	"bytes"
	"io"
	"path/filepath"
	"strings"
	"testing"
)

func TestSendHandshake(t *testing.T) {
	t.Setenv("ANTIGRAVITY_CONVERSATION_ID", "AbC-123")
	t.Setenv("ANTIGRAVITY_LS_ADDRESS", "127.0.0.1:42100")
	t.Setenv("ANTIGRAVITY_CSRF_TOKEN", "ToKeN-9")

	dir := t.TempDir()
	socket := filepath.Join(dir, "8.sock")
	writeJSON(t, filepath.Join(dir, "8.json"), registration(8, "worker", "/w", socket))
	stubSessions(t, dir, 8)
	received := listenInbox(t, socket)

	if code := run([]string{"send", "--to", "8", "--handshake"}, io.Discard, io.Discard); code != exitOK {
		t.Fatalf("run(--handshake) = %d, want %d", code, exitOK)
	}
	lines := decodeLines(t, waitReceived(t, received))
	if len(lines) != 1 || lines[0]["from"] != "agy-abc-123" {
		t.Fatalf("delivered %v, want from agy-abc-123", lines)
	}
	content := lines[0]["message"].(map[string]any)["content"].(string)
	for _, sub := range []string{
		"agy agentapi send-message AbC-123 <message>",
		"ANTIGRAVITY_CONVERSATION_ID=AbC-123",
		"ANTIGRAVITY_LS_ADDRESS=127.0.0.1:42100",
		"ANTIGRAVITY_CSRF_TOKEN=ToKeN-9",
		"agy-abc-123",
		"scratchpad",
	} {
		if !strings.Contains(content, sub) {
			t.Errorf("handshake content %q missing %q", content, sub)
		}
	}

	errDir := t.TempDir()
	errSocket := filepath.Join(errDir, "8.sock")
	writeJSON(t, filepath.Join(errDir, "8.json"), registration(8, "worker", "/w", errSocket))
	stubSessions(t, errDir, 8)
	listenInbox(t, errSocket)
	for _, missing := range []string{
		"ANTIGRAVITY_CONVERSATION_ID",
		"ANTIGRAVITY_LS_ADDRESS",
		"ANTIGRAVITY_CSRF_TOKEN",
	} {
		t.Setenv("ANTIGRAVITY_CONVERSATION_ID", "AbC-123")
		t.Setenv("ANTIGRAVITY_LS_ADDRESS", "127.0.0.1:42100")
		t.Setenv("ANTIGRAVITY_CSRF_TOKEN", "ToKeN-9")
		t.Setenv(missing, "")
		var stderr bytes.Buffer
		if code := run([]string{"send", "--to", "8", "--handshake"}, io.Discard, &stderr); code != exitError || !strings.Contains(stderr.String(), missing) {
			t.Errorf("missing %s: code=%d stderr=%q, want exitError naming %s", missing, code, stderr.String(), missing)
		}
	}
}
