package httpdelivery

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"metaldocs/internal/platform/problem"
)

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

func TestIAMAdminHandler_MethodNotAllowed(t *testing.T) {
	h := NewAdminHandler(nil, nil)
	t.Run("overview", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/iam/admin/overview", nil)
		rec := httptest.NewRecorder()
		h.handleAdminOverview(rec, req)
		assertMethodNotAllowedProblem(t, rec, "GET")
	})
	t.Run("role-upsert", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/iam/users/u-1/roles", nil)
		rec := httptest.NewRecorder()
		h.handleUserRoleUpsert(rec, req, "u-1")
		assertMethodNotAllowedProblem(t, rec, "POST")
	})
}

func TestIAMSessionsHandler_MethodNotAllowed(t *testing.T) {
	h := NewSessionsHandler(nil)
	t.Run("list", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/sessions", nil)
		rec := httptest.NewRecorder()
		h.handleSessions(rec, req)
		assertMethodNotAllowedProblem(t, rec, "GET")
	})
	t.Run("by-id", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/sessions/s-1", nil)
		rec := httptest.NewRecorder()
		h.handleSessionByID(rec, req)
		assertMethodNotAllowedProblem(t, rec, "DELETE")
	})
}
