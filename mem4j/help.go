package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"strings"
)

type helpRequest struct {
	text string
}

func (h helpRequest) Error() string { return h.text }

type helpItem struct {
	name    string
	summary string
}

var resourceHelpItems = []helpItem{
	{"scope", "Manage scope hubs (list, create, edit, delete)"},
	{"instruction", "Manage always-on instructions and recall merged output"},
	{"reminder", "Manage prompt-triggered instant reminders with embedding recall"},
	{"graph", "Run read-only Cypher queries against the Neo4j graph"},
}

var actionHelpItems = map[string][]helpItem{
	"scope": {
		{"list", "List all scopes with descriptions"},
		{"create", "Create a new scope"},
		{"edit", "Update a scope description"},
		{"delete", "Delete a scope with no linked instructions or reminders"},
	},
	"instruction": {
		{"recall", "Merge instructions matching the intersection of active scopes"},
		{"create", "Create an instruction linked to one or more scopes"},
		{"edit", "Update an instruction title or body"},
		{"link", "Link an instruction to an additional scope"},
		{"unlink", "Unlink an instruction from a scope"},
		{"delete", "Delete an instruction"},
	},
	"reminder": {
		{"recall", "Find reminders matching active scopes and prompt embedding similarity"},
		{"create", "Create a reminder and embed its trigger description"},
		{"edit", "Update a reminder title, description (re-embedded), or body"},
		{"link", "Link a reminder to an additional scope"},
		{"unlink", "Unlink a reminder from a scope"},
		{"delete", "Delete a reminder"},
	},
	"graph": {
		{"query", "Execute a read-only Cypher query and print JSON Lines"},
	},
}

func isHelpFlag(s string) bool {
	return s == "-h" || s == "--help" || s == "-help"
}

func handleResourceHelp(resource string, args []string, stdin io.Reader) (parsedCommand, bool, error) {
	if isHelpFlag(args[0]) {
		text, _ := resourceHelp(resource)
		return parsedCommand{helpText: text}, true, nil
	}
	if args[0] == "help" {
		cmd, err := parseHelpCommand(append([]string{resource}, args[1:]...), stdin)
		return cmd, true, err
	}
	return parsedCommand{}, false, nil
}

func rootHelp() string {
	var b strings.Builder
	b.WriteString("mem4j manages shared agent memory in a Neo4j graph.\n\n")
	b.WriteString("Usage:\n")
	b.WriteString("  mem4j <resource> <action> [flags]\n")
	b.WriteString("  mem4j help [resource [action]]\n\n")
	b.WriteString("Resources:\n")
	writeHelpItems(&b, resourceHelpItems)
	b.WriteString("\nRun 'mem4j <resource> --help' or 'mem4j <resource> <action> --help' for details.\n")
	return b.String()
}

func resourceHelp(resource string) (string, bool) {
	items, ok := actionHelpItems[resource]
	if !ok {
		return "", false
	}
	var b strings.Builder
	fmt.Fprintf(&b, "Usage:\n  mem4j %s <action> [flags]\n\nActions:\n", resource)
	writeHelpItems(&b, items)
	fmt.Fprintf(&b, "\nRun 'mem4j %s <action> --help' for flags.\n", resource)
	return b.String(), true
}

func actionSummary(fullName string) string {
	resource, action, ok := strings.Cut(fullName, " ")
	if !ok {
		return ""
	}
	for _, item := range actionHelpItems[resource] {
		if item.name == action {
			return item.summary
		}
	}
	return ""
}

func formatFlagHelp(fs *flag.FlagSet) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Usage:\n  mem4j %s [flags]\n", fs.Name())
	if summary := actionSummary(fs.Name()); summary != "" {
		fmt.Fprintf(&b, "\n%s\n", summary)
	}
	var defaults bytes.Buffer
	fs.SetOutput(&defaults)
	fs.PrintDefaults()
	fs.SetOutput(io.Discard)
	if defaults.Len() > 0 {
		b.WriteString("\nFlags:\n")
		b.Write(defaults.Bytes())
	}
	return b.String()
}

func parseHelpCommand(args []string, stdin io.Reader) (parsedCommand, error) {
	switch len(args) {
	case 0:
		return parsedCommand{helpText: rootHelp()}, nil
	case 1:
		text, ok := resourceHelp(args[0])
		if !ok {
			return parsedCommand{}, fmt.Errorf("unknown resource %q (expected scope, instruction, reminder, or graph)", args[0])
		}
		return parsedCommand{helpText: text}, nil
	case 2:
		if _, ok := actionHelpItems[args[0]]; !ok {
			return parsedCommand{}, fmt.Errorf("unknown resource %q (expected scope, instruction, reminder, or graph)", args[0])
		}
		if actionSummary(args[0]+" "+args[1]) == "" {
			return parsedCommand{}, fmt.Errorf("unknown %s action %q", args[0], args[1])
		}
		return parseCommand([]string{args[0], args[1], "-h"}, stdin)
	default:
		return parsedCommand{}, fmt.Errorf("unexpected arguments: %v", args[2:])
	}
}

func writeHelpItems(b *strings.Builder, items []helpItem) {
	width := 0
	for _, item := range items {
		if len(item.name) > width {
			width = len(item.name)
		}
	}
	for _, item := range items {
		fmt.Fprintf(b, "  %-*s  %s\n", width, item.name, item.summary)
	}
}
