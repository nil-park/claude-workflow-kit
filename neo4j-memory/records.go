package main

import (
	"fmt"
	"strings"

	"golang.org/x/text/unicode/norm"
)

// normalizeText unifies line endings and Hangul composition so that text read
// on any OS compares equal.
func normalizeText(text string) string {
	return norm.NFC.String(strings.ReplaceAll(text, "\r\n", "\n"))
}

// toInstruction rejects a record whose fields have unexpected types, since
// other agents also write to the shared graph.
func toInstruction(m map[string]any) (Instruction, error) {
	key, ok := m["key"].(string)
	if !ok {
		return Instruction{}, fmt.Errorf("instruction key has type %T", m["key"])
	}
	title, okTitle := m["title"].(string)
	body, okBody := m["body"].(string)
	scopes, errScopes := toStrings(m["scopes"])
	if !okTitle || !okBody || errScopes != nil {
		return Instruction{}, fmt.Errorf("instruction %s has a field of an unexpected type", key)
	}
	return Instruction{Key: key, Title: title, Body: body, Scopes: scopes}, nil
}

func toStrings(v any) ([]string, error) {
	items, ok := v.([]any)
	if !ok {
		return nil, fmt.Errorf("expected a list, got %T", v)
	}
	out := make([]string, 0, len(items))
	for _, item := range items {
		s, ok := item.(string)
		if !ok {
			return nil, fmt.Errorf("expected a string, got %T", item)
		}
		out = append(out, s)
	}
	return out, nil
}
