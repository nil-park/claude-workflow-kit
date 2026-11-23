package main

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

func TestValidateRelType(t *testing.T) {
	for _, valid := range []string{"USES", "DEPENDS_ON", "BLOCKS_READ", "V2_CALLS"} {
		if err := validateRelType(valid); err != nil {
			t.Errorf("validateRelType(%q) = %v, want nil", valid, err)
		}
	}
	for _, bad := range []string{"", "uses", "USES-DB", "123", "APPLIES_IN", "ABOUT", "A B"} {
		if err := validateRelType(bad); err == nil {
			t.Errorf("validateRelType(%q) = nil, want error", bad)
		}
	}
}

func TestNormalizeOccurredAt(t *testing.T) {
	got, err := normalizeOccurredAt("2026-11-17")
	if err != nil || got != "2026-11-17" {
		t.Fatalf("date only: got (%q, %v)", got, err)
	}
	got, err = normalizeOccurredAt("2026-11-23T23:30:00+09:00")
	if err != nil || got != "2026-11-23T14:30:00Z" {
		t.Fatalf("kst to utc: got (%q, %v)", got, err)
	}
	if _, err := normalizeOccurredAt("2026/11/17"); err == nil {
		t.Error("expected error for invalid date format")
	}
}

func TestBackDataUsageErrorsBeforeConnecting(t *testing.T) {
	cases := map[string]struct {
		args []string
		want string
	}{
		"entity missing action":       {[]string{"entity"}, "usage:"},
		"entity create missing desc":  {[]string{"entity", "create", "-name", "e"}, "-description"},
		"entity recall no query":      {[]string{"entity", "recall"}, "-query or -file"},
		"entity link reserved rel":    {[]string{"entity", "link", "-from", "a", "-to", "b", "-type", "ABOUT"}, "reserved"},
		"entity link self":            {[]string{"entity", "link", "-from", "a", "-to", "a", "-type", "USES"}, "cannot be the same"},
		"entity neighbors bad depth":  {[]string{"entity", "neighbors", "-name", "a", "-depth", "4"}, "between 1 and 3"},
		"scope link reserved rel":     {[]string{"scope", "link", "-from", "a", "-to", "b", "-type", "APPLIES_IN"}, "reserved"},
		"statement create no title":   {[]string{"statement", "create", "-entity", "e", "-body", "b"}, "-title"},
		"statement create no entity":  {[]string{"statement", "create", "-title", "t", "-body", "b"}, "-entity"},
		"statement edit no fields":    {[]string{"statement", "edit", "-key", "k"}, "at least one of -title, -description, -body, or -file"},
		"statement edit empty title":  {[]string{"statement", "edit", "-key", "k", "-title", ""}, "cannot be empty"},
		"statement edit empty desc":   {[]string{"statement", "edit", "-key", "k", "-description", ""}, "cannot be empty"},
		"statement create empty desc": {[]string{"statement", "create", "-title", "t", "-entity", "e", "-body", "b", "-description", ""}, "cannot be empty"},
		"statement list no entity":    {[]string{"statement", "list"}, "-entity"},
		"event create no occurred-at": {[]string{"event", "create", "-title", "t", "-entity", "e", "-body", "b"}, "-occurred-at"},
		"event create bad date":       {[]string{"event", "create", "-title", "t", "-occurred-at", "bad", "-entity", "e", "-body", "b"}, "invalid -occurred-at"},
		"event edit no fields":        {[]string{"event", "edit", "-key", "k"}, "-occurred-at"},
	}
	for name, tc := range cases {
		args := tc.args
		if len(args) >= 2 {
			args = append([]string{args[0], args[1], "-credentials", "missing.txt"}, args[2:]...)
		}
		err := run(args, strings.NewReader(""), io.Discard, io.Discard)
		if err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Errorf("%s: err = %v, want it to mention %q", name, err, tc.want)
		}
	}
}

func TestWriteBackDataFormats(t *testing.T) {
	var buf bytes.Buffer
	writeEntities(&buf, []Entity{{Name: "mem4j", Description: "CLI", IsScope: false, Score: 0.912}}, true)
	writeStatements(&buf, []Statement{{Key: "k1", Title: "T1", Body: "B1", Entities: []string{"mem4j"}}}, false)
	writeEvents(&buf, []Event{{Key: "e1", Title: "ET1", OccurredAt: "2026-11-23", Body: "EB1", Entities: []string{"mem4j"}}}, false)
	writeNeighbors(&buf, NeighborView{
		Root:  "mem4j",
		Edges: []HubEdge{{From: "mem4j", Type: "USES", To: "Neo4j AuraDB"}},
	})
	got := buf.String()
	for _, want := range []string{`name="mem4j"`, `score="0.912"`, `title="T1"`, `occurredAt="2026-11-23"`, `type="USES"`} {
		if !strings.Contains(got, want) {
			t.Errorf("output missing %q:\n%s", want, got)
		}
	}
}
