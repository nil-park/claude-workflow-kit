package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestSegContextRoundsHalfToEvenAndColors(t *testing.T) {
	cases := []struct {
		percent float64
		want    string
	}{
		{22.5, magenta + "1.0k (22%)"},
		{49.4, magenta + "1.0k (49%)"},
		{49.5, yellow + "1.0k (50%)"},
		{79.4, yellow + "1.0k (79%)"},
		{79.5, red + "1.0k (80%)"},
	}
	for _, c := range cases {
		want := c.want + reset + " " + gray + "ctx" + reset
		if got := segContext(new(1000.0), new(c.percent)); got != want {
			t.Errorf("segContext(%v) = %q, want %q", c.percent, got, want)
		}
	}
}

func TestSegContextNeedsBothFields(t *testing.T) {
	if got := segContext(nil, new(10.0)); got != "" {
		t.Errorf("without tokens = %q", got)
	}
	if got := segContext(new(10.0), nil); got != "" {
		t.Errorf("without percent = %q", got)
	}
}

func TestFormatTokens(t *testing.T) {
	cases := map[float64]string{
		0:         "0",
		999:       "999",
		1000:      "1.0k",
		216_297:   "216.3k",
		999_999:   "1000.0k",
		1_000_000: "1.0M",
		1_250_000: "1.2M",
	}
	for tokens, want := range cases {
		if got := formatTokens(tokens); got != want {
			t.Errorf("formatTokens(%v) = %q, want %q", tokens, got, want)
		}
	}
}

func TestSegCost(t *testing.T) {
	want := defaultFg + "$11.94" + reset + " " + gray + "session" + reset
	if got := segCost(new(11.942875)); got != want {
		t.Errorf("segCost = %q, want %q", got, want)
	}
	if got := segCost(nil); got != "" {
		t.Errorf("segCost(nil) = %q", got)
	}
}

func TestSegQuota(t *testing.T) {
	at := time.Date(2026, 1, 5, 3, 7, 0, 0, kst)
	arrow := " " + gray + "→" + reset + " "
	cases := []struct {
		name   string
		window *quotaWindow
		label  string
		layout string
		want   string
	}{
		{"5h pads hour and minute", &quotaWindow{usedPercent: 8, resetsAt: &at}, "5h", "15:04",
			green + "8%" + reset + " " + gray + "5h" + reset + arrow + defaultFg + "03:07" + reset},
		{"7d pads month and day", &quotaWindow{usedPercent: 50.5, resetsAt: &at}, "7d", "01-02 15:04",
			yellow + "50%" + reset + " " + gray + "7d" + reset + arrow + defaultFg + "01-05 03:07" + reset},
		{"no reset time", &quotaWindow{usedPercent: 80}, "5h", "15:04",
			red + "80%" + reset + " " + gray + "5h" + reset},
		{"no window", nil, "5h", "15:04", ""},
	}
	for _, c := range cases {
		if got := segQuota(c.window, c.label, c.layout); got != c.want {
			t.Errorf("%s: segQuota = %q, want %q", c.name, got, c.want)
		}
	}
}

func TestSegIdle(t *testing.T) {
	path := filepath.Join(t.TempDir(), "transcript.jsonl")
	if err := os.WriteFile(path, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	modified := time.Date(2026, 9, 19, 12, 0, 0, 0, time.UTC)
	if err := os.Chtimes(path, modified, modified); err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		elapsed time.Duration
		want    string
	}{
		{-5 * time.Second, green + "0s"},
		{59*time.Second + 900*time.Millisecond, green + "59s"},
		{60 * time.Second, green + "1m0s"},
		{299 * time.Second, green + "4m59s"},
		{300 * time.Second, yellow + "5m0s"},
		{3599 * time.Second, yellow + "59m59s"},
		{3600 * time.Second, red + "1h0m"},
		{2*time.Hour + 15*time.Minute + 59*time.Second, red + "2h15m"},
	}
	for _, c := range cases {
		want := gray + "idle" + reset + " " + c.want + reset
		if got := segIdle(path, modified.Add(c.elapsed)); got != want {
			t.Errorf("elapsed %v: segIdle = %q, want %q", c.elapsed, got, want)
		}
	}
}

func TestSegIdleWithoutRegularFile(t *testing.T) {
	dir := t.TempDir()
	for _, path := range []string{"", filepath.Join(dir, "missing.jsonl"), dir} {
		if got := segIdle(path, time.Now()); got != "" {
			t.Errorf("segIdle(%q) = %q, want empty", path, got)
		}
	}
}
