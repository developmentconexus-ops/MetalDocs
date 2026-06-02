package unit

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	authapp "metaldocs/internal/modules/auth/application"
	authdomain "metaldocs/internal/modules/auth/domain"
	authmemory "metaldocs/internal/modules/auth/infrastructure/memory"
	iamapp "metaldocs/internal/modules/iam/application"
	iamdelivery "metaldocs/internal/modules/iam/delivery/http"
	iamdomain "metaldocs/internal/modules/iam/domain"
	iammemory "metaldocs/internal/modules/iam/infrastructure/memory"
	"metaldocs/internal/platform/tenant"
)

const testTenantID = "test-tenant"

func newAuthServiceForTest(t *testing.T, authRepo *authmemory.Repository, roleAdminRepo iamdomain.RoleAdminRepository, userRoles map[string][]iamdomain.Role) *authapp.Service {
	t.Helper()
	provider := iamapp.NewDevRoleProvider(userRoles, testTenantID)
	svc, err := authapp.NewService(authRepo, provider, roleAdminRepo, authapp.Config{
		SessionSecret:          "0123456789abcdef0123456789abcdef",
		SessionTTL:             time.Hour,
		PasswordMinLength:      8,
		LoginMaxFailedAttempts: 5,
		LoginLockDuration:      time.Minute,
	})
	if err != nil {
		t.Fatalf("new auth service: %v", err)
	}
	return svc
}

func registerPeopleHandlerForTest(t *testing.T, mux *http.ServeMux, authSvc *authapp.Service, roleAdmin iamdomain.RoleAdminRepository, inv iamapp.RoleCacheInvalidator) {
	t.Helper()
	peopleSvc := iamapp.NewPeopleService(authSvc, peopleRoleProviderShim{}, roleAdmin, nil, nil, inv)
	iamdelivery.NewPeopleHandler(peopleSvc, authSvc, nil).RegisterRoutes(mux)
}

// peopleRoleProviderShim satisfies the peopleRoleProvider interface that
// PeopleService accepts. We do not exercise role-driven branches in these
// tests so a stub returning the viewer role for any user is sufficient.
type peopleRoleProviderShim struct{}

func (peopleRoleProviderShim) RolesByUserID(_ context.Context, _, _ string) ([]iamdomain.Role, error) {
	return []iamdomain.Role{iamdomain.RoleViewer}, nil
}

type fakeInvalidator struct {
	called bool
	userID string
}

func (f *fakeInvalidator) InvalidateUser(userID string) {
	f.called = true
	f.userID = userID
}

func (f *fakeInvalidator) InvalidateUserTenant(userID, _ string) {
	f.called = true
	f.userID = userID
}

func TestIAMAdminHandlerUpsertRole(t *testing.T) {
	repo := iammemory.NewRoleAdminRepository()
	inv := &fakeInvalidator{}
	service := iamapp.NewAdminService(repo, inv)
	handler := iamdelivery.NewAdminHandler(service, nil)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/iam/users/test-user/roles", strings.NewReader(`{"displayName":"Test User","role":"viewer"}`))
	req = req.WithContext(tenant.WithTenantID(req.Context(), "test-tenant"))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-Id", "admin-local")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if !inv.called || inv.userID != "test-user" {
		t.Fatalf("expected cache invalidation for test-user, got called=%v user=%s", inv.called, inv.userID)
	}
}

func TestIAMAdminHandlerReplaceRoles(t *testing.T) {
	repo := iammemory.NewRoleAdminRepository()
	inv := &fakeInvalidator{}
	service := iamapp.NewAdminService(repo, inv)
	handler := iamdelivery.NewAdminHandler(service, nil)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/iam/users/test-user/roles", strings.NewReader(`{"displayName":"Test User","roles":["editor"]}`))
	req = req.WithContext(tenant.WithTenantID(req.Context(), "test-tenant"))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-Id", "admin-local")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rr.Code, rr.Body.String())
	}
	if !inv.called || inv.userID != "test-user" {
		t.Fatalf("expected cache invalidation for test-user, got called=%v user=%s", inv.called, inv.userID)
	}
}

