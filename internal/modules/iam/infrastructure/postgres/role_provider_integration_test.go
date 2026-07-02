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

// seedWithCapsIAM asserts capsJSON via metaldocs.asserted_caps tx-locally
// (SEC-05 / migration 0259: metaldocs.iam_users and metaldocs.iam_user_roles now
// carry trg_require_cap_asserted, requiring user.manage for INSERT/DELETE and for
// UPDATE of privileged columns) then runs fn inside that same transaction and
// commits. Pool-safe local equivalent of tests/integration/testdb/fixtures.go's
// seedWithCaps — this package (postgres_test, package-local live-DB probes) does
// not import the tests/integration/testdb framework to avoid an import cycle
// (that framework itself exercises this package's repositories).
func seedWithCapsIAM(t *testing.T, db *sql.DB, capsJSON string, fn func(tx *sql.Tx) error) {
	t.Helper()
	ctx := context.Background()

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("seedWithCapsIAM: begin tx: %v", err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx,
		`SELECT set_config('metaldocs.asserted_caps', $1, true)`, capsJSON,
	); err != nil {
		t.Fatalf("seedWithCapsIAM: assert caps %s: %v", capsJSON, err)
	}
	if err := fn(tx); err != nil {
		t.Fatalf("seedWithCapsIAM: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("seedWithCapsIAM: commit: %v", err)
	}
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
