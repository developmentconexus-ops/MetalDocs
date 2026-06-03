package httpdelivery

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	auditdomain "metaldocs/internal/modules/audit/domain"
	authdomain "metaldocs/internal/modules/auth/domain"
	iamapp "metaldocs/internal/modules/iam/application"
	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/platform/tenant"
)

type stubRoleAdminRepository struct{}

func (stubRoleAdminRepository) HasAnyRole(context.Context, iamdomain.Role, string) (bool, error) {
	return false, nil
}

func (stubRoleAdminRepository) UpsertUserAndAssignRole(context.Context, string, string, string, iamdomain.Role, string) error {
	return nil
}

func (stubRoleAdminRepository) ReplaceUserRoles(context.Context, string, string, string, iamdomain.Role, string) error {
	return nil
}

type stubUserAdminService struct {
	listUsersTenantID       string
	listOnlineUsersTenantID string
}

func (s *stubUserAdminService) ListUsers(context.Context, string) ([]authdomain.ManagedUser, error) {
	return nil, nil
}

func (s *stubUserAdminService) ListOnlineUsers(_ context.Context, tenantID string, _ time.Time) ([]authdomain.OnlineUser, error) {
	s.listOnlineUsersTenantID = tenantID
	return nil, nil
}

func (s *stubUserAdminService) CreateUser(context.Context, string, string, string, string, string, string, []iamdomain.Role, string) error {
	return nil
}

func (s *stubUserAdminService) UpdateUser(context.Context, authdomain.UpdateUserParams, string) error {
	return nil
}

func (s *stubUserAdminService) AdminResetPassword(context.Context, string, string) error {
	return nil
}

func (s *stubUserAdminService) UnlockUser(context.Context, string) error {
	return nil
}

func TestHandleReplaceUserRoles_RejectsMultipleRoles(t *testing.T) {
	handler := NewAdminHandler(iamapp.NewAdminService(stubRoleAdminRepository{}, nil), nil)

	body := bytes.NewBufferString(`{"displayName":"Alice","roles":["editor","viewer"]}`)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/iam/users/alice/roles", body)
	req = req.WithContext(tenant.WithTenantID(req.Context(), "11111111-1111-1111-1111-111111111111"))
	rec := httptest.NewRecorder()

	handler.handleReplaceUserRoles(rec, req, "alice")

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
	if got := rec.Body.String(); !bytes.Contains([]byte(got), []byte("Exactly one role is required")) {
		t.Fatalf("response body %q does not mention single-role requirement", got)
	}
}

// TestHandleCreateUser_RejectsMultipleRoles was removed by PR-4: the legacy
// POST /iam/users endpoint and its CreateUserRequest type were replaced by
// POST /iam/users/invite (UserInviteRequest), which carries a single
// tenantRole field instead of a multi-role array — so a "rejects multiple
// roles" assertion is no longer expressible at this layer. Equivalent
// canonical-role-only coverage lives in tests/unit/iam_people/.

