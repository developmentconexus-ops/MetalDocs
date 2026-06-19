package application

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"

	authdomain "metaldocs/internal/modules/auth/domain"
	"metaldocs/internal/modules/auth/infrastructure/memory"
	authpostgres "metaldocs/internal/modules/auth/infrastructure/postgres"
	iamdomain "metaldocs/internal/modules/iam/domain"
	iampostgres "metaldocs/internal/modules/iam/infrastructure/postgres"
	"metaldocs/internal/platform/tenant"

	"golang.org/x/crypto/bcrypt"
)

// mockRoleProvider implements iamdomain.RoleProvider for testing.
type mockRoleProvider struct {
	roles map[string][]iamdomain.Role
	errs  map[string]error
}

func (m *mockRoleProvider) RolesByUserID(_ context.Context, userID, tenantID string) ([]iamdomain.Role, error) {
	key := userID + ":" + tenantID
	if err, ok := m.errs[key]; ok {
		return nil, err
	}
	roles, ok := m.roles[key]
	if !ok {
		return nil, iamdomain.ErrUserNotFound
	}
	return roles, nil
}

func (m *mockRoleProvider) UserActiveInTenant(_ context.Context, _, _ string) (bool, error) {
	return true, nil
}

func (m *mockRoleProvider) RolesByUserIDs(_ context.Context, tenantID string, userIDs []string) (map[string][]iamdomain.Role, error) {
	out := make(map[string][]iamdomain.Role, len(userIDs))
	for _, uid := range userIDs {
		key := uid + ":" + tenantID
		if err, hasErr := m.errs[key]; hasErr {
			// ErrNoRolesAssigned → user is active but has no roles: include with empty slice.
			// Other errors (ErrUserNotFound, ErrUserInactive) → exclude.
			if errors.Is(err, iamdomain.ErrNoRolesAssigned) {
				out[uid] = []iamdomain.Role{}
			}
			continue
		}
		if roles, ok := m.roles[key]; ok {
			clone := make([]iamdomain.Role, len(roles))
			copy(clone, roles)
			out[uid] = clone
		}
	}
	return out, nil
}

func newMockRoleProvider() *mockRoleProvider {
	return &mockRoleProvider{
		roles: map[string][]iamdomain.Role{},
		errs:  map[string]error{},
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

func (m *mockRoleAdminRepository) ReplaceUserRoles(_ context.Context, userID, displayName, tenantID string, role iamdomain.Role, assignedBy string) error {
	key := userID + ":" + tenantID
	m.roles[key] = []iamdomain.Role{role}
	return nil
}

func (m *mockRoleAdminRepository) UpsertUserAndAssignRoleTx(_ context.Context, _ *sql.Tx, userID, _, tenantID string, role iamdomain.Role, _ string) error {
	key := userID + ":" + tenantID
	m.roles[key] = append(m.roles[key], role)
	return nil
}

func (m *mockRoleAdminRepository) ReplaceUserRolesTx(_ context.Context, _ *sql.Tx, userID, _, tenantID string, role iamdomain.Role, _ string) error {
	key := userID + ":" + tenantID
	m.roles[key] = []iamdomain.Role{role}
	return nil
}

func newMockRoleAdminRepository() *mockRoleAdminRepository {
	return &mockRoleAdminRepository{
		roles: map[string][]iamdomain.Role{},
	}
}

// noopLoginCtxPort satisfies iamdomain.LoginContextPort with no-op writes.
// Used in tests that do not require governance metadata recording.
type noopLoginCtxPort struct{}

func (noopLoginCtxPort) RecordLoginContext(_ context.Context, _, _, _, _, _ string) error {
	return nil
}

const testSessionSecret = "0123456789abcdef0123456789abcdef"

func mustHashPassword(t *testing.T, password string) authdomain.PasswordHash {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("GenerateFromPassword: %v", err)
	}
	return authdomain.PasswordHash(string(hash))
}

func mustNewService(t *testing.T, repo authdomain.Repository, roleProvider iamdomain.RoleProvider, roleAdmin iamdomain.RoleAdminRepository, cfg Config) *Service {
	t.Helper()
	svc, err := NewService(repo, roleProvider, roleAdmin, noopLoginCtxPort{}, cfg)
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}
	return svc
}

