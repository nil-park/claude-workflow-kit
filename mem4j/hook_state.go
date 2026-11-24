package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"
)

const stateRetention = 7 * 24 * time.Hour

type sessionState struct {
	Scopes              []string `json:"scopes,omitempty"`
	LastTurnKey         string   `json:"lastTurnKey,omitempty"`
	LastCheckpointCount int      `json:"lastCheckpointCount,omitempty"`
}

func resolveSessionStatePath(isAgy bool, sessionDir, sessionID, stateDirOverride string) (string, error) {
	if stateDirOverride != "" {
		if sessionID == "" {
			return "", nil
		}
		return filepath.Join(stateDirOverride, sanitizeStateID(sessionID)+".json"), nil
	}
	if trimmed := strings.TrimSpace(sessionDir); trimmed != "" {
		if isAgy {
			return filepath.Join(trimmed, "scratch", "mem4j-state.json"), nil
		}
		return filepath.Join(trimmed, "mem4j-state.json"), nil
	}
	if sessionID == "" {
		return "", nil
	}
	dir, err := defaultStateDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, sanitizeStateID(sessionID)+".json"), nil
}

func sanitizeStateID(id string) string {
	var b strings.Builder
	for _, r := range id {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			b.WriteRune(r)
		}
	}
	if b.Len() > 0 && b.Len() <= 80 {
		return b.String()
	}
	sum := sha256.Sum256([]byte(id))
	return hex.EncodeToString(sum[:12])
}

func loadSessionState(path string) (sessionState, error) {
	if path == "" {
		return sessionState{}, nil
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return sessionState{}, nil
		}
		return sessionState{}, err
	}
	var st sessionState
	if err := json.Unmarshal(raw, &st); err != nil {
		return sessionState{}, err
	}
	st.Scopes = mergeScopeSlices(st.Scopes, nil)
	return st, nil
}

// saveSessionState writes to a temp file in the same directory and renames it
// so concurrent hook processes never observe a half-written JSON file.
func saveSessionState(path string, st sessionState) error {
	if path == "" {
		return nil
	}
	st.Scopes = mergeScopeSlices(st.Scopes, nil)
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(st); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, ".mem4j-state-*.tmp")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	if _, err := tmp.Write(buf.Bytes()); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpPath)
		return err
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpPath)
		return err
	}
	if err := os.Chmod(tmpPath, 0o600); err != nil {
		_ = os.Remove(tmpPath)
		return err
	}
	if err := os.Rename(tmpPath, path); err != nil {
		_ = os.Remove(tmpPath)
		return err
	}
	return nil
}

func pruneStaleStateFiles(dir string, maxAge time.Duration, now time.Time) {
	if dir == "" {
		var err error
		dir, err = defaultStateDir()
		if err != nil {
			return
		}
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	cutoff := now.Add(-maxAge)
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		info, err := entry.Info()
		if err != nil || !info.ModTime().Before(cutoff) {
			continue
		}
		_ = os.Remove(filepath.Join(dir, entry.Name()))
	}
}

func mergeScopeSlices(a, b []string) []string {
	merged := make([]string, 0, len(a)+len(b))
	for _, list := range [][]string{a, b} {
		for _, s := range list {
			if n := normalizeText(s); n != "" {
				merged = append(merged, n)
			}
		}
	}
	out := dedup(merged)
	slices.Sort(out)
	return out
}
