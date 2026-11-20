package main

import (
	"encoding/json"
	"fmt"
)

// toInstruction rejects a record whose fields have unexpected types, since
// other agents also write to the shared graph.
func toInstruction(m map[string]any) (Instruction, error) {
	key, ok := m["key"].(string)
	if !ok {
		return Instruction{}, fmt.Errorf("instruction key has type %T", m["key"])
	}
	title, okTitle := m["title"].(string)
	body, okBody := m["body"].(string)
	rawSlots, okSlots := m["slots"].(string)
	scopes, errScopes := toStrings(m["scopes"])
	includes, errIncludes := toStrings(m["includes"])
	if !okTitle || !okBody || !okSlots || errScopes != nil || errIncludes != nil {
		return Instruction{}, fmt.Errorf("instruction %s has a field of an unexpected type", key)
	}
	var slots map[string]string
	if rawSlots != "" {
		if err := json.Unmarshal([]byte(rawSlots), &slots); err != nil {
			return Instruction{}, fmt.Errorf("instruction %s: slots: %w", key, err)
		}
	}
	return Instruction{Key: key, Title: title, Body: body, Slots: slots, Scopes: scopes, Includes: includes}, nil
}

func scopeParams(scopes []ScopeSpec) []map[string]any {
	out := make([]map[string]any, len(scopes))
	for i, s := range scopes {
		out[i] = map[string]any{"name": s.Name, "description": s.Description, "disableModelInvocation": s.DisableModelInvocation}
	}
	return out
}

// Neo4j has no map property type, so a variant's slots are stored as a JSON
// object string.
func instructionParams(insts []Instruction) ([]map[string]any, error) {
	out := make([]map[string]any, len(insts))
	for i, inst := range insts {
		slots := ""
		if inst.isVariant() {
			b, err := json.Marshal(inst.Slots)
			if err != nil {
				return nil, err
			}
			slots = string(b)
		}
		out[i] = map[string]any{
			"key": inst.Key, "title": inst.Title, "body": inst.Body, "slots": slots,
			"scopes": inst.Scopes, "includes": nonNil(inst.Includes),
		}
	}
	return out, nil
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

func nonNil(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}
