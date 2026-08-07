package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"metaldocs/tools/cilint/internal/analyzers"
)

func mkFinding(analyzer, file, message string, line int) analyzers.Finding {
	return analyzers.Finding{Analyzer: analyzer, File: file, Line: line, Message: message}
}

func writeBaselineFixture(t *testing.T, bl *Baseline) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "baseline.json")
	if err := writeBaseline(path, bl); err != nil {
		t.Fatalf("writeBaseline: %v", err)
	}
	return path
}

func TestCompareBaseline_ExactMatch_Passes(t *testing.T) {
	bl := &Baseline{
		Doc: baselineDoc,
		Entries: []BaselineEntry{
			{Analyzer: "hgcrossmodule", File: "internal/modules/approval/x.go", Message: "reads foreign table", Count: 2},
		},
	}
	findings := []analyzers.Finding{
		mkFinding("hgcrossmodule", "internal/modules/approval/x.go", "reads foreign table", 10),
		mkFinding("hgcrossmodule", "internal/modules/approval/x.go", "reads foreign table", 20),
	}

	ok, lines := compareBaseline(findings, bl)
	if !ok {
		t.Fatalf("expected exact match to pass, got lines: %v", lines)
	}
	if len(lines) != 0 {
		t.Fatalf("expected no diagnostic lines on pass, got: %v", lines)
	}
}

