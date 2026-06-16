package postgres

import (
	"context"
	"database/sql"
	"fmt"

	iamdomain "metaldocs/internal/modules/iam/domain"
)

var _ iamdomain.MfaUserReader = (*MfaUserRepository)(nil)

// MfaUserRepository implements iamdomain.MfaUserReader. It reads MFA coverage
// counts from metaldocs.iam_users / iam_user_roles on the connection pool (never
// on a caller's transaction), so cross-module consumers never hold iam table reads
// inside a lock-holding transaction (H-PRE-1, M4/F4.3).
type MfaUserRepository struct {
	db *sql.DB
}

// NewMfaUserRepository constructs a pool-backed MfaUserReader.
func NewMfaUserRepository(db *sql.DB) *MfaUserRepository {
	return &MfaUserRepository{db: db}
}

// TenantMfaCounts returns total active user count and MFA-enabled count for tenantID.
func (r *MfaUserRepository) TenantMfaCounts(ctx context.Context, tenantID string) (total, mfaEnabled int, err error) {
	err = r.db.QueryRowContext(ctx, `
SELECT COUNT(*) FILTER (WHERE deactivated_at IS NULL)                  AS total,
       COUNT(*) FILTER (WHERE deactivated_at IS NULL AND mfa_enabled)  AS enabled
  FROM metaldocs.iam_users
 WHERE tenant_id = $1::uuid`, tenantID).Scan(&total, &mfaEnabled)
	if err != nil {
		return 0, 0, fmt.Errorf("iam: TenantMfaCounts(%s): %w", tenantID, err)
	}
	return total, mfaEnabled, nil
}

// TenantMfaCountsByRole returns per-role (total, MFA-enabled) counts for active
// users in tenantID, ordered by role_code.
func (r *MfaUserRepository) TenantMfaCountsByRole(ctx context.Context, tenantID string) ([]iamdomain.RoleMfaCounts, error) {
	rows, err := r.db.QueryContext(ctx, `
SELECT ur.role_code,
       COUNT(*)                                      AS total,
       COUNT(*) FILTER (WHERE u.mfa_enabled)         AS enabled
  FROM metaldocs.iam_user_roles ur
  JOIN metaldocs.iam_users u
    ON u.user_id   = ur.user_id
   AND u.tenant_id = ur.tenant_id
 WHERE ur.tenant_id = $1::uuid
   AND u.deactivated_at IS NULL
 GROUP BY ur.role_code
 ORDER BY ur.role_code`, tenantID)
	if err != nil {
		return nil, fmt.Errorf("iam: TenantMfaCountsByRole(%s): %w", tenantID, err)
	}
	defer rows.Close()

	var out []iamdomain.RoleMfaCounts
	for rows.Next() {
		var s iamdomain.RoleMfaCounts
		if err := rows.Scan(&s.Role, &s.Total, &s.MfaEnabled); err != nil {
			return nil, fmt.Errorf("iam: TenantMfaCountsByRole scan: %w", err)
		}
		out = append(out, s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iam: TenantMfaCountsByRole rows: %w", err)
	}
	return out, nil
}
