package ratelimit_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"metaldocs/internal/platform/ratelimit"
)

func mustConfig(t *testing.T, q map[ratelimit.RouteKey]int, sweep, idle time.Duration) ratelimit.Config {
	t.Helper()
	cfg, err := ratelimit.NewConfig(q)
	if err != nil {
		t.Fatalf("NewConfig: %v", err)
	}
	cfg.SweepInterval = sweep
	cfg.IdleThreshold = idle
	return cfg
}

// TestSweep_EvictsIdleEntries proves that limiter entries idle longer than
// the configured IdleThreshold are dropped by the sweeper, so 10_000 unique
// identities collapse back to baseline after one window — driven against a
// deterministic clock so the test is not racy.
func TestSweep_EvictsIdleEntries(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	t0 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	clock := t0
	now := func() time.Time { return clock }

	cfg := mustConfig(t,
		map[ratelimit.RouteKey]int{ratelimit.RouteExportPDF: 60},
		time.Hour, // background sweeper effectively off for this test
		time.Minute,
	)
	mw := ratelimit.New(ctx, cfg).WithClock(now)
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(204) })
	h := mw.Limit(ratelimit.RouteExportPDF, func(r *http.Request) string { return r.Header.Get("x-user") }, next)

	const N = 10_000
	for i := 0; i < N; i++ {
		req := httptest.NewRequest("POST", "/x", nil)
		req.Header.Set("x-user", strconv.Itoa(i))
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)
		if rr.Code != 204 {
			t.Fatalf("unique user %d: want 204, got %d", i, rr.Code)
		}
	}

	if got := mw.Size(); got != N {
		t.Fatalf("pre-sweep size: want %d, got %d", N, got)
	}

	// Advance past idle threshold, then run sweep synchronously.
	clock = t0.Add(2 * time.Minute)
	mw.SweepNow()

	if got := mw.Size(); got != 0 {
		t.Fatalf("post-sweep size: want 0, got %d", got)
	}
}

// TestSweeper_ExitsOnContextCancel proves the sweeper goroutine terminates
// when the constructor-provided context is canceled (no goroutine leak).
//
// Deliberately built on the Middleware WaitGroup join (Wait) rather than
// runtime.NumGoroutine(): the process-global goroutine count flaked under
// full-suite parallel load (TST-04) — unrelated tests start/stop goroutines
// inside the observation windows. Wait() returning IS the leak proof: it
// joins the sweeper's WaitGroup, which is Done()ed only when sweepLoop
// returns. External goroutine churn cannot make Wait return early or late.
func TestSweeper_ExitsOnContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cfg := mustConfig(t,
		map[ratelimit.RouteKey]int{ratelimit.RouteExportPDF: 60},
		10*time.Millisecond,
		20*time.Millisecond,
	)
	mw := ratelimit.New(ctx, cfg)

	done := make(chan struct{})
	go func() {
		mw.Wait()
		close(done)
	}()

	// Pre-cancel the sweeper must still be running: Wait must not have
	// returned. One-directional check — external load cannot fake either
	// outcome (only a sweeper that already exited closes done early).
	select {
	case <-done:
		t.Fatal("sweeper exited before context cancel")
	case <-time.After(50 * time.Millisecond):
	}

	cancel()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("sweeper did not exit within 1s of context cancel")
	}
}

// TestLimit_LimiterReusedAcrossRequests proves the LoadOrStore fix: a second
// hit on the same key reuses the existing limiter (quota=1 → second request
// rejected because the bucket persisted).
func TestLimit_LimiterReusedAcrossRequests(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	cfg := mustConfig(t,
		map[ratelimit.RouteKey]int{ratelimit.RouteExportPDF: 1},
		time.Hour,
		time.Hour,
	)
	mw := ratelimit.New(ctx, cfg)
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(204) })
	h := mw.Limit(ratelimit.RouteExportPDF, func(r *http.Request) string { return "u1" }, next)

	rr1 := httptest.NewRecorder()
	h.ServeHTTP(rr1, httptest.NewRequest("POST", "/x", nil))
	if rr1.Code != 204 {
		t.Fatalf("first: want 204, got %d", rr1.Code)
	}
	rr2 := httptest.NewRecorder()
	h.ServeHTTP(rr2, httptest.NewRequest("POST", "/x", nil))
	if rr2.Code != http.StatusTooManyRequests {
		t.Fatalf("second: want 429 (proves limiter persisted, not freshly allocated), got %d", rr2.Code)
	}
	if mw.Size() != 1 {
		t.Fatalf("want exactly 1 limiter entry, got %d", mw.Size())
	}
}

// TestLimit_RaceUnderConcurrentFirstHit guards the LoadOrStore-on-miss path:
// regardless of how many goroutines race the first hit, exactly one limiter
// survives per distinct key.
func TestLimit_RaceUnderConcurrentFirstHit(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	cfg := mustConfig(t,
		map[ratelimit.RouteKey]int{ratelimit.RouteExportPDF: 1000},
		time.Hour,
		time.Hour,
	)
	mw := ratelimit.New(ctx, cfg)
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(204) })
	h := mw.Limit(ratelimit.RouteExportPDF, func(r *http.Request) string { return r.Header.Get("x-user") }, next)

	const goroutines = 64
	done := make(chan struct{}, goroutines)
	for g := 0; g < goroutines; g++ {
		go func(gid int) {
			defer func() { done <- struct{}{} }()
			rr := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/x", nil)
			req.Header.Set("x-user", fmt.Sprintf("u%d", gid%4))
			h.ServeHTTP(rr, req)
		}(g)
	}
	for i := 0; i < goroutines; i++ {
		<-done
	}
	if got := mw.Size(); got != 4 {
		t.Fatalf("want exactly 4 limiter entries (one per distinct key), got %d", got)
	}
}
