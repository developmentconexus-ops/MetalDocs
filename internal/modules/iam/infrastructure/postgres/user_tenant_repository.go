package postgres

import (
	"context"
	"database/sql"
	"fmt"

	iamdomain "metaldocs/internal/modules/iam/domain"
)

// Ensure UserTenantRepository satisfies iamdomain.UserTenantReader.
var _ iamdomain.UserTenantReader = (*UserTenantRepository)(nil)

// userTenantsQuery returns the distinct tenant IDs a user holds any role in,
// sorted ascending. It is the canonical source of the membership read previously
// embedded directly in auth/infrastructure/postgres (H-G site #1, M5/F5.2) — kept
// character-identical on DISTINCT / WHERE / ORDER BY so the move is behavior-
// preserving parity.
const userTenantsQuery = `
SELECT DISTINCT tenant_id::text
FROM metaldocs.iam_user_roles
WHERE user_id = $1
ORDER BY tenant_id
`

// UserTenantRepository implements iamdomain.UserTenantReader. It reads the user's
// tenant memberships from metaldocs.iam_user_roles on the connection pool (never on
// a caller's transaction), so cross-module consumers never hold an iam_user_roles
// read inside a lock-holding transaction (H-PRE-1, M4/F4.5; see ADR 0031).
type UserTenantRepository struct {
	db *sql.DB
}

// NewUserTenantRepository constructs a pool-backed UserTenantReader.
func NewUserTenantRepository(db *sql.DB) *UserTenantRepository {
	return &UserTenantRepository{db: db}
}

// UserTenantIDs returns the distinct, ascending-sorted tenant IDs the user holds
// any role in. Returns an empty (non-nil) slice when the user has no roles.
func (r *UserTenantRepository) UserTenantIDs(ctx context.Context, userID string) ([]string, error) {
	rows, err := r.db.QueryContext(ctx, userTenantsQuery, userID)
	if err != nil {
		return nil, fmt.Errorf("iam: UserTenantIDs(%s): %w", userID, err)
	}
	defer func() { _ = rows.Close() }()

	out := []string{}
	for rows.Next() {
		var tid string
		if err := rows.Scan(&tid); err != nil {
			return nil, fmt.Errorf("iam: UserTenantIDs scan: %w", err)
		}
		out = append(out, tid)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iam: UserTenantIDs rows: %w", err)
	}
	return out, nil
}
