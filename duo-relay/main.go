// Command duo-relay relays messages between a Claude Code main session and an
// agy peer session through a shared directory.
//
// Subcommands:
// - register: records the agy session's endpoint for direct notices.
// - push: writes a round message and sends a direct notice when the receiver
// has an endpoint.
// - wait: prints the round message for a role once it arrives.
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const (
	exitOK      = 0
	exitError   = 1
	exitTimeout = 2
	exitBusy    = 3
)

const usage = `usage:
  duo-relay register --dir <dir> --role <main|peer>
  duo-relay push --dir <dir> --to <main|peer> --round <N> --file <body file>
  duo-relay wait --dir <dir> --role <main|peer> --round <N> --timeout <seconds>`

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprintln(stderr, usage)
		return exitError
	}
	var err error
	code := exitOK
	switch args[0] {
	case "register":
		err = runRegister(args[1:], stdout, stderr, os.Getenv)
	case "push":
		err = runPush(args[1:], stderr)
	case "wait":
		code, err = runWait(args[1:], stdout, stderr)
	default:
		err = fmt.Errorf("unknown subcommand %q\n%s", args[0], usage)
	}
	if err != nil {
		fmt.Fprintln(stderr, "duo-relay:", err)
		return exitError
	}
	return code
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

func checkDir(dir string) error {
	if !filepath.IsAbs(dir) {
		return fmt.Errorf("--dir must be an absolute path: %q", dir)
	}
	return nil
}

func checkRole(flagName, role string) error {
	if role != "main" && role != "peer" {
		return fmt.Errorf("--%s must be main or peer: %q", flagName, role)
	}
	return nil
}

func checkRound(round int) error {
	if round < 1 {
		return errors.New("--round must be 1 or greater")
	}
	return nil
}
