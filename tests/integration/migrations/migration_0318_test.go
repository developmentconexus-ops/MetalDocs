//go:build integration

package migrations_test

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/google/uuid"

	"metaldocs/internal/platform/tenant"
	"metaldocs/tests/integration/testdb"
)

// TestMigration0318_SchemaLanded proves the two new relations and the
// prerequisite unique keys exist after bootstrap, and that the ledger row was
// written.
func TestMigration0318_SchemaLanded(t *testing.T) {
	ctx := context.Background()
	db, _ := testdb.OpenFreshDatabase(t)

	if got := regclass(t, ctx, db, "metaldocs.capability_bindings"); !got.Valid {
		t.Fatal("metaldocs.capability_bindings does not exist after bootstrap")
	}
	if got := regclass(t, ctx, db, "metaldocs.roles"); !got.Valid {
		t.Fatal("metaldocs.roles does not exist after bootstrap")
	}
	if got := regclass(t, ctx, db, "metaldocs.iam_users_tenant_user_uk"); !got.Valid {
		t.Fatal("iam_users_tenant_user_uk (promoted unique constraint) is missing after bootstrap")
	}
	if got := regclass(t, ctx, db, "metaldocs.iam_groups_tenant_id_id_uk"); !got.Valid {
		t.Fatal("iam_groups_tenant_id_id_uk is missing after bootstrap")
	}

	if n := ledgerCount(t, ctx, db, "0318"); n != 1 {
		t.Fatalf("expected exactly one schema_migrations row for version '0318', got %d", n)
	}

	var roleCount int
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM metaldocs.roles`).Scan(&roleCount); err != nil {
		t.Fatalf("count metaldocs.roles: %v", err)
	}
	if roleCount != 8 {
		t.Fatalf("expected 8 hand-seeded roles, got %d", roleCount)
	}
}

// TestMigration0318_ReplaySafe proves migration 0318 is safe to execute a
// second time against a database where it already succeeded (partial
// restore / operator replay -- P1-2 from the PR #113 bot review). This is
// the negative case the DDL/backfill guards in 0318 exist to prevent: a
// naive (unguarded) version of this migration fails the DDL outright with
// duplicate_object on the second CREATE TABLE/CREATE POLICY, or -- worse,
// silently -- doubles every capability_bindings row on the second backfill
// INSERT (capability_bindings carries no uniqueness constraint at all over
// historical/revoked rows, only over the active slice). The test seeds real
// source rows, runs the file once to backfill them for real, snapshots, runs
// it again as the replay under test, and asserts the row counts are
// byte-identical before and after that replay -- what would catch a guard
// regression; asserting only "no error" would not have caught the
// silent-duplication failure mode.
//
// This single fixture also discharges the duplicate-iam_group_roles-seeding
// regression deferred from the PR #113 fix round's Finding 5: a duplicate
// (group_id, role) row is unrepresentable at the source (iam_group_roles has
// a composite PRIMARY KEY on (group_id, role) -- see the corrected comment
// on the Source 3 backfill in the migration itself), so a fixture that seeds
// one cannot be built; replay safety is the only remaining edge this
// migration's guards need to prove, and it is the same DISTINCT/backfill
// code path.
func TestMigration0318_ReplaySafe(t *testing.T) {
	ctx := context.Background()
	db, _ := testdb.OpenFreshDatabase(t)
	// Sequential-only usage in this test; pinning the pool to one physical
	// connection makes execWithSchedulerBypass's pinned-conn SET/RESET and
	// assertBypassNotInherited's leak check deterministic (same idiom
	// testdb.SetCapsOnDB documents: "Safe only for isolated per-test
	// databases (MaxOpenConns=1 ...)").
	db.SetMaxOpenConns(1)

	// testdb.OpenFreshDatabase's bootstrap already runs 0318 once, but
	// against an otherwise-empty database (no dev-seed grants), so all
	// three backfill sources are empty and capability_bindings ends up
	// empty too -- a replay of "0 rows before, 0 rows after" would pass
	// even with the backfill guards deleted entirely, proving nothing.
	// Seed one row into each source first, matching a production
	// deployment where 0318 lands against a live DB that already has
	// grants to carry forward.
	seedReplaySafeSourceRows(t, ctx, db)

	type counts struct {
		total          int
		fromUserRoles  int
		fromAreaGrants int
		fromGroupRoles int
		roles          int
	}
	snapshot := func() counts {
		t.Helper()
		var c counts
		if err := db.QueryRowContext(ctx, `SELECT count(*) FROM metaldocs.capability_bindings`).Scan(&c.total); err != nil {
			t.Fatalf("count capability_bindings: %v", err)
		}
		if err := db.QueryRowContext(ctx,
			`SELECT count(*) FROM metaldocs.capability_bindings WHERE source_relation = 'iam_user_roles'`).Scan(&c.fromUserRoles); err != nil {
			t.Fatalf("count capability_bindings (iam_user_roles): %v", err)
		}
		if err := db.QueryRowContext(ctx,
			`SELECT count(*) FROM metaldocs.capability_bindings WHERE source_relation = 'user_process_areas'`).Scan(&c.fromAreaGrants); err != nil {
			t.Fatalf("count capability_bindings (user_process_areas): %v", err)
		}
		if err := db.QueryRowContext(ctx,
			`SELECT count(*) FROM metaldocs.capability_bindings WHERE source_relation = 'iam_group_roles'`).Scan(&c.fromGroupRoles); err != nil {
			t.Fatalf("count capability_bindings (iam_group_roles): %v", err)
		}
		if err := db.QueryRowContext(ctx, `SELECT count(*) FROM metaldocs.roles`).Scan(&c.roles); err != nil {
			t.Fatalf("count metaldocs.roles: %v", err)
		}
		return c
	}

	sqlBytes, err := os.ReadFile(migration0318Path(t))
	if err != nil {
		t.Fatalf("read migration 0318 file: %v", err)
	}
	sqlText := string(sqlBytes)

	// By the time this test runs, bootstrap has already applied 0319, which
	// ATTACHES trg_require_cap_asserted to metaldocs.roles and
	// metaldocs.capability_bindings (arms #21/#22). Re-executing 0318's raw
	// SQL now -- outside any request/service context that would assert a
	// capability -- hits that now-live trigger on its own backfill INSERTs,
	// which a real 0318-only run (before 0319 has run) would not. The
	// scheduler bypass is the same escape hatch background/administrative
	// SQL uses elsewhere (see metaldocs.bypass_authz='scheduler' in the
	// tenant-erasure path). execWithSchedulerBypass sets and resets it on a
	// single pinned *sql.Conn (B2, PR #113 bot review) -- SET is
	// session-scoped, so leaving it set on a connection returned to db's
	// pool would hand a later caller on this same *sql.DB an unintended
	// tripwire bypass. SET LOCAL is not usable here: this runs outside an
	// explicit Go-managed transaction (0318's own top-level BEGIN/COMMIT is
	// inside sqlText itself).

	// First pass: bootstrap's own 0318 run already happened over an empty
	// database, so the backfill guards were never blocked and never had
	// anything to insert either. Running the file again now -- the first
	// time capability_bindings' per-source guard markers are absent AND
	// the sources are non-empty -- is the real backfill this migration is
	// meant to perform in production; it is not yet the replay under test.
	// These are genuine new-row inserts into capability_bindings (three
	// source rows), so they DO need a capability assertion -- unlike the
	// true replay pass below, this is not a B1 scenario.
	if err := execWithSchedulerBypass(ctx, t, db, sqlText); err != nil {
		t.Fatalf("running migration 0318 against freshly-seeded source rows failed: %v", err)
	}
	assertBypassNotInherited(t, ctx, db)

	before := snapshot()
	if before.total == 0 {
		t.Fatal("precondition failed: capability_bindings is empty after the seeded backfill pass, replay would prove nothing")
	}

	// Second pass: the actual replay under test (P1-2/B1) -- every guard,
	// including the roles seed (B1 fix), must now see zero candidate rows
	// and never fire the tripwire trigger at all. Deliberately NO scheduler
	// bypass and NO asserted_caps here: a true replay against an
	// already-fully-migrated database needs neither.
	if _, err := db.ExecContext(ctx, sqlText); err != nil {
		t.Fatalf("replaying migration 0318 against an already-migrated database failed (not replay-safe): %v", err)
	}

	after := snapshot()
	if after != before {
		t.Fatalf("migration 0318 replay changed row counts (backfill duplicated rows): before=%+v after=%+v", before, after)
	}
}

// TestMigration0318_ReplaySafeAfterTripwireAttached proves migration 0318 is
// safe to replay AFTER migration 0319 has attached trg_require_cap_asserted
// to metaldocs.roles and metaldocs.capability_bindings, WITH NO capability
// bypass or assertion at all -- the exact "apply 0318, apply 0319, replay
// 0318" sequence PR #113 bot review found broken (B1, finding on
// db/migrations/0318_capability_bindings_schema_backfill.sql:214).
//
// RED before the fix: the roles seed used `ON CONFLICT (code) DO NOTHING`,
// which still evaluates trg_require_cap_asserted for every candidate row
// BEFORE Postgres resolves the conflict -- so a bare replay (no
// bypass_authz, no asserted_caps) hit ErrCapabilityNotAsserted on all 8 seed
// rows. GREEN after the fix: the seed now uses the same
// `WHERE NOT EXISTS (SELECT 1 FROM metaldocs.roles)` guard shape as the
// three backfill INSERTs, so a replay's candidate row count is zero and the
// trigger never fires -- no capability assertion is needed for a true
// replay, matching every other guarded statement in this file.
//
// Deliberately decoupled from TestMigration0318_ReplaySafe's real-backfill
// scenario (which legitimately needs a capability bypass for its first,
// genuine-write pass): this test starts from bootstrap's own state --
// nothing seeded into any of the three source tables -- so both passes here
// are true, zero-candidate-row replays.
func TestMigration0318_ReplaySafeAfterTripwireAttached(t *testing.T) {
	ctx := context.Background()
	db, _ := testdb.OpenFreshDatabase(t)

	// testdb.OpenFreshDatabase's bootstrap has already applied every
	// migration through the current head over an empty database -- 0318
	// (roles seeded, capability_bindings created with nothing to backfill)
	// and 0319 (trg_require_cap_asserted attached to both new tables).
	before := roleAndBindingCounts(t, ctx, db)
	if before.roles != 8 {
		t.Fatalf("precondition failed: expected 8 roles after bootstrap, got %d", before.roles)
	}

	sqlBytes, err := os.ReadFile(migration0318Path(t))
	if err != nil {
		t.Fatalf("read migration 0318 file: %v", err)
	}

	// Deliberately NO SET metaldocs.bypass_authz and NO
	// metaldocs.asserted_caps -- this is the crux of B1: a true replay
	// against an already-fully-migrated database must need neither, because
	// every guarded statement must produce zero candidate rows and never
	// fire trg_require_cap_asserted at all.
	if _, err := db.ExecContext(ctx, string(sqlBytes)); err != nil {
		t.Fatalf("replaying migration 0318 after 0319 attached the tripwire failed with no capability bypass or assertion set (B1 regression): %v", err)
	}

	after := roleAndBindingCounts(t, ctx, db)
	if after != before {
		t.Fatalf("migration 0318 replay after 0319 changed row counts: before=%+v after=%+v", before, after)
	}
}

// roleAndBindingCounts snapshots metaldocs.roles and metaldocs.capability_bindings
// row counts, for TestMigration0318_ReplaySafeAfterTripwireAttached's
// before/after comparison.
func roleAndBindingCounts(t *testing.T, ctx context.Context, db *sql.DB) (counts struct{ roles, bindings int }) {
	t.Helper()
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM metaldocs.roles`).Scan(&counts.roles); err != nil {
		t.Fatalf("count metaldocs.roles: %v", err)
	}
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM metaldocs.capability_bindings`).Scan(&counts.bindings); err != nil {
		t.Fatalf("count metaldocs.capability_bindings: %v", err)
	}
	return counts
}

// execWithSchedulerBypass runs sqlText against a single pinned *sql.Conn with
// metaldocs.bypass_authz='scheduler' set for its duration, then RESETs the
// GUC on that SAME connection before returning it to db's pool (B2, PR #113
// bot review). SET (not SET LOCAL) is session-scoped: leaving it set on a
// connection handed back to the pool would give a later caller on this same
// *sql.DB an unintended capability-tripwire bypass. SET LOCAL is not usable
// here because this runs outside an explicit Go-managed transaction --
// sqlText carries its own top-level BEGIN/COMMIT.
func execWithSchedulerBypass(ctx context.Context, t *testing.T, db *sql.DB, sqlText string) error {
	t.Helper()
	conn, err := db.Conn(ctx)
	if err != nil {
		return fmt.Errorf("acquire pinned connection: %w", err)
	}
	defer func() {
		if _, resetErr := conn.ExecContext(ctx, "RESET metaldocs.bypass_authz"); resetErr != nil {
			t.Errorf("reset scheduler bypass on pinned connection: %v", resetErr)
		}
		if closeErr := conn.Close(); closeErr != nil {
			t.Errorf("close pinned connection: %v", closeErr)
		}
	}()

	if _, err := conn.ExecContext(ctx, "SET metaldocs.bypass_authz = 'scheduler'"); err != nil {
		return fmt.Errorf("set scheduler bypass: %w", err)
	}
	if _, err := conn.ExecContext(ctx, sqlText); err != nil {
		return err
	}
	return nil
}

// assertBypassNotInherited proves a connection acquired from db's pool AFTER
// execWithSchedulerBypass returned does not see the scheduler bypass -- the
// B2 regression this guards against. db must have SetMaxOpenConns(1) so this
// deterministically observes the same physical connection
// execWithSchedulerBypass just released, not a fresh one that would trivially
// show no GUC set regardless of whether RESET ran.
func assertBypassNotInherited(t *testing.T, ctx context.Context, db *sql.DB) {
	t.Helper()
	conn, err := db.Conn(ctx)
	if err != nil {
		t.Fatalf("acquire connection to check bypass leak: %v", err)
	}
	defer conn.Close()

	var bypass sql.NullString
	if err := conn.QueryRowContext(ctx, `SELECT current_setting('metaldocs.bypass_authz', true)`).Scan(&bypass); err != nil {
		t.Fatalf("read metaldocs.bypass_authz on the next pooled connection: %v", err)
	}
	if bypass.Valid && bypass.String != "" {
		t.Fatalf("metaldocs.bypass_authz leaked onto a later connection from this pool: %q", bypass.String)
	}
}

// seedReplaySafeSourceRows seeds exactly one row into each of 0318's three
// backfill sources (iam_user_roles, user_process_areas, iam_group_roles) so
// TestMigration0318_ReplaySafe has real, non-empty source data to carry
// across a replay -- see that test's own comment for why an empty-to-empty
// replay would not exercise the guards at all.
func seedReplaySafeSourceRows(t *testing.T, ctx context.Context, db *sql.DB) {
	t.Helper()
	tenantID := tenant.DevTenantID
	userID := seedUser(t, ctx, db, tenantID)
	area := seedArea(t, ctx, db, tenantID)

	// Source 1: iam_user_roles (arm #3, user.manage).
	func() {
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			t.Fatalf("seed iam_user_roles: begin tx: %v", err)
		}
		defer func() { _ = tx.Rollback() }()

		testdb.SetCapsOnTx(t, tx, `[{"cap":"user.manage"}]`)
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO metaldocs.iam_user_roles (user_id, role_code, tenant_id, assigned_by)
			 VALUES ($1, 'viewer', $2::uuid, $1)`,
			userID, tenantID); err != nil {
			t.Fatalf("seed iam_user_roles: %v", err)
		}
		if err := tx.Commit(); err != nil {
			t.Fatalf("commit seed iam_user_roles: %v", err)
		}
	}()

	// Source 2: public.user_process_areas (arm #4, membership.manage).
	func() {
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			t.Fatalf("seed user_process_areas: begin tx: %v", err)
		}
		defer func() { _ = tx.Rollback() }()

		testdb.SetCapsOnTx(t, tx, `[{"cap":"membership.manage"}]`)
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO public.user_process_areas
			    (user_id, tenant_id, area_code, role, effective_from, effective_to, granted_by, revoked_by)
			 VALUES ($1, $2::uuid, $3, 'author', now(), NULL, $1, NULL)`,
			userID, tenantID, area); err != nil {
			t.Fatalf("seed user_process_areas: %v", err)
		}
		if err := tx.Commit(); err != nil {
			t.Fatalf("commit seed user_process_areas: %v", err)
		}
	}()

	// Source 3: metaldocs.iam_group_roles (arm #18, user.manage) — needs a
	// group first (arm #16, also user.manage).
	groupID := uuid.NewString()
	func() {
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			t.Fatalf("seed iam_groups: begin tx: %v", err)
		}
		defer func() { _ = tx.Rollback() }()

		testdb.SetCapsOnTx(t, tx, `[{"cap":"user.manage"}]`)
		if _, err := tx.ExecContext(ctx,
			`SELECT set_config('metaldocs.tenant_id', $1, true)`, tenantID); err != nil {
			t.Fatalf("seed iam_groups: set tenant GUC: %v", err)
		}
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO metaldocs.iam_groups (id, tenant_id, name) VALUES ($1::uuid, $2::uuid, $3)`,
			groupID, tenantID, "fixture-group-"+groupID[:8]); err != nil {
			t.Fatalf("seed iam_groups: %v", err)
		}
		if err := tx.Commit(); err != nil {
			t.Fatalf("commit seed iam_groups: %v", err)
		}
	}()
	func() {
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			t.Fatalf("seed iam_group_roles: begin tx: %v", err)
		}
		defer func() { _ = tx.Rollback() }()

		testdb.SetCapsOnTx(t, tx, `[{"cap":"user.manage"}]`)
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO metaldocs.iam_group_roles (group_id, role) VALUES ($1::uuid, 'viewer')`,
			groupID); err != nil {
			t.Fatalf("seed iam_group_roles: %v", err)
		}
		if err := tx.Commit(); err != nil {
			t.Fatalf("commit seed iam_group_roles: %v", err)
		}
	}()
}

// migration0318Path resolves db/migrations/0318_capability_bindings_schema_backfill.sql
// by walking up from this test file to the repo root (identified by go.mod),
// mirroring the same walk testdb.ApplyCuratedBootstrap uses internally.
func migration0318Path(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed to resolve this test file's own path")
	}
	dir := filepath.Dir(file)
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return filepath.Join(dir, "db", "migrations", "0318_capability_bindings_schema_backfill.sql")
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("could not find repo root walking up from migration_0318_test.go")
		}
		dir = parent
	}
}

// TestMigration0318_BackfillPreservesHistory proves the backfill is a
// straight row-count carryover from each source relation (provenance
// preserved, nothing flattened).
//
// CodeRabbit (PR #113 review round) caught this test vacuous as originally
// written: testdb.OpenFreshDatabase's bootstrap seeds reference-data only
// (db/reference-data, never db/dev-seeds — see ApplyCuratedBootstrap), and
// reference-data inserts zero rows into iam_user_roles, user_process_areas,
// or iam_group_roles. Every comparison below would have read srcCount=0,
// boundCount=0 regardless of whether the three backfill INSERT...SELECT
// statements in 0318 existed at all -- deleting the whole backfill would
// still pass. TestMigration0318_ReplaySafe (above) already guards this exact
// shape with an explicit non-zero precondition; this test had the identical
// hole one function down.
//
// Fix: seed, don't just guard. A bare "fail if the count is zero"
// precondition (CodeRabbit's proposed floor) would make the test *honest*
// but still non-deterministic -- pass or fail would depend on whatever
// db/reference-data happens to seed into these three tables today, which
// this test does not own and should not need to track. Seeding rows this
// test knows the shape of, using the SAME canonical fixture the sibling
// replay-safety test already established (seedReplaySafeSourceRows,
// testdb.SetCapsOnTx), makes the carryover proof self-contained and
// deterministic instead. testdb.OpenFreshDatabase's own bootstrap already
// ran 0318 once over the still-empty sources (a real no-op backfill), so
// after seeding, this test must replay 0318's backfill INSERTs itself --
// exactly the "seed, then run the file for real" first pass
// TestMigration0318_ReplaySafe performs -- before the source-vs-backfilled
// comparison means anything.
//
// Proven RED (2026-08-11): with the three backfill INSERT...SELECT
// statements deleted from a scratch copy of 0318, this test failed with
// "source_relation=iam_user_roles: capability_bindings has 0 rows, source
// table has 1 rows" (and the same for the other two sources) -- proof the
// comparison actually exercises the backfill, not 0==0.
func TestMigration0318_BackfillPreservesHistory(t *testing.T) {
	ctx := context.Background()
	db, _ := testdb.OpenFreshDatabase(t)

	// Seed one real row into each of 0318's three backfill sources (the same
	// canonical fixture TestMigration0318_ReplaySafe uses), then replay
	// 0318's own backfill INSERTs so those rows actually reach
	// capability_bindings -- bootstrap's earlier run of 0318 saw only empty
	// sources and backfilled nothing, so this is the first genuine backfill
	// pass, not a replay.
	seedReplaySafeSourceRows(t, ctx, db)
	sqlBytes, err := os.ReadFile(migration0318Path(t))
	if err != nil {
		t.Fatalf("read migration 0318 file: %v", err)
	}
	if err := execWithSchedulerBypass(ctx, t, db, string(sqlBytes)); err != nil {
		t.Fatalf("running migration 0318 against freshly-seeded source rows failed: %v", err)
	}

	for _, tc := range []struct {
		source string
		countQ string
	}{
		{"iam_user_roles", `SELECT count(*) FROM metaldocs.iam_user_roles`},
		{"user_process_areas", `SELECT count(*) FROM public.user_process_areas`},
		{"iam_group_roles", `SELECT count(*) FROM metaldocs.iam_group_roles`},
	} {
		var srcCount, boundCount int
		if err := db.QueryRowContext(ctx, tc.countQ).Scan(&srcCount); err != nil {
			t.Fatalf("count %s: %v", tc.source, err)
		}
		if srcCount == 0 {
			t.Fatalf("precondition failed: source %s has zero rows after seeding — comparison would prove nothing", tc.source)
		}
		if err := db.QueryRowContext(ctx,
			`SELECT count(*) FROM metaldocs.capability_bindings WHERE source_relation = $1`,
			tc.source).Scan(&boundCount); err != nil {
			t.Fatalf("count capability_bindings backfilled from %s: %v", tc.source, err)
		}
		if boundCount != srcCount {
			t.Fatalf("source_relation=%s: capability_bindings has %d rows, source table has %d — backfill lost or fabricated rows", tc.source, boundCount, srcCount)
		}
	}
}

// TestMigration0318_RejectsDanglingSubject proves a capability_bindings row
// cannot name a subject that does not exist — RED: before this migration's
// FK, a dangling subject_user_id/subject_group_id was representable and
// silently unenforced by anything but application code.
func TestMigration0318_RejectsDanglingSubject(t *testing.T) {
	ctx := context.Background()
	db, _ := testdb.OpenFreshDatabase(t)

	t.Run("dangling subject_user_id", func(t *testing.T) {
		err := insertCapabilityBinding(t, ctx, db, capabilityBindingRow{
			TenantID:      tenant.DevTenantID,
			SubjectKind:   "user",
			SubjectUserID: sql.NullString{String: "no-such-user-" + uuid.NewString(), Valid: true},
			RoleCode:      "viewer",
			ScopeKind:     "tenant",
		})
		if err == nil {
			t.Fatal("insert with a subject_user_id naming a nonexistent iam_users row succeeded — FK is not enforced")
		}
		if !hasSQLState(err, "23503") {
			t.Fatalf("expected SQLSTATE 23503 (foreign_key_violation), got: %v", err)
		}
	})

	t.Run("dangling subject_group_id", func(t *testing.T) {
		err := insertCapabilityBinding(t, ctx, db, capabilityBindingRow{
			TenantID:       tenant.DevTenantID,
			SubjectKind:    "group",
			SubjectGroupID: uuid.NullUUID{UUID: uuid.New(), Valid: true},
			RoleCode:       "viewer",
			ScopeKind:      "tenant",
		})
		if err == nil {
			t.Fatal("insert with a subject_group_id naming a nonexistent iam_groups row succeeded — FK is not enforced")
		}
		if !hasSQLState(err, "23503") {
			t.Fatalf("expected SQLSTATE 23503 (foreign_key_violation), got: %v", err)
		}
	})
}

// TestMigration0318_RejectsCrossTenantScope proves a binding cannot name an
// area (scope_ref) that belongs to a DIFFERENT tenant than the binding's own
// tenant_id — the composite FK (tenant_id, scope_ref) -> document_process_areas
// makes a cross-tenant scope reference structurally impossible, matching the
// "cross-tenant URL -> 404, not a leaked row" invariant one level down at the
// schema line.
func TestMigration0318_RejectsCrossTenantScope(t *testing.T) {
	ctx := context.Background()
	db, _ := testdb.OpenFreshDatabase(t)

	otherTenant := seedTenant(t, ctx, db)
	otherTenantUser := seedUser(t, ctx, db, otherTenant)
	areaInOtherTenant := seedArea(t, ctx, db, otherTenant)

	// A valid user in DevTenantID's own tenant, so the failure is attributable
	// to the scope FK, not the subject FK.
	devTenantUser := seedUser(t, ctx, db, tenant.DevTenantID)

	err := insertCapabilityBinding(t, ctx, db, capabilityBindingRow{
		TenantID:      tenant.DevTenantID,
		SubjectKind:   "user",
		SubjectUserID: sql.NullString{String: devTenantUser, Valid: true},
		RoleCode:      "signer",
		ScopeKind:     "area",
		ScopeRef:      sql.NullString{String: areaInOtherTenant, Valid: true},
	})
	if err == nil {
		t.Fatal("insert binding DevTenantID to an area owned by a different tenant succeeded — composite (tenant_id, scope_ref) FK is not enforced")
	}
	if !hasSQLState(err, "23503") {
		t.Fatalf("expected SQLSTATE 23503 (foreign_key_violation), got: %v", err)
	}

	_ = otherTenantUser // seeded for symmetry / future extension; unused directly here
}

// TestMigration0318_RejectsDuplicateActiveBinding proves at most one ACTIVE
// (effective_to IS NULL) binding can exist per (tenant, subject, role,
// scope) — the partial unique index ux_capability_bindings_active_identity.
// A second, already-revoked binding for the same identity must NOT collide
// (history is preserved, not deduplicated away).
func TestMigration0318_RejectsDuplicateActiveBinding(t *testing.T) {
	ctx := context.Background()
	db, _ := testdb.OpenFreshDatabase(t)

	userID := seedUser(t, ctx, db, tenant.DevTenantID)

	row := capabilityBindingRow{
		TenantID:      tenant.DevTenantID,
		SubjectKind:   "user",
		SubjectUserID: sql.NullString{String: userID, Valid: true},
		RoleCode:      "editor",
		ScopeKind:     "tenant",
	}

	if err := insertCapabilityBinding(t, ctx, db, row); err != nil {
		t.Fatalf("first (active) binding insert failed unexpectedly: %v", err)
	}

	err := insertCapabilityBinding(t, ctx, db, row)
	if err == nil {
		t.Fatal("second ACTIVE binding with the identical (tenant,subject,role,scope) identity succeeded — duplicate active grant is representable")
	}
	if !hasSQLState(err, "23505") {
		t.Fatalf("expected SQLSTATE 23505 (unique_violation) for duplicate active identity, got: %v", err)
	}

	// A REVOKED binding for the exact same identity must NOT collide — history
	// is preserved, not deduplicated. Revoke the first row, then insert a
	// second active one for the same identity: this must succeed. The UPDATE
	// itself is tripwire-guarded (arm #21, same as INSERT), so it needs its
	// own capability assertion and tenant GUC, not just a bare ExecContext.
	func() {
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			t.Fatalf("revoke first binding: begin tx: %v", err)
		}
		defer func() { _ = tx.Rollback() }()

		testdb.SetCapsOnTx(t, tx, `[{"cap":"user.manage"}]`)
		if _, err := tx.ExecContext(ctx,
			`SELECT set_config('metaldocs.tenant_id', $1, true)`, tenant.DevTenantID,
		); err != nil {
			t.Fatalf("revoke first binding: set tenant GUC: %v", err)
		}
		if _, err := tx.ExecContext(ctx, `
			UPDATE metaldocs.capability_bindings
			   SET effective_to = now(), revoked_by = 'test-harness'
			 WHERE tenant_id = $1 AND subject_user_id = $2 AND role_code = $3
			   AND scope_kind = 'tenant' AND effective_to IS NULL`,
			tenant.DevTenantID, userID, row.RoleCode); err != nil {
			t.Fatalf("revoke first binding: update: %v", err)
		}
		if err := tx.Commit(); err != nil {
			t.Fatalf("revoke first binding: commit: %v", err)
		}
	}()

	if err := insertCapabilityBinding(t, ctx, db, row); err != nil {
		t.Fatalf("re-granting the same identity after revocation should succeed (history, not dedup), got: %v", err)
	}

	var n int
	if err := db.QueryRowContext(ctx, `
		SELECT count(*) FROM metaldocs.capability_bindings
		 WHERE tenant_id = $1 AND subject_user_id = $2 AND role_code = $3 AND scope_kind = 'tenant'`,
		tenant.DevTenantID, userID, row.RoleCode).Scan(&n); err != nil {
		t.Fatalf("count history rows: %v", err)
	}
	if n != 2 {
		t.Fatalf("expected 2 history rows (1 revoked + 1 active) for the identity, got %d — history was flattened", n)
	}
}

// TestMigration0318_RejectsSubjectShapeMismatch and
// TestMigration0318_RejectsScopeShapeMismatch prove the discriminated-union
// CHECKs: a 'user' subject cannot also carry a group id (or vice versa), and
// a 'tenant' scope cannot carry an area ref (or vice versa).
func TestMigration0318_RejectsSubjectShapeMismatch(t *testing.T) {
	ctx := context.Background()
	db, _ := testdb.OpenFreshDatabase(t)

	userID := seedUser(t, ctx, db, tenant.DevTenantID)

	err := insertCapabilityBinding(t, ctx, db, capabilityBindingRow{
		TenantID:       tenant.DevTenantID,
		SubjectKind:    "user",
		SubjectUserID:  sql.NullString{String: userID, Valid: true},
		SubjectGroupID: uuid.NullUUID{UUID: uuid.New(), Valid: true}, // both set — illegal
		RoleCode:       "viewer",
		ScopeKind:      "tenant",
	})
	if err == nil {
		t.Fatal("insert with both subject_user_id and subject_group_id set succeeded — subject-shape CHECK is not enforced")
	}
	if !hasSQLState(err, "23514") {
		t.Fatalf("expected SQLSTATE 23514 (check_violation), got: %v", err)
	}
}

func TestMigration0318_RejectsScopeShapeMismatch(t *testing.T) {
	ctx := context.Background()
	db, _ := testdb.OpenFreshDatabase(t)

	userID := seedUser(t, ctx, db, tenant.DevTenantID)
	area := seedArea(t, ctx, db, tenant.DevTenantID)

	err := insertCapabilityBinding(t, ctx, db, capabilityBindingRow{
		TenantID:      tenant.DevTenantID,
		SubjectKind:   "user",
		SubjectUserID: sql.NullString{String: userID, Valid: true},
		RoleCode:      "viewer",
		ScopeKind:     "tenant",
		ScopeRef:      sql.NullString{String: area, Valid: true}, // scope_kind='tenant' but scope_ref set — illegal
	})
	if err == nil {
		t.Fatal("insert with scope_kind='tenant' and a non-NULL scope_ref succeeded — scope-shape CHECK is not enforced")
	}
	if !hasSQLState(err, "23514") {
		t.Fatalf("expected SQLSTATE 23514 (check_violation), got: %v", err)
	}
}

// ── fixtures ────────────────────────────────────────────────────────────

type capabilityBindingRow struct {
	TenantID       string
	SubjectKind    string
	SubjectUserID  sql.NullString
	SubjectGroupID uuid.NullUUID
	RoleCode       string
	ScopeKind      string
	ScopeRef       sql.NullString
}

// insertCapabilityBinding attempts the write with the same capability
// (user.manage — arm #21, match-one with membership.manage) and tenant GUC
// context the application itself would hold, so a failure is attributable to
// the constraint under test, not to the tripwire or RLS.
func insertCapabilityBinding(t *testing.T, ctx context.Context, db *sql.DB, row capabilityBindingRow) error {
	t.Helper()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	testdb.SetCapsOnTx(t, tx, `[{"cap":"user.manage"}]`)
	if _, err := tx.ExecContext(ctx, `SELECT set_config('metaldocs.tenant_id', $1, true)`, row.TenantID); err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO metaldocs.capability_bindings
			(tenant_id, subject_kind, subject_user_id, subject_group_id, role_code, scope_kind, scope_ref)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		row.TenantID, row.SubjectKind, row.SubjectUserID, row.SubjectGroupID, row.RoleCode, row.ScopeKind, row.ScopeRef,
	)
	if err != nil {
		return err
	}
	return tx.Commit()
}

// seedTenant inserts a fresh tenant row (asserting tenant.onboard, the arm
// this table's trigger requires) and returns its id.
func seedTenant(t *testing.T, ctx context.Context, db *sql.DB) string {
	t.Helper()
	id := uuid.NewString()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	defer func() { _ = tx.Rollback() }()

	testdb.SetCapsOnTx(t, tx, `[{"cap":"tenant.onboard"}]`)
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO metaldocs.tenants (id, name, slug) VALUES ($1::uuid, $2, $2)`,
		id, "fixture-"+id[:8]); err != nil {
		t.Fatalf("seed tenant: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("commit seed tenant: %v", err)
	}
	return id
}

