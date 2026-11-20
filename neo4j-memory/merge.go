package main

import (
	"fmt"
	"maps"
	"regexp"
	"slices"
	"strings"
)

// Instruction is an Instruction node as the merge step sees it.
type Instruction struct {
	Key   string
	Title string
	Body  string
	// Slots makes this node a variant: it holds the values for the
	// placeholders of the template that includes it, and has no body of its
	// own. An ordinary included node is appended to the including node.
	Slots    map[string]string
	Scopes   []string
	Includes []string
}

func (inst Instruction) isVariant() bool { return len(inst.Slots) > 0 }

type Section struct {
	Key    string
	Title  string
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
// variant is never emitted on its own: without its template it means nothing.
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
		case inst.isVariant():
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
		inst := byKey[key]
		sections = append(sections, Section{Key: key, Title: inst.Title, Scopes: inst.Scopes, Body: body})
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

func (e *expander) warn(format string, args ...any) {
	e.warnings = append(e.warnings, fmt.Sprintf(format, args...))
}

func (e *expander) expand(key string) string {
	e.onPath[key] = true
	e.reached[key] = true
	defer delete(e.onPath, key)

	inst := e.byKey[key]
	values := map[string]string{}
	var appended []string
	children := slices.Clone(inst.Includes)
	slices.Sort(children)
	for _, childKey := range children {
		child, ok := e.byKey[childKey]
		switch {
		case !ok:
		case e.onPath[childKey]:
			e.warn("%s: cycle through %s skipped", key, childKey)
		case child.isVariant():
			e.reached[childKey] = true
			for _, slot := range slices.Sorted(maps.Keys(child.Slots)) {
				if _, taken := values[slot]; taken {
					e.warn("%s: {{%s}} is filled by more than one variant; the value from %s skipped", key, slot, childKey)
					continue
				}
				values[slot] = child.Slots[slot]
			}
		default:
			appended = append(appended, e.expand(childKey))
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
