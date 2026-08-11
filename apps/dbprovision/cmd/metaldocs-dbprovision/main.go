// Command metaldocs-dbprovision is the one-shot database provisioning binary
// introduced by issue #88's A6.1 re-cut. It is the ONLY thing in this repo
// that opens a connection as the bootstrap superuser (today's metaldocs_app /
// ${POSTGRES_USER}) after initial cluster bootstrap; every other process
// (metaldocs-api, metaldocs-worker, metaldocs-jobs) connects exclusively as
// metaldocs_runtime and boot-fatals if it doesn't (see
// internal/platform/db/postgres.AssertSafeIdentity).
//
// Splitting this out of application startup closes two findings against the
// original A6.1 slice:
//
//   - F1 (bootstrap deadlock): db/grants/0000_identity_roles.sql, which
//     creates metaldocs_runtime, previously could only reach an existing
//     database via internal/platform/migrate.ApplyGrants running INSIDE
//     metaldocs-api's own startup — but that startup now boot-fatals before
//     ApplyGrants runs if the connected identity is unsafe. The only
//     mechanism that could provision the safe role was gated behind the
//     check requiring that role to already exist. Moving provisioning to a
//     separate binary that connects as the bootstrap superuser and NEVER
//     calls AssertSafeIdentity makes that deadlock structurally
//     unrepresentable: nothing about the runtime pool's safety gates
//     what this binary can do.
//
//   - F2 (migration identity vs serving identity): DDL (db/migrations,
//     River's schema) must not run over the same identity that serves
//     requests, because metaldocs_runtime deliberately holds no DDL rights
//     (see db/grants/0001_role_grants.sql). This binary runs DDL under
//     metaldocs_owner instead — a NOLOGIN, NOSUPERUSER role reached only via
//     SET ROLE from the bootstrap superuser session — so migrated objects
//     are owned by metaldocs_owner, never by the bootstrap superuser and
//     never by metaldocs_runtime (RLS does not apply to a table's owner
//     unless FORCE ROW LEVEL SECURITY is set, so metaldocs_runtime owning
//     anything would be a silent RLS bypass).
//
// Invocation: two paths reach this one binary/entrypoint --
//   - deploy/compose/docker-compose.yml's db-provision service, which api/
//     worker/jobs depend on via `condition: service_completed_successfully`.
//   - scripts/start-api.ps1, which runs compiled binaries directly (not
//     through compose) and now runs this one first, pointed at the
//     bootstrap-superuser DSN, before launching worker/jobs/api.
//
// Boot-fatal-assertion pattern (name this explicitly so a later A6.3 slice
// extends it, not reinvents it): every composition root that must refuse to
// run under a specific unsafe identity/config posture opens its connection
// through a dedicated constructor that performs the check internally and
// fails the open itself — see postgres.OpenServing (calls
// postgres.AssertSafeIdentity) and, in this binary, postgres.OpenAsRole
// (pins DDL to metaldocs_owner via a connector hook, not a hand-paired
// statement a future call site could forget). A future boot-fatal check
// belongs in a matching dedicated constructor, not a second ad hoc call
// sprinkled into run() below.
//
// DDL execution identity (A6.1 review round 2 correction): DDL is provably
// pinned to metaldocs_owner via postgres.OpenAsRole, NOT via
// db.SetMaxOpenConns(1) plus a single SET ROLE statement on the shared
// bootstrap-superuser *sql.DB. That earlier shape looked like it pinned
// every later call to one physical connection, but did not: SET ROLE is
// session state, and nothing in database/sql's pool contract guarantees a
// logical *sql.DB stays on one physical connection across separate top-level
// calls (a pool reconnect after driver.ErrBadConn silently reverts to the
// connection's original bootstrap-superuser identity, with no error). See
// postgres.OpenAsRole's doc comment for the pgx stdlib AfterConnect/
// ResetSession mechanism that replaces it, and
// TestProvision_DDLObjectsAreOwnedByMetaldocsOwner (provision_integration_test.go)
// for the catalog-inspection (pg_class.relowner / pg_namespace.nspowner)
// proof that no DDL path falls back to the bootstrap superuser.
package main

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/jackc/pgx/v5"

	"metaldocs/internal/platform/bootstrap"
	"metaldocs/internal/platform/config"
	pgdb "metaldocs/internal/platform/db/postgres"
	"metaldocs/internal/platform/migrate"
)

