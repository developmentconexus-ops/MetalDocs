package domain

import (
	"context"
)

// RoleProvider resolves effective roles for a given user identity within a tenant.
type RoleProvider interface {
	// RolesByUserID returns the effective roles held by userID in tenantID.
	RolesByUserID(ctx context.Context, userID, tenantID string) ([]Role, error)
	// RolesByUserIDs resolves roles for multiple users in a single query.
	// Returns a map of userID → []Role. Missing or inactive users are absent
	// from the map (same semantics as ErrUserNotFound from RolesByUserID).
	// Users that are active but have no roles are present with an empty slice.
	RolesByUserIDs(ctx context.Context, tenantID string, userIDs []string) (map[string][]Role, error)
	// UserActiveInTenant returns true iff the user exists in the tenant and is
	// not deactivated. Implementations MUST return identical semantics to
	// "the user appears in ListUsers(tenantID) with deactivated_at IS NULL".
	// Never caches negatives (deactivation must take effect within one call).
	UserActiveInTenant(ctx context.Context, tenantID, userID string) (bool, error)
}

// RoleAdminRepository writes IAM user and role assignments.
type RoleAdminRepository interface {
	// Bootstrap operations.
	//
	// HasAnyRole reports whether any user in tenantID currently holds role.
	HasAnyRole(ctx context.Context, role Role, tenantID string) (bool, error)
	// Lifecycle operations — autocommit wrappers (begin own tx).
	//
	// UpsertUserAndAssignRole creates or updates the user's display name and
	// assigns role in tenantID, in its own transaction.
	UpsertUserAndAssignRole(ctx context.Context, userID, displayName, tenantID string, role Role, assignedBy string) error
	// ReplaceUserRoles replaces the user's role assignment in tenantID with
	// role, in its own transaction.
	ReplaceUserRoles(ctx context.Context, userID, displayName, tenantID string, role Role, assignedBy string) error
}
