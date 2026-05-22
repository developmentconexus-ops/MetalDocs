package ratelimit_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"runtime"
	"strconv"
	"testing"
	"time"

	"metaldocs/internal/platform/ratelimit"
)

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

	mw := ratelimit.New(ctx, ratelimit.Config{
		Quotas:        map[ratelimit.RouteKey]int{ratelimit.RouteExportPDF: 60},
		SweepInterval: time.Hour, // background sweeper effectively off for this test
		IdleThreshold: time.Minute,
	}).WithClock(now)
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
func TestSweeper_ExitsOnContextCancel(t *testing.T) {
	before := runtime.NumGoroutine()

	ctx, cancel := context.WithCancel(context.Background())
	mw := ratelimit.New(ctx, ratelimit.Config{
		Quotas:        map[ratelimit.RouteKey]int{ratelimit.RouteExportPDF: 60},
		SweepInterval: 10 * time.Millisecond,
		IdleThreshold: 20 * time.Millisecond,
	})

	deadline := time.Now().Add(200 * time.Millisecond)
	for time.Now().Before(deadline) {
		if runtime.NumGoroutine() > before {
			break
		}
		time.Sleep(5 * time.Millisecond)
	}
	if runtime.NumGoroutine() <= before {
		t.Fatalf("expected sweeper goroutine running; before=%d now=%d", before, runtime.NumGoroutine())
	}

	cancel()

	done := make(chan struct{})
	go func() {
		mw.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatalf("sweeper did not exit within 1s of context cancel")
	}

	for i := 0; i < 20; i++ {
		runtime.Gosched()
		if runtime.NumGoroutine() <= before {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	if runtime.NumGoroutine() > before {
		t.Fatalf("goroutine leak: before=%d after=%d", before, runtime.NumGoroutine())
	}
}

// TestLimit_LimiterReusedAcrossRequests proves the LoadOrStore fix: a second
// hit on the same key reuses the existing limiter (quota=1 → second request
// rejected because the bucket persisted).
func TestLimit_LimiterReusedAcrossRequests(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	mw := ratelimit.New(ctx, ratelimit.Config{
		Quotas:        map[ratelimit.RouteKey]int{ratelimit.RouteExportPDF: 1},
		SweepInterval: time.Hour,
		IdleThreshold: time.Hour,
	})
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

	mw := ratelimit.New(ctx, ratelimit.Config{
		Quotas:        map[ratelimit.RouteKey]int{ratelimit.RouteExportPDF: 1000},
		SweepInterval: time.Hour,
		IdleThreshold: time.Hour,
	})
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
