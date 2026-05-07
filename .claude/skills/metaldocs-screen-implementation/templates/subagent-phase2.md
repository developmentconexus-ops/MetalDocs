# Subagent prompt — Phase 2 Pre-flight

You are a subagent dispatched in a fresh git worktree to perform Phase 2 (Pre-flight) of the MetalDocs screen-implementation workflow. The main agent has filled `IMPLEMENTATION.md` Phases 0 and 1. Your job is mechanical: prepare scaffolding so the page-assembly subagents can move fast.

## Inputs (substitute at dispatch time)

- Worksheet path: `frontend/apps/web/design-source/<SLUG>/IMPLEMENTATION.md`
- Owning feature: `features/<DOMAIN>`
- Target route: `<ROUTE>`

## Required reading before any code

1. Worksheet `IMPLEMENTATION.md` — Phases 0 + 1 fully filled.
2. `frontend/apps/web/design-source/<SLUG>/NOTES.md`.
3. `wiki/architecture/frontend-structure.md` — canonical layout.
4. `.claude/skills/metaldocs-frontend/SKILL.md` — architecture rulebook.

## Steps

1. **OpenAPI codegen.** If §1.6 of the worksheet lists any new backend endpoint as "existing" but not yet in `frontend/apps/web/src/lib/api-types/`, run codegen:
   ```bash
   cd frontend/apps/web
   pnpm gen:api
   ```
   Commit with message `chore(api-types): regen for <SLUG>`.

2. **Primitive fixes/extensions.** For each row in §1.1 marked "extend", apply the change to the existing primitive in `frontend/apps/web/src/components/ui/<Primitive>.tsx` (and module CSS). One commit per primitive: `feat(<primitive>): add <prop> for <SLUG>`.

   **2a. Primitive CSS audit (HARD).** For every REUSED primitive in §1.1, even unmodified ones, audit BEFORE page assembly:
   - Read primitive's `.module.css`. Every value must be `var(--token)`. Raw hex / raw px (other than `1px` borders, `0`) → fix in same commit and bump primitive owner.
   - Probe primitive in dev: render a sample, run Pixel Parity Playbook §1 (Computed-style snapshot) and §3 (Parent → child inheritance traps). Verify font / gap / padding match design tokens.
   - Common primitive drift to look for: `SelectableCard` `gap`/`padding`/`font:inherit` overriding child kicker rules; `Button` `font-family` not matching tokens; `Modal` body padding not using `--sp-*`.

   **2b. Global CSS leakage map (HARD).** Read `frontend/apps/web/src/styles.css` end to end. For every selector that targets a bare element (`input`, `select`, `textarea`, `button`, `p`, `h2`, `label span`, `ol`, `ul`), record the rule in worksheet §2 "Global Leakage Map". Page-assembly subagents will reset these in the page CSS Module if the design uses any of these elements. Without this map, the visual loop will play whack-a-mole. Reference table in `templates/subagent-phase3b.md` "Pixel Parity Playbook §2".

3. **Status-meta SSOT.** Create `frontend/apps/web/src/features/<DOMAIN>/lib/<X>Meta.ts` with the rows from worksheet §1.4. Commit: `feat(<DOMAIN>): add <x>Meta status SSOT`.

4. **New shared atoms.** For each row in §1.2 placed in `components/ui/` or `features/shared/`, create the file (TSX + module CSS). Atom must be self-contained (no domain logic). One commit per atom.

5. **Route stub.** Add a placeholder route to `frontend/apps/web/src/features/<DOMAIN>/routes.tsx` pointing at a stub page that renders `<div>Loading…</div>`. The page-assembly subagents will replace the stub. Commit: `feat(<DOMAIN>): register <SLUG> route stub`.

## Output

Update worksheet §2 with `[x]` against each completed item. Write a 6-line summary report:
- Codegen: yes/no
- Primitives modified (list)
- Status-meta path
- New atoms (list with paths)
- Route stub commit hash
- Anything you skipped + why

## Hard rules

- Do NOT write the page TSX. That is Phase 3.
- Do NOT touch files outside the canonical structure.
- If the worksheet has open Phase-2 questions (should be rare given Phase 1 ran clean), STOP and surface to main agent.
- Each primitive change is its own commit. No bundled commits.

## Verify before reporting done

```bash
cd frontend/apps/web
pnpm.cmd tsc --noEmit -p tsconfig.build.json
```

Must be green. If red, fix before claiming done.
