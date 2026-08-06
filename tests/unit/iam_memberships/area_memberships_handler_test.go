// area_memberships_handler_test.go — PR-1 (IAM area-memberships rebuild)
// handler coverage.
//
// Lives in its own sub-package (tests/unit/iam_memberships) mirroring the PR-7b
// tests/unit/iam_people layout — exercises the actual MembershipHandler via
// httptest against in-memory dependencies (no DB, no codegen server stubs).
//
//	go test ./tests/unit/iam_memberships/...
//
// Surface covered (matches OpenAPI spec at api/openapi/v1/openapi.yaml):
//   - listAreaMemberships  — response shape contract
//   - grantAreaMembership  — audit emission, 404 cross-tenant guard, 409 duplicate
//   - revokeAreaMembership — 404 cross-tenant guard, audit emission
package iammembershipstest

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"io"

	auditdomain "metaldocs/internal/modules/audit/domain"
	iamapp "metaldocs/internal/modules/iam/application"
	iamdelivery "metaldocs/internal/modules/iam/delivery/http"
	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/platform/db"
	"metaldocs/internal/platform/tenant"
)

const (
	tenantAlpha = "11111111-1111-1111-1111-111111111111"
	tenantBeta  = "22222222-2222-2222-2222-222222222222"
	adminID     = "admin-1"
	targetID    = "alice"
)

// ─── fakes ───────────────────────────────────────────────────────────────

// noopMemTx satisfies iamdomain.MembershipTx for the in-memory repo.
type noopMemTx struct{}

func (noopMemTx) ExecContext(_ context.Context, _ string, _ ...any) (sql.Result, error) {
	return nil, nil
}
func (noopMemTx) QueryContext(_ context.Context, _ string, _ ...any) (*sql.Rows, error) {
	return nil, nil
}
func (noopMemTx) QueryRowContext(_ context.Context, _ string, _ ...any) *sql.Row { return nil }
func (noopMemTx) Commit() error                                                  { return nil }
func (noopMemTx) Rollback() error                                                { return nil }

// memAreaRepo is an in-memory UserAreaWriteRepository. Tracks active rows
// keyed by (userID, tenantID, areaCode); revoked rows are dropped (the
// service only ever lists active anyway).
type memAreaRepo struct {
	mu     sync.Mutex
	active map[string]iamdomain.UserProcessArea
	// directory-scope config (ADR 0022 Phase 4). The real repo derives these
	// from system_admin roles + the role_capabilities↔user_process_areas join;
	// the fake lets each test declare the actor's scope directly.
	tenantWide   map[string]bool     // actorID → full tenant directory
	managedAreas map[string][]string // actorID → areas where actor holds the cap
}

func newMemAreaRepo() *memAreaRepo {
	return &memAreaRepo{
		active:       map[string]iamdomain.UserProcessArea{},
		tenantWide:   map[string]bool{},
		managedAreas: map[string][]string{},
	}
}

func memKey(userID, tenantID, areaCode string) string {
	return userID + "|" + tenantID + "|" + areaCode
}

func (r *memAreaRepo) ListActive(_ context.Context, userID, tenantID string, _ time.Time) ([]iamdomain.UserProcessArea, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]iamdomain.UserProcessArea, 0)
	for _, m := range r.active {
		if m.UserID == userID && m.TenantID == tenantID {
			out = append(out, m)
		}
	}
	return out, nil
}

func (r *memAreaRepo) ListActiveForUsers(_ context.Context, tenantID string, userIDs []string, _ time.Time) (map[string][]iamdomain.UserProcessArea, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	want := make(map[string]struct{}, len(userIDs))
	for _, uid := range userIDs {
		want[uid] = struct{}{}
	}
	out := make(map[string][]iamdomain.UserProcessArea, len(userIDs))
	for _, m := range r.active {
		if m.TenantID != tenantID {
			continue
		}
		if _, ok := want[m.UserID]; !ok {
			continue
		}
		out[m.UserID] = append(out[m.UserID], m)
	}
	return out, nil
}

func (r *memAreaRepo) ListByTenant(_ context.Context, tenantID, userID, areaCode, role string, _ time.Time) ([]iamdomain.UserProcessArea, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]iamdomain.UserProcessArea, 0)
	for _, m := range r.active {
		if m.TenantID != tenantID {
			continue
		}
		if userID != "" && m.UserID != userID {
			continue
		}
		if areaCode != "" && m.AreaCode != areaCode {
			continue
		}
		if role != "" && string(m.Role) != role {
			continue
		}
		out = append(out, m)
	}
	return out, nil
}