func main() {
	os.Exit(runMain())
}

// runMain holds main's body so deferred cleanups run on every exit path.
func runMain() int {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := run(ctx); err != nil {
		slog.Error("db provisioning failed", "err", err)
		return 1
	}
	slog.Info("db provisioning complete")
	return 0
}

func run(ctx context.Context) error {
	migrationCfg, err := config.LoadMigrationConfig()
	if err != nil {
		return fmt.Errorf("load migration config: %w", err)
	}
	if migrationCfg.Skip {
		slog.Info("db provisioning skipped by configuration (METALDOCS_SKIP_STARTUP_MIGRATIONS=true)")
		return nil
	}

	pgCfg, err := config.LoadPostgresConfig()
	if err != nil {
		return fmt.Errorf("load postgres config: %w", err)
	}
	runtimeCfg := config.LoadRuntimeIdentityConfig()
	jobsCfg, err := config.LoadJobsConfig()
	if err != nil {
		return fmt.Errorf("load jobs config: %w", err)
	}

	// Deliberately pgdb.Open, NOT bootstrap.Build*Dependencies: this binary
	// must NEVER call postgres.AssertSafeIdentity. It is the one thing in
	// the system that is SUPPOSED to connect as the bootstrap superuser —
	// see api.go's "not folded into Open() itself" comment, which already
	// carves out exactly this kind of internal tooling.
	db, err := pgdb.Open(ctx, pgCfg.DSN)
	if err != nil {
		return fmt.Errorf("open postgres: %w", err)
	}
	defer func() { _ = db.Close() }()

	// ── Stage 1: superuser-only extensions (db/prerequisites) ──────────────
	// CREATE EXTENSION requires superuser (absent a pre-trusted allowlist);
	// nothing else in this sequence needs it, so this runs first, entirely
	// on the bootstrap-superuser connection.
	if err := migrate.ApplyGrants(ctx, db, migrationCfg.PrerequisitesDir, slog.Default()); err != nil {
		return fmt.Errorf("apply prerequisites stage: %w", err)
	}

	// ── Stage 2: identity roles + privilege grants (db/grants) ─────────────
	// Still running as the bootstrap superuser for BOTH files in this stage:
	// 0000_identity_roles.sql creates metaldocs_owner/metaldocs_runtime/
	// metaldocs_ci (idempotent, skips cleanly if already present) and
	// transfers schema + object ownership from the bootstrap superuser to
	// metaldocs_owner via a scoped per-object-kind ALTER ... OWNER TO loop
	// (NOT a blanket REASSIGN OWNED BY CURRENT_USER -- that statement always
	// fails for the literal bootstrap/initdb role; see that file's header for
	// why); 0001_role_grants.sql then grants DML privileges to
	// metaldocs_runtime/metaldocs_ci. Both grants files are idempotent and
	// safe to re-run against an already-provisioned volume.
	if err := migrate.ApplyGrants(ctx, db, migrationCfg.GrantsDir, slog.Default()); err != nil {
		return fmt.Errorf("apply grants stage: %w", err)
	}

	// ── Stage 2.5: runtime credential + custom River schema ownership ──────
	// Both still on the bootstrap-superuser connection, both must happen
	// AFTER role creation (Stage 2) and BEFORE any DDL runs under
	// metaldocs_owner (Stage 3) so the roles/schema they touch already exist
	// and so River's own migrator (Stage 3) finds a schema it can create
	// tables in.
	if err := rotateRuntimePassword(ctx, db, runtimeCfg.RuntimePassword); err != nil {
		return fmt.Errorf("rotate metaldocs_runtime password: %w", err)
	}
	if err := ensureRiverSchemaOwnership(ctx, db, jobsCfg.RiverSchema); err != nil {
		return fmt.Errorf("ensure river schema ownership: %w", err)
	}

	// ── Stage 3: DDL, under metaldocs_owner (NOT superuser) ─────────────────
	// ownerDB is a SEPARATE, dedicated connection pool: every physical
	// connection it ever hands out issues `SET ROLE metaldocs_owner` before
	// it is usable (postgres.OpenAsRole) and re-asserts that on every pooled
	// reuse, so DDL cannot silently fall back to running as the bootstrap
	// superuser across a pool reconnect. From here on, every object created
	// is owned by metaldocs_owner, never by the bootstrap superuser and
	// never by metaldocs_runtime — the hard constraint the serving identity
	// must never own a table (see db/grants/0000_identity_roles.sql header
	// for why).
	ownerDB, err := pgdb.OpenAsRole(ctx, pgCfg.DSN, "metaldocs_owner")
	if err != nil {
		return fmt.Errorf("open as metaldocs_owner: %w", err)
	}
	defer func() { _ = ownerDB.Close() }()

	if err := migrate.Apply(ctx, ownerDB, migrationCfg.Dir, slog.Default()); err != nil {
		return fmt.Errorf("apply forward migrations: %w", err)
	}

	if err := migrateRiverSchema(ctx, ownerDB, jobsCfg.RiverSchema); err != nil {
		return fmt.Errorf("migrate river schema: %w", err)
	}

	// Grant metaldocs_runtime DML access on a custom River schema now that
	// River's own migrator has created its tables there. Runs as
	// metaldocs_owner (ownerDB): metaldocs_owner owns everything this stage
	// just created, so it can GRANT on its own objects without superuser,
	// and ALTER DEFAULT PRIVILEGES with no FOR ROLE clause defaults to the
	// current role -- metaldocs_owner, via the same SET ROLE this connection
	// already carries -- which is exactly the identity future River
	// migrations will run as too.
	if err := grantRuntimeRiverSchemaAccess(ctx, ownerDB, jobsCfg.RiverSchema); err != nil {
		return fmt.Errorf("grant metaldocs_runtime river schema access: %w", err)
	}

	return nil
}

