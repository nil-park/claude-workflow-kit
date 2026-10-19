// Command agy-compaction-hook detects context compaction in agy sessions
// and injects a reload message prompting the model to re-read skills and
// standards documents.
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
)

type hookInput struct {
	TranscriptPath string `json:"transcriptPath"`
}

type systemMessageInner struct {
	SystemMessage string `json:"systemMessage"`
}

type systemMessageStep struct {
	SystemMessage systemMessageInner `json:"systemMessage"`
}

type injectOutput struct {
	InjectSteps []systemMessageStep `json:"injectSteps"`
}

func main() {
	run(os.Stdin, os.Stdout, os.Stderr)
}

func run(stdin io.Reader, stdout, stderr io.Writer) {
	inject, err := shouldInject(stdin)
	if err != nil {
		fmt.Fprintln(stderr, "agy-compaction-hook:", err)
		fmt.Fprintln(stdout, "{}")
		return
	}
	if !inject {
		fmt.Fprintln(stdout, "{}")
		return
	}
	out := injectOutput{
		InjectSteps: []systemMessageStep{
			{SystemMessage: systemMessageInner{SystemMessage: reloadMessage}},
		},
	}
	data, err := json.Marshal(out)
	if err != nil {
		fmt.Fprintln(stderr, "agy-compaction-hook:", err)
		fmt.Fprintln(stdout, "{}")
		return
	}
	if _, err := stdout.Write(append(data, '\n')); err != nil {
		fmt.Fprintln(stderr, "agy-compaction-hook:", err)
	}
}

func shouldInject(stdin io.Reader) (bool, error) {
	var input hookInput
	if err := json.NewDecoder(stdin).Decode(&input); err != nil {
		return false, fmt.Errorf("decode stdin: %w", err)
	}
	if input.TranscriptPath == "" {
		return false, errors.New("empty transcriptPath")
	}
	file, err := os.Open(input.TranscriptPath)
	if err != nil {
		return false, fmt.Errorf("open transcript: %w", err)
	}
	defer file.Close()
	return needsReload(file)
}
