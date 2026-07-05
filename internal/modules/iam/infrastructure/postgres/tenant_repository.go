package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/lib/pq"

	iamdomain "metaldocs/internal/modules/iam/domain"
)

// TenantRepository writes the tenant root row (metaldocs.tenants) for M7
// F7.2 onboarding. Tx-scoped only: the tripwire arm (migration 0277) requires
// tenant.onboard to already be asserted on tx before the INSERT runs, so
// there is no non-tx convenience wrapper — callers always compose this inside
// OnboardTenantService's single business tx.
type TenantRepository struct {
	db *sql.DB
}

// NewTenantRepository constructs a pool-backed TenantRepository. db is kept
// only so the type matches the constructor convention of its siblings in this
// package (e.g. NewRoleAdminRepository); InsertTenantTx never uses it
// directly — it always runs against the caller-supplied *sql.Tx.
func NewTenantRepository(db *sql.DB) *TenantRepository {
	return &TenantRepository{db: db}
}

// InsertTenantTx inserts the tenant root row inside the caller's tx. Returns
// iamdomain.ErrTenantSlugConflict on a tenants_slug_key unique violation.
func (r *TenantRepository) InsertTenantTx(ctx context.Context, tx *sql.Tx, tenantID, name, slug string) error {
	if _, err := tx.ExecContext(ctx, `
INSERT INTO metaldocs.tenants (id, name, slug)
VALUES ($1::uuid, $2, $3)
`, tenantID, name, slug); err != nil {
		if isTenantUniqueViolation(err) {
			return iamdomain.ErrTenantSlugConflict
		}
		return fmt.Errorf("insert tenant: %w", err)
	}
	return nil
}

func isTenantUniqueViolation(err error) bool {
	var pgxErr *pgconn.PgError
	if errors.As(err, &pgxErr) && pgxErr.Code == "23505" {
		return true
	}
	var pqErr *pq.Error
	return errors.As(err, &pqErr) && string(pqErr.Code) == "23505"
}
