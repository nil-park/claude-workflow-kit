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
//
// A node is emitted on its own unless an active node includes it, so an
// instruction stays visible when only the node including it is inactive. A
// slot value is never emitted on its own: without its template it means
// nothing.
func Merge(candidates []Instruction, active []string) MergeResult {
	byKey := map[string]Instruction{}
	for _, inst := range candidates {
		if len(inst.Scopes) > 0 && isSubset(inst.Scopes, active) {
			byKey[inst.Key] = inst
		}
	}
	includedByActive := map[string]bool{}
	for _, inst := range byKey {
		for _, child := range inst.Includes {
			includedByActive[child] = true
		}
	}
	var roots, rest []string
	for key, inst := range byKey {
		switch {
		case inst.Slot != "":
		case includedByActive[key]:
			rest = append(rest, key)
		default:
			roots = append(roots, key)
		}
	}
	slices.Sort(roots)
	slices.Sort(rest)

	e := expander{byKey: byKey, onPath: map[string]bool{}, reached: map[string]bool{}}
	var sections []Section
	emit := func(key string) {
		body := e.expand(key)
		for _, m := range placeholder.FindAllStringSubmatch(body, -1) {
			e.warnings = append(e.warnings, fmt.Sprintf("%s: placeholder {{%s}} is not filled", key, m[1]))
		}
		sections = append(sections, Section{Key: key, Scopes: byKey[key].Scopes, Body: body})
	}
	for _, key := range roots {
		emit(key)
	}
	// Nodes that include each other with no root above them would otherwise
	// vanish; the first unreached one stands in as their root.
	for _, key := range rest {
		if !e.reached[key] {
			e.warnings = append(e.warnings, fmt.Sprintf("%s: reachable only through a cycle, emitted on its own", key))
			emit(key)
		}
	}
	return MergeResult{Sections: sections, Warnings: e.warnings}
}

type expander struct {
	byKey map[string]Instruction
	// onPath holds the current include path so that a cycle stops instead of
	// recursing forever; reached holds every node expanded so far.
	onPath   map[string]bool
	reached  map[string]bool
	warnings []string
}

func (e *expander) expand(key string) string {
	e.onPath[key] = true
	e.reached[key] = true
	defer delete(e.onPath, key)

	byKey, warnings := e.byKey, &e.warnings
	inst := byKey[key]
	values := map[string]string{}
	var appended []string
	children := slices.Clone(inst.Includes)
	slices.Sort(children)
	for _, child := range children {
		if _, ok := byKey[child]; !ok {
			continue
		}
		if e.onPath[child] {
			*warnings = append(*warnings, fmt.Sprintf("%s: cycle through %s skipped", key, child))
			continue
		}
		text := e.expand(child)
		if slot := byKey[child].Slot; slot != "" {
			if _, taken := values[slot]; taken {
				*warnings = append(*warnings, fmt.Sprintf("%s: {{%s}} is filled by more than one node; %s skipped", key, slot, child))
				continue
			}
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
