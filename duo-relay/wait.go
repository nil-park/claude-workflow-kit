package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const pollInterval = 200 * time.Millisecond

func runWait(args []string, stdout, stderr io.Writer) (int, error) {
	flags := newFlagSet("wait", stderr)
	dir := flags.String("dir", "", "shared directory")
	role := flags.String("role", "", "main or peer")
	round := flags.Int("round", 0, "round number")
	timeout := flags.Int("timeout", 0, "seconds to wait")
	if err := parseFlags(flags, args); err != nil {
		return exitError, err
	}
	if err := checkDir(*dir); err != nil {
		return exitError, err
	}
	if err := checkRole("role", *role); err != nil {
		return exitError, err
	}
	if err := checkRound(*round); err != nil {
		return exitError, err
	}
	if *timeout < 1 {
		return exitError, errors.New("--timeout must be 1 or greater")
	}

	lock := pidPath(*dir, *role)
	if pid, ok := readPID(lock); ok && pid != os.Getpid() && processAlive(pid) {
		fmt.Fprintf(stderr, "duo-relay: wait for %s is already running as pid %d\n", *role, pid)
		return exitBusy, nil
	}
	self := os.Getpid()
	if err := writeFileAtomic(lock, []byte(strconv.Itoa(self)+"\n")); err != nil {
		return exitError, err
	}
	defer releaseLock(lock, self)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	deadline := time.NewTimer(time.Duration(*timeout) * time.Second)
	defer deadline.Stop()
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	path := roundPath(*dir, *round, *role)
	for {
		body, err := os.ReadFile(path)
		if err == nil {
			stdout.Write(body)
			return exitOK, nil
		}
		if !errors.Is(err, fs.ErrNotExist) {
			return exitError, err
		}
		select {
		case <-ctx.Done():
			return exitError, errors.New("interrupted")
		case <-deadline.C:
			fmt.Fprintf(stdout, "duo-relay: no round %d message for %s within %d seconds\n", *round, *role, *timeout)
			return exitTimeout, nil
		case <-ticker.C:
		}
	}
}

func readPID(path string) (int, bool) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return 0, false
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(raw)))
	if err != nil || pid <= 0 {
		return 0, false
	}
	return pid, true
}

// releaseLock leaves the file alone when another wait has taken it over after
// this one was judged dead.
func releaseLock(path string, self int) {
	if pid, ok := readPID(path); ok && pid == self {
		os.Remove(path)
	}
}
