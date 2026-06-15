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

	approvalapp "metaldocs/internal/modules/documents/approval/application"
	approvaljobs "metaldocs/internal/modules/documents/approval/jobs"
	approvalrepo "metaldocs/internal/modules/documents/approval/repository"
	iampg "metaldocs/internal/modules/iam/infrastructure/postgres"
	"metaldocs/internal/platform/bootstrap"
	"metaldocs/internal/platform/config"
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

	deps, err := bootstrap.BuildJobsDependencies(ctx, jobsCfg, func(db *sql.DB) (*river.Workers, error) {
		// The scheduled-publish job never calls LoadActorDisplayName, but we pass
		// the real reader so the binary is correct if the code path ever is reached.
		displayNameRepo := iampg.NewUserDisplayNameRepository(db)
		repo := approvalrepo.NewPostgresApprovalRepository(db, displayNameRepo)
		services := approvalapp.NewServices(repo, approvalapp.NewSQLEmitter(), approvalapp.RealClock{})
		return approvaljobs.NewWorkers(services.Scheduler, db), nil
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
