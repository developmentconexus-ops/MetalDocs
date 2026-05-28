package main

import (
	"context"
	"errors"
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
		runWorkerLoop(ctx, runner, 3, ticks)
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
	if runner.ctxs[0] != ctx {
		t.Fatal("runner did not receive the caller context")
	}
}
