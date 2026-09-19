package main

import (
	"math"
	"strconv"
	"time"
)

const (
	fiveHourMinutes = 300
	sevenDayMinutes = 10080
)

// codexWindowPrefixes are the credential's own windows. Additional limits use
// other prefixes (X-Codex-<short name>-*, X-Codex-Additional-<name>-*) and
// carry the same window lengths, so matching on the suffix alone would mix in
// another product's quota.
var codexWindowPrefixes = []string{"X-Codex-Primary-", "X-Codex-Secondary-"}

type codexSnapshot struct {
	ObservedAt time.Time         `json:"observed_at"`
	Signals    map[string]string `json:"signals"`
}

type codexAuth struct {
	Name        string                   `json:"name"`
	Quota       codexSnapshot            `json:"quota"`
	ModelQuotas map[string]codexSnapshot `json:"model_quotas,omitempty"`
}

// selectSnapshot prefers the snapshot recorded for the session's model, since
// the credential-level one follows whichever model was called last.
func selectSnapshot(auths []codexAuth, modelID string) (codexSnapshot, bool) {
	if snap, ok := newestSnapshot(auths, func(a codexAuth) codexSnapshot { return a.ModelQuotas[modelID] }); ok {
		return snap, true
	}
	return newestSnapshot(auths, func(a codexAuth) codexSnapshot { return a.Quota })
}

func newestSnapshot(auths []codexAuth, pick func(codexAuth) codexSnapshot) (codexSnapshot, bool) {
	var best codexSnapshot
	var bestName string
	found := false
	for _, auth := range auths {
		snap := pick(auth)
		if len(snap.Signals) == 0 {
			continue
		}
		newer := snap.ObservedAt.After(best.ObservedAt)
		tie := snap.ObservedAt.Equal(best.ObservedAt) && auth.Name < bestName
		if !found || newer || tie {
			best, bestName, found = snap, auth.Name, true
		}
	}
	return best, found
}

func quotaWindows(snap codexSnapshot, now time.Time) (fiveHour, sevenDay *quotaWindow) {
	for _, prefix := range codexWindowPrefixes {
		minutes, window := codexWindow(snap, prefix, now)
		switch {
		case window == nil:
		case minutes == fiveHourMinutes && fiveHour == nil:
			fiveHour = window
		case minutes == sevenDayMinutes && sevenDay == nil:
			sevenDay = window
		}
	}
	return fiveHour, sevenDay
}

func codexWindow(snap codexSnapshot, prefix string, now time.Time) (int64, *quotaWindow) {
	used, err := strconv.ParseFloat(snap.Signals[prefix+"Used-Percent"], 64)
	if err != nil || math.IsNaN(used) || used < 0 || used > 100 {
		return 0, nil
	}
	minutes, err := strconv.ParseInt(snap.Signals[prefix+"Window-Minutes"], 10, 64)
	if err != nil {
		return 0, nil
	}
	window := &quotaWindow{usedPercent: used}
	if at, ok := codexResetAt(snap, prefix); ok {
		if at.Before(now) {
			return 0, nil
		}
		window.resetsAt = &at
	}
	return minutes, window
}

func codexResetAt(snap codexSnapshot, prefix string) (time.Time, bool) {
	if epoch, err := strconv.ParseInt(snap.Signals[prefix+"Reset-At"], 10, 64); err == nil && epoch > 0 {
		return time.Unix(epoch, 0), true
	}
	after, err := strconv.ParseInt(snap.Signals[prefix+"Reset-After-Seconds"], 10, 64)
	if err != nil || after < 0 || after > sevenDayMinutes*60 || snap.ObservedAt.IsZero() {
		return time.Time{}, false
	}
	return snap.ObservedAt.Add(time.Duration(after) * time.Second).Local(), true
}
