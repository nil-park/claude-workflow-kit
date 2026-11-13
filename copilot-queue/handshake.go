package main

import "fmt"

const handshakeTemplate = "- Reply to this session with `claude-queue send --to %s --from copilot-%d --message <message>`; use `--file <path>` for long or multi-line text.\n" +
	"- Messages whose `from` attribute is `%s` come from this session.\n" +
	"- Keep this reply command; every reply needs it."

func handshakeBody(from string, port int) string {
	return fmt.Sprintf(handshakeTemplate, from, port, from)
}
