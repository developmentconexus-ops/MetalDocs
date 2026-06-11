# Binary: metaldocs-jobs

> **Last verified:** 2026-06-11
> **Scope:** The `apps/jobs` binary — its River-based scheduling model, the single business job it executes, configuration, lifecycle, and deployment status. This document also covers the River client factory package and the approval module's River job definitions.
> **Key files:**
> - `apps/jobs/cmd/metaldocs-jobs/main.go` — binary entrypoint
> - `internal/platform/jobs/river/client.go` — River client factory
> - `internal/platform/bootstrap/jobs.go` — dependency wiring
> - `internal/platform/config/jobs.go` — configuration schema
> - `internal/modules/documents/approval/jobs/scheduled_publish_job.go` — River worker and enqueuer
> - `internal/modules/documents/approval/jobs/scheduled_publish_args.go` — River job args struct
> - `internal/modules/documents/approval/application/scheduler_service.go` — execution logic

---

## 1. Why the jobs binary exists

The jobs binary is a River-based job queue executor. It exists as a separate process to execute future-scheduled domain jobs — specifically the scheduled-publish cutover transaction that promotes a document from `scheduled` to `published` at a user-specified effective date. River provides durable, transactionally-safe job scheduling with at-least-once delivery semantics backed by Postgres.

The jobs binary is entirely independent of the outbox/worker pipeline. It shares no code path, no table, and no queue name with `apps/worker`. The two binaries coexist to serve orthogonal concerns: the worker handles event-driven I/O work (PDF generation, DOCX materialization); the jobs binary handles time-scheduled domain transactions.

---

## 2. Scheduling model

The jobs binary uses `github.com/riverqueue/river` with a single named queue: `temporal`.

