package domain

import (
	"context"
	"time"

	iamdomain "metaldocs/internal/modules/iam/domain"
)

// CapabilityProvider resolves the effective capability codes a user holds in a
// tenant. Optional dependency for the auth Service: when wired, CurrentUser
// responses carry a capabilities[] hint for the frontend defense-in-depth gate.
// Backend remains the sole enforcer (wiki/concepts/authz-tiers.md).
type CapabilityProvider interface {
	CapsByUserID(ctx context.Context, userID, tenantID string) ([]iamdomain.Capability, error)
}

// LoginState is the authoritative credential state read inside the per-identity
// login lock. It carries only what Authenticate needs to make the lockout
// decision without re-reading outside the serialized region.
type LoginState struct {
	PasswordHash PasswordHash
	IsActive     bool
	LockedUntil  *time.Time
}

// LoginTx is the serialized critical-section view handed to WithinLoginLock's
// callback. Every call runs while the per-identity lock is held, so the lockout
// check and the failed-attempt write cannot interleave with a concurrent login.
type LoginTx interface {
	LoadLoginState(ctx context.Context, userID string) (LoginState, error)
	RecordFailedLogin(ctx context.Context, userID string, maxAttempts int, lockDurationSeconds int, ip string) (attempts int, lockedUntil *time.Time, err error)
}

type Repository interface {
	FindIdentityByIdentifier(ctx context.Context, identifier string) (Identity, error)
	FindIdentityByUserID(ctx context.Context, userID string) (Identity, error)
	// WithinLoginLock runs fn while holding an exclusive lock scoped to userID,
	// serializing concurrent login attempts on the same identity. This closes the
	// TOCTOU window between the lockout check and the failed-attempt write: a burst
	// of parallel attempts is processed one at a time, so once the threshold lock
	// is set the remaining in-flight attempts observe it and are rejected before
	// the password is verified. Postgres uses pg_advisory_xact_lock inside a
	// transaction (the LoginTx runs every read/write on that transaction); the
	// in-memory double uses a per-identity mutex. fn's returned error aborts the
	// unit (Postgres rolls back).
	WithinLoginLock(ctx context.Context, userID string, fn func(LoginTx) error) error
	CreateSession(ctx context.Context, session Session) error
	FindSession(ctx context.Context, sessionID string) (Session, error)
	TouchSession(ctx context.Context, sessionID string, seenAt time.Time) error
	RevokeSession(ctx context.Context, sessionID string, revokedAt time.Time) error
	RevokeSessionsByUserID(ctx context.Context, userID string, revokedAt time.Time) error
	RecordSuccessfulLogin(ctx context.Context, userID string, loginAt time.Time) error
	// RecordLastLoginContext records the client IP, User-Agent, and (optionally)
	// derived device label that produced the most recent successful login. Stored
	// on metaldocs.iam_users (governance metadata, tenant-scoped) — separate from
	// auth_identities.last_login_at which lives on the credential row. PR-4 leaves
	// deviceLabel empty; PR-7 populates it via UA parsing.
	RecordLastLoginContext(ctx context.Context, userID, tenantID, ip, userAgent, deviceLabel string) error
	// RecordFailedLogin increments failed_login_attempts and, on threshold,
	// sets locked_until. ip is the request remote address — implementations
	// may stamp it onto last_failed_login_ip alongside last_failed_login_at
	// so the Sessions & Security tab can surface the most recent failure.
	RecordFailedLogin(ctx context.Context, userID string, maxAttempts int, lockDurationSeconds int, ip string) (attempts int, lockedUntil *time.Time, err error)
	CreateUser(ctx context.Context, params CreateUserParams) error
	ListUsers(ctx context.Context) ([]ManagedUser, error)
	UpdateUser(ctx context.Context, params UpdateUserParams) error
	ListOnlineUsers(ctx context.Context, tenantID string, activeSince time.Time) ([]OnlineUser, error)
	BootstrapAdmin(ctx context.Context, params BootstrapAdminParams) (bool, error)
	// GetUserTenants returns the distinct tenant IDs from iam_user_roles for the
	// given user. Used by Login to verify the X-Tenant-ID claim and bind it to
	// the session row.
	GetUserTenants(ctx context.Context, userID string) ([]string, error)
	GetTenantByID(ctx context.Context, tenantID string) (Tenant, error)
}
