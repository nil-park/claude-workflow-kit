package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"
)

const hookTimeout = 10 * time.Second

type claudeSessionStartInput struct {
	Cwd    string `json:"cwd"`
	Source string `json:"source"`
}

type agySessionStartInput struct {
	WorkspacePaths []string `json:"workspacePaths"`
}

type hookOptions struct {
	configPath  string
	credPath    string
	gcpCredPath string
	logPath     string
	stateDir    string
	threshold   float64
	limit       int
}

func parseHookCommand(args []string, stdin io.Reader) (parsedCommand, error) {
	if len(args) == 0 {
		return parsedCommand{}, errors.New("usage: mem4j hook <claude-session-start|claude-user-prompt|agy-session-start|agy-pre-invocation> [flags]")
	}
	if cmd, ok, err := handleResourceHelp("hook", args, stdin); ok {
		return cmd, err
	}
	switch args[0] {
	case "claude-session-start", "agy-session-start":
		return parseSessionStartHook(args[0], args[1:], stdin)
	case "claude-user-prompt", "agy-pre-invocation":
		return parsePromptHook(args[0], args[1:], stdin)
	default:
		return parsedCommand{}, fmt.Errorf("unknown hook action %q", args[0])
	}
}

func parseSessionStartHook(action string, args []string, stdin io.Reader) (parsedCommand, error) {
	fs, cred := newFlagSet("hook " + action)
	cfgPath := fs.String("config", "", "hook config file (default ~/.config/mem4j/config.json)")
	logPath := fs.String("log", "", "hook JSONL log file (default ~/.config/mem4j/hook.log)")
	if err := parseFlags(fs, args); err != nil {
		return parsedCommand{}, err
	}
	opts := hookOptions{configPath: *cfgPath, credPath: *cred, logPath: *logPath}
	isAgy := action == "agy-session-start"
	return parsedCommand{
		execNoStore: func(stdout, stderr io.Writer) error {
			runSessionStartHook(action, isAgy, opts, stdin, stdout, stderr)
			return nil
		},
	}, nil
}

func runSessionStartHook(action string, isAgy bool, opts hookOptions, stdin io.Reader, stdout, stderr io.Writer) {
	start := time.Now()
	entry := hookLogEntry{Event: action, Status: "ok"}
	defer func() {
		entry.DurationMs = time.Since(start).Milliseconds()
		appendHookLog(opts.logPath, entry)
	}()

	workDirs, err := decodeSessionWorkDirs(isAgy, stdin)
	if err != nil {
		failOpenHook(isAgy, &entry, stdout, stderr, fmt.Errorf("decode stdin: %w", err))
		return
	}
	cfg, err := loadHookConfig(opts.configPath)
	if err != nil {
		failOpenHook(isAgy, &entry, stdout, stderr, err)
		return
	}
	baseScopes := resolveConfiguredScopes(cfg, workDirs)
	if len(baseScopes) == 0 {
		entry.Status = "skipped"
		entry.Reason = "no_scopes"
		if isAgy {
			writeAgyEmptyOutput(stdout)
		}
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), hookTimeout)
	defer cancel()
	s, closeStore, err := openStore(ctx, opts.credPath)
	if err != nil {
		failOpenHook(isAgy, &entry, stdout, stderr, err)
		return
	}
	defer closeStore()

	activeScopes, err := s.expandScopesOutgoing(ctx, baseScopes)
	if err != nil {
		failOpenHook(isAgy, &entry, stdout, stderr, err)
		return
	}
	entry.ActiveScopes = activeScopes
	candidates, err := s.fetchCandidates(ctx, activeScopes)
	if err != nil {
		failOpenHook(isAgy, &entry, stdout, stderr, err)
		return
	}
	allScopes, err := s.listScopes(ctx)
	if err != nil {
		failOpenHook(isAgy, &entry, stdout, stderr, err)
		return
	}
	merged := Merge(candidates, activeScopes)
	text := formatSessionContext(activeScopes, merged, allScopes)
	if isAgy {
		err = writeAgySessionOutput(stdout, text)
	} else {
		err = writeClaudeHookOutput(stdout, "SessionStart", text)
	}
	if err != nil {
		entry.Status = "error"
		entry.Error = err.Error()
		fmt.Fprintln(stderr, "mem4j hook:", err)
	}
}

func decodeSessionWorkDirs(isAgy bool, stdin io.Reader) ([]string, error) {
	if isAgy {
		var in agySessionStartInput
		if err := json.NewDecoder(stdin).Decode(&in); err != nil {
			return nil, err
		}
		return in.WorkspacePaths, nil
	}
	var in claudeSessionStartInput
	if err := json.NewDecoder(stdin).Decode(&in); err != nil {
		return nil, err
	}
	if in.Cwd == "" {
		return nil, nil
	}
	return []string{in.Cwd}, nil
}

func failOpenHook(isAgy bool, entry *hookLogEntry, stdout, stderr io.Writer, err error) {
	entry.Status = "error"
	entry.Error = err.Error()
	fmt.Fprintln(stderr, "mem4j hook:", err)
	if isAgy {
		writeAgyEmptyOutput(stdout)
	}
}
