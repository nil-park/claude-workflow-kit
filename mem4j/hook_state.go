package main

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"
)

type hookLogEntry struct {
	Timestamp    string   `json:"ts"`
	Event        string   `json:"event"`
	Status       string   `json:"status"`
	Reason       string   `json:"reason,omitempty"`
	ActiveScopes []string `json:"activeScopes,omitempty"`
	MatchedKeys  []string `json:"matchedKeys,omitempty"`
	TopScore     float64  `json:"topScore,omitempty"`
	DurationMs   int64    `json:"durationMs"`
	Error        string   `json:"error,omitempty"`
}

type sessionState struct {
	Scopes      []string `json:"scopes,omitempty"`
	LastTurnKey string   `json:"lastTurnKey,omitempty"`
}

func appendHookLog(logPath string, entry hookLogEntry) {
	if logPath == "" {
		var err error
		logPath, err = defaultHookLogPath()
		if err != nil {
			return
		}
	}
	if entry.Timestamp == "" {
		entry.Timestamp = time.Now().UTC().Format(time.RFC3339)
	}
	if err := os.MkdirAll(filepath.Dir(logPath), 0o700); err != nil {
		return
	}
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(entry); err != nil {
		return
	}
	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return
	}
	defer f.Close()
	_, _ = f.Write(buf.Bytes())
}

type agyTranscriptLine struct {
	Type string `json:"type"`
}

func agyTurnKey(transcriptPath, prompt string) string {
	sum := sha256.Sum256([]byte(normalizeText(prompt)))
	hash := hex.EncodeToString(sum[:8])
	if transcriptPath != "" {
		if count, err := countTranscriptUserInputs(transcriptPath); err == nil {
			return fmt.Sprintf("u:%d:%s", count, hash)
		}
	}
	return "p:" + hash
}

func countTranscriptUserInputs(path string) (int, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	reader := bufio.NewReader(f)
	count := 0
	for {
		line, err := reader.ReadBytes('\n')
		if len(line) > 0 {
			trimmed := bytes.TrimSpace(line)
			if len(trimmed) > 0 {
				var step agyTranscriptLine
				if json.Unmarshal(trimmed, &step) == nil && step.Type == "USER_INPUT" {
					count++
				}
			}
		}
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return 0, err
		}
	}
	return count, nil
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

func saveSessionState(path string, st sessionState) error {
	if path == "" {
		return nil
	}
	st.Scopes = mergeScopeSlices(st.Scopes, nil)
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(st); err != nil {
		return err
	}
	return os.WriteFile(path, buf.Bytes(), 0o600)
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
