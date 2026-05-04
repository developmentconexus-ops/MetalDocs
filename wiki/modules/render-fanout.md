# Module: render-fanout

> **Last verified:** 2026-05-04
> **Status:** Stub. Cross-references `workflows/freeze-and-fanout.md` which has the full pipeline.
> **Scope:** DOCX → PDF rendering pipeline, token substitution engine, outbox-driven dispatch.
> **Out of scope:** Approval-triggered freeze invocation (see `modules/approval.md`).
> **Key files:**
> - `internal/modules/render/fanout/client.go` — HTTP client to docgen-v2
> - `internal/modules/render/fanout/pdf_dispatcher.go` — outbox publisher
> - `internal/modules/render/fanout/pdf_dispatch_adapter.go` — invoker bridge
> - `internal/modules/render/fanout/pdf_outbox_repository.go:97` — `ReadState(ctx, tenantID, revisionID)` — returns latest outbox status; used by `view_service.go` to report `pdf_status=failed`
> - `internal/modules/render/fanout/pdf_outbox_worker.go` — background worker polls + dispatches
> - `internal/platform/worker/pdf_job_runner.go` — outbox consumer
> - `internal/modules/render/resolvers/builtins.go` — resolver implementations
> - `apps/docgen-v2/src/routes/fanout.ts` — docgen-v2 fanout route
> - `apps/docgen-v2/src/render/fanout.ts` — eigenpal headless substitution

## Pipeline (high-level)

1. Freeze service substitutes the 7 fixed tokens in the DOCX (eigenpal-native format).
2. Frozen DOCX uploaded to MinIO.
3. PDFDispatcher publishes a `docgen_v2_pdf` outbox event.
4. PDFJobRunner picks up the event, calls Gotenberg via docgen-v2.
5. Resulting PDF stored alongside the DOCX in MinIO.

## See also

- [workflows/freeze-and-fanout.md](../workflows/freeze-and-fanout.md) — full pipeline (this is the canonical doc)
- [concepts/placeholders.md](../concepts/placeholders.md) — what the resolvers substitute
- [concepts/freeze-and-hashing.md](../concepts/freeze-and-hashing.md) — content/values/schema hashes
