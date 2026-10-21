package main

import (
	"errors"
	"strconv"
	"syscall"
)

const (
	processQueryLimitedInformation = 0x1000
	stillActive                    = 259
)

// processAlive checks the exit code because OpenProcess still succeeds for an
// exited process while any handle to it remains open.
func processAlive(pid int) bool {
	h, err := syscall.OpenProcess(processQueryLimitedInformation, false, uint32(pid))
	if errors.Is(err, syscall.ERROR_ACCESS_DENIED) {
		return true
	}
	if err != nil {
		return false
	}
	defer syscall.CloseHandle(h)
	var code uint32
	if err := syscall.GetExitCodeProcess(h, &code); err != nil {
		return false
	}
	return code == stillActive
}

// processStart returns the creation time as a decimal FILETIME, the form
// Claude Code writes to procStart on Windows.
func processStart(pid int) (string, bool) {
	h, err := syscall.OpenProcess(processQueryLimitedInformation, false, uint32(pid))
	if err != nil {
		return "", false
	}
	defer syscall.CloseHandle(h)
	var creation, exit, kernel, user syscall.Filetime
	if err := syscall.GetProcessTimes(h, &creation, &exit, &kernel, &user); err != nil {
		return "", false
	}
	filetime := uint64(creation.HighDateTime)<<32 | uint64(creation.LowDateTime)
	return strconv.FormatUint(filetime, 10), true
}
