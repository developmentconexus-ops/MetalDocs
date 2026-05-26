package domain

import "context"

// RoleProvider resolves effective roles for a given user identity within a tenant.
type RoleProvider interface {
	RolesByUserID(ctx context.Context, userID, tenantID string) ([]Role, error)
}

// RoleAdminRepository writes IAM user and role assignments.
type RoleAdminRepository interface {
	// Bootstrap operations.
	HasAnyRole(ctx context.Context, role Role, tenantID string) (bool, error)
	// Lifecycle operations.
	UpsertUserAndAssignRole(ctx context.Context, userID, displayName, tenantID string, role Role, assignedBy string) error
	ReplaceUserRoles(ctx context.Context, userID, displayName, tenantID string, roles []Role, assignedBy string) error
}
