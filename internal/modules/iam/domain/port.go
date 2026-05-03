package domain

import "context"

// RoleProvider resolves effective roles for a given user identity.
type RoleProvider interface {
	RolesByUserID(ctx context.Context, userID string) ([]Role, error)
}

// RoleAdminRepository writes IAM user and role assignments.
type RoleAdminRepository interface {
	HasAnyRole(ctx context.Context, role Role) (bool, error)
	UpsertUserAndAssignRole(ctx context.Context, userID, displayName string, role Role, assignedBy string) error
	ReplaceUserRoles(ctx context.Context, userID, displayName string, roles []Role, assignedBy string) error
}
