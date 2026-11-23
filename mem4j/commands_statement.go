package main

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"
)

func parseFactCommand(resource, label string, args []string, stdin io.Reader) (parsedCommand, error) {
	if len(args) == 0 {
		return parsedCommand{}, fmt.Errorf("usage: mem4j %s <list|recall|create|edit|link|unlink|delete> [flags]", resource)
	}
	if cmd, ok, err := handleResourceHelp(resource, args, stdin); ok {
		return cmd, err
	}
	switch args[0] {
	case "list":
		return parseFactList(resource, label, args[1:])
	case "recall":
		return parseFactRecall(resource, label, args[1:], stdin)
	case "create":
		return parseFactCreate(resource, label, args[1:], stdin)
	case "edit":
		return parseFactEdit(resource, label, args[1:], stdin)
	case "link":
		return parseFactLink(resource, label, args[1:], true)
	case "unlink":
		return parseFactLink(resource, label, args[1:], false)
	case "delete":
		return parseFactDelete(resource, label, args[1:])
	default:
		return parsedCommand{}, fmt.Errorf("unknown %s action %q", resource, args[0])
	}
}

func normalizeOccurredAt(raw string) (string, error) {
	v := strings.TrimSpace(raw)
	if _, err := time.Parse(time.DateOnly, v); err == nil {
		return v, nil
	}
	if t, err := time.Parse(time.RFC3339Nano, v); err == nil {
		return t.UTC().Format(time.RFC3339), nil
	}
	return "", fmt.Errorf("invalid -occurred-at %q (expected YYYY-MM-DD or RFC3339 timestamp)", raw)
}

func parseFactList(resource, label string, args []string) (parsedCommand, error) {
	fs, cred := newFlagSet(resource + " list")
	entity := fs.String("entity", "", "entity name")
	if err := parseFlags(fs, args); err != nil {
		return parsedCommand{}, err
	}
	e := normalizeText(*entity)
	if e == "" {
		return parsedCommand{}, fmt.Errorf("%s list requires -entity", resource)
	}
	return parsedCommand{
		credPath: *cred,
		exec: func(ctx context.Context, s *store, stdout, _ io.Writer) error {
			if label == "Event" {
				events, err := s.listEvents(ctx, e)
				if err != nil {
					return err
				}
				writeEvents(stdout, events, false)
				return nil
			}
			stmts, err := s.listStatements(ctx, e)
			if err != nil {
				return err
			}
			writeStatements(stdout, stmts, false)
			return nil
		},
	}, nil
}

func parseFactRecall(resource, label string, args []string, stdin io.Reader) (parsedCommand, error) {
	fs, cred := newFlagSet(resource + " recall")
	gcpCred := fs.String("gcp-credentials", "", "GCP service account JSON (default ~/.config/mem4j/.secrets/gcp.json)")
	query := fs.String("query", "", "query text to match against embeddings")
	file := fs.String("file", "", "path to query text file, or '-' for stdin")
	threshold := fs.Float64("threshold", defaultReminderThreshold, "minimum cosine similarity score (-1..1)")
	limit := fs.Int("limit", defaultReminderLimit, "maximum number of results to return")
	var entities scopeList
	fs.Var(&entities, "entity", "optional entity name filter, repeatable")
	if err := parseFlags(fs, args); err != nil {
		return parsedCommand{}, err
	}
	if *threshold < -1 || *threshold > 1 {
		return parsedCommand{}, fmt.Errorf("%s recall: -threshold must be between -1 and 1", resource)
	}
	if *limit < 1 {
		return parsedCommand{}, fmt.Errorf("%s recall: -limit must be at least 1", resource)
	}
	text, provided, err := resolveTextInput(fs, "query", *query, *file, stdin)
	if err != nil {
		return parsedCommand{}, err
	}
	if !provided || text == "" {
		return parsedCommand{}, fmt.Errorf("%s recall requires a non-empty -query or -file", resource)
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
			if label == "Event" {
				events, err := s.recallEvents(ctx, entities, vec, *threshold, *limit)
				if err != nil {
					return err
				}
				if len(events) == 0 {
					fmt.Fprintln(stderr, "mem4j: no events match the query")
					return nil
				}
				writeEvents(stdout, events, true)
				return nil
			}
			stmts, err := s.recallStatements(ctx, entities, vec, *threshold, *limit)
			if err != nil {
				return err
			}
			if len(stmts) == 0 {
				fmt.Fprintln(stderr, "mem4j: no statements match the query")
				return nil
			}
			writeStatements(stdout, stmts, true)
			return nil
		},
	}, nil
}

func parseFactLink(resource, label string, args []string, link bool) (parsedCommand, error) {
	cmdName := resource + " unlink"
	if link {
		cmdName = resource + " link"
	}
	fs, cred := newFlagSet(cmdName)
	key := fs.String("key", "", resource+" key")
	entity := fs.String("entity", "", "entity name")
	if err := parseFlags(fs, args); err != nil {
		return parsedCommand{}, err
	}
	k, e := normalizeText(*key), normalizeText(*entity)
	if k == "" || e == "" {
		return parsedCommand{}, fmt.Errorf("%s requires -key and -entity", cmdName)
	}
	return parsedCommand{
		credPath: *cred,
		exec: func(ctx context.Context, s *store, stdout, _ io.Writer) error {
			if err := s.linkFact(ctx, label, k, e, link); err != nil {
				return err
			}
			if link {
				fmt.Fprintf(stdout, "linked %q to %q\n", k, e)
				return nil
			}
			fmt.Fprintf(stdout, "unlinked %q from %q\n", k, e)
			return nil
		},
	}, nil
}

func parseFactDelete(resource, label string, args []string) (parsedCommand, error) {
	fs, cred := newFlagSet(resource + " delete")
	key := fs.String("key", "", resource+" key")
	if err := parseFlags(fs, args); err != nil {
		return parsedCommand{}, err
	}
	k := normalizeText(*key)
	if k == "" {
		return parsedCommand{}, fmt.Errorf("%s delete requires -key", resource)
	}
	return parsedCommand{
		credPath: *cred,
		exec: func(ctx context.Context, s *store, stdout, _ io.Writer) error {
			if err := s.deleteFact(ctx, label, k); err != nil {
				return err
			}
			fmt.Fprintf(stdout, "deleted %s %q\n", resource, k)
			return nil
		},
	}, nil
}
