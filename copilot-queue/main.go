// Command copilot-queue posts messages into an interactive GitHub Copilot CLI
// session running on the same host through its UI server socket.
//
// Subcommands:
// - send: delivers one message to the foreground Copilot session on a port.
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
  copilot-queue send --port <port> --from <name> [--handshake] [--message <text> | --file <path>]`

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
	case "send":
		err = runSend(args[1:], stdout, stderr)
	default:
		err = fmt.Errorf("unknown subcommand %q\n%s", args[0], usage)
	}
	if err != nil {
		fmt.Fprintln(stderr, "copilot-queue:", err)
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
