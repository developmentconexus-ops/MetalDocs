package presence

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	iamdomain "metaldocs/internal/modules/iam/domain"
)

// A3.3 class B proof: identity is OPTIONAL here by an explicit, documented
// policy, and the policy is what the test pins.
//
// Every other migrated actor consumer in this slice fails closed. This
// middleware deliberately does not, and the reason is in Wrap's comment: a
// presence bump is a best-effort last-seen UPDATE that gates nothing, and this
// middleware sits AHEAD of the routes' own authorization — 401-ing here would
// reject unauthenticated traffic that the login route (and every other
// deliberately public path) must still serve.
//
// So the policy is "no actor → skip the bump, serve the request". The two
// halves are equally load-bearing and both are asserted below: absence must not
// reach the repository (a bump of the "" user), and absence must not turn into
// a rejection either.
func TestBumpMiddlewareClassB_AbsentActorSkipsBumpAndStillServes(t *testing.T) {
	cases := []struct {
		name string
		auth func(context.Context) context.Context
	}{
		{name: "no auth context at all", auth: nil},
		{
			name: "empty actor in auth context",
			auth: func(ctx context.Context) context.Context {
				return iamdomain.WithAuthContext(ctx, "", nil)
			},
		},
		{
			// Before A3.3 this middleware read the low-level accessor and
			// bumped whatever came back. "   " is non-empty at that layer, so
			// it reached BumpLastSeen as a user id and the debounce map grew an
			// entry keyed on whitespace. The canonical accessor trims, so the
			// policy now covers this shape too.
			name: "whitespace-only actor in auth context",
			auth: func(ctx context.Context) context.Context {
				return iamdomain.WithAuthContext(ctx, "   ", nil)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := newFakeRepo()
			mw := NewBumpMiddleware(repo, nil)

			served := 0
			h := mw.Wrap(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				served++
				w.WriteHeader(http.StatusNoContent)
			}))

			req := httptest.NewRequest(http.MethodGet, "/api/v1/anything", nil)
			if tc.auth != nil {
				req = req.WithContext(tc.auth(req.Context()))
			}
			rr := httptest.NewRecorder()
			h.ServeHTTP(rr, req)

			// Half one: the request is served, not rejected. A 401 here would
			// be a behavior change this slice explicitly must not make.
			if served != 1 {
				t.Fatalf("downstream handler called %d times, want 1", served)
			}
			if rr.Code != http.StatusNoContent {
				t.Fatalf("status = %d, want %d (the middleware must not answer for an anonymous request)", rr.Code, http.StatusNoContent)
			}

			// Half two: nothing was written on behalf of a non-principal. The
			// bump is asynchronous, so give the goroutine a window in which it
			// WOULD have landed.
			time.Sleep(50 * time.Millisecond)
			if got := repo.Bumps(); len(got) != 0 {
				t.Fatalf("bumped last_seen_at for a request with no actor: %+v", got)
			}
			mw.mu.Lock()
			tracked := len(mw.lastBump)
			mw.mu.Unlock()
			if tracked != 0 {
				t.Fatalf("debounce map holds %d entries for a request with no actor, want 0", tracked)
			}
		})
	}
}

// TestBumpMiddlewareClassB_PresentActorStillBumps is the other side of the
// policy: classifying this site as "optional" must not have cost it its actual
// job. A request that carries a principal is bumped exactly as before.
func TestBumpMiddlewareClassB_PresentActorStillBumps(t *testing.T) {
	repo := newFakeRepo()
	mw := NewBumpMiddleware(repo, nil)

	h := mw.Wrap(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/anything", nil).
		WithContext(iamdomain.WithAuthContext(context.Background(), "alice", nil))
	h.ServeHTTP(httptest.NewRecorder(), req)

	if !waitForBump(repo, "alice", time.Second) {
		t.Fatal("expected a bump for the authenticated actor within 1s")
	}
}
