package postgres_test

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"

	"metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/modules/iam/infrastructure/postgres"
	"metaldocs/internal/platform/tenant"
)

const testTenant = tenant.DevTenantID

// singleRoundTripSQL is the LEFT JOIN query that replaces the two-query path.
const singleRoundTripSQL = `
SELECT r.role_code
FROM metaldocs.iam_users u
LEFT JOIN metaldocs.iam_user_roles r
       ON r.user_id = u.user_id
      AND r.tenant_id = u.tenant_id
WHERE u.user_id = $1
  AND u.tenant_id = $2::uuid
  AND u.deactivated_at IS NULL
ORDER BY r.role_code ASC
`

func TestRolesByUserID_ActiveUserWithRole(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock new: %v", err)
	}
	defer db.Close()

	// Single query: LEFT JOIN returns one row with role_code = "author".
	mock.ExpectQuery(regexp.QuoteMeta(singleRoundTripSQL)).
		WithArgs("alice", testTenant).
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

func TestRolesByUserID_FiltersByTenant(t *testing.T) {
	// Alias for backward compat — same as ActiveUserWithRole but named for the
	// original test's intent (ensuring tenant_id is passed in the query).
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock new: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery(regexp.QuoteMeta(singleRoundTripSQL)).
		WithArgs("alice", testTenant).
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

func TestRolesByUserID_UnknownUser_ReturnsErrUserNotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock new: %v", err)
	}
	defer db.Close()

	// LEFT JOIN returns 0 rows → user not found.
	mock.ExpectQuery(regexp.QuoteMeta(singleRoundTripSQL)).
		WithArgs("ghost", testTenant).
		WillReturnRows(sqlmock.NewRows([]string{"role_code"}))

	provider := postgres.NewRoleProvider(db)
	_, roleErr := provider.RolesByUserID(context.Background(), "ghost", testTenant)
	if !errors.Is(roleErr, domain.ErrUserNotFound) {
		t.Fatalf("want ErrUserNotFound, got %v", roleErr)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestRolesByUserID_ActiveUserNoRoles_ReturnsErrNoRolesAssigned(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock new: %v", err)
	}
	defer db.Close()

	// LEFT JOIN returns 1 row with NULL role_code → user exists but has no roles.
	mock.ExpectQuery(regexp.QuoteMeta(singleRoundTripSQL)).
		WithArgs("noroles", testTenant).
		WillReturnRows(sqlmock.NewRows([]string{"role_code"}).AddRow(nil))

	provider := postgres.NewRoleProvider(db)
	_, roleErr := provider.RolesByUserID(context.Background(), "noroles", testTenant)
	if !errors.Is(roleErr, domain.ErrNoRolesAssigned) {
		t.Fatalf("want ErrNoRolesAssigned, got %v", roleErr)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestRolesByUserID_IsSingleRoundTrip(t *testing.T) {
	// Verify exactly ONE query is issued (not two as in the old path).
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock new: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery(regexp.QuoteMeta(singleRoundTripSQL)).
		WithArgs("alice", testTenant).
		WillReturnRows(sqlmock.NewRows([]string{"role_code"}).AddRow("editor"))

	provider := postgres.NewRoleProvider(db)
	if _, err := provider.RolesByUserID(context.Background(), "alice", testTenant); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("more than one query issued (N+1 regression): %v", err)
	}
}

func TestUserActiveInTenant_ActiveUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock new: %v", err)
	}
	defer db.Close()

	const existsSQL = `
SELECT EXISTS (
    SELECT 1
    FROM metaldocs.iam_users
    WHERE user_id = $1
      AND tenant_id = $2::uuid
      AND deactivated_at IS NULL
)
`
	mock.ExpectQuery(regexp.QuoteMeta(existsSQL)).
		WithArgs("alice", testTenant).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	provider := postgres.NewRoleProvider(db)
	active, err := provider.UserActiveInTenant(context.Background(), testTenant, "alice")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !active {
		t.Fatal("expected active=true")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestUserActiveInTenant_InactiveOrUnknownUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock new: %v", err)
	}
	defer db.Close()

	const existsSQL = `
SELECT EXISTS (
    SELECT 1
    FROM metaldocs.iam_users
    WHERE user_id = $1
      AND tenant_id = $2::uuid
      AND deactivated_at IS NULL
)
`
	mock.ExpectQuery(regexp.QuoteMeta(existsSQL)).
		WithArgs("ghost", testTenant).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	provider := postgres.NewRoleProvider(db)
	active, err := provider.UserActiveInTenant(context.Background(), testTenant, "ghost")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if active {
		t.Fatal("expected active=false for unknown/inactive user")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

// TestRolesByUserID_InactiveUser verifies that a user whose deactivated_at IS NOT
// NULL produces 0 rows from the LEFT JOIN, which the provider maps to ErrUserNotFound
// (identical semantics to the old two-query path).
func TestRolesByUserID_InactiveUser_ReturnsErrUserNotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock new: %v", err)
	}
	defer db.Close()

	// Inactive user: deactivated_at IS NOT NULL → the WHERE clause filters it
	// out → 0 rows returned → ErrUserNotFound.
	mock.ExpectQuery(regexp.QuoteMeta(singleRoundTripSQL)).
		WithArgs("inactive", testTenant).
		WillReturnRows(sqlmock.NewRows([]string{"role_code"}))

	provider := postgres.NewRoleProvider(db)
	_, roleErr := provider.RolesByUserID(context.Background(), "inactive", testTenant)

	// The LEFT JOIN collapses "not found" and "inactive" into the same 0-row
	// result, so we expect ErrUserNotFound in both cases — identical to the
	// old two-query path where the liveness check used ErrNoRows → ErrUserNotFound.
	if roleErr == nil || (!errors.Is(roleErr, domain.ErrUserNotFound)) {
		t.Fatalf("want ErrUserNotFound for inactive user, got %v", roleErr)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

// TestRolesByUserIDs_EmptyInput returns an empty map without querying the DB.
func TestRolesByUserIDs_EmptyInput(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock new: %v", err)
	}
	defer db.Close()

	provider := postgres.NewRoleProvider(db)
	out, err := provider.RolesByUserIDs(context.Background(), testTenant, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out) != 0 {
		t.Fatalf("expected empty map, got %v", out)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

// TestRolesByUserIDs_ActiveUsers verifies batch results and ordering.
func TestRolesByUserIDs_ActiveUsers(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock new: %v", err)
	}
	defer db.Close()

	const batchSQL = `
SELECT u.user_id, r.role_code
FROM metaldocs.iam_users u
LEFT JOIN metaldocs.iam_user_roles r
       ON r.user_id = u.user_id
      AND r.tenant_id = u.tenant_id
WHERE u.user_id = ANY($1)
  AND u.tenant_id = $2::uuid
  AND u.deactivated_at IS NULL
ORDER BY u.user_id, r.role_code ASC
`
	rows := sqlmock.NewRows([]string{"user_id", "role_code"}).
		AddRow("alice", "author").
		AddRow("bob", sql.NullString{}) // active, no roles → NULL

	mock.ExpectQuery(regexp.QuoteMeta(batchSQL)).
		WithArgs(sqlmock.AnyArg(), testTenant).
		WillReturnRows(rows)

	provider := postgres.NewRoleProvider(db)
	out, err := provider.RolesByUserIDs(context.Background(), testTenant, []string{"alice", "bob", "ghost"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// alice should have [author]
	aliceRoles, ok := out["alice"]
	if !ok || len(aliceRoles) != 1 || string(aliceRoles[0]) != "author" {
		t.Fatalf("alice: got %v, want [author]", aliceRoles)
	}
	// bob is present with empty slice (active, no roles)
	bobRoles, ok := out["bob"]
	if !ok {
		t.Fatal("bob should be in map (active user with no roles)")
	}
	if len(bobRoles) != 0 {
		t.Fatalf("bob: got %v, want []", bobRoles)
	}
	// ghost is absent (not found)
	if _, ok := out["ghost"]; ok {
		t.Fatal("ghost should be absent from map")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
