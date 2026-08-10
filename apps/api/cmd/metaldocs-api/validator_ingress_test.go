package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	openapiv1 "metaldocs/api/openapi/v1"
	platformmw "metaldocs/internal/platform/middleware"
	"metaldocs/internal/platform/openapivalidate"
)

// A3.2's proofs, taken through the composed ingress rather than around it.
//
// internal/platform/openapivalidate has its own tests, and they pin the
// classification table. They cannot pin the thing that actually matters here:
// that the validator is IN the chain production builds, in the right position,
// with the real contract behind it. A validator that works perfectly and is
// wired nowhere passes every unit test it has.
//
// So this file builds the handler through the same apiChain + buildChain that
// main.go calls, with the same openapivalidate.New(openapiv1.SpecYAML()) the
// binary constructs. The links that need live infrastructure (authn, authz,
// rate limiting, presence) are nil — buildChain skips them, exactly as it does
// on the SQLDB-less boot path — but the two links under test, contract
// validation and method_not_allowed, are the production functions.

// specSpy stands in for a business handler mounted at a real spec route. If it
// ever runs on invalid input, the API executed business logic against a request
// the contract forbids — which is the entire defect A3.2 closes.
//
// It records more than "was I reached", because reachability is only half of
// what this link must not break. Validating a request body means CONSUMING it,
// and a handler downstream of the validator still has to be able to decode the
// same bytes the client sent. So the spy also reads r.Body and keeps both the
// bytes and the read error, letting the control assert replay rather than
// assume it.
type specSpy struct {
	called  bool
	body    []byte
	readErr error
}

func (s *specSpy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.called = true
	s.body, s.readErr = io.ReadAll(r.Body)
	w.WriteHeader(http.StatusOK)
}

func buildValidatingIngress(t *testing.T) (http.Handler, *specSpy) {
	t.Helper()

	validator, err := openapivalidate.New(openapiv1.SpecYAML(), openapivalidate.DefaultMaxBodyBytes)
	if err != nil {
		t.Fatalf("contract validator failed to build from the shipped spec: %v", err)
	}
	return buildIngressWith(validator.Wrap)
}

func buildIngressWith(contractValidation func(http.Handler) http.Handler) (http.Handler, *specSpy) {
	spy := &specSpy{}
	mux := http.NewServeMux()
	// Three real spec routes: one with a request body, one with a numeric
	// bound on a query parameter, one with an enum. All three patterns are the
	// ones httpSurface carries.
	mux.Handle("POST /api/v1/auth/login", spy)
	mux.Handle("GET /api/v1/iam/users", spy)
	mux.Handle("GET /api/v1/controlled-documents", spy)

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
		contractValidation,
		platformmw.MethodNotAllowedJSON,
	))
	return h, spy
}

