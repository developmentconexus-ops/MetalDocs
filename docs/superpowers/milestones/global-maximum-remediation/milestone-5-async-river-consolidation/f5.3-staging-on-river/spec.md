# F5.3 — staging dispatch on River (spec)

> **Milestone:** M5 · **Status:** in progress
> **Binding parent:** `../validation-contract.md` §3 (staging dispatch idempotency + backoff-deletion).
> On any conflict the **contract wins**; divergence is HS-7.
> **Rails:** ADR 0067 (River is the single async primitive; dispatch/scheduling moves, render *execution*
> under ADR 0009/0015 does NOT — contract §6/§8).

## Consumer contract (who consumes the output, and the shape required)

- **Consumer 1 — the `metaldocs-jobs` runtime.** Requires the two staging dispatch flows (pdf, materialize)
  registered as **River queue workers** on the existing business queue **`temporal`** (MaxWorkers 10, same
  queue as scheduled-publish/notifications — NOT `maintenance`). Each worker: builds the render event from
  the job Args, publishes it via `messaging.Publisher` (into `outbox_events`, `ON CONFLICT` idempotent),
  then `MarkDispatched` the staging row in a tenant-seeded tx. Shape: a `river.Worker[Args]` per kind
  (`pdf_dispatch`, `materialize_dispatch`), registered via `river.AddWorker`, mirroring
  `ScheduledPublishWorker` (`internal/modules/documents/approval/jobs/scheduled_publish_job.go`).
- **Consumer 2 — the three enqueue sites** (business-tx producers). Require an **enqueue-only** path that,
  inside the caller's business tx, inserts the staging outbox row **and** (only when the row is newly
  inserted) inserts a River dispatch job via `InsertTx` on the same `*sql.Tx` — transactional-outbox
  preserved. Sites: `documents/approval/.../decision_service.go:547` (pdf, in `metaldocs-api`),
  `documents/.../freeze_service.go:212` (materialize, api), `platform/worker/materialize_job_runner.go:90`
  (pdf, in `metaldocs-worker`). ⇒ **`metaldocs-worker` gains an enqueue-only River client** (no queue
  subscription, no workers, `SkipUnknownJobCheck:true`), mirroring the api enqueue-only client.
- **Consumer 3 — the milestone-validator.** Requires: the §3.2 backoff/claim census = 0 (no hand-rolled
  backoff, no `ClaimPending`/`ResetStaleClaims`, `StagingOutboxWorker` poll loop deleted); a pdf + a
  materialize dispatch equivalence integration proof (testdb); `go build ./...` green on all 4 binaries.

## Design decision (global-maximum — recorded)

**River job carries the work; the staging outbox row stays as the dedup + audit + retention record.**

- **Dedup unchanged:** the outbox row's `ON CONFLICT (tenant_id, revision_id) DO NOTHING` remains the
  single dedup point. `Enqueue` changes to `... DO NOTHING RETURNING id`. A **new** row id ⇒ also
  `InsertTx` one River dispatch job carrying `{TenantID, RevisionID, ContentHash, OutboxID}`. `ON CONFLICT`
  skip (no id) ⇒ **no** River insert. Duplicate enqueue = one row = one dispatch (contract §3.1).
- **River owns retry/backoff/DLQ:** the duplicated exp-backoff math (`worker.go:99,109`), `ClaimPending`,
  `ResetStaleClaims`, and the poll loop are **deleted**. River `MaxAttempts` + `RetryPolicy` schedule
  retries; terminal exhaustion ⇒ the worker dead-letters the row (`MarkFailed` finalize) and returns the
  error so River records `discarded` (inspectable DLQ, REQ-ASYNC-3). The row lifecycle simplifies to
  `pending` → `dispatched` (success) | `failed`+`dead_lettered_at` (terminal). Intermediate River retries
  do not touch the row; `processing`/`claimed_at`/`next_retry_at` become vestigial (columns kept — no
  schema drop; retention §4 reads only `status`/`dispatched_at`/`dead_lettered_at`).
- **NOT collapsing the double-outbox:** the staging row → `outbox_events` → render-worker chain (ADR
  0009/0015 execution) is preserved; folding staging into a single River layer is the contract §8 defer,
  **out of F5.3 scope** (would be HS-2/HS-6).

## What to implement (per contract §3)

