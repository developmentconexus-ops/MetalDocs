package observability

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestNewInfraServer_ReadyReflectsDependencyCheck is the A7.1 RED-first proof
// for the shared worker/jobs infra-port mechanism: GET /ready must NOT be a
// decoration that always returns 200. It has to be capable of reporting NOT
// ready when the wired RuntimeStatusProvider says the process genuinely
// cannot do its job (forced here via a failing DependencyCheck, mirroring
// how the worker/jobs binaries wire their own "loop/River client not
// started" checks).
func TestNewInfraServer_ReadyReflectsDependencyCheck(t *testing.T) {
	failing := DependencyCheck{
		Name: "outbox_consumer",
		Check: func(context.Context) (DependencyCheckResult, error) {
			return DependencyCheckResult{}, errors.New("consumer loop has not started")
		},
	}
	provider := NewStaticRuntimeStatusProvider("postgres", "n/a", false, failing)
	server := NewInfraServer(":0", provider, nil)

	req := httptest.NewRequest(http.MethodGet, "/ready", nil)
	rec := httptest.NewRecorder()
	server.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("GET /ready with a failing dependency check = %d, want 503", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "consumer loop has not started") {
		t.Fatalf("body = %q, want it to surface the failing check detail", rec.Body.String())
	}
}

// TestNewInfraServer_ReadyGreenWhenChecksPass proves the endpoint can also
// report ready — the RED test above shows it can fail; this shows the same
// wiring reports 200 once the dependency is healthy, so the test suite pins
// both sides of the semantics, not just the failure path.
func TestNewInfraServer_ReadyGreenWhenChecksPass(t *testing.T) {
	passing := DependencyCheck{
		Name: "outbox_consumer",
		Check: func(context.Context) (DependencyCheckResult, error) {
			return DependencyCheckResult{Status: "up"}, nil
		},
	}
	provider := NewStaticRuntimeStatusProvider("postgres", "n/a", false, passing)
	server := NewInfraServer(":0", provider, nil)

	req := httptest.NewRequest(http.MethodGet, "/ready", nil)
	rec := httptest.NewRecorder()
	server.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET /ready with a passing dependency check = %d, want 200", rec.Code)
	}
}

// TestNewInfraServer_LiveIsProcessLocal pins liveness as the coarse,
// always-cheap check (k8s convention: liveness triggers a restart and must
// not fail on a degraded-but-alive dependency); it is intentionally NOT
// gated on the same DependencyCheck as readiness.
func TestNewInfraServer_LiveIsProcessLocal(t *testing.T) {
	failing := DependencyCheck{
		Name: "outbox_consumer",
		Check: func(context.Context) (DependencyCheckResult, error) {
			return DependencyCheckResult{}, errors.New("down")
		},
	}
	provider := NewStaticRuntimeStatusProvider("postgres", "n/a", false, failing)
	server := NewInfraServer(":0", provider, nil)

	req := httptest.NewRequest(http.MethodGet, "/live", nil)
	rec := httptest.NewRecorder()
	server.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET /live = %d, want 200 (liveness is process-local, not dependency-gated)", rec.Code)
	}
}

// TestNewInfraServer_MetricsServesPrometheusExposition proves the metrics
// route is wired to whatever handler the caller supplies (worker/jobs pass
// httpObs.PrometheusHandler()), served on the SAME dedicated mux as
// live/ready — no public/business route ever shares this listener.
func TestNewInfraServer_MetricsServesPrometheusExposition(t *testing.T) {
	provider := NewStaticRuntimeStatusProvider("postgres", "n/a", false)
	metrics := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4")
		_, _ = w.Write([]byte("metaldocs_probe 1\n"))
	})
	server := NewInfraServer(":0", provider, metrics)

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()
	server.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET /metrics = %d, want 200", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "metaldocs_probe 1") {
		t.Fatalf("body = %q, want the wired metrics handler's output", rec.Body.String())
	}
}

// TestNewInfraServer_NilMetricsHandlerOmitsRoute documents that a nil
// metricsHandler leaves /metrics unmounted rather than panicking — callers
// that have not wired Prometheus yet still get a working live/ready server.
func TestNewInfraServer_NilMetricsHandlerOmitsRoute(t *testing.T) {
	provider := NewStaticRuntimeStatusProvider("postgres", "n/a", false)
	server := NewInfraServer(":0", provider, nil)

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()
	server.Handler.ServeHTTP(rec, req)

	if rec.Code == http.StatusOK {
		t.Fatalf("GET /metrics with no metrics handler wired = 200, want non-200 (no route mounted)")
	}
}

// TestNewInfraServer_BoundsAllTimeouts is the F4 review round 2 RED-first
// proof: NewInfraServer used to set only ReadHeaderTimeout, leaving
// ReadTimeout/WriteTimeout/IdleTimeout at their zero-value (unbounded)
// default — a reachable client could hold a connection open indefinitely
// (Slowloris, CWE-400). All four must now be non-zero.
func TestNewInfraServer_BoundsAllTimeouts(t *testing.T) {
	provider := NewStaticRuntimeStatusProvider("postgres", "n/a", false)
	server := NewInfraServer(":0", provider, nil)

	if server.ReadHeaderTimeout <= 0 {
		t.Error("ReadHeaderTimeout is unbounded (<= 0)")
	}
	if server.ReadTimeout <= 0 {
		t.Error("ReadTimeout is unbounded (<= 0)")
	}
	if server.WriteTimeout <= 0 {
		t.Error("WriteTimeout is unbounded (<= 0)")
	}
	if server.IdleTimeout <= 0 {
		t.Error("IdleTimeout is unbounded (<= 0)")
	}
	if server.ReadTimeout < server.ReadHeaderTimeout {
		t.Error("ReadTimeout should be >= ReadHeaderTimeout")
	}
}
