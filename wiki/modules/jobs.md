# Module: jobs

> **Last verified:** 2026-07-28 (ADR 0085 Stage B — the `apps/jobs` binary's `scheduled_publish_cutover` job (`ScheduledPublishArgs`/`ScheduledPublishWorker`, `internal/modules/documents/approval/jobs/`) is DELETED, along with its upstream `PublishApproved`/`SchedulePublish`/`RunScheduledPublishJob` callers in the `approval` module. Publication is no longer a client-invoked transaction — it is now driven by the ADR 0085 release coordinator's `release_evaluate` job kind (`internal/modules/approval/jobs/release_evaluate_job.go`, `ReleaseEvaluateWorker`), enqueued in the same tx as an approval fact or artifact fact write, on the same `temporal` queue the old cutover job used. §"River job: release evaluation" section (renamed from "scheduled-publish cutover") and Public surface/Persistence rows below rewritten; see [`wiki/modules/approval.md`](approval.md) and [ADR 0085](../decisions/0085-release-coordinator-approval-driven-publication.md).) | prior: 2026-07-04 (M6 F6.2 — ADR 0069: added the `document_review_surfacer` River periodic job, hourly, `maintenance` queue, `RunOnStart:false`; consumes the documents-owned `ReviewDueReader`/`ReviewSurfaceWriter` ports, zero raw SQL on `public.documents`) | prior: 2026-07-04 (ADR 0067/0068 — River is the single async primitive; 3 janitors + lease scheduler retired; watchdog is alert-only. NOTE: this doc's body below still narrates the pre-ADR-0067 lease-scheduler architecture and has not yet been fully rewritten to the post-ADR-0067 shape — treat the River periodic-job facts as current and the Scheduler/lease-reaper narrative as historical pending a full rewrite) | prior: 2026-06-11 (Wave 1)
> **Status:** active
> **Maturity:** L0 — Stage-1 audit draft — not yet promoted via metaldocs-module-doc
> **Scope:** The `internal/modules/jobs` module — the in-API Scheduler, its four maintenance jobs (stuck-instance watchdog, idempotency janitor, audit-integrity validator, lease reaper), the lightweight document sweepers under `internal/modules/documents/jobs`, and the River periodic jobs registered in `internal/modules/jobs/maintenance`. The River-based `ReleaseEvaluateWorker` under `internal/modules/approval/jobs` (ADR 0085) is included because it is a job consumed by the `apps/jobs` binary — it replaces the deleted `ScheduledPublishWorker`/`scheduled_publish_cutover`; the M6 `document_review_surfacer` (`internal/modules/jobs/document_review_surfacer`) is included as the newest River periodic job on the same `maintenance` queue.
> **Out of scope:** The platform packages that underpin async execution (`internal/platform/worker`, `internal/platform/messaging`, `internal/platform/jobs/river`) — those are documented in [../backend/platform/async-messaging.md](../backend/platform/async-messaging.md). The worker binary itself — [../backend/binaries/worker.md](../backend/binaries/worker.md). The jobs binary — [../backend/binaries/jobs.md](../backend/binaries/jobs.md).
> **Key files:**
> - `internal/modules/jobs/stuck_instance_watchdog/job.go` — now a River `WorkerDefaults` job (ADR 0067); the old `internal/modules/jobs/scheduler` lease-scheduler package and `lease_reaper.go` cited in earlier verifications of this doc no longer exist in the tree — retired by ADR 0067 (M5 F5.1), superseded by River periodic jobs. §Scheduler/§Registered maintenance jobs below still narrate the pre-0067 architecture (Known gap, not yet rewritten).
> - `internal/modules/jobs/idempotency_janitor/job.go`
> - `internal/modules/jobs/audit_integrity_validator/job.go`
> - `internal/modules/jobs/maintenance/periodic.go` — shared River `PeriodicJobs()` definitions (4 jobs, `maintenance` queue), consumed by both `metaldocs-api` and `metaldocs-jobs`
> - `internal/modules/jobs/document_review_surfacer/job.go` — M6 F6.2 review-due surfacer (ADR 0069)
> - `internal/modules/documents/jobs/session_sweeper.go`
> - `internal/modules/documents/jobs/orphan_pending_sweeper.go`
> - `internal/modules/approval/jobs/release_evaluate_job.go` — ADR 0085 `ReleaseEvaluateWorker` (River job kind `release_evaluate`); supersedes the deleted `scheduled_publish_cutover` worker
> - `internal/modules/approval/jobs/release_evaluate_args.go` — River job args struct (payload IS the release generation key)
> - `internal/modules/approval/application/release_coordinator.go` — `ReleaseCoordinator.Evaluate`, the release predicate execution logic the worker calls

---

## Approach

> **Drift notice (2026-07-04):** The narrative below (Scheduler / lease-based leader election) describes the pre-ADR-0067 architecture. As of ADR 0067 (M5 F5.1, 2026-07-04) the custom `internal/modules/jobs/scheduler` lease scheduler and `lease_reaper.go` were **deleted** — all four maintenance jobs (stuck-instance watchdog, idempotency janitor, audit-integrity validator, and now the M6 `document_review_surfacer`) are River periodic jobs registered via `internal/modules/jobs/maintenance.PeriodicJobs()` and run on the `maintenance` queue, leader-elected by River itself. This section is retained as historical context pending a full Arc42 rewrite of this L0 doc; treat statements about `Scheduler`/`JobConfig`/lease TTL as **superseded**, and the River job facts in this doc's other sections (River job: scheduled-publish cutover, River job: document-review-surfacer) as current truth.

The `jobs` module provides three distinct async execution mechanisms:

1. **Distributed lease scheduler** (`internal/modules/jobs/scheduler`): a custom per-job ticker loop backed by Postgres advisory leases. Multiple API replicas can run simultaneously; each job is executed by at most one replica at a time (leader election per job via `metaldocs.acquire_lease`). Used for maintenance work that runs inside the API process.

2. **Lightweight goroutine sweepers** (`internal/modules/documents/jobs`): simple fire-and-forget goroutines with time tickers. No distributed coordination, no restart logic. Used for low-stakes row-expiry operations.

3. **River job workers** (`internal/modules/approval/jobs`): River-based `WorkerDefaults` implementing time-scheduled domain transactions. Consumed by the `apps/jobs` binary. As of ADR 0085, used for release-coordinator evaluation (`release_evaluate`) — both immediate re-evaluation after a fact lands and the effective-date timer for a future-dated publication share the same idempotent job kind.

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
- `RunLeaseReaper(db)` returns a `JobFunc` that: queries expired `job_leases` rows, deletes each, writes a structured `slog.WarnContext` log line per reclaim (Wave 1, F-07/1.7 — replaces the broken `governance_events` INSERT that had a cross-schema JOIN bug; `governance_events.tenant_id` is NOT NULL with no system-tenant convention).
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

## River job: release evaluation (ADR 0085, supersedes scheduled-publish cutover)

`internal/modules/approval/jobs/`

This sub-package is the domain-facing side of the `apps/jobs` binary. It previously hosted `ScheduledPublishArgs`/`ScheduledPublishWorker` (`Kind()` `"scheduled_publish_cutover"`), which called `SchedulerService.RunScheduledPublishJob` — both are **deleted**. Publication is no longer a client-invoked transaction; it is a side effect the release coordinator decides on, driven entirely by facts + timers.

### Types

- `ReleaseEvaluateArgs` (`release_evaluate_args.go:16`): River job args struct. `Kind()` returns `"release_evaluate"` (`:27`). The payload IS the release generation key (`TenantID`, `SubjectKind`, `DocumentID`, `ApprovalInstanceID`, `RevisionID`, `RevisionVersion`, `FrozenContentHash`) — nothing else is needed, since `Evaluate` re-reads every input under lock. One job kind serves both triggers: immediate re-evaluation after a fact lands (`ScheduledAt = now`) and the effective-date timer (`ScheduledAt = planned_effective_from`).
- `ReleaseEvaluateWorker` (`release_evaluate_job.go:23`): River `WorkerDefaults[ReleaseEvaluateArgs]`. `Work` (`:48`) sets `authz.WithBackgroundBypass(ctx)` and calls `ReleaseCoordinator.Evaluate`; a `application.ErrReleaseGenerationNotFound` result (tenant erased / row purged) is treated as success, not retried.
- `RiverReleaseEvaluationEnqueuer` / `DeferredReleaseEvaluationEnqueuer` (`release_evaluate_job.go:120,94`): the latter breaks a wiring cycle in the jobs binary — the worker set (which needs a coordinator, which needs an enqueuer) must be constructed before the River client the enqueuer inserts into exists; `Bind` supplies the real client once it does.

### Enqueue path

Any fact write that can change a release generation's readiness enqueues its own re-evaluation in the SAME transaction as the fact (transactional outbox) — `RecordArtifactFactTx` (`approval/application/release_facts.go:302`) and the approval-fact recorder both call `ReleaseEvaluationEnqueuer.EnqueueReleaseEvaluationTx` (`release_evaluate_job.go:109,131`), which does `client.InsertTx` on the `temporal` queue. A zero `runAt` means immediate; a future `runAt` (the plan's `planned_effective_from`) arms the effective-date timer.

