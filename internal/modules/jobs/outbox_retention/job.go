// Package outbox_retention holds the River worker that purges terminal rows
// out of metaldocs.outbox_events — the platform relay outbox written by
// internal/platform/messaging/outbox/postgres.Publisher and drained by the
// same package's Consumer (and therefore by the metaldocs-worker binary).
//
// It is a sibling of idempotency_janitor: a jobs-module worker over a platform
// table. It holds no SQL of its own — the delete statements live beside the
// publisher/consumer in internal/platform/messaging/outbox/postgres, which
// remains the sole owner of SQL against outbox_events. It is deliberately NOT
// folded into internal/modules/render/fanout/retention, whose PurgeWorker
// purges the render module's own pdf/materialize STAGING tables; reaching from
// a module package into a platform table would cross the module boundary.
//
// ADR 0067 dual-define: the periodic definition is registered in
// internal/modules/jobs/maintenance.PeriodicJobs() (which both metaldocs-api
// and metaldocs-jobs put on their River client config, so whichever wins
// leader election enqueues the tick), while only metaldocs-jobs subscribes the
// "maintenance" queue and registers the Worker below.
package outbox_retention

import (
	"context"
	"log/slog"
	"time"

	"github.com/riverqueue/river"

	outboxpg "metaldocs/internal/platform/messaging/outbox/postgres"
)

const (
	// JobName is the River job kind for the outbox_events retention tick.
	JobName = "outbox-events-retention"

	// PublishedRetention is how long a successfully published row is kept
	// before it becomes eligible for purge. Matches the 7-day window the
	// staging outbox purge uses (internal/modules/render/fanout/retention),
	// so both stages of the two-stage outbox chain age out together.
	PublishedRetention = 7 * 24 * time.Hour

	// DeadLetterRetention is how long a dead-lettered row is kept. It is an
	// order of magnitude longer than PublishedRetention on purpose: a
	// dead-lettered row is the only durable record that a domain event never
	// reached its consumer, so it must survive a full incident-review cycle.
	// It is not kept forever because the row also pins its idempotency_key
	// against the UNIQUE constraint the publisher dedupes on, which blocks any
	// later legitimate re-publish of that key.
	DeadLetterRetention = 90 * 24 * time.Hour

	// BatchSize is the maximum number of rows deleted per batch iteration.
	BatchSize = 5000

	// MaxIterations bounds the batch iterations per class per run, so at most
	// BatchSize*MaxIterations published rows and the same again in
	// dead-lettered rows are purged per tick.
	MaxIterations = 10
)

// Args is the (empty) River job payload for the retention tick. The job
// carries no per-run parameters; the windows and batch bounds are constants.
type Args struct{}

// Kind implements river.JobArgs, identifying this job type to River.
func (Args) Kind() string { return JobName }

// purger captures the two Retention methods this worker uses, so tests can
// supply a fake without a live database — mirrors
// internal/modules/render/fanout/retention.purger.
type purger interface {
	PurgePublished(ctx context.Context, cutoff time.Time, batchSize, maxIterations int) (int, error)
	PurgeDeadLettered(ctx context.Context, cutoff time.Time, batchSize, maxIterations int) (int, error)
}

// Worker is the River worker that purges terminal outbox_events rows.
type Worker struct {
	river.WorkerDefaults[Args]

	retention purger
}

// NewWorker constructs a Worker bound to the platform outbox retention store.
func NewWorker(retention *outboxpg.Retention) *Worker {
	return &Worker{retention: retention}
}

// Work purges aged published rows and aged dead-lettered rows. Both passes are
// attempted even if one fails, so a failure in one class never starves the
// other; the first non-nil error is returned so River retries the whole job.
// The purge is idempotent — already-deleted rows simply stop matching the
// predicate — so a partial purge is safe to re-attempt.
func (w *Worker) Work(ctx context.Context, job *river.Job[Args]) error {
	now := time.Now()
	publishedCutoff := now.Add(-PublishedRetention)
	deadLetterCutoff := now.Add(-DeadLetterRetention)

	publishedDeleted, publishedErr := w.retention.PurgePublished(ctx, publishedCutoff, BatchSize, MaxIterations)
	if publishedErr != nil {
		slog.ErrorContext(ctx, "outbox-events-retention: published purge failed", "error", publishedErr)
	}

	deadDeleted, deadErr := w.retention.PurgeDeadLettered(ctx, deadLetterCutoff, BatchSize, MaxIterations)
	if deadErr != nil {
		slog.ErrorContext(ctx, "outbox-events-retention: dead-letter purge failed", "error", deadErr)
	}

	slog.InfoContext(ctx, "outbox-events-retention: tick complete",
		"job", JobName,
		"published_deleted", publishedDeleted,
		"dead_lettered_deleted", deadDeleted,
		"published_cutoff", publishedCutoff,
		"dead_letter_cutoff", deadLetterCutoff,
	)

	if publishedErr != nil {
		return publishedErr
	}
	return deadErr
}
