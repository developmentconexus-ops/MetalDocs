package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// methodMux builds a Go 1.22 method-routed ServeMux with a single GET-only
// route, so a DELETE produces the stdlib text/plain 405 the interceptor must
// rewrite, and an unknown path produces the stdlib text/plain 404.
func methodMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /thing", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	return mux
}

func decodeProblem(t *testing.T, body []byte) map[string]any {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal(body, &m); err != nil {
		t.Fatalf("body is not JSON (%v): %q", err, string(body))
	}
	return m
}

func TestMethodNotAllowedJSON_RewritesStdlib405(t *testing.T) {
	h := MethodNotAllowedJSON(methodMux())
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodDelete, "/thing", nil))

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want 405", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/problem+json" {
		t.Fatalf("Content-Type = %q, want application/problem+json", ct)
	}
	if allow := rec.Header().Get("Allow"); !strings.Contains(allow, http.MethodGet) {
		t.Fatalf("Allow = %q, want it to contain GET", allow)
	}
	m := decodeProblem(t, rec.Body.Bytes())
	if m["code"] != "METHOD_NOT_ALLOWED" {
		t.Fatalf("code = %v, want METHOD_NOT_ALLOWED", m["code"])
	}
}

func TestMethodNotAllowedJSON_RewritesStdlib404(t *testing.T) {
	h := MethodNotAllowedJSON(methodMux())
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/nope", nil))

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/problem+json" {
		t.Fatalf("Content-Type = %q, want application/problem+json", ct)
	}
	m := decodeProblem(t, rec.Body.Bytes())
	if m["code"] != "NOT_FOUND" {
		t.Fatalf("code = %v, want NOT_FOUND", m["code"])
	}
}

func TestMethodNotAllowedJSON_PassesThroughHandcodedProblem(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/hand", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Allow", "GET")
		w.Header().Set("Content-Type", "application/problem+json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		_, _ = w.Write([]byte(`{"code":"METHOD_NOT_ALLOWED","title":"hand"}`))
	})
	h := MethodNotAllowedJSON(mux)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/hand", nil))

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want 405", rec.Code)
	}
	m := decodeProblem(t, rec.Body.Bytes())
	if m["title"] != "hand" {
		t.Fatalf("hand-coded problem body was rewritten: %q", rec.Body.String())
	}
	if got := strings.Count(rec.Body.String(), "{"); got != 1 {
		t.Fatalf("expected single JSON object, got %d (double-write?): %q", got, rec.Body.String())
	}
}

func TestMethodNotAllowedJSON_PassesThroughSuccess(t *testing.T) {
	h := MethodNotAllowedJSON(methodMux())
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/thing", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if rec.Body.String() != "ok" {
		t.Fatalf("body = %q, want ok", rec.Body.String())
	}
}

func TestMethodNotAllowedJSON_PreservesFlusher(t *testing.T) {
	var sawFlusher bool
	h := MethodNotAllowedJSON(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, sawFlusher = w.(http.Flusher)
		w.WriteHeader(http.StatusOK)
	}))
	rec := httptest.NewRecorder() // httptest.ResponseRecorder implements http.Flusher
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/x", nil))
	if !sawFlusher {
		t.Fatal("wrapped writer must expose http.Flusher when the underlying writer does")
	}
}