func (r *memAreaRepo) MembershipDirectoryScope(_ context.Context, _, actorID, _ string, _ time.Time) (bool, bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.tenantWide[actorID], len(r.managedAreas[actorID]) > 0, nil
}

func (r *memAreaRepo) ListByTenantInManagedAreas(_ context.Context, tenantID, userID, areaCode, role, actorID, _ string, _ time.Time) ([]iamdomain.UserProcessArea, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	managed := map[string]struct{}{}
	for _, a := range r.managedAreas[actorID] {
		managed[a] = struct{}{}
	}
	out := make([]iamdomain.UserProcessArea, 0)
	for _, m := range r.active {
		if m.TenantID != tenantID {
			continue
		}
		if _, ok := managed[m.AreaCode]; !ok {
			continue
		}
		if userID != "" && m.UserID != userID {
			continue
		}
		if areaCode != "" && m.AreaCode != areaCode {
			continue
		}
		if role != "" && string(m.Role) != role {
			continue
		}
		out = append(out, m)
	}
	return out, nil
}

func (r *memAreaRepo) Insert(_ context.Context, m iamdomain.UserProcessArea) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.active[memKey(m.UserID, m.TenantID, m.AreaCode)] = m
	return nil
}

func (r *memAreaRepo) CloseActive(_ context.Context, userID, tenantID, areaCode string, _ time.Time, _ string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.active, memKey(userID, tenantID, areaCode))
	return nil
}

func (r *memAreaRepo) GrantAtomic(_ context.Context, oldM, newM iamdomain.UserProcessArea) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.active, memKey(oldM.UserID, oldM.TenantID, oldM.AreaCode))
	r.active[memKey(newM.UserID, newM.TenantID, newM.AreaCode)] = newM
	return nil
}

func (r *memAreaRepo) BeginTx(_ context.Context) (iamdomain.MembershipTx, error) {
	return noopMemTx{}, nil
}

func (r *memAreaRepo) InsertTx(ctx context.Context, _ iamdomain.MembershipTx, m iamdomain.UserProcessArea) error {
	return r.Insert(ctx, m)
}

func (r *memAreaRepo) CloseActiveTx(ctx context.Context, _ iamdomain.MembershipTx, userID, tenantID, areaCode string, effectiveTo time.Time, actorID string) error {
	return r.CloseActive(ctx, userID, tenantID, areaCode, effectiveTo, actorID)
}

func (r *memAreaRepo) GrantAtomicTx(ctx context.Context, _ iamdomain.MembershipTx, oldM, newM iamdomain.UserProcessArea) error {
	return r.GrantAtomic(ctx, oldM, newM)
}

func (r *memAreaRepo) GetActiveByUserAndArea(_ context.Context, userID, tenantID, areaCode string, _ time.Time) (*iamdomain.UserProcessArea, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if m, ok := r.active[memKey(userID, tenantID, areaCode)]; ok {
		return &m, nil
	}
	return nil, nil
}

// tenantScopedVerifier returns ErrUserNotInTenant when the (tenantID, userID)
// pair is not in the seeded set — mirrors PeopleService.VerifyUserInTenant.
type tenantScopedVerifier struct {
	in map[string]struct{} // key: tenantID|userID
}

func newTenantScopedVerifier() *tenantScopedVerifier {
	return &tenantScopedVerifier{in: map[string]struct{}{}}
}

func (v *tenantScopedVerifier) seed(tenantID, userID string) {
	v.in[tenantID+"|"+userID] = struct{}{}
}

func (v *tenantScopedVerifier) VerifyUserInTenant(_ context.Context, tenantID, userID string) error {
	if _, ok := v.in[tenantID+"|"+userID]; ok {
		return nil
	}
	return iamapp.ErrUserNotInTenant
}

// recordingAudit captures every audit Event for assertion.
type recordingAudit struct {
	mu     sync.Mutex
	events []auditdomain.Event
}

func (a *recordingAudit) Record(_ context.Context, ev auditdomain.Event) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.events = append(a.events, ev)
	return nil
}

func (a *recordingAudit) RecordTx(_ context.Context, _ db.Tx, ev auditdomain.Event) error {
	return a.Record(context.Background(), ev)
}