// migrateRiverSchema runs bootstrap.MigrateRiverSchema against schema. River's
// job-queue schema is DDL like any other migration, so it runs here — on the
// metaldocs_owner-pinned connection, see run()'s Stage 3 — rather than in
// metaldocs-api at startup (superseding F-19/REQ-ASYNC-4's original "API
// binary alone" ownership; see internal/platform/bootstrap/jobs.go's
// MigrateRiverSchema doc comment for the full rationale).
func migrateRiverSchema(ctx context.Context, db *sql.DB, schema string) error {
	return bootstrap.MigrateRiverSchema(ctx, db, schema)
}

// alterRolePasswordStatement renders `ALTER ROLE role PASSWORD ...` for an
// arbitrary role/password pair. Split out from rotateRuntimePassword so its
// SQL-escaping correctness (the actual security-relevant property here: no
// injection via a password value that itself contains a single quote) is
// unit-testable without a database connection, and so an integration test can
// exercise the exact same statement shape against a disposable role instead
// of the cluster-global metaldocs_runtime (issue #88 / A6.1 review C2 — see
// provision_integration_test.go).
//
// ALTER ROLE ... PASSWORD takes a grammar-level string literal (Sconst), not
// a bind parameter -- extended-protocol placeholders like $1 do not parse in
// that position. Manual quote-doubling is the correct (and only) way to embed
// an arbitrary value there. role is identifier-quoted the same way (defense
// in depth; every call site today passes the literal "metaldocs_runtime").
func alterRolePasswordStatement(role, password string) string {
	quoted := "'" + strings.ReplaceAll(password, "'", "''") + "'"
	return fmt.Sprintf("ALTER ROLE %s PASSWORD %s", pgx.Identifier{role}.Sanitize(), quoted)
}

