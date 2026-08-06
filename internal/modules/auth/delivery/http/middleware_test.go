package httpdelivery

import (
	"context"
	sqlpkg "database/sql"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	authapp "metaldocs/internal/modules/auth/application"
	authdomain "metaldocs/internal/modules/auth/domain"
	"metaldocs/internal/modules/auth/infrastructure/memory"
	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/platform/iamtypes"
	platformtenant "metaldocs/internal/platform/tenant"

	"golang.org/x/crypto/bcrypt"
)

// middlewareMockRoleProvider satisfies iamdomain.RoleProvider with a fixed
// role set so Authenticate/ResolveSession succeed without a real DB.
type middlewareMockRoleProvider struct {
	roles map[string][]iamtypes.Role
}

func (m *middlewareMockRoleProvider) RolesByUserID(_ context.Context, userID, tenantID string) ([]iamtypes.Role, error) {
	roles, ok := m.roles[userID+":"+tenantID]
	if !ok {
		return nil, iamdomain.ErrUserNotFound
	}
	return roles, nil
}

func (m *middlewareMockRoleProvider) UserActiveInTenant(_ context.Context, _, _ string) (bool, error) {
	return true, nil
}

func (m *middlewareMockRoleProvider) RolesByUserIDs(_ context.Context, tenantID string, userIDs []string) (map[string][]iamtypes.Role, error) {
	out := make(map[string][]iamtypes.Role, len(userIDs))
	for _, uid := range userIDs {
		if roles, ok := m.roles[uid+":"+tenantID]; ok {
			out[uid] = roles
		}
	}
	return out, nil
}

// middlewareNoopLoginCtxPort satisfies iamdomain.LoginContextPort for tests
// that do not exercise governance metadata recording.
type middlewareNoopLoginCtxPort struct{}

func (middlewareNoopLoginCtxPort) RecordLoginContext(_ context.Context, _, _, _, _, _ string) error {
	return nil
}

// middlewareNoopRoleAdminRepository satisfies iamdomain.RoleAdminRepository;
// unused by the authenticate/resolve-session path under test.
type middlewareNoopRoleAdminRepository struct{}

func (middlewareNoopRoleAdminRepository) HasAnyRole(_ context.Context, _ iamtypes.Role, _ string) (bool, error) {
	return false, nil
}

func (middlewareNoopRoleAdminRepository) UpsertUserAndAssignRole(_ context.Context, _, _, _ string, _ iamtypes.Role, _ string) error {
	return nil
}

func (middlewareNoopRoleAdminRepository) ReplaceUserRoles(_ context.Context, _, _, _ string, _ iamtypes.Role, _ string) error {
	return nil
}

func (middlewareNoopRoleAdminRepository) UpsertUserAndAssignRoleTx(_ context.Context, _ *sqlpkg.Tx, _, _, _ string, _ iamtypes.Role, _ string) error {
	return nil
}

func (middlewareNoopRoleAdminRepository) ReplaceUserRolesTx(_ context.Context, _ *sqlpkg.Tx, _, _, _ string, _ iamtypes.Role, _ string) error {
	return nil
}

// passthrough is a sentinel handler that records whether it was reached.
func passthrough(reached *bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*reached = true
		w.WriteHeader(http.StatusOK)
	})
}

func TestMiddleware_PublicPathChecker_Injection(t *testing.T) {
	// A custom checker that marks /api/v1/feature-flags as public and nothing else.
	customChecker := func(r *http.Request) bool {
		return r.Method == http.MethodGet && r.URL.Path == "/api/v1/feature-flags"
	}

	m := NewMiddleware(nil, authapp.Config{}, true).
		WithPublicPathChecker(customChecker)

	reached := false
	handler := m.Wrap(passthrough(&reached))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/feature-flags", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 for public feature-flags, got %d", rec.Code)
	}
	if !reached {
		t.Error("passthrough handler was not reached for public path")
	}
}

func TestMiddleware_NoCookie_PrivateRoute_Returns401(t *testing.T) {
	m := NewMiddleware(nil, authapp.Config{SessionCookieName: "session"}, true).
		WithPublicPathChecker(func(*http.Request) bool { return false })

	reached := false
	handler := m.Wrap(passthrough(&reached))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/documents", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for unauthenticated private route, got %d", rec.Code)
	}
	if reached {
		t.Error("passthrough handler must NOT be reached for unauthenticated private route")
	}
}

func TestMiddleware_Disabled_PassesThrough(t *testing.T) {
	m := NewMiddleware(nil, authapp.Config{}, false)

	reached := false
	handler := m.Wrap(passthrough(&reached))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/documents", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if !reached {
		t.Error("disabled middleware must pass all requests through")
	}
}

// TestMiddleware_NoPublicPathChecker_FailsClosed pins the C3 idiom: a nil
// PublicPathChecker is a boot-time misconfiguration, never a silent
// "nothing is public" or "everything is public" fallback. defaultPublicPaths
// was deleted for exactly this reason — its doc comment carried a
// synchronization instruction ("keep this in sync with the composition
// root's authoritative list") that had already drifted.
func TestMiddleware_NoPublicPathChecker_FailsClosed(t *testing.T) {
	m := NewMiddleware(nil, authapp.Config{}, true) // no WithPublicPathChecker

	reached := false
	handler := m.Wrap(passthrough(&reached))
	req := httptest.NewRequest(http.MethodGet, "/api/v1/health/live", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("got %d, want 500 (nil checker must fail closed)", rec.Code)
	}
	if reached {
		t.Error("passthrough handler must NOT be reached when no checker is configured")
	}
}

