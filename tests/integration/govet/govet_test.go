//go:build integration
// +build integration

// Package govet_test isolates the repo-wide `go vet` gate in its own test
// binary. With a cold vet cache the command takes minutes; sharing a package
// (and its per-binary timeout budget) with DB scenario suites made the whole
// binary blow the 10m timeout. Keeping it alone gives vet the full budget.
package govet_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestGoVetPasses(t *testing.T) {
	root := repoRoot()
	cmd := exec.Command("go", "vet", "./internal/...")
	cmd.Dir = root
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go vet ./internal/... failed: %v\n%s", err, strings.TrimSpace(string(out)))
	}
}

func repoRoot() string {
	_, file, _, _ := runtime.Caller(0)
	dir := filepath.Dir(file)
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			panic("could not find repo root")
		}
		dir = parent
	}
}
