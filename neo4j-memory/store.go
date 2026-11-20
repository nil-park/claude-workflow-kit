package main

import (
	"context"
	"fmt"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

// seedOrigin, joined with the manifest owner, marks nodes written by `load`,
// so a reload soft-deletes only that owner's dropped instructions.
const seedOrigin = "plugins-seed"

// candidateLimit bounds the read; hitting it means the graph outgrew this PoC.
const candidateLimit = 1000

const txTimeout = 20 * time.Second

var constraintQueries = []string{
	"CREATE CONSTRAINT instruction_key IF NOT EXISTS FOR (i:Instruction) REQUIRE i.key IS UNIQUE",
	"CREATE CONSTRAINT scope_name IF NOT EXISTS FOR (s:Scope) REQUIRE s.name IS UNIQUE",
}

// Locked nodes are skipped everywhere: the user protects them from agents,
// and the loader writes as an agent. An entity becomes a scope only when it
// carries no other instruction-group label, since those labels exclude each
// other, and only when the name points at that one node: otherwise the
// scope_name constraint would abort the whole load.
var writeQueries = []string{
	`UNWIND $scopes AS s
	 MATCH (e:Entity {name: s.name})
	 WITH s, collect(e) AS entities
	 WHERE size(entities) = 1 AND NOT EXISTS { MATCH (:Scope {name: s.name}) }
	 WITH entities[0] AS e
	 WHERE NOT e:Instruction AND NOT e:Reminder
	   AND e.deletedAt IS NULL AND coalesce(e.locked, false) = false
	 SET e:Scope`,
	`UNWIND $scopes AS s
	 MERGE (n:Scope {name: s.name})
	 WITH n, s WHERE coalesce(n.locked, false) = false
	 SET n.description = s.description, n.deletedAt = null,
	     n.disableModelInvocation = CASE WHEN s.disableModelInvocation THEN true END`,
	`UNWIND $instructions AS x
	 MERGE (i:Instruction {key: x.key})
	 WITH i, x WHERE coalesce(i.locked, false) = false
	 SET i.title = x.title, i.origin = $origin, i.deletedAt = null,
	     i.body = CASE WHEN x.body = '' THEN null ELSE x.body END,
	     i.slots = CASE WHEN x.slots = '' THEN null ELSE x.slots END
	 WITH i
	 OPTIONAL MATCH (i)-[r:APPLIES_IN|INCLUDES]->()
	 DELETE r`,
	`UNWIND $instructions AS x
	 MATCH (i:Instruction {key: x.key}) WHERE coalesce(i.locked, false) = false
	 UNWIND x.scopes AS name
	 MATCH (s:Scope {name: name})
	 MERGE (i)-[:APPLIES_IN]->(s)`,
	`UNWIND $instructions AS x
	 MATCH (i:Instruction {key: x.key}) WHERE coalesce(i.locked, false) = false
	 UNWIND x.includes AS childKey
	 MATCH (c:Instruction {key: childKey})
	 MERGE (i)-[:INCLUDES]->(c)`,
	`MATCH (i:Instruction {origin: $origin})
	 WHERE NOT i.key IN [x IN $instructions | x.key]
	   AND i.deletedAt IS NULL AND coalesce(i.locked, false) = false
	 SET i.deletedAt = datetime()`,
}

// An instruction linked to a soft-deleted scope never applies: dropping that
// scope from its list would widen the AND condition instead.
const candidateQuery = `
MATCH (s:Scope)<-[:APPLIES_IN]-(i:Instruction)
WHERE s.name IN $scopes AND s.deletedAt IS NULL
  AND i.deletedAt IS NULL AND i.key IS NOT NULL
  AND NOT EXISTS { (i)-[:APPLIES_IN]->(d:Scope) WHERE d.deletedAt IS NOT NULL }
WITH DISTINCT i LIMIT $limit
RETURN i.key AS key, coalesce(i.title, '') AS title, coalesce(i.body, '') AS body, coalesce(i.slots, '') AS slots,
       [(i)-[:APPLIES_IN]->(t:Scope) | t.name] AS scopes,
       [(i)-[:INCLUDES]->(c:Instruction) WHERE c.deletedAt IS NULL | c.key] AS includes`

func writeSeed(ctx context.Context, driver neo4j.DriverWithContext, db string, seed Seed) error {
	for _, q := range constraintQueries {
		if _, err := neo4j.ExecuteQuery(ctx, driver, q, nil, neo4j.EagerResultTransformer,
			neo4j.ExecuteQueryWithDatabase(db), neo4j.ExecuteQueryWithTransactionConfig(neo4j.WithTxTimeout(txTimeout))); err != nil {
			return err
		}
	}
	instructions, err := instructionParams(seed.Instructions)
	if err != nil {
		return err
	}
	params := map[string]any{"scopes": scopeParams(seed.Scopes), "instructions": instructions, "origin": seed.Origin}
	session := driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: db})
	defer session.Close(ctx)
	_, err = session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		for _, q := range writeQueries {
			res, err := tx.Run(ctx, q, params)
			if err != nil {
				return nil, err
			}
			// Consume so that a failing statement reports here, not at commit.
			if _, err := res.Consume(ctx); err != nil {
				return nil, err
			}
		}
		return nil, nil
	}, neo4j.WithTxTimeout(txTimeout))
	return err
}

// fetchCandidates returns every instruction linked to at least one of the
// given scopes; Merge then keeps those whose scopes all apply.
func fetchCandidates(ctx context.Context, driver neo4j.DriverWithContext, db string, scopes []string) ([]Instruction, error) {
	res, err := neo4j.ExecuteQuery(ctx, driver, candidateQuery,
		map[string]any{"scopes": scopes, "limit": candidateLimit},
		neo4j.EagerResultTransformer, neo4j.ExecuteQueryWithDatabase(db), neo4j.ExecuteQueryWithReadersRouting(),
		neo4j.ExecuteQueryWithTransactionConfig(neo4j.WithTxTimeout(txTimeout)))
	if err != nil {
		return nil, err
	}
	if len(res.Records) >= candidateLimit {
		return nil, fmt.Errorf("candidate instructions reached the limit of %d", candidateLimit)
	}
	insts := make([]Instruction, 0, len(res.Records))
	for _, rec := range res.Records {
		inst, err := toInstruction(rec.AsMap())
		if err != nil {
			return nil, err
		}
		insts = append(insts, inst)
	}
	return insts, nil
}
