package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"strings"
)

const hookMarker = "[agy-compaction-hook]"

const reloadMessage = hookMarker + `
Context was just compacted.
Re-read, from the original files, every skill and standards document loaded in this conversation that is still being followed.
AGENTS.md, GEMINI.md, rules, and the agent definition are part of the system prompt and need no re-reading.`

type transcriptStep struct {
	Source  string `json:"source"`
	Type    string `json:"type"`
	Content string `json:"content"`
}

// needsReload reads a transcript line by line in a single pass and returns
// whether the newest checkpoint still requires a reload message.
func needsReload(r io.Reader) (bool, error) {
	reader := bufio.NewReader(r)
	pending := false

	for {
		line, err := reader.ReadBytes('\n')
		if len(line) > 0 {
			pending = updatePending(line, pending)
		}
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return false, err
		}
	}

	return pending, nil
}

func updatePending(line []byte, pending bool) bool {
	trimmed := bytes.TrimSpace(line)
	if len(trimmed) == 0 {
		return pending
	}
	var step transcriptStep
	if err := json.Unmarshal(trimmed, &step); err != nil {
		return pending
	}
	if step.Type == "CHECKPOINT" {
		return true
	}
	if pending &&
		step.Source == "SYSTEM_SDK" &&
		step.Type == "SYSTEM_MESSAGE" &&
		strings.Contains(step.Content, hookMarker) {
		return false
	}
	return pending
}
