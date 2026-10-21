package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestKeyFileHashLowercasesPipeName(t *testing.T) {
	// sha256 of `\\.\pipe\local\cc-msg-0123456789abcdef0123456789abcdef`,
	// computed outside this package.
	const want = "a22f6f21321c8636e75cbe65adbb453652c33c690dd6007c8937a49f23132a9d"
	got, err := keyFileHash(`\\.\pipe\LOCAL\cc-msg-0123456789ABCDEF0123456789abcdef`)
	if err != nil || got != want {
		t.Errorf("keyFileHash() = %q, %v; want %q", got, err, want)
	}
}

func writeKey(t *testing.T, dir string, s session, token string) {
	t.Helper()
	hash, err := keyFileHash(s.SocketPath)
	if err != nil {
		t.Fatal(err)
	}
	writeJSON(t, filepath.Join(dir, fmt.Sprintf("%d.%s.key", s.PID, hash)), map[string]string{"peerToken": token})
}

func TestReadPeerToken(t *testing.T) {
	dir := t.TempDir()
	s := session{PID: 42, SocketPath: `\\.\pipe\LOCAL\cc-msg-x`}
	if token, err := readPeerToken(dir, s); token != "" || err != nil {
		t.Errorf("without a key file: %q, %v; want no token and no error", token, err)
	}

	const valid = "0123456789abcdef0123456789abcdef"
	writeKey(t, dir, s, valid)
	if token, err := readPeerToken(dir, s); token != valid || err != nil {
		t.Errorf("with a key file: %q, %v; want %q", token, err, valid)
	}

	other := session{PID: 43, SocketPath: s.SocketPath}
	if token, _ := readPeerToken(dir, other); token != "" {
		t.Error("a key file of another pid was used")
	}

	for _, bad := range []string{"", "0123456789ABCDEF0123456789ABCDEF", valid + "0"} {
		writeKey(t, dir, s, bad)
		if _, err := readPeerToken(dir, s); err == nil {
			t.Errorf("peerToken %q was accepted", bad)
		}
	}
	hash, _ := keyFileHash(s.SocketPath)
	if err := os.WriteFile(filepath.Join(dir, fmt.Sprintf("42.%s.key", hash)), []byte("{"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := readPeerToken(dir, s); err == nil {
		t.Error("a key file that is not JSON was accepted")
	}
}
