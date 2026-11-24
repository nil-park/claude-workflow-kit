package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"
)

type claudeUserPromptInput struct {
	Cwd    string `json:"cwd"`
	Prompt string `json:"prompt"`
}

type agyPreInvocationInput struct {
	ConversationID string   `json:"conversationId"`
	WorkspacePaths []string `json:"workspacePaths"`
	TranscriptPath string   `json:"transcriptPath"`
	LastUserInput  string   `json:"lastUserInput"`
}

func parsePromptHook(action string, args []string, stdin io.Reader) (parsedCommand, error) {
	isAgy := action == "agy-pre-invocation"
	fs, cred := newFlagSet("hook " + action)
	cfgPath := fs.String("config", "", "hook config file (default ~/.config/mem4j/config.json)")
	gcpCred := fs.String("gcp-credentials", "", "GCP service account JSON (default ~/.config/mem4j/.secrets/gcp.json)")
	logPath := fs.String("log", "", "hook JSONL log file (default ~/.config/mem4j/hook.log)")
	var stateDir *string
	if isAgy {
		stateDir = fs.String("state-dir", "", "state directory for agy turn deduplication (default ~/.config/mem4j/state)")
	} else {
		empty := ""
		stateDir = &empty
	}
	threshold := fs.Float64("threshold", defaultReminderThreshold, "minimum cosine similarity score (-1..1)")
	limit := fs.Int("limit", defaultReminderLimit, "maximum number of reminders to return")
	if err := parseFlags(fs, args); err != nil {
		return parsedCommand{}, err
	}
	if *threshold < -1 || *threshold > 1 {
		return parsedCommand{}, fmt.Errorf("hook %s: -threshold must be between -1 and 1", action)
	}
	if *limit < 1 {
		return parsedCommand{}, fmt.Errorf("hook %s: -limit must be at least 1", action)
	}
	opts := hookOptions{
		configPath: *cfgPath, credPath: *cred, gcpCredPath: *gcpCred,
		logPath: *logPath, stateDir: *stateDir, threshold: *threshold, limit: *limit,
	}
	return parsedCommand{
		execNoStore: func(stdout, stderr io.Writer) error {
			runPromptHook(action, isAgy, opts, stdin, stdout, stderr)
			return nil
		},
	}, nil
}

func runPromptHook(action string, isAgy bool, opts hookOptions, stdin io.Reader, stdout, stderr io.Writer) {
	start := time.Now()
	entry := hookLogEntry{Event: action, Status: "ok"}
	defer func() {
		entry.DurationMs = time.Since(start).Milliseconds()
		appendHookLog(opts.logPath, entry)
	}()

	prompt, workDirs, convID, transcriptPath, err := decodePromptInput(isAgy, stdin)
	if err != nil {
		failOpenHook(isAgy, &entry, stdout, stderr, fmt.Errorf("decode stdin: %w", err))
		return
	}
	var turnKey string
	if isAgy {
		turnKey = agyTurnKey(transcriptPath, prompt)
		if isDuplicateAgyTurn(opts.stateDir, convID, turnKey) {
			entry.Status = "skipped"
			entry.Reason = "duplicate_turn"
			writeAgyEmptyOutput(stdout)
			return
		}
	}
	if !shouldRecallPrompt(prompt) {
		entry.Status = "skipped"
		entry.Reason = "short_prompt"
		if isAgy {
			_ = saveAgyTurn(opts.stateDir, convID, turnKey)
			writeAgyEmptyOutput(stdout)
		}
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
			_ = saveAgyTurn(opts.stateDir, convID, turnKey)
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
	emb, err := newVertexEmbedder(ctx, opts.gcpCredPath)
	if err != nil {
		failOpenHook(isAgy, &entry, stdout, stderr, err)
		return
	}
	vec, err := emb.Embed(ctx, prompt, taskTypeQuery)
	if err != nil {
		failOpenHook(isAgy, &entry, stdout, stderr, err)
		return
	}
	reminders, err := s.recallReminders(ctx, activeScopes, vec, opts.threshold, opts.limit)
	if err != nil {
		failOpenHook(isAgy, &entry, stdout, stderr, err)
		return
	}
	for i, r := range reminders {
		entry.MatchedKeys = append(entry.MatchedKeys, r.Key)
		if i == 0 {
			entry.TopScore = r.Score
		}
	}
	if isAgy {
		_ = saveAgyTurn(opts.stateDir, convID, turnKey)
	}
	text := formatReminderContext(activeScopes, reminders)
	if isAgy {
		err = writeAgyPromptOutput(stdout, text)
	} else {
		err = writeClaudeHookOutput(stdout, "UserPromptSubmit", text)
	}
	if err != nil {
		entry.Status = "error"
		entry.Error = err.Error()
		fmt.Fprintln(stderr, "mem4j hook:", err)
	}
}

func decodePromptInput(isAgy bool, stdin io.Reader) (prompt string, workDirs []string, convID, transcriptPath string, err error) {
	if isAgy {
		var in agyPreInvocationInput
		if err := json.NewDecoder(stdin).Decode(&in); err != nil {
			return "", nil, "", "", err
		}
		return in.LastUserInput, in.WorkspacePaths, in.ConversationID, in.TranscriptPath, nil
	}
	var in claudeUserPromptInput
	if err := json.NewDecoder(stdin).Decode(&in); err != nil {
		return "", nil, "", "", err
	}
	if in.Cwd != "" {
		workDirs = []string{in.Cwd}
	}
	return in.Prompt, workDirs, "", "", nil
}
