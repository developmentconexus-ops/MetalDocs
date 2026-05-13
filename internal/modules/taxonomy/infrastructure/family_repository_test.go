package infrastructure

import (
	"context"
	"regexp"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"metaldocs/internal/platform/tenant"
)

func TestFamilyRepository_HasActiveProfiles_TenantScoped(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	repo := NewFamilyRepository(db)
	ctx := tenant.WithTenantID(context.Background(), "tenant-a")

	mock.ExpectQuery(regexp.QuoteMeta(`
SELECT EXISTS(
  SELECT 1 FROM metaldocs.document_profiles
  WHERE tenant_id = $1 AND family_code = $2 AND archived_at IS NULL
)`)).
		WithArgs("tenant-a", "fam-1").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	exists, err := repo.HasActiveProfiles(ctx, "tenant-a", "fam-1")
	if err != nil {
		t.Fatalf("HasActiveProfiles: %v", err)
	}
	if !exists {
		t.Fatal("expected exists=true")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
