# Feature F0.2 — Evidence

> **Milestone:** 0 — Docs Progression De-Staling  ·  **Feature:** `f0.2-stale-ref-repair`  ·  **Closed:** 2026-06-14
> A feature is closed only when every row below is filled with real output.

## What was implemented

Repaired every wiki reference that pointed at a legacy `docs/` tree removed at the v1
re-baseline (commit `c7f06f2e`). 11 refs across 7 files, made truthful to post-deletion
reality — past-tense "removed at re-baseline" notes naming the surviving canonical wiki
home where one exists. No archive repointing (F0.5 owns `wiki/_archive/`).

| File | Ref repaired | Repair |
|------|--------------|--------|
| `wiki/modules/frontend/iam.md:43` | md link → `docs/audits/QA-evidence-admin-center-rebuild.md` | link dropped; "removed at re-baseline" note |
| `wiki/modules/frontend/iam.md:119` | md link → same | link dropped; removal note |
| `wiki/quality/index.md:25` | md link → `docs/runbooks/release-readiness.md` | replaced with canonical sibling link + removal note |
| `wiki/architecture/data-model.md:6` | backtick `docs/db-research/` | dead path dropped; removal note |
| `wiki/decisions/0022-…:259` | prose `docs/runbooks/docx-v2-w2/w3-*.md` | path prefix dropped; historical filenames kept + "since-removed" |
| `wiki/quality/qa-operating-system.md:342` | promotion step `docs/runbooks/release-readiness.md` | marked done — source removed |
| `wiki/quality/release-readiness.md:77` | "Source normalization note" path | updated: sole canonical, no staging source remains |
| `wiki/standards/documentation-governance.md:88-90` | migration-map rows `docs/runbooks/…`, `docs/adr/*`, `docs/ck5-wiki/*` | Decisions rewritten to "Removed at re-baseline" + canonical wiki home |
| `wiki/standards/documentation-governance.md:96` | explicit-defer "reconcile legacy `docs/adr/`" | struck through — obsolete (tree gone, reconciled in F0.1) |

`decisions/index.md:34` (spec-named) was already repaired in F0.1 — verified truthful, no
action. Migration-map rows for `docs/superpowers/specs/*` left intact (those paths survive).

Not yet committed — staged for the M0 close-out commit batch (operator gate HS-1).

## Verification

| Check | Command / action | Result (evidence) |
|-------|------------------|-------------------|
| Gate A — zero broken md links to deleted `docs/` | `grep -rnoP "\]\(\.{0,2}/?(?:\.\./)*docs/(?!superpowers)[^)]+\)" wiki --include="*.md"` | **empty** (exit 1, no match) |
| Gate B — every surviving `docs/<non-superpowers>` token is external URL / prose / historical note | `grep -rnoP "(?<![A-Za-z])docs/(?!superpowers)[A-Za-z0-9/_.-]+" wiki` → 32 tokens, each classified | **0 present a deleted path as live**: 20 external URLs (stripe/river/postgres/redis/OTel/K8s/GCP/OpenFGA/OPA), 6 prose substrings (`controlled-docs/taxonomy` etc.), 6 historical removal notes (my edits + governance map rows) |
| Gate C — docs-only, no decision text altered | `git status --short` | 7 `wiki/**` modified + feature folder untracked; **0 code files**; only stale path refs changed, no ADR decision section touched |

> Docs-only feature — no build/test/runtime surface. The doc-QA greps above are the proof.

## Acceptance vs milestone spec

From `../milestone.md` F0.2: *"`grep` proves **0** wiki links to deleted `docs/` paths;
broken-link sweep clean."*

| Acceptance criterion (from milestone.md) | Met? | Evidence |
|------------------------------------------|------|----------|
| 0 wiki links to deleted `docs/` paths | yes | Gate A empty |
| Broken-link sweep clean (`docs/` class) | yes | Gate B — every residual token is external/prose/historical, none a live broken link |
| Spec-named locations repaired | yes | governance map 73-96, index:34 (F0.1), quality/*, data-model.md, iam.md all covered above |

## Review disposition

- **Spec-compliance review:** ✅ compliant. Self-reviewed all 11 edits: each repairs a stale
  `docs/` ref and nothing else; no ADR *decision* rewritten (HS-6 guard — 0022 edit touched
  only the Origin-investigation evidence-path list, not the Phase 9 decision); no markdown
  link left pointing at a deleted path. Re-ran Gates A/B/C post-edit — all pass.
- **Code-quality review:** N/A — docs-only markdown, no code surface. Wording reviewed inline
  by controller: past-tense + commit `c7f06f2e` cited consistently across all notes; canonical
  targets are live links (`release-readiness.md`, `wiki/decisions/`).

## Bounded defers

| Defer | Why bounded | Written trigger / owner |
|-------|-------------|-------------------------|
| **HS-6 — `.agents/skills/` is a second deleted-tree class (RESOLVED — operator ruling 2026-06-14).** The same re-baseline (`c7f06f2e`) removed `.agents/skills/`; **13 wiki files** still link to it (e.g. `frontend-structure.md:25`, `system-map.md:28`, `iam.md:118`) plus `CLAUDE.md` routes to `.agents/skills/*/SKILL.md`. | **Out of F0.2 spec scope** — F0.2 is explicitly "deleted `docs/` trees". `.agents/skills` is a different class the M0 plan did not enumerate. Surfaced not silently fixed (surgical / HS-6). | **Resolution:** operator ruled `.agents/skills/` was **intentionally deleted** (stale `.md` files / old logic) — **not to be repaired or tracked**. These refs are deliberately dead; the M0 broken-link sweep is scoped to the `docs/` class only. No follow-up feature. |
