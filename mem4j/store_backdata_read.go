package main

import (
	"context"
	"fmt"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func (s *store) listStatements(ctx context.Context, entity string) ([]Statement, error) {
	res, err := neo4j.ExecuteQuery(ctx, s.driver, `
		MATCH (n:Statement)-[:ABOUT]->(:Entity {name: $entity})
		WHERE n.key IS NOT NULL
		WITH DISTINCT n LIMIT $limit
		RETURN n.key AS key, coalesce(n.title, '') AS title, coalesce(n.body, '') AS body,
		       [(n)-[:ABOUT]->(e:Entity) | e.name] AS entities
		ORDER BY n.key`,
		map[string]any{"entity": entity, "limit": candidateLimit},
		neo4j.EagerResultTransformer, neo4j.ExecuteQueryWithDatabase(s.db), neo4j.ExecuteQueryWithReadersRouting(),
		neo4j.ExecuteQueryWithTransactionConfig(neo4j.WithTxTimeout(txTimeout)))
	if err != nil {
		return nil, err
	}
	return collectStatements(res.Records, false)
}

func (s *store) listEvents(ctx context.Context, entity string) ([]Event, error) {
	res, err := neo4j.ExecuteQuery(ctx, s.driver, `
		MATCH (n:Event)-[:ABOUT]->(:Entity {name: $entity})
		WHERE n.key IS NOT NULL
		WITH DISTINCT n LIMIT $limit
		RETURN n.key AS key, coalesce(n.title, '') AS title,
		       coalesce(n.occurredAt, '') AS occurredAt, coalesce(n.body, '') AS body,
		       [(n)-[:ABOUT]->(e:Entity) | e.name] AS entities
		ORDER BY n.occurredAt DESC, n.key ASC`,
		map[string]any{"entity": entity, "limit": candidateLimit},
		neo4j.EagerResultTransformer, neo4j.ExecuteQueryWithDatabase(s.db), neo4j.ExecuteQueryWithReadersRouting(),
		neo4j.ExecuteQueryWithTransactionConfig(neo4j.WithTxTimeout(txTimeout)))
	if err != nil {
		return nil, err
	}
	return collectEvents(res.Records, false)
}

func (s *store) recallStatements(ctx context.Context, entities []string, queryVec []float64, threshold float64, limit int) ([]Statement, error) {
	q := `MATCH (n:Statement) WHERE n.key IS NOT NULL AND n.embedding IS NOT NULL
		WITH n LIMIT $candidateLimit`
	if len(entities) > 0 {
		q = `MATCH (n:Statement)-[:ABOUT]->(e:Entity)
			WHERE e.name IN $entities AND n.key IS NOT NULL AND n.embedding IS NOT NULL
			WITH DISTINCT n LIMIT $candidateLimit`
	}
	q += `
		WITH n, vector.similarity.cosine(n.embedding, $queryVec) AS score
		WHERE score >= $threshold
		RETURN n.key AS key, coalesce(n.title, '') AS title, coalesce(n.body, '') AS body,
		       [(n)-[:ABOUT]->(t:Entity) | t.name] AS entities, score
		ORDER BY score DESC, n.key ASC LIMIT $limit`
	res, err := neo4j.ExecuteQuery(ctx, s.driver, q,
		map[string]any{
			"entities":       dedup(entities),
			"candidateLimit": candidateLimit,
			"queryVec":       queryVec,
			"threshold":      threshold,
			"limit":          limit,
		},
		neo4j.EagerResultTransformer, neo4j.ExecuteQueryWithDatabase(s.db), neo4j.ExecuteQueryWithReadersRouting(),
		neo4j.ExecuteQueryWithTransactionConfig(neo4j.WithTxTimeout(txTimeout)))
	if err != nil {
		return nil, err
	}
	return collectStatements(res.Records, true)
}

func (s *store) recallEvents(ctx context.Context, entities []string, queryVec []float64, threshold float64, limit int) ([]Event, error) {
	q := `MATCH (n:Event) WHERE n.key IS NOT NULL AND n.embedding IS NOT NULL
		WITH n LIMIT $candidateLimit`
	if len(entities) > 0 {
		q = `MATCH (n:Event)-[:ABOUT]->(e:Entity)
			WHERE e.name IN $entities AND n.key IS NOT NULL AND n.embedding IS NOT NULL
			WITH DISTINCT n LIMIT $candidateLimit`
	}
	q += `
		WITH n, vector.similarity.cosine(n.embedding, $queryVec) AS score
		WHERE score >= $threshold
		RETURN n.key AS key, coalesce(n.title, '') AS title,
		       coalesce(n.occurredAt, '') AS occurredAt, coalesce(n.body, '') AS body,
		       [(n)-[:ABOUT]->(t:Entity) | t.name] AS entities, score
		ORDER BY score DESC, n.occurredAt DESC, n.key ASC LIMIT $limit`
	res, err := neo4j.ExecuteQuery(ctx, s.driver, q,
		map[string]any{
			"entities":       dedup(entities),
			"candidateLimit": candidateLimit,
			"queryVec":       queryVec,
			"threshold":      threshold,
			"limit":          limit,
		},
		neo4j.EagerResultTransformer, neo4j.ExecuteQueryWithDatabase(s.db), neo4j.ExecuteQueryWithReadersRouting(),
		neo4j.ExecuteQueryWithTransactionConfig(neo4j.WithTxTimeout(txTimeout)))
	if err != nil {
		return nil, err
	}
	return collectEvents(res.Records, true)
}

func (s *store) hubNeighbors(ctx context.Context, label, name string, depth int) (NeighborView, error) {
	checkQ := fmt.Sprintf("MATCH (n:%s {name: $name}) RETURN n.name AS name LIMIT 1", label)
	checkRes, err := neo4j.ExecuteQuery(ctx, s.driver, checkQ,
		map[string]any{"name": name},
		neo4j.EagerResultTransformer, neo4j.ExecuteQueryWithDatabase(s.db), neo4j.ExecuteQueryWithReadersRouting(),
		neo4j.ExecuteQueryWithTransactionConfig(neo4j.WithTxTimeout(txTimeout)))
	if err != nil {
		return NeighborView{}, err
	}
	if len(checkRes.Records) == 0 {
		return NeighborView{}, fmt.Errorf("%s %q not found", label, name)
	}
	q := fmt.Sprintf(`
		MATCH (root:%s {name: $name})
		OPTIONAL MATCH p = (root)-[r*1..%d]-(other:%s)
		WHERE all(rel IN r WHERE type(rel) <> 'APPLIES_IN' AND type(rel) <> 'ABOUT')
		  AND all(n IN nodes(p) WHERE n:%s)
		UNWIND coalesce(relationships(p), [null]) AS rel
		WITH DISTINCT rel WHERE rel IS NOT NULL
		RETURN startNode(rel).name AS fromName, type(rel) AS relType, endNode(rel).name AS toName
		ORDER BY fromName, relType, toName LIMIT $limit`, label, depth, label, label)
	res, err := neo4j.ExecuteQuery(ctx, s.driver, q,
		map[string]any{"name": name, "limit": candidateLimit},
		neo4j.EagerResultTransformer, neo4j.ExecuteQueryWithDatabase(s.db), neo4j.ExecuteQueryWithReadersRouting(),
		neo4j.ExecuteQueryWithTransactionConfig(neo4j.WithTxTimeout(txTimeout)))
	if err != nil {
		return NeighborView{}, err
	}
	view := NeighborView{Root: name}
	for _, rec := range res.Records {
		m := rec.AsMap()
		from, _ := m["fromName"].(string)
		relType, _ := m["relType"].(string)
		to, _ := m["toName"].(string)
		if from != "" && relType != "" && to != "" {
			view.Edges = append(view.Edges, HubEdge{From: from, Type: relType, To: to})
		}
	}
	if label == "Entity" {
		stmts, err := s.listStatements(ctx, name)
		if err != nil {
			return NeighborView{}, err
		}
		events, err := s.listEvents(ctx, name)
		if err != nil {
			return NeighborView{}, err
		}
		view.Statements = stmts
		view.Events = events
	}
	return view, nil
}
