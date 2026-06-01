# ADR 0015: Async Freeze — Split into Pin (in-tx) and Materialize (outbox)

**Status:** Accepted (2026-06-01)
**Affects:** approval signoff flow, freeze service, worker, docx-renderer coupling

## Context

Approval signoff (`DecisionService.RecordSignoff`) called `FreezeService.Freeze` inside
the open DB transaction. `Freeze` made a synchronous HTTP call to the docx-renderer
(`POST /render/fanout`) before committing. This created a hard coupling:

1. **docx-renderer down → approval fails and rolls back.** Availability ≤ docx-renderer uptime.
2. **Postgres row locks held across an HTTP call.** Under load this serialises signoffs
   and pressures the connection pool.
3. **Recovery on transient failure = user retry.** The freeze marker was also rolled back
   (same tx), so the next attempt re-ran the full resolve → WriteFreeze cycle.

The PDF path (ADR 0009) already proved the correct pattern: transactional outbox.
The reconstruct HTTP call is a *derived artifact producer*, not a legal commitment.

## ISO 9001 Compliance Invariant

| Guarantee | Mechanism | Affected by this change? |
|---|---|---|
| Content immutability at approval | `values_hash` + `frozen_at` written atomically with the signoff | **No** — still in the same tx |
| Auditable | `governance_events` row same tx | **No** |
| Idempotent | Early-return when `frozen_at` already set | **No** |
| Deterministic reconstruct | `final docx = f(body_docx, values, composition)` | **No** — just happens later |

The artifact bytes (frozen.docx) are derived from already-committed, immutable inputs.
Producing them asynchronously does not weaken the legal commitment.

## Decision

Split `Freeze` into two halves:

- **`Pin` (in-tx, fast, no network):**
  Validates required placeholders, resolves computed ones, writes `values_hash` +
  `frozen_at`, inserts a `materialize_dispatch_outbox` row — all inside the signoff tx.
  Commits without any network call to docx-renderer.

- **`Materialize` (async, in worker):**
  Reads the already-committed placeholder values, calls `POST /render/fanout`,
  writes `final_docx_key` + `content_hash`, and enqueues a `pdf_dispatch_outbox` row
  (which chains the existing PDF conversion path, ADR 0009).

The outbox relay (`MaterializeOutboxWorker`) runs in the API process alongside
`PDFOutboxWorker`. The job handler (`MaterializeJobRunner`) runs in the worker process
alongside `PDFJobRunner`.

## Consequences

**Positive:**
- Approval availability is independent of docx-renderer uptime.
- No Postgres row locks held across HTTP calls.
- Transient docx-renderer failures are retried automatically (same policy as PDF — max 5 attempts, exponential backoff, dead-letter on exhaustion).
- The ISO 9001 freeze guarantees are fully preserved.
- Pattern is identical to ADR 0009 — no new architectural concepts.

**Negative:**
- A brief window (~seconds, normally) exists between `approved` status and the artifact
  being available. UI must represent this (see `materialize_status` in the outbox).
- Any consumer that assumed `final_docx_key` is set the instant status flips to
  `approved` must now handle the `pending` case.
- The worker must be reachable to docx-renderer (added `depends_on: docx-renderer`).

## Implementation

- Table: `materialize_dispatch_outbox` (migration 0210) — mirrors `pdf_dispatch_outbox`.
- `MaterializeOutboxRepository` — read/insert/claim/mark.
- `MaterializeOutboxWorker` — polling loop, publishes `docx_materialize` events.
- `MaterializeJobRunner` — handles `docx_materialize`, calls fanout, writes final docx
  + pdf outbox atomically.
- `FreezeService.Pin` / `FreezeService.Materialize` — split of the old `Freeze`.
- `DecisionService.WithPinInvoker` — enables the async path; falls back to old `Freeze`
  if not set (backward compatibility for tests and gradual rollout).

## See also

- ADR 0009 — PDF dispatch transactional outbox (the pattern this mirrors).
- `wiki/diagrams/sequence-signoff-freeze.md` — updated sequence diagrams.
- `wiki/diagrams/c4-container-backend.md` — removed sync API→docx-renderer arrow.
- `migrations/0210_materialize_dispatch_outbox.sql`.
