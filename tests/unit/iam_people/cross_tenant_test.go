// cross_tenant_test.go — PR-7b hardening regressions for cross-tenant ATO.
// Asserts handleResetPassword + handleUnlock + handlePatch reject a target
// userID that belongs to a different tenant, returning 404 NOT 403 so the
// existence of the cross-tenant user is not leaked.
package iampeopletest

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	authdomain "metaldocs/internal/modules/auth/domain"
	iamapp "metaldocs/internal/modules/iam/application"
	iamdelivery "metaldocs/internal/modules/iam/delivery/http"
	iamdomain "metaldocs/internal/modules/iam/domain"
	iammemory "metaldocs/internal/modules/iam/infrastructure/memory"
)

const (
	crossTenantA = "11111111-1111-1111-1111-111111111111"
	crossTenantB = "22222222-2222-2222-2222-222222222222"
)

// tenantAwareAuth scopes ListUsers by tenant so VerifyUserInTenant gets a
// realistic answer. It also satisfies the UserAdminService interface
// PeopleHandler holds on h.authSvc — same single type covers both surfaces.
type tenantAwareAuth struct {
	usersByTenant map[string][]authdomain.ManagedUser
	resets        []string
	unlocks       []string
	patches       int
}

func (a *tenantAwareAuth) seed(tenantID string, u authdomain.ManagedUser) {
	if a.usersByTenant == nil {
		a.usersByTenant = map[string][]authdomain.ManagedUser{}
	}
	a.usersByTenant[tenantID] = append(a.usersByTenant[tenantID], u)
}

// peopleAuthService surface.
func (a *tenantAwareAuth) CreateUserWithInput(context.Context, authdomain.CreateUserInput) error {
	return nil
}
func (a *tenantAwareAuth) UpdateUser(context.Context, authdomain.UpdateUserParams, string) error {
	a.patches++
	return nil
}
func (a *tenantAwareAuth) AdminResetPassword(_ context.Context, userID, _ string) error {
	a.resets = append(a.resets, userID)
	return nil
}
func (a *tenantAwareAuth) UnlockUser(_ context.Context, userID string) error {
	a.unlocks = append(a.unlocks, userID)
	return nil
}
func (a *tenantAwareAuth) ListUsers(_ context.Context, tenantID string) ([]authdomain.ManagedUser, error) {
	return append([]authdomain.ManagedUser(nil), a.usersByTenant[tenantID]...), nil
}

// UserAdminService surface (extra methods).
func (a *tenantAwareAuth) ListOnlineUsers(context.Context, string, time.Time) ([]authdomain.OnlineUser, error) {
	return nil, nil
}
func (a *tenantAwareAuth) CreateUser(context.Context, string, string, string, string, string, string, []iamdomain.Role, string) error {
	return nil
}

func newCrossTenantMux(auth *tenantAwareAuth) *http.ServeMux {
	roleAdmin := iammemory.NewRoleAdminRepository()
	svc := iamapp.PeopleServiceFromInterfaces(auth, fakeRoleProvider{}, roleAdmin, nil, iamapp.PermissiveAreaCatalog{}, nil)
	h := iamdelivery.NewPeopleHandler(svc, auth, nil)
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)
	return mux
}

func TestResetPassword_RejectsCrossTenantUserWith404(t *testing.T) {
	auth := &tenantAwareAuth{}
	auth.seed(crossTenantB, authdomain.ManagedUser{UserID: "victim", Username: "victim"})

	mux := newCrossTenantMux(auth)

	body := strings.NewReader(`{"newPassword":"NewSecret123!"}`)
	req := withTenant(httptest.NewRequest(http.MethodPost, "/api/v1/iam/users/victim/reset-password", body), crossTenantA)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d body=%s; want 404 NOT_FOUND", rec.Code, rec.Body.String())
	}
	if len(auth.resets) != 0 {
		t.Fatalf("AdminResetPassword was reached for cross-tenant target: %v", auth.resets)
	}
}

func TestUnlock_RejectsCrossTenantUserWith404(t *testing.T) {
	auth := &tenantAwareAuth{}
	auth.seed(crossTenantB, authdomain.ManagedUser{UserID: "victim", Username: "victim"})

	mux := newCrossTenantMux(auth)

	req := withTenant(httptest.NewRequest(http.MethodPost, "/api/v1/iam/users/victim/unlock", nil), crossTenantA)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d body=%s; want 404 NOT_FOUND", rec.Code, rec.Body.String())
	}
	if len(auth.unlocks) != 0 {
		t.Fatalf("UnlockUser was reached for cross-tenant target: %v", auth.unlocks)
	}
}

func TestPatch_RejectsCrossTenantUserWith404(t *testing.T) {
	auth := &tenantAwareAuth{}
	auth.seed(crossTenantB, authdomain.ManagedUser{UserID: "victim", Username: "victim"})

	mux := newCrossTenantMux(auth)

	body := strings.NewReader(`{"displayName":"hijacked"}`)
	req := withTenant(httptest.NewRequest(http.MethodPatch, "/api/v1/iam/users/victim", body), crossTenantA)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d body=%s; want 404 NOT_FOUND", rec.Code, rec.Body.String())
	}
	if auth.patches != 0 {
		t.Fatalf("UpdateUser was reached for cross-tenant target")
	}
}
