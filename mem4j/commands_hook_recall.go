package main

import (
	"context"
	"sync"
)

func recallPromptContext(ctx context.Context, s *store, opts hookOptions, activeScopes []string, prompt string, entry *hookLogEntry) (string, error) {
	emb, err := newVertexEmbedder(ctx, opts.gcpCredPath)
	if err != nil {
		return "", err
	}
	vec, err := emb.Embed(ctx, prompt, taskTypeQuery)
	if err != nil {
		return "", err
	}

	var (
		reminders  []Reminder
		statements []Statement
		events     []Event
		remErr     error
		stmtErr    error
		evtErr     error
		wg         sync.WaitGroup
	)
	if len(activeScopes) > 0 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			reminders, remErr = s.recallReminders(ctx, activeScopes, vec, opts.threshold, opts.limit)
		}()
	}
	wg.Add(2)
	go func() {
		defer wg.Done()
		statements, stmtErr = s.recallStatements(ctx, nil, vec, opts.threshold, opts.limit)
	}()
	go func() {
		defer wg.Done()
		events, evtErr = s.recallEvents(ctx, nil, vec, opts.threshold, opts.limit)
	}()
	wg.Wait()

	for _, qErr := range []error{remErr, stmtErr, evtErr} {
		if qErr != nil {
			return "", qErr
		}
	}
	recordPromptMatches(entry, reminders, statements, events)
	return formatPromptContext(activeScopes, reminders, statements, events), nil
}
