package main

import (
	"context"
	"fmt"
	"slices"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

const instructionExistsQuery = "MATCH (i:Instruction {key: $key}) RETURN i.key AS key"
const scopeExistsQuery = "MATCH (s:Scope {name: $scope}) RETURN s.name AS name"
const lockInstructionQuery = "MATCH (i:Instruction {key: $key}) SET i._lock = true REMOVE i._lock RETURN i.key AS key"

func (s *store) createInstruction(ctx context.Context, key, title, body string, scopes []string) error {
	if err := s.ensureConstraints(ctx); err != nil {
		return err
	}
	uniqueScopes := dedup(scopes)
	slices.Sort(uniqueScopes)
	params := map[string]any{"key": key, "title": title, "body": body, "scopes": uniqueScopes}
	return s.withWriteTx(ctx, func(tx neo4j.ManagedTransaction) error {
		_, exists, err := querySingleRow(ctx, tx, "MATCH (i:Instruction {key: $key}) RETURN i.key AS key LIMIT 1", params)
		if err != nil {
			return err
		}
		if exists {
			return fmt.Errorf("instruction %q already exists", key)
		}
		for _, name := range uniqueScopes {
			if err := checkNodeExists(ctx, tx, lockScopeQuery, map[string]any{"name": name}, "scope", name); err != nil {
				return err
			}
		}
		return runAndConsume(ctx, tx, `
			CREATE (i:Instruction {key: $key, title: $title, body: $body})
			WITH i
			UNWIND $scopes AS name
			MATCH (s:Scope {name: name})
			CREATE (i)-[:APPLIES_IN]->(s)`, params)
	})
}

func (s *store) editInstruction(ctx context.Context, key string, title, body *string) error {
	params := map[string]any{"key": key, "title": nil, "body": nil}
	if title != nil {
		params["title"] = *title
	}
	if body != nil {
		params["body"] = *body
	}
	return s.withWriteTx(ctx, func(tx neo4j.ManagedTransaction) error {
		if err := checkNodeExists(ctx, tx, instructionExistsQuery, params, "instruction", key); err != nil {
			return err
		}
		return runAndConsume(ctx, tx, `
			MATCH (i:Instruction {key: $key})
			SET i.title = coalesce($title, i.title),
			    i.body = coalesce($body, i.body)`, params)
	})
}

func (s *store) linkInstruction(ctx context.Context, key, scope string) error {
	params := map[string]any{"key": key, "name": scope}
	return s.withWriteTx(ctx, func(tx neo4j.ManagedTransaction) error {
		if err := checkNodeExists(ctx, tx, lockScopeQuery, params, "scope", scope); err != nil {
			return err
		}
		if err := checkNodeExists(ctx, tx, instructionExistsQuery, params, "instruction", key); err != nil {
			return err
		}
		return runAndConsume(ctx, tx, `
			MATCH (i:Instruction {key: $key}), (s:Scope {name: $name})
			MERGE (i)-[:APPLIES_IN]->(s)`, params)
	})
}

func (s *store) unlinkInstruction(ctx context.Context, key, scope string) error {
	params := map[string]any{"key": key, "scope": scope}
	return s.withWriteTx(ctx, func(tx neo4j.ManagedTransaction) error {
		if err := checkNodeExists(ctx, tx, lockInstructionQuery, params, "instruction", key); err != nil {
			return err
		}
		if err := checkNodeExists(ctx, tx, scopeExistsQuery, params, "scope", scope); err != nil {
			return err
		}
		row, _, err := querySingleRow(ctx, tx, `
			MATCH (i:Instruction {key: $key})
			RETURN EXISTS { (i)-[:APPLIES_IN]->(:Scope {name: $scope}) } AS linked,
			       COUNT { (i)-[:APPLIES_IN]->(:Scope) } AS scopeCount`, params)
		if err != nil {
			return err
		}
		linked, _ := row["linked"].(bool)
		if !linked {
			return fmt.Errorf("instruction %q is not linked to scope %q", key, scope)
		}
		scopeCount, _ := row["scopeCount"].(int64)
		if scopeCount <= 1 {
			return fmt.Errorf("cannot unlink the last scope %q from instruction %q", scope, key)
		}
		return runAndConsume(ctx, tx, `
			MATCH (i:Instruction {key: $key})-[r:APPLIES_IN]->(:Scope {name: $scope})
			DELETE r`, params)
	})
}

func (s *store) deleteInstruction(ctx context.Context, key string) error {
	params := map[string]any{"key": key}
	return s.withWriteTx(ctx, func(tx neo4j.ManagedTransaction) error {
		if err := checkNodeExists(ctx, tx, instructionExistsQuery, params, "instruction", key); err != nil {
			return err
		}
		return runAndConsume(ctx, tx, "MATCH (i:Instruction {key: $key}) DETACH DELETE i", params)
	})
}

func dedup(items []string) []string {
	out := make([]string, 0, len(items))
	for _, item := range items {
		if !slices.Contains(out, item) {
			out = append(out, item)
		}
	}
	return out
}
