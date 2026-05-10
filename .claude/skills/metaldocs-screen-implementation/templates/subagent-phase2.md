# Subagent prompt — Phase 2 Pre-flight

Fresh git worktree subagent. Main agent has filled `IMPLEMENTATION.md` Phases 0+1 and tier-classified.

## Inputs

- Worksheet: `frontend/apps/web/design-source/<SLUG>/IMPLEMENTATION.md`
- Feature: `features/<DOMAIN>` · Route: `<ROUTE>` · Tier: `<TIER>`

## Required reading

- Worksheet (Phases 0+1).
- `wiki/architecture/frontend-structure.md`.
- `wiki/modules/frontend-primitives.md` — check primitive `Last verified` stamps.

## Steps

1. **Codegen** (if §1.6 lists new endpoints not in `lib/api-types/`):
   `cd frontend/apps/web && pnpm gen:api`
   Commit: `chore(api-types): regen for <SLUG>`.

2. **Primitive audit cache.** For each REUSED primitive in §1.1:
   - If `wiki/modules/frontend-primitives.md` shows `Last verified` within 14 days → SKIP audit, link prior `phase2-preflight.md` in your output.
   - Else: read primitive `.module.css`, every value `var(--token)`, raw hex/px (besides `1px`, `0`) → fix in primitive's own commit. Bump `Last verified` stamp.

3. **Primitive extensions.** For each §1.1 row marked "extend": apply change. Commit per primitive: `feat(<primitive>): add <prop> for <SLUG>`.

4. **Global CSS leakage map.** Read `frontend/apps/web/src/styles.css`. Record every bare-element selector (`input`, `select`, `textarea`, `button`, `p`, `h2`, `label span`, `ol`, `ul`) → worksheet §2 "Global Leakage Map". Required when design has form inputs OR semantic page content.

5. **Status-meta SSOT** (if §1.4 has rows): create `features/<DOMAIN>/lib/<X>Meta.ts`. Commit: `feat(<DOMAIN>): add <x>Meta SSOT`.

6. **New shared atoms** (§1.2 placed in `components/ui/` or `features/shared/`): one commit per atom.

7. **Route stub.** Register `<ROUTE>` in `features/<DOMAIN>/routes.tsx` rendering `<div>Loading…</div>`. Commit: `feat(<DOMAIN>): register <SLUG> stub`.

## Output (≤30 lines, bullets)

`phase2-preflight.md`:
- Codegen: yes/no
- Primitives audited (with cache-skip notes)
- Primitives extended
- Leakage map: N rows added
- Status-meta path
- New atoms (paths)
- Route stub commit hash
- Skipped + why

## Hard rules

- NO page TSX (that is Phase 3).
- NO tsc — runs in Phase 4 main session.
- Each primitive change = own commit.
- Stop on open Phase-2 questions.
