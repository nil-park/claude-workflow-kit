package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type credentials struct {
	URI      string
	Username string
	Password string
	Database string
}

// readCredentials reads the KEY=VALUE file that Aura hands out when an
// instance is created.
func readCredentials(path string) (credentials, error) {
	f, err := os.Open(path)
	if err != nil {
		return credentials{}, err
	}
	defer f.Close()

	values := map[string]string{}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if key, value, ok := strings.Cut(line, "="); ok {
			values[strings.TrimSpace(key)] = strings.TrimSpace(value)
		}
	}
	if err := scanner.Err(); err != nil {
		return credentials{}, err
	}

	cred := credentials{
		URI:      values["NEO4J_URI"],
		Username: values["NEO4J_USERNAME"],
		Password: values["NEO4J_PASSWORD"],
		Database: values["NEO4J_DATABASE"],
	}
	for key, value := range map[string]string{"NEO4J_URI": cred.URI, "NEO4J_USERNAME": cred.Username, "NEO4J_PASSWORD": cred.Password} {
		if value == "" {
			return credentials{}, fmt.Errorf("%s: %s is missing", path, key)
		}
	}
	return cred, nil
}
