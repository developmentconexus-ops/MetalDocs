package main

import (
	"os"
	"path/filepath"
	"testing"
)

// TestScanTestCitations_Fixture builds a small tree with a *_test.go file that
// cites a REQ ID, a non-test .go file that cites one too (must be ignored —
// only *_test.go counts as evidence kind (a) per contract §2.1), and a nested
// directory to prove the walk recurses.
func TestScanTestCitations_Fixture(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, dir, "internal/modules/foo/application/service_test.go",
		"package application\n// covers REQ-FOO-1 and REQ-FOO-2 in the same file.\nfunc TestX(t *testing.T) {}\n")
	mustWrite(t, dir, "internal/modules/foo/application/service.go",
		"package application\n// REQ-FOO-3 mentioned here must NOT count — not a _test.go file.\n")
	mustWrite(t, dir, "apps/api/cmd/metaldocs-api/chain_test.go",
		"package main\n// REQ-MW-7 chain order.\n")

	got, err := ScanTestCitations(dir, []string{"internal", "apps"})
	if err != nil {
		t.Fatalf("ScanTestCitations: %v", err)
	}

	if files, ok := got["REQ-FOO-1"]; !ok || len(files) != 1 {
		t.Errorf("REQ-FOO-1 citations = %v, want 1 file", files)
	}
	if files, ok := got["REQ-FOO-2"]; !ok || len(files) != 1 {
		t.Errorf("REQ-FOO-2 citations = %v, want 1 file", files)
	}
	if _, ok := got["REQ-FOO-3"]; ok {
		t.Errorf("REQ-FOO-3 must not be counted — only cited in a non-test file")
	}
	if files, ok := got["REQ-MW-7"]; !ok || len(files) != 1 {
		t.Errorf("REQ-MW-7 citations = %v, want 1 file", files)
	}
}

// TestScanTestCitations_RealTree smoke-tests the scan against the real repo:
// the F9.2 spec/plan record ~10 REQ-* matches across internal/+apps/ *_test.go
// today. Assert a lower bound so the scan proves it is actually walking the
// real tree rather than a stricter-than-necessary exact count that would be
// brittle across unrelated future test edits.
func TestScanTestCitations_RealTree(t *testing.T) {
	root := repoRootForTest(t)
	got, err := ScanTestCitations(root, []string{"internal", "apps"})
	if err != nil {
		t.Fatalf("ScanTestCitations real tree: %v", err)
	}
	total := 0
	for _, files := range got {
		total += len(files)
	}
	if total < 5 {
		t.Errorf("real-tree citation count = %d, want >= 5 (spec records ~10 REQ-* matches)", total)
	}
}

func mustWrite(t *testing.T, root, rel, content string) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", rel, err)
	}
}

func repoRootForTest(t *testing.T) string {
	t.Helper()
	abs, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("resolve repo root: %v", err)
	}
	return abs
}
