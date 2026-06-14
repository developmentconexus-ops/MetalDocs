# Feature F0.5 — Evidence

> **Milestone:** 0 — Docs Progression De-Staling  ·  **Feature:** `f0.5-archive-convention`  ·  **Closed:** 2026-06-14
> A feature is closed only when every row below is filled with real output.

## What was implemented

Established the `wiki/_archive/` convention and relocated the two F0.4 closed/superseded backlog
docs into it (git mv — history preserved), repointing every inbound link and recording the moves
in the governance migration map. The two large historical roadmaps were de-staled in place by F0.3
and are deliberately **retained in place** (documented scope decision, below).

| Change | Detail |
|--------|--------|
| `wiki/_archive/` (new) + `wiki/_archive/README.md` | Archive convention: what belongs here, the fix-relative-links-on-move rule (a `_archive/backlog/` doc sits one level deeper than `backlog/`), governance map = index of record |
| `git mv` ×2 | `backlog/api-contract-hardening.md` → `_archive/backlog/api-contract-hardening.md`; `backlog/contract-first-followups.md` → `_archive/backlog/contract-first-followups.md` (git `R`/`RM` = rename, not delete) |
| moved-file banner | api-contract-hardening banner updated "scheduled for relocation" → "archived at `wiki/_archive/backlog/`" |
| moved-file outbound links | api-contract-hardening: `roadmap.md` ×2 + `planned-endpoints.md` ×1 → `../../backlog/…` (siblings that stayed); sibling↔sibling link to contract-first-followups unchanged (co-moved, same dir) |
| inbound links repointed (3 files, 6 links) | `backlog/index.md` ×2, `backlog/planned-endpoints.md` ×1, `backlog/roadmap.md` ×3 → `../_archive/backlog/…` |
| governance migration map | `documentation-governance.md`: +2 rows (archived docs, new `_archive/` home) +2 rows (roadmaps `historical — retained in place`) |

Not yet committed — staged for the M0 close-out commit batch (operator gate HS-1).

## Scope decision — roadmaps retained in place (not a defer)

`wiki/backend/roadmap.md` + `wiki/backlog/roadmap.md` were de-staled in F0.3 with HISTORICAL
banners. **Decided not to relocate them:** the banner already does the de-staling, and physically
moving them would only spread link churn into backend tracker docs (`current-agent-handoff.md`,
`wave-h-plan.md`, `architecture-audit-2026-06-13.md`) that are part of the same frozen historical
record — no de-staling gain, added risk. The milestone objective ("exactly one forward roadmap",
F0.3) is met without relocation. Recorded in the governance map as `historical — retained in place`.

## Verification

| Check | Command | Result |
|-------|---------|--------|
| Gate A — `_archive/` tree + 2 docs | `ls wiki/_archive wiki/_archive/backlog` | `README.md`, `backlog/` → `api-contract-hardening.md`, `contract-first-followups.md` |
| Gate B — zero dangling links | `grep -rnoE "\]\([^)]*(api-contract-hardening\|contract-first-followups)\.md\)" wiki --include="*.md" \| grep -v "_archive/backlog/"` | **NONE — clean** (every link to the moved docs goes through `_archive/backlog/`) |
| Gate B — broken-link sweep on co-moved + retained siblings | per-target `test -f` on `backlog/roadmap.md`, `backlog/planned-endpoints.md`, both archived docs | all OK exists |
| Gate C — governance map rows accurate | read `documentation-governance.md` migration map | 2 archived-doc rows (new `_archive/` home) + 2 retained-in-place roadmap rows present |
| Gate D — moved not destroyed; docs-only | `git status --short` | `R`/`RM` renames (history preserved); `wiki/backlog/{both}` absent from working tree; 0 code files |

> Docs-only feature — no build/test/runtime surface. The move + link greps above are the proof.

## Acceptance vs milestone spec

From `../milestone.md` F0.5: *"`wiki/_archive/` tree exists; domain indexes + governance migration
map accurately reflect moved docs; no dangling index entries."*

| Acceptance criterion | Met? | Evidence |
|----------------------|------|----------|
| `wiki/_archive/` tree exists | yes | Gate A — dir + README + `backlog/` with both docs |
| Domain indexes reflect moved docs | yes | `backlog/index.md` "Closed/superseded" section links the new `_archive/backlog/` paths |
| Governance migration map reflects moved docs | yes | Gate C — 4 new rows (2 moved + 2 retained-in-place) |
| No dangling index entries | yes | Gate B — zero links to the moved docs bypass `_archive/`; git rename (not delete) |

## Review disposition

- **Spec-compliance review:** ✅ compliant. `_archive/` established; both queued docs relocated with
  history; every inbound link repointed; governance map is the single source of truth for the moves.
  Roadmap retain-in-place is a documented scope decision, recorded in the map (not silent).
- **Code-quality review:** N/A — docs-only. **Self-introduced bug caught + fixed at the milestone
  gate (QA-5 broken-link sweep):** the original plan assumed the move *preserved* link depth — it
  does not. `wiki/_archive/backlog/X.md` sits **one directory deeper** than `wiki/backlog/X.md`, so
  the moved docs' untouched `../`-relative outbound links dangled: 4 ADR links in api-contract-hardening
  (`../decisions/0023..0026`) and 1 tech-debt link in contract-first-followups (`../modules/…`). Fixed
  by adding one `../` to each (5 links, `../`→`../../`). Re-verified post-fix: all 7 outbound + 6 inbound
  M0-touched links resolve (`test -f` from each file's dir). The false "depth-preserved" claim was
  corrected in both `wiki/_archive/README.md` and this evidence row.

## Bounded defers

| Defer | Why bounded | Trigger / owner |
|-------|-------------|-----------------|
| (none) | F0.5 is self-contained; the F0.4 relocation hand-off is fully discharged here | — |
