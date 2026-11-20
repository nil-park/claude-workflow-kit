package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"text/template"

	"gopkg.in/yaml.v3"
)

// The seed manifest records the agent's judgment of which scopes the
// repository's skills belong to, and how variants split into nodes.
type seedManifest struct {
	Scopes       []ScopeSpec       `yaml:"scopes"`
	Instructions []seedInstruction `yaml:"instructions"`
	Variants     []seedVariant     `yaml:"variants"`
}

type ScopeSpec struct {
	Name                   string `yaml:"name"`
	Description            string `yaml:"description"`
	DisableModelInvocation bool   `yaml:"disableModelInvocation"`
}

// A seed instruction takes its body from File, or from Body when it is a short
// value such as a slot filled per repository.
type seedInstruction struct {
	Key        string   `yaml:"key"`
	File       string   `yaml:"file"`
	Body       string   `yaml:"body"`
	Slot       string   `yaml:"slot"`
	IncludedBy string   `yaml:"includedBy"`
	Scopes     []string `yaml:"scopes"`
}

// A seed variant becomes one slot-filling instruction per section of File,
// keyed "<KeyPrefix>/<section name>".
type seedVariant struct {
	KeyPrefix  string   `yaml:"keyPrefix"`
	File       string   `yaml:"file"`
	IncludedBy string   `yaml:"includedBy"`
	Scopes     []string `yaml:"scopes"`
}

type Seed struct {
	Scopes       []ScopeSpec
	Instructions []Instruction
}

// renderManifest fills the manifest template with the values file, the way
// Helm renders a chart. A key missing from the values file is an error.
func renderManifest(manifestPath, valuesPath string) (seedManifest, error) {
	rawValues, err := os.ReadFile(valuesPath)
	if err != nil {
		return seedManifest{}, err
	}
	values := map[string]any{}
	if err := yaml.Unmarshal(rawValues, &values); err != nil {
		return seedManifest{}, fmt.Errorf("%s: %w", valuesPath, err)
	}
	// JSON string syntax is valid YAML double-quoted scalar syntax.
	quote := func(v any) (string, error) {
		b, err := json.Marshal(fmt.Sprint(v))
		return string(b), err
	}
	tmpl, err := template.New(filepath.Base(manifestPath)).
		Option("missingkey=error").
		Funcs(template.FuncMap{"quote": quote}).
		ParseFiles(manifestPath)
	if err != nil {
		return seedManifest{}, err
	}
	var rendered bytes.Buffer
	if err := tmpl.Execute(&rendered, values); err != nil {
		return seedManifest{}, err
	}
	var manifest seedManifest
	if err := yaml.Unmarshal(rendered.Bytes(), &manifest); err != nil {
		return seedManifest{}, fmt.Errorf("%s rendered: %w", manifestPath, err)
	}
	return manifest, nil
}

// loadSeed renders the manifest and reads the files it names, resolving file
// paths against root.
func loadSeed(manifestPath, valuesPath, root string) (Seed, error) {
	manifest, err := renderManifest(manifestPath, valuesPath)
	if err != nil {
		return Seed{}, err
	}
	for i := range manifest.Scopes {
		manifest.Scopes[i].Name = normalizeText(manifest.Scopes[i].Name)
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

	seed := Seed{Scopes: manifest.Scopes, Instructions: insts}
	return seed, link(&seed, parents)
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
	for i, inst := range seed.Instructions {
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
		seed.Instructions[i].Included = true
	}
	return errors.Join(errs...)
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
