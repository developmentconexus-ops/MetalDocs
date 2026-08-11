package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	openapiv1 "metaldocs/api/openapi/v1"
	platformmw "metaldocs/internal/platform/middleware"
	"metaldocs/internal/platform/openapivalidate"
)

// PR #112 round 3 cold-review (Thread 4): revokeSession
// (DELETE /api/v1/auth/sessions/{session_id}) accepts an optional
// best-effort {"reason": "..."} body — an absent, empty, or malformed body
// must never fail the revoke; internal/modules/iam/delivery/http/sessions_handler.go's
// parseRevokeSessionReason decodes it manually and swallows decode errors.
//
// A prior edit in this same round briefly gave the operation a requestBody
// schema (to close a different, unrelated gap). OpenAPI/JSON-Schema
// validation has no "validate if present, else tolerate malformed" mode:
// declaring ANY content schema for a requestBody — required or not — makes
// contract_validation reject malformed JSON or a wrong-typed field before
// the handler's own best-effort decode ever runs. That silently broke this
// route's documented contract. The fix removed the requestBody declaration
// entirely (see the operation's `description` in api/openapi/v1/openapi.yaml
// for the long-form why) and this file proves the real ingress chain — same
// openapivalidate.Wrap the binary mounts, not a bespoke stand-in — now lets
// all three shapes below reach the handler rather than 400ing upstream of it.
func buildRevokeSessionIngress(t *testing.T) (http.Handler, *specSpy) {
	t.Helper()

	validator, err := openapivalidate.New(openapiv1.SpecYAML(), openapivalidate.DefaultMaxBodyBytes)
	if err != nil {
		t.Fatalf("contract validator failed to build from the shipped spec: %v", err)
	}

	spy := &specSpy{}
	mux := http.NewServeMux()
	mux.Handle("DELETE /api/v1/auth/sessions/{session_id}", spy)

	h := buildChain(mux, apiChain(
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
		validator.Wrap,
		platformmw.MethodNotAllowedJSON,
	))
	return h, spy
}

func revokeRequest(body string) *http.Request {
	var r *http.Request
	if body == "" {
		r = httptest.NewRequest(http.MethodDelete, "/api/v1/auth/sessions/11111111-1111-4111-8111-111111111111", nil)
	} else {
		r = httptest.NewRequest(http.MethodDelete, "/api/v1/auth/sessions/11111111-1111-4111-8111-111111111111", strings.NewReader(body))
		r.Header.Set("Content-Type", "application/json")
	}
	return r
}

func TestIngress_RevokeSession_AbsentBody_ReachesHandler(t *testing.T) {
	h, spy := buildRevokeSessionIngress(t)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, revokeRequest(""))

	if !spy.called {
		t.Fatalf("an absent body must reach the handler (status %d, body %s)", rr.Code, rr.Body.String())
	}
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 from the spy", rr.Code)
	}
}

func TestIngress_RevokeSession_MalformedJSONBody_ReachesHandler(t *testing.T) {
	h, spy := buildRevokeSessionIngress(t)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, revokeRequest(`{not valid json`))

	if !spy.called {
		t.Fatalf("malformed JSON must still reach the handler, not be rejected by contract_validation "+
			"(status %d, body %s) — this is exactly the regression cold-review Thread 4 caught", rr.Code, rr.Body.String())
	}
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 from the spy", rr.Code)
	}
}

func TestIngress_RevokeSession_WrongTypedReason_ReachesHandler(t *testing.T) {
	h, spy := buildRevokeSessionIngress(t)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, revokeRequest(`{"reason":12345}`))

	if !spy.called {
		t.Fatalf("a wrong-typed reason field must still reach the handler, where parseRevokeSessionReason "+
			"swallows the decode error and falls back to \"\" (status %d, body %s)", rr.Code, rr.Body.String())
	}
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 from the spy", rr.Code)
	}
}

func TestIngress_RevokeSession_ValidReasonBody_ReachesHandler(t *testing.T) {
	// Control: a well-formed body still reaches the handler, so the passes
	// above aren't equally satisfied by a chain that never rejects anything.
	h, spy := buildRevokeSessionIngress(t)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, revokeRequest(`{"reason":"admin forced logout"}`))

	if !spy.called {
		t.Fatalf("a contract-shaped body did not reach the handler (status %d, body %s)", rr.Code, rr.Body.String())
	}
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 from the spy", rr.Code)
	}
}
