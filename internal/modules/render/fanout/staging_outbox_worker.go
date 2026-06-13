package fanout

import (
	"context"
	"log/slog"
	"math"
	"time"

	"metaldocs/internal/platform/messaging"
)

type stagingOutboxRepoAPI interface {
	ClaimPending(ctx context.Context, limit, maxAttempts int) ([]OutboxRow, error)
	MarkDispatched(ctx context.Context, id string) error
	MarkFailed(ctx context.Context, id, errStr string, nextRetryAt time.Time, finalize bool) error
	ResetStaleClaims(ctx context.Context, olderThan time.Duration) (int, error)
}

// StagingOutboxWorker is a generic poll-and-dispatch worker for staging outbox tables.
// Callers supply buildEvent to construct the event envelope from an OutboxRow;
// this is the only per-table difference between PDF and materialize workers.
type StagingOutboxWorker struct {
	repo       stagingOutboxRepoAPI
	pub        messaging.Publisher
	buildEvent func(OutboxRow) messaging.Event
	pollEvery  time.Duration
	batchSize  int
	maxAttempt int
	staleAfter time.Duration
	log        *slog.Logger
}

// NewStagingOutboxWorker constructs a worker.
// buildEvent constructs the messaging.Event for a given row; it encapsulates
// the per-table event type, idempotency-key prefix, and payload shape.
func NewStagingOutboxWorker(
	repo stagingOutboxRepoAPI,
	pub messaging.Publisher,
	buildEvent func(OutboxRow) messaging.Event,
	log *slog.Logger,
) *StagingOutboxWorker {
	return &StagingOutboxWorker{
		repo: repo, pub: pub, buildEvent: buildEvent,
		pollEvery: 5 * time.Second, batchSize: 10,
		maxAttempt: 5, staleAfter: 5 * time.Minute,
		log: log,
	}
}

// Run polls the outbox until ctx is cancelled. Returns nil when ctx is done;
// the loop inside Run() never returns a non-nil error.
func (w *StagingOutboxWorker) Run(ctx context.Context) error {
	t := time.NewTicker(w.pollEvery)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-t.C:
			w.tick(ctx)
		}
	}
}

func (w *StagingOutboxWorker) tick(ctx context.Context) {
	if _, err := w.repo.ResetStaleClaims(ctx, w.staleAfter); err != nil {
		w.log.Warn("staging outbox reset stale claims", "err", err)
	}
	rows, err := w.repo.ClaimPending(ctx, w.batchSize, w.maxAttempt)
	if err != nil {
		w.log.Warn("staging outbox claim pending", "err", err)
		return
	}
	for _, r := range rows {
		w.dispatchOne(ctx, r)
	}
}

func (w *StagingOutboxWorker) dispatchOne(ctx context.Context, r OutboxRow) {
	err := w.pub.Publish(ctx, w.buildEvent(r))
	if err == nil {
		if mErr := w.repo.MarkDispatched(ctx, r.ID); mErr != nil {
			w.log.Error("staging outbox mark dispatched", "id", r.ID, "err", mErr)
		}
		return
	}
	finalize := r.Attempts+1 >= w.maxAttempt
	cappedAttempts := min(max(r.Attempts, 0), 30)
	backoff := time.Duration(math.Min(float64(30*time.Minute), float64(time.Duration(1<<cappedAttempts)*30*time.Second)))
	nextRetry := time.Now().Add(backoff)
	if mErr := w.repo.MarkFailed(ctx, r.ID, err.Error(), nextRetry, finalize); mErr != nil {
		w.log.Error("staging outbox mark failed", "id", r.ID, "err", mErr)
	}
	if finalize {
		w.log.Error("staging outbox dispatch permanently failed",
			"id", r.ID, "revision_id", r.RevisionID,
			"tenant_id", r.TenantID, "attempts", r.Attempts, "err", err)
	}
}
