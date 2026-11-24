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

	qCtx, cancel := context.WithCancel(ctx)
	defer cancel()

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
			reminders, remErr = s.recallReminders(qCtx, activeScopes, vec, opts.threshold, opts.limit)
			if remErr != nil {
				cancel()
			}
		}()
	}
	wg.Add(2)
	go func() {
		defer wg.Done()
		statements, stmtErr = s.recallStatements(qCtx, nil, vec, opts.threshold, opts.limit)
		if stmtErr != nil {
			cancel()
		}
	}()
	go func() {
		defer wg.Done()
		events, evtErr = s.recallEvents(qCtx, nil, vec, opts.threshold, opts.limit)
		if evtErr != nil {
			cancel()
		}
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
