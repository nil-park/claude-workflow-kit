package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestShouldRecallPrompt(t *testing.T) {
	cases := []struct {
		prompt string
		want   bool
	}{
		{"", false},
		{"   ", false},
		{"응", false},
		{"좋아.", false},
		{"맞아!!", false},
		{"오케이~", false},
		{"커밋해", false},
		{"커밋해.", false},
		{"커밋푸시", true},
		{"커밋 푸시", true},
		{"커밋하자.", true},
		{"진행하자.", true},
		{"PR 올려", true},
		{"ping", false},
		{"hello!", false},
		{"commit", false},
		{"commit push", true},
		{"fix bug", true},
	}
	for _, tc := range cases {
		if got := shouldRecallPrompt(tc.prompt); got != tc.want {
			t.Errorf("shouldRecallPrompt(%q) = %v, want %v", tc.prompt, got, tc.want)
		}
	}
}

func TestResolveConfiguredScopes(t *testing.T) {
	cfg := hookConfig{
		Scopes: []string{"박광석", "kwang@NilPark-A520M-K"},
		Workspaces: map[string][]string{
			"C:/Git/claude-workflow-kit": {"claude-workflow-kit 리포지토리"},
		},
	}

	got := resolveConfiguredScopes(cfg, []string{`C:\Git\claude-workflow-kit\mem4j`})
	want := []string{"claude-workflow-kit 리포지토리", "kwang@NilPark-A520M-K", "박광석"}
	if !slices.Equal(got, want) {
		t.Fatalf("subdir match = %v, want %v", got, want)
	}

	gotSibling := resolveConfiguredScopes(cfg, []string{`C:\Git\claude-workflow-kit-2`})
	wantBase := []string{"kwang@NilPark-A520M-K", "박광석"}
	if !slices.Equal(gotSibling, wantBase) {
		t.Fatalf("sibling prefix must not match: got %v, want %v", gotSibling, wantBase)
	}
}

func TestAgyTurnDeduplicationAndTranscriptCount(t *testing.T) {
	dir := t.TempDir()
	transcript := filepath.Join(dir, "transcript.jsonl")
	lines := strings.Join([]string{
		`{"source":"USER_EXPLICIT","type":"USER_INPUT","content":"계속 진행해줘"}`,
		`{"source":"MODEL","type":"PLANNER_RESPONSE","content":"ok"}`,
	}, "\n") + "\n"
	if err := os.WriteFile(transcript, []byte(lines), 0o600); err != nil {
		t.Fatal(err)
	}

	stateDir := filepath.Join(dir, "state")
	convID := "../evil/conv-1"
	key1 := agyTurnKey(transcript, "계속 진행해줘")
	if isDuplicateAgyTurn(stateDir, convID, key1) {
		t.Fatal("first turn should not be duplicate")
	}
	if err := saveAgyTurn(stateDir, convID, key1); err != nil {
		t.Fatal(err)
	}
	if !isDuplicateAgyTurn(stateDir, convID, key1) {
		t.Fatal("second invocation in same turn should be duplicate")
	}

	// Ensure state file stays inside stateDir despite traversal chars in convID.
	entries, err := os.ReadDir(stateDir)
	if err != nil || len(entries) != 1 {
		t.Fatalf("expected 1 state file in stateDir, got %v (err=%v)", entries, err)
	}

	// Appending a second identical user prompt increments USER_INPUT count.
	lines += `{"source":"USER_EXPLICIT","type":"USER_INPUT","content":"계속 진행해줘"}` + "\n"
	if err := os.WriteFile(transcript, []byte(lines), 0o600); err != nil {
		t.Fatal(err)
	}
	key2 := agyTurnKey(transcript, "계속 진행해줘")
	if isDuplicateAgyTurn(stateDir, convID, key2) {
		t.Fatal("new turn with identical prompt text should not be duplicate")
	}
}

func TestHookShortPromptAndFailOpen(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "hook.log")
	stateDir := filepath.Join(dir, "state")

	var stdout, stderr bytes.Buffer
	if err := run(
		[]string{"hook", "claude-user-prompt", "-log", logPath},
		strings.NewReader(`{"cwd":"C:/Git/claude-workflow-kit","prompt":"응"}`),
		&stdout, &stderr,
	); err != nil || stdout.Len() != 0 {
		t.Fatalf("claude short prompt: err=%v stdout=%q", err, stdout.String())
	}

	stdout.Reset()
	stderr.Reset()
	if err := run(
		[]string{"hook", "agy-pre-invocation", "-log", logPath, "-state-dir", stateDir},
		strings.NewReader(`{"conversationId":"c1","lastUserInput":"ping"}`),
		&stdout, &stderr,
	); err != nil || strings.TrimSpace(stdout.String()) != "{}" {
		t.Fatalf("agy short prompt: err=%v stdout=%q", err, stdout.String())
	}

	stdout.Reset()
	stderr.Reset()
	if err := run(
		[]string{"hook", "agy-session-start", "-config", filepath.Join(dir, "missing.json"), "-log", logPath},
		strings.NewReader(`{"workspacePaths":["C:/Git/claude-workflow-kit"]}`),
		&stdout, &stderr,
	); err != nil || strings.TrimSpace(stdout.String()) != "{}" {
		t.Fatalf("agy missing config: err=%v stdout=%q", err, stdout.String())
	}

	rawLog, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("read hook log: %v", err)
	}
	logLines := strings.Split(strings.TrimSpace(string(rawLog)), "\n")
	if len(logLines) != 3 {
		t.Fatalf("expected 3 log lines, got %d: %s", len(logLines), string(rawLog))
	}
	var last hookLogEntry
	if err := json.Unmarshal([]byte(logLines[2]), &last); err != nil {
		t.Fatalf("unmarshal log line: %v", err)
	}
	if last.Status != "error" || last.Event != "agy-session-start" {
		t.Fatalf("unexpected last log entry: %+v", last)
	}
}

func TestHookOutputFormatting(t *testing.T) {
	sessionText := formatSessionContext(
		[]string{"박광석"},
		[]Instruction{{Key: "k1", Title: "규칙", Scopes: []string{"박광석"}, Body: "본문"}},
		[]Scope{{Name: "박광석", Description: "사용자"}},
	)
	if !strings.Contains(sessionText, "활성 스코프: 박광석") || !strings.Contains(sessionText, "<instruction") {
		t.Fatalf("unexpected session text: %s", sessionText)
	}

	var claudeOut bytes.Buffer
	if err := writeClaudeHookOutput(&claudeOut, "SessionStart", sessionText); err != nil {
		t.Fatal(err)
	}
	var parsed claudeHookOutput
	if err := json.Unmarshal(claudeOut.Bytes(), &parsed); err != nil {
		t.Fatalf("invalid claude json: %v", err)
	}
	if parsed.HookSpecificOutput.HookEventName != "SessionStart" {
		t.Fatalf("unexpected hookEventName: %+v", parsed)
	}

	var agyOut bytes.Buffer
	if err := writeAgyPromptOutput(&agyOut, "리마인더 내용"); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(agyOut.String(), `"ephemeralMessage":"리마인더 내용"`) {
		t.Fatalf("unexpected agy prompt json: %s", agyOut.String())
	}
}
