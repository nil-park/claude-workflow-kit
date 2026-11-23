package main

import "slices"

type Instruction struct {
	Key    string
	Title  string
	Body   string
	Scopes []string
}

// Merge collects every instruction whose scopes are all in the active set.
// Output is sorted by key only for reproducibility; merging does not depend on
// order. An instruction with no scopes never applies.
func Merge(candidates []Instruction, active []string) []Instruction {
	byKey := map[string]Instruction{}
	for _, inst := range candidates {
		if len(inst.Scopes) > 0 && isSubset(inst.Scopes, active) {
			byKey[inst.Key] = inst
		}
	}
	keys := make([]string, 0, len(byKey))
	for key := range byKey {
		keys = append(keys, key)
	}
	slices.Sort(keys)

	merged := make([]Instruction, 0, len(keys))
	for _, key := range keys {
		inst := byKey[key]
		scopes := slices.Clone(inst.Scopes)
		slices.Sort(scopes)
		inst.Scopes = scopes
		merged = append(merged, inst)
	}
	return merged
}

func isSubset(items, set []string) bool {
	for _, item := range items {
		if !slices.Contains(set, item) {
			return false
		}
	}
	return true
}
