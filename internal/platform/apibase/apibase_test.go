package apibase_test

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"metaldocs/internal/platform/apibase"
)

func TestBaseURLValue(t *testing.T) {
	if apibase.BaseURL != "/api/v1" {
		t.Fatalf("BaseURL = %q, want %q", apibase.BaseURL, "/api/v1")
	}
}

// TestNoHardCodedBaseURL is the guard that makes this constant load-bearing:
// a re-introduced literal is the hand-synced-constant defect class this
// program deletes, so it fails the build rather than drifting quietly.
func TestNoHardCodedBaseURL(t *testing.T) {
	lit := regexp.MustCompile(`BaseURL:\s*"/api/v1"`)
	root := filepath.Join("..", "..", "..")
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// Skip dot-directories (.git, .claude/worktrees/...): the latter can
		// hold gitignored, physically-nested git worktrees with their own
		// stale copies of these same files — not part of the tracked source
		// tree this guard polices.
		if info.IsDir() && strings.HasPrefix(info.Name(), ".") && path != root {
			return filepath.SkipDir
		}
		if info.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}
		if strings.Contains(path, string(filepath.Separator)+"api.gen.go") {
			return nil
		}
		b, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		if lit.Match(b) {
			t.Errorf("%s hard-codes BaseURL; use apibase.BaseURL", path)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
