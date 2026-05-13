# Module: render-fanout

> **Last verified:** 2026-05-13
> **Status:** active (pipeline module)`r`n> **Maturity:** L2
> **Scope:** DOCX ÃƒÂ¢Ã¢â‚¬Â Ã¢â‚¬â„¢ PDF rendering pipeline, token substitution engine, outbox-driven dispatch.
> **Out of scope:** Approval-triggered freeze invocation (see `modules/approval.md`).
> **Key files:**
> - `internal/modules/render/fanout/client.go` ÃƒÂ¢Ã¢â€šÂ¬Ã¢â‚¬Â HTTP client to docgen-v2
> - `internal/modules/render/fanout/pdf_dispatcher.go` ÃƒÂ¢Ã¢â€šÂ¬Ã¢â‚¬Â outbox publisher
> - `internal/modules/render/fanout/pdf_dispatch_adapter.go` ÃƒÂ¢Ã¢â€šÂ¬Ã¢â‚¬Â invoker bridge
> - `internal/modules/render/fanout/pdf_outbox_repository.go:97` ÃƒÂ¢Ã¢â€šÂ¬Ã¢â‚¬Â `ReadState(ctx, tenantID, revisionID)` ÃƒÂ¢Ã¢â€šÂ¬Ã¢â‚¬Â returns latest outbox status; used by `view_service.go` to report `pdf_status=failed`
> - `internal/modules/render/fanout/pdf_outbox_worker.go` ÃƒÂ¢Ã¢â€šÂ¬Ã¢â‚¬Â background worker polls + dispatches
> - `internal/platform/worker/pdf_job_runner.go` ÃƒÂ¢Ã¢â€šÂ¬Ã¢â‚¬Â outbox consumer
> - `internal/modules/render/resolvers/builtins.go` ÃƒÂ¢Ã¢â€šÂ¬Ã¢â‚¬Â resolver implementations
> - `apps/docgen-v2/src/routes/fanout.ts` ÃƒÂ¢Ã¢â€šÂ¬Ã¢â‚¬Â docgen-v2 fanout route
> - `apps/docgen-v2/src/render/fanout.ts` ÃƒÂ¢Ã¢â€šÂ¬Ã¢â‚¬Â eigenpal headless substitution

## Pipeline (high-level)

1. Freeze service substitutes the 7 fixed tokens in the DOCX (eigenpal-native format).
2. Frozen DOCX uploaded to MinIO.
3. PDFDispatcher publishes a `docgen_v2_pdf` outbox event.
4. PDFJobRunner picks up the event, calls Gotenberg via docgen-v2.
5. Resulting PDF stored alongside the DOCX in MinIO.

## See also

- [workflows/freeze-and-fanout.md](../workflows/freeze-and-fanout.md) ÃƒÂ¢Ã¢â€šÂ¬Ã¢â‚¬Â full pipeline (this is the canonical doc)
- [concepts/placeholders.md](../concepts/placeholders.md) ÃƒÂ¢Ã¢â€šÂ¬Ã¢â‚¬Â what the resolvers substitute
- [concepts/freeze-and-hashing.md](../concepts/freeze-and-hashing.md) ÃƒÂ¢Ã¢â€šÂ¬Ã¢â‚¬Â content/values/schema hashes

- [render-fanout-tech-debt.md](render-fanout-tech-debt.md)
- [backlog/render-fanout-refactor.md](../backlog/render-fanout-refactor.md)


## 11. Risks & Technical Debt

- Critical: 0
- Major: 2
- Minor: 1

Refactor backlog: [../backlog/render-fanout-refactor.md](../backlog/render-fanout-refactor.md)
