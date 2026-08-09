package main

import (
	"fmt"
	"os"
)

// GateInput bundles every path/parameter the gate needs, so tests can point
// it at a synthetic tree instead of the real repo.
type GateInput struct {
	DocPath    string
	ReportPath string
	ScanRoot   string
	ScanDirs   []string
	MapPath    string
}

// GateResult is the outcome of one gate run.
type GateResult struct {
	OK             bool
	UncoveredMusts []string
	Stale          bool
	Report         Report
	Rendered       string
}

// RunGate performs the default-mode check: extract, scan, merge, build the
// in-memory report, and compare it against the committed report file
// (anti-rot). It never writes to disk. OK is true only when there are zero
// uncovered MUST REQs AND the committed report is byte-identical to a fresh
// regeneration.
func RunGate(in GateInput) (GateResult, error) {
	reqs, err := ExtractReqs(in.DocPath)
	if err != nil {
		return GateResult{}, fmt.Errorf("extract reqs: %w", err)
	}

	testCitations, err := ScanTestCitations(in.ScanRoot, in.ScanDirs)
	if err != nil {
		return GateResult{}, fmt.Errorf("scan test citations: %w", err)
	}

	mapEntries, err := ParseMap(in.MapPath)
	if err != nil {
		return GateResult{}, fmt.Errorf("parse map: %w", err)
	}

	rep := BuildReport(reqs, testCitations, mapEntries)
	rendered := RenderReport(rep)

	committed, readErr := os.ReadFile(in.ReportPath)
	stale := readErr != nil || string(committed) != rendered

	uncovered := rep.UncoveredMusts()

	return GateResult{
		OK:             len(uncovered) == 0 && !stale,
		UncoveredMusts: uncovered,
		Stale:          stale,
		Report:         rep,
		Rendered:       rendered,
	}, nil
}

// WriteReport regenerates and writes the committed report file (the `-write`
// mode). It performs no gate check — callers should run RunGate afterward to
// confirm coverage.
func WriteReport(in GateInput) error {
	reqs, err := ExtractReqs(in.DocPath)
	if err != nil {
		return fmt.Errorf("extract reqs: %w", err)
	}
	testCitations, err := ScanTestCitations(in.ScanRoot, in.ScanDirs)
	if err != nil {
		return fmt.Errorf("scan test citations: %w", err)
	}
	mapEntries, err := ParseMap(in.MapPath)
	if err != nil {
		return fmt.Errorf("parse map: %w", err)
	}
	rep := BuildReport(reqs, testCitations, mapEntries)
	rendered := RenderReport(rep)
	if err := os.WriteFile(in.ReportPath, []byte(rendered), 0o644); err != nil { // #nosec G306 -- report is a committed wiki doc (wiki/architecture/req-traceability.md) that must stay git/CI readable at 0644; contains only REQ-ID coverage rows, no secret material.
		return fmt.Errorf("write report %s: %w", in.ReportPath, err)
	}
	return nil
}
