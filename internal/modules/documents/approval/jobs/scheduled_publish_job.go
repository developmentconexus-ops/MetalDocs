package jobs

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/riverqueue/river"

	"metaldocs/internal/modules/documents/approval/application"
	"metaldocs/internal/modules/iam/authz"
	"metaldocs/internal/platform/db"
)

var (
	ErrScheduledPublishJobNil = errors.New("scheduled publish job is nil")
)

type ScheduledPublishWorker struct {
	river.WorkerDefaults[ScheduledPublishArgs]

	service *application.SchedulerService
	runner  db.TxRunner
}

func NewScheduledPublishWorker(service *application.SchedulerService, database *sql.DB) *ScheduledPublishWorker {
	if service == nil {
		panic("scheduled_publish_worker: service is nil")
	}
	if database == nil {
		panic("scheduled_publish_worker: db is nil")
	}
	return &ScheduledPublishWorker{
		service: service,
		runner:  db.NewTxRunner(database),
	}
}

func (w *ScheduledPublishWorker) Work(ctx context.Context, job *river.Job[ScheduledPublishArgs]) error {
	if job == nil {
		return ErrScheduledPublishJobNil
	}
	// Background root: permit RunScheduledPublishJob's authz.BypassSystem
	// (fail-closed off any HTTP path — ADR 0022 Phase 7, CWE-269).
	ctx = authz.WithBackgroundBypass(ctx)
	return w.service.RunScheduledPublishJob(ctx, w.runner, application.ScheduledPublishJobInput{
		TenantID:                job.Args.TenantID,
		DocumentID:              job.Args.DocumentID,
		ExpectedRevisionVersion: job.Args.ExpectedRevisionVersion,
		ScheduledEffectiveAt:    job.Args.ScheduledEffectiveAt,
		ScheduleGeneration:      job.Args.ScheduleGeneration,
	})
}

type RiverScheduledPublishEnqueuer struct {
	Client *river.Client[*sql.Tx]
}

func (e *RiverScheduledPublishEnqueuer) EnqueueScheduledPublishTx(ctx context.Context, tx db.Tx, input application.ScheduledPublishJobInput) error {
	sqlTx, ok := tx.(*sql.Tx)
	if !ok {
		return fmt.Errorf("scheduled_publish: river requires *sql.Tx, got %T", tx)
	}
	_, err := e.Client.InsertTx(ctx, sqlTx, ScheduledPublishArgs{
		TenantID:                input.TenantID,
		DocumentID:              input.DocumentID,
		ExpectedRevisionVersion: input.ExpectedRevisionVersion,
		ScheduledEffectiveAt:    input.ScheduledEffectiveAt.UTC(),
		ScheduleGeneration:      input.ScheduleGeneration,
	}, &river.InsertOpts{
		Queue:       "temporal",
		ScheduledAt: input.ScheduledEffectiveAt.UTC(),
	})
	return err
}

func NewWorkers(scheduler *application.SchedulerService, db *sql.DB) *river.Workers {
	workers := river.NewWorkers()
	river.AddWorker(workers, NewScheduledPublishWorker(scheduler, db))
	return workers
}

func NewScheduledPublishEnqueuer(client *river.Client[*sql.Tx]) application.ScheduledPublishEnqueuer {
	return &RiverScheduledPublishEnqueuer{Client: client}
}
