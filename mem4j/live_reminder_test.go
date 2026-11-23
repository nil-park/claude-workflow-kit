package main

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func TestLiveReminderCRUDAndRecall(t *testing.T) {
	ctx, s := openLiveStore(t)
	prefix := fmt.Sprintf("_livetest-%d-", time.Now().UnixNano())
	var autoKey string
	t.Cleanup(func() {
		cleanCtx, cleanCancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cleanCancel()
		if _, err := neo4j.ExecuteQuery(cleanCtx, s.driver, `
			MATCH (n)
			WHERE (n:Scope AND n.name STARTS WITH $prefix)
			   OR (n:Reminder AND (n.key STARTS WITH $prefix OR n.key = $autoKey))
			DETACH DELETE n`,
			map[string]any{"prefix": prefix, "autoKey": autoKey},
			neo4j.EagerResultTransformer, neo4j.ExecuteQueryWithDatabase(s.db)); err != nil {
			t.Errorf("cleanup reminder test nodes: %v", err)
		}
	})

	scopeA := prefix + "scope-a"
	scopeB := prefix + "scope-b"
	if err := s.createScope(ctx, scopeA, "scope A", false); err != nil {
		t.Fatal(err)
	}
	if err := s.createScope(ctx, scopeB, "scope B", false); err != nil {
		t.Fatal(err)
	}

	vecCommit := make([]float64, embeddingDimensions)
	vecCommit[0] = 1.0
	vecOther := make([]float64, embeddingDimensions)
	vecOther[1] = 1.0

	var err error
	autoKey, err = s.createReminder(ctx, "", "Commit lang", "when committing", "Use English.", vecCommit, []string{scopeA})
	if err != nil || len(autoKey) != shortIDLength {
		t.Fatalf("createReminder autoKey: key=%q err=%v", autoKey, err)
	}
	otherKey := prefix + "rem-other"
	if _, err := s.createReminder(ctx, otherKey, "Other", "other task", "Other body.", vecOther, []string{scopeA, scopeB}); err != nil {
		t.Fatalf("createReminder explicit: %v", err)
	}

	// Pre-filtering by scope intersection + cosine similarity
	got, err := s.recallReminders(ctx, []string{scopeA}, vecCommit, 0.8, 5)
	if err != nil {
		t.Fatalf("recallReminders: %v", err)
	}
	if len(got) != 1 || got[0].Key != autoKey || got[0].Score < 0.99 {
		t.Fatalf("recallReminders = %+v, want only %q", got, autoKey)
	}

	// Scope deletion refused while reminder is linked
	if err := s.deleteScope(ctx, scopeA); err == nil || !strings.Contains(err.Error(), "linked") {
		t.Fatalf("expected linked refusal on deleteScope, got %v", err)
	}

	// Last-scope unlink refused
	if err := s.unlinkReminder(ctx, autoKey, scopeA); err == nil || !strings.Contains(err.Error(), "last scope") {
		t.Fatalf("expected last-scope refusal, got %v", err)
	}

	// Edit, link, unlink, delete
	newDesc := "when creating a git commit"
	if err := s.editReminder(ctx, autoKey, nil, &newDesc, nil, vecCommit); err != nil {
		t.Fatalf("editReminder: %v", err)
	}
	if err := s.linkReminder(ctx, autoKey, scopeB); err != nil {
		t.Fatalf("linkReminder: %v", err)
	}
	if err := s.unlinkReminder(ctx, autoKey, scopeB); err != nil {
		t.Fatalf("unlinkReminder: %v", err)
	}
	if err := s.deleteReminder(ctx, autoKey); err != nil {
		t.Fatalf("deleteReminder autoKey: %v", err)
	}
	if err := s.deleteReminder(ctx, otherKey); err != nil {
		t.Fatalf("deleteReminder otherKey: %v", err)
	}
}
