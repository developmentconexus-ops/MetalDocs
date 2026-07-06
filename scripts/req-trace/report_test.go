package main

import (
	"strings"
	"testing"
)

// TestBuildReport_CoverageJoin exercises the three evidence kinds and the
// uncovered-MUST detection (plan.md Task 3): a MUST covered by a test
// citation, a MUST covered by a manual commit-map entry, a MUST covered by a
// doc-annotation (auto-derived from Req.Annotation), a SHOULD with no
// evidence (reported, non-blocking), and a MUST with zero evidence (blocking).
func TestBuildReport_CoverageJoin(t *testing.T) {
	reqs := []Req{
		{ID: "REQ-A-1", Class: ClassMust},                                 // covered by test
		{ID: "REQ-B-1", Class: ClassMust},                                 // covered by map
		{ID: "REQ-C-1", Class: ClassMust, Annotation: "satisfied Wave 1"}, // covered by doc annotation
		{ID: "REQ-D-1", Class: ClassMust},                                 // UNCOVERED
		{ID: "REQ-E-1", Class: ClassShould},                               // uncovered SHOULD — non-blocking
	}
	testCitations := map[string][]string{
		"REQ-A-1": {"internal/foo/bar_test.go"},
	}
	mapEntries := []MapEntry{
		{Req: "REQ-B-1", Kind: "commit", Ref: "abc1234", Note: "M2 evidence, commit abc1234"},
	}

	rep := BuildReport(reqs, testCitations, mapEntries)

	byID := map[string]ReportRow{}
	for _, row := range rep.Rows {
		byID[row.ID] = row
	}

	if byID["REQ-A-1"].Evidence != EvidenceTest {
		t.Errorf("REQ-A-1 evidence = %v, want test", byID["REQ-A-1"].Evidence)
	}
	if byID["REQ-B-1"].Evidence != EvidenceCommit {
		t.Errorf("REQ-B-1 evidence = %v, want commit", byID["REQ-B-1"].Evidence)
	}
	if byID["REQ-C-1"].Evidence != EvidenceDoc {
		t.Errorf("REQ-C-1 evidence = %v, want doc", byID["REQ-C-1"].Evidence)
	}
	if byID["REQ-D-1"].Evidence != EvidenceNone {
		t.Errorf("REQ-D-1 evidence = %v, want none (uncovered)", byID["REQ-D-1"].Evidence)
	}
	if byID["REQ-E-1"].Evidence != EvidenceNone {
		t.Errorf("REQ-E-1 evidence = %v, want none", byID["REQ-E-1"].Evidence)
	}

	uncoveredMusts := rep.UncoveredMusts()
	if len(uncoveredMusts) != 1 || uncoveredMusts[0] != "REQ-D-1" {
		t.Fatalf("UncoveredMusts() = %v, want [REQ-D-1] only (SHOULD is non-blocking)", uncoveredMusts)
	}
}

func TestRenderReport_ContainsHeaderAndCommand(t *testing.T) {
	reqs := []Req{{ID: "REQ-A-1", Class: ClassMust}}
	rep := BuildReport(reqs, map[string][]string{"REQ-A-1": {"x_test.go"}}, nil)
	md := RenderReport(rep)
	if !strings.Contains(md, "go run ./scripts/req-trace") {
		t.Errorf("rendered report missing local regeneration command")
	}
	if !strings.Contains(md, "REQ-A-1") {
		t.Errorf("rendered report missing REQ-A-1 row")
	}
}
