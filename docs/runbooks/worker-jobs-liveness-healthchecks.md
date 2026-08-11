# Runbook — metaldocs-worker / metaldocs-jobs container healthchecks (A7.1)

**Files:** `deploy/compose/docker-compose.yml` (`worker`/`jobs` service
`healthcheck:` blocks), `apps/worker/cmd/metaldocs-worker/infraserver.go`,
`apps/jobs/cmd/metaldocs-jobs/infraserver.go`,
`internal/platform/observability/infraserver.go`,
`internal/platform/config/worker.go` (the second, independently-maintained
reading of "is `METALDOCS_WORKER_RUN_ONCE` truthy" — see the caveat below)
**Owner:** on-call (whoever is triaging `docker compose ps` showing
`worker`/`jobs` as `unhealthy`, or a deploy tool stuck waiting on their
health state)
**Context:** A7.1 (issue #95, PR #109) gave both binaries a dedicated
infra-port listener (`GET /live`, `GET /ready`, `GET /metrics` — default
`:9091` worker, `:9092` jobs, overridable via `WORKER_METRICS_ADDR` /
`JOBS_METRICS_ADDR`) and wired `docker compose` `healthcheck:` blocks that
probe `GET /live` on each. This runbook covers what the check does, its one
structural caveat (worker batch mode), and what to do when it goes red.

## What the healthcheck actually probes

Both blocks probe `/live` (liveness — "process is up and its HTTP loop is
responsive"), **not** `/ready` (readiness — "this process's actual work
loop is making progress", per A7.1's heartbeat design). That choice is
deliberate: combined with `restart: unless-stopped`, a readiness-based
container healthcheck would restart-loop a worker or jobs process that is
merely waiting on a slow-but-recovering dependency (e.g. Postgres mid
crash-recovery) — killing a process that was about to succeed on its own.
`/ready` is for orchestration-level dependency gating (a human or a
gateway deciding whether to route to it), not for Docker's own
kill-and-restart decision.

## Caveat: worker batch mode (`METALDOCS_WORKER_RUN_ONCE=true`)

The `worker` service's healthcheck (`docker-compose.yml`, `worker.healthcheck.
test`) special-cases `METALDOCS_WORKER_RUN_ONCE` (same env var already wired
into the service's `environment:` block, read at container-runtime by the
healthcheck's shell command — no compose-level `depends_on`/ordering
changes): when it's truthy, the healthcheck is a deliberate no-op pass
instead of probing `/live`. Reason: `apps/worker/cmd/metaldocs-worker/main.go`
never starts the infra-port listener in `RunOnce` mode at all — "a one-shot
batch invocation exits within this call; there is no persistent process for
an orchestrator to probe" — so a plain `/live` probe would be structurally
unsatisfiable for that mode's entire container lifetime and misreport a
correctly-behaving batch worker as unhealthy. Continuous mode (the default —
`false`/unset) is unaffected and still gets the real probe.

**"Truthy" is defined twice, and the two definitions are only proven to
agree, not structurally identical (F9, review round 4).** The shell `case`
guard in the healthcheck and `config.LoadWorkerConfig`'s Go truthy check
(`internal/platform/config/worker.go`) are independently-maintained readings
of the same rule — a hand-synced enumeration. They have been cross-checked
value-by-value (Go's `TrimSpace`+`EqualFold("true")`/`"1"` vs. the shell's
POSIX trim + `case [Tt][Rr][Uu][Ee]|1`, the shell side run under both `dash`
and busybox `ash` — the actual interpreter Docker's `CMD-SHELL` invokes —
not just `bash`) and agree on every value tested, including
leading/trailing-whitespace-padded ones (`" true"`, `"true "`, `" true "`,
`"1 "`, `" 1"`, tab-padded, `"TRUE1"`, `"1true"`, empty, and all-case
variants of `true`/`false`). Before the F9 fix the shell guard did not trim,
so it silently disagreed with Go on padded input — the failure direction was
safe (falls through to the real `/live` probe, which fails closed as
"unhealthy" rather than false-reporting "healthy" for a batch-mode container
Go itself considers correctly configured), but it was a real divergence, not
a merely theoretical one. Do not re-introduce that gap by editing one side
without re-running the parity check against the other — the compose comment
above the `test:` line explains why deletion (removing this second reading
entirely) was investigated and rejected for this slice: the worker's
infra-port listener isn't started in batch mode at all, so there's no
process for the shell to ask instead of parsing the env var itself.

If you are running the `worker` service in batch mode and want an actual
per-run success/failure signal, use the container's own exit code (`0` =
batch completed and drained; nonzero = `runWorkerBatch` failed), not the
Docker healthcheck — the healthcheck cannot express that in this mode by
construction.

## When it goes red (continuous mode)

- **`worker`/`jobs` shows `unhealthy` and stays that way.** `docker exec` in
  and `wget -qO- http://127.0.0.1:9091/live` (worker) or `:9092/live` (jobs)
  by hand. If it hangs or refuses, the infra-port listener bind/serve
  failed — check the process's own logs for `"worker infra server failed"` /
  the jobs equivalent (logged, not fatal, per A7.1's "must not block the
  actual job" bound, so the process is still doing real work even with a
  dead health port — treat as degraded observability, not a full outage).
- **`worker`/`jobs` restart-loops (`Restarting` in `docker compose ps`, not
  just `unhealthy`).** That is the container's own process exiting, not the
  healthcheck — the healthcheck only marks state, `restart: unless-stopped`
  is what relaunches it. Check `docker logs` for the exit reason before
  assuming it's healthcheck-related.
- **Compose image build fails before any of this is reachable
  (`go.mod requires go >= 1.26.5` vs. `deploy/docker/*.Dockerfile`'s
  `golang:1.25-alpine`).** Known, tracked separately on
  `fix/dockerfile-go-version` — not a healthcheck defect.

## Deliberately deferred (tracked, not silently dropped)

`docker compose ps` / `docker inspect .State.Health` proof of the worker and
jobs healthchecks going `healthy` through a real built image has **not**
been captured as of A7.1 landing — blocked on the Dockerfile Go-version fix
above. `/live` itself is live-verified directly against the compiled
binaries (see PR #109 review thread replies). Once that fix merges: rebase,
build, `docker compose ps`, and attach the `healthy` transcript here or to
the follow-up that closes this note.
