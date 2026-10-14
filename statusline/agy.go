package main

import (
	"math"
	"strings"
	"time"
)

const productAgy = "antigravity"

func isAgy(root map[string]any) bool {
	return text(field(root, "product")) == productAgy
}

func parseAgyInput(root map[string]any) statusInput {
	modelID := text(field(root, "model", "id"))
	modelName := text(field(root, "model", "display_name"))

	isGemini := strings.HasPrefix(modelName, "Gemini")
	key5h := "3p-5h"
	keyWeekly := "3p-weekly"
	if isGemini {
		key5h = "gemini-5h"
		keyWeekly = "gemini-weekly"
	}

	return statusInput{
		product:        productAgy,
		modelID:        modelID,
		modelName:      modelName,
		inputTokens:    number(field(root, "context_window", "total_input_tokens")),
		contextPercent: number(field(root, "context_window", "used_percentage")),
		fiveHour:       parseAgyQuota(field(root, "quota", key5h)),
		sevenDay:       parseAgyQuota(field(root, "quota", keyWeekly)),
		agentState:     text(field(root, "agent_state")),
	}
}

func parseAgyQuota(v any) *quotaWindow {
	rem := number(field(v, "remaining_fraction"))
	if rem == nil {
		return nil
	}

	resetStr := text(field(v, "reset_time"))

	// Placeholder quotas (remaining_fraction: 1 with 1970 reset) are omitted.
	if *rem == 1.0 && resetStr == "1970-01-01T00:00:00Z" {
		return nil
	}

	window := &quotaWindow{
		usedPercent: math.Max(0, math.Min(100, (1.0-*rem)*100.0)),
	}

	if resetStr != "" {
		if t, err := time.Parse(time.RFC3339, resetStr); err == nil {
			local := t.Local()
			window.resetsAt = &local
		}
	}
	return window
}
