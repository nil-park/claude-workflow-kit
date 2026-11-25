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
		{"", false}, {"   ", false}, {"응", false}, {"좋아.", false},
		{"맞아!!", false}, {"오케이~", false}, {"커밋해", false}, {"커밋해.", false},
		{"커밋푸시", true}, {"커밋 푸시", true}, {"커밋하자.", true}, {"진행하자.", true},
		{"PR 올려", true}, {"ping", false}, {"hello!", false}, {"commit", false},
		{"commit push", true}, {"fix bug", true},
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

func TestSessionStatePathsAndAgyTurnDedup(t *testing.T) {
	dir := t.TempDir()
	claudePath, err := resolveSessionStatePath(false, filepath.Join(dir, "scratchpad"), "s1", "")
	if err != nil || claudePath != filepath.Join(dir, "scratchpad", "mem4j-state.json") {
		t.Fatalf("claude state path = %q (err=%v)", claudePath, err)
	}
	agyPath, err := resolveSessionStatePath(true, filepath.Join(dir, "brain", "c1"), "c1", "")
	if err != nil || agyPath != filepath.Join(dir, "brain", "c1", "scratch", "mem4j-state.json") {
		t.Fatalf("agy state path = %q (err=%v)", agyPath, err)
	}
	safePath, err := resolveSessionStatePath(true, "", "../evil/conv-1", filepath.Join(dir, "override"))
	if err != nil || filepath.Dir(safePath) != filepath.Join(dir, "override") {
		t.Fatalf("traversal must stay in override dir: %q (err=%v)", safePath, err)
	}

	transcript := filepath.Join(dir, "transcript.jsonl")
	lines := strings.Join([]string{
		`{"source":"USER_EXPLICIT","type":"USER_INPUT","content":"계속 진행해줘"}`,
		`{"source":"MODEL","type":"PLANNER_RESPONSE","content":"ok"}`,
	}, "\n") + "\n"
	if err := os.WriteFile(transcript, []byte(lines), 0o600); err != nil {
		t.Fatal(err)
	}
	u1, cp1, err := scanAgyTranscript(transcript)
	if err != nil || u1 != 1 || cp1 != 0 {
		t.Fatalf("scan1 = (%d, %d, %v), want (1, 0, nil)", u1, cp1, err)
	}
	key1 := agyTurnKeyFromCount(u1, true, "계속 진행해줘")
	st := sessionState{Scopes: []string{"박광석", "코드 작성"}, LastTurnKey: key1, LastCheckpointCount: cp1}
	if err := saveSessionState(agyPath, st); err != nil {
		t.Fatal(err)
	}
	loaded, err := loadSessionState(agyPath)
	if err != nil || loaded.LastTurnKey != key1 || !slices.Equal(loaded.Scopes, []string{"박광석", "코드 작성"}) {
		t.Fatalf("unexpected loaded state: %+v (err=%v)", loaded, err)
	}

	lines += `{"source":"SYSTEM_SDK","type":"CHECKPOINT","content":"compacted"}` + "\n"
	lines += `{"source":"USER_EXPLICIT","type":"USER_INPUT","content":"계속 진행해줘"}` + "\n"
	if err := os.WriteFile(transcript, []byte(lines), 0o600); err != nil {
		t.Fatal(err)
	}
	u2, cp2, err := scanAgyTranscript(transcript)
	if err != nil || u2 != 2 || cp2 != 1 {
		t.Fatalf("scan2 = (%d, %d, %v), want (2, 1, nil)", u2, cp2, err)
	}
	if key2 := agyTurnKeyFromCount(u2, true, "계속 진행해줘"); key2 == loaded.LastTurnKey {
		t.Fatalf("new turn with identical prompt must produce new key: %q", key2)
	}
}

