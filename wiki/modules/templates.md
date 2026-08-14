# Module: templates — LEGACY PARALLEL LIFECYCLE

> **Status:** LEGACY / target lifecycle being retired
> **Marked:** 2026-08-14

The current `internal/modules/templates` implementation still runs, but the target no longer accepts Template as a peer aggregate with its own document-like lifecycle/version counter.

Locked direction:

- template is a **designation/role of an exact governed DocumentRevision**;
- changing template DOCX/body, placeholders, types, constraints, visibility or resolver binding requires a new DocumentRevision;
- derived documents preserve the exact source revision/hash used to seed them;
- later template revisions never silently rebind existing documents;
- after creation, the derived document's own Revision is the content truth being edited/reviewed/approved.

Read:

- `wiki/architecture/cohesive-platform-redesign.md`
- `docs/superpowers/analysis/2026-08-14-cohesive-platform-redesign-ledger.md`

The old template schema/placeholder implementation is still important migration evidence, but it no longer defines a separate target lifecycle. Detailed historical architecture remains in Git history.
