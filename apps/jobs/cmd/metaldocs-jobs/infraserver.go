package main

import (
	"net/http"

	"metaldocs/internal/platform/bootstrap"
	"metaldocs/internal/platform/config"
	"metaldocs/internal/platform/db/postgres"
	"metaldocs/internal/platform/observability"
)

// buildInfraServer wires the A7.1 liveness/readiness/metrics server for
// metaldocs-jobs: a dedicated infra-port listener (JOBS_METRICS_ADDR,
// default :9092) built from the shared observability.NewInfraServer
// mechanism (internal/platform/observability/infraserver.go) — the same one
// metaldocs-worker uses.
//
// readiness's Check method is wired as the ONE DependencyCheck that gates
// GET /ready beyond the DB ping PostgresRuntimeStatusProvider already runs:
// it reports the River client's actual run state (see readiness.go), which
// is this binary's ADR-0067 truth — only metaldocs-jobs subscribes
// "maintenance" and executes it, so "River not started" IS "this process
// cannot run a periodic job right now", not a cosmetic detail.
func buildInfraServer(deps bootstrap.JobsDependencies, readiness *jobsReadiness) (*http.Server, error) {
	addr, err := config.LoadListenAddr("JOBS_METRICS_ADDR", ":9092")
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
	// readiness. metaldocs-jobs always runs against Postgres and performs
	// no end-user authentication of its own.
	provider := observability.NewPostgresRuntimeStatusProvider(deps.SQLDB, "postgres", "n/a", false,
		observability.DependencyCheck{Name: "river_client", Check: readiness.Check})

	return observability.NewInfraServer(addr, provider, httpObs.PrometheusHandler()), nil
}
