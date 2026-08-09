package approvalhttp

import (
	"errors"
	"testing"
)

func TestParseIfMatch(t *testing.T) {
	// The "*" wildcard is NOT a valid precondition on a governed transition. It
	// was previously encoded as version 0 and fed straight into the OCC CAS, so
	// it silently meant "expect v0" and 409'd on any real document; the contract
	// no longer offers it and the parser rejects it as malformed.
	t.Run("wildcard rejected", func(t *testing.T) {
		version, err := parseIfMatch("*")
		if !errors.Is(err, ErrIfMatchMalformed) {
			t.Fatalf("err = %v (version %d), want %v", err, version, ErrIfMatchMalformed)
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
		if !errors.Is(err, ErrIfMatchMalformed) {
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

	// Submit is the one caller whose minVersion is 0, which is exactly why the
	// old wildcard-as-0 sentinel was unsound here: "*" and "v0" collapsed to the
	// same parsed value, so the handler could not tell "I don't care about the
	// version" from "I require version 0". Rejecting "*" keeps v0 unambiguous.
	t.Run("wildcard rejected (not confusable with v0)", func(t *testing.T) {
		version, err := parseSubmitIfMatch("*")
		if !errors.Is(err, ErrIfMatchMalformed) {
			t.Fatalf("err = %v (version %d), want %v", err, version, ErrIfMatchMalformed)
		}
	})

	t.Run("negative version rejected", func(t *testing.T) {
		if _, err := parseSubmitIfMatch(`"v-1"`); !errors.Is(err, ErrIfMatchMalformed) {
			t.Fatalf("err = %v, want %v", err, ErrIfMatchMalformed)
		}
	})

	t.Run("missing header rejected", func(t *testing.T) {
		if _, err := parseSubmitIfMatch(""); !errors.Is(err, ErrIfMatchRequired) {
			t.Fatalf("err = %v, want %v", err, ErrIfMatchRequired)
		}
	})
}
