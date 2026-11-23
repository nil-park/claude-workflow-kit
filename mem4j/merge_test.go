package main

import (
	"slices"
	"testing"
)

func TestMergeRequiresEveryScope(t *testing.T) {
	candidates := []Instruction{
		{Key: "both", Title: "Both", Body: "b", Scopes: []string{"user", "repo"}},
		{Key: "user-only", Title: "User", Body: "u", Scopes: []string{"user"}},
		{Key: "no-scope", Title: "None", Body: "n"},
	}
	got := Merge(candidates, []string{"user"})
	if len(got) != 1 || got[0].Key != "user-only" || got[0].Title != "User" || got[0].Body != "u" {
		t.Fatalf("got %+v, want only user-only", got)
	}
}

func TestMergeSortsByKeyRegardlessOfInputOrder(t *testing.T) {
	candidates := []Instruction{
		{Key: "z", Title: "Z", Body: "zb", Scopes: []string{"b", "a"}},
		{Key: "a", Title: "A", Body: "ab", Scopes: []string{"a"}},
	}
	got := Merge(candidates, []string{"b", "a"})
	if len(got) != 2 || got[0].Key != "a" || got[1].Key != "z" {
		t.Fatalf("got %+v, want keys [a z]", got)
	}
	if !slices.Equal(got[1].Scopes, []string{"a", "b"}) {
		t.Fatalf("scopes = %v, want [a b]", got[1].Scopes)
	}
}