func TestHookShortPromptAndFailOpen(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "hook.log")
	brainDir := filepath.Join(dir, "brain", "c1")

	var stdout, stderr bytes.Buffer
	if err := run(
		[]string{"hook", "claude-user-prompt", "-log", logPath},
		strings.NewReader(`{"session_id":"s1","cwd":"C:/Git/claude-workflow-kit","prompt":"응"}`),
		&stdout, &stderr,
	); err != nil || stdout.Len() != 0 {
		t.Fatalf("claude short prompt: err=%v stdout=%q", err, stdout.String())
	}

	stdout.Reset()
	stderr.Reset()
	inJSON := `{"conversationId":"c1","artifactDirectoryPath":` + mustJSON(t, brainDir) + `,"lastUserInput":"ping"}`
	if err := run(
		[]string{"hook", "agy-pre-invocation", "-log", logPath},
		strings.NewReader(inJSON),
		&stdout, &stderr,
	); err != nil || strings.TrimSpace(stdout.String()) != "{}" {
		t.Fatalf("agy short prompt: err=%v stdout=%q", err, stdout.String())
	}

	stdout.Reset()
	stderr.Reset()
	if err := run(
		[]string{"hook", "agy-session-start", "-config", filepath.Join(dir, "missing.json"), "-log", logPath},
		strings.NewReader(`{"conversationId":"c1","workspacePaths":["C:/Git/claude-workflow-kit"]}`),
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
		t.Fatalf("expected 3 log lines, got %d", len(logLines))
	}
	var last hookLogEntry
	if err := json.Unmarshal([]byte(logLines[2]), &last); err != nil || last.Status != "error" {
		t.Fatalf("unexpected last log entry: %+v (err=%v)", last, err)
	}
}

func TestHookOutputFormatting(t *testing.T) {
	sessionText := formatSessionContext(
		[]string{"박광석"},
		`C:\tmp\scratch\mem4j-state.json`,
		[]Instruction{{Key: "k1", Title: "규칙", Scopes: []string{"박광석"}, Body: "본문"}},
		[]Scope{{Name: "박광석", Description: "사용자"}},
		claudeSessionSoftLimit,
	)
	if !strings.Contains(sessionText, "활성 스코프: 박광석") || !strings.Contains(sessionText, "세션 상태 파일: C:/tmp/scratch/mem4j-state.json") {
		t.Fatalf("unexpected session text: %s", sessionText)
	}

	var claudeOut bytes.Buffer
	if err := writeClaudeHookOutput(&claudeOut, "SessionStart", sessionText); err != nil {
		t.Fatal(err)
	}
	var parsed claudeHookOutput
	if err := json.Unmarshal(claudeOut.Bytes(), &parsed); err != nil || parsed.HookSpecificOutput.HookEventName != "SessionStart" {
		t.Fatalf("invalid claude json: %+v (err=%v)", parsed, err)
	}

	promptText := formatPromptContext(
		[]string{"박광석"},
		[]Reminder{{Key: "r1", Title: "규칙", Scopes: []string{"박광석"}, Score: 0.85, Body: "리마인더"}},
		[]Statement{{Key: "s1", Title: "사실", Entities: []string{"Recallar"}, Score: 0.91, Body: "안드로이드 앱"}},
		[]Event{{Key: "e1", Title: "이력", OccurredAt: "2026-11-01", Entities: []string{"Recallar"}, Score: 0.88, Body: "이름 변경"}},
	)
	if !strings.Contains(promptText, "현재 프롬프트만 비교해 조회했으므로 현재 작업과 무관한 항목이 포함될 수 있다.") {
		t.Fatalf("prompt text missing similarity caveat: %s", promptText)
	}
	if !strings.Contains(promptText, "<reminder") || !strings.Contains(promptText, "<statement") || !strings.Contains(promptText, "<event") {
		t.Fatalf("unexpected prompt text: %s", promptText)
	}
	var agyOut bytes.Buffer
	if err := writeAgyStepsOutput(&agyOut, sessionText, promptText); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(agyOut.String(), `"systemMessage":`) || !strings.Contains(agyOut.String(), `"ephemeralMessage":`) {
		t.Fatalf("unexpected agy steps json: %s", agyOut.String())
	}
}

func mustJSON(t *testing.T, v any) string {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}
