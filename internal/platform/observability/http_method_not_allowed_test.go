package observability_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"metaldocs/internal/platform/observability"
	"metaldocs/internal/platform/problem"
)

func TestMetricsHandler_MethodNotAllowedIsProblemJSON(t *testing.T) {
	o := observability.NewHTTPObservability(func(*http.Request) string { return "" })
	h := o.MetricsHandler()

	req := httptest.NewRequest(http.MethodPost, "/metrics", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusMethodNotAllowed)
	}
	if got := rr.Header().Get("Content-Type"); got != "application/problem+json" {
		t.Fatalf("Content-Type = %q, want application/problem+json", got)
	}
	if got := rr.Header().Get("Allow"); got != "GET" {
		t.Fatalf("Allow = %q, want GET", got)
	}
	var body problem.Problem
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.Code != problem.CodeMethodNotAllowed {
		t.Fatalf("code = %q, want %q", body.Code, problem.CodeMethodNotAllowed)
	}
}
