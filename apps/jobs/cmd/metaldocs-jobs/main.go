package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/riverqueue/river"

	"metaldocs/internal/composition/tenantdata/registry"
	approvalapp "metaldocs/internal/modules/approval/application"
	approvalrepo "metaldocs/internal/modules/approval/infrastructure"
	approvaljobs "metaldocs/internal/modules/approval/jobs"
	auditpg "metaldocs/internal/modules/audit/infrastructure/postgres"
	documentsrepo "metaldocs/internal/modules/documents/infrastructure"
	iamapp "metaldocs/internal/modules/iam/application"
	iampg "metaldocs/internal/modules/iam/infrastructure/postgres"
	iamjobs "metaldocs/internal/modules/iam/jobs"
	"metaldocs/internal/modules/jobs/approval_sla_surfacer"
	"metaldocs/internal/modules/jobs/audit_integrity_validator"
	"metaldocs/internal/modules/jobs/document_review_surfacer"
	"metaldocs/internal/modules/jobs/idempotency_janitor"
	"metaldocs/internal/modules/jobs/maintenance"
	"metaldocs/internal/modules/jobs/outbox_retention"
	"metaldocs/internal/modules/jobs/release_hold_reconciler"
	"metaldocs/internal/modules/jobs/stuck_instance_watchdog"
	notificationsinfra "metaldocs/internal/modules/notifications/infrastructure"
	"metaldocs/internal/modules/render/fanout"
	"metaldocs/internal/modules/render/fanout/dispatchjobs"
	"metaldocs/internal/modules/render/fanout/retention"
	securityapp "metaldocs/internal/modules/security/application"
	securitydomain "metaldocs/internal/modules/security/domain"
	securitypg "metaldocs/internal/modules/security/infrastructure/postgres"
	taxonomyrepo "metaldocs/internal/modules/taxonomy/infrastructure"
	"metaldocs/internal/platform/bootstrap"
	"metaldocs/internal/platform/config"
	platformdb "metaldocs/internal/platform/db"
	outboxpg "metaldocs/internal/platform/messaging/outbox/postgres"
	"metaldocs/internal/platform/objectstore"
	"metaldocs/internal/platform/observability"
)

// auditPayloadCryptoAdapter adapts security's published TenantCrypto port to
// the audit module's narrow auditpg.PayloadCrypto port — same composition-root
// adapter as apps/api (audit never imports security): ErrKeyNotFound /
// ErrKeyDestroyed map to the (_, encrypted=false, nil) fall-through-to-
// plaintext contract; any other error propagates unchanged.
type auditPayloadCryptoAdapter struct {
	crypto securitydomain.TenantCrypto
}

func (a auditPayloadCryptoAdapter) EncryptForTenant(ctx context.Context, tenantID string, plaintext []byte) (string, bool, error) {
	envelope, err := a.crypto.EncryptForTenant(ctx, tenantID, plaintext)
	if err != nil {
		if errors.Is(err, securitydomain.ErrKeyNotFound) || errors.Is(err, securitydomain.ErrKeyDestroyed) {
			return "", false, nil
		}
		return "", false, err
	}
	return envelope, true, nil
}

// EncryptForTenantTx is the tx-aware variant RecordTx uses whenever it holds
// a live *sql.Tx (always, in practice) — same-tx key-visibility fix mirrored
// from apps/api's twin adapter (see that file's comment for the full
// rationale): delegates to TenantCrypto.EncryptForTenantTx so the DEK
// lookup reads through the SAME transaction as the audit INSERT.
func (a auditPayloadCryptoAdapter) EncryptForTenantTx(ctx context.Context, tx *sql.Tx, tenantID string, plaintext []byte) (string, bool, error) {
	envelope, err := a.crypto.EncryptForTenantTx(ctx, tx, tenantID, plaintext)
	if err != nil {
		if errors.Is(err, securitydomain.ErrKeyNotFound) || errors.Is(err, securitydomain.ErrKeyDestroyed) {
			return "", false, nil
		}
		return "", false, err
	}
	return envelope, true, nil
}

func (a auditPayloadCryptoAdapter) DecryptForTenant(ctx context.Context, tenantID, envelope string) ([]byte, error) {
	return a.crypto.DecryptForTenant(ctx, tenantID, envelope)
}