| Kind | Enqueue site(s) | River Args | Worker body | Idempotency |
|---|---|---|---|---|
| `pdf_dispatch` | decision_service (api), materialize_job_runner (worker) | `{TenantID, RevisionID, ContentHash, OutboxID}` | build `EventTypePDFConvert` event (idempotency-key `docgen_v2_pdf:<t>:<rev>`) → `Publisher.Publish` → `MarkDispatched` | outbox `ON CONFLICT (tenant_id,revision_id)`; `outbox_events ON CONFLICT (idempotency_key)`; River unique per job |
| `materialize_dispatch` | freeze_service (api) | same shape | build `EventTypeMaterializeFanout` event (key `materialize_fanout:<t>:<rev>`) → `Publisher.Publish` → `MarkDispatched` | same |

`InsertOpts{Queue:"temporal", MaxAttempts:<staging cfg MaxAttempts>}`. Tenant seed preserved: `MarkDispatched`
/dead-letter run through the existing `inSeededTx` (`SeedTxTenant`, M3 F3.2). The `buildEvent` closures move
out of `apps/api/.../main.go:938-966` into the render-module job package (they are the per-kind event
builders); no event-shape change.

### Migration ordering (contract §1, expand/contract — do NOT reorder)
1. Add the 2 River dispatch workers (jobs) + the enqueue-only client in worker + the enqueuer at the 3
   sites; both binaries build and start. River jobs now created alongside the existing poll worker.
2. Remove `startOutboxWorkers` + the 2 `NewStagingOutboxWorker` goroutines from api main; delete
   `StagingOutboxWorker`, `ClaimPending`, `ResetStaleClaims`, the backoff math, the non-finalize retry
   branch. Backoff/claim census = 0.

## Non-goals

- **No render-execution change.** pdf/materialize render worker pipeline (ADR 0009/0015) untouched — F5.3
  moves dispatch scheduling only (contract §6). Collapsing the double-outbox = §8 defer (HS-2/HS-6 if attempted).
- **No staging-table schema drop.** Vestigial columns (`processing`/`claimed_at`/`next_retry_at`) kept.
  Retention (`dispatched_at`) is F5.4, not here.
- **No event-shape / idempotency-key change.** Same `messaging.Event` the poll worker built.
- **No new capability, no openapi edit, no capability-registry-size change.**
- **No metrics/Prometheus** (M8 defer, contract §8).

## Validation Gate (acceptance — all must hold)

1. **pdf + materialize dispatch equivalence integration tests (testdb, targeted `-run`)** (contract §3.3):
   - enqueue-in-business-tx → the River dispatch job runs → the render event is published to `outbox_events`
     (correct EventType + idempotency-key + payload, identical to the pre-migration `buildEvent`).
   - duplicate enqueue (same tenant+revision) ⇒ **one** outbox row ⇒ **one** dispatch (`ON CONFLICT` holds).
   - the `MarkDispatched` write runs tenant-seeded (RLS engages, M3 F3.2) — row ends `status='dispatched'`.
2. **§3.2 backoff/claim census = 0** (grep, recorded in evidence): no `1<<`…`*time.Second` backoff in
   `render/fanout`; `StagingOutboxWorker`/`ClaimPending`/`ResetStaleClaims` gone; `startOutboxWorkers` gone
   from api main. Any retained thin adapter explicitly allowlisted with a reason.
3. **All 4 binaries build + start:** `go build ./...` green; api starts with **no** staging poll goroutine;
   jobs registers the 2 dispatch workers on `temporal`; worker has an enqueue-only River client.
4. **Transactional-outbox preserved:** the River `InsertTx` shares the caller's business tx at all 3 sites;
   a nil/non-`*sql.Tx` fails loud (mirror `EnqueueScheduledPublishTx`).
5. Section-by-section match to contract §3. No divergence, or divergence surfaced as HS-7.

## Interview record

No operator interview — contract §3 (operator-committed at D4) is the locked spec; F5.3 is its execution.
The only design freedom (River-carries-work vs periodic-claim-loop) is resolved above toward the contract's
explicit "River job dequeue (no bespoke poll loop)" (§3.1 dispatch-trigger row) — the periodic-claim
alternative is rejected as it retains `ClaimPending`, which §3.2 requires removed.

## ADR

No new ADR — F5.3 executes ADR 0067's accepted decision. The double-outbox collapse it deliberately does
NOT do is recorded as a contract §8 bounded defer.
