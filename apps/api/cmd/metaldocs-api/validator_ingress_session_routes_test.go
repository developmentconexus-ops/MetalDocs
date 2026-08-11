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

// A3.4 cold-review Finding 2, taken through the composed ingress, same
// discipline as validator_ingress_test.go: a claim about what the contract
// now requires is only proven by running the real openapivalidate.Wrap in
// the real apiChain position, not by asserting against the bare mux the
// module's own handler_test.go uses (newMux there never mounts
// contract_validation — see internal/modules/documents/delivery/http/handler_test.go).
//
// A3.4 declared `session_id` required on DocumentSessionIdRequest (used by
// heartbeat/release/force-release) and `value` required on
// PutPlaceholderValueRequest. Before that, the handler read the field
// optimistically: an absent/empty session_id reached the application layer
// and then a Postgres `uuid`-typed column (editor_sessions.id), which
// rejects an empty string at the driver/server level — not the friendly
// domain error `mapErr`'s dispatch table knows how to translate, so it fell
// through to the `default:` case (500 `internal.unknown`). Declaring the
// field required moves that rejection into contract_validation, before the
// handler runs: 500 becomes 400 `request.invalid`. This file proves the new
// 400-before-handler behaviour for all four routes, plus the control case
// (a present field still reaches the handler) so the rejection assertions
// cannot be satisfied by a chain that rejects everything.

func buildSessionRoutesIngress(t *testing.T) (http.Handler, *specSpy) {
	t.Helper()

	validator, err := openapivalidate.New(openapiv1.SpecYAML(), openapivalidate.DefaultMaxBodyBytes)
	if err != nil {
		t.Fatalf("contract validator failed to build from the shipped spec: %v", err)
	}

	spy := &specSpy{}
	mux := http.NewServeMux()
	mux.Handle("POST /api/v1/documents/{id}/session/heartbeat", spy)
	mux.Handle("POST /api/v1/documents/{id}/session/release", spy)
	mux.Handle("POST /api/v1/documents/{id}/session/force-release", spy)
	mux.Handle("PUT /api/v1/documents/{id}/placeholders/{pid}", spy)

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

const validSessionID = "5f0f4b0a-8a1e-4e0a-9c3d-1e2f3a4b5c6d"

func TestIngress_SessionRoutes_MissingSessionID_RejectedBeforeHandler(t *testing.T) {
	for _, route := range []string{
		"/api/v1/documents/doc-1/session/heartbeat",
		"/api/v1/documents/doc-1/session/release",
		"/api/v1/documents/doc-1/session/force-release",
	} {
		t.Run(route, func(t *testing.T) {
			h, spy := buildSessionRoutesIngress(t)
			rr := httptest.NewRecorder()
			// The contract-shaped request with the required field absent —
			// exactly what the pre-A3.4 handler code accepted and forwarded.
			h.ServeHTTP(rr, postJSON(route, `{}`))

			if rr.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want 400 (body: %s)", rr.Code, rr.Body.String())
			}
			if spy.called {
				t.Fatal("the handler ran without the contract-required session_id field")
			}
			if code := problemCode(t, rr); code != "request.invalid" {
				t.Fatalf("code = %q, want request.invalid", code)
			}
		})
	}
}

func TestIngress_SessionRoutes_EmptyStringSessionID_StillReachesHandler_KnownGap(t *testing.T) {
	// This is NOT the fix — it documents what A3.4 did not close. `required:
	// [session_id]` on DocumentSessionIdRequest only demands the key be
	// present; `session_id` is `type: string` with no `minLength`, so an
	// explicit "" still satisfies the schema and reaches the handler exactly
	// as it did before this PR. The Postgres uuid-cast 500 this finding's
	// analysis describes is therefore only fixed for the ABSENT-key case
	// (see the RejectedBeforeHandler test above), not for an explicit "".
	// Closing this too would mean adding `minLength: 1` to session_id in
	// openapi.yaml — a real, scoped follow-up, deliberately left out of this
	// fix round rather than folded in silently (CLAUDE.md: keep changes
	// scoped to the request).
	for _, route := range []string{
		"/api/v1/documents/doc-1/session/heartbeat",
		"/api/v1/documents/doc-1/session/release",
		"/api/v1/documents/doc-1/session/force-release",
	} {
		t.Run(route, func(t *testing.T) {
			h, spy := buildSessionRoutesIngress(t)
			rr := httptest.NewRecorder()
			h.ServeHTTP(rr, postJSON(route, `{"session_id":""}`))

			if rr.Code != http.StatusOK {
				t.Fatalf("status = %d, want 200 (contract_validation currently accepts an empty session_id string) (body: %s)", rr.Code, rr.Body.String())
			}
			if !spy.called {
				t.Fatal("expected the empty-string session_id to still reach the handler, matching pre-A3.4 behaviour; " +
					"if this now fails, someone closed the minLength gap and this test should flip to a rejection assertion")
			}
		})
	}
}

