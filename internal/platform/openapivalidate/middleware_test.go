package openapivalidate_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	openapiv1 "metaldocs/api/openapi/v1"
	"metaldocs/internal/platform/openapivalidate"
)

// These tests exercise the middleware against the REAL contract — the same
// bytes the binary embeds — rather than a toy spec. A toy spec would prove the
// library works, which nobody doubted; what is worth proving is that MetalDocs'
// own 7700-line contract loads, that its relative server URL ("/api/v1") is
// matched against real request paths, and that its declared constraints fire.
//
// The end-to-end proof that this runs inside the composed production chain
// lives in apps/api/cmd/metaldocs-api (validator_ingress_test.go). Both exist
// on purpose: this file pins the classification table, that one pins the
// composition.

func newMiddleware(t *testing.T) *openapivalidate.Middleware {
	t.Helper()
	m, err := openapivalidate.New(openapiv1.SpecYAML(), 0)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return m
}

// handlerSpy records whether the business handler ran. "Rejected before the
// handler" is the whole property; asserting only the status code would pass
// even if the handler had already written to the database.
type handlerSpy struct{ called bool }

func (h *handlerSpy) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	h.called = true
	w.WriteHeader(http.StatusOK)
}

func do(t *testing.T, m *openapivalidate.Middleware, req *http.Request) (*httptest.ResponseRecorder, *handlerSpy) {
	t.Helper()
	spy := &handlerSpy{}
	rec := httptest.NewRecorder()
	m.Wrap(spy).ServeHTTP(rec, req)
	return rec, spy
}

func jsonReq(method, path, body string) *http.Request {
	r := httptest.NewRequest(method, path, strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	return r
}

func decodeProblem(t *testing.T, rec *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	if ct := rec.Header().Get("Content-Type"); ct != "application/problem+json" {
		t.Fatalf("Content-Type = %q, want application/problem+json", ct)
	}
	var out map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("body is not JSON: %v (%s)", err, rec.Body.String())
	}
	return out
}

func TestSpecLoads(t *testing.T) {
	if _, err := openapivalidate.New(openapiv1.SpecYAML(), 0); err != nil {
		t.Fatalf("the shipped contract does not load or does not validate: %v", err)
	}
}

func TestNew_EmptySpecFailsClosed(t *testing.T) {
	if _, err := openapivalidate.New(nil, 0); err == nil {
		t.Fatal("New(nil) returned no error: a missing contract must be a boot failure, not silent pass-through")
	}
}

// POST /api/v1/auth/login declares a body with required "identifier" and "password".
func TestMissingRequiredField_RejectedBeforeHandler(t *testing.T) {
	m := newMiddleware(t)
	rec, spy := do(t, m, jsonReq(http.MethodPost, "/api/v1/auth/login", `{"identifier":"a@b.com"}`))

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400 (body: %s)", rec.Code, rec.Body.String())
	}
	if spy.called {
		t.Fatal("the business handler ran on a request the contract forbids")
	}
	p := decodeProblem(t, rec)
	if p["code"] != "request.invalid" {
		t.Fatalf("code = %v, want request.invalid", p["code"])
	}
}

func TestValidRequest_ReachesHandler(t *testing.T) {
	m := newMiddleware(t)
	rec, spy := do(t, m,
		jsonReq(http.MethodPost, "/api/v1/auth/login", `{"identifier":"a@b.com","password":"correct-horse"}`))

	if !spy.called {
		t.Fatalf("valid request did not reach the handler (status %d, body %s)", rec.Code, rec.Body.String())
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
}

func TestMalformedJSON_Rejected(t *testing.T) {
	m := newMiddleware(t)
	rec, spy := do(t, m, jsonReq(http.MethodPost, "/api/v1/auth/login", `{"identifier":`))

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400 (body: %s)", rec.Code, rec.Body.String())
	}
	if spy.called {
		t.Fatal("handler ran on malformed JSON")
	}
	if p := decodeProblem(t, rec); p["code"] != "request.json_decode" {
		t.Fatalf("code = %v, want request.json_decode", p["code"])
	}
}

