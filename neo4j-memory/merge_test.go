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

func TestMergeFillsSlotsFromVariantAndAppendsOthers(t *testing.T) {
	candidates := []Instruction{
		{Key: "tpl", Title: "Template", Scopes: []string{"s"}, Includes: []string{"tpl@v", "tail"},
			Body: "intro {{name}}\n- item\n  {{block}}\n{{empty}}\nend"},
		{Key: "tpl@v", Scopes: []string{"s"},
			Slots: map[string]string{"name": "agy", "block": "    - indented\n    - kept", "empty": ""}},
		{Key: "tail", Body: "appended", Scopes: []string{"s"}},
	}
	got := Merge(candidates, []string{"s"})
	want := "intro agy\n- item\n    - indented\n    - kept\nend\n\nappended"
	if len(got.Sections) != 1 || got.Sections[0].Body != want || got.Sections[0].Title != "Template" || len(got.Warnings) > 0 {
		t.Fatalf("got %+v\nwant body %q", got, want)
	}
}

func TestMergeSkipsInactiveVariantAndWarnsOnUnfilledSlot(t *testing.T) {
	candidates := []Instruction{
		{Key: "tpl", Body: "use {{peer}}", Scopes: []string{"s"}, Includes: []string{"tpl@agy"}},
		{Key: "tpl@agy", Slots: map[string]string{"peer": "agy"}, Scopes: []string{"s", "agy"}},
	}
	got := Merge(candidates, []string{"s"})
	if got.Sections[0].Body != "use {{peer}}" || len(got.Warnings) != 1 || !strings.Contains(got.Warnings[0], "{{peer}}") {
		t.Fatalf("got %+v", got)
	}
}

// When the including node is inactive, an ordinary instruction it includes
// still applies on its own, but a variant does not.
func TestMergeEmitsChildOfInactiveParentUnlessVariant(t *testing.T) {
	candidates := []Instruction{
		{Key: "parent", Body: "p {{x}}", Scopes: []string{"user", "repo"}, Includes: []string{"shared", "parent@v"}},
		{Key: "shared", Body: "shared", Scopes: []string{"user"}},
		{Key: "parent@v", Slots: map[string]string{"x": "x"}, Scopes: []string{"user"}},
	}
	got := sectionKeys(Merge(candidates, []string{"user"}))
	if !slices.Equal(got, []string{"shared"}) {
		t.Fatalf("sections = %v, want [shared]", got)
	}
}

func TestMergeWarnsWhenTwoVariantsFillOneSlot(t *testing.T) {
	candidates := []Instruction{
		{Key: "tpl", Body: "{{peer}} {{only-b}}", Scopes: []string{"s"}, Includes: []string{"tpl@a", "tpl@b"}},
		{Key: "tpl@a", Slots: map[string]string{"peer": "a"}, Scopes: []string{"s"}},
		{Key: "tpl@b", Slots: map[string]string{"peer": "b", "only-b": "kept"}, Scopes: []string{"s"}},
	}
	got := Merge(candidates, []string{"s"})
	if got.Sections[0].Body != "a kept" || len(got.Warnings) != 1 || !strings.Contains(got.Warnings[0], "tpl@b") {
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
