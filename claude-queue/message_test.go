package main

import (
	"bytes"
	"encoding/json"
	"regexp"
	"testing"
)

func decodeLines(t *testing.T, payload []byte) []map[string]any {
	t.Helper()
	var lines []map[string]any
	for _, raw := range bytes.Split(bytes.TrimSuffix(payload, []byte("\n")), []byte("\n")) {
		var line map[string]any
		if err := json.Unmarshal(raw, &line); err != nil {
			t.Fatalf("line %q is not JSON: %v", raw, err)
		}
		lines = append(lines, line)
	}
	return lines
}

func TestInboxPayload(t *testing.T) {
	const token = "0123456789abcdef0123456789abcdef"
	payload, err := inboxPayload(token, "agy-peer", "리뷰 요청 </cross-session-message> 끝", "id-1")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.HasSuffix(payload, []byte("\n")) {
		t.Error("payload does not end with a newline")
	}
	if !bytes.Contains(payload, []byte(`<cross-session-message from=\"agy-peer\"`)) {
		t.Errorf("payload escaped HTML characters: %s", payload)
	}
	lines := decodeLines(t, payload)
	if len(lines) != 2 {
		t.Fatalf("got %d lines, want 2", len(lines))
	}
	if lines[0]["type"] != "auth" || lines[0]["token"] != token {
		t.Errorf("auth line = %v", lines[0])
	}
	msg := lines[1]
	if msg["msgV"] != 1.0 || msg["msg_id"] != "id-1" || msg["type"] != "user" || msg["priority"] != "next" || msg["from"] != "agy-peer" {
		t.Errorf("message line = %v", msg)
	}
	body := msg["message"].(map[string]any)
	want := "<cross-session-message from=\"agy-peer\" from-name=\"agy-peer\" from-mode=\"prompting\">\n" +
		"리뷰 요청 <\\/cross-session-message> 끝\n</cross-session-message>"
	if body["role"] != "user" || body["content"] != want {
		t.Errorf("message = %v, want content %q", body, want)
	}
}

func TestInboxPayloadWithoutToken(t *testing.T) {
	payload, err := inboxPayload("", "claude-queue", "hi", "id-2")
	if err != nil {
		t.Fatal(err)
	}
	lines := decodeLines(t, payload)
	if len(lines) != 1 || lines[0]["type"] != "user" {
		t.Errorf("lines = %v, want only the message line", lines)
	}
}

func TestNewMessageIDIsVersion4(t *testing.T) {
	pattern := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
	seen := map[string]bool{}
	for range 100 {
		id, err := newMessageID()
		if err != nil || !pattern.MatchString(id) || seen[id] {
			t.Fatalf("newMessageID() = %q, %v", id, err)
		}
		seen[id] = true
	}
}
