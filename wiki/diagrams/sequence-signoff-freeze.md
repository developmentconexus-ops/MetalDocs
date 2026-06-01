# Sequence — Signoff + Freeze

> **Last verified:** 2026-06-01 (async freeze refactor — ADR 0015)
> **Flow:** Approver clicks "approve" → API pins the document (fast, no network) → commits → worker materializes the frozen DOCX asynchronously.
> **Code anchors:**
> - [`internal/modules/documents/approval/application/decision_service.go`](../../internal/modules/documents/approval/application/decision_service.go) — signoff calls `pinInvoker.Pin` in-tx
> - [`internal/modules/documents/application/freeze_service.go`](../../internal/modules/documents/application/freeze_service.go) — `Pin` + `Materialize`
> - [`internal/modules/render/fanout/materialize_outbox_worker.go`](../../internal/modules/render/fanout/materialize_outbox_worker.go) — outbox relay (API process)
> - [`internal/platform/worker/materialize_job_runner.go`](../../internal/platform/worker/materialize_job_runner.go) — job handler (worker process)

## Current design (as of 2026-06-01 — async freeze, ADR 0015)

```mermaid
sequenceDiagram
    autonumber
    actor Approver
    participant Web as metaldocs-web
    participant API as metaldocs-api
    participant PG as Postgres
    participant Worker as metaldocs-worker
    participant Docgen as docx-renderer
    participant Gotenberg as gotenberg
    participant Minio as MinIO

    Approver->>Web: Click "Approve" in inbox
    Web->>API: POST /api/v1/documents/{id}/signoff (If-Match: "v{n}")

    activate API
    Note over API,PG: Begin signoff transaction (open until COMMIT below)
    API->>PG: BEGIN
    API->>PG: load route + instance + stage + caps + eligibility checks
    API->>PG: validate password / SoD / quorum
    API->>PG: INSERT approval_signoffs row
    API->>PG: UPDATE approval_stage_instances (stage state)

    Note right of API: Final stage satisfies quorum → Pin.

    API->>PG: read snapshot + freeze marker (idempotency)
    API->>PG: validate required placeholders + resolve computed ones
    API->>PG: compute values_hash
    API->>PG: UPDATE revisions SET values_hash, frozen_at (Pin.WriteFreeze)
    API->>PG: INSERT materialize_dispatch_outbox row
    API->>PG: UPDATE documents SET status='approved'
    API->>PG: INSERT governance_events (signoff_recorded)
    API->>PG: COMMIT
    deactivate API
    API-->>Web: 200 {result, next_state}
    Web-->>Approver: "Approved" (+ "Finalizando artefato…" badge)

    Note over API: Async — MaterializeOutboxWorker polls the outbox
    API->>PG: claim materialize_dispatch_outbox row
    API->>PG: publish docx_materialize event

    Note over Worker: MaterializeJobRunner handles docx_materialize event
    Worker->>PG: read frozen placeholder values + snapshot
    Worker->>Docgen: POST /render/fanout {body_docx_key, placeholder_values, composition}
    Docgen->>Minio: GET template body docx
    Docgen->>Docgen: eigenpal headless reconstruct
    Docgen->>Minio: PUT final docx → tenants/{t}/revisions/{r}/frozen.docx
    Docgen-->>Worker: {final_docx_s3_key, content_hash}
    Worker->>PG: BEGIN
    Worker->>PG: UPDATE revisions SET final_docx_key, content_hash (WriteFinalDocx)
    Worker->>PG: INSERT pdf_dispatch_outbox row
    Worker->>PG: COMMIT

    Note over Worker: PDF chain unchanged from sequence-pdf-export.md
    Worker->>PG: poll pdf_dispatch_outbox → claim row → publish docx_to_pdf event
    Worker->>Minio: GET frozen docx
    Worker->>Gotenberg: POST /forms/libreoffice/convert
    Gotenberg-->>Worker: PDF bytes
    Worker->>Minio: PUT PDF
    Worker->>PG: write artifact metadata (final pdf key + hash)
```

## What freeze guarantees (compliance)

| Guarantee | Mechanism |
|---|---|
| **Content immutability at approval** | `values_hash` + `frozen_at` written atomically with the signoff in one tx — **unchanged** |
| **Auditable** | `governance_events` row written same tx; `audit_log` records actor + action |
| **Idempotent** | Second freeze attempt early-returns when `frozen_at` is already set |
| **Deterministic reconstruct** | `final docx = f(body_docx, values, composition)` — re-runnable, content-addressed by hash |

## Failure modes

| Failure | Outcome |
|---|---|
| docx-renderer down | Signoff commits; outbox retries with backoff |
| docx-renderer slow | Worker retries with backoff |
| Reconstruct fails for content reason | Approval committed; artifact retried; `materialize_status=failed` after max_attempts |
| Worker crashes mid-materialize | Outbox claim TTL expires; another worker reclaims |

## Previous design (deprecated 2026-06-01)

The previous design called `FreezeService.Freeze` synchronously inside the signoff
transaction, making an HTTP call to docx-renderer before commit. See ADR 0015 for why
this was problematic and the full before/after analysis.

## Related

- [c4-container-backend.md](c4-container-backend.md)
- [sequence-pdf-export.md](sequence-pdf-export.md)
- [`wiki/decisions/0015-async-freeze-pin-materialize.md`](../decisions/0015-async-freeze-pin-materialize.md)
- [`wiki/decisions/0009-pdf-dispatch-outbox.md`](../decisions/0009-pdf-dispatch-outbox.md)
- [`wiki/modules/approval.md`](../modules/approval.md), [`wiki/concepts/freeze-and-hashing.md`](../concepts/freeze-and-hashing.md)
