package observability_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"metaldocs/internal/platform/observability"
)

// stubSchedulerProvider satisfies observability.SchedulerMetricsProvider.
type stubSchedulerProvider struct {
	data map[string]any
}

func (s *stubSchedulerProvider) SchedulerMetrics() map[string]any {
	return s.data
}

func TestMetricsHandler_IncludesSchedulerCounters(t *testing.T) {
	stub := &stubSchedulerProvider{
		data: map[string]any{
			"jobs": map[string]any{
				"probe": map[string]int64{"runs": 3, "errors": 1, "skips": 0},
			},
		},
	}
	o := observability.NewHTTPObservability(func(*http.Request) string { return "" })
	o.SetSchedulerMetrics(stub)
	h := o.MetricsHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/metrics", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	var body map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	sched, ok := body["scheduler"]
	if !ok {
		t.Fatal("response body missing 'scheduler' key")
	}
	schedMap, ok := sched.(map[string]any)
	if !ok {
		t.Fatalf("scheduler value is not map[string]any: %T", sched)
	}
	jobs, ok := schedMap["jobs"].(map[string]any)
	if !ok {
		t.Fatalf("scheduler.jobs is not map[string]any: %T", schedMap["jobs"])
	}
	probe, ok := jobs["probe"].(map[string]any)
	if !ok {
		t.Fatalf("scheduler.jobs.probe is not map[string]any: %T", jobs["probe"])
	}
	// JSON numbers decode as float64.
	if got := probe["runs"]; got != float64(3) {
		t.Fatalf("scheduler.jobs.probe.runs = %v; want 3", got)
	}
	if got := probe["errors"]; got != float64(1) {
		t.Fatalf("scheduler.jobs.probe.errors = %v; want 1", got)
	}
}

func TestMetricsHandler_NoSchedulerKey_WhenNotWired(t *testing.T) {
	o := observability.NewHTTPObservability(func(*http.Request) string { return "" })
	// No SetSchedulerMetrics call — nil guard must suppress the key.
	h := o.MetricsHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/metrics", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	var body map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if _, ok := body["scheduler"]; ok {
		t.Fatal("response body must not contain 'scheduler' key when no provider is wired")
	}
}
