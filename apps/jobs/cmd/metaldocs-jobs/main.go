package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/riverqueue/river"

	auditpg "metaldocs/internal/modules/audit/infrastructure/postgres"
	approvalapp "metaldocs/internal/modules/documents/approval/application"
	approvaljobs "metaldocs/internal/modules/documents/approval/jobs"
	cdinfra "metaldocs/internal/modules/controlleddocuments/infrastructure"
	approvalrepo "metaldocs/internal/modules/documents/approval/repository"
	iampg "metaldocs/internal/modules/iam/infrastructure/postgres"
	"metaldocs/internal/modules/jobs/audit_integrity_validator"
	"metaldocs/internal/modules/jobs/idempotency_janitor"
	"metaldocs/internal/modules/jobs/maintenance"
	"metaldocs/internal/modules/jobs/stuck_instance_watchdog"
	notificationsinfra "metaldocs/internal/modules/notifications/infrastructure"
	"metaldocs/internal/modules/render/fanout"
	"metaldocs/internal/modules/render/fanout/dispatchjobs"
	"metaldocs/internal/modules/render/fanout/retention"
	"metaldocs/internal/platform/bootstrap"
	"metaldocs/internal/platform/config"
	outboxpg "metaldocs/internal/platform/messaging/outbox/postgres"
	"metaldocs/internal/platform/observability"
)

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

	deps, err := bootstrap.BuildJobsDependencies(ctx, jobsCfg, func(db *sql.DB) (*river.Workers, []*river.PeriodicJob, error) {
		// The scheduled-publish job never calls LoadActorDisplayName, but we pass
		// the real reader so the binary is correct if the code path ever is reached.
		displayNameRepo := iampg.NewUserDisplayNameRepository(db)
		repo := approvalrepo.NewPostgresApprovalRepository(db, displayNameRepo)
		approvalEmitter := approvalapp.NewSQLEmitter()
		services := approvalapp.NewServices(repo, approvalEmitter, approvalapp.RealClock{}, cdinfra.NewCDFieldReaderPG())
		workers := approvaljobs.NewWorkers(services.Scheduler, db)
		river.AddWorker(workers, notificationsinfra.NewNotificationsFanoutWorker(db))
		river.AddWorker(workers, stuck_instance_watchdog.NewWorker(db, services.Cancel, approvalEmitter))
		river.AddWorker(workers, idempotency_janitor.NewWorker(db))
		river.AddWorker(workers, audit_integrity_validator.NewWorker(auditpg.NewWriter(db)))

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

		periodicJobs := append(maintenance.PeriodicJobs(), retention.PeriodicJob())
		return workers, periodicJobs, nil
	})
	if err != nil {
		return fmt.Errorf("build jobs dependencies: %w", err)
	}
	defer deps.Cleanup()

	slog.Info("MetalDocs Jobs running", "queues", "temporal")
	if err := deps.River.Client.Start(ctx); err != nil {
		return fmt.Errorf("run jobs host: %w", err)
	}

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := deps.River.Client.Stop(shutdownCtx); err != nil && !errors.Is(err, context.DeadlineExceeded) && !errors.Is(err, context.Canceled) {
		return fmt.Errorf("stop jobs host: %w", err)
	}

	return nil
}

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// OpenTelemetry: inert unless an exporter is configured (Z-1, REQ-OBS-3).
	// otelShutdown is a no-op when disabled; otelEnabled gates the chain link.
	otelShutdown, otelEnabled, err := observability.SetupOTel(ctx, "metaldocs-jobs")
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

	if err := run(ctx); err != nil {
		slog.Error("jobs exited with error", "err", err)
		os.Exit(1)
	}
}
