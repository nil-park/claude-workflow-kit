package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"unicode"
)

type hookConfig struct {
	Scopes     []string            `json:"scopes"`
	Workspaces map[string][]string `json:"workspaces"`
}

func defaultMem4jDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home directory: %w", err)
	}
	return filepath.Join(home, ".config", "mem4j"), nil
}

func defaultConfigPath() (string, error) {
	dir, err := defaultMem4jDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}

func defaultHookLogPath() (string, error) {
	dir, err := defaultMem4jDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "hook.log"), nil
}

func defaultStateDir() (string, error) {
	dir, err := defaultMem4jDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "state"), nil
}

func loadHookConfig(path string) (hookConfig, error) {
	if path == "" {
		var err error
		path, err = defaultConfigPath()
		if err != nil {
			return hookConfig{}, err
		}
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return hookConfig{}, fmt.Errorf("read hook config %s: %w", path, err)
	}
	var cfg hookConfig
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return hookConfig{}, fmt.Errorf("parse hook config %s: %w", path, err)
	}
	return cfg, nil
}

func resolveConfiguredScopes(cfg hookConfig, workDirs []string) []string {
	scopes := make([]string, 0, len(cfg.Scopes))
	for _, s := range cfg.Scopes {
		if n := normalizeText(s); n != "" {
			scopes = append(scopes, n)
		}
	}
	normDirs := make([]string, 0, len(workDirs))
	for _, d := range workDirs {
		if nd := normalizeDirPath(d); nd != "" {
			normDirs = append(normDirs, nd)
		}
	}
	for wsPath, wsScopes := range cfg.Workspaces {
		normWS := normalizeDirPath(wsPath)
		if normWS == "" {
			continue
		}
		if matchesWorkspace(normWS, normDirs) {
			for _, s := range wsScopes {
				if n := normalizeText(s); n != "" {
					scopes = append(scopes, n)
				}
			}
		}
	}
	out := dedup(scopes)
	slices.Sort(out)
	return out
}

func matchesWorkspace(normWS string, normDirs []string) bool {
	prefix := normWS + "/"
	for _, dir := range normDirs {
		if dir == normWS || strings.HasPrefix(dir, prefix) {
			return true
		}
	}
	return false
}

func normalizeDirPath(p string) string {
	trimmed := strings.TrimSpace(p)
	if trimmed == "" {
		return ""
	}
	cleaned := path.Clean(strings.ReplaceAll(trimmed, "\\", "/"))
	if cleaned != "/" {
		cleaned = strings.TrimRight(cleaned, "/")
	}
	if runtime.GOOS == "windows" {
		cleaned = strings.ToLower(cleaned)
	}
	return normalizeText(cleaned)
}

func shouldRecallPrompt(prompt string) bool {
	trimmed := strings.TrimSpace(normalizeText(prompt))
	if trimmed == "" {
		return false
	}
	if containsHangul(trimmed) {
		stripped := strings.TrimRightFunc(trimmed, func(r rune) bool {
			return unicode.IsSpace(r) || unicode.IsPunct(r) || unicode.IsSymbol(r)
		})
		count := 0
		for _, r := range stripped {
			if !unicode.IsSpace(r) {
				count++
			}
		}
		return count >= 4
	}
	return len(strings.Fields(trimmed)) >= 2
}

func containsHangul(s string) bool {
	for _, r := range s {
		if unicode.In(r, unicode.Hangul) {
			return true
		}
	}
	return false
}
