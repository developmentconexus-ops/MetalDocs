package main

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"metaldocs/internal/platform/observability"
)

// workerReadiness tracks whether the outbox consumer poll loop
// (runWorkerLoop) is actually running AND making genuine progress. It is the
// A7.1 readiness truth for metaldocs-worker: GET /ready must report NOT
// ready whenever this binary genuinely cannot process outbox events.
//
// Two independent conditions gate readiness:
//
//  1. Start/stop latch (started): "the loop has not started yet" (still
//     bootstrapping) or "the loop has stopped" (shutdown in progress / a
//     fatal loop exit).
//  2. Progress heartbeat (lastProgressAt/staleAfter): review round 1 (F1)
//     correctly flagged that (1) alone only proves the loop *began*, not
//     that it is still doing anything — a hung RunOnce with a reachable
//     Postgres would leave /ready reporting 200 forever. RecordProgress
//     must be called ONLY from a call site that observed a genuinely
//     completed poll cycle (runWorkerLoop, on RunOnce returning nil) —
//     never unconditionally, never on a timer, and never from Check itself.
//     A heartbeat written from any of those places would be a synthetic
//     always-fresh signal that looks like progress detection while
//     detecting nothing.
//
// runMain calls MarkStarted() immediately before entering runWorkerLoop
// (which also seeds the heartbeat so a slow first cycle does not read as
// instantly stale) and MarkStopped() both the moment the shutdown signal
// arrives AND once runWorkerLoop returns (F3: the first call, from a
// goroutine watching ctx.Done(), flips readiness during the up-to-30s
// batch-drain window instead of only after it; the second is an idempotent
// fallback for the no-drain-in-flight case).
type workerReadiness struct {
	started atomic.Bool

	mu             sync.Mutex
	lastProgressAt time.Time
	staleAfter     time.Duration
	clock          func() time.Time
}

// workerHeartbeatFallbackStaleAfter is used ONLY when ConfigureHeartbeat was
// never called (e.g. a unit test exercising just the start/stop latch via a
// bare &workerReadiness{}). Production wiring (runMain) always calls
// ConfigureHeartbeat with a threshold derived from workerCfg.PollIntervalSeconds,
// so this constant never governs a real deployment; it exists so an
// unconfigured heartbeat still fails closed eventually instead of never
// going stale.
const workerHeartbeatFallbackStaleAfter = 5 * time.Minute

// MarkStarted flips readiness to healthy and seeds the progress heartbeat
// (startup grace): a process that has just started has not completed a poll
// cycle yet, but it should not read as instantly stale before it gets the
// chance to. Call once, right before the poll loop begins ticking.
func (r *workerReadiness) MarkStarted() {
	r.started.Store(true)
	r.recordProgressAt(r.clockFunc()())
}

// MarkStopped flips readiness back to unhealthy. Safe to call from multiple
// goroutines and multiple times (idempotent) — F3 calls it both from a
// goroutine watching ctx.Done() and, as a fallback, after the poll loop
// returns; atomic.Bool.Store needs no additional synchronization for that.
func (r *workerReadiness) MarkStopped() { r.started.Store(false) }

// ConfigureHeartbeat sets the staleness threshold and clock this readiness
// check uses to judge poll-cycle freshness. Call once, before MarkStarted,
// with staleAfter derived from the binary's own configured poll interval
// (see buildInfraServer) — never a fixed duration unrelated to that
// configuration, or the check silently goes wrong the moment someone changes
// the poll interval. clock may be nil to use time.Now (tests inject a fake
// clock to drive staleness deterministically, without sleeping).
func (r *workerReadiness) ConfigureHeartbeat(staleAfter time.Duration, clock func() time.Time) {
	r.mu.Lock()
	r.staleAfter = staleAfter
	r.clock = clock
	r.mu.Unlock()
}

// RecordProgress marks that a poll cycle genuinely completed successfully
// (RunOnce returned nil). See the type doc comment: this must only be
// called from a site that observed real forward progress.
func (r *workerReadiness) RecordProgress() {
	r.recordProgressAt(r.clockFunc()())
}

func (r *workerReadiness) recordProgressAt(t time.Time) {
	r.mu.Lock()
	r.lastProgressAt = t
	r.mu.Unlock()
}

// clockFunc returns the configured clock, or time.Now when unconfigured.
func (r *workerReadiness) clockFunc() func() time.Time {
	r.mu.Lock()
	clock := r.clock
	r.mu.Unlock()
	if clock != nil {
		return clock
	}
	return time.Now
}

// Check implements the observability.DependencyCheck.Check contract: it
// reports the outbox consumer loop's actual run state AND progress
// freshness.
func (r *workerReadiness) Check(context.Context) (observability.DependencyCheckResult, error) {
	if !r.started.Load() {
		return observability.DependencyCheckResult{}, errors.New("outbox consumer loop has not started")
	}

	r.mu.Lock()
	last := r.lastProgressAt
	staleAfter := r.staleAfter
	r.mu.Unlock()
	if staleAfter <= 0 {
		staleAfter = workerHeartbeatFallbackStaleAfter
	}

	age := r.clockFunc()().Sub(last)
	if age > staleAfter {
		return observability.DependencyCheckResult{}, fmt.Errorf(
			"outbox consumer poll loop stalled: last successful cycle %s ago (threshold %s)",
			age.Round(time.Second), staleAfter,
		)
	}
	return observability.DependencyCheckResult{Status: "up"}, nil
}
