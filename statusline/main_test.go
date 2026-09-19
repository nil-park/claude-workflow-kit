package main

import (
	"os"
	"testing"
	"time"
)

var kst = time.FixedZone("KST", 9*60*60)

func TestParseInputCaptured(t *testing.T) {
	raw, err := os.ReadFile("testdata/captured-stdin.json")
	if err != nil {
		t.Fatal(err)
	}
	in := parseInput(raw)

	if in.modelID != "claude-opus-5" || in.modelName != "Opus 5" {
		t.Errorf("modelID = %q, modelName = %q", in.modelID, in.modelName)
	}
	if in.inputTokens == nil || *in.inputTokens != 216297 {
		t.Errorf("inputTokens = %v", in.inputTokens)
	}
	if in.contextPercent == nil || *in.contextPercent != 22 {
		t.Errorf("contextPercent = %v", in.contextPercent)
	}
	if in.costUSD == nil || *in.costUSD != 11.942875 {
		t.Errorf("costUSD = %v", in.costUSD)
	}
	if in.fiveHour == nil || in.fiveHour.usedPercent != 19 ||
		in.fiveHour.resetsAt == nil || in.fiveHour.resetsAt.Unix() != 1789798800 {
		t.Errorf("fiveHour = %+v", in.fiveHour)
	}
	if in.sevenDay == nil || in.sevenDay.usedPercent != 7 ||
		in.sevenDay.resetsAt == nil || in.sevenDay.resetsAt.Unix() != 1790287200 {
		t.Errorf("sevenDay = %+v", in.sevenDay)
	}
}

func TestParseInputUnusableInput(t *testing.T) {
	cases := map[string]string{
		"empty":         "",
		"syntax error":  `{"model":{"display_name":"Opus 5"}`,
		"invalid UTF-8": "{\"model\":{\"display_name\":\"Opus \xff\"}}",
		"array":         `[{"model":{"display_name":"Opus 5"}}]`,
		"null":          "null",
	}
	for name, raw := range cases {
		t.Run(name, func(t *testing.T) {
			if got := parseInput([]byte(raw)); got != (statusInput{}) {
				t.Errorf("parseInput = %+v, want zero value", got)
			}
		})
	}
}

func TestParseInputWrongTypeDropsOnlyThatField(t *testing.T) {
	in := parseInput([]byte(`{
		"model": {"display_name": 5},
		"cost": {"total_cost_usd": "1.00"},
		"context_window": {"total_input_tokens": 1e400, "used_percentage": 22},
		"rate_limits": {"five_hour": {"used_percentage": 19, "resets_at": "15:20"}, "seven_day": "7%"},
		"transcript_path": ["a"]
	}`))

	if in.modelName != "" || in.costUSD != nil || in.inputTokens != nil ||
		in.sevenDay != nil || in.transcriptPath != "" {
		t.Errorf("wrongly typed fields should be missing: %+v", in)
	}
	if in.contextPercent == nil || *in.contextPercent != 22 {
		t.Errorf("contextPercent = %v, want 22", in.contextPercent)
	}
	if in.fiveHour == nil || in.fiveHour.usedPercent != 19 || in.fiveHour.resetsAt != nil {
		t.Errorf("fiveHour = %+v, want 19%% without reset time", in.fiveHour)
	}
}

func TestRender(t *testing.T) {
	in := statusInput{
		modelName:      "Opus 5",
		inputTokens:    new(145_700.0),
		contextPercent: new(15.0),
		costUSD:        new(5.18),
		fiveHour:       &quotaWindow{usedPercent: 8, resetsAt: new(time.Date(2026, 9, 19, 15, 20, 0, 0, kst))},
		sevenDay:       &quotaWindow{usedPercent: 7, resetsAt: new(time.Date(2026, 9, 25, 7, 0, 0, 0, kst))},
	}
	want := "\x1b[0;32mOpus 5\x1b[0m" +
		" \x1b[0;90m|\x1b[0m \x1b[0;35m145.7k (15%)\x1b[0m \x1b[0;90mctx\x1b[0m" +
		" \x1b[0;90m|\x1b[0m \x1b[0;39m$5.18\x1b[0m \x1b[0;90msession\x1b[0m" +
		" \x1b[0;90m|\x1b[0m \x1b[0;32m8%\x1b[0m \x1b[0;90m5h\x1b[0m \x1b[0;90m→\x1b[0m \x1b[0;39m15:20\x1b[0m" +
		" \x1b[0;90m|\x1b[0m \x1b[0;32m7%\x1b[0m \x1b[0;90m7d\x1b[0m \x1b[0;90m→\x1b[0m \x1b[0;39m09-25 07:00\x1b[0m"
	if got := render(in, time.Now()); got != want {
		t.Errorf("render =\n%q\nwant\n%q", got, want)
	}
}

func TestRenderJoinsOnlyPresentSegments(t *testing.T) {
	in := statusInput{
		modelName: "Opus 5",
		sevenDay:  &quotaWindow{usedPercent: 7},
	}
	want := green + "Opus 5" + reset + separator + green + "7%" + reset + " " + gray + "7d" + reset
	if got := render(in, time.Now()); got != want {
		t.Errorf("render = %q, want %q", got, want)
	}
	if got := render(statusInput{}, time.Now()); got != "" {
		t.Errorf("render of empty input = %q, want empty", got)
	}
}

func TestSafeRecoversPanic(t *testing.T) {
	if got := safe(func() string { panic("segment failed") }); got != "" {
		t.Errorf("safe = %q, want empty", got)
	}
}
