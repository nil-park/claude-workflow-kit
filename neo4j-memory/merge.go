package main

import (
	"fmt"
	"regexp"
	"slices"
	"strings"
)

// Instruction is an Instruction node as the merge step sees it.
type Instruction struct {
	Key  string
	Body string
	// Slot names the placeholder this node fills in the node that includes it.
	// An empty slot appends the body to that node instead.
	Slot     string
	Scopes   []string
	Includes []string
	// Included nodes are only reached through the node that includes them.
	Included bool
}

type Section struct {
	Key    string
	Scopes []string
	Body   string
}

type MergeResult struct {
	Sections []Section
	Warnings []string
}

var placeholder = regexp.MustCompile(`\{\{([A-Za-z0-9_-]+)\}\}`)

// Merge picks the instructions whose scopes all lie in the active set and
// expands what they include. A node with no scope never applies, because
// every instruction must belong to at least one scope.
func Merge(candidates []Instruction, active []string) MergeResult {
	byKey := map[string]Instruction{}
	for _, inst := range candidates {
		if len(inst.Scopes) > 0 && isSubset(inst.Scopes, active) {
			byKey[inst.Key] = inst
		}
	}
	keys := make([]string, 0, len(byKey))
	for key, inst := range byKey {
		if !inst.Included {
			keys = append(keys, key)
		}
	}
	slices.Sort(keys)

	var result MergeResult
	for _, key := range keys {
		body := expand(key, byKey, map[string]bool{}, &result.Warnings)
		for _, m := range placeholder.FindAllStringSubmatch(body, -1) {
			result.Warnings = append(result.Warnings, fmt.Sprintf("%s: placeholder {{%s}} is not filled", key, m[1]))
		}
		result.Sections = append(result.Sections, Section{Key: key, Scopes: byKey[key].Scopes, Body: body})
	}
	return result
}

// expand tracks the current path so that a cycle stops instead of recursing
// forever.
func expand(key string, byKey map[string]Instruction, onPath map[string]bool, warnings *[]string) string {
	onPath[key] = true
	defer delete(onPath, key)

	inst := byKey[key]
	values := map[string]string{}
	var appended []string
	children := slices.Clone(inst.Includes)
	slices.Sort(children)
	for _, child := range children {
		if _, ok := byKey[child]; !ok {
			continue
		}
		if onPath[child] {
			*warnings = append(*warnings, fmt.Sprintf("%s: cycle through %s skipped", key, child))
			continue
		}
		text := expand(child, byKey, onPath, warnings)
		if slot := byKey[child].Slot; slot != "" {
			values[slot] = text
		} else {
			appended = append(appended, text)
		}
	}
	return strings.Join(append([]string{fillSlots(inst.Body, values)}, appended...), "\n\n")
}

// fillSlots follows the multi-session assembly rule: a line holding only a
// placeholder becomes the value as written, and an empty value removes the
// line.
func fillSlots(body string, values map[string]string) string {
	var out []string
	for _, line := range strings.Split(body, "\n") {
		trimmed := strings.TrimSpace(line)
		if m := placeholder.FindStringSubmatch(trimmed); m != nil && m[0] == trimmed {
			if value, ok := values[m[1]]; ok {
				if value != "" {
					out = append(out, value)
				}
				continue
			}
		}
		out = append(out, placeholder.ReplaceAllStringFunc(line, func(p string) string {
			if value, ok := values[p[2:len(p)-2]]; ok {
				return value
			}
			return p
		}))
	}
	return strings.Join(out, "\n")
}

func isSubset(items, set []string) bool {
	for _, item := range items {
		if !slices.Contains(set, item) {
			return false
		}
	}
	return true
}
