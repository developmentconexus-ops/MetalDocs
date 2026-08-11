//go:build integration

// Integration proof ladder for issue #88 / A6.1 re-cut, PR #110 review
// findings C1-C3. Every test here drives the REAL run() entrypoint (main.go)
// against a database this file owns end-to-end, never the ambient leased test
// database tests/integration/testdb.Open hands out to ~130 other callers --
// provisioning mutates cluster-global roles and per-database ownership/grants,
// which must never leak onto a shared fixture.
package main

import (
	"context"
	"crypto/rand"
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"

	"metaldocs/tests/integration/testdb"
)

// bareTemplateName is the fixed name of the one bare-plus-baseline-schema
// (no roles/grants/migrations) database this package builds ONCE per test
// process, guarded by bareTemplateOnce below. Applying
// db/baseline/0001_current_schema.sql (5000+ lines of DDL) measures multiple
// minutes on this virtualized-storage/shared-cluster host -- confirmed
// directly via `psql -f` during this test's own development, not assumed --
// so every test that needs a bare database CLONES this one via
// CREATE DATABASE ... TEMPLATE ... STRATEGY = WAL_LOG (a physical file copy,
// measured elsewhere in this repo at low single-digit seconds; see
// tests/integration/testdb.OpenFreshDatabase's comment) instead of re-running
// the slow DDL apply per test.
const bareTemplateName = "metaldocs_provtest_bare_template"

var (
	bareTemplateOnce sync.Once
	bareTemplateErr  error
)

// ensureBareTemplate builds bareTemplateName once (idempotent across the
// whole `go test` process -- sync.Once, not per-test), then returns nothing;
// callers clone FROM it by name. Not content-addressed/cached across runs
// like testdb's own template (unnecessary here: this package's tests are the
// only consumer, and a stale bare template would just get dropped and
// rebuilt the next time this file's baseline SQL changes cause a schema
// mismatch inside a test -- acceptable for a provisioning-binary test suite
// that already budgets minutes for its slow path once per run).
func ensureBareTemplate(t *testing.T) {
	t.Helper()
	bareTemplateOnce.Do(func() {
		bareTemplateErr = buildBareTemplate()
	})
	if bareTemplateErr != nil {
		t.Fatalf("build bare provisioning template: %v", bareTemplateErr)
	}
}

