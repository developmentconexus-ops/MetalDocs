package httpdelivery

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"testing"
	"time"

	authdomain "metaldocs/internal/modules/auth/domain"
	"metaldocs/internal/platform/tenant"
)

type fakeSessionAdmin struct {
	sessions map[string]authdomain.Session
	revoked  []string
}

// ListActiveSessions mirrors the real authpg.Repository behavior relevant to
// the handler's over-fetch-by-one has_more logic (security-tech-debt.md #7):
// it truncates to q.Limit when set (>0), so tests can pin has_more=true (more
// matching sessions than the requested page) and has_more=false (all
// sessions fit). Rows are sorted by SessionID for deterministic pagination in
// tests, unlike production's last_seen_at DESC ordering.
func (f *fakeSessionAdmin) ListActiveSessions(_ context.Context, q authdomain.SessionAdminQuery) ([]authdomain.SessionListItem, error) {
	out := make([]authdomain.SessionListItem, 0)
	for _, s := range f.sessions {
		if s.TenantID != q.TenantID {
			continue
		}
		out = append(out, authdomain.SessionListItem{SessionID: s.SessionID, UserID: s.UserID})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].SessionID < out[j].SessionID })
	if q.Limit > 0 && len(out) > q.Limit {
		out = out[:q.Limit]
	}
	return out, nil
}

// fakeDisplayNameReader is a fixture iamdomain.UserDisplayNameReader: DisplayNames
// returns the preloaded map (omitting any user_id not present, so the handler must
// apply its own missing->user_id fallback).
type fakeDisplayNameReader struct {
	names map[string]string
}

func (f *fakeDisplayNameReader) DisplayName(_ context.Context, _, _ string) (string, error) {
	return "", nil
}

func (f *fakeDisplayNameReader) DisplayNames(_ context.Context, _ string, userIDs []string) (map[string]string, error) {
	out := make(map[string]string, len(userIDs))
	for _, id := range userIDs {
		if n, ok := f.names[id]; ok {
			out[id] = n
		}
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
	// Caller is in tenant-b, session belongs to tenant-a.
	// MUST return 404 (not 403) to avoid leaking session existence.
	rr := httptest.NewRecorder()
	h.handleSessionByID(rr, newSessionsRequest(http.MethodDelete, "/api/v1/auth/sessions/sess-1", "tenant-b"))
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

	rr := httptest.NewRecorder()
	h.handleSessionByID(rr, newSessionsRequest(http.MethodDelete, "/api/v1/auth/sessions/sess-1", "tenant-a"))
	if rr.Code != http.StatusNoContent {
		t.Fatalf("same-tenant revoke: want 204, got %d body=%s", rr.Code, rr.Body.String())
	}
	if len(fake.revoked) != 1 || fake.revoked[0] != "sess-1" {
		t.Fatalf("expected revoke to fire, got %v", fake.revoked)
	}
}

func TestSessionsHandler_Revoke404WhenMissing(t *testing.T) {
	h := NewSessionsHandler(&fakeSessionAdmin{sessions: map[string]authdomain.Session{}}, nil)

	rr := httptest.NewRecorder()
	h.handleSessionByID(rr, newSessionsRequest(http.MethodDelete, "/api/v1/auth/sessions/missing", "tenant-a"))
	if rr.Code != http.StatusNotFound {
		t.Fatalf("missing session: want 404, got %d", rr.Code)
	}
}

func TestSessionsHandler_List401WithoutTenant(t *testing.T) {
	h := NewSessionsHandler(&fakeSessionAdmin{sessions: map[string]authdomain.Session{}}, nil)

	rr := httptest.NewRecorder()
	h.handleSessions(rr, newSessionsRequest(http.MethodGet, "/api/v1/auth/sessions", ""))
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("missing tenant: want 401, got %d", rr.Code)
	}
}

func TestSessionsHandler_ListEnrichesDisplayNameViaPort(t *testing.T) {
	fake := &fakeSessionAdmin{
		sessions: map[string]authdomain.Session{
			"sess-1": {SessionID: "sess-1", UserID: "user-1", TenantID: "tenant-a"},
			"sess-2": {SessionID: "sess-2", UserID: "user-2", TenantID: "tenant-a"},
		},
	}
	// Reader knows user-1's name, omits user-2 -> handler must fall back to user_id.
	reader := &fakeDisplayNameReader{names: map[string]string{"user-1": "Alice"}}
	h := NewSessionsHandler(fake, nil).WithDisplayNameReader(reader)

	rr := httptest.NewRecorder()
	h.handleSessions(rr, newSessionsRequest(http.MethodGet, "/api/v1/auth/sessions", "tenant-a"))
	if rr.Code != http.StatusOK {
		t.Fatalf("list: want 200, got %d body=%s", rr.Code, rr.Body.String())
	}
	body := rr.Body.String()
	if !strings.Contains(body, `"display_name":"Alice"`) {
		t.Fatalf("user-1 should render port name Alice; body=%s", body)
	}
	if !strings.Contains(body, `"display_name":"user-2"`) {
		t.Fatalf("user-2 should fall back to user_id; body=%s", body)
	}
}

