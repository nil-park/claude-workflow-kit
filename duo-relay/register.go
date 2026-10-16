package main

import (
	"encoding/json"
	"fmt"
	"io"
)

const conversationIDKey = "ANTIGRAVITY_CONVERSATION_ID"

// endpointKeys are the variables an interactive agy session sets for its shell
// commands. The endpoint file stores them under the same names so push can
// hand them back to agy unchanged.
var endpointKeys = []string{
	conversationIDKey,
	"ANTIGRAVITY_LS_ADDRESS",
	"ANTIGRAVITY_CSRF_TOKEN",
}

func runRegister(args []string, stdout, stderr io.Writer, getenv func(string) string) error {
	fs := newFlagSet("register", stderr)
	dir := fs.String("dir", "", "shared directory")
	role := fs.String("role", "", "main or peer")
	if err := parseFlags(fs, args); err != nil {
		return err
	}
	if err := checkDir(*dir); err != nil {
		return err
	}
	if err := checkRole("role", *role); err != nil {
		return err
	}
	endpoint := make(map[string]string, len(endpointKeys))
	for _, key := range endpointKeys {
		value := getenv(key)
		if value == "" {
			return fmt.Errorf("%s is not set; run register from an agy session", key)
		}
		endpoint[key] = value
	}
	data, err := json.MarshalIndent(endpoint, "", "  ")
	if err != nil {
		return err
	}
	if err := writeFileAtomic(endpointPath(*dir, *role), append(data, '\n')); err != nil {
		return err
	}
	fmt.Fprintln(stdout, endpoint[conversationIDKey])
	return nil
}
