package fanout

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"metaldocs/internal/platform/messaging"
)

// fakeOutboxRepo implements outboxRepoAPI for testing.
type fakeOutboxRepo struct {
	rows          []OutboxRow
	dispatchedIDs []string
	failedIDs     []string
	finalizedIDs  []string
}

func (f *fakeOutboxRepo) ClaimPending(_ context.Context, _, _ int) ([]OutboxRow, error) {
	return f.rows, nil
}
func (f *fakeOutboxRepo) MarkDispatched(_ context.Context, id string) error {
	f.dispatchedIDs = append(f.dispatchedIDs, id)
	return nil
}
func (f *fakeOutboxRepo) MarkFailed(_ context.Context, id, _ string, _ time.Time, finalize bool) error {
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
type fakeWorkerPublisher struct{ err error }

func (p *fakeWorkerPublisher) Publish(_ context.Context, _ messaging.Event) error {
	return p.err
}

func TestPDFOutboxWorker_PublishSuccessMarksDispatched(t *testing.T) {
	repo := &fakeOutboxRepo{rows: []OutboxRow{{ID: "id-1", TenantID: "t1", RevisionID: "r1"}}}
	pub := &fakeWorkerPublisher{err: nil}
	w := NewPDFOutboxWorker(repo, pub, slog.Default())
	w.tick(context.Background())
	if len(repo.dispatchedIDs) != 1 || repo.dispatchedIDs[0] != "id-1" {
		t.Fatalf("want id-1 dispatched, got %v", repo.dispatchedIDs)
	}
}

func TestPDFOutboxWorker_PublishFailIncrementsAttempts(t *testing.T) {
	repo := &fakeOutboxRepo{rows: []OutboxRow{{ID: "id-2", TenantID: "t1", RevisionID: "r1", Attempts: 0}}}
	pub := &fakeWorkerPublisher{err: errors.New("bus down")}
	w := NewPDFOutboxWorker(repo, pub, slog.Default())
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
	w := NewPDFOutboxWorker(repo, pub, slog.Default())
	w.tick(context.Background())
	if len(repo.finalizedIDs) != 1 || repo.finalizedIDs[0] != "id-3" {
		t.Fatalf("want id-3 finalized, got %v", repo.finalizedIDs)
	}
}

func TestPDFOutboxWorker_StopOnContext(t *testing.T) {
	repo := &fakeOutboxRepo{}
	pub := &fakeWorkerPublisher{}
	w := NewPDFOutboxWorker(repo, pub, slog.Default())
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
