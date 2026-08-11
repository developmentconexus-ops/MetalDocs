package main

import (
	"net/http"
	"time"

	"metaldocs/internal/platform/bootstrap"
	"metaldocs/internal/platform/config"
	"metaldocs/internal/platform/db/postgres"
	"metaldocs/internal/platform/observability"
)

// riverQueueReportInterval mirrors River's own internal, non-configurable
// queue-report cadence (vendor/github.com/riverqueue/river/producer.go's
// queueReportIntervalDefault = 10 * time.Minute). river.Config exposes no
// field to override it, so this is the real, honestly-derived basis for the
// jobsReadiness heartbeat threshold (see readiness.go) — not a magic
// duration invented independently of it.
const riverQueueReportInterval = 10 * time.Minute

// jobsQueueNames returns the queue names this binary subscribes (whatever
// jobsCfg.Queues carries — "maintenance" plus "temporal" by default) for the
// F1 queue-report heartbeat: jobsReadiness.Check reads
// river_queue.updated_at for each of these.
func jobsQueueNames(jobsCfg config.JobsConfig) []string {
	names := make([]string, 0, len(jobsCfg.Queues))
	for name := range jobsCfg.Queues {
		names = append(names, name)
	}
	return names
}

// buildInfraServer wires the A7.1 liveness/readiness/metrics server for
// metaldocs-jobs: a dedicated infra-port listener (JOBS_METRICS_ADDR,
// default :9092) built from the shared observability.NewInfraServer
// mechanism (internal/platform/observability/infraserver.go) — the same one
// metaldocs-worker uses.
//
// F5 (review round 2): JOBS_METRICS_ADDR is deliberately bare, not
// METALDOCS_-prefixed — see apps/worker/cmd/metaldocs-worker/infraserver.go's
// buildInfraServer doc comment for the full rationale (it mirrors
// metaldocs-api's own bare APP_PORT/METRICS_ADDR convention for infra-port
// listen addresses, a convention deliberately separate from METALDOCS_*
// business config).
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
