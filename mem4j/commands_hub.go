package main

import (
	"context"
	"fmt"
	"io"
)

const maxNeighborDepth = 3

func parseHubLink(label, cmdName string, args []string, link bool) (parsedCommand, error) {
	fs, cred := newFlagSet(cmdName)
	from := fs.String("from", "", "source hub name")
	to := fs.String("to", "", "target hub name")
	relType := fs.String("type", "", "relationship type in UPPER_SNAKE_CASE (e.g. USES, DEPENDS_ON)")
	if err := parseFlags(fs, args); err != nil {
		return parsedCommand{}, err
	}
	f, t, rt := normalizeText(*from), normalizeText(*to), normalizeText(*relType)
	if f == "" || t == "" || rt == "" {
		return parsedCommand{}, fmt.Errorf("%s requires -from, -to, and -type", cmdName)
	}
	if err := validateRelType(rt); err != nil {
		return parsedCommand{}, err
	}
	if f == t {
		return parsedCommand{}, fmt.Errorf("%s: -from and -to cannot be the same", cmdName)
	}
	return parsedCommand{
		credPath: *cred,
		exec: func(ctx context.Context, s *store, stdout, _ io.Writer) error {
			if err := s.linkHubs(ctx, label, f, t, rt, link); err != nil {
				return err
			}
			verb := "unlinked"
			if link {
				verb = "linked"
			}
			fmt.Fprintf(stdout, "%s (%q)-[:%s]->(%q)\n", verb, f, rt, t)
			return nil
		},
	}, nil
}

func parseHubNeighbors(label, cmdName string, args []string) (parsedCommand, error) {
	fs, cred := newFlagSet(cmdName)
	name := fs.String("name", "", "starting hub name")
	depth := fs.Int("depth", 1, "traversal depth (1..3)")
	if err := parseFlags(fs, args); err != nil {
		return parsedCommand{}, err
	}
	n := normalizeText(*name)
	if n == "" {
		return parsedCommand{}, fmt.Errorf("%s requires -name", cmdName)
	}
	if *depth < 1 || *depth > maxNeighborDepth {
		return parsedCommand{}, fmt.Errorf("%s: -depth must be between 1 and %d", cmdName, maxNeighborDepth)
	}
	return parsedCommand{
		credPath: *cred,
		exec: func(ctx context.Context, s *store, stdout, _ io.Writer) error {
			view, err := s.hubNeighbors(ctx, label, n, *depth)
			if err != nil {
				return err
			}
			writeNeighbors(stdout, view)
			return nil
		},
	}, nil
}
