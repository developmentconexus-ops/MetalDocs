package infrastructure

import (
	"context"
	"database/sql"
	"errors"

	"metaldocs/internal/modules/taxonomy/domain"
)

type FamilyRepository struct {
	db *sql.DB
}

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
	return out, rows.Err()
}

func (r *FamilyRepository) Create(ctx context.Context, f *domain.DocumentFamily) error {
	const q = `
INSERT INTO metaldocs.document_families (code, name, description, is_active)
VALUES ($1, $2, $3, $4)`
	_, err := r.db.ExecContext(ctx, q, f.Code, f.Name, f.Description, f.IsActive)
	return err
}

func (r *FamilyRepository) Update(ctx context.Context, f *domain.DocumentFamily) error {
	const q = `
UPDATE metaldocs.document_families
SET name = $1, description = $2, is_active = $3
WHERE code = $4`
	result, err := r.db.ExecContext(ctx, q, f.Name, f.Description, f.IsActive, f.Code)
	if err != nil {
		return err
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return domain.ErrFamilyNotFound
	}
	return nil
}

func (r *FamilyRepository) HasActiveProfiles(ctx context.Context, familyCode string) (bool, error) {
	const q = `
SELECT EXISTS(
  SELECT 1 FROM metaldocs.document_profiles
  WHERE family_code = $1 AND archived_at IS NULL
)`
	var exists bool
	err := r.db.QueryRowContext(ctx, q, familyCode).Scan(&exists)
	return exists, err
}
