# Plan 12.1 - Templates Screen (Reality-First Execution)

> Date: 2026-05-13
> Mode: implementation plan, one screen PR
> Scope: `frontend/apps/web/design-source/templates/` -> real integration only

## Decision

Plan 12.1 is not a "build everything from the visual" pass.
It is a bounded real-integration pass based on the `metaldocs-screen-integration-audit` output in:

- `wiki/backlog/templates.md` (`Integration Audit - 2026-05-13`)
- `frontend/apps/web/design-source/templates/NOTES.md` (`Integration audit summary`)

## Required Skills

- `metaldocs-frontend`
- `metaldocs-screen-integration-audit` (already run; use its output as the boundary)
- `metaldocs-screen-implementation`
- `metaldocs-tanstack-query` when touching list query/api wrapper behavior
- `runtime-contract-prereq` only if startup/auth/route/contract gates fail again
- `verification-before-completion`

## Scope Boundary

Implement now (inside Plan 12.1):
- templates list fetch/mapping hardening on the real `v1` path (`templatesV2.ts`)
- status tabs/counts behavior
- card shell rendering consistency
- local route integrity for `/templates` (already fixed in `da8875ae`)
- mobile tab clipping fix
- card grid spacing fix
- list -> wizard handoff polish (if needed)

Prerequisites (not in this PR):
- backend/API support for true `updated_at` semantics (if required by product)
- backend/API support for author display name (replace raw `created_by`)

Deferred (not in this PR):
- description text on cards
- bound profile pills
- any new "em revisao" tab semantics beyond current 4-tab behavior

## Hard Rules

1. No fake data or placeholder backend behavior.
2. No backend feature invention inside the screen PR.
3. Keep the implementation on `api/templatesV2.ts` for this path.
4. Do not expand this PR into legacy templates API cleanup unless the touched route path still depends on it.
5. If a decision requires assumptions, stop and ask a focused clarification question.

## Task Flow

- [ ] **Task 1 - Reconfirm runnable gate before edits**
  - Run `scripts/check-system-runnable.ps1 -TargetRoute /api/v1/templates`.
  - If fail: stop and switch to `runtime-contract-prereq`.

- [ ] **Task 2 - Reconfirm contract surface gate**
  - Run `scripts/check-module-contract-sync.ps1 -Module templates`.
  - Use manual drift checks from the audit as the binding context for what is in/out of scope.

- [ ] **Task 3 - Implement screen-local keeps only**
  - Touch only list-screen files required by the "implement now" lane.
  - Keep changes bounded to templates list behavior and related shared UI styles.
  - Do not introduce new backend fields/endpoints in this PR.

- [ ] **Task 4 - Verify**
  - Frontend checks:
    - `cd frontend/apps/web`
    - `pnpm.cmd tsc --noEmit -p tsconfig.build.json`
    - `pnpm test`
  - Runtime smoke:
    - API up through canonical startup path
    - login + `/templates` load
    - tab behavior
    - create navigation
  - Capture screenshots for PR evidence.

- [ ] **Task 5 - Update backlog/notes honestly**
  - In `wiki/backlog/templates.md`, close only completed screen-local items.
  - Preserve prerequisite and deferred rows with explicit blocker rationale.
  - Keep `frontend/apps/web/design-source/templates/NOTES.md` aligned with final Keep/Prerequisite/Defer outcomes.

## Expected File Surface

Likely in-scope:
- `frontend/apps/web/src/features/templates/TemplatesListPage.tsx`
- `frontend/apps/web/src/features/templates/queries/useTemplatesQuery.ts` (if needed)
- `frontend/apps/web/src/features/templates/api/templatesV2.ts` (only list-path hardening)
- `frontend/apps/web/src/components/ui/TabBar.module.css`
- focused tests near touched screen/list behavior
- `wiki/backlog/templates.md`
- `frontend/apps/web/design-source/templates/NOTES.md`

Out-of-scope for this PR:
- `api/openapi/v1/openapi.yaml`
- `internal/modules/templates/**`
- migrations for templates timestamp/author-display behavior

## Stop Conditions

Stop and split prerequisite work if:
- list behavior needs a missing backend field/endpoint to proceed
- route/runtime/spec/generated/frontend-wrapper mismatch is shared, not local
- startup/auth/session/target route gates fail

## Completion Criteria

- Only screen-local items from the "implement now" lane are shipped.
- Prerequisites and deferred items remain explicit and honest in backlog/notes.
- No mock data path is introduced.
- Verification checks pass.
- PR diff stays bounded to templates screen integration scope.
