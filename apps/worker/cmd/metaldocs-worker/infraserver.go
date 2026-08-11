package main

import (
	"net/http"

	"metaldocs/internal/platform/bootstrap"
	"metaldocs/internal/platform/config"
	"metaldocs/internal/platform/db/postgres"
	"metaldocs/internal/platform/observability"
)

// buildInfraServer wires the A7.1 liveness/readiness/metrics server for
// metaldocs-worker: a dedicated infra-port listener (WORKER_METRICS_ADDR,
// default :9091) built from the shared observability.NewInfraServer
// mechanism (internal/platform/observability/infraserver.go) — the same one
// metaldocs-jobs uses, so the mux/handler bootstrap is written once, not
// duplicated per binary.
//
// F5 (review round 2): WORKER_METRICS_ADDR is deliberately NOT
// METALDOCS_-prefixed. That prefix is reserved for METALDOCS_* business/
// domain config; infra-port listen-address settings already have an
// established, separate, bare-name convention — metaldocs-api's own
// APP_PORT and METRICS_ADDR (config.LoadServerConfig) — which
// WORKER_METRICS_ADDR/JOBS_METRICS_ADDR follow exactly (binary-scoped
// prefix + "_METRICS_ADDR", same as api's bare METRICS_ADDR). Renaming to
// METALDOCS_WORKER_METRICS_ADDR would be the actual drift, not a fix.
//
// readiness's Check method is wired as the ONE DependencyCheck that gates
// GET /ready beyond the DB ping PostgresRuntimeStatusProvider already runs:
// it reports the outbox consumer poll loop's actual run state (see
// readiness.go), so readiness reflects "can this process genuinely do its
// job" rather than "is the process running".
func buildInfraServer(deps bootstrap.WorkerDependencies, readiness *workerReadiness) (*http.Server, error) {
	addr, err := config.LoadListenAddr("WORKER_METRICS_ADDR", ":9091")
	if err != nil {
		return nil, err
	}

	httpObs := observability.NewHTTPObservability(func(*http.Request) string { return "" })
	if deps.SQLDB != nil {
		httpObs.SetDBPool(postgres.NewPoolStatsAdapter(deps.SQLDB))
	}

	// repositoryMode/storageProvider/authEnabled are purely descriptive
	// fields in the JSON /ready and /metrics payloads (see
	// internal/platform/observability/runtime.go) — they never gate
	// readiness. metaldocs-worker always runs against Postgres and performs
	// no end-user authentication of its own; its actual storage provider is
	// resolved deeper, per job, in buildRevisionBodyReader.
	provider := observability.NewPostgresRuntimeStatusProvider(deps.SQLDB, "postgres", "n/a", false,
		observability.DependencyCheck{Name: "outbox_consumer", Check: readiness.Check})

	return observability.NewInfraServer(addr, provider, httpObs.PrometheusHandler()), nil
}
