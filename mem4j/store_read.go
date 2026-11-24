package main

import (
	"context"
	"fmt"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

// candidateLimit bounds reads; hitting it means the graph outgrew this CLI.
const candidateLimit = 1000

type Scope struct {
	Name                   string
	Description            string
	DisableModelInvocation bool
}

const listScopesQuery = `
MATCH (s:Scope)
RETURN s.name AS name, coalesce(s.description, '') AS description,
       coalesce(s.disableModelInvocation, false) AS disableModelInvocation
ORDER BY s.name
LIMIT $limit`

const candidateQuery = `
MATCH (s:Scope)<-[:APPLIES_IN]-(i:Instruction)
WHERE s.name IN $scopes AND i.key IS NOT NULL
WITH DISTINCT i LIMIT $limit
RETURN i.key AS key, coalesce(i.title, '') AS title, coalesce(i.body, '') AS body,
       [(i)-[:APPLIES_IN]->(t:Scope) | t.name] AS scopes`

func (s *store) fetchCandidates(ctx context.Context, scopes []string) ([]Instruction, error) {
	res, err := neo4j.ExecuteQuery(ctx, s.driver, candidateQuery,
		map[string]any{"scopes": scopes, "limit": candidateLimit},
		neo4j.EagerResultTransformer, neo4j.ExecuteQueryWithDatabase(s.db), neo4j.ExecuteQueryWithReadersRouting(),
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

func (s *store) missingScopes(ctx context.Context, scopes []string) ([]string, error) {
	res, err := neo4j.ExecuteQuery(ctx, s.driver, `
		UNWIND $scopes AS name
		OPTIONAL MATCH (s:Scope {name: name})
		WITH name, s
		WHERE s IS NULL
		RETURN name`,
		map[string]any{"scopes": dedup(scopes)},
		neo4j.EagerResultTransformer, neo4j.ExecuteQueryWithDatabase(s.db), neo4j.ExecuteQueryWithReadersRouting(),
		neo4j.ExecuteQueryWithTransactionConfig(neo4j.WithTxTimeout(txTimeout)))
	if err != nil {
		return nil, err
	}
	missing := make([]string, 0, len(res.Records))
	for _, rec := range res.Records {
		raw, _ := rec.Get("name")
		name, ok := raw.(string)
		if !ok {
			return nil, fmt.Errorf("scope name has unexpected type %T", raw)
		}
		missing = append(missing, name)
	}
	return missing, nil
}

func (s *store) listScopes(ctx context.Context) ([]Scope, error) {
	res, err := neo4j.ExecuteQuery(ctx, s.driver, listScopesQuery,
		map[string]any{"limit": candidateLimit},
		neo4j.EagerResultTransformer, neo4j.ExecuteQueryWithDatabase(s.db), neo4j.ExecuteQueryWithReadersRouting(),
		neo4j.ExecuteQueryWithTransactionConfig(neo4j.WithTxTimeout(txTimeout)))
	if err != nil {
		return nil, err
	}
	if len(res.Records) >= candidateLimit {
		return nil, fmt.Errorf("scopes reached the limit of %d", candidateLimit)
	}
	scopes := make([]Scope, 0, len(res.Records))
	for _, rec := range res.Records {
		m := rec.AsMap()
		name, ok := m["name"].(string)
		if !ok {
			return nil, fmt.Errorf("scope name has unexpected type %T", m["name"])
		}
		desc, ok := m["description"].(string)
		if !ok {
			return nil, fmt.Errorf("scope description has unexpected type %T", m["description"])
		}
		disable, ok := m["disableModelInvocation"].(bool)
		if !ok {
			return nil, fmt.Errorf("scope disableModelInvocation has unexpected type %T", m["disableModelInvocation"])
		}
		scopes = append(scopes, Scope{Name: name, Description: desc, DisableModelInvocation: disable})
	}
	return scopes, nil
}
