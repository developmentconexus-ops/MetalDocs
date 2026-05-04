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

func TestUpsertUserAndAssignRole_PassesTenantID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO metaldocs.iam_users`)).
		WithArgs("alice", "Alice").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM metaldocs.iam_user_roles WHERE tenant_id = $1::uuid AND user_id = $2`)).
		WithArgs(testTenant, "alice").
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO metaldocs.iam_user_roles (user_id, tenant_id, role_code, assigned_at, assigned_by)`)).
		WithArgs("alice", testTenant, "author", "admin").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	repo := postgres.NewRoleAdminRepository(db)
	if err := repo.UpsertUserAndAssignRole(context.Background(), "alice", "Alice", testTenant, iamdomain.Role("author"), "admin"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestReplaceUserRoles_DeleteThenInsert_OnlyLastSurvives(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO metaldocs.iam_users`)).
		WithArgs("alice", "Alice").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM metaldocs.iam_user_roles WHERE tenant_id = $1::uuid AND user_id = $2`)).
		WithArgs(testTenant, "alice").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO metaldocs.iam_user_roles (user_id, tenant_id, role_code, assigned_at, assigned_by)`)).
		WithArgs("alice", testTenant, "approver", "admin").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	repo := postgres.NewRoleAdminRepository(db)
	if err := repo.ReplaceUserRoles(context.Background(), "alice", "Alice", testTenant,
		[]iamdomain.Role{"author", "approver"}, "admin"); err != nil {
		t.Fatalf("unexpected error: %v", err)
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
