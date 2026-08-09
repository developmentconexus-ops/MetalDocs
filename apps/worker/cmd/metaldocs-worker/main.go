package main

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	approvalapp "metaldocs/internal/modules/approval/application"
	approvaljobs "metaldocs/internal/modules/approval/jobs"
	docapp "metaldocs/internal/modules/documents/application"
	docrepo "metaldocs/internal/modules/documents/infrastructure"
	"metaldocs/internal/modules/iam/authz"
	fanoutpkg "metaldocs/internal/modules/render/fanout"
	"metaldocs/internal/modules/render/fanout/dispatchjobs"
	"metaldocs/internal/modules/render/resolvers"
	"metaldocs/internal/platform/bootstrap"
	"metaldocs/internal/platform/config"
	"metaldocs/internal/platform/db"
	"metaldocs/internal/platform/httpclient"
	riverjobs "metaldocs/internal/platform/jobs/river"
	"metaldocs/internal/platform/objectstore"
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

// authzSeamAdapter bridges internal/modules/iam/authz to
// workerapp.TxAuthzSeam (REQ-TOP-2: internal/platform must not import
// internal/modules — this adapter lives at the composition root instead).
type authzSeamAdapter struct{}

func (authzSeamAdapter) WithBackgroundBypass(ctx context.Context) context.Context {
	return authz.WithBackgroundBypass(ctx)
}

func (authzSeamAdapter) SeedTxTenant(ctx context.Context, tx *sql.Tx, tenantID string) error {
	return authz.SeedTxTenant(ctx, tx, tenantID)
}

func (authzSeamAdapter) BypassSystem(ctx context.Context, tx *sql.Tx) error {
	return authz.BypassSystem(ctx, tx)
}

// buildRevisionBodyReader builds the blob reader Materialize uses to fetch the
// PINNED revision body and verify its bytes against document_revisions.content_hash
// before rendering (F-QA3-1 lineage). Returns (nil, nil) when the deployment has
// no MinIO attachments provider — the caller decides whether that is fatal.
func buildRevisionBodyReader() (*objectstore.VerifiedStore, error) {
	attachmentsCfg, err := config.LoadAttachmentsConfig()
	if err != nil {
		return nil, fmt.Errorf("invalid attachments config: %w", err)
	}
	if attachmentsCfg.Provider != config.StorageProviderMinIO {
		return nil, nil
	}
	minioClient, minioPublicClient, minioBucket, err := bootstrap.BuildMinioClients(attachmentsCfg)
	if err != nil {
		return nil, fmt.Errorf("build minio clients: %w", err)
	}
	return objectstore.NewVerifiedStore(minioClient, minioPublicClient, minioBucket, maxRevisionBodyBytes), nil
}

// maxRevisionBodyBytes mirrors the cap the freeze service passes to
// AssertedReadObject; the store's own cap must not be tighter than the read.
const maxRevisionBodyBytes = 25 * 1024 * 1024

func main() {
	os.Exit(runMain())
}

// runMain holds main's body so deferred cleanups (signal stop, otel shutdown,
// deps.Cleanup) run on every exit path; os.Exit in main proper would skip them.
func runMain() int {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// OpenTelemetry: inert unless an exporter is configured (Z-1, REQ-OBS-3).
	// otelShutdown is a no-op when disabled; otelEnabled gates the chain link.
	otelShutdown, otelEnabled, err := observability.SetupOTel(ctx, "metaldocs-worker")
	if err != nil {
		slog.Error("setup otel", "err", err)
		return 1
	}
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := otelShutdown(shutdownCtx); err != nil {
			slog.Warn("otel shutdown", "err", err)
		}
	}()
	if otelEnabled {
		slog.Info("OpenTelemetry tracing enabled", "exporter", os.Getenv("OTEL_TRACES_EXPORTER")) //nolint:gosec // G706: slog default is JSONHandler (set at process start) — control chars are JSON-escaped, log-line injection not possible
	}

	workerCfg, err := config.LoadWorkerConfig()
	if err != nil {
		slog.Error("invalid worker config", "err", err)
		return 1
	}
	deps, err := bootstrap.BuildWorkerDependencies(ctx, workerCfg)
	if err != nil {
		slog.Error("build worker dependencies", "err", err)
		return 1
	}
	defer deps.Cleanup()

	workerSvc, ok := buildWorkerService(deps, workerCfg)
	if !ok {
		return 1
	}

	if workerCfg.RunOnce {
		if err := runWorkerBatch(ctx, workerSvc, workerCfg.BatchSize); err != nil {
			slog.Error("worker run failed", "err", err)
			return 1
		}
		return 0
	}

	ticker := time.NewTicker(time.Duration(workerCfg.PollIntervalSeconds) * time.Second)
	defer ticker.Stop()
	slog.Info("MetalDocs Worker running",
		"poll_interval_s", workerCfg.PollIntervalSeconds, "batch_size", workerCfg.BatchSize,
		"max_attempts", workerCfg.MaxAttempts, "retry_base_seconds", workerCfg.RetryBaseSeconds,
		"retry_max_seconds", workerCfg.RetryMaxSeconds)

	runWorkerLoop(ctx, workerSvc, workerCfg.BatchSize, ticker.C)
	return 0
}

