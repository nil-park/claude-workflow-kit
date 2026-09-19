package main

import (
	"fmt"
	"math"
	"os"
	"time"
)

const (
	red       = "\033[0;31m"
	green     = "\033[0;32m"
	yellow    = "\033[0;33m"
	magenta   = "\033[0;35m"
	defaultFg = "\033[0;39m"
	gray      = "\033[0;90m"
	reset     = "\033[0m"

	separator = " " + gray + "|" + reset + " "
)

// roundPercent rounds half to even, so 22.5 becomes 22.
func roundPercent(percent float64) int64 {
	return int64(math.RoundToEven(percent))
}

func usageColor(rounded int64, base string) string {
	switch {
	case rounded >= 80:
		return red
	case rounded >= 50:
		return yellow
	}
	return base
}

func segModel(name string) string {
	if name == "" {
		return ""
	}
	return green + name + reset
}

func segContext(inputTokens, usedPercent *float64) string {
	if inputTokens == nil || usedPercent == nil {
		return ""
	}
	pct := roundPercent(*usedPercent)
	return fmt.Sprintf("%s%s (%d%%)%s %sctx%s",
		usageColor(pct, magenta), formatTokens(*inputTokens), pct, reset, gray, reset)
}

func formatTokens(tokens float64) string {
	switch {
	case tokens >= 1_000_000:
		return fmt.Sprintf("%.1fM", tokens/1_000_000)
	case tokens >= 1000:
		return fmt.Sprintf("%.1fk", tokens/1000)
	}
	return fmt.Sprintf("%d", int64(tokens))
}

func segCost(usd *float64) string {
	if usd == nil {
		return ""
	}
	return fmt.Sprintf("%s$%.2f%s %ssession%s", defaultFg, *usd, reset, gray, reset)
}

func segQuota(window *quotaWindow, label, resetLayout string) string {
	if window == nil {
		return ""
	}
	pct := roundPercent(window.usedPercent)
	s := fmt.Sprintf("%s%d%%%s %s%s%s", usageColor(pct, green), pct, reset, gray, label, reset)
	if window.resetsAt != nil {
		s += " " + gray + "→" + reset + " " + defaultFg + window.resetsAt.Format(resetLayout) + reset
	}
	return s
}

func segIdle(transcriptPath string, now time.Time) string {
	if transcriptPath == "" {
		return ""
	}
	info, err := os.Stat(transcriptPath)
	if err != nil || !info.Mode().IsRegular() {
		return ""
	}
	elapsed := max(0, int(now.Sub(info.ModTime()).Seconds()))
	color := green
	switch {
	case elapsed >= 3600:
		color = red
	case elapsed >= 300:
		color = yellow
	}
	return gray + "idle" + reset + " " + color + formatElapsed(elapsed) + reset
}

func formatElapsed(seconds int) string {
	switch {
	case seconds >= 3600:
		return fmt.Sprintf("%dh%dm", seconds/3600, seconds%3600/60)
	case seconds >= 60:
		return fmt.Sprintf("%dm%ds", seconds/60, seconds%60)
	}
	return fmt.Sprintf("%ds", seconds)
}
