package main

import (
	"context"
	"errors"
	"fmt"
	"io"
)

const (
	defaultReminderThreshold = 0.5
	defaultReminderLimit     = 5
)

func parseReminderCommand(args []string, stdin io.Reader) (parsedCommand, error) {
	if len(args) == 0 {
		return parsedCommand{}, errors.New("usage: mem4j reminder <recall|create|edit|link|unlink|delete> [flags]")
	}
	if cmd, ok, err := handleResourceHelp("reminder", args, stdin); ok {
		return cmd, err
	}
	switch args[0] {
	case "recall":
		return parseReminderRecall(args[1:], stdin)
	case "create":
		return parseReminderCreate(args[1:], stdin)
	case "edit":
		return parseReminderEdit(args[1:], stdin)
	case "link":
		return parseReminderLink(args[1:], true)
	case "unlink":
		return parseReminderLink(args[1:], false)
	case "delete":
		return parseReminderDelete(args[1:])
	default:
		return parsedCommand{}, fmt.Errorf("unknown reminder action %q", args[0])
	}
}

func parseReminderRecall(args []string, stdin io.Reader) (parsedCommand, error) {
	fs, cred := newFlagSet("reminder recall")
	gcpCred := fs.String("gcp-credentials", "", "GCP service account JSON (default ~/.config/mem4j/.secrets/gcp.json)")
	query := fs.String("query", "", "conversation context text to match against reminder descriptions")
	file := fs.String("file", "", "path to query text file, or '-' for stdin")
	threshold := fs.Float64("threshold", defaultReminderThreshold, "minimum cosine similarity score (0..1)")
	limit := fs.Int("limit", defaultReminderLimit, "maximum number of reminders to return")
	var scopes scopeList
	fs.Var(&scopes, "scope", "active scope name, repeatable")
	if err := parseFlags(fs, args); err != nil {
		return parsedCommand{}, err
	}
	if len(scopes) == 0 {
		return parsedCommand{}, errors.New("reminder recall requires at least one -scope")
	}
	if *threshold < -1 || *threshold > 1 {
		return parsedCommand{}, errors.New("reminder recall: -threshold must be between -1 and 1")
	}
	if *limit < 1 {
		return parsedCommand{}, errors.New("reminder recall: -limit must be at least 1")
	}
	text, provided, err := resolveTextInput(fs, "query", *query, *file, stdin)
	if err != nil {
		return parsedCommand{}, err
	}
	if !provided || text == "" {
		return parsedCommand{}, errors.New("reminder recall requires a non-empty -query or -file")
	}
	return parsedCommand{
		credPath: *cred,
		exec: func(ctx context.Context, s *store, stdout, stderr io.Writer) error {
			emb, err := newVertexEmbedder(ctx, *gcpCred)
			if err != nil {
				return err
			}
			vec, err := emb.Embed(ctx, text, taskTypeQuery)
			if err != nil {
				return err
			}
			reminders, err := s.recallReminders(ctx, scopes, vec, *threshold, *limit)
			if err != nil {
				return err
			}
			if len(reminders) == 0 {
				missing, err := s.missingScopes(ctx, scopes)
				if err != nil {
					return err
				}
				writeEmptyReminderNotice(stderr, missing)
				return nil
			}
			writeReminders(stdout, reminders)
			return nil
		},
	}, nil
}
