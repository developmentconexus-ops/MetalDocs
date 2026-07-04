package fanout

import (
	"context"
	"encoding/hex"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/google/uuid"

	"metaldocs/internal/platform/config"
	"metaldocs/internal/platform/messaging"
)

// fakeOutboxRepo implements stagingOutboxRepoAPI for testing.
type fakeOutboxRepo struct {
	rows              []OutboxRow
	dispatchedIDs     []string
	failedIDs         []string
	finalizedIDs      []string
	markDispatchedErr error // if non-nil, MarkDispatched returns this error
}

func (f *fakeOutboxRepo) ClaimPending(_ context.Context, _, _ int) ([]OutboxRow, error) {
	return f.rows, nil
}
func (f *fakeOutboxRepo) MarkDispatched(_ context.Context, _, id string) error {
	if f.markDispatchedErr != nil {
		return f.markDispatchedErr
	}
	f.dispatchedIDs = append(f.dispatchedIDs, id)
	return nil
}
func (f *fakeOutboxRepo) MarkFailed(_ context.Context, _, id, _ string, _ time.Time, finalize bool) error {
	f.failedIDs = append(f.failedIDs, id)
	if finalize {
		f.finalizedIDs = append(f.finalizedIDs, id)
	}
	return nil
}
func (f *fakeOutboxRepo) ResetStaleClaims(_ context.Context, _ time.Duration) (int, error) {
	return 0, nil
}

// fakeWorkerPublisher satisfies messaging.Publisher.
type fakeWorkerPublisher struct {
	err    error
	events []messaging.Event
}

func (p *fakeWorkerPublisher) Publish(_ context.Context, e messaging.Event) error {
	if p.err != nil {
		return p.err
	}
	p.events = append(p.events, e)
	return nil
}

// defaultTestWorkerCfg returns a StagingOutboxWorkerConfig with the same
// defaults as the production hardcoded values, suitable for unit tests.
func defaultTestWorkerCfg() config.StagingOutboxWorkerConfig {
	return config.StagingOutboxWorkerConfig{
		PollIntervalSeconds: 5,
		BatchSize:           10,
		MaxAttempts:         5,
		StaleAfterSeconds:   300,
	}
}

// buildPDFEvent is the canonical PDF buildEvent closure (mirrors main.go wiring).
func buildPDFEvent(r OutboxRow) messaging.Event {
	return messaging.Event{
		EventID:        messaging.EventID(uuid.NewString()),
		EventType:      messaging.EventTypePDFConvert,
		AggregateType:  messaging.AggregateType("document_revision"),
		AggregateID:    messaging.AggregateID(r.RevisionID),
		IdempotencyKey: messaging.IdempotencyKey("docgen_v2_pdf:" + r.TenantID + ":" + r.RevisionID),
		Payload: messaging.PDFConvertPayload{
			TenantID:    r.TenantID,
			RevisionID:  r.RevisionID,
			ContentHash: hex.EncodeToString(r.ContentHash),
		},
	}
}

// buildMaterializeEvent is the canonical materialize buildEvent closure (mirrors main.go wiring).
func buildMaterializeEvent(r OutboxRow) messaging.Event {
	return messaging.Event{
		EventID:        messaging.EventID(uuid.NewString()),
		EventType:      messaging.EventTypeMaterializeFanout,
		AggregateType:  messaging.AggregateType("document_revision"),
		AggregateID:    messaging.AggregateID(r.RevisionID),
		IdempotencyKey: messaging.IdempotencyKey("materialize_fanout:" + r.TenantID + ":" + r.RevisionID),
		Payload: messaging.MaterializeFanoutPayload{
			TenantID:   r.TenantID,
			RevisionID: r.RevisionID,
		},
	}
}

func TestPDFOutboxWorker_PublishSuccessMarksDispatched(t *testing.T) {
	repo := &fakeOutboxRepo{rows: []OutboxRow{{ID: "id-1", TenantID: "t1", RevisionID: "r1"}}}
	pub := &fakeWorkerPublisher{err: nil}
	w := NewStagingOutboxWorker(repo, pub, buildPDFEvent, defaultTestWorkerCfg(), slog.Default())
	w.tick(context.Background())
	if len(repo.dispatchedIDs) != 1 || repo.dispatchedIDs[0] != "id-1" {
		t.Fatalf("want id-1 dispatched, got %v", repo.dispatchedIDs)
	}
	if got, want := pub.events[0].IdempotencyKey, messaging.IdempotencyKey("docgen_v2_pdf:t1:r1"); got != want {
		t.Fatalf("IdempotencyKey = %q, want %q", got, want)
	}
}

