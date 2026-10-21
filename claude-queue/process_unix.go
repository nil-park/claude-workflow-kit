//go:build !windows

package main

import (
	"errors"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
)

// processAlive sends signal 0, which checks the pid without delivering
// anything. EPERM means the process exists under another user.
func processAlive(pid int) bool {
	err := syscall.Kill(pid, 0)
	return err == nil || errors.Is(err, syscall.EPERM)
}

// processStart runs ps with the same locale and time zone Claude Code uses,
// so the output matches procStart byte for byte.
func processStart(pid int) (string, bool) {
	cmd := exec.Command("ps", "-o", "lstart=", "-p", strconv.Itoa(pid))
	cmd.Env = append(os.Environ(), "LC_ALL=C", "TZ=UTC")
	out, err := cmd.Output()
	if err != nil {
		return "", false
	}
	start := strings.TrimSpace(string(out))
	return start, start != ""
}
