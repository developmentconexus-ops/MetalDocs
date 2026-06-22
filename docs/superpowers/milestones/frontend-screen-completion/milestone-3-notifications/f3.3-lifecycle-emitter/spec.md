# Feature F3.3 — Spec (lifecycle-emitter / domain-event pattern)

> **Milestone:** 3 — Notifications (full-stack; surface + document-lifecycle emitters)  ·  **Folder:** `f3.3-lifecycle-emitter`
> **Status:** **DRAFT — awaiting operator approval (pre-code).** Per the milestone gate, code may begin only after the operator approves this contract.
> **Governing decision:** [ADR-0044](../../../../../wiki/decisions/0044-domain-event-pattern-and-river-dispatch.md) (domain-event pattern, River dispatch; supersedes ADR-0043 §6). Evidence: `research-and-design.md`.

> This is the feature's **contract**, written and approved **before any code**. It defines what the
> feature must do and how it will be proven — not how it will be built (`plan.md`). The
> milestone-validator judges the feature against *this* file (C1).

## Interview record (fail-closed gate)

Producer design driven by (a) ADR-0044 (the operator-ratified root-cause redesign), (b) a read-only
recon of the live backend at HEAD (River presence + same-tx enqueue, the five emit-site tx handles,
the `CDFieldReader` port, the four CI guards that govern this seam), and (c) operator decisions below.

