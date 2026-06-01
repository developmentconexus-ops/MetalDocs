package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	docapp "metaldocs/internal/modules/documents/application"
	docrepo "metaldocs/internal/modules/documents/repository"
	fanoutpkg "metaldocs/internal/modules/render/fanout"
	"metaldocs/internal/modules/render/resolvers"
	"metaldocs/internal/platform/bootstrap"
	"metaldocs/internal/platform/config"
	"metaldocs/internal/platform/httpclient"
	workerapp "metaldocs/internal/platform/worker"
)

type workerBatchRunner interface {
	RunOnce(ctx context.Context, batchSize int) error
}

// materializeInvokerAdapter bridges docapp.FreezeService to workerapp.MaterializeInvoker.
type materializeInvokerAdapter struct {
	svc *docapp.FreezeService
}

func (a materializeInvokerAdapter) Materialize(ctx context.Context, tenantID, revisionID string) (workerapp.MaterializeFanoutResult, error) {
	res, err := a.svc.Materialize(ctx, tenantID, revisionID)
	if err != nil {
		return workerapp.MaterializeFanoutResult{}, err
	}
	return workerapp.MaterializeFanoutResult{
		FinalDocxS3Key: res.FinalDocxS3Key,
		ContentHash:    res.ContentHash,
	}, nil
}

// snapshotFinalDocxAdapter bridges SnapshotRepository to workerapp.MaterializeFinalDocxPersister.
type snapshotFinalDocxAdapter struct {
	repo *docrepo.SnapshotRepository
}

func (a snapshotFinalDocxAdapter) WriteFinalDocxInTx(ctx context.Context, tx *sql.Tx, tenantID, revisionID, s3Key string, contentHash []byte) error {
	return a.repo.WriteFinalDocx(ctx, tenantID, revisionID, s3Key, contentHash, tx)
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	workerCfg, err := config.LoadWorkerConfig()
	if err != nil {
		log.Fatalf("invalid worker config: %v", err)
	}
	deps, err := bootstrap.BuildWorkerDependencies(ctx, workerCfg)
	if err != nil {
		log.Fatalf("build worker dependencies: %v", err)
	}
	defer deps.Cleanup()

	workerSvc := workerapp.NewService(deps.Consumer, workerCfg)

	if deps.PDFConverter != nil && deps.SQLDB != nil {
		snapRepo := docrepo.NewSnapshotRepository(deps.SQLDB)
		pdfRunner := workerapp.NewPDFJobRunner(deps.PDFConverter, workerapp.NewSnapshotPDFPersister(snapRepo))
		workerSvc = workerSvc.WithPDFRunner(pdfRunner)
	}

	if deps.FanoutURL != "" && deps.SQLDB != nil {
		fanoutClient := fanoutpkg.NewClient(deps.FanoutURL, deps.FanoutToken, httpclient.NewInternalClient())
		snapRepo := docrepo.NewSnapshotRepository(deps.SQLDB)
		fillInRepo := docrepo.NewFillInRepository(deps.SQLDB)
		schemaReader := docapp.NewSnapshotSchemaReader(deps.SQLDB)
		resolverReg := resolvers.NewRegistry()
		resolvers.RegisterBuiltins(resolverReg)

		// ctxBuilder not needed for Materialize (reads already-resolved values).
		freezeSvc := docapp.NewFreezeService(
			schemaReader, fillInRepo, fillInRepo,
			resolverReg, snapRepo, nil,
			snapRepo, snapRepo, fanoutClient,
		)

		pdfOutboxRepo := fanoutpkg.NewPDFOutboxRepository(deps.SQLDB)
		materializeRunner := workerapp.NewMaterializeJobRunner(
			materializeInvokerAdapter{svc: freezeSvc},
			snapshotFinalDocxAdapter{repo: snapRepo},
			pdfOutboxRepo,
			deps.SQLDB,
		)
		workerSvc = workerSvc.WithMaterializeRunner(materializeRunner)
		log.Printf("materialize runner active (fanout_url=%s)", deps.FanoutURL)
	}

	if workerCfg.RunOnce {
		if err := runWorkerBatch(ctx, workerSvc, workerCfg.BatchSize); err != nil {
			log.Fatalf("worker run failed: %v", err)
		}
		return
	}

	ticker := time.NewTicker(time.Duration(workerCfg.PollIntervalSeconds) * time.Second)
	defer ticker.Stop()
	log.Printf("MetalDocs Worker running (poll_interval_s=%d batch_size=%d review_reminder_days=%d max_attempts=%d retry_base_seconds=%d retry_max_seconds=%d)",
		workerCfg.PollIntervalSeconds, workerCfg.BatchSize, workerCfg.ReviewReminderDays, workerCfg.MaxAttempts, workerCfg.RetryBaseSeconds, workerCfg.RetryMaxSeconds)

	runWorkerLoop(ctx, workerSvc, workerCfg.BatchSize, ticker.C)
}

func runWorkerBatch(ctx context.Context, runner workerBatchRunner, batchSize int) error {
	if err := runner.RunOnce(ctx, batchSize); err != nil {
		if ctx.Err() != nil {
			return nil
		}
		return err
	}
	log.Printf("worker batch completed")
	return nil
}

func runWorkerLoop(ctx context.Context, runner workerBatchRunner, batchSize int, ticks <-chan time.Time) {
	for {
		if err := runWorkerBatch(ctx, runner, batchSize); err != nil {
			log.Printf("worker run failed: %v", err)
		}
		select {
		case <-ctx.Done():
			return
		case <-ticks:
		}
	}
}
