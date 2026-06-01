package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	docrepo "metaldocs/internal/modules/documents/repository"
	"metaldocs/internal/platform/bootstrap"
	"metaldocs/internal/platform/config"
	workerapp "metaldocs/internal/platform/worker"
)

type workerBatchRunner interface {
	RunOnce(ctx context.Context, batchSize int) error
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
