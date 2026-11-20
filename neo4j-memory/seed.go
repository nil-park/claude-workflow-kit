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
		text, err := readText(filepath.Join(root, s.File))
		if err != nil {
			return Seed{}, err
		}
		insts = append(insts, Instruction{Key: s.Key, Title: normalizeText(s.Title), Body: stripFrontmatter(text), Scopes: normalizeAll(s.Scopes)})
		parents[s.Key] = s.IncludedBy
	}
	for _, v := range manifest.Variants {
		slots, err := variantSlots(v, root)
		if err != nil {
			return Seed{}, err
		}
		insts = append(insts, Instruction{Key: v.Key, Title: normalizeText(v.Title), Slots: slots, Scopes: normalizeAll(v.Scopes)})
		parents[v.Key] = v.IncludedBy
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
		if inst.Title == "" {
			errs = append(errs, fmt.Errorf("%s: title is missing", inst.Key))
		}
		if inst.isVariant() && parents[inst.Key] == "" {
			errs = append(errs, fmt.Errorf("%s: a variant needs includedBy", inst.Key))
		}
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

// variantSlots takes the values from the variant file's sections, or from the
// manifest when it has no file. A variant without values would be an ordinary
// node with an empty body, so that is an error.
func variantSlots(v seedVariant, root string) (map[string]string, error) {
	slots := map[string]string{}
	for name, value := range v.Slots {
		slots[name] = normalizeText(value)
	}
	if v.File != "" {
		text, err := readText(filepath.Join(root, v.File))
		if err != nil {
			return nil, err
		}
		sections, err := parseVariant(text)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", v.File, err)
		}
		for _, sec := range sections {
			slots[sec.Name] = sec.Value
		}
	}
	if len(slots) == 0 {
		return nil, fmt.Errorf("%s: a variant needs at least one slot value", v.Key)
	}
	return slots, nil
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
