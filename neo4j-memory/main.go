// Command neo4j-memory stores agent instructions as a Neo4j graph and merges
// the ones that apply to a set of scopes.
//
//	neo4j-memory load  [-seed seed/plugins.yaml.tmpl] [-values ../.secrets/neo4j-memory-values.yaml] [-root ..]
//	neo4j-memory merge -scope <name> [-scope <name> ...]
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

type scopeList []string

func (s *scopeList) String() string { return strings.Join(*s, ", ") }

func (s *scopeList) Set(v string) error {
	*s = append(*s, normalizeText(v))
	return nil
}

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, "neo4j-memory:", err)
		os.Exit(1)
	}
}

func run(args []string, stdout, stderr io.Writer) error {
	if len(args) == 0 {
		return errors.New("usage: neo4j-memory load|merge [flags]")
	}
	fs := flag.NewFlagSet(args[0], flag.ContinueOnError)
	credPath := fs.String("credentials", "../.secrets/neo4j.txt", "Aura credential file (KEY=VALUE)")
	seedPath := fs.String("seed", "seed/plugins.yaml.tmpl", "seed manifest template (load)")
	valuesPath := fs.String("values", "../.secrets/neo4j-memory-values.yaml", "values the manifest template is rendered with (load)")
	root := fs.String("root", "..", "directory the manifest's file paths resolve against (load)")
	var scopes scopeList
	fs.Var(&scopes, "scope", "active scope name, repeatable (merge)")
	if err := fs.Parse(args[1:]); err != nil {
		return err
	}

	// Check everything that needs no database before connecting to one.
	var seed Seed
	switch args[0] {
	case "load":
		var err error
		if seed, err = loadSeed(*seedPath, *valuesPath, *root); err != nil {
			return err
		}
	case "merge":
		if len(scopes) == 0 {
			return errors.New("merge needs at least one -scope")
		}
	default:
		return fmt.Errorf("unknown command %q", args[0])
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	cred, err := readCredentials(*credPath)
	if err != nil {
		return err
	}
	driver, err := neo4j.NewDriverWithContext(cred.URI, neo4j.BasicAuth(cred.Username, cred.Password, ""))
	if err != nil {
		return err
	}
	defer driver.Close(ctx)

	if args[0] == "load" {
		if err := writeSeed(ctx, driver, cred.Database, seed); err != nil {
			return err
		}
		fmt.Fprintf(stdout, "loaded %d scopes and %d instructions\n", len(seed.Scopes), len(seed.Instructions))
		return nil
	}
	candidates, err := fetchCandidates(ctx, driver, cred.Database, scopes)
	if err != nil {
		return err
	}
	result := Merge(candidates, scopes)
	writeMerged(stdout, result)
	for _, w := range result.Warnings {
		fmt.Fprintln(stderr, "warning:", w)
	}
	return nil
}

// writeMerged tags each instruction with its scopes, so the agent can weigh
// conflicting instructions by where they come from.
func writeMerged(w io.Writer, result MergeResult) {
	for _, s := range result.Sections {
		fmt.Fprintf(w, "<instruction key=%q title=%q scopes=%q>\n%s\n</instruction>\n\n",
			s.Key, s.Title, strings.Join(s.Scopes, ", "), strings.TrimRight(s.Body, "\n"))
	}
}
