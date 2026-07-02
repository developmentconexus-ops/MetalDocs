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

// db.Tx forwarding — satisfies domain.FamilyTx which embeds db.Tx (F-07).
func (f familyTx) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return f.tx.ExecContext(ctx, query, args...)
}
func (f familyTx) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return f.tx.QueryContext(ctx, query, args...)
}
func (f familyTx) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	return f.tx.QueryRowContext(ctx, query, args...)
}

func NewFamilyRepository(db *sql.DB) *FamilyRepository {
	return &FamilyRepository{db: db}
}

func (r *FamilyRepository) GetByCode(ctx context.Context, tenantID string, code domain.FamilyCode) (*domain.DocumentFamily, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin get family tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if err := setAuthzGUC(ctx, tx); err != nil {
		return nil, fmt.Errorf("query family %q: %w", code, err)
	}
	if err := authz.Require(ctx, tx, string(iamdomain.CapTaxonomyView), "tenant"); err != nil {
		return nil, fmt.Errorf("taxonomy: authz check Get family: %w", err)
	}

	const q = `
SELECT code, tenant_id, name, description, is_active, created_at
FROM metaldocs.document_families
WHERE tenant_id = $1 AND code = $2`

	var f domain.DocumentFamily
	err = tx.QueryRowContext(ctx, q, tenantID, code).Scan(
		&f.Code, &f.TenantID, &f.Name, &f.Description, &f.IsActive, &f.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrFamilyNotFound
	}
	if err != nil {
		return nil, err
	}
	return &f, nil
}

