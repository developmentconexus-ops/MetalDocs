package httpdelivery

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestAuthHandler_MethodNotAllowed asserts that each auth route rejects the
// wrong HTTP method with a 405 and the correct Allow header. Prior to T-008
// (task 6 of the HTTP surface protocol) these bare mux.HandleFunc patterns
// were method-less, so the handler itself had to guard the method and emit
// the RFC 9457 problem+json 405. Now that RegisterRoutes mounts through the
// generated authapi.ServerInterface router (method-qualified net/http 1.22
// patterns), the stdlib mux rejects the wrong method itself — the assertion
// class is routing, so it registers through the real RegisterRoutes on a
// real *http.ServeMux and asserts the 405 the mux produces (mirrors
// internal/modules/audit/delivery/http/handler_allow_test.go, T-008).
func TestAuthHandler_MethodNotAllowed(t *testing.T) {
	cases := []struct {
		name      string
		method    string
		target    string
		wantAllow string
	}{
		{"login", http.MethodGet, "/api/v1/auth/login", "POST"},
		{"logout", http.MethodGet, "/api/v1/auth/logout", "POST"},
		{"me", http.MethodPost, "/api/v1/auth/me", "GET, HEAD"},
		{"change-password", http.MethodGet, "/api/v1/auth/change-password", "POST"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := &Handler{}
			mux := http.NewServeMux()
			h.RegisterRoutes(mux)

			req := httptest.NewRequest(tc.method, tc.target, nil)
			rec := httptest.NewRecorder()
			mux.ServeHTTP(rec, req)

			if rec.Code != http.StatusMethodNotAllowed {
				t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusMethodNotAllowed, rec.Body.String())
			}
			if got := rec.Header().Get("Allow"); got != tc.wantAllow {
				t.Errorf("Allow = %q, want %q", got, tc.wantAllow)
			}
		})
	}
}
