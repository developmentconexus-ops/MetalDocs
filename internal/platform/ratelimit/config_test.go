package ratelimit_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"metaldocs/internal/platform/ratelimit"
)

// TestNewConfig_RejectsInvalidQuotas ensures NewConfig returns an error for
// zero and negative quota values — regression against the divide-by-zero panic.
func TestNewConfig_RejectsInvalidQuotas(t *testing.T) {
	cases := []struct {
		name string
		q    map[ratelimit.RouteKey]int
	}{
		{"zero quota", map[ratelimit.RouteKey]int{ratelimit.RouteExportPDF: 0}},
		{"negative quota", map[ratelimit.RouteKey]int{ratelimit.RouteExportPDF: -5}},
		{"mixed valid and zero", map[ratelimit.RouteKey]int{
			ratelimit.RouteExportPDF:      20,
			ratelimit.RouteAutosaveCommit: 0,
		}},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			_, err := ratelimit.NewConfig(tc.q)
			if err == nil {
				t.Fatalf("expected error for quota map %v, got nil", tc.q)
			}
		})
	}
}

// TestNewConfig_AcceptsValidQuotas ensures NewConfig succeeds for all-positive values.
func TestNewConfig_AcceptsValidQuotas(t *testing.T) {
	cfg, err := ratelimit.NewConfig(map[ratelimit.RouteKey]int{
		ratelimit.RouteExportPDF:      20,
		ratelimit.RouteAutosaveCommit: 1,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v, ok := cfg.QuotaFor(ratelimit.RouteExportPDF); !ok || v != 20 {
		t.Fatalf("QuotaFor(RouteExportPDF): want (20, true), got (%d, %v)", v, ok)
	}
	if v, ok := cfg.QuotaFor(ratelimit.RouteAutosaveCommit); !ok || v != 1 {
		t.Fatalf("QuotaFor(RouteAutosaveCommit): want (1, true), got (%d, %v)", v, ok)
	}
	if _, ok := cfg.QuotaFor(ratelimit.RouteUploadsPresign); ok {
		t.Fatal("QuotaFor(RouteUploadsPresign): expected absent, got present")
	}
}

// TestNewConfig_IsolatesInputMap proves NewConfig copies the input map so
// mutations to the original after construction cannot alter the Config.
func TestNewConfig_IsolatesInputMap(t *testing.T) {
	q := map[ratelimit.RouteKey]int{ratelimit.RouteExportPDF: 20}
	cfg, err := ratelimit.NewConfig(q)
	if err != nil {
		t.Fatalf("NewConfig: %v", err)
	}
	q[ratelimit.RouteExportPDF] = 0 // mutate original
	if v, ok := cfg.QuotaFor(ratelimit.RouteExportPDF); !ok || v != 20 {
		t.Fatalf("mutation of input map leaked into Config: got (%d, %v)", v, ok)
	}
}

// TestMiddlewareBelt_ZeroQuotaDegradesToNoLimit proves that a Config with a
// zero quota (injected via test helper, bypassing NewConfig validation) causes
// the middleware to degrade to no-limit rather than panicking on integer divide.
// Regression test for H10.
func TestMiddlewareBelt_ZeroQuotaDegradesToNoLimit(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	cfg := ratelimit.NewConfigUnchecked(map[ratelimit.RouteKey]int{ratelimit.RouteExportPDF: 0})
	mw := ratelimit.New(ctx, cfg)

	reached := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reached = true
		w.WriteHeader(204)
	})
	h := mw.Limit(ratelimit.RouteExportPDF, func(r *http.Request) string { return "u1" }, next)

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest("POST", "/x", nil))

	if !reached {
		t.Fatal("expected request to pass through to next handler when quota=0 (no-limit degradation)")
	}
	if rr.Code != 204 {
		t.Fatalf("want 204, got %d", rr.Code)
	}
}

// TestDivideByZero_NoPanic is an explicit regression test that exercises the
// exact panic path: time.Minute / time.Duration(0) inside Limit.
// A zero-quota Config injected via NewConfigUnchecked must not crash the process.
func TestDivideByZero_NoPanic(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("panic on zero quota: %v", r)
		}
	}()

	cfg := ratelimit.NewConfigUnchecked(map[ratelimit.RouteKey]int{ratelimit.RouteExportPDF: 0})
	mw := ratelimit.New(ctx, cfg)
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(204) })
	h := mw.Limit(ratelimit.RouteExportPDF, func(r *http.Request) string { return "u1" }, next)

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest("POST", "/x", nil))
}