func TestHandleAdminOverview_PassesTenantIDToOnlineUsers(t *testing.T) {
	authSvc := &stubUserAdminService{}
	handler := NewAdminHandler(nil, authSvc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/iam/admin/overview", nil)
	req = req.WithContext(tenant.WithTenantID(req.Context(), "11111111-1111-1111-1111-111111111111"))
	rec := httptest.NewRecorder()

	handler.handleAdminOverview(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if authSvc.listOnlineUsersTenantID != "11111111-1111-1111-1111-111111111111" {
		t.Fatalf("ListOnlineUsers tenant = %q", authSvc.listOnlineUsersTenantID)
	}
}

// --- PR-8 overview composition tests ----------------------------------------

type stubKpiReader struct {
	tenantID string
	delay    time.Duration
	snap     iamdomain.KpiSnapshot
}

func (s *stubKpiReader) GetKpi(_ context.Context, tenantID string) (iamdomain.KpiSnapshot, error) {
	s.tenantID = tenantID
	if s.delay > 0 {
		time.Sleep(s.delay)
	}
	return s.snap, nil
}

type stubAuditLister struct {
	tenantID string
	delay    time.Duration
	events   []auditdomain.Event
}

func (s *stubAuditLister) ListEvents(_ context.Context, q auditdomain.ListEventsQuery) ([]auditdomain.Event, error) {
	s.tenantID = q.TenantID
	if s.delay > 0 {
		time.Sleep(s.delay)
	}
	return s.events, nil
}

type delayedOnlineService struct {
	stubUserAdminService
	delay time.Duration
}

func (s *delayedOnlineService) ListOnlineUsers(ctx context.Context, tenantID string, _ time.Time) ([]authdomain.OnlineUser, error) {
	s.listOnlineUsersTenantID = tenantID
	if s.delay > 0 {
		time.Sleep(s.delay)
	}
	return nil, nil
}

func TestHandleAdminOverview_DropsUsersField_ReturnsTypedShape(t *testing.T) {
	authSvc := &stubUserAdminService{}
	handler := NewAdminHandler(nil, authSvc).
		WithObservabilityService(&stubKpiReader{snap: iamdomain.KpiSnapshot{LockedAccounts: 3}}).
		WithAuditEventLister(&stubAuditLister{events: []auditdomain.Event{{ID: "evt_1", TenantID: tenantA, Action: "x", ResourceType: "user", ResourceID: "u1"}}})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/iam/admin/overview", nil)
	req = req.WithContext(tenant.WithTenantID(req.Context(), tenantA))
	rec := httptest.NewRecorder()
	handler.handleAdminOverview(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if _, ok := body["users"]; ok {
		t.Fatalf("response still carries legacy users[] field: %v", body)
	}
	if _, ok := body["kpi"]; !ok {
		t.Fatalf("response missing kpi: %v", body)
	}
	if _, ok := body["presence"]; !ok {
		t.Fatalf("response missing presence: %v", body)
	}
	if _, ok := body["recentActivities"]; !ok {
		t.Fatalf("response missing recentActivities: %v", body)
	}
	kpi := body["kpi"].(map[string]any)
	if kpi["lockedAccounts"] != float64(3) {
		t.Fatalf("kpi.lockedAccounts = %v, want 3", kpi["lockedAccounts"])
	}
}

func TestHandleAdminOverview_TenantIsolation(t *testing.T) {
	authSvc := &stubUserAdminService{}
	kpi := &stubKpiReader{}
	audit := &stubAuditLister{}
	handler := NewAdminHandler(nil, authSvc).
		WithObservabilityService(kpi).
		WithAuditEventLister(audit)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/iam/admin/overview", nil)
	req = req.WithContext(tenant.WithTenantID(req.Context(), tenantB))
	rec := httptest.NewRecorder()
	handler.handleAdminOverview(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	if kpi.tenantID != tenantB {
		t.Fatalf("kpi tenant = %q, want %q", kpi.tenantID, tenantB)
	}
	if audit.tenantID != tenantB {
		t.Fatalf("audit tenant = %q, want %q", audit.tenantID, tenantB)
	}
	if authSvc.listOnlineUsersTenantID != tenantB {
		t.Fatalf("online users tenant = %q, want %q", authSvc.listOnlineUsersTenantID, tenantB)
	}
}

// TestHandleAdminOverview_RunsInParallel asserts the three composition reads
// run concurrently. Each downstream sleeps 80ms; sequential = 240ms,
// parallel = ~80ms. A 200ms ceiling leaves slack for slow CI without
// allowing a regression to a sequential composition.
func TestHandleAdminOverview_RunsInParallel(t *testing.T) {
	const delay = 80 * time.Millisecond
	authSvc := &delayedOnlineService{delay: delay}
	kpi := &stubKpiReader{delay: delay}
	audit := &stubAuditLister{delay: delay}
	handler := NewAdminHandler(nil, authSvc).
		WithObservabilityService(kpi).
		WithAuditEventLister(audit)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/iam/admin/overview", nil)
	req = req.WithContext(tenant.WithTenantID(req.Context(), tenantA))
	rec := httptest.NewRecorder()

	start := time.Now()
	handler.handleAdminOverview(rec, req)
	elapsed := time.Since(start)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	// Sequential composition would take 3×80ms ≈ 240ms. Parallel runs
	// max(80ms) plus a small overhead. 200ms ceiling catches regressions
	// without flaking on slow CI.
	if elapsed > 200*time.Millisecond {
		t.Fatalf("overview composition took %v, expected parallel (<200ms with 3×80ms downstreams)", elapsed)
	}
}

const tenantA = "11111111-1111-1111-1111-111111111111"
const tenantB = "22222222-2222-2222-2222-222222222222"
