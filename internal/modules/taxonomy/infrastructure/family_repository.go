package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"metaldocs/internal/modules/iam/authz"
	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/modules/taxonomy/domain"
)

type FamilyRepository struct {
	db *sql.DB
}

type familyTx struct {
	tx *sql.Tx
}

func (f familyTx) Commit() error   { return f.tx.Commit() }
func (f familyTx) Rollback() error { return f.tx.Rollback() }

func NewFamilyRepository(db *sql.DB) *FamilyRepository {
	return &FamilyRepository{db: db}
}

func (r *FamilyRepository) GetByCode(ctx context.Context, code string) (*domain.DocumentFamily, error) {
	const q = `
SELECT code, name, description, is_active, created_at
FROM metaldocs.document_families
WHERE code = $1`

	var f domain.DocumentFamily
	err := r.db.QueryRowContext(ctx, q, code).Scan(
		&f.Code, &f.Name, &f.Description, &f.IsActive, &f.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrFamilyNotFound
	}
	if err != nil {
		return nil, err
	}
	return &f, nil
}

func (r *FamilyRepository) List(ctx context.Context, includeInactive bool) ([]domain.DocumentFamily, error) {
	q := `
SELECT code, name, description, is_active, created_at
FROM metaldocs.document_families`
	if !includeInactive {
		q += " WHERE is_active = TRUE"
	}
	q += " ORDER BY code ASC"

	rows, err := r.db.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]domain.DocumentFamily, 0)
	for rows.Next() {
		var f domain.DocumentFamily
		if err := rows.Scan(&f.Code, &f.Name, &f.Description, &f.IsActive, &f.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, f)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *FamilyRepository) Create(ctx context.Context, f *domain.DocumentFamily) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin create family tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if err := authz.Require(ctx, tx, string(iamdomain.CapTaxonomyManage), "tenant"); err != nil {
		return fmt.Errorf("taxonomy: authz check Create family: %w", err)
	}
	const q = `
INSERT INTO metaldocs.document_families (code, name, description, is_active)
VALUES ($1, $2, $3, $4)`
	if _, err := tx.ExecContext(ctx, q, f.Code, f.Name, f.Description, f.IsActive); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *FamilyRepository) Update(ctx context.Context, f *domain.DocumentFamily) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin update family tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if err := authz.Require(ctx, tx, string(iamdomain.CapTaxonomyManage), "tenant"); err != nil {
		return fmt.Errorf("taxonomy: authz check Update family: %w", err)
	}
	const q = `
UPDATE metaldocs.document_families
SET name = $1, description = $2, is_active = $3
WHERE code = $4`
	result, err := tx.ExecContext(ctx, q, f.Name, f.Description, f.IsActive, f.Code)
	if err != nil {
		return err
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return domain.ErrFamilyNotFound
	}
	return tx.Commit()
}

func (r *FamilyRepository) HasActiveProfiles(ctx context.Context, tenantID, familyCode string) (bool, error) {
	const q = `
SELECT EXISTS(
  SELECT 1 FROM metaldocs.document_profiles
  WHERE tenant_id = $1 AND family_code = $2 AND archived_at IS NULL
)`
	var exists bool
	err := r.db.QueryRowContext(ctx, q, tenantID, familyCode).Scan(&exists)
	return exists, err
}

func (r *FamilyRepository) BeginTx(ctx context.Context) (domain.FamilyTx, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin family tx: %w", err)
	}
	return familyTx{tx: tx}, nil
}

func (r *FamilyRepository) GetByCodeForUpdate(ctx context.Context, tx domain.FamilyTx, code string) (*domain.DocumentFamily, error) {
	sqlTx, ok := tx.(familyTx)
	if !ok {
		return nil, fmt.Errorf("invalid family tx type %T", tx)
	}
	const q = `
SELECT code, name, description, is_active, created_at
FROM metaldocs.document_families
WHERE code = $1
FOR UPDATE`

	var f domain.DocumentFamily
	err := sqlTx.tx.QueryRowContext(ctx, q, code).Scan(
		&f.Code, &f.Name, &f.Description, &f.IsActive, &f.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrFamilyNotFound
	}
	if err != nil {
		return nil, err
	}
	return &f, nil
}

func (r *FamilyRepository) HasActiveProfilesTx(ctx context.Context, tx domain.FamilyTx, tenantID, familyCode string) (bool, error) {
	sqlTx, ok := tx.(familyTx)
	if !ok {
		return false, fmt.Errorf("invalid family tx type %T", tx)
	}
	const q = `
SELECT EXISTS(
  SELECT 1 FROM metaldocs.document_profiles
  WHERE tenant_id = $1 AND family_code = $2 AND archived_at IS NULL
)`
	var exists bool
	err := sqlTx.tx.QueryRowContext(ctx, q, tenantID, familyCode).Scan(&exists)
	return exists, err
}

func (r *FamilyRepository) UpdateTx(ctx context.Context, tx domain.FamilyTx, f *domain.DocumentFamily) error {
	sqlTx, ok := tx.(familyTx)
	if !ok {
		return fmt.Errorf("invalid family tx type %T", tx)
	}
	if err := authz.Require(ctx, sqlTx.tx, string(iamdomain.CapTaxonomyManage), "tenant"); err != nil {
		return fmt.Errorf("taxonomy: authz check Update family: %w", err)
	}
	const q = `
UPDATE metaldocs.document_families
SET name = $1, description = $2, is_active = $3
WHERE code = $4`
	result, err := sqlTx.tx.ExecContext(ctx, q, f.Name, f.Description, f.IsActive, f.Code)
	if err != nil {
		return err
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return domain.ErrFamilyNotFound
	}
	return nil
}
