// Command mem4j manages agent instructions and scopes in a Neo4j graph
// and merges the instructions that apply to a set of scopes.
//
//	mem4j scope list
//	mem4j scope create -name <name> -description <text> [-disable-model-invocation]
//	mem4j scope edit -name <name> -description <text>
//	mem4j scope delete -name <name>
//	mem4j instruction recall -scope <name> [-scope <name> ...]
//	mem4j instruction create -key <key> -title <title> -scope <name> [-scope ...] (-body <text> | -file <path|->)
//	mem4j instruction edit -key <key> [-title <title>] [-body <text> | -file <path|->]
//	mem4j instruction link -key <key> -scope <name>
//	mem4j instruction unlink -key <key> -scope <name>
//	mem4j instruction delete -key <key>
//	mem4j reminder recall -scope <name> [-scope ...] (-query <text> | -file <path|->) [-threshold 0.5] [-limit 5]
//	mem4j reminder create [-key <key>] -title <title> -description <text> -scope <name> [-scope ...] (-body <text> | -file <path|->)
//	mem4j reminder edit -key <key> [-title <title>] [-description <text>] [-body <text> | -file <path|->]
//	mem4j reminder link -key <key> -scope <name>
//	mem4j reminder unlink -key <key> -scope <name>
//	mem4j reminder delete -key <key>
//	mem4j graph query (-cypher <text> | -file <path|->)
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
	if err := run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, "mem4j:", err)
		os.Exit(1)
	}
}

func run(args []string, stdin io.Reader, stdout, stderr io.Writer) error {
	cmd, err := parseCommand(args, stdin)
	if err != nil {
		return err
	}
	if cmd.helpText != "" {
		_, err := io.WriteString(stdout, cmd.helpText)
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	credPath := cmd.credPath
	if credPath == "" {
		credPath, err = defaultCredentialsPath()
		if err != nil {
			return err
		}
	}
	cred, err := readCredentials(credPath)
	if err != nil {
		return err
	}
	driver, err := neo4j.NewDriverWithContext(cred.URI, neo4j.BasicAuth(cred.Username, cred.Password, ""))
	if err != nil {
		return err
	}
	defer driver.Close(ctx)

	return cmd.exec(ctx, &store{driver: driver, db: cred.Database}, stdout, stderr)
}

// writeMerged tags each instruction with its scopes, so the agent can weigh
// conflicting instructions by where they come from.
func writeMerged(w io.Writer, instructions []Instruction) {
	for _, inst := range instructions {
		fmt.Fprintf(w, "<instruction key=%q title=%q scopes=%q>\n%s\n</instruction>\n\n",
			inst.Key, inst.Title, strings.Join(inst.Scopes, ", "), strings.TrimRight(inst.Body, "\n"))
	}
}

func writeEmptyMergeNotice(w io.Writer, missing []string) {
	if len(missing) == 0 {
		fmt.Fprintln(w, "mem4j: no instructions apply to the requested scopes")
		return
	}
	fmt.Fprintf(w, "mem4j: no instructions apply (unknown scopes: %s)\n", strings.Join(missing, ", "))
}

func writeScopes(w io.Writer, scopes []Scope) {
	for _, s := range scopes {
		fmt.Fprintf(w, "<scope name=%q disableModelInvocation=%t>\n%s\n</scope>\n\n",
			s.Name, s.DisableModelInvocation, strings.TrimRight(s.Description, "\n"))
	}
}

func writeReminders(w io.Writer, reminders []Reminder) {
	for _, r := range reminders {
		fmt.Fprintf(w, "<reminder key=%q title=%q scopes=%q score=\"%.3f\">\n%s\n</reminder>\n\n",
			r.Key, r.Title, strings.Join(r.Scopes, ", "), r.Score, strings.TrimRight(r.Body, "\n"))
	}
}

func writeEmptyReminderNotice(w io.Writer, missing []string) {
	if len(missing) == 0 {
		fmt.Fprintln(w, "mem4j: no reminders match the query in the requested scopes")
		return
	}
	fmt.Fprintf(w, "mem4j: no reminders match (unknown scopes: %s)\n", strings.Join(missing, ", "))
}
