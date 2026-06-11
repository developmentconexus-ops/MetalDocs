# Stage-1 Audit Artifact — async-runtime

**Area:** async-runtime (jobs vs worker)
**Last verified:** 2026-06-10
**Author:** Stage-1 audit agent

---

## 1. Identity & purpose

MetalDocs runs **two separate async binaries** with distinct responsibilities. `apps/worker` is a polling outbox-consumer that owns the heavy I/O pipeline: it claims rows from `metaldocs.outbox_events`, dispatches them to Gotenberg (PDF) or to the docx-renderer fanout (DOCX materialization), and writes results back to Postgres. `apps/jobs` is a River-based scheduled-job executor that owns **one** business job: running scheduled-publish cutover transactions for the approval workflow. The two binaries share no code path at runtime: the worker uses a homegrown outbox-polling protocol; the jobs binary uses the River queue backed by its own Postgres schema.

Alongside these two external binaries, the **API binary** hosts three in-process async subsystems: (1) a custom Scheduler with distributed Postgres lease-based coordination running four maintenance jobs; (2) two lightweight outbox-relay goroutines (`PDFOutboxWorker`, `MaterializeOutboxWorker`) that move rows from domain-level staging tables into `metaldocs.outbox_events`; and (3) two fire-and-forget sweeper goroutines (`StartSessionSweeper`, `StartOrphanPendingSweeper`) that clean up transient Postgres rows.

---

## 2. File inventory

### `apps/jobs/cmd/metaldocs-jobs/`
| File | Role |
|------|------|
| `main.go` | Binary entrypoint — loads `JobsConfig`, calls `BuildJobsDependencies`, starts the River client on the `temporal` queue, blocks until SIGTERM, then calls `Client.Stop` with a 15 s timeout |

### `apps/worker/cmd/metaldocs-worker/`
| File | Role |
|------|------|
| `main.go` | Binary entrypoint — loads `WorkerConfig`, calls `BuildWorkerDependencies`, wires `PDFJobRunner` and `MaterializeJobRunner` into the platform `worker.Service`, then runs either a single-shot batch (`RunOnce`) or a `time.Ticker`-driven loop |
| `main_test.go` | Unit tests for `runWorkerBatch` and `runWorkerLoop` (context cancellation, error propagation) |

### `internal/platform/jobs/river/`
| File | Role |
|------|------|
| `client.go` | `ClientBundle` factory: wraps `github.com/riverqueue/river` + `riverdatabasesql` driver into a single struct; exposes `Client *river.Client[*sql.Tx]` and `Driver` |
| `client_test.go` | Integration test verifying transactional visibility of `InsertTx` (job invisible outside tx before commit, visible after) |

### `internal/platform/worker/`
| File | Role |
|------|------|
| `service.go` | `Service` — poll-and-dispatch loop; calls `consumer.ClaimUnpublished`, routes each `messaging.Event` to the correct runner by `EventType`, calls `MarkPublished`/`MarkFailed`, implements exponential backoff and dead-letter logic |
| `pdf_job_runner.go` | `PDFJobRunner` — handles `docgen_v2_pdf` events; extracts `PDFConvertPayload`, calls `PDFConverter.ConvertPDF`, persists result via `PDFPersister.WritePDF`; also defines adapter types `PDFConverter`, `PDFPersister`, `SnapshotPDFPersister` |
| `materialize_job_runner.go` | `MaterializeJobRunner` — handles `docx_materialize` events; calls `MaterializeInvoker.Materialize` (HTTP to docx-renderer), then in a single transaction writes the final-docx key and enqueues a `pdf_dispatch_outbox` row via `MaterializePDFEnqueuer.Enqueue` |
| `service_test.go` | Unit tests for `Service.RunOnce` and retry/DLQ logic |
| `pdf_job_runner_test.go` | Unit tests for `PDFJobRunner.Handle` |
| `materialize_job_runner_test.go` | Unit tests for `MaterializeJobRunner.Handle` |
| `pdf_pipeline_test.go` | Integration-style tests for the combined PDF pipeline path |

### `internal/platform/messaging/`
| File | Role |
|------|------|
| `consumer.go` | `Consumer` interface: `ClaimUnpublished`, `MarkPublished`, `MarkFailed` |
| `events.go` | Core types: `EventID`, `EventType`, `Event` envelope, `Publisher` interface; defines `EventTypePDFConvert = "docgen_v2_pdf"` and `EventTypeMaterializeFanout = "docx_materialize"` |
| `payloads.go` | `DecodePayload` switch; `PDFConvertPayloadFrom`/`MaterializeFanoutPayloadFrom` typed extractors |
| `noop/publisher.go` | No-op `Publisher` implementation (used in in-memory / test mode) |
| `outbox/postgres/consumer.go` | Postgres `Consumer`: `ClaimUnpublished` uses `FOR UPDATE SKIP LOCKED` CTE with lease timeout; `MarkPublished` sets `published_at`; `MarkFailed` updates retry schedule or DLQ timestamp |
| `outbox/postgres/publisher.go` | Postgres `Publisher`: inserts into `metaldocs.outbox_events` with `ON CONFLICT (idempotency_key) DO NOTHING` |
| `.gitkeep` | Empty placeholder |

