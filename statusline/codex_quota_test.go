package main

import (
	"strconv"
	"testing"
	"time"
)

func TestSelectSnapshot(t *testing.T) {
	snap := func(minute int, used string) codexSnapshot {
		return codexSnapshot{
			ObservedAt: observed.Add(time.Duration(minute) * time.Minute),
			Signals:    map[string]string{"X-Codex-Primary-Used-Percent": used},
		}
	}
	auths := []codexAuth{
		{Name: "b", Quota: snap(9, "newest auth"), ModelQuotas: map[string]codexSnapshot{"gpt-x": snap(1, "b model")}},
		{Name: "c", Quota: snap(0, "c auth"), ModelQuotas: map[string]codexSnapshot{"gpt-x": snap(2, "c model")}},
		{Name: "a", Quota: snap(9, "tied auth"), ModelQuotas: map[string]codexSnapshot{"gpt-y": {}}},
	}
	cases := []struct {
		model string
		want  string
	}{
		{"gpt-x", "c model"},   // model snapshot beats a newer credential snapshot
		{"gpt-y", "tied auth"}, // empty model snapshot falls back; tie goes to name "a"
		{"gpt-z", "tied auth"},
	}
	for _, c := range cases {
		got, ok := selectSnapshot(auths, c.model)
		if !ok || got.Signals["X-Codex-Primary-Used-Percent"] != c.want {
			t.Errorf("selectSnapshot(%q) = %v, %v, want %q", c.model, got, ok, c.want)
		}
	}
	if _, ok := selectSnapshot([]codexAuth{{Name: "x"}}, "gpt-x"); ok {
		t.Error("auths without signals should select nothing")
	}
}

func TestQuotaWindows(t *testing.T) {
	now := observed.Add(time.Hour)
	five := observed.Add(2 * time.Hour).Unix()
	cases := map[string]struct {
		signals             map[string]string
		wantFive, wantSeven string
	}{
		"matched by window length, not by prefix": {map[string]string{
			"X-Codex-Primary-Used-Percent": "40", "X-Codex-Primary-Window-Minutes": "10080",
			"X-Codex-Primary-Reset-After-Seconds": "86400",
			"X-Codex-Secondary-Used-Percent":      "12.5", "X-Codex-Secondary-Window-Minutes": "300",
			"X-Codex-Secondary-Reset-At": itoa(five),
		}, "12.5 @" + itoa(five), "40 @" + itoa(observed.Add(24*time.Hour).Unix())},
		"no reset time keeps the percentage": {map[string]string{
			"X-Codex-Primary-Used-Percent": "3", "X-Codex-Primary-Window-Minutes": "300",
		}, "3", ""},
		"reset already passed": {map[string]string{
			"X-Codex-Primary-Used-Percent": "3", "X-Codex-Primary-Window-Minutes": "300",
			"X-Codex-Primary-Reset-After-Seconds": "60",
		}, "", ""},
		"percentage out of range": {map[string]string{
			"X-Codex-Primary-Used-Percent": "101", "X-Codex-Primary-Window-Minutes": "300",
			"X-Codex-Secondary-Used-Percent": "NaN", "X-Codex-Secondary-Window-Minutes": "10080",
		}, "", ""},
		"window length other than 5h or 7d": {map[string]string{
			"X-Codex-Primary-Used-Percent": "3", "X-Codex-Primary-Window-Minutes": "60",
		}, "", ""},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			five, seven := quotaWindows(codexSnapshot{ObservedAt: observed, Signals: c.signals}, now)
			if got := windowString(five); got != c.wantFive {
				t.Errorf("5h = %q, want %q", got, c.wantFive)
			}
			if got := windowString(seven); got != c.wantSeven {
				t.Errorf("7d = %q, want %q", got, c.wantSeven)
			}
			for _, w := range []*quotaWindow{five, seven} {
				if w != nil && w.resetsAt != nil && w.resetsAt.Location() != time.Local {
					t.Errorf("reset time in %v, want local time for display", w.resetsAt.Location())
				}
			}
		})
	}
}

func windowString(w *quotaWindow) string {
	if w == nil {
		return ""
	}
	s := strconv.FormatFloat(w.usedPercent, 'g', -1, 64)
	if w.resetsAt != nil {
		s += " @" + itoa(w.resetsAt.Unix())
	}
	return s
}

func itoa(n int64) string { return strconv.FormatInt(n, 10) }
