# ADR 0044 — Domain-event pattern (River-job dispatch) + notifications fan-out; supersedes ADR-0043 §6

> **Status:** Accepted 2026-06-22
> **Last verified:** 2026-06-22
> **Deciders:** leandrotca.work (operator), MetalDocs backend
> **Context window:** Mission `frontend-screen-completion` · Milestone M3 (Notifications) · Feature F3.3 (re-scoped).
> **Supersedes:** **ADR-0043 §6** (recipient resolution via additive `v_approval_instance_submitter` +
> `v_document_cd_mapping` views read by a projector over the audit log). The rest of ADR-0043 (the
> `notifications` module, `CapNotificationRead`, the read surface, open `event_type`, the
> per-recipient idempotent inbox table) **stands**.
> **Related ADRs:** 0039 (cross-module read boundary), 0040 (`v_cd_obligated_readers`), 0043
> (notifications module), 0015 (transactional outbox / PDF dispatch precedent), 0022 (authz tiers).

## Context

ADR-0043 specced notifications as a **projector over `governance_events`** plus two compensating
published views to recover the submitter and the document→CD mapping. F3.3 pre-spec recon
(2026-06-22) established this is a **workaround over a category error**:

- `governance_events` is an **actor-centric audit log** — it records `actor_user_id` (*who acted*),
  immutably, in the state-change tx. It does **not** carry the business **outcome** or **who should be
  notified**.
- Consequence: `signoff.rejected` is an **eligibility-failure** event (not document rejection);
  `signoff_recorded` fires on **every** signoff with **no terminal-outcome marker**; submitter +
  controlled-document-id live on producer base tables. A notifications consumer must reverse-engineer
  outcome + recipients via extra views — the workaround the operator rejected, directing a root-cause fix.

Industry practice (DDD, transactional outbox, CloudEvents, Knock/novu notification-infra, modular-
monolith event ownership; research synthesis in
`docs/superpowers/milestones/frontend-screen-completion/milestone-3-notifications/f3.3-lifecycle-emitter/research-and-design.md`)
distinguishes three event kinds and keeps them distinct: **audit** (who acted), **domain** (what
outcome happened, carrying consumer-relevant data), **integration/cross-module** (the public contract
between modules). We conflated audit with domain.

## Decision

### 1 — Three distinct event kinds; the audit log is untouched

`governance_events` remains the **audit log**, unchanged (no payload widening, no new consumer-driven
columns). We **add** a **domain-event** layer. Domain events are owned by the module where the business
action happens and carry the **outcome + the recipient-resolution identifiers the producer already has
in scope** — never full entity state (medium-fat, not thin-ID, not full-ECST).

### 2 — River jobs ARE the transactional outbox + dispatcher

River (Postgres-backed queue, already in production for scheduled publish — ADR-0015 lineage) is the
domain-event dispatch mechanism. At each **canonical business action**, in the **same `*sql.Tx`** as
the state change, the owning module **enqueues a typed River job** whose args are the domain-event
payload. River's same-tx enqueue is the transactional-outbox guarantee (a job is invisible to workers
until the state change commits); no separate `domain_events` table, no Debezium/CDC, no external broker
(we are a single-process modular monolith). A dedicated **`domain_events` table is explicitly rejected**
for now (no near-term second consumer or replay need); revisit if one appears.

### 3 — Canonical per-outcome domain events (the M3 bundle)

The owning modules emit these typed domain events at their canonical transition points (the same code
sites that already emit the audit event, in the same tx):

| Domain event | Owner module · emit site | Payload carries | Recipient set (resolved by the consumer worker) |
|---|---|---|---|
| `document.published` | documents/approval · `publish_service.go` (approved→published) | tenant, document_id, **controlled_document_id**, revision | obligated readers via published `v_cd_obligated_readers` |
| `document.superseded` | documents/approval · `supersede_service.go` | tenant, new+old document_id, controlled_document_id | obligated readers (same) |
| `document.obsoleted` | documents/approval · `obsolete_service.go` | tenant, document_id, controlled_document_id | obligated readers (same) |
| `document.approved` | documents/approval · `decision_service.go` at the **`InstanceApproved`** transition | tenant, document_id, **submitted_by** | the submitter (1 recipient) |
| `document.rejected` | documents/approval · `decision_service.go` at the **`InstanceRejected`** transition | tenant, document_id, **submitted_by** | the submitter (1 recipient) |

