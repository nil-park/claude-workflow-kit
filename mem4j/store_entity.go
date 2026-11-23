package main

import (
	"context"
	"fmt"
	"regexp"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

const (
	lockEntityQuery = "MATCH (e:Entity {name: $name}) SET e._lock = true REMOVE e._lock RETURN e.name AS name"
)

var relTypePattern = regexp.MustCompile(`^[A-Z][A-Z0-9_]*$`)

type Entity struct {
	Name        string
	Description string
	IsScope     bool
	Score       float64
}

type HubEdge struct {
	From string
	Type string
	To   string
}

type NeighborView struct {
	Root       string
	Edges      []HubEdge
	Statements []Statement
	Events     []Event
}

func validateRelType(relType string) error {
	if !relTypePattern.MatchString(relType) {
		return fmt.Errorf("invalid relationship type %q (expected UPPER_SNAKE_CASE)", relType)
	}
	if relType == "APPLIES_IN" || relType == "ABOUT" {
		return fmt.Errorf("relationship type %q is reserved for leaf links", relType)
	}
	return nil
}

func (s *store) createEntity(ctx context.Context, name, description string, embedding []float64) error {
	if err := s.ensureConstraints(ctx); err != nil {
		return err
	}
	params := map[string]any{"name": name, "description": description, "embedding": embedding}
	return s.withWriteTx(ctx, func(tx neo4j.ManagedTransaction) error {
		_, exists, err := querySingleRow(ctx, tx, "MATCH (e:Entity {name: $name}) RETURN e.name AS name LIMIT 1", params)
		if err != nil {
			return err
		}
		if exists {
			return fmt.Errorf("entity %q already exists", name)
		}
		row, found, err := querySingleRow(ctx, tx, `
			MATCH (sc:Scope {name: $name})
			WITH collect(sc) AS scopes
			WHERE size(scopes) = 1
			WITH scopes[0] AS sc
			WHERE NOT sc:Instruction AND NOT sc:Reminder
			SET sc:Entity, sc.description = $description, sc.embedding = $embedding
			RETURN true AS promoted`, params)
		if err != nil {
			return err
		}
		if found && row["promoted"] == true {
			return nil
		}
		return runAndConsume(ctx, tx,
			"CREATE (:Entity {name: $name, description: $description, embedding: $embedding})", params)
	})
}

func (s *store) editEntity(ctx context.Context, name, description string, embedding []float64) error {
	params := map[string]any{"name": name, "description": description, "embedding": embedding}
	return s.withWriteTx(ctx, func(tx neo4j.ManagedTransaction) error {
		if err := checkNodeExists(ctx, tx, "MATCH (e:Entity {name: $name}) RETURN e.name AS name", params, "entity", name); err != nil {
			return err
		}
		return runAndConsume(ctx, tx,
			"MATCH (e:Entity {name: $name}) SET e.description = $description, e.embedding = $embedding", params)
	})
}

func (s *store) deleteEntity(ctx context.Context, name string) error {
	params := map[string]any{"name": name}
	return s.withWriteTx(ctx, func(tx neo4j.ManagedTransaction) error {
		if err := checkNodeExists(ctx, tx, lockEntityQuery, params, "entity", name); err != nil {
			return err
		}
		row, _, err := querySingleRow(ctx, tx,
			"MATCH (e:Entity {name: $name})<-[:ABOUT]-(n) WHERE n:Statement OR n:Event RETURN count(n) AS linked",
			params)
		if err != nil {
			return err
		}
		if linked, _ := row["linked"].(int64); linked > 0 {
			return fmt.Errorf("entity %q still has %d linked statement(s) or event(s)", name, linked)
		}
		dual, _, err := querySingleRow(ctx, tx, "MATCH (e:Entity {name: $name}) RETURN e:Scope AS isScope", params)
		if err != nil {
			return err
		}
		if isScope, _ := dual["isScope"].(bool); isScope {
			return runAndConsume(ctx, tx, "MATCH (e:Entity {name: $name}) REMOVE e:Entity, e.embedding", params)
		}
		return runAndConsume(ctx, tx, "MATCH (e:Entity {name: $name}) DETACH DELETE e", params)
	})
}

func (s *store) linkHubs(ctx context.Context, label, from, to, relType string, link bool) error {
	if err := validateRelType(relType); err != nil {
		return err
	}
	if from == to {
		return fmt.Errorf("cannot link %s %q to itself", label, from)
	}
	existsQ := fmt.Sprintf("MATCH (n:%s {name: $name}) RETURN n.name AS name", label)
	return s.withWriteTx(ctx, func(tx neo4j.ManagedTransaction) error {
		if err := checkNodeExists(ctx, tx, existsQ, map[string]any{"name": from}, label, from); err != nil {
			return err
		}
		if err := checkNodeExists(ctx, tx, existsQ, map[string]any{"name": to}, label, to); err != nil {
			return err
		}
		params := map[string]any{"from": from, "to": to}
		if link {
			q := fmt.Sprintf("MATCH (a:%s {name: $from}), (b:%s {name: $to}) MERGE (a)-[:%s]->(b)", label, label, relType)
			return runAndConsume(ctx, tx, q, params)
		}
		checkQ := fmt.Sprintf("MATCH (a:%s {name: $from})-[r:%s]->(b:%s {name: $to}) RETURN count(r) AS cnt", label, relType, label)
		row, _, err := querySingleRow(ctx, tx, checkQ, params)
		if err != nil {
			return err
		}
		if cnt, _ := row["cnt"].(int64); cnt == 0 {
			return fmt.Errorf("%s link (%q)-[:%s]->(%q) not found", label, from, relType, to)
		}
		delQ := fmt.Sprintf("MATCH (a:%s {name: $from})-[r:%s]->(b:%s {name: $to}) DELETE r", label, relType, label)
		return runAndConsume(ctx, tx, delQ, params)
	})
}

func (s *store) listEntities(ctx context.Context) ([]Entity, error) {
	res, err := neo4j.ExecuteQuery(ctx, s.driver, `
		MATCH (e:Entity) WHERE e.name IS NOT NULL
		RETURN e.name AS name, coalesce(e.description, '') AS description, e:Scope AS isScope
		ORDER BY e.name LIMIT $limit`,
		map[string]any{"limit": candidateLimit},
		neo4j.EagerResultTransformer, neo4j.ExecuteQueryWithDatabase(s.db), neo4j.ExecuteQueryWithReadersRouting(),
		neo4j.ExecuteQueryWithTransactionConfig(neo4j.WithTxTimeout(txTimeout)))
	if err != nil {
		return nil, err
	}
	out := make([]Entity, 0, len(res.Records))
	for _, rec := range res.Records {
		e, err := toEntity(rec.AsMap(), false)
		if err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, nil
}

func (s *store) recallEntities(ctx context.Context, queryVec []float64, threshold float64, limit int) ([]Entity, error) {
	res, err := neo4j.ExecuteQuery(ctx, s.driver, `
		MATCH (e:Entity) WHERE e.name IS NOT NULL AND e.embedding IS NOT NULL
		WITH e LIMIT $candidateLimit
		WITH e, vector.similarity.cosine(e.embedding, $queryVec) AS score
		WHERE score >= $threshold
		RETURN e.name AS name, coalesce(e.description, '') AS description, e:Scope AS isScope, score
		ORDER BY score DESC, e.name ASC LIMIT $limit`,
		map[string]any{"candidateLimit": candidateLimit, "queryVec": queryVec, "threshold": threshold, "limit": limit},
		neo4j.EagerResultTransformer, neo4j.ExecuteQueryWithDatabase(s.db), neo4j.ExecuteQueryWithReadersRouting(),
		neo4j.ExecuteQueryWithTransactionConfig(neo4j.WithTxTimeout(txTimeout)))
	if err != nil {
		return nil, err
	}
	out := make([]Entity, 0, len(res.Records))
	for _, rec := range res.Records {
		e, err := toEntity(rec.AsMap(), true)
		if err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, nil
}