func TestPDFOutboxWorker_PublishFailIncrementsAttempts(t *testing.T) {
	repo := &fakeOutboxRepo{rows: []OutboxRow{{ID: "id-2", TenantID: "t1", RevisionID: "r1", Attempts: 0}}}
	pub := &fakeWorkerPublisher{err: errors.New("bus down")}
	w := NewStagingOutboxWorker(repo, pub, buildPDFEvent, defaultTestWorkerCfg(), slog.Default())
	w.tick(context.Background())
	if len(repo.failedIDs) != 1 {
		t.Fatalf("want failure recorded, got %v", repo.failedIDs)
	}
	if len(repo.finalizedIDs) != 0 {
		t.Fatalf("should not finalize on attempt 0")
	}
}

func TestPDFOutboxWorker_MaxAttemptsMarksFinal(t *testing.T) {
	repo := &fakeOutboxRepo{rows: []OutboxRow{{ID: "id-3", TenantID: "t1", RevisionID: "r1", Attempts: 4}}}
	pub := &fakeWorkerPublisher{err: errors.New("bus down")}
	w := NewStagingOutboxWorker(repo, pub, buildPDFEvent, defaultTestWorkerCfg(), slog.Default())
	w.tick(context.Background())
	if len(repo.finalizedIDs) != 1 || repo.finalizedIDs[0] != "id-3" {
		t.Fatalf("want id-3 finalized, got %v", repo.finalizedIDs)
	}
}

func TestPDFOutboxWorker_StopOnContext(t *testing.T) {
	repo := &fakeOutboxRepo{}
	pub := &fakeWorkerPublisher{}
	w := NewStagingOutboxWorker(repo, pub, buildPDFEvent, defaultTestWorkerCfg(), slog.Default())
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- w.Run(ctx) }()
	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Run returned error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Run did not stop after context cancel")
	}
}

func TestMaterializeOutboxWorker_PublishSuccessMarksDispatched(t *testing.T) {
	repo := &fakeOutboxRepo{rows: []OutboxRow{{ID: "id-m1", TenantID: "t1", RevisionID: "r1"}}}
	pub := &fakeWorkerPublisher{err: nil}
	w := NewStagingOutboxWorker(repo, pub, buildMaterializeEvent, defaultTestWorkerCfg(), slog.Default())
	w.tick(context.Background())
	if len(repo.dispatchedIDs) != 1 || repo.dispatchedIDs[0] != "id-m1" {
		t.Fatalf("want id-m1 dispatched, got %v", repo.dispatchedIDs)
	}
	if got, want := pub.events[0].IdempotencyKey, messaging.IdempotencyKey("materialize_fanout:t1:r1"); got != want {
		t.Fatalf("IdempotencyKey = %q, want %q", got, want)
	}
}

// TestPDFOutboxWorker_MarkDispatchedFailFallsBackToMarkFailed covers F-R4:
// when Publish succeeds but MarkDispatched fails, MarkFailed must be called so
// the row is retried rather than stuck forever in 'processing' state.
func TestPDFOutboxWorker_MarkDispatchedFailFallsBackToMarkFailed(t *testing.T) {
	repo := &fakeOutboxRepo{
		rows:              []OutboxRow{{ID: "id-fr4", TenantID: "t1", RevisionID: "r1", Attempts: 0}},
		markDispatchedErr: errors.New("db write failed"),
	}
	pub := &fakeWorkerPublisher{err: nil} // Publish succeeds
	w := NewStagingOutboxWorker(repo, pub, buildPDFEvent, defaultTestWorkerCfg(), slog.Default())
	w.tick(context.Background())
	// Publish succeeded so dispatchedIDs must be empty (MarkDispatched returned err).
	if len(repo.dispatchedIDs) != 0 {
		t.Fatalf("MarkDispatched must not record success on error; got %v", repo.dispatchedIDs)
	}
	// MarkFailed must have been called as the fallback.
	if len(repo.failedIDs) != 1 || repo.failedIDs[0] != "id-fr4" {
		t.Fatalf("MarkFailed must be called after MarkDispatched error; failedIDs=%v", repo.failedIDs)
	}
}

func TestStagingOutboxWorker_RunNeverReturnsError(t *testing.T) {
	// Run() only returns nil; the restart loop in main.go was dead and has been removed.
	repo := &fakeOutboxRepo{}
	pub := &fakeWorkerPublisher{}
	w := NewStagingOutboxWorker(repo, pub, buildPDFEvent, defaultTestWorkerCfg(), slog.Default())
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := w.Run(ctx); err != nil {
		t.Fatalf("Run must return nil, got %v", err)
	}
}
