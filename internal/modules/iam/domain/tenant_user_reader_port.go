package domain

import "context"

// TenantUserReader is the narrow read surface cross-module consumers use to learn
// which users belong to a tenant, without reaching across the module boundary into
// metaldocs.iam_users. IAM owns iam_users; a (user_id, tenant_id) row IS the
// tenant membership.
//
// It exists for consumers that must tenant-scope a table which has no tenant_id
// column of its own — notably metaldocs.auth_identities (a global-PK table, ADR
// 0027). Such a consumer fetches the tenant's member ids here and filters its own
// table with WHERE ... = ANY(ids), instead of JOINing iam_users (H-G class
// remediation, M4/F4.5; see ADR 0031).
//
// The implementation reads from the connection pool — never from a caller's
// lock-holding transaction (H-PRE-1).
type TenantUserReader interface {
	// TenantUserIDs returns every metaldocs.iam_users.user_id for tenantID.
	//
	// All members are returned regardless of deactivated_at — this reproduces the
	// INNER JOIN membership semantics of the consumers it replaces (which do NOT
	// filter on deactivated_at). Returns an empty (non-nil) slice when the tenant
	// has no members or does not exist.
	TenantUserIDs(ctx context.Context, tenantID string) ([]string, error)
}

// NoopTenantUserReader is the explicit null-object for callers that do not resolve
// tenant membership (tests, or paths where it is irrelevant). It satisfies
// TenantUserReader and always returns an empty slice with a nil error. Mirrors
// NoopUserDisplayNameReader.
type NoopTenantUserReader struct{}

// TenantUserIDs always returns an empty, non-nil slice and a nil error.
func (NoopTenantUserReader) TenantUserIDs(_ context.Context, _ string) ([]string, error) {
	return []string{}, nil
}
