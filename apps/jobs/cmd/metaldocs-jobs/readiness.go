package main

import (
	"context"
	"errors"
	"sync/atomic"

	"metaldocs/internal/platform/observability"
)

// jobsReadiness tracks whether this binary's River client is actually
// running. It is the A7.1 readiness truth for metaldocs-jobs: per ADR 0067
// (dual-define), metaldocs-jobs is the ONLY binary that subscribes the
// "maintenance" queue and executes its periodic jobs — metaldocs-api merely
// enqueues them when it wins leader election. Readiness must therefore
// reflect this binary's own River client run state, not "the process is
// running": before Client.Start succeeds (or once Client.Stop begins), this
// binary cannot execute a single job, maintenance or temporal.
//
// run() calls MarkStarted() immediately after deps.River.Client.Start(ctx)
// returns nil, and MarkStopped() the moment the shutdown signal fires
// (ctx.Done()), before Client.Stop drains.
type jobsReadiness struct {
	started atomic.Bool
}

// MarkStarted flips readiness to healthy. Call once, right after the River
// client has started successfully.
func (r *jobsReadiness) MarkStarted() { r.started.Store(true) }

// MarkStopped flips readiness back to unhealthy. Call when a shutdown
// signal is received, before the River client stops.
func (r *jobsReadiness) MarkStopped() { r.started.Store(false) }

// Check implements the observability.DependencyCheck.Check contract: it
// reports the River client's actual run state.
func (r *jobsReadiness) Check(context.Context) (observability.DependencyCheckResult, error) {
	if !r.started.Load() {
		return observability.DependencyCheckResult{}, errors.New("river client has not started")
	}
	return observability.DependencyCheckResult{Status: "up"}, nil
}
