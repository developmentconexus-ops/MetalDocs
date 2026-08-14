# Module: documents — LEGACY CURRENT-STATE MODULE

> **Status:** LEGACY current-state implementation / responsibilities being re-homed
> **Marked:** 2026-08-14

`internal/modules/documents` still contains the current editor/draft/revision/comments/freeze/view/export implementation, but its historical module boundary is not the target architecture.

Target Controlled Information centers on:

```text
Document
DocumentRevision
immutable submission evidence
Template-as-revision-role
Release/effectivity
```

The crucial invariant is that the exact Revision/hash reviewed and submitted is the only legal source for Approval evidence, freeze and official Rendition. The historical QA defect where the editor revision was reviewed while freeze rendered a separate blank template snapshot must become structurally impossible.

Read:

- `wiki/architecture/cohesive-platform-redesign.md`
- `docs/superpowers/analysis/2026-08-14-cohesive-platform-redesign-ledger.md`

Do not extend current bridge fields, foreign domain types or duplicate lifecycle paths merely to preserve this module shape. Git history contains the former detailed living architecture for migration archaeology.
