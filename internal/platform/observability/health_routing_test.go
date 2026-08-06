package observability_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"metaldocs/internal/platform/observability"
)

// mountHealth mounts both the health and metrics publishers (Task 15a,
// Ruling 1 split them into two Go objects mounting two generated packages)
// onto one mux, matching what buildRouter does in production -- the mounted
// patterns are unchanged, only which Go object registers them moved.
func mountHealth(t *testing.T) *http.ServeMux {
	t.Helper()
	mux := http.NewServeMux()
	observability.NewHealthHandler(nil).Mount(mux)
	observability.NewMetricsHandler(nil).Mount(mux)
	return mux
}

// §10.3 row 1: authenticated non-GET on a health path was 200 (the handlers
// checked no method at all); the method-qualified pattern makes the mux answer
// 405. Deliberate tightening, pinned so it cannot regress silently.
func TestHealthRejectsNonGET(t *testing.T) {
	mux := mountHealth(t)
	for _, path := range []string{"/api/v1/health/live", "/api/v1/health/ready"} {
		for _, method := range []string{http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch} {
			rr := httptest.NewRecorder()
			mux.ServeHTTP(rr, httptest.NewRequest(method, path, nil))
			if rr.Code != http.StatusMethodNotAllowed {
				t.Errorf("%s %s = %d, want 405", method, path, rr.Code)
			}
		}
	}
}

// §10.3 / Task 11 review finding: mounting /api/v1/metrics through
// HandlerWithOptions registers the method-qualified pattern
// "GET /api/v1/metrics" on the mux. Go's stdlib mux intercepts a non-GET/HEAD
// request to a method-qualified pattern BEFORE the handler runs, so
// MetricsHandler's own in-handler 405 (http.go:211-216) never fires for the
// mounted route. The status code stays 405, but the body/Content-Type/Allow
// header now come from the stdlib mux's own rejection, not from the
// handler's httpresponse.WriteMethodNotAllowed problem+json output. Pinned
// here so that delta cannot regress or drift silently; health.go's
// NewMetricsHandler(nil) never invokes h.metrics because the mux never
// reaches GetMetrics for these requests.
func TestMetricsRejectsNonGET(t *testing.T) {
	mux := mountHealth(t)
	for _, method := range []string{http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch} {
		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, httptest.NewRequest(method, "/api/v1/metrics", nil))
		if rr.Code != http.StatusMethodNotAllowed {
			t.Errorf("%s /api/v1/metrics = %d, want 405", method, rr.Code)
		}
		if got := rr.Header().Get("Content-Type"); got != "text/plain; charset=utf-8" {
			t.Errorf("%s /api/v1/metrics Content-Type = %q, want stdlib mux's text/plain; charset=utf-8 (not application/problem+json)", method, got)
		}
		if got := rr.Header().Get("Allow"); got != "GET, HEAD" {
			t.Errorf("%s /api/v1/metrics Allow = %q, want %q", method, got, "GET, HEAD")
		}
		if got := rr.Body.String(); got != "Method Not Allowed\n" {
			t.Errorf("%s /api/v1/metrics body = %q, want stdlib mux's plain-text %q (not the problem+json body)", method, got, "Method Not Allowed\n")
		}
	}
}

func TestHealthGETStillServes(t *testing.T) {
	mux := mountHealth(t)
	for _, path := range []string{"/api/v1/health/live", "/api/v1/health/ready"} {
		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, path, nil))
		if rr.Code != http.StatusOK {
			t.Errorf("GET %s = %d, want 200", path, rr.Code)
		}
	}
}

// §10.3 row 3: /healthz is deleted, not excepted (operator ruling C).
func TestHealthzIsGone(t *testing.T) {
	mux := mountHealth(t)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	if rr.Code != http.StatusNotFound {
		t.Fatalf("GET /healthz = %d, want 404 from the mux (authn turns this into 401 in the real chain)", rr.Code)
	}
}
