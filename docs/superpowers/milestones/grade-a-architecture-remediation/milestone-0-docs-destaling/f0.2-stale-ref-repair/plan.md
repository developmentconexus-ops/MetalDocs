# Feature F0.2 — Stale `docs/` Reference Repair

> **Milestone:** 0 — Docs Progression De-Staling  ·  **Folder:** `f0.2-stale-ref-repair`
> **Status:** Done (evidence: `evidence.md`)

## Source

- Milestone spec row F0.2: *Implement* — "Fix the wiki refs that point at deleted `docs/`
  trees (`documentation-governance.md:73-96`, `decisions/index.md:34`, `quality/*`,
  `architecture/data-model.md`, `modules/frontend/iam.md`)." *Validate* — "`grep` proves
  **0** wiki links to deleted `docs/` paths; broken-link sweep clean."
- Governing spec: M0 (docs-only, decision D5 docs-first). No REQ id (no code grade moves).

## Discovery (grounding, 2026-06-14)

Post-deletion repo state: only `docs/superpowers/` survives under `docs/` (re-baseline
commit `c7f06f2e` removed legacy `docs/` **and** `.agents/skills/` trees). So any wiki ref
to `docs/<non-superpowers>` is stale.

A raw `docs/` substring sweep is noisy — most hits are **out of scope** and must be left:
- **External URLs** — `stripe.com/docs/api/...`, `riverqueue.com/docs/migrations`,
  `postgresql.org/docs/16|current/...`, `redis.io/docs/latest/...`,
  `go.opentelemetry.io/docs/languages/go/...`, OPA/AWS docs. Not repo paths.
- **Prose substrings** — `controlled-docs/taxonomy`→`docs/taxonomy`,
  `controlled-docs/documents-wrapper`→`docs/documents`, "P5 is docs/governance",
  "irrelevant for docs/code", "latent/docs/missing-ADR", "workflow docs/skills hardening".
  Not paths.

**True stale internal refs (the F0.2 work set):**

| # | Location | Current ref | Repair |
|---|----------|-------------|--------|
| 1 | `wiki/modules/frontend/iam.md:43` | md link → `../../../docs/audits/QA-evidence-admin-center-rebuild.md` | drop dead link; keep prose, note evidence removed at re-baseline |
| 2 | `wiki/modules/frontend/iam.md:119` | md link → same | same |
| 3 | `wiki/quality/index.md:25` | md link → `../../docs/runbooks/release-readiness.md` | replace with historical note; canonical = sibling `release-readiness.md` |
| 4 | `wiki/architecture/data-model.md:6` | backtick `docs/db-research/` | drop dead path; past-tense removal note |
| 5 | `wiki/decisions/0022-…:259` | prose `docs/runbooks/docx-v2-w2-templates.md` + sibling | keep historical filenames, drop `docs/runbooks/` path prefix; mark removed |
| 6 | `wiki/quality/qa-operating-system.md:342` | historical promotion step `docs/runbooks/release-readiness.md` | mark source removed (promotion already completed) |
| 7 | `wiki/quality/release-readiness.md:77` | "Source normalization note" → `docs/runbooks/release-readiness.md` | update: source removed, this page sole canonical |
| 8 | `wiki/standards/documentation-governance.md:88` | migration-map row `docs/runbooks/release-readiness.md` (`defer`) | row → removed-at-re-baseline; canonical wiki home |
| 9 | `wiki/standards/documentation-governance.md:89` | migration-map row `docs/adr/*` | row → removed-at-re-baseline; superseded by `wiki/decisions/` |
| 10 | `wiki/standards/documentation-governance.md:90` | migration-map row `docs/ck5-wiki/*` | row → removed-at-re-baseline |
| 11 | `wiki/standards/documentation-governance.md:96` | explicit-defer bullet "reconcile legacy `docs/adr/` tree" | defer obsolete (tree gone); update to reflect removal |

`decisions/index.md:34` (spec-named) was already repaired in F0.1 — verified present and
truthful ("now-deleted `docs/` documentation set"). No action.

## Repair principle

`wiki/_archive/` does not exist yet (F0.5). So F0.2 does **not** repoint to an archive —
it makes each ref **truthful to post-deletion reality**: the legacy `docs/` tree was
**removed at the v1 re-baseline (commit `c7f06f2e`)**, naming the surviving canonical wiki
home where one exists. F0.5 later adds `_archive/` rows to the governance migration map.

## Acceptance gates (run after edits)

- **Gate A (hard):** zero markdown link targets to deleted `docs/` paths —
  `grep -rnoP "\]\(\.{0,2}/?(?:\.\./)*docs/(?!superpowers)[^)]+\)" wiki` → **empty**.
- **Gate B (hard):** every surviving `docs/<non-superpowers>` token in wiki is an external
  URL or a clearly historical/past-tense removal note — **no** ref presents a deleted path
  as a current/live location. (Manual classification over the token sweep.)
- **Gate C:** the 11 edits above each land; no decision text altered (HS-6 guard); no code
  file touched.

## Out of scope (surfaced, not fixed — HS-6)

`.agents/skills/` was deleted in the same re-baseline; **13 wiki files** still link to it
(plus `CLAUDE.md`). This is a **separate deleted-tree class**, not a `docs/` ref → outside
F0.2's spec scope. Recorded in `evidence.md` as a discovered deviation for operator triage
(candidate new feature / fold into F0.5). F0.2 does not touch it.

## Execution notes

Surgical doc edits (≤6 files, past-tense truthful rewrites). Direct edits — no implementer
subagent (simplicity-first; nothing to compile/test). Durable record → `evidence.md`.
