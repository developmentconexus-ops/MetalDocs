package observability_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"metaldocs/internal/platform/observability"
)

type stubDBPool struct {
	stats map[string]any
}

func (s *stubDBPool) DBPoolStats() map[string]any { return s.stats }

func TestMetricsHandler_IncludesDBPoolStats(t *testing.T) {
	obs := observability.NewHTTPObservability(func(*http.Request) string { return "" })
	obs.SetDBPool(&stubDBPool{stats: map[string]any{"in_use": 2, "idle": 3}})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/metrics", nil)
	rec := httptest.NewRecorder()
	obs.MetricsHandler().ServeHTTP(rec, req)

	var payload map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode: %v", err)
	}
	pool, ok := payload["db_pool"]
	if !ok {
		t.Fatal("expected db_pool key in metrics payload")
	}
	poolMap, ok := pool.(map[string]any)
	if !ok {
		t.Fatalf("db_pool should be map, got %T", pool)
	}
	if v, ok := poolMap["in_use"]; !ok || v != float64(2) {
		t.Errorf("expected in_use=2, got %v", poolMap["in_use"])
	}
}

func TestMetricsHandler_NoDBPoolKey_WhenNotWired(t *testing.T) {
	obs := observability.NewHTTPObservability(func(*http.Request) string { return "" })
	// SetDBPool NOT called

	req := httptest.NewRequest(http.MethodGet, "/api/v1/metrics", nil)
	rec := httptest.NewRecorder()
	obs.MetricsHandler().ServeHTTP(rec, req)

	var payload map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if _, ok := payload["db_pool"]; ok {
		t.Fatal("db_pool key should be absent when SetDBPool not called")
	}
}
