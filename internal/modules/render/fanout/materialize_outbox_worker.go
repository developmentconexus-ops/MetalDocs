package fanout

import (
	"context"
	"log/slog"
	"math"
	"time"

	"github.com/google/uuid"

	"metaldocs/internal/platform/messaging"
)

type materializeOutboxRepoAPI interface {
	ClaimPending(ctx context.Context, limit, maxAttempts int) ([]MaterializeOutboxRow, error)
	MarkDispatched(ctx context.Context, id string) error
	MarkFailed(ctx context.Context, id, errStr string, nextRetryAt time.Time, finalize bool) error
	ResetStaleClaims(ctx context.Context, olderThan time.Duration) (int, error)
}

type MaterializeOutboxWorker struct {
	repo       materializeOutboxRepoAPI
	pub        messaging.Publisher
	pollEvery  time.Duration
	batchSize  int
	maxAttempt int
	staleAfter time.Duration
	log        *slog.Logger
}

func NewMaterializeOutboxWorker(repo materializeOutboxRepoAPI, pub messaging.Publisher, log *slog.Logger) *MaterializeOutboxWorker {
	return &MaterializeOutboxWorker{
		repo: repo, pub: pub,
		pollEvery: 5 * time.Second, batchSize: 10,
		maxAttempt: 5, staleAfter: 5 * time.Minute,
		log: log,
	}
}

func (w *MaterializeOutboxWorker) Run(ctx context.Context) error {
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

func (w *MaterializeOutboxWorker) tick(ctx context.Context) {
	if _, err := w.repo.ResetStaleClaims(ctx, w.staleAfter); err != nil {
		w.log.Warn("materialize outbox reset stale claims", "err", err)
	}
	rows, err := w.repo.ClaimPending(ctx, w.batchSize, w.maxAttempt)
	if err != nil {
		w.log.Warn("materialize outbox claim pending", "err", err)
		return
	}
	for _, r := range rows {
		w.dispatchOne(ctx, r)
	}
}

func (w *MaterializeOutboxWorker) dispatchOne(ctx context.Context, r MaterializeOutboxRow) {
	err := w.pub.Publish(ctx, messaging.Event{
		EventID:        messaging.EventID(uuid.NewString()),
		EventType:      messaging.EventTypeMaterializeFanout,
		AggregateType:  messaging.AggregateType("document_revision"),
		AggregateID:    messaging.AggregateID(r.RevisionID),
		IdempotencyKey: messaging.IdempotencyKey("materialize_fanout:" + r.TenantID + ":" + r.RevisionID),
		Payload: messaging.MaterializeFanoutPayload{
			TenantID:   r.TenantID,
			RevisionID: r.RevisionID,
		},
	})
	if err == nil {
		if mErr := w.repo.MarkDispatched(ctx, r.ID); mErr != nil {
			w.log.Error("materialize outbox mark dispatched", "id", r.ID, "err", mErr)
		}
		return
	}
	finalize := r.Attempts+1 >= w.maxAttempt
	cappedAttempts := min(max(r.Attempts, 0), 30)
	backoff := time.Duration(math.Min(float64(30*time.Minute), float64(time.Duration(1<<cappedAttempts)*30*time.Second)))
	nextRetry := time.Now().Add(backoff)
	if mErr := w.repo.MarkFailed(ctx, r.ID, err.Error(), nextRetry, finalize); mErr != nil {
		w.log.Error("materialize outbox mark failed", "id", r.ID, "err", mErr)
	}
	if finalize {
		w.log.Error("materialize dispatch permanently failed",
			"id", r.ID, "revision_id", r.RevisionID,
			"tenant_id", r.TenantID, "attempts", r.Attempts, "err", err)
	}
}
