package main

import (
	"context"
	"fmt"
	"slices"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

const (
	reminderExistsQuery = "MATCH (r:Reminder {key: $key}) RETURN r.key AS key"
	lockReminderQuery   = "MATCH (r:Reminder {key: $key}) SET r._lock = true REMOVE r._lock RETURN r.key AS key"
	maxShortIDRetries   = 5
)

type Reminder struct {
	Key         string
	Title       string
	Description string
	Body        string
	Scopes      []string
	Score       float64
}

func (s *store) createReminder(ctx context.Context, key, title, desc, body string, embedding []float64, scopes []string) (string, error) {
	if err := s.ensureConstraints(ctx); err != nil {
		return "", err
	}
	uniqueScopes := dedup(scopes)
	slices.Sort(uniqueScopes)
	autoKey := key == ""

	var createdKey string
	err := s.withWriteTx(ctx, func(tx neo4j.ManagedTransaction) error {
		chosenKey, err := resolveReminderKey(ctx, tx, key, autoKey)
		if err != nil {
			return err
		}
		for _, name := range uniqueScopes {
			if err := checkNodeExists(ctx, tx, lockScopeQuery, map[string]any{"name": name}, "scope", name); err != nil {
				return err
			}
		}
		params := map[string]any{
			"key":         chosenKey,
			"title":       title,
			"description": desc,
			"body":        body,
			"embedding":   embedding,
			"scopes":      uniqueScopes,
		}
		if err := runAndConsume(ctx, tx, `
			CREATE (r:Reminder {key: $key, title: $title, description: $description, body: $body, embedding: $embedding})
			WITH r
			UNWIND $scopes AS name
			MATCH (s:Scope {name: name})
			CREATE (r)-[:APPLIES_IN]->(s)`, params); err != nil {
			return err
		}
		createdKey = chosenKey
		return nil
	})
	return createdKey, err
}

func resolveReminderKey(ctx context.Context, tx neo4j.ManagedTransaction, key string, autoKey bool) (string, error) {
	if !autoKey {
		_, exists, err := querySingleRow(ctx, tx, "MATCH (r:Reminder {key: $key}) RETURN r.key AS key LIMIT 1", map[string]any{"key": key})
		if err != nil {
			return "", err
		}
		if exists {
			return "", fmt.Errorf("reminder %q already exists", key)
		}
		return key, nil
	}
	for range maxShortIDRetries {
		cand, err := generateShortID()
		if err != nil {
			return "", err
		}
		_, exists, err := querySingleRow(ctx, tx, "MATCH (r:Reminder {key: $key}) RETURN r.key AS key LIMIT 1", map[string]any{"key": cand})
		if err != nil {
			return "", err
		}
		if !exists {
			return cand, nil
		}
	}
	return "", fmt.Errorf("failed to allocate unique reminder key after %d attempts", maxShortIDRetries)
}

func (s *store) editReminder(ctx context.Context, key string, title, desc, body *string, embedding []float64) error {
	params := map[string]any{"key": key, "title": nil, "description": nil, "body": nil, "embedding": nil}
	if title != nil {
		params["title"] = *title
	}
	if desc != nil {
		params["description"] = *desc
		params["embedding"] = embedding
	}
	if body != nil {
		params["body"] = *body
	}
	return s.withWriteTx(ctx, func(tx neo4j.ManagedTransaction) error {
		if err := checkNodeExists(ctx, tx, reminderExistsQuery, params, "reminder", key); err != nil {
			return err
		}
		return runAndConsume(ctx, tx, `
			MATCH (r:Reminder {key: $key})
			SET r.title = coalesce($title, r.title),
			    r.description = coalesce($description, r.description),
			    r.body = coalesce($body, r.body),
			    r.embedding = coalesce($embedding, r.embedding)`, params)
	})
}

func (s *store) linkReminder(ctx context.Context, key, scope string) error {
	params := map[string]any{"key": key, "name": scope}
	return s.withWriteTx(ctx, func(tx neo4j.ManagedTransaction) error {
		if err := checkNodeExists(ctx, tx, lockScopeQuery, params, "scope", scope); err != nil {
			return err
		}
		if err := checkNodeExists(ctx, tx, reminderExistsQuery, params, "reminder", key); err != nil {
			return err
		}
		return runAndConsume(ctx, tx, `
			MATCH (r:Reminder {key: $key}), (s:Scope {name: $name})
			MERGE (r)-[:APPLIES_IN]->(s)`, params)
	})
}

func (s *store) unlinkReminder(ctx context.Context, key, scope string) error {
	params := map[string]any{"key": key, "scope": scope}
	return s.withWriteTx(ctx, func(tx neo4j.ManagedTransaction) error {
		if err := checkNodeExists(ctx, tx, lockReminderQuery, params, "reminder", key); err != nil {
			return err
		}
		if err := checkNodeExists(ctx, tx, scopeExistsQuery, params, "scope", scope); err != nil {
			return err
		}
		row, _, err := querySingleRow(ctx, tx, `
			MATCH (r:Reminder {key: $key})
			RETURN EXISTS { (r)-[:APPLIES_IN]->(:Scope {name: $scope}) } AS linked,
			       COUNT { (r)-[:APPLIES_IN]->(:Scope) } AS scopeCount`, params)
		if err != nil {
			return err
		}
		linked, _ := row["linked"].(bool)
		if !linked {
			return fmt.Errorf("reminder %q is not linked to scope %q", key, scope)
		}
		scopeCount, _ := row["scopeCount"].(int64)
		if scopeCount <= 1 {
			return fmt.Errorf("cannot unlink the last scope %q from reminder %q", scope, key)
		}
		return runAndConsume(ctx, tx, `
			MATCH (r:Reminder {key: $key})-[rel:APPLIES_IN]->(:Scope {name: $scope})
			DELETE rel`, params)
	})
}

func (s *store) deleteReminder(ctx context.Context, key string) error {
	params := map[string]any{"key": key}
	return s.withWriteTx(ctx, func(tx neo4j.ManagedTransaction) error {
		if err := checkNodeExists(ctx, tx, reminderExistsQuery, params, "reminder", key); err != nil {
			return err
		}
		return runAndConsume(ctx, tx, "MATCH (r:Reminder {key: $key}) DETACH DELETE r", params)
	})
}
