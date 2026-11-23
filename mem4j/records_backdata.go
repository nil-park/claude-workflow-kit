package main

import (
	"fmt"
	"slices"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func toEntity(m map[string]any, withScore bool) (Entity, error) {
	name, ok := m["name"].(string)
	if !ok {
		return Entity{}, fmt.Errorf("entity name has unexpected type %T", m["name"])
	}
	desc, ok := m["description"].(string)
	if !ok {
		return Entity{}, fmt.Errorf("entity description has unexpected type %T", m["description"])
	}
	isScope, _ := m["isScope"].(bool)
	var score float64
	if withScore {
		score, ok = m["score"].(float64)
		if !ok {
			return Entity{}, fmt.Errorf("entity score has unexpected type %T", m["score"])
		}
	}
	return Entity{Name: name, Description: desc, IsScope: isScope, Score: score}, nil
}

func collectStatements(records []*neo4j.Record, withScore bool) ([]Statement, error) {
	out := make([]Statement, 0, len(records))
	for _, rec := range records {
		m := rec.AsMap()
		key, ok := m["key"].(string)
		title, okTitle := m["title"].(string)
		body, okBody := m["body"].(string)
		entities, err := toStrings(m["entities"])
		if !ok || !okTitle || !okBody || err != nil {
			return nil, fmt.Errorf("statement record has unexpected field types")
		}
		slices.Sort(entities)
		var score float64
		if withScore {
			score, ok = m["score"].(float64)
			if !ok {
				return nil, fmt.Errorf("statement score has unexpected type %T", m["score"])
			}
		}
		out = append(out, Statement{Key: key, Title: title, Body: body, Entities: entities, Score: score})
	}
	return out, nil
}

func collectEvents(records []*neo4j.Record, withScore bool) ([]Event, error) {
	out := make([]Event, 0, len(records))
	for _, rec := range records {
		m := rec.AsMap()
		key, ok := m["key"].(string)
		title, okTitle := m["title"].(string)
		occ, okOcc := m["occurredAt"].(string)
		body, okBody := m["body"].(string)
		entities, err := toStrings(m["entities"])
		if !ok || !okTitle || !okOcc || !okBody || err != nil {
			return nil, fmt.Errorf("event record has unexpected field types")
		}
		slices.Sort(entities)
		var score float64
		if withScore {
			score, ok = m["score"].(float64)
			if !ok {
				return nil, fmt.Errorf("event score has unexpected type %T", m["score"])
			}
		}
		out = append(out, Event{Key: key, Title: title, OccurredAt: occ, Body: body, Entities: entities, Score: score})
	}
	return out, nil
}
