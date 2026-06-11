# Module: jobs

> **Last verified:** 2026-06-10
> **Status:** active
> **Maturity:** L0 — Stage-1 audit draft — not yet promoted via metaldocs-module-doc
> **Scope:** The `internal/modules/jobs` module — the in-API Scheduler, its four maintenance jobs (stuck-instance watchdog, idempotency janitor, audit-integrity validator, lease reaper), and the lightweight document sweepers under `internal/modules/documents/jobs`. The River-based `ScheduledPublishWorker` under `internal/modules/documents/approval/jobs` is included because it is the sole job consumed by the `apps/jobs` binary.
> **Out of scope:** The platform packages that underpin async execution (`internal/platform/worker`, `internal/platform/messaging`, `internal/platform/jobs/river`) — those are documented in [../backend/platform/async-messaging.md](../backend/platform/async-messaging.md). The worker binary itself — [../backend/binaries/worker.md](../backend/binaries/worker.md). The jobs binary — [../backend/binaries/jobs.md](../backend/binaries/jobs.md).
> **Key files:**
> - `internal/modules/jobs/scheduler/scheduler.go` — distributed lease scheduler
> - `internal/modules/jobs/scheduler/lease_reaper.go` — lease-reclaim job
> - `internal/modules/jobs/stuck_instance_watchdog/job.go`
> - `internal/modules/jobs/idempotency_janitor/job.go`
> - `internal/modules/jobs/audit_integrity_validator/job.go`
> - `internal/modules/documents/jobs/session_sweeper.go`
> - `internal/modules/documents/jobs/orphan_pending_sweeper.go`
> - `internal/modules/documents/approval/jobs/scheduled_publish_job.go`
> - `internal/modules/documents/approval/jobs/scheduled_publish_args.go`

---

## Approach

The `jobs` module provides three distinct async execution mechanisms:

1. **Distributed lease scheduler** (`internal/modules/jobs/scheduler`): a custom per-job ticker loop backed by Postgres advisory leases. Multiple API replicas can run simultaneously; each job is executed by at most one replica at a time (leader election per job via `metaldocs.acquire_lease`). Used for maintenance work that runs inside the API process.

2. **Lightweight goroutine sweepers** (`internal/modules/documents/jobs`): simple fire-and-forget goroutines with time tickers. No distributed coordination, no restart logic. Used for low-stakes row-expiry operations.

3. **River job workers** (`internal/modules/documents/approval/jobs`): River-based `WorkerDefaults` implementing time-scheduled domain transactions. Consumed by the `apps/jobs` binary. Used for the scheduled-publish cutover that must fire atomically at a future effective date.

---

## Scheduler

`internal/modules/jobs/scheduler/scheduler.go`

### Registration

Jobs are registered as `JobConfig` values with a `JobFunc` (`func(ctx context.Context, epoch int64) error`) and a `BackpressurePolicy`. `New(db, leaderID)` takes the Postgres handle and a `leaderID` string (`hostname:pid` built by `schedulerLeaderID()` at `apps/api/cmd/metaldocs-api/main.go:839-845`).

### Execution loop (per registered job)

Each call to `Scheduler.Start` (`scheduler.go:141`) spawns a goroutine:

1. `time.Ticker` fires at the configured interval.
2. `probePressure` queries `pg_stat_activity` — if active/max_connections ratio > 0.70, apply `BackpressurePolicy` (all four jobs use `SkipOnPressure`).
3. `acquireLease` calls the Postgres function `metaldocs.acquire_lease($job, $leaderID, '5 minutes')`, which returns `(acquired bool, epoch int64)`.
4. If acquired: spawn a heartbeat goroutine (`heartbeat_lease` every ~1 min), call `cfg.Fn(jobCtx, epoch)`, stop heartbeat, call `releaseLease`.
5. If not acquired: skip (another replica holds the lease).

Lease TTL: 5 minutes. `leaderID` is `hostname:pid`. Container restarts with the same hostname but a new PID will allow the old lease to lapse naturally within the 5-min TTL.

### Metrics

`Metrics` counter struct exposed via `Snapshot()`: per-job runs, errors, and skips. No export to Prometheus or any external sink — the snapshot is accessible programmatically only.

### Graceful drain

`drain()` waits up to 30 s for in-flight jobs, then up to 5 s more with forced cancellation, then releases all leases.

---

## Registered maintenance jobs

### stuck-instance-watchdog

`internal/modules/jobs/stuck_instance_watchdog/job.go`

