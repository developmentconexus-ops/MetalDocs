package main

import (
	"os"
	"path/filepath"
	"testing"
)

// TestRunGate_CleanTreeExitZero drives RunGate against a tiny synthetic repo
// tree where every MUST REQ is covered and the committed report matches a
// fresh regeneration exactly — must return ok=true, no uncovered, not stale.
func TestRunGate_CleanTreeExitZero(t *testing.T) {
	root := t.TempDir()
	docPath := filepath.Join(root, "wiki", "architecture", "backend-target-architecture.md")
	mustWrite(t, root, "wiki/architecture/backend-target-architecture.md",
		"- **REQ-A-1** Covered by test. (MUST)\n")
	mustWrite(t, root, "internal/foo/bar_test.go", "// REQ-A-1 covered here.\n")

	reqs, err := ExtractReqs(docPath)
	if err != nil {
		t.Fatalf("ExtractReqs: %v", err)
	}
	testCitations, err := ScanTestCitations(root, []string{"internal", "apps"})
	if err != nil {
		t.Fatalf("ScanTestCitations: %v", err)
	}
	rep := BuildReport(reqs, testCitations, nil)
	rendered := RenderReport(rep)

	reportPath := filepath.Join(root, "wiki", "architecture", "req-traceability.md")
	mustWrite(t, root, "wiki/architecture/req-traceability.md", rendered)

	result, err := RunGate(GateInput{
		DocPath:    docPath,
		ReportPath: reportPath,
		ScanRoot:   root,
		ScanDirs:   []string{"internal", "apps"},
		MapPath:    filepath.Join(root, "wiki", "architecture", "req-trace-map.yaml"),
	})
	if err != nil {
		t.Fatalf("RunGate: %v", err)
	}
	if !result.OK {
		t.Fatalf("result.OK = false, want true: uncovered=%v stale=%v", result.UncoveredMusts, result.Stale)
	}
	if len(result.UncoveredMusts) != 0 {
		t.Errorf("UncoveredMusts = %v, want none", result.UncoveredMusts)
	}
	if result.Stale {
		t.Errorf("Stale = true, want false")
	}
}

// TestRunGate_UncoveredMustFails is the NEGATIVE proof shape (a planted
// uncovered MUST REQ) exercised as a unit test.
func TestRunGate_UncoveredMustFails(t *testing.T) {
	root := t.TempDir()
	docPath := filepath.Join(root, "wiki", "architecture", "backend-target-architecture.md")
	mustWrite(t, root, "wiki/architecture/backend-target-architecture.md",
		"- **REQ-TST-99** Fake requirement. (MUST)\n")

	reqs, _ := ExtractReqs(docPath)
	rep := BuildReport(reqs, nil, nil)
	rendered := RenderReport(rep)
	mustWrite(t, root, "wiki/architecture/req-traceability.md", rendered)

	result, err := RunGate(GateInput{
		DocPath:    docPath,
		ReportPath: filepath.Join(root, "wiki", "architecture", "req-traceability.md"),
		ScanRoot:   root,
		ScanDirs:   []string{"internal", "apps"},
		MapPath:    filepath.Join(root, "wiki", "architecture", "req-trace-map.yaml"),
	})
	if err != nil {
		t.Fatalf("RunGate: %v", err)
	}
	if result.OK {
		t.Fatalf("result.OK = true, want false (REQ-TST-99 has no evidence)")
	}
	if len(result.UncoveredMusts) != 1 || result.UncoveredMusts[0] != "REQ-TST-99" {
		t.Fatalf("UncoveredMusts = %v, want [REQ-TST-99]", result.UncoveredMusts)
	}
}

