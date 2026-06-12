package unit

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	authdomain "metaldocs/internal/modules/auth/domain"
	"metaldocs/internal/platform/ratelimit"
)

// TestRateLimiterSweepsExpiredWindowEntries asserts that idle limiter entries
// are evicted by the background sweeper, bounding memory against accumulation
// of distinct identities (C2 fix — now in platform/ratelimit).
func TestRateLimiterSweepsExpiredWindowEntries(t *testing.T) {
	t0 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	clock := t0
	clockFn := func() time.Time { return clock }

	cfg, err := ratelimit.NewConfig(map[ratelimit.RouteKey]int{ratelimit.RouteGlobalEnvelope: 10})
	if err != nil {
		t.Fatalf("ratelimit.NewConfig: %v", err)
	}
	cfg.IdleThreshold = time.Minute

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	m := ratelimit.New(ctx, cfg).WithClock(clockFn)

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/documents", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	h := m.GlobalEnvelopeWrap(userIDExtractorForTest, mux)

	// Smaller than the spec's 10k to keep the count=100 stress run inside budget.
	const N = 2_000
	for i := 0; i < N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/documents", nil)
		req = req.WithContext(authdomain.WithCurrentUser(req.Context(), authdomain.CurrentUser{UserID: "user-" + strconv.Itoa(i)}))
		req.RemoteAddr = "203.0.113.1:1234"
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("user %d: want 200, got %d", i, rr.Code)
		}
	}

	if got := m.Size(); got != N {
		t.Fatalf("pre-sweep size: want %d, got %d", N, got)
	}

	// Advance clock past idle threshold so all entries age out on next sweep.
	clock = t0.Add(5 * time.Minute)
	m.SweepNow()

	if got := m.Size(); got != 0 {
		t.Fatalf("post-sweep size: want 0 (all idle evicted), got %d", got)
	}

	// A new admission after sweep succeeds and creates exactly one entry.
	req := httptest.NewRequest(http.MethodGet, "/api/v1/documents", nil)
	req = req.WithContext(authdomain.WithCurrentUser(req.Context(), authdomain.CurrentUser{UserID: "post-sweep-user"}))
	req.RemoteAddr = "203.0.113.1:1234"
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("post-sweep request: want 200, got %d", rr.Code)
	}

	if got := m.Size(); got != 1 {
		t.Fatalf("post-sweep size: want 1 (only post-sweep-user), got %d", got)
	}
}

// TestRateLimiterFailsClosedOnMapOverflow asserts that once the entry cap is
// reached, new identities are denied (fail-closed) rather than the map growing
// without bound. Existing identities inside the cap continue to be evaluated.
func TestRateLimiterFailsClosedOnMapOverflow(t *testing.T) {
	cfg, err := ratelimit.NewConfig(map[ratelimit.RouteKey]int{ratelimit.RouteGlobalEnvelope: 10})
	if err != nil {
		t.Fatalf("ratelimit.NewConfig: %v", err)
	}
	const capacity = 8
	cfg.MaxEntries = capacity

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	m := ratelimit.New(ctx, cfg)

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/documents", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	h := m.GlobalEnvelopeWrap(userIDExtractorForTest, mux)

	for i := 0; i < capacity; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/documents", nil)
		req = req.WithContext(authdomain.WithCurrentUser(req.Context(), authdomain.CurrentUser{UserID: "user-" + strconv.Itoa(i)}))
		req.RemoteAddr = "203.0.113.1:1234"
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("fill user %d: want 200, got %d", i, rr.Code)
		}
	}
	if got := m.Size(); got != capacity {
		t.Fatalf("after fill: want size=%d, got %d", capacity, got)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/documents", nil)
	req = req.WithContext(authdomain.WithCurrentUser(req.Context(), authdomain.CurrentUser{UserID: "overflow-user"}))
	req.RemoteAddr = "203.0.113.2:1234"
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
	req2.RemoteAddr = "203.0.113.1:1235"
	rr2 := httptest.NewRecorder()
	h.ServeHTTP(rr2, req2)
	if rr2.Code != http.StatusOK {
		t.Fatalf("existing identity: want 200, got %d", rr2.Code)
	}
}
