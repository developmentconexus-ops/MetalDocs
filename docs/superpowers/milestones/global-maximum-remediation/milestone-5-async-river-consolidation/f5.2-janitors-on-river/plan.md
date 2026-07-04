# F5.2 — janitors on River (plan)

> Executes `spec.md` / contract §2. Engine: `superpowers:subagent-driven-development` (fresh subagent
> per task; sonnet implement+review; main session reviews + commits). Expand/contract ordering (contract §1).

## Load-bearing design facts (from runtime map, 2026-07-04)

1. **`metaldocs-api` runs its own River client** (`apps/api/cmd/metaldocs-api/main.go:509-513`,
   `SkipUnknownJobCheck:true`, `Queues=jobsCfg.Queues={temporal}`, `Workers=nil`) for HTTP-side
   `InsertTx` (scheduled-publish + lifecycle enqueue). It **joins leader election** (same River schema).
2. **River enqueues periodic jobs only on the elected leader**, using *that client's* `Config.PeriodicJobs`
   (`vendor/.../river/client.go:955-956`; `internal/maintenance/queue_maintainer_leader.go` starts the
   PeriodicJobEnqueuer only on `notification.IsLeader`). ⇒ **If api is leader and lacks the periodic-job
   definitions, the janitors never enqueue.**
3. **Job execution is singleton per row** across all clients subscribing to a queue (River claims rows
   `FOR UPDATE SKIP LOCKED`). A client that subscribes to a queue but has no worker for a fetched Kind
   **errors that job** — so api must NOT subscribe to `maintenance`.

### ⇒ Rail: dual-definition, single-subscription (the correct topology)

- **Define** the 3 janitor periodic jobs on **both** clients (`Config.PeriodicJobs`), so whichever client
  is leader enqueues them. The definitions are pure `(schedule, emptyArgs, InsertOpts{Queue:"maintenance"})`
  — no worker/deps needed to *enqueue*.
- **Subscribe + register workers** for `maintenance` **only in `metaldocs-jobs`** (add
  `Queues["maintenance"]` in the jobs path only — NOT in the shared `LoadJobsConfig` default; register the
  3 janitor `river.Worker`s there). api stays enqueue-when-leader-only, never executes.
- Result: cluster-wide singleton execution on `metaldocs-jobs`, enqueue resilient to which client leads.
  This is consistent with contract §1 ("jobs binary hosts execution") and ADR 0067; the enqueue-on-leader
  detail is River mechanics, **not** a contract divergence. Recorded in evidence.

## Task breakdown (ordered — expand, then contract, then drop, then prove, then docs)

### T1 (sonnet) — platform: thread PeriodicJobs + maintenance queue
- `internal/platform/jobs/river/client.go`: add `PeriodicJobs []*river.PeriodicJob` to `Config`; pass to
  `river.Config{... PeriodicJobs: cfg.PeriodicJobs}`. Leave `Queues` as-is.
