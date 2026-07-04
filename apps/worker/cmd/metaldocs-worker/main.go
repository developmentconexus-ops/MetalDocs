package main

import (
	"context"
	"log/slog"
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
	"metaldocs/internal/platform/db"
	"metaldocs/internal/platform/httpclient"
	"metaldocs/internal/platform/observability"
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

func (a snapshotFinalDocxAdapter) WriteFinalDocxInTx(ctx context.Context, tx db.Tx, tenantID, revisionID, s3Key string, contentHash []byte) error {
	return a.repo.WriteFinalDocx(ctx, tenantID, revisionID, s3Key, contentHash, tx)
}

// snapshotPDFTxAdapter bridges SnapshotRepository to workerapp.PDFPersisterInTx
// (M3 F3.2 — validation-contract.md §2.2 site 2: the pdf job runner's write
// must run inside a SeedTxTenant-seeded tx).
type snapshotPDFTxAdapter struct {
	repo *docrepo.SnapshotRepository
}

func (a snapshotPDFTxAdapter) WritePDFInTx(ctx context.Context, tx db.Tx, req workerapp.PDFWriteRequest) error {
	return a.repo.WritePDF(
		ctx,
		string(req.TenantID),
		string(req.DocumentID),
		string(req.StorageKey),
		req.PDFHash,
		req.GeneratedAt,
		tx,
	)
}

// WritePDF satisfies workerapp.PDFPersister so snapshotPDFTxAdapter can be
// passed directly to NewPDFJobRunnerWithDB (which requires both interfaces);
// the runner always prefers WritePDFInTx when db != nil, so this untransacted
// path is dead code in production wiring but keeps the adapter a complete
// PDFPersister for type-safety.
func (a snapshotPDFTxAdapter) WritePDF(ctx context.Context, req workerapp.PDFWriteRequest) error {
	return a.repo.WritePDF(
		ctx,
		string(req.TenantID),
		string(req.DocumentID),
		string(req.StorageKey),
		req.PDFHash,
		req.GeneratedAt,
	)
}

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// OpenTelemetry: inert unless an exporter is configured (Z-1, REQ-OBS-3).
	// otelShutdown is a no-op when disabled; otelEnabled gates the chain link.
	otelShutdown, otelEnabled, err := observability.SetupOTel(ctx, "metaldocs-worker")
	if err != nil {
		slog.Error("setup otel", "err", err)
		os.Exit(1)
	}
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := otelShutdown(shutdownCtx); err != nil {
			slog.Warn("otel shutdown", "err", err)
		}
	}()
	if otelEnabled {
		slog.Info("OpenTelemetry tracing enabled", "exporter", os.Getenv("OTEL_TRACES_EXPORTER"))
	}

	workerCfg, err := config.LoadWorkerConfig()
	if err != nil {
		slog.Error("invalid worker config", "err", err)
		os.Exit(1)
	}
	deps, err := bootstrap.BuildWorkerDependencies(ctx, workerCfg)
	if err != nil {
		slog.Error("build worker dependencies", "err", err)
		os.Exit(1)
	}
	defer deps.Cleanup()

	workerSvc := workerapp.NewService(deps.Consumer, workerCfg)

	if deps.PDFConverter != nil && deps.SQLDB != nil {
		snapRepo := docrepo.NewSnapshotRepository(deps.SQLDB)
		// M3 F3.2 (validation-contract.md §2.2 site 2): wrap the PDF write in a
		// SeedTxTenant-seeded tx so the FORCE RLS backstop engages.
		pdfRunner := workerapp.NewPDFJobRunnerWithDB(deps.PDFConverter, snapshotPDFTxAdapter{repo: snapRepo}, deps.SQLDB)
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
			snapRepo, fanoutClient,
		)

		pdfOutboxRepo := fanoutpkg.NewPDFOutboxRepository(deps.SQLDB)
		materializeRunner := workerapp.NewMaterializeJobRunner(
			materializeInvokerAdapter{svc: freezeSvc},
			snapshotFinalDocxAdapter{repo: snapRepo},
			pdfOutboxRepo,
			deps.SQLDB,
		)
		workerSvc = workerSvc.WithMaterializeRunner(materializeRunner)
		slog.Info("materialize runner active", "fanout_url", deps.FanoutURL)
	}

	if workerCfg.RunOnce {
		if err := runWorkerBatch(ctx, workerSvc, workerCfg.BatchSize); err != nil {
			slog.Error("worker run failed", "err", err)
			deps.Cleanup()
			os.Exit(1)
		}
		return
	}

	ticker := time.NewTicker(time.Duration(workerCfg.PollIntervalSeconds) * time.Second)
	defer ticker.Stop()
	slog.Info("MetalDocs Worker running",
		"poll_interval_s", workerCfg.PollIntervalSeconds, "batch_size", workerCfg.BatchSize,
		"max_attempts", workerCfg.MaxAttempts, "retry_base_seconds", workerCfg.RetryBaseSeconds,
		"retry_max_seconds", workerCfg.RetryMaxSeconds)

	runWorkerLoop(ctx, workerSvc, workerCfg.BatchSize, ticker.C)
}

func runWorkerBatch(ctx context.Context, runner workerBatchRunner, batchSize int) error {
	if err := runner.RunOnce(ctx, batchSize); err != nil {
		if ctx.Err() != nil {
			return nil
		}
		return err
	}
	slog.Info("worker batch completed")
	return nil
}

// runWorkerLoop polls on ticks and runs a batch each iteration.
// Graceful drain: when the signal arrives (ctx cancelled), any in-flight batch
// is allowed to finish using a detached context with a 30 s deadline before the
// loop exits. This prevents mid-batch abandonment while keeping the drain
// bounded so the process does not hang indefinitely on a slow job.
func runWorkerLoop(ctx context.Context, runner workerBatchRunner, batchSize int, ticks <-chan time.Time) {
	for {
		// Use a detached context for the batch so that a signal arriving
		// mid-batch does not abort it; the outer loop's post-batch select
		// will detect ctx.Done() and exit cleanly after the batch finishes.
		batchCtx, batchCancel := context.WithTimeout(context.Background(), 30*time.Second)
		err := runner.RunOnce(batchCtx, batchSize)
		batchCancel()

		if err != nil {
			slog.Error("worker run failed", "err", err)
		}

		select {
		case <-ctx.Done():
			return
		case <-ticks:
		}
	}
}
