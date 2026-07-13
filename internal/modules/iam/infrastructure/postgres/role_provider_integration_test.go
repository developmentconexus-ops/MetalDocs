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
	"metaldocs/tests/integration/testdb"
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
// commits. Thin package-local alias for the canonical testdb.SeedWithCaps helper
// (importing tests/integration/testdb from here is cycle-free: the framework's
// only internal/modules dependency is controlleddocuments/{domain,infrastructure};
// the historical import-cycle constraint applies to THAT package only).
func seedWithCapsIAM(t *testing.T, db *sql.DB, capsJSON string, fn func(tx *sql.Tx) error) {
	t.Helper()
	testdb.SeedWithCaps(t, db, capsJSON, fn)
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
	// Register close as a t.Cleanup (not defer) so it runs AFTER the seed-teardown
	// Cleanup below — t.Cleanup is LIFO and this is registered first, so the DELETE
	// runs while the pool is still open (a plain `defer db.Close()` would close the
	// pool before any t.Cleanup, breaking the teardown with "database is closed").
	t.Cleanup(func() { db.Close() })

	provider := iampg.NewRoleProvider(db)
	ctx := context.Background()

	const systemTenant = "ffffffff-ffff-ffff-ffff-ffffffffffff"

	// Shared-DB hardening: this Live probe reads a dev-seeded `admin` identity,
	// but runs against the shared metaldocs database whose seed state is not
	// guaranteed (testdb.DSN — no per-test clone; a rebuilt DB may omit it).
	// Provision an active admin in the system tenant if absent and remove ONLY
	// what we created, so the probe is hermetic while still validating the real
	// EXISTS query. iam_users carries trg_require_cap_asserted (SEC-05 / 0259,
	// user.manage) — seed with that capability asserted.
	seededAdmin := false
	seedWithCapsIAM(t, db, `[{"cap":"user.manage"}]`, func(tx *sql.Tx) error {
		res, err := tx.ExecContext(ctx,
			`INSERT INTO metaldocs.iam_users (user_id, display_name, tenant_id)
			 VALUES ('admin', 'Seeded Admin Probe', $1::uuid)
			 ON CONFLICT (user_id) DO NOTHING`,
			systemTenant,
		)
		if err != nil {
			return err
		}
		if n, _ := res.RowsAffected(); n == 1 {
			seededAdmin = true
		}
		return nil
	})
	if seededAdmin {
		t.Cleanup(func() {
			seedWithCapsIAM(t, db, `[{"cap":"user.manage"}]`, func(tx *sql.Tx) error {
				_, err := tx.ExecContext(context.Background(),
					`DELETE FROM metaldocs.iam_users WHERE user_id = 'admin'`)
				return err
			})
		})
	}

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
