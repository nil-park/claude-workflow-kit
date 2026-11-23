package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
)

type parsedCommand struct {
	credPath string
	helpText string
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
	cmd, err := dispatchCommand(args, stdin)
	var h helpRequest
	if errors.As(err, &h) {
		return parsedCommand{helpText: h.text}, nil
	}
	return cmd, err
}

func dispatchCommand(args []string, stdin io.Reader) (parsedCommand, error) {
	if len(args) == 0 {
		return parsedCommand{}, errors.New("usage: mem4j <scope|instruction|reminder|entity|statement|event|graph> <action> [flags]")
	}
	switch args[0] {
	case "help":
		return parseHelpCommand(args[1:], stdin)
	case "-h", "--help", "-help":
		return parsedCommand{helpText: rootHelp()}, nil
	case "scope":
		return parseScopeCommand(args[1:], stdin)
	case "instruction":
		return parseInstructionCommand(args[1:], stdin)
	case "reminder":
		return parseReminderCommand(args[1:], stdin)
	case "entity":
		return parseEntityCommand(args[1:], stdin)
	case "statement":
		return parseFactCommand("statement", "Statement", args[1:], stdin)
	case "event":
		return parseFactCommand("event", "Event", args[1:], stdin)
	case "graph":
		return parseGraphCommand(args[1:], stdin)
	default:
		return parsedCommand{}, fmt.Errorf("unknown resource %q (expected scope, instruction, reminder, entity, statement, event, or graph)", args[0])
	}
}

func parseScopeCommand(args []string, stdin io.Reader) (parsedCommand, error) {
	if len(args) == 0 {
		return parsedCommand{}, errors.New("usage: mem4j scope <list|create|edit|link|unlink|neighbors|delete> [flags]")
	}
	if cmd, ok, err := handleResourceHelp("scope", args, stdin); ok {
		return cmd, err
	}
	switch args[0] {
	case "list":
		return parseScopeList(args[1:])
	case "create":
		return parseScopeCreate(args[1:])
	case "edit":
		return parseScopeEdit(args[1:])
	case "link":
		return parseHubLink("Scope", "scope link", args[1:], true)
	case "unlink":
		return parseHubLink("Scope", "scope unlink", args[1:], false)
	case "neighbors":
		return parseHubNeighbors("Scope", "scope neighbors", args[1:])
	case "delete":
		return parseScopeDelete(args[1:])
	default:
		return parsedCommand{}, fmt.Errorf("unknown scope action %q", args[0])
	}
}

func parseInstructionCommand(args []string, stdin io.Reader) (parsedCommand, error) {
	if len(args) == 0 {
		return parsedCommand{}, errors.New("usage: mem4j instruction <recall|create|edit|link|unlink|delete> [flags]")
	}
	if cmd, ok, err := handleResourceHelp("instruction", args, stdin); ok {
		return cmd, err
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
	cred := fs.String("credentials", "", "Aura credential file (KEY=VALUE; default ~/.config/mem4j/.secrets/neo4j.txt)")
	return fs, cred
}

func parseFlags(fs *flag.FlagSet, args []string) error {
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return helpRequest{text: formatFlagHelp(fs)}
		}
		return err
	}
	if fs.NArg() > 0 {
		return fmt.Errorf("unexpected arguments: %v", fs.Args())
	}
	return nil
}

func flagProvided(fs *flag.FlagSet, name string) bool {
	found := false
	fs.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}

func resolveTextInput(fs *flag.FlagSet, textFlag, text, file string, stdin io.Reader) (string, bool, error) {
	hasText := flagProvided(fs, textFlag)
	hasFile := flagProvided(fs, "file")
	if hasText && hasFile {
		return "", false, fmt.Errorf("specify only one of -%s or -file", textFlag)
	}
	if hasText {
		return normalizeText(text), true, nil
	}
	if !hasFile {
		return "", false, nil
	}
	var raw []byte
	var err error
	if file == "-" {
		raw, err = io.ReadAll(stdin)
	} else {
		raw, err = os.ReadFile(file)
	}
	if err != nil {
		return "", false, err
	}
	return normalizeText(string(raw)), true, nil
}
