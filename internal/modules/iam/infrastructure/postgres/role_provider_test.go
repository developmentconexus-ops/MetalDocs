package postgres_test

import (
	"context"
	"regexp"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"

	"metaldocs/internal/modules/iam/infrastructure/postgres"
	"metaldocs/internal/platform/tenant"
)

const testTenant = tenant.DevTenantID

func TestRolesByUserID_FiltersByTenant(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock new: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery(regexp.QuoteMeta(`
SELECT 1
FROM metaldocs.iam_users
WHERE user_id = $1 AND tenant_id = $2::uuid AND deactivated_at IS NULL
`)).WithArgs("alice", testTenant).WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(1))

	mock.ExpectQuery(regexp.QuoteMeta(`
SELECT role_code
FROM metaldocs.iam_user_roles
WHERE user_id = $1
  AND tenant_id = $2::uuid
ORDER BY role_code ASC
`)).WithArgs("alice", testTenant).
		WillReturnRows(sqlmock.NewRows([]string{"role_code"}).AddRow("author"))

	provider := postgres.NewRoleProvider(db)
	roles, err := provider.RolesByUserID(context.Background(), "alice", testTenant)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(roles) != 1 || string(roles[0]) != "author" {
		t.Fatalf("got %v, want [author]", roles)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
