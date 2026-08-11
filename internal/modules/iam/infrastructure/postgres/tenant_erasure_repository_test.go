package postgres_test

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"

	"metaldocs/internal/modules/iam/infrastructure/postgres"
)

const tenantErasureQuery = `SELECT erased_at FROM metaldocs.tenants WHERE id = $1::uuid`

func TestTenantErasureRepositoryIsErased_ErasedTenant(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer func() { _ = db.Close() }()

	mock.ExpectQuery(regexp.QuoteMeta(tenantErasureQuery)).
		WithArgs("tenant-2").
		WillReturnRows(sqlmock.NewRows([]string{"erased_at"}).AddRow(time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)))

	repo := postgres.NewTenantErasureRepository(db)

	erased, err := repo.IsErased(context.Background(), "tenant-2")
	if err != nil {
		t.Fatalf("IsErased: %v", err)
	}
	if !erased {
		t.Fatalf("erased = false, want true for a non-null erased_at")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestTenantErasureRepositoryIsErased_ActiveTenant(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer func() { _ = db.Close() }()

	mock.ExpectQuery(regexp.QuoteMeta(tenantErasureQuery)).
		WithArgs("tenant-1").
		WillReturnRows(sqlmock.NewRows([]string{"erased_at"}).AddRow(nil))

	repo := postgres.NewTenantErasureRepository(db)

	erased, err := repo.IsErased(context.Background(), "tenant-1")
	if err != nil {
		t.Fatalf("IsErased: %v", err)
	}
	if erased {
		t.Fatalf("erased = true, want false for a null erased_at")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

// TestTenantErasureRepositoryIsErased_NotFoundFailsClosed pins the
// deliberate deviation from TenantRepository.TenantLookupTx: a tenantID with
// no metaldocs.tenants row must come back as an ERROR, not (false, nil).
// This is a read-gate deciding whether to withhold PII, not a best-effort
// existence check — treating "not found" as "safe to serve" would be exactly
// the kind of silent fallback the DEC-8 fix exists to close.
func TestTenantErasureRepositoryIsErased_NotFoundFailsClosed(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer func() { _ = db.Close() }()

	mock.ExpectQuery(regexp.QuoteMeta(tenantErasureQuery)).
		WithArgs("ghost-tenant").
		WillReturnError(sql.ErrNoRows)

	repo := postgres.NewTenantErasureRepository(db)

	_, err = repo.IsErased(context.Background(), "ghost-tenant")
	if err == nil {
		t.Fatalf("IsErased: want error for a not-found tenant, got nil")
	}
}

func TestTenantErasureRepositoryIsErased_QueryErrorPropagates(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer func() { _ = db.Close() }()

	dbErr := errors.New("connection reset")
	mock.ExpectQuery(regexp.QuoteMeta(tenantErasureQuery)).
		WithArgs("tenant-1").
		WillReturnError(dbErr)

	repo := postgres.NewTenantErasureRepository(db)

	_, err = repo.IsErased(context.Background(), "tenant-1")
	if err == nil {
		t.Fatalf("IsErased: want error, got nil")
	}
	if !errors.Is(err, dbErr) {
		t.Fatalf("IsErased error = %v, want it to wrap %v", err, dbErr)
	}
}
