package application

import (
	"context"
	"errors"
	"net/http/httptest"
	"testing"
	"time"

	authdomain "metaldocs/internal/modules/auth/domain"
	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/modules/auth/infrastructure/memory"
	"metaldocs/internal/platform/tenant"

	"golang.org/x/crypto/bcrypt"
)

// mockRoleProvider implements iamdomain.RoleProvider for testing.
type mockRoleProvider struct {
	roles map[string][]iamdomain.Role
}

func (m *mockRoleProvider) RolesByUserID(_ context.Context, userID, tenantID string) ([]iamdomain.Role, error) {
	key := userID + ":" + tenantID
	roles, ok := m.roles[key]
	if !ok {
		return nil, iamdomain.ErrUserNotFound
	}
	return roles, nil
}

func newMockRoleProvider() *mockRoleProvider {
	return &mockRoleProvider{
		roles: map[string][]iamdomain.Role{},
	}
}

// mockRoleAdminRepository implements iamdomain.RoleAdminRepository for testing.
type mockRoleAdminRepository struct {
	roles map[string][]iamdomain.Role
}

func (m *mockRoleAdminRepository) HasAnyRole(_ context.Context, role iamdomain.Role, _ string) (bool, error) {
	for _, roles := range m.roles {
		for _, r := range roles {
			if r == role {
				return true, nil
			}
		}
	}
	return false, nil
}

func (m *mockRoleAdminRepository) UpsertUserAndAssignRole(_ context.Context, userID, displayName, tenantID string, role iamdomain.Role, assignedBy string) error {
	key := userID + ":" + tenantID
	m.roles[key] = append(m.roles[key], role)
	return nil
}

func (m *mockRoleAdminRepository) ReplaceUserRoles(_ context.Context, userID, displayName, tenantID string, roles []iamdomain.Role, assignedBy string) error {
	key := userID + ":" + tenantID
	m.roles[key] = append([]iamdomain.Role(nil), roles...)
	return nil
}

func newMockRoleAdminRepository() *mockRoleAdminRepository {
	return &mockRoleAdminRepository{
		roles: map[string][]iamdomain.Role{},
	}
}

func mustHashPassword(t *testing.T, password string) string {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("GenerateFromPassword: %v", err)
	}
	return string(hash)
}

// TestAuthenticate_DevFallback_NoTenantClaim verifies that when a user has no
// roles (empty tenant list) and AllowDevTenantFallback is true, the login
// succeeds with session.TenantID set to the dev tenant constant.
func TestAuthenticate_DevFallback_NoTenantClaim(t *testing.T) {
	repo := memory.NewRepository()
	roleProvider := newMockRoleProvider()
	roleAdmin := newMockRoleAdminRepository()

	ctx := context.Background()

	// Create a test user
	userID := "dev-user"
	password := "TestPassword123!"
	hash := mustHashPassword(t, password)
	if err := repo.CreateUser(ctx, authdomain.CreateUserParams{
		UserID:       userID,
		Username:     userID,
		Email:        "dev@example.com",
		DisplayName:  "Dev User",
		PasswordHash: hash,
		PasswordAlgo: "bcrypt",
		IsActive:     true,
	}); err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	// Seed the role provider with the user having a role in the dev tenant.
	// This is required because buildCurrentUser calls RolesByUserID and fails
	// if the user has no roles.
	roleProvider.roles[userID+":"+tenant.DevTenantID] = []iamdomain.Role{iamdomain.RoleViewer}

	// Service with AllowDevTenantFallback=true
	svc := NewService(repo, roleProvider, roleAdmin, Config{
		SessionCookieName:      "session",
		SessionTTL:             24 * time.Hour,
		SessionSecret:          "test-secret-key-32-bytes-long!!",
		PasswordMinLength:      8,
		LoginMaxFailedAttempts: 5,
		LoginLockDuration:      15 * time.Minute,
		AllowDevTenantFallback: true,
		CookieSecure:           false,
	})

	// Authenticate without X-Tenant-ID header (GetUserTenants returns nil for memory repo)
	req := httptest.NewRequest("POST", "/api/v1/auth/login", nil)
	session, err := svc.Authenticate(ctx, userID, password, req)
	if err != nil {
		t.Fatalf("Authenticate: %v", err)
	}

	// Verify TenantID is set to dev tenant
	if session.CurrentUser.TenantID != tenant.DevTenantID {
		t.Errorf("TenantID mismatch: got %q, want %q", session.CurrentUser.TenantID, tenant.DevTenantID)
	}
}

