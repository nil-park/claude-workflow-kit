package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// stubSessions points the package at dir and treats only the given pids as
// running processes.
func stubSessions(t *testing.T, dir string, livePIDs ...int) {
	t.Helper()
	originalDir, originalAlive := sessionsDir, sessionAlive
	sessionsDir = func() (string, error) { return dir, nil }
	sessionAlive = func(pid int, _ string) bool {
		for _, live := range livePIDs {
			if pid == live {
				return true
			}
		}
		return false
	}
	t.Cleanup(func() { sessionsDir, sessionAlive = originalDir, originalAlive })
}

func writeJSON(t *testing.T, path string, v any) {
	t.Helper()
	raw, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, raw, 0o600); err != nil {
		t.Fatal(err)
	}
}

func registration(pid int, name, cwd, socket string) map[string]any {
	return map[string]any{
		"pid": pid, "name": name, "kind": "interactive", "status": "idle",
		"cwd": cwd, "messagingSocketPath": socket, "procStart": "1",
	}
}

func TestLiveSessionsSkipsUnusableEntries(t *testing.T) {
	dir := t.TempDir()
	writeJSON(t, filepath.Join(dir, "30.json"), registration(30, "beta", "/w/b", "/s/30.sock"))
	writeJSON(t, filepath.Join(dir, "20.json"), registration(20, "", "/w/alpha", "/s/20.sock"))
	writeJSON(t, filepath.Join(dir, "10.json"), registration(10, "beta", "/w/c", "/s/10.sock"))
	// Each entry below breaks exactly one acceptance rule.
	writeJSON(t, filepath.Join(dir, "40.json"), registration(40, "dead", "/w", "/s/40.sock"))
	writeJSON(t, filepath.Join(dir, "050.json"), registration(50, "padded", "/w", "/s/50.sock"))
	writeJSON(t, filepath.Join(dir, "60.json"), registration(61, "mismatch", "/w", "/s/60.sock"))
	writeJSON(t, filepath.Join(dir, "70.json"), registration(70, "nosocket", "/w", ""))
	writeJSON(t, filepath.Join(dir, "80.abc.key"), registration(80, "key", "/w", "/s/80.sock"))
	if err := os.WriteFile(filepath.Join(dir, "90.json"), []byte("{"), 0o600); err != nil {
		t.Fatal(err)
	}
	stubSessions(t, dir, 10, 20, 30, 50, 60, 61, 70, 80, 90)

	sessions, err := liveSessions()
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, s := range sessions {
		got = append(got, s.displayName()+":"+s.SocketPath)
	}
	want := "alpha:/s/20.sock beta:/s/10.sock beta:/s/30.sock"
	if strings.Join(got, " ") != want {
		t.Errorf("liveSessions() = %v, want %s", got, want)
	}
}

func TestLiveSessionsWithoutDirectory(t *testing.T) {
	stubSessions(t, filepath.Join(t.TempDir(), "missing"))
	sessions, err := liveSessions()
	if err != nil || len(sessions) != 0 {
		t.Errorf("liveSessions() = %v, %v; want no sessions and no error", sessions, err)
	}
}

func TestResolveSession(t *testing.T) {
	sessions := []session{
		{PID: 10, Name: "beta", Cwd: "/w/one"},
		{PID: 20, Name: "beta", Cwd: "/w/two"},
		{PID: 30, Cwd: "/w/gamma"},
	}
	for target, wantPID := range map[string]int{"10": 10, "20": 20, "gamma": 30} {
		got, err := resolveSession(sessions, target)
		if err != nil || got.PID != wantPID {
			t.Errorf("resolveSession(%q) = pid %d, %v; want pid %d", target, got.PID, err, wantPID)
		}
	}
	_, err := resolveSession(sessions, "beta")
	if err == nil || !strings.Contains(err.Error(), "/w/one") || !strings.Contains(err.Error(), "/w/two") {
		t.Errorf("resolveSession(beta) error = %v, want both candidate cwds", err)
	}
	for _, target := range []string{"99", "missing"} {
		if _, err := resolveSession(sessions, target); err == nil {
			t.Errorf("resolveSession(%q) succeeded, want an error", target)
		}
	}
}

func TestProcessMatchesComparesStartTime(t *testing.T) {
	pid := os.Getpid()
	start, ok := processStart(pid)
	if !ok {
		t.Skip("process start time is unavailable here")
	}
	if !processMatches(pid, start) {
		t.Errorf("processMatches(self, %q) = false, want true", start)
	}
	if !processMatches(pid, "") {
		t.Error("processMatches(self, empty) = false, want true")
	}
	if processMatches(pid, start+"0") {
		t.Error("processMatches(self, wrong start) = true, want false")
	}
}
