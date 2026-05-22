package idempotency_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"metaldocs/internal/platform/idempotency"
	"metaldocs/internal/testsupport/pgtest"
)

// TestMiddleware_ConcurrentSameKey_HandlerExecutesOnce wires the real Require
// middleware against the real Postgres store and proves the C3 fix end-to-end:
// two HTTP goroutines sharing the same Idempotency-Key and body cause the
// handler to run exactly once, with both clients receiving the same response.
func TestMiddleware_ConcurrentSameKey_HandlerExecutesOnce(t *testing.T) {
	db := pgtest.OpenAndMigrate(t)
	store := idempotency.New(db, "POST /mwc")

	var executed int32
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&executed, 1)
		// Hold the in_flight row so the loser is forced to block on FOR UPDATE.
		time.Sleep(150 * time.Millisecond)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(201)
		_, _ = w.Write([]byte(`{"id":"one"}`))
	})
	chain := withIDs(testTenantMW, testActorMW)(
		idempotency.Require(store, actorFromCtx)(handler),
	)

	key := "44444444-4444-4444-8444-444444444" + uniqSuffix()
	mkReq := func() *http.Request {
		req := httptest.NewRequest("POST", "/mwc", bytes.NewReader([]byte(`{"k":"v"}`)))
		req.Header.Set("Idempotency-Key", key)
		return req
	}

	var wg sync.WaitGroup
	codes := [2]int{}
	bodies := [2]string{}

	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			rec := httptest.NewRecorder()
			chain.ServeHTTP(rec, mkReq())
			codes[i] = rec.Code
			bodies[i] = rec.Body.String()
		}(i)
	}
	wg.Wait()

	if got := atomic.LoadInt32(&executed); got != 1 {
		t.Fatalf("handler ran %d times, want 1", got)
	}
	if codes[0] != 201 || codes[1] != 201 {
		t.Fatalf("both clients should get 201, got %v", codes)
	}
	if bodies[0] != bodies[1] {
		t.Fatalf("client bodies diverged: %q vs %q", bodies[0], bodies[1])
	}
}

// TestMiddleware_HandlerPanic_FreesSlot proves the FailReplay defer releases
// the slot when the handler panics, so a retry with the same key can run.
func TestMiddleware_HandlerPanic_FreesSlot(t *testing.T) {
	db := pgtest.OpenAndMigrate(t)
	store := idempotency.New(db, "POST /mwp")

	var panicked, recovered int32
	panicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&panicked, 1)
		panic("simulated handler crash")
	})
	chain := withIDs(testTenantMW, testActorMW)(
		idempotency.Require(store, actorFromCtx)(panicHandler),
	)

	key := "55555555-5555-4555-8555-555555555" + uniqSuffix()

	func() {
		defer func() {
			if r := recover(); r != nil {
				atomic.AddInt32(&recovered, 1)
			}
		}()
		req := httptest.NewRequest("POST", "/mwp", bytes.NewReader([]byte(`{}`)))
		req.Header.Set("Idempotency-Key", key)
		chain.ServeHTTP(httptest.NewRecorder(), req)
	}()

	if atomic.LoadInt32(&panicked) != 1 {
		t.Fatalf("handler should have panicked")
	}
	if atomic.LoadInt32(&recovered) != 1 {
		t.Fatalf("test harness should have caught the panic")
	}

	// Second call: handler must execute again because FailReplay released
	// the slot. Swap to a non-panicking handler since the slot is fresh.
	var retryRan int32
	retryHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&retryRan, 1)
		w.WriteHeader(200)
	})
	retryChain := withIDs(testTenantMW, testActorMW)(
		idempotency.Require(store, actorFromCtx)(retryHandler),
	)
	req2 := httptest.NewRequest("POST", "/mwp", bytes.NewReader([]byte(`{}`)))
	req2.Header.Set("Idempotency-Key", key)
	rec2 := httptest.NewRecorder()
	retryChain.ServeHTTP(rec2, req2)
	if rec2.Code != 200 {
		t.Fatalf("retry status: got %d want 200", rec2.Code)
	}
	if atomic.LoadInt32(&retryRan) != 1 {
		t.Fatalf("retry handler should have executed after Fail released the slot")
	}
}

// TestMiddleware_NonSuccess_DoesNotCacheResponse proves a 4xx/5xx response
// releases the slot via FailReplay so retries are not pinned to the failure.
func TestMiddleware_NonSuccess_DoesNotCacheResponse(t *testing.T) {
	db := pgtest.OpenAndMigrate(t)
	store := idempotency.New(db, "POST /mwn")

	var ran int32
	flakyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&ran, 1)
		// First call: 500 (transient). After Fail release, retry handler succeeds.
		if atomic.LoadInt32(&ran) == 1 {
			w.WriteHeader(500)
			_, _ = w.Write([]byte("transient"))
			return
		}
		w.WriteHeader(200)
		_, _ = w.Write([]byte("ok"))
	})
	chain := withIDs(testTenantMW, testActorMW)(
		idempotency.Require(store, actorFromCtx)(flakyHandler),
	)
	key := "66666666-6666-4666-8666-666666666" + uniqSuffix()

	req1 := httptest.NewRequest("POST", "/mwn", bytes.NewReader([]byte(`{}`)))
	req1.Header.Set("Idempotency-Key", key)
	rec1 := httptest.NewRecorder()
	chain.ServeHTTP(rec1, req1)
	if rec1.Code != 500 {
		t.Fatalf("first response: got %d want 500", rec1.Code)
	}

	req2 := httptest.NewRequest("POST", "/mwn", bytes.NewReader([]byte(`{}`)))
	req2.Header.Set("Idempotency-Key", key)
	rec2 := httptest.NewRecorder()
	chain.ServeHTTP(rec2, req2)
	if rec2.Code != 200 {
		t.Fatalf("retry response: got %d want 200 (slot must release on non-2xx)", rec2.Code)
	}
	if got := atomic.LoadInt32(&ran); got != 2 {
		t.Fatalf("handler should have run twice (failed then succeeded), ran %d", got)
	}
}

func uniqSuffix() string {
	// 3-digit deterministic-but-time-derived suffix to keep UUID shape.
	now := time.Now().UnixNano() % 1000
	return string([]byte{
		'0' + byte((now/100)%10),
		'0' + byte((now/10)%10),
		'0' + byte(now%10),
	})
}

// silence the otherwise-unused-import lint when this file is the only thing
// touching context.
var _ = context.Background
