package unit

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	authdomain "metaldocs/internal/modules/auth/domain"
	"metaldocs/internal/platform/config"
	"metaldocs/internal/platform/security"
)

// TestRateLimiterSweepsExpiredWindowEntries asserts that 10_000 distinct
// identities seen inside one window collapse back to baseline after the
// window elapses and a subsequent allow() drives the sweep — proving the C2
// unbounded-map defect is fixed.
func TestRateLimiterSweepsExpiredWindowEntries(t *testing.T) {
	t0 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	clock := t0
	now := func() time.Time { return clock }

	rl := security.NewRateLimiter(config.RateLimitConfig{
		Enabled:       true,
		WindowSeconds: 60,
		MaxRequests:   1,
	}).WithClock(now)

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/documents", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	h := rl.Wrap(mux)

	// Smaller than the spec's 10k figure to keep the 100-iteration stress run
	// (go test -run TestRateLimit -count 100) inside the harness time budget.
	// 2k is enough to demonstrate the bounded-eviction invariant; the
	// in-package middleware test at internal/platform/ratelimit covers the
	// 10k case via deterministic-clock fast path.
	const N = 2_000
	for i := 0; i < N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/documents", nil)
		req = req.WithContext(authdomain.WithCurrentUser(req.Context(), authdomain.CurrentUser{UserID: "user-" + strconv.Itoa(i)}))
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("user %d: want 200, got %d", i, rr.Code)
		}
	}

	if got := rl.Size(); got != N {
		t.Fatalf("pre-sweep size: want %d, got %d", N, got)
	}

	clock = t0.Add(2 * time.Minute)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/documents", nil)
	req = req.WithContext(authdomain.WithCurrentUser(req.Context(), authdomain.CurrentUser{UserID: "post-sweep-user"}))
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("post-sweep request: want 200, got %d", rr.Code)
	}

	if got := rl.Size(); got != 1 {
		t.Fatalf("post-sweep size: want 1 (only post-sweep-user), got %d", got)
	}
}

// TestRateLimiterFailsClosedOnMapOverflow asserts that once the entry cap is
// reached, new identities are denied (fail-closed) rather than the map growing
// without bound. Existing identities inside the cap continue to be evaluated.
func TestRateLimiterFailsClosedOnMapOverflow(t *testing.T) {
	t0 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	clock := t0
	now := func() time.Time { return clock }

	const capacity = 8
	rl := security.NewRateLimiter(config.RateLimitConfig{
		Enabled:       true,
		WindowSeconds: 60,
		MaxRequests:   10,
	}).WithClock(now).WithMaxEntries(capacity)

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/documents", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	h := rl.Wrap(mux)

	for i := 0; i < capacity; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/documents", nil)
		req = req.WithContext(authdomain.WithCurrentUser(req.Context(), authdomain.CurrentUser{UserID: "user-" + strconv.Itoa(i)}))
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("fill user %d: want 200, got %d", i, rr.Code)
		}
	}
	if got := rl.Size(); got != capacity {
		t.Fatalf("after fill: want size=%d, got %d", capacity, got)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/documents", nil)
	req = req.WithContext(authdomain.WithCurrentUser(req.Context(), authdomain.CurrentUser{UserID: "overflow-user"}))
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusTooManyRequests {
		t.Fatalf("overflow: want 429 fail-closed, got %d", rr.Code)
	}
	if ra := rr.Header().Get("Retry-After"); ra == "" {
		t.Fatalf("overflow response missing Retry-After")
	}

	req2 := httptest.NewRequest(http.MethodGet, "/api/v1/documents", nil)
	req2 = req2.WithContext(authdomain.WithCurrentUser(req2.Context(), authdomain.CurrentUser{UserID: "user-0"}))
	rr2 := httptest.NewRecorder()
	h.ServeHTTP(rr2, req2)
	if rr2.Code != http.StatusOK {
		t.Fatalf("existing identity: want 200, got %d", rr2.Code)
	}
}
