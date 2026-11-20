package main

import (
	"context"
	"fmt"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

const txTimeout = 20 * time.Second

var constraintQueries = []string{
	"CREATE CONSTRAINT instruction_key IF NOT EXISTS FOR (i:Instruction) REQUIRE i.key IS UNIQUE",
	"CREATE CONSTRAINT scope_name IF NOT EXISTS FOR (s:Scope) REQUIRE s.name IS UNIQUE",
}

type store struct {
	driver neo4j.DriverWithContext
	db     string
}

func (s *store) ensureConstraints(ctx context.Context) error {
	for _, q := range constraintQueries {
		if _, err := neo4j.ExecuteQuery(ctx, s.driver, q, nil, neo4j.EagerResultTransformer,
			neo4j.ExecuteQueryWithDatabase(s.db),
			neo4j.ExecuteQueryWithTransactionConfig(neo4j.WithTxTimeout(txTimeout))); err != nil {
			return err
		}
	}
	return nil
}

func (s *store) withWriteTx(ctx context.Context, fn func(neo4j.ManagedTransaction) error) error {
	session := s.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: s.db})
	defer session.Close(ctx)
	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		return nil, fn(tx)
	}, neo4j.WithTxTimeout(txTimeout))
	return err
}

func runAndConsume(ctx context.Context, tx neo4j.ManagedTransaction, query string, params map[string]any) error {
	res, err := tx.Run(ctx, query, params)
	if err != nil {
		return err
	}
	_, err = res.Consume(ctx)
	return err
}

func querySingleRow(ctx context.Context, tx neo4j.ManagedTransaction, query string, params map[string]any) (map[string]any, bool, error) {
	res, err := tx.Run(ctx, query, params)
	if err != nil {
		return nil, false, err
	}
	records, err := res.Collect(ctx)
	if err != nil {
		return nil, false, err
	}
	if len(records) == 0 {
		return nil, false, nil
	}
	return records[0].AsMap(), true, nil
}

func checkNodeExists(ctx context.Context, tx neo4j.ManagedTransaction, query string, params map[string]any, kind, id string) error {
	_, found, err := querySingleRow(ctx, tx, query, params)
	if err != nil {
		return err
	}
	if !found {
		return fmt.Errorf("%s %q not found", kind, id)
	}
	return nil
}

// createScope creates a Scope node, or adds :Scope to a single matching
// Entity that carries no Instruction or Reminder label.
func (s *store) createScope(ctx context.Context, name, description string, disableModelInvocation bool) error {
	if err := s.ensureConstraints(ctx); err != nil {
		return err
	}
	params := map[string]any{
		"name":                   name,
		"description":            description,
		"disableModelInvocation": disableModelInvocation,
	}
	return s.withWriteTx(ctx, func(tx neo4j.ManagedTransaction) error {
		_, exists, err := querySingleRow(ctx, tx, "MATCH (s:Scope {name: $name}) RETURN s.name AS name LIMIT 1", params)
		if err != nil {
			return err
		}
		if exists {
			return fmt.Errorf("scope %q already exists", name)
		}
		row, found, err := querySingleRow(ctx, tx, `
			MATCH (e:Entity {name: $name})
			WITH collect(e) AS entities
			WHERE size(entities) = 1
			WITH entities[0] AS e
			WHERE NOT e:Instruction AND NOT e:Reminder
			SET e:Scope, e.description = $description,
			    e.disableModelInvocation = CASE WHEN $disableModelInvocation THEN true END
			RETURN true AS promoted`, params)
		if err != nil {
			return err
		}
		if found && row["promoted"] == true {
			return nil
		}
		return runAndConsume(ctx, tx, `
			CREATE (n:Scope {name: $name, description: $description})
			SET n.disableModelInvocation = CASE WHEN $disableModelInvocation THEN true END`, params)
	})
}

func (s *store) editScope(ctx context.Context, name, description string) error {
	params := map[string]any{"name": name, "description": description}
	return s.withWriteTx(ctx, func(tx neo4j.ManagedTransaction) error {
		if err := checkNodeExists(ctx, tx,
			"MATCH (s:Scope {name: $name}) RETURN s.name AS name",
			params, "scope", name); err != nil {
			return err
		}
		return runAndConsume(ctx, tx, "MATCH (s:Scope {name: $name}) SET s.description = $description", params)
	})
}

// SET+REMOVE on a scratch property takes an exclusive write lock on the node
// for the transaction (Neo4j Java Reference, Transaction management -> Isolation levels).
const lockScopeQuery = "MATCH (s:Scope {name: $name}) SET s._lock = true REMOVE s._lock RETURN s.name AS name"

func (s *store) deleteScope(ctx context.Context, name string) error {
	params := map[string]any{"name": name}
	return s.withWriteTx(ctx, func(tx neo4j.ManagedTransaction) error {
		if err := checkNodeExists(ctx, tx, lockScopeQuery, params, "scope", name); err != nil {
			return err
		}
		row, _, err := querySingleRow(ctx, tx,
			"MATCH (s:Scope {name: $name})<-[:APPLIES_IN]-(i:Instruction) RETURN count(i) AS linked",
			params)
		if err != nil {
			return err
		}
		if linked, _ := row["linked"].(int64); linked > 0 {
			return fmt.Errorf("scope %q still has %d linked instruction(s)", name, linked)
		}
		return runAndConsume(ctx, tx, "MATCH (s:Scope {name: $name}) DETACH DELETE s", params)
	})
}