- Interval: 5 min. Acquires `pg_try_advisory_lock` at job start (belt-and-suspenders, in addition to the scheduler lease).
- Lists `approval_instances` stuck > 7 days.
- Per instance: auto-cancels via `cancelSvc` or emits a governance alert via `emitter`, per `drift_policy`.
- Uses `authz.WithBackgroundBypass` — no tenant context required.
- Accumulates per-instance errors via `errors.Join`; partial failures do not abort the epoch.

### idempotency-janitor

`internal/modules/jobs/idempotency_janitor/job.go`

- Interval: 15 min.
- Batched DELETE of expired rows from `metaldocs.idempotency_keys` (batch=5000, max 10 iterations per run).
- Orphan detection pass: warns on `in_flight` rows past the grace window.

### audit-integrity-validator

`internal/modules/jobs/audit_integrity_validator/job.go`

- Interval: 1 h.
- Calls `auditdomain.IntegrityValidator.ValidateIntegrity(ctx)` (one parameter — `epoch` is in the enclosing `JobFunc` scope and is logged but not forwarded to the validator).
- Returns `ErrIntegrityViolation` if issues found (logged by the scheduler loop).

### lease-reaper (built-in)

`internal/modules/jobs/scheduler/lease_reaper.go`

- Interval: 10 min.
- `RunLeaseReaper(db)` returns a `JobFunc` that: queries expired `job_leases` rows, deletes each, inserts a `governance_events` row per reclaim.
- **High-severity flag:** `lease_reaper.go:38` uses `SELECT doc.tenant_id FROM public.documents doc WHERE doc.id::text = d.job_name LIMIT 1`. The `job_leases.job_name` values for the four registered scheduler jobs are strings like `"stuck-instance-watchdog"`, not document UUIDs. This subquery always returns NULL, causing every reaped lease to be skipped (`rowErrs` appended). The `reclaimed` counter is always 0. Governance rows for scheduler-level lease reaps are never written. [runtime-unverified in a live system, but confirmed as a code-reading finding.]

---

## Lightweight sweepers

`internal/modules/documents/jobs/`

| Sweeper | Package | Interval | Action |
|---------|---------|---------|--------|
| `StartSessionSweeper` | `documents/jobs` | 60 s | `repo.ExpireStaleSessions(ctx, now)` |
| `StartOrphanPendingSweeper` | `documents/jobs` | 1 h | `repo.DeleteExpiredPending(ctx, now-24h)` |

Both goroutines are started at `apps/api/cmd/metaldocs-api/main.go:568-569` and stopped via deferred stop functions. No distributed coordination (these are single-process goroutines, duplicated work if multiple replicas run). Both use `authz.WithBackgroundBypass`.

---

## River job: scheduled-publish cutover

`internal/modules/documents/approval/jobs/`

This sub-package is the domain-facing side of the `apps/jobs` binary.

### Types

- `ScheduledPublishArgs` (`scheduled_publish_args.go`): River job args struct. `Kind()` returns `"scheduled_publish_cutover"`.
- `ScheduledPublishWorker` (`scheduled_publish_job.go`): River `WorkerDefaults[ScheduledPublishArgs]`. `Work` sets `authz.WithBackgroundBypass(ctx)` and calls `SchedulerService.RunScheduledPublishJob`.
- `RiverScheduledPublishEnqueuer`: wraps the River client to satisfy the `ScheduledPublishEnqueuer` interface consumed by the approval service.
- `NewWorkers`: assembles the `*river.Workers` registry for the jobs binary.

### Enqueue path

`EnqueueScheduledPublishTx` (`scheduled_publish_job.go:56`) calls `client.InsertTx` inside the active approvals transaction — the River job row is inserted atomically with the document row update. `ScheduledAt = effectiveDate`.

### Execution path

River fires `ScheduledPublishWorker.Work` at `ScheduledAt`. Work calls `service.RunScheduledPublishJob` (`approval/application/scheduler_service.go:44`), which opens a transaction, loads the document with `FOR UPDATE` (stale-job guard), and calls `publishScheduledDocumentTx` (UPDATE + governance INSERT) or returns cleanly if the document is already published/withdrawn.

---

## Public surface

