package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	platformmw "metaldocs/internal/platform/middleware"
)

// TestPanicRecoveryYieldsProblemJSON verifies REQ-MW-1 (a handler panic is
// recovered into a 500 problem+json response, not a killed process) directly
// over the same chain builder main() uses (apiChain/buildChain, chain.go:25),
// with panic_recovery as the one real, non-nil link — mirroring
// health_delta_test.go's pattern for driving a real middleware link in
// isolation.
//
// This replaces the deleted GET /internal/test/panic runtime probe
// (main.go, formerly mounted only when METALDOCS_E2E=1): that probe had
// exactly one occurrence in the repo — its own registration — and no
// workflow, script, or Playwright flow ever requested it, so it never
// verified REQ-MW-1. This test drives the real invariant directly and can
// actually fail.
func TestPanicRecoveryYieldsProblemJSON(t *testing.T) {
	inner := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("REQ-MW-1 probe")
	})

	h := buildChain(inner, apiChain(
		platformmw.Recovery,
		nil, // otel
		nil, // http_obs
		nil, // cors
		nil, // origin_protection
		nil, // pre_auth_login_rate_limit
		nil, // authn
		nil, // iam_authz
		nil, // presence_bump
		nil, // rate_limit
		nil, // contract_validation
		nil, // method_not_allowed
	))

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/api/v1/anything", nil))

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", rr.Code)
	}
	if ct := rr.Header().Get("Content-Type"); !strings.HasPrefix(ct, "application/problem+json") {
		t.Fatalf("Content-Type = %q, want problem+json", ct)
	}
}
