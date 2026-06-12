//go:build integration

package postgres_test

// Live integration probe for postgres.RoleProvider.UserActiveInTenant (Task 4, 2.12 step 4/4).
//
// Verifies the EXISTS query returns the correct answer against a live Postgres
// instance. Reads the known admin user from the system tenant (ffffffff-...) and
// asserts true; uses a nil UUID for the false case.
//
// Run:
//
//	DATABASE_URL=postgres://... go test -tags integration -run TestRoleProvider_UserActiveInTenant_Live ./internal/modules/iam/infrastructure/postgres/
//	METALDOCS_DATABASE_URL=postgres://... go test -tags integration -run TestRoleProvider_UserActiveInTenant_Live ./internal/modules/iam/infrastructure/postgres/
import (
	"context"
	"database/sql"
	"os"
	"strings"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"

	iampg "metaldocs/internal/modules/iam/infrastructure/postgres"
)

func openLiveIAMDB(t *testing.T) *sql.DB {
	t.Helper()
	dsn := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if dsn == "" {
		dsn = strings.TrimSpace(os.Getenv("METALDOCS_DATABASE_URL"))
	}
	if dsn == "" {
		t.Skip("no DATABASE_URL or METALDOCS_DATABASE_URL set — skipping live DB probe")
	}
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("sql.Open(pgx): %v", err)
	}
	if err := db.PingContext(context.Background()); err != nil {
		t.Fatalf("db.Ping: %v", err)
	}
	return db
}

// TestRoleProvider_UserActiveInTenant_Live exercises the EXISTS query against
// the real metaldocs-postgres container.
//
// Assertions:
//
//	(i)  admin / ffffffff-ffff-ffff-ffff-ffffffffffff → true  (known active user)
//	(ii) 00000000-0000-0000-0000-000000000000 / system tenant → false (nil UUID)
func TestRoleProvider_UserActiveInTenant_Live(t *testing.T) {
	db := openLiveIAMDB(t)
	defer db.Close()

	provider := iampg.NewRoleProvider(db)
	ctx := context.Background()

	const systemTenant = "ffffffff-ffff-ffff-ffff-ffffffffffff"

	// (i) Known admin user must be active in system tenant.
	active, err := provider.UserActiveInTenant(ctx, systemTenant, "admin")
	if err != nil {
		t.Fatalf("UserActiveInTenant(admin, system_tenant): %v", err)
	}
	if !active {
		t.Fatal("expected active=true for known admin user in system tenant")
	}
	t.Logf("PASS: admin in system-tenant → active=true")

	// (ii) Nil UUID must not be found.
	active, err = provider.UserActiveInTenant(ctx, systemTenant, "00000000-0000-0000-0000-000000000000")
	if err != nil {
		t.Fatalf("UserActiveInTenant(nil-uuid, system_tenant): %v", err)
	}
	if active {
		t.Fatal("expected active=false for fictitious nil-UUID user")
	}
	t.Logf("PASS: nil-uuid in system-tenant → active=false")
}
