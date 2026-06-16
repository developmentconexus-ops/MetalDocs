package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/lib/pq"

	iamdomain "metaldocs/internal/modules/iam/domain"
)

var _ iamdomain.AdminRoleMemberReader = (*AdminRoleMemberRepository)(nil)

// AdminRoleMemberRepository implements iamdomain.AdminRoleMemberReader. It reads
// role membership from metaldocs.iam_user_roles on the connection pool (never on
// a caller's transaction), so cross-module consumers never hold an iam_user_roles
// read inside a lock-holding transaction (H-PRE-1, M4/F4.2).
type AdminRoleMemberRepository struct {
	db *sql.DB
}

// NewAdminRoleMemberRepository constructs a pool-backed AdminRoleMemberReader.
func NewAdminRoleMemberRepository(db *sql.DB) *AdminRoleMemberRepository {
	return &AdminRoleMemberRepository{db: db}
}

// AdminRoleMembers returns a map of userID → roleCode for each user in tenantID
// whose role_code is in roleCodes. Where a user holds multiple matching roles,
// the alphabetically first (MIN) is returned — mirrors the JOIN semantics it
// replaces. Returns an empty map when no matching members exist.
func (r *AdminRoleMemberRepository) AdminRoleMembers(ctx context.Context, tenantID string, roleCodes []string) (map[string]string, error) {
	if len(roleCodes) == 0 {
		return map[string]string{}, nil
	}
	rows, err := r.db.QueryContext(ctx,
		`SELECT user_id, MIN(role_code) AS role_code
		   FROM metaldocs.iam_user_roles
		  WHERE tenant_id = $1::uuid
		    AND role_code = ANY($2)
		  GROUP BY user_id`,
		tenantID, pq.Array(roleCodes),
	)
	if err != nil {
		return nil, fmt.Errorf("iam: AdminRoleMembers(%s): %w", tenantID, err)
	}
	defer rows.Close()

	out := map[string]string{}
	for rows.Next() {
		var userID, role string
		if err := rows.Scan(&userID, &role); err != nil {
			return nil, fmt.Errorf("iam: AdminRoleMembers scan: %w", err)
		}
		out[userID] = role
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iam: AdminRoleMembers rows: %w", err)
	}
	return out, nil
}
