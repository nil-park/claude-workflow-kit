package main

import (
	"context"
	"errors"
	"fmt"
	"io"
)

func parseFactCreate(resource, label string, args []string, stdin io.Reader) (parsedCommand, error) {
	fs, cred := newFlagSet(resource + " create")
	gcpCred := fs.String("gcp-credentials", "", "GCP service account JSON (default ~/.config/mem4j/.secrets/gcp.json)")
	key := fs.String("key", "", "unique key (auto-generated 10-char ID if omitted)")
	title := fs.String("title", "", "short "+resource+" title")
	desc := fs.String("description", "", "embedding description (defaults to body if omitted)")
	body := fs.String("body", "", "body text")
	file := fs.String("file", "", "path to body file, or '-' for stdin")
	var occurredAt *string
	if label == "Event" {
		occurredAt = fs.String("occurred-at", "", "UTC timestamp or date (YYYY-MM-DD or RFC3339)")
	}
	var entities scopeList
	fs.Var(&entities, "entity", "entity name, repeatable")
	if err := parseFlags(fs, args); err != nil {
		return parsedCommand{}, err
	}
	t := normalizeText(*title)
	if t == "" || len(entities) == 0 {
		return parsedCommand{}, fmt.Errorf("%s create requires -title and at least one -entity", resource)
	}
	var occ string
	if label == "Event" {
		if *occurredAt == "" {
			return parsedCommand{}, errors.New("event create requires -occurred-at")
		}
		var err error
		occ, err = normalizeOccurredAt(*occurredAt)
		if err != nil {
			return parsedCommand{}, err
		}
	}
	text, provided, err := resolveTextInput(fs, "body", *body, *file, stdin)
	if err != nil {
		return parsedCommand{}, err
	}
	if !provided || text == "" {
		return parsedCommand{}, fmt.Errorf("%s create requires a non-empty -body or -file", resource)
	}
	d := text
	if flagProvided(fs, "description") {
		d = normalizeText(*desc)
		if d == "" {
			return parsedCommand{}, fmt.Errorf("%s create: -description cannot be empty", resource)
		}
	}
	k := normalizeText(*key)
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
			createdKey, err := s.createFact(ctx, label, k, t, d, occ, text, vec, entities)
			if err != nil {
				return err
			}
			fmt.Fprintf(stdout, "created %s %q\n", resource, createdKey)
			return nil
		},
	}, nil
}

func parseFactEdit(resource, label string, args []string, stdin io.Reader) (parsedCommand, error) {
	fs, cred := newFlagSet(resource + " edit")
	gcpCred := fs.String("gcp-credentials", "", "GCP service account JSON (default ~/.config/mem4j/.secrets/gcp.json)")
	key := fs.String("key", "", resource+" key")
	title := fs.String("title", "", "new "+resource+" title")
	desc := fs.String("description", "", "new embedding description (re-embedded)")
	body := fs.String("body", "", "new body text")
	file := fs.String("file", "", "path to new body file, or '-' for stdin")
	var occurredAt *string
	if label == "Event" {
		occurredAt = fs.String("occurred-at", "", "new UTC timestamp or date (YYYY-MM-DD or RFC3339)")
	}
	if err := parseFlags(fs, args); err != nil {
		return parsedCommand{}, err
	}
	k := normalizeText(*key)
	if k == "" {
		return parsedCommand{}, fmt.Errorf("%s edit requires -key", resource)
	}
	var newTitle, newDesc, newOcc, newBody *string
	if flagProvided(fs, "title") {
		t := normalizeText(*title)
		if t == "" {
			return parsedCommand{}, fmt.Errorf("%s edit: -title cannot be empty", resource)
		}
		newTitle = &t
	}
	if flagProvided(fs, "description") {
		d := normalizeText(*desc)
		if d == "" {
			return parsedCommand{}, fmt.Errorf("%s edit: -description cannot be empty", resource)
		}
		newDesc = &d
	}
	if label == "Event" && flagProvided(fs, "occurred-at") {
		normOcc, err := normalizeOccurredAt(*occurredAt)
		if err != nil {
			return parsedCommand{}, err
		}
		newOcc = &normOcc
	}
	text, bodyProvided, err := resolveTextInput(fs, "body", *body, *file, stdin)
	if err != nil {
		return parsedCommand{}, err
	}
	if bodyProvided {
		if text == "" {
			return parsedCommand{}, fmt.Errorf("%s edit: body cannot be empty", resource)
		}
		newBody = &text
	}
	if newTitle == nil && newDesc == nil && newOcc == nil && newBody == nil {
		if label == "Event" {
			return parsedCommand{}, errors.New("event edit requires at least one of -title, -description, -occurred-at, -body, or -file")
		}
		return parsedCommand{}, errors.New("statement edit requires at least one of -title, -description, -body, or -file")
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
			if err := s.editFact(ctx, label, k, newTitle, newDesc, newOcc, newBody, vec); err != nil {
				return err
			}
			fmt.Fprintf(stdout, "updated %s %q\n", resource, k)
			return nil
		},
	}, nil
}
