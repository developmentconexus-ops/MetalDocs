//go:build integration

package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"metaldocs/tests/integration/testdb"
)

// §10.3 row 5. Before this program routeRule.matches demanded exact method
// equality and no row carried HEAD, so ServeMux's GET-also-matches-HEAD rule
// made a capability-guarded GET reachable by HEAD with only a session — status
// and headers leaked. mux.Handler(r) returns the GET pattern for a HEAD
// request, so HEAD now inherits the GET rule's capability structurally.
//
// The claim under test is generic — HEAD resolves to the SAME capability as
// GET — not specific to any one route. The carrier is picked by
// pickSubsetCarrierRoute (see its doc comment): a route gated on a capability
// every assignable role holds (e.g. document.view on documents/{id}) cannot
// express the negative half of this test at all, since no real principal
// lacks it — that would make the claim look unconstructible when only the
// carrier was badly chosen.
func TestHEADInheritsGETCapability(t *testing.T) {
	db := testdb.New(t)
	tenant := testdb.NewTenant(t, db)

	pattern, path := pickSubsetCarrierRoute(t, db)
	grantingRoles := rolesGranting(t, db, string(httpSurface[pattern].capability))
	nonGranting := firstRoleNotIn(assignableRoleNames, grantingRoles)
	// "noCaps" relative to the carrier's own capability — not a truly
	// role-less user (see TestSurfaceConformance's minimalUser comment: a
	// role-less principal cannot complete a real login at all).
	noCaps := testdb.NewUser(t, db, testdb.WithTenantID(tenant.ID), testdb.WithRole(nonGranting))

	var reached bool
	sentinel := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reached = true
		w.WriteHeader(http.StatusOK)
	})
	srv := chainOverSentinel(t, db, pattern, sentinel)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodHead, fillPathParams(path), nil)
	srv.ServeHTTP(rr, withSession(t, req, noCaps))

	if reached {
		t.Fatalf("HEAD bypassed tier-1: the sentinel ran for a principal (role %q) holding not %s's capability", nonGranting, pattern)
	}
}

// §10.3 row 7. The old resolver matched decoded substrings of the WHOLE path,
// so a parameter VALUE could select a rule. The new table is keyed by the
// route, so a parameter value is never part of its key. Each case asserts the
// new behavior is correct, not that it matches the old one.
//
// The two parameter-steering cases are re-homed onto the same computed
// carrier as TestHEADInheritsGETCapability, for the identical reason: the
// claim (resolved policy is invariant under parameter CONTENT) is generic,
// but documents/{id} is gated on a floor capability every assignable role
// holds, so no real fixture can express "this principal must be refused"
// against it — that made the carrier, not the claim, unconstructible.
func TestParameterContentCannotSteerThePolicy(t *testing.T) {
	db := testdb.New(t)
	tenant := testdb.NewTenant(t, db)
	noCaps := testdb.NewUser(t, db, testdb.WithTenantID(tenant.ID), testdb.WithRole("viewer"))

	carrierPattern, carrierPath := pickSubsetCarrierRoute(t, db)
	carrierMethod, _ := splitPattern(t, carrierPattern)
	carrierGranting := rolesGranting(t, db, string(httpSurface[carrierPattern].capability))
	carrierNonGranting := firstRoleNotIn(assignableRoleNames, carrierGranting)
	carrierNoCaps := testdb.NewUser(t, db, testdb.WithTenantID(tenant.ID), testdb.WithRole(carrierNonGranting))

	cases := []struct {
		name    string
		pattern string
		method  string
		target  string
		fixture testdb.User
	}{
		// user_id="roles" hit notSuffix:"/roles" and fell through to the
		// session-required fallback under the old table.
		{"user id spelled roles", "PATCH /api/v1/iam/users/{user_id}", http.MethodPatch, "/api/v1/iam/users/roles", noCaps},
		// A parameter value that looks like it could be mistaken for a
		// sibling route segment/tag must still resolve via the matched
		// PATTERN, never the string content of the value.
		{"param spelled approval-preview", carrierPattern, carrierMethod, fillFirstPathParam(carrierPath, "approval-preview"), carrierNoCaps},
		// %2F kept the segment single for the mux while an old decoded-match
		// resolver could see a suffix rule the router never routed to.
		{"encoded slash in param", carrierPattern, carrierMethod, fillFirstPathParam(carrierPath, "foo%2Fapproval-preview"), carrierNoCaps},
		// The §9 edge table: trailing slash, doubled slash, and dot-dot are
		// resolved by the REAL mux before any rule is consulted — these stay
		// on documents/{id} because the property they prove is entirely
		// mux-level and independent of which capability guards the route.
		{"trailing slash", "GET /api/v1/documents/{id}", http.MethodGet, "/api/v1/documents/x/", noCaps},
		{"doubled slash", "GET /api/v1/documents/{id}", http.MethodGet, "/api/v1//documents/x", noCaps},
		{"dot dot", "GET /api/v1/documents/{id}", http.MethodGet, "/api/v1/documents/../documents/x", noCaps},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			var reached bool
			sentinel := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				reached = true
				w.WriteHeader(http.StatusOK)
			})
			srv := chainOverSentinel(t, db, tc.pattern, sentinel)

			rr := httptest.NewRecorder()
			srv.ServeHTTP(rr, withSession(t, httptest.NewRequest(tc.method, tc.target, nil), tc.fixture))

			// Either the mux never routed this target to the guarded pattern
			// (the sentinel is the only thing behind it, so not reaching it is
			// a 404) or tier-1 refused it. Both are correct; what must NEVER
			// happen is the sentinel running for a principal holding nothing.
			if reached {
				t.Fatalf("parameter content steered the policy: the sentinel ran for %s %s", tc.method, tc.target)
			}
		})
	}
}
