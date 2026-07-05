package postgres

// tenant_key_repository.go — TenantKeyRepository (M7 F7.3 Task B) is the
// Postgres adapter backing application.TenantCryptoService's
// tenantKeyRepository port. Every statement carries an explicit tenant_id
// predicate (M6 F6.4 idiom) — reads run off the shared pool with no tx/GUC
// seeded, so correctness never depends on the RLS session GUC being set.
//
// Runtime DB driver is pgx (database/sql via pgx stdlib); unique-violation
// detection checks *pgconn.PgError first, falling back to *pq.Error for any
// lib/pq-driven test harness path (dual check idiom, mirrors
// iam/infrastructure/postgres/tenant_repository.go isTenantUniqueViolation).

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/lib/pq"
)

// TenantKeyRepository implements the tenantKeyRepository port the security
// module's TenantCryptoService depends on.
type TenantKeyRepository struct {
	db *sql.DB
}

// NewTenantKeyRepository constructs the adapter over the shared pool.
func NewTenantKeyRepository(db *sql.DB) *TenantKeyRepository {
	return &TenantKeyRepository{db: db}
}

// InsertIfAbsentTx inserts the tenant_keys row for tenantID with wrappedDEK,
// inside tx, unless a row already exists (ON CONFLICT DO NOTHING —
// idempotent, never rotates an existing key).
func (r *TenantKeyRepository) InsertIfAbsentTx(ctx context.Context, tx *sql.Tx, tenantID string, wrappedDEK []byte) error {
	const q = `
INSERT INTO metaldocs.tenant_keys (tenant_id, wrapped_dek)
VALUES ($1::uuid, $2)
ON CONFLICT (tenant_id) DO NOTHING
`
	if _, err := tx.ExecContext(ctx, q, tenantID, wrappedDEK); err != nil {
		if isTenantKeyUniqueViolation(err) {
			// Race with a concurrent provisioner for the same tenant — the
			// conflicting insert already satisfied the idempotency
			// contract, so treat it as success rather than surfacing a
			// spurious error.
			return nil
		}
		return fmt.Errorf("insert tenant key: %w", err)
	}
	return nil
}

// WrappedDEK returns tenantID's wrapped DEK and destroyed state. Explicit
// tenant_id = $1 predicate — this is a pool read with no tx/GUC seeded, so
// it must not depend on the RLS session GUC for tenant isolation.
func (r *TenantKeyRepository) WrappedDEK(ctx context.Context, tenantID string) ([]byte, bool, error) {
	const q = `
SELECT wrapped_dek, destroyed_at IS NOT NULL
FROM metaldocs.tenant_keys
WHERE tenant_id = $1::uuid
`
	var wrapped []byte
	var destroyed bool
	err := r.db.QueryRowContext(ctx, q, tenantID).Scan(&wrapped, &destroyed)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, fmt.Errorf("select tenant key: %w", err)
	}
	if destroyed {
		return nil, true, nil
	}
	return wrapped, false, nil
}

// DestroyTx crypto-shreds tenantID's key inside tx: sets destroyed_at and
// zeroes wrapped_dek. Explicit tenant_id predicate; no-op (0 rows affected)
// if the tenant has no key row or the key is already destroyed — both are
// success, not an error (idempotent erasure retries).
func (r *TenantKeyRepository) DestroyTx(ctx context.Context, tx *sql.Tx, tenantID string) error {
	const q = `
UPDATE metaldocs.tenant_keys
SET destroyed_at = now(), wrapped_dek = ''::bytea
WHERE tenant_id = $1::uuid
  AND destroyed_at IS NULL
`
	if _, err := tx.ExecContext(ctx, q, tenantID); err != nil {
		return fmt.Errorf("destroy tenant key: %w", err)
	}
	return nil
}

// isTenantKeyUniqueViolation reports whether err is a Postgres unique-key
// violation (23505), checking both the pgx and lib/pq error shapes.
func isTenantKeyUniqueViolation(err error) bool {
	var pgxErr *pgconn.PgError
	if errors.As(err, &pgxErr) && pgxErr.Code == "23505" {
		return true
	}
	var pqErr *pq.Error
	return errors.As(err, &pqErr) && string(pqErr.Code) == "23505"
}