func postJSON(target, body string) *http.Request {
	r := httptest.NewRequest(http.MethodPost, target, strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	return r
}

func problemCode(t *testing.T, rr *httptest.ResponseRecorder) string {
	t.Helper()
	if ct := rr.Header().Get("Content-Type"); !strings.HasPrefix(ct, "application/problem+json") {
		t.Fatalf("Content-Type = %q, want application/problem+json (body: %s)", ct, rr.Body.String())
	}
	var p struct {
		Code string `json:"code"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &p); err != nil {
		t.Fatalf("rejection body is not JSON: %v (%s)", err, rr.Body.String())
	}
	return p.Code
}

// The composition assertion: contract_validation must exist as a named link,
// and it must sit after every security link and before the mux. Position is not
// cosmetic — validating before authn would let an anonymous caller probe the
// shape of bodies they may not submit.
func TestContractValidationLink_PositionedAfterSecurity(t *testing.T) {
	stub := func(http.Handler) http.Handler { return nil }
	links := apiChain(stub, stub, stub, stub, stub, stub, stub, stub, stub, stub, stub, stub)

	index := map[string]int{}
	for i, l := range links {
		index[l.name] = i
	}
	pos, ok := index["contract_validation"]
	if !ok {
		t.Fatal("apiChain has no contract_validation link: A3.2's guarantee is not composed")
	}
	for _, earlier := range []string{"authn", "iam_authz", "rate_limit"} {
		if index[earlier] >= pos {
			t.Fatalf("contract_validation (%d) must run after %s (%d)", pos, earlier, index[earlier])
		}
	}
	if pos >= index["method_not_allowed"] {
		t.Fatalf("contract_validation (%d) must run outside method_not_allowed (%d), which owns 404/405",
			pos, index["method_not_allowed"])
	}
}

func TestIngress_MissingRequiredField_RejectedBeforeHandler(t *testing.T) {
	h, spy := buildValidatingIngress(t)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, postJSON("/api/v1/auth/login", `{"identifier":"someone@example.com"}`))

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400 (body: %s)", rr.Code, rr.Body.String())
	}
	if spy.called {
		t.Fatal("the handler ran without the contract-required password field")
	}
	if code := problemCode(t, rr); code != "request.invalid" {
		t.Fatalf("code = %q, want request.invalid", code)
	}
}

func TestIngress_NumericBound_RejectedBeforeHandler(t *testing.T) {
	h, spy := buildValidatingIngress(t)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/api/v1/iam/users?limit=999", nil))

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400 (body: %s)", rr.Code, rr.Body.String())
	}
	if spy.called {
		t.Fatal("the handler ran with limit=999, which the contract caps at 100")
	}
	if code := problemCode(t, rr); code != "request.invalid" {
		t.Fatalf("code = %q, want request.invalid", code)
	}
}

func TestIngress_InvalidEnum_RejectedBeforeHandler(t *testing.T) {
	h, spy := buildValidatingIngress(t)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/api/v1/controlled-documents?status=bogus", nil))

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400 (body: %s)", rr.Code, rr.Body.String())
	}
	if spy.called {
		t.Fatal("the handler ran with a status outside the contract's enum")
	}
	if code := problemCode(t, rr); code != "request.invalid" {
		t.Fatalf("code = %q, want request.invalid", code)
	}
}

// The recorded RED state, kept permanently rather than run once and deleted.
//
// Compose the identical ingress with the contract_validation link removed —
// which is what this repo shipped before A3.2 — and the request every test
// above proves is rejected lands in business code with a 200. This is the
// counterfactual that makes the other assertions mean something: without it,
// "the handler was not called" could be true for any number of unrelated
// reasons.
func TestIngress_WithoutTheValidatorLink_InvalidRequestReachesHandler(t *testing.T) {
	h, spy := buildIngressWith(nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, postJSON("/api/v1/auth/login", `{"identifier":"someone@example.com"}`))

	if !spy.called {
		t.Fatal("expected the pre-A3.2 chain to hand invalid input to the handler; " +
			"if this now fails, something else is rejecting the request and the other proofs are measuring it")
	}
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 from the unguarded chain", rr.Code)
	}
}

func TestIngress_MalformedJSON_RejectedBeforeHandler(t *testing.T) {
	h, spy := buildValidatingIngress(t)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, postJSON("/api/v1/auth/login", `{"identifier": "a@b.com",`))

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400 (body: %s)", rr.Code, rr.Body.String())
	}
	if spy.called {
		t.Fatal("the handler ran on a body that is not JSON")
	}
	if code := problemCode(t, rr); code != "request.json_decode" {
		t.Fatalf("code = %q, want request.json_decode", code)
	}
}

// The control. Without it, every assertion above is satisfiable by a chain that
// rejects everything.
//
// It also proves the property the rejection tests structurally cannot: that the
// body survives validation intact. openapi3filter has to READ the body to check
// it against the schema, which consumes the stream; every body-decoding handler
// in this API is downstream of that read. If the validator returned a body that
// was empty, truncated, or re-serialized, no status assertion anywhere would
// notice — the request would be accepted and then silently decode into the
// wrong thing.
//
// The submitted bytes are deliberately NOT canonical JSON: they carry newlines,
// indentation and a space before a colon that any marshal/unmarshal round trip
// would normalize away. Comparing byte-for-byte therefore proves replay of the
// original stream, not merely that something semantically equivalent arrived.
// A weaker assertion (valid JSON, or equal after re-marshalling) would pass on
// a validator that rebuilt the body, which is exactly the failure mode worth
// catching.
func TestIngress_ValidRequest_ReachesHandlerWithTheExactSubmittedBody(t *testing.T) {
	h, spy := buildValidatingIngress(t)
	submitted := "{\n  \"identifier\" : \"someone@example.com\",\n" +
		"  \"password\": \"correct-horse-battery\"\n}"

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, postJSON("/api/v1/auth/login", submitted))

	if !spy.called {
		t.Fatalf("a contract-valid request did not reach the handler (status %d, body %s)",
			rr.Code, rr.Body.String())
	}
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rr.Code)
	}
	if spy.readErr != nil {
		t.Fatalf("the handler could not read r.Body after the validator consumed it: %v", spy.readErr)
	}
	if !bytes.Equal(spy.body, []byte(submitted)) {
		t.Fatalf("the handler received a body the client never sent\nsubmitted (%d bytes): %q\nreceived  (%d bytes): %q",
			len(submitted), submitted, len(spy.body), string(spy.body))
	}
}

// method_not_allowed still owns 405 with its Allow header: the validator sits
// outside it and must not have taken that over.
func TestIngress_MethodNotAllowed_StillAnsweredByTheMuxLayer(t *testing.T) {
	h, spy := buildValidatingIngress(t)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodDelete, "/api/v1/iam/users", nil))

	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want 405 (body: %s)", rr.Code, rr.Body.String())
	}
	if spy.called {
		t.Fatal("handler ran for a method it is not mounted for")
	}
	if allow := rr.Header().Get("Allow"); allow == "" {
		t.Fatal("405 lost its Allow header")
	}
}
