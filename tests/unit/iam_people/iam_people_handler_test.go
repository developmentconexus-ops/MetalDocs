// iam_people_handler_test.go — PR-4 People-tab handler coverage.
//
// Lives in its own sub-package (tests/unit/iam_people) because the broader
// tests/unit package has pre-existing breakage that PR-4 is not allowed to
// repair (auth Service signature drift, RoleAdminRepository interface
// changes, etc.). Running:
//
//   go test ./tests/unit/iam_people/...
//
// exercises Invite, Patch, Bulk, and force-logout deferral via the actual
// HTTP handlers wired against in-memory dependencies — no DB or bcrypt.
package iampeopletest

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
	"unicode"

	auditdomain "metaldocs/internal/modules/audit/domain"
	authdomain "metaldocs/internal/modules/auth/domain"
	iamapp "metaldocs/internal/modules/iam/application"
	iamdelivery "metaldocs/internal/modules/iam/delivery/http"
	iamdomain "metaldocs/internal/modules/iam/domain"
	iammemory "metaldocs/internal/modules/iam/infrastructure/memory"
	"metaldocs/internal/platform/tenant"
)

// ──────────────────────────────────────────────────────────────────────────
// Fakes — small enough to read in one screen each.

type fakeAuthService struct {
	mu       sync.Mutex
	users    map[string]authdomain.ManagedUser
	creates  []authdomain.CreateUserInput
	patches  []authdomain.UpdateUserParams
	resets   []string
	unlocks  []string
	failNext error
}

func newFakeAuthService() *fakeAuthService {
	return &fakeAuthService{users: map[string]authdomain.ManagedUser{}}
}

func (f *fakeAuthService) CreateUserWithInput(_ context.Context, input authdomain.CreateUserInput) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.failNext != nil {
		err := f.failNext
		f.failNext = nil
		return err
	}
	id := strings.TrimSpace(input.UserID.String())
	if id == "" {
		id = input.Username
	}
	if _, ok := f.users[id]; ok {
		return authdomain.ErrUserAlreadyExists
	}
	now := time.Now().UTC()
	f.users[id] = authdomain.ManagedUser{
		UserID:      id,
		Username:    input.Username,
		Email:       input.Email,
		DisplayName: input.DisplayName,
		IsActive:    true,
		Roles:       []iamdomain.Role{input.Roles[0]},
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	f.creates = append(f.creates, input)
	return nil
}

func (f *fakeAuthService) UpdateUser(_ context.Context, params authdomain.UpdateUserParams, _ string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	u, ok := f.users[params.UserID]
	if !ok {
		return authdomain.ErrIdentityNotFound
	}
	if params.DisplayName != nil {
		u.DisplayName = *params.DisplayName
	}
	if params.Email != nil {
		u.Email = *params.Email
	}
	if params.IsActive != nil {
		u.IsActive = *params.IsActive
	}
	if params.MustChangePassword != nil {
		u.MustChangePassword = *params.MustChangePassword
	}
	if params.ResetLockState {
		u.FailedLoginAttempts = 0
		u.LockedUntil = nil
	}
	f.users[params.UserID] = u
	f.patches = append(f.patches, params)
	return nil
}

func (f *fakeAuthService) AdminResetPassword(_ context.Context, userID, _ string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.resets = append(f.resets, userID)
	return nil
}

func (f *fakeAuthService) UnlockUser(_ context.Context, userID string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if _, ok := f.users[userID]; !ok {
		return authdomain.ErrIdentityNotFound
	}
	f.unlocks = append(f.unlocks, userID)
	return nil
}

func (f *fakeAuthService) ListUsers(_ context.Context, _ string) ([]authdomain.ManagedUser, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]authdomain.ManagedUser, 0, len(f.users))
	for _, u := range f.users {
		out = append(out, u)
	}
	return out, nil
}

// fakeAuthAdminService is the UserAdminService interface that PeopleHandler
// needs for reset-password and unlock. We adapt fakeAuthService to it.
type fakeAuthAdminService struct {
	inner *fakeAuthService
}

func (f *fakeAuthAdminService) ListUsers(ctx context.Context, tenantID string) ([]authdomain.ManagedUser, error) {
	return f.inner.ListUsers(ctx, tenantID)
}

func (f *fakeAuthAdminService) ListOnlineUsers(_ context.Context, _ string, _ time.Time) ([]authdomain.OnlineUser, error) {
	return nil, nil
}

func (f *fakeAuthAdminService) CreateUser(context.Context, string, string, string, string, string, string, []iamdomain.Role, string) error {
	return nil
}

func (f *fakeAuthAdminService) UpdateUser(ctx context.Context, params authdomain.UpdateUserParams, pwd string) error {
	return f.inner.UpdateUser(ctx, params, pwd)
}

