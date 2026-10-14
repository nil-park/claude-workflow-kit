package main

import (
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

	var fiveHour, sevenDay *quotaWindow
	if modelID != "" || modelName != "" {
		key5h := "3p-5h"
		keyWeekly := "3p-weekly"
		if strings.HasPrefix(modelID, "Gemini") || strings.HasPrefix(modelName, "Gemini") {
			key5h = "gemini-5h"
			keyWeekly = "gemini-weekly"
		}
		fiveHour = parseAgyQuota(field(root, "quota", key5h))
		sevenDay = parseAgyQuota(field(root, "quota", keyWeekly))
	}

	return statusInput{
		product:        productAgy,
		modelID:        modelID,
		modelName:      modelName,
		inputTokens:    number(field(root, "context_window", "total_input_tokens")),
		contextPercent: number(field(root, "context_window", "used_percentage")),
		fiveHour:       fiveHour,
		sevenDay:       sevenDay,
		agentState:     text(field(root, "agent_state")),
	}
}

func parseAgyQuota(v any) *quotaWindow {
	rem := number(field(v, "remaining_fraction"))
	if rem == nil {
		return nil
	}

	var resetAt *time.Time
	resetStr := text(field(v, "reset_time"))
	if resetStr != "" {
		if t, err := time.Parse(time.RFC3339, resetStr); err == nil {
			if *rem == 1.0 && t.Unix() <= 0 {
				return nil
			}
			local := t.Local()
			resetAt = &local
		}
	}

	return &quotaWindow{
		usedPercent: (1.0 - *rem) * 100.0,
		resetsAt:    resetAt,
	}
}
