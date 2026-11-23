package main

import (
	"fmt"
	"io"
	"strings"
)

func writeEntities(w io.Writer, entities []Entity, withScore bool) {
	for _, e := range entities {
		if withScore {
			fmt.Fprintf(w, "<entity name=%q isScope=%t score=\"%.3f\">\n%s\n</entity>\n\n",
				e.Name, e.IsScope, e.Score, strings.TrimRight(e.Description, "\n"))
			continue
		}
		fmt.Fprintf(w, "<entity name=%q isScope=%t>\n%s\n</entity>\n\n",
			e.Name, e.IsScope, strings.TrimRight(e.Description, "\n"))
	}
}

func writeStatements(w io.Writer, stmts []Statement, withScore bool) {
	for _, s := range stmts {
		if withScore {
			fmt.Fprintf(w, "<statement key=%q title=%q entities=%q score=\"%.3f\">\n%s\n</statement>\n\n",
				s.Key, s.Title, strings.Join(s.Entities, ", "), s.Score, strings.TrimRight(s.Body, "\n"))
			continue
		}
		fmt.Fprintf(w, "<statement key=%q title=%q entities=%q>\n%s\n</statement>\n\n",
			s.Key, s.Title, strings.Join(s.Entities, ", "), strings.TrimRight(s.Body, "\n"))
	}
}

func writeEvents(w io.Writer, events []Event, withScore bool) {
	for _, e := range events {
		if withScore {
			fmt.Fprintf(w, "<event key=%q title=%q occurredAt=%q entities=%q score=\"%.3f\">\n%s\n</event>\n\n",
				e.Key, e.Title, e.OccurredAt, strings.Join(e.Entities, ", "), e.Score, strings.TrimRight(e.Body, "\n"))
			continue
		}
		fmt.Fprintf(w, "<event key=%q title=%q occurredAt=%q entities=%q>\n%s\n</event>\n\n",
			e.Key, e.Title, e.OccurredAt, strings.Join(e.Entities, ", "), strings.TrimRight(e.Body, "\n"))
	}
}

func writeNeighbors(w io.Writer, view NeighborView) {
	for _, edge := range view.Edges {
		fmt.Fprintf(w, "<link from=%q type=%q to=%q />\n", edge.From, edge.Type, edge.To)
	}
	if len(view.Edges) > 0 && (len(view.Statements) > 0 || len(view.Events) > 0) {
		fmt.Fprintln(w)
	}
	writeStatements(w, view.Statements, false)
	writeEvents(w, view.Events, false)
}
