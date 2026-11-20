package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const (
	manifestTemplate = "seed/plugins.yaml.tmpl"
	exampleValues    = "seed/values.example.yaml"
	exampleUser      = "user"
)

func loadRepoSeed(t *testing.T, valuesPath string) Seed {
	t.Helper()
	seed, err := loadSeed(manifestTemplate, valuesPath, "..")
	if err != nil {
		t.Fatal(err)
	}
	return seed
}

// Every multi-session variant must fill every placeholder of the template;
// a missing or misnamed section would leave a {{...}} behind.
func TestRepoSeedFillsEveryMultiSessionVariant(t *testing.T) {
	seed := loadRepoSeed(t, exampleValues)
	for _, peer := range []string{"Antigravity CLI", "Claude Code", "Codex CLI", "GitHub Copilot CLI", "Gemini CLI"} {
		t.Run(peer, func(t *testing.T) {
			got := Merge(seed.Instructions, []string{exampleUser, "멀티 세션", "peer CLI: " + peer})
			if len(got.Warnings) > 0 {
				t.Fatalf("warnings: %v", got.Warnings)
			}
			if len(got.Sections) != 1 || got.Sections[0].Key != "multi-session" {
				t.Fatalf("sections = %v, want only multi-session", sectionKeys(got))
			}
			if strings.Contains(got.Sections[0].Body, "## peer-setup") {
				t.Fatal("a variant heading leaked into the merged body")
			}
		})
	}
}

func TestRepoSeedLeavesTemplateUnfilledWithoutPeerScope(t *testing.T) {
	got := Merge(loadRepoSeed(t, exampleValues).Instructions, []string{exampleUser, "멀티 세션"})
	if len(got.Warnings) == 0 {
		t.Fatal("expected unfilled placeholder warnings without a peer CLI scope")
	}
}

func TestRepoSeedFillsScratchDirOnlyInThisRepository(t *testing.T) {
	seed := loadRepoSeed(t, exampleValues)
	inRepo := Merge(seed.Instructions, []string{exampleUser, "작업 파일 저장", "claude-workflow-kit 리포지토리"})
	if len(inRepo.Warnings) > 0 || !strings.Contains(inRepo.Sections[0].Body, "`<repo-root>/.refs/`") {
		t.Fatalf("in repo: warnings %v", inRepo.Warnings)
	}
	elsewhere := Merge(seed.Instructions, []string{exampleUser, "작업 파일 저장"})
	if len(elsewhere.Warnings) == 0 {
		t.Fatal("expected {{scratch_dir}} to stay unfilled outside the repository")
	}
}

func TestRepoSeedPicksGitWorkflowByHosting(t *testing.T) {
	got := Merge(loadRepoSeed(t, exampleValues).Instructions, []string{exampleUser, "브랜치와 PR 작업", "GitHub 리포지토리"})
	if keys := sectionKeys(got); len(keys) != 1 || keys[0] != "git-workflow@github" {
		t.Fatalf("sections = %v, want only git-workflow@github", keys)
	}
}

func TestRenderManifestQuotesValuesAndRejectsMissingKeys(t *testing.T) {
	dir := t.TempDir()
	manifest := writeFile(t, dir, "m.tmpl", "scopes:\n  - name: {{ .user | quote }}\n    description: d\n")
	tricky := writeFile(t, dir, "tricky.yaml", "user: 'a \"b\": c'\n")
	got, err := renderManifest(manifest, tricky)
	if err != nil || len(got.Scopes) != 1 || got.Scopes[0].Name != `a "b": c` {
		t.Fatalf("got %+v, err %v", got, err)
	}
	empty := writeFile(t, dir, "empty.yaml", "other: x\n")
	if _, err := renderManifest(manifest, empty); err == nil {
		t.Fatal("expected an error for a value missing from the values file")
	}
}

func writeFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func sectionKeys(r MergeResult) []string {
	var keys []string
	for _, s := range r.Sections {
		keys = append(keys, s.Key)
	}
	return keys
}
