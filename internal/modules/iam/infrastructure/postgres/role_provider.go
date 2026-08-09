package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/lib/pq"

	"metaldocs/internal/modules/iam/domain"
)

// RoleProvider resolves user role assignments from metaldocs.iam_users /
// iam_user_roles on the connection pool for tier-2-adjacent membership
// verification and role lookups.
type RoleProvider struct {
	db *sql.DB
}

// NewRoleProvider constructs a pool-backed RoleProvider.
func NewRoleProvider(db *sql.DB) *RoleProvider {
	return &RoleProvider{db: db}
}

// RolesByUserID resolves tenant roles for a user in a single round trip using a
// LEFT JOIN. The join produces:
//   - 0 rows          → user not found or inactive   → ErrUserNotFound
//   - 1 row, NULL code → active user, no roles        → ErrNoRolesAssigned
//   - N rows, non-NULL → active user with N roles     → (roles, nil)
//
// Semantics are identical to the previous two-query implementation.
func (p *RoleProvider) RolesByUserID(ctx context.Context, userID, tenantID string) ([]domain.Role, error) {
	const q = `
SELECT r.role_code
FROM metaldocs.iam_users u
LEFT JOIN metaldocs.iam_user_roles r
       ON r.user_id = u.user_id
      AND r.tenant_id = u.tenant_id
WHERE u.user_id = $1
  AND u.tenant_id = $2::uuid
  AND u.deactivated_at IS NULL
ORDER BY r.role_code ASC
`
	rows, err := p.db.QueryContext(ctx, q, userID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("query iam roles: %w", err)
	}
	defer func() { _ = rows.Close() }()

	found := false
	roles := make([]domain.Role, 0, 4)
	for rows.Next() {
		found = true
		var roleCode sql.NullString
		if err := rows.Scan(&roleCode); err != nil {
			return nil, fmt.Errorf("scan iam role: %w", err)
		}
		if roleCode.Valid {
			roles = append(roles, domain.Role(roleCode.String))
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate iam roles: %w", err)
	}

	if !found {
		return nil, domain.ErrUserNotFound
	}
	if len(roles) == 0 {
		return nil, domain.ErrNoRolesAssigned
	}
	return roles, nil
}

// RolesByUserIDs resolves roles for a set of users in a single round trip using
// = ANY($1). The returned map contains only active users; absent entries mean
// the user was not found or is inactive (mirrors ErrUserNotFound semantics from
// RolesByUserID). Active users with no roles are present with an empty slice.
func (p *RoleProvider) RolesByUserIDs(ctx context.Context, tenantID string, userIDs []string) (map[string][]domain.Role, error) {
	if len(userIDs) == 0 {
		return map[string][]domain.Role{}, nil
	}
	const q = `
SELECT u.user_id, r.role_code
FROM metaldocs.iam_users u
LEFT JOIN metaldocs.iam_user_roles r
       ON r.user_id = u.user_id
      AND r.tenant_id = u.tenant_id
WHERE u.user_id = ANY($1)
  AND u.tenant_id = $2::uuid
  AND u.deactivated_at IS NULL
ORDER BY u.user_id, r.role_code ASC
`
	rows, err := p.db.QueryContext(ctx, q, pq.Array(userIDs), tenantID)
	if err != nil {
		return nil, fmt.Errorf("batch query iam roles: %w", err)
	}
	defer func() { _ = rows.Close() }()

	out := make(map[string][]domain.Role, len(userIDs))
	for rows.Next() {
		var uid string
		var roleCode sql.NullString
		if err := rows.Scan(&uid, &roleCode); err != nil {
			return nil, fmt.Errorf("scan batch iam role: %w", err)
		}
		if _, seen := out[uid]; !seen {
			out[uid] = []domain.Role{}
		}
		if roleCode.Valid {
			out[uid] = append(out[uid], domain.Role(roleCode.String))
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate batch iam roles: %w", err)
	}
	return out, nil
}

// UserActiveInTenant reports whether userID is an active member of tenantID
// in the IAM store. Used by PeopleService.VerifyUserInTenant for a fast
// EXISTS point-lookup in place of loading the full user list.
func (p *RoleProvider) UserActiveInTenant(ctx context.Context, tenantID, userID string) (bool, error) {
	const q = `
SELECT EXISTS (
    SELECT 1
    FROM metaldocs.iam_users
    WHERE user_id = $1
      AND tenant_id = $2::uuid
      AND deactivated_at IS NULL
)
`
	var exists bool
	if err := p.db.QueryRowContext(ctx, q, userID, tenantID).Scan(&exists); err != nil {
		return false, fmt.Errorf("check iam user active in tenant: %w", err)
	}
	return exists, nil
}
