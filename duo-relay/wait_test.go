package main

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"
	"time"
)

func waitArgs(dir string, timeout int) []string {
	return []string{"--dir", dir, "--role", "main", "--round", "3", "--timeout", strconv.Itoa(timeout)}
}

func TestWaitPrintsMessageAndReleasesLock(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(roundPath(dir, 3, "main"), []byte("회신"), 0o600); err != nil {
		t.Fatal(err)
	}
	var stdout bytes.Buffer
	code, err := runWait(waitArgs(dir, 5), &stdout, io.Discard)
	if err != nil || code != exitOK {
		t.Fatalf("code = %d, err = %v", code, err)
	}
	if stdout.String() != "회신" {
		t.Errorf("stdout = %q", stdout.String())
	}
	if _, err := os.Stat(pidPath(dir, "main")); !os.IsNotExist(err) {
		t.Errorf("pid file remains: %v", err)
	}
}

func TestWaitPicksUpMessageWrittenLater(t *testing.T) {
	dir := t.TempDir()
	go func() {
		time.Sleep(3 * pollInterval)
		writeFileAtomic(roundPath(dir, 3, "main"), []byte("늦은 회신"))
	}()
	var stdout bytes.Buffer
	code, err := runWait(waitArgs(dir, 5), &stdout, io.Discard)
	if err != nil || code != exitOK {
		t.Fatalf("code = %d, err = %v", code, err)
	}
	if stdout.String() != "늦은 회신" {
		t.Errorf("stdout = %q", stdout.String())
	}
}

func TestWaitTimesOut(t *testing.T) {
	dir := t.TempDir()
	var stdout bytes.Buffer
	code, err := runWait(waitArgs(dir, 1), &stdout, io.Discard)
	if err != nil || code != exitTimeout {
		t.Fatalf("code = %d, err = %v", code, err)
	}
	if !strings.Contains(stdout.String(), "within 1 seconds") {
		t.Errorf("stdout = %q", stdout.String())
	}
	if _, err := os.Stat(pidPath(dir, "main")); !os.IsNotExist(err) {
		t.Errorf("pid file remains: %v", err)
	}
}

func TestWaitExitsWhenAnotherWaitIsAlive(t *testing.T) {
	dir := t.TempDir()
	// The test binary's parent (go test) outlives this test.
	other := strconv.Itoa(os.Getppid()) + "\n"
	if err := os.WriteFile(pidPath(dir, "main"), []byte(other), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(roundPath(dir, 3, "main"), []byte("회신"), 0o600); err != nil {
		t.Fatal(err)
	}
	var stdout bytes.Buffer
	code, err := runWait(waitArgs(dir, 5), &stdout, io.Discard)
	if err != nil || code != exitBusy {
		t.Fatalf("code = %d, err = %v", code, err)
	}
	if stdout.Len() != 0 {
		t.Errorf("stdout = %q", stdout.String())
	}
	if got := readFile(t, pidPath(dir, "main")); got != other {
		t.Errorf("pid file = %q, want %q", got, other)
	}
}

func TestWaitTakesOverLockOfExitedProcess(t *testing.T) {
	dir := t.TempDir()
	exited := exec.Command(os.Args[0], "-test.run=^$")
	if err := exited.Run(); err != nil {
		t.Fatal(err)
	}
	stale := strconv.Itoa(exited.Process.Pid)
	if err := os.WriteFile(pidPath(dir, "main"), []byte(stale), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(roundPath(dir, 3, "main"), []byte("회신"), 0o600); err != nil {
		t.Fatal(err)
	}
	code, err := runWait(waitArgs(dir, 5), io.Discard, io.Discard)
	if err != nil || code != exitOK {
		t.Fatalf("code = %d, err = %v", code, err)
	}
	if _, err := os.Stat(pidPath(dir, "main")); !os.IsNotExist(err) {
		t.Errorf("pid file remains: %v", err)
	}
}

func TestReleaseLockKeepsLockOfAnotherProcess(t *testing.T) {
	path := pidPath(t.TempDir(), "main")
	if err := os.WriteFile(path, []byte("12345\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	releaseLock(path, 54321)
	if _, err := os.Stat(path); err != nil {
		t.Errorf("pid file removed: %v", err)
	}
}
