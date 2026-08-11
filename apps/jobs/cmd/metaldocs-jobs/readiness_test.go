package main

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/riverqueue/river/rivertype"

	"metaldocs/internal/platform/observability"
)

var errQueueGetBoom = errors.New("boom: queue get failed")

// fakeJobsClock is an injectable, manually-advanced clock so heartbeat
// staleness tests drive time deterministically instead of sleeping.
type fakeJobsClock struct {
	mu sync.Mutex
	t  time.Time
}

func (c *fakeJobsClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.t
}

func (c *fakeJobsClock) Advance(d time.Duration) {
	c.mu.Lock()
	c.t = c.t.Add(d)
	c.mu.Unlock()
}

// fakeQueueSource is a test double for queueUpdatedAtSource: no live
// Postgres or River schema, just a map of queue name -> UpdatedAt the test
// controls directly.
type fakeQueueSource struct {
	mu   sync.Mutex
	rows map[string]time.Time
	err  error
}

func (f *fakeQueueSource) QueueGet(_ context.Context, name string) (*rivertype.Queue, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.err != nil {
		return nil, f.err
	}
	return &rivertype.Queue{Name: name, UpdatedAt: f.rows[name]}, nil
}

func (f *fakeQueueSource) setUpdatedAt(name string, t time.Time) {
	f.mu.Lock()
	f.rows[name] = t
	f.mu.Unlock()
}

// TestJobsReadiness_NotReadyBeforeRiverStarts is the A7.1 RED-first proof
// for metaldocs-jobs: readiness must report NOT ready before the River
// client has started — this is the concrete, testable form of "the
// maintenance queue is not subscribed yet" (ADR 0067 dual-define: only
// metaldocs-jobs subscribes "maintenance" and executes those periodic
// jobs; before Client.Start succeeds this binary cannot run any job at
// all, not just the maintenance set).
func TestJobsReadiness_NotReadyBeforeRiverStarts(t *testing.T) {
	r := &jobsReadiness{}

	_, err := r.Check(context.Background())
	if err == nil {
		t.Fatal("Check() before MarkStarted() = nil error, want an error (River client not started)")
	}
}

// TestJobsReadiness_ReadyAfterRiverStarts proves the same object reports
// healthy once River has genuinely started.
func TestJobsReadiness_ReadyAfterRiverStarts(t *testing.T) {
	r := &jobsReadiness{}
	r.MarkStarted()

	result, err := r.Check(context.Background())
	if err != nil {
		t.Fatalf("Check() after MarkStarted() = %v, want nil error", err)
	}
	if result.Status != "up" {
		t.Fatalf("Status = %q, want up", result.Status)
	}
}

// TestJobsReadiness_NotReadyAfterStop proves readiness flips back to NOT
// ready the moment shutdown begins — an orchestrator must not keep treating
// this process as able to execute periodic/temporal jobs once the River
// client is stopping/stopped.
func TestJobsReadiness_NotReadyAfterStop(t *testing.T) {
	r := &jobsReadiness{}
	r.MarkStarted()
	r.MarkStopped()

	_, err := r.Check(context.Background())
	if err == nil {
		t.Fatal("Check() after MarkStopped() = nil error, want an error (River client stopped)")
	}
}

// TestJobsReadinessEndpoint_ReflectsRiverState drives the full stack
// (jobsReadiness -> observability.DependencyCheck ->
// PostgresRuntimeStatusProvider -> NewInfraServer) exactly as main.go wires
// it, over real HTTP via httptest: GET /ready must be drivable to 503 by
// the same condition ("River client not started") the binary would hit on
// a genuine bootstrap failure or mid-shutdown.
func TestJobsReadinessEndpoint_ReflectsRiverState(t *testing.T) {
	r := &jobsReadiness{}
	provider := observability.NewPostgresRuntimeStatusProvider(nil, "postgres", "n/a", false,
		observability.DependencyCheck{Name: "river_client", Check: r.Check})
	server := observability.NewInfraServer(":0", provider, nil)

	req := httptest.NewRequest(http.MethodGet, "/ready", nil)
	rec := httptest.NewRecorder()
	server.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("GET /ready with nil DB and River not started = %d, want 503", rec.Code)
	}
}

// TestJobsReadiness_StaleQueueHeartbeatFlipsNotReady is the F1 review round 2
// RED-first proof for metaldocs-jobs: once the River client has started,
// readiness must ALSO go stale if River's own queue-report heartbeat
// (river_queue.updated_at, read via QueueGet) has not advanced within the
// configured threshold — the start latch alone only proves Client.Start
// once succeeded, not that the client's internal fetch/report loops are
// still alive. The clock is injected and advanced manually; no sleeping.
func TestJobsReadiness_StaleQueueHeartbeatFlipsNotReady(t *testing.T) {
	clock := &fakeJobsClock{t: time.Unix(1_700_000_000, 0)}
	src := &fakeQueueSource{rows: map[string]time.Time{
		"maintenance": clock.Now(),
		"temporal":    clock.Now(),
	}}
	r := &jobsReadiness{}
	r.ConfigureHeartbeat(src, []string{"maintenance", "temporal"}, 5*time.Minute, clock.Now)
	r.MarkStarted()

	if _, err := r.Check(context.Background()); err != nil {
		t.Fatalf("Check() immediately after MarkStarted() = %v, want nil error", err)
	}

	clock.Advance(6 * time.Minute) // no queue ever reports again

	_, err := r.Check(context.Background())
	if err == nil {
		t.Fatal("Check() after the queue heartbeat went stale = nil error, want an error")
	}
}

// TestJobsReadiness_QueueHeartbeatAdvanceKeepsReady proves the other half:
// when River genuinely re-stamps a subscribed queue's UpdatedAt (its own
// internal report loop is alive), readiness stays healthy even though a lot
// of wall-clock time has passed overall.
func TestJobsReadiness_QueueHeartbeatAdvanceKeepsReady(t *testing.T) {
	clock := &fakeJobsClock{t: time.Unix(1_700_000_000, 0)}
	src := &fakeQueueSource{rows: map[string]time.Time{
		"maintenance": clock.Now(),
	}}
	r := &jobsReadiness{}
	r.ConfigureHeartbeat(src, []string{"maintenance"}, 5*time.Minute, clock.Now)
	r.MarkStarted()

	clock.Advance(4 * time.Minute)
	src.setUpdatedAt("maintenance", clock.Now()) // River's own report loop ticked

	clock.Advance(4 * time.Minute) // 8 min since start, but only 4 min since the report

	if _, err := r.Check(context.Background()); err != nil {
		t.Fatalf("Check() 4 min after the last queue report (threshold 5 min) = %v, want nil error", err)
	}
}

// TestJobsReadiness_QueueGetErrorFailsReadiness proves a QueueGet failure
// (e.g. the DB connection backing River is down) fails closed rather than
// being silently ignored.
func TestJobsReadiness_QueueGetErrorFailsReadiness(t *testing.T) {
	src := &fakeQueueSource{err: errQueueGetBoom}
	r := &jobsReadiness{}
	r.ConfigureHeartbeat(src, []string{"maintenance"}, 5*time.Minute, nil)
	r.MarkStarted()

	if _, err := r.Check(context.Background()); err == nil {
		t.Fatal("Check() with a failing QueueGet = nil error, want an error")
	}
}
