package main

import (
	"context"
	"fmt"
	"slices"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

const recallRemindersQuery = `
MATCH (r:Reminder)-[:APPLIES_IN]->(s:Scope)
WHERE s.name IN $scopes AND r.key IS NOT NULL AND r.embedding IS NOT NULL
WITH DISTINCT r LIMIT $candidateLimit
WITH r, [(r)-[:APPLIES_IN]->(t:Scope) | t.name] AS scopes
WHERE size(scopes) > 0 AND all(x IN scopes WHERE x IN $scopes)
WITH r, scopes, vector.similarity.cosine(r.embedding, $queryVec) AS score
WHERE score >= $threshold
RETURN r.key AS key, coalesce(r.title, '') AS title,
       coalesce(r.description, '') AS description,
       coalesce(r.body, '') AS body, scopes, score
ORDER BY score DESC, r.key ASC
LIMIT $limit`

func (s *store) recallReminders(ctx context.Context, scopes []string, queryVec []float64, threshold float64, limit int) ([]Reminder, error) {
	res, err := neo4j.ExecuteQuery(ctx, s.driver, recallRemindersQuery,
		map[string]any{
			"scopes":         dedup(scopes),
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
	out := make([]Reminder, 0, len(res.Records))
	for _, rec := range res.Records {
		r, err := toReminder(rec.AsMap())
		if err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, nil
}

func toReminder(m map[string]any) (Reminder, error) {
	key, ok := m["key"].(string)
	if !ok {
		return Reminder{}, fmt.Errorf("reminder key has unexpected type %T", m["key"])
	}
	title, ok := m["title"].(string)
	if !ok {
		return Reminder{}, fmt.Errorf("reminder title has unexpected type %T", m["title"])
	}
	desc, ok := m["description"].(string)
	if !ok {
		return Reminder{}, fmt.Errorf("reminder description has unexpected type %T", m["description"])
	}
	body, ok := m["body"].(string)
	if !ok {
		return Reminder{}, fmt.Errorf("reminder body has unexpected type %T", m["body"])
	}
	score, ok := m["score"].(float64)
	if !ok {
		return Reminder{}, fmt.Errorf("reminder score has unexpected type %T", m["score"])
	}
	rawScopes, ok := m["scopes"].([]any)
	if !ok {
		return Reminder{}, fmt.Errorf("reminder scopes has unexpected type %T", m["scopes"])
	}
	scopes := make([]string, len(rawScopes))
	for i, raw := range rawScopes {
		s, ok := raw.(string)
		if !ok {
			return Reminder{}, fmt.Errorf("reminder scope has unexpected type %T", raw)
		}
		scopes[i] = s
	}
	slices.Sort(scopes)
	return Reminder{Key: key, Title: title, Description: desc, Body: body, Scopes: scopes, Score: score}, nil
}
