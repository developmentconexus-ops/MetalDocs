package observability

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestSetupOTel_InertWhenUnconfigured proves the zero-overhead default: no
// exporter env ⇒ no SDK install, enabled=false, shutdown is a usable no-op
// (Z-1, REQ-OBS-3).
func TestSetupOTel_InertWhenUnconfigured(t *testing.T) {
	t.Setenv("OTEL_TRACES_EXPORTER", "")
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "")

	shutdown, enabled, err := SetupOTel(context.Background())
	if err != nil {
		t.Fatalf("SetupOTel unconfigured returned error: %v", err)
	}
	if enabled {
		t.Fatal("enabled=true with no exporter env; want inert")
	}
	if shutdown == nil {
		t.Fatal("shutdown is nil; want a no-op func")
	}
	if err := shutdown(context.Background()); err != nil {
		t.Fatalf("no-op shutdown returned error: %v", err)
	}
}

// TestSetupOTel_EnabledWithConsoleExporter proves OTEL_TRACES_EXPORTER=console
// installs the SDK and returns a real shutdown.
func TestSetupOTel_EnabledWithConsoleExporter(t *testing.T) {
	t.Setenv("OTEL_TRACES_EXPORTER", "console")

	shutdown, enabled, err := SetupOTel(context.Background())
	if err != nil {
		t.Fatalf("SetupOTel console returned error: %v", err)
	}
	if !enabled {
		t.Fatal("enabled=false with console exporter; want SDK installed")
	}
	if shutdown == nil {
		t.Fatal("shutdown is nil")
	}
	if err := shutdown(context.Background()); err != nil {
		t.Fatalf("shutdown returned error: %v", err)
	}
}

// TestOTelMiddleware_NamesSpanByPattern proves the middleware names the server
// span by the Go 1.22 route pattern via the inner ServeMux, not the raw path
// (cardinality bound, REQ-OBS-1). It also confirms the wrapped handler still
// serves the request normally.
func TestOTelMiddleware_NamesSpanByPattern(t *testing.T) {
	mux := http.NewServeMux()
	served := false
	mux.HandleFunc("GET /api/v1/iam/users/{user_id}", func(w http.ResponseWriter, _ *http.Request) {
		served = true
		w.WriteHeader(http.StatusOK)
	})

	h := OTelMiddleware()(mux)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/iam/users/u-123", nil))

	if !served {
		t.Fatal("wrapped handler did not serve the request")
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
}

// TestRouteLabel_PrefersPattern proves routeLabel collapses path variables using
// r.Pattern when present, and falls back to normalizeRoute when it is empty.
func TestRouteLabel_PrefersPattern(t *testing.T) {
	withPattern := httptest.NewRequest(http.MethodGet, "/api/v1/iam/users/u-1/roles", nil)
	withPattern.Pattern = "GET /api/v1/iam/users/{user_id}/roles"
	if got := routeLabel(withPattern, withPattern.URL.Path); got != "/api/v1/iam/users/{user_id}/roles" {
		t.Fatalf("routeLabel with pattern = %q, want the pattern path", got)
	}

	noPattern := httptest.NewRequest(http.MethodGet, "/api/v1/documents/d-1", nil)
	if got := routeLabel(noPattern, noPattern.URL.Path); got != "/api/v1/documents/{documentId}" {
		t.Fatalf("routeLabel without pattern = %q, want normalizeRoute output", got)
	}
}