func (f *fakeAuthAdminService) AdminResetPassword(ctx context.Context, userID, pwd string) error {
	return f.inner.AdminResetPassword(ctx, userID, pwd)
}

func (f *fakeAuthAdminService) UnlockUser(ctx context.Context, userID string) error {
	return f.inner.UnlockUser(ctx, userID)
}

type fakeRoleProvider struct{}

func (fakeRoleProvider) RolesByUserID(_ context.Context, _, _ string) ([]iamdomain.Role, error) {
	return []iamdomain.Role{iamdomain.RoleAuthor}, nil
}

type fakeAuditWriter struct {
	mu     sync.Mutex
	events []auditdomain.Event
}

func (f *fakeAuditWriter) Record(_ context.Context, ev auditdomain.Event) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.events = append(f.events, ev)
	return nil
}

func (f *fakeAuditWriter) RecordTx(_ context.Context, _ *sql.Tx, ev auditdomain.Event) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.events = append(f.events, ev)
	return nil
}

func (f *fakeAuditWriter) snapshot() []auditdomain.Event {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]auditdomain.Event, len(f.events))
	copy(out, f.events)
	return out
}

type peopleServiceTestBuilder struct {
	auth        *fakeAuthService
	roleAdmin   *iammemory.RoleAdminRepository
	memberships *iamapp.AreaMembershipService
}

func newPeopleServiceForTest(t *testing.T) (*iamapp.PeopleService, *peopleServiceTestBuilder) {
	t.Helper()
	auth := newFakeAuthService()
	roleAdmin := iammemory.NewRoleAdminRepository()

	// We use the same iammemory repo as the UserAreaWriteRepository to keep the
	// surface small; the iammemory package does not currently provide one, so
	// we use a noop in-memory shim. The Invite happy-path test does not pass
	// memberships, and the deferred PR-7 force-logout test does not need it.
	noopMembership := &noopAreaRepo{}
	memberships := iamapp.NewAreaMembershipService(noopMembership, nil)

	// We pass our fakeAuthService through the iface-only constructor by
	// reaching into the package; PeopleService accepts an interface in
	// newPeopleServiceFromIfaces (unexported). Using NewPeopleService instead
	// requires a real *authapp.Service. Workaround: we call NewPeopleService
	// with nil-typed *authapp.Service and immediately overwrite via the
	// public WithTempPasswordGenerator hook, but we cannot inject our fake
	// authService that way. So we use a thin shim that satisfies the
	// constructor by using a real authapp.Service is too heavy.
	//
	// Strategy: PeopleService is exported and the constructor takes a
	// *authapp.Service concrete type. We avoid that ceiling by going through
	// the HTTP layer with a fake CreateUserWithInput executor wired into the
	// PeopleHandler indirectly via the AuthSvc dependency... but PeopleService
	// itself owns the invite path. Therefore: we build PeopleService through
	// a small adaptor that re-implements the public surface as a thin
	// wrapper. See newPeopleServiceFakeWrapper below.
	svc := newPeopleServiceFakeWrapper(auth, roleAdmin, memberships)
	return svc, &peopleServiceTestBuilder{auth: auth, roleAdmin: roleAdmin, memberships: memberships}
}

// noopAreaRepo satisfies iamapp.UserAreaWriteRepository for the membership
// service in tests that do not exercise the membership path.
type noopAreaRepo struct{}

func (noopAreaRepo) ListActive(context.Context, string, string, time.Time) ([]iamdomain.UserProcessArea, error) {
	return nil, nil
}
func (noopAreaRepo) Insert(context.Context, iamdomain.UserProcessArea) error { return nil }
func (noopAreaRepo) CloseActive(context.Context, string, string, string, time.Time, string) error {
	return nil
}
func (noopAreaRepo) GrantAtomic(context.Context, iamdomain.UserProcessArea, iamdomain.UserProcessArea) error {
	return nil
}
func (noopAreaRepo) GetActiveByUserAndArea(context.Context, string, string, string, time.Time) (*iamdomain.UserProcessArea, error) {
	return nil, nil
}

// Builds a PeopleService using the test fakes. We cannot reach the unexported
// iface-constructor from this package, so we go through the
// HandlerWithFakeService shim below: it directly drives the HTTP handler with
// a PeopleService built around the auth Service's interface contract via a
// production constructor that accepts a CreateUserWithInput-compatible value.
//
// In practice the constructor we use is the test-only helper exported in the
// iam application package as PeopleServiceFromInterfaces — added in PR-4 for
// this purpose.
func newPeopleServiceFakeWrapper(
	auth *fakeAuthService,
	roleAdmin iamdomain.RoleAdminRepository,
	memberships *iamapp.AreaMembershipService,
) *iamapp.PeopleService {
	return iamapp.PeopleServiceFromInterfaces(auth, fakeRoleProvider{}, roleAdmin, memberships, iamapp.PermissiveAreaCatalog{}, nil)
}

