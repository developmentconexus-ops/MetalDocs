# F5.3 — staging dispatch on River (plan)

> Executes `spec.md` / contract §3. Engine: `superpowers:subagent-driven-development` (fresh subagent per
> task; sonnet implement+review; main session reviews + commits). Expand→contract ordering (contract §1).

## Load-bearing facts (from runtime map, 2026-07-04)

1. **3 enqueue sites, 2 binaries.** pdf: `documents/approval/application/decision_service.go:547` (api, in
   `runner.Do` tx) + `platform/worker/materialize_job_runner.go:90` (**worker**, in a `BeginTx`+`SeedTxTenant`
   tx). materialize: `documents/application/freeze_service.go:212` (api, in the Pin tx). `Enqueue(ctx, tx,
   tenantID, revisionID, contentHash)`.
2. **Poll worker today:** `fanout.StagingOutboxWorker` (`staging_outbox_worker.go`) — `Run`/`tick`/
   `dispatchOne`, constructed twice in `apps/api/.../main.go:923-968` (`startOutboxWorkers`, launched at
   `:546`, joined `:701`), each with a `buildEvent func(OutboxRow) messaging.Event` closure (pdf =
   `EventTypePDFConvert` key `docgen_v2_pdf:<t>:<rev>`; materialize = `EventTypeMaterializeFanout` key
   `materialize_fanout:<t>:<rev>`). Publisher = `messaging.Publisher` → `outbox_events ON CONFLICT
   (idempotency_key)`.
3. **River pattern to mirror:** `internal/modules/documents/approval/jobs/scheduled_publish_*.go` — Args w/
   `Kind()`, `Worker` embedding `river.WorkerDefaults[Args]`+`Work`, `NewWorkers()`→`river.AddWorker`,
   `EnqueueScheduledPublishTx(ctx, tx, ...)` asserting `tx.(*sql.Tx)` then `Client.InsertTx(..., &river.InsertOpts{Queue:"temporal",...})`.
4. **Client wiring:** `riverjobs.NewClientBundle(db, Config{Queues, PeriodicJobs, SkipUnknownJobCheck}, workers)`
   (`internal/platform/jobs/river/client.go:24`). api has an enqueue-only bundle (`main.go:506-517`,
   `SkipUnknownJobCheck:true`, workers nil). jobs subscribes `temporal`(MaxWorkers 10)+`maintenance`(2) and
   registers workers. **worker binary has NO River client** — must add an enqueue-only bundle.
5. **Both outbox tables** identical schema; `ON CONFLICT (tenant_id, revision_id)`; status CHECK
   `pending|processing|dispatched|failed`; `dispatched_at`/`dead_lettered_at` drive retention (F5.4).

## Task breakdown (ordered — expand, then contract)

### T1 (sonnet) — render-module River dispatch jobs (package + Args + Workers)
- New package `internal/modules/render/fanout/dispatchjobs` (render module owns staging outbox):
  - `Args` fields struct `{TenantID, RevisionID, ContentHash []byte, OutboxID string}` (JSON tags).
  - Two Args types `PDFDispatchArgs`/`MaterializeDispatchArgs` embedding the fields struct;
    `Kind()` → `"pdf_dispatch"` / `"materialize_dispatch"`.
  - A `Worker` per kind embedding `river.WorkerDefaults[…]`, closing over `messaging.Publisher` + the
    `*fanout.StagingOutboxRepository` (or the `MarkDispatched`/dead-letter port) + the per-kind `buildEvent`.
    Shared unexported `run(ctx, fields, buildEvent, repo, job)`:
    1. `ctx = authz.WithBackgroundBypass(ctx)` (system path, REQ-ASYNC-6).
    2. build event from fields → `pub.Publish(ctx, evt)`; on error return it (River retries) — but if
       `job.Attempt >= job.MaxAttempts` first dead-letter the row (`MarkFailed` finalize=true) then return err.
    3. on publish success → `repo.MarkDispatched(ctx, fields.TenantID, fields.OutboxID)`; if that errors,
       return err (River retries; terminal ⇒ dead-letter).
  - Move the two `buildEvent` closures verbatim from `apps/api/.../main.go:938-966` into this package (the
    per-kind event builders). No event-shape change.
  - `NewWorkers(pub, pdfRepo, materializeRepo) *river.Workers` registering both. `go build ./...` green.

### T2 (sonnet) — enqueuer + Enqueue RETURNING id
- `fanout.StagingOutboxRepository.Enqueue` → return `(insertedID string, err error)`; SQL `... ON CONFLICT
  (tenant_id, revision_id) DO NOTHING RETURNING id` (empty string when the conflict skipped — scan with
  `sql.ErrNoRows` → "" not error).
- New `dispatchjobs.Enqueuer{ Client *river.Client[*sql.Tx] }` with
  `EnqueueDispatchTx(ctx, tx db.Tx, kind Kind, tenantID, revisionID string, contentHash []byte) error`:
  1. `id, err := repo.Enqueue(...)` inside the caller tx (repo passed in or the enqueuer holds both repos).
  2. `if id == "" { return nil }` (dedup skip — no River job).
  3. assert `tx.(*sql.Tx)`; `Client.InsertTx(ctx, sqlTx, argsFor(kind, fields{…, OutboxID:id}), &river.InsertOpts{Queue:"temporal", MaxAttempts:maxAttempts})`.
  - nil/non-`*sql.Tx` fails loud (mirror `EnqueueScheduledPublishTx`).