| Export | Package | Consumers |
|--------|---------|-----------|
| `Scheduler`, `JobConfig`, `JobFunc`, `BackpressurePolicy`, `SkipOnPressure`, `DegradeOnPressure`, `MetricsSnapshot`, `New`, `RunLeaseReaper` | `modules/jobs/scheduler` | `apps/api/cmd/metaldocs-api/main.go` |
| `New` (audit validator), `JobName`, `ErrIntegrityViolation` | `modules/jobs/audit_integrity_validator` | `apps/api/cmd/metaldocs-api/main.go` |
| `New` (idempotency janitor), `JobName` | `modules/jobs/idempotency_janitor` | `apps/api/cmd/metaldocs-api/main.go` |
| `New` (watchdog), `JobName` | `modules/jobs/stuck_instance_watchdog` | `apps/api/cmd/metaldocs-api/main.go` |
| `StartSessionSweeper`, `StartOrphanPendingSweeper` | `modules/documents/jobs` | `apps/api/cmd/metaldocs-api/main.go` |
| `NewWorkers`, `NewScheduledPublishEnqueuer`, `ScheduledPublishArgs`, `ScheduledPublishWorker`, `RiverScheduledPublishEnqueuer` | `modules/documents/approval/jobs` | `apps/jobs/main.go` (execution), `apps/api/cmd/metaldocs-api/main.go` (enqueue) |

HTTP routes: none. The jobs module exposes no HTTP surface.

---

## Persistence

| Table | Written by | Read by | Notes |
|-------|-----------|---------|-------|
| `metaldocs.job_leases` | `acquire_lease`, `heartbeat_lease`, `release_lease` (Postgres functions) | `lease_reaper.go` (DELETE on expiry), scheduler goroutines | Schema in `db/baseline/0001_current_schema.sql` |
| `governance_events` | `lease_reaper.go` (INSERT per reclaim), `stuck_instance_watchdog` (INSERT alert) | — | Lease reaper writes are currently a no-op due to the JOIN bug |
| `metaldocs.idempotency_keys` | API handlers (outside this module) | `idempotency_janitor` (DELETE expired rows) | — |
| `metaldocs.audit_events` | (outside this module) | `audit_integrity_validator` (read-only) | — |
| `approval_instances` | Approval module | `stuck_instance_watchdog` (read + cancel) | — |
| `documents` | Approval module | `scheduled_publish_job.go` (UPDATE status) | Via `approval/application/scheduler_service.go` |
| River schema tables | `MigrateRiverSchema` at startup | River client | Schema location determined by `METALDOCS_JOBS_RIVER_SCHEMA` |

---

## Failure modes

| Failure | Symptom | Response |
|---------|---------|---------|
| Postgres advisory lease not acquired (another replica holds it) | Job silently skipped for that tick — expected behavior | Normal under multi-replica deployment |
| `stuck_instance_watchdog` partial failure | Partial errors joined; epoch still released; stuck instances may remain | Check `slog` error output; per-instance failure does not abort run |
| `idempotency_janitor` slow (large expired batch) | Runs up to 50,000 deletions per execution cycle | Monitor batch loop duration; table may lag under high idempotency-key volume |
| `lease_reaper` governance writes silent no-op | `governance_events` rows never written for scheduler job leaps | Known code bug — fix requires removing the `documents` JOIN |
| `apps/jobs` binary not running | `scheduled_publish_cutover` River rows accumulate; documents never become `published` | High-severity deployment gap — see [../backend/binaries/jobs.md](../backend/binaries/jobs.md) |
| `sweeper` goroutine exits (context cancel) | Goroutine stops without restart | Context cancel is graceful shutdown; no restart needed |

---

## Risks and technical debt

- **Critical: 0**
- **High: 2**
  - `lease_reaper` governance writes are always a no-op (wrong JOIN on `job_leases.job_name`)
  - `apps/jobs` binary absent from Docker Compose; scheduled-publish non-functional in containers
- **Medium: 1**
  - `METALDOCS_WORKER_REVIEW_REMINDER_DAYS` loaded in the worker config (adjacent area) but never consumed; review-reminder feature not implemented
- **Low: several** — deferred; will be enumerated in the tech-debt register on module promotion

---

## See also

- [../backend/binaries/worker.md](../backend/binaries/worker.md) — worker binary and in-API maintenance subsystems
- [../backend/binaries/jobs.md](../backend/binaries/jobs.md) — jobs binary, River model, deployment status
- [../backend/flows/async-job-pipeline.md](../backend/flows/async-job-pipeline.md) — end-to-end async flows with Mermaid diagrams
- [../backend/platform/async-messaging.md](../backend/platform/async-messaging.md) — messaging platform packages
- [./approval.md](./approval.md) — approval module (owns `ScheduledPublishWorker`)
- [../decisions/](../decisions/) — ADR 0015 (async DOCX materialization)

---

## Sources

Stage-1 artifact: [../backend/_artifacts/stage1/async-runtime.md](../backend/_artifacts/stage1/async-runtime.md).
Strategic context: [../architecture/backend-blueprint.md](../architecture/backend-blueprint.md).
Target normative spec: [../architecture/backend-target-architecture.md](../architecture/backend-target-architecture.md).
