package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunErrorsPrintEmptyObjectAndStderr(t *testing.T) {
	cases := []struct {
		name       string
		stdinJSON  string
		wantStderr string
	}{
		{
			name:       "stdin not JSON",
			stdinJSON:  "not-json",
			wantStderr: "decode stdin",
		},
		{
			name:       "empty transcriptPath",
			stdinJSON:  `{"transcriptPath":""}`,
			wantStderr: "empty transcriptPath",
		},
		{
			name:       "missing transcript file",
			stdinJSON:  `{"transcriptPath":"missing-transcript-file-12345.jsonl"}`,
			wantStderr: "open transcript",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			run(strings.NewReader(tc.stdinJSON), &stdout, &stderr)

			if got := strings.TrimSpace(stdout.String()); got != "{}" {
				t.Errorf("stdout = %q, want %q", got, "{}")
			}
			if !strings.Contains(stderr.String(), tc.wantStderr) {
				t.Errorf("stderr = %q, want it to contain %q", stderr.String(), tc.wantStderr)
			}
		})
	}
}

func TestRunInjectCase(t *testing.T) {
	dir := t.TempDir()
	transcriptPath := filepath.Join(dir, "transcript.jsonl")
	content := `{"type":"CHECKPOINT","content":"# Resuming from a compaction"}` + "\n"
	if err := os.WriteFile(transcriptPath, []byte(content), 0o600); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	stdinJSON := `{"transcriptPath":` + quoteJSON(transcriptPath) + `}`
	var stdout, stderr bytes.Buffer
	run(strings.NewReader(stdinJSON), &stdout, &stderr)

	if stderr.Len() != 0 {
		t.Errorf("unexpected stderr: %s", stderr.String())
	}

	var out injectOutput
	if err := json.Unmarshal(stdout.Bytes(), &out); err != nil {
		t.Fatalf("stdout is not valid injectOutput JSON: %v, raw: %q", err, stdout.String())
	}
	if len(out.InjectSteps) != 1 {
		t.Fatalf("got %d inject steps, want 1", len(out.InjectSteps))
	}
	msg := out.InjectSteps[0].SystemMessage.SystemMessage
	if !strings.HasPrefix(msg, hookMarker) {
		t.Errorf("message does not start with marker: %q", msg)
	}
}

func TestRunNoInjectCase(t *testing.T) {
	dir := t.TempDir()
	transcriptPath := filepath.Join(dir, "transcript.jsonl")
	content := `{"type":"USER_MESSAGE","content":"hello"}` + "\n"
	if err := os.WriteFile(transcriptPath, []byte(content), 0o600); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	stdinJSON := `{"transcriptPath":` + quoteJSON(transcriptPath) + `}`
	var stdout, stderr bytes.Buffer
	run(strings.NewReader(stdinJSON), &stdout, &stderr)

	if stderr.Len() != 0 {
		t.Errorf("unexpected stderr: %s", stderr.String())
	}
	if got := strings.TrimSpace(stdout.String()); got != "{}" {
		t.Errorf("stdout = %q, want %q", got, "{}")
	}
}

func quoteJSON(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}
