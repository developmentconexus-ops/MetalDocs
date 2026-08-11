package http_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/modules/templates/domain"
	"metaldocs/internal/platform/problem"
	"metaldocs/internal/platform/tenant"
)

// A3.3 behavioral proof for a SECOND, independent module.
//
// The approval proof (internal/modules/approval/http/submit_actor_fail_closed_test.go)
// covers a handler that stops in front of a service DOUBLE. This one covers a
// different module and a different shape: POST /templates/{id}/versions runs
// the REAL application service over an in-memory repository, so "no side
// effect" is asserted against the repository's own state — no version row, no
// audit event, no bump of the template's latest-version pointer — rather than
// against a spy's boolean.
//
// The route is deliberately the non-idempotent one. POST /templates is wrapped
// by the platform idempotency middleware, whose actor resolver also fails
// closed now; testing that route would prove the MIDDLEWARE's check and leave
// the handler's own check unproven. This route reaches the handler.

const (
	actorTenantID   = "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"
	actorTemplateID = "00000000-0000-0000-0000-0000000009a1"
)

// newActorProofRepo seeds a plain (non-system, non-archived) template with one
// published-shaped version, i.e. a template on which createNextVersion WOULD
// succeed. Every rejection below is therefore attributable to the actor and not
// to the aggregate's state.
func newActorProofRepo() *fakeRepo {
	repo := newFakeRepo()
	repo.templates[actorTemplateID] = &domain.Template{
		ID:            actorTemplateID,
		TenantID:      actorTenantID,
		Key:           "contract-default",
		Name:          "Contract Template",
		LatestVersion: 1,
	}
	repo.versions["ver-1"] = &domain.TemplateVersion{
		ID:            "ver-1",
		TemplateID:    actorTemplateID,
		VersionNumber: 1,
		Status:        domain.VersionStatusDraft,
		// ADR 0088: spawnNextDraft refuses a source without a 64-char verified
		// hash, so the seed carries one — otherwise the valid-actor case would
		// fail for a reason that has nothing to do with the actor.
		DocxStorageKey: "tenants/" + actorTenantID + "/templates/" + actorTemplateID + "/v1.docx",
		ContentHash:    strings.Repeat("a", 64),
	}
	return repo
}

func newCreateVersionRequest(auth func(context.Context) context.Context) *http.Request {
	req := httptest.NewRequest(http.MethodPost, "/api/v1/templates/"+actorTemplateID+"/versions", nil)
	req.Header.Set("content-type", "application/json")
	ctx := tenant.WithTenantID(req.Context(), actorTenantID)
	if auth != nil {
		ctx = auth(ctx)
	}
	return req.WithContext(ctx)
}

// TestTemplatesActorRED_LowLevelAccessorStillYieldsBlankActor is the RED half:
// the counterfactual, asserted rather than described.
//
// createNextVersion passes its actor into application.CreateVersionCmd
// .ActorUserID, which becomes both the new draft's creator and the ActorID of
// the audit event the service appends. Reading the low-level accessor — what
// this handler did before the migration — hands back "" for all three
// no-principal shapes, and nothing at the assignment site could tell that apart
// from a real user id.
func TestTemplatesActorRED_LowLevelAccessorStillYieldsBlankActor(t *testing.T) {
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
			if got := strings.TrimSpace(iamdomain.UserIDFromContext(tc.ctx)); got != "" {
				t.Fatalf("counterfactual precondition broken: low-level accessor returned %q, want a blank actor", got)
			}
		})
	}
}

