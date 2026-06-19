package domain

import "context"

// UserTenantReader is the narrow read surface cross-module consumers use to learn
// which tenants a user belongs to, without reaching across the module boundary
// into metaldocs.iam_user_roles. IAM owns iam_user_roles; a (user_id, tenant_id)
// row IS the user's membership in that tenant.
//
// It is the inverse of TenantUserReader (which answers tenant→users); this answers
// user→tenants. It exists for consumers that must resolve a user's tenant set
// without JOINing or directly reading iam_user_roles — notably auth session
// establishment, which picks the active tenant from the user's memberships
// (H-G class remediation, M5/F5.2; mirrors M4/F4.5, ADR 0031).
//
// The implementation reads from the connection pool — never from a caller's
// lock-holding transaction (H-PRE-1).
type UserTenantReader interface {
	// UserTenantIDs returns the distinct tenant IDs the user holds any role in,
	// sorted ascending. Returns an empty (non-nil) slice when the user has no
	// roles or does not exist.
	UserTenantIDs(ctx context.Context, userID string) ([]string, error)
}

// NoopUserTenantReader is the explicit null-object for callers that do not resolve
// a user's tenant set (tests, or paths where it is irrelevant). It satisfies
// UserTenantReader and always returns an empty slice with a nil error. Mirrors
// NoopTenantUserReader.
type NoopUserTenantReader struct{}

func (NoopUserTenantReader) UserTenantIDs(_ context.Context, _ string) ([]string, error) {
	return []string{}, nil
}
