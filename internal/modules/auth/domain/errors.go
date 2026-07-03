package domain

import "errors"

var (
	// ErrInvalidCredentials is returned for a wrong password or an unknown identity
	// (same error in both cases — disclosure-safe, never reveals which).
	ErrInvalidCredentials = errors.New("auth invalid credentials")
	// ErrSessionNotFound is returned when the session token does not match any stored session.
	ErrSessionNotFound = errors.New("auth session not found")
	// ErrSessionExpired is returned when the session's ExpiresAt has passed.
	ErrSessionExpired = errors.New("auth session expired")
	// ErrSessionRevoked is returned when the session's RevokedAt is set.
	ErrSessionRevoked = errors.New("auth session revoked")
	// ErrPasswordPolicy is returned when a new password fails the configured policy checks.
	ErrPasswordPolicy = errors.New("auth password policy violation")
	// ErrPasswordChangeRequired is returned when the identity's MustChangePassword is set and
	// the caller attempted an operation that requires a fresh password first.
	ErrPasswordChangeRequired = errors.New("auth password change required")
	// ErrIdentityLocked is returned when the identity's LockedUntil is in the future.
	ErrIdentityLocked = errors.New("auth identity locked")
	// ErrIdentityInactive is returned when the identity's IsActive is false.
	ErrIdentityInactive = errors.New("auth identity inactive")
	// ErrIdentityNotFound is returned when no identity matches the given lookup key.
	ErrIdentityNotFound = errors.New("auth identity not found")
	// ErrUserAlreadyExists is returned when creating a user whose username or email is already taken.
	ErrUserAlreadyExists = errors.New("auth user already exists")
	// ErrTenantNotPermitted is returned when the login X-Tenant-ID claim names a
	// tenant the authenticated user has no role in.
	ErrTenantNotPermitted = errors.New("auth tenant not permitted")
	// ErrTenantClaimRequired is returned when the user belongs to multiple tenants
	// and no X-Tenant-ID claim was provided at login.
	ErrTenantClaimRequired = errors.New("auth tenant claim required")
	// ErrTenantNotFound is returned when the resolved tenant ID does not match any stored tenant.
	ErrTenantNotFound = errors.New("auth tenant not found")
)
