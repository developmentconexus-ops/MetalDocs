# Module: render-fanout

> **Last verified:** 2026-06-01 (P2 consolidation: added Failure modes section; prior: 2026-05-24)
> **Last verified:** 2026-06-01 (docgen-v2 → docx-renderer rename)
> **Status:** active (pipeline module)`r`n> **Maturity:** L2
> **Scope:** DOCX → PDF rendering pipeline, token substitution engine, outbox-driven dispatch.
> **Out of scope:** Approval-triggered freeze invocation (see `modules/approval.md`).
> **Key files:**
> - `internal/platform/messaging/events.go` - typed event envelope and `PDFConvertPayload`
> - `internal/modules/render/fanout/client.go` — HTTP client to docx-renderer
> - `apps/docx-renderer/src/routes/fanout.ts` — docx-renderer fanout route
> - `internal/modules/render/fanout/pdf_dispatcher.go` — outbox publisher
> - `internal/modules/render/fanout/pdf_dispatch_adapter.go` — invoker bridge
> - `internal/modules/render/fanout/pdf_outbox_repository.go:97` — `ReadState(ctx, tenantID, revisionID)` — returns latest outbox status; used by `view_service.go` to report `pdf_status=failed`
> - `internal/modules/render/fanout/pdf_outbox_worker.go` — background worker polls + dispatches
> - `internal/platform/worker/pdf_job_runner.go` — outbox consumer
> - `internal/modules/render/resolvers/builtins.go` — resolver implementations
> - `apps/docx-renderer/src/routes/fanout.ts` — docx-renderer fanout route
> - `apps/docx-renderer/src/render/fanout.ts` — eigenpal headless substitution

## Pipeline (high-level)

1. Freeze service substitutes the 7 fixed tokens in the DOCX (eigenpal-native format).
2. Frozen DOCX uploaded to MinIO.
3. PDFDispatcher publishes a `docgen_v2_pdf` outbox event with `messaging.PDFConvertPayload`.
4. PDFJobRunner picks up the typed payload, calls Gotenberg via docx-renderer.
5. Resulting PDF stored alongside the DOCX in MinIO.

## Failure modes

| Failure | Symptom | Detection | Response |
|---|---|---|---|
| Gotenberg / LibreOffice down | PDF outbox rows accumulate in `pdf_dispatch_outbox` with `pending`; published doc shows `pdf_status=pending` for >SLA | `pdf_outbox_worker` logs HTTP error; `pdf_outbox_repository.ReadState` returns `pending`/`failed` | Restart Gotenberg container; failed rows retried by worker until `max_attempts`, then marked `failed` (see [`render-fanout-tech-debt.md`](render-fanout-tech-debt.md)) |
| docx-renderer fanout substitution error | Freeze emits `docgen.substitution_failed`; outbox row may be `failed` | `apps/docx-renderer/src/routes/fanout.ts` returns 5xx; `internal/modules/render/fanout/client.go` surfaces error | Inspect docx-renderer logs for token/resolver mismatch; check `concepts/placeholders.md` 7-token catalog drift |
| Resolver returns empty value for required token | Frozen DOCX renders blank placeholder | `internal/modules/render/resolvers/builtins.go` returns `""`; tracked by `concepts/placeholders.md` | Confirm upstream data populated (e.g. controlled-document fields); add resolver coverage |
| Outbox replay (worker restart mid-dispatch) | Same `(tenant_id, revision_id)` processed twice | `ON CONFLICT (tenant_id, revision_id) DO NOTHING` on outbox INSERT dedupes; consumer is idempotent | Expected; no operator action |
| MinIO upload fails for frozen PDF | Worker logs S3 PUT error; outbox row marked `failed` | Worker error log + outbox status | MinIO healthcheck; retry by clearing `failed` status or reissuing freeze |

## See also

- [workflows/freeze-and-fanout.md](../workflows/freeze-and-fanout.md) — full pipeline (this is the canonical doc)
- [concepts/placeholders.md](../concepts/placeholders.md) — what the resolvers substitute
- [concepts/freeze-and-hashing.md](../concepts/freeze-and-hashing.md) — content/values/schema hashes

- [render-fanout-tech-debt.md](render-fanout-tech-debt.md)
- [backlog/render-fanout-refactor.md](../backlog/render-fanout-refactor.md)


## 11. Risks & Technical Debt

- Critical: 0
- Major: 2
- Minor: 1

Refactor backlog: [../backlog/render-fanout-refactor.md](../backlog/render-fanout-refactor.md)