func (a *recordingAudit) snapshot() []auditdomain.Event {
	a.mu.Lock()
	defer a.mu.Unlock()
	out := make([]auditdomain.Event, len(a.events))
	copy(out, a.events)
	return out
}

// ─── harness ─────────────────────────────────────────────────────────────

type harness struct {
	mux      *http.ServeMux
	repo     *memAreaRepo
	verifier *tenantScopedVerifier
	audit    *recordingAudit
}

func newHarness(t *testing.T) *harness {
	t.Helper()
	repo := newMemAreaRepo()
	// adminID is the system_admin actor used by adminReq — grant it tenant-wide
	// directory scope (ADR 0022 Phase 4; the real repo derives this from the
	// system_admin role bypass).
	repo.tenantWide[adminID] = true
	verifier := newTenantScopedVerifier()
	audit := &recordingAudit{}
	// Audit is now written in-tx by the service's AuditMembershipLogger (H-3a);
	// the handler no longer emits a post-commit row. recordingAudit.RecordTx
	// appends to the same events slice, so the assertions below are unchanged.
	svc := iamapp.NewAreaMembershipService(repo, iamapp.NewAuditMembershipLogger(audit))
	h := iamdelivery.NewMembershipHandler(svc, verifier)
	// MembershipHandler.RegisterRoutes was deleted as orphaned dead code
	// (Task 18 step 3); re-pointed to the real production mount, Router.Mount.
	mux := http.NewServeMux()
	iamdelivery.NewRouter(nil, nil, h, nil, nil, nil, nil).Mount(mux)
	return &harness{mux: mux, repo: repo, verifier: verifier, audit: audit}
}

// adminReq seeds tenant + system_admin actor context on the request so the
// handler permits tenant-wide directory access and non-self grant/revoke for
// any target user (system_admin bypasses tier-2 area scoping at the repo layer).
func adminReq(method, url string, body string, tenantID, actorID string) *http.Request {
	var r io.Reader
	if body != "" {
		r = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, url, r)
	ctx := tenant.WithTenantID(req.Context(), tenantID)
	ctx = iamdomain.WithAuthContext(ctx, actorID, []iamdomain.Role{iamdomain.RoleSystemAdmin})
	return req.WithContext(ctx)
}

// ─── tests ───────────────────────────────────────────────────────────────

func TestListAreaMemberships_ContractShape(t *testing.T) {
	h := newHarness(t)
	h.verifier.seed(tenantAlpha, targetID)
	h.repo.Insert(context.Background(), iamdomain.UserProcessArea{
		UserID:        targetID,
		TenantID:      tenantAlpha,
		AreaCode:      "QMS",
		Role:          iamdomain.RoleAuthor,
		EffectiveFrom: time.Now().UTC().Add(-time.Hour),
	})

	req := adminReq(http.MethodGet, "/api/v1/iam/area-memberships?user_id="+targetID, "", tenantAlpha, adminID)
	rec := httptest.NewRecorder()
	h.mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s want 200", rec.Code, rec.Body.String())
	}
	var resp struct {
		Items []map[string]any `json:"items"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v body=%s", err, rec.Body.String())
	}
	if len(resp.Items) != 1 {
		t.Fatalf("items=%d want 1; body=%s", len(resp.Items), rec.Body.String())
	}
	row := resp.Items[0]
	for _, key := range []string{"user_id", "tenant_id", "area_code", "role", "effective_from"} {
		if _, ok := row[key]; !ok {
			t.Errorf("row missing %q (snake_case contract violation); row=%v", key, row)
		}
	}
	if row["user_id"] != targetID || row["tenant_id"] != tenantAlpha || row["area_code"] != "QMS" {
		t.Errorf("row content mismatch: %v", row)
	}
}

// userReq seeds tenant + a non-system-admin actor context, so the handler
// treats the caller as a plain membership.view holder (self-scoped listing).
func userReq(method, url string, body string, tenantID, actorID string, roles ...iamdomain.Role) *http.Request {
	var r io.Reader
	if body != "" {
		r = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, url, r)
	ctx := tenant.WithTenantID(req.Context(), tenantID)
	ctx = iamdomain.WithAuthContext(ctx, actorID, roles)
	return req.WithContext(ctx)
}

func TestListAreaMemberships_SystemAdminSeesTenantWideDirectory(t *testing.T) {
	h := newHarness(t)
	h.repo.Insert(context.Background(), iamdomain.UserProcessArea{
		UserID: targetID, TenantID: tenantAlpha, AreaCode: "QMS", Role: iamdomain.RoleAuthor,
		EffectiveFrom: time.Now().UTC().Add(-time.Hour),
	})
	h.repo.Insert(context.Background(), iamdomain.UserProcessArea{
		UserID: "bob", TenantID: tenantAlpha, AreaCode: "RH", Role: iamdomain.RoleApprover,
		EffectiveFrom: time.Now().UTC().Add(-time.Hour),
	})

	// No userId filter + system_admin -> whole tenant directory (both users).
	req := adminReq(http.MethodGet, "/api/v1/iam/area-memberships", "", tenantAlpha, adminID)
	rec := httptest.NewRecorder()
	h.mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s want 200", rec.Code, rec.Body.String())
	}
	var resp struct {
		Items []map[string]any `json:"items"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(resp.Items) != 2 {
		t.Fatalf("items=%d want 2 (tenant-wide directory); body=%s", len(resp.Items), rec.Body.String())
	}
}