// TestRunGate_StaleReportFails is the ANTI-ROT proof shape (hand-edited
// committed report differs from a fresh regeneration) exercised as a unit
// test — a one-char corruption must be caught even when coverage is fine.
func TestRunGate_StaleReportFails(t *testing.T) {
	root := t.TempDir()
	docPath := filepath.Join(root, "wiki", "architecture", "backend-target-architecture.md")
	mustWrite(t, root, "wiki/architecture/backend-target-architecture.md",
		"- **REQ-A-1** Covered by test. (MUST)\n")
	mustWrite(t, root, "internal/foo/bar_test.go", "// REQ-A-1 covered here.\n")

	reqs, _ := ExtractReqs(docPath)
	testCitations, _ := ScanTestCitations(root, []string{"internal", "apps"})
	rep := BuildReport(reqs, testCitations, nil)
	rendered := RenderReport(rep)
	// Corrupt one character of the committed report.
	corrupted := rendered[:len(rendered)-2] + "X\n"
	mustWrite(t, root, "wiki/architecture/req-traceability.md", corrupted)

	result, err := RunGate(GateInput{
		DocPath:    docPath,
		ReportPath: filepath.Join(root, "wiki", "architecture", "req-traceability.md"),
		ScanRoot:   root,
		ScanDirs:   []string{"internal", "apps"},
		MapPath:    filepath.Join(root, "wiki", "architecture", "req-trace-map.yaml"),
	})
	if err != nil {
		t.Fatalf("RunGate: %v", err)
	}
	if result.OK {
		t.Fatalf("result.OK = true, want false (committed report is stale/corrupted)")
	}
	if !result.Stale {
		t.Errorf("Stale = false, want true")
	}
}

// TestRunGate_MissingReportIsStale: a gate run before the report has ever
// been generated (-write) must also fail as "stale" (there's nothing to
// diff against — treat absence as drift, not a silent pass).
func TestRunGate_MissingReportIsStale(t *testing.T) {
	root := t.TempDir()
	docPath := filepath.Join(root, "wiki", "architecture", "backend-target-architecture.md")
	mustWrite(t, root, "wiki/architecture/backend-target-architecture.md",
		"- **REQ-A-1** Covered by test. (MUST)\n")
	mustWrite(t, root, "internal/foo/bar_test.go", "// REQ-A-1 covered here.\n")

	result, err := RunGate(GateInput{
		DocPath:    docPath,
		ReportPath: filepath.Join(root, "wiki", "architecture", "req-traceability.md"),
		ScanRoot:   root,
		ScanDirs:   []string{"internal", "apps"},
		MapPath:    filepath.Join(root, "wiki", "architecture", "req-trace-map.yaml"),
	})
	if err != nil {
		t.Fatalf("RunGate: %v", err)
	}
	if result.OK || !result.Stale {
		t.Fatalf("want OK=false Stale=true for a missing committed report, got OK=%v Stale=%v", result.OK, result.Stale)
	}
}

func TestWriteReport_RegeneratesFile(t *testing.T) {
	root := t.TempDir()
	docPath := filepath.Join(root, "wiki", "architecture", "backend-target-architecture.md")
	mustWrite(t, root, "wiki/architecture/backend-target-architecture.md",
		"- **REQ-A-1** Covered by test. (MUST)\n")
	mustWrite(t, root, "internal/foo/bar_test.go", "// REQ-A-1 covered here.\n")
	reportPath := filepath.Join(root, "wiki", "architecture", "req-traceability.md")

	if err := WriteReport(GateInput{
		DocPath:    docPath,
		ReportPath: reportPath,
		ScanRoot:   root,
		ScanDirs:   []string{"internal", "apps"},
		MapPath:    filepath.Join(root, "wiki", "architecture", "req-trace-map.yaml"),
	}); err != nil {
		t.Fatalf("WriteReport: %v", err)
	}
	if _, err := os.Stat(reportPath); err != nil {
		t.Fatalf("report not written: %v", err)
	}

	result, err := RunGate(GateInput{
		DocPath:    docPath,
		ReportPath: reportPath,
		ScanRoot:   root,
		ScanDirs:   []string{"internal", "apps"},
		MapPath:    filepath.Join(root, "wiki", "architecture", "req-trace-map.yaml"),
	})
	if err != nil {
		t.Fatalf("RunGate after WriteReport: %v", err)
	}
	if !result.OK {
		t.Fatalf("gate not clean immediately after -write: %+v", result)
	}
}
