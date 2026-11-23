package main

import (
	"context"
	"fmt"
	"slices"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

type Statement struct {
	Key      string
	Title    string
	Body     string
	Entities []string
	Score    float64
}

type Event struct {
	Key        string
	Title      string
	OccurredAt string
	Body       string
	Entities   []string
	Score      float64
}

func (s *store) createFact(ctx context.Context, label, key, title, desc, occurredAt, body string, embedding []float64, entities []string) (string, error) {
	if err := s.ensureConstraints(ctx); err != nil {
		return "", err
	}
	uniqueEntities := dedup(entities)
	slices.Sort(uniqueEntities)
	autoKey := key == ""

	var createdKey string
	err := s.withWriteTx(ctx, func(tx neo4j.ManagedTransaction) error {
		chosenKey, err := resolveFactKey(ctx, tx, label, key, autoKey)
		if err != nil {
			return err
		}
		for _, name := range uniqueEntities {
			if err := checkNodeExists(ctx, tx, lockEntityQuery, map[string]any{"name": name}, "entity", name); err != nil {
				return err
			}
		}
		params := map[string]any{
			"key":         chosenKey,
			"title":       title,
			"description": desc,
			"body":        body,
			"embedding":   embedding,
			"entities":    uniqueEntities,
			"occurredAt":  occurredAt,
		}
		var q string
		if label == "Event" {
			q = `CREATE (n:Event {key: $key, title: $title, description: $description, occurredAt: $occurredAt, body: $body, embedding: $embedding})
				WITH n UNWIND $entities AS name MATCH (e:Entity {name: name}) CREATE (n)-[:ABOUT]->(e)`
		} else {
			q = `CREATE (n:Statement {key: $key, title: $title, description: $description, body: $body, embedding: $embedding})
				WITH n UNWIND $entities AS name MATCH (e:Entity {name: name}) CREATE (n)-[:ABOUT]->(e)`
		}
		if err := runAndConsume(ctx, tx, q, params); err != nil {
			return err
		}
		createdKey = chosenKey
		return nil
	})
	return createdKey, err
}

func resolveFactKey(ctx context.Context, tx neo4j.ManagedTransaction, label, key string, autoKey bool) (string, error) {
	existsQ := fmt.Sprintf("MATCH (n:%s {key: $key}) RETURN n.key AS key LIMIT 1", label)
	if !autoKey {
		_, exists, err := querySingleRow(ctx, tx, existsQ, map[string]any{"key": key})
		if err != nil {
			return "", err
		}
		if exists {
			return "", fmt.Errorf("%s %q already exists", label, key)
		}
		return key, nil
	}
	for range maxShortIDRetries {
		cand, err := generateShortID()
		if err != nil {
			return "", err
		}
		_, exists, err := querySingleRow(ctx, tx, existsQ, map[string]any{"key": cand})
		if err != nil {
			return "", err
		}
		if !exists {
			return cand, nil
		}
	}
	return "", fmt.Errorf("failed to allocate unique %s key after %d attempts", label, maxShortIDRetries)
}

func (s *store) editFact(ctx context.Context, label, key string, title, desc, occurredAt, body *string, embedding []float64) error {
	params := map[string]any{"key": key, "title": nil, "description": nil, "occurredAt": nil, "body": nil, "embedding": nil}
	if title != nil {
		params["title"] = *title
	}
	if desc != nil {
		params["description"] = *desc
		params["embedding"] = embedding
	}
	if occurredAt != nil {
		params["occurredAt"] = *occurredAt
	}
	if body != nil {
		params["body"] = *body
	}
	existsQ := fmt.Sprintf("MATCH (n:%s {key: $key}) RETURN n.key AS key", label)
	return s.withWriteTx(ctx, func(tx neo4j.ManagedTransaction) error {
		if err := checkNodeExists(ctx, tx, existsQ, params, label, key); err != nil {
			return err
		}
		if label == "Event" {
			return runAndConsume(ctx, tx, `
				MATCH (n:Event {key: $key})
				SET n.title = coalesce($title, n.title),
				    n.description = coalesce($description, n.description),
				    n.occurredAt = coalesce($occurredAt, n.occurredAt),
				    n.body = coalesce($body, n.body),
				    n.embedding = coalesce($embedding, n.embedding)`, params)
		}
		return runAndConsume(ctx, tx, `
			MATCH (n:Statement {key: $key})
			SET n.title = coalesce($title, n.title),
			    n.description = coalesce($description, n.description),
			    n.body = coalesce($body, n.body),
			    n.embedding = coalesce($embedding, n.embedding)`, params)
	})
}

func (s *store) linkFact(ctx context.Context, label, key, entity string, link bool) error {
	existsQ := fmt.Sprintf("MATCH (n:%s {key: $key}) RETURN n.key AS key", label)
	lockQ := fmt.Sprintf("MATCH (n:%s {key: $key}) SET n._lock = true REMOVE n._lock RETURN n.key AS key", label)
	return s.withWriteTx(ctx, func(tx neo4j.ManagedTransaction) error {
		if link {
			if err := checkNodeExists(ctx, tx, lockEntityQuery, map[string]any{"name": entity}, "entity", entity); err != nil {
				return err
			}
			if err := checkNodeExists(ctx, tx, existsQ, map[string]any{"key": key}, label, key); err != nil {
				return err
			}
			q := fmt.Sprintf("MATCH (n:%s {key: $key}), (e:Entity {name: $name}) MERGE (n)-[:ABOUT]->(e)", label)
			return runAndConsume(ctx, tx, q, map[string]any{"key": key, "name": entity})
		}
		if err := checkNodeExists(ctx, tx, lockQ, map[string]any{"key": key}, label, key); err != nil {
			return err
		}
		if err := checkNodeExists(ctx, tx, "MATCH (e:Entity {name: $name}) RETURN e.name AS name", map[string]any{"name": entity}, "entity", entity); err != nil {
			return err
		}
		checkQ := fmt.Sprintf(`
			MATCH (n:%s {key: $key})
			RETURN EXISTS { (n)-[:ABOUT]->(:Entity {name: $name}) } AS linked,
			       COUNT { (n)-[:ABOUT]->(:Entity) } AS cnt`, label)
		row, _, err := querySingleRow(ctx, tx, checkQ, map[string]any{"key": key, "name": entity})
		if err != nil {
			return err
		}
		if linked, _ := row["linked"].(bool); !linked {
			return fmt.Errorf("%s %q is not linked to entity %q", label, key, entity)
		}
		if cnt, _ := row["cnt"].(int64); cnt <= 1 {
			return fmt.Errorf("cannot unlink the last entity %q from %s %q", entity, label, key)
		}
		delQ := fmt.Sprintf("MATCH (n:%s {key: $key})-[r:ABOUT]->(:Entity {name: $name}) DELETE r", label)
		return runAndConsume(ctx, tx, delQ, map[string]any{"key": key, "name": entity})
	})
}

func (s *store) deleteFact(ctx context.Context, label, key string) error {
	existsQ := fmt.Sprintf("MATCH (n:%s {key: $key}) RETURN n.key AS key", label)
	delQ := fmt.Sprintf("MATCH (n:%s {key: $key}) DETACH DELETE n", label)
	return s.withWriteTx(ctx, func(tx neo4j.ManagedTransaction) error {
		if err := checkNodeExists(ctx, tx, existsQ, map[string]any{"key": key}, label, key); err != nil {
			return err
		}
		return runAndConsume(ctx, tx, delQ, map[string]any{"key": key})
	})
}