func (r *FamilyRepository) List(ctx context.Context, tenantID string, includeInactive bool) ([]domain.DocumentFamily, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin list families tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if err := setAuthzGUC(ctx, tx); err != nil {
		return nil, err
	}
	if err := authz.Require(ctx, tx, string(iamdomain.CapTaxonomyView), "tenant"); err != nil {
		return nil, fmt.Errorf("taxonomy: authz check List families: %w", err)
	}

	q := `
SELECT code, tenant_id, name, description, is_active, created_at
FROM metaldocs.document_families
WHERE tenant_id = $1`
	if !includeInactive {
		q += " AND is_active = TRUE"
	}
	q += " ORDER BY code ASC"

	rows, err := tx.QueryContext(ctx, q, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]domain.DocumentFamily, 0)
	for rows.Next() {
		var f domain.DocumentFamily
		if err := rows.Scan(&f.Code, &f.TenantID, &f.Name, &f.Description, &f.IsActive, &f.CreatedAt); err != nil {
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

	if err := setAuthzGUC(ctx, tx); err != nil {
		return fmt.Errorf("insert family %q: %w", f.Code, err)
	}
	if err := authz.Require(ctx, tx, string(iamdomain.CapTaxonomyManage), "tenant"); err != nil {
		return fmt.Errorf("taxonomy: authz check Create family: %w", err)
	}
	const q = `
INSERT INTO metaldocs.document_families (code, tenant_id, name, description, is_active)
VALUES ($1, $2, $3, $4, $5)`
	if _, err := tx.ExecContext(ctx, q, f.Code, f.TenantID, f.Name, f.Description, f.IsActive); err != nil {
		return err
	}
	return tx.Commit()
}

// CreateTx inserts a new DocumentFamily using the caller-owned transaction.
// The caller is responsible for committing or rolling back the transaction.
func (r *FamilyRepository) CreateTx(ctx context.Context, tx domain.FamilyTx, f *domain.DocumentFamily) error {
	sqlTx, ok := tx.(familyTx)
	if !ok {
		return fmt.Errorf("taxonomy: FamilyRepository.CreateTx: unsupported tx type %T", tx)
	}
	if err := setAuthzGUC(ctx, sqlTx.tx); err != nil {
		return fmt.Errorf("insert family %q: %w", f.Code, err)
	}
	if err := authz.Require(ctx, sqlTx.tx, string(iamdomain.CapTaxonomyManage), "tenant"); err != nil {
		return fmt.Errorf("taxonomy: authz check CreateTx family: %w", err)
	}
	const q = `
INSERT INTO metaldocs.document_families (code, tenant_id, name, description, is_active)
VALUES ($1, $2, $3, $4, $5)`
	_, err := sqlTx.tx.ExecContext(ctx, q, f.Code, f.TenantID, f.Name, f.Description, f.IsActive)
	return err
}

func (r *FamilyRepository) Update(ctx context.Context, f *domain.DocumentFamily) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin update family tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if err := setAuthzGUC(ctx, tx); err != nil {
		return err
	}
	if err := authz.Require(ctx, tx, string(iamdomain.CapTaxonomyManage), "tenant"); err != nil {
		return fmt.Errorf("taxonomy: authz check Update family: %w", err)
	}
	const q = `
UPDATE metaldocs.document_families
SET name = $1, description = $2, is_active = $3
WHERE tenant_id = $4 AND code = $5`
	result, err := tx.ExecContext(ctx, q, f.Name, f.Description, f.IsActive, f.TenantID, f.Code)
	if err != nil {
		return fmt.Errorf("update family %q: %w", f.Code, err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("family update rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return domain.ErrFamilyNotFound
	}
	return tx.Commit()
}

func (r *FamilyRepository) HasActiveProfiles(ctx context.Context, tenantID string, familyCode domain.FamilyCode) (bool, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return false, fmt.Errorf("begin has active profiles tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if err := setAuthzGUC(ctx, tx); err != nil {
		return false, err
	}
	if err := authz.Require(ctx, tx, string(iamdomain.CapTaxonomyView), "tenant"); err != nil {
		return false, fmt.Errorf("taxonomy: authz check HasActiveProfiles: %w", err)
	}

	const q = `
SELECT EXISTS(
  SELECT 1 FROM metaldocs.document_profiles
  WHERE tenant_id = $1 AND family_code = $2 AND archived_at IS NULL
)`
	var exists bool
	err = tx.QueryRowContext(ctx, q, tenantID, familyCode).Scan(&exists)
	return exists, err
}

func (r *FamilyRepository) BeginTx(ctx context.Context) (domain.FamilyTx, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin family tx: %w", err)
	}
	return familyTx{tx: tx}, nil
}

func (r *FamilyRepository) GetByCodeForUpdate(ctx context.Context, tx domain.FamilyTx, tenantID string, code domain.FamilyCode) (*domain.DocumentFamily, error) {
	sqlTx, ok := tx.(familyTx)
	if !ok {
		return nil, fmt.Errorf("invalid family tx type %T", tx)
	}
	if err := setAuthzGUC(ctx, sqlTx.tx); err != nil {
		return nil, fmt.Errorf("query family for update %q: %w", code, err)
	}
	if err := authz.Require(ctx, sqlTx.tx, string(iamdomain.CapTaxonomyManage), "tenant"); err != nil {
		return nil, fmt.Errorf("taxonomy: authz check Get family for update: %w", err)
	}
	const q = `
SELECT code, tenant_id, name, description, is_active, created_at
FROM metaldocs.document_families
WHERE tenant_id = $1 AND code = $2
FOR UPDATE`

	var f domain.DocumentFamily
	err := sqlTx.tx.QueryRowContext(ctx, q, tenantID, code).Scan(
		&f.Code, &f.TenantID, &f.Name, &f.Description, &f.IsActive, &f.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrFamilyNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("query family for update %q: %w", code, err)
	}
	return &f, nil
}

func (r *FamilyRepository) HasActiveProfilesTx(ctx context.Context, tx domain.FamilyTx, tenantID string, familyCode domain.FamilyCode) (bool, error) {
	sqlTx, ok := tx.(familyTx)
	if !ok {
		return false, fmt.Errorf("invalid family tx type %T", tx)
	}
	if err := setAuthzGUC(ctx, sqlTx.tx); err != nil {
		return false, err
	}
	if err := authz.Require(ctx, sqlTx.tx, string(iamdomain.CapTaxonomyManage), "tenant"); err != nil {
		return false, fmt.Errorf("taxonomy: authz check HasActiveProfilesTx: %w", err)
	}
	const q = `
SELECT EXISTS(
  SELECT 1 FROM metaldocs.document_profiles
  WHERE tenant_id = $1 AND family_code = $2 AND archived_at IS NULL
)`
	var exists bool
	err := sqlTx.tx.QueryRowContext(ctx, q, tenantID, familyCode).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("query active profiles for family %q: %w", familyCode, err)
	}
	return exists, nil
}

func (r *FamilyRepository) UpdateTx(ctx context.Context, tx domain.FamilyTx, f *domain.DocumentFamily) error {
	sqlTx, ok := tx.(familyTx)
	if !ok {
		return fmt.Errorf("invalid family tx type %T", tx)
	}
	if err := setAuthzGUC(ctx, sqlTx.tx); err != nil {
		return err
	}
	if err := authz.Require(ctx, sqlTx.tx, string(iamdomain.CapTaxonomyManage), "tenant"); err != nil {
		return fmt.Errorf("taxonomy: authz check Update family: %w", err)
	}
	const q = `
UPDATE metaldocs.document_families
SET name = $1, description = $2, is_active = $3
WHERE tenant_id = $4 AND code = $5`
	result, err := sqlTx.tx.ExecContext(ctx, q, f.Name, f.Description, f.IsActive, f.TenantID, f.Code)
	if err != nil {
		return fmt.Errorf("update family tx %q: %w", f.Code, err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("family tx update rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return domain.ErrFamilyNotFound
	}
	return nil
}