// TestSessionsHandler_ListHasMoreTrueWhenExtraRowExists pins the
// over-fetch-by-one fix (security-tech-debt.md #7): with more matching
// sessions than the requested page size, the handler must report
// has_more=true and must NOT leak the extra (limit+1'th) row onto the wire.
func TestSessionsHandler_ListHasMoreTrueWhenExtraRowExists(t *testing.T) {
	sessions := make(map[string]authdomain.Session, 3)
	for i := 0; i < 3; i++ {
		id := fmt.Sprintf("sess-%d", i)
		sessions[id] = authdomain.Session{SessionID: id, UserID: fmt.Sprintf("user-%d", i), TenantID: "tenant-a"}
	}
	fake := &fakeSessionAdmin{sessions: sessions}
	h := NewSessionsHandler(fake, nil)

	rr := httptest.NewRecorder()
	h.handleSessions(rr, newSessionsRequest(http.MethodGet, "/api/v1/auth/sessions?limit=2", "tenant-a"))
	if rr.Code != http.StatusOK {
		t.Fatalf("list: want 200, got %d body=%s", rr.Code, rr.Body.String())
	}
	body := rr.Body.String()
	if !strings.Contains(body, `"has_more":true`) {
		t.Fatalf("expected has_more:true with 3 sessions and limit=2; body=%s", body)
	}
	if strings.Contains(body, "sess-2") {
		t.Fatalf("the limit+1'th row (sess-2) must be truncated, not leaked onto the wire; body=%s", body)
	}
	if got := strings.Count(body, `"session_id"`); got != 2 {
		t.Fatalf("expected exactly 2 items in the page, got %d; body=%s", got, body)
	}
}

// TestSessionsHandler_ListHasMoreFalseWhenAllSessionsFit pins the honest
// has_more=false case: when the tenant has fewer matching sessions than the
// requested page size, has_more must be false and every session must appear.
func TestSessionsHandler_ListHasMoreFalseWhenAllSessionsFit(t *testing.T) {
	fake := &fakeSessionAdmin{
		sessions: map[string]authdomain.Session{
			"sess-1": {SessionID: "sess-1", UserID: "user-1", TenantID: "tenant-a"},
			"sess-2": {SessionID: "sess-2", UserID: "user-2", TenantID: "tenant-a"},
		},
	}
	h := NewSessionsHandler(fake, nil)

	rr := httptest.NewRecorder()
	h.handleSessions(rr, newSessionsRequest(http.MethodGet, "/api/v1/auth/sessions?limit=10", "tenant-a"))
	if rr.Code != http.StatusOK {
		t.Fatalf("list: want 200, got %d body=%s", rr.Code, rr.Body.String())
	}
	body := rr.Body.String()
	if !strings.Contains(body, `"has_more":false`) {
		t.Fatalf("expected has_more:false when all sessions fit in one page; body=%s", body)
	}
	if !strings.Contains(body, "sess-1") || !strings.Contains(body, "sess-2") {
		t.Fatalf("expected both sessions present; body=%s", body)
	}
}

// TestSessionsHandler_ListDefaultLimitAppliedWhenOmitted pins that the
// default page size (50, mirroring authpg.ListActiveSessions's own default)
// is still applied server-side even though the handler now always requests
// limit+1 from the port — omitting ?limit must not change observable
// behavior for tenants with a small number of sessions.
func TestSessionsHandler_ListDefaultLimitAppliedWhenOmitted(t *testing.T) {
	fake := &fakeSessionAdmin{
		sessions: map[string]authdomain.Session{
			"sess-1": {SessionID: "sess-1", UserID: "user-1", TenantID: "tenant-a"},
		},
	}
	h := NewSessionsHandler(fake, nil)

	rr := httptest.NewRecorder()
	h.handleSessions(rr, newSessionsRequest(http.MethodGet, "/api/v1/auth/sessions", "tenant-a"))
	if rr.Code != http.StatusOK {
		t.Fatalf("list: want 200, got %d body=%s", rr.Code, rr.Body.String())
	}
	body := rr.Body.String()
	if !strings.Contains(body, `"has_more":false`) {
		t.Fatalf("expected has_more:false for a single session with default limit; body=%s", body)
	}
	if !strings.Contains(body, "sess-1") {
		t.Fatalf("expected sess-1 present; body=%s", body)
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

	rr := httptest.NewRecorder()
	h.handleSessions(rr, newSessionsRequest(http.MethodGet, "/api/v1/auth/sessions", "tenant-a"))
	body := rr.Body.String()
	if !strings.Contains(body, "sess-a") || strings.Contains(body, "sess-b") {
		t.Fatalf("tenant isolation broken; body=%s", body)
	}
}
