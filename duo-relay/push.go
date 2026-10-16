package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"time"
)

const notifyTimeout = 30 * time.Second

func runPush(args []string, stderr io.Writer) error {
	flags := newFlagSet("push", stderr)
	dir := flags.String("dir", "", "shared directory")
	to := flags.String("to", "", "main or peer")
	round := flags.Int("round", 0, "round number")
	file := flags.String("file", "", "message body file")
	if err := parseFlags(flags, args); err != nil {
		return err
	}
	if err := checkDir(*dir); err != nil {
		return err
	}
	if err := checkRole("to", *to); err != nil {
		return err
	}
	if err := checkRound(*round); err != nil {
		return err
	}
	if *file == "" {
		return errors.New("--file is required")
	}
	body, err := os.ReadFile(*file)
	if err != nil {
		return err
	}
	path := roundPath(*dir, *round, *to)
	if err := writeFileAtomic(path, body); err != nil {
		return err
	}
	endpoint, err := readEndpoint(endpointPath(*dir, *to))
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	notice := fmt.Sprintf("duo-relay: round %d message for %s is at %s", *round, *to, path)
	return notify(endpoint, notice, stderr)
}

func readEndpoint(path string) (map[string]string, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var endpoint map[string]string
	if err := json.Unmarshal(raw, &endpoint); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	for _, key := range endpointKeys {
		if endpoint[key] == "" {
			return nil, fmt.Errorf("%s: %s is empty", path, key)
		}
	}
	return endpoint, nil
}

// notify is a variable so tests can run push without an agy session.
var notify = func(endpoint map[string]string, notice string, stderr io.Writer) error {
	ctx, cancel := context.WithTimeout(context.Background(), notifyTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "agy", "agentapi", "send-message", endpoint[conversationIDKey], notice)
	cmd.Env = os.Environ()
	for _, key := range endpointKeys {
		cmd.Env = append(cmd.Env, key+"="+endpoint[key])
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		stderr.Write(out)
		return fmt.Errorf("agy agentapi send-message: %w", err)
	}
	return nil
}
