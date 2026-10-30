package main

import (
	"fmt"
	"os"
	"strings"
)

const handshakeTemplate = `- Reply to this session with: agy agentapi send-message %s <message>
  Set these environment variables on that command:
  ANTIGRAVITY_CONVERSATION_ID=%s
  ANTIGRAVITY_LS_ADDRESS=%s
  ANTIGRAVITY_CSRF_TOKEN=%s
- Messages whose from attribute is %s come from this session.
- Save ANTIGRAVITY_CONVERSATION_ID, ANTIGRAVITY_LS_ADDRESS, and ANTIGRAVITY_CSRF_TOKEN in the harness scratchpad so they survive context compaction.`

func agyFrom(id string) string {
	return "agy-" + strings.ToLower(id)
}

func handshakeBody() (string, error) {
	vals := make([]string, 3)
	for i, name := range []string{
		"ANTIGRAVITY_CONVERSATION_ID",
		"ANTIGRAVITY_LS_ADDRESS",
		"ANTIGRAVITY_CSRF_TOKEN",
	} {
		vals[i] = os.Getenv(name)
		if vals[i] == "" {
			return "", fmt.Errorf("%s is empty", name)
		}
	}
	return fmt.Sprintf(handshakeTemplate, vals[0], vals[0], vals[1], vals[2], agyFrom(vals[0])), nil
}
