// Package jobs wires the approval subsystem's River-delivered async work
// (scheduled-publish cutover, lifecycle-event enqueue) using the transactional
// outbox pattern: jobs are enqueued via InsertTx in the same transaction as the
// state write that scheduled them, and workers run under a background-bypass
// authz context (ADR 0022 Phase 7) since they have no HTTP-request identity.
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
	// ErrScheduledPublishJobNil is returned by Work when River delivers a nil job.
	ErrScheduledPublishJobNil = errors.New("scheduled publish job is nil")
)

// ScheduledPublishWorker is the River worker that runs SchedulerService.RunScheduledPublishJob
// for each delivered ScheduledPublishArgs job.
type ScheduledPublishWorker struct {
	river.WorkerDefaults[ScheduledPublishArgs]

	service *application.SchedulerService
	runner  db.TxRunner
}

// NewScheduledPublishWorker constructs a ScheduledPublishWorker. Panics if
// service or database is nil — both are required for the worker to run.
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

// Work runs the scheduled-publish cutover for one delivered job under a
// background-bypass authz context (no HTTP-request identity exists here).
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

// RiverScheduledPublishEnqueuer implements application.ScheduledPublishEnqueuer
// using River's same-tx InsertTx.
type RiverScheduledPublishEnqueuer struct {
	Client *river.Client[*sql.Tx]
}

// EnqueueScheduledPublishTx enqueues a ScheduledPublishArgs job within tx (outbox
// pattern), scheduled to run at input.ScheduledEffectiveAt. tx must be a *sql.Tx.
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

// NewWorkers returns the River worker set for the approval subsystem's jobs
// (currently just ScheduledPublishWorker), ready to register with a river.Client.
func NewWorkers(scheduler *application.SchedulerService, db *sql.DB) *river.Workers {
	workers := river.NewWorkers()
	river.AddWorker(workers, NewScheduledPublishWorker(scheduler, db))
	return workers
}

// NewScheduledPublishEnqueuer wraps a River client as an application.ScheduledPublishEnqueuer.
func NewScheduledPublishEnqueuer(client *river.Client[*sql.Tx]) application.ScheduledPublishEnqueuer {
	return &RiverScheduledPublishEnqueuer{Client: client}
}
