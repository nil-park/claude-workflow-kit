package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestHookLogRotation(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "hook.log")

	appendHookLogWithLimit(logPath, hookLogEntry{Event: "first", Status: "ok"}, 50)
	appendHookLogWithLimit(logPath, hookLogEntry{Event: "second", Status: "ok"}, 50)

	backupRaw, err := os.ReadFile(logPath + ".1")
	if err != nil {
		t.Fatalf("expected rotated backup file: %v", err)
	}
	if !strings.Contains(string(backupRaw), `"event":"first"`) {
		t.Fatalf("backup missing first entry: %s", string(backupRaw))
	}
	curRaw, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("read current log: %v", err)
	}
	if !strings.Contains(string(curRaw), `"event":"second"`) || strings.Contains(string(curRaw), `"event":"first"`) {
		t.Fatalf("current log should only contain second entry: %s", string(curRaw))
	}
}

func TestPruneStaleStateFilesAndAtomicSave(t *testing.T) {
	dir := t.TempDir()
	oldFile := filepath.Join(dir, "old.json")
	activeOldFile := filepath.Join(dir, "active.json")
	freshFile := filepath.Join(dir, "fresh.json")
	oldTmp := filepath.Join(dir, ".mem4j-state-123.tmp")
	keepTxt := filepath.Join(dir, "note.txt")

	for _, p := range []string{oldFile, activeOldFile, freshFile} {
		if err := saveSessionState(p, sessionState{Scopes: []string{"박광석"}}); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(oldTmp, []byte("partial"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(keepTxt, []byte("keep"), 0o600); err != nil {
		t.Fatal(err)
	}
	now := time.Now()
	staleTime := now.Add(-10 * 24 * time.Hour)
	for _, p := range []string{oldFile, activeOldFile, oldTmp, keepTxt} {
		if err := os.Chtimes(p, staleTime, staleTime); err != nil {
			t.Fatal(err)
		}
	}

	pruneStaleStateFiles(dir, activeOldFile, 7*24*time.Hour, now)

	if _, err := os.Stat(oldFile); !os.IsNotExist(err) {
		t.Fatalf("old.json should be pruned, err=%v", err)
	}
	if _, err := os.Stat(oldTmp); !os.IsNotExist(err) {
		t.Fatalf("stale temp file should be pruned, err=%v", err)
	}
	if _, err := os.Stat(activeOldFile); err != nil {
		t.Fatalf("excluded active session file must remain: %v", err)
	}
	if _, err := os.Stat(freshFile); err != nil {
		t.Fatalf("fresh.json should remain: %v", err)
	}
	if _, err := os.Stat(keepTxt); err != nil {
		t.Fatalf("non-json file should remain: %v", err)
	}
}

func TestFormatSessionContextFallsBackToActiveFetchOverLimit(t *testing.T) {
	longBody := strings.Repeat("가", 9500)
	text := formatSessionContext(
		[]string{"박광석", "코드 작성"},
		`C:\tmp\scratchpad\mem4j-state.json`,
		[]Instruction{{Key: "k1", Title: "긴 지침", Scopes: []string{"박광석", "코드 작성"}, Body: longBody}},
		[]Scope{{Name: "박광석", Description: "사용자"}},
		claudeSessionSoftLimit,
	)
	if strings.Contains(text, "<instruction") {
		t.Fatalf("expected instructions to be omitted over soft limit")
	}
	if !strings.Contains(text, `mem4j instruction recall -state "C:/tmp/scratchpad/mem4j-state.json"`) {
		t.Fatalf("expected active fetch hint, got: %s", text)
	}
	if !strings.Contains(text, "<scope name=\"박광석\"") {
		t.Fatalf("expected scope list to remain present")
	}

	noStateText := formatSessionContext(
		[]string{"박광석", "코드 작성"},
		"",
		[]Instruction{{Key: "k1", Title: "긴 지침", Scopes: []string{"박광석", "코드 작성"}, Body: longBody}},
		[]Scope{{Name: "박광석", Description: "사용자"}},
		claudeSessionSoftLimit,
	)
	if !strings.Contains(noStateText, `mem4j instruction recall -scope "박광석" -scope "코드 작성"`) {
		t.Fatalf("expected scope-flag fetch hint when statePath is empty, got: %s", noStateText)
	}
}

func TestPromptHookLogsInvalidConfigError(t *testing.T) {
	dir := t.TempDir()
	badCfg := filepath.Join(dir, "bad.json")
	if err := os.WriteFile(badCfg, []byte("{invalid"), 0o600); err != nil {
		t.Fatal(err)
	}
	logPath := filepath.Join(dir, "hook.log")
	var stdout, stderr bytes.Buffer
	err := run(
		[]string{"hook", "claude-user-prompt", "-config", badCfg, "-credentials", filepath.Join(dir, "missing.txt"), "-log", logPath},
		strings.NewReader(`{"session_id":"s1","prompt":"변경 사항 커밋해줘"}`),
		&stdout, &stderr,
	)
	if err != nil {
		t.Fatalf("hook must fail open: %v", err)
	}
	if !strings.Contains(stderr.String(), "parse hook config") {
		t.Fatalf("expected config parse error on stderr, got: %q", stderr.String())
	}
	raw, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatal(err)
	}
	var entry hookLogEntry
	if err := json.Unmarshal(bytes.TrimSpace(raw), &entry); err != nil {
		t.Fatal(err)
	}
	if entry.Status != "error" || !strings.Contains(entry.Error, "parse hook config") {
		t.Fatalf("expected config parse error preserved in hook.log, got: %+v", entry)
	}
}
