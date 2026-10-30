// Command claude-queue posts messages into Claude Code sessions running on
// the same host through each session's inbox socket.
//
// Subcommands:
// - list: prints the live sessions.
// - send: delivers one message to a session chosen by name or pid.
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
)

const (
	exitOK    = 0
	exitError = 1
)

const usage = `usage:
  claude-queue list
  claude-queue send --to <name|pid> [--from <name>] (--message <text> | --file <path> | --handshake)`

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprintln(stderr, usage)
		return exitError
	}
	var err error
	switch args[0] {
	case "list":
		err = runList(args[1:], stdout, stderr)
	case "send":
		err = runSend(args[1:], stdout, stderr)
	default:
		err = fmt.Errorf("unknown subcommand %q\n%s", args[0], usage)
	}
	if err != nil {
		fmt.Fprintln(stderr, "claude-queue:", err)
		return exitError
	}
	return exitOK
}

func newFlagSet(name string, stderr io.Writer) *flag.FlagSet {
	flags := flag.NewFlagSet(name, flag.ContinueOnError)
	flags.SetOutput(stderr)
	return flags
}

// parseFlags rejects leftover arguments so a mistyped flag value is not
// silently dropped.
func parseFlags(flags *flag.FlagSet, args []string) error {
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() > 0 {
		return fmt.Errorf("unexpected argument %q", flags.Arg(0))
	}
	return nil
}

func runList(args []string, stdout, stderr io.Writer) error {
	if err := parseFlags(newFlagSet("list", stderr), args); err != nil {
		return err
	}
	sessions, err := liveSessions()
	if err != nil {
		return err
	}
	fmt.Fprintln(stdout, "NAME\tPID\tKIND\tSTATUS\tCWD")
	for _, s := range sessions {
		fmt.Fprintf(stdout, "%s\t%d\t%s\t%s\t%s\n", s.displayName(), s.PID, s.Kind, s.Status, s.Cwd)
	}
	return nil
}
