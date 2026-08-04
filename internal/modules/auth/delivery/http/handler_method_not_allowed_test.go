package httpdelivery

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"metaldocs/internal/platform/problem"
)

// assertMethodNotAllowedProblem asserts the canonical RFC 9457 405 contract (D-03).
func assertMethodNotAllowedProblem(t *testing.T, rec *httptest.ResponseRecorder, wantAllow string) {
	t.Helper()
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusMethodNotAllowed)
	}
	if got := rec.Header().Get("Content-Type"); got != "application/problem+json" {
		t.Fatalf("Content-Type = %q, want application/problem+json", got)
	}
	if got := rec.Header().Get("Allow"); got != wantAllow {
		t.Fatalf("Allow = %q, want %q", got, wantAllow)
	}
	var body problem.Problem
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body.Code != problem.CodeRequestMethodNotAllowed {
		t.Fatalf("code = %q, want %q", body.Code, problem.CodeRequestMethodNotAllowed)
	}
}

func TestAuthHandler_MethodNotAllowedIsProblemJSON(t *testing.T) {
	h := &Handler{}
	cases := []struct {
		name      string
		method    string
		target    string
		fn        func(http.ResponseWriter, *http.Request)
		wantAllow string
	}{
		{"login", http.MethodGet, "/api/v1/auth/login", h.handleLogin, "POST"},
		{"logout", http.MethodGet, "/api/v1/auth/logout", h.handleLogout, "POST"},
		{"me", http.MethodPost, "/api/v1/auth/me", h.handleMe, "GET"},
		{"change-password", http.MethodGet, "/api/v1/auth/change-password", h.handleChangePassword, "POST"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.target, nil)
			rec := httptest.NewRecorder()
			tc.fn(rec, req)
			assertMethodNotAllowedProblem(t, rec, tc.wantAllow)
		})
	}
}
