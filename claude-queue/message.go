package main

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"strings"
)

type authLine struct {
	Type  string `json:"type"`
	Token string `json:"token"`
}

type inboxMessage struct {
	Version  int         `json:"msgV"`
	ID       string      `json:"msg_id"`
	Type     string      `json:"type"`
	Message  messageBody `json:"message"`
	Priority string      `json:"priority"`
	From     string      `json:"from"`
}

type messageBody struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// inboxPayload returns the bytes written to the socket: an optional auth
// line followed by the message line.
func inboxPayload(token, from, body, msgID string) ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	// The receiver reads the content as text, so "<" and ">" stay unescaped.
	enc.SetEscapeHTML(false)
	if token != "" {
		if err := enc.Encode(authLine{Type: "auth", Token: token}); err != nil {
			return nil, err
		}
	}
	msg := inboxMessage{
		Version:  1,
		ID:       msgID,
		Type:     "user",
		Message:  messageBody{Role: "user", Content: wrapBody(from, body)},
		Priority: "next",
		From:     from,
	}
	if err := enc.Encode(msg); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// wrapBody escapes closing tags in the body so it cannot end the wrapper
// early. from needs no escaping because checkFromName allows only
// [a-z0-9-].
func wrapBody(from, body string) string {
	escaped := strings.ReplaceAll(body, "</cross-session-message", `<\/cross-session-message`)
	return fmt.Sprintf("<cross-session-message from=\"%s\" from-name=\"%s\" from-mode=\"prompting\">\n%s\n</cross-session-message>",
		from, from, escaped)
}

func newMessageID() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	b[6] = b[6]&0x0f | 0x40
	b[8] = b[8]&0x3f | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16]), nil
}