func TestListAreaMemberships_NonAdminIsSelfScoped(t *testing.T) {
	h := newHarness(t)
	h.verifier.seed(tenantAlpha, "carol")
	h.repo.Insert(context.Background(), iamdomain.UserProcessArea{
		UserID: "carol", TenantID: tenantAlpha, AreaCode: "QMS", Role: iamdomain.RoleAuthor,
		EffectiveFrom: time.Now().UTC().Add(-time.Hour),
	})
	h.repo.Insert(context.Background(), iamdomain.UserProcessArea{
		UserID: targetID, TenantID: tenantAlpha, AreaCode: "RH", Role: iamdomain.RoleApprover,
		EffectiveFrom: time.Now().UTC().Add(-time.Hour),
	})

	// Non-admin omitting userId -> only own rows, never the full directory.
	req := userReq(http.MethodGet, "/api/v1/iam/area-memberships", "", tenantAlpha, "carol", iamdomain.RoleViewer)
	rec := httptest.NewRecorder()
	h.mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s want 200", rec.Code, rec.Body.String())
	}
	var resp struct {
		Items []struct {
			UserID string `json:"user_id"`
		} `json:"items"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(resp.Items) != 1 || resp.Items[0].UserID != "carol" {
		t.Fatalf("non-admin self-scope breach: %+v", resp.Items)
	}
}

// TestListAreaMemberships_AreaAdminSeesManagedAreasOnly locks ADR 0022 Phase 4:
// an area_admin (membership.manage in QMS only) sees every membership in QMS but
// NOT memberships in RH (an unmanaged area). Proves the BOLA guard — the
// directory is scoped to managed areas, not the actor's own row, and not the
// whole tenant.
func TestListAreaMemberships_AreaAdminSeesManagedAreasOnly(t *testing.T) {
	h := newHarness(t)
	h.repo.managedAreas["area-admin-1"] = []string{"QMS"}
	// QMS rows for two other users (managed) + one RH row (unmanaged).
	h.repo.Insert(context.Background(), iamdomain.UserProcessArea{
		UserID: targetID, TenantID: tenantAlpha, AreaCode: "QMS", Role: iamdomain.RoleAuthor,
		EffectiveFrom: time.Now().UTC().Add(-time.Hour),
	})
	h.repo.Insert(context.Background(), iamdomain.UserProcessArea{
		UserID: "bob", TenantID: tenantAlpha, AreaCode: "QMS", Role: iamdomain.RoleApprover,
		EffectiveFrom: time.Now().UTC().Add(-time.Hour),
	})
	h.repo.Insert(context.Background(), iamdomain.UserProcessArea{
		UserID: "carol", TenantID: tenantAlpha, AreaCode: "RH", Role: iamdomain.RoleViewer,
		EffectiveFrom: time.Now().UTC().Add(-time.Hour),
	})

	req := userReq(http.MethodGet, "/api/v1/iam/area-memberships", "", tenantAlpha, "area-admin-1", iamdomain.RoleAreaAdmin)
	rec := httptest.NewRecorder()
	h.mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s want 200", rec.Code, rec.Body.String())
	}
	var resp struct {
		Items []struct {
			UserID   string `json:"user_id"`
			AreaCode string `json:"area_code"`
		} `json:"items"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(resp.Items) != 2 {
		t.Fatalf("items=%d want 2 (QMS only); body=%s", len(resp.Items), rec.Body.String())
	}
	for _, it := range resp.Items {
		if it.AreaCode != "QMS" {
			t.Errorf("area_admin saw unmanaged area row: %+v", it)
		}
	}
}