// seedUser inserts a fresh iam_users row in the given tenant (asserting
// user.manage, arm #15) and returns its user_id.
func seedUser(t *testing.T, ctx context.Context, db *sql.DB, tenantID string) string {
	t.Helper()
	userID := "fixture-user-" + uuid.NewString()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	defer func() { _ = tx.Rollback() }()

	testdb.SetCapsOnTx(t, tx, `[{"cap":"user.manage"}]`)
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO metaldocs.iam_users (user_id, display_name, tenant_id) VALUES ($1, $1, $2::uuid)`,
		userID, tenantID); err != nil {
		t.Fatalf("seed user: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("commit seed user: %v", err)
	}
	return userID
}

// seedArea inserts a fresh document_process_areas row in the given tenant
// (asserting taxonomy.manage, arm #11) and returns its code.
func seedArea(t *testing.T, ctx context.Context, db *sql.DB, tenantID string) string {
	t.Helper()
	code := "fx" + uuid.NewString()[:8]
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	defer func() { _ = tx.Rollback() }()

	testdb.SetCapsOnTx(t, tx, `[{"cap":"taxonomy.manage"}]`)
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO metaldocs.document_process_areas (code, name, tenant_id) VALUES ($1, $1, $2::uuid)`,
		code, tenantID); err != nil {
		t.Fatalf("seed area: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("commit seed area: %v", err)
	}
	return code
}