func TestIAMAdminHandlerResetPassword(t *testing.T) {
	_ = iamapp.NewAdminService // legacy reference retained
	inv := &fakeInvalidator{}
	authRepo := authmemory.NewRepository()
	roleAdminRepo := iammemory.NewRoleAdminRepository()
	authService := newAuthServiceForTest(t, authRepo, roleAdminRepo, map[string][]iamdomain.Role{
		"test-user": {iamdomain.RoleViewer},
	})
	if err := authRepo.CreateUser(context.Background(), authdomain.CreateUserParams{
		UserID:             "test-user",
		Username:           "test-user",
		DisplayName:        "Test User",
		PasswordHash:       "$2a$10$W8v2u5Q2m4B8T7d7YQxM5eK/94B2SxA4yJ1mQ1MNhbbX6YsJhBKtC",
		PasswordAlgo:       "bcrypt",
		MustChangePassword: false,
		IsActive:           true,
		Roles:              []iamdomain.Role{iamdomain.RoleViewer},
		CreatedBy:          "system",
	}); err != nil {
		t.Fatalf("seed auth user: %v", err)
	}

	mux := http.NewServeMux()
	registerPeopleHandlerForTest(t, mux, authService, roleAdminRepo, inv)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/iam/users/test-user/reset-password", strings.NewReader(`{"newPassword":"Reset1234"}`))
	req = req.WithContext(tenant.WithTenantID(req.Context(), testTenantID))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-Id", "admin-local")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rr.Code, rr.Body.String())
	}
	identity, err := authRepo.FindIdentityByUserID(context.Background(), "test-user")
	if err != nil {
		t.Fatalf("load user: %v", err)
	}
	if !identity.MustChangePassword {
		t.Fatalf("expected must change password to be true after reset")
	}
}

func TestIAMAdminHandlerUnlockUser(t *testing.T) {
	inv := &fakeInvalidator{}
	authRepo := authmemory.NewRepository()
	roleAdminRepo := iammemory.NewRoleAdminRepository()
	authService := newAuthServiceForTest(t, authRepo, roleAdminRepo, map[string][]iamdomain.Role{
		"test-user": {iamdomain.RoleViewer},
	})
	if err := authRepo.CreateUser(context.Background(), authdomain.CreateUserParams{
		UserID:             "test-user",
		Username:           "test-user",
		DisplayName:        "Test User",
		PasswordHash:       "$2a$10$W8v2u5Q2m4B8T7d7YQxM5eK/94B2SxA4yJ1mQ1MNhbbX6YsJhBKtC",
		PasswordAlgo:       "bcrypt",
		MustChangePassword: false,
		IsActive:           true,
		Roles:              []iamdomain.Role{iamdomain.RoleViewer},
		CreatedBy:          "system",
	}); err != nil {
		t.Fatalf("seed auth user: %v", err)
	}
	if _, _, err := authRepo.RecordFailedLogin(context.Background(), "test-user", 1, 60, ""); err != nil {
		t.Fatalf("seed lock state: %v", err)
	}

	mux := http.NewServeMux()
	registerPeopleHandlerForTest(t, mux, authService, roleAdminRepo, inv)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/iam/users/test-user/unlock", nil)
	req = req.WithContext(tenant.WithTenantID(req.Context(), testTenantID))
	req.Header.Set("X-User-Id", "admin-local")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rr.Code, rr.Body.String())
	}
	identity, err := authRepo.FindIdentityByUserID(context.Background(), "test-user")
	if err != nil {
		t.Fatalf("load user: %v", err)
	}
	if identity.FailedLoginAttempts != 0 {
		t.Fatalf("expected failed attempts to be reset, got %d", identity.FailedLoginAttempts)
	}
	if identity.LockedUntil != nil {
		t.Fatalf("expected user to be unlocked")
	}
}
