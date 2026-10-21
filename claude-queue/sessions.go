package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// session holds the fields of ~/.claude/sessions/<pid>.json that
// claude-queue reads.
type session struct {
	PID        int    `json:"pid"`
	Name       string `json:"name"`
	Kind       string `json:"kind"`
	Status     string `json:"status"`
	Cwd        string `json:"cwd"`
	SocketPath string `json:"messagingSocketPath"`
	ProcStart  string `json:"procStart"`
}

// displayName falls back to the last cwd element the way ListAgents names an
// unnamed session.
func (s session) displayName() string {
	if s.Name != "" {
		return s.Name
	}
	return filepath.Base(s.Cwd)
}

// sessionsDir and sessionAlive are variables so tests can point them at a
// temporary directory and fake processes.
var (
	sessionsDir  = defaultSessionsDir
	sessionAlive = processMatches
)

func defaultSessionsDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot locate the home directory: %w", err)
	}
	return filepath.Join(home, ".claude", "sessions"), nil
}

// processMatches skips the start time check when it cannot read the start
// time, so a session is not dropped only because ps or the process query
// failed.
func processMatches(pid int, procStart string) bool {
	if !processAlive(pid) {
		return false
	}
	if procStart == "" {
		return true
	}
	start, ok := processStart(pid)
	return !ok || start == procStart
}

// liveSessions returns the reachable sessions sorted by name, then pid.
func liveSessions() ([]session, error) {
	dir, err := sessionsDir()
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(dir)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var sessions []session
	for _, entry := range entries {
		pid, ok := registrationPID(entry.Name())
		if !ok {
			continue
		}
		s, ok := readSession(filepath.Join(dir, entry.Name()))
		if !ok || s.PID != pid || s.SocketPath == "" || !sessionAlive(s.PID, s.ProcStart) {
			continue
		}
		sessions = append(sessions, s)
	}
	sort.Slice(sessions, func(i, j int) bool {
		if a, b := sessions[i].displayName(), sessions[j].displayName(); a != b {
			return a < b
		}
		return sessions[i].PID < sessions[j].PID
	})
	return sessions, nil
}

// registrationPID rejects padded names such as 050.json because Claude Code
// names a registration after the pid without leading zeros.
func registrationPID(fileName string) (int, bool) {
	digits, ok := strings.CutSuffix(fileName, ".json")
	if !ok {
		return 0, false
	}
	pid, err := strconv.Atoi(digits)
	if err != nil || pid <= 0 || strconv.Itoa(pid) != digits {
		return 0, false
	}
	return pid, true
}

func readSession(path string) (session, bool) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return session{}, false
	}
	var s session
	if err := json.Unmarshal(raw, &s); err != nil {
		return session{}, false
	}
	return s, true
}

// resolveSession treats a decimal target as a pid and anything else as a
// session name.
func resolveSession(sessions []session, target string) (session, error) {
	if pid, err := strconv.Atoi(target); err == nil {
		for _, s := range sessions {
			if s.PID == pid {
				return s, nil
			}
		}
		return session{}, fmt.Errorf("no live session with pid %d", pid)
	}
	var matches []session
	for _, s := range sessions {
		if s.displayName() == target {
			matches = append(matches, s)
		}
	}
	switch len(matches) {
	case 0:
		return session{}, fmt.Errorf("no live session named %q; run claude-queue list", target)
	case 1:
		return matches[0], nil
	}
	var b strings.Builder
	fmt.Fprintf(&b, "%d live sessions are named %q; pass --to <pid>:", len(matches), target)
	for _, s := range matches {
		fmt.Fprintf(&b, "\n  %d\t%s", s.PID, s.Cwd)
	}
	return session{}, errors.New(b.String())
}
