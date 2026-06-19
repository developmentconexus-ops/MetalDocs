package httpdelivery

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	auditdomain "metaldocs/internal/modules/audit/domain"
	authdomain "metaldocs/internal/modules/auth/domain"
	iamapi "metaldocs/internal/modules/iam/api"
	iamapp "metaldocs/internal/modules/iam/application"
	iamdomain "metaldocs/internal/modules/iam/domain"
	iampresence "metaldocs/internal/modules/iam/presence"
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

func (stubRoleAdminRepository) UpsertUserAndAssignRoleTx(_ context.Context, _ *sql.Tx, _, _, _ string, _ iamdomain.Role, _ string) error {
	return nil
}

func (stubRoleAdminRepository) ReplaceUserRolesTx(_ context.Context, _ *sql.Tx, _, _, _ string, _ iamdomain.Role, _ string) error {
	return nil
}

type stubUserAdminService struct {
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
	handler := NewAdminHandler(iamapp.NewAdminService(stubRoleAdminRepository{}, nil, nil, nil), nil)

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

func (s *stubAuditLister) ListEvents(_ context.Context, q auditdomain.ListEventsQuery) ([]auditdomain.Event, bool, error) {
	s.tenantID = q.TenantID
	if s.delay > 0 {
		time.Sleep(s.delay)
	}
	return s.events, false, nil
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
	if _, ok := body["recent_activities"]; !ok {
		t.Fatalf("response missing recent_activities: %v", body)
	}
	kpi := body["kpi"].(map[string]any)
	if kpi["locked_accounts"] != float64(3) {
		t.Fatalf("kpi.locked_accounts = %v, want 3", kpi["locked_accounts"])
	}
}

type stubPresenceReader struct {
	tenantID string
	items    []iampresence.Item
}

func (s *stubPresenceReader) Snapshot(_ context.Context, tenantID string, _ time.Time) ([]iampresence.Item, error) {
	s.tenantID = tenantID
	return s.items, nil
}

// TestHandleAdminOverview_PresenceCarriesStatus is the F2.2 contract-emit
// proof: when the presence reader is wired, each presence item emits a
// `status` field carrying the online/idle enum value declared on
// OnlinePresenceItem. Guards the contract↔emit agreement that the live
// runtime could not show (no online users → empty presence[]).
func TestHandleAdminOverview_PresenceCarriesStatus(t *testing.T) {
	authSvc := &stubUserAdminService{}
	presence := &stubPresenceReader{items: []iampresence.Item{{
		UserID:      "u1",
		Username:    "alice",
		DisplayName: "Alice",
		LastSeenAt:  time.Unix(1700000000, 0).UTC(),
		Status:      iampresence.StatusOnline,
	}}}
	handler := NewAdminHandler(nil, authSvc).
		WithObservabilityService(&stubKpiReader{}).
		WithAuditEventLister(&stubAuditLister{}).
		WithPresenceReader(presence)

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
	items, ok := body["presence"].([]any)
	if !ok || len(items) != 1 {
		t.Fatalf("presence = %v, want 1 item", body["presence"])
	}
	item := items[0].(map[string]any)
	if item["status"] != "online" {
		t.Fatalf("presence[0].status = %v, want \"online\" (the declared OnlinePresenceItem enum)", item["status"])
	}
}

// TestHandleAdminOverview_DecodesIntoGeneratedContract is the F5.5 (Major #4)
// guard: the overview route body must decode strictly into the strict-server
// generated iamapi.AdminOverviewResponse — no unknown fields, every key typed.
// Locks the route to the OpenAPI contract so future drift fails the suite.
func TestHandleAdminOverview_DecodesIntoGeneratedContract(t *testing.T) {
	authSvc := &stubUserAdminService{}
	presence := &stubPresenceReader{items: []iampresence.Item{{
		UserID:      "u1",
		Username:    "alice",
		DisplayName: "Alice",
		LastSeenAt:  time.Unix(1700000000, 0).UTC(),
		Status:      iampresence.StatusOnline,
	}}}
	handler := NewAdminHandler(nil, authSvc).
		WithObservabilityService(&stubKpiReader{snap: iamdomain.KpiSnapshot{
			LockedAccounts:   3,
			FailedLogins24h:  7,
			RoleDistribution: []iamdomain.RoleCount{{Role: iamdomain.RoleEditor, Count: 2}},
		}}).
		WithAuditEventLister(&stubAuditLister{events: []auditdomain.Event{{
			ID: "evt_1", TenantID: tenantA, Action: "login", ResourceType: "user", ResourceID: "u1",
		}}}).
		WithPresenceReader(presence)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/iam/admin/overview", nil)
	req = req.WithContext(tenant.WithTenantID(req.Context(), tenantA))
	rec := httptest.NewRecorder()
	handler.handleAdminOverview(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}

	dec := json.NewDecoder(bytes.NewReader(rec.Body.Bytes()))
	dec.DisallowUnknownFields()
	var out iamapi.AdminOverviewResponse
	if err := dec.Decode(&out); err != nil {
		t.Fatalf("body does not conform to iamapi.AdminOverviewResponse: %v\nbody=%s", err, rec.Body.String())
	}

	if out.Kpi.LockedAccounts != 3 {
		t.Fatalf("kpi.locked_accounts = %d, want 3", out.Kpi.LockedAccounts)
	}
	if out.Kpi.FailedLogins24h != 7 {
		t.Fatalf("kpi.failed_logins24h = %d, want 7", out.Kpi.FailedLogins24h)
	}
	if len(out.Kpi.RoleDistribution) != 1 || string(out.Kpi.RoleDistribution[0].Role) != "editor" {
		t.Fatalf("kpi.role_distribution = %+v, want [{editor 2}]", out.Kpi.RoleDistribution)
	}
	if len(out.Presence) != 1 || out.Presence[0].Status == nil || *out.Presence[0].Status != iamapi.Online {
		t.Fatalf("presence = %+v, want 1 item with status=online", out.Presence)
	}
	if len(out.RecentActivities) != 1 || out.RecentActivities[0].Action != "login" {
		t.Fatalf("recent_activities = %+v, want 1 item action=login", out.RecentActivities)
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
