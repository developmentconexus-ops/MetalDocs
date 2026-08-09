package bootstrap

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverdatabasesql"
	"github.com/riverqueue/river/rivermigrate"

	"metaldocs/internal/platform/config"
	pgdb "metaldocs/internal/platform/db/postgres"
	riverjobs "metaldocs/internal/platform/jobs/river"
)

// JobsWorkerFactory builds the River workers and periodic jobs registered
// against the given database handle.
type JobsWorkerFactory func(db *sql.DB) (*river.Workers, []*river.PeriodicJob, error)

// JobsDependencies holds the constructed infrastructure dependencies the
// metaldocs-jobs binary wires into its River client at startup.
type JobsDependencies struct {
	River   *riverjobs.ClientBundle
	SQLDB   *sql.DB
	Cleanup func()
}

// BuildJobsDependencies constructs the jobs binary's infrastructure
// dependencies, including the River client bundle built from workerFactory.
func BuildJobsDependencies(ctx context.Context, cfg config.JobsConfig, workerFactory JobsWorkerFactory) (JobsDependencies, error) {
	pgCfg, err := config.LoadPostgresConfig()
	if err != nil {
		return JobsDependencies{}, fmt.Errorf("load postgres config: %w", err)
	}

	db, err := pgdb.Open(ctx, pgCfg.DSN)
	if err != nil {
		return JobsDependencies{}, fmt.Errorf("open postgres: %w", err)
	}

	// River schema migration is owned by the API binary alone (F-19,
	// REQ-ASYNC-4): metaldocs-api runs MigrateRiverSchema at startup, and
	// the jobs compose service has depends_on: api (healthy), so the schema
	// exists before this binary starts. Running it here too gave the schema
	// two owners with no declared order.

	var workers *river.Workers
	var periodicJobs []*river.PeriodicJob
	if workerFactory != nil {
		workers, periodicJobs, err = workerFactory(db)
		if err != nil {
			_ = closeDB(db)
			return JobsDependencies{}, fmt.Errorf("build jobs workers: %w", err)
		}
	}

	riverBundle, err := riverjobs.NewClientBundle(db, riverjobs.Config{
		Queues:       cfg.Queues,
		Schema:       cfg.RiverSchema,
		PeriodicJobs: periodicJobs,
		// M5 F5.4 T2: explicit retention values (binding per contract §4.1),
		// matching River's own documented defaults — set explicitly so a
		// future River default change can't silently drift the policy.
		// metaldocs-jobs is the designated host binary for the purge worker
		// (contract §1); metaldocs-api's enqueue-only Config intentionally
		// leaves these at zero (inherits River's own defaults, harmless
		// since api never runs the cleaner as the primary contract-pinned
		// host).
		CompletedJobRetentionPeriod: 24 * time.Hour,
		CancelledJobRetentionPeriod: 24 * time.Hour,
		DiscardedJobRetentionPeriod: 7 * 24 * time.Hour,
	}, workers)
	if err != nil {
		_ = closeDB(db)
		return JobsDependencies{}, fmt.Errorf("build river client: %w", err)
	}

	return JobsDependencies{
		River: riverBundle,
		SQLDB: db,
		Cleanup: func() {
			_ = closeDB(db)
		},
	}, nil
}

// MigrateRiverSchema runs River's schema migrations up against the given
// schema. Owned exclusively by the API binary at startup (F-19, REQ-ASYNC-4).
func MigrateRiverSchema(ctx context.Context, db *sql.DB, schema string) error {
	migrator, err := rivermigrate.New(riverdatabasesql.New(db), &rivermigrate.Config{Schema: schema})
	if err != nil {
		return fmt.Errorf("build river migrator: %w", err)
	}

	if _, err := migrator.Migrate(ctx, rivermigrate.DirectionUp, nil); err != nil {
		return fmt.Errorf("migrate river schema: %w", err)
	}

	return nil
}
