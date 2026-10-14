package main

import (
	"math"
	"os"
	"strings"
	"testing"
	"time"
)

func TestParseInputAgyPersonal(t *testing.T) {
	raw, err := os.ReadFile("testdata/captured-stdin-agy-personal.json")
	if err != nil {
		t.Fatal(err)
	}
	in := parseInput(raw)

	if in.product != "antigravity" {
		t.Errorf("product = %q, want %q", in.product, "antigravity")
	}
	if in.modelID != "Gemini 3.8 Flash (Medium)" || in.modelName != "Gemini 3.8 Flash (Medium)" {
		t.Errorf("modelID = %q, modelName = %q", in.modelID, in.modelName)
	}
	if in.inputTokens == nil || *in.inputTokens != 19473 {
		t.Errorf("inputTokens = %v, want 19473", in.inputTokens)
	}
	if in.contextPercent == nil || math.Abs(*in.contextPercent-1.8570899963378906) > 1e-6 {
		t.Errorf("contextPercent = %v", in.contextPercent)
	}
	if in.costUSD != nil {
		t.Errorf("costUSD = %v, want nil", in.costUSD)
	}
	if in.agentState != "working" {
		t.Errorf("agentState = %q, want %q", in.agentState, "working")
	}
	if in.fiveHour == nil {
		t.Fatal("fiveHour is nil, want gemini-5h parsed")
	}
	// remaining_fraction: 0.9881002 -> used: ~1.18998%
	if math.Abs(in.fiveHour.usedPercent-1.18998) > 1e-4 {
		t.Errorf("fiveHour.usedPercent = %f, want ~1.18998", in.fiveHour.usedPercent)
	}
	if in.fiveHour.resetsAt == nil || in.fiveHour.resetsAt.UTC().Format(time.RFC3339) != "2026-10-14T17:31:13Z" {
		t.Errorf("fiveHour.resetsAt = %v, want 2026-10-14T17:31:13Z", in.fiveHour.resetsAt)
	}

	if in.sevenDay == nil {
		t.Fatal("sevenDay is nil, want gemini-weekly parsed")
	}
	// remaining_fraction: 0.9926717 -> used: ~0.73283%
	if math.Abs(in.sevenDay.usedPercent-0.73283) > 1e-4 {
		t.Errorf("sevenDay.usedPercent = %f, want ~0.73283", in.sevenDay.usedPercent)
	}
	if in.sevenDay.resetsAt == nil || in.sevenDay.resetsAt.UTC().Format(time.RFC3339) != "2026-10-21T12:31:13Z" {
		t.Errorf("sevenDay.resetsAt = %v, want 2026-10-21T12:31:13Z", in.sevenDay.resetsAt)
	}
}

func TestParseInputAgyEnterprise(t *testing.T) {
	raw, err := os.ReadFile("testdata/captured-stdin-agy-enterprise.json")
	if err != nil {
		t.Fatal(err)
	}
	in := parseInput(raw)

	if in.product != "antigravity" {
		t.Errorf("product = %q, want %q", in.product, "antigravity")
	}
	if in.agentState != "idle" {
		t.Errorf("agentState = %q, want %q", in.agentState, "idle")
	}
	if in.fiveHour != nil {
		t.Errorf("fiveHour = %+v, want nil (omitted placeholder)", in.fiveHour)
	}
	if in.sevenDay != nil {
		t.Errorf("sevenDay = %+v, want nil (omitted placeholder)", in.sevenDay)
	}
}

func TestParseInputAgy3PModel(t *testing.T) {
	raw := []byte(`{
		"product": "antigravity",
		"model": {"id": "claude-3-7-sonnet", "display_name": "Claude 3.7 Sonnet"},
		"quota": {
			"3p-5h": {"remaining_fraction": 0.8, "reset_time": "2026-10-14T18:00:00Z"},
			"3p-weekly": {"remaining_fraction": 0.9, "reset_time": "2026-10-21T18:00:00Z"},
			"gemini-5h": {"remaining_fraction": 0.99, "reset_time": "2026-10-14T19:00:00Z"},
			"gemini-weekly": {"remaining_fraction": 0.99, "reset_time": "2026-10-21T19:00:00Z"}
		}
	}`)
	in := parseInput(raw)

	if in.fiveHour == nil || math.Abs(in.fiveHour.usedPercent-20.0) > 1e-4 {
		t.Fatalf("fiveHour = %+v, want 20%% used", in.fiveHour)
	}
	if in.sevenDay == nil || math.Abs(in.sevenDay.usedPercent-10.0) > 1e-4 {
		t.Fatalf("sevenDay = %+v, want 10%% used", in.sevenDay)
	}
}

func TestRenderAgy(t *testing.T) {
	reset5h := time.Date(2026, 10, 15, 2, 31, 0, 0, kst)
	resetWeekly := time.Date(2026, 10, 21, 21, 31, 0, 0, kst)

	in := statusInput{
		product:        "antigravity",
		modelName:      "Gemini 3.8 Flash (Medium)",
		inputTokens:    new(19_473.0),
		contextPercent: new(1.857),
		fiveHour:       &quotaWindow{usedPercent: 1.19, resetsAt: &reset5h},
		sevenDay:       &quotaWindow{usedPercent: 0.73, resetsAt: &resetWeekly},
		agentState:     "working",
	}

	got := render(in, time.Now())

	// Model and cost segments must NOT appear.
	if strings.Contains(got, "Gemini") || strings.Contains(got, "session") {
		t.Errorf("renderAgy should not contain model or session cost: %q", got)
	}

	// Must contain context, quota with "weekly" label, and working state in green.
	want := "\x1b[0;35m19.5k (2%)\x1b[0m \x1b[0;90mctx\x1b[0m" +
		" \x1b[0;90m|\x1b[0m \x1b[0;32m1%\x1b[0m \x1b[0;90m5h\x1b[0m \x1b[0;90m→\x1b[0m \x1b[0;39m02:31\x1b[0m" +
		" \x1b[0;90m|\x1b[0m \x1b[0;32m1%\x1b[0m \x1b[0;90mweekly\x1b[0m \x1b[0;90m→\x1b[0m \x1b[0;39m10-21 21:31\x1b[0m" +
		" \x1b[0;90m|\x1b[0m \x1b[0;32mworking\x1b[0m"
	if got != want {
		t.Errorf("renderAgy =\n%q\nwant\n%q", got, want)
	}
}

func TestSegState(t *testing.T) {
	if got := segState(""); got != "" {
		t.Errorf("segState(\"\") = %q, want empty", got)
	}
	if got := segState("working"); got != green+"working"+reset {
		t.Errorf("segState(\"working\") = %q, want green working", got)
	}
	if got := segState("idle"); got != defaultFg+"idle"+reset {
		t.Errorf("segState(\"idle\") = %q, want defaultFg idle", got)
	}
	if got := segState("authenticating"); got != defaultFg+"authenticating"+reset {
		t.Errorf("segState(\"authenticating\") = %q, want defaultFg authenticating", got)
	}
}
