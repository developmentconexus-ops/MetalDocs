package jobs

import (
	"context"
	"database/sql"

	"github.com/riverqueue/river"

	"metaldocs/internal/modules/documents/approval/application"
)

type ScheduledPublishWorker struct {
	river.WorkerDefaults[ScheduledPublishArgs]

	Service *application.SchedulerService
	DB      *sql.DB
}

func (w *ScheduledPublishWorker) Work(ctx context.Context, job *river.Job[ScheduledPublishArgs]) error {
	return w.Service.RunScheduledPublishJob(ctx, w.DB, application.ScheduledPublishJobInput{
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

func (e *RiverScheduledPublishEnqueuer) EnqueueScheduledPublishTx(ctx context.Context, tx *sql.Tx, input application.ScheduledPublishJobInput) error {
	_, err := e.Client.InsertTx(ctx, tx, ScheduledPublishArgs{
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
	river.AddWorker(workers, &ScheduledPublishWorker{
		Service: scheduler,
		DB:      db,
	})
	return workers
}

func NewScheduledPublishEnqueuer(client *river.Client[*sql.Tx]) application.ScheduledPublishEnqueuer {
	return &RiverScheduledPublishEnqueuer{Client: client}
}
