package approvalhttp

import "testing"

func TestParseIfMatch(t *testing.T) {
	t.Run("wildcard", func(t *testing.T) {
		version, err := parseIfMatch("*")
		if err != nil {
			t.Fatalf("parseIfMatch(*): %v", err)
		}
		if version != 0 {
			t.Fatalf("version = %d, want 0", version)
		}
	})

	t.Run("positive version", func(t *testing.T) {
		version, err := parseIfMatch(`"v7"`)
		if err != nil {
			t.Fatalf("parseIfMatch(v7): %v", err)
		}
		if version != 7 {
			t.Fatalf("version = %d, want 7", version)
		}
	})

	// Governed transitions (signoff/cancel/decision) act on an already-submitted
	// document at revision_version >= 1, so an explicit "v0" precondition is stale
	// and MUST be rejected by the shared parser.
	t.Run("zero version rejected for governed transitions", func(t *testing.T) {
		version, err := parseIfMatch(`"v0"`)
		if err == nil {
			t.Fatalf("expected malformed If-Match error, got version %d", version)
		}
		if err != ErrIfMatchMalformed {
			t.Fatalf("err = %v, want %v", err, ErrIfMatchMalformed)
		}
	})
}

// TestParseSubmitIfMatch pins the ADR 0073 submit-only relaxation: the canonical
// /submit handler targets a fresh draft at revision_version = 0, so "v0" MUST be
// accepted here even though parseIfMatch rejects it for governed transitions.
func TestParseSubmitIfMatch(t *testing.T) {
	t.Run("zero version accepted (fresh draft)", func(t *testing.T) {
		version, err := parseSubmitIfMatch(`"v0"`)
		if err != nil {
			t.Fatalf("parseSubmitIfMatch(v0): %v", err)
		}
		if version != 0 {
			t.Fatalf("version = %d, want 0", version)
		}
	})

	t.Run("positive version accepted", func(t *testing.T) {
		version, err := parseSubmitIfMatch(`"v3"`)
		if err != nil {
			t.Fatalf("parseSubmitIfMatch(v3): %v", err)
		}
		if version != 3 {
			t.Fatalf("version = %d, want 3", version)
		}
	})

	t.Run("negative version rejected", func(t *testing.T) {
		if _, err := parseSubmitIfMatch(`"v-1"`); err != ErrIfMatchMalformed {
			t.Fatalf("err = %v, want %v", err, ErrIfMatchMalformed)
		}
	})

	t.Run("missing header rejected", func(t *testing.T) {
		if _, err := parseSubmitIfMatch(""); err != ErrIfMatchRequired {
			t.Fatalf("err = %v, want %v", err, ErrIfMatchRequired)
		}
	})
}
