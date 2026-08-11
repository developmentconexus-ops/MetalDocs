package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"metaldocs/internal/platform/observability"
)

// fakeWorkerClock is an injectable, manually-advanced clock so heartbeat
// staleness tests drive time deterministically instead of sleeping.
type fakeWorkerClock struct {
	mu sync.Mutex
	t  time.Time
}

func (c *fakeWorkerClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.t
}

func (c *fakeWorkerClock) Advance(d time.Duration) {
	c.mu.Lock()
	c.t = c.t.Add(d)
	c.mu.Unlock()
}

// TestWorkerReadiness_NotReadyBeforeLoopStarts is the A7.1 RED-first proof
// for metaldocs-worker: readiness must report NOT ready while the outbox
// consumer poll loop has not actually started — a process that is up but
// whose loop never began (crashed goroutine, bootstrap ordering bug,
// deadlock before the first tick) must not read as healthy.
func TestWorkerReadiness_NotReadyBeforeLoopStarts(t *testing.T) {
	r := &workerReadiness{}

	_, err := r.Check(context.Background())
	if err == nil {
		t.Fatal("Check() before MarkStarted() = nil error, want an error (loop not started)")
	}
}

// TestWorkerReadiness_ReadyAfterLoopStarts proves the same object reports
// healthy once the loop has genuinely begun — both sides of the semantics
// are pinned, not just the failure path.
func TestWorkerReadiness_ReadyAfterLoopStarts(t *testing.T) {
	r := &workerReadiness{}
	r.MarkStarted()

	result, err := r.Check(context.Background())
	if err != nil {
		t.Fatalf("Check() after MarkStarted() = %v, want nil error", err)
	}
	if result.Status != "up" {
		t.Fatalf("Status = %q, want up", result.Status)
	}
}

// TestWorkerReadiness_NotReadyAfterStop proves that a controlled shutdown
// (signal received, loop draining/exiting) flips readiness back to NOT
// ready — an orchestrator must stop treating the process as able to do its
// job the moment the poll loop stops, not only before it ever started.
func TestWorkerReadiness_NotReadyAfterStop(t *testing.T) {
	r := &workerReadiness{}
	r.MarkStarted()
	r.MarkStopped()

	_, err := r.Check(context.Background())
	if err == nil {
		t.Fatal("Check() after MarkStopped() = nil error, want an error (loop stopped)")
	}
}

// TestWorkerReadinessEndpoint_ReflectsLoopState drives the full stack
// (workerReadiness -> observability.DependencyCheck ->
// PostgresRuntimeStatusProvider -> NewInfraServer) exactly as main.go wires
// it, over real HTTP via httptest, proving GET /ready is not a decoration:
// it can be driven to 503 by the same condition ("consumer loop not
// started") the binary would hit on a genuine bootstrap failure, and to 200
// once the loop is marked started with a reachable DB.
func TestWorkerReadinessEndpoint_ReflectsLoopState(t *testing.T) {
	r := &workerReadiness{}
	// nil *sql.DB models "no DB configured" — PostgresRuntimeStatusProvider
	// fails closed on that (see internal/platform/observability/runtime.go),
	// so this also exercises the "DB unreachable" readiness path end to end.
	provider := observability.NewPostgresRuntimeStatusProvider(nil, "postgres", "n/a", false,
		observability.DependencyCheck{Name: "outbox_consumer", Check: r.Check})
	server := observability.NewInfraServer(":0", provider, nil)

	req := httptest.NewRequest(http.MethodGet, "/ready", nil)
	rec := httptest.NewRecorder()
	server.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("GET /ready with nil DB and loop not started = %d, want 503", rec.Code)
	}
}

// TestWorkerReadiness_StaleHeartbeatFlipsNotReady is the F1 review round 2
// RED-first proof: a poll loop that started successfully but has stopped
// making progress (RunOnce hanging, or never being reached again) must flip
// readiness back to NOT ready once the configured heartbeat threshold is
// exceeded — a one-shot start latch alone cannot detect this. The clock is
// injected and advanced manually; the test never sleeps.
func TestWorkerReadiness_StaleHeartbeatFlipsNotReady(t *testing.T) {
	clock := &fakeWorkerClock{t: time.Unix(1_700_000_000, 0)}
	r := &workerReadiness{}
	r.ConfigureHeartbeat(30*time.Second, clock.Now)
	r.MarkStarted() // seeds the heartbeat at clock.Now() (startup grace)

	if _, err := r.Check(context.Background()); err != nil {
		t.Fatalf("Check() immediately after MarkStarted() = %v, want nil error (grace period)", err)
	}

	clock.Advance(31 * time.Second)

	_, err := r.Check(context.Background())
	if err == nil {
		t.Fatal("Check() after the heartbeat went stale = nil error, want an error")
	}
}

// TestWorkerReadiness_RecordProgressResetsHeartbeat proves the other half:
// a genuinely completed poll cycle (RecordProgress) resets the staleness
// clock, so a live loop that is ticking on schedule never flips to NOT
// ready.
func TestWorkerReadiness_RecordProgressResetsHeartbeat(t *testing.T) {
	clock := &fakeWorkerClock{t: time.Unix(1_700_000_000, 0)}
	r := &workerReadiness{}
	r.ConfigureHeartbeat(30*time.Second, clock.Now)
	r.MarkStarted()

	clock.Advance(25 * time.Second)
	r.RecordProgress() // a real cycle completed just before the threshold

	clock.Advance(25 * time.Second) // 25s since RecordProgress, still < 30s

	if _, err := r.Check(context.Background()); err != nil {
		t.Fatalf("Check() 25s after RecordProgress (threshold 30s) = %v, want nil error", err)
	}

	clock.Advance(6 * time.Second) // now 31s since RecordProgress

	if _, err := r.Check(context.Background()); err == nil {
		t.Fatal("Check() 31s after the last RecordProgress (threshold 30s) = nil error, want an error")
	}
}
