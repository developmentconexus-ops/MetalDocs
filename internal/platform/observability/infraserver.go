package observability

import (
	"net/http"
	"time"

	platformmw "metaldocs/internal/platform/middleware"
)

// NewInfraServer builds a dedicated infra-port HTTP server exposing
// liveness (GET /live), readiness (GET /ready), and — when metricsHandler is
// non-nil — Prometheus metrics (GET /metrics) for a binary that has no
// public HTTP surface of its own (A7.1: metaldocs-worker, metaldocs-jobs).
//
// This is the ONE shared mechanism for that pattern; worker and jobs are its
// two consumers, each supplying their own RuntimeStatusProvider (with a
// DependencyCheck that reflects whether THAT binary can genuinely do its
// job — outbox consumer loop running, River client started and subscribed —
// not merely "process exists") and their own httpObs.PrometheusHandler().
//
// It mirrors metaldocs-api's already-landed dedicated METRICS_ADDR listener
// (apps/api/cmd/metaldocs-api/main.go buildServers): wrapped only in
// platformmw.Recovery, never chained through authn/contract-validation, and
// not declared in any OpenAPI spec — these are infra-port endpoints, not
// product API routes (A7.1 dispatch fence). Unlike api, worker/jobs have no
// other HTTP surface to split health from metrics against, so all three
// routes are consolidated onto this one dedicated listener per binary
// rather than mirroring api's own two-listener split (public contract
// health routes + separate metrics-only port) — api's shape does not apply
// here because worker/jobs have no public contract surface at all.
func NewInfraServer(addr string, provider RuntimeStatusProvider, metricsHandler http.Handler) *http.Server {
	mux := http.NewServeMux()
	health := NewHealthHandler(provider)
	mux.HandleFunc("GET /live", health.CheckLiveness)
	mux.HandleFunc("GET /ready", health.CheckReadiness)
	if metricsHandler != nil {
		mux.Handle("GET /metrics", metricsHandler)
	}

	return &http.Server{
		Addr:              addr,
		Handler:           platformmw.Recovery(mux),
		ReadHeaderTimeout: 5 * time.Second,
		// F4 (review round 2): only ReadHeaderTimeout was set, leaving
		// ReadTimeout/WriteTimeout/IdleTimeout unbounded — a reachable
		// client could hold a connection open indefinitely (Slowloris,
		// CWE-400) and stall /metrics scrapes or /live /ready probes until
		// fds/goroutines exhaust. Values match metaldocs-api's own public
		// server (apps/api/cmd/metaldocs-api/main.go buildServers,
		// REQ-REL-1/2 / F-16) — api's own dedicated metrics listener next to
		// it sets none of these either (a separate, pre-existing gap, out of
		// this slice's fence), so the public server is the only bounded
		// precedent this codebase has for an http.Server of this shape.
		// live/ready/metrics payloads are all small and fast, so these
		// bounds are generous headroom, not a tight fit.
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  90 * time.Second,
	}
}
