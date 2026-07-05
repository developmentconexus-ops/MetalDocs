package domain

import "errors"

var (
	// ErrUserNotFound is returned when no iam_users row exists for the requested
	// (tenant, user) pair.
	ErrUserNotFound = errors.New("iam user not found")
	// ErrUserInactive is returned when the resolved user exists but is
	// deactivated — callers must treat this the same as "not found" for
	// authorization purposes.
	ErrUserInactive = errors.New("iam user inactive")
	// ErrNoRolesAssigned is returned when a user has no role assignments in the
	// tenant (RoleProvider.RolesByUserID and related lookups).
	ErrNoRolesAssigned = errors.New("iam no roles assigned")
	// ErrInvalidRole is returned by ParseRole when the input does not match one
	// of the eight canonical roles.
	ErrInvalidRole = errors.New("iam invalid role")
	// ErrTenantOnboardValidation is returned by OnboardTenantService for
	// blank/malformed input (name, slug, admin_user_id, admin_display_name,
	// admin_password) before any write is attempted. Maps to 400.
	ErrTenantOnboardValidation = errors.New("iam: tenant onboarding validation failed")
	// ErrTenantSlugConflict is returned when the requested slug already exists
	// on metaldocs.tenants (unique violation on tenants_slug_key). Maps to 409.
	ErrTenantSlugConflict = errors.New("iam: tenant slug already exists")
	// ErrTenantAdminUserConflict is returned when the requested admin_user_id
	// already exists (auth_identities unique violation). Maps to 409.
	ErrTenantAdminUserConflict = errors.New("iam: tenant admin user already exists")
	// ErrTenantNotFound is returned by TenantLifecycleService.RequestExport /
	// RequestErase when the target tenantId does not exist in
	// metaldocs.tenants. Maps to 404 (M7 F7.3).
	ErrTenantNotFound = errors.New("iam: tenant not found")
	// ErrTenantAlreadyErased is returned when the target tenant's erased_at is
	// already set — export of an erased tenant and any second erase request
	// both hit this. Maps to 409 (M7 F7.3).
	ErrTenantAlreadyErased = errors.New("iam: tenant already erased")
)
