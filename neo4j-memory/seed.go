package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
)

type Seed struct {
	// Origin tags every node this seed loads, so a reload soft-deletes only
	// this owner's dropped instructions.
	Origin       string
	Scopes       []ScopeSpec
	Instructions []Instruction
}

// loadSeed renders the manifest and reads the files it names, resolving file
// paths against root. Keys come out prefixed with the owner.
func loadSeed(manifestPath, valuesPath, root string) (Seed, error) {
	manifest, err := renderManifest(manifestPath, valuesPath)
	if err != nil {
		return Seed{}, err
	}
	owner := normalizeText(manifest.Owner)
	for i := range manifest.Scopes {
		manifest.Scopes[i].Name = normalizeText(manifest.Scopes[i].Name)
		manifest.Scopes[i].Description = normalizeText(manifest.Scopes[i].Description)
	}

	var insts []Instruction
	parents := map[string]string{}
	for _, s := range manifest.Instructions {
		body := s.Body
		if s.File != "" {
			text, err := readText(filepath.Join(root, s.File))
			if err != nil {
				return Seed{}, err
			}
			body = stripFrontmatter(text)
		}
		insts = append(insts, Instruction{Key: s.Key, Body: normalizeText(body), Slot: s.Slot, Scopes: normalizeAll(s.Scopes)})
		parents[s.Key] = s.IncludedBy
	}
	for _, v := range manifest.Variants {
		text, err := readText(filepath.Join(root, v.File))
		if err != nil {
			return Seed{}, err
		}
		sections, err := parseVariant(text)
		if err != nil {
			return Seed{}, fmt.Errorf("%s: %w", v.File, err)
		}
		for _, sec := range sections {
			key := v.KeyPrefix + "/" + sec.Name
			insts = append(insts, Instruction{Key: key, Body: sec.Value, Slot: sec.Name, Scopes: normalizeAll(v.Scopes)})
			parents[key] = v.IncludedBy
		}
	}

	seed := Seed{Origin: seedOrigin + ":" + owner, Scopes: manifest.Scopes, Instructions: insts}
	if err := link(&seed, parents); err != nil {
		return Seed{}, err
	}
	prefixKeys(seed.Instructions, owner+"/")
	return seed, nil
}

// link records each includedBy as an edge on the parent and checks that every
// name the manifest uses points at something it defines.
func link(seed *Seed, parents map[string]string) error {
	index := map[string]int{}
	for i, inst := range seed.Instructions {
		if _, dup := index[inst.Key]; dup {
			return fmt.Errorf("duplicate instruction key %q", inst.Key)
		}
		index[inst.Key] = i
	}
	scopeNames := make([]string, 0, len(seed.Scopes))
	for _, s := range seed.Scopes {
		scopeNames = append(scopeNames, s.Name)
	}
	var errs []error
	for _, inst := range seed.Instructions {
		if len(inst.Scopes) == 0 {
			errs = append(errs, fmt.Errorf("%s: an instruction needs at least one scope", inst.Key))
		}
		for _, s := range inst.Scopes {
			if !slices.Contains(scopeNames, s) {
				errs = append(errs, fmt.Errorf("%s: unknown scope %q", inst.Key, s))
			}
		}
		parent := parents[inst.Key]
		if parent == "" {
			continue
		}
		p, ok := index[parent]
		if !ok {
			errs = append(errs, fmt.Errorf("%s: unknown includedBy %q", inst.Key, parent))
			continue
		}
		seed.Instructions[p].Includes = append(seed.Instructions[p].Includes, inst.Key)
	}
	return errors.Join(errs...)
}

func prefixKeys(insts []Instruction, prefix string) {
	for i := range insts {
		insts[i].Key = prefix + insts[i].Key
		for j := range insts[i].Includes {
			insts[i].Includes[j] = prefix + insts[i].Includes[j]
		}
	}
}

func readText(path string) (string, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return normalizeText(string(raw)), nil
}

func normalizeAll(names []string) []string {
	out := make([]string, len(names))
	for i, n := range names {
		out[i] = normalizeText(n)
	}
	return out
}