// buildWorkerService wires the base outbox service plus the optional PDF and
// materialize runners. Wiring failures are logged at the failing site; the
// false return tells the caller to exit non-zero (cleanup runs via its defers).
func buildWorkerService(deps bootstrap.WorkerDependencies, workerCfg config.WorkerConfig) (*workerapp.Service, bool) {
	workerSvc := workerapp.NewService(deps.Consumer, workerCfg)

	// ADR 0085: both artifact writers record their release fact through the
	// approval module's port, in the same transaction as the artifact write.
	// The enqueue-only River bundle is built once here because BOTH runners
	// need it (the fact producer enqueues the coordinator evaluation).
	var artifactFactRecorder *approvalapp.ArtifactFactWriter
	var sharedRiverBundle *riverjobs.ClientBundle
	if deps.SQLDB != nil {
		jobsCfg, err := config.LoadJobsConfig()
		if err != nil {
			slog.Error("invalid jobs config", "err", err)
			return nil, false
		}
		sharedRiverBundle, err = riverjobs.NewClientBundle(deps.SQLDB, riverjobs.Config{
			Schema:              jobsCfg.RiverSchema,
			SkipUnknownJobCheck: true,
		}, nil)
		if err != nil {
			slog.Error("build release evaluation enqueuer client", "err", err)
			return nil, false
		}
		artifactFactRecorder = approvalapp.NewArtifactFactWriter(
			approvaljobs.NewReleaseEvaluationEnqueuer(sharedRiverBundle.Client))
	}

	if deps.PDFConverter != nil && deps.SQLDB != nil {
		snapRepo := docrepo.NewSnapshotRepository(deps.SQLDB)
		// M3 F3.2 (validation-contract.md §2.2 site 2): wrap the PDF write in a
		// SeedTxTenant-seeded tx so the FORCE RLS backstop engages.
		pdfRunner := workerapp.NewPDFJobRunnerWithDB(deps.PDFConverter, snapshotPDFTxAdapter{repo: snapRepo}, authzSeamAdapter{}, deps.SQLDB).
			WithArtifactFactRecorder(artifactFactRecorder)
		workerSvc = workerSvc.WithPDFRunner(pdfRunner)
	}

	if deps.FanoutURL != "" && deps.SQLDB != nil {
		materializeRunner, ok := buildMaterializeRunner(deps, sharedRiverBundle, artifactFactRecorder)
		if !ok {
			return nil, false
		}
		workerSvc = workerSvc.WithMaterializeRunner(materializeRunner)
		slog.Info("materialize runner active", "fanout_url", deps.FanoutURL)
	}

	return workerSvc, true
}

// buildMaterializeRunner wires the freeze service, verified blob reader and
// pdf-dispatch enqueuer behind the materialize job runner (ADR 0015 path).
func buildMaterializeRunner(deps bootstrap.WorkerDependencies, sharedRiverBundle *riverjobs.ClientBundle, artifactFactRecorder *approvalapp.ArtifactFactWriter) (*workerapp.MaterializeJobRunner, bool) {
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

	// Materialize verifies the pinned revision's bytes against that
	// revision's recorded content_hash before rendering anything (F-QA3-1
	// lineage, fail closed). Without a blob reader it cannot verify, and it
	// refuses to render rather than skip the check — so a Gotenberg/MinIO-less
	// deploy fails loudly here instead of silently producing unverified
	// artifacts.
	blobs, err := buildRevisionBodyReader()
	if err != nil {
		slog.Error("build revision body reader", "err", err)
		return nil, false
	}
	if blobs == nil {
		slog.Error("materialize requires a MinIO attachments provider to verify frozen lineage")
		return nil, false
	}
	freezeSvc = freezeSvc.WithRevisionBodyReader(blobs)

	pdfOutboxRepo := fanoutpkg.NewPDFOutboxRepository(deps.SQLDB)
	// The Enqueuer constructor needs both staging outbox repos even though
	// this binary only ever enqueues pdf dispatch (M5 F5.3 T3).
	materializeOutboxRepo := fanoutpkg.NewMaterializeOutboxRepository(deps.SQLDB)

	stagingOutboxWorkerCfg, err := config.LoadStagingOutboxWorkerConfig()
	if err != nil {
		slog.Error("invalid staging outbox worker config", "err", err)
		return nil, false
	}

	// Enqueue-only River client bundle (built once above): no Queues, no
	// Workers, no PeriodicJobs — this binary only calls InsertTx via the
	// Enqueuers, it never subscribes a queue or executes jobs, mirroring
	// how metaldocs-api builds its own enqueue-only bundle.
	pdfDispatchEnqueuer := dispatchjobs.NewEnqueuer(sharedRiverBundle.Client, pdfOutboxRepo, materializeOutboxRepo, stagingOutboxWorkerCfg.MaxAttempts)

	materializeRunner := workerapp.NewMaterializeJobRunner(
		materializeInvokerAdapter{svc: freezeSvc},
		snapshotFinalDocxAdapter{repo: snapRepo},
		pdfDispatchEnqueuer,
		authzSeamAdapter{},
		deps.SQLDB,
	).WithArtifactFactRecorder(artifactFactRecorder)
	return materializeRunner, true
}

func runWorkerBatch(ctx context.Context, runner workerBatchRunner, batchSize int) error {
	if err := runner.RunOnce(ctx, batchSize); err != nil {
		if ctx.Err() != nil {
			// Shutdown signal arrived mid-batch: not a batch failure.
			return nil //nolint:nilerr // ctx cancellation is a clean exit, not an error
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
		// Intentional cancellation detach (WithoutCancel): a signal arriving
		// mid-batch must not abort it; the outer loop's post-batch select
		// will detect ctx.Done() and exit cleanly after the batch finishes.
		// Context VALUES (trace, baggage) still flow from ctx.
		batchCtx, batchCancel := context.WithTimeout(context.WithoutCancel(ctx), 30*time.Second)
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