// TestTemplatesActorGREEN_MissingOrBlankActorProducesNoSideEffect drives the
// same three contexts through the real router + real service.
func TestTemplatesActorGREEN_MissingOrBlankActorProducesNoSideEffect(t *testing.T) {
	cases := []struct {
		name string
		auth func(context.Context) context.Context
	}{
		{name: "no authenticated principal", auth: nil},
		{
			name: "empty actor in auth context",
			auth: func(ctx context.Context) context.Context {
				return iamdomain.WithAuthContext(ctx, "", []iamdomain.Role{})
			},
		},
		{
			// Whitespace is its own case: a `== ""` guard accepts "   " as a
			// present actor and would persist it verbatim as the version's
			// creator and the audit event's ActorID.
			name: "whitespace-only actor in auth context",
			auth: func(ctx context.Context) context.Context {
				return iamdomain.WithAuthContext(ctx, "   ", []iamdomain.Role{})
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := newActorProofRepo()
			authzCalled := false
			mux := newMux(t, func(_ *http.Request, _, _, _ string) error {
				authzCalled = true
				return nil
			}, repo)

			rr := httptest.NewRecorder()
			mux.ServeHTTP(rr, newCreateVersionRequest(tc.auth))

			if rr.Code != http.StatusUnauthorized {
				t.Fatalf("status = %d, want %d; body=%s", rr.Code, http.StatusUnauthorized, rr.Body.String())
			}
			if ct := rr.Header().Get("Content-Type"); !strings.Contains(ct, "problem+json") {
				t.Fatalf("content-type = %q, want the RFC 9457 problem media type", ct)
			}
			var prob problem.Problem
			if err := json.Unmarshal(rr.Body.Bytes(), &prob); err != nil {
				t.Fatalf("decode problem body: %v", err)
			}
			// Same platform code the approval module answers with: the wire
			// contract for "no principal" does not fork per module.
			if prob.Code != problem.CodeAuthUnauthenticated {
				t.Fatalf("problem code = %q, want %q", prob.Code, problem.CodeAuthUnauthenticated)
			}

			// The actor gate runs BEFORE the capability check, so an absent
			// principal is never handed to authorization as a principal with no
			// capabilities.
			if authzCalled {
				t.Fatal("authz was consulted for a request with no authenticated actor")
			}

			// The load-bearing assertions: the service was not merely
			// unsuccessful, it never ran.
			if len(repo.versions) != 1 {
				t.Fatalf("repo.versions = %d, want 1 (a new draft version was created for a blank actor)", len(repo.versions))
			}
			if len(repo.audit) != 0 {
				t.Fatalf("repo.audit = %d events, want 0 (an audit row was written under a blank actor)", len(repo.audit))
			}
			if got := repo.templates[actorTemplateID].LatestVersion; got != 1 {
				t.Fatalf("template.LatestVersion = %d, want 1 (the aggregate advanced for a blank actor)", got)
			}
		})
	}
}

// TestTemplatesActorGREEN_ValidActorIsUnaffected is the behavior-preservation
// half (A3.3 §10): a request that already carried a valid actor keeps its old
// outcome, and the attribution it writes is the same non-empty id as before.
func TestTemplatesActorGREEN_ValidActorIsUnaffected(t *testing.T) {
	repo := newActorProofRepo()
	authzCalled := false
	mux := newMux(t, func(_ *http.Request, _, _, _ string) error {
		authzCalled = true
		return nil
	}, repo)

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, newCreateVersionRequest(func(ctx context.Context) context.Context {
		return iamdomain.WithAuthContext(ctx, "user-a", []iamdomain.Role{})
	}))

	if rr.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d; body=%s", rr.Code, http.StatusCreated, rr.Body.String())
	}
	if !authzCalled {
		t.Fatal("authz was not consulted for a request with a valid actor")
	}
	if len(repo.versions) != 2 {
		t.Fatalf("repo.versions = %d, want 2 (the new draft was not created)", len(repo.versions))
	}
	if len(repo.audit) != 1 {
		t.Fatalf("repo.audit = %d events, want 1", len(repo.audit))
	}
	if got := repo.audit[0].ActorID; got != "user-a" {
		t.Fatalf("audit ActorID = %q, want %q", got, "user-a")
	}
}
