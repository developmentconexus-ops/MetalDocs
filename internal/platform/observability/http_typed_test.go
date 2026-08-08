package observability_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"metaldocs/internal/platform/observability"
)

type stubRuntime struct{ m map[string]any }

func (s stubRuntime) Live(context.Context) (int, map[string]any)    { return http.StatusOK, nil }
func (s stubRuntime) Ready(context.Context) (int, map[string]any)   { return http.StatusOK, nil }
func (s stubRuntime) RuntimeMetrics(context.Context) map[string]any { return s.m }

// TestMetricsHandler_TypedEnvelope_KeySet is the F8.2 parity-lock for the typed
// observability.MetricsResponse envelope. With all three providers wired the body
// must carry exactly {items, runtime, scheduler, db_pool}; items must be an array.
// This pins the typed envelope to the same wire key-set the prior map[string]any
// literal produced.
func TestMetricsHandler_TypedEnvelope_KeySet(t *testing.T) {
	o := observability.NewHTTPObservability(nil, stubRuntime{m: map[string]any{"goroutines": 7}})
	o.SetSchedulerMetrics(&stubSchedulerProvider{data: map[string]any{"jobs": map[string]any{}}})
	o.SetDBPool(&stubDBPool{stats: map[string]any{"in_use": 1}})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/metrics", nil)
	rec := httptest.NewRecorder()
	o.MetricsHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status: got %d want 200", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("content-type: got %q want application/json", ct)
	}

	var body map[string]json.RawMessage
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v; body=%s", err, rec.Body.String())
	}
	for _, k := range []string{"items", "runtime", "scheduler", "db_pool"} {
		if _, ok := body[k]; !ok {
			t.Fatalf("key %q missing; body=%s", k, rec.Body.String())
		}
	}
	if len(body) != 4 {
		t.Fatalf("unexpected key set (%d keys); body=%s", len(body), rec.Body.String())
	}
	var items []any
	if err := json.Unmarshal(body["items"], &items); err != nil {
		t.Fatalf("items not an array: %v", err)
	}
}

// TestMetricsHandler_TypedEnvelope_ItemsAlwaysPresent confirms the unconditional
// `items` key (no omitempty) survives with no providers wired, and the three
// optional keys are absent.
func TestMetricsHandler_TypedEnvelope_ItemsAlwaysPresent(t *testing.T) {
	o := observability.NewHTTPObservability(nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/metrics", nil)
	rec := httptest.NewRecorder()
	o.MetricsHandler().ServeHTTP(rec, req)

	var body map[string]json.RawMessage
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if _, ok := body["items"]; !ok {
		t.Fatalf("items key absent with no providers; body=%s", rec.Body.String())
	}
	for _, k := range []string{"runtime", "scheduler", "db_pool"} {
		if _, ok := body[k]; ok {
			t.Fatalf("key %q present without provider; body=%s", k, rec.Body.String())
		}
	}
}