| # | Question | Answer (recon-grounded) |
|---|----------|--------------------------|
| 1 | Is River actually present + can we enqueue inside the state-change tx? | **Yes.** `github.com/riverqueue/river v0.37.1` (go.mod:16). The client is `*river.Client[*sql.Tx]` (`internal/platform/jobs/river/client.go:23`). Same-tx enqueue is the **established** pattern: `EnqueueScheduledPublishTx` calls `e.Client.InsertTx(ctx, sqlTx, args, &river.InsertOpts{...})` inside `runner.Do(ctx, func(tx *sql.Tx) …)` (`approval/jobs/scheduled_publish_job.go:60–76`; live use at `publish_service.go:266–276`). River's same-tx insert **is** the transactional-outbox guarantee. **No new infra, no `domain_events` table** (ADR-0044 §2). |
| 1b | **`db.Tx` vs `*sql.Tx` (operator question).** | The **application/service + port layer speaks `db.Tx`** (the codebase abstraction); only the **River infra adapter** narrows to the concrete `*sql.Tx` the River client is generically bound to. The established `RiverScheduledPublishEnqueuer.EnqueueScheduledPublishTx(ctx, tx db.Tx, …)` takes `db.Tx`, then does one localized `sqlTx, ok := tx.(*sql.Tx)` assertion (`scheduled_publish_job.go:60–64`) before `Client.InsertTx(ctx, sqlTx, …)`. `runner.Do` happens to hand the closure a concrete `*sql.Tx` (`publish_service.go:54`), but the **enqueuer port** F3.3 defines takes `db.Tx` to stay domain-clean — the `*sql.Tx` assertion is isolated to the one infra adapter, mirroring the existing enqueuer. *(The prior draft over-stated `*sql.Tx` at the port boundary — corrected.)* |
| 2 | Are the five emit sites tx-capable? | **Yes — all five run inside `runner.Do(ctx, func(tx *sql.Tx) error{…})`:** `publish_service.go:54` (PublishApproved, approved→published), `supersede_service.go:48` (PublishSuperseding), `obsolete_service.go:47` (MarkObsolete), `decision_service.go:190` (RecordSignoff — terminal `InstanceApproved` UPDATE @ :408, terminal `InstanceRejected` UPDATE @ :464). The enqueue sits next to the existing `events.Emit` audit call in each. |
| 3 | Reader vs author recipient data in scope at emit? | **Reader events** (`document.published/.superseded/.obsoleted`) need only `controlled_document_id` — readable in-tx from the `documents` row (the `LoadDocumentAreaCode` precedent reads `documents.controlled_document_id` via `s.cdRead`/in-tx, `documents/application/document_area.go:54`). The recon "blocker" that supersede/obsolete lack `SubmittedBy` is **moot**: they are reader events, they do not carry a submitter. **Author events** (`document.approved/.rejected`) need `submitted_by` — `instance.SubmittedBy` **is** in scope at `decision_service.go:198/291`. ADR-0044 §3 confirmed against runtime. |
| 4 | Where does the domain-event **contract type** (River job args + `event_type` constants + enqueuer port) live? **(operator challenged the first answer — resolved by guard + precedent, not preference.)** | **The producer module's top-level `domain` package: `internal/modules/documents/domain`.** Decisive evidence: the **`module-boundaries` guard** (`scripts/check-module-boundaries.ps1:43–54`) permits a cross-module import **only** when the imported path is exactly `<module>/domain` — any deeper layer (`approval/jobs`, `approval/application`, even `approval/domain`) is a **violation**. And `<module>/domain` is the architecture's **pervasive, sanctioned cross-module contract surface**: `documents/domain` is imported **47×**, `iam/domain` **140×**, `controlleddocuments/domain` **38×** today. So the notifications worker legally imports `documents/domain` for the contract type; the DDD-owner (documents/approval produces the events) keeps ownership; zero new infra concept. **The earlier "neutral `internal/platform/jobs/domainevents`" answer is withdrawn** — `platform` is module-*agnostic infrastructure* (platformboundary), so putting document-lifecycle **domain vocabulary** there fights the architecture's intent; `documents/domain` is strictly better and is the established pattern. **Decision made — no operator action needed.** |
| 5 | One River job kind or five? | **One job kind + one worker, discriminated by `event_type`** (matches the re-scoped milestone.md: "one fan-out worker that switches on `event_type`"). Args = the domain-event envelope: `EventType`, `TenantID`, `DocumentID`, `ControlledDocumentID *string` (reader events), `SubmittedBy *string` (author events), `ResourceType`, `ResourceID`, `OccurredAt`, and a stable `EventID` (uuid minted at emit → the idempotency key). Five **typed `event_type` values** = the five canonical domain events; the typed envelope is the medium-fat payload (ADR-0044 §1/§3). |
| 6 | Idempotency key for at-least-once redelivery? | The F3.2 partial unique index `(recipient_user_id, source_event_id) WHERE source_event_id IS NOT NULL`. The worker sets `source_event_id = EventID` (the per-emit uuid carried in the job). River at-least-once redelivery → `ON CONFLICT DO NOTHING` → no duplicate inbox row (ADR-0044 §4; F3.2 spec Q1). |
| 7 | Does adding the enqueue break the `outboxpair` guard? | **No — it reinforces it.** `outboxpair.go` requires every approval-state mutation to pair with `events.Emit` (the **audit** event). ADR-0044 keeps the audit Emit **untouched**; the River enqueue is **additive** beside it. The audit pairing still holds; no `//cilint:allow-no-outbox` needed. |
| 8 | `postcommitaudit` / off-tx (H-PRE-1)? | The **enqueue** is in-tx, **before** Commit (correct — `postcommitaudit` only flags audit sinks *after* Commit). The **fan-out worker** runs after commit in a separate River-invoked function — it writes **notification** rows (not an audit sink in `postCommitAuditSinks`), and reads `v_cd_obligated_readers` off any publish-path lock. **H-PRE-1 satisfied** (no authz-recording read inside the lock-holding tx). |
| 9 | Reader fan-out cross-module read — `hgcrossmodule`? | The worker reads only the **published** `metaldocs.v_cd_obligated_readers` view (M2 contract) for readers, and the `submitted_by` carried in the job for authors. **No base-table read** of CD/approval/documents. Register `"notifications"` ownership already done in F3.2 (`hgOwnerByTable`). `hgcrossmodule` = 0. |
| 10 | Worker registration point? | `main.go` builds the River client via `riverjobs.NewClientBundle(deps.SQLDB, cfg, workers *river.Workers)` (`main.go:484`). F3.3 composes a `*river.Workers` that registers the notifications fan-out worker (and preserves any existing approval workers) and passes it in. Exact composition (combine-workers factory vs. per-module `AddWorker`) is a **`plan.md`** decision within this constraint. notifications currently has **no** worker/jobs file (only the F3.2 read surface) — F3.3 adds one. |

## Consumer contract (FIRST — before any producer)

- **Consumer:** the **notifications fan-out River worker** (new, this feature) is the consumer of the
  five domain-event jobs. Its *output* is rows in the F2/F3.2-owned `metaldocs.notifications` table,
  which the F3.2 read surface and the F3.4 screen already consume. So the binding downstream contract is
  the **F3.2 `Notification` row shape** (`id, tenant_id, recipient_user_id, event_type, resource_type,
  resource_id, title, message, status, created_at, read_at?, source_event_id`) — F3.3 must produce rows
  structurally valid for that surface (correct `event_type`, a real pt-BR `title`/`message`,
  `status='PENDING'`, `source_event_id` set).
- **The domain-event contract (the wire seam between producer and consumer):** the
  `internal/modules/documents/domain` `LifecycleEventArgs` envelope (Q5) + the five `event_type`
  constants + the `LifecycleEventEnqueuer` port (Q4). The producer (approval, via its infra adapter)
  binds to it to enqueue; the notifications worker imports `documents/domain` to fan out. This is the
  cross-module contract boundary (ADR-0044 §5), placed in the architecture's sanctioned `<module>/domain`
  contract layer (module-boundaries-legal; 47× precedent).