// withTenant + withActor — request-context helpers.

func withTenant(req *http.Request, tenantID string) *http.Request {
	return req.WithContext(tenant.WithTenantID(req.Context(), tenantID))
}

// ──────────────────────────────────────────────────────────────────────────
// Test cases.

const testTenantID = "11111111-1111-1111-1111-111111111111"

func newHandlerForTest(t *testing.T) (*http.ServeMux, *peopleServiceTestBuilder, *fakeAuditWriter) {
	t.Helper()
	svc, builder := newPeopleServiceForTest(t)
	auditWriter := &fakeAuditWriter{}
	authAdmin := &fakeAuthAdminService{inner: builder.auth}
	handler := iamdelivery.NewPeopleHandler(svc, authAdmin, auditWriter)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)
	return mux, builder, auditWriter
}

func TestListUsers_PaginationBoundaryAndQFilter(t *testing.T) {
	mux, builder, _ := newHandlerForTest(t)

	// Seed three users via the fake auth service so ListUsers returns them.
	for _, u := range []struct {
		id, dn, em string
	}{
		{"alice", "Alice Anderson", "alice@example.com"},
		{"bob", "Bob Builder", "bob@example.com"},
		{"carol", "Carol Curie", "carol@example.com"},
	} {
		_ = builder.auth.CreateUserWithInput(context.Background(), authdomain.CreateUserInput{
			UserID:      authdomain.UserID(u.id),
			Username:    u.id,
			Email:       u.em,
			DisplayName: u.dn,
			Password:    "TempPassword1!",
			TenantID:    testTenantID,
			Roles:       []iamdomain.Role{iamdomain.RoleAuthor},
		})
	}

	// Page size 10 — no hasMore.
	req := withTenant(httptest.NewRequest(http.MethodGet, "/api/v1/iam/users?limit=10", nil), testTenantID)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("list status = %d body=%s", rec.Code, rec.Body.String())
	}
	var page struct {
		Items []map[string]any `json:"items"`
		Page  struct {
			HasMore    bool    `json:"has_more"`
			NextCursor *string `json:"next_cursor"`
		} `json:"page"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &page); err != nil {
		t.Fatalf("decode: %v body=%s", err, rec.Body.String())
	}
	if page.Page.HasMore {
		t.Fatalf("expected hasMore=false on last page, got true")
	}
	if len(page.Items) != 3 {
		t.Fatalf("expected 3 items, got %d", len(page.Items))
	}

	// q filter — case insensitive 'CAROL' should hit one row.
	req2 := withTenant(httptest.NewRequest(http.MethodGet, "/api/v1/iam/users?q=CAROL", nil), testTenantID)
	rec2 := httptest.NewRecorder()
	mux.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusOK {
		t.Fatalf("q filter status = %d body=%s", rec2.Code, rec2.Body.String())
	}
	var page2 struct {
		Items []map[string]any `json:"items"`
	}
	_ = json.Unmarshal(rec2.Body.Bytes(), &page2)
	if len(page2.Items) != 1 {
		t.Fatalf("expected 1 item for q=CAROL, got %d", len(page2.Items))
	}
}

func TestInviteUser_ReturnsTempPasswordAndMeetsPolicy(t *testing.T) {
	mux, _, audit := newHandlerForTest(t)

	body := `{"username":"new.user","email":"new.user@example.com","displayName":"New User","tenantRole":"author"}`
	req := withTenant(httptest.NewRequest(http.MethodPost, "/api/v1/iam/users/invite", strings.NewReader(body)), testTenantID)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("invite status = %d body=%s", rec.Code, rec.Body.String())
	}
	var resp struct {
		UserId       string `json:"userId"`
		TempPassword string `json:"tempPassword"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.UserId != "new.user" {
		t.Fatalf("userId = %q want new.user", resp.UserId)
	}
	if len(resp.TempPassword) != 16 {
		t.Fatalf("tempPassword length = %d, want 16 (value=%q)", len(resp.TempPassword), resp.TempPassword)
	}
	if err := assertPasswordPolicy(resp.TempPassword); err != nil {
		t.Fatalf("tempPassword %q violates policy: %v", resp.TempPassword, err)
	}

	// Audit emitted, no password in payload.
	events := audit.snapshot()
	if len(events) != 1 {
		t.Fatalf("expected 1 audit event, got %d", len(events))
	}
	if events[0].Action != "iam.user.invited" {
		t.Fatalf("audit action = %q", events[0].Action)
	}
	if strings.Contains(events[0].PayloadJSON, resp.TempPassword) {
		t.Fatalf("CRITICAL: tempPassword leaked into audit payload: %s", events[0].PayloadJSON)
	}
}

