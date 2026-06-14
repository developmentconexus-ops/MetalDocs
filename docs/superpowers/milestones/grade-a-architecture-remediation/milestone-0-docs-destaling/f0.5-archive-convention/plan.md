# Feature F0.5 — Archive Convention

> **Milestone:** 0 — Docs Progression De-Staling  ·  **Folder:** `f0.5-archive-convention`
> **Status:** Implementing

## Source

- Milestone spec row F0.5: *Implement* — "Establish `wiki/_archive/`; move superseded-historical
  docs there; update domain `index.md`s + the governance migration map to post-deletion reality."
  *Validate* — "`wiki/_archive/` tree exists; domain indexes + governance migration map accurately
  reflect moved docs; no dangling index entries."
- Constraint (milestone §Dependencies): superseded docs **archived, not destroyed**; the governance
  migration map stays the single source of truth for where moved docs went.
- Inherited from **F0.4**: `api-contract-hardening.md` (CLOSED) + `contract-first-followups.md`
  (superseded) are queued here for relocation.

## Discovery — blast radius (true markdown links only, 2026-06-14)

A raw mention-grep over-counts (21 "refs" to api-contract-hardening) — most are **prose mentions**
that do not break on a move. Counting only real `](path)` links:

| Doc to move | Inbound `](…)` link sources (excl. self + co-moved sibling) |
|-------------|------------------------------------------------------------|
| `backlog/api-contract-hardening.md` | `backlog/index.md` ×1, `backlog/planned-endpoints.md` ×1, `backlog/roadmap.md` ×3 |
| `backlog/contract-first-followups.md` | `backlog/index.md` ×1 (sibling link from api-contract-hardening co-moves) |

**Link-depth rule (corrected at gate — original assumption was WRONG):** moving
`wiki/backlog/X` → `wiki/_archive/backlog/X` puts the file **one directory deeper** (depth 2 → 3
under `wiki/`), NOT at the same depth. So every `../`-relative **outbound** link inside the moved
files (to `../decisions/`, `../modules/`, …) must gain one `../`. The co-moved siblings keep their
link to each other (same dir). Inbound: **6 links across 3 non-moved files** rewritten to point into
`_archive/backlog/`. The original "depth-preserved" wording was false; QA-5 caught 5 dangling
outbound links (`../decisions/0023..0026`, `../modules/documents-tech-debt.md`), all fixed `../`→`../../`.

## Scope decision — roadmaps retained in place (documented, not a defer)

`wiki/backend/roadmap.md` + `wiki/backlog/roadmap.md` were de-staled in **F0.3** with top-of-file
HISTORICAL banners. Physically relocating them would (a) add no de-staling value — the banner
already does it — and (b) spread link churn into backend tracker docs (`current-agent-handoff.md`,
`wave-h-plan.md`, `architecture-audit-2026-06-13.md`) that are themselves part of the same frozen
historical record. **Decision:** retain both roadmaps in place; record them in the governance
migration map as *historical — retained in place (banner-deStaled, not relocated)*. The milestone
objective ("exactly one forward roadmap") is already met by F0.3 and does **not** require physical
relocation. (Surgical-change discipline; avoids dragging the backend historical area into churn.)

## Approach

1. **Create** `wiki/_archive/` with `wiki/_archive/README.md` documenting the convention (what
   belongs here, the fix-relative-links-on-move rule, that the governance map is the index of record).
2. **Relocate** (git mv, preserve history) into `wiki/_archive/backlog/`:
   - `api-contract-hardening.md` (CLOSED program)
   - `contract-first-followups.md` (superseded)
3. **Fix the 6 inbound links** (3 files): `backlog/index.md` ×2, `backlog/planned-endpoints.md` ×1,
   `backlog/roadmap.md` ×3 → `../_archive/backlog/…`.
4. **Governance migration map** (`wiki/standards/documentation-governance.md`): add rows for the two
   relocated docs (new `wiki/_archive/backlog/…` home) + the two retained-in-place roadmaps.
5. **Verify** no dangling links to the old `backlog/` paths; `_archive/` tree exists.

## Acceptance gates (run after edits)

- **Gate A:** `wiki/_archive/` exists and contains the 2 relocated docs.
- **Gate B (hard):** zero dangling links — no surviving `](…/backlog/api-contract-hardening.md)` or
  `…/backlog/contract-first-followups.md` link anywhere in `wiki/`; broken-link sweep clean.
- **Gate C:** governance migration map has accurate rows for the 2 moved docs + 2 retained roadmaps.
- **Gate D:** moved files not destroyed (git mv, history preserved); docs-only; no code.

## Execution notes

git mv ×2 + 1 new dir/README + 6 link edits + governance rows. Direct edits. Record → `evidence.md`.
