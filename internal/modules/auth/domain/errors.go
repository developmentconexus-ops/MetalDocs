package domain

import "errors"

var (
	ErrInvalidCredentials     = errors.New("auth invalid credentials")
	ErrSessionNotFound        = errors.New("auth session not found")
	ErrSessionExpired         = errors.New("auth session expired")
	ErrSessionRevoked         = errors.New("auth session revoked")
	ErrPasswordPolicy         = errors.New("auth password policy violation")
	ErrPasswordChangeRequired = errors.New("auth password change required")
	ErrIdentityLocked         = errors.New("auth identity locked")
	ErrIdentityInactive       = errors.New("auth identity inactive")
	ErrIdentityNotFound       = errors.New("auth identity not found")
	ErrUserAlreadyExists      = errors.New("auth user already exists")
	// ErrTenantNotPermitted is returned when the login X-Tenant-ID claim names a
	// tenant the authenticated user has no role in.
	ErrTenantNotPermitted = errors.New("auth tenant not permitted")
	// ErrTenantClaimRequired is returned when the user belongs to multiple tenants
	// and no X-Tenant-ID claim was provided at login.
	ErrTenantClaimRequired = errors.New("auth tenant claim required")
)
