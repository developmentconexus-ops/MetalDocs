# F3.3 — Domain-Event Pattern: Research Synthesis + Design Options

> **Milestone:** 3 — Notifications  ·  **Feature:** `f3.3-lifecycle-emitter` (re-scoped)
> **Date:** 2026-06-22  ·  **Status:** Design proposal — awaiting operator option pick.
> **Evidence base:** 113 verified claims salvaged from the (aborted) deep-research run
> `wf_832393c2-19e` (industry/OSS/SaaS practice: DDD domain events, transactional outbox,
> CloudEvents, Knock/novu notification-infra, modular-monolith event ownership). Citations are the
> claim corpus in that run's `journal.jsonl`; this note distills, it does not re-research.

## 1. The root cause (named precisely)

`governance_events` is an **actor-centric audit log** — it records *who acted* (`actor_user_id`),
immutably, in the state-change tx. That is a correct and valuable audit trail. It is **not** a
**domain-event substrate** — it does not carry *what business outcome occurred* or *who should be
notified*. Building notifications on it forces the consumer to reverse-engineer outcome + recipients
via compensating views (`v_approval_instance_submitter`, `v_document_cd_mapping`) and outcome-guessing
(`signoff_recorded` fires per-stage with no terminal marker; `signoff.rejected` is eligibility-failure,
not document rejection). That reverse-engineering is the workaround.

Industry distinguishes three event kinds and keeps them **distinct** (claims 44, 62–66, 109):
- **Audit events** — immutable record of an action (what `governance_events` already is). *Keep, untouched.*
- **Domain events** — "something interesting happened in the domain," carrying the **outcome** + the
  data downstream consumers need (claims 62–63). *This is what we lack.*
- **Integration / cross-module events** — the public contract between modules; the event **is** the
  module boundary (claims 36, 44–45). Consumers never read another module's tables.

## 2. What the evidence says to build (for a notifications consumer)

- **Event fatness: "medium," not thin, not full-ECST.** Default is thin (Thoughtworks, claims 87–90),
  but a thin ID-only event forces the consumer to call back / read producer tables — reintroducing the
  coupling we're trying to kill (claim 58). For notifications, the event must carry **the outcome +
  the recipient-resolution IDs** the owning module already has in hand at emit time (submitter_user_id,
  document_id, controlled_document_id, outcome=approved|rejected). Not full document state (claims 59–61, 90).
- **River IS our transactional outbox + dispatcher — no new infra** (claims 67–71). River enqueues a
  job in the **same Postgres tx** as the state change (claim 67 = the outbox guarantee); has **unique
  jobs / dedup, batch insert, periodic jobs** (claim 69); LISTEN/NOTIFY push, sub-ms latency, ~10k
  jobs/s (claims 68, 71). For a modular monolith, a Postgres-backed queue beats external broker / CDC /
  Debezium (claims 70, 109, 113 — CDC+outbox is for cross-service; we are one process).
- **At-least-once → idempotent consumer** (claims 48–49, 54, 84). The fix is an idempotent inbox keyed
  on the unique event id (claim 49). **We already built this** — F3.2's partial unique index
  `(recipient_user_id, source_event_id)`. The F3.2 table is already the correct idempotent inbox.
- **Fan-out-on-write** — one row per recipient, fast inbox reads (claims 77, 79, 96; Knock does exactly
  this, claim 104). **We already built this** — F3.2's per-recipient table. Company-scope high-fan-out
  is the only volume risk (milestone R4); bound it only if measured (claims 96–97, 100 — hybrid only
  when proven necessary).
- **Schema/versioning: keep our open `event_type` string + additive-only fields** (claims 6, 16, 20,
  31–32). Forward-compatible by default; breaking change ⇒ new type name (`x.v2`) (claims 21, 23).
  ADR-0043 already locked open-enum — correct, keep it.
- **Module ownership** (claims 34–37, 45–47): the domain-event definition lives in the **owning
  module**; consumers subscribe. Approval owns the signoff outcome event; documents owns
  publish/supersede/obsolete. Notifications **subscribes** — it does not read approval/CD base tables.

