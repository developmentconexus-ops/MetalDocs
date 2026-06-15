# Feature F4.3 — Plan

> Documentation-only. One ordered pass; no TDD (no code).

## Slices

1. **Author ADR 0029** — `wiki/decisions/0029-user-display-name-reader-port.md` from the F4.1 boundary
   (iam-owned port, single+batch, reads-live/off-tx, Approach 2 rejected, security JOIN deferred).
2. **Author ADR 0030** — `wiki/decisions/0030-template-version-state-port.md` from the F4.2 boundary
   (extend existing `TemplateVersionPort`, raw `GetTemplateVersionState`, `IsPublished` kept, reads-live/
   off-tx, parallel-reader + Approach 2 rejected).
3. **Register** both in `wiki/decisions/index.md` (one row each, matching the table format).
4. **Cross-link** from F4.1 `spec.md` (replace ADR placeholder) and F4.2 `spec.md` (replace
   `<f4.3 adr link>` placeholder).
5. **Module-doc references** — dispatch `wiki-curator` (CLAUDE.md canonical path) to add a one-line ADR
   cross-link + stamp bump to iam.md, documents.md, approval.md, templates.md, controlled-documents.md.
6. **Gate verify** — run the F4.3 Validation Gate greps; write `evidence.md`; commit.

## Files touched

- `wiki/decisions/0029-user-display-name-reader-port.md` (new)
- `wiki/decisions/0030-template-version-state-port.md` (new)
- `wiki/decisions/index.md` (two rows)
- `f4.1-user-display-name-reader/spec.md`, `f4.2-template-version-state-reader/spec.md` (ADR links)
- `wiki/modules/{iam,documents,approval,templates,controlled-documents}.md` (cross-link + stamp, via curator)
- this feature's `spec.md` / `plan.md` / `evidence.md`
