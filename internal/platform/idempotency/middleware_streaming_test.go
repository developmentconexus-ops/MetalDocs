package idempotency_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"metaldocs/internal/platform/idempotency"
	"metaldocs/internal/testsupport/pgtest"
)

// streamingHandler asserts the writer is a Flusher and calls Flush.
// Without idempotency.WithStreamingOptOut, the wrapped recorder MUST
// expose Flush in a fail-closed way (panic with a clear directive)
// rather than silently buffering or panicking with an opaque
// interface-conversion error.
func streamingHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(200)
		_, _ = w.Write([]byte("data: one\n\n"))
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "no flusher", 500)
			return
		}
		flusher.Flush()
		_, _ = w.Write([]byte("data: two\n\n"))
		flusher.Flush()
	})
}

// TestMiddleware_StreamingHandler_WithoutOptOut_FailsClosed proves that
// a handler invoking Flush on the idempotency-wrapped writer triggers a
// clear panic naming WithStreamingOptOut. The middleware's deferred
// recover logs and re-panics, so the panic surfaces here.
func TestMiddleware_StreamingHandler_WithoutOptOut_FailsClosed(t *testing.T) {
	db := pgtest.OpenAndMigrate(t)
	store := idempotency.New(db, "POST /stream")
	h := withIDs(testTenantMW, testActorMW)(
		idempotency.Require(store, actorFromCtx)(streamingHandler()),
	)
	req := httptest.NewRequest("POST", "/stream", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Idempotency-Key", "55555555-5555-4555-8555-555555555555")
	rec := httptest.NewRecorder()

	defer func() {
		p := recover()
		if p == nil {
			t.Fatalf("expected panic from Flush on wrapped recorder; got none")
		}
		msg, ok := p.(string)
		if !ok {
			t.Fatalf("panic type: got %T want string", p)
		}
		if !strings.Contains(msg, "WithStreamingOptOut") {
			t.Fatalf("panic message missing WithStreamingOptOut directive: %q", msg)
		}
		if !strings.Contains(msg, "idempotency:") {
			t.Fatalf("panic message missing idempotency prefix: %q", msg)
		}
	}()
	h.ServeHTTP(rec, req)
}

// TestMiddleware_StreamingHandler_WithOptOut_PassesThroughUnwrapped
// proves that a route marked streaming bypasses idempotency entirely:
// no Idempotency-Key required, Flush works on the raw httptest recorder,
// and the full streamed body reaches the client.
func TestMiddleware_StreamingHandler_WithOptOut_PassesThroughUnwrapped(t *testing.T) {
	db := pgtest.OpenAndMigrate(t)
	store := idempotency.New(db, "POST /stream")

	optOut := func(r *http.Request) bool {
		return r.URL.Path == "/stream"
	}
	h := withIDs(testTenantMW, testActorMW)(
		idempotency.Require(store, actorFromCtx, idempotency.WithStreamingOptOut(optOut))(streamingHandler()),
	)

	req := httptest.NewRequest("POST", "/stream", bytes.NewReader([]byte(`{}`)))
	// Deliberately no Idempotency-Key — opt-out must bypass the header check.
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("status: got %d want 200", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "data: one") || !strings.Contains(body, "data: two") {
		t.Fatalf("missing streamed chunks: %q", body)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "text/event-stream" {
		t.Fatalf("content-type: got %q want text/event-stream", ct)
	}
}

// TestMiddleware_OptOutFalse_StillEnforcesIdempotency proves the matcher
// is consulted per-request: when it returns false, the middleware enforces
// the Idempotency-Key contract as usual.
func TestMiddleware_OptOutFalse_StillEnforcesIdempotency(t *testing.T) {
	db := pgtest.OpenAndMigrate(t)
	store := idempotency.New(db, "POST /test")
	neverOptOut := func(*http.Request) bool { return false }
	h := withIDs(testTenantMW, testActorMW)(
		idempotency.Require(store, actorFromCtx, idempotency.WithStreamingOptOut(neverOptOut))(handler201(`{}`)),
	)
	req := httptest.NewRequest("POST", "/test", bytes.NewReader([]byte(`{}`)))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != 400 {
		t.Fatalf("status: got %d want 400 (idempotency still enforced)", rec.Code)
	}
}
