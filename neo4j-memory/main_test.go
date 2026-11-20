package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Usage errors must surface before the credential file is read, so every case
// passes a credential path that does not exist.
func TestRunReportsUsageErrorsBeforeConnecting(t *testing.T) {
	cases := map[string]struct {
		args  []string
		stdin string
		want  string
	}{
		"no command":                   {nil, "", "usage:"},
		"unknown resource":             {[]string{"drop"}, "", "unknown resource"},
		"scope missing action":         {[]string{"scope"}, "", "usage:"},
		"scope unknown action":         {[]string{"scope", "drop"}, "", "unknown scope action"},
		"scope list extra arg":         {[]string{"scope", "list", "extra"}, "", "unexpected arguments"},
		"scope create missing desc":    {[]string{"scope", "create", "-name", "s"}, "", "-description"},
		"scope edit missing name":      {[]string{"scope", "edit", "-description", "d"}, "", "-name"},
		"scope delete missing name":    {[]string{"scope", "delete"}, "", "-name"},
		"scope delete extra arg":       {[]string{"scope", "delete", "-name", "s", "extra"}, "", "unexpected arguments"},
		"instruction missing action":   {[]string{"instruction"}, "", "usage:"},
		"instruction unknown action":   {[]string{"instruction", "drop"}, "", "unknown instruction action"},
		"instruction create no scope":  {[]string{"instruction", "create", "-key", "k", "-title", "t", "-body", "b"}, "", "-scope"},
		"instruction create no body":   {[]string{"instruction", "create", "-key", "k", "-title", "t", "-scope", "s"}, "", "-body or -file"},
		"instruction create both body": {[]string{"instruction", "create", "-key", "k", "-title", "t", "-scope", "s", "-body", "b", "-file", "-"}, "x", "only one of -body or -file"},
		"instruction create extra arg": {[]string{"instruction", "create", "-key", "k", "-title", "t", "-scope", "s", "-body", "b", "extra"}, "", "unexpected arguments"},
		"instruction edit no fields":   {[]string{"instruction", "edit", "-key", "k"}, "", "at least one of -title, -body, or -file"},
		"instruction edit empty title": {[]string{"instruction", "edit", "-key", "k", "-title", ""}, "", "cannot be empty"},
		"instruction link no scope":    {[]string{"instruction", "link", "-key", "k"}, "", "-scope"},
		"instruction unlink no key":    {[]string{"instruction", "unlink", "-scope", "s"}, "", "-key"},
		"instruction delete no key":    {[]string{"instruction", "delete"}, "", "-key"},
		"instruction recall no scope":  {[]string{"instruction", "recall"}, "", "-scope"},
		"instruction recall extra arg": {[]string{"instruction", "recall", "-scope", "a", "b"}, "", "unexpected arguments"},
	}
	for name, tc := range cases {
		args := tc.args
		if len(args) >= 2 {
			args = append([]string{args[0], args[1], "-credentials", "missing.txt"}, args[2:]...)
		}
		err := run(args, strings.NewReader(tc.stdin), io.Discard, io.Discard)
		if err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Errorf("%s: err = %v, want it to mention %q", name, err, tc.want)
		}
	}
}

func TestResolveBodyReadsFileAndStdin(t *testing.T) {
	dir := t.TempDir()
	bodyPath := filepath.Join(dir, "body.md")
	if err := os.WriteFile(bodyPath, []byte("line1\r\n가"), 0o600); err != nil {
		t.Fatal(err)
	}
	resolve := func(args []string, stdin io.Reader) (string, bool, error) {
		fs, _ := newFlagSet("test")
		body := fs.String("body", "", "")
		file := fs.String("file", "", "")
		if err := parseFlags(fs, args); err != nil {
			t.Fatalf("parseFlags(%v): %v", args, err)
		}
		return resolveBody(fs, *body, *file, stdin)
	}
	got, ok, err := resolve([]string{"-file", bodyPath}, nil)
	if err != nil || !ok || got != "line1\n가" {
		t.Fatalf("from file: got (%q, %v, %v), want (%q, true, nil)", got, ok, err, "line1\n가")
	}
	got, ok, err = resolve([]string{"-file", "-"}, strings.NewReader("from\r\nstdin"))
	if err != nil || !ok || got != "from\nstdin" {
		t.Fatalf("from stdin: got (%q, %v, %v), want (%q, true, nil)", got, ok, err, "from\nstdin")
	}
	if _, err := parseCommand([]string{"instruction", "create", "-key", "k", "-title", "t", "-scope", "s", "-file", "-"}, strings.NewReader("")); err == nil {
		t.Fatal("expected error for empty stdin body")
	}
}

func TestNormalizeTextComposesHangul(t *testing.T) {
	if got := normalizeText("가\r\n"); got != "가\n" {
		t.Errorf("got %q, want %q", got, "가\n")
	}
}

func TestToInstructionRejectsUnexpectedTypes(t *testing.T) {
	valid := map[string]any{"key": "k", "title": "t", "body": "b", "scopes": []any{"s"}}
	if got, err := toInstruction(valid); err != nil || got.Key != "k" || got.Title != "t" {
		t.Fatalf("valid record: %+v, %v", got, err)
	}
	for field, bad := range map[string]any{"key": nil, "title": nil, "body": 1, "scopes": []any{1}} {
		rec := map[string]any{}
		for k, v := range valid {
			rec[k] = v
		}
		rec[field] = bad
		if _, err := toInstruction(rec); err == nil {
			t.Errorf("%s = %v: expected an error", field, bad)
		}
	}
}

func TestWriteEmptyMergeNotice(t *testing.T) {
	var buf bytes.Buffer
	writeEmptyMergeNotice(&buf, nil)
	if got := buf.String(); !strings.Contains(got, "no instructions apply") || strings.Contains(got, "unknown scopes") {
		t.Errorf("existing scopes notice = %q", got)
	}
	buf.Reset()
	writeEmptyMergeNotice(&buf, []string{"missing-a", "missing-b"})
	if got := buf.String(); !strings.Contains(got, "missing-a") || !strings.Contains(got, "missing-b") {
		t.Errorf("unknown scopes notice = %q, want both scope names", got)
	}
}

func TestWriteScopes(t *testing.T) {
	var buf bytes.Buffer
	writeScopes(&buf, []Scope{{Name: "듀오 세션", Description: "2인 작업", DisableModelInvocation: true}})
	got := buf.String()
	if !strings.Contains(got, `name="듀오 세션"`) || !strings.Contains(got, "disableModelInvocation=true") || !strings.Contains(got, "2인 작업") {
		t.Errorf("writeScopes output = %q", got)
	}
}