- `internal/platform/bootstrap/jobs.go`: thread `PeriodicJobs` through `BuildJobsDependencies` →
  `NewClientBundle` (add a field/param; keep api's separate `NewClientBundle` call able to pass its own).
- No behavior change yet; `go build ./...` green.

### T2 (sonnet) — each janitor exposes a River Worker (body IDENTICAL)
For stuck-instance-watchdog, idempotency-janitor, audit-integrity-validator: add to each package an
`Args` empty struct with `Kind()` (kinds: `stuck_instance_watchdog`, `idempotency_janitor`,
`audit_integrity_validator` — reuse existing `JobName` consts) and a `Worker` embedding
`river.WorkerDefaults[Args]` whose `Work(ctx, *river.Job[Args]) error` runs the **existing body verbatim**.
- Refactor: extract each current `New(...) scheduler.JobFunc` closure body into an unexported
  `run(ctx context.Context) error` (drop the unused `epoch` — or pass `job.Attempt`); the Worker + any
  residual scheduler shim both call it. **Do not change the query, batch size, drift policy, or the
  `SeedTxTenant`/`WithBackgroundBypass` calls.**
- **stuck-instance-watchdog: remove `acquireRunLock`/`pg_try_advisory_lock`** (`job.go:114`) — the River
  elector+queue provide the singleton guard (contract §2.6). **HS-7:** this removes an *unrelated*
  single-runner lock; it is NOT an H-PRE-1 retirement (H-PRE-1 stays LIVE — audit-writer lock).
- Break the dependency on `internal/modules/jobs/scheduler` (that package is deleted in T4). Keep
  constructors dependency-compatible (db / cancelSvc+emitter / auditValidator).

### T3 (sonnet) — wire the 3 workers + shared periodic-job definitions
- New shared `MaintenancePeriodicJobs()` (e.g. `internal/modules/jobs/maintenance` or reuse `jobs` pkg):
  returns `[]*river.PeriodicJob` — watchdog `PeriodicInterval(5*time.Minute)`, idempotency
  `15*time.Minute`, audit `1*time.Hour`; each `PeriodicJobOpts{ID:<name>, RunOnStart:false}`,
  `InsertOpts{Queue:"maintenance"}`.
- `apps/jobs/cmd/metaldocs-jobs/main.go`: in the worker factory add the 3 janitor workers via
  `river.AddWorker`; construct deps — approval `Cancel`+emitter already built there for scheduled-publish;
  audit validator via `auditpg.NewWriter(db)` (satisfies `auditdomain.IntegrityValidator`, per map C.12).
  Add `Queues["maintenance"]={MaxWorkers:2}` (jobs-only). Pass `MaintenancePeriodicJobs()` to the client.
- `apps/api/cmd/metaldocs-api/main.go`: pass the **same** `MaintenancePeriodicJobs()` into api's
  `NewClientBundle` Config.PeriodicJobs (enqueue-when-leader). Do NOT add the maintenance queue or workers.
- `go build ./...` green; both binaries buildable.

### T4 (sonnet) — contract: delete the lease scheduler + api registration
- Delete `internal/modules/jobs/scheduler/` (scheduler.go, lease_reaper.go, lease helpers).
- `apps/api/cmd/metaldocs-api/main.go`: remove `registerScheduledJobs`, the `jobscheduler.New` build, the
  `schedulerWG` goroutine (`:599-617,999-1037`), `httpObs.SetSchedulerMetrics`. api starts with **no
  scheduler goroutine**. Leave the unrelated in-process sweepers (`StartSessionSweeper`,
  `StartOrphanPendingSweeper`, template reconciliation) — out of M5 scope (not lease-based; §8 note).
- Remove now-unused `ENABLE_JOB_*` gates for the migrated janitors (or repoint if trivially reused).
- Retirement census (spec gate 3) = 0. `go build ./...` green.

### T5 (mechanical→migration, sonnet) — drop DB lease objects (forward-only, ordered AFTER T4)
- New forward-only migration: `DROP FUNCTION acquire_lease/heartbeat_lease/release_lease`,
  `DROP TABLE metaldocs.job_leases`. Follow the repo's migration convention (check `db/` layout +
  `check-system-runnable`). No down needed (forward-only per repo norm); note in evidence.

### T6 (sonnet) — proofs (testdb, targeted -run)
- 3 equivalence integration tests (spec gate 1) — watchdog / idempotency / audit, each proving the River
  Worker body == pre-migration effect.
- §2.6 singleton proof (spec gate 2): two River clients on one testdb, `maintenance` queue on both for the
  test, assert the watchdog job executes exactly once per tick (elector singleton; advisory lock gone).
- testdb factory; targeted `-run`; NOT the full suite (box). Record any deferred full-suite run in evidence.

### T7 (haiku→main) — docs
- `developing-new-work/references/invariant-checklist.md` H-PRE-1 line → **keep LIVE**; note M5 removed the
  watchdog's *unrelated* single-runner lock (HS-7 correction).
- memory `advisory-lock-deadlock-constraint.md` → clarifier (LIVE, not retired; distinguish watchdog lock).
  Update `MEMORY.md` line.
- ADR 0067 §H-PRE-1 erratum; record HS-7 in program README.

## Test strategy
TDD where it bites: T6 proofs are the acceptance. For T2 refactor, the equivalence tests are the guard that
body behavior is unchanged. Build-green after every task. Commit per task (or per coherent pair) with
`docs(...)`/`feat(...)`/`refactor(...)` scoped messages; never push.

## Files touched (census)
`internal/platform/jobs/river/client.go`, `internal/platform/bootstrap/jobs.go`,
`internal/modules/jobs/stuck_instance_watchdog/*`, `internal/modules/jobs/idempotency_janitor/*`,
`internal/modules/jobs/audit_integrity_validator/*`, new `internal/modules/jobs/maintenance/*` (periodic
defs), `apps/jobs/cmd/metaldocs-jobs/main.go`, `apps/api/cmd/metaldocs-api/main.go`,
DELETE `internal/modules/jobs/scheduler/*`, new migration under `db/`, docs.
