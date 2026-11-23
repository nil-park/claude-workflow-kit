package main

import (
	"context"
	"errors"
	"fmt"
	"io"
)

func parseReminderCreate(args []string, stdin io.Reader) (parsedCommand, error) {
	fs, cred := newFlagSet("reminder create")
	gcpCred := fs.String("gcp-credentials", "", "GCP service account JSON (default ~/.config/mem4j/.secrets/gcp.json)")
	key := fs.String("key", "", "unique reminder key (auto-generated 10-char ID if omitted)")
	title := fs.String("title", "", "short reminder title")
	desc := fs.String("description", "", "trigger description used for embedding")
	body := fs.String("body", "", "reminder body text")
	file := fs.String("file", "", "path to body file, or '-' for stdin")
	var scopes scopeList
	fs.Var(&scopes, "scope", "scope name, repeatable")
	if err := parseFlags(fs, args); err != nil {
		return parsedCommand{}, err
	}
	k, t, d := normalizeText(*key), normalizeText(*title), normalizeText(*desc)
	if t == "" || d == "" || len(scopes) == 0 {
		return parsedCommand{}, errors.New("reminder create requires -title, -description, and at least one -scope")
	}
	text, provided, err := resolveTextInput(fs, "body", *body, *file, stdin)
	if err != nil {
		return parsedCommand{}, err
	}
	if !provided || text == "" {
		return parsedCommand{}, errors.New("reminder create requires a non-empty -body or -file")
	}
	return parsedCommand{
		credPath: *cred,
		exec: func(ctx context.Context, s *store, stdout, _ io.Writer) error {
			emb, err := newVertexEmbedder(ctx, *gcpCred)
			if err != nil {
				return err
			}
			vec, err := emb.Embed(ctx, d, taskTypeDocument)
			if err != nil {
				return err
			}
			createdKey, err := s.createReminder(ctx, k, t, d, text, vec, scopes)
			if err != nil {
				return err
			}
			fmt.Fprintf(stdout, "created reminder %q\n", createdKey)
			return nil
		},
	}, nil
}

func parseReminderEdit(args []string, stdin io.Reader) (parsedCommand, error) {
	fs, cred := newFlagSet("reminder edit")
	gcpCred := fs.String("gcp-credentials", "", "GCP service account JSON (default ~/.config/mem4j/.secrets/gcp.json)")
	key := fs.String("key", "", "reminder key")
	title := fs.String("title", "", "new reminder title")
	desc := fs.String("description", "", "new trigger description")
	body := fs.String("body", "", "new reminder body text")
	file := fs.String("file", "", "path to new body file, or '-' for stdin")
	if err := parseFlags(fs, args); err != nil {
		return parsedCommand{}, err
	}
	k := normalizeText(*key)
	if k == "" {
		return parsedCommand{}, errors.New("reminder edit requires -key")
	}
	var newTitle, newDesc, newBody *string
	if flagProvided(fs, "title") {
		t := normalizeText(*title)
		if t == "" {
			return parsedCommand{}, errors.New("reminder edit: -title cannot be empty")
		}
		newTitle = &t
	}
	if flagProvided(fs, "description") {
		d := normalizeText(*desc)
		if d == "" {
			return parsedCommand{}, errors.New("reminder edit: -description cannot be empty")
		}
		newDesc = &d
	}
	text, bodyProvided, err := resolveTextInput(fs, "body", *body, *file, stdin)
	if err != nil {
		return parsedCommand{}, err
	}
	if bodyProvided {
		if text == "" {
			return parsedCommand{}, errors.New("reminder edit: body cannot be empty")
		}
		newBody = &text
	}
	if newTitle == nil && newDesc == nil && newBody == nil {
		return parsedCommand{}, errors.New("reminder edit requires at least one of -title, -description, -body, or -file")
	}
	return parsedCommand{
		credPath: *cred,
		exec: func(ctx context.Context, s *store, stdout, _ io.Writer) error {
			var vec []float64
			if newDesc != nil {
				emb, err := newVertexEmbedder(ctx, *gcpCred)
				if err != nil {
					return err
				}
				vec, err = emb.Embed(ctx, *newDesc, taskTypeDocument)
				if err != nil {
					return err
				}
			}
			if err := s.editReminder(ctx, k, newTitle, newDesc, newBody, vec); err != nil {
				return err
			}
			fmt.Fprintf(stdout, "updated reminder %q\n", k)
			return nil
		},
	}, nil
}

func parseReminderLink(args []string, link bool) (parsedCommand, error) {
	cmdName := "reminder unlink"
	if link {
		cmdName = "reminder link"
	}
	fs, cred := newFlagSet(cmdName)
	key := fs.String("key", "", "reminder key")
	scope := fs.String("scope", "", "scope name")
	if err := parseFlags(fs, args); err != nil {
		return parsedCommand{}, err
	}
	k, sc := normalizeText(*key), normalizeText(*scope)
	if k == "" || sc == "" {
		return parsedCommand{}, fmt.Errorf("%s requires -key and -scope", cmdName)
	}
	return parsedCommand{
		credPath: *cred,
		exec: func(ctx context.Context, s *store, stdout, _ io.Writer) error {
			if link {
				if err := s.linkReminder(ctx, k, sc); err != nil {
					return err
				}
				fmt.Fprintf(stdout, "linked %q to %q\n", k, sc)
				return nil
			}
			if err := s.unlinkReminder(ctx, k, sc); err != nil {
				return err
			}
			fmt.Fprintf(stdout, "unlinked %q from %q\n", k, sc)
			return nil
		},
	}, nil
}

func parseReminderDelete(args []string) (parsedCommand, error) {
	fs, cred := newFlagSet("reminder delete")
	key := fs.String("key", "", "reminder key")
	if err := parseFlags(fs, args); err != nil {
		return parsedCommand{}, err
	}
	k := normalizeText(*key)
	if k == "" {
		return parsedCommand{}, errors.New("reminder delete requires -key")
	}
	return parsedCommand{
		credPath: *cred,
		exec: func(ctx context.Context, s *store, stdout, _ io.Writer) error {
			if err := s.deleteReminder(ctx, k); err != nil {
				return err
			}
			fmt.Fprintf(stdout, "deleted reminder %q\n", k)
			return nil
		},
	}, nil
}
