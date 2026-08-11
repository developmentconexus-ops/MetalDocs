package http

import (
	"net/http"
	"testing"

	"metaldocs/internal/platform/authn"
	"metaldocs/internal/platform/problem"
)

// TestMapErrMissingActor is the RED-first proof for the review-round-1 gap:
// mapErr had no case for authn.ErrMissingActor and fell through to
// `default: 500 internal.unknown`. approval/http (errors.go:251) and
// templates/delivery/http (errors.go:123) both already carry
// `{matchIs/isErr(authn.ErrMissingActor), http.StatusUnauthorized,
// .../CodeAuthUnauthenticated}`; documents/delivery/http was the odd mapper
// out even though the sentinel is the identical platform condition — no
// authenticated principal in context. That inconsistency across the six
// per-module sentinel->401 mappers is itself the hand-synced-enumeration
// defect class this repo tracks.
//
// The sentinel is dormant today: no live caller feeds authn.ErrMissingActor
// into this mapErr today (userIDFromReq's callers stop at the RequireUserID
// boundary and never delegate the raw error into the generic mapErr switch).
// The gap is real regardless — any future caller that does route the
// sentinel through mapErr would silently get 500 instead of 401 — so this
// test pins the mapping directly rather than through a live handler path.
func TestMapErrMissingActor(t *testing.T) {
	status, code := mapErr(authn.ErrMissingActor)

	if status != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d (authn.ErrMissingActor must map to 401, not fall through to the default 500)", status, http.StatusUnauthorized)
	}
	if code != problem.CodeAuthUnauthenticated {
		t.Fatalf("code = %q, want %q", code, problem.CodeAuthUnauthenticated)
	}
}
