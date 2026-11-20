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
	exec     func(ctx context.Context, s *store, stdout, stderr io.Writer) error
}

type scopeList []string

func (s *scopeList) String() string { return strings.Join(*s, ", ") }

func (s *scopeList) Set(v string) error {
	*s = append(*s, normalizeText(v))
	return nil
}

// parseCommand validates the resource, action, flags, and any body input
// before credentials are read or Neo4j is contacted.
func parseCommand(args []string, stdin io.Reader) (parsedCommand, error) {
	if len(args) == 0 {
		return parsedCommand{}, errors.New("usage: neo4j-memory <scope|instruction> <action> [flags]")
	}
	switch args[0] {
	case "scope":
		return parseScopeCommand(args[1:])
	case "instruction":
		return parseInstructionCommand(args[1:], stdin)
	default:
		return parsedCommand{}, fmt.Errorf("unknown resource %q (expected scope or instruction)", args[0])
	}
}

func parseScopeCommand(args []string) (parsedCommand, error) {
	if len(args) == 0 {
		return parsedCommand{}, errors.New("usage: neo4j-memory scope <list|create|edit|delete> [flags]")
	}
	switch args[0] {
	case "list":
		return parseScopeList(args[1:])
	case "create":
		return parseScopeCreate(args[1:])
	case "edit":
		return parseScopeEdit(args[1:])
	case "delete":
		return parseScopeDelete(args[1:])
	default:
		return parsedCommand{}, fmt.Errorf("unknown scope action %q", args[0])
	}
}

func parseInstructionCommand(args []string, stdin io.Reader) (parsedCommand, error) {
	if len(args) == 0 {
		return parsedCommand{}, errors.New("usage: neo4j-memory instruction <recall|create|edit|link|unlink|delete> [flags]")
	}
	switch args[0] {
	case "recall":
		return parseInstructionRecall(args[1:])
	case "create":
		return parseInstructionCreate(args[1:], stdin)
	case "edit":
		return parseInstructionEdit(args[1:], stdin)
	case "link":
		return parseInstructionLink(args[1:], true)
	case "unlink":
		return parseInstructionLink(args[1:], false)
	case "delete":
		return parseInstructionDelete(args[1:])
	default:
		return parsedCommand{}, fmt.Errorf("unknown instruction action %q", args[0])
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

func parseInstructionRecall(args []string) (parsedCommand, error) {
	fs, cred := newFlagSet("instruction recall")
	var scopes scopeList
	fs.Var(&scopes, "scope", "active scope name, repeatable")
	if err := parseFlags(fs, args); err != nil {
		return parsedCommand{}, err
	}
	if len(scopes) == 0 {
		return parsedCommand{}, errors.New("instruction recall requires at least one -scope")
	}
	return parsedCommand{
		credPath: *cred,
		exec: func(ctx context.Context, s *store, stdout, stderr io.Writer) error {
			candidates, err := s.fetchCandidates(ctx, scopes)
			if err != nil {
				return err
			}
			merged := Merge(candidates, scopes)
			if len(merged) == 0 {
				missing, err := s.missingScopes(ctx, scopes)
				if err != nil {
					return err
				}
				writeEmptyMergeNotice(stderr, missing)
				return nil
			}
			writeMerged(stdout, merged)
			return nil
		},
	}, nil
}
