# Binary: metaldocs-worker

> **Last verified:** 2026-07-02 (StagingOutboxWorker consolidation: `PDFOutboxWorker`/`MaterializeOutboxWorker` replaced by two instances of the generic `fanout.StagingOutboxWorker`; restart loop removed) | **Prior:** 2026-06-16
> **Scope:** The `apps/worker` binary — its purpose, what it consumes from Postgres, what external services it calls, how it is configured, and its full lifecycle from startup to shutdown. This document also covers the three in-API async goroutine subsystems that are architecturally part of the same worker concern (outbox relay, maintenance jobs, sweepers) even though they run inside the API process.
> **Key files:**
> - `apps/worker/cmd/metaldocs-worker/main.go` — binary entrypoint
> - `apps/worker/cmd/metaldocs-worker/main_test.go` — unit tests for batch and loop modes
> - `internal/platform/worker/service.go` — poll-and-dispatch core
> - `internal/platform/worker/pdf_job_runner.go` — PDF event handler
> - `internal/platform/worker/materialize_job_runner.go` — DOCX materialize handler
> - `internal/platform/bootstrap/worker.go` — dependency wiring
> - `internal/platform/config/worker.go` — configuration schema
> - `internal/modules/jobs/scheduler/scheduler.go` — in-API maintenance scheduler
> - `internal/modules/render/fanout/staging_outbox_worker.go:23` — generic in-API staging outbox relay worker (PDF + materialize instances)
> - `internal/modules/render/fanout/staging_outbox.go:33` — generic staging outbox repository (allowlist-validated table binding)
> - `internal/modules/documents/jobs/session_sweeper.go` — in-API session sweeper
> - `internal/modules/documents/jobs/orphan_pending_sweeper.go` — in-API orphan sweeper

---

## 1. Why the worker binary exists

The worker binary owns the heavy I/O pipeline for document generation. It exists as a separate process so that long-running Gotenberg PDF conversions and docx-renderer HTTP calls do not block or compete for resources with the API process. It is the sole consumer of the `metaldocs.outbox_events` table. See [../flows/async-job-pipeline.md](../flows/async-job-pipeline.md) for the end-to-end flow.

---

## 2. What it consumes and processes

The worker binary processes exactly two event types from `metaldocs.outbox_events`:

| Event type constant | String value | Handler | External dependency |
|--------------------|-------------|---------|---------------------|
| `EventTypePDFConvert` | `docgen_v2_pdf` | `PDFJobRunner` | Gotenberg (HTTP), MinIO |
| `EventTypeMaterializeFanout` | `docx_materialize` | `MaterializeJobRunner` | docx-renderer (HTTP), MinIO, Postgres |

Events are placed into `outbox_events` by the in-API relay workers (two instances of the generic `fanout.StagingOutboxWorker` — PDF and materialize) — not by the API handlers directly. See [../platform/async-messaging.md](../platform/async-messaging.md) for the staging outbox layer.

### PDF event (`docgen_v2_pdf`)

`PDFJobRunner.Handle` (`internal/platform/worker/pdf_job_runner.go:68`):

1. Extracts `PDFConvertPayload` from `Event.Payload`.
2. Derives the DOCX S3 key (`tenants/{tenantID}/revisions/{revisionID}/frozen.docx` if not already in the payload).
3. Calls `GotenbergPDFClient.ConvertPDF` (`internal/platform/servicebus/gotenberg_pdf.go:70`): reads the DOCX blob from MinIO, POSTs to Gotenberg, writes the resulting PDF back to MinIO.
4. Calls `PDFPersister.WritePDF` to record the PDF S3 key and content hash in the snapshot repository.

### DOCX materialize event (`docx_materialize`)

`MaterializeJobRunner.Handle` (`internal/platform/worker/materialize_job_runner.go:58`):

1. Extracts `MaterializeFanoutPayload`.
2. Calls `MaterializeInvoker.Materialize` — HTTP POST to the docx-renderer fanout service. This call is made **outside any transaction**.
3. Opens a new Postgres transaction to call `WriteFinalDocxInTx` (persists the final DOCX S3 key) and `pdfOutbox.Enqueue` atomically (inserts a `pdf_dispatch_outbox` row). This enqueue re-enters the PDF pipeline: the API's PDF `StagingOutboxWorker` instance will relay it into `outbox_events` and the worker will process it as a `docgen_v2_pdf` event.

---

## 3. Configuration

Loaded from `internal/platform/config/worker.go` via `config.LoadWorkerConfig`.

