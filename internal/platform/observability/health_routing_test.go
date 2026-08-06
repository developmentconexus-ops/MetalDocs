package observability_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"metaldocs/internal/platform/observability"
)

func mountHealth(t *testing.T) *http.ServeMux {
	t.Helper()
	mux := http.NewServeMux()
	observability.NewHealthHandler(nil, nil).RegisterRoutes(mux)
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