func run(ctx context.Context) error {
	jobsCfg, err := config.LoadJobsConfig()
	if err != nil {
		return fmt.Errorf("invalid jobs config: %w", err)
	}
	if !jobsCfg.Enabled {
		slog.Info("MetalDocs Jobs disabled by configuration")
		return nil
	}
	if jobsCfg.Queues == nil {
		jobsCfg.Queues = map[string]river.QueueConfig{}
	}
	// metaldocs-jobs is the only binary that subscribes the maintenance queue
	// and registers the janitor Workers below (ADR 0067 dual-define,
	// jobs-only execute topology); metaldocs-api only enqueues-when-leader.
	jobsCfg.Queues["maintenance"] = river.QueueConfig{MaxWorkers: 2}

	// ADR 0085: the release coordinator re-enqueues its own effective-date
	// timer through River, but the client does not exist until
	// BuildJobsDependencies returns. The deferred enqueuer (and the lifecycle
	// enqueuer) are bound to it below, before the client starts working jobs.
	releaseEnqueuer := approvaljobs.NewDeferredReleaseEvaluationEnqueuer()
	var releaseCoordinator *approvalapp.ReleaseCoordinator
	// F8/Task 7: the SLA surfacer's overdue notification enqueuer has the same
	// chicken-and-egg wiring problem as releaseEnqueuer above — bound after
	// the River client exists, below.
	slaNotifier := approvaljobs.NewDeferredApprovalNotificationEnqueuer()

	deps, err := bootstrap.BuildJobsDependencies(ctx, jobsCfg, func(db *sql.DB) (*river.Workers, []*river.PeriodicJob, error) {
		// No worker in this binary calls LoadActorDisplayName today, but we pass
		// the real reader so the binary stays correct if one ever does.
		displayNameRepo := iampg.NewUserDisplayNameRepository(db)
		repo := approvalrepo.NewPostgresApprovalRepository(db, displayNameRepo)
		approvalEmitter := approvalapp.NewSQLEmitter()
		workers := river.NewWorkers()
		// ADR 0085 release coordinator: the single writer of the released
		// state. Every trigger (approval fact, artifact fact, effective-date
		// timer) funnels into its idempotent Evaluate through this worker.
		releaseCoordinator = approvalapp.NewReleaseCoordinator(repo, approvalEmitter, approvalapp.RealClock{}).
			WithEvaluationEnqueuer(releaseEnqueuer).
			WithProfileReviewIntervalReader(approvalrepo.NewProfileReviewIntervalReader(taxonomyrepo.NewProfileRepository(db)))
		river.AddWorker(workers, approvaljobs.NewReleaseEvaluateWorker(releaseCoordinator, db))
		river.AddWorker(workers, notificationsinfra.NewNotificationsFanoutWorker(db))
		river.AddWorker(workers, notificationsinfra.NewApprovalNotifyWorker(db))
		river.AddWorker(workers, stuck_instance_watchdog.NewWorker(db, approvalEmitter))
		river.AddWorker(workers, idempotency_janitor.NewWorker(db))
		river.AddWorker(workers, audit_integrity_validator.NewWorker(auditpg.NewWriter(db)))
		river.AddWorker(workers, document_review_surfacer.NewWorker(db,
			documentsrepo.NewReviewDueReaderPG(db),
			documentsrepo.NewReviewSurfaceWriterPG(db)))
		// F8 (approval-kernel-backend): approval stage SLA surfacer — a
		// genuine sibling to document_review_surfacer, not an extension of
		// it (distinct per-stage due_at clock vs per-document
		// review_due_at cadence; see approval_sla_surfacer package docs).
		river.AddWorker(workers, approval_sla_surfacer.NewWorker(db,
			approvalrepo.NewSLAOverdueReaderPG(db),
			approvalrepo.NewSLASurfaceWriterPG(db),
			slaNotifier))
		// ADR 0085 Stage C W2: reconciliation sweep over release generations
		// stuck in readiness hold (lost fact, dead-lettered consumer, dead
		// timer). Alert-only (ADR 0068) — it reads through the approval
		// module's ReleaseHoldReader port and emits governance alerts; it never
		// mutates a generation and never re-enqueues an evaluation.
		river.AddWorker(workers, release_hold_reconciler.NewWorker(db,
			approvalrepo.NewReleaseHoldReaderPG(db),
			approvalEmitter))

		// Staging pdf/materialize dispatch workers (M5 F5.3 T3): consume the
		// River jobs the api/worker Enqueuers insert and run on the already-
		// subscribed "temporal" queue, publishing the corresponding domain
		// event and marking the paired staging outbox row dispatched.
		publisher := outboxpg.NewPublisher(db)
		pdfRepo := fanout.NewPDFOutboxRepository(db)
		matRepo := fanout.NewMaterializeOutboxRepository(db)
		river.AddWorker(workers, dispatchjobs.NewPDFDispatchWorker(publisher, pdfRepo))
		river.AddWorker(workers, dispatchjobs.NewMaterializeDispatchWorker(publisher, matRepo))

		// Staging outbox retention purge (M5 F5.4 T2): reuses the same pdfRepo/
		// matRepo instances built above for the dispatch workers.
		river.AddWorker(workers, retention.NewPurgeWorker(pdfRepo, matRepo))

		// Relay outbox retention purge: the second stage of the same chain.
		// The staging purge above bounds pdf_dispatch_outbox /
		// materialize_dispatch_outbox; this one bounds metaldocs.outbox_events,
		// which had no retention at all and so pinned every terminal row's
		// idempotency_key against the publisher's ON CONFLICT dedup forever.
		river.AddWorker(workers, outbox_retention.NewWorker(outboxpg.NewRetention(db)))

		// M7 F7.3 Task E: tenant lifecycle (export/erase) worker. Consumes
		// iamdomain.TenantLifecycleJobArgs jobs the api binary's
		// TenantLifecycleService.RequestExport/RequestErase enqueue (paired
		// with the tenant_lifecycle_jobs row insert, same tx — see
		// internal/modules/iam/jobs/tenant_lifecycle_enqueuer.go). Runs on the
		// already-subscribed "temporal" queue, same as the staging dispatch
		// workers above.
		tenantLifecycleWorker, err := buildTenantLifecycleWorker(db)
		if err != nil {
			return nil, nil, fmt.Errorf("build tenant lifecycle worker: %w", err)
		}
		river.AddWorker(workers, tenantLifecycleWorker)

		periodicJobs := append(maintenance.PeriodicJobs(), retention.PeriodicJob())
		return workers, periodicJobs, nil
	})
	if err != nil {
		return fmt.Errorf("build jobs dependencies: %w", err)
	}
	defer deps.Cleanup()
	releaseEnqueuer.Bind(deps.River.Client)
	releaseCoordinator.WithLifecycleEnqueuer(approvaljobs.NewLifecycleEventEnqueuer(deps.River.Client))
	slaNotifier.Bind(deps.River.Client)

	// A7.1: liveness/readiness/metrics on a dedicated infra-port listener.
	// Built and served before Client.Start so an orchestrator can observe
	// this process during bootstrap too (readiness correctly reports NOT
	// ready until Start succeeds below). A bind/serve failure here is
	// logged, not fatal — it must not block this binary's actual job
	// (executing periodic + temporal River jobs), matching the roadmap's
	// "additive endpoints, no behavior change" acceptance bound for A7.1.
	jobsReady := &jobsReadiness{}
	// F1: wire the River queue-report heartbeat (see readiness.go's type doc
	// comment for the full derivation and stated limitation) across every
	// queue THIS binary subscribes — both "maintenance" (jobsCfg.Queues,
	// added above) and whatever else jobsCfg.Queues carries ("temporal" by
	// default, see config.LoadJobsConfig). staleAfter = 2x River's own fixed
	// ~10-minute queue-report cadence + a 1-minute margin: a small multiple
	// plus margin, derived from that cadence rather than an unrelated fixed
	// value, the same discipline the worker binary applies to its own poll
	// interval.
	jobsReady.ConfigureHeartbeat(deps.River.Client, jobsQueueNames(jobsCfg), 2*riverQueueReportInterval+time.Minute, nil)
	infraServer, err := buildInfraServer(deps, jobsReady)
	if err != nil {
		return fmt.Errorf("invalid jobs infra server config: %w", err)
	}
	go func() {
		if err := infraServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("jobs infra server failed", "err", err)
		}
	}()

	slog.Info("MetalDocs Jobs running", "queues", "temporal", "infra_addr", infraServer.Addr)
	if err := deps.River.Client.Start(ctx); err != nil {
		return fmt.Errorf("run jobs host: %w", err)
	}
	jobsReady.MarkStarted()

	// F3 (review round 2): confirmed no equivalent gap exists here to the
	// worker binary's batch-drain window — MarkStopped() below runs
	// synchronously the instant ctx.Done() unblocks, BEFORE
	// deps.River.Client.Stop's up-to-15s drain begins, so /ready already
	// reports NOT ready for the entire duration of that drain. No goroutine
	// restructuring is needed to close this finding for metaldocs-jobs.
	<-ctx.Done()
	jobsReady.MarkStopped()

	// Intentional cancellation detach (WithoutCancel): ctx is already done at
	// this point — the shutdown needs its own fresh 15 s deadline while still
	// carrying ctx's values (trace, baggage).
	shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 15*time.Second)
	defer cancel()

	if err := deps.River.Client.Stop(shutdownCtx); err != nil && !errors.Is(err, context.DeadlineExceeded) && !errors.Is(err, context.Canceled) {
		return fmt.Errorf("stop jobs host: %w", err)
	}

	infraShutdownCtx, infraCancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer infraCancel()
	if err := infraServer.Shutdown(infraShutdownCtx); err != nil {
		slog.Warn("jobs infra server shutdown incomplete", "err", err)
	}

	return nil
}