- **The five domain events (ADR-0044 §3 — the M3 bundle):**

  | `event_type` | Emit site (in-tx, beside the audit Emit) | Carries | Recipient set (resolved by the worker) | Inbox message (pt-BR) |
  |---|---|---|---|---|
  | `document.published` | `publish_service.go` PublishApproved (approved→published) | `controlled_document_id` | obligated readers via `v_cd_obligated_readers` | "Novo documento controlado para leitura" |
  | `document.superseded` | `supersede_service.go` PublishSuperseding | `controlled_document_id` | obligated readers (same) | "Documento substituído por nova revisão" |
  | `document.obsoleted` | `obsolete_service.go` MarkObsolete | `controlled_document_id` | obligated readers (same) | "Documento que você lê foi obsoletado" |
  | `document.approved` | `decision_service.go` @ terminal `InstanceApproved` (:408) | `submitted_by` | the submitter (1) | "Seu documento foi aprovado" |
  | `document.rejected` | `decision_service.go` @ terminal `InstanceRejected` (:464) | `submitted_by` | the submitter (1) | "Documento rejeitado — ajustes solicitados" |

  `document.approved`/`.rejected` fire **only** at the instance terminal transition — the clean
  final-outcome signal the per-stage `signoff_recorded` audit event cannot give. The eligibility-failure
  `signoff.rejected` audit event is **not** a producer (honest absence).

## What this feature implements

1. **`internal/modules/documents/domain/notification_events.go` (new — in the producer module's
   top-level domain, the sanctioned cross-module contract layer per Q4):** the `LifecycleEventArgs`
   River job-args struct (the envelope from Q5; satisfies `river.JobArgs` via a `Kind() = "notification_fanout"`
   string method — **no `river` import** in this file, so `domain` stays infra-free), the five `event_type`
   string constants, and the producer **enqueuer port** interface `LifecycleEventEnqueuer` (in-tx method
   `EnqueueLifecycleEventTx(ctx, tx db.Tx, args LifecycleEventArgs) error` — `db.Tx`, not `*sql.Tx`, per Q1b).
   The notifications worker imports this package (`documents/domain`, 47× precedent — module-boundaries-legal).
2. **River producer adapter** (`documents/approval/infrastructure`, mirroring `RiverScheduledPublishEnqueuer`):
   implements `documentsdomain.LifecycleEventEnqueuer` — takes `db.Tx`, does the one localized
   `tx.(*sql.Tx)` assertion (Q1b), calls `Client.InsertTx(ctx, sqlTx, args, …)`. The `river`/`*sql.Tx`
   coupling lives **only** here, not in domain or the services.
3. **Five additive in-tx enqueues** — at each emit site, *immediately after* the existing in-tx
   `events.Emit(...)` audit call, call the injected `LifecycleEventEnqueuer.EnqueueLifecycleEventTx(ctx, tx, args)`
   with a freshly minted `EventID` (uuid) and the in-scope recipient id (`controlled_document_id` read
   in-tx for reader events; `instance.SubmittedBy` for author events). **No other change** to the five
   services — publish/approval **semantics unchanged** (existing assertions hold; only diff = the enqueue call).
   - `publish_service.go` PublishApproved → `document.published`
   - `supersede_service.go` PublishSuperseding → `document.superseded`
   - `obsolete_service.go` MarkObsolete → `document.obsoleted`
   - `decision_service.go` @ `InstanceApproved` → `document.approved`; @ `InstanceRejected` → `document.rejected`
4. **`internal/modules/notifications/infrastructure/fanout_worker.go` (new):** a
   `river.Worker[documentsdomain.LifecycleEventArgs]` that, **after commit**, switches on `EventType`:
   - reader events → query the published `metaldocs.v_cd_obligated_readers` for the carried
     `controlled_document_id` → one recipient per obligated reader;
   - author events → the single carried `submitted_by`;
   then renders the pt-BR `title`/`message` per the bundle table and inserts one row per recipient into
   `metaldocs.notifications` with `source_event_id = EventID`, `status='PENDING'`,
   `ON CONFLICT (recipient_user_id, source_event_id) DO NOTHING` (idempotent). Reads only the published
   view (hgcrossmodule-clean); runs off any publish-path lock (H-PRE-1).
5. **Worker registration** at the composition root (`apps/api/cmd/metaldocs-api/main.go`): compose the
   `*river.Workers` passed to `NewClientBundle` so the notifications fan-out worker is registered
   (preserving existing workers — today the client is built with `nil` workers at `main.go:484`). Exact
   factory shape per `plan.md`.
6. **`tests/integration/testdb/factory.go`** (if needed): a helper to drive a real emit path end-to-end
   (publish/supersede/obsolete/approve/reject) so the integration tests assert the produced rows — reuse
   the existing approval/CD factories; add only what the per-event tests need.

