package main

import (
	"context"
	"errors"
	"fmt"
	"io"
)

func parseEntityCommand(args []string, stdin io.Reader) (parsedCommand, error) {
	if len(args) == 0 {
		return parsedCommand{}, errors.New("usage: mem4j entity <list|recall|create|edit|link|unlink|neighbors|delete> [flags]")
	}
	if cmd, ok, err := handleResourceHelp("entity", args, stdin); ok {
		return cmd, err
	}
	switch args[0] {
	case "list":
		return parseEntityList(args[1:])
	case "recall":
		return parseEntityRecall(args[1:], stdin)
	case "create":
		return parseEntityCreate(args[1:])
	case "edit":
		return parseEntityEdit(args[1:])
	case "link":
		return parseHubLink("Entity", "entity link", args[1:], true)
	case "unlink":
		return parseHubLink("Entity", "entity unlink", args[1:], false)
	case "neighbors":
		return parseHubNeighbors("Entity", "entity neighbors", args[1:])
	case "delete":
		return parseEntityDelete(args[1:])
	default:
		return parsedCommand{}, fmt.Errorf("unknown entity action %q", args[0])
	}
}

func parseEntityList(args []string) (parsedCommand, error) {
	fs, cred := newFlagSet("entity list")
	if err := parseFlags(fs, args); err != nil {
		return parsedCommand{}, err
	}
	return parsedCommand{
		credPath: *cred,
		exec: func(ctx context.Context, s *store, stdout, _ io.Writer) error {
			entities, err := s.listEntities(ctx)
			if err != nil {
				return err
			}
			writeEntities(stdout, entities, false)
			return nil
		},
	}, nil
}

func parseEntityRecall(args []string, stdin io.Reader) (parsedCommand, error) {
	fs, cred := newFlagSet("entity recall")
	gcpCred := fs.String("gcp-credentials", "", "GCP service account JSON (default ~/.config/mem4j/.secrets/gcp.json)")
	query := fs.String("query", "", "query text to match against entity embeddings")
	file := fs.String("file", "", "path to query text file, or '-' for stdin")
	threshold := fs.Float64("threshold", defaultReminderThreshold, "minimum cosine similarity score (-1..1)")
	limit := fs.Int("limit", defaultReminderLimit, "maximum number of entities to return")
	if err := parseFlags(fs, args); err != nil {
		return parsedCommand{}, err
	}
	if *threshold < -1 || *threshold > 1 {
		return parsedCommand{}, errors.New("entity recall: -threshold must be between -1 and 1")
	}
	if *limit < 1 {
		return parsedCommand{}, errors.New("entity recall: -limit must be at least 1")
	}
	text, provided, err := resolveTextInput(fs, "query", *query, *file, stdin)
	if err != nil {
		return parsedCommand{}, err
	}
	if !provided || text == "" {
		return parsedCommand{}, errors.New("entity recall requires a non-empty -query or -file")
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
			entities, err := s.recallEntities(ctx, vec, *threshold, *limit)
			if err != nil {
				return err
			}
			if len(entities) == 0 {
				fmt.Fprintln(stderr, "mem4j: no entities match the query")
				return nil
			}
			writeEntities(stdout, entities, true)
			return nil
		},
	}, nil
}

func parseEntityCreate(args []string) (parsedCommand, error) {
	fs, cred := newFlagSet("entity create")
	gcpCred := fs.String("gcp-credentials", "", "GCP service account JSON (default ~/.config/mem4j/.secrets/gcp.json)")
	name := fs.String("name", "", "unique entity name")
	desc := fs.String("description", "", "short entity description and aliases")
	if err := parseFlags(fs, args); err != nil {
		return parsedCommand{}, err
	}
	n, d := normalizeText(*name), normalizeText(*desc)
	if n == "" || d == "" {
		return parsedCommand{}, errors.New("entity create requires -name and -description")
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
			if err := s.createEntity(ctx, n, d, vec); err != nil {
				return err
			}
			fmt.Fprintf(stdout, "created entity %q\n", n)
			return nil
		},
	}, nil
}

func parseEntityEdit(args []string) (parsedCommand, error) {
	fs, cred := newFlagSet("entity edit")
	gcpCred := fs.String("gcp-credentials", "", "GCP service account JSON (default ~/.config/mem4j/.secrets/gcp.json)")
	name := fs.String("name", "", "entity name")
	newNameFlag := fs.String("new-name", "", "new entity name")
	desc := fs.String("description", "", "new entity description (re-embedded)")
	if err := parseFlags(fs, args); err != nil {
		return parsedCommand{}, err
	}
	n := normalizeText(*name)
	if n == "" {
		return parsedCommand{}, errors.New("entity edit requires -name")
	}
	var newName, newDesc *string
	if flagProvided(fs, "new-name") {
		nn := normalizeText(*newNameFlag)
		if nn == "" {
			return parsedCommand{}, errors.New("entity edit: -new-name cannot be empty")
		}
		if nn == n {
			return parsedCommand{}, errors.New("entity edit: -new-name must differ from -name")
		}
		newName = &nn
	}
	if flagProvided(fs, "description") {
		d := normalizeText(*desc)
		if d == "" {
			return parsedCommand{}, errors.New("entity edit: -description cannot be empty")
		}
		newDesc = &d
	}
	if newName == nil && newDesc == nil {
		return parsedCommand{}, errors.New("entity edit requires at least one of -new-name or -description")
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
			if err := s.editEntity(ctx, n, newName, newDesc, vec); err != nil {
				return err
			}
			switch {
			case newName != nil && newDesc != nil:
				fmt.Fprintf(stdout, "renamed entity %q to %q and updated description\n", n, *newName)
			case newName != nil:
				fmt.Fprintf(stdout, "renamed entity %q to %q\n", n, *newName)
			default:
				fmt.Fprintf(stdout, "updated entity %q\n", n)
			}
			return nil
		},
	}, nil
}

func parseEntityDelete(args []string) (parsedCommand, error) {
	fs, cred := newFlagSet("entity delete")
	name := fs.String("name", "", "entity name")
	if err := parseFlags(fs, args); err != nil {
		return parsedCommand{}, err
	}
	n := normalizeText(*name)
	if n == "" {
		return parsedCommand{}, errors.New("entity delete requires -name")
	}
	return parsedCommand{
		credPath: *cred,
		exec: func(ctx context.Context, s *store, stdout, _ io.Writer) error {
			if err := s.deleteEntity(ctx, n); err != nil {
				return err
			}
			fmt.Fprintf(stdout, "deleted entity %q\n", n)
			return nil
		},
	}, nil
}
