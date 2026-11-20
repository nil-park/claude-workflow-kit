package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func parseGraphCommand(args []string, stdin io.Reader) (parsedCommand, error) {
	if len(args) == 0 {
		return parsedCommand{}, errors.New("usage: mem4j graph <query> [flags]")
	}
	switch args[0] {
	case "query":
		return parseGraphQuery(args[1:], stdin)
	default:
		return parsedCommand{}, fmt.Errorf("unknown graph action %q", args[0])
	}
}

func parseGraphQuery(args []string, stdin io.Reader) (parsedCommand, error) {
	fs, cred := newFlagSet("graph query")
	cypher := fs.String("cypher", "", "read-only Cypher query")
	file := fs.String("file", "", "path to Cypher file, or '-' for stdin")
	if err := parseFlags(fs, args); err != nil {
		return parsedCommand{}, err
	}
	q, provided, err := resolveTextInput(fs, "cypher", *cypher, *file, stdin)
	if err != nil {
		return parsedCommand{}, err
	}
	if !provided || strings.TrimSpace(q) == "" {
		return parsedCommand{}, errors.New("graph query requires a non-empty -cypher or -file")
	}
	return parsedCommand{
		credPath: *cred,
		exec: func(ctx context.Context, s *store, stdout, _ io.Writer) error {
			rows, err := s.queryRead(ctx, q)
			if err != nil {
				return err
			}
			return writeJSONLines(stdout, rows)
		},
	}, nil
}

func (s *store) queryRead(ctx context.Context, cypher string) ([]map[string]any, error) {
	session := s.driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: s.db,
		AccessMode:   neo4j.AccessModeRead,
	})
	defer session.Close(ctx)

	out, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		res, err := tx.Run(ctx, cypher, nil)
		if err != nil {
			return nil, err
		}
		rows := make([]map[string]any, 0)
		for res.Next(ctx) {
			if len(rows) >= candidateLimit {
				return nil, fmt.Errorf("query exceeded the limit of %d rows; narrow the query with LIMIT", candidateLimit)
			}
			rows = append(rows, convertMap(res.Record().AsMap()))
		}
		if err := res.Err(); err != nil {
			return nil, err
		}
		return rows, nil
	}, neo4j.WithTxTimeout(txTimeout))
	if err != nil {
		return nil, err
	}
	rows, ok := out.([]map[string]any)
	if !ok {
		return nil, fmt.Errorf("query result has unexpected type %T", out)
	}
	return rows, nil
}

func convertMap(m map[string]any) map[string]any {
	out := make(map[string]any, len(m))
	for k, v := range m {
		out[k] = convertValue(v)
	}
	return out
}

func convertValue(v any) any {
	switch x := v.(type) {
	case nil, bool, int64, float64, string:
		return x
	case neo4j.Node:
		return map[string]any{"labels": x.Labels, "properties": convertMap(x.Props)}
	case neo4j.Relationship:
		return map[string]any{"type": x.Type, "properties": convertMap(x.Props)}
	case neo4j.Path:
		nodes := make([]any, len(x.Nodes))
		for i, n := range x.Nodes {
			nodes[i] = convertValue(n)
		}
		rels := make([]any, len(x.Relationships))
		for i, r := range x.Relationships {
			rels[i] = convertValue(r)
		}
		return map[string]any{"nodes": nodes, "relationships": rels}
	case []any:
		out := make([]any, len(x))
		for i, elem := range x {
			out[i] = convertValue(elem)
		}
		return out
	case map[string]any:
		return convertMap(x)
	case time.Time:
		return x.Format(time.RFC3339Nano)
	case fmt.Stringer:
		return x.String()
	default:
		return fmt.Sprint(x)
	}
}

func writeJSONLines(w io.Writer, rows []map[string]any) error {
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	for _, row := range rows {
		if err := enc.Encode(row); err != nil {
			return err
		}
	}
	return nil
}
