package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

const maxHookLogBytes int64 = 5 * 1024 * 1024

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

func appendHookLog(logPath string, entry hookLogEntry) {
	appendHookLogWithLimit(logPath, entry, maxHookLogBytes)
}

func appendHookLogWithLimit(logPath string, entry hookLogEntry, maxBytes int64) {
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
	rotateHookLogIfNeeded(logPath, maxBytes)
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

func rotateHookLogIfNeeded(logPath string, maxBytes int64) {
	if maxBytes <= 0 {
		return
	}
	info, err := os.Stat(logPath)
	if err != nil || info.Size() < maxBytes {
		return
	}
	// os.Rename replaces an existing backup on POSIX and Windows alike;
	// avoiding a prior Remove preserves the backup if Rename fails.
	_ = os.Rename(logPath, logPath+".1")
}

func appendEntryError(entry *hookLogEntry, err error) {
	if err == nil {
		return
	}
	msg := err.Error()
	if entry.Error == "" {
		entry.Error = msg
		return
	}
	entry.Error += "; " + msg
}
