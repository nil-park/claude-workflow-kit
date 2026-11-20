package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
)

func parseInstructionCreate(args []string, stdin io.Reader) (parsedCommand, error) {
	fs, cred := newFlagSet("instruction-create")
	key := fs.String("key", "", "unique instruction key")
	title := fs.String("title", "", "short instruction title")
	body := fs.String("body", "", "instruction body text")
	file := fs.String("file", "", "path to body file, or '-' for stdin")
	var scopes scopeList
	fs.Var(&scopes, "scope", "scope name, repeatable")
	if err := parseFlags(fs, args); err != nil {
		return parsedCommand{}, err
	}
	k, t := normalizeText(*key), normalizeText(*title)
	if k == "" || t == "" || len(scopes) == 0 {
		return parsedCommand{}, errors.New("instruction-create requires -key, -title, and at least one -scope")
	}
	text, provided, err := resolveBody(fs, *body, *file, stdin)
	if err != nil {
		return parsedCommand{}, err
	}
	if !provided || text == "" {
		return parsedCommand{}, errors.New("instruction-create requires a non-empty -body or -file")
	}
	return parsedCommand{
		credPath: *cred,
		exec: func(ctx context.Context, s *store, stdout io.Writer) error {
			if err := s.createInstruction(ctx, k, t, text, scopes); err != nil {
				return err
			}
			fmt.Fprintf(stdout, "created instruction %q\n", k)
			return nil
		},
	}, nil
}

func parseInstructionEdit(args []string, stdin io.Reader) (parsedCommand, error) {
	fs, cred := newFlagSet("instruction-edit")
	key := fs.String("key", "", "instruction key")
	title := fs.String("title", "", "new instruction title")
	body := fs.String("body", "", "new instruction body text")
	file := fs.String("file", "", "path to new body file, or '-' for stdin")
	if err := parseFlags(fs, args); err != nil {
		return parsedCommand{}, err
	}
	k := normalizeText(*key)
	if k == "" {
		return parsedCommand{}, errors.New("instruction-edit requires -key")
	}
	var newTitle *string
	if flagProvided(fs, "title") {
		t := normalizeText(*title)
		if t == "" {
			return parsedCommand{}, errors.New("instruction-edit: -title cannot be empty")
		}
		newTitle = &t
	}
	text, bodyProvided, err := resolveBody(fs, *body, *file, stdin)
	if err != nil {
		return parsedCommand{}, err
	}
	var newBody *string
	if bodyProvided {
		if text == "" {
			return parsedCommand{}, errors.New("instruction-edit: body cannot be empty")
		}
		newBody = &text
	}
	if newTitle == nil && newBody == nil {
		return parsedCommand{}, errors.New("instruction-edit requires at least one of -title, -body, or -file")
	}
	return parsedCommand{
		credPath: *cred,
		exec: func(ctx context.Context, s *store, stdout io.Writer) error {
			if err := s.editInstruction(ctx, k, newTitle, newBody); err != nil {
				return err
			}
			fmt.Fprintf(stdout, "updated instruction %q\n", k)
			return nil
		},
	}, nil
}

func parseInstructionLink(args []string, link bool) (parsedCommand, error) {
	cmdName := "instruction-unlink"
	if link {
		cmdName = "instruction-link"
	}
	fs, cred := newFlagSet(cmdName)
	key := fs.String("key", "", "instruction key")
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
		exec: func(ctx context.Context, s *store, stdout io.Writer) error {
			if link {
				if err := s.linkInstruction(ctx, k, sc); err != nil {
					return err
				}
				fmt.Fprintf(stdout, "linked %q to %q\n", k, sc)
				return nil
			}
			if err := s.unlinkInstruction(ctx, k, sc); err != nil {
				return err
			}
			fmt.Fprintf(stdout, "unlinked %q from %q\n", k, sc)
			return nil
		},
	}, nil
}

func parseInstructionDelete(args []string) (parsedCommand, error) {
	fs, cred := newFlagSet("instruction-delete")
	key := fs.String("key", "", "instruction key")
	if err := parseFlags(fs, args); err != nil {
		return parsedCommand{}, err
	}
	k := normalizeText(*key)
	if k == "" {
		return parsedCommand{}, errors.New("instruction-delete requires -key")
	}
	return parsedCommand{
		credPath: *cred,
		exec: func(ctx context.Context, s *store, stdout io.Writer) error {
			if err := s.deleteInstruction(ctx, k); err != nil {
				return err
			}
			fmt.Fprintf(stdout, "deleted instruction %q\n", k)
			return nil
		},
	}, nil
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

func resolveBody(fs *flag.FlagSet, body, file string, stdin io.Reader) (string, bool, error) {
	hasBody := flagProvided(fs, "body")
	hasFile := flagProvided(fs, "file")
	if hasBody && hasFile {
		return "", false, errors.New("specify only one of -body or -file")
	}
	if hasBody {
		return normalizeText(body), true, nil
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
