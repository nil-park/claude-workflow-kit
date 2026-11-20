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
		{Key: "tpl/block", Slot: "block", Body: "    - indented\n    - kept", Scopes: []string{"s"}},
		{Key: "tpl/empty", Slot: "empty", Body: "", Scopes: []string{"s"}},
		{Key: "tpl/name", Slot: "name", Body: "agy", Scopes: []string{"s"}},
		{Key: "tpl/tail", Body: "appended", Scopes: []string{"s"}},
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
		{Key: "tpl/peer", Slot: "peer", Body: "agy", Scopes: []string{"s", "agy"}},
	}
	got := Merge(candidates, []string{"s"})
	if got.Sections[0].Body != "use {{peer}}" || len(got.Warnings) != 1 || !strings.Contains(got.Warnings[0], "{{peer}}") {
		t.Fatalf("got %+v", got)
	}
}

// When the including node is inactive, an ordinary instruction it includes
// still applies on its own, but a slot value does not.
func TestMergeEmitsChildOfInactiveParentUnlessSlot(t *testing.T) {
	candidates := []Instruction{
		{Key: "parent", Body: "p {{x}}", Scopes: []string{"user", "repo"}, Includes: []string{"shared", "parent/x"}},
		{Key: "shared", Body: "shared", Scopes: []string{"user"}},
		{Key: "parent/x", Slot: "x", Body: "x", Scopes: []string{"user"}},
	}
	got := sectionKeys(Merge(candidates, []string{"user"}))
	if !slices.Equal(got, []string{"shared"}) {
		t.Fatalf("sections = %v, want [shared]", got)
	}
}

func TestMergeWarnsWhenTwoNodesFillOneSlot(t *testing.T) {
	candidates := []Instruction{
		{Key: "tpl", Body: "{{peer}}", Scopes: []string{"s"}, Includes: []string{"tpl@a/peer", "tpl@b/peer"}},
		{Key: "tpl@a/peer", Slot: "peer", Body: "a", Scopes: []string{"s"}},
		{Key: "tpl@b/peer", Slot: "peer", Body: "b", Scopes: []string{"s"}},
	}
	got := Merge(candidates, []string{"s"})
	if got.Sections[0].Body != "a" || len(got.Warnings) != 1 || !strings.Contains(got.Warnings[0], "tpl@b/peer") {
		t.Fatalf("got %+v", got)
	}
}

func TestMergeEmitsRootlessCycleAndSelfLoop(t *testing.T) {
	candidates := []Instruction{
		{Key: "a", Body: "a", Scopes: []string{"s"}, Includes: []string{"b"}},
		{Key: "b", Body: "b", Scopes: []string{"s"}, Includes: []string{"a"}},
		{Key: "self", Body: "self", Scopes: []string{"s"}, Includes: []string{"self"}},
	}
	got := Merge(candidates, []string{"s"})
	if keys := sectionKeys(got); !slices.Equal(keys, []string{"a", "self"}) || got.Sections[0].Body != "a\n\nb" {
		t.Fatalf("sections = %+v", got.Sections)
	}
	if len(got.Warnings) != 4 {
		t.Fatalf("warnings = %v, want a stand-in and a cycle warning for each loop", got.Warnings)
	}
}

func TestMergeStopsAtCycle(t *testing.T) {
	candidates := []Instruction{
		{Key: "root", Body: "root", Scopes: []string{"s"}, Includes: []string{"a"}},
		{Key: "a", Body: "a", Scopes: []string{"s"}, Includes: []string{"b"}},
		{Key: "b", Body: "b", Scopes: []string{"s"}, Includes: []string{"a"}},
	}
	got := Merge(candidates, []string{"s"})
	if len(got.Sections) != 1 || got.Sections[0].Body != "root\n\na\n\nb" {
		t.Fatalf("sections = %+v", got.Sections)
	}
	if len(got.Warnings) != 1 || !strings.Contains(got.Warnings[0], "cycle") {
		t.Fatalf("warnings = %v", got.Warnings)
	}
}