## Non-goals (mandatory — anything here in the diff is scope drift, validator C6)

- **No change to publish/approval *semantics*** — the only edit to the five services is the additive
  enqueue (HS-2 lift is scoped to exactly that). No reordering, no new business logic, no changed return
  shapes. Existing `publish_service.go`/`decision_service.go`/etc. service-test assertions must still pass.
- **No `domain_events` table, no CDC/Debezium, no external broker** — River jobs are the outbox (ADR-0044 §2).
- **No new audit event / no edit to `governance_events` payloads** — the audit log stays untouched (ADR-0044 §1). The enqueue is *additive*, the existing `events.Emit` is unchanged.
- **No migration to existing audit events** — only the five new domain events follow this pattern (strangler-fig, ADR-0044 §6).
- **No approver-routing / template-lifecycle emitter, no channels (email/push), no preferences, no SSE rebuild** — parked emitter mission.
- **No change to `v_cd_grantee` / `v_cd_obligated_readers`** (diff on their migrations empty) or any search/CD/iam base table.
- **No FE work** (F3.4).
- **No new capability** — read surface auth is F3.1/F3.2's `CapNotificationRead`; the worker is system-internal (no HTTP route).

## Validation Gate (concrete — to be approved before code)

| Acceptance criterion | Named test / proof command | Real vs fixture |
|----------------------|----------------------------|-----------------|
| **`document.published` → reader rows** — publishing a CD yields exactly one `PENDING` notification per obligated reader (== `v_cd_obligated_readers`), zero for non-readers | `go test -tags integration ./internal/modules/notifications/... -run TestFanout/published_to_obligated_readers` | real (live PG) |
| **`document.superseded` → reader rows** | `…-run TestFanout/superseded_to_obligated_readers` | real |
| **`document.obsoleted` → reader rows** | `…-run TestFanout/obsoleted_to_obligated_readers` | real |
| **`document.approved` → submitter row** — terminal approval yields exactly one notification for `submitted_by`, zero for the approver/others | `…-run TestFanout/approved_to_submitter` | real |
| **`document.rejected` → submitter row** — terminal rejection only (not per-stage signoff) | `…-run TestFanout/rejected_to_submitter` | real |
| **Idempotent redelivery** — re-running the worker for the same `EventID` inserts no duplicate row | `…-run TestFanout/redelivery_is_noop` | real |
| **In-tx enqueue + atomicity** — each emit site enqueues exactly one job in the state-change tx; if the tx rolls back, no job is visible (outbox guarantee) | `…-run TestEmit/enqueues_one_job_per_action` + a rollback case | real |
| **Publish/approval semantics unchanged** — the existing approval/publish service tests pass unmodified | `go test ./internal/modules/documents/approval/...` (no assertion edits) | real |
| **`outboxpair` still green** — audit Emit retained at all five sites | `go test ./tools/cilint/...` (outboxpair = 0) | real |
| **`module-boundaries` green** — the only cross-module import is notifications→`documents/domain` (exactly `<module>/domain`, the sanctioned layer); the new `documents/domain` event file imports no `river` | `.\scripts\check-module-boundaries.ps1` → `OK`; `grep -L riverqueue documents/domain/notification_events.go` | real |
| **`hgcrossmodule` green** — worker reads only the published view | cilint hgcrossmodule = 0 | real |
| **`api-lint -strict` = 0** (no new routes; surface unchanged) | `go run ./scripts/api-lint -strict api/openapi/v1/openapi.yaml .` → `0 violation(s)` | real |
| **All 6 CI guards green** | `go vet ./...`; `go test ./tools/cilint/...`; `go build ./...` | real |
| **`go build` / `go test ./...` green** | `go test ./...` | real |
| **System migrates + runs** | `.\scripts\check-system-runnable.ps1` | real |

> TDD: each per-event fan-out test is written **failing-first** against the live-PG worker before the
> worker/enqueue exists, then implemented to green. The reader-set and submitter assertions run against
> real Postgres + the real emit paths — no fixture-only substitution for the recipient-resolution proofs.

## ADR needed?

- [x] **No new ADR — covered by [ADR-0044](../../../../../wiki/decisions/0044-domain-event-pattern-and-river-dispatch.md).**
  ADR-0044 already records the domain-event pattern, River dispatch, the five canonical events, the
  fan-out worker, the idempotent inbox, module ownership, and the strangler-fig rollout. F3.3 *implements*
  that decision. The one refinement here — the contract-type **placement** in `internal/modules/documents/domain`
  (Q4) — is a mechanical application of the `module-boundaries` guard (only `<module>/domain` is
  cross-module-importable) + ADR-0044 §5 ("the event is the contract"), recorded in `evidence.md`, not a
  new durable decision. No open decision remains before code.
