package jobs

import (
	"context"
	"database/sql"
	"errors"

	"github.com/riverqueue/river"

	"metaldocs/internal/modules/documents/approval/application"
)

var (
	ErrScheduledPublishJobNil     = errors.New("scheduled publish job is nil")
	ErrScheduledPublishWorkerDeps = errors.New("scheduled publish worker dependencies not configured")
)

type ScheduledPublishWorker struct {
	river.WorkerDefaults[ScheduledPublishArgs]

	service *application.SchedulerService
	db      *sql.DB
}

func NewScheduledPublishWorker(service *application.SchedulerService, db *sql.DB) *ScheduledPublishWorker {
	return &ScheduledPublishWorker{
		service: service,
		db:      db,
	}
}

func (w *ScheduledPublishWorker) Work(ctx context.Context, job *river.Job[ScheduledPublishArgs]) error {
	if w == nil || w.service == nil || w.db == nil {
		return ErrScheduledPublishWorkerDeps
	}
	if job == nil {
		return ErrScheduledPublishJobNil
	}
	return w.service.RunScheduledPublishJob(ctx, w.db, application.ScheduledPublishJobInput{
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
	river.AddWorker(workers, NewScheduledPublishWorker(scheduler, db))
	return workers
}

func NewScheduledPublishEnqueuer(client *river.Client[*sql.Tx]) application.ScheduledPublishEnqueuer {
	return &RiverScheduledPublishEnqueuer{Client: client}
}
