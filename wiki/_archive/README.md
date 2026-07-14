# wiki/_archive

> **Purpose:** Frozen home for **superseded / closed** wiki docs that are kept for the record but
> are no longer live planning surfaces. Archiving here means "historical, not deleted" — the doc
> stays readable and git-tracked; it just stops competing with active docs for agent context.

## What belongs here

- **Closed programs** — a program doc whose work fully shipped and carries no active deferred work
  (e.g. a backlog program marked CLOSED with a passing closing re-audit).
- **Superseded docs** — a doc whose content was folded into / replaced by a canonical successor.

## What does **not** belong here

- Anything with active `open` rows or live deferred-stub items (intentional stubs, missing-backend
  capability). Those stay in their live home.
- Docs that were merely de-staled in place with a HISTORICAL banner and are still cross-referenced
  as living history (e.g. the backend execution roadmap) — those are recorded in the governance
  migration map as *retained in place*, not relocated here.

## Conventions

- **Mirror the original path under `_archive/`.** A doc moved from `wiki/backlog/foo.md` lands at
  `wiki/_archive/backlog/foo.md` (provenance is legible in the path). Note this **adds one directory
  level**, so the moved doc's own parent-relative (`../`) links must each gain one `../` to keep
  resolving — fix them in the same change (see next bullet).
- **The index of record is the governance migration map** —
  [`wiki/standards/documentation-governance.md`](../standards/documentation-governance.md). Every
  relocation gets a row there naming the new `_archive/` home. That map, not this README, is the
  single source of truth for where moved docs went.
- **Fix inbound links on move.** When a doc is relocated, every real markdown link into it is
  repointed in the same change — no dangling entries.

## Provenance

Established in **M0 F0.5** (Grade-A Architecture Remediation, Milestone 0 — Docs De-Staling).
