package main

import (
	"io"
	"testing"
)

// Each case would succeed if its check were removed, so the exit code shows
// that the check itself rejected the arguments.
func TestRunRejectsBadArguments(t *testing.T) {
	stubNotify(t, nil)
	dir := t.TempDir()
	t.Chdir(dir)
	body := writeBody(t, "본문")
	cases := [][]string{
		{},
		{"unknown"},
		{"push", "--dir", ".", "--to", "peer", "--round", "1", "--file", body},
		{"push", "--dir", dir, "--to", "peer", "--round", "0", "--file", body},
		{"push", "--dir", dir, "--to", "peer", "--round", "1", "--file", body, "extra"},
		{"wait", "--dir", dir, "--role", "main", "--round", "1", "--timeout", "0"},
	}
	for _, badName := range []string{"", "Main", "a_b", "a/b", "..", "a b"} {
		cases = append(cases,
			[]string{"push", "--dir", dir, "--to", badName, "--round", "1", "--file", body},
			[]string{"wait", "--dir", dir, "--role", badName, "--round", "1", "--timeout", "1"},
		)
	}
	for _, args := range cases {
		if code := run(args, io.Discard, io.Discard); code != exitError {
			t.Errorf("run(%q) = %d, want %d", args, code, exitError)
		}
	}
}