func TestCompareBaseline_NewFinding_Fails(t *testing.T) {
	bl := &Baseline{
		Doc: baselineDoc,
		Entries: []BaselineEntry{
			{Analyzer: "hgcrossmodule", File: "internal/modules/approval/x.go", Message: "reads foreign table", Count: 1},
		},
	}
	findings := []analyzers.Finding{
		mkFinding("hgcrossmodule", "internal/modules/approval/x.go", "reads foreign table", 10),
		mkFinding("platformboundary", "internal/platform/new.go", "brand new violation", 5),
	}

	ok, lines := compareBaseline(findings, bl)
	if ok {
		t.Fatal("expected new (un-baselined) finding to fail")
	}
	found := false
	for _, l := range lines {
		if containsAll(l, "NEW VIOLATION", "internal/platform/new.go", "brand new violation") {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected a NEW VIOLATION line naming the new finding, got: %v", lines)
	}
}

func TestCompareBaseline_CountIncrease_Fails(t *testing.T) {
	bl := &Baseline{
		Doc: baselineDoc,
		Entries: []BaselineEntry{
			{Analyzer: "hgcrossmodule", File: "internal/modules/approval/x.go", Message: "reads foreign table", Count: 1},
		},
	}
	findings := []analyzers.Finding{
		mkFinding("hgcrossmodule", "internal/modules/approval/x.go", "reads foreign table", 10),
		mkFinding("hgcrossmodule", "internal/modules/approval/x.go", "reads foreign table", 20),
	}

	ok, lines := compareBaseline(findings, bl)
	if ok {
		t.Fatal("expected count increase (1 -> 2) to fail")
	}
	found := false
	for _, l := range lines {
		if containsAll(l, "COUNT INCREASED", "baseline=1", "actual=2") {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected a COUNT INCREASED line with baseline/actual counts, got: %v", lines)
	}
}

func TestCompareBaseline_CountDecrease_FailsWithStaleMessage(t *testing.T) {
	bl := &Baseline{
		Doc: baselineDoc,
		Entries: []BaselineEntry{
			{Analyzer: "hgcrossmodule", File: "internal/modules/approval/x.go", Message: "reads foreign table", Count: 2},
		},
	}
	findings := []analyzers.Finding{
		mkFinding("hgcrossmodule", "internal/modules/approval/x.go", "reads foreign table", 10),
	}

	ok, lines := compareBaseline(findings, bl)
	if ok {
		t.Fatal("expected count decrease (2 -> 1) to fail")
	}
	found := false
	for _, l := range lines {
		if containsAll(l, "STALE BASELINE", "--update-baseline", "baseline=2", "actual=1") {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected a STALE BASELINE line pointing at --update-baseline, got: %v", lines)
	}
}

func TestCompareBaseline_BaselineEntryWithZeroFindings_FailsAsStale(t *testing.T) {
	// A baselined violation that was entirely fixed (actual count 0) is the
	// same "stale baseline" case as a partial decrease — the debt is gone but
	// the record wasn't tightened.
	bl := &Baseline{
		Doc: baselineDoc,
		Entries: []BaselineEntry{
			{Analyzer: "hgcrossmodule", File: "internal/modules/approval/x.go", Message: "reads foreign table", Count: 1},
		},
	}
	var findings []analyzers.Finding

	ok, lines := compareBaseline(findings, bl)
	if ok {
		t.Fatal("expected a baseline entry with zero actual findings to fail as stale")
	}
	found := false
	for _, l := range lines {
		if containsAll(l, "STALE BASELINE", "actual=0") {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected a STALE BASELINE line with actual=0, got: %v", lines)
	}
}

func TestLoadBaseline_AbsentFile_ReturnsNilNoError(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "does-not-exist.json")

	bl, err := loadBaseline(path)
	if err != nil {
		t.Fatalf("expected no error for absent baseline file, got: %v", err)
	}
	if bl != nil {
		t.Fatalf("expected nil baseline for absent file, got: %+v", bl)
	}
}

func TestLoadBaseline_NormalizesBackslashPaths(t *testing.T) {
	bl := &Baseline{
		Doc: baselineDoc,
		Entries: []BaselineEntry{
			{Analyzer: "platformboundary", File: `internal\platform\tripwire\arms.go`, Message: "boundary violation", Count: 1},
		},
	}
	path := writeBaselineFixture(t, bl)

	loaded, err := loadBaseline(path)
	if err != nil {
		t.Fatalf("loadBaseline: %v", err)
	}
	if loaded.Entries[0].File != "internal/platform/tripwire/arms.go" {
		t.Fatalf("expected backslash path normalized to forward slashes, got: %q", loaded.Entries[0].File)
	}
}

func TestCompareBaseline_PathSeparatorNormalization_BothDirections(t *testing.T) {
	// Baseline stores backslashes (as if authored/generated on Windows); the
	// findings under test come in with forward slashes (as cilint would print
	// on Linux CI). They must still match after normalization.
	bl := &Baseline{
		Doc: baselineDoc,
		Entries: []BaselineEntry{
			// toForwardSlash, not filepath.ToSlash: this fixture is built by hand
			// and never passes through loadBaseline, so nothing else normalizes
			// it. On Linux filepath.ToSlash is the identity function, which left
			// the entry backslashed while the finding below was normalized — the
			// test could only pass on Windows.
			{Analyzer: "platformboundary", File: toForwardSlash(`internal\platform\tripwire\arms.go`), Message: "boundary violation", Count: 1},
		},
	}
	findings := normalizeFindings([]analyzers.Finding{
		mkFinding("platformboundary", `internal\platform\tripwire\arms.go`, "boundary violation", 17),
	})

	ok, lines := compareBaseline(findings, bl)
	if !ok {
		t.Fatalf("expected normalized backslash finding to match forward-slash baseline entry, got: %v", lines)
	}
}

func TestNormalizeFindings_ConvertsBackslashesToForwardSlashes(t *testing.T) {
	findings := []analyzers.Finding{
		mkFinding("platformboundary", `internal\platform\tripwire\arms.go`, "msg", 17),
	}
	out := normalizeFindings(findings)
	if out[0].File != "internal/platform/tripwire/arms.go" {
		t.Fatalf("expected forward-slash normalized path, got: %q", out[0].File)
	}
	// Original slice must be untouched (defensive copy).
	if findings[0].File != `internal\platform\tripwire\arms.go` {
		t.Fatalf("normalizeFindings must not mutate its input, got: %q", findings[0].File)
	}
}

func TestBuildBaseline_CarriesOverReasonAndOwner(t *testing.T) {
	existing := &Baseline{
		Doc: baselineDoc,
		Entries: []BaselineEntry{
			{Analyzer: "hgcrossmodule", File: "internal/modules/approval/x.go", Message: "reads foreign table", Count: 1, Reason: "pre-existing debt", Owner: "M3"},
		},
	}
	findings := []analyzers.Finding{
		mkFinding("hgcrossmodule", "internal/modules/approval/x.go", "reads foreign table", 10),
		mkFinding("hgcrossmodule", "internal/modules/approval/x.go", "reads foreign table", 20),
	}

	rebuilt := buildBaseline(findings, existing)
	if len(rebuilt.Entries) != 1 {
		t.Fatalf("expected 1 entry, got %d: %+v", len(rebuilt.Entries), rebuilt.Entries)
	}
	e := rebuilt.Entries[0]
	if e.Count != 2 {
		t.Fatalf("expected count updated to 2, got %d", e.Count)
	}
	if e.Reason != "pre-existing debt" || e.Owner != "M3" {
		t.Fatalf("expected reason/owner carried over, got reason=%q owner=%q", e.Reason, e.Owner)
	}
}

func TestBuildBaseline_NewKeyGetsNoAnnotation(t *testing.T) {
	findings := []analyzers.Finding{
		mkFinding("hgcrossmodule", "internal/modules/approval/x.go", "reads foreign table", 10),
	}
	rebuilt := buildBaseline(findings, nil)
	if len(rebuilt.Entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(rebuilt.Entries))
	}
	if rebuilt.Entries[0].Reason != "" || rebuilt.Entries[0].Owner != "" {
		t.Fatalf("expected no reason/owner for a fresh key, got reason=%q owner=%q",
			rebuilt.Entries[0].Reason, rebuilt.Entries[0].Owner)
	}
	if rebuilt.Doc != baselineDoc {
		t.Fatalf("expected the mandatory transitional doc string, got: %q", rebuilt.Doc)
	}
}

func TestWriteBaseline_RoundTrips(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "baseline.json")
	bl := &Baseline{
		Doc: baselineDoc,
		Entries: []BaselineEntry{
			{Analyzer: "hgcrossmodule", File: "a/b.go", Message: "m", Count: 3, Reason: "r", Owner: "o"},
		},
	}
	if err := writeBaseline(path, bl); err != nil {
		t.Fatalf("writeBaseline: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected file to exist: %v", err)
	}
	loaded, err := loadBaseline(path)
	if err != nil {
		t.Fatalf("loadBaseline: %v", err)
	}
	if loaded.Entries[0].Count != 3 || loaded.Entries[0].Reason != "r" || loaded.Entries[0].Owner != "o" {
		t.Fatalf("round-trip mismatch: %+v", loaded.Entries[0])
	}
}

func containsAll(s string, substrs ...string) bool {
	for _, sub := range substrs {
		if !strings.Contains(s, sub) {
			return false
		}
	}
	return true
}
