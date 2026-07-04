package retention

import (
	"context"
	"log/slog"
	"time"

	"github.com/riverqueue/river"

	"metaldocs/internal/modules/render/fanout"
)

const (
	// RetentionPeriod is how long a dispatched row must sit before it is
	// eligible for purge. Binding per contract §4.1 (M5 F5.4) — changing it
	// is HS-7.
	RetentionPeriod = 7 * 24 * time.Hour
	// BatchSize is the maximum number of rows deleted per batch iteration.
	// Binding per contract §4.1 — changing it is HS-7.
	BatchSize = 5000
	// MaxIterations bounds the number of batch iterations per table per run
	// (so at most BatchSize*MaxIterations rows are purged per table per
	// tick). Binding per contract §4.1 — changing it is HS-7.
	MaxIterations = 10
)

// purger captures the single StagingOutboxRepository method the purge worker
// uses. Kept unexported and minimal so tests can supply a fake without
// standing up a real *fanout.StagingOutboxRepository (whose methods are
// concrete, not behind an interface, at the repo's own package boundary) —
// mirrors dispatchjobs.outboxMarker.
type purger interface {
	PurgeDispatched(ctx context.Context, cutoff time.Time, batchSize, maxIterations int) (int, error)
}

// PurgeWorker is the River worker that purges aged dispatched rows from both
// the pdf and materialize staging outbox tables on each periodic tick.
type PurgeWorker struct {
	river.WorkerDefaults[PurgeArgs]

	pdfRepo         purger
	materializeRepo purger
}

// NewPurgeWorker constructs a PurgeWorker bound to the pdf and materialize
// staging outbox repositories.
func NewPurgeWorker(pdfRepo, materializeRepo *fanout.StagingOutboxRepository) *PurgeWorker {
	return &PurgeWorker{pdfRepo: pdfRepo, materializeRepo: materializeRepo}
}

// Work purges aged dispatched rows from both staging outbox tables. Both
// purges are attempted even if one fails, so a single table's failure never
// starves the other; the first non-nil error is returned so River retries
// the whole job (purge is idempotent — a partial purge is simply re-attempted
// on the next try, since already-deleted rows just won't match the filter
// again).
func (w *PurgeWorker) Work(ctx context.Context, job *river.Job[PurgeArgs]) error {
	cutoff := time.Now().Add(-RetentionPeriod)

	pdfDeleted, pdfErr := w.pdfRepo.PurgeDispatched(ctx, cutoff, BatchSize, MaxIterations)
	if pdfErr != nil {
		slog.ErrorContext(ctx, "staging-outbox-purge: pdf purge failed", "error", pdfErr)
	}

	matDeleted, matErr := w.materializeRepo.PurgeDispatched(ctx, cutoff, BatchSize, MaxIterations)
	if matErr != nil {
		slog.ErrorContext(ctx, "staging-outbox-purge: materialize purge failed", "error", matErr)
	}

	slog.InfoContext(ctx, "staging-outbox-purge: tick complete",
		"pdf_deleted", pdfDeleted,
		"materialize_deleted", matDeleted,
		"cutoff", cutoff,
	)

	if pdfErr != nil {
		return pdfErr
	}
	return matErr
}
