package main

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestLiveBackDataCRUDLinksAndRecall(t *testing.T) {
	if os.Getenv("MEM4J_LIVE") != "1" {
		t.Skip("set MEM4J_LIVE=1 to run against the live graph")
	}
	ctx, s := openLiveStore(t)
	prefix := "_livetest-bd-"
	e1, e2 := prefix+"tool", prefix+"db"
	stmtKey, evKey := prefix+"s1", prefix+"e1"

	defer func() {
		_ = s.deleteFact(ctx, "Statement", stmtKey)
		_ = s.deleteFact(ctx, "Event", evKey)
		_ = s.deleteEntity(ctx, e1)
		_ = s.deleteEntity(ctx, e2)
	}()

	runCLI := func(args ...string) string {
		t.Helper()
		var out, errBuf bytes.Buffer
		if err := run(args, strings.NewReader(""), &out, &errBuf); err != nil {
			t.Fatalf("run(%v) failed: %v (stderr=%q)", args, err, errBuf.String())
		}
		return out.String()
	}

	runCLI("entity", "create", "-name", e1, "-description", "Go CLI tool for shared agent memory")
	runCLI("entity", "create", "-name", e2, "-description", "Neo4j AuraDB cloud graph database")
	runCLI("entity", "link", "-from", e1, "-to", e2, "-type", "USES")

	runCLI("statement", "create", "-key", stmtKey, "-title", "CLI query mode", "-entity", e1, "-body", "The CLI executes read-only Cypher queries via graph query.")
	runCLI("event", "create", "-key", evKey, "-title", "Initial launch", "-occurred-at", "2026-11-23", "-entity", e1, "-body", "Launched back-data support in mem4j.")

	nbOut := runCLI("entity", "neighbors", "-name", e1, "-depth", "1")
	if !strings.Contains(nbOut, `type="USES"`) || !strings.Contains(nbOut, `title="CLI query mode"`) || !strings.Contains(nbOut, `title="Initial launch"`) {
		t.Fatalf("unexpected neighbors output:\n%s", nbOut)
	}

	eRecall := runCLI("entity", "recall", "-query", "shared agent memory CLI", "-threshold", "0.5", "-limit", "5")
	if !strings.Contains(eRecall, e1) {
		t.Fatalf("entity recall missed %q:\n%s", e1, eRecall)
	}
	sRecall := runCLI("statement", "recall", "-entity", e1, "-query", "read-only Cypher queries", "-threshold", "0.5", "-limit", "5")
	if !strings.Contains(sRecall, stmtKey) {
		t.Fatalf("statement recall missed %q:\n%s", stmtKey, sRecall)
	}
	evRecall := runCLI("event", "recall", "-entity", e1, "-query", "back-data support launch", "-threshold", "0.5", "-limit", "5")
	if !strings.Contains(evRecall, evKey) {
		t.Fatalf("event recall missed %q:\n%s", evKey, evRecall)
	}

	runCLI("statement", "edit", "-key", stmtKey, "-title", "CLI read-only mode")
	listOut := runCLI("statement", "list", "-entity", e1)
	if !strings.Contains(listOut, `title="CLI read-only mode"`) {
		t.Fatalf("statement list did not reflect updated title:\n%s", listOut)
	}
}