func buildBareTemplate() error {
	dsn := os.Getenv("METALDOCS_DATABASE_URL")
	if dsn == "" {
		dsn = os.Getenv("DATABASE_URL")
	}
	if dsn == "" {
		return fmt.Errorf("DATABASE_URL/METALDOCS_DATABASE_URL not set")
	}
	cfg, err := pgx.ParseConfig(dsn)
	if err != nil {
		return fmt.Errorf("parse base DSN: %w", err)
	}
	cfg.Database = "postgres"
	admin := stdlib.OpenDB(*cfg)
	defer admin.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	pingErr := admin.PingContext(ctx)
	cancel()
	if pingErr != nil {
		return fmt.Errorf("ping admin db: %w", pingErr)
	}

	// Drop any stale template left by a previous killed/crashed run before
	// rebuilding -- CREATE DATABASE below is not IF NOT EXISTS.
	dropCtx, dropCancel := context.WithTimeout(context.Background(), 60*time.Second)
	_, _ = admin.ExecContext(dropCtx, fmt.Sprintf(
		"SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = %s AND pid <> pg_backend_pid()",
		quoteLiteralLocal(bareTemplateName)))
	_, _ = admin.ExecContext(dropCtx, fmt.Sprintf("DROP DATABASE IF EXISTS %s WITH (FORCE)", quoteIdent(bareTemplateName)))
	dropCancel()

	createCtx, createCancel := context.WithTimeout(context.Background(), 5*time.Minute)
	_, err = admin.ExecContext(createCtx, fmt.Sprintf("CREATE DATABASE %s STRATEGY = WAL_LOG", quoteIdent(bareTemplateName)))
	createCancel()
	if err != nil {
		return fmt.Errorf("create bare template database: %w", err)
	}

	baseCfg := *cfg
	baseCfg.Database = bareTemplateName
	templateDB := stdlib.OpenDB(baseCfg)
	// Must be closed before any clone reads this template, else
	// CREATE DATABASE ... TEMPLATE fails with "source database is being
	// accessed by other users" (same constraint testdb.buildTemplate documents).
	defer func() { _ = templateDB.Close() }()

	pingCtx, pingCancel := context.WithTimeout(context.Background(), 60*time.Second)
	pingErr = templateDB.PingContext(pingCtx)
	pingCancel()
	if pingErr != nil {
		return fmt.Errorf("ping bare template database: %w", pingErr)
	}

	// Real init order (deploy/compose/docker-compose.yml's
	// docker-entrypoint-initdb.d mounts, 01_extensions.sql before
	// 02_baseline.sql): prerequisites (CREATE EXTENSION pgcrypto/pg_trgm)
	// MUST land before the baseline schema, whose indexes reference
	// extension-provided operator classes (e.g. gin_trgm_ops) at CREATE INDEX
	// time, not merely at query time -- applying baseline alone against a
	// bare database fails partway through with "operator class ... does not
	// exist", confirmed empirically while developing this test.
	root := repoRoot()
	for _, rel := range []string{
		filepath.Join("db", "prerequisites", "0001_extensions.sql"),
		filepath.Join("db", "baseline", "0001_current_schema.sql"),
	} {
		path := filepath.Join(root, rel)
		sqlBytes, readErr := os.ReadFile(path)
		if readErr != nil {
			return fmt.Errorf("read sql bundle %s: %w", path, readErr)
		}
		applyCtx, applyCancel := context.WithTimeout(context.Background(), 10*time.Minute)
		_, execErr := templateDB.ExecContext(applyCtx, string(sqlBytes))
		applyCancel()
		if execErr != nil {
			return fmt.Errorf("apply sql bundle %s to template: %w", rel, execErr)
		}
	}
	return nil
}

// repoRoot walks up from this file to the directory containing go.mod. A
// small local duplicate of tests/integration/testdb's unexported repoRoot --
// not worth exporting cross-package for one caller.
func repoRoot() string {
	_, file, _, _ := runtime.Caller(0)
	dir := filepath.Dir(file)
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			panic("could not find repo root")
		}
		dir = parent
	}
}

func quoteIdent(v string) string {
	return `"` + strings.ReplaceAll(v, `"`, `""`) + `"`
}

// adminDBOnMaintenance opens a *sql.DB against the ambient test DSN's
// identity, but pointed at the "postgres" maintenance database, for
// CREATE DATABASE / DROP DATABASE -- exactly the pattern
// tests/integration/testdb.OpenFreshDatabase uses for the same reason.
func adminDBOnMaintenance(t *testing.T) *sql.DB {
	t.Helper()
	cfg, err := pgx.ParseConfig(testdb.DSN(t))
	if err != nil {
		t.Fatalf("parse base DSN: %v", err)
	}
	cfg.Database = "postgres"
	db := stdlib.OpenDB(*cfg)
	t.Cleanup(func() { _ = db.Close() })

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		t.Skipf("integration DB unreachable: %v", err)
	}
	return db
}

