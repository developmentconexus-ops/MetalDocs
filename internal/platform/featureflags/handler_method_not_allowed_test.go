package featureflags_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"metaldocs/internal/platform/config"
	"metaldocs/internal/platform/featureflags"
)

// TestHandler_MethodNotAllowed asserts that GET /api/v1/feature-flags
// rejects the wrong HTTP method with a 405 and the correct Allow header.
// Prior to Task 10 (HTTP surface protocol) this was a bare, method-less
// mux.HandleFunc pattern, so the handler itself had to guard the method and
// emit the RFC 9457 problem+json 405. Now that Mount mounts
// through the generated featureflagsapi.ServerInterface router
// (method-qualified net/http 1.22 patterns), the stdlib mux rejects the
// wrong method itself -- the assertion class is routing, so it registers
// through the real Mount on a real *http.ServeMux and asserts the
// 405 the mux produces (mirrors internal/modules/auth's
// TestAuthHandler_MethodNotAllowed, Task 6).
func TestHandler_MethodNotAllowed(t *testing.T) {
	h := featureflags.NewHandler(config.FeatureFlagsConfig{})
	mux := http.NewServeMux()
	h.Mount(mux)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/feature-flags", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d, body=%s", rr.Code, http.StatusMethodNotAllowed, rr.Body.String())
	}
	if got := rr.Header().Get("Allow"); got != "GET, HEAD" {
		t.Fatalf("Allow = %q, want %q", got, "GET, HEAD")
	}
}
