package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func endpointPath(dir, role string) string {
	return filepath.Join(dir, "endpoint-"+role+".json")
}

func roundPath(dir string, round int, role string) string {
	return filepath.Join(dir, fmt.Sprintf("round-%03d-to-%s.md", round, role))
}

func pidPath(dir, role string) string {
	return filepath.Join(dir, "wait-"+role+".pid")
}

// writeFileAtomic writes through a temporary file in the same directory so a
// reader polling for path never sees a partial file.
func writeFileAtomic(path string, data []byte) error {
	tmp, err := os.CreateTemp(filepath.Dir(path), "."+filepath.Base(path)+".tmp-*")
	if err != nil {
		return err
	}
	_, writeErr := tmp.Write(data)
	closeErr := tmp.Close()
	if writeErr == nil {
		writeErr = closeErr
	}
	if writeErr == nil {
		writeErr = os.Rename(tmp.Name(), path)
	}
	if writeErr != nil {
		os.Remove(tmp.Name())
	}
	return writeErr
}