// buildTenantLifecycleWorker composes the M7 F7.3 tenant lifecycle worker:
// VerifiedStore (for the export artifact write / blob deletion during
// erasure) + TenantCrypto (crypto-shred, nil-safe when METALDOCS_TENANT_KEK
// is unset — same F7.2/F7.3 established pattern as apps/api's tenantCrypto
// wiring) + registry.AllTenantDataPorts(db) (every module's erase/export
// fan-out port). The worker's run-side service needs no TxRunner/tenant
// lookup/River-enqueuer dependencies (those are enqueue-side-only, wired in
// apps/api) — nil is passed for those four constructor args since RunJob
// never calls RequestExport/RequestErase.
func buildTenantLifecycleWorker(db *sql.DB) (*iamjobs.TenantLifecycleWorker, error) {
	var tenantCrypto securitydomain.TenantCrypto
	kek, kekConfigured, err := config.LoadTenantKEK()
	if err != nil {
		return nil, fmt.Errorf("invalid tenant crypto KEK: %w", err)
	}
	if kekConfigured {
		svc, err := securityapp.NewTenantCryptoService(securitypg.NewTenantKeyRepository(db), kek)
		if err != nil {
			return nil, fmt.Errorf("construct tenant crypto service: %w", err)
		}
		tenantCrypto = svc
	} else {
		slog.Info("tenant crypto disabled: METALDOCS_TENANT_KEK not set (tenant erasure will not crypto-shred)")
	}

	// Same audit payload-envelope wiring as apps/api: without it any
	// worker-side audit event for an active tenant would silently skip
	// sealing while the api binary seals. tenant.erased itself is written
	// after key destruction, so the adapter's ErrKeyDestroyed fall-through
	// keeps that tombstone plaintext-survivable either way.
	auditWriter := auditpg.NewWriter(db)
	if tenantCrypto != nil {
		auditWriter.WithPayloadCrypto(auditPayloadCryptoAdapter{crypto: tenantCrypto})
	}

	var blobs *objectstore.VerifiedStore
	attachmentsCfg, err := config.LoadAttachmentsConfig()
	if err != nil {
		return nil, fmt.Errorf("invalid attachments config: %w", err)
	}
	if attachmentsCfg.Provider == config.StorageProviderMinIO {
		minioClient, minioPublicClient, minioBucket, err := bootstrap.BuildMinioClients(attachmentsCfg)
		if err != nil {
			return nil, fmt.Errorf("build minio clients: %w", err)
		}
		blobs = objectstore.NewVerifiedStore(minioClient, minioPublicClient, minioBucket, 25*1024*1024)
	} else {
		slog.Info("tenant lifecycle export disabled: attachments provider is not minio")
	}

	svc := iamapp.NewTenantLifecycleService(
		nil, // tenantLookupTx: enqueue-side only, wired in apps/api
		nil, // lifecycleJobInserter: enqueue-side only, wired in apps/api
		iampg.NewTenantLifecycleRepository(db),
		nil, // iamdomain.TenantLifecycleEnqueuer: enqueue-side only, wired in apps/api
		// TxRunner is NOT enqueue-side-only: runErase's phase-1 (row erase)
		// and phase-3 (key-destroy + tombstone) txs run through it. A nil
		// here panics on the first erase job (caught by
		// TestTenantErasure_ChainStaysGreen).
		platformdb.NewTxRunner(db),
		auditWriter,
		db,
		registry.AllTenantDataPorts(db),
		blobs,
		tenantCrypto,
	)
	return iamjobs.NewTenantLifecycleWorker(svc), nil
}

func main() {
	os.Exit(runMain())
}

// runMain holds main's body so deferred cleanups (signal stop, otel shutdown)
// run on every exit path; os.Exit in main proper would skip them.
func runMain() int {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// OpenTelemetry: inert unless an exporter is configured (Z-1, REQ-OBS-3).
	// otelShutdown is a no-op when disabled; otelEnabled gates the chain link.
	otelShutdown, otelEnabled, err := observability.SetupOTel(ctx, "metaldocs-jobs")
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

	if err := run(ctx); err != nil {
		slog.Error("jobs exited with error", "err", err)
		return 1
	}
	return 0
}
