# ADR 0009: PDF Dispatch Transactional Outbox

**Status:** Accepted (2026-05-03)  
**Affects:** freeze approval flow, document finalization, background workers

## Context

When a document approval is finalized, the freeze service substitutes tokens into the DOCX, uploads it to MinIO, and triggers PDF rendering via `pdfDispatcher.Dispatch()`.

**Problem:** If the process dies between the approval transaction commit and the dispatch call, the PDF rendering event is lost. The frozen DOCX exists but no PDF is generated. Manual re-dispatch requires operator intervention.

Approval is transactional (atomic within a database transaction). Dispatch is a side effect that happens after commit. If the server crashes or a network error occurs between commit and dispatch, the event disappears.

## Decision

Implement a **transactional outbox pattern**:

1. **Enqueue phase (in approval tx):** `pdfDispatcher.Dispatch()` inserts a row into `pdf_dispatch_outbox` table (in the same transaction as approval finalize). The write commits atomically with the approval state.

2. **Dequeue phase (background worker):** `PDFOutboxWorker` polls the outbox table with `SELECT FOR UPDATE SKIP LOCKED`, processes rows in batches, calls the render pipeline, and deletes on success.

3. **Stale recovery:** Rows older than 1 hour are considered stale claims. `ResetStaleClaims()` periodically updates stale rows to `claimed_at = NULL`, allowing other workers to retry.

## Consequences

**Positive:**
- Guaranteed at-least-once delivery — no events lost after approval commit
- Decoupled approval from rendering (approval completes before PDF generation)
- Multi-worker safe via `FOR UPDATE SKIP LOCKED` (no duplicate processing)
- Recovers from stale claims automatically

**Negative:**
- Requires background worker process running in bootstrap (added dependency)
- PDF generation now asynchronous (~5–10s latency, not immediate)
- Adds two new tables: `pdf_dispatch_outbox` + migration schema
- Operator must monitor worker health

## Implementation

- Table: `pdf_dispatch_outbox` (migration 0176)
  - `id`, `tenant_id`, `revision_id`, `claimed_by`, `claimed_at`, `error_count`, `last_error`, `created_at`
- `PDFOutboxRepository` — read/insert/delete/reset
- `PDFOutboxWorker` — polling loop, calls `pdf_dispatcher.Dispatch()` for each row
- Integrated into `internal/platform/worker/` bootstrap lifecycle

## See also

- `wiki/modules/render-fanout.md` — full rendering pipeline
- `wiki/workflows/freeze-and-fanout.md` — approval → freeze → fanout flow
- `migrations/0176_pdf_dispatch_outbox.sql` — schema