// bareProvisionableDatabase creates a genuinely bare database -- NOT cloned
// from testdb's curated template (which already carries db/grants' roles and
// per-database privilege posture baked in, per curatedBundlePaths -- that
// makes it representative of an ALREADY-PROVISIONED volume, not a clean
// one) -- carries only prerequisites + the baseline schema (the same two
// files docker-entrypoint-initdb.d applies, in the same order, before
// db-provision ever runs in compose; see the runbook's "fails on a fresh
// volume" note for why this step is required first), and returns a DSN
// pointed at it. Cloned via WAL_LOG from the once-built bareTemplateName
// (see ensureBareTemplate) rather than re-running the ~5000-line baseline
// DDL apply per test. Dropped via t.Cleanup.
func bareProvisionableDatabase(t *testing.T) string {
	t.Helper()
	ensureBareTemplate(t)
	admin := adminDBOnMaintenance(t)

	dbName := "metaldocs_provtest_" + randomSuffix(t)

	// STRATEGY = WAL_LOG, not the bare server default: same reasoning as
	// tests/integration/testdb.OpenFreshDatabase's own CREATE DATABASE (see
	// that function's comment) -- a physical WAL_LOG clone from an existing
	// template measures low single-digit seconds on this host.
	createCtx, createCancel := context.WithTimeout(context.Background(), 60*time.Second)
	_, err := admin.ExecContext(createCtx, fmt.Sprintf(
		"CREATE DATABASE %s TEMPLATE %s STRATEGY = WAL_LOG", quoteIdent(dbName), quoteIdent(bareTemplateName)))
	createCancel()
	if err != nil {
		t.Fatalf("clone bare database %s from template %s: %v", dbName, bareTemplateName, err)
	}

	t.Cleanup(func() {
		dropCtx, dropCancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer dropCancel()
		_, _ = admin.ExecContext(dropCtx, fmt.Sprintf(
			"SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = %s AND pid <> pg_backend_pid()",
			quoteLiteralLocal(dbName)))
		if _, err := admin.ExecContext(dropCtx, fmt.Sprintf("DROP DATABASE IF EXISTS %s WITH (FORCE)", quoteIdent(dbName))); err != nil {
			t.Logf("cleanup: drop bare database %s: %v", dbName, err)
		}
	})

	baseCfg, err := pgx.ParseConfig(testdb.DSN(t))
	if err != nil {
		t.Fatalf("parse base DSN: %v", err)
	}
	baseCfg.Database = dbName
	dsn := stdlibDSNString(baseCfg)

	newDB := stdlib.OpenDB(*baseCfg)
	defer func() { _ = newDB.Close() }()

	pingCtx, pingCancel := context.WithTimeout(context.Background(), 60*time.Second)
	pingErr := newDB.PingContext(pingCtx)
	pingCancel()
	if pingErr != nil {
		t.Fatalf("ping bare database %s: %v", dbName, pingErr)
	}

	return dsn
}

func quoteLiteralLocal(v string) string {
	return "'" + strings.ReplaceAll(v, "'", "''") + "'"
}

// stdlibDSNString renders cfg back into a postgres:// URL run()'s own
// config.LoadPostgresConfig can parse via METALDOCS_DATABASE_URL.
func stdlibDSNString(cfg *pgx.ConnConfig) string {
	u := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(cfg.User, cfg.Password),
		Host:   fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Path:   cfg.Database,
	}
	q := u.Query()
	q.Set("sslmode", "disable")
	u.RawQuery = q.Encode()
	return u.String()
}

func randomSuffix(t *testing.T) string {
	t.Helper()
	b := make([]byte, 6)
	if _, err := rand.Read(b); err != nil {
		t.Fatalf("rand.Read: %v", err)
	}
	return fmt.Sprintf("%x", b)
}

// runProvisionAgainst points run()'s env-sourced config at dsn (the exact
// same env vars config.LoadPostgresConfig/LoadMigrationConfig/
// LoadRuntimeIdentityConfig/LoadJobsConfig read at real boot) and invokes it
// directly -- this is the actual production entrypoint, not a reimplementation
// of it, so a passing test proves run() itself, not a stand-in.
func runProvisionAgainst(t *testing.T, dsn string, extraEnv map[string]string) error {
	t.Helper()
	root := repoRoot()
	t.Setenv("METALDOCS_DATABASE_URL", dsn)
	t.Setenv("METALDOCS_SKIP_STARTUP_MIGRATIONS", "false")
	t.Setenv("METALDOCS_PREREQUISITES_DIR", filepath.Join(root, "db", "prerequisites"))
	t.Setenv("METALDOCS_GRANTS_DIR", filepath.Join(root, "db", "grants"))
	t.Setenv("METALDOCS_MIGRATIONS_DIR", filepath.Join(root, "db", "migrations"))
	t.Setenv("METALDOCS_RUNTIME_DB_PASSWORD", "provtest_runtime_pw_"+randomSuffix(t))
	t.Setenv("METALDOCS_JOBS_RIVER_SCHEMA", "")
	for k, v := range extraEnv {
		t.Setenv(k, v)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	return run(ctx)
}

// openAsDBUser opens dsn as-is (its own embedded database name). database
// overrides that database when non-empty.
func openAsDBUser(dsn, database string) (*sql.DB, error) {
	cfg, err := pgx.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}
	if database != "" {
		cfg.Database = database
	}
	return stdlib.OpenDB(*cfg), nil
}

