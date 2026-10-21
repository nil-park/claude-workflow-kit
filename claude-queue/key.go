package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const pipePrefix = `\\.\pipe\`

var peerTokenPattern = regexp.MustCompile(`^[0-9a-f]{32}$`)

// keyFileHash names the key file of the session that listens on socketPath.
func keyFileHash(socketPath string) (string, error) {
	canonical, err := canonicalSocketPath(socketPath)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256([]byte(canonical))
	return hex.EncodeToString(sum[:]), nil
}

// canonicalSocketPath lowercases only ASCII letters of a pipe name because
// Claude Code derives the key file name from that form.
func canonicalSocketPath(socketPath string) (string, error) {
	if name, ok := strings.CutPrefix(socketPath, pipePrefix); ok {
		var b strings.Builder
		for i := 0; i < len(name); i++ {
			c := name[i]
			if 'A' <= c && c <= 'Z' {
				c += 'a' - 'A'
			}
			b.WriteByte(c)
		}
		return pipePrefix + b.String(), nil
	}
	return filepath.Abs(socketPath)
}

// readPeerToken returns an empty token when the session has no key file, in
// which case the message goes out without an auth line.
func readPeerToken(dir string, s session) (string, error) {
	hash, err := keyFileHash(s.SocketPath)
	if err != nil {
		return "", err
	}
	path := filepath.Join(dir, fmt.Sprintf("%d.%s.key", s.PID, hash))
	raw, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	var key struct {
		PeerToken string `json:"peerToken"`
	}
	if err := json.Unmarshal(raw, &key); err != nil || !peerTokenPattern.MatchString(key.PeerToken) {
		return "", fmt.Errorf("key file %s has no usable peerToken", path)
	}
	return key.PeerToken, nil
}