### Execution path

River fires `ReleaseEvaluateWorker.Work` at `ScheduledAt`. `Work` calls `ReleaseCoordinator.Evaluate` (`approval/application/release_coordinator.go:123`), which re-derives the release predicate (approval fact × artifact facts × effective-date gate × supersession head, `domain.EvaluateRelease`) under an in-tx row lock and either holds (records a hold reason: `awaiting_approval_fact`/`materializing`/`awaiting_effective_date`/`supersede_conflict`/`plan_invalid`/`failed`), schedules (re-arms a future evaluation), or releases — publishing and supersession happen atomically inside the same winning transaction. A lost optimistic-concurrency CAS against a concurrent evaluation (`IsLostReleaseCAS`) is a benign no-op, not an error.

---

## River job: document-review-surfacer (M6 F6.2, ADR 0069)

`internal/modules/jobs/document_review_surfacer/job.go`

Periodic eQMS review/expiry surfacer. Registered as a River periodic job (not the custom lease scheduler) — `PeriodicInterval(time.Hour)`, `PeriodicJobOpts{ID: "document-review-surfacer", RunOnStart: false}`, queue `"maintenance"` (`internal/modules/jobs/maintenance/periodic.go:52-58`). Definitions are shared between `metaldocs-api` (which enqueues via its own elected leader's periodic scheduler) and `metaldocs-jobs` (which registers the `Workers` and actually executes) — only `metaldocs-jobs` subscribes the `maintenance` queue.

### Types

- `DocumentReviewSurfacerArgs` (`job.go:32`): empty River job args struct. `Kind()` returns `"document_review_surfacer"`.
- `DocumentReviewSurfacerWorker` (`job.go:42`): River `WorkerDefaults[DocumentReviewSurfacerArgs]`. Cluster-wide single-runner comes from River's leader-elected periodic insert + queue dequeue semantics — no advisory lock (mirrors `stuck_instance_watchdog` post-ADR-0067, ADR 0067 §H-PRE-1).
- `NewWorker(database, reader, writer)` (`job.go:51`) wires the documents-owned `ReviewDueReader`/`ReviewSurfaceWriter` ports (see [`documents.md`](documents.md) §8.7a) — this package holds **zero raw SQL** against `public.documents`.

### Execution (`Work` → `run`, `job.go:61-117`)

1. `authz.WithBackgroundBypass(ctx)` — no HTTP-request identity exists for a periodic job.
2. Opens a tx, calls `authz.BypassSystem(ctx, tx)` (scheduler bypass token, satisfies the `documents/UPDATE` tripwire for the write step).
3. `reader.ListDueForReview(ctx, tx, now, BatchSize=100)` — read-port call, for the observability count only.
4. `writer.MarkSurfaced(ctx, tx, now)` — write-port call, the idempotent side effect (sets `review_surfaced_at`; a rerun in the same review cycle is a no-op).
5. Commits; logs `due_count`/`surfaced_count`.

**Cross-tenant scope by design:** the tx runs under the scheduler bypass with no tenant GUC seeded — `public.documents`' RLS `tenant_isolation` policy treats an unset `metaldocs.tenant_id` GUC as "all tenants" (mirrors `stuck_instance_watchdog.listStuckInstances`), so one tx sweeps every tenant's due documents per tick. This is deliberate: there is one idempotent marker write per due document per cycle and no per-tenant side effect requiring isolation (contrast `stuck_instance_watchdog`'s per-instance governance-event emission, which does iterate).

## Public surface

| Export | Package | Consumers |
|--------|---------|-----------|
| `Scheduler`, `JobConfig`, `JobFunc`, `BackpressurePolicy`, `SkipOnPressure`, `DegradeOnPressure`, `MetricsSnapshot`, `New`, `RunLeaseReaper` | `modules/jobs/scheduler` | `apps/api/cmd/metaldocs-api/main.go` |
| `New` (audit validator), `JobName`, `ErrIntegrityViolation` | `modules/jobs/audit_integrity_validator` | `apps/api/cmd/metaldocs-api/main.go` |
| `New` (idempotency janitor), `JobName` | `modules/jobs/idempotency_janitor` | `apps/api/cmd/metaldocs-api/main.go` |
| `New` (watchdog), `JobName` | `modules/jobs/stuck_instance_watchdog` | `apps/api/cmd/metaldocs-api/main.go` |
| `StartSessionSweeper`, `StartOrphanPendingSweeper` | `modules/documents/jobs` | `apps/api/cmd/metaldocs-api/main.go` |
| `NewReleaseEvaluateWorker`, `ReleaseEvaluateArgs`, `ReleaseEvaluateWorker`, `NewReleaseEvaluationEnqueuer`, `DeferredReleaseEvaluationEnqueuer` | `modules/approval/jobs` | `apps/jobs/cmd/metaldocs-jobs/main.go:124` (registration/execution); enqueued from `approval/application/release_facts.go` (fact writes) |
| `NewWorker`, `DocumentReviewSurfacerArgs`, `DocumentReviewSurfacerWorker`, `JobName`, `BatchSize` | `modules/jobs/document_review_surfacer` | `apps/jobs/cmd/metaldocs-jobs/main.go:67` (registration) |
| `PeriodicJobs` | `modules/jobs/maintenance` | `apps/jobs/cmd/metaldocs-jobs/main.go`, `apps/api/cmd/metaldocs-api/main.go` (shared periodic-job definitions across both binaries) |

HTTP routes: none. The jobs module exposes no HTTP surface.

---

## Persistence

| Table | Written by | Read by | Notes |
|-------|-----------|---------|-------|
| `metaldocs.job_leases` | `acquire_lease`, `heartbeat_lease`, `release_lease` (Postgres functions) | `lease_reaper.go` (DELETE on expiry), scheduler goroutines | Schema in `db/baseline/0001_current_schema.sql` |
| `governance_events` | `stuck_instance_watchdog` (INSERT alert) | — | `lease_reaper` no longer writes here (Wave 1 fix): reclaim is now logged via `slog.WarnContext` |
| `metaldocs.idempotency_keys` | API handlers (outside this module) | `idempotency_janitor` (DELETE expired rows) | — |
| `metaldocs.audit_events` | (outside this module) | `audit_integrity_validator` (read-only) | — |
| `approval_instances` | Approval module | `stuck_instance_watchdog` (read + cancel) | — |
| `documents` | Approval module | `release_evaluate_job.go`/`ReleaseCoordinator.Evaluate` (UPDATE status, supersede) | ADR 0085; via `approval/application/release_coordinator.go`, replaces the deleted `scheduler_service.go`/`scheduled_publish_cutover` path |
| `public.release_generations` | Approval module (`RecordApprovalFactTx`/`RecordArtifactFactTx`) | `ReleaseCoordinator.Evaluate` | ADR 0085; migration `0310_release_coordinator.sql` |
| `documents` (`review_due_at`, `last_reviewed_at`, `review_surfaced_at`) | Documents module (mark-reviewed) | `document_review_surfacer` (read via `ReviewDueReader`, write via `ReviewSurfaceWriter` — never raw SQL from this module) | ADR 0069, migrations 0274/0276 |
| River schema tables | `MigrateRiverSchema` at startup | River client | Schema location determined by `METALDOCS_JOBS_RIVER_SCHEMA` |

---

## Failure modes

| Failure | Symptom | Response |
|---------|---------|---------|
| Postgres advisory lease not acquired (another replica holds it) | Job silently skipped for that tick — expected behavior | Normal under multi-replica deployment |
| `stuck_instance_watchdog` partial failure | Partial errors joined; epoch still released; stuck instances may remain | Check `slog` error output; per-instance failure does not abort run |
| `idempotency_janitor` slow (large expired batch) | Runs up to 50,000 deletions per execution cycle | Monitor batch loop duration; table may lag under high idempotency-key volume |
| ~~`lease_reaper` governance writes silent no-op~~ | **FIXED Wave 1 (1.7):** JOIN bug removed; reclaim now emits `slog.WarnContext` instead of broken `governance_events` INSERT | — |
| ~~`apps/jobs` binary not running~~| **FIXED Wave 0 (F-19):** `jobs.Dockerfile` and compose service added; runtime-verified. River schema migration sole owner = API binary (Wave 1, F-19). | — |
| `sweeper` goroutine exits (context cancel) | Goroutine stops without restart | Context cancel is graceful shutdown; no restart needed |

---

## Risks and technical debt

- **Critical: 0**
- **High: 0** — both former High risks are closed: the `lease_reaper` JOIN bug
  was fixed in Wave 1 (1.7) and the Postgres-lease scheduler + reaper were
  retired outright in M5; the `apps/jobs` binary has a Dockerfile and compose
  service since Wave 0 (F-19) (`deploy/docker/jobs.Dockerfile`,
  `deploy/compose/docker-compose.yml:291`), and publication is now the
  `release_evaluate` coordinator job (ADR 0085 Stage B), runtime-verified in
  containers.
- **Medium: 1**
  - `METALDOCS_WORKER_REVIEW_REMINDER_DAYS` loaded in the worker config (adjacent area) but never consumed; review-reminder feature not implemented
- **Low: several** — deferred; will be enumerated in the tech-debt register on module promotion

---

## See also

- [../backend/binaries/worker.md](../backend/binaries/worker.md) — worker binary and in-API maintenance subsystems
- [../backend/binaries/jobs.md](../backend/binaries/jobs.md) — jobs binary, River model, deployment status
- [../backend/flows/async-job-pipeline.md](../backend/flows/async-job-pipeline.md) — end-to-end async flows with Mermaid diagrams
- [../backend/platform/async-messaging.md](../backend/platform/async-messaging.md) — messaging platform packages
- [./approval.md](./approval.md) — approval module (owns `ReleaseEvaluateWorker`/`ReleaseCoordinator`, ADR 0085)
- [./documents.md](documents.md) §8.7a — owns the `ReviewDueReader`/`ReviewSurfaceWriter` ports consumed by `document_review_surfacer`
- [../decisions/0069-document-periodic-review-and-reason-for-change.md](../decisions/0069-document-periodic-review-and-reason-for-change.md) — ADR 0069 (M6 eQMS periodic review)
- [../decisions/](../decisions/) — ADR 0015 (async DOCX materialization), ADR 0067 (River consolidation), ADR 0068 (watchdog alert-only), [ADR 0085](../decisions/0085-release-coordinator-approval-driven-publication.md) (release coordinator — `release_evaluate` job kind)

---

## Sources

Stage-1 artifact: [../backend/_artifacts/stage1/async-runtime.md](../backend/_artifacts/stage1/async-runtime.md).
Strategic context: [../architecture/backend-blueprint.md](../architecture/backend-blueprint.md).
Target normative spec: [../architecture/backend-target-architecture.md](../architecture/backend-target-architecture.md).
