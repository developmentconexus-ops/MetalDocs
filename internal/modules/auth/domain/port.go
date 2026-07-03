// Package domain holds the auth bounded-context's pure domain model:
// identities, sessions, tenants, and the Repository port they are persisted
// through. Password material (PlainPassword, PasswordHash) and session tokens
// are handled here with redaction/hashing contracts that callers must respect;
// this package has no HTTP or SQL dependencies.
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

// Repository is the persistence port for the auth module: identities,
// sessions, login-lock serialization, and user administration.
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