### `internal/platform/servicebus/`
| File | Role |
|------|------|
| `gotenberg_pdf.go` | `GotenbergPDFClient`: reads the source DOCX from MinIO, calls `converter.ConvertDocxToPDFWithOptions`, writes PDF back to MinIO, returns `ConvertPDFResult{OutputKey, ContentHash, SizeBytes}` |
| `gotenberg_pdf_test.go` | Unit tests for the converter adapter |

### `internal/modules/jobs/scheduler/`
| File | Role |
|------|------|
| `scheduler.go` | `Scheduler`: per-job ticker loop with Postgres advisory-lease (`acquire_lease`/`heartbeat_lease`/`release_lease`), backpressure probe (active/max_connections ratio), drain-with-grace-period shutdown, `Metrics` counter struct |
| `lease_reaper.go` | `RunLeaseReaper(db)` — `JobFunc` that reclaims expired `job_leases` rows and inserts `governance_events` rows for each reaped lease |
| `scheduler_test.go`, `lease_reaper_test.go`, `integration_test.go` | Tests for scheduler and lease reaper |

### `internal/modules/jobs/audit_integrity_validator/`
| File | Role |
|------|------|
| `job.go` | `New(validator)` — `JobFunc` that calls `auditdomain.IntegrityValidator.ValidateIntegrity`; returns `ErrIntegrityViolation` if issues found |
| `job_test.go` | Unit tests |

### `internal/modules/jobs/idempotency_janitor/`
| File | Role |
|------|------|
| `job.go` | `New(db)` — `JobFunc`; batched DELETE of expired `metaldocs.idempotency_keys` rows (batch=5000, max 10 iterations); orphan detection pass warns on `in_flight` rows past the grace window |
| `job_test.go` | Unit tests |

### `internal/modules/jobs/stuck_instance_watchdog/`
| File | Role |
|------|------|
| `job.go` | `New(db, cancelSvc, emitter)` — `JobFunc`; acquires `pg_try_advisory_lock`, lists `approval_instances` stuck >7 days, auto-cancels or emits governance alert per `drift_policy`; uses `authz.WithBackgroundBypass` |
| `job_test.go` | Unit tests |

### `internal/modules/documents/jobs/`
| File | Role |
|------|------|
| `session_sweeper.go` | `StartSessionSweeper(ctx, repo, interval)` — fire-and-forget goroutine; calls `repo.ExpireStaleSessions` every 60 s |
| `orphan_pending_sweeper.go` | `StartOrphanPendingSweeper(ctx, repo, interval, maxAge)` — fire-and-forget goroutine; calls `repo.DeleteExpiredPending` every 1 h with a 24 h max-age cutoff |
| `jobs_test.go` | Tests for the sweeper functions |

### `internal/modules/documents/approval/jobs/`
| File | Role |
|------|------|
| `scheduled_publish_args.go` | `ScheduledPublishArgs` — River job args struct; `Kind()` returns `"scheduled_publish_cutover"` |
| `scheduled_publish_job.go` | `ScheduledPublishWorker` — River `WorkerDefaults[ScheduledPublishArgs]`; calls `SchedulerService.RunScheduledPublishJob` with `authz.WithBackgroundBypass`; `RiverScheduledPublishEnqueuer` wraps the River client to satisfy `ScheduledPublishEnqueuer`; `NewWorkers` assembles a `*river.Workers` registry |
| `scheduled_publish_job_test.go` | Unit tests |

### `internal/platform/bootstrap/` (async-relevant files)
| File | Role |
|------|------|
| `jobs.go` | `BuildJobsDependencies`: opens Postgres, runs River schema migration, invokes the caller-supplied worker factory, returns `JobsDependencies{River, SQLDB, Cleanup}`; also exports `MigrateRiverSchema` (used by the API binary too) |
| `worker.go` | `BuildWorkerDependencies`: opens Postgres, builds `GotenbergPDFClient` (conditional on Gotenberg config + MinIO provider), constructs `outboxpg.Consumer`, reads `METALDOCS_FANOUT_URL` and `METALDOCS_DOCX_RENDERER_SERVICE_TOKEN`, returns `WorkerDependencies` |

### `internal/platform/config/` (async-relevant files)
| File | Role |
|------|------|
| `jobs.go` | `JobsConfig` + `LoadJobsConfig`; env vars: `METALDOCS_JOBS_ENABLED`, `METALDOCS_JOBS_RIVER_SCHEMA`, `METALDOCS_JOBS_TEMPORAL_MAX_WORKERS` |
| `worker.go` | `WorkerConfig` + `LoadWorkerConfig`; env vars: `METALDOCS_WORKER_POLL_INTERVAL_SECONDS`, `METALDOCS_WORKER_BATCH_SIZE`, `METALDOCS_WORKER_REVIEW_REMINDER_DAYS`, `METALDOCS_WORKER_RUN_ONCE`, `METALDOCS_WORKER_MAX_ATTEMPTS`, `METALDOCS_WORKER_RETRY_BASE_SECONDS`, `METALDOCS_WORKER_RETRY_MAX_SECONDS` |

