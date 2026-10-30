package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"syscall"
	"time"
	"unicode/utf8"
)

const (
	defaultFrom     = "claude-queue"
	deliveryTimeout = 10 * time.Second
)

func runSend(args []string, stdout, stderr io.Writer) error {
	flags := newFlagSet("send", stderr)
	to := flags.String("to", "", "session name or pid")
	from := flags.String("from", defaultFrom, "sender name shown to the session (agy-<id> inside agy)")
	message := flags.String("message", "", "message body")
	file := flags.String("file", "", "path of a file holding the message body")
	handshake := flags.Bool("handshake", false, "send an agy reply handshake message")
	if err := parseFlags(flags, args); err != nil {
		return err
	}
	if *to == "" {
		return errors.New("--to is required")
	}
	given := givenFlags(flags)
	body, err := readBody(given, *message, *file, *handshake)
	if err != nil {
		return err
	}
	sender := resolveFrom(given, *from)
	if err := checkFromName(sender); err != nil {
		return err
	}
	sessions, err := liveSessions()
	if err != nil {
		return err
	}
	target, err := resolveSession(sessions, *to)
	if err != nil {
		return err
	}
	dir, err := sessionsDir()
	if err != nil {
		return err
	}
	token, err := readPeerToken(dir, target)
	if err != nil {
		return err
	}
	msgID, err := newMessageID()
	if err != nil {
		return err
	}
	payload, err := inboxPayload(token, sender, body, msgID)
	if err != nil {
		return err
	}
	if err := deliver(target.SocketPath, payload, deliveryTimeout); err != nil {
		return fmt.Errorf("delivering to %s (pid %d): %w", target.displayName(), target.PID, err)
	}
	fmt.Fprintln(stdout, msgID)
	return nil
}

// givenFlags uses flags.Visit so an explicit empty value such as --message ""
// counts as given instead of as missing.
func givenFlags(flags *flag.FlagSet) map[string]bool {
	given := map[string]bool{}
	flags.Visit(func(f *flag.Flag) { given[f.Name] = true })
	return given
}

func resolveFrom(given map[string]bool, from string) string {
	if given["from"] {
		return from
	}
	id := os.Getenv("ANTIGRAVITY_CONVERSATION_ID")
	if id == "" {
		return defaultFrom
	}
	return agyFrom(id)
}

func checkFromName(name string) error {
	if name == "" {
		return errors.New("--from must not be empty")
	}
	for i := 0; i < len(name); i++ {
		c := name[i]
		if (c < 'a' || c > 'z') && (c < '0' || c > '9') && c != '-' {
			return fmt.Errorf("--from must contain only lowercase ASCII letters, digits, and '-': %q", name)
		}
	}
	return nil
}

func readBody(given map[string]bool, message, file string, handshake bool) (string, error) {
	if given["message"] && given["file"] {
		return "", errors.New("pass at most one of --message and --file")
	}
	hasMsg := given["message"] || given["file"]
	if !hasMsg && !handshake {
		return "", errors.New("pass --handshake, --message, or --file")
	}
	var parts []string
	if handshake {
		if given["from"] {
			return "", errors.New("--handshake cannot be combined with --from")
		}
		hs, err := handshakeBody()
		if err != nil {
			return "", err
		}
		parts = append(parts, hs)
	}
	if hasMsg {
		body := message
		if given["file"] {
			raw, err := os.ReadFile(file)
			if err != nil {
				return "", err
			}
			if !utf8.Valid(raw) {
				return "", fmt.Errorf("%s is not valid UTF-8", file)
			}
			body = string(raw)
		}
		if strings.TrimSpace(body) == "" {
			return "", errors.New("message body is empty")
		}
		parts = append(parts, body)
	}
	return strings.Join(parts, "\n\n"), nil
}

// deliver gives up after timeout because a receiver that never reads would
// otherwise block the write or the pipe flush forever.
func deliver(socketPath string, payload []byte, timeout time.Duration) error {
	done := make(chan error, 1)
	go func() { done <- writeInbox(socketPath, payload) }()
	select {
	case err := <-done:
		return err
	case <-time.After(timeout):
		return fmt.Errorf("the inbox did not take the message within %s", timeout)
	}
}

func writeInbox(socketPath string, payload []byte) error {
	if strings.HasPrefix(socketPath, pipePrefix) {
		return writePipe(socketPath, payload)
	}
	conn, err := net.DialUnix("unix", nil, &net.UnixAddr{Name: socketPath, Net: "unix"})
	if err != nil {
		return err
	}
	defer conn.Close()
	if _, err := conn.Write(payload); err != nil {
		return err
	}
	// Without an explicit shutdown, a Windows AF_UNIX reader can miss EOF
	// after Close.
	return conn.CloseWrite()
}

// errorPipeBusy is ERROR_PIPE_BUSY, returned while every server instance of
// the pipe is connected to another client.
const errorPipeBusy = syscall.Errno(231)

// writePipe retries a busy open because nothing has been sent yet, and
// deliver's timeout bounds the retries. It flushes before closing because
// FlushFileBuffers on a pipe returns only after the server has read
// everything written.
func writePipe(path string, payload []byte) error {
	f, err := os.OpenFile(path, os.O_RDWR, 0)
	for errors.Is(err, errorPipeBusy) {
		time.Sleep(50 * time.Millisecond)
		f, err = os.OpenFile(path, os.O_RDWR, 0)
	}
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := f.Write(payload); err != nil {
		return err
	}
	return f.Sync()
}