// schemaOwner returns the rolname owning schema, queried against the
// provisioned database itself (as the bootstrap-superuser identity, which
// can read every catalog).
func schemaOwner(t *testing.T, db *sql.DB, schema string) string {
	t.Helper()
	var owner string
	err := db.QueryRowContext(context.Background(),
		`SELECT r.rolname FROM pg_namespace n JOIN pg_roles r ON r.oid = n.nspowner WHERE n.nspname = $1`,
		schema).Scan(&owner)
	if err != nil {
		t.Fatalf("query schema owner for %q: %v", schema, err)
	}
	return owner
}

// nonOwnerRelations returns every table/sequence in schema NOT owned by want
// -- the catalog-inspection proof named in main.go's own doc comment
// (pg_class.relowner), proving DDL never fell back to the bootstrap
// superuser across a pool reconnect.
func nonOwnerRelations(t *testing.T, db *sql.DB, schema, want string) []string {
	t.Helper()
	rows, err := db.QueryContext(context.Background(),
		`SELECT c.relname, r.rolname
		   FROM pg_class c
		   JOIN pg_namespace n ON n.oid = c.relnamespace
		   JOIN pg_roles r ON r.oid = c.relowner
		  WHERE n.nspname = $1 AND c.relkind IN ('r', 'S')`,
		schema)
	if err != nil {
		t.Fatalf("query relation owners in %q: %v", schema, err)
	}
	defer rows.Close()

	var bad []string
	for rows.Next() {
		var relname, owner string
		if err := rows.Scan(&relname, &owner); err != nil {
			t.Fatalf("scan relation owner: %v", err)
		}
		if owner != want {
			bad = append(bad, fmt.Sprintf("%s.%s owned by %s, want %s", schema, relname, owner, want))
		}
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate relation owners in %q: %v", schema, err)
	}
	return bad
}

// TestProvision_CleanDatabaseEndToEnd is the C1/C2/C3 baseline proof: run()
// itself, driven end-to-end against a bare (schema-only, no roles/grants/
// migrations applied) database -- the fresh-volume case the runbook
// describes -- exits nil and leaves metaldocs_owner/metaldocs_runtime safe.
func TestProvision_CleanDatabaseEndToEnd(t *testing.T) {
	dsn := bareProvisionableDatabase(t)

	if err := runProvisionAgainst(t, dsn, nil); err != nil {
		t.Fatalf("run() against clean database: %v", err)
	}

	db, err := openAsDBUser(dsn, "")
	if err != nil {
		t.Fatalf("open provisioned database: %v", err)
	}
	defer db.Close()

	var superuser, bypassrls, canLogin bool
	if err := db.QueryRowContext(context.Background(),
		`SELECT rolsuper, rolbypassrls, rolcanlogin FROM pg_roles WHERE rolname = 'metaldocs_owner'`).
		Scan(&superuser, &bypassrls, &canLogin); err != nil {
		t.Fatalf("query metaldocs_owner attributes: %v", err)
	}
	if superuser || bypassrls || canLogin {
		t.Fatalf("metaldocs_owner unsafe after provisioning: superuser=%t bypassrls=%t canlogin=%t", superuser, bypassrls, canLogin)
	}

	if err := db.QueryRowContext(context.Background(),
		`SELECT rolsuper, rolbypassrls, rolcanlogin FROM pg_roles WHERE rolname = 'metaldocs_runtime'`).
		Scan(&superuser, &bypassrls, &canLogin); err != nil {
		t.Fatalf("query metaldocs_runtime attributes: %v", err)
	}
	if superuser || bypassrls || !canLogin {
		t.Fatalf("metaldocs_runtime unsafe after provisioning: superuser=%t bypassrls=%t canlogin=%t (want canlogin=true)", superuser, bypassrls, canLogin)
	}
}