### `internal/modules/render/fanout/` (in-process outbox relay)
| File | Role |
|------|------|
| `pdf_outbox_repository.go` | `PDFOutboxRepository`: Enqueue/ClaimPending/MarkDispatched/MarkFailed/ResetStaleClaims against `metaldocs.pdf_dispatch_outbox`; used by both the API's `PDFOutboxWorker` and `MaterializeJobRunner` |
| `pdf_outbox_worker.go` | `PDFOutboxWorker`: polls `pdf_dispatch_outbox` every 5 s; claims up to 10 rows; calls `messaging.Publisher.Publish` to relay each row as a `docgen_v2_pdf` event into `outbox_events` |
| `materialize_outbox_repository.go` | `MaterializeOutboxRepository`: same pattern against `metaldocs.materialize_dispatch_outbox` |
| `materialize_outbox_worker.go` | `MaterializeOutboxWorker`: polls `materialize_dispatch_outbox` every 5 s; relays rows as `docx_materialize` events |

---

## 3. Public surface

### Exported types consumed outside the area

| Package | Export | Consumer |
|---------|--------|----------|
| `internal/platform/jobs/river` | `ClientBundle`, `Config`, `NewClientBundle` | `bootstrap/jobs.go`, `apps/api/main.go` |
| `internal/platform/worker` | `Service`, `PDFJobRunner`, `MaterializeJobRunner`, `PDFConverter`, `PDFPersister`, `MaterializeInvoker`, `MaterializeFanoutResult`, `MaterializeFinalDocxPersister`, `MaterializePDFEnqueuer`, `NewSnapshotPDFPersister` | `apps/worker/cmd/metaldocs-worker/main.go` |
| `internal/platform/messaging` | `Consumer`, `Publisher`, `Event`, `EventType`, `EventTypePDFConvert`, `EventTypeMaterializeFanout`, `Payload`, `PDFConvertPayload`, `MaterializeFanoutPayload`, `DecodePayload`, `PDFConvertPayloadFrom`, `MaterializeFanoutPayloadFrom`, `FailedEvent` | `platform/worker/*`, `platform/messaging/outbox/postgres/*`, `modules/render/fanout/*`, `bootstrap/*` |
| `internal/platform/servicebus` | `GotenbergPDFClient`, `ConvertPDFRequest`, `ConvertPDFResult`, `PaperSize`, `PDFRenderOpts` | `platform/worker/pdf_job_runner.go`, `bootstrap/worker.go`, `bootstrap/api.go`, `apps/api/main.go` |
| `internal/modules/jobs/scheduler` | `Scheduler`, `JobConfig`, `JobFunc`, `BackpressurePolicy`, `SkipOnPressure`, `DegradeOnPressure`, `MetricsSnapshot`, `New`, `RunLeaseReaper` | `apps/api/cmd/metaldocs-api/main.go` |
| `internal/modules/jobs/audit_integrity_validator` | `New`, `JobName`, `ErrIntegrityViolation` | `apps/api/cmd/metaldocs-api/main.go` |
| `internal/modules/jobs/idempotency_janitor` | `New`, `JobName` | `apps/api/cmd/metaldocs-api/main.go` |
| `internal/modules/jobs/stuck_instance_watchdog` | `New`, `JobName` | `apps/api/cmd/metaldocs-api/main.go` |
| `internal/modules/documents/jobs` | `StartSessionSweeper`, `StartOrphanPendingSweeper` | `apps/api/cmd/metaldocs-api/main.go` |
| `internal/modules/documents/approval/jobs` | `NewWorkers`, `NewScheduledPublishEnqueuer`, `ScheduledPublishArgs`, `ScheduledPublishWorker`, `RiverScheduledPublishEnqueuer` | `apps/jobs/cmd/metaldocs-jobs/main.go`, `apps/api/cmd/metaldocs-api/main.go` |
| `internal/platform/bootstrap` | `BuildJobsDependencies`, `BuildWorkerDependencies`, `JobsDependencies`, `WorkerDependencies`, `JobsWorkerFactory`, `MigrateRiverSchema` | `apps/jobs/main.go`, `apps/worker/main.go`, `apps/api/main.go` |

### HTTP routes

None. The async binaries expose no HTTP endpoints.

---

## 4. Logic flows

### Flow 1 — Scheduled-publish cutover (River-based, `apps/jobs` binary)

