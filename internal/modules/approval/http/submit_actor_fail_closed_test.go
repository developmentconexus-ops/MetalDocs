package approvalhttp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/platform/problem"
	"metaldocs/internal/platform/tenant"
)

// A3.3 behavioral proof for the approval MUTATION path.
//
// SubmitHandler builds application.SubmitRequest{SubmittedBy: actorID}: the
// actor is not a filter or a log field here, it is the accountability record
// for the entire approval instance — who put this document into review.
//
// The whole slice is about one substitution. BEFORE, the handler read the
// low-level accessor, which answers "" for a request carrying no principal;
// that "" went into SubmittedBy and the submit service opened a real approval
// instance attributed to nobody. AFTER, absence terminates the request before
// the service is reached.
//
// The tests below prove both halves: the counterfactual (what the old read
// would still return today, so the RED is a fact about this code rather than a
// story about a deleted line) and the live behavior (401, canonical problem
// body, service untouched).

const submitBody = `{"route_id":"11111111-1111-1111-1111-111111111111","content_hash":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}`

// newSubmitRequestWithAuth builds a well-formed submit request — valid
// Idempotency-Key, valid If-Match, valid body — differing from the happy path
// ONLY in what auth context it carries. Every rejection below is therefore
// attributable to the actor and to nothing else.
func newSubmitRequestWithAuth(auth func(context.Context) context.Context) *http.Request {
	req := httptest.NewRequest(http.MethodPost, "/api/v1/documents/doc-1/submit", strings.NewReader(submitBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", "22222222-2222-2222-2222-222222222222")
	req.Header.Set("If-Match", "\"v3\"")
	req.Header.Set("X-Request-ID", "req-123")
	ctx := tenant.WithTenantID(req.Context(), "tenant-1")
	if auth != nil {
		ctx = auth(ctx)
	}
	return req.WithContext(ctx)
}

// TestSubmitActorRED_LowLevelAccessorStillYieldsBlankActor is the RED half: the
// counterfactual, asserted rather than asserted-about.
//
// It reads the low-level accessor exactly as the handler used to, against the
// same contexts the GREEN tests reject, and shows it hands back "" — a value
// indistinguishable from a real user id at the assignment site. This is the
// input SubmittedBy would receive if the fail-closed check were removed, so it
// pins WHY the check exists rather than merely that it is present.
func TestSubmitActorRED_LowLevelAccessorStillYieldsBlankActor(t *testing.T) {
	cases := []struct {
		name string
		ctx  context.Context
	}{
		{"no auth context at all", context.Background()},
		{"empty actor stored", iamdomain.WithAuthContext(context.Background(), "", []iamdomain.Role{})},
		{"whitespace-only actor stored", iamdomain.WithAuthContext(context.Background(), "   ", []iamdomain.Role{})},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := iamdomain.UserIDFromContext(tc.ctx)
			if strings.TrimSpace(got) != "" {
				t.Fatalf("counterfactual precondition broken: low-level accessor returned %q, want a blank actor", got)
			}
			// The defect this reproduces: nothing at the old call site could
			// tell this apart from a legitimate id, so it became
			// SubmitRequest.SubmittedBy and an approval instance was opened
			// with no accountable submitter.
		})
	}
}

// TestSubmitActorGREEN_MissingOrBlankActorIsRejectedBeforeService is the GREEN
// half: the same three contexts, driven through the real handler.
func TestSubmitActorGREEN_MissingOrBlankActorIsRejectedBeforeService(t *testing.T) {
	cases := []struct {
		name string
		auth func(context.Context) context.Context
	}{
		{
			name: "no authenticated principal",
			auth: nil,
		},
		{
			name: "empty actor in auth context",
			auth: func(ctx context.Context) context.Context {
				return iamdomain.WithAuthContext(ctx, "", []iamdomain.Role{})
			},
		},
		{
			// Whitespace is its own case, not a variant of the empty one: the
			// old `== ""` comparisons this slice replaced accepted "   " as a
			// present actor and would have persisted it verbatim.
			name: "whitespace-only actor in auth context",
			auth: func(ctx context.Context) context.Context {
				return iamdomain.WithAuthContext(ctx, "   ", []iamdomain.Role{})
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc := &fakeSubmitService{}
			mux := submitTestMux(&Handler{submitSvc: svc})

			rr := httptest.NewRecorder()
			mux.ServeHTTP(rr, newSubmitRequestWithAuth(tc.auth))

			if rr.Code != http.StatusUnauthorized {
				t.Fatalf("status = %d, want %d", rr.Code, http.StatusUnauthorized)
			}
			if ct := rr.Header().Get("Content-Type"); !strings.Contains(ct, "problem+json") {
				t.Fatalf("content-type = %q, want the RFC 9457 problem media type", ct)
			}

			var prob problem.Problem
			if err := json.NewDecoder(rr.Body).Decode(&prob); err != nil {
				t.Fatalf("decode problem body: %v", err)
			}
			// The wire contract must NOT fork: a missing principal on an
			// approval route is the same condition the authn middleware
			// answers with, so it carries the platform code, not an
			// approval-specific dialect.
			if prob.Code != problem.CodeAuthUnauthenticated {
				t.Fatalf("problem code = %q, want %q", prob.Code, problem.CodeAuthUnauthenticated)
			}
			if prob.Status != http.StatusUnauthorized {
				t.Fatalf("problem status = %d, want %d", prob.Status, http.StatusUnauthorized)
			}

			// The load-bearing assertion. A 401 alone would still be
			// compatible with the service having run and written an instance
			// before something downstream complained; this proves no business
			// mutation was even attempted.
			if svc.called {
				t.Fatalf("submit service was called with a missing/blank actor (SubmittedBy would be %q)", svc.gotReq.SubmittedBy)
			}
		})
	}
}

// TestSubmitActorGREEN_ValidActorIsUnaffected is the behavior-preservation half
// of the same property (A3.3 §10): a request that already had a valid actor
// must reach the service unchanged, with the same non-empty attribution. The
// migration may only change what happens on ABSENCE.
func TestSubmitActorGREEN_ValidActorIsUnaffected(t *testing.T) {
	svc := &fakeSubmitService{}
	mux := submitTestMux(&Handler{submitSvc: svc})

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, newSubmitRequestWithAuth(func(ctx context.Context) context.Context {
		return iamdomain.WithAuthContext(ctx, "actor-1", []iamdomain.Role{})
	}))

	if rr.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusCreated)
	}
	if !svc.called {
		t.Fatalf("submit service was not called for a request with a valid actor")
	}
	if svc.gotReq.SubmittedBy != "actor-1" {
		t.Fatalf("SubmittedBy = %q, want %q", svc.gotReq.SubmittedBy, "actor-1")
	}
}
