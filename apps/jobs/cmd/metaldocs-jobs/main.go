package main

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/riverqueue/river"

	approvalapp "metaldocs/internal/modules/documents/approval/application"
	approvaljobs "metaldocs/internal/modules/documents/approval/jobs"
	approvalrepo "metaldocs/internal/modules/documents/approval/repository"
	"metaldocs/internal/platform/bootstrap"
	"metaldocs/internal/platform/config"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	jobsCfg, err := config.LoadJobsConfig()
	if err != nil {
		log.Fatalf("invalid jobs config: %v", err)
	}
	if !jobsCfg.Enabled {
		log.Printf("MetalDocs Jobs disabled by configuration")
		return
	}

	deps, err := bootstrap.BuildJobsDependencies(ctx, jobsCfg, func(db *sql.DB) (*river.Workers, error) {
		repo := approvalrepo.NewPostgresApprovalRepository(db)
		services := approvalapp.NewServices(repo, approvalapp.NewSQLEmitter(), approvalapp.RealClock{})
		return approvaljobs.NewWorkers(services.Scheduler, db), nil
	})
	if err != nil {
		log.Fatalf("build jobs dependencies: %v", err)
	}
	defer deps.Cleanup()

	log.Printf("MetalDocs Jobs running (queues=temporal)")
	if err := deps.River.Client.Start(ctx); err != nil {
		log.Fatalf("run jobs host: %v", err)
	}

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := deps.River.Client.Stop(shutdownCtx); err != nil && !errors.Is(err, context.DeadlineExceeded) && !errors.Is(err, context.Canceled) {
		log.Fatalf("stop jobs host: %v", err)
	}
}
