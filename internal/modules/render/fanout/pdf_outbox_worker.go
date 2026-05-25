package fanout

import (
	"context"
	"encoding/hex"
	"log/slog"
	"math"
	"time"

	"github.com/google/uuid"

	"metaldocs/internal/platform/messaging"
)

type outboxRepoAPI interface {
	ClaimPending(ctx context.Context, limit, maxAttempts int) ([]OutboxRow, error)
	MarkDispatched(ctx context.Context, id string) error
	MarkFailed(ctx context.Context, id, errStr string, nextRetryAt time.Time, finalize bool) error
	ResetStaleClaims(ctx context.Context, olderThan time.Duration) (int, error)
}

type PDFOutboxWorker struct {
	repo       outboxRepoAPI
	pub        messaging.Publisher
	pollEvery  time.Duration
	batchSize  int
	maxAttempt int
	staleAfter time.Duration
	log        *slog.Logger
}

func NewPDFOutboxWorker(repo outboxRepoAPI, pub messaging.Publisher, log *slog.Logger) *PDFOutboxWorker {
	return &PDFOutboxWorker{
		repo: repo, pub: pub,
		pollEvery: 5 * time.Second, batchSize: 10,
		maxAttempt: 5, staleAfter: 5 * time.Minute,
		log: log,
	}
}

func (w *PDFOutboxWorker) Run(ctx context.Context) error {
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

func (w *PDFOutboxWorker) tick(ctx context.Context) {
	if _, err := w.repo.ResetStaleClaims(ctx, w.staleAfter); err != nil {
		w.log.Warn("reset stale claims", "err", err)
	}
	rows, err := w.repo.ClaimPending(ctx, w.batchSize, w.maxAttempt)
	if err != nil {
		w.log.Warn("claim pending", "err", err)
		return
	}
	for _, r := range rows {
		w.dispatchOne(ctx, r)
	}
}

func (w *PDFOutboxWorker) dispatchOne(ctx context.Context, r OutboxRow) {
	err := w.pub.Publish(ctx, messaging.Event{
		EventID:        messaging.EventID(uuid.NewString()),
		EventType:      messaging.EventTypePDFConvert,
		AggregateType:  messaging.AggregateType("document_revision"),
		AggregateID:    messaging.AggregateID(r.RevisionID),
		IdempotencyKey: messaging.IdempotencyKey("docgen_v2_pdf:" + r.RevisionID),
		Payload: messaging.PDFConvertPayload{
			TenantID:    r.TenantID,
			RevisionID:  r.RevisionID,
			ContentHash: hex.EncodeToString(r.ContentHash),
		},
	})
	if err == nil {
		if mErr := w.repo.MarkDispatched(ctx, r.ID); mErr != nil {
			w.log.Error("mark dispatched", "id", r.ID, "err", mErr)
		}
		return
	}
	finalize := r.Attempts+1 >= w.maxAttempt
	backoff := time.Duration(math.Min(float64(30*time.Minute), float64(time.Duration(1<<r.Attempts)*30*time.Second)))
	nextRetry := time.Now().Add(backoff)
	if mErr := w.repo.MarkFailed(ctx, r.ID, err.Error(), nextRetry, finalize); mErr != nil {
		w.log.Error("mark failed", "id", r.ID, "err", mErr)
	}
	if finalize {
		w.log.Error("pdf dispatch permanently failed", "id", r.ID, "revision_id", r.RevisionID, "err", err)
	}
}
