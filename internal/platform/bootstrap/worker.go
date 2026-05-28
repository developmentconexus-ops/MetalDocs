package bootstrap

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"metaldocs/internal/platform/config"
	pgdb "metaldocs/internal/platform/db/postgres"
	"metaldocs/internal/platform/messaging"
	outboxpg "metaldocs/internal/platform/messaging/outbox/postgres"
	"metaldocs/internal/platform/servicebus"
)

type WorkerDependencies struct {
	Consumer       messaging.Consumer
	DocgenV2Client *servicebus.DocgenV2Client
	SQLDB          *sql.DB
	Cleanup        func()
}

func BuildWorkerDependencies(ctx context.Context, workerCfg config.WorkerConfig) (WorkerDependencies, error) {
	pgCfg, err := config.LoadPostgresConfig()
	if err != nil {
		return WorkerDependencies{}, fmt.Errorf("load postgres config: %w", err)
	}
	db, err := pgdb.Open(ctx, pgCfg.DSN)
	if err != nil {
		return WorkerDependencies{}, fmt.Errorf("open postgres: %w", err)
	}

	docgenV2Cfg, err := config.LoadDocgenV2Config()
	if err != nil {
		_ = closeDB(db)
		return WorkerDependencies{}, fmt.Errorf("load docgen-v2 config: %w", err)
	}
	var docgenV2Client *servicebus.DocgenV2Client
	if docgenV2Cfg.Enabled {
		docgenV2Client = servicebus.NewDocgenV2Client(
			docgenV2Cfg.APIURL,
			docgenV2Cfg.ServiceToken,
			time.Duration(docgenV2Cfg.RequestTimeoutSeconds)*time.Second,
		)
	}

	consumer := outboxpg.NewConsumer(db, workerClaimLease(workerCfg))

	return WorkerDependencies{
		Consumer:       consumer,
		DocgenV2Client: docgenV2Client,
		SQLDB:          db,
		Cleanup:        func() { _ = closeDB(db) },
	}, nil
}

func workerClaimLease(cfg config.WorkerConfig) time.Duration {
	const minClaimLease = 5 * time.Minute

	lease := time.Duration(cfg.RetryMaxSeconds) * time.Second
	if lease < minClaimLease {
		return minClaimLease
	}
	return lease
}
