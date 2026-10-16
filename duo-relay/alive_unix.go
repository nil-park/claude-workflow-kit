//go:build !windows

package main

import (
	"errors"
	"syscall"
)

// processAlive sends signal 0, which checks the pid without delivering
// anything. EPERM means the process exists under another user.
func processAlive(pid int) bool {
	err := syscall.Kill(pid, 0)
	return err == nil || errors.Is(err, syscall.EPERM)
}
