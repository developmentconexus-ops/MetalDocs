# Binary: metaldocs-jobs

> **Last verified:** 2026-08-12 (A5 queue-topology sync: the jobs binary subscribes to `temporal` (default max 10) and `maintenance` (max 2); `maintenance` is registered in `apps/jobs/cmd/metaldocs-jobs/main.go:102-105` and both queues are included in startup readiness.) | **Prior:** 2026-07-29 (ADR 0085 Stage C — the shared maintenance periodic jobs registered in this binary gained a 6th job, `release_hold_reconciler` (15-min alert-only reconciliation sweep over stuck release holds; full detail in [`wiki/modules/jobs.md`](../../modules/jobs.md)); Scope line's job list corrected — it was missing `approval-sla-surfacer` (F8) too.) | prior: 2026-07-28 (ADR 0085 Stage B — `scheduled_publish_cutover` is DELETED; the binary now executes the release coordinator's `release_evaluate` job kind, plus the shared maintenance periodic jobs and the other `temporal`-queue workers registered in `apps/jobs/cmd/metaldocs-jobs/main.go`. §3 rewritten below; see [`wiki/modules/approval.md`](../../modules/approval.md) and [ADR 0085](../../decisions/0085-release-coordinator-approval-driven-publication.md).) | prior: 2026-06-11
> **Scope:** The `apps/jobs` binary — its River-based scheduling model, the business jobs it executes (dominated by the ADR 0085 release coordinator's `release_evaluate` job), configuration, lifecycle, and deployment status. This document also covers the River client factory package and the approval module's River job definitions. It also hosts the 6 shared maintenance periodic jobs (stuck-instance watchdog, idempotency janitor, audit-integrity validator, document-review-surfacer, approval-sla-surfacer, release-hold-reconciler) documented in [`wiki/modules/jobs.md`](../../modules/jobs.md) — not repeated in full here.
> **Key files:**
> - `apps/jobs/cmd/metaldocs-jobs/main.go` — binary entrypoint
> - `internal/platform/jobs/river/client.go` — River client factory
> - `internal/platform/bootstrap/jobs.go` — dependency wiring
> - `internal/platform/config/jobs.go` — configuration schema
> - `internal/modules/approval/jobs/release_evaluate_job.go` — River worker and enqueuer (ADR 0085)
> - `internal/modules/approval/jobs/release_evaluate_args.go` — River job args struct
> - `internal/modules/approval/application/release_coordinator.go` — `ReleaseCoordinator.Evaluate`, execution logic
> - `internal/modules/jobs/release_hold_reconciler/job.go` — ADR 0085 Stage C release-hold reconciliation sweep (full detail in `wiki/modules/jobs.md`)

---

## 1. Why the jobs binary exists

The jobs binary is a River-based job queue executor. It exists as a separate process to execute future-scheduled and fact-triggered domain jobs — principally the ADR 0085 release coordinator's `release_evaluate` job, which decides whether a document is ready to publish (and executes the publish + supersede transaction when it is) off an approval-fact × artifact-fact × effective-date-gate predicate, rather than a client-invoked publish call. River provides durable, transactionally-safe job scheduling with at-least-once delivery semantics backed by Postgres.

The jobs binary is entirely independent of the outbox/worker pipeline. It shares no code path, no table, and no queue name with `apps/worker`. The two binaries coexist to serve orthogonal concerns: the worker handles event-driven I/O work (PDF generation, DOCX materialization); the jobs binary handles time-scheduled domain transactions.

---

## 2. Scheduling model

The jobs binary uses `github.com/riverqueue/river` with two named queues: `temporal` and `maintenance`. `temporal` is the default business-job queue; the jobs binary adds `maintenance` at startup for the shared periodic maintenance workers.

| Aspect | Value | Source |
|--------|-------|--------|
| Queues | `temporal` (default) and `maintenance` | `internal/platform/config/jobs.go:21-28`; `apps/jobs/cmd/metaldocs-jobs/main.go:102-105` |
| Max workers | `temporal`: 10 (default); `maintenance`: 2 | `internal/platform/config/jobs.go:26-40`; `apps/jobs/cmd/metaldocs-jobs/main.go:105` |
| Job schema | Configurable (default: River's default schema) | `METALDOCS_JOBS_RIVER_SCHEMA` |
| Scheduling model | River-internal scheduler fires jobs at `ScheduledAt` | River client internals |
| Delivery semantics | At-least-once (River retries on failure) | River framework |

River's internal scheduler is embedded in the client process — there is no separate cron process. The client polls its own Postgres tables and fires workers when a job's `ScheduledAt` has elapsed. [runtime-unverified — this is a River framework architectural claim; it cannot be confirmed by reading the MetalDocs codebase alone.]

---

## 3. What it processes

The binary registers many River workers on the shared `temporal` queue (release evaluation, notifications fanout, staging PDF/materialize dispatch, tenant lifecycle) plus the `maintenance`-queue periodic jobs (`apps/jobs/cmd/metaldocs-jobs/main.go:111-198`; the periodic-job set is documented in [`wiki/modules/jobs.md`](../../modules/jobs.md), not repeated here). The composition root adds `maintenance` with `MaxWorkers: 2` (`main.go:102-105`) and returns the periodic definitions alongside the worker registry (`main.go:196-197`). It also constructs one shared `platform/db.TxRunner` (`main.go:129`) and injects it into both notifications workers (`main.go:137-138`), centralizing their transaction lifecycle without changing their River registration or queue. The dominant business job — and the one that used to be `scheduled_publish_cutover` — is the ADR 0085 release coordinator's `release_evaluate`.

### Release evaluation (`release_evaluate`, ADR 0085 — supersedes `scheduled_publish_cutover`)

Publication is no longer a client-invoked transaction. A document becomes `published` only when a `release_evaluate` job evaluates its generation as ready and wins the release transaction.

**Enqueue path** (inside whichever process/transaction records the triggering fact):

1. Two fact recorders can trigger evaluation: the approval-terminal-decision recorder and `RecordArtifactFactTx` (`internal/modules/approval/application/release_facts.go:302`, called when the final DOCX or final PDF materializes). Both call the injected `ReleaseEvaluationEnqueuer.EnqueueReleaseEvaluationTx` in the SAME transaction as the fact write (transactional outbox).
2. `RiverReleaseEvaluationEnqueuer.EnqueueReleaseEvaluationTx` (`internal/modules/approval/jobs/release_evaluate_job.go:131`) calls `client.InsertTx` on `Queue="temporal"`. A zero `runAt` means "evaluate immediately"; a future `runAt` (the plan's `planned_effective_from`) arms the effective-date timer — the SAME job kind serves both triggers, so no separate cutover job type exists.
3. In the jobs binary itself, `DeferredReleaseEvaluationEnqueuer` (`release_evaluate_job.go:94`) breaks a wiring cycle: the worker set needs a coordinator, which needs an enqueuer, but the enqueuer needs the River client that doesn't exist until `BuildJobsDependencies` returns — `Bind` (`apps/jobs/cmd/metaldocs-jobs/main.go:174`) supplies the real client once it does.

**Execution path** (in the jobs binary):

1. River's internal scheduler fires `ReleaseEvaluateWorker.Work` (`internal/modules/approval/jobs/release_evaluate_job.go:48`) at `ScheduledAt`.
2. `Work` sets `authz.WithBackgroundBypass(ctx)` — jobs run outside any HTTP session — and calls `ReleaseCoordinator.Evaluate` (`internal/modules/approval/application/release_coordinator.go:123`), re-reading the release generation key straight from the job payload (`argsToKey`, `release_evaluate_job.go:72`) rather than trusting anything captured at enqueue time.
3. `Evaluate` locks the generation + document row, re-derives the release predicate (`domain.EvaluateRelease`: approval fact × artifact facts × effective-date gate; supersession-head checked separately inside the release tx) and either:
   - **holds** — records a hold reason (`awaiting_approval_fact`/`materializing`/`awaiting_effective_date`/`supersede_conflict`/`plan_invalid`/`failed`) and stops;
   - **schedules** — re-arms a future evaluation at `planned_effective_from`;
   - **releases** — inside one transaction: source-CAS UPDATE `approved|scheduled → published`, supersedes the discovered target(s), recomputes the review-cycle due date, records the release event, and commits.
4. A `application.ErrReleaseGenerationNotFound` result (tenant erased / row purged) is treated as success, not retried. A lost optimistic-concurrency CAS against a concurrent evaluation (`IsLostReleaseCAS`, `release_coordinator.go:157,538`) is also a benign no-op — a duplicate delivery decided the same way another evaluation already did.

---

## 4. Configuration

Loaded from `internal/platform/config/jobs.go` via `config.LoadJobsConfig`.

| Env var | Default | Effect |
|---------|---------|--------|
| `METALDOCS_JOBS_ENABLED` | `true` | If `false`, binary logs and exits immediately |
| `METALDOCS_JOBS_RIVER_SCHEMA` | `""` | River table schema (empty = River's default) |
| `METALDOCS_JOBS_TEMPORAL_MAX_WORKERS` | `10` | Concurrency cap for the `temporal` queue |
| `maintenance` queue | `2` | Fixed composition-root concurrency cap; registered by `metaldocs-jobs` at `apps/jobs/cmd/metaldocs-jobs/main.go:102-105` (no environment override) |
| `PGHOST`, `PGPORT`, `PGDATABASE`, `PGUSER`, `PGPASSWORD`, `PGSSLMODE` | — | Postgres connection |

---

## 5. Lifecycle

### Startup (`apps/jobs/cmd/metaldocs-jobs/main.go`)

1. `config.LoadJobsConfig` — reads env vars; exits if `METALDOCS_JOBS_ENABLED=false`.
2. `bootstrap.BuildJobsDependencies` (`internal/platform/bootstrap/jobs.go`):
   a. Opens Postgres connection.
   b. Calls `MigrateRiverSchema` — runs River's `IF NOT EXISTS` schema migration. Note: `apps/api/cmd/metaldocs-api/main.go:439` also calls `MigrateRiverSchema` at startup, so both binaries attempt this migration; River migrations are idempotent.
   c. Invokes the caller-supplied worker-factory closure (`apps/jobs/cmd/metaldocs-jobs/main.go:111-168`) to build a `*river.Workers` registry — `ReleaseEvaluateWorker` (ADR 0085), plus notifications fanout, staging PDF/materialize dispatch, tenant-lifecycle, and the shared maintenance periodic jobs.
   d. Returns `JobsDependencies{River, SQLDB, Cleanup}`.
3. Starts the River client with both configured queues (`temporal` and `maintenance`): `deps.River.Client.Start(ctx)`.
4. Logs `MetalDocs Jobs running` with `queues=temporal` (`main.go:235`); the current log field names the default queue even though `maintenance` is also configured and subscribed. Readiness heartbeat coverage includes both queues (`main.go:217-224`).
5. Blocks on `<-ctx.Done()`.

### Shutdown

`apps/jobs/cmd/metaldocs-jobs/main.go:52–57`: on SIGTERM / context cancellation, calls `deps.River.Client.Stop(shutdownCtx)` with a **15 second timeout**. River drains in-flight workers within the timeout before exiting.

---

## 6. River client package

`internal/platform/jobs/river/client.go` — `ClientBundle` factory:

- Wraps `github.com/riverqueue/river` + `riverdatabasesql` driver.
- Exposes `Client *river.Client[*sql.Tx]` and `Driver`.
- Used by `bootstrap/jobs.go`, `apps/jobs/cmd/metaldocs-jobs/main.go`, and `apps/api/cmd/metaldocs-api/main.go` (for `InsertTx` at enqueue time).

The package has no dependency on `platform/messaging`, `platform/worker`, or any outbox type.

---

## 7. Deployment status

**The jobs binary is containerised and deployed.** This section previously said the opposite; that claim was stale, and the correction is the point of this pass.

- `deploy/docker/` contains `api.Dockerfile`, `worker.Dockerfile` **and** `jobs.Dockerfile`.
- `deploy/compose/docker-compose.yml:291-294` defines a `jobs` service (`container_name: metaldocs-jobs`, `restart: unless-stopped`) built from `deploy/docker/jobs.Dockerfile`, gated on `postgres` healthy and `api`.
- `scripts/start-jobs.ps1` remains the local-development path; it is no longer the only one.

This closes what used to be the highest-severity flag on this binary: nothing consumes `release_evaluate` except the jobs host, so an absent `jobs` service meant no document could ever reach `published` in a containerised environment. That gap was filled by GMR milestone 8 (F8.1); the doc simply had not caught up. Compose-file truth is verified above; end-to-end container behaviour is still [runtime-unverified] here.

---

## 8. Legacy and open flags

| Flag | Severity | Notes |
|------|----------|-------|
| ~~No Dockerfile; absent from Docker Compose~~ | **Closed 2026-07-28** | `deploy/docker/jobs.Dockerfile` and the compose `jobs` service (`deploy/compose/docker-compose.yml:291-294`) both exist; see §7 |
| Both `apps/jobs` and `apps/api` call `MigrateRiverSchema` at startup — River schema lifecycle ownership is unclear | Low | Idempotent in practice; ownership should be documented |
| River schema name defaults to empty string; actual schema location at runtime depends on Postgres role search path [runtime-unverified] | Low | Set `METALDOCS_JOBS_RIVER_SCHEMA` explicitly in deployment config |

Full flag registry: [../legacy-register.md](../legacy-register.md).

---

## Sources

Stage-1 artifact: [../_artifacts/stage1/async-runtime.md](../_artifacts/stage1/async-runtime.md).
Strategic context: [../../architecture/backend-blueprint.md](../../architecture/backend-blueprint.md).
Target normative spec: [../../architecture/backend-target-architecture.md](../../architecture/backend-target-architecture.md).
