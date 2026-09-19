package main

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestLoadCachesWithinTTL(t *testing.T) {
	proxy := newFakeProxy(t)
	source := proxy.source(t)
	now := observed

	if auths := source.load(now); len(auths) != 2 || proxy.hits.Load() != 1 {
		t.Fatalf("first load: %d auths, %d hits", len(auths), proxy.hits.Load())
	}
	if auths := source.load(now.Add(cliproxyCacheTTL - time.Second)); len(auths) != 2 || proxy.hits.Load() != 1 {
		t.Errorf("load within TTL: %d auths, %d hits, want cache only", len(auths), proxy.hits.Load())
	}
	if source.load(now.Add(cliproxyCacheTTL)); proxy.hits.Load() != 2 {
		t.Errorf("load at TTL: %d hits, want a refetch", proxy.hits.Load())
	}
	if source.load(now.Add(-time.Second)); proxy.hits.Load() != 3 {
		t.Errorf("load before fetched_at: %d hits, want a refetch", proxy.hits.Load())
	}

	raw, err := os.ReadFile(source.cachePath)
	if err != nil {
		t.Fatal(err)
	}
	for _, secret := range []string{"secret-key", "user@example.com", "codex-b.json", "recent_requests"} {
		if strings.Contains(string(raw), secret) {
			t.Errorf("cache contains %q", secret)
		}
	}
}

func TestLoadKeepsCacheOnFailureAndWaitsTTL(t *testing.T) {
	proxy := newFakeProxy(t)
	source := proxy.source(t)
	source.load(observed)

	proxy.status.Store(http.StatusInternalServerError)
	failedAt := observed.Add(cliproxyCacheTTL)
	if auths := source.load(failedAt); len(auths) != 2 {
		t.Errorf("load after failure: %d auths, want the cached two", len(auths))
	}
	if source.load(failedAt.Add(time.Second)); proxy.hits.Load() != 2 {
		t.Errorf("hits = %d, want no retry within TTL of the failed attempt", proxy.hits.Load())
	}
	if entries, _ := os.ReadDir(filepath.Dir(source.cachePath)); len(entries) != 1 {
		t.Errorf("cache dir has %d entries, want no leftover temp file", len(entries))
	}
}

func TestLoadStopsAfterRejectedKeyUntilKeyChanges(t *testing.T) {
	proxy := newFakeProxy(t)
	source := proxy.source(t)
	source.key = "wrong"

	if auths := source.load(observed); auths != nil || proxy.hits.Load() != 1 {
		t.Fatalf("first load: %v after %d hits", auths, proxy.hits.Load())
	}
	if source.load(observed.Add(time.Hour)); proxy.hits.Load() != 1 {
		t.Errorf("hits = %d, want no retry with the same rejected key", proxy.hits.Load())
	}
	raw, _ := os.ReadFile(source.cachePath)
	if strings.Contains(string(raw), "wrong") {
		t.Error("cache contains the rejected key")
	}

	source.key = "secret-key"
	if auths := source.load(observed.Add(time.Hour + time.Second)); len(auths) != 2 || proxy.hits.Load() != 2 {
		t.Errorf("load with a new key: %d auths after %d hits, want an immediate refetch", len(auths), proxy.hits.Load())
	}
}
