package main

import (
	"context"
	"errors"
	"sync/atomic"

	"metaldocs/internal/platform/observability"
)

// workerReadiness tracks whether the outbox consumer poll loop
// (runWorkerLoop) is actually running. It is the A7.1 readiness truth for
// metaldocs-worker: GET /ready must report NOT ready whenever this binary
// genuinely cannot process outbox events, which for a single-loop poller
// means "the loop has not started yet" (still bootstrapping) or "the loop
// has stopped" (shutdown in progress / a fatal loop exit) — not merely
// "the process is still running".
//
// runMain calls MarkStarted() immediately before entering runWorkerLoop and
// MarkStopped() once runWorkerLoop returns (the shutdown signal has been
// received AND any in-flight batch has finished draining).
type workerReadiness struct {
	started atomic.Bool
}

// MarkStarted flips readiness to healthy. Call once, right before the poll
// loop begins ticking.
func (r *workerReadiness) MarkStarted() { r.started.Store(true) }

// MarkStopped flips readiness back to unhealthy. Call when a shutdown
// signal is received, before draining any in-flight batch.
func (r *workerReadiness) MarkStopped() { r.started.Store(false) }

// Check implements the observability.DependencyCheck.Check contract: it
// reports the outbox consumer loop's actual run state.
func (r *workerReadiness) Check(context.Context) (observability.DependencyCheckResult, error) {
	if !r.started.Load() {
		return observability.DependencyCheckResult{}, errors.New("outbox consumer loop has not started")
	}
	return observability.DependencyCheckResult{Status: "up"}, nil
}
