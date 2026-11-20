package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

// TestLiveWriteRulesAndMerge exercises the write commands and merge against
// the live Aura graph under a unique prefix, cleaning up all created nodes
// even when the test fails.
func TestLiveWriteRulesAndMerge(t *testing.T) {
	if os.Getenv("NEO4J_MEMORY_LIVE") == "" {
		t.Skip("set NEO4J_MEMORY_LIVE=1 to run against the live graph")
	}
	cred, err := readCredentials(defaultCredentials)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	driver, err := neo4j.NewDriverWithContext(cred.URI, neo4j.BasicAuth(cred.Username, cred.Password, ""))
	if err != nil {
		t.Fatal(err)
	}
	// t.Cleanup runs in LIFO order: register driver.Close first so the node
	// cleanup query runs while the driver is still open.
	t.Cleanup(func() {
		closeCtx, closeCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer closeCancel()
		if err := driver.Close(closeCtx); err != nil {
			t.Errorf("close driver: %v", err)
		}
	})

	s := &store{driver: driver, db: cred.Database}
	prefix := fmt.Sprintf("test-%d-", time.Now().UnixNano())
	t.Cleanup(func() {
		cleanCtx, cleanCancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cleanCancel()
		if _, err := neo4j.ExecuteQuery(cleanCtx, driver, `
			MATCH (n)
			WHERE (n:Scope AND n.name STARTS WITH $prefix)
			   OR (n:Entity AND n.name STARTS WITH $prefix)
			   OR (n:Instruction AND n.key STARTS WITH $prefix)
			DETACH DELETE n`,
			map[string]any{"prefix": prefix},
			neo4j.EagerResultTransformer, neo4j.ExecuteQueryWithDatabase(cred.Database)); err != nil {
			t.Errorf("cleanup test nodes: %v", err)
		}
	})

	scopeA := prefix + "scope-a"
	scopeB := prefix + "scope-b"
	entityScope := prefix + "entity-scope"
	instKey := prefix + "inst-1"

	// Entity -> Scope promotion
	if _, err := neo4j.ExecuteQuery(ctx, driver, "CREATE (:Entity {name: $name})",
		map[string]any{"name": entityScope}, neo4j.EagerResultTransformer, neo4j.ExecuteQueryWithDatabase(cred.Database)); err != nil {
		t.Fatal(err)
	}
	if err := s.createScope(ctx, entityScope, "promoted scope", false); err != nil {
		t.Fatalf("createScope promoted: %v", err)
	}
	if err := s.createScope(ctx, scopeA, "scope A", true); err != nil {
		t.Fatalf("createScope A: %v", err)
	}
	if err := s.createScope(ctx, scopeB, "scope B", false); err != nil {
		t.Fatalf("createScope B: %v", err)
	}
	if err := s.editScope(ctx, scopeA, "scope A updated"); err != nil {
		t.Fatalf("editScope: %v", err)
	}

	// Create, edit, link, unlink, last-scope refusal
	if err := s.createInstruction(ctx, instKey, "Title 1", "Body 1", []string{scopeA}); err != nil {
		t.Fatalf("createInstruction: %v", err)
	}
	newTitle, newBody := "Title 2", "Body 2"
	if err := s.editInstruction(ctx, instKey, &newTitle, &newBody); err != nil {
		t.Fatalf("editInstruction: %v", err)
	}
	if err := s.unlinkInstruction(ctx, instKey, scopeA); err == nil || !strings.Contains(err.Error(), "last scope") {
		t.Fatalf("expected last-scope refusal, got %v", err)
	}
	if err := s.linkInstruction(ctx, instKey, scopeB); err != nil {
		t.Fatalf("linkInstruction: %v", err)
	}

	// Merge requires all linked scopes
	cands, err := s.fetchCandidates(ctx, []string{scopeA})
	if err != nil {
		t.Fatal(err)
	}
	if got := Merge(cands, []string{scopeA}); len(got) != 0 {
		t.Fatalf("merge with only scopeA: got %+v, want empty", got)
	}
	cands, err = s.fetchCandidates(ctx, []string{scopeA, scopeB})
	if err != nil {
		t.Fatal(err)
	}
	got := Merge(cands, []string{scopeA, scopeB})
	if len(got) != 1 || got[0].Key != instKey || got[0].Title != "Title 2" || got[0].Body != "Body 2" {
		t.Fatalf("merge with both scopes: got %+v", got)
	}

	// Unlink works when more than one scope is linked
	if err := s.unlinkInstruction(ctx, instKey, scopeB); err != nil {
		t.Fatalf("unlinkInstruction: %v", err)
	}

	// Scope deletion refused while an instruction is still linked
	if err := s.deleteScope(ctx, scopeA); err == nil || !strings.Contains(err.Error(), "1 linked") {
		t.Fatalf("expected linked-instruction refusal on deleteScope, got %v", err)
	}

	// Hard delete and verify nodes are gone
	if err := s.deleteInstruction(ctx, instKey); err != nil {
		t.Fatalf("deleteInstruction: %v", err)
	}
	if err := s.editInstruction(ctx, instKey, &newTitle, nil); err == nil || !strings.Contains(err.Error(), "not found") {
		t.Fatalf("expected not-found after delete, got %v", err)
	}
	if err := s.deleteScope(ctx, scopeB); err != nil {
		t.Fatalf("deleteScope: %v", err)
	}
	if err := s.editScope(ctx, scopeB, "nope"); err == nil || !strings.Contains(err.Error(), "not found") {
		t.Fatalf("expected not-found scope after delete, got %v", err)
	}
}