// TestMiddleware_NoPasswordChangeAllowedChecker_FailsClosed mirrors the same
// idiom for the password-change-allowed checker: a MustChangePassword
// principal hitting a nil checker must be denied (500), never silently
// admitted.
func TestMiddleware_NoPasswordChangeAllowedChecker_FailsClosed(t *testing.T) {
	repo := memory.NewRepository()
	roleProvider := &middlewareMockRoleProvider{roles: map[string][]iamtypes.Role{}}
	roleAdmin := middlewareNoopRoleAdminRepository{}

	ctx := context.Background()
	userID := "mw-pw-change-test-user"
	password := "TestPassword123!"
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("GenerateFromPassword: %v", err)
	}
	if err := repo.CreateUser(ctx, authdomain.CreateUserParams{
		UserID:             userID,
		Username:           userID,
		Email:              "mw-pw-change@example.com",
		DisplayName:        "MW Password Change Test",
		PasswordHash:       authdomain.PasswordHash(string(hash)),
		PasswordAlgo:       "bcrypt",
		IsActive:           true,
		MustChangePassword: true,
	}); err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	roleProvider.roles[userID+":"+platformtenant.DevTenantID] = []iamtypes.Role{iamtypes.RoleViewer}

	svc, err := authapp.NewService(repo, roleProvider, roleAdmin, middlewareNoopLoginCtxPort{}, authapp.Config{
		SessionCookieName:      "session",
		SessionTTL:             24 * time.Hour,
		SessionSecret:          "0123456789abcdef0123456789abcdef",
		PasswordMinLength:      8,
		LoginMaxFailedAttempts: 5,
		LoginLockDuration:      15 * time.Minute,
		AllowDevTenantFallback: true,
		CookieSecure:           false,
	})
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}

	loginReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", nil)
	authSession, err := svc.Authenticate(ctx, userID, password, loginReq)
	if err != nil {
		t.Fatalf("Authenticate: %v", err)
	}

	m := NewMiddleware(svc, authapp.Config{SessionCookieName: "session"}, true).
		WithPublicPathChecker(func(*http.Request) bool { return false })
		// no WithPasswordChangeAllowedChecker

	reached := false
	handler := m.Wrap(passthrough(&reached))
	req := httptest.NewRequest(http.MethodGet, "/api/v1/documents", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: authSession.RawToken})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("got %d, want 500 (nil password-change checker must fail closed)", rec.Code)
	}
	if reached {
		t.Error("passthrough handler must NOT be reached when no password-change checker is configured")
	}
}

// TestMiddleware_AuthenticatedRequest_InjectsTenantAndActorCtx pins F3.1 T2:
// on a successfully authenticated request, Wrap must inject BOTH the tenant
// ID (existing behavior) and the actor ID (new — platformtenant.WithActorID)
// into the outgoing context, so the platform/db chokepoint (T3) can read
// both without importing internal/modules/iam.
func TestMiddleware_AuthenticatedRequest_InjectsTenantAndActorCtx(t *testing.T) {
	repo := memory.NewRepository()
	roleProvider := &middlewareMockRoleProvider{roles: map[string][]iamtypes.Role{}}
	roleAdmin := middlewareNoopRoleAdminRepository{}

	ctx := context.Background()
	userID := "mw-actor-test-user"
	password := "TestPassword123!"
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("GenerateFromPassword: %v", err)
	}
	if err := repo.CreateUser(ctx, authdomain.CreateUserParams{
		UserID:       userID,
		Username:     userID,
		Email:        "mw-actor@example.com",
		DisplayName:  "MW Actor Test",
		PasswordHash: authdomain.PasswordHash(string(hash)),
		PasswordAlgo: "bcrypt",
		IsActive:     true,
	}); err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	roleProvider.roles[userID+":"+platformtenant.DevTenantID] = []iamtypes.Role{iamtypes.RoleViewer}

	svc, err := authapp.NewService(repo, roleProvider, roleAdmin, middlewareNoopLoginCtxPort{}, authapp.Config{
		SessionCookieName:      "session",
		SessionTTL:             24 * time.Hour,
		SessionSecret:          "0123456789abcdef0123456789abcdef",
		PasswordMinLength:      8,
		LoginMaxFailedAttempts: 5,
		LoginLockDuration:      15 * time.Minute,
		AllowDevTenantFallback: true,
		CookieSecure:           false,
	})
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}

	loginReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", nil)
	authSession, err := svc.Authenticate(ctx, userID, password, loginReq)
	if err != nil {
		t.Fatalf("Authenticate: %v", err)
	}

	m := NewMiddleware(svc, authapp.Config{SessionCookieName: "session"}, true).
		WithPublicPathChecker(func(*http.Request) bool { return false })

	var gotTenant, gotActor string
	var tenantErr, actorErr error
	handler := m.Wrap(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotTenant, tenantErr = platformtenant.FromContext(r.Context())
		gotActor, actorErr = platformtenant.ActorFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/documents", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: authSession.RawToken})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if tenantErr != nil {
		t.Fatalf("tenant ctx missing: %v", tenantErr)
	}
	if actorErr != nil {
		t.Fatalf("actor ctx missing: %v", actorErr)
	}
	if gotTenant != platformtenant.DevTenantID {
		t.Errorf("tenant = %q, want %q", gotTenant, platformtenant.DevTenantID)
	}
	if gotActor != userID {
		t.Errorf("actor = %q, want %q", gotActor, userID)
	}
}