// TestProvision_DDLObjectsAreOwnedByMetaldocsOwner is the catalog-inspection
// proof named in main.go's own package doc comment: after run(), every table
// and sequence in the metaldocs schema is owned by metaldocs_owner -- never
// the bootstrap superuser (which would mean OpenAsRole's SET ROLE pin
// silently didn't take on some connection) and never metaldocs_runtime
// (which would be a silent RLS bypass, since RLS does not apply to an
// owner unless FORCE ROW LEVEL SECURITY is set).
func TestProvision_DDLObjectsAreOwnedByMetaldocsOwner(t *testing.T) {
	dsn := bareProvisionableDatabase(t)
	if err := runProvisionAgainst(t, dsn, nil); err != nil {
		t.Fatalf("run() against clean database: %v", err)
	}

	db, err := openAsDBUser(dsn, "")
	if err != nil {
		t.Fatalf("open provisioned database: %v", err)
	}
	defer db.Close()

	if owner := schemaOwner(t, db, "metaldocs"); owner != "metaldocs_owner" {
		t.Errorf("schema metaldocs owner = %q, want metaldocs_owner", owner)
	}
	if owner := schemaOwner(t, db, "public"); owner != "metaldocs_owner" {
		t.Errorf("schema public owner = %q, want metaldocs_owner", owner)
	}
	if bad := nonOwnerRelations(t, db, "metaldocs", "metaldocs_owner"); len(bad) > 0 {
		t.Errorf("relations not owned by metaldocs_owner:\n%v", bad)
	}
}

// TestProvision_ExistingVolumeReplayIsIdempotent proves the second, third,
// ... boot of the stack (every boot, per the runbook -- not just the first)
// re-running db-provision against an already-provisioned database is a safe
// no-op, never a failure and never a second, conflicting set of grants.
func TestProvision_ExistingVolumeReplayIsIdempotent(t *testing.T) {
	dsn := bareProvisionableDatabase(t)

	if err := runProvisionAgainst(t, dsn, nil); err != nil {
		t.Fatalf("first run(): %v", err)
	}
	if err := runProvisionAgainst(t, dsn, nil); err != nil {
		t.Fatalf("replay run() against already-provisioned database: %v", err)
	}

	db, err := openAsDBUser(dsn, "")
	if err != nil {
		t.Fatalf("open provisioned database: %v", err)
	}
	defer db.Close()
	if bad := nonOwnerRelations(t, db, "metaldocs", "metaldocs_owner"); len(bad) > 0 {
		t.Errorf("relations not owned by metaldocs_owner after replay:\n%v", bad)
	}
}

// TestProvision_CustomRiverSchemaOwnershipAndGrants is the C3 proof:
// METALDOCS_JOBS_RIVER_SCHEMA naming a schema other than public gets created
// (or re-owned, if pre-existing) by the bootstrap superuser, River's migrator
// creates its tables there under metaldocs_owner, and metaldocs_runtime
// receives DML-only access -- mirroring db/grants/0001_role_grants.sql's
// public/metaldocs posture.
func TestProvision_CustomRiverSchemaOwnershipAndGrants(t *testing.T) {
	dsn := bareProvisionableDatabase(t)
	schema := "river_provtest_" + randomSuffix(t)

	if err := runProvisionAgainst(t, dsn, map[string]string{"METALDOCS_JOBS_RIVER_SCHEMA": schema}); err != nil {
		t.Fatalf("run() with custom river schema %q: %v", schema, err)
	}

	db, err := openAsDBUser(dsn, "")
	if err != nil {
		t.Fatalf("open provisioned database: %v", err)
	}
	defer db.Close()

	if owner := schemaOwner(t, db, schema); owner != "metaldocs_owner" {
		t.Fatalf("river schema %q owner = %q, want metaldocs_owner", schema, owner)
	}
	if bad := nonOwnerRelations(t, db, schema, "metaldocs_owner"); len(bad) > 0 {
		t.Errorf("river schema relations not owned by metaldocs_owner:\n%v", bad)
	}

	var hasAccess bool
	if err := db.QueryRowContext(context.Background(),
		`SELECT has_table_privilege('metaldocs_runtime', quote_ident($1) || '.river_job', 'SELECT, INSERT, UPDATE, DELETE')`,
		schema).Scan(&hasAccess); err != nil {
		t.Fatalf("check metaldocs_runtime privilege on %s.river_job: %v", schema, err)
	}
	if !hasAccess {
		t.Errorf("metaldocs_runtime lacks DML access on %s.river_job", schema)
	}
}

