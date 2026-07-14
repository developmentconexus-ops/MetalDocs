# Sync log - render-fanout

> Append-only log of `metaldocs-module-doc-sync` runs against this module. Newest at top.

## 2026-05-24 - Typed messaging event envelope C10

- **Context:** uncommitted diff for branch `fix/messaging-2b-c10`
- **Mode:** structural refresh
- **Anchors moved:** none
- **Public surface:** `messaging.Event`, `messaging.Consumer.MarkPublished`, render PDF publishers, and worker PDF runner now use typed event IDs/event type/payload boundaries.
- **Routes/API:** none
- **Runtime flows:** render-fanout pipeline now records that `docgen_v2_pdf` dispatch carries `messaging.PDFConvertPayload`.
- **Persistence:** outbox JSON payload storage unchanged; consumer decodes JSON into typed payloads at claim time.
- **Dependencies:** render-fanout dependency on platform messaging now includes `PDFConvertPayload`; worker app wraps the document repository with `NewSnapshotPDFPersister`.
- **T-NNN touched:** none
- **R-NNN touched:** none
- **Counts after:** Critical=0 Major=2 Minor=1; missing-ADR=3
- **Tally gate:** PASS
- **Patched files:** `wiki/modules/render-fanout.md`, `wiki/modules/render-fanout-tech-debt.md`, `wiki/backlog/render-fanout-refactor.md`, `wiki/modules/render-fanout/_artifacts/00-context.md`, `wiki/modules/render-fanout/_artifacts/sync-log.md`
