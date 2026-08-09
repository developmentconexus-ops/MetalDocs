package middleware_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"metaldocs/internal/platform/middleware"
	"metaldocs/internal/platform/requesttrace"
)

func TestRecovery_PanicProduces500ProblemJSON(t *testing.T) {
	h := middleware.Recovery(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("boom")
	}))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/anything", nil))

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/problem+json" {
		t.Fatalf("content-type = %q, want application/problem+json", ct)
	}
	var body struct {
		Code   string `json:"code"`
		Status int    `json:"status"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("body not JSON: %v", err)
	}
	if body.Code != "internal.unknown" || body.Status != 500 {
		t.Fatalf("body = %+v, want code INTERNAL_ERROR status 500", body)
	}
}

func TestRecovery_SetsTraceIDInContext(t *testing.T) {
	var got string
	h := middleware.Recovery(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		got, _ = requesttrace.FromContext(r.Context())
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Trace-Id", "incoming-trace-id")
	h.ServeHTTP(httptest.NewRecorder(), req)

	if got != "incoming-trace-id" {
		t.Fatalf("trace id in handler context = %q, want %q", got, "incoming-trace-id")
	}

	got = ""
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))
	if got == "" {
		t.Fatal("expected a generated trace id when no header is present")
	}
}

func TestRecovery_RepanicsOnErrAbortHandler(t *testing.T) {
	h := middleware.Recovery(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic(http.ErrAbortHandler)
	}))

	defer func() {
		recErr, ok := recover().(error)
		if !ok || !errors.Is(recErr, http.ErrAbortHandler) {
			t.Fatal("expected http.ErrAbortHandler to propagate")
		}
	}()
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))
	t.Fatal("unreachable: panic expected")
}

func TestRecovery_NormalRequestPassesThrough(t *testing.T) {
	h := middleware.Recovery(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusTeapot {
		t.Fatalf("status = %d, want 418", rec.Code)
	}
}