func TestNewService_RejectsShortSessionSecret(t *testing.T) {
	svc, err := NewService(memory.NewRepository(), newMockRoleProvider(), newMockRoleAdminRepository(), noopLoginCtxPort{}, Config{
		SessionSecret: "too-short",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if svc != nil {
		t.Fatalf("expected nil service, got %#v", svc)
	}
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
	svc := mustNewService(t, repo, roleProvider, roleAdmin, Config{
		SessionCookieName:      "session",
		SessionTTL:             24 * time.Hour,
		SessionSecret:          testSessionSecret,
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
	if session.CurrentUser.TenantName != "System Tenant" {
		t.Errorf("TenantName mismatch: got %q, want %q", session.CurrentUser.TenantName, "System Tenant")
	}
}

func TestAuthenticate_PreservesPasswordWhitespace(t *testing.T) {
	repo := memory.NewRepository()
	roleProvider := newMockRoleProvider()
	roleAdmin := newMockRoleAdminRepository()
	ctx := context.Background()

	userID := "space-user"
	password := " secret "
	if err := repo.CreateUser(ctx, authdomain.CreateUserParams{
		UserID:       userID,
		Username:     userID,
		Email:        "space@example.com",
		DisplayName:  "Space User",
		PasswordHash: mustHashPassword(t, password),
		PasswordAlgo: "bcrypt",
		IsActive:     true,
	}); err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	roleProvider.roles[userID+":"+tenant.DevTenantID] = []iamdomain.Role{iamdomain.RoleViewer}
	svc := mustNewService(t, repo, roleProvider, roleAdmin, Config{
		SessionCookieName:      "session",
		SessionTTL:             24 * time.Hour,
		SessionSecret:          testSessionSecret,
		PasswordMinLength:      1,
		LoginMaxFailedAttempts: 5,
		LoginLockDuration:      15 * time.Minute,
		AllowDevTenantFallback: true,
	})

	req := httptest.NewRequest("POST", "/api/v1/auth/login", nil)
	if _, err := svc.Authenticate(ctx, userID, strings.TrimSpace(password), req); !errors.Is(err, authdomain.ErrInvalidCredentials) {
		t.Fatalf("trimmed password error = %v, want ErrInvalidCredentials", err)
	}
	if _, err := svc.Authenticate(ctx, userID, password, req); err != nil {
		t.Fatalf("Authenticate with exact password: %v", err)
	}
}

// TestAuthenticate_TimingConstant guards the constant-time login fix (A1): an
// unknown identifier must spend bcrypt-equivalent time rather than returning
// immediately, so wall-clock latency cannot reveal whether an account exists.
// We assert both the unknown-identifier path and the known-user/wrong-password
// path are bcrypt-dominated (> 50ms floor) and within a generous 3x ratio of
// each other (lenient to absorb scheduler/GC jitter without flaking).
func TestAuthenticate_TimingConstant(t *testing.T) {
	repo := memory.NewRepository()
	roleProvider := newMockRoleProvider()
	roleAdmin := newMockRoleAdminRepository()
	ctx := context.Background()

	// Seed a real user with a cost-12 hash so the known-user wrong-password path
	// runs the same bcrypt cost as the precomputed dummy hash.
	const userID = "timing-user"
	knownHash, err := bcrypt.GenerateFromPassword([]byte("CorrectPassword123!"), 12)
	if err != nil {
		t.Fatalf("GenerateFromPassword: %v", err)
	}
	if err := repo.CreateUser(ctx, authdomain.CreateUserParams{
		UserID:       userID,
		Username:     userID,
		Email:        "timing@example.com",
		DisplayName:  "Timing User",
		PasswordHash: authdomain.PasswordHash(string(knownHash)),
		PasswordAlgo: "bcrypt",
		IsActive:     true,
	}); err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	svc := mustNewService(t, repo, roleProvider, roleAdmin, Config{
		SessionCookieName:      "session",
		SessionTTL:             24 * time.Hour,
		SessionSecret:          testSessionSecret,
		PasswordMinLength:      8,
		LoginMaxFailedAttempts: 100, // intentionally high so the repeated timing-probe attempts below never trip account lockout
		LoginLockDuration:      15 * time.Minute,
		AllowDevTenantFallback: true,
	})
	req := httptest.NewRequest("POST", "/api/v1/auth/login", nil)

	// minElapsed runs the call a few times and returns the fastest run, which is
	// the most stable estimate (least contaminated by transient pauses).
	minElapsed := func(identifier string) time.Duration {
		best := time.Hour
		for i := 0; i < 3; i++ {
			start := time.Now()
			_, gotErr := svc.Authenticate(ctx, identifier, "WrongPassword!", req)
			if !errors.Is(gotErr, authdomain.ErrInvalidCredentials) {
				t.Fatalf("Authenticate(%q) error = %v, want ErrInvalidCredentials", identifier, gotErr)
			}
			if d := time.Since(start); d < best {
				best = d
			}
		}
		return best
	}

	unknown := minElapsed("does-not-exist")
	known := minElapsed(userID)

	const floor = 50 * time.Millisecond
	if unknown < floor {
		t.Fatalf("unknown-identifier path too fast (%v < %v): timing oracle not closed", unknown, floor)
	}
	if known < floor {
		t.Fatalf("known-user path too fast (%v < %v)", known, floor)
	}
	ratio := float64(unknown) / float64(known)
	if known > unknown {
		ratio = float64(known) / float64(unknown)
	}
	if ratio > 3.0 {
		t.Fatalf("timing paths diverge too much (ratio %.2f, unknown=%v known=%v)", ratio, unknown, known)
	}
}

// TestAuthenticate_InactiveConstantTime proves the inactive (and, by the same
// code path, locked) branch is checked AFTER the bcrypt comparison, so an
// inactive account cannot be distinguished from an active wrong-password attempt
// by wall-clock. The distinct ErrIdentityInactive is intentionally kept (admin
// UX, operator decision) — only the timing is equalized.
func TestAuthenticate_InactiveConstantTime(t *testing.T) {
	repo := memory.NewRepository()
	roleProvider := newMockRoleProvider()
	roleAdmin := newMockRoleAdminRepository()
	ctx := context.Background()

	const userID = "inactive-user"
	knownHash, err := bcrypt.GenerateFromPassword([]byte("CorrectPassword123!"), 12)
	if err != nil {
		t.Fatalf("GenerateFromPassword: %v", err)
	}
	if err := repo.CreateUser(ctx, authdomain.CreateUserParams{
		UserID:       userID,
		Username:     userID,
		Email:        "inactive@example.com",
		DisplayName:  "Inactive User",
		PasswordHash: authdomain.PasswordHash(string(knownHash)),
		PasswordAlgo: "bcrypt",
		IsActive:     false,
	}); err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	svc := mustNewService(t, repo, roleProvider, roleAdmin, Config{
		SessionCookieName:      "session",
		SessionTTL:             24 * time.Hour,
		SessionSecret:          testSessionSecret,
		PasswordMinLength:      8,
		LoginMaxFailedAttempts: 100,
		LoginLockDuration:      15 * time.Minute,
		AllowDevTenantFallback: true,
	})
	req := httptest.NewRequest("POST", "/api/v1/auth/login", nil)

	start := time.Now()
	_, gotErr := svc.Authenticate(ctx, userID, "WrongPassword!", req)
	elapsed := time.Since(start)
	if !errors.Is(gotErr, authdomain.ErrIdentityInactive) {
		t.Fatalf("Authenticate(inactive) error = %v, want ErrIdentityInactive", gotErr)
	}
	if elapsed < 50*time.Millisecond {
		t.Fatalf("inactive path too fast (%v): bcrypt not run before the inactive check — timing leak", elapsed)
	}
}

// TestAuthenticate_ConcurrentWrongPasswordBoundedByLock proves the lockout is
// atomic: firing many parallel wrong-password attempts must NOT drive
// failed_login_attempts past the threshold. Pre-fix (stale-snapshot lock check),
// every concurrent attempt passed the gate and incremented, so the counter
// climbed to N. With the per-identity lock, attempts serialize and the ones
// after the threshold short-circuit on the locked state without recording —
// the counter caps at exactly LoginMaxFailedAttempts.
func TestAuthenticate_ConcurrentWrongPasswordBoundedByLock(t *testing.T) {
	t.Parallel()
	repo := memory.NewRepository()
	roleProvider := newMockRoleProvider()
	roleAdmin := newMockRoleAdminRepository()
	ctx := context.Background()

	const userID = "burst-user"
	const maxAttempts = 5
	if err := repo.CreateUser(ctx, authdomain.CreateUserParams{
		UserID:       userID,
		Username:     userID,
		Email:        "burst@example.com",
		DisplayName:  "Burst User",
		PasswordHash: mustHashPassword(t, "CorrectPassword123!"),
		PasswordAlgo: "bcrypt",
		IsActive:     true,
	}); err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	svc := mustNewService(t, repo, roleProvider, roleAdmin, Config{
		SessionCookieName:      "session",
		SessionTTL:             24 * time.Hour,
		SessionSecret:          testSessionSecret,
		PasswordMinLength:      8,
		LoginMaxFailedAttempts: maxAttempts,
		LoginLockDuration:      15 * time.Minute,
		AllowDevTenantFallback: true,
	})

	const burst = 50
	var wg sync.WaitGroup
	var invalid, locked, other int64
	wg.Add(burst)
	for i := 0; i < burst; i++ {
		go func() {
			defer wg.Done()
			req := httptest.NewRequest("POST", "/api/v1/auth/login", nil)
			_, err := svc.Authenticate(ctx, userID, "WrongPassword!", req)
			switch {
			case errors.Is(err, authdomain.ErrInvalidCredentials):
				atomic.AddInt64(&invalid, 1)
			case errors.Is(err, authdomain.ErrIdentityLocked):
				atomic.AddInt64(&locked, 1)
			default:
				atomic.AddInt64(&other, 1)
			}
		}()
	}
	wg.Wait()

	if other != 0 {
		t.Fatalf("unexpected non-credential errors: %d", other)
	}
	// Exactly maxAttempts attempts may verify the password and increment; the
	// rest must be rejected as already-locked. Counts depend on serialized order
	// but the partition is exact.
	if invalid != maxAttempts {
		t.Fatalf("ErrInvalidCredentials = %d, want %d (lockout must cap recorded failures)", invalid, maxAttempts)
	}
	if locked != burst-maxAttempts {
		t.Fatalf("ErrIdentityLocked = %d, want %d", locked, burst-maxAttempts)
	}

	identity, err := repo.FindIdentityByUserID(ctx, userID)
	if err != nil {
		t.Fatalf("FindIdentityByUserID: %v", err)
	}
	if identity.FailedLoginAttempts != maxAttempts {
		t.Fatalf("failed_login_attempts = %d, want %d (must not exceed threshold under concurrency)", identity.FailedLoginAttempts, maxAttempts)
	}
	if identity.LockedUntil == nil {
		t.Fatal("account must be locked after threshold breached")
	}
}

func TestHashPassword_UsesCost12(t *testing.T) {
	svc := mustNewService(t, memory.NewRepository(), newMockRoleProvider(), newMockRoleAdminRepository(), Config{
		SessionSecret: testSessionSecret,
	})
	hash, err := svc.hashPassword("Password123!")
	if err != nil {
		t.Fatalf("hashPassword: %v", err)
	}
	cost, err := bcrypt.Cost([]byte(hash))
	if err != nil {
		t.Fatalf("bcrypt cost: %v", err)
	}
	if cost != bcryptCost {
		t.Fatalf("cost = %d, want %d", cost, bcryptCost)
	}
}

func TestSessionCookiesUseStrictSameSite(t *testing.T) {
	svc := mustNewService(t, memory.NewRepository(), newMockRoleProvider(), newMockRoleAdminRepository(), Config{
		SessionCookieName: "session",
		SessionSecret:     testSessionSecret,
	})
	expiresAt := time.Now().UTC().Add(time.Hour)
	if got := svc.SessionCookie("raw", expiresAt).SameSite; got != http.SameSiteStrictMode {
		t.Fatalf("SessionCookie SameSite = %v, want %v", got, http.SameSiteStrictMode)
	}
	if got := svc.ExpiredSessionCookie().SameSite; got != http.SameSiteStrictMode {
		t.Fatalf("ExpiredSessionCookie SameSite = %v, want %v", got, http.SameSiteStrictMode)
	}
}

func TestConfig_RedactsBootstrapPassword(t *testing.T) {
	cfg := Config{SessionSecret: testSessionSecret, BootstrapAdminPassword: "secret"}
	rendered := cfg.String()
	if strings.Contains(rendered, "secret") || !strings.Contains(rendered, "***") {
		t.Fatalf("String() did not redact password: %s", rendered)
	}
	raw, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("MarshalJSON: %v", err)
	}
	if strings.Contains(string(raw), "secret") || !strings.Contains(string(raw), "***") {
		t.Fatalf("MarshalJSON did not redact password: %s", raw)
	}
}

func TestBootstrapLocalAdmin_ClearsPlaintextPassword(t *testing.T) {
	repo := memory.NewRepository()
	roleProvider := newMockRoleProvider()
	roleAdmin := newMockRoleAdminRepository()
	svc := mustNewService(t, repo, roleProvider, roleAdmin, Config{
		SessionSecret:          testSessionSecret,
		BootstrapAdminEnabled:  true,
		BootstrapAdminUserID:   "admin",
		BootstrapAdminUsername: "admin",
		BootstrapAdminEmail:    "admin@example.com",
		BootstrapAdminPassword: " secret ",
		BootstrapAdminName:     "Admin",
	})

	if err := svc.BootstrapLocalAdmin(context.Background()); err != nil {
		t.Fatalf("BootstrapLocalAdmin: %v", err)
	}
	if svc.cfg.BootstrapAdminPassword != "" {
		t.Fatalf("BootstrapAdminPassword retained plaintext")
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
	svc := mustNewService(t, repo, roleProvider, roleAdmin, Config{
		SessionCookieName:      "session",
		SessionTTL:             24 * time.Hour,
		SessionSecret:          testSessionSecret,
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
	if resolved.TenantName != "System Tenant" {
		t.Errorf("TenantName mismatch: got %q, want %q", resolved.TenantName, "System Tenant")
	}
	if resolved.UserID != userID {
		t.Errorf("UserID mismatch: got %q, want %q", resolved.UserID, userID)
	}
}

// TestResolveSession_IdleTimeout guards the sliding idle timeout (A2). A session
// whose absolute TTL is still valid but whose last activity is older than the
// idle window must be rejected with ErrSessionExpired; activity within the window
// resolves normally. The check is measured against the pre-touch LastSeenAt.
func TestResolveSession_IdleTimeout(t *testing.T) {
	const idleWindow = 30 * time.Minute
	base := time.Date(2026, 6, 8, 9, 0, 0, 0, time.UTC)

	newSession := func(t *testing.T) (*Service, string) {
		t.Helper()
		repo := memory.NewRepository()
		roleProvider := newMockRoleProvider()
		roleAdmin := newMockRoleAdminRepository()
		ctx := context.Background()

		userID := "idle-user"
		password := "TestPassword123!"
		if err := repo.CreateUser(ctx, authdomain.CreateUserParams{
			UserID:       userID,
			Username:     userID,
			Email:        "idle@example.com",
			DisplayName:  "Idle User",
			PasswordHash: mustHashPassword(t, password),
			PasswordAlgo: "bcrypt",
			IsActive:     true,
		}); err != nil {
			t.Fatalf("CreateUser: %v", err)
		}
		roleProvider.roles[userID+":"+tenant.DevTenantID] = []iamdomain.Role{iamdomain.RoleViewer}

		svc := mustNewService(t, repo, roleProvider, roleAdmin, Config{
			SessionCookieName:      "session",
			SessionTTL:             24 * time.Hour, // absolute TTL stays valid throughout
			SessionIdleTimeout:     idleWindow,
			SessionSecret:          testSessionSecret,
			PasswordMinLength:      8,
			LoginMaxFailedAttempts: 5,
			LoginLockDuration:      15 * time.Minute,
			AllowDevTenantFallback: true,
		})
		// Pin the clock so the session's LastSeenAt == base.
		svc.now = func() time.Time { return base }
		req := httptest.NewRequest("POST", "/api/v1/auth/login", nil)
		authSession, err := svc.Authenticate(ctx, userID, password, req)
		if err != nil {
			t.Fatalf("Authenticate: %v", err)
		}
		return svc, authSession.RawToken
	}

	t.Run("idle beyond window expires", func(t *testing.T) {
		svc, token := newSession(t)
		svc.now = func() time.Time { return base.Add(idleWindow + time.Minute) }
		if _, err := svc.ResolveSession(context.Background(), token); !errors.Is(err, authdomain.ErrSessionExpired) {
			t.Fatalf("ResolveSession after idle window: err = %v, want ErrSessionExpired", err)
		}
	})

	t.Run("within window resolves", func(t *testing.T) {
		svc, token := newSession(t)
		svc.now = func() time.Time { return base.Add(idleWindow - time.Minute) }
		user, err := svc.ResolveSession(context.Background(), token)
		if err != nil {
			t.Fatalf("ResolveSession within idle window: %v", err)
		}
		if user.UserID != "idle-user" {
			t.Fatalf("UserID = %q, want idle-user", user.UserID)
		}
	})
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

	svc := mustNewService(t, repo, roleProvider, roleAdmin, Config{
		SessionCookieName:      "session",
		SessionTTL:             24 * time.Hour,
		SessionSecret:          testSessionSecret,
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

func TestAuthenticate_TenantClaimRequired(t *testing.T) {
	repo := memory.NewRepository()
	roleProvider := newMockRoleProvider()
	roleAdmin := newMockRoleAdminRepository()

	ctx := context.Background()

	userID := "tenant-claim-user"
	password := "TestPassword123!"
	hash := mustHashPassword(t, password)
	if err := repo.CreateUser(ctx, authdomain.CreateUserParams{
		UserID:       userID,
		Username:     userID,
		Email:        "tenant-claim@example.com",
		DisplayName:  "Tenant Claim User",
		PasswordHash: hash,
		PasswordAlgo: "bcrypt",
		IsActive:     true,
	}); err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	repo.SeedUserTenants(userID, []string{"tenant-a", "tenant-b"})

	svc := mustNewService(t, repo, roleProvider, roleAdmin, Config{
		SessionCookieName:      "session",
		SessionTTL:             24 * time.Hour,
		SessionSecret:          testSessionSecret,
		PasswordMinLength:      8,
		LoginMaxFailedAttempts: 5,
		LoginLockDuration:      15 * time.Minute,
		CookieSecure:           false,
	})

	req := httptest.NewRequest("POST", "/api/v1/auth/login", nil)
	_, err := svc.Authenticate(ctx, userID, password, req)
	if !errors.Is(err, authdomain.ErrTenantClaimRequired) {
		t.Errorf("expected ErrTenantClaimRequired, got %v", err)
	}
}

func TestAuthenticate_TenantNotPermitted(t *testing.T) {
	repo := memory.NewRepository()
	roleProvider := newMockRoleProvider()
	roleAdmin := newMockRoleAdminRepository()

	ctx := context.Background()

	userID := "tenant-not-permitted-user"
	password := "TestPassword123!"
	hash := mustHashPassword(t, password)
	if err := repo.CreateUser(ctx, authdomain.CreateUserParams{
		UserID:       userID,
		Username:     userID,
		Email:        "tenant-not-permitted@example.com",
		DisplayName:  "Tenant Not Permitted User",
		PasswordHash: hash,
		PasswordAlgo: "bcrypt",
		IsActive:     true,
	}); err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	repo.SeedUserTenants(userID, []string{"tenant-a"})

	svc := mustNewService(t, repo, roleProvider, roleAdmin, Config{
		SessionCookieName:      "session",
		SessionTTL:             24 * time.Hour,
		SessionSecret:          testSessionSecret,
		PasswordMinLength:      8,
		LoginMaxFailedAttempts: 5,
		LoginLockDuration:      15 * time.Minute,
		CookieSecure:           false,
	})

	req := httptest.NewRequest("POST", "/api/v1/auth/login", nil)
	req.Header.Set("X-Tenant-ID", "tenant-z")
	_, err := svc.Authenticate(ctx, userID, password, req)
	if !errors.Is(err, authdomain.ErrTenantNotPermitted) {
		t.Errorf("expected ErrTenantNotPermitted, got %v", err)
	}
}

func TestListUsers_FiltersByTenantMembership(t *testing.T) {
	repo := memory.NewRepository()
	roleProvider := newMockRoleProvider()
	roleAdmin := newMockRoleAdminRepository()
	ctx := context.Background()

	for _, user := range []string{"alice", "bob", "carol"} {
		if err := repo.CreateUser(ctx, authdomain.CreateUserParams{
			UserID:       user,
			Username:     user,
			Email:        user + "@example.com",
			DisplayName:  user,
			PasswordHash: mustHashPassword(t, "Password123!"),
			PasswordAlgo: "bcrypt",
			IsActive:     true,
		}); err != nil {
			t.Fatalf("CreateUser(%s): %v", user, err)
		}
	}
	roleProvider.roles["alice:tenant-a"] = []iamdomain.Role{iamdomain.RoleViewer}
	roleProvider.errs["carol:tenant-a"] = iamdomain.ErrNoRolesAssigned

	svc := mustNewService(t, repo, roleProvider, roleAdmin, Config{
		SessionCookieName:      "session",
		SessionTTL:             24 * time.Hour,
		SessionSecret:          testSessionSecret,
		PasswordMinLength:      8,
		LoginMaxFailedAttempts: 5,
		LoginLockDuration:      15 * time.Minute,
		CookieSecure:           false,
	})

	items, err := svc.ListUsers(ctx, "tenant-a")
	if err != nil {
		t.Fatalf("ListUsers: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 users in tenant-a, got %d", len(items))
	}
	byUserID := map[string]authdomain.ManagedUser{}
	for _, item := range items {
		byUserID[item.UserID] = item
	}
	alice, ok := byUserID["alice"]
	if !ok {
		t.Fatal("expected alice in tenant-a listing")
	}
	if len(alice.Roles) != 1 || alice.Roles[0] != iamdomain.RoleViewer {
		t.Fatalf("alice roles = %v, want [viewer]", alice.Roles)
	}
	carol, ok := byUserID["carol"]
	if !ok {
		t.Fatal("expected carol in tenant-a listing")
	}
	if len(carol.Roles) != 0 {
		t.Fatalf("carol roles = %v, want empty", carol.Roles)
	}
}

func TestCreateUserWithInput_DefaultsUserIDFromUsername(t *testing.T) {
	repo := memory.NewRepository()
	roleProvider := newMockRoleProvider()
	roleAdmin := newMockRoleAdminRepository()
	svc := mustNewService(t, repo, roleProvider, roleAdmin, Config{
		PasswordMinLength:      8,
		LoginMaxFailedAttempts: 5,
		LoginLockDuration:      15 * time.Minute,
		SessionSecret:          testSessionSecret,
		SessionTTL:             24 * time.Hour,
		SessionCookieName:      "session",
		CookieSecure:           false,
	})

	err := svc.CreateUserWithInput(context.Background(), authdomain.CreateUserInput{
		Username: "new-user",
		Password: authdomain.PlainPassword("Password123!"),
		Roles:    []iamdomain.Role{iamdomain.RoleViewer},
	})
	if err != nil {
		t.Fatalf("CreateUserWithInput: %v", err)
	}

	identity, err := repo.FindIdentityByUserID(context.Background(), "new-user")
	if err != nil {
		t.Fatalf("FindIdentityByUserID: %v", err)
	}
	if identity.Username != "new-user" {
		t.Fatalf("username = %q, want new-user", identity.Username)
	}
}

func TestCreateUserWithInput_RejectsMultipleRoles(t *testing.T) {
	repo := memory.NewRepository()
	roleProvider := newMockRoleProvider()
	roleAdmin := newMockRoleAdminRepository()
	svc := mustNewService(t, repo, roleProvider, roleAdmin, Config{
		PasswordMinLength:      8,
		LoginMaxFailedAttempts: 5,
		LoginLockDuration:      15 * time.Minute,
		SessionSecret:          testSessionSecret,
		SessionTTL:             24 * time.Hour,
		SessionCookieName:      "session",
		CookieSecure:           false,
	})

	err := svc.CreateUserWithInput(context.Background(), authdomain.CreateUserInput{
		Username: "new-user",
		Password: authdomain.PlainPassword("Password123!"),
		Roles:    []iamdomain.Role{iamdomain.RoleViewer, iamdomain.RoleEditor},
	})
	if !errors.Is(err, iamdomain.ErrInvalidRole) {
		t.Fatalf("CreateUserWithInput error = %v, want ErrInvalidRole", err)
	}
}

func TestCreateUser_RollbackWhenReplaceUserRolesFails(t *testing.T) {
	const tenantID = "11111111-1111-1111-1111-111111111111"
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`
INSERT INTO metaldocs.auth_identities (user_id, username, email, display_name, is_active, password_hash, password_algo, must_change_password, last_login_at, failed_login_attempts, locked_until, created_at, updated_at)
VALUES ($1, $2, NULLIF($3, ''), $4, $5, $6, $7, $8, NULL, 0, NULL, NOW(), NOW())
`)).
		WithArgs("alice", "alice", "alice@example.com", "Alice", true, sqlmock.AnyArg(), "bcrypt", true).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(`
SELECT
	set_config('metaldocs.tenant_id', $1, true),
	set_config('metaldocs.actor_id', $2, true)
`)).
		WithArgs(tenantID, "admin").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT current_setting('metaldocs.actor_id', true)")).
		WillReturnRows(sqlmock.NewRows([]string{"current_setting"}).AddRow("admin"))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT current_setting('metaldocs.tenant_id', true)")).
		WillReturnRows(sqlmock.NewRows([]string{"current_setting"}).AddRow(tenantID))
	mock.ExpectQuery(regexp.QuoteMeta(`
SELECT EXISTS (
  SELECT 1
    FROM metaldocs.iam_user_roles ur
   WHERE ur.user_id   = $1
     AND ur.tenant_id = $2::uuid
     AND ur.role_code = 'system_admin'
  UNION ALL
  SELECT 1
    FROM metaldocs.iam_group_members gm
    JOIN metaldocs.iam_groups g ON g.id = gm.group_id
    JOIN metaldocs.iam_group_roles gr ON gr.group_id = gm.group_id
   WHERE gm.user_id = $1
     AND gm.tenant_id = $2::uuid
     AND g.tenant_id = $2::uuid
     AND gr.role = 'system_admin'
)`)).
		WithArgs("admin", tenantID).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
	mock.ExpectQuery(regexp.QuoteMeta(`
SELECT EXISTS (
  SELECT 1
  FROM metaldocs.role_capabilities rc
  JOIN metaldocs.user_process_areas upa
    ON upa.role = rc.role
   AND upa.tenant_id = $4::uuid
   AND upa.user_id   = $3
   AND upa.effective_from <= now()
   AND upa.effective_to IS NULL
  WHERE rc.capability = $1
    AND ($2 = 'tenant' OR upa.area_code = $2)
)`)).
		WithArgs(string(iamdomain.CapUserManage), "tenant", "admin", tenantID).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT current_setting('metaldocs.asserted_caps', true)")).
		WillReturnRows(sqlmock.NewRows([]string{"current_setting"}).AddRow(""))
	mock.ExpectExec(regexp.QuoteMeta("SELECT set_config('metaldocs.asserted_caps', $1, true)")).
		WithArgs(`[{"area":"tenant","cap":"user.manage"}]`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO metaldocs.iam_users`)).
		WithArgs("alice", "Alice").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM metaldocs.iam_user_roles WHERE tenant_id = $1::uuid AND user_id = $2`)).
		WithArgs(tenantID, "alice").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO metaldocs.iam_user_roles (user_id, tenant_id, role_code, assigned_at, assigned_by)`)).
		WithArgs("alice", tenantID, "author", "admin").
		WillReturnError(sqlmock.ErrCancelled)
	mock.ExpectRollback()

	repo := authpostgres.NewRepository(db, nil)
	roleProvider := newMockRoleProvider()
	roleAdmin := iampostgres.NewRoleAdminRepository(db)
	svc := mustNewService(t, repo, roleProvider, roleAdmin, Config{
		PasswordMinLength:      8,
		LoginMaxFailedAttempts: 5,
		LoginLockDuration:      15 * time.Minute,
		SessionSecret:          testSessionSecret,
		SessionTTL:             24 * time.Hour,
		SessionCookieName:      "session",
		CookieSecure:           false,
	})

	err = svc.CreateUser(context.Background(), "alice", "alice", "alice@example.com", "Alice", "Password123!", tenantID, []iamdomain.Role{iamdomain.Role("author")}, "admin")
	if err == nil {
		t.Fatal("expected error")
	}
	if got, want := err.Error(), "insert iam role"; !regexp.MustCompile(want).MatchString(got) {
		t.Fatalf("expected %q in error, got %q", want, got)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestLogout_EmptyAndMalformedTokenReturnError(t *testing.T) {
	svc := mustNewService(t, memory.NewRepository(), newMockRoleProvider(), newMockRoleAdminRepository(), Config{
		SessionSecret: testSessionSecret,
	})

	if err := svc.Logout(context.Background(), ""); !errors.Is(err, authdomain.ErrSessionNotFound) {
		t.Fatalf("Logout(empty) error = %v, want ErrSessionNotFound", err)
	}
	if err := svc.Logout(context.Background(), "not-a-valid-cookie"); !errors.Is(err, authdomain.ErrSessionNotFound) {
		t.Fatalf("Logout(malformed) error = %v, want ErrSessionNotFound", err)
	}
}

// TestChangePasswordForUser_RevokesSessions guards A3 (CWE-613): a self-service
// password change must revoke the user's existing sessions so a stolen or stale
// session cannot survive the change.
func TestChangePasswordForUser_RevokesSessions(t *testing.T) {
	repo := memory.NewRepository()
	roleProvider := newMockRoleProvider()
	roleAdmin := newMockRoleAdminRepository()
	ctx := context.Background()

	userID := "pw-change-user"
	password := "TestPassword123!"
	hash := mustHashPassword(t, password)
	if err := repo.CreateUser(ctx, authdomain.CreateUserParams{
		UserID:       userID,
		Username:     userID,
		Email:        "pwchange@example.com",
		DisplayName:  "PW Change User",
		PasswordHash: hash,
		PasswordAlgo: "bcrypt",
		IsActive:     true,
	}); err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	roleProvider.roles[userID+":"+tenant.DevTenantID] = []iamdomain.Role{iamdomain.RoleViewer}

	svc := mustNewService(t, repo, roleProvider, roleAdmin, Config{
		SessionCookieName:      "session",
		SessionTTL:             24 * time.Hour,
		SessionSecret:          testSessionSecret,
		PasswordMinLength:      8,
		LoginMaxFailedAttempts: 5,
		LoginLockDuration:      15 * time.Minute,
		AllowDevTenantFallback: true,
		CookieSecure:           false,
	})

	req := httptest.NewRequest("POST", "/api/v1/auth/login", nil)
	authSession, err := svc.Authenticate(ctx, userID, password, req)
	if err != nil {
		t.Fatalf("Authenticate: %v", err)
	}
	// Sanity: the freshly minted session resolves.
	if _, err := svc.ResolveSession(ctx, authSession.RawToken); err != nil {
		t.Fatalf("ResolveSession before change: %v", err)
	}

	if err := svc.ChangePasswordForUser(ctx, authSession.CurrentUser, password, "NewPassword456!"); err != nil {
		t.Fatalf("ChangePasswordForUser: %v", err)
	}

	// The pre-change session must no longer resolve (revoked).
	if _, err := svc.ResolveSession(ctx, authSession.RawToken); err == nil {
		t.Fatalf("expected session to be revoked after password change, but ResolveSession succeeded")
	}
}
