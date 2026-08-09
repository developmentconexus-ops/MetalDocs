// Package infrastructure is the tokens Postgres adapter. It operates on the
// *sql.Tx the application service supplies (the service owns the tx boundary and
// has already seeded the authz GUC + checked the capability). The repo touches
// ONLY token_dictionary_entries.
package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"metaldocs/internal/modules/tokens/domain"
	"metaldocs/internal/platform/db"
)

// PostgresRepository is the Postgres-backed implementation of
// domain.Repository for the tokens module.
type PostgresRepository struct{}

// NewPostgresRepository constructs a PostgresRepository.
func NewPostgresRepository() *PostgresRepository { return &PostgresRepository{} }

var _ domain.Repository = (*PostgresRepository)(nil)

const selectColumns = `id, tenant_id, name, value, label, description, created_by, updated_by, created_at, updated_at`

func scanEntry(row interface {
	Scan(dest ...any) error
}) (*domain.Entry, error) {
	var e domain.Entry
	var description sql.NullString
	if err := row.Scan(
		&e.ID, &e.TenantID, &e.Name, &e.Value, &e.Label, &description,
		&e.CreatedBy, &e.UpdatedBy, &e.CreatedAt, &e.UpdatedAt,
	); err != nil {
		return nil, err
	}
	if description.Valid {
		e.Description = &description.String
	}
	return &e, nil
}

// Create inserts a new token dictionary entry and returns the created entry.
func (r *PostgresRepository) Create(ctx context.Context, tx db.Tx, e *domain.Entry) (*domain.Entry, error) {
	const q = `
INSERT INTO metaldocs.token_dictionary_entries
    (tenant_id, name, value, label, description, created_by, updated_by)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING ` + selectColumns
	row := tx.QueryRowContext(ctx, q,
		e.TenantID, e.Name, e.Value, e.Label, descArg(e.Description), e.CreatedBy, e.UpdatedBy)
	out, err := scanEntry(row)
	if err != nil {
		return nil, fmt.Errorf("tokens: create entry: %w", err)
	}
	return out, nil
}

// Update modifies an existing token dictionary entry and returns the updated entry.
func (r *PostgresRepository) Update(ctx context.Context, tx db.Tx, e *domain.Entry) (*domain.Entry, error) {
	const q = `
UPDATE metaldocs.token_dictionary_entries
   SET value = $1, label = $2, description = $3, updated_by = $4, updated_at = now()
 WHERE tenant_id = $5 AND id = $6
RETURNING ` + selectColumns
	row := tx.QueryRowContext(ctx, q,
		e.Value, e.Label, descArg(e.Description), e.UpdatedBy, e.TenantID, e.ID)
	out, err := scanEntry(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("tokens: update entry: %w", err)
	}
	return out, nil
}

// Delete removes a token dictionary entry by ID.
func (r *PostgresRepository) Delete(ctx context.Context, tx db.Tx, tenantID, id string) error {
	const q = `DELETE FROM metaldocs.token_dictionary_entries WHERE tenant_id = $1 AND id = $2`
	res, err := tx.ExecContext(ctx, q, tenantID, id)
	if err != nil {
		return fmt.Errorf("tokens: delete entry: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("tokens: delete entry rows: %w", err)
	}
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// GetByID retrieves a token dictionary entry by ID.
func (r *PostgresRepository) GetByID(ctx context.Context, tx db.Tx, tenantID, id string) (*domain.Entry, error) {
	const q = `SELECT ` + selectColumns + ` FROM metaldocs.token_dictionary_entries WHERE tenant_id = $1 AND id = $2`
	out, err := scanEntry(tx.QueryRowContext(ctx, q, tenantID, id))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("tokens: get entry by id: %w", err)
	}
	return out, nil
}

// GetByName retrieves a token dictionary entry by name.
func (r *PostgresRepository) GetByName(ctx context.Context, tx db.Tx, tenantID, name string) (*domain.Entry, error) {
	const q = `SELECT ` + selectColumns + ` FROM metaldocs.token_dictionary_entries WHERE tenant_id = $1 AND name = $2`
	out, err := scanEntry(tx.QueryRowContext(ctx, q, tenantID, name))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("tokens: get entry by name: %w", err)
	}
	return out, nil
}

// List retrieves all token dictionary entries for a tenant, ordered by name.
func (r *PostgresRepository) List(ctx context.Context, tx db.Tx, tenantID string) ([]domain.Entry, error) {
	const q = `SELECT ` + selectColumns + ` FROM metaldocs.token_dictionary_entries WHERE tenant_id = $1 ORDER BY name ASC`
	rows, err := tx.QueryContext(ctx, q, tenantID)
	if err != nil {
		return nil, fmt.Errorf("tokens: list entries: %w", err)
	}
	defer func() { _ = rows.Close() }()
	var out []domain.Entry
	for rows.Next() {
		e, err := scanEntry(rows)
		if err != nil {
			return nil, fmt.Errorf("tokens: scan entry: %w", err)
		}
		out = append(out, *e)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("tokens: iterate entries: %w", err)
	}
	return out, nil
}

func descArg(d *string) any {
	if d == nil {
		return nil
	}
	return *d
}