func assertPasswordPolicy(password string) error {
	var hasUpper, hasLower, hasDigit, hasSymbol bool
	for _, r := range password {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsDigit(r):
			hasDigit = true
		case strings.ContainsRune("!@#$%^&*-_=+?", r):
			hasSymbol = true
		}
	}
	if !hasUpper || !hasLower || !hasDigit || !hasSymbol {
		return errors.New("missing one of upper/lower/digit/symbol")
	}
	return nil
}

func TestInviteUser_MissingTenantContext(t *testing.T) {
	mux, _, _ := newHandlerForTest(t)

	body := `{"username":"x","email":"x@example.com","displayName":"X","tenantRole":"author"}`
	// Note: NO tenant in context.
	req := httptest.NewRequest(http.MethodPost, "/api/v1/iam/users/invite", strings.NewReader(body))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code < 400 {
		t.Fatalf("expected error status (>=400) when tenant missing, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestInviteUser_InvalidAreaRoleReturns400(t *testing.T) {
	mux, _, _ := newHandlerForTest(t)

	// "viewer" is a canonical tenant role but NOT in the area-only subset
	// {signer, area_admin, qms_admin}.
	body := `{
		"username":"x.user",
		"email":"x.user@example.com",
		"displayName":"X User",
		"tenantRole":"author",
		"areaMemberships":[{"areaCode":"AREA1","role":"viewer"}]
	}`
	req := withTenant(httptest.NewRequest(http.MethodPost, "/api/v1/iam/users/invite", strings.NewReader(body)), testTenantID)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid area role, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestPatchUser_EmitsAuditWithoutPassword(t *testing.T) {
	mux, builder, audit := newHandlerForTest(t)

	// Seed.
	_ = builder.auth.CreateUserWithInput(context.Background(), authdomain.CreateUserInput{
		UserID: "alice", Username: "alice", DisplayName: "Alice", Password: "Temp123!Abc", TenantID: testTenantID,
		Roles: []iamdomain.Role{iamdomain.RoleAuthor},
	})

	body := `{"displayName":"Alice Renamed","newPassword":"NEVER_LOG_THIS_SECRET_42!"}`
	req := withTenant(httptest.NewRequest(http.MethodPatch, "/api/v1/iam/users/alice", strings.NewReader(body)), testTenantID)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("patch status = %d body=%s", rec.Code, rec.Body.String())
	}
	events := audit.snapshot()
	if len(events) != 1 || events[0].Action != "iam.user.updated" {
		t.Fatalf("expected single iam.user.updated audit event, got %+v", events)
	}
	if strings.Contains(events[0].PayloadJSON, "NEVER_LOG_THIS_SECRET_42") {
		t.Fatalf("CRITICAL: newPassword leaked into audit payload: %s", events[0].PayloadJSON)
	}
}

func TestBulkUsers_ForceLogoutReturns501NotImplemented(t *testing.T) {
	mux, _, _ := newHandlerForTest(t)

	body := `{"userIds":["alice"],"action":"force-logout"}`
	req := withTenant(httptest.NewRequest(http.MethodPost, "/api/v1/iam/users/bulk", strings.NewReader(body)), testTenantID)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotImplemented {
		t.Fatalf("expected 501 for force-logout, got %d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "NOT_IMPLEMENTED") {
		t.Fatalf("expected NOT_IMPLEMENTED code in body, got %s", rec.Body.String())
	}
}

func TestBulkUsers_PerUserIsolation(t *testing.T) {
	mux, builder, _ := newHandlerForTest(t)

	// Seed two users.
	for _, u := range []string{"alice", "bob"} {
		_ = builder.auth.CreateUserWithInput(context.Background(), authdomain.CreateUserInput{
			UserID: authdomain.UserID(u), Username: u, DisplayName: u, Password: "Temp1!Abcdefg",
			TenantID: testTenantID, Roles: []iamdomain.Role{iamdomain.RoleAuthor},
		})
	}

	// Mix one valid + one missing — both run, only the valid one succeeds.
	body := `{"userIds":["alice","does-not-exist","bob"],"action":"deactivate"}`
	req := withTenant(httptest.NewRequest(http.MethodPost, "/api/v1/iam/users/bulk", strings.NewReader(body)), testTenantID)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("bulk status = %d body=%s", rec.Code, rec.Body.String())
	}
	var resp struct {
		Succeeded []string `json:"succeeded"`
		Failed    []struct {
			UserId  string `json:"userId"`
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"failed"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v body=%s", err, rec.Body.String())
	}
	if len(resp.Succeeded) != 2 {
		t.Fatalf("expected 2 succeeded (alice, bob), got %v", resp.Succeeded)
	}
	if len(resp.Failed) != 1 || resp.Failed[0].UserId != "does-not-exist" {
		t.Fatalf("expected exactly 1 failure for does-not-exist, got %+v", resp.Failed)
	}
}
