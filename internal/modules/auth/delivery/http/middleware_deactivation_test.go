package httpdelivery

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	authapp "metaldocs/internal/modules/auth/application"
	authdomain "metaldocs/internal/modules/auth/domain"
	"metaldocs/internal/modules/auth/infrastructure/memory"
	"metaldocs/internal/platform/iamtypes"
	"metaldocs/internal/platform/tenant"

	"golang.org/x/crypto/bcrypt"
)

// stubRoleProvider / stubLoginCtxPort are the minimal real collaborators needed to
// mint a session via Authenticate. The inactive-resolve path returns before roles
// are consulted, so the role set only matters for the initial (active) login.
type stubRoleProvider struct{ roles []iamtypes.Role }

func (s stubRoleProvider) RolesByUserID(context.Context, string, string) ([]iamtypes.Role, error) {
	return s.roles, nil
}
func (stubRoleProvider) RolesByUserIDs(context.Context, string, []string) (map[string][]iamtypes.Role, error) {
	return map[string][]iamtypes.Role{}, nil
}
func (stubRoleProvider) UserActiveInTenant(context.Context, string, string) (bool, error) {
	return true, nil
}

type stubLoginCtxPort struct{}

func (stubLoginCtxPort) RecordLoginContext(context.Context, string, string, string, string, string) error {
	return nil
}

// TestMiddleware_DeactivatedIdentity_Returns401 is the F8.4 delivery-side proof:
// a request bearing a still-live session cookie for an identity that has since been
// deactivated is rejected with 401 (treated as revoked), not 500 — i.e.
// ErrIdentityInactive is in the resolve 401 set. End-to-end through a real service.
func TestMiddleware_DeactivatedIdentity_Returns401(t *testing.T) {
	repo := memory.NewRepository()
	ctx := context.Background()
	userID := "mw-deact-user"
	password := "TestPassword123!"

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	if err := repo.CreateUser(ctx, authdomain.CreateUserParams{
		UserID:       userID,
		Username:     userID,
		Email:        "mwdeact@example.com",
		DisplayName:  "MW Deact",
		PasswordHash: authdomain.PasswordHash(string(hash)),
		PasswordAlgo: "bcrypt",
		IsActive:     true,
	}); err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	cfg := authapp.Config{
		SessionCookieName:      "session",
		SessionTTL:             24 * time.Hour,
		SessionSecret:          authapp.Secret("0123456789abcdef0123456789abcdef"),
		PasswordMinLength:      8,
		LoginMaxFailedAttempts: 5,
		LoginLockDuration:      15 * time.Minute,
		AllowDevTenantFallback: true,
		CookieSecure:           false,
	}
	svc, err := authapp.NewService(repo, stubRoleProvider{roles: []iamtypes.Role{iamtypes.RoleViewer}}, nil, stubLoginCtxPort{}, cfg)
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}
	_ = tenant.DevTenantID // dev-tenant fallback is exercised by Authenticate

	loginReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", nil)
	authSession, err := svc.Authenticate(ctx, userID, password, loginReq)
	if err != nil {
		t.Fatalf("Authenticate: %v", err)
	}

	// Deactivate the identity at the repo (session row stays live) to isolate the
	// resolve-time rejection the middleware must map to 401.
	inactive := false
	if err := repo.UpdateUser(ctx, authdomain.UpdateUserParams{UserID: userID, IsActive: &inactive}); err != nil {
		t.Fatalf("repo.UpdateUser(deactivate): %v", err)
	}

	m := NewMiddleware(svc, cfg, true)
	reached := false
	handler := m.Wrap(passthrough(&reached))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/documents", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: authSession.RawToken})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("deactivated identity: status = %d, want 401", rec.Code)
	}
	if reached {
		t.Fatal("downstream handler must NOT be reached for a deactivated identity")
	}
}
