package main

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
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
	mu    sync.Mutex
	rows  map[string]time.Time
	err   error
	asNil map[string]bool // names that QueueGet should return (nil, nil) for
}

func (f *fakeQueueSource) QueueGet(_ context.Context, name string) (*rivertype.Queue, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.err != nil {
		return nil, f.err
	}
	if f.asNil[name] {
		return nil, nil
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
// a genuine bootstrap failure or mid-shutdown, and back to 200 once River
// has started.
//
// Round-5 review (PR #109): a nil *sql.DB is NOT used here, on purpose.
// PostgresRuntimeStatusProvider.Ready fails closed with 503 whenever
// p.db == nil (internal/platform/observability/runtime.go), independent of
// any DependencyCheck — so a nil-DB version of this test would report 503
// even if the River wiring were deleted entirely, proving nothing about the
// River condition it claims to name. A healthy sqlmock *sql.DB (Ping is
// unmonitored by sqlmock by default, so it succeeds without an
// ExpectPing/expectation) isolates the DB leg as always-up, so the 503 -> 200
// transition below can only be explained by jobsReadiness.Check itself.
func TestJobsReadinessEndpoint_ReflectsRiverState(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer func() { _ = db.Close() }()

	r := &jobsReadiness{}
	provider := observability.NewPostgresRuntimeStatusProvider(db, "postgres", "n/a", false,
		observability.DependencyCheck{Name: "river_client", Check: r.Check})
	server := observability.NewInfraServer(":0", provider, nil)

	req := httptest.NewRequest(http.MethodGet, "/ready", nil)
	rec := httptest.NewRecorder()
	server.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("GET /ready with healthy DB and River not started = %d, want 503", rec.Code)
	}

	r.MarkStarted()

	req2 := httptest.NewRequest(http.MethodGet, "/ready", nil)
	rec2 := httptest.NewRecorder()
	server.Handler.ServeHTTP(rec2, req2)

	if rec2.Code != http.StatusOK {
		t.Fatalf("GET /ready with healthy DB and River started = %d, want 200", rec2.Code)
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

// TestJobsReadiness_OneStaleQueueAmongManyFlipsNotReady is the F6 (review
// round 3) RED-first proof: a live "maintenance" queue must NOT mask a dead
// "temporal" queue. Per-queue freshness is the contract this binary's own
// type doc promises ("every subscribed queue"); aggregating across queues
// by taking the MOST recent UpdatedAt (the pre-fix "newest" logic) lets one
// healthy queue paper over another that has stopped reporting entirely —
// exactly the "readiness tied to process existence, not progress" failure
// A7.1 exists to eliminate, just moved one level down (per-process ->
// per-queue).
//
// Non-vacuous by construction: "maintenance" is kept fresh on every tick
// (never goes stale on its own) while "temporal" is stamped once and then
// left to rot past the threshold. A test that advanced every queue together
// (see TestJobsReadiness_StaleQueueHeartbeatFlipsNotReady above) cannot
// distinguish newest-across-queues from oldest-across-queues — it passes
// either way. This one can: under the old "newest" aggregation it FAILS
// (reports ready, because "maintenance" is always recent), and only passes
// once the fix lands (oldest-across-queues correctly flags "temporal" as
// stale).
func TestJobsReadiness_OneStaleQueueAmongManyFlipsNotReady(t *testing.T) {
	clock := &fakeJobsClock{t: time.Unix(1_700_000_000, 0)}
	src := &fakeQueueSource{rows: map[string]time.Time{
		"maintenance": clock.Now(),
		"temporal":    clock.Now(),
	}}
	r := &jobsReadiness{}
	r.ConfigureHeartbeat(src, []string{"maintenance", "temporal"}, 5*time.Minute, clock.Now)
	r.MarkStarted()

	// "temporal" never reports again from here on; "maintenance" keeps
	// reporting right up to the threshold, so a newest-across-queues
	// aggregation would see it as fresh forever.
	for i := 0; i < 3; i++ {
		clock.Advance(4 * time.Minute)
		src.setUpdatedAt("maintenance", clock.Now())
	}
	// 12 minutes since "temporal" last reported (threshold 5 min);
	// "maintenance" reported 0 minutes ago.

	_, err := r.Check(context.Background())
	if err == nil {
		t.Fatal("Check() with \"temporal\" stale 12m past threshold (\"maintenance\" fresh) = nil error, want an error — a live queue must not mask a dead one")
	}
}

// TestJobsReadiness_NilQueueFailsReadiness proves QueueGet returning a nil
// *rivertype.Queue (no error — the shape River's public API allows for a
// queue name that was never created/subscribed, distinct from a QueueGet
// error) fails closed explicitly, rather than being silently skipped from
// the aggregation. Skipping is the same failure shape as F6: paired here
// with a "maintenance" queue that keeps reporting fresh, a skip-on-nil
// implementation would let "maintenance" mask "temporal" never having a
// queue row at all — this must still flip readiness to not-ready.
func TestJobsReadiness_NilQueueFailsReadiness(t *testing.T) {
	clock := &fakeJobsClock{t: time.Unix(1_700_000_000, 0)}
	src := &fakeQueueSource{
		rows:  map[string]time.Time{"maintenance": clock.Now()},
		asNil: map[string]bool{"temporal": true},
	}
	r := &jobsReadiness{}
	r.ConfigureHeartbeat(src, []string{"maintenance", "temporal"}, 5*time.Minute, clock.Now)
	r.MarkStarted()

	if _, err := r.Check(context.Background()); err == nil {
		t.Fatal("Check() with QueueGet returning a nil queue for \"temporal\" (\"maintenance\" fresh) = nil error, want an error")
	}
}
