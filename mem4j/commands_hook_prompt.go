package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"
)

type claudeUserPromptInput struct {
	SessionID     string `json:"session_id"`
	ScratchpadDir string `json:"scratchpad_dir"`
	Cwd           string `json:"cwd"`
	Prompt        string `json:"prompt"`
}

type agyPreInvocationInput struct {
	ConversationID        string   `json:"conversationId"`
	ArtifactDirectoryPath string   `json:"artifactDirectoryPath"`
	WorkspacePaths        []string `json:"workspacePaths"`
	TranscriptPath        string   `json:"transcriptPath"`
	LastUserInput         string   `json:"lastUserInput"`
}

type promptHookInput struct {
	prompt         string
	workDirs       []string
	sessionDir     string
	sessionID      string
	transcriptPath string
}

func parsePromptHook(action string, args []string, stdin io.Reader) (parsedCommand, error) {
	isAgy := action == "agy-pre-invocation"
	fs, cred := newFlagSet("hook " + action)
	cfgPath := fs.String("config", "", "hook config file (default ~/.config/mem4j/config.json)")
	gcpCred := fs.String("gcp-credentials", "", "GCP service account JSON (default ~/.config/mem4j/.secrets/gcp.json)")
	logPath := fs.String("log", "", "hook JSONL log file (default ~/.config/mem4j/hook.log)")
	stateDir := fs.String("state-dir", "", "override directory for session state files")
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

	in, err := decodePromptInput(isAgy, stdin)
	if err != nil {
		failOpenHook(isAgy, &entry, stdout, stderr, fmt.Errorf("decode stdin: %w", err))
		return
	}
	statePath, _ := resolveSessionStatePath(isAgy, in.sessionDir, in.sessionID, opts.stateDir)
	st, _ := loadSessionState(statePath)
	var turnKey string
	if isAgy {
		turnKey = agyTurnKey(in.transcriptPath, in.prompt)
		if turnKey != "" && st.LastTurnKey == turnKey {
			entry.Status = "skipped"
			entry.Reason = "duplicate_turn"
			writeAgyEmptyOutput(stdout)
			return
		}
	}
	if !shouldRecallPrompt(in.prompt) {
		entry.Status = "skipped"
		entry.Reason = "short_prompt"
		if isAgy {
			st.LastTurnKey = turnKey
			_ = saveSessionState(statePath, st)
			writeAgyEmptyOutput(stdout)
		}
		return
	}
	cfg, cfgErr := loadHookConfig(opts.configPath)
	if cfgErr != nil && len(st.Scopes) == 0 {
		failOpenHook(isAgy, &entry, stdout, stderr, cfgErr)
		return
	}
	activeScopes := mergeScopeSlices(st.Scopes, resolveConfiguredScopes(cfg, in.workDirs))
	if len(activeScopes) == 0 {
		entry.Status = "skipped"
		entry.Reason = "no_scopes"
		if isAgy {
			st.LastTurnKey = turnKey
			_ = saveSessionState(statePath, st)
			writeAgyEmptyOutput(stdout)
		}
		return
	}
	entry.ActiveScopes = activeScopes
	ctx, cancel := context.WithTimeout(context.Background(), hookTimeout)
	defer cancel()
	s, closeStore, err := openStore(ctx, opts.credPath)
	if err != nil {
		failOpenHook(isAgy, &entry, stdout, stderr, err)
		return
	}
	defer closeStore()

	emb, err := newVertexEmbedder(ctx, opts.gcpCredPath)
	if err != nil {
		failOpenHook(isAgy, &entry, stdout, stderr, err)
		return
	}
	vec, err := emb.Embed(ctx, in.prompt, taskTypeQuery)
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
		st.LastTurnKey = turnKey
		st.Scopes = activeScopes
		_ = saveSessionState(statePath, st)
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

func decodePromptInput(isAgy bool, stdin io.Reader) (promptHookInput, error) {
	if isAgy {
		var in agyPreInvocationInput
		if err := json.NewDecoder(stdin).Decode(&in); err != nil {
			return promptHookInput{}, err
		}
		return promptHookInput{
			prompt:         in.LastUserInput,
			workDirs:       in.WorkspacePaths,
			sessionDir:     in.ArtifactDirectoryPath,
			sessionID:      in.ConversationID,
			transcriptPath: in.TranscriptPath,
		}, nil
	}
	var in claudeUserPromptInput
	if err := json.NewDecoder(stdin).Decode(&in); err != nil {
		return promptHookInput{}, err
	}
	var workDirs []string
	if in.Cwd != "" {
		workDirs = []string{in.Cwd}
	}
	return promptHookInput{
		prompt:     in.Prompt,
		workDirs:   workDirs,
		sessionDir: in.ScratchpadDir,
		sessionID:  in.SessionID,
	}, nil
}
