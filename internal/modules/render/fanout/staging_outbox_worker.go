package fanout

import (
	"context"
	"log/slog"
	"math"
	"time"

	"metaldocs/internal/platform/config"
	"metaldocs/internal/platform/messaging"
)

// stagingOutboxRepoAPI's MarkDispatched/MarkFailed take tenantID so the
// implementation can seed authz.SeedTxTenant(ctx, tx, tenantID) in its own
// processing tx before the write (M3 F3.2 — validation-contract.md §2.2 site
// 5: the claim step stays GUC-unset per ADR 0054 rule 1, but every per-row
// processing write must be scoped to that row's tenant).
type stagingOutboxRepoAPI interface {
	ClaimPending(ctx context.Context, limit, maxAttempts int) ([]OutboxRow, error)
	MarkDispatched(ctx context.Context, tenantID, id string) error
	MarkFailed(ctx context.Context, tenantID, id, errStr string, nextRetryAt time.Time, finalize bool) error
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

// NewStagingOutboxWorker constructs a worker from an externally-loaded config.
// buildEvent constructs the messaging.Event for a given row; it encapsulates
// the per-table event type, idempotency-key prefix, and payload shape.
func NewStagingOutboxWorker(
	repo stagingOutboxRepoAPI,
	pub messaging.Publisher,
	buildEvent func(OutboxRow) messaging.Event,
	cfg config.StagingOutboxWorkerConfig,
	log *slog.Logger,
) *StagingOutboxWorker {
	return &StagingOutboxWorker{
		repo:       repo,
		pub:        pub,
		buildEvent: buildEvent,
		pollEvery:  time.Duration(cfg.PollIntervalSeconds) * time.Second,
		batchSize:  cfg.BatchSize,
		maxAttempt: cfg.MaxAttempts,
		staleAfter: time.Duration(cfg.StaleAfterSeconds) * time.Second,
		log:        log,
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
		// F-R4: if MarkDispatched fails, fall through to MarkFailed so the row
		// is retried rather than silently abandoned in 'processing' state.
		if mErr := w.repo.MarkDispatched(ctx, r.TenantID, r.ID); mErr != nil {
			w.log.Error("staging outbox mark dispatched", "id", r.ID, "err", mErr)
			finalize := r.Attempts+1 >= w.maxAttempt
			cappedAttempts := min(max(r.Attempts, 0), 30)
			backoff := time.Duration(math.Min(float64(30*time.Minute), float64(time.Duration(1<<cappedAttempts)*30*time.Second)))
			nextRetry := time.Now().Add(backoff)
			if fErr := w.repo.MarkFailed(ctx, r.TenantID, r.ID, mErr.Error(), nextRetry, finalize); fErr != nil {
				w.log.Error("staging outbox mark failed after dispatch error", "id", r.ID, "err", fErr)
			}
		}
		return
	}
	finalize := r.Attempts+1 >= w.maxAttempt
	cappedAttempts := min(max(r.Attempts, 0), 30)
	backoff := time.Duration(math.Min(float64(30*time.Minute), float64(time.Duration(1<<cappedAttempts)*30*time.Second)))
	nextRetry := time.Now().Add(backoff)
	if mErr := w.repo.MarkFailed(ctx, r.TenantID, r.ID, err.Error(), nextRetry, finalize); mErr != nil {
		w.log.Error("staging outbox mark failed", "id", r.ID, "err", mErr)
	}
	if finalize {
		w.log.Error("staging outbox dispatch permanently failed",
			"id", r.ID, "revision_id", r.RevisionID,
			"tenant_id", r.TenantID, "attempts", r.Attempts, "err", err)
	}
}
