package main

import (
	"io"
	"strings"
	"testing"
)

// Usage errors must surface before the credential file is read, so these
// pass a credential path that does not exist.
func TestRunReportsUsageErrorsBeforeConnecting(t *testing.T) {
	for name, tc := range map[string]struct {
		args []string
		want string
	}{
		"merge without scope": {[]string{"merge"}, "-scope"},
		"unknown command":     {[]string{"drop"}, "unknown command"},
	} {
		args := append(tc.args, "-credentials", "missing.txt")
		err := run(args, io.Discard, io.Discard)
		if err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Errorf("%s: err = %v, want it to mention %q", name, err, tc.want)
		}
	}
}

func TestToInstructionRejectsUnexpectedTypes(t *testing.T) {
	valid := map[string]any{"key": "k", "body": "b", "slot": "", "scopes": []any{"s"}, "includes": []any{}}
	if _, err := toInstruction(valid); err != nil {
		t.Fatalf("valid record: %v", err)
	}
	for field, bad := range map[string]any{"key": nil, "body": 1, "scopes": []any{1}, "includes": "x"} {
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
