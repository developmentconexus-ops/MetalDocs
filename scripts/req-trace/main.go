package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

// scripts/req-trace is the F9.2 traceability gate. It extracts every REQ ID
// from wiki/architecture/backend-target-architecture.md, merges in test
// citations (internal/, apps/ *_test.go) and the manual evidence map
// (wiki/architecture/req-trace-map.yaml), and fails (exit 1) if any
// MUST-classified REQ has zero evidence, or if the committed report
// (wiki/architecture/req-traceability.md) has drifted from a fresh
// regeneration (anti-rot). SHOULD/MAY are reported only, never exit-affecting.
//
// Usage:
//
//	go run ./scripts/req-trace              # gate check (default)
//	go run ./scripts/req-trace -write       # regenerate the committed report
//	go run ./scripts/req-trace -repo <root> # point at a different repo root
func main() {
	repoRoot := flag.String("repo", ".", "repo root (paths below are resolved relative to this)")
	docPath := flag.String("doc", "", "override path to the governing architecture doc (default: <repo>/wiki/architecture/backend-target-architecture.md)")
	write := flag.Bool("write", false, "regenerate the committed report file instead of gating")
	flag.Parse()

	root, err := filepath.Abs(*repoRoot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "resolve repo root: %v\n", err)
		os.Exit(1)
	}

	doc := *docPath
	if doc == "" {
		doc = filepath.Join(root, "wiki", "architecture", "backend-target-architecture.md")
	}

	in := GateInput{
		DocPath:    doc,
		ReportPath: filepath.Join(root, "wiki", "architecture", "req-traceability.md"),
		ScanRoot:   root,
		ScanDirs:   []string{"internal", "apps"},
		MapPath:    filepath.Join(root, "wiki", "architecture", "req-trace-map.yaml"),
	}

	if *write {
		if err := WriteReport(in); err != nil {
			fmt.Fprintf(os.Stderr, "write report: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("wrote %s\n", in.ReportPath)
		return
	}

	result, err := RunGate(in)
	if err != nil {
		fmt.Fprintf(os.Stderr, "gate: %v\n", err)
		os.Exit(1)
	}

	if result.Stale {
		fmt.Println("STALE REPORT: committed wiki/architecture/req-traceability.md does not match a fresh regeneration.")
		fmt.Println("Run `go run ./scripts/req-trace -write` and commit the result.")
	}
	if len(result.UncoveredMusts) > 0 {
		fmt.Printf("UNCOVERED MUST REQ(s) (%d):\n", len(result.UncoveredMusts))
		for _, id := range result.UncoveredMusts {
			fmt.Printf("  %s\n", id)
		}
	}

	must, should, may, uncoveredMust := 0, 0, 0, 0
	for _, row := range result.Report.Rows {
		switch row.Class {
		case ClassMust:
			must++
			if row.Evidence == EvidenceNone {
				uncoveredMust++
			}
		case ClassShould:
			should++
			if row.Evidence == EvidenceNone {
				fmt.Printf("reported (SHOULD, non-blocking, no evidence): %s\n", row.ID)
			}
		case ClassMay:
			may++
		}
	}
	fmt.Printf("%d REQ IDs (%d MUST, %d SHOULD, %d MAY); %d MUST uncovered; stale=%v\n",
		len(result.Report.Rows), must, should, may, uncoveredMust, result.Stale)

	if !result.OK {
		os.Exit(1)
	}
}