## 3. The fix in one sentence

At each **canonical business action** (publish / supersede / obsolete / final-signoff / reject), the
**owning module emits a domain event carrying outcome + recipient-resolution IDs**, enqueued via River
**in the same tx** as the state change; the **notifications module subscribes** and fans out
per-recipient rows **idempotently** (keyed on event id). `governance_events` stays as the audit log,
untouched. The two compensating views from ADR-0043 §6 are **eliminated** — the event carries
`submitted_by` and `controlled_document_id` directly from the module that already has them in scope, so
the consumer no longer reverse-engineers them. (Reader fan-out still reads the **published**
`v_cd_obligated_readers` contract — that is a legitimate published read, not base-table reverse-engineering.)

## 4. Options (pick one)

| | **A — Typed River-job domain events (recommended)** | **B — Dedicated `domain_events` outbox table + projector** | **C — Enrich audit events in place** |
|---|---|---|---|
| **Mechanism** | Each emit site, in the state-change tx, enqueues a **typed River job** (the job args ARE the domain-event payload: outcome + recipient IDs). A notifications **fan-out worker** consumes it, resolves recipients (`v_cd_obligated_readers` for readers; submitter from the event), inserts N idempotent rows. | A new `domain_events` table (envelope: id, type, occurred_at, aggregate, payload, outcome). Emit sites insert in-tx; a River **periodic projector** reads new rows past a watermark → notifications. | Widen the 5 `governance_events` payloads to carry outcome + submitter + cd_id; keep the projector reading `governance_events`. |
| **New infra** | None (River exists; F3.2 table is the inbox). | +1 table, +watermark/checkpoint. | None. |
| **Event log durability** | Ephemeral (River job consumed & gone — fine for notifications; claim 110). | Durable, queryable domain-event log (replay, future consumers). | None (audit log polluted with consumer fields). |
| **Module cleanliness** | Clean — typed event owned by emit module; notifications subscribes. | Clean + explicit shared envelope. | **Dirty** — conflates audit + domain; keeps reverse-engineering shape. |
| **Blast radius** | 5 emit sites get a River enqueue next to the existing audit emit; 1 new worker; typed event structs. `publish_service.go` **is** edited (operator lifted HS-2). | 5 emit sites + new table + projector + migration. | 5 payload edits; projector stays. |
| **Effort / risk** | **Medium / low** — additive enqueues, idempotent consumer, no new infra. | Medium-high / low — more moving parts, but most "proper event-sourcing-lite." | Low / **rejected** — the workaround with lipstick; operator already refused it. |
| **Extensible to parked emitter mission** | Yes — new event = new typed job + worker subscription. | Yes — new event = new row type + consumer. | Poorly — every consumer re-reads the audit log. |

**Recommendation: Option A.** Smallest footprint that fixes the root cause, uses infra we already run
(River) and the inbox table F3.2 already built, gives clean per-module event ownership, and stays
incremental (claims 28, 37, 47, 113 — strangler-fig, start coarse, keep source-of-truth in our own DB,
no big-bang). If you want a durable/queryable domain-event log for future replay or non-notification
consumers, that's **Option B** (worth it only if a second consumer is near-term).

## 5. If Option A — open recon items for the spec (not blockers)

1. Confirm `controlled_document_id` is in scope at the publish/supersede/obsolete emit sites (so the
   event carries it and `v_document_cd_mapping` is unnecessary). `submitted_by` is confirmed in scope
   (`instance.SubmittedBy`, decision_service.go).
2. Confirm a River worker can resolve `v_cd_obligated_readers` off-tx (H-PRE-1: no authz-recording read
   inside the publish lock — the fan-out worker runs after commit, so it's clear).
3. Decide the "final approval" signal: emit a distinct `document_approved` domain event only on the
   `InstanceApproved` transition (the canonical terminal point in decision_service.go), and a
   `document_rejected` domain event on the `InstanceRejected` transition — instead of overloading
   `signoff_recorded`. This is the clean per-outcome event the notifications consumer wants.