func TestListAreaMemberships_NonAdminCannotTargetOther(t *testing.T) {
	h := newHarness(t)
	req := userReq(http.MethodGet, "/api/v1/iam/area-memberships?user_id="+targetID, "", tenantAlpha, "carol", iamdomain.RoleViewer)
	rec := httptest.NewRecorder()
	h.mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status=%d body=%s want 403 (non-admin probing other user)", rec.Code, rec.Body.String())
	}
}

func TestListAreaMemberships_AreaAndRoleFilters(t *testing.T) {
	h := newHarness(t)
	h.repo.Insert(context.Background(), iamdomain.UserProcessArea{
		UserID: targetID, TenantID: tenantAlpha, AreaCode: "QMS", Role: iamdomain.RoleAuthor,
		EffectiveFrom: time.Now().UTC().Add(-time.Hour),
	})
	h.repo.Insert(context.Background(), iamdomain.UserProcessArea{
		UserID: "bob", TenantID: tenantAlpha, AreaCode: "RH", Role: iamdomain.RoleApprover,
		EffectiveFrom: time.Now().UTC().Add(-time.Hour),
	})

	req := adminReq(http.MethodGet, "/api/v1/iam/area-memberships?area_code=RH&role=approver", "", tenantAlpha, adminID)
	rec := httptest.NewRecorder()
	h.mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s want 200", rec.Code, rec.Body.String())
	}
	var resp struct {
		Items []struct {
			UserID   string `json:"user_id"`
			AreaCode string `json:"area_code"`
			Role     string `json:"role"`
		} `json:"items"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(resp.Items) != 1 || resp.Items[0].AreaCode != "RH" || resp.Items[0].Role != "approver" {
		t.Fatalf("area/role filter breach: %+v", resp.Items)
	}
}

func TestGrantMembership_EmitsAudit(t *testing.T) {
	h := newHarness(t)
	h.verifier.seed(tenantAlpha, targetID)

	body := `{"user_id":"` + targetID + `","area_code":"QMS","role":"author"}`
	req := adminReq(http.MethodPost, "/api/v1/iam/area-memberships", body, tenantAlpha, adminID)
	rec := httptest.NewRecorder()
	h.mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status=%d body=%s want 201", rec.Code, rec.Body.String())
	}
	events := h.audit.snapshot()
	if len(events) != 1 {
		t.Fatalf("audit events=%d want 1", len(events))
	}
	ev := events[0]
	if ev.Action != "iam.area_membership.granted" {
		t.Errorf("action=%q want iam.area_membership.granted", ev.Action)
	}
	if ev.ResourceType != "area_membership" {
		t.Errorf("resourceType=%q want area_membership", ev.ResourceType)
	}
	if ev.ResourceID != targetID {
		t.Errorf("resourceID=%q want %q", ev.ResourceID, targetID)
	}
	if ev.TenantID != tenantAlpha {
		t.Errorf("tenantID=%q want %q", ev.TenantID, tenantAlpha)
	}
	if !strings.Contains(ev.PayloadJSON, `"area_code":"QMS"`) || !strings.Contains(ev.PayloadJSON, `"role":"author"`) {
		t.Errorf("payload missing area_code/role: %s", ev.PayloadJSON)
	}
}

func TestRevokeMembership_RejectsCrossTenantUserWith404(t *testing.T) {
	h := newHarness(t)
	// Target lives in tenantBeta; admin is acting under tenantAlpha.
	h.verifier.seed(tenantBeta, targetID)
	h.repo.Insert(context.Background(), iamdomain.UserProcessArea{
		UserID:        targetID,
		TenantID:      tenantBeta,
		AreaCode:      "QMS",
		Role:          iamdomain.RoleAuthor,
		EffectiveFrom: time.Now().UTC().Add(-time.Hour),
	})

	req := adminReq(http.MethodDelete, "/api/v1/iam/area-memberships/"+targetID+"/QMS", "", tenantAlpha, adminID)
	rec := httptest.NewRecorder()
	h.mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status=%d body=%s want 404 (cross-tenant probe must not leak existence)", rec.Code, rec.Body.String())
	}
	// Confirm membership in tenantBeta was NOT revoked.
	rows, err := h.repo.ListActive(context.Background(), targetID, tenantBeta, time.Now().UTC())
	if err != nil {
		t.Fatalf("ListActive: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("cross-tenant revoke leaked: tenantBeta active=%d want 1", len(rows))
	}
	// And no audit event was emitted (the guard ran before service).
	if events := h.audit.snapshot(); len(events) != 0 {
		t.Errorf("cross-tenant 404 must not emit audit; got %d events", len(events))
	}
}

func TestGrantMembership_DuplicateReturns409(t *testing.T) {
	h := newHarness(t)
	h.verifier.seed(tenantAlpha, targetID)

	body := `{"user_id":"` + targetID + `","area_code":"QMS","role":"author"}`

	// First grant — 201.
	req1 := adminReq(http.MethodPost, "/api/v1/iam/area-memberships", body, tenantAlpha, adminID)
	rec1 := httptest.NewRecorder()
	h.mux.ServeHTTP(rec1, req1)
	if rec1.Code != http.StatusCreated {
		t.Fatalf("first grant status=%d body=%s want 201", rec1.Code, rec1.Body.String())
	}

	// Same role re-grant — 409 MEMBERSHIP_EXISTS.
	req2 := adminReq(http.MethodPost, "/api/v1/iam/area-memberships", body, tenantAlpha, adminID)
	rec2 := httptest.NewRecorder()
	h.mux.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusConflict {
		t.Fatalf("duplicate grant status=%d body=%s want 409", rec2.Code, rec2.Body.String())
	}
	if ct := rec2.Header().Get("Content-Type"); ct != "application/problem+json" {
		t.Errorf("duplicate Content-Type=%q want application/problem+json", ct)
	}
	var p struct {
		Code   string `json:"code"`
		Status int    `json:"status"`
	}
	if err := json.Unmarshal(rec2.Body.Bytes(), &p); err != nil {
		t.Fatalf("decode problem: %v body=%s", err, rec2.Body.String())
	}
	if p.Code != "conflict.membership_exists" || p.Status != http.StatusConflict {
		t.Errorf("problem code=%q status=%d want MEMBERSHIP_EXISTS/409", p.Code, p.Status)
	}
	// Only one audit event (the first successful grant).
	if events := h.audit.snapshot(); len(events) != 1 {
		t.Errorf("audit events=%d want 1 (only first grant); got %+v", len(events), events)
	}
}

func TestListAreaMemberships_AreaScopedUnderTenantIsolation(t *testing.T) {
	h := newHarness(t)
	// Target is in BOTH tenants (e.g. cross-org user) with different area sets.
	h.verifier.seed(tenantAlpha, targetID)
	h.verifier.seed(tenantBeta, targetID)
	h.repo.Insert(context.Background(), iamdomain.UserProcessArea{
		UserID:        targetID,
		TenantID:      tenantAlpha,
		AreaCode:      "QMS",
		Role:          iamdomain.RoleAuthor,
		EffectiveFrom: time.Now().UTC().Add(-time.Hour),
	})
	h.repo.Insert(context.Background(), iamdomain.UserProcessArea{
		UserID:        targetID,
		TenantID:      tenantBeta,
		AreaCode:      "RH",
		Role:          iamdomain.RoleApprover,
		EffectiveFrom: time.Now().UTC().Add(-time.Hour),
	})

	// List from tenantAlpha context — must only see QMS row.
	req := adminReq(http.MethodGet, "/api/v1/iam/area-memberships?user_id="+targetID, "", tenantAlpha, adminID)
	rec := httptest.NewRecorder()
	h.mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("tenantAlpha list status=%d body=%s want 200", rec.Code, rec.Body.String())
	}
	var resp struct {
		Items []struct {
			TenantID string `json:"tenant_id"`
			AreaCode string `json:"area_code"`
		} `json:"items"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v body=%s", err, rec.Body.String())
	}
	if len(resp.Items) != 1 {
		t.Fatalf("items=%d want 1 (tenant isolation breach)", len(resp.Items))
	}
	got := resp.Items[0]
	if got.TenantID != tenantAlpha || got.AreaCode != "QMS" {
		t.Errorf("tenantAlpha listing leaked tenantBeta row: %+v", got)
	}
}

