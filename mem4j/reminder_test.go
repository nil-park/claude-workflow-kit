package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateShortID(t *testing.T) {
	seen := make(map[string]struct{}, 100)
	for range 100 {
		id, err := generateShortID()
		if err != nil {
			t.Fatalf("generateShortID: %v", err)
		}
		if len(id) != shortIDLength {
			t.Fatalf("len(%q) = %d, want %d", id, len(id), shortIDLength)
		}
		for _, ch := range id {
			if !strings.ContainsRune(shortIDAlphabet, ch) {
				t.Fatalf("id %q contains invalid char %q", id, ch)
			}
		}
		if _, dup := seen[id]; dup {
			t.Fatalf("duplicate short id %q", id)
		}
		seen[id] = struct{}{}
	}
}

func TestDefaultGCPCredentialsPath(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	got, err := defaultGCPCredentialsPath()
	if err != nil {
		t.Fatalf("defaultGCPCredentialsPath: %v", err)
	}
	want := filepath.Join(home, ".config", "mem4j", ".secrets", "gcp.json")
	if got != want {
		t.Errorf("defaultGCPCredentialsPath() = %q, want %q", got, want)
	}
}

func TestVertexEmbedderEmbed(t *testing.T) {
	wantVec := make([]float64, embeddingDimensions)
	wantVec[0] = 0.25
	wantVec[embeddingDimensions-1] = 0.75

	var gotReq embedRequest
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if err := json.NewDecoder(r.Body).Decode(&gotReq); err != nil {
			t.Errorf("decode request: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"embedding": map[string]any{"values": wantVec},
		})
	}))
	defer srv.Close()

	emb := &vertexEmbedder{endpoint: srv.URL, client: srv.Client()}
	vec, err := emb.Embed(context.Background(), "when committing code", taskTypeDocument)
	if err != nil {
		t.Fatalf("Embed: %v", err)
	}
	if gotReq.OutputDimensionality != embeddingDimensions || gotReq.TaskType != taskTypeDocument {
		t.Errorf("request = %+v", gotReq)
	}
	if len(gotReq.Content.Parts) != 1 || gotReq.Content.Parts[0].Text != "when committing code" {
		t.Errorf("request parts = %+v", gotReq.Content.Parts)
	}
	if len(vec) != embeddingDimensions || vec[0] != 0.25 || vec[embeddingDimensions-1] != 0.75 {
		t.Errorf("unexpected vector len=%d first=%v last=%v", len(vec), vec[0], vec[len(vec)-1])
	}
}

func TestVertexEmbedderErrors(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/bad-status" {
			http.Error(w, "permission denied", http.StatusForbidden)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"embedding": map[string]any{"values": []float64{0.1, 0.2}},
		})
	}))
	defer srv.Close()

	emb := &vertexEmbedder{endpoint: srv.URL + "/bad-status", client: srv.Client()}
	if _, err := emb.Embed(context.Background(), "   ", taskTypeQuery); err == nil {
		t.Error("expected error for blank text")
	}
	if _, err := emb.Embed(context.Background(), "hello", taskTypeQuery); err == nil || !strings.Contains(err.Error(), "HTTP 403") {
		t.Errorf("bad status err = %v, want HTTP 403", err)
	}
	emb.endpoint = srv.URL + "/bad-dims"
	if _, err := emb.Embed(context.Background(), "hello", taskTypeQuery); err == nil || !strings.Contains(err.Error(), "expected 768") {
		t.Errorf("bad dims err = %v, want dimension mismatch", err)
	}
}

func TestReminderUsageErrorsBeforeConnecting(t *testing.T) {
	cases := map[string]struct {
		args  []string
		stdin string
		want  string
	}{
		"missing action":       {[]string{"reminder"}, "", "usage:"},
		"unknown action":       {[]string{"reminder", "drop"}, "", "unknown reminder action"},
		"create missing title": {[]string{"reminder", "create", "-description", "d", "-scope", "s", "-body", "b"}, "", "-title"},
		"create missing desc":  {[]string{"reminder", "create", "-title", "t", "-scope", "s", "-body", "b"}, "", "-description"},
		"create missing scope": {[]string{"reminder", "create", "-title", "t", "-description", "d", "-body", "b"}, "", "-scope"},
		"create missing body":  {[]string{"reminder", "create", "-title", "t", "-description", "d", "-scope", "s"}, "", "-body or -file"},
		"edit missing key":     {[]string{"reminder", "edit", "-title", "t"}, "", "-key"},
		"edit no fields":       {[]string{"reminder", "edit", "-key", "k"}, "", "at least one of"},
		"edit empty desc":      {[]string{"reminder", "edit", "-key", "k", "-description", ""}, "", "-description cannot be empty"},
		"link missing scope":   {[]string{"reminder", "link", "-key", "k"}, "", "-scope"},
		"unlink missing key":   {[]string{"reminder", "unlink", "-scope", "s"}, "", "-key"},
		"delete missing key":   {[]string{"reminder", "delete"}, "", "-key"},
		"recall missing scope": {[]string{"reminder", "recall", "-query", "q"}, "", "-scope"},
		"recall missing query": {[]string{"reminder", "recall", "-scope", "s"}, "", "-query or -file"},
		"recall bad threshold": {[]string{"reminder", "recall", "-scope", "s", "-query", "q", "-threshold", "2"}, "", "-threshold"},
		"recall bad limit":     {[]string{"reminder", "recall", "-scope", "s", "-query", "q", "-limit", "0"}, "", "-limit"},
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

func TestWriteRemindersAndEmptyNotice(t *testing.T) {
	var buf bytes.Buffer
	writeReminders(&buf, []Reminder{{
		Key:    "abc2345678",
		Title:  "English commits",
		Body:   "Write commit messages in English.",
		Scopes: []string{"git", "user"},
		Score:  0.875,
	}})
	got := buf.String()
	if !strings.Contains(got, `key="abc2345678"`) || !strings.Contains(got, `score="0.875"`) || !strings.Contains(got, "Write commit messages in English.") {
		t.Errorf("writeReminders output = %q", got)
	}
	buf.Reset()
	writeEmptyReminderNotice(&buf, []string{"no-such-scope"})
	if !strings.Contains(buf.String(), "unknown scopes: no-such-scope") {
		t.Errorf("writeEmptyReminderNotice output = %q", buf.String())
	}
}
