# Feature F0.3 — Evidence

> **Milestone:** 0 — Docs Progression De-Staling  ·  **Feature:** `f0.3-roadmap-consolidation`  ·  **Closed:** 2026-06-14
> A feature is closed only when every row below is filled with real output.

## What was implemented

Consolidated two competing roadmap surfaces into **one** forward roadmap. Created
`wiki/roadmap.md` as the single canonical forward progression surface; bannered both old
roadmaps HISTORICAL/superseded (bodies retained as the record); repointed `backlog/index.md`.

| File | Change |
|------|--------|
| `wiki/roadmap.md` (new) | Single forward roadmap: active **Grade-A Architecture Remediation** program (links README + governing spec; notes H-G trigger = M4); **post-v1 carried-forward** (Plan 12 screens, eigenpal packaging, Wave-3 triggers) **by reference**; superseded-roadmaps section linking both historical files |
| `wiki/backend/roadmap.md:3` | Top-of-file ⚠️ HISTORICAL banner → `wiki/roadmap.md`; body untouched |
| `wiki/backlog/roadmap.md:3` | Top-of-file ⚠️ HISTORICAL banner → `wiki/roadmap.md`; body untouched |
| `wiki/backlog/index.md:6-11` | Forward-roadmap pointer added; refactor-roadmap link relabeled historical; `Last verified` bumped 2026-05-27 → 2026-06-14 |

Carried-forward items are referenced, **not re-adjudicated** (HS-6 guard) — each item's
status stays owned by its source doc. Cited spec/eval paths verified to exist
(`grade-a-…-design.md`, `eigenpal-vendor-path-design.md`, `backend/stage2-evaluation.md`).

Not yet committed — staged for the M0 close-out commit batch (operator gate HS-1).

## Verification

| Check | Command | Result |
|-------|---------|--------|
| Roadmap files present | `grep -rln "^# .*[Rr]oadmap" wiki --include="*.md"` | 3: `wiki/roadmap.md`, `wiki/backend/roadmap.md`, `wiki/backlog/roadmap.md` |
| Both old roadmaps bannered historical | `grep -rln "HISTORICAL — superseded" wiki --include="*.md"` | 2: `wiki/backend/roadmap.md`, `wiki/backlog/roadmap.md` |
| Exactly 1 forward roadmap (not historical) | set difference of the two above | **1** — `wiki/roadmap.md` only |
| Forward roadmap exists | `ls wiki/roadmap.md` | present |
| Backlog index repointed | `grep -n "wiki/roadmap.md\|historical" wiki/backlog/index.md` | forward pointer (l.6-7) + historical relabel (l.11) |
| Cited target paths exist | `ls docs/superpowers/specs/2026-06-14-{grade-a-…,eigenpal-vendor-path}-design.md wiki/backend/stage2-evaluation.md` | all present |

> Docs-only feature — no build/test/runtime surface.

## Acceptance vs milestone spec

From `../milestone.md` F0.3: *"Mark `wiki/backlog/roadmap.md` (May) and `wiki/backend/roadmap.md`
(June) **historical**; create **one** forward roadmap... | Exactly **1** forward roadmap exists;
the 2 old roadmaps are clearly labeled historical."*

| Acceptance criterion | Met? | Evidence |
|----------------------|------|----------|
| Exactly 1 forward roadmap exists | yes | 3 roadmap files − 2 historical = 1 forward (`wiki/roadmap.md`) |
| 2 old roadmaps clearly labeled historical | yes | both carry top-of-file ⚠️ HISTORICAL — superseded banner |
| Forward roadmap carries this program + post-v1 progression | yes | Grade-A M0–M5 section + post-v1 carried-forward (Plan 12, eigenpal, Wave 3) |

## Review disposition

- **Spec-compliance review:** ✅ compliant. One forward roadmap; both predecessors bannered;
  carried-forward items referenced not re-adjudicated (HS-6 guard held — no status of Plan
  10/12/13 or Wave-3 items rewritten). Backlog index no longer presents the historical
  roadmap as the plan-from surface.
- **Code-quality review:** N/A — docs-only markdown. Links verified against on-disk targets;
  relative paths (`../roadmap.md` from `backend/`, `backlog/`; `../docs/...` from `wiki/`) correct.

## Bounded defers

| Defer | Why bounded | Trigger / owner |
|-------|-------------|-----------------|
| Plan 10/12/13 + Wave-3 item **status** not reconciled into the forward roadmap | F0.3 spec is consolidation + labeling, not re-adjudication (HS-6). Forward roadmap carries them by reference; truth stays in source docs | F0.4 (backlog-hygiene) / F0.5 (archive-convention) own any status/archival reconciliation |
| Old roadmap bodies left in place (not moved to `wiki/_archive/`) | `wiki/_archive/` does not exist yet — **F0.5** owns the archive convention | F0.5 may relocate bannered historical roadmaps under `_archive/` and add governance-map rows |
