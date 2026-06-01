# Diagrams

> **Last verified:** 2026-06-01
> **Scope:** Canonical system-level diagrams. One source of truth, embedded elsewhere via Mermaid blocks.
> **Style:** Mermaid only. No PNG/JPEG. GitHub renders Mermaid natively.

## What lives here

Diagrams that describe the system *as a whole* (cross-module) and the load-bearing end-to-end flows. Module-internal diagrams stay inside their owning `wiki/modules/<name>.md`.

| File | Type | Shows | Embedded in |
|---|---|---|---|
| [c4-context.md](c4-context.md) | C4 L1 (Context) | External actors + the MetalDocs system as one black box | [system-overview.md](../architecture/system-overview.md) |
| [c4-container-backend.md](c4-container-backend.md) | C4 L2 (Container) | Backend services, ports, dependencies | [system-overview.md](../architecture/system-overview.md) |
| [sequence-create-document.md](sequence-create-document.md) | Sequence | Browser → API → Postgres + MinIO | [workflows/document-fillin.md](../workflows/document-fillin.md) |
| [sequence-edit-autosave.md](sequence-edit-autosave.md) | Sequence | Browser ↔ MinIO direct via presigned URLs | [workflows/document-fillin.md](../workflows/document-fillin.md) |
| [sequence-signoff-freeze.md](sequence-signoff-freeze.md) | Sequence | Approval signoff + freeze (current sync design + planned async) | [workflows/approval.md](../workflows/approval.md), [workflows/freeze-and-fanout.md](../workflows/freeze-and-fanout.md) |
| [sequence-pdf-export.md](sequence-pdf-export.md) | Sequence | Transactional outbox → worker → Gotenberg | [workflows/freeze-and-fanout.md](../workflows/freeze-and-fanout.md) |

## How to use

- **Add a diagram here** when it describes more than one module or a cross-cutting flow.
- **Embed by transclusion** — link to the file, do not copy the Mermaid block elsewhere. Single source of truth, no drift.
- **Keep diagrams minimal** — show the load-bearing actors and arrows, not every helper. Detail belongs in the module page.
- **Stamp `Last verified:`** at the top of every diagram file. Update when the underlying code changes.

## Reading order for new engineers

1. [c4-context.md](c4-context.md) — what is MetalDocs from outside.
2. [c4-container-backend.md](c4-container-backend.md) — the moving parts inside.
3. [sequence-create-document.md](sequence-create-document.md) — simplest end-to-end flow.
4. [sequence-edit-autosave.md](sequence-edit-autosave.md) — the scalability pattern (browser ↔ object store direct).
5. [sequence-signoff-freeze.md](sequence-signoff-freeze.md) — the compliance moment.
6. [sequence-pdf-export.md](sequence-pdf-export.md) — async derivation pattern.

After this, dive into any [`wiki/modules/<name>.md`](../modules/) for depth.
