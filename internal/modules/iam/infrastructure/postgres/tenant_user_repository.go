package postgres

import (
	"context"
	"database/sql"
	"fmt"

	iamdomain "metaldocs/internal/modules/iam/domain"
)

// Ensure TenantUserRepository satisfies iamdomain.TenantUserReader.
var _ iamdomain.TenantUserReader = (*TenantUserRepository)(nil)

// TenantUserRepository implements iamdomain.TenantUserReader. It reads the tenant's
// member user_ids from metaldocs.iam_users on the connection pool (never on a
// caller's transaction), so cross-module consumers never hold an iam_users read
// inside a lock-holding transaction (H-PRE-1, M4/F4.5; see ADR 0031).
type TenantUserRepository struct {
	db *sql.DB
}

// NewTenantUserRepository constructs a pool-backed TenantUserReader.
func NewTenantUserRepository(db *sql.DB) *TenantUserRepository {
	return &TenantUserRepository{db: db}
}

// TenantUserIDs returns every metaldocs.iam_users.user_id for tenantID. All members
// are returned regardless of deactivated_at (matching the INNER JOIN membership of
// the consumers this replaces). Returns an empty slice when the tenant has no members.
func (r *TenantUserRepository) TenantUserIDs(ctx context.Context, tenantID string) ([]string, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT user_id
		   FROM metaldocs.iam_users
		  WHERE tenant_id = $1::uuid`,
		tenantID,
	)
	if err != nil {
		return nil, fmt.Errorf("iam: TenantUserIDs(%s): %w", tenantID, err)
	}
	defer func() { _ = rows.Close() }()

	out := []string{}
	for rows.Next() {
		var uid string
		if err := rows.Scan(&uid); err != nil {
			return nil, fmt.Errorf("iam: TenantUserIDs scan: %w", err)
		}
		out = append(out, uid)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iam: TenantUserIDs rows: %w", err)
	}
	return out, nil
}