func TestIngress_SessionRoutes_ValidSessionID_ReachesHandler(t *testing.T) {
	// The control: a contract-valid body still reaches the handler. Without
	// this, the rejection tests above would be equally satisfied by a chain
	// that rejects every request on these routes.
	for _, route := range []string{
		"/api/v1/documents/doc-1/session/heartbeat",
		"/api/v1/documents/doc-1/session/release",
		"/api/v1/documents/doc-1/session/force-release",
	} {
		t.Run(route, func(t *testing.T) {
			h, spy := buildSessionRoutesIngress(t)
			rr := httptest.NewRecorder()
			h.ServeHTTP(rr, postJSON(route, `{"session_id":"`+validSessionID+`"}`))

			if !spy.called {
				t.Fatalf("a contract-valid session_id did not reach the handler (status %d, body %s)",
					rr.Code, rr.Body.String())
			}
			if rr.Code != http.StatusOK {
				t.Fatalf("status = %d, want 200 from the spy", rr.Code)
			}
		})
	}
}

func TestIngress_PutPlaceholderValue_MissingValue_RejectedBeforeHandler(t *testing.T) {
	h, spy := buildSessionRoutesIngress(t)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/documents/doc-1/placeholders/ph-1", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400 (body: %s)", rr.Code, rr.Body.String())
	}
	if spy.called {
		t.Fatal("the handler ran without the contract-required value field")
	}
	if code := problemCode(t, rr); code != "request.invalid" {
		t.Fatalf("code = %q, want request.invalid", code)
	}
}

func TestIngress_PutPlaceholderValue_ValuePresent_ReachesHandler(t *testing.T) {
	h, spy := buildSessionRoutesIngress(t)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/documents/doc-1/placeholders/ph-1", strings.NewReader(`{"value":"anything"}`))
	req.Header.Set("Content-Type", "application/json")
	h.ServeHTTP(rr, req)

	if !spy.called {
		t.Fatalf("a contract-valid value did not reach the handler (status %d, body %s)",
			rr.Code, rr.Body.String())
	}
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 from the spy", rr.Code)
	}
}

// The recorded RED state for this finding: compose the identical ingress
// with the validator link removed, which is the chain shape mapErr's
// default case had to cope with before A3.4's spec tightening. The `{}`
// body reaches the handler instead of being rejected — proof that today's
// rejections above are caused by the contract requiring session_id, not by
// some unrelated always-reject failure mode.
func TestIngress_SessionRoutes_WithoutTheValidatorLink_EmptyBodyReachesHandler(t *testing.T) {
	spy := &specSpy{}
	mux := http.NewServeMux()
	mux.Handle("POST /api/v1/documents/{id}/session/heartbeat", spy)

	h := buildChain(mux, apiChain(
		platformmw.Recovery,
		nil, nil, nil, nil, nil, nil, nil, nil, nil,
		nil, // contract_validation removed
		platformmw.MethodNotAllowedJSON,
	))

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, postJSON("/api/v1/documents/doc-1/session/heartbeat", `{}`))

	if !spy.called {
		t.Fatal("expected the unguarded chain to hand the empty body to the handler; " +
			"if this now fails, something else is rejecting the request and the rejection proofs above are measuring it")
	}
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 from the unguarded chain", rr.Code)
	}
}