// TestProvision_RuntimePasswordRotation_OldFailsNewSucceeds proves
// alterRolePasswordStatement's rendering is correct (old password fails
// authentication, new password succeeds) WITHOUT ever mutating the
// cluster-global metaldocs_runtime role -- see alterRolePasswordStatement's
// own doc comment (main.go) for why: mutating a shared role's password from
// an automated test is worse than the role-attribute hazard
// testdb.CreateThrowawayRole was written to close (C5), since a hung window
// would break every OTHER test package's ability to authenticate as
// metaldocs_runtime at all. A throwaway role proves the exact same
// statement-construction/execution path exactly as well.
func TestProvision_RuntimePasswordRotation_OldFailsNewSucceeds(t *testing.T) {
	adminDB, dbName := testdb.Open(t)
	roleName, oldPassword := testdb.CreateThrowawayRole(t, adminDB)

	// Sanity: old password works before rotation.
	if _, err := probeOpen(dbName, roleName, oldPassword); err != nil {
		t.Fatalf("throwaway role does not authenticate with its own initial password: %v", err)
	}

	newPassword := "rotated_" + randomSuffix(t)
	stmt := alterRolePasswordStatement(roleName, newPassword)
	if _, err := adminDB.ExecContext(context.Background(), stmt); err != nil {
		t.Fatalf("%s: %v", stmt, err)
	}

	if _, err := probeOpen(dbName, roleName, oldPassword); err == nil {
		t.Fatalf("role %s still authenticates with old password after rotation", roleName)
	}
	if _, err := probeOpen(dbName, roleName, newPassword); err != nil {
		t.Fatalf("role %s does not authenticate with new password after rotation: %v", roleName, err)
	}
}

// TestAlterRolePasswordStatement_EscapesEmbeddedQuote is the injection-safety
// unit proof for alterRolePasswordStatement: a password containing a single
// quote must not break out of the Sconst literal. No database needed.
func TestAlterRolePasswordStatement_EscapesEmbeddedQuote(t *testing.T) {
	stmt := alterRolePasswordStatement("metaldocs_runtime", `p'; DROP ROLE metaldocs_owner; --`)
	want := `ALTER ROLE "metaldocs_runtime" PASSWORD 'p''; DROP ROLE metaldocs_owner; --'`
	if stmt != want {
		t.Fatalf("alterRolePasswordStatement() = %q, want %q", stmt, want)
	}
}

// probeOpen attempts to connect to dbName as role/password against the
// ambient test cluster, returning (rather than t.Fatal-ing on) a failed ping
// -- callers need to assert on the FAILURE case too.
func probeOpen(dbName, role, password string) (*sql.DB, error) {
	// t is not available here (needs to stay a plain error-returning helper
	// for the negative-path assertion); DSN discovery duplicates
	// testdb.DSN's env precedence directly since that helper itself requires
	// *testing.T.
	dsn := os.Getenv("METALDOCS_DATABASE_URL")
	if dsn == "" {
		dsn = os.Getenv("DATABASE_URL")
	}
	cfg, err := pgx.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}
	cfg.User = role
	cfg.Password = password
	cfg.Database = dbName

	db := stdlib.OpenDB(*cfg)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	return db, nil
}
