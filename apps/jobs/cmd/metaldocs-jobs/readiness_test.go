package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"metaldocs/internal/platform/observability"
)

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