// TestResolveSession_ReadsTenantFromSession verifies that ResolveSession reads
// the TenantID from the stored session row and returns it in CurrentUser.
// This test uses AllowDevTenantFallback so the user can authenticate without
// providing an X-Tenant-ID header (since memory repo GetUserTenants always returns nil).
func TestResolveSession_ReadsTenantFromSession(t *testing.T) {
	repo := memory.NewRepository()
	roleProvider := newMockRoleProvider()
	roleAdmin := newMockRoleAdminRepository()

	ctx := context.Background()

	// Create a test user
	userID := "session-test-user"
	password := "TestPassword123!"
	hash := mustHashPassword(t, password)
	if err := repo.CreateUser(ctx, authdomain.CreateUserParams{
		UserID:       userID,
		Username:     userID,
		Email:        "session@example.com",
		DisplayName:  "Session User",
		PasswordHash: hash,
		PasswordAlgo: "bcrypt",
		IsActive:     true,
	}); err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	// Mock role provider to return a role for this user in the dev tenant
	// (when AllowDevTenantFallback=true, user will be logged into the dev tenant)
	roleProvider.roles[userID+":"+tenant.DevTenantID] = []iamdomain.Role{iamdomain.RoleViewer}

	// Service with AllowDevTenantFallback=true so user can login without X-Tenant-ID
	svc := NewService(repo, roleProvider, roleAdmin, Config{
		SessionCookieName:      "session",
		SessionTTL:             24 * time.Hour,
		SessionSecret:          "test-secret-key-32-bytes-long!!",
		PasswordMinLength:      8,
		LoginMaxFailedAttempts: 5,
		LoginLockDuration:      15 * time.Minute,
		AllowDevTenantFallback: true,
		CookieSecure:           false,
	})

	// Authenticate without X-Tenant-ID header; will use dev tenant fallback.
	// This will bind session.TenantID to the dev tenant.
	req := httptest.NewRequest("POST", "/api/v1/auth/login", nil)
	authSession, err := svc.Authenticate(ctx, userID, password, req)
	if err != nil {
		t.Fatalf("Authenticate: %v", err)
	}

	// Verify the session contains the dev tenant
	if authSession.CurrentUser.TenantID != tenant.DevTenantID {
		t.Fatalf("Authenticate returned wrong tenant: got %q, want %q", authSession.CurrentUser.TenantID, tenant.DevTenantID)
	}

	// Now resolve the session using the raw token from authenticate.
	// ResolveSession should read the TenantID from the session row and include it in CurrentUser.
	resolved, err := svc.ResolveSession(ctx, authSession.RawToken)
	if err != nil {
		t.Fatalf("ResolveSession: %v", err)
	}

	// Verify TenantID is preserved from the session row
	if resolved.TenantID != tenant.DevTenantID {
		t.Errorf("TenantID mismatch: got %q, want %q", resolved.TenantID, tenant.DevTenantID)
	}
	if resolved.UserID != userID {
		t.Errorf("UserID mismatch: got %q, want %q", resolved.UserID, userID)
	}
}

// TestAuthenticate_InvalidCredentials verifies that authentication fails
// with invalid credentials and returns ErrInvalidCredentials.
func TestAuthenticate_InvalidCredentials(t *testing.T) {
	repo := memory.NewRepository()
	roleProvider := newMockRoleProvider()
	roleAdmin := newMockRoleAdminRepository()

	ctx := context.Background()

	// Create a test user
	userID := "invalid-test-user"
	hash := mustHashPassword(t, "TestPassword123!")
	if err := repo.CreateUser(ctx, authdomain.CreateUserParams{
		UserID:       userID,
		Username:     userID,
		Email:        "invalid@example.com",
		DisplayName:  "Invalid Test User",
		PasswordHash: hash,
		PasswordAlgo: "bcrypt",
		IsActive:     true,
	}); err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	svc := NewService(repo, roleProvider, roleAdmin, Config{
		SessionCookieName:      "session",
		SessionTTL:             24 * time.Hour,
		SessionSecret:          "test-secret-key-32-bytes-long!!",
		PasswordMinLength:      8,
		LoginMaxFailedAttempts: 5,
		LoginLockDuration:      15 * time.Minute,
		AllowDevTenantFallback: true,
		CookieSecure:           false,
	})

	req := httptest.NewRequest("POST", "/api/v1/auth/login", nil)
	_, err := svc.Authenticate(ctx, userID, "WrongPassword", req)
	if !errors.Is(err, authdomain.ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}
}

// TestAuthenticate_TenantClaimRequired verifies that when a user belongs to
// multiple tenants and no X-Tenant-ID header is provided, authentication fails
// with ErrTenantClaimRequired. This test requires a postgres setup with
// iam_user_roles table, so we skip it for memory-only tests.
func TestAuthenticate_TenantClaimRequired(t *testing.T) {
	t.Skip("requires postgres iam_user_roles; memory repo GetUserTenants returns nil")
}

// TestAuthenticate_TenantNotPermitted verifies that when a user tries to
// authenticate with an X-Tenant-ID header naming a tenant they don't have
// a role in, authentication fails with ErrTenantNotPermitted.
func TestAuthenticate_TenantNotPermitted(t *testing.T) {
	t.Skip("requires postgres iam_user_roles; memory repo GetUserTenants returns nil")
}
