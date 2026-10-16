package main

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func fakeEnv(values map[string]string) func(string) string {
	return func(key string) string { return values[key] }
}

var agyEnv = map[string]string{
	"ANTIGRAVITY_CONVERSATION_ID": "cd5e6efd-6ae7-4519-9ec8-bab59cd271f7",
	"ANTIGRAVITY_LS_ADDRESS":      "localhost:60993",
	"ANTIGRAVITY_CSRF_TOKEN":      "token",
}

func TestRegisterWritesEndpointAndPrintsID(t *testing.T) {
	dir := t.TempDir()
	var stdout bytes.Buffer
	args := []string{"--dir", dir, "--role", "peer"}
	if err := runRegister(args, &stdout, io.Discard, fakeEnv(agyEnv)); err != nil {
		t.Fatal(err)
	}
	if got := stdout.String(); got != agyEnv["ANTIGRAVITY_CONVERSATION_ID"]+"\n" {
		t.Errorf("stdout = %q", got)
	}
	endpoint, err := readEndpoint(endpointPath(dir, "peer"))
	if err != nil {
		t.Fatal(err)
	}
	for key, want := range agyEnv {
		if endpoint[key] != want {
			t.Errorf("%s = %q, want %q", key, endpoint[key], want)
		}
	}
}

func TestRegisterFailsWithoutFileWhenVariableIsEmpty(t *testing.T) {
	dir := t.TempDir()
	env := map[string]string{}
	for key, value := range agyEnv {
		env[key] = value
	}
	env["ANTIGRAVITY_CSRF_TOKEN"] = ""
	err := runRegister([]string{"--dir", dir, "--role", "peer"}, io.Discard, io.Discard, fakeEnv(env))
	if err == nil || !strings.Contains(err.Error(), "ANTIGRAVITY_CSRF_TOKEN") {
		t.Fatalf("err = %v", err)
	}
	if _, err := os.Stat(endpointPath(dir, "peer")); !os.IsNotExist(err) {
		t.Errorf("endpoint file exists: %v", err)
	}
}

func TestRegisterOverwritesEndpoint(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(endpointPath(dir, "peer"), []byte("stale"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := runRegister([]string{"--dir", dir, "--role", "peer"}, io.Discard, io.Discard, fakeEnv(agyEnv)); err != nil {
		t.Fatal(err)
	}
	if _, err := readEndpoint(endpointPath(dir, "peer")); err != nil {
		t.Fatal(err)
	}
}
