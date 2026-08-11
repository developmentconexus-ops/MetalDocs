package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	iamdomain "metaldocs/internal/modules/iam/domain"
)

// Ensure TenantErasureRepository satisfies iamdomain.TenantErasureReader.
var _ iamdomain.TenantErasureReader = (*TenantErasureRepository)(nil)

// TenantErasureRepository implements iamdomain.TenantErasureReader. It reads
// metaldocs.tenants.erased_at on the connection pool (never on a caller's
// transaction), matching TenantUserRepository / UserDisplayNameRepository
// (H-PRE-1). metaldocs.tenants carries no RLS policy (it is the tenancy
// root, not a per-tenant table), so this is a plain pool read with no tx-GUC
// setup required.
type TenantErasureRepository struct {
	db *sql.DB
}

// NewTenantErasureRepository constructs a pool-backed TenantErasureReader.
func NewTenantErasureRepository(db *sql.DB) *TenantErasureRepository {
	return &TenantErasureRepository{db: db}
}

// IsErased reports whether tenantID has been erased. A tenantID that
// resolves to no metaldocs.tenants row is returned as an error rather than
// (false, nil) — see the interface doc for why this read-gate deliberately
// does not share TenantLookupTx's lenient not-found convention.
func (r *TenantErasureRepository) IsErased(ctx context.Context, tenantID string) (bool, error) {
	var erasedAt sql.NullTime
	err := r.db.QueryRowContext(ctx,
		`SELECT erased_at FROM metaldocs.tenants WHERE id = $1::uuid`,
		tenantID,
	).Scan(&erasedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return false, fmt.Errorf("iam: IsErased(%s): tenant not found", tenantID)
	}
	if err != nil {
		return false, fmt.Errorf("iam: IsErased(%s): %w", tenantID, err)
	}
	return erasedAt.Valid, nil
}
