package domain

import "context"

// AdminRoleMemberReader is the narrow read surface for resolving which users
// hold admin roles in a tenant. Cross-module consumers (security) use this
// instead of JOINing metaldocs.iam_user_roles directly (H-G class, M4/F4.2).
// The implementation reads from the connection pool — never from a caller's
// lock-holding transaction (H-PRE-1).
type AdminRoleMemberReader interface {
	// AdminRoleMembers returns a map of userID → roleCode for each user in
	// tenantID whose role_code is in roleCodes. Where a user holds multiple
	// matching roles, the alphabetically first is returned (mirrors MIN()).
	// Returns an empty (non-nil) map when no matching members exist.
	AdminRoleMembers(ctx context.Context, tenantID string, roleCodes []string) (map[string]string, error)
}

// NoopAdminRoleMemberReader satisfies AdminRoleMemberReader and always returns
// an empty map with a nil error.
type NoopAdminRoleMemberReader struct{}

func (NoopAdminRoleMemberReader) AdminRoleMembers(_ context.Context, _ string, _ []string) (map[string]string, error) {
	return map[string]string{}, nil
}