// TestGrantMembership_AreaAdminReachesServiceAfterGateRemoval locks ADR 0022
// Phase 3: the canManageMembershipTarget RoleSystemAdmin handler gate is gone, so
// a non-system actor (area_admin) granting another target is no longer blocked at
// the handler. The in-memory repo has no tier-2 authz, so the request reaches the
// service and succeeds (201); the actual area-scope enforcement (area_admin denied
// outside a managed area) is asserted in tests/integration/iam against a real DB.
func TestGrantMembership_AreaAdminReachesServiceAfterGateRemoval(t *testing.T) {
	h := newHarness(t)
	h.verifier.seed(tenantAlpha, targetID)

	body := `{"user_id":"` + targetID + `","area_code":"QMS","role":"author"}`
	req := userReq(http.MethodPost, "/api/v1/iam/area-memberships", body, tenantAlpha, "area-admin-1", iamdomain.RoleAreaAdmin)
	rec := httptest.NewRecorder()
	h.mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("area_admin grant status=%d body=%s want 201 (handler role gate must be removed)", rec.Code, rec.Body.String())
	}
	if events := h.audit.snapshot(); len(events) != 1 {
		t.Errorf("area_admin grant audit events=%d want 1", len(events))
	}
}

// TestRevokeMembership_AreaAdminReachesServiceAfterGateRemoval mirrors the grant
// case for the revoke path (DELETE), which also dropped the role gate.
func TestRevokeMembership_AreaAdminReachesServiceAfterGateRemoval(t *testing.T) {
	h := newHarness(t)
	h.verifier.seed(tenantAlpha, targetID)
	h.repo.Insert(context.Background(), iamdomain.UserProcessArea{
		UserID: targetID, TenantID: tenantAlpha, AreaCode: "QMS", Role: iamdomain.RoleAuthor,
		EffectiveFrom: time.Now().UTC().Add(-time.Hour),
	})

	req := userReq(http.MethodDelete, "/api/v1/iam/area-memberships/"+targetID+"/QMS", "", tenantAlpha, "area-admin-1", iamdomain.RoleAreaAdmin)
	rec := httptest.NewRecorder()
	h.mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("area_admin revoke status=%d body=%s want 204 (handler role gate must be removed)", rec.Code, rec.Body.String())
	}
}