// GET /controlled-documents declares status as enum [active, obsolete,
// superseded]. Nothing in the handler chain checked that before this link.
func TestInvalidEnum_RejectedBeforeHandler(t *testing.T) {
	m := newMiddleware(t)
	rec, spy := do(t, m,
		httptest.NewRequest(http.MethodGet, "/api/v1/controlled-documents?status=not-a-status", nil))

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400 (body: %s)", rec.Code, rec.Body.String())
	}
	if spy.called {
		t.Fatal("handler ran on an out-of-enum query value")
	}
	p := decodeProblem(t, rec)
	if p["code"] != "request.invalid" {
		t.Fatalf("code = %v, want request.invalid", p["code"])
	}
	if title, _ := p["title"].(string); !strings.Contains(title, "status") {
		t.Fatalf("title %q does not name the offending parameter", title)
	}
}

// GET /iam/users declares limit as integer with maximum 100. This is the
// numeric-bound half of the proof: a value that parses fine and is still
// outside the contract.
func TestNumericBound_RejectedBeforeHandler(t *testing.T) {
	m := newMiddleware(t)
	rec, spy := do(t, m, httptest.NewRequest(http.MethodGet, "/api/v1/iam/users?limit=999", nil))

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400 (body: %s)", rec.Code, rec.Body.String())
	}
	if spy.called {
		t.Fatal("handler ran with limit=999, which the contract caps at 100")
	}
	if p := decodeProblem(t, rec); p["code"] != "request.invalid" {
		t.Fatalf("code = %v, want request.invalid", p["code"])
	}
}

// The control for both of the above: the same routes with in-contract values
// must reach the handler untouched.
func TestInContractQueryValues_ReachHandler(t *testing.T) {
	m := newMiddleware(t)
	for _, target := range []string{
		"/api/v1/controlled-documents?status=active",
		"/api/v1/iam/users?limit=100",
	} {
		rec, spy := do(t, m, httptest.NewRequest(http.MethodGet, target, nil))
		if !spy.called {
			t.Fatalf("%s did not reach the handler (status %d, body %s)", target, rec.Code, rec.Body.String())
		}
	}
}

func TestUndeclaredContentType_Rejected(t *testing.T) {
	m := newMiddleware(t)
	r := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader("identifier=a@b.com"))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec, spy := do(t, m, r)

	if rec.Code != http.StatusUnsupportedMediaType {
		t.Fatalf("status = %d, want 415 (body: %s)", rec.Code, rec.Body.String())
	}
	if spy.called {
		t.Fatal("handler ran on an undeclared content type")
	}
}

func TestOversizedBody_Rejected(t *testing.T) {
	m, err := openapivalidate.New(openapiv1.SpecYAML(), 64)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	spy := &handlerSpy{}
	rec := httptest.NewRecorder()
	body := `{"identifier":"a@b.com","password":"` + strings.Repeat("x", 512) + `"}`
	m.Wrap(spy).ServeHTTP(rec, jsonReq(http.MethodPost, "/api/v1/auth/login", body))

	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, want 413 (body: %s)", rec.Code, rec.Body.String())
	}
	if spy.called {
		t.Fatal("handler ran on an oversized body")
	}
}

// A route the contract does not describe belongs to the mux (which owns 404 and
// 405), so it must pass through rather than be answered here.
func TestUnknownPath_PassesThroughToMux(t *testing.T) {
	m := newMiddleware(t)
	rec, spy := do(t, m, httptest.NewRequest(http.MethodGet, "/healthz", nil))

	if !spy.called {
		t.Fatalf("unknown path did not pass through (status %d) — the validator must not own 404", rec.Code)
	}
}

// The validator must not leak the payload that failed. A password that misses
// minLength is still a password.
func TestRejection_NeverEchoesTheSubmittedValue(t *testing.T) {
	m := newMiddleware(t)
	const secret = "hunter2-do-not-echo"
	rec, _ := do(t, m,
		jsonReq(http.MethodPost, "/api/v1/auth/login", `{"identifier":`+`123`+`,"password":"`+secret+`"}`))

	if rec.Code == http.StatusOK {
		t.Fatal("expected a rejection")
	}
	if strings.Contains(rec.Body.String(), secret) {
		t.Fatalf("the rejection echoed the submitted password: %s", rec.Body.String())
	}
}
