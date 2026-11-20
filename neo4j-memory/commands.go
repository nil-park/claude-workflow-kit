package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"strings"
)

const defaultCredentials = "../.secrets/neo4j.txt"

type parsedCommand struct {
	credPath string
	exec     func(ctx context.Context, s *store, stdout io.Writer) error
}

type scopeList []string

func (s *scopeList) String() string { return strings.Join(*s, ", ") }

func (s *scopeList) Set(v string) error {
	*s = append(*s, normalizeText(v))
	return nil
}

// parseCommand validates the subcommand, flags, and any body input before
// credentials are read or Neo4j is contacted.
func parseCommand(args []string, stdin io.Reader) (parsedCommand, error) {
	if len(args) == 0 {
		return parsedCommand{}, errors.New("usage: neo4j-memory <scope-create|scope-edit|scope-delete|instruction-create|instruction-edit|instruction-link|instruction-unlink|instruction-delete|merge> [flags]")
	}
	switch args[0] {
	case "scope-create":
		return parseScopeCreate(args[1:])
	case "scope-edit":
		return parseScopeEdit(args[1:])
	case "scope-delete":
		return parseScopeDelete(args[1:])
	case "instruction-create":
		return parseInstructionCreate(args[1:], stdin)
	case "instruction-edit":
		return parseInstructionEdit(args[1:], stdin)
	case "instruction-link":
		return parseInstructionLink(args[1:], true)
	case "instruction-unlink":
		return parseInstructionLink(args[1:], false)
	case "instruction-delete":
		return parseInstructionDelete(args[1:])
	case "merge":
		return parseMerge(args[1:])
	default:
		return parsedCommand{}, fmt.Errorf("unknown command %q", args[0])
	}
}

func newFlagSet(name string) (*flag.FlagSet, *string) {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	cred := fs.String("credentials", defaultCredentials, "Aura credential file (KEY=VALUE)")
	return fs, cred
}

func parseFlags(fs *flag.FlagSet, args []string) error {
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() > 0 {
		return fmt.Errorf("unexpected arguments: %v", fs.Args())
	}
	return nil
}

func parseScopeCreate(args []string) (parsedCommand, error) {
	fs, cred := newFlagSet("scope-create")
	name := fs.String("name", "", "scope name")
	desc := fs.String("description", "", "short description of the scope")
	disable := fs.Bool("disable-model-invocation", false, "require explicit user invocation")
	if err := parseFlags(fs, args); err != nil {
		return parsedCommand{}, err
	}
	n, d := normalizeText(*name), normalizeText(*desc)
	if n == "" || d == "" {
		return parsedCommand{}, errors.New("scope-create requires -name and -description")
	}
	return parsedCommand{
		credPath: *cred,
		exec: func(ctx context.Context, s *store, stdout io.Writer) error {
			if err := s.createScope(ctx, n, d, *disable); err != nil {
				return err
			}
			fmt.Fprintf(stdout, "created scope %q\n", n)
			return nil
		},
	}, nil
}

func parseScopeEdit(args []string) (parsedCommand, error) {
	fs, cred := newFlagSet("scope-edit")
	name := fs.String("name", "", "scope name")
	desc := fs.String("description", "", "new short description")
	if err := parseFlags(fs, args); err != nil {
		return parsedCommand{}, err
	}
	n, d := normalizeText(*name), normalizeText(*desc)
	if n == "" || d == "" {
		return parsedCommand{}, errors.New("scope-edit requires -name and -description")
	}
	return parsedCommand{
		credPath: *cred,
		exec: func(ctx context.Context, s *store, stdout io.Writer) error {
			if err := s.editScope(ctx, n, d); err != nil {
				return err
			}
			fmt.Fprintf(stdout, "updated scope %q\n", n)
			return nil
		},
	}, nil
}

func parseScopeDelete(args []string) (parsedCommand, error) {
	fs, cred := newFlagSet("scope-delete")
	name := fs.String("name", "", "scope name")
	if err := parseFlags(fs, args); err != nil {
		return parsedCommand{}, err
	}
	n := normalizeText(*name)
	if n == "" {
		return parsedCommand{}, errors.New("scope-delete requires -name")
	}
	return parsedCommand{
		credPath: *cred,
		exec: func(ctx context.Context, s *store, stdout io.Writer) error {
			if err := s.deleteScope(ctx, n); err != nil {
				return err
			}
			fmt.Fprintf(stdout, "deleted scope %q\n", n)
			return nil
		},
	}, nil
}

func parseMerge(args []string) (parsedCommand, error) {
	fs, cred := newFlagSet("merge")
	var scopes scopeList
	fs.Var(&scopes, "scope", "active scope name, repeatable")
	if err := parseFlags(fs, args); err != nil {
		return parsedCommand{}, err
	}
	if len(scopes) == 0 {
		return parsedCommand{}, errors.New("merge requires at least one -scope")
	}
	return parsedCommand{
		credPath: *cred,
		exec: func(ctx context.Context, s *store, stdout io.Writer) error {
			candidates, err := s.fetchCandidates(ctx, scopes)
			if err != nil {
				return err
			}
			writeMerged(stdout, Merge(candidates, scopes))
			return nil
		},
	}, nil
}
