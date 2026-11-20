package main

import (
	"context"
	"os"
	"slices"
	"testing"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

// TestLiveGraphMatchesSeed checks the round trip through Neo4j: after `load`,
// merging from the graph must give the same text as merging the seed in
// memory. It reads the shared database, so it runs only on request.
func TestLiveGraphMatchesSeed(t *testing.T) {
	if os.Getenv("NEO4J_MEMORY_LIVE") == "" {
		t.Skip("set NEO4J_MEMORY_LIVE=1 to read the live graph")
	}
	cred, err := readCredentials("../.secrets/neo4j.txt")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	driver, err := neo4j.NewDriverWithContext(cred.URI, neo4j.BasicAuth(cred.Username, cred.Password, ""))
	if err != nil {
		t.Fatal(err)
	}
	defer driver.Close(ctx)

	// The graph was loaded with the real values, so compare against the same.
	const liveValues = "../.secrets/neo4j-memory-values.yaml"
	seed := loadRepoSeed(t, liveValues)
	manifest, err := renderManifest(manifestTemplate, liveValues)
	if err != nil {
		t.Fatal(err)
	}
	user := manifest.Scopes[0].Name
	for _, scopes := range [][]string{
		{user, "멀티 세션", "peer CLI: Antigravity CLI"},
		{user, "멀티 세션", "peer CLI: Codex CLI"},
		{user, "파일 작성", "코드 작성", "브랜치와 PR 작업", "GitHub 리포지토리", "작업 파일 저장", "claude-workflow-kit 리포지토리"},
	} {
		candidates, err := fetchCandidates(ctx, driver, cred.Database, scopes)
		if err != nil {
			t.Fatal(err)
		}
		live, want := Merge(candidates, scopes), Merge(seed.Instructions, scopes)
		if len(live.Sections) != len(want.Sections) || len(live.Warnings) != len(want.Warnings) {
			t.Fatalf("%v: live %v / %v, seed %v / %v", scopes, sectionKeys(live), live.Warnings, sectionKeys(want), want.Warnings)
		}
		for i := range want.Sections {
			l, w := live.Sections[i], want.Sections[i]
			slices.Sort(l.Scopes)
			slices.Sort(w.Scopes)
			if l.Key != w.Key || l.Title != w.Title || l.Body != w.Body || !slices.Equal(l.Scopes, w.Scopes) {
				t.Errorf("%v: section %s differs from the seed", scopes, w.Key)
			}
		}
	}
}
