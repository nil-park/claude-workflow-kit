package main

import (
	"context"
	"errors"
	"fmt"
	"io"
)

func parseScopeList(args []string) (parsedCommand, error) {
	fs, cred := newFlagSet("scope list")
	if err := parseFlags(fs, args); err != nil {
		return parsedCommand{}, err
	}
	return parsedCommand{
		credPath: *cred,
		exec: func(ctx context.Context, s *store, stdout, _ io.Writer) error {
			scopes, err := s.listScopes(ctx)
			if err != nil {
				return err
			}
			writeScopes(stdout, scopes)
			return nil
		},
	}, nil
}

func parseScopeCreate(args []string) (parsedCommand, error) {
	fs, cred := newFlagSet("scope create")
	name := fs.String("name", "", "scope name")
	desc := fs.String("description", "", "short description of the scope")
	disable := fs.Bool("disable-model-invocation", false, "require explicit user invocation")
	if err := parseFlags(fs, args); err != nil {
		return parsedCommand{}, err
	}
	n, d := normalizeText(*name), normalizeText(*desc)
	if n == "" || d == "" {
		return parsedCommand{}, errors.New("scope create requires -name and -description")
	}
	return parsedCommand{
		credPath: *cred,
		exec: func(ctx context.Context, s *store, stdout, _ io.Writer) error {
			if err := s.createScope(ctx, n, d, *disable); err != nil {
				return err
			}
			fmt.Fprintf(stdout, "created scope %q\n", n)
			return nil
		},
	}, nil
}

func parseScopeEdit(args []string) (parsedCommand, error) {
	fs, cred := newFlagSet("scope edit")
	name := fs.String("name", "", "scope name")
	desc := fs.String("description", "", "new short description")
	if err := parseFlags(fs, args); err != nil {
		return parsedCommand{}, err
	}
	n, d := normalizeText(*name), normalizeText(*desc)
	if n == "" || d == "" {
		return parsedCommand{}, errors.New("scope edit requires -name and -description")
	}
	return parsedCommand{
		credPath: *cred,
		exec: func(ctx context.Context, s *store, stdout, _ io.Writer) error {
			if err := s.editScope(ctx, n, d); err != nil {
				return err
			}
			fmt.Fprintf(stdout, "updated scope %q\n", n)
			return nil
		},
	}, nil
}

func parseScopeDelete(args []string) (parsedCommand, error) {
	fs, cred := newFlagSet("scope delete")
	name := fs.String("name", "", "scope name")
	if err := parseFlags(fs, args); err != nil {
		return parsedCommand{}, err
	}
	n := normalizeText(*name)
	if n == "" {
		return parsedCommand{}, errors.New("scope delete requires -name")
	}
	return parsedCommand{
		credPath: *cred,
		exec: func(ctx context.Context, s *store, stdout, _ io.Writer) error {
			if err := s.deleteScope(ctx, n); err != nil {
				return err
			}
			fmt.Fprintf(stdout, "deleted scope %q\n", n)
			return nil
		},
	}, nil
}
