package postgres_test

import (
	"context"
	"errors"
	"regexp"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"

	"metaldocs/internal/modules/iam/authz"
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
	mock.ExpectExec(regexp.QuoteMeta(`
SELECT
	set_config('metaldocs.tenant_id', $1, true),
	set_config('metaldocs.actor_id', $2, true)
`)).
		WithArgs(testTenant, "admin").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT current_setting('metaldocs.actor_id', true)")).
		WillReturnRows(sqlmock.NewRows([]string{"current_setting"}).AddRow("actor-1"))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT current_setting('metaldocs.tenant_id', true)")).
		WillReturnRows(sqlmock.NewRows([]string{"current_setting"}).AddRow(testTenant))
	expectSystemAdminBypass(mock, "actor-1", testTenant, false)
	mock.ExpectQuery(regexp.QuoteMeta(`
SELECT EXISTS (
  SELECT 1
  FROM metaldocs.role_capabilities rc
  JOIN metaldocs.user_process_areas upa
    ON upa.role = rc.role
   AND upa.tenant_id = $4::uuid
   AND upa.user_id   = $3
   AND upa.effective_to IS NULL
  WHERE rc.capability = $1
    AND ($2 = 'tenant' OR upa.area_code = $2)
)`)).WithArgs(string(iamdomain.CapUserManage), "tenant", "actor-1", testTenant).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT current_setting('metaldocs.asserted_caps', true)")).
		WillReturnRows(sqlmock.NewRows([]string{"current_setting"}).AddRow(""))
	mock.ExpectExec(regexp.QuoteMeta("SELECT set_config('metaldocs.asserted_caps', $1, true)")).
		WithArgs(`[{"area":"tenant","cap":"user.manage"}]`).
		WillReturnResult(sqlmock.NewResult(0, 1))
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

func TestUpsertUserAndAssignRole_WrapsAuthzError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectRollback()

	repo := postgres.NewRoleAdminRepository(db)
	err = repo.UpsertUserAndAssignRole(context.Background(), "alice", "Alice", testTenant, iamdomain.Role("author"), "")
	if err == nil {
		t.Fatal("expected authz error")
	}
	if !errors.Is(err, authz.ErrActorContextMissing) {
		t.Fatalf("expected ErrActorContextMissing, got %v", err)
	}
	if got := err.Error(); got != "seed iam user.manage authorization: authz: metaldocs.actor_id GUC not set on transaction" {
		t.Fatalf("unexpected error: %q", got)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestReplaceUserRoles_DeleteThenInsert_PersistsSingleRole(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`
SELECT
	set_config('metaldocs.tenant_id', $1, true),
	set_config('metaldocs.actor_id', $2, true)
`)).
		WithArgs(testTenant, "admin").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT current_setting('metaldocs.actor_id', true)")).
		WillReturnRows(sqlmock.NewRows([]string{"current_setting"}).AddRow("actor-1"))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT current_setting('metaldocs.tenant_id', true)")).
		WillReturnRows(sqlmock.NewRows([]string{"current_setting"}).AddRow(testTenant))
	expectSystemAdminBypass(mock, "actor-1", testTenant, false)
	mock.ExpectQuery(regexp.QuoteMeta(`
SELECT EXISTS (
  SELECT 1
  FROM metaldocs.role_capabilities rc
  JOIN metaldocs.user_process_areas upa
    ON upa.role = rc.role
   AND upa.tenant_id = $4::uuid
   AND upa.user_id   = $3
   AND upa.effective_to IS NULL
  WHERE rc.capability = $1
    AND ($2 = 'tenant' OR upa.area_code = $2)
)`)).WithArgs(string(iamdomain.CapUserManage), "tenant", "actor-1", testTenant).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT current_setting('metaldocs.asserted_caps', true)")).
		WillReturnRows(sqlmock.NewRows([]string{"current_setting"}).AddRow(""))
	mock.ExpectExec(regexp.QuoteMeta("SELECT set_config('metaldocs.asserted_caps', $1, true)")).
		WithArgs(`[{"area":"tenant","cap":"user.manage"}]`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO metaldocs.iam_users`)).
		WithArgs("alice", "Alice").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM metaldocs.iam_user_roles WHERE tenant_id = $1::uuid AND user_id = $2`)).
		WithArgs(testTenant, "alice").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(`
INSERT INTO metaldocs.iam_user_roles (user_id, tenant_id, role_code, assigned_at, assigned_by)
VALUES ($1, $2::uuid, $3, NOW(), $4)
`)).
		WithArgs("alice", testTenant, "editor", "admin").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	repo := postgres.NewRoleAdminRepository(db)
	if err := repo.ReplaceUserRoles(context.Background(), "alice", "Alice", testTenant, iamdomain.RoleEditor, "admin"); err != nil {
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

func expectSystemAdminBypass(mock sqlmock.Sqlmock, actorID, tenantID string, exists bool) {
	mock.ExpectQuery(regexp.QuoteMeta(`
SELECT EXISTS (
  SELECT 1
    FROM metaldocs.iam_user_roles ur
   WHERE ur.user_id   = $1
     AND ur.tenant_id = $2::uuid
     AND ur.role_code = 'system_admin'
  UNION ALL
  SELECT 1
    FROM metaldocs.iam_group_members gm
    JOIN metaldocs.iam_groups g ON g.id = gm.group_id
    JOIN metaldocs.iam_group_roles gr ON gr.group_id = gm.group_id
   WHERE gm.user_id = $1
     AND gm.tenant_id = $2::uuid
     AND g.tenant_id = $2::uuid
     AND gr.role = 'system_admin'
)`)).
		WithArgs(actorID, tenantID).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(exists))
}
