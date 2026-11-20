package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func TestConvertValueHandlesGraphAndTemporalTypes(t *testing.T) {
	node := neo4j.Node{Labels: []string{"Scope"}, Props: map[string]any{"name": "파일 작성"}}
	rel := neo4j.Relationship{Type: "APPLIES_IN", Props: map[string]any{"weight": int64(1)}}
	path := neo4j.Path{Nodes: []neo4j.Node{node}, Relationships: []neo4j.Relationship{rel}}
	ts := time.Date(2026, 4, 21, 1, 2, 3, 0, time.UTC)
	date := neo4j.DateOf(ts)
	dur := neo4j.DurationOf(1, 2, 3, 4)

	row := convertMap(map[string]any{
		"n":    node,
		"r":    rel,
		"p":    path,
		"list": []any{node, ts},
		"map":  map[string]any{"d": date, "dur": dur},
	})

	var buf bytes.Buffer
	if err := writeJSONLines(&buf, []map[string]any{row}); err != nil {
		t.Fatal(err)
	}
	raw := buf.String()
	if strings.Contains(raw, `\u`) {
		t.Fatalf("expected UTF-8 Korean without \\u escaping, got %s", raw)
	}
	if !strings.Contains(raw, "파일 작성") {
		t.Fatalf("expected Korean property value in JSON, got %s", raw)
	}
	var decoded map[string]any
	if err := json.Unmarshal([]byte(strings.TrimSpace(raw)), &decoded); err != nil {
		t.Fatalf("invalid JSON line: %v", err)
	}
	nMap := decoded["n"].(map[string]any)
	if nMap["labels"].([]any)[0] != "Scope" {
		t.Fatalf("node labels = %v", nMap["labels"])
	}
	rMap := decoded["r"].(map[string]any)
	if rMap["type"] != "APPLIES_IN" {
		t.Fatalf("rel type = %v", rMap["type"])
	}
	pMap := decoded["p"].(map[string]any)
	if len(pMap["nodes"].([]any)) != 1 || len(pMap["relationships"].([]any)) != 1 {
		t.Fatalf("path = %v", pMap)
	}
	list := decoded["list"].([]any)
	if list[1] != "2026-04-21T01:02:03Z" {
		t.Fatalf("time = %v", list[1])
	}
	mMap := decoded["map"].(map[string]any)
	if mMap["d"] != "2026-04-21" || !strings.HasPrefix(mMap["dur"].(string), "P") {
		t.Fatalf("temporal map = %v", mMap)
	}
}