1. **Enqueue at HTTP time.** `apps/api/main.go:450` wires `RiverScheduledPublishEnqueuer` into `approvalServices.Publish`. When a user schedules a document for future publication, `publish_service.go` calls `EnqueueScheduledPublishTx` (`internal/modules/documents/approval/jobs/scheduled_publish_job.go:56`), which calls `client.InsertTx` inside the approvals transaction, inserting a River job row with `Queue="temporal"` and `ScheduledAt=effectiveDate` — atomically with the document row update.
2. **Binary startup.** `apps/jobs/cmd/metaldocs-jobs/main.go:22` calls `config.LoadJobsConfig` and `bootstrap.BuildJobsDependencies`. The bootstrap calls `MigrateRiverSchema` (ensuring River's schema exists), then constructs `ClientBundle` with the `temporal` queue and the worker factory that returns `NewWorkers(scheduler, db)`.
3. **Job delivery.** River's internal scheduler (in-process, built into the River client) fires the job at `ScheduledAt`. It calls `ScheduledPublishWorker.Work` (`internal/modules/documents/approval/jobs/scheduled_publish_job.go:33`).
4. **Execution.** `Work` sets `authz.WithBackgroundBypass(ctx)` and calls `service.RunScheduledPublishJob` (`internal/modules/documents/approval/application/scheduler_service.go:44`), which opens a transaction, loads the document state with `FOR UPDATE`, verifies the generation/version are still current (stale-job guard), then calls `publishScheduledDocumentTx` which updates `documents.status='published'` and emits a `governance_events` row.
5. **Shutdown.** On SIGTERM, `apps/jobs/main.go:52–57` calls `deps.River.Client.Stop(shutdownCtx)` with a 15 s timeout.

### Flow 2 — PDF generation via outbox (three stages across two binaries + API)

Stage A — **domain write → staging outbox** (in API process):

1. When an approval decision triggers PDF dispatch, `fanout.PDFDispatchAdapter.DispatchPDF` (`internal/modules/render/fanout/pdf_dispatch_adapter.go`) calls `pdfOutboxRepo.Enqueue` inside the approvals transaction, inserting a row into `metaldocs.pdf_dispatch_outbox` with status `pending`.
2. The `PDFOutboxWorker` goroutine (started at `apps/api/main.go:488`) polls `pdf_dispatch_outbox` every 5 s (`internal/modules/render/fanout/pdf_outbox_worker.go:55`). It claims up to 10 rows with `FOR UPDATE SKIP LOCKED`, then calls `messaging.Publisher.Publish` for each row, inserting a `docgen_v2_pdf` event into `metaldocs.outbox_events` with idempotency key `"docgen_v2_pdf:{tenantID}:{revisionID}"`.

Stage B — **outbox relay → worker** (in `apps/worker` process):

3. `platform/worker.Service.RunOnce` (`internal/platform/worker/service.go:42`) calls `consumer.ClaimUnpublished` (`internal/platform/messaging/outbox/postgres/consumer.go:25`), which uses a `FOR UPDATE SKIP LOCKED` CTE to claim up to `batchSize` rows from `metaldocs.outbox_events` and bump `attempt_count`, setting `next_attempt_at = now() + claimLease` as an in-flight lock.
4. For each claimed event with type `docgen_v2_pdf`, `service.go:62` calls `pdfRunner.Handle` (`internal/platform/worker/pdf_job_runner.go:68`). The runner extracts `PDFConvertPayload`, derives the docx S3 key (`tenants/{tenantID}/revisions/{revisionID}/frozen.docx` if not already in the payload), calls `GotenbergPDFClient.ConvertPDF` (`internal/platform/servicebus/gotenberg_pdf.go:70`), and calls `persister.WritePDF` to record the PDF S3 key and hash in the snapshot repository.
5. On success, `service.go:84` calls `consumer.MarkPublished`, setting `outbox_events.published_at = now()`. On failure, `markFailure` updates `next_attempt_at` with exponential backoff or sets `dead_lettered_at` if `attempt >= MaxAttempts`.

### Flow 3 — Async DOCX materialization (ADR 0015, two binaries + API)

Stage A — **Pin → materialize outbox** (in API process):

1. When the approval decision service calls `FreezeService.Pin`, the freeze service writes the snapshot and then calls `materializeOutboxRepo.Enqueue` (inside the freeze transaction), inserting a row into `metaldocs.materialize_dispatch_outbox` (migration 0215).
2. `MaterializeOutboxWorker` (started at `apps/api/main.go:491`) polls `materialize_dispatch_outbox` every 5 s, claims pending rows, and calls `Publisher.Publish` with `EventTypeMaterializeFanout`, inserting into `metaldocs.outbox_events`.

Stage B — **outbox → worker materialize** (in `apps/worker`):

3. `Service.RunOnce` claims `docx_materialize` events from `outbox_events`.
4. `MaterializeJobRunner.Handle` (`internal/platform/worker/materialize_job_runner.go:58`) extracts `MaterializeFanoutPayload`, calls `MaterializeInvoker.Materialize` (HTTP to docx-renderer fanout — outside any transaction), then opens a new transaction to call `WriteFinalDocxInTx` and `pdfOutbox.Enqueue` atomically. This enqueue writes a `pdf_dispatch_outbox` row, which re-enters Flow 2 Stage A for PDF generation.

### Flow 4 — In-API scheduler (maintenance jobs)

1. At API startup, `apps/api/main.go:523` calls `jobscheduler.New(deps.SQLDB, leaderID)` where `leaderID = hostname:pid`.
2. Each registered `JobConfig` gets its own goroutine via `Scheduler.Start` (`internal/modules/jobs/scheduler/scheduler.go:141`). Each loop iteration: (a) probe Postgres for connection-ratio backpressure (`probePressure`); (b) if not under pressure, call `acquireLease` — a Postgres function `metaldocs.acquire_lease($job, $leader, '5 minutes')` that returns `(acquired bool, epoch int64)`; (c) if acquired, spawn a heartbeat goroutine and call `cfg.Fn(jobCtx, epoch)`; (d) on completion, stop the heartbeat and call `releaseLease`.
3. The four registered jobs (all use `SkipOnPressure` policy): `stuck-instance-watchdog` (5 min interval), `idempotency-janitor` (15 min), `audit-integrity-validator` (1 h), `lease-reaper` (10 min).
4. On SIGTERM, the scheduler calls `drain()` which waits up to 30 s for in-flight jobs then up to 5 s more with forced cancellation before releasing leases.

### Flow 5 — Lightweight sweepers (in-API goroutines)

1. `jobs.StartSessionSweeper` (`internal/modules/documents/jobs/session_sweeper.go:12`) starts a goroutine with a 60 s ticker; calls `repo.ExpireStaleSessions(ctx, time.Now())` each tick with `authz.WithBackgroundBypass`.
2. `jobs.StartOrphanPendingSweeper` (`internal/modules/documents/jobs/orphan_pending_sweeper.go:12`) starts a goroutine with a 1 h ticker; calls `repo.DeleteExpiredPending(ctx, cutoff)` where cutoff = `now() - 24h`.
3. Both are started at `apps/api/main.go:568–569` and stopped via `defer stopSessions(); defer stopOrphans()`.

---

## 5. Dependencies

### `apps/jobs` (metaldocs-jobs binary)
**Outbound imports:**
- `github.com/riverqueue/river` — job queue framework
- `internal/modules/documents/approval/application` — `SchedulerService`, `NewServices`, `SQLEmitter`, `RealClock`
- `internal/modules/documents/approval/jobs` — `NewWorkers`
- `internal/modules/documents/approval/repository` — `NewPostgresApprovalRepository`
- `internal/platform/bootstrap` — `BuildJobsDependencies`
- `internal/platform/config` — `LoadJobsConfig`

**Inbound:** nothing imports `apps/jobs` (it is a binary).

### `apps/worker` (metaldocs-worker binary)
**Outbound imports:**
- `internal/modules/documents/application` — `FreezeService`, `NewSnapshotSchemaReader`
- `internal/modules/documents/repository` — `SnapshotRepository`, `FillInRepository`
- `internal/modules/render/fanout` — `NewClient`, `NewPDFOutboxRepository`
- `internal/modules/render/resolvers` — `NewRegistry`, `RegisterBuiltins`
- `internal/platform/bootstrap` — `BuildWorkerDependencies`
- `internal/platform/config` — `LoadWorkerConfig`
- `internal/platform/httpclient` — `NewInternalClient`
- `internal/platform/worker` — `NewService`, `NewPDFJobRunner`, `NewMaterializeJobRunner`, `NewSnapshotPDFPersister`

**Inbound:** nothing imports `apps/worker`.

### `internal/platform/jobs/river`
**Outbound:** `github.com/riverqueue/river`, `riverdriver/riverdatabasesql`
**Inbound:** `bootstrap/jobs.go`, `apps/jobs/main.go`, `apps/api/main.go`

### `internal/platform/worker`
**Outbound:** `platform/messaging`, `platform/servicebus`, `platform/config`, `database/sql`, `log/slog`
**Inbound:** `apps/worker/main.go` (primary), `bootstrap/worker.go` (indirect via WorkerDependencies)

### `internal/platform/messaging`
**Outbound:** `encoding/json`, `database/sql`, `time`
**Inbound:** `platform/worker`, `platform/bootstrap`, `modules/render/fanout` (pdf_outbox_worker, materialize_outbox_worker, pdf_dispatcher), `modules/documents/application/export_service.go`

### `internal/platform/servicebus`
**Outbound:** `crypto/sha256`, `io` (no MetalDocs imports)
**Inbound:** `platform/worker/pdf_job_runner.go`, `bootstrap/worker.go`, `bootstrap/api.go`

### `internal/modules/jobs/scheduler`
**Outbound:** `database/sql`, `log/slog`, `sync`, `time`, `github.com/lib/pq` (lease_reaper only)
**Inbound:** `apps/api/main.go`, `modules/jobs/audit_integrity_validator`, `modules/jobs/idempotency_janitor`, `modules/jobs/stuck_instance_watchdog` (all three import scheduler for the `JobFunc` type)

### `internal/modules/jobs/audit_integrity_validator`
**Outbound:** `modules/audit/domain`, `modules/jobs/scheduler`
**Inbound:** `apps/api/main.go`

### `internal/modules/jobs/idempotency_janitor`
**Outbound:** `database/sql`, `log/slog`, `modules/jobs/scheduler`
**Inbound:** `apps/api/main.go`

### `internal/modules/jobs/stuck_instance_watchdog`
**Outbound:** `database/sql`, `modules/documents/approval/application`, `modules/iam/authz`, `modules/jobs/scheduler`
**Inbound:** `apps/api/main.go`

### `internal/modules/documents/jobs`
**Outbound:** `modules/documents/repository`, `modules/iam/authz`, `log`, `time`
**Inbound:** `apps/api/main.go`

### `internal/modules/documents/approval/jobs`
**Outbound:** `github.com/riverqueue/river`, `modules/documents/approval/application`, `modules/iam/authz`, `database/sql`
**Inbound:** `apps/jobs/main.go` (primary — `NewWorkers`), `apps/api/main.go` (`NewScheduledPublishEnqueuer`)

---

## 6. Persistence

### Tables written by the async subsystems

| Table | Schema | Written by | Read by |
|-------|--------|-----------|---------|
| `metaldocs.outbox_events` | `outbox_events_pkey`, `outbox_events_idempotency_key_key` | `outbox/postgres/publisher.go` (API) | `outbox/postgres/consumer.go` (worker) |
| `metaldocs.pdf_dispatch_outbox` | status enum, unique (tenant_id, revision_id) | `fanout/pdf_outbox_repository.go` Enqueue (API), `platform/worker/materialize_job_runner.go` Enqueue (worker) | `fanout/pdf_outbox_repository.go` ClaimPending (API's PDFOutboxWorker) |
| `metaldocs.materialize_dispatch_outbox` | status enum, unique (tenant_id, revision_id); migration 0215 | `fanout/materialize_outbox_repository.go` Enqueue (API's FreezeService) | `fanout/materialize_outbox_repository.go` ClaimPending (API's MaterializeOutboxWorker) |
| `metaldocs.job_leases` | `job_leases_pkey (job_name)` | Postgres functions `acquire_lease`, `heartbeat_lease`, `release_lease` (called by scheduler) | `lease_reaper.go` (deletion), scheduler goroutines |
| `metaldocs.idempotency_keys` | — | (not in this area — written by API handlers) | `idempotency_janitor` (DELETE expired rows) |
| `metaldocs.audit_events` (subset) | — | `stuck_instance_watchdog` via `GovernanceEvent` emitter | `audit_integrity_validator` (read-only validation) |
| `governance_events` (public schema) | — | `lease_reaper.go` (INSERT on reclaim), `stuck_instance_watchdog` (INSERT alert) | — |
| `documents` | — | `scheduler_service.go` UPDATE status=published | — |
| River schema (configurable, default empty) | River internal tables | `MigrateRiverSchema` (up migration at jobs/API startup), River client | River client |

**Migration files:**
- `db/migrations/0215_materialize_dispatch_outbox.sql` — creates `metaldocs.materialize_dispatch_outbox`
- `db/baseline/0001_current_schema.sql` — contains `metaldocs.outbox_events`, `metaldocs.pdf_dispatch_outbox`, `metaldocs.job_leases`, and the three Postgres functions (`acquire_lease`, `heartbeat_lease`, `release_lease`)

---

## 7. Config & environment

### `apps/jobs` (`config.JobsConfig`)
| Env var | Default | Effect |
|---------|---------|--------|
| `METALDOCS_JOBS_ENABLED` | `true` | If false, binary exits immediately after logging |
| `METALDOCS_JOBS_RIVER_SCHEMA` | `""` (River uses its default schema) | Schema name for River tables |
| `METALDOCS_JOBS_TEMPORAL_MAX_WORKERS` | `10` | `MaxWorkers` for the `temporal` queue |
| `PGHOST`, `PGPORT`, `PGDATABASE`, `PGUSER`, `PGPASSWORD`, `PGSSLMODE` | — | Postgres connection (via `config.LoadPostgresConfig`) |

### `apps/worker` (`config.WorkerConfig`)
| Env var | Default | Effect |
|---------|---------|--------|
| `METALDOCS_WORKER_POLL_INTERVAL_SECONDS` | `10` | Ticker interval for the outbox-poll loop |
| `METALDOCS_WORKER_BATCH_SIZE` | `25` | `ClaimUnpublished` limit |
| `METALDOCS_WORKER_REVIEW_REMINDER_DAYS` | `14` | **Loaded but not used** — see section 10 |
| `METALDOCS_WORKER_RUN_ONCE` | `false` | If true, run a single batch then exit (CI/test mode) |
| `METALDOCS_WORKER_MAX_ATTEMPTS` | `5` | Dead-letter threshold |
| `METALDOCS_WORKER_RETRY_BASE_SECONDS` | `10` | Exponential backoff base |
| `METALDOCS_WORKER_RETRY_MAX_SECONDS` | `300` | Exponential backoff ceiling; also used as `claimLease` floor (min 5 min) |
| `METALDOCS_FANOUT_URL` | `""` | If set, enables `MaterializeJobRunner` |
| `METALDOCS_DOCX_RENDERER_SERVICE_TOKEN` | `""` | Bearer token for fanout HTTP calls |
| `METALDOCS_GOTENBERG_URL`, Gotenberg/MinIO env | — | Enable PDF runner (conditional; see `buildWorkerPDFConverter`) |

### In-API scheduler jobs
| Env var | Default | Effect |
|---------|---------|--------|
| `ENABLE_JOB_STUCK_INSTANCE_WATCHDOG` | enabled (opt-out) | Set to `false` to skip registration |
| `ENABLE_JOB_IDEMPOTENCY_JANITOR` | enabled | Set to `false` to skip |
| `ENABLE_JOB_AUDIT_INTEGRITY_VALIDATOR` | enabled | Set to `false` to skip |
| `ENABLE_JOB_LEASE_REAPER` | enabled | Set to `false` to skip |
| `AUDIT_RETENTION_DAYS` | `0` (disabled) | Positive int enables a 24 h audit purge goroutine |

---

## 8. Concurrency & async

### `apps/jobs`
- Single goroutine running the River client event loop; River internals manage worker goroutines up to `MaxWorkers=10`.
- Graceful shutdown via `Client.Stop(shutdownCtx)` with 15 s timeout.

### `apps/worker`
- Main loop: single `time.Ticker`-driven goroutine (`runWorkerLoop`); no parallelism within a batch (events processed sequentially in `RunOnce`).
- `MaterializeJobRunner.Handle` makes an outbound HTTP call (docx-renderer fanout) outside any transaction, then opens a new transaction for the write+enqueue pair — deliberate "HTTP then commit" pattern.

### In-API async goroutines
- **Scheduler:** one goroutine per registered job (`scheduler.go:149`), each running a ticker loop. In-flight jobs tracked via `sync.Mutex`-protected `map[*inFlightJob]struct{}`. Each running job gets its own heartbeat goroutine (`scheduler.go:241`).
- **PDFOutboxWorker / MaterializeOutboxWorker:** each spawned via `startOutboxWorker` (`apps/api/main.go:462`), which auto-restarts on error with exponential backoff up to 1 minute.
- **Session sweeper / orphan sweeper:** simple goroutines with `time.Ticker`, no restart logic.
- **Presence hub:** two goroutines — `hub.Run(ctx)` and `hub.RunHeartbeat(ctx)` — unrelated to async jobs.
- **Audit retention:** one goroutine with 24 h ticker, only started if `AUDIT_RETENTION_DAYS > 0`.

### Outbox claim mechanics
`outbox/postgres/consumer.go:25–126`: `ClaimUnpublished` runs inside a transaction — CTE selects candidates, UPDATE bumps `attempt_count` and sets `next_attempt_at = now() + claimLease`, RETURNING rows. Commit finalizes the claim. `FOR UPDATE SKIP LOCKED` prevents double-claiming across concurrent worker processes.

`pdf_dispatch_outbox` and `materialize_dispatch_outbox` use a `status='processing'` + `claimed_at` pattern with `ResetStaleClaims` to recover rows where the worker crashed after claiming but before marking dispatched.

---

## 9. Error handling & observability

### Error patterns
- `apps/jobs`: fatal on config/bootstrap failure (`log.Fatalf`); River handles job-level retries internally.
- `apps/worker/service.go`: per-event failure calls `markFailure`; retries with exponential backoff (`backoffDuration`); dead-letters after `MaxAttempts`; `runErr` accumulates all marking errors and is returned to the loop caller, which logs but does not exit.
- Scheduler jobs: errors are logged (`slog.Error`) and counted in `Metrics.ErrorsTotal`; the scheduler loop continues. Backpressure skip streak is logged as a warning at `maxSkipStreak=10`.
- `stuck_instance_watchdog.go`: uses `errors.Join` to accumulate per-instance failures; returns partial errors so the epoch is still released.

### Logging
- `apps/worker/service.go:88`: structured `log.Printf` per event: `worker_event event_id=... event_type=... attempt_count=... result=published trace_id=...`
- `apps/worker/service.go:93`: `worker_batch result=completed processed=... failed=... dead_lettered=... duration_ms=...`
- Scheduler jobs all use `slog.InfoContext` / `slog.ErrorContext` / `slog.WarnContext`.
- `apps/jobs/main.go:45`: `log.Printf("MetalDocs Jobs running (queues=temporal)")`.

### RFC 9457 / problem JSON
None in the async binaries. These processes do not expose HTTP.

### Metrics
- Scheduler exposes in-process `Metrics.Snapshot()` (runs/errors/skips per job name) but **does not export metrics to Prometheus or any external sink** — the snapshot is accessible programmatically but no HTTP metrics endpoint is wired for it.
- Worker logs batch completion stats but has no structured metrics export.

### Tracing
- `Event.TraceID` is propagated from the outbox through the worker log lines but is not forwarded to any tracing backend.

---

## 10. Legacy / duplication / smell flags

- **`ReviewReminderDays` field loaded but never consumed.** `config.WorkerConfig.ReviewReminderDays` is populated from `METALDOCS_WORKER_REVIEW_REMINDER_DAYS` (`internal/platform/config/worker.go:13,25,50`) and logged at startup (`apps/worker/cmd/metaldocs-worker/main.go:109`) but is not read anywhere in `internal/platform/worker/service.go` or any runner. The git log shows this was added in commit `2218bd016 feat: add worker notifications and review reminders`, suggesting a review-reminder feature was planned but never implemented. Severity: **medium** (dead config field with an associated env var that operators may set expecting behavior).

- **`apps/jobs` binary has no Dockerfile and is absent from the Docker Compose deployment.** `deploy/docker/` contains `api.Dockerfile` and `worker.Dockerfile` but no `jobs.Dockerfile`. `deploy/compose/docker-compose.yml` defines `api` and `worker` services but no `jobs` service. The binary is only invokable via `scripts/start-jobs.ps1` (local dev only). The compose `worker` service has `depends_on: docx-renderer: condition: service_healthy`, which the jobs binary does not depend on, suggesting the jobs binary was never integrated into the compose deployment. Severity: **high** (scheduled-publish cutover jobs will never fire in a containerised deployment).

- **`lease_reaper.go` queries `public.documents` by `id` for tenant attribution.** `internal/modules/jobs/scheduler/lease_reaper.go:37`: `SELECT doc.tenant_id FROM public.documents doc WHERE doc.id::text = d.job_name LIMIT 1`. The `job_leases.job_name` column is documented as a job identifier (e.g. `"stuck-instance-watchdog"`), not a document UUID. For the four registered scheduler jobs, this subquery will always return `NULL`, causing every reaped lease to log an error and be skipped (`rowErrs` appended). The reclaim counter will report 0 committed rows even when leases are reaped. Severity: **high** (governance audit rows are never written for scheduler-level lease reaps).

- **Duplicate outbox-relay pattern between `pdf_dispatch_outbox` / `materialize_dispatch_outbox` and `outbox_events`.** The system uses three distinct outbox tables: a domain-level staging outbox for PDF dispatch, a domain-level staging outbox for materialize dispatch, and a generic `outbox_events` relay. The two staging-outbox workers (`PDFOutboxWorker`, `MaterializeOutboxWorker`) in the API process exist solely to relay rows from the domain tables into `outbox_events`, which the worker binary then consumes. This is two-stage outbox chaining with duplicate claim/fail/retry logic across all three tables. Severity: **medium** (complexity/maintenance; `materialize_outbox_worker.go` and `pdf_outbox_worker.go` are structurally identical — RF-OB1 candidate).

- **`TODO(phase11)` comments in outbox consumer.** `outbox/postgres/consumer.go:37–38`: two TODO comments referencing `phase11` improvements to the claim query (heartbeat interval configurability and partial-index alignment). These are unresolved tech-debt markers. Severity: **low**.

- **`TODO(phase11)` comment in outbox publisher.** `outbox/postgres/publisher.go:29`: `event_id/idempotency_key are still TEXT-backed`. Severity: **low**.

- **`TODO(render)` comment in `pdf_outbox_repository.go`.** `internal/modules/render/fanout/pdf_outbox_repository.go:43`: the claim query intentionally lacks a tenant predicate; a comment says "thread tenant scope through the worker claim path before adding a tenant_id predicate". Severity: **low** (documented deferral, but means tenant isolation at the outbox layer is absent).

- **Backpressure probe uses `pg_stat_activity` active/max_connections ratio** with hard-coded thresholds (0.70 enter, 0.60 exit). This is a blunt proxy for DB pressure; it does not distinguish background vs. query connections and has no configuration surface. `scheduler.go:308–358`. Severity: **low** (functional but fragile under non-standard Postgres configurations).

- **`startOutboxWorker` restart loop in the API binary** (`apps/api/main.go:462–486`) re-starts `PDFOutboxWorker` and `MaterializeOutboxWorker` in-process after a returned error, with exponential backoff. However, `PDFOutboxWorker.Run` and `MaterializeOutboxWorker.Run` only return `nil` (they swallow all errors internally and log them). The restart loop will never trigger in practice — it is dead code for the restart path. Severity: **low**.

- **`apps/jobs` and `apps/api` both call `MigrateRiverSchema` at startup.** `apps/api/main.go:439` and `apps/jobs/main.go:35` (via `BuildJobsDependencies`) both run River schema migrations on startup. If the jobs binary runs concurrently with the API, this is idempotent (River migrations are `IF NOT EXISTS`) but represents unclear ownership of the River schema lifecycle. Severity: **low**.

- **Scheduler `leaderID` is `hostname:pid`** (`apps/api/main.go:840`). In a containerized deployment where multiple API replicas run, each replica will compete for leases using different IDs, which is the intended multi-leader protection. However, if an API pod restarts with the same hostname (common in Kubernetes), the new process will have a different PID, and `heartbeat_lease` will fail to extend the old epoch, eventually allowing the lease to be reaped or re-acquired. The 5 min lease TTL makes the gap bounded, but there is no documentation of this operational contract. Severity: **info**.

---

## 11. Wiki drift

No existing wiki doc covers this area. Section is recorded as **No existing doc**.

---

## 12. Open questions

1. **[runtime-unverified] Does `apps/jobs` actually run in production?** The binary has no Dockerfile and is absent from `deploy/compose/docker-compose.yml`. The scheduled-publish feature (`ScheduledPublishArgs.Kind() = "scheduled_publish_cutover"`) requires River jobs to be consumed. Without the jobs binary running, scheduled documents will accumulate River rows in the database and never become published. A live environment check would confirm whether (a) the binary is deployed outside the compose stack, (b) the River rows are consumed by a differently-named service, or (c) the scheduled-publish feature is non-functional in deployed environments.

2. **[runtime-unverified] River schema name in production.** `METALDOCS_JOBS_RIVER_SCHEMA` defaults to `""`. River with an empty schema uses the search path's current schema. If the search path is not explicitly set for the jobs/API process, the River tables may be created in `public` or `metaldocs` depending on the Postgres role's default. The baseline schema (`0001_current_schema.sql`) does not contain River tables, confirming they are created at runtime by `MigrateRiverSchema`.

3. **[runtime-unverified] `lease_reaper` governance writes.** Given the `public.documents` JOIN issue identified in section 10, it is unverified whether any `governance_events` rows have ever been written by the lease reaper in a running system. The `reclaimed` counter at `lease_reaper.go:119` may always be 0 when the four scheduler jobs are the only lease holders.

4. **[runtime-unverified] Worker `ReviewReminderDays` feature.** The git history shows this was added in `2218bd016`. Whether the intended review-reminder email/notification feature was intentionally deferred or accidentally omitted is not determinable from code alone.

5. **Outbox claim lease vs. retry max.** `bootstrap/worker.go:90–97` computes `claimLease = max(RetryMaxSeconds, 5 min)`. With default `RetryMaxSeconds=300` (5 min), `claimLease = 5 min`. This means a claimed-but-not-completed event will be re-available after 5 min if the worker crashes. For long-running materializations (docx-renderer latency is unknown), this could cause duplicate execution before the first attempt finishes. Runtime verification of typical materialization duration is needed to assess the risk.