// rotateRuntimePassword sets metaldocs_runtime's login password to password
// (the same value api/worker/jobs read as their own PGPASSWORD via
// METALDOCS_RUNTIME_DB_PASSWORD — see config.LoadRuntimeIdentityConfig), on
// the bootstrap-superuser connection. Idempotent: always safe to re-run, even
// when the password is unchanged. Never logged (the error path below never
// includes password or the rendered statement).
func rotateRuntimePassword(ctx context.Context, db *sql.DB, password string) error {
	if _, err := db.ExecContext(ctx, alterRolePasswordStatement("metaldocs_runtime", password)); err != nil {
		return fmt.Errorf("set metaldocs_runtime password: %w", err)
	}
	return nil
}

// ensureRiverSchemaOwnership makes a non-default METALDOCS_JOBS_RIVER_SCHEMA
// provisionable the same way public/metaldocs already are:
// db/grants/0000_identity_roles.sql's ownership transfer only ever covers
// those two hardcoded schemas, because a schema named only by a runtime env
// var cannot be baked into hand-authored SQL. schema == "" or "public" is a
// no-op — both are already covered by that file.
//
// Runs on the bootstrap-superuser connection: creating a brand-new schema, or
// re-owning a pre-existing one (the "existing deployment" case named in the
// PR #110 review — a custom schema created by an older, pre-A6.1 volume,
// still owned by the bootstrap superuser), both require database-level CREATE
// / ownership-transfer privilege that metaldocs_owner does not hold.
func ensureRiverSchemaOwnership(ctx context.Context, db *sql.DB, schema string) error {
	if schema == "" || schema == "public" {
		return nil
	}
	ident := pgx.Identifier{schema}.Sanitize()
	if _, err := db.ExecContext(ctx, fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s AUTHORIZATION metaldocs_owner", ident)); err != nil {
		return fmt.Errorf("create river schema %q: %w", schema, err)
	}
	// CREATE SCHEMA IF NOT EXISTS skips its AUTHORIZATION clause entirely
	// when the schema already exists, so a pre-existing schema still needs
	// an explicit, idempotent ownership transfer.
	if _, err := db.ExecContext(ctx, fmt.Sprintf("ALTER SCHEMA %s OWNER TO metaldocs_owner", ident)); err != nil {
		return fmt.Errorf("transfer ownership of river schema %q to metaldocs_owner: %w", schema, err)
	}
	return nil
}

// grantRuntimeRiverSchemaAccess grants metaldocs_runtime the same DML-only
// posture on a custom River schema that db/grants/0001_role_grants.sql
// already grants it on public/metaldocs — including a default-privileges
// rule so a FUTURE River migration's tables auto-grant too, the same
// reasoning that file's own comment gives for metaldocs_owner's default
// privileges. schema == "" or "public" is a no-op — 0001_role_grants.sql
// already covers both.
func grantRuntimeRiverSchemaAccess(ctx context.Context, db *sql.DB, schema string) error {
	if schema == "" || schema == "public" {
		return nil
	}
	ident := pgx.Identifier{schema}.Sanitize()
	stmts := []string{
		fmt.Sprintf("GRANT USAGE ON SCHEMA %s TO metaldocs_runtime", ident),
		fmt.Sprintf("GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA %s TO metaldocs_runtime", ident),
		fmt.Sprintf("GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA %s TO metaldocs_runtime", ident),
		fmt.Sprintf("ALTER DEFAULT PRIVILEGES IN SCHEMA %s GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO metaldocs_runtime", ident),
		fmt.Sprintf("ALTER DEFAULT PRIVILEGES IN SCHEMA %s GRANT USAGE, SELECT ON SEQUENCES TO metaldocs_runtime", ident),
	}
	for _, stmt := range stmts {
		if _, err := db.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("grant metaldocs_runtime access on river schema %q: %w", schema, err)
		}
	}
	return nil
}
