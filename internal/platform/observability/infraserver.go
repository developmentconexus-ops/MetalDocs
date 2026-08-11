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
	}
}
