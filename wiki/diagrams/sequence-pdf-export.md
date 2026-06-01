# Sequence — PDF Generation (post-freeze)

> **Last verified:** 2026-06-01 (docgen-v2 → docx-renderer rename)
> **Flow:** Final approval enqueues a row in the **transactional PDF outbox** inside the same DB transaction. After the tx commits, a worker claims the row, reads the frozen docx from MinIO, converts it via Gotenberg, and writes the PDF back to MinIO. **Failure here never rolls back the approval** — ADR 0009.
> **Code anchors:**
> - [`internal/modules/render/fanout/pdf_outbox_repository.go:33`](../../internal/modules/render/fanout/pdf_outbox_repository.go) — `pdf_dispatch_outbox` INSERT (in-tx)
> - [`internal/modules/render/fanout/pdf_outbox_worker.go`](../../internal/modules/render/fanout/pdf_outbox_worker.go) — relay: outbox row → published event
> - [`internal/platform/worker/pdf_job_runner.go:82`](../../internal/platform/worker/pdf_job_runner.go) — `ConvertPDF` call
> - [`internal/platform/servicebus/gotenberg_pdf.go`](../../internal/platform/servicebus/gotenberg_pdf.go) — direct Gotenberg adapter (replaces the old docx-renderer PDF proxy; on the qa/approval-inbox branch)
> - [ADR 0009](../decisions/0009-pdf-dispatch-outbox.md) — transactional outbox pattern

```mermaid
sequenceDiagram
    autonumber
    actor Approver
    participant API as metaldocs-api
    participant PG as Postgres
    participant Outbox as PDFOutboxWorker (in metaldocs-worker)
    participant Pub as Outbox publisher
    participant Runner as PDFJobRunner (in metaldocs-worker)
    participant Minio as MinIO
    participant Gotenberg as gotenberg

    Note over Approver,API: (Approval signoff — see sequence-signoff-freeze.md)
    API->>PG: BEGIN ... INSERT pdf_dispatch_outbox(tenant,revision,content_hash) ... COMMIT
    Note right of API: Outbox row is INSERTED inside the signoff tx →<br/>atomic with approval. No event published yet.

    loop every ~5s
        Outbox->>PG: ClaimPending (UPDATE … RETURNING; claim lease)
        PG-->>Outbox: rows (or empty)
        alt rows found
            Outbox->>Pub: Publish(EventTypePDFConvert{tenant, revision, content_hash})
            Outbox->>PG: MarkDispatched(id)
        end
    end

    Note over Runner,PG: Worker service consumes events from the outbox_events table
    Runner->>PG: claim event (EventTypePDFConvert)
    Runner->>Minio: GET frozen docx (tenants/{t}/revisions/{r}/frozen.docx)
    Minio-->>Runner: docx bytes
    Runner->>Gotenberg: POST /forms/libreoffice/convert
    Gotenberg-->>Runner: PDF bytes
    Runner->>Runner: sha256(pdf) → content hash
    Runner->>Minio: PUT PDF (tenants/{t}/revisions/{r}/final.pdf)
    Runner->>PG: WritePDF metadata (storage key, hash, generated_at)
    Runner->>PG: mark event handled

    Note right of Runner: On any failure (Gotenberg down, MinIO error,<br/>conversion failure) the worker re-claims with backoff.<br/>Approval is already committed — no rollback.
```

## Why this pattern (ADR 0009)

| Requirement | How it's met |
|---|---|
| **Approval must not fail because PDF is slow/down** | PDF row enqueued in-tx; the actual conversion is async and retried. |
| **At-least-once delivery** | Outbox row is durable in Postgres. Worker reclaims if it crashes mid-flight. |
| **Idempotency** | Event has `IdempotencyKey = "docgen_v2_pdf:{tenant}:{revision}"`; consumers dedupe on it. |
| **Backpressure** | Worker batch size + poll interval bound throughput; no thundering herd. |
| **Observable** | `pdf_dispatch_outbox.status` + `attempts` + `last_error` are inspectable in SQL. |

## Gotenberg path (no eigenpal)

PDF generation does **not** use eigenpal. It's a straight docx → PDF conversion via Gotenberg's LibreOffice route. The (frozen) docx is the input; bytes never re-enter the eigenpal engine for PDF. This is the cleanest provider boundary in the system.

> Note: a Go-side adapter (`GotenbergPDFClient`, in `internal/platform/servicebus/gotenberg_pdf.go`) replaces the previous docx-renderer `/convert/pdf` HTTP hop. Worker now talks to Gotenberg directly. The Node docx-renderer service is no longer in the PDF path. On the `qa/approval-inbox` branch as of 2026-06-01.

## Failure modes

| Failure | Worker behavior | Operator visibility |
|---|---|---|
| Gotenberg down | Retry with exponential backoff (capped at 30 min) | `pdf_dispatch_outbox.attempts` increments; `last_error` set |
| Frozen docx missing in MinIO | Same | Same — likely indicates a freeze materialize lag/failure |
| Worker crash mid-conversion | Stale-claim TTL expires; another worker picks up | Lease holder rotates |
| `max_attempts` exhausted | Row marked `dead_letter`; no further auto-retry | SQL query on dead-lettered rows |

## Related

- [c4-container-backend.md](c4-container-backend.md) — see `worker → gotenberg` arrow.
- [sequence-signoff-freeze.md](sequence-signoff-freeze.md) — where the outbox row originates.
- [ADR 0009 — PDF Dispatch Outbox](../decisions/0009-pdf-dispatch-outbox.md).
