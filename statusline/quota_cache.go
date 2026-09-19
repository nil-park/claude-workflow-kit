package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"
)

const (
	cliproxyCacheName = "claude-statusline-cliproxy-quota.json"
	cliproxyCacheTTL  = 30 * time.Second
)

type quotaCache struct {
	FetchedAt time.Time   `json:"fetched_at"`
	Auths     []codexAuth `json:"auths"`
	// RejectedKey is a truncated hash of the last rejected key, enough to
	// notice that the key changed without storing the key itself.
	RejectedKey string `json:"rejected_key,omitempty"`
}

// load bumps fetched_at even when the fetch fails, so an unreachable proxy is
// retried once per TTL instead of on every refresh.
func (s cliproxySource) load(now time.Time) []codexAuth {
	cached, ok := readQuotaCache(s.cachePath)
	if ok && cached.RejectedKey != "" && cached.RejectedKey == keyFingerprint(s.key) {
		return nil
	}
	if age := now.Sub(cached.FetchedAt); ok && cached.RejectedKey == "" && age >= 0 && age < cliproxyCacheTTL {
		return cached.Auths
	}
	auths, err := s.fetch()
	next := quotaCache{FetchedAt: now, Auths: auths}
	switch {
	case errors.Is(err, errKeyRejected):
		next.Auths = nil
		next.RejectedKey = keyFingerprint(s.key)
	case err != nil && cached.RejectedKey == "":
		next.Auths = cached.Auths
	}
	writeQuotaCache(s.cachePath, next)
	return next.Auths
}

func keyFingerprint(key string) string {
	sum := sha256.Sum256([]byte(key))
	return hex.EncodeToString(sum[:4])
}

func readQuotaCache(path string) (quotaCache, bool) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return quotaCache{}, false
	}
	var cache quotaCache
	if err := json.Unmarshal(raw, &cache); err != nil {
		return quotaCache{}, false
	}
	return cache, true
}

// writeQuotaCache renames a finished file into place so a statusline process
// from another session never reads a half-written cache.
func writeQuotaCache(path string, cache quotaCache) {
	raw, err := json.Marshal(cache)
	if err != nil {
		return
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), filepath.Base(path)+".*.tmp")
	if err != nil {
		return
	}
	_, writeErr := tmp.Write(raw)
	closeErr := tmp.Close()
	if writeErr != nil || closeErr != nil || os.Rename(tmp.Name(), path) != nil {
		os.Remove(tmp.Name())
	}
}