// TestGrantMembership_AreaAdminSelfGrantStill403 confirms the self-grant business
// invariant survives the gate removal for a non-system actor.
func TestGrantMembership_AreaAdminSelfGrantStill403(t *testing.T) {
	h := newHarness(t)
	h.verifier.seed(tenantAlpha, "area-admin-1")

	body := `{"user_id":"area-admin-1","area_code":"QMS","role":"approver"}`
	req := userReq(http.MethodPost, "/api/v1/iam/area-memberships", body, tenantAlpha, "area-admin-1", iamdomain.RoleAreaAdmin)
	rec := httptest.NewRecorder()
	h.mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("area_admin self-grant status=%d body=%s want 403", rec.Code, rec.Body.String())
	}
	if rows, _ := h.repo.ListActive(context.Background(), "area-admin-1", tenantAlpha, time.Now().UTC()); len(rows) != 0 {
		t.Errorf("self-grant created %d rows; must be 0", len(rows))
	}
}

// TestGrantMembership_RejectsSelfGrant locks the PR-1 security hardening: a
// CapMembershipManage holder must not be able to grant themselves additional
// roles even though canManageMembershipTarget treats actor==target as
// permitted (legitimate for self-list and self-revoke).
func TestGrantMembership_RejectsSelfGrant(t *testing.T) {
	h := newHarness(t)
	h.verifier.seed(tenantAlpha, adminID)

	body := `{"user_id":"` + adminID + `","area_code":"QMS","role":"approver"}`
	req := adminReq(http.MethodPost, "/api/v1/iam/area-memberships", body, tenantAlpha, adminID)
	rec := httptest.NewRecorder()
	h.mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("self-grant status=%d body=%s want 403", rec.Code, rec.Body.String())
	}
	// And no audit / no row created.
	if rows, _ := h.repo.ListActive(context.Background(), adminID, tenantAlpha, time.Now().UTC()); len(rows) != 0 {
		t.Errorf("self-grant created %d rows; must be 0", len(rows))
	}
	if events := h.audit.snapshot(); len(events) != 0 {
		t.Errorf("self-grant emitted %d audit events; must be 0", len(events))
	}
}
