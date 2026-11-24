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
		firstErr   error
		errOnce    sync.Once
		wg         sync.WaitGroup
	)
	recordErr := func(err error) {
		if err != nil {
			errOnce.Do(func() {
				firstErr = err
				cancel()
			})
		}
	}
	if len(activeScopes) > 0 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var err error
			reminders, err = s.recallReminders(qCtx, activeScopes, vec, opts.threshold, opts.limit)
			recordErr(err)
		}()
	}
	wg.Add(2)
	go func() {
		defer wg.Done()
		var err error
		statements, err = s.recallStatements(qCtx, nil, vec, opts.threshold, opts.limit)
		recordErr(err)
	}()
	go func() {
		defer wg.Done()
		var err error
		events, err = s.recallEvents(qCtx, nil, vec, opts.threshold, opts.limit)
		recordErr(err)
	}()
	wg.Wait()

	if firstErr != nil {
		return "", firstErr
	}
	recordPromptMatches(entry, reminders, statements, events)
	return formatPromptContext(activeScopes, reminders, statements, events), nil
}
