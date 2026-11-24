package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestRunHelpSucceedsWithoutCredentials(t *testing.T) {
	cases := map[string]struct {
		args []string
		want []string
	}{
		"root -h":           {[]string{"-h"}, []string{"Resources:", "scope", "instruction", "reminder", "graph"}},
		"root --help":       {[]string{"--help"}, []string{"Usage:", "mem4j help [resource [action]]"}},
		"root help":         {[]string{"help"}, []string{"Resources:", "reminder"}},
		"scope --help":      {[]string{"scope", "--help"}, []string{"mem4j scope <action>", "list", "create", "edit", "delete"}},
		"help scope":        {[]string{"help", "scope"}, []string{"mem4j scope <action>", "delete"}},
		"instruction -h":    {[]string{"instruction", "-h"}, []string{"mem4j instruction <action>", "recall", "unlink"}},
		"reminder --help":   {[]string{"reminder", "--help"}, []string{"mem4j reminder <action>", "recall", "create"}},
		"graph -h":          {[]string{"graph", "-h"}, []string{"mem4j graph <action>", "query"}},
		"hook -h":           {[]string{"hook", "-h"}, []string{"mem4j hook <action>", "claude-session-start", "agy-pre-invocation"}},
		"hook action -h":    {[]string{"hook", "agy-pre-invocation", "-h"}, []string{"mem4j hook agy-pre-invocation [flags]", "-config", "-threshold", "-state-dir", "-log"}},
		"action -h":         {[]string{"reminder", "recall", "-h"}, []string{"mem4j reminder recall [flags]", "-threshold", "-limit", "-gcp-credentials", "-scope"}},
		"action --help":     {[]string{"scope", "create", "--help"}, []string{"mem4j scope create [flags]", "-name", "-description", "-disable-model-invocation"}},
		"help action":       {[]string{"help", "graph", "query"}, []string{"mem4j graph query [flags]", "-cypher", "-file"}},
		"resource help act": {[]string{"reminder", "help", "create"}, []string{"mem4j reminder create [flags]", "-description", "-key"}},
	}
	for name, tc := range cases {
		var stdout, stderr bytes.Buffer
		if err := run(tc.args, strings.NewReader(""), &stdout, &stderr); err != nil {
			t.Errorf("%s: unexpected error: %v", name, err)
			continue
		}
		if stderr.Len() != 0 {
			t.Errorf("%s: stderr = %q, want empty", name, stderr.String())
		}
		got := stdout.String()
		for _, want := range tc.want {
			if !strings.Contains(got, want) {
				t.Errorf("%s: stdout missing %q:\n%s", name, want, got)
			}
		}
	}
}

func TestHelpSubcommandMatchesFlagHelp(t *testing.T) {
	var viaSubcmd, viaFlag bytes.Buffer
	if err := run([]string{"help", "reminder", "recall"}, strings.NewReader(""), &viaSubcmd, &bytes.Buffer{}); err != nil {
		t.Fatalf("help subcommand failed: %v", err)
	}
	if err := run([]string{"reminder", "recall", "--help"}, strings.NewReader(""), &viaFlag, &bytes.Buffer{}); err != nil {
		t.Fatalf("--help failed: %v", err)
	}
	if viaSubcmd.String() != viaFlag.String() {
		t.Errorf("help outputs differ:\n--- subcommand ---\n%s\n--- flag ---\n%s", viaSubcmd.String(), viaFlag.String())
	}
}

func TestHelpRejectsUnknownTargets(t *testing.T) {
	cases := [][]string{
		{"help", "unknown"},
		{"help", "scope", "unknown"},
		{"help", "scope", "help"},
		{"help", "help", "help"},
		{"help", "scope", "list", "extra"},
	}
	for _, args := range cases {
		var stdout, stderr bytes.Buffer
		if err := run(args, strings.NewReader(""), &stdout, &stderr); err == nil {
			t.Errorf("run(%v) = nil, want error", args)
		}
	}
}
