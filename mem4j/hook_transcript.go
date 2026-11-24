package main

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
)

func agyTurnKeyFromCount(userInputs int, hasTranscript bool, prompt string) string {
	sum := sha256.Sum256([]byte(normalizeText(prompt)))
	hash := hex.EncodeToString(sum[:8])
	if hasTranscript {
		return fmt.Sprintf("u:%d:%s", userInputs, hash)
	}
	return "p:" + hash
}

func scanAgyTranscript(path string) (userInputs, checkpoints int, err error) {
	if path == "" {
		return 0, 0, os.ErrNotExist
	}
	f, err := os.Open(path)
	if err != nil {
		return 0, 0, err
	}
	defer f.Close()

	reader := bufio.NewReader(f)
	for {
		line, err := reader.ReadBytes('\n')
		if trimmed := bytes.TrimSpace(line); len(trimmed) > 0 {
			var step struct {
				Type string `json:"type"`
			}
			if json.Unmarshal(trimmed, &step) == nil {
				if step.Type == "USER_INPUT" {
					userInputs++
				} else if step.Type == "CHECKPOINT" {
					checkpoints++
				}
			}
		}
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return 0, 0, err
		}
	}
	return userInputs, checkpoints, nil
}
