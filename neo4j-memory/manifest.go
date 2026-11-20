package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"gopkg.in/yaml.v3"
)

// The seed manifest records the agent's judgment of which scopes the
// repository's skills belong to, and how variants split into nodes.
type seedManifest struct {
	// Owner keeps one user's seeded instructions apart from another's in the
	// shared database: it prefixes every key and tags every node it loads.
	Owner        string            `yaml:"owner"`
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
	if manifest.Owner == "" {
		return seedManifest{}, fmt.Errorf("%s: owner is missing", manifestPath)
	}
	return manifest, nil
}