| Env var | Default | Effect |
|---------|---------|--------|
| `METALDOCS_WORKER_POLL_INTERVAL_SECONDS` | `10` | Ticker interval for the outbox-poll loop |
| `METALDOCS_WORKER_BATCH_SIZE` | `25` | `ClaimUnpublished` row limit per tick |
| `METALDOCS_WORKER_RUN_ONCE` | `false` | If `true`, run a single batch then exit (CI / test mode) |
| `METALDOCS_WORKER_MAX_ATTEMPTS` | `5` | Dead-letter threshold (`attempt_count >= MaxAttempts`) |
| `METALDOCS_WORKER_RETRY_BASE_SECONDS` | `10` | Exponential backoff base |
| `METALDOCS_WORKER_RETRY_MAX_SECONDS` | `300` | Backoff ceiling; also sets `claimLease = max(RetryMaxSeconds, 5 min)` |
| `METALDOCS_WORKER_REVIEW_REMINDER_DAYS` | `14` | **Loaded and logged but not consumed** — see flags section |
| `METALDOCS_FANOUT_URL` | `""` | If set, enables `MaterializeJobRunner` |
| `METALDOCS_DOCX_RENDERER_SERVICE_TOKEN` | `""` | Bearer token for docx-renderer calls |
| `METALDOCS_GOTENBERG_URL` + Gotenberg/MinIO vars | — | Enables `PDFJobRunner` (conditional on `buildWorkerPDFConverter`) |
| `PGHOST`, `PGPORT`, `PGDATABASE`, `PGUSER`, `PGPASSWORD`, `PGSSLMODE` | — | Postgres connection |

---

## 4. Lifecycle

### Startup (`apps/worker/cmd/metaldocs-worker/main.go`)

1. `config.LoadWorkerConfig` — reads env vars, fails fast on missing required values.
2. `bootstrap.BuildWorkerDependencies` (`internal/platform/bootstrap/worker.go`): opens Postgres connection; builds `GotenbergPDFClient` (conditional); constructs `outboxpg.Consumer`; reads fanout URL and renderer token; returns `WorkerDependencies`.
3. Wires `PDFJobRunner` and `MaterializeJobRunner` into `platform/worker.Service` via `worker.NewService`.
4. Dispatches based on `RUN_ONCE`:
   - `RUN_ONCE=true`: calls `runWorkerBatch(ctx, svc)` once and exits.
   - Default: enters `runWorkerLoop(ctx, svc, interval)` — a `time.Ticker`-driven loop calling `RunOnce` each tick.

### Processing loop (`internal/platform/worker/service.go`)

```
tick
  └─ ClaimUnpublished(ctx, batchSize)          [FOR UPDATE SKIP LOCKED on outbox_events]
       for each event:
         route by EventType → call handler.Handle
         success → MarkPublished
         failure → markFailure (backoff | dead-letter)
  └─ log worker_batch line
```

Events are processed **sequentially within a batch** — no goroutine-per-event parallelism (`service.go:42`). Batch completion is logged as `worker_batch result=completed processed=... failed=... dead_lettered=... duration_ms=...` (`service.go:93`).

### Shutdown

A `signal.NotifyContext` for `os.Interrupt` and `syscall.SIGTERM` is registered at `apps/worker/cmd/metaldocs-worker/main.go:52–53`, so SIGTERM does cancel the context. However, there is no graceful drain logic: when the context is cancelled mid-batch, in-flight Gotenberg/HTTP calls are abandoned immediately with no wait for completion. [runtime-unverified] — actual abandonment behavior depends on whether the external HTTP clients respect context cancellation.

---

## 5. In-API async subsystems

Three async subsystems run as goroutines inside `apps/api` rather than in `apps/worker`. They are architecturally part of the worker concern.

### 5.1 Outbox relay workers

Both relays are instances of the generic `fanout.StagingOutboxWorker` (`internal/modules/render/fanout/staging_outbox_worker.go:23`), started by `startOutboxWorkers` (`apps/api/cmd/metaldocs-api/main.go:945`, called at `main.go:543`). The per-instance difference is only the repository table binding (`NewPDFOutboxRepository`/`NewMaterializeOutboxRepository`, `staging_outbox.go:215/220`) and the `buildEvent` callback:

| Instance (wired at) | Source table | Event type | Poll | Batch |
|--------|-------------|-----------|------|-------|
| PDF (`main.go:960`) | `metaldocs.pdf_dispatch_outbox` | `docgen_v2_pdf` | 5 s | 10 |
| Materialize (`main.go:976`) | `metaldocs.materialize_dispatch_outbox` | `docx_materialize` | 5 s | 10 |

Each relay instance: claims rows with `FOR UPDATE SKIP LOCKED`, calls `Publisher.Publish` (inserts into `outbox_events`), marks rows dispatched. `ResetStaleClaims` recovers rows stuck in `processing` status.

Poll/batch/retry knobs come from the shared `config.StagingOutboxWorkerConfig` (`internal/platform/config/staging_outbox_worker.go`; env `METALDOCS_STAGING_OUTBOX_POLL_INTERVAL_SECONDS` / `_BATCH_SIZE` / `_MAX_ATTEMPTS` / `_STALE_AFTER_SECONDS`; defaults 5 s / 10 / 5 / 300 s). There is no restart wrapper: `Run` returns only `nil` on context cancellation, and the goroutines are joined via `workerWG` (`main.go:542`) at shutdown.

