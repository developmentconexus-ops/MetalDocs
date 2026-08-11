package main

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

type captureWorkerRunner struct {
	calls int
	ctxs  []context.Context
	err   error
}

func (r *captureWorkerRunner) RunOnce(ctx context.Context, _ int) error {
	r.calls++
	r.ctxs = append(r.ctxs, ctx)
	return r.err
}

func TestRunWorkerBatchPassesCancellationContext(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	runner := &captureWorkerRunner{err: errors.New("context canceled")}

	if err := runWorkerBatch(ctx, runner, 5); err != nil {
		t.Fatalf("runWorkerBatch: %v", err)
	}

	if runner.calls != 1 {
		t.Fatalf("calls = %d, want 1", runner.calls)
	}
	if runner.ctxs[0] != ctx {
		t.Fatal("runner did not receive the caller context")
	}
}

func TestRunWorkerBatchReturnsFailureInRunOnceMode(t *testing.T) {
	t.Parallel()

	runner := &captureWorkerRunner{err: errors.New("boom")}

	err := runWorkerBatch(context.Background(), runner, 5)
	if err == nil || err.Error() != "boom" {
		t.Fatalf("expected runWorkerBatch to return boom, got %v", err)
	}
}

func TestRunWorkerLoopStopsAfterCancellation(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	ticks := make(chan time.Time)
	runner := &captureWorkerRunner{}
	done := make(chan struct{})

	go func() {
		runWorkerLoop(ctx, runner, 3, ticks, nil)
		close(done)
	}()

	cancel()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("runWorkerLoop did not stop after cancellation")
	}

	if runner.calls != 1 {
		t.Fatalf("calls = %d, want 1", runner.calls)
	}
	// Graceful drain: the in-flight batch runs under a detached, bounded
	// context (not the cancellable caller ctx) so a mid-batch signal does not
	// abort it. The loop still exits after the batch via the post-batch select.
	if runner.ctxs[0] == ctx {
		t.Fatal("batch should run under a detached drain context, not the caller ctx")
	}
	if _, ok := runner.ctxs[0].Deadline(); !ok {
		t.Fatal("batch context should carry a bounded drain deadline")
	}
}

// blockingWorkerRunner blocks in RunOnce until release is closed, signaling
// started once it is actually inside the call — used to simulate an
// in-flight batch that has NOT yet observed the outer context's
// cancellation (runWorkerLoop deliberately detaches the batch context from
// ctx via context.WithoutCancel, so a blocked batch keeps blocking across a
// shutdown signal; see runWorkerLoop's comment).
type blockingWorkerRunner struct {
	started chan struct{}
	release chan struct{}
	once    sync.Once
}

func (r *blockingWorkerRunner) RunOnce(context.Context, int) error {
	r.once.Do(func() { close(r.started) })
	<-r.release
	return nil
}

// TestRunWorkerLifecycle_MarksStoppedImmediatelyOnCancelDuringBlockedBatch is
// the F1/F3 review round 2 RED-first proof: readiness.MarkStopped() must
// fire the instant ctx is cancelled, not only after runWorkerLoop returns
// (which can be up to 30s later, waiting on an in-flight batch to drain).
// This drives readiness to NOT-ready while a batch is still genuinely
// blocked, proving the fix independent of the F1 heartbeat threshold.
func TestRunWorkerLifecycle_MarksStoppedImmediatelyOnCancelDuringBlockedBatch(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	runner := &blockingWorkerRunner{started: make(chan struct{}), release: make(chan struct{})}
	readiness := &workerReadiness{}
	ticks := make(chan time.Time)
	done := make(chan struct{})

	go func() {
		runWorkerLifecycle(ctx, readiness, runner, 1, ticks)
		close(done)
	}()

	select {
	case <-runner.started:
	case <-time.After(2 * time.Second):
		t.Fatal("batch never started")
	}

	cancel()

	// The batch is still blocked (release not yet closed); readiness must
	// flip to not-ready promptly regardless.
	deadline := time.Now().Add(2 * time.Second)
	for {
		if _, err := readiness.Check(context.Background()); err != nil {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("readiness did not flip to not-ready promptly on ctx cancellation while the batch was still blocked")
		}
		time.Sleep(5 * time.Millisecond)
	}

	close(runner.release)

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("runWorkerLifecycle did not exit after the blocked batch was released")
	}

	if _, err := readiness.Check(context.Background()); err == nil {
		t.Fatal("readiness should remain not-ready after the loop exits (post-loop MarkStopped fallback)")
	}
}
