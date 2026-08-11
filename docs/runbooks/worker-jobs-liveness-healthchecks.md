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
responsive"), **not** `/ready`.

`/ready` is a **readiness latch**, not a live progress meter: it reports
database reachability (`PostgresRuntimeStatusProvider.Ready` pinging
Postgres), plus whether the worker's poll loop / jobs' River client has
started (`MarkStarted`), has not since stopped (`MarkStopped`), and has
reported a heartbeat within its configured staleness threshold. It proves
the process was alive and reporting recently enough to trust — it does
**not** prove that the most recent poll or queue iteration actually
succeeded; a hung iteration inside the threshold still reads as ready.

Round-5 correction: preferring `/live` over `/ready` here is **not** about
avoiding a restart loop. A Docker Compose healthcheck only updates the
container's reported health state (`docker compose ps` / `docker inspect
.State.Health`, and any future `depends_on: condition: service_healthy`
gate on this container) — it does not by itself cause `restart:
unless-stopped` to fire; Compose's restart policy reacts only to the
container's own process exiting, not to healthcheck status
(docs.docker.com/reference/compose-file/services,
docs.docker.com/engine/containers/start-containers-automatically). The
actual reason: probing `/ready` here would report a worker or jobs process
that is merely waiting on a slow-but-recovering dependency (e.g. Postgres
mid crash-recovery) as `unhealthy`, even though the process itself is fine
and about to succeed on its own. `/ready` is for orchestration-level
dependency gating (a human or a gateway deciding whether to route to or
depend on it); `/live` is the right signal for this container-level probe.

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
**Tracked as ME-16** (`docs/engineering/mechanical-enforcement-register.md`,
[issue #115](https://github.com/developmentconexus-ops/MetalDocs/issues/115))
— closes when a batch-specific compose service/override lands, which
removes the shell guard (and this second truthy reading) entirely rather
than patching it in place.

If you are running the `worker` service in batch mode and want an actual
per-run success/failure signal, use the container's own exit code (`0` =
batch completed and drained; nonzero = `runWorkerBatch` failed), not the
Docker healthcheck — the healthcheck cannot express that in this mode by
construction.

## When it goes red (continuous mode)

- **`worker`/`jobs` shows `unhealthy` and stays that way.** Run the same
  probe the healthcheck runs, by hand, inside the container:
  `docker exec metaldocs-worker wget -qO- http://127.0.0.1:9091/live` for
  the worker, or `docker exec metaldocs-jobs wget -qO- http://127.0.0.1:9092/live`
  for jobs (container names from `container_name:` in
  `deploy/compose/docker-compose.yml`; `docker compose exec worker ...` /
  `docker compose exec jobs ...` work equivalently if you're in the compose
  project directory). If it hangs or refuses, the infra-port listener bind/serve
  failed — check the process's own logs for `"worker infra server failed"` /
  the jobs equivalent (logged, not fatal, per A7.1's "must not block the
  actual job" bound, so the process is still doing real work even with a
  dead health port — treat as degraded observability, not a full outage).
- **`worker`/`jobs` restart-loops (`Restarting` in `docker compose ps`, not
  just `unhealthy`).** That is the container's own process exiting, not the
  healthcheck — the healthcheck only marks state, `restart: unless-stopped`
  is what relaunches it. Check `docker logs` for the exit reason before
  assuming it's healthcheck-related. **Known case:** if `worker` is running
  with `METALDOCS_WORKER_RUN_ONCE=true`, this is expected today, not a bug
  to chase — `runWorkerBatch` exits by design once the batch drains, and
  `restart: unless-stopped` restarts the container on any exit, so a batch
  run currently loops. Tracked as ME-16 Surface B (issue #115 above); use
  `docker compose run --rm worker` (which does not carry `restart:`, unlike
  `up`) or manually `docker compose stop worker` after the batch you care
  about completes.
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

The `worker`/`jobs` healthcheck `test:` lines hard-code `127.0.0.1:9091` /
`:9092` — a second, unforced reading of `WORKER_METRICS_ADDR` /
`JOBS_METRICS_ADDR` (`infraserver.go:34`/`:53`, default `:9091`/`:9092`).
They agree today only because neither service's `environment:` overrides
those vars; nothing enforces that they keep agreeing. **Tracked as ME-16
Surface A** (`docs/engineering/mechanical-enforcement-register.md`,
[issue #115](https://github.com/developmentconexus-ops/MetalDocs/issues/115))
— closes when one compose variable drives both the process's listen address
and the probe's target.
