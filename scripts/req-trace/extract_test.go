package main

import (
	"path/filepath"
	"testing"
)

// TestExtractReqs_Fixture drives ExtractReqs against a small fixture doc that
// exercises every extraction case called out in plan.md Task 1: plain (MUST),
// an annotated (MUST — satisfied ...) line, (SHOULD), (MAY), a non-REQ bullet
// that must be ignored, a duplicate ID mention (first normative line wins),
// and a MUST annotation that wraps onto a second physical line before the
// closing paren (must still collapse to one logical Req).
func TestExtractReqs_Fixture(t *testing.T) {
	path := filepath.Join("testdata", "fixture_doc.md")
	reqs, err := ExtractReqs(path)
	if err != nil {
		t.Fatalf("ExtractReqs: %v", err)
	}

	byID := map[string]Req{}
	for _, r := range reqs {
		byID[r.ID] = r
	}

	if len(reqs) != 5 {
		t.Fatalf("want 5 unique reqs (FOO-1, FOO-2, BAR-1, BAR-2, BAZ-1; the dup FOO-1 mention collapses), got %d: %+v", len(reqs), reqs)
	}

	foo1, ok := byID["REQ-FOO-1"]
	if !ok {
		t.Fatalf("REQ-FOO-1 missing")
	}
	if foo1.Class != ClassMust {
		t.Errorf("REQ-FOO-1 class = %v, want MUST", foo1.Class)
	}
	if foo1.Annotation != "" {
		t.Errorf("REQ-FOO-1 annotation = %q, want empty (plain MUST)", foo1.Annotation)
	}

	foo2, ok := byID["REQ-FOO-2"]
	if !ok {
		t.Fatalf("REQ-FOO-2 missing")
	}
	if foo2.Class != ClassMust {
		t.Errorf("REQ-FOO-2 class = %v, want MUST", foo2.Class)
	}
	if foo2.Annotation == "" {
		t.Errorf("REQ-FOO-2 annotation empty, want satisfied-by text captured")
	}

	bar1, ok := byID["REQ-BAR-1"]
	if !ok {
		t.Fatalf("REQ-BAR-1 missing")
	}
	if bar1.Class != ClassShould {
		t.Errorf("REQ-BAR-1 class = %v, want SHOULD", bar1.Class)
	}

	bar2, ok := byID["REQ-BAR-2"]
	if !ok {
		t.Fatalf("REQ-BAR-2 missing")
	}
	if bar2.Class != ClassMay {
		t.Errorf("REQ-BAR-2 class = %v, want MAY", bar2.Class)
	}

	baz1, ok := byID["REQ-BAZ-1"]
	if !ok {
		t.Fatalf("REQ-BAZ-1 missing (wrapped-annotation case)")
	}
	if baz1.Class != ClassMust {
		t.Errorf("REQ-BAZ-1 class = %v, want MUST", baz1.Class)
	}
	if baz1.Annotation == "" {
		t.Errorf("REQ-BAZ-1 annotation empty, want wrapped satisfied-by text captured")
	}

	if foo1.Line != 3 {
		t.Errorf("REQ-FOO-1 first mention line = %d, want 3 (the dup at a later line must not win)", foo1.Line)
	}
}

// TestExtractReqs_RealDoc smoke-tests extraction against the actual governing
// architecture doc: 69 unique REQ IDs, 62 MUST-classified, 7 SHOULD-classified,
// 0 MAY-classified. Was 67/61/6/0 (F9.2 spec/plan) until ADR 0095 split
// REQ-SEARCH-1 into REQ-SEARCH-1 (kept, MUST) + REQ-SEARCH-2 (new, SHOULD) —
// total +1, MUST unchanged, SHOULD +1. Then 68/61/7/0 until A3.2 added
// REQ-MW-8 (single ingress OpenAPI request validator) — total +1, MUST +1.
func TestExtractReqs_RealDoc(t *testing.T) {
	path := filepath.Join("..", "..", "wiki", "architecture", "backend-target-architecture.md")
	reqs, err := ExtractReqs(path)
	if err != nil {
		t.Fatalf("ExtractReqs real doc: %v", err)
	}
	if len(reqs) != 69 {
		t.Fatalf("want 69 unique REQ IDs, got %d", len(reqs))
	}
	var must, should, may int
	for _, r := range reqs {
		switch r.Class {
		case ClassMust:
			must++
		case ClassShould:
			should++
		case ClassMay:
			may++
		}
	}
	if must != 62 {
		t.Errorf("MUST count = %d, want 62", must)
	}
	if should != 7 {
		t.Errorf("SHOULD count = %d, want 7", should)
	}
	if may != 0 {
		t.Errorf("MAY count = %d, want 0", may)
	}
}
