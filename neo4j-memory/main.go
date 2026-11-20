// Command neo4j-memory manages agent instructions and scopes in a Neo4j graph
// and merges the instructions that apply to a set of scopes.
//
//	neo4j-memory scope-create -name <name> -description <text> [-disable-model-invocation]
//	neo4j-memory scope-edit -name <name> -description <text>
//	neo4j-memory scope-delete -name <name>
//	neo4j-memory instruction-create -key <key> -title <title> -scope <name> [-scope ...] (-body <text> | -file <path|->)
//	neo4j-memory instruction-edit -key <key> [-title <title>] [-body <text> | -file <path|->]
//	neo4j-memory instruction-link -key <key> -scope <name>
//	neo4j-memory instruction-unlink -key <key> -scope <name>
//	neo4j-memory instruction-delete -key <key>
//	neo4j-memory merge -scope <name> [-scope <name> ...]
package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func main() {
	if err := run(os.Args[1:], os.Stdin, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "neo4j-memory:", err)
		os.Exit(1)
	}
}

func run(args []string, stdin io.Reader, stdout io.Writer) error {
	cmd, err := parseCommand(args, stdin)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	cred, err := readCredentials(cmd.credPath)
	if err != nil {
		return err
	}
	driver, err := neo4j.NewDriverWithContext(cred.URI, neo4j.BasicAuth(cred.Username, cred.Password, ""))
	if err != nil {
		return err
	}
	defer driver.Close(ctx)

	return cmd.exec(ctx, &store{driver: driver, db: cred.Database}, stdout)
}

// writeMerged tags each instruction with its scopes, so the agent can weigh
// conflicting instructions by where they come from.
func writeMerged(w io.Writer, instructions []Instruction) {
	for _, inst := range instructions {
		fmt.Fprintf(w, "<instruction key=%q title=%q scopes=%q>\n%s\n</instruction>\n\n",
			inst.Key, inst.Title, strings.Join(inst.Scopes, ", "), strings.TrimRight(inst.Body, "\n"))
	}
}