| Aspect | Value | Source |
|--------|-------|--------|
| Queue name | `temporal` | `internal/platform/config/jobs.go:24` |
| Max workers | 10 (default) | `METALDOCS_JOBS_TEMPORAL_MAX_WORKERS` |
| Job schema | Configurable (default: River's default schema) | `METALDOCS_JOBS_RIVER_SCHEMA` |
| Scheduling model | River-internal scheduler fires jobs at `ScheduledAt` | River client internals |
| Delivery semantics | At-least-once (River retries on failure) | River framework |

River's internal scheduler is embedded in the client process — there is no separate cron process. The client polls its own Postgres tables and fires workers when a job's `ScheduledAt` has elapsed. [runtime-unverified — this is a River framework architectural claim; it cannot be confirmed by reading the MetalDocs codebase alone.]

---

## 3. What it processes

The binary processes exactly one job kind: `scheduled_publish_cutover`.

### Scheduled publish cutover (`scheduled_publish_cutover`)

**Enqueue path** (at HTTP time, inside the API process):

1. When a user schedules a document for future publication, `publish_service.go` calls `EnqueueScheduledPublishTx` (`internal/modules/documents/approval/jobs/scheduled_publish_job.go:56`).
2. `EnqueueScheduledPublishTx` calls `client.InsertTx` inside the active approvals transaction, inserting a River job row with `Queue="temporal"` and `ScheduledAt=effectiveDate`. The insert is **atomic with the document row update** — if the document write rolls back, the River job row rolls back with it.
3. The `RiverScheduledPublishEnqueuer` type wraps the River client to satisfy the `ScheduledPublishEnqueuer` interface consumed by the approval service. It is wired in `apps/api/cmd/metaldocs-api/main.go:450`.

**Execution path** (in the jobs binary):

1. River's internal scheduler fires `ScheduledPublishWorker.Work` (`internal/modules/documents/approval/jobs/scheduled_publish_job.go:33`) at `ScheduledAt`.
2. `Work` sets `authz.WithBackgroundBypass(ctx)` — jobs run outside any HTTP session.
3. `Work` calls `service.RunScheduledPublishJob` (`internal/modules/documents/approval/application/scheduler_service.go:44`).
4. The service opens a transaction, loads the document state with `FOR UPDATE` (stale-job guard: verifies generation/version are still current), then calls `publishScheduledDocumentTx` which executes `updateScheduledDocSQL` (`internal/modules/documents/approval/application/scheduler_service.go:23-31`):
   - Sets `documents.status = 'published'`
   - Sets `effective_from = NULL`
   - Increments `revision_version = revision_version + 1`
   - Inserts a `governance_events` row

If the stale-job guard fails (document was already published or withdrawn), the job returns without error and is marked complete by River.

---

## 4. Configuration

Loaded from `internal/platform/config/jobs.go` via `config.LoadJobsConfig`.

| Env var | Default | Effect |
|---------|---------|--------|
| `METALDOCS_JOBS_ENABLED` | `true` | If `false`, binary logs and exits immediately |
| `METALDOCS_JOBS_RIVER_SCHEMA` | `""` | River table schema (empty = River's default) |
| `METALDOCS_JOBS_TEMPORAL_MAX_WORKERS` | `10` | Concurrency cap for the `temporal` queue |
| `PGHOST`, `PGPORT`, `PGDATABASE`, `PGUSER`, `PGPASSWORD`, `PGSSLMODE` | — | Postgres connection |

---

## 5. Lifecycle

### Startup (`apps/jobs/cmd/metaldocs-jobs/main.go`)

1. `config.LoadJobsConfig` — reads env vars; exits if `METALDOCS_JOBS_ENABLED=false`.
2. `bootstrap.BuildJobsDependencies` (`internal/platform/bootstrap/jobs.go`):
   a. Opens Postgres connection.
   b. Calls `MigrateRiverSchema` — runs River's `IF NOT EXISTS` schema migration. Note: `apps/api/cmd/metaldocs-api/main.go:439` also calls `MigrateRiverSchema` at startup, so both binaries attempt this migration; River migrations are idempotent.
   c. Invokes the caller-supplied `JobsWorkerFactory` to construct `NewWorkers(scheduler, db)` — builds a `*river.Workers` registry containing `ScheduledPublishWorker`.
   d. Returns `JobsDependencies{River, SQLDB, Cleanup}`.
3. Starts the River client on the `temporal` queue: `deps.River.Client.Start(ctx)`.
4. Logs `MetalDocs Jobs running (queues=temporal)`.
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

**The jobs binary has no Dockerfile and is absent from the Docker Compose deployment.**

- `deploy/docker/` contains `api.Dockerfile` and `worker.Dockerfile` but no `jobs.Dockerfile`.
- `deploy/compose/docker-compose.yml` defines `api` and `worker` services but no `jobs` service.
- The binary is only reachable via `scripts/start-jobs.ps1` (local development only).

Consequence: **scheduled-publish cutover jobs will never fire in a containerised deployment.** River job rows inserted by `EnqueueScheduledPublishTx` will accumulate in the River Postgres tables and never be consumed. The scheduled-publish feature (`documents.status = 'published'` at a future effective date) is non-functional in any environment that uses the Docker Compose stack. [runtime-unverified — an alternate deployment mechanism outside compose is possible but not evidenced in the repository.]

---

## 8. Legacy and open flags

| Flag | Severity | Notes |
|------|----------|-------|
| No Dockerfile; absent from Docker Compose — scheduled-publish is non-functional in containerised deployment | High | Add `jobs.Dockerfile` and compose service |
| Both `apps/jobs` and `apps/api` call `MigrateRiverSchema` at startup — River schema lifecycle ownership is unclear | Low | Idempotent in practice; ownership should be documented |
| River schema name defaults to empty string; actual schema location at runtime depends on Postgres role search path [runtime-unverified] | Low | Set `METALDOCS_JOBS_RIVER_SCHEMA` explicitly in deployment config |

Full flag registry: [../legacy-register.md](../legacy-register.md).

---

## Sources

Stage-1 artifact: [../_artifacts/stage1/async-runtime.md](../_artifacts/stage1/async-runtime.md).
Strategic context: [../../architecture/backend-blueprint.md](../../architecture/backend-blueprint.md).
Target normative spec: [../../architecture/backend-target-architecture.md](../../architecture/backend-target-architecture.md).
