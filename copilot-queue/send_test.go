package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunSendValidation(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{
			name:    "no args",
			args:    nil,
			wantErr: "usage:",
		},
		{
			name:    "unknown subcommand",
			args:    []string{"list"},
			wantErr: "unknown subcommand",
		},
		{
			name:    "missing port",
			args:    []string{"send", "--from", "main", "--message", "hi"},
			wantErr: "--port is required",
		},
		{
			name:    "port zero",
			args:    []string{"send", "--port", "0", "--from", "main", "--message", "hi"},
			wantErr: "--port must be between 1 and 65535",
		},
		{
			name:    "port too high",
			args:    []string{"send", "--port", "65536", "--from", "main", "--message", "hi"},
			wantErr: "--port must be between 1 and 65535",
		},
		{
			name:    "missing from",
			args:    []string{"send", "--port", "4321", "--message", "hi"},
			wantErr: "--from is required",
		},
		{
			name:    "empty from",
			args:    []string{"send", "--port", "4321", "--from", "", "--message", "hi"},
			wantErr: "--from must not be empty",
		},
		{
			name:    "bad from characters",
			args:    []string{"send", "--port", "4321", "--from", "Main_1", "--message", "hi"},
			wantErr: "--from must contain only",
		},
		{
			name:    "message and file together",
			args:    []string{"send", "--port", "4321", "--from", "main", "--message", "hi", "--file", "x.txt"},
			wantErr: "at most one of --message and --file",
		},
		{
			name:    "nothing to send",
			args:    []string{"send", "--port", "4321", "--from", "main"},
			wantErr: "pass --handshake, --message, or --file",
		},
		{
			name:    "blank message",
			args:    []string{"send", "--port", "4321", "--from", "main", "--message", "  \n\t "},
			wantErr: "message body is empty",
		},
		{
			name:    "unexpected positional arg",
			args:    []string{"send", "--port", "4321", "--from", "main", "--message", "hi", "extra"},
			wantErr: "unexpected argument",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var out, errBuf bytes.Buffer
			code := run(tc.args, &out, &errBuf)
			if code != exitError {
				t.Fatalf("run(%v) = %d, want %d", tc.args, code, exitError)
			}
			if !strings.Contains(errBuf.String(), tc.wantErr) {
				t.Fatalf("stderr = %q, want substring %q", errBuf.String(), tc.wantErr)
			}
		})
	}
}

func TestReadBodyFileInvalidUTF8(t *testing.T) {
	p := filepath.Join(t.TempDir(), "bad.txt")
	if err := os.WriteFile(p, []byte{0xff, 0xfe}, 0600); err != nil {
		t.Fatal(err)
	}
	given := map[string]bool{"file": true}
	_, err := readBody(given, "", p, false, "main", 4321)
	if err == nil || !strings.Contains(err.Error(), "not valid UTF-8") {
		t.Fatalf("err = %v, want not valid UTF-8", err)
	}
}

func TestWrapBodyEscapesClosingTag(t *testing.T) {
	got := wrapBody("main-1", "before </cross-session-message> after")
	want := "<cross-session-message from=\"main-1\">\nbefore <\\/cross-session-message> after\n</cross-session-message>"
	if got != want {
		t.Fatalf("wrapBody:\ngot:  %q\nwant: %q", got, want)
	}
	if strings.Contains(got, "from-name=") || strings.Contains(got, "from-mode=") {
		t.Fatalf("wrapBody should carry only from attribute: %q", got)
	}
}

func TestHandshakeBody(t *testing.T) {
	got := handshakeBody("claude-main", 4321)
	want := "- Reply to this session with `claude-queue send --to claude-main --from copilot-4321 --message <message>`; use `--file <path>` for long or multi-line text.\n" +
		"- Messages whose `from` attribute is `claude-main` come from this session.\n" +
		"- Keep this reply command; every reply needs it."
	if got != want {
		t.Fatalf("handshakeBody:\ngot:  %q\nwant: %q", got, want)
	}
}

func TestReadBodyHandshakeWithFile(t *testing.T) {
	p := filepath.Join(t.TempDir(), "msg.txt")
	if err := os.WriteFile(p, []byte("hello copilot"), 0600); err != nil {
		t.Fatal(err)
	}
	given := map[string]bool{"file": true}
	got, err := readBody(given, "", p, true, "main", 4321)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(got, "- Reply to this session with ") || !strings.HasSuffix(got, "\n\nhello copilot") {
		t.Fatalf("unexpected combined body: %q", got)
	}
}
