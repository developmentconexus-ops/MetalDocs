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
//     `SET ROLE metaldocs_owner` instead — a NOLOGIN, NOSUPERUSER role
//     reached only from the bootstrap superuser session — so migrated
//     objects are owned by metaldocs_owner, never by the bootstrap
//     superuser and never by metaldocs_runtime (RLS does not apply to a
//     table's owner unless FORCE ROW LEVEL SECURITY is set, so
//     metaldocs_runtime owning anything would be a silent RLS bypass).
//
// Invocation: two paths reach this one binary/entrypoint --
//   - deploy/compose/docker-compose.yml's db-provision service, which api/
//     worker/jobs depend on via `condition: service_completed_successfully`.
//   - scripts/start-api.ps1, which runs compiled binaries directly (not
//     through compose) and now runs this one first, pointed at the
//     bootstrap-superuser DSN, before launching worker/jobs/api.
//
// Ordering and identity-persistence contract: db.SetMaxOpenConns(1) /
// SetMaxIdleConns(1) below pin this process to a single physical connection
// for its entire (short, single-goroutine, sequential) run, so `SET ROLE
// metaldocs_owner` — a session-scoped statement — reliably persists across
// the separate migrate.Apply/MigrateRiverSchema calls that follow it, each
// of which internally acquires its own *sql.Conn from the pool. (Verified
// against github.com/jackc/pgx/v5/stdlib's Conn.ResetSession: it only pings
// and checks transaction status before reuse, it does not discard or reset
// server-side session state, so this does not get silently undone between
// calls.)
package main

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

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

	// Pin to a single physical connection for the whole run so SET ROLE
	// (below) persists across every subsequent call on this *sql.DB. See the
	// package doc comment for why this is safe with the pgx stdlib driver.
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	// ── Stage 1: superuser-only extensions (db/prerequisites) ──────────────
	// CREATE EXTENSION requires superuser (absent a pre-trusted allowlist);
	// nothing else in this sequence needs it, so this runs first and is the
	// only stage that never touches SET ROLE.
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
	// metaldocs_runtime/metaldocs_ci. Neither file runs under SET ROLE
	// metaldocs_owner -- that only starts at Stage 3 below. Both grants files
	// are idempotent and safe to re-run against an already-provisioned volume.
	if err := migrate.ApplyGrants(ctx, db, migrationCfg.GrantsDir, slog.Default()); err != nil {
		return fmt.Errorf("apply grants stage: %w", err)
	}

	// ── Stage 3: DDL, under metaldocs_owner (NOT superuser) ─────────────────
	// From here on, every object created is owned by metaldocs_owner, never
	// by the bootstrap superuser and never by metaldocs_runtime — the hard
	// constraint the serving identity must never own a table (see
	// db/grants/0000_identity_roles.sql header for why).
	if _, err := db.ExecContext(ctx, "SET ROLE metaldocs_owner"); err != nil {
		return fmt.Errorf("set role metaldocs_owner: %w", err)
	}
	defer func() {
		if _, err := db.ExecContext(context.WithoutCancel(ctx), "RESET ROLE"); err != nil {
			slog.Warn("reset role after provisioning", "err", err)
		}
	}()

	if err := migrate.Apply(ctx, db, migrationCfg.Dir, slog.Default()); err != nil {
		return fmt.Errorf("apply forward migrations: %w", err)
	}

	if err := migrateRiverSchema(ctx, db); err != nil {
		return fmt.Errorf("migrate river schema: %w", err)
	}

	return nil
}

// migrateRiverSchema loads the River schema name from jobs config and runs
// bootstrap.MigrateRiverSchema. River's job-queue schema is DDL like any
// other migration, so it runs here — under SET ROLE metaldocs_owner, set
// above — rather than in metaldocs-api at startup (superseding F-19/
// REQ-ASYNC-4's original "API binary alone" ownership; see
// internal/platform/bootstrap/jobs.go's MigrateRiverSchema doc comment for
// the full rationale).
func migrateRiverSchema(ctx context.Context, db *sql.DB) error {
	jobsCfg, err := config.LoadJobsConfig()
	if err != nil {
		return fmt.Errorf("load jobs config: %w", err)
	}
	return bootstrap.MigrateRiverSchema(ctx, db, jobsCfg.RiverSchema)
}
