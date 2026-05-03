package postgres_test

import (
	"context"
	"regexp"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"

	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/modules/iam/infrastructure/postgres"
)

func TestHasAnyRole_FiltersByTenant(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock new: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery(regexp.QuoteMeta(`
SELECT COUNT(*)
FROM metaldocs.iam_user_roles
WHERE role_code = $1
  AND tenant_id = $2::uuid
`)).WithArgs("system_admin", testTenant).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	repo := postgres.NewRoleAdminRepository(db)
	found, err := repo.HasAnyRole(context.Background(), iamdomain.RoleSystemAdmin, testTenant)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !found {
		t.Fatal("expected HasAnyRole to return true")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestHasAnyRole_OtherTenantReturnsZero(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock new: %v", err)
	}
	defer db.Close()

	otherTenant := "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	mock.ExpectQuery(regexp.QuoteMeta(`
SELECT COUNT(*)
FROM metaldocs.iam_user_roles
WHERE role_code = $1
  AND tenant_id = $2::uuid
`)).WithArgs("system_admin", otherTenant).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	repo := postgres.NewRoleAdminRepository(db)
	found, err := repo.HasAnyRole(context.Background(), iamdomain.RoleSystemAdmin, otherTenant)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found {
		t.Fatal("expected HasAnyRole to return false for different tenant")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
