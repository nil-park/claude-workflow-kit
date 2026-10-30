package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSendHandshake(t *testing.T) {
	t.Setenv("ANTIGRAVITY_CONVERSATION_ID", "AbC-123")
	t.Setenv("ANTIGRAVITY_LS_ADDRESS", "127.0.0.1:42100")
	t.Setenv("ANTIGRAVITY_CSRF_TOKEN", "ToKeN-9")

	bodyFile := filepath.Join(t.TempDir(), "body.md")
	if err := os.WriteFile(bodyFile, []byte("from file"), 0o600); err != nil {
		t.Fatal(err)
	}

	for _, tc := range []struct {
		name       string
		args       []string
		wantSuffix string
	}{
		{"alone", []string{"send", "--to", "8", "--handshake"}, "survive context compaction.\n</cross-session-message>"},
		{"with_message", []string{"send", "--to", "8", "--handshake", "--message", "hello body"}, "survive context compaction.\n\nhello body\n</cross-session-message>"},
		{"with_file", []string{"send", "--to", "8", "--handshake", "--file", bodyFile}, "survive context compaction.\n\nfrom file\n</cross-session-message>"},
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
			if !strings.HasSuffix(content, tc.wantSuffix) {
				t.Errorf("handshake content %q does not end with %q", content, tc.wantSuffix)
			}
		})
	}

	errDir := t.TempDir()
	errSocket := filepath.Join(errDir, "8.sock")
	writeJSON(t, filepath.Join(errDir, "8.json"), registration(8, "worker", "/w", errSocket))
	stubSessions(t, errDir, 8)
	listenInbox(t, errSocket)
	invalidUTF8 := filepath.Join(errDir, "bad.md")
	if err := os.WriteFile(invalidUTF8, []byte{0xff, 0xfe}, 0o600); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{
		{"send", "--to", "8", "--handshake", "--message", ""},
		{"send", "--to", "8", "--handshake", "--file", invalidUTF8},
	} {
		if code := run(args, io.Discard, io.Discard); code != exitError {
			t.Errorf("run(%q) = %d, want %d", args, code, exitError)
		}
	}
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
