package main

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/riverqueue/river/rivertype"

	"metaldocs/internal/platform/observability"
)

// jobsReadiness tracks whether this binary's River client is actually
// running AND making genuine progress. It is the A7.1 readiness truth for
// metaldocs-jobs: per ADR 0067 (dual-define), metaldocs-jobs is the ONLY
// binary that subscribes the "maintenance" queue and executes its periodic
// jobs — metaldocs-api merely enqueues them when it wins leader election.
//
// Two independent conditions gate readiness:
//
//  1. Start/stop latch (started): before Client.Start succeeds (or once
//     Client.Stop begins), this binary cannot execute a single job at all.
//  2. Queue-report heartbeat (queueSource/queues/staleAfter): review round 1
//     (F1) correctly flagged that (1) alone only proves the River client
//     *started*, not that its internal fetch/report loops are still alive.
//
// F1's "harder case" for jobs: River's periodic jobs range from every 5
// minutes (stuck-instance-watchdog) to every 24 hours (outbox retention
// purge), and several "temporal" queue jobs are purely event-driven with no
// schedule at all — so "no job has completed recently" is routinely normal
// and cannot be used as a staleness signal without either flapping on idle
// periods or being too coarse (24h+) to mean anything operationally.
//
// Instead this uses River's OWN internal health surface: river.Client.
// QueueGet(name).UpdatedAt. Every producer synchronously stamps this at
// Client.Start (see river's producer.go StartWorkContext -> initial
// QueueCreateOrSetUpdatedAt) and then periodically re-stamps it via its own
// internal report loop for as long as that queue's producer goroutines are
// alive and can reach Postgres — regardless of whether any job is queued or
// which process currently holds periodic-job leadership. That makes it a
// genuine, externally-persisted progress signal scoped to exactly what this
// binary can honestly claim ("my own job-fetching machinery is alive"),
// rather than a completed-job signal borrowed from a schedule this binary
// does not control.
//
// LIMITATION (stated plainly, per the review instruction): River's public
// Client API does not expose the finer-grained (~1 minute) internal
// producer keepalive; the only publicly readable heartbeat is the queue
// report, which River re-stamps on a fixed, non-configurable ~10 minute
// cadence (producer.go's queueReportIntervalDefault; river.Config has no
// field to override it). So this heartbeat's staleness window is measured
// in tens of minutes, not seconds — coarser than the worker's poll-based
// heartbeat. That is accepted as the honest ceiling of what River's API
// exposes, rather than fabricating a tighter threshold from job-completion
// events that would either flap on legitimate idle periods or need a 24h+
// window to avoid doing so.
//
// run() calls MarkStarted() immediately after deps.River.Client.Start(ctx)
// returns nil (at which point every subscribed queue's UpdatedAt is already
// fresh, per the synchronous stamp above — no separate startup-grace seed is
// needed), and MarkStopped() the moment the shutdown signal fires
// (ctx.Done()), before Client.Stop drains — confirmed below (F3) that this
// already happens with no gap, unlike the worker binary's batch-drain
// window.
type jobsReadiness struct {
	started atomic.Bool

	mu          sync.Mutex
	queueSource queueUpdatedAtSource
	queues      []string
	staleAfter  time.Duration
	clock       func() time.Time
}

// queueUpdatedAtSource is the subset of *river.Client[*sql.Tx] jobsReadiness
// needs — QueueGet — kept as a narrow interface so tests can inject a fake
// without a live Postgres + River schema.
type queueUpdatedAtSource interface {
	QueueGet(ctx context.Context, name string) (*rivertype.Queue, error)
}

// jobsHeartbeatFallbackStaleAfter is used ONLY when ConfigureHeartbeat was
// never called (e.g. a unit test exercising just the start/stop latch via a
// bare &jobsReadiness{}). Production wiring (run()) always calls
// ConfigureHeartbeat, so this constant never governs a real deployment.
const jobsHeartbeatFallbackStaleAfter = 5 * time.Minute

// MarkStarted flips readiness to healthy. Call once, right after the River
// client has started successfully.
func (r *jobsReadiness) MarkStarted() { r.started.Store(true) }