`document.approved`/`document.rejected` are emitted **only at the instance terminal transition** — this
is the clean "final approval / rejection" signal that `signoff_recorded` could not provide. The
eligibility-failure `signoff.rejected` audit event is **not** a notification producer (honest absence).
`controlled_document_id` is resolved at emit time via the approval module's existing `CDFieldReader`
port (`s.cdRead`) — within module boundaries, no raw cross-module read. `submitted_by` is already in
scope (`instance.SubmittedBy`). **This eliminates both ADR-0043 §6 views.**

### 4 — Notifications fan-out worker (idempotent, off-tx)

A new River worker in the `notifications` module consumes each domain-event job, resolves the recipient
set (reader events → the **published** `v_cd_obligated_readers` contract; author events → the
`submitted_by` carried in the event), and inserts one **fan-out-on-write** row per recipient into the
F3.2 `metaldocs.notifications` table. Idempotency is the F3.2 partial unique index
`(recipient_user_id, source_event_id)` — the worker keys `source_event_id` on the domain-event id, so
at-least-once redelivery is a no-op. The worker runs **after commit** (off any publish-path lock —
H-PRE-1 satisfied). Reader fan-out reads only the published view (no base-table read; `hgcrossmodule`
clean).

### 5 — Module ownership & the contract boundary

Domain-event type definitions (the typed River job args + a stable `event_type` string + the producer's
enqueuer port) live in the **owning module**, specifically in its **top-level `domain` package**
(`internal/modules/<owner>/domain`). This is mandated by the `module-boundaries` guard
(`scripts/check-module-boundaries.ps1`): a cross-module import is legal **only** when the imported path
is exactly `<module>/domain` — deeper layers (`.../application`, `.../jobs`, even a nested
`.../approval/domain`) are violations. So a consumer module legally imports only the producer's
top-level `domain` (the architecture's sanctioned contract surface — `documents/domain` is imported 47×,
`iam/domain` 140× today). The args struct stays **infra-free** (satisfies `river.JobArgs` via a `Kind()`
string method with **no `river` import** in the `domain` file); the `river`/`*sql.Tx` coupling lives in
the producer's infra adapter (mirroring `RiverScheduledPublishEnqueuer`). The args/port use `db.Tx`, not
`*sql.Tx`, at the boundary — the concrete-tx assertion is isolated to that one adapter. A neutral
`internal/platform/...` package is **rejected** for the event contract: `platform` is module-agnostic
infrastructure (the `platformboundary` guard), so domain vocabulary does not belong there. The
notifications module **subscribes** (registers the worker) — it does not read approval/CD/documents base
tables. The event is the cross-module contract. `event_type` stays an **open string** (ADR-0043 §2)
evolved **additively** (new fields optional/defaulted; breaking change ⇒ new `type` name) per
CloudEvents-style forward-compatibility.

### 6 — Incremental, strangler-fig rollout (the operator's "redesign, not fully")

This pattern is introduced **alongside** the audit log, **module-by-module**, starting with the five
document-lifecycle actions notifications needs. There is **one canonical emit point per business
action** (the existing audit-emit site). Existing audit events are **not** migrated; future domain
events follow this ADR as the standard. No big-bang rewrite.

## Consequences

- **Positive:** root-cause fix — events carry truth, no reverse-engineering; both ADR-0043 §6 views
  eliminated; no new infra (River + the F3.2 table already exist); clean per-module event ownership;
  the audit log stays a pure audit log; pattern is the reusable standard for the parked emitter mission.
- **Negative / watch:** the **publish path is now edited** — `publish_service.go` and the other four
  emit sites gain an in-tx River enqueue. The milestone's "publish path sacred" rule (HS-2) is
  **explicitly lifted by the operator** for this redesign (2026-06-22). The enqueue must be additive
  (no change to existing publish/approval *semantics*), in the same tx, and covered by the existing
  service tests + new emit assertions. Company-scope reader fan-out volume (R4) is bounded only if a
  real measurement shows a problem.
- **Verification:** each emit site asserts exactly one domain-event job enqueued per action in the
  state-change tx; the fan-out worker integration test proves per-recipient rows + idempotent
  redelivery (no duplicates) + correct recipient set (readers == `v_cd_obligated_readers`; author ==
  `submitted_by`) + zero for non-recipients; `git diff` shows publish/approval *semantics* unchanged
  (only additive enqueues); `hgcrossmodule` + all 6 CI guards green; `go build/vet/test ./...` green.
