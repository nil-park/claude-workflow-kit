package main

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"
)

type notifyCall struct {
	endpoint map[string]string
	notice   string
}

// stubNotify replaces notify for one test and records every call.
func stubNotify(t *testing.T, err error) *[]notifyCall {
	t.Helper()
	calls := &[]notifyCall{}
	original := notify
	notify = func(endpoint map[string]string, notice string, _ io.Writer) error {
		*calls = append(*calls, notifyCall{endpoint, notice})
		return err
	}
	t.Cleanup(func() { notify = original })
	return calls
}

func writeBody(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "body.md")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func registerPeer(t *testing.T, dir string) {
	t.Helper()
	if err := runRegister([]string{"--dir", dir, "--role", "peer"}, io.Discard, io.Discard, fakeEnv(agyEnv)); err != nil {
		t.Fatal(err)
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}

func TestPushWithoutEndpointOnlyWritesFile(t *testing.T) {
	calls := stubNotify(t, nil)
	dir := t.TempDir()
	args := []string{"--dir", dir, "--to", "main", "--round", "7", "--file", writeBody(t, "회신 본문")}
	if err := runPush(args, io.Discard); err != nil {
		t.Fatal(err)
	}
	if got := readFile(t, filepath.Join(dir, "round-007-to-main.md")); got != "회신 본문" {
		t.Errorf("message = %q", got)
	}
	if len(*calls) != 0 {
		t.Errorf("notify called %d times", len(*calls))
	}
}

func TestPushNotifiesRegisteredReceiver(t *testing.T) {
	calls := stubNotify(t, nil)
	dir := t.TempDir()
	registerPeer(t, dir)
	args := []string{"--dir", dir, "--to", "peer", "--round", "12", "--file", writeBody(t, "요청")}
	if err := runPush(args, io.Discard); err != nil {
		t.Fatal(err)
	}
	if len(*calls) != 1 {
		t.Fatalf("notify called %d times", len(*calls))
	}
	call := (*calls)[0]
	path := filepath.Join(dir, "round-012-to-peer.md")
	if want := "duo-relay: round 12 message for peer is at " + path; call.notice != want {
		t.Errorf("notice = %q, want %q", call.notice, want)
	}
	if call.endpoint[conversationIDKey] != agyEnv[conversationIDKey] {
		t.Errorf("endpoint = %v", call.endpoint)
	}
}

func TestPushOverwritesSameRound(t *testing.T) {
	stubNotify(t, nil)
	dir := t.TempDir()
	for _, body := range []string{"first", "second"} {
		args := []string{"--dir", dir, "--to", "main", "--round", "1", "--file", writeBody(t, body)}
		if err := runPush(args, io.Discard); err != nil {
			t.Fatal(err)
		}
	}
	if got := readFile(t, roundPath(dir, 1, "main")); got != "second" {
		t.Errorf("message = %q", got)
	}
}

func TestPushFailsWhenNotifyFails(t *testing.T) {
	stubNotify(t, errors.New("agy failed"))
	dir := t.TempDir()
	registerPeer(t, dir)
	args := []string{"--dir", dir, "--to", "peer", "--round", "1", "--file", writeBody(t, "요청")}
	if code := run(append([]string{"push"}, args...), io.Discard, io.Discard); code != exitError {
		t.Errorf("exit code = %d", code)
	}
}

func TestPushFailsOnEndpointWithEmptyField(t *testing.T) {
	calls := stubNotify(t, nil)
	dir := t.TempDir()
	broken := `{"ANTIGRAVITY_CONVERSATION_ID": "id", "ANTIGRAVITY_LS_ADDRESS": ""}`
	if err := os.WriteFile(endpointPath(dir, "peer"), []byte(broken), 0o600); err != nil {
		t.Fatal(err)
	}
	args := []string{"--dir", dir, "--to", "peer", "--round", "1", "--file", writeBody(t, "요청")}
	if err := runPush(args, io.Discard); err == nil {
		t.Fatal("push succeeded with a broken endpoint")
	}
	if len(*calls) != 0 {
		t.Errorf("notify called %d times", len(*calls))
	}
}
