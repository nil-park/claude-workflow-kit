package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

var observed = time.Date(2026, 9, 19, 6, 0, 0, 0, time.UTC)

type fakeProxy struct {
	server *httptest.Server
	hits   int
	status int
}

func newFakeProxy(t *testing.T) *fakeProxy {
	t.Helper()
	body, err := os.ReadFile("testdata/cliproxy-auth-files.json")
	if err != nil {
		t.Fatal(err)
	}
	p := &fakeProxy{status: http.StatusOK}
	p.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p.hits++
		if r.URL.Path != "/v0/management/auth-files" || r.Header.Get("Authorization") != "Bearer secret-key" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(p.status)
		w.Write(body)
	}))
	t.Cleanup(p.server.Close)
	return p
}

func (p *fakeProxy) source(t *testing.T) cliproxySource {
	return cliproxySource{
		baseURL:   p.server.URL + "/",
		key:       "secret-key",
		cachePath: filepath.Join(t.TempDir(), cliproxyCacheName),
		client:    &http.Client{Timeout: time.Second},
	}
}

func TestFetchKeepsOnlyCodexWindows(t *testing.T) {
	proxy := newFakeProxy(t)
	auths, err := proxy.source(t).fetch()
	if err != nil {
		t.Fatal(err)
	}
	if len(auths) != 2 || auths[0].Name != "codex-b.json" || auths[1].Name != "codex-a.json" {
		t.Fatalf("auths = %+v, want the two codex entries in response order", auths)
	}
	b := auths[0]
	if !b.Quota.ObservedAt.Equal(observed) || len(b.Quota.Signals) != 6 {
		t.Errorf("quota = %+v, want observed_at %v and the six Primary/Secondary signals", b.Quota, observed)
	}
	if _, ok := b.Quota.Signals["X-Codex-Bengalfox-Primary-Used-Percent"]; ok {
		t.Error("additional limit signals must be dropped")
	}
	if len(b.ModelQuotas["gpt-5.6-luna"].Signals) != 2 {
		t.Errorf("model_quotas = %+v", b.ModelQuotas)
	}
	if a := auths[1].Quota; len(a.Signals) != 0 {
		t.Errorf("snapshot with a malformed observed_at = %+v, want empty", a)
	}
}

func TestFetchRejectsNonOK(t *testing.T) {
	proxy := newFakeProxy(t)
	source := proxy.source(t)
	source.key = "wrong"
	if _, err := source.fetch(); err == nil {
		t.Error("fetch with a rejected key should fail")
	}
}

func TestLoadCachesWithinTTL(t *testing.T) {
	proxy := newFakeProxy(t)
	source := proxy.source(t)
	now := observed

	if auths := source.load(now); len(auths) != 2 || proxy.hits != 1 {
		t.Fatalf("first load: %d auths, %d hits", len(auths), proxy.hits)
	}
	if auths := source.load(now.Add(cliproxyCacheTTL - time.Second)); len(auths) != 2 || proxy.hits != 1 {
		t.Errorf("load within TTL: %d auths, %d hits, want cache only", len(auths), proxy.hits)
	}
	if source.load(now.Add(cliproxyCacheTTL)); proxy.hits != 2 {
		t.Errorf("load at TTL: %d hits, want a refetch", proxy.hits)
	}
	if source.load(now.Add(-time.Second)); proxy.hits != 3 {
		t.Errorf("load before fetched_at: %d hits, want a refetch", proxy.hits)
	}

	raw, err := os.ReadFile(source.cachePath)
	if err != nil {
		t.Fatal(err)
	}
	for _, secret := range []string{"secret-key", "user@example.com", "recent_requests"} {
		if strings.Contains(string(raw), secret) {
			t.Errorf("cache contains %q", secret)
		}
	}
}

func TestLoadKeepsCacheOnFailureAndWaitsTTL(t *testing.T) {
	proxy := newFakeProxy(t)
	source := proxy.source(t)
	source.load(observed)

	proxy.status = http.StatusInternalServerError
	failedAt := observed.Add(cliproxyCacheTTL)
	if auths := source.load(failedAt); len(auths) != 2 {
		t.Errorf("load after failure: %d auths, want the cached two", len(auths))
	}
	if source.load(failedAt.Add(time.Second)); proxy.hits != 2 {
		t.Errorf("hits = %d, want no retry within TTL of the failed attempt", proxy.hits)
	}
	if entries, _ := os.ReadDir(filepath.Dir(source.cachePath)); len(entries) != 1 {
		t.Errorf("cache dir has %d entries, want no leftover temp file", len(entries))
	}
}

func TestNeedsCLIProxyQuota(t *testing.T) {
	window := &quotaWindow{usedPercent: 1}
	cases := []struct {
		in   statusInput
		want bool
	}{
		{statusInput{modelID: "gpt-5.6-luna"}, true},
		{statusInput{modelID: "gpt-5.6-luna", sevenDay: window}, false},
		{statusInput{modelID: "claude-opus-5"}, false},
		{statusInput{}, false},
	}
	for _, c := range cases {
		if got := needsCLIProxyQuota(c.in); got != c.want {
			t.Errorf("needsCLIProxyQuota(%+v) = %v, want %v", c.in, got, c.want)
		}
	}
}

func TestWithCLIProxyQuota(t *testing.T) {
	proxy := newFakeProxy(t)
	// A cache directory that does not exist yet, set for every OS's lookup.
	home := filepath.Join(t.TempDir(), "home")
	for _, name := range []string{"LocalAppData", "XDG_CACHE_HOME", "HOME"} {
		t.Setenv(name, home)
	}
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("CLIPROXY_URL", proxy.server.URL)
	in := statusInput{modelID: "gpt-5.6-luna"}

	t.Setenv("CLIPROXY_MANAGEMENT_KEY", "")
	if got := withCLIProxyQuota(in, observed); got != in || proxy.hits != 0 {
		t.Errorf("without a key: %+v after %d hits, want input unchanged and no request", got, proxy.hits)
	}

	t.Setenv("CLIPROXY_MANAGEMENT_KEY", "secret-key")
	got := withCLIProxyQuota(in, observed)
	if windowString(got.fiveHour) != "10" || got.sevenDay != nil {
		t.Errorf("5h = %q, 7d = %+v, want the model snapshot's 10%% and no 7d", windowString(got.fiveHour), got.sevenDay)
	}
	if _, err := os.Stat(filepath.Join(cacheDir, cliproxyCacheName)); err != nil {
		t.Errorf("cache not written under the user cache dir: %v", err)
	}
}
