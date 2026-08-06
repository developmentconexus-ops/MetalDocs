package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// §10.3 row 4: on the SQLDB-less boot path the presence stream used to be
// absent from the mux entirely (404). Under §4's Mount-is-total rule it is
// mounted and answers 501, matching its snapshot sibling — which is the
// repo's existing majority convention (eight IAM operations already do this).
//
// This is the degraded dev path only; no production boot has SQLDB == nil.
func TestPresenceStreamAnswers501WhenUnavailable(t *testing.T) {
	mux := http.NewServeMux()
	mountRoutesWithNilPresence(t, mux) // helper: builds the router with presence == nil

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/api/v1/iam/presence/stream", nil))

	if rr.Code != http.StatusNotImplemented {
		t.Fatalf("status = %d, want 501 (was 404 before Mount-is-total)", rr.Code)
	}
	if ct := rr.Header().Get("Content-Type"); !strings.HasPrefix(ct, "application/problem+json") {
		t.Fatalf("Content-Type = %q, want problem+json", ct)
	}
}

// mountRoutesWithNilPresence builds a real routeHandlers instance exactly the
// way buildTestRouteHandlers does (same constructors buildRouter's family
// registration binds against), then nils out presence and mounts through
// buildRouter — the same function main() calls. This is the production mount
// path with only h.presence forced nil, mirroring the one boot condition
// (deps.SQLDB == nil) that used to make buildPresence return nil.
func mountRoutesWithNilPresence(t *testing.T, mux *http.ServeMux) {
	t.Helper()
	h := buildTestRouteHandlers()
	h.presence = nil
	buildRouter(mux, h)
}
