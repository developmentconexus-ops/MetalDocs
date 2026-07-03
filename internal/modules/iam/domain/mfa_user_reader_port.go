package domain

import "context"

// RoleMfaCounts holds raw MFA counts for one role (percentages are the consumer's job).
type RoleMfaCounts struct {
	Role       string
	Total      int
	MfaEnabled int
}

// MfaUserReader is the narrow read surface for per-tenant MFA coverage counts.
// Cross-module consumers (security) use this instead of querying iam_users /
// iam_user_roles directly (H-G class, M4/F4.3).
// The implementation reads from the connection pool — never from a caller's
// lock-holding transaction (H-PRE-1).
type MfaUserReader interface {
	// TenantMfaCounts returns total active user count and MFA-enabled count for tenantID.
	TenantMfaCounts(ctx context.Context, tenantID string) (total, mfaEnabled int, err error)
	// TenantMfaCountsByRole returns per-role (total, MFA-enabled) counts for active
	// users in tenantID, ordered by role_code.
	TenantMfaCountsByRole(ctx context.Context, tenantID string) ([]RoleMfaCounts, error)
}

// NoopMfaUserReader satisfies MfaUserReader and always returns zeros / empty slice.
type NoopMfaUserReader struct{}

// TenantMfaCounts always returns (0, 0, nil).
func (NoopMfaUserReader) TenantMfaCounts(_ context.Context, _ string) (int, int, error) {
	return 0, 0, nil
}

// TenantMfaCountsByRole always returns an empty, non-nil slice and a nil error.
func (NoopMfaUserReader) TenantMfaCountsByRole(_ context.Context, _ string) ([]RoleMfaCounts, error) {
	return []RoleMfaCounts{}, nil
}