// MarkStopped flips readiness back to unhealthy. Call when a shutdown
// signal is received, before the River client stops. Safe to call more than
// once (idempotent, a single atomic.Bool.Store).
func (r *jobsReadiness) MarkStopped() { r.started.Store(false) }

// ConfigureHeartbeat wires the River queue-report heartbeat: source reads
// river_queue.updated_at for each name in queues (this binary's subscribed
// queues — "maintenance" and "temporal"), staleAfter is the freshness
// threshold (see buildInfraServer for its derivation from River's own fixed
// report cadence), and clock may be nil to use time.Now (tests inject a
// fake clock). Call once, before MarkStarted. Leaving this unconfigured
// (queueSource nil) is tolerated so existing latch-only tests keep working —
// production wiring always configures it.
func (r *jobsReadiness) ConfigureHeartbeat(source queueUpdatedAtSource, queues []string, staleAfter time.Duration, clock func() time.Time) {
	r.mu.Lock()
	r.queueSource = source
	r.queues = queues
	r.staleAfter = staleAfter
	r.clock = clock
	r.mu.Unlock()
}

func (r *jobsReadiness) clockFunc() func() time.Time {
	r.mu.Lock()
	clock := r.clock
	r.mu.Unlock()
	if clock != nil {
		return clock
	}
	return time.Now
}

// Check implements the observability.DependencyCheck.Check contract: it
// reports the River client's actual run state AND, when the heartbeat is
// configured, the freshness of River's own queue-report heartbeat across
// every queue this binary subscribes.
func (r *jobsReadiness) Check(ctx context.Context) (observability.DependencyCheckResult, error) {
	if !r.started.Load() {
		return observability.DependencyCheckResult{}, errors.New("river client has not started")
	}

	r.mu.Lock()
	source := r.queueSource
	queues := r.queues
	staleAfter := r.staleAfter
	r.mu.Unlock()

	if source == nil || len(queues) == 0 {
		// Heartbeat not wired — only true for tests exercising the latch in
		// isolation (see the type doc comment). Production always wires it.
		return observability.DependencyCheckResult{Status: "up"}, nil
	}
	if staleAfter <= 0 {
		staleAfter = jobsHeartbeatFallbackStaleAfter
	}

	// F6 (review round 3): must be the OLDEST UpdatedAt across queues, not
	// the newest. Readiness here means "every subscribed queue's producer
	// is alive", per this type's own doc comment above ("the freshness of
	// River's own queue-report heartbeat across every queue this binary
	// subscribes") — taking the newest lets one live queue mask another
	// that has stopped reporting entirely, which is exactly the
	// "readiness tied to process existence, not progress" failure this
	// slice exists to eliminate, just one level down (per-queue instead of
	// per-process). A QueueGet returning a nil queue (no error — River's
	// public API returns this for a name that was never created/
	// subscribed) fails closed explicitly rather than being silently
	// skipped, which would reintroduce the same masking bug for a queue
	// this binary declares it subscribes but River has no row for.
	var (
		oldest    time.Time
		oldestSet bool
		staleName string
	)
	for _, name := range queues {
		q, err := source.QueueGet(ctx, name)
		if err != nil {
			return observability.DependencyCheckResult{}, fmt.Errorf("river queue %q: %w", name, err)
		}
		if q == nil {
			return observability.DependencyCheckResult{}, fmt.Errorf("river queue %q: QueueGet returned no queue (not created/subscribed?)", name)
		}
		if !oldestSet || q.UpdatedAt.Before(oldest) {
			oldest = q.UpdatedAt
			oldestSet = true
			staleName = name
		}
	}

	age := r.clockFunc()().Sub(oldest)
	if age > staleAfter {
		return observability.DependencyCheckResult{}, fmt.Errorf(
			"river queue report heartbeat stale: queue %q last reported %s ago across queues %v (threshold %s)",
			staleName, age.Round(time.Second), queues, staleAfter,
		)
	}
	return observability.DependencyCheckResult{Status: "up"}, nil
}
