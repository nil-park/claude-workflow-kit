// Command statusline renders the Claude Code status line from the session
// JSON that Claude Code writes to stdin, taking the quota from CLIProxyAPI
// when a gpt-* session's stdin has none:
// model | context usage | session cost | 5h quota | 7d quota | idle.
package main

import (
	"encoding/json"
	"errors"
	"io"
	"math"
	"os"
	"strings"
	"time"
	"unicode/utf8"
)

type quotaWindow struct {
	usedPercent float64
	resetsAt    *time.Time
}

type statusInput struct {
	modelID        string
	modelName      string
	inputTokens    *float64
	contextPercent *float64
	costUSD        *float64
	fiveHour       *quotaWindow
	sevenDay       *quotaWindow
	transcriptPath string
}

func main() {
	raw, err := io.ReadAll(os.Stdin)
	if err != nil {
		raw = nil
	}
	now := time.Now()
	os.Stdout.WriteString(render(withCLIProxyQuota(parseInput(raw), now), now))
}

// parseInput decodes into generic values rather than a tagged struct so a
// field with an unexpected type reads as missing without touching the others.
func parseInput(raw []byte) statusInput {
	if !utf8.Valid(raw) {
		return statusInput{}
	}
	var root map[string]any
	// A syntax error fails before anything is decoded. A type error (an
	// out-of-range number) still decodes the rest, so only that value is lost.
	var typeErr *json.UnmarshalTypeError
	if err := json.Unmarshal(raw, &root); err != nil && !errors.As(err, &typeErr) {
		return statusInput{}
	}
	return statusInput{
		modelID:        text(field(root, "model", "id")),
		modelName:      text(field(root, "model", "display_name")),
		inputTokens:    number(field(root, "context_window", "total_input_tokens")),
		contextPercent: number(field(root, "context_window", "used_percentage")),
		costUSD:        number(field(root, "cost", "total_cost_usd")),
		fiveHour:       parseQuota(field(root, "rate_limits", "five_hour")),
		sevenDay:       parseQuota(field(root, "rate_limits", "seven_day")),
		transcriptPath: text(field(root, "transcript_path")),
	}
}

func parseQuota(v any) *quotaWindow {
	used := number(field(v, "used_percentage"))
	if used == nil {
		return nil
	}
	window := &quotaWindow{usedPercent: *used}
	if epoch := number(field(v, "resets_at")); epoch != nil {
		at := time.Unix(int64(*epoch), 0)
		window.resetsAt = &at
	}
	return window
}

// field returns nil when any step of the path is missing or is not an object.
func field(v any, path ...string) any {
	for _, key := range path {
		obj, ok := v.(map[string]any)
		if !ok {
			return nil
		}
		v = obj[key]
	}
	return v
}

func text(v any) string {
	s, _ := v.(string)
	return s
}

// number also rejects ±Inf, which the decoder stores alongside the type error
// for an out-of-range literal such as 1e400.
func number(v any) *float64 {
	f, ok := v.(float64)
	if !ok || math.IsInf(f, 0) {
		return nil
	}
	return &f
}

func render(in statusInput, now time.Time) string {
	segments := []func() string{
		func() string { return segModel(in.modelName) },
		func() string { return segContext(in.inputTokens, in.contextPercent) },
		func() string { return segCost(in.costUSD) },
		func() string { return segQuota(in.fiveHour, "5h", "15:04") },
		func() string { return segQuota(in.sevenDay, "7d", "01-02 15:04") },
		func() string { return segIdle(in.transcriptPath, now) },
	}
	var shown []string
	for _, segment := range segments {
		if s := safe(segment); s != "" {
			shown = append(shown, s)
		}
	}
	return strings.Join(shown, separator)
}

func safe(segment func() string) (out string) {
	defer func() {
		if recover() != nil {
			out = ""
		}
	}()
	return segment()
}