### 5.2 In-API maintenance scheduler

`jobscheduler.New(deps.SQLDB, leaderID, slog.Default())` called at `apps/api/cmd/metaldocs-api/main.go:525`. `leaderID = hostname:pid`.

Runs four jobs, each in its own goroutine with a `time.Ticker`, protected by Postgres advisory lease (`metaldocs.acquire_lease`/`heartbeat_lease`/`release_lease` functions):

| Job name | Package | Interval | Backpressure |
|----------|---------|---------|-------------|
| `stuck-instance-watchdog` | `modules/jobs/stuck_instance_watchdog` | 5 min | `SkipOnPressure` |
| `idempotency-janitor` | `modules/jobs/idempotency_janitor` | 15 min | `SkipOnPressure` |
| `audit-integrity-validator` | `modules/jobs/audit_integrity_validator` | 1 h | `SkipOnPressure` |
| `lease-reaper` | `modules/jobs/scheduler` (built-in) | 10 min | `SkipOnPressure` |

> **Retired (annotation 2026-08-09):** this scheduler table describes the pre-ADR-0067 Postgres-lease ticker scheduler, which is **retired (M5)** along with `lease-reaper`. The periodic maintenance jobs now run as River periodic jobs enqueued by the API's leader election and executed by `metaldocs-jobs` (`internal/modules/jobs/maintenance/periodic.go`). Kept as history; do not use as current runtime truth.

Backpressure probe (`scheduler.go:302–359`): if `pg_stat_activity` active / max_connections ratio > 0.70 (enter) or < 0.60 (exit), jobs are skipped for that tick. Warning logged at 10 consecutive skips.

Graceful drain at `drain()`: waits up to 30 s for in-flight jobs, then up to 5 s more with forced cancellation before releasing leases.

`leaderID = hostname:pid` note: in a container environment where the pod restarts with the same hostname but a new PID, the old lease epoch will lapse (5 min TTL) before being re-acquired.

### 5.3 Lightweight sweepers

Started at `apps/api/main.go:568–569`:

| Sweeper | Interval | Action | Auth |
|---------|---------|--------|------|
| `StartSessionSweeper` | 60 s | `repo.ExpireStaleSessions(ctx, now)` | `authz.WithBackgroundBypass` |
| `StartOrphanPendingSweeper` | 1 h | `repo.DeleteExpiredPending(ctx, now-24h)` | `authz.WithBackgroundBypass` |

Both use simple `time.Ticker` goroutines with no restart logic. Stopped via `defer stopSessions(); defer stopOrphans()` at `apps/api/main.go:568–569`.

---

## 6. Deployment

The worker binary is containerised via `deploy/docker/worker.Dockerfile` and defined as the `worker` service in `deploy/compose/docker-compose.yml`. The compose service includes `depends_on: docx-renderer: condition: service_healthy`.

---

## 7. Legacy and open flags

| Flag | Severity | RF reference |
|------|----------|-------------|
| `METALDOCS_WORKER_REVIEW_REMINDER_DAYS` loaded and logged but never consumed anywhere in service or runner code | Medium | — |
| Sequential batch processing — no per-event goroutine parallelism | Info | — |
| No graceful drain in `apps/worker/main.go`: SIGTERM cancels context (`main.go:52–53`) but in-flight HTTP calls are abandoned immediately [runtime-unverified] | Info | — |
| ~~`startOutboxWorker` restart loop is dead code~~ **CLOSED:** restart loop deleted with the `StagingOutboxWorker` consolidation — `startOutboxWorkers` (`main.go:945`) starts each instance directly | ~~Low~~ | — |
| `lease_reaper` governance events are never written for scheduler jobs: the subquery `SELECT tenant_id FROM public.documents WHERE id::text = job_name` (`lease_reaper.go:38`) always returns NULL because `job_name` is a string like `'stuck-instance-watchdog'`, never a document UUID. The `tenant attribution unavailable` error is logged (`lease_reaper.go:79–84`) and propagated via `errors.Join` (`lease_reaper.go:122`), but every reaped lease row is skipped. | High | — |
| Backpressure uses hard-coded pg_stat_activity ratio thresholds with no configuration surface | Low | — |

Full flag registry: [../legacy-register.md](../legacy-register.md).

---

## Sources

Stage-1 artifact: [../_artifacts/stage1/async-runtime.md](../_artifacts/stage1/async-runtime.md).
Strategic context: [../../architecture/backend-blueprint.md](../../architecture/backend-blueprint.md).
Target normative spec: [../../architecture/backend-target-architecture.md](../../architecture/backend-target-architecture.md).