- Update the 3 call sites' repo `Enqueue` usages of the return value (they currently ignore it). Unit-test
  the enqueuer's dedup-skip (no InsertTx when id=="") with a fake client. `go build ./...` green.

### T3 (sonnet) — wire enqueuer into the 3 sites + worker River client
- `apps/api/.../main.go`: build a `dispatchjobs.Enqueuer` from the existing `riverBundle.Client`; inject
  into the approval `DecisionService` (pdf) and the `FreezeService` (materialize) in place of the raw repo
  `Enqueue`. Keep the repos for the workers/retention.
- `apps/worker/cmd/metaldocs-worker/main.go`: add an **enqueue-only** River bundle
  (`riverjobs.NewClientBundle(deps.SQLDB, Config{Queues:nil or temporal-def, SkipUnknownJobCheck:true, PeriodicJobs:maintenance.PeriodicJobs()? NO — enqueue only}, nil)`) — mirror api's enqueue-only client
  (no queue subscription, no workers). Thread its enqueuer into `MaterializeJobRunner` (site C, pdf).
  > Decision: worker client passes **no** `PeriodicJobs` and **no** `Queues` (pure InsertTx). Confirm
  > `NewClientBundle` tolerates empty Queues for an insert-only client (api already runs one — check its
  > exact Config and copy it).
- Register the 2 dispatch workers in `apps/jobs/.../main.go` worker factory (`river.AddWorker` via
  `dispatchjobs.NewWorkers(pub, pdfRepo, materializeRepo)` merged into the returned `*river.Workers`).
  Requires the jobs binary to construct the `messaging.Publisher` + the 2 repos (check they're already
  built there for other jobs; if not, build from `db`). Workers land on `temporal` (already subscribed).
- `go build ./...` green; all 4 binaries buildable. Both dispatch paths now create River jobs.

### T4 (sonnet) — contract: delete the poll worker + api registration
- Remove `startOutboxWorkers` + the 2 `NewStagingOutboxWorker` goroutines + their `buildEvent` closures
  from `apps/api/.../main.go` (moved to the job package in T1); drop the `workerWG` staging entries if now
  unused (keep the WG if other workers use it — check). api starts with **no** staging poll goroutine.
- Delete `internal/modules/render/fanout/staging_outbox_worker.go` (`StagingOutboxWorker`).
- `fanout.StagingOutboxRepository`: delete `ClaimPending`, `ResetStaleClaims`; collapse `MarkFailed` to the
  finalize/dead-letter path only (River owns retry — the non-finalize `status='pending'`+`next_retry_at`
  branch is unreachable). Keep `MarkDispatched`, `CountDeadLettered`, `ReadState`, `inSeededTx`, the repos.
- Remove now-dead `StagingOutboxWorkerConfig` fields if fully unused (or leave `MaxAttempts` if the
  enqueuer reads it; keep config surface minimal). `go build ./...` green.

### T5 (sonnet) — backoff/claim census (spec gate 2)
- `grep` census recorded in evidence: 0 hits for the duplicated backoff (`1<<`…`30*time.Second`),
  `ClaimPending`, `ResetStaleClaims`, `StagingOutboxWorker`, `startOutboxWorkers`. Any residual explicitly
  allowlisted with a reason. (The pre-existing unrelated `claimLease` TODO comment in
  `outbox/postgres/consumer.go:37` is platform-messaging, not staging — note it as out-of-scope.)

### T6 (sonnet) — proofs (testdb, targeted -run)
- pdf + materialize dispatch equivalence integration tests (spec gate 1), testdb factory, `//go:build
  integration`: enqueue in a business tx (with a real River client on the testdb) → run the River worker →
  assert (a) an `outbox_events` row with the correct EventType/idempotency-key/payload, (b) duplicate
  enqueue ⇒ one staging row ⇒ one dispatch, (c) staging row `status='dispatched'` written tenant-seeded.
- Targeted `-run`; NOT the full suite (box). Execution deferred to M5-close live drive if no DSN (M1–M4
  precedent) — authored + compiled + vetted regardless; record the exact `-run` cmds in evidence.

## Test strategy
TDD where it bites: T2 enqueuer dedup-skip unit test; T6 dispatch equivalence is the acceptance. Build-green
after every task. Commit per task (or coherent pair) `feat/refactor(...)` scoped; never push.

## Files touched (census)
NEW `internal/modules/render/fanout/dispatchjobs/*` (Args, Workers, Enqueuer); MODIFY
`internal/modules/render/fanout/staging_outbox.go` (Enqueue RETURNING, MarkFailed collapse, delete
ClaimPending/ResetStaleClaims); DELETE `internal/modules/render/fanout/staging_outbox_worker.go`; MODIFY
`apps/api/cmd/metaldocs-api/main.go` (enqueuer inject, delete startOutboxWorkers), `apps/worker/cmd/
metaldocs-worker/main.go` (add River client + enqueuer), `apps/jobs/cmd/metaldocs-jobs/main.go` (register 2
workers), the 3 enqueue call sites (`decision_service.go`, `freeze_service.go`, `materialize_job_runner.go`),
`internal/platform/config/*` (staging worker config trim if unused). NEW integration tests.
