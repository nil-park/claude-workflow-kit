package main

import (
	"context"
	"fmt"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

// seedOrigin marks nodes written by `load`, so a reload can soft-delete the
// ones the manifest no longer lists without touching other nodes.
const seedOrigin = "plugins-seed"

// candidateLimit bounds the read; hitting it means the graph outgrew this PoC.
const candidateLimit = 1000

const txTimeout = 20 * time.Second

var constraintQueries = []string{
	"CREATE CONSTRAINT instruction_key IF NOT EXISTS FOR (i:Instruction) REQUIRE i.key IS UNIQUE",
	"CREATE CONSTRAINT scope_name IF NOT EXISTS FOR (s:Scope) REQUIRE s.name IS UNIQUE",
}

// Locked nodes are skipped everywhere: the user protects them from agents,
// and the loader writes as an agent.
var writeQueries = []string{
	`UNWIND $scopes AS s
	 MATCH (e:Entity {name: s.name}) WHERE NOT e:Scope
	 SET e:Scope`,
	`UNWIND $scopes AS s
	 MERGE (n:Scope {name: s.name})
	 WITH n, s WHERE coalesce(n.locked, false) = false
	 SET n.description = s.description,
	     n.disableModelInvocation = CASE WHEN s.disableModelInvocation THEN true END`,
	`UNWIND $instructions AS x
	 MERGE (i:Instruction {key: x.key})
	 WITH i, x WHERE coalesce(i.locked, false) = false
	 SET i.body = x.body, i.slot = CASE WHEN x.slot = '' THEN null ELSE x.slot END, i.origin = $origin, i.deletedAt = null
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

const candidateQuery = `
MATCH (s:Scope)<-[:APPLIES_IN]-(i:Instruction)
WHERE s.name IN $scopes AND i.deletedAt IS NULL
WITH DISTINCT i LIMIT $limit
RETURN i.key AS key, coalesce(i.body, '') AS body, coalesce(i.slot, '') AS slot,
       [(i)-[:APPLIES_IN]->(t:Scope) | t.name] AS scopes,
       [(i)-[:INCLUDES]->(c:Instruction) WHERE c.deletedAt IS NULL | c.key] AS includes,
       EXISTS { (p:Instruction)-[:INCLUDES]->(i) WHERE p.deletedAt IS NULL } AS included`

func writeSeed(ctx context.Context, driver neo4j.DriverWithContext, db string, seed Seed) error {
	for _, q := range constraintQueries {
		if _, err := neo4j.ExecuteQuery(ctx, driver, q, nil, neo4j.EagerResultTransformer,
			neo4j.ExecuteQueryWithDatabase(db), neo4j.ExecuteQueryWithTransactionConfig(neo4j.WithTxTimeout(txTimeout))); err != nil {
			return err
		}
	}
	params := map[string]any{"scopes": scopeParams(seed.Scopes), "instructions": instructionParams(seed.Instructions), "origin": seedOrigin}
	session := driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: db})
	defer session.Close(ctx)
	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		for _, q := range writeQueries {
			if _, err := tx.Run(ctx, q, params); err != nil {
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
		m := rec.AsMap()
		insts = append(insts, Instruction{
			Key:      m["key"].(string),
			Body:     m["body"].(string),
			Slot:     m["slot"].(string),
			Scopes:   toStrings(m["scopes"]),
			Includes: toStrings(m["includes"]),
			Included: m["included"].(bool),
		})
	}
	return insts, nil
}

func scopeParams(scopes []ScopeSpec) []map[string]any {
	out := make([]map[string]any, len(scopes))
	for i, s := range scopes {
		out[i] = map[string]any{"name": s.Name, "description": s.Description, "disableModelInvocation": s.DisableModelInvocation}
	}
	return out
}

func instructionParams(insts []Instruction) []map[string]any {
	out := make([]map[string]any, len(insts))
	for i, inst := range insts {
		out[i] = map[string]any{
			"key": inst.Key, "body": inst.Body, "slot": inst.Slot,
			"scopes": inst.Scopes, "includes": nonNil(inst.Includes),
		}
	}
	return out
}

func toStrings(v any) []string {
	items, _ := v.([]any)
	out := make([]string, 0, len(items))
	for _, item := range items {
		out = append(out, item.(string))
	}
	return out
}

func nonNil(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}
