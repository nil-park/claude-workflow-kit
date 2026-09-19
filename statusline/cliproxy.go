package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	cliproxyDefaultURL = "http://127.0.0.1:8317"
	cliproxyTimeout    = 300 * time.Millisecond
	// The auth-files listing also carries per-request history, so it can be
	// far larger than the few signals kept from it.
	cliproxyMaxBody = 8 << 20
)

// errKeyRejected marks a 401 or 403. CLIProxyAPI bans the client IP from the
// management API for 30 minutes after five failed keys, so a rejected key must
// not be retried on every TTL.
var errKeyRejected = errors.New("management key rejected")

type cliproxySource struct {
	baseURL   string
	key       string
	cachePath string
	client    *http.Client
}

func cliproxySourceFromEnv() (cliproxySource, bool) {
	key := os.Getenv("CLIPROXY_MANAGEMENT_KEY")
	if key == "" {
		return cliproxySource{}, false
	}
	// A per-user directory, unlike a shared /tmp, keeps another account from
	// planting the cache first. Without it the proxy would be hit every second.
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return cliproxySource{}, false
	}
	baseURL := os.Getenv("CLIPROXY_URL")
	if baseURL == "" {
		baseURL = cliproxyDefaultURL
	}
	return cliproxySource{
		baseURL:   baseURL,
		key:       key,
		cachePath: filepath.Join(cacheDir, cliproxyCacheName),
		client:    &http.Client{Timeout: cliproxyTimeout},
	}, true
}

func needsCLIProxyQuota(in statusInput) bool {
	return in.fiveHour == nil && in.sevenDay == nil && strings.HasPrefix(in.modelID, "gpt")
}

// withCLIProxyQuota never lets a proxy failure cost the other segments, so a
// panic here leaves the input as it was.
func withCLIProxyQuota(in statusInput, now time.Time) (out statusInput) {
	out = in
	defer func() {
		if recover() != nil {
			out = in
		}
	}()
	if !needsCLIProxyQuota(in) {
		return in
	}
	source, ok := cliproxySourceFromEnv()
	if !ok {
		return in
	}
	if snap, ok := selectSnapshot(source.load(now), in.modelID); ok {
		out.fiveHour, out.sevenDay = quotaWindows(snap, now)
	}
	return out
}

type authFilesResponse struct {
	Files []struct {
		Name        string                 `json:"name"`
		Provider    string                 `json:"provider"`
		Quota       rawSnapshot            `json:"quota"`
		ModelQuotas map[string]rawSnapshot `json:"model_quotas"`
	} `json:"files"`
}

// rawSnapshot keeps observed_at as text so one malformed timestamp drops
// only its own snapshot instead of failing the whole response.
type rawSnapshot struct {
	ObservedAt string            `json:"observed_at"`
	Signals    map[string]string `json:"signals"`
}

func (s cliproxySource) fetch() ([]codexAuth, error) {
	req, err := http.NewRequest(http.MethodGet, strings.TrimRight(s.baseURL, "/")+"/v0/management/auth-files", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+s.key)
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return nil, errKeyRejected
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("auth-files: status %d", resp.StatusCode)
	}
	var body authFilesResponse
	if err := json.NewDecoder(io.LimitReader(resp.Body, cliproxyMaxBody)).Decode(&body); err != nil {
		return nil, err
	}
	var auths []codexAuth
	for _, file := range body.Files {
		if !strings.EqualFold(strings.TrimSpace(file.Provider), "codex") {
			continue
		}
		auth := codexAuth{Name: file.Name, Quota: file.Quota.windowsOnly()}
		for model, snap := range file.ModelQuotas {
			if auth.ModelQuotas == nil {
				auth.ModelQuotas = map[string]codexSnapshot{}
			}
			auth.ModelQuotas[model] = snap.windowsOnly()
		}
		auths = append(auths, auth)
	}
	return auths, nil
}

func (r rawSnapshot) windowsOnly() codexSnapshot {
	observedAt, err := time.Parse(time.RFC3339Nano, r.ObservedAt)
	if err != nil {
		return codexSnapshot{}
	}
	signals := map[string]string{}
	for name, value := range r.Signals {
		name = http.CanonicalHeaderKey(name)
		for _, prefix := range codexWindowPrefixes {
			if strings.HasPrefix(name, prefix) {
				signals[name] = value
			}
		}
	}
	return codexSnapshot{ObservedAt: observedAt, Signals: signals}
}
