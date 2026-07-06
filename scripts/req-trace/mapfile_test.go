package main

import (
	"path/filepath"
	"testing"
)

func TestParseMap_Good(t *testing.T) {
	entries, err := ParseMap(filepath.Join("testdata", "map_good.yaml"))
	if err != nil {
		t.Fatalf("ParseMap: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("want 2 entries, got %d: %+v", len(entries), entries)
	}
	if entries[0].Req != "REQ-FOO-1" || entries[0].Kind != "commit" || entries[0].Ref != "abc1234" {
		t.Errorf("entry 0 = %+v, want REQ-FOO-1/commit/abc1234", entries[0])
	}
	if entries[0].Note == "" {
		t.Errorf("entry 0 note must be non-empty (source-naming requirement)")
	}
}

// TestParseMap_RejectsNonCommitKind is the anti-gaming guard from plan.md
// Task 2: kind `doc` is NOT accepted from the map — doc evidence only
// auto-derives from the REQ line's own annotation in the architecture doc.
func TestParseMap_RejectsNonCommitKind(t *testing.T) {
	_, err := ParseMap(filepath.Join("testdata", "map_bad_kind.yaml"))
	if err == nil {
		t.Fatalf("want error for kind: doc in the map (only kind: commit is accepted), got nil")
	}
}

// TestParseMap_RejectsMissingNote: every map entry's note must name its
// source (contract §2.3 anti-gaming / spec HARD RULES).
func TestParseMap_RejectsMissingNote(t *testing.T) {
	_, err := ParseMap(filepath.Join("testdata", "map_missing_note.yaml"))
	if err == nil {
		t.Fatalf("want error for a map entry with an empty note, got nil")
	}
}

func TestParseMap_MissingFileIsEmptyNotError(t *testing.T) {
	entries, err := ParseMap(filepath.Join("testdata", "does_not_exist.yaml"))
	if err != nil {
		t.Fatalf("ParseMap on missing file: %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("want 0 entries for a missing map file, got %d", len(entries))
	}
}
