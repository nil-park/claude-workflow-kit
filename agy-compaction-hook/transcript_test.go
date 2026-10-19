package main

import (
	"strings"
	"testing"
)

func TestNeedsReload(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{
			name:  "no checkpoint",
			input: `{"type":"USER_MESSAGE","content":"hello"}` + "\n",
			want:  false,
		},
		{
			name:  "checkpoint with nothing after it",
			input: `{"type":"CHECKPOINT","content":"# Resuming from a compaction"}` + "\n",
			want:  true,
		},
		{
			name: "checkpoint followed by marked SYSTEM_SDK SYSTEM_MESSAGE",
			input: `{"type":"CHECKPOINT","content":"# Resuming from a compaction"}` + "\n" +
				`{"source":"SYSTEM_SDK","type":"SYSTEM_MESSAGE","content":"[agy-compaction-hook]\nreload"}` + "\n",
			want: false,
		},
		{
			name: "marker only inside checkpoint own content",
			input: `{"type":"CHECKPOINT","content":"# Resuming from a compaction with [agy-compaction-hook]"}` + "\n" +
				`{"source":"USER","type":"USER_MESSAGE","content":"next"}` + "\n",
			want: true,
		},
		{
			name: "marker only in MODEL step after checkpoint",
			input: `{"type":"CHECKPOINT","content":"# Resuming from a compaction"}` + "\n" +
				`{"source":"MODEL","type":"MODEL_RESPONSE","content":"echo [agy-compaction-hook]"}` + "\n",
			want: true,
		},
		{
			name: "SYSTEM_MESSAGE with marker but source is SYSTEM",
			input: `{"type":"CHECKPOINT","content":"# Resuming from a compaction"}` + "\n" +
				`{"source":"SYSTEM","type":"SYSTEM_MESSAGE","content":"[agy-compaction-hook]"}` + "\n",
			want: true,
		},
		{
			name: "older checkpoint handled then newer checkpoint with no marked step",
			input: `{"type":"CHECKPOINT","content":"# Resuming 1"}` + "\n" +
				`{"source":"SYSTEM_SDK","type":"SYSTEM_MESSAGE","content":"[agy-compaction-hook]"}` + "\n" +
				`{"type":"CHECKPOINT","content":"# Resuming 2"}` + "\n",
			want: true,
		},
		{
			name: "malformed JSON line is skipped",
			input: `{"type":"CHECKPOINT","content":"# Resuming"}` + "\n" +
				`not-valid-json` + "\n" +
				`{"source":"USER","type":"USER_MESSAGE","content":"hi"}` + "\n",
			want: true,
		},
		{
			name: "line over 64 KiB is handled",
			input: `{"type":"CHECKPOINT","content":"# Resuming"}` + "\n" +
				`{"source":"MODEL","type":"MODEL_RESPONSE","content":"` + strings.Repeat("a", 70000) + `"}` + "\n" +
				`{"source":"SYSTEM_SDK","type":"SYSTEM_MESSAGE","content":"[agy-compaction-hook]"}` + "\n",
			want: false,
		},
		{
			name: "last line without trailing newline",
			input: `{"type":"CHECKPOINT","content":"# Resuming"}` + "\n" +
				`{"source":"SYSTEM_SDK","type":"SYSTEM_MESSAGE","content":"[agy-compaction-hook]"}`,
			want: false,
		},
		{
			name:  "last line checkpoint without trailing newline",
			input: `{"type":"CHECKPOINT","content":"# Resuming"}`,
			want:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := needsReload(strings.NewReader(tc.input))
			if err != nil {
				t.Fatalf("needsReload() unexpected error: %v", err)
			}
			if got != tc.want {
				t.Errorf("needsReload() = %v, want %v", got, tc.want)
			}
		})
	}
}
