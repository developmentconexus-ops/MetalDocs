package observability_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"metaldocs/internal/platform/observability"
)

// stubDBPoolProm is a minimal DBPoolStatsProvider for the Prometheus gauge test.
type stubDBPoolProm struct {
	stats map[string]any
}

func (s *stubDBPoolProm) DBPoolStats() map[string]any { return s.stats }

func TestPrometheusHandler_ExposesBindingMetricFamilies(t *testing.T) {
	obs := observability.NewHTTPObservability(func(*http.Request) string { return "" })
	obs.SetDBPool(&stubDBPoolProm{stats: map[string]any{"in_use": 2, "idle": 5}})

	// Drive a couple of requests through Wrap: one success, one error.
	okHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	errHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	mux := http.NewServeMux()
	mux.Handle("GET /widgets", obs.Wrap(okHandler))
	mux.Handle("GET /explode", obs.Wrap(errHandler))

	req1 := httptest.NewRequest(http.MethodGet, "/widgets", nil)
	rec1 := httptest.NewRecorder()
	mux.ServeHTTP(rec1, req1)

	req2 := httptest.NewRequest(http.MethodGet, "/explode", nil)
	rec2 := httptest.NewRecorder()
	mux.ServeHTTP(rec2, req2)

	// Scrape.
	scrapeReq := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	scrapeRec := httptest.NewRecorder()
	obs.PrometheusHandler().ServeHTTP(scrapeRec, scrapeReq)

	if scrapeRec.Code != http.StatusOK {
		t.Fatalf("expected 200 from PrometheusHandler, got %d", scrapeRec.Code)
	}
	body := scrapeRec.Body.String()

	// 1. requests_total
	if !strings.Contains(body, `metaldocs_http_requests_total{method="GET",route="/widgets"} 1`) {
		t.Errorf("missing metaldocs_http_requests_total for /widgets, body:\n%s", body)
	}
	if !strings.Contains(body, `metaldocs_http_requests_total{method="GET",route="/explode"} 1`) {
		t.Errorf("missing metaldocs_http_requests_total for /explode, body:\n%s", body)
	}

	// 2. errors_total — only the 500 route should have an error counted.
	if !strings.Contains(body, `metaldocs_http_request_errors_total{method="GET",route="/explode"} 1`) {
		t.Errorf("missing metaldocs_http_request_errors_total for /explode, body:\n%s", body)
	}
	if strings.Contains(body, `metaldocs_http_request_errors_total{method="GET",route="/widgets"}`) {
		t.Errorf("did not expect error counter series for /widgets (no error occurred), body:\n%s", body)
	}

	// 3. duration histogram
	if !strings.Contains(body, "metaldocs_http_request_duration_seconds_bucket") {
		t.Errorf("missing metaldocs_http_request_duration_seconds_bucket, body:\n%s", body)
	}
	if !strings.Contains(body, `metaldocs_http_request_duration_seconds_count{method="GET",route="/widgets"} 1`) {
		t.Errorf("missing duration_seconds_count for /widgets, body:\n%s", body)
	}

	// 4. Go runtime + process collectors.
	if !strings.Contains(body, "go_goroutines") {
		t.Errorf("missing go_goroutines (Go collector), body:\n%s", body)
	}
	if !strings.Contains(body, "process_start_time_seconds") {
		t.Errorf("missing process_start_time_seconds (process collector), body:\n%s", body)
	}

	// 5. DB pool gauges.
	if !strings.Contains(body, "metaldocs_db_pool_in_use 2") {
		t.Errorf("missing metaldocs_db_pool_in_use gauge, body:\n%s", body)
	}
	if !strings.Contains(body, "metaldocs_db_pool_idle 5") {
		t.Errorf("missing metaldocs_db_pool_idle gauge, body:\n%s", body)
	}
}

func TestPrometheusHandler_NoDBPoolGauges_WhenNotWired(t *testing.T) {
	obs := observability.NewHTTPObservability(func(*http.Request) string { return "" })
	// SetDBPool NOT called.

	scrapeReq := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	scrapeRec := httptest.NewRecorder()
	obs.PrometheusHandler().ServeHTTP(scrapeRec, scrapeReq)

	body := scrapeRec.Body.String()
	if strings.Contains(body, "metaldocs_db_pool_") {
		t.Errorf("did not expect any metaldocs_db_pool_ series when DBPool provider not wired, body:\n%s", body)
	}
}

// TestMetricsHandler_JSONUnaffectedByPrometheusInstrumentation pins that the
// existing JSON /api/v1/metrics snapshot behavior is untouched by adding
// Prometheus instrumentation to the same Wrap hot path.
func TestMetricsHandler_JSONUnaffectedByPrometheusInstrumentation(t *testing.T) {
	obs := observability.NewHTTPObservability(func(*http.Request) string { return "" })
	okHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux := http.NewServeMux()
	mux.Handle("GET /widgets", obs.Wrap(okHandler))
	req := httptest.NewRequest(http.MethodGet, "/widgets", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	jsonReq := httptest.NewRequest(http.MethodGet, "/api/v1/metrics", nil)
	jsonRec := httptest.NewRecorder()
	obs.MetricsHandler().ServeHTTP(jsonRec, jsonReq)
	if jsonRec.Code != http.StatusOK {
		t.Fatalf("expected 200 from MetricsHandler, got %d", jsonRec.Code)
	}
	if !strings.Contains(jsonRec.Body.String(), `"route":"/widgets"`) {
		t.Errorf("expected JSON snapshot to still include /widgets route, body:\n%s", jsonRec.Body.String())
	}
}
