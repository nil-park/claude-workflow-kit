package main

import (
	"slices"
	"strings"
	"testing"
)

func TestMergeRequiresEveryScope(t *testing.T) {
	candidates := []Instruction{
		{Key: "both", Body: "b", Scopes: []string{"user", "repo"}},
		{Key: "user-only", Body: "u", Scopes: []string{"user"}},
		{Key: "no-scope", Body: "n"},
	}
	got := sectionKeys(Merge(candidates, []string{"user"}))
	if !slices.Equal(got, []string{"user-only"}) {
		t.Fatalf("sections = %v, want [user-only]", got)
	}
}

func TestMergeFillsSlotsAndHidesIncludedNodes(t *testing.T) {
	candidates := []Instruction{
		{Key: "tpl", Scopes: []string{"s"}, Includes: []string{"tpl/block", "tpl/empty", "tpl/name", "tpl/tail"},
			Body: "intro {{name}}\n- item\n  {{block}}\n{{empty}}\nend"},
		{Key: "tpl/block", Slot: "block", Body: "    - indented\n    - kept", Scopes: []string{"s"}, Included: true},
		{Key: "tpl/empty", Slot: "empty", Body: "", Scopes: []string{"s"}, Included: true},
		{Key: "tpl/name", Slot: "name", Body: "agy", Scopes: []string{"s"}, Included: true},
		{Key: "tpl/tail", Body: "appended", Scopes: []string{"s"}, Included: true},
	}
	got := Merge(candidates, []string{"s"})
	want := "intro agy\n- item\n    - indented\n    - kept\nend\n\nappended"
	if len(got.Sections) != 1 || got.Sections[0].Body != want || len(got.Warnings) > 0 {
		t.Fatalf("got %+v\nwant body %q", got, want)
	}
}

func TestMergeSkipsInactiveChildAndWarnsOnUnfilledSlot(t *testing.T) {
	candidates := []Instruction{
		{Key: "tpl", Body: "use {{peer}}", Scopes: []string{"s"}, Includes: []string{"tpl/peer"}},
		{Key: "tpl/peer", Slot: "peer", Body: "agy", Scopes: []string{"s", "agy"}, Included: true},
	}
	got := Merge(candidates, []string{"s"})
	if got.Sections[0].Body != "use {{peer}}" || len(got.Warnings) != 1 || !strings.Contains(got.Warnings[0], "{{peer}}") {
		t.Fatalf("got %+v", got)
	}
}

func TestMergeStopsAtCycle(t *testing.T) {
	candidates := []Instruction{
		{Key: "root", Body: "root", Scopes: []string{"s"}, Includes: []string{"a"}},
		{Key: "a", Body: "a", Scopes: []string{"s"}, Includes: []string{"b"}, Included: true},
		{Key: "b", Body: "b", Scopes: []string{"s"}, Includes: []string{"a"}, Included: true},
	}
	got := Merge(candidates, []string{"s"})
	if got.Sections[0].Body != "root\n\na\n\nb" {
		t.Fatalf("body = %q", got.Sections[0].Body)
	}
	if len(got.Warnings) != 1 || !strings.Contains(got.Warnings[0], "cycle") {
		t.Fatalf("warnings = %v", got.Warnings)
	}
}
