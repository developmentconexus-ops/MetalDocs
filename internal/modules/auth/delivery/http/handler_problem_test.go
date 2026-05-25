package httpdelivery

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	auditdomain "metaldocs/internal/modules/audit/domain"
	authdomain "metaldocs/internal/modules/auth/domain"
	"metaldocs/internal/platform/problem"
)

func TestAuthHandler_UnauthorizedIsProblemJSON(t *testing.T) {
	h := &Handler{}
	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	rec := httptest.NewRecorder()

	h.handleMe(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
	if got := rec.Header().Get("Content-Type"); got != "application/problem+json" {
		t.Fatalf("Content-Type = %q, want %q", got, "application/problem+json")
	}

	var body problem.Problem
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body.Code != "AUTH_UNAUTHORIZED" {
		t.Fatalf("code = %q, want %q", body.Code, "AUTH_UNAUTHORIZED")
	}
}

type failingAuditWriter struct{}

func (f failingAuditWriter) Record(context.Context, auditdomain.Event) error {
	return errors.New("boom")
}

func (f failingAuditWriter) RecordTx(context.Context, *sql.Tx, auditdomain.Event) error {
	return errors.New("boom")
}

func TestRecordAudit_AuditFailureDoesNotPanic(t *testing.T) {
	h := &Handler{
		audit: failingAuditWriter{},
		now:   func() time.Time { return time.Date(2026, 5, 25, 12, 0, 0, 0, time.UTC) },
	}
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	req = req.WithContext(authdomain.WithCurrentUser(req.Context(), authdomain.CurrentUser{
		UserID:   "u-1",
		TenantID: "t-1",
	}))

	h.recordAudit(req, "u-1", "auth.logout", "u-1", map[string]any{})
}
