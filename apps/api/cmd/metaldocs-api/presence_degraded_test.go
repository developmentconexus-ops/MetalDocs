package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	iamdelivery "metaldocs/internal/modules/iam/delivery/http"
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

// mountRoutesWithNilPresence builds a real publisherDeps instance exactly the
// way buildTestRouteHandlers does (same constructors buildPublishers'
// SurfacePublisher list binds against), then rebuilds iam with a nil presence
// dependency and mounts every publisher — the same construction main() runs.
// This is the production mount path with only the presence dependency forced
// nil, mirroring the one boot condition (deps.SQLDB == nil) that used to make
// buildPresence return nil.
//
// Task 15a (Ruling 2) folded presence's stream mount into iamRouter.Mount, so
// presence is no longer a publisherDeps field of its own — the nil injection
// point moved to iamRouter's own presence dependency (unexported, so it's
// re-supplied via NewRouter rather than a field assignment on d).
func mountRoutesWithNilPresence(t *testing.T, mux *http.ServeMux) {
	t.Helper()
	d := buildTestRouteHandlers()
	d.iam = iamdelivery.NewRouter(
		iamdelivery.NewAdminHandler(nil, nil),
		iamdelivery.NewPeopleHandler(nil, nil, nil),
		iamdelivery.NewMembershipHandler(nil, nil),
		iamdelivery.NewRolesCapsHandler(nil),
		iamdelivery.NewSessionsHandler(nil),
		iamdelivery.NewObservabilityHandler(nil),
		nil, // presence: forced nil, mirroring the SQLDB-less boot path
	)
	for _, p := range buildPublishers(d) {
		p.Mount(mux)
	}
}
