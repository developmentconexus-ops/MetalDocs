package bootstrap

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverdatabasesql"
	"github.com/riverqueue/river/rivermigrate"

	"metaldocs/internal/platform/config"
	pgdb "metaldocs/internal/platform/db/postgres"
	riverjobs "metaldocs/internal/platform/jobs/river"
)

type JobsWorkerFactory func(db *sql.DB) (*river.Workers, error)

type JobsDependencies struct {
	River   *riverjobs.ClientBundle
	SQLDB   *sql.DB
	Cleanup func()
}

func BuildJobsDependencies(ctx context.Context, cfg config.JobsConfig, workerFactory JobsWorkerFactory) (JobsDependencies, error) {
	pgCfg, err := config.LoadPostgresConfig()
	if err != nil {
		return JobsDependencies{}, fmt.Errorf("load postgres config: %w", err)
	}

	db, err := pgdb.Open(ctx, pgCfg.DSN)
	if err != nil {
		return JobsDependencies{}, fmt.Errorf("open postgres: %w", err)
	}

	if err := MigrateRiverSchema(ctx, db, cfg.RiverSchema); err != nil {
		_ = closeDB(db)
		return JobsDependencies{}, err
	}

	var workers *river.Workers
	if workerFactory != nil {
		workers, err = workerFactory(db)
		if err != nil {
			_ = closeDB(db)
			return JobsDependencies{}, fmt.Errorf("build jobs workers: %w", err)
		}
	}

	riverBundle, err := riverjobs.NewClientBundle(db, riverjobs.Config{
		Queues: cfg.Queues,
		Schema: cfg.RiverSchema,
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
