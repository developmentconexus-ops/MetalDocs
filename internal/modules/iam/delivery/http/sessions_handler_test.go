package httpdelivery

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	authdomain "metaldocs/internal/modules/auth/domain"
	authpg "metaldocs/internal/modules/auth/infrastructure/postgres"
	"metaldocs/internal/platform/tenant"
)

type fakeSessionAdmin struct {
	sessions map[string]authdomain.Session
	revoked  []string
}

func (f *fakeSessionAdmin) ListActiveSessions(_ context.Context, q authpg.SessionAdminQuery) ([]authpg.SessionListItem, error) {
	out := make([]authpg.SessionListItem, 0)
	for _, s := range f.sessions {
		if s.TenantID != q.TenantID {
			continue
		}
		out = append(out, authpg.SessionListItem{SessionID: s.SessionID, UserID: s.UserID, DisplayName: s.UserID})
	}
	return out, nil
}
func (f *fakeSessionAdmin) FindSession(_ context.Context, sessionID string) (authdomain.Session, error) {
	s, ok := f.sessions[sessionID]
	if !ok {
		return authdomain.Session{}, authdomain.ErrSessionNotFound
	}
	return s, nil
}
func (f *fakeSessionAdmin) RevokeSession(_ context.Context, sessionID string, _ time.Time) error {
	if _, ok := f.sessions[sessionID]; !ok {
		return authdomain.ErrSessionNotFound
	}
	f.revoked = append(f.revoked, sessionID)
	return nil
}
func (f *fakeSessionAdmin) RevokeSessionsByUserID(_ context.Context, _ string, _ time.Time) error {
	return errors.New("not used in this test")
}

func newSessionsRequest(method, path, tenantID string) *http.Request {
	r := httptest.NewRequest(method, path, nil)
	if tenantID != "" {
		r = r.WithContext(tenant.WithTenantID(r.Context(), tenantID))
	}
	return r
}

func TestSessionsHandler_RevokeReturns404WhenCrossTenant(t *testing.T) {
	fake := &fakeSessionAdmin{
		sessions: map[string]authdomain.Session{
			"sess-1": {SessionID: "sess-1", UserID: "user-1", TenantID: "tenant-a"},
		},
	}
	h := NewSessionsHandler(fake, nil)
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	// Caller is in tenant-b, session belongs to tenant-a.
	// MUST return 404 (not 403) to avoid leaking session existence.
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, newSessionsRequest(http.MethodDelete, "/api/v1/auth/sessions/sess-1", "tenant-b"))
	if rr.Code != http.StatusNotFound {
		t.Fatalf("cross-tenant revoke: want 404, got %d body=%s", rr.Code, rr.Body.String())
	}
	if len(fake.revoked) != 0 {
		t.Fatalf("cross-tenant revoke must NOT touch the session, got revoked=%v", fake.revoked)
	}
}

func TestSessionsHandler_Revoke204WhenSameTenant(t *testing.T) {
	fake := &fakeSessionAdmin{
		sessions: map[string]authdomain.Session{
			"sess-1": {SessionID: "sess-1", UserID: "user-1", TenantID: "tenant-a"},
		},
	}
	h := NewSessionsHandler(fake, nil)
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, newSessionsRequest(http.MethodDelete, "/api/v1/auth/sessions/sess-1", "tenant-a"))
	if rr.Code != http.StatusNoContent {
		t.Fatalf("same-tenant revoke: want 204, got %d body=%s", rr.Code, rr.Body.String())
	}
	if len(fake.revoked) != 1 || fake.revoked[0] != "sess-1" {
		t.Fatalf("expected revoke to fire, got %v", fake.revoked)
	}
}

func TestSessionsHandler_Revoke404WhenMissing(t *testing.T) {
	h := NewSessionsHandler(&fakeSessionAdmin{sessions: map[string]authdomain.Session{}}, nil)
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, newSessionsRequest(http.MethodDelete, "/api/v1/auth/sessions/missing", "tenant-a"))
	if rr.Code != http.StatusNotFound {
		t.Fatalf("missing session: want 404, got %d", rr.Code)
	}
}

func TestSessionsHandler_List401WithoutTenant(t *testing.T) {
	h := NewSessionsHandler(&fakeSessionAdmin{sessions: map[string]authdomain.Session{}}, nil)
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, newSessionsRequest(http.MethodGet, "/api/v1/auth/sessions", ""))
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("missing tenant: want 401, got %d", rr.Code)
	}
}

func TestSessionsHandler_ListOnlyOwnTenant(t *testing.T) {
	fake := &fakeSessionAdmin{
		sessions: map[string]authdomain.Session{
			"sess-a": {SessionID: "sess-a", UserID: "user-1", TenantID: "tenant-a"},
			"sess-b": {SessionID: "sess-b", UserID: "user-2", TenantID: "tenant-b"},
		},
	}
	h := NewSessionsHandler(fake, nil)
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, newSessionsRequest(http.MethodGet, "/api/v1/auth/sessions", "tenant-a"))
	body := rr.Body.String()
	if !strings.Contains(body, "sess-a") || strings.Contains(body, "sess-b") {
		t.Fatalf("tenant isolation broken; body=%s", body)
	}
}
