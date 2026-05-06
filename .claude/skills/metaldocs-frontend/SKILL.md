---
name: metaldocs-frontend
description: Use this skill for ANY frontend work in MetalDocs (`frontend/apps/web/`) — new screens, new features, new components, refactors, design implementation from `frontend/apps/web/design-source/`, routing changes, state changes, API wiring. Triggers on phrases like "new screen", "implement this design", "add component", "frontend feature", "refactor frontend", "build the X page", "wire up the X UI", "add a route", or any UI work in the web app. ALWAYS use this skill before writing or editing any file under `frontend/apps/web/src/`. It enforces the canonical architecture defined in `wiki/architecture/frontend-structure.md` and prevents drift back into legacy patterns.
---

# MetalDocs Frontend

You are working on `frontend/apps/web` — a feature-sliced React SPA that is the operator UI for the MetalDocs QMS. This skill is the playbook. Follow it.

## Why this skill exists

The frontend was mid-migration from a flat layout to a feature-sliced architecture. The migration was completed in the foundation refactor (Blocks 0–6, branch `refactor/frontend-foundation`). This skill keeps every subsequent change inside the canonical structure and prevents reintroduction of the legacy patterns that were deleted.

The canonical structure, hard rules, naming, decision rules, anti-patterns, and migration policy are defined in:

- **[wiki/architecture/frontend-structure.md](../../../wiki/architecture/frontend-structure.md)** — the single source of truth. Read it first.

Companion docs to consult when relevant:

- [wiki/concepts/error-ux.md](../../../wiki/concepts/error-ux.md) — `apiFetch` / `ApiError` / auth-bus contract for error handling
- [wiki/modules/editor-ui-eigenpal.md](../../../wiki/modules/editor-ui-eigenpal.md) — eigenpal integration if you touch editor code
- [wiki/README.md](../../../wiki/README.md) — index for everything else

## Workflow (every frontend task)

### 1. Orient before coding

- Read `wiki/architecture/frontend-structure.md` end to end if you have not in the current session. It is short and load-bearing.
- If the task references a designed screen, read `frontend/apps/web/design-source/<slug>/NOTES.md` and view the `<slug>.html` reference.
- **Audit the design before implementing.** The `design-source/*.jsx` mockups were AI-generated without domain context. Run the audit from [wiki/concepts/design-workflow-audit.md](../../../wiki/concepts/design-workflow-audit.md) — walk every UI element, verify it maps to real document states / RBAC / personas, record Keep/Cut/Defer in the screen's `NOTES.md`. Cut decorative widgets that imply behavior we don't support. **No TSX before audit.**
- Locate the right feature folder (`frontend/apps/web/src/features/<domain>/`). The list of domains is in section 1 of the spec.

### 2. Apply the decision rules

Before placing any new file, run the decision rules from spec section 5:

1. Generic primitive (Button, Modal, Card, Tabs)? → `components/ui/`
2. Specific to one domain? → `features/<domain>/<correct subfolder>/`
3. Cross-cutting infra (HTTP, auth bus, error mapping, codegen types)? → `lib/`
4. Routing? → feature owns `routes.tsx`; `app/AppRouter.tsx` composes
5. Global UI state? → `store/ui.store.ts`. Otherwise feature-local.

When in doubt, stay feature-local. Promote to shared only when a second caller appears.

### 3. Wire the building blocks

Use the right tool for the right job. The defaults are not negotiable because they were chosen to keep the codebase coherent at scale:

- **Server state** — TanStack Query hooks in `features/<x>/queries/`. Never `useEffect` + `setState` for fetching.
- **API call** — thin function in `features/<x>/api/<domain>.ts` using `lib/api/client.ts` (the `openapi-fetch` instance) with types from `lib/api-types/`.
- **Local UI state** — `useState`/`useReducer`. Cross-cutting UI state — `store/ui.store.ts`. Domain state spanning components — `features/<x>/state/<x>.store.ts`.
- **Routing** — feature `routes.tsx` exporting `RouteObject[]`, lazy-loaded pages. Composed in `app/AppRouter.tsx`. Never `HashRouter`. Never string-pattern path dispatchers.
- **Styling** — CSS Modules (`<Component>.module.css`) using tokens from `styles/tokens.css` or `@metaldocs/shared-tokens`. No inline theme styles. No CSS-in-JS.
- **Error UX** — `ApiError` + `resolveErrorMessage` from `lib/api/`, toasts via `sonner`. Never raw `alert`.
- **Naming** — see spec section 4. PascalCase components, `useXxx.ts` hooks, `useXxxQuery.ts` for query hooks.

### 4. Touched legacy file? Migrate it.

Migration policy from spec section 13:

- No new code in legacy paths. (After the foundation refactor those paths are gone, but watch for re-introduction in stacked branches or merges.)
- When you edit a file that lives outside the canonical layout, move it to its canonical home in the same change. Update every importer. Delete the legacy file. No re-export shims. No "moved to X" comments.

### 5. Verify before claiming done

Run all three before reporting completion:

```bash
cd frontend/apps/web
pnpm.cmd tsc --noEmit -p tsconfig.build.json
pnpm test
# If a flow changed: pnpm e2e:smoke
```

For UI changes also exercise the screen in the dev server (`./scripts/start-api.ps1` from repo root, then `pnpm dev` in `frontend/apps/web`). Type-checking and unit tests verify code correctness, not feature correctness — if you cannot test the UI manually, say so.

### 6. Update the wiki

After each meaningful change, dispatch the `wiki-curator` agent (`.claude/agents/wiki-curator.md`). It refreshes `Last verified` stamps, fixes broken `Key files:` anchors, updates the index, and creates new docs for new modules. Do not let wiki drift accumulate — the wiki is the cold-start brain for future sessions.

## Implementing screens from design-source

When the task is "implement this designed screen":

1. Read `frontend/apps/web/design-source/<slug>/NOTES.md` to know the target route, owning feature, and any reused primitives.
2. View `<slug>.html` and `<slug>.png` to understand the visual.
3. Map HTML structure → existing `components/ui/` primitives. Extract a new primitive only when the same composition appears in two or more design files.
4. Place the page in `features/<domain>/pages/<PageName>.tsx`. CSS Module sits next to it.
5. Wire data via `features/<domain>/queries/`. Mock data is acceptable in the first commit but a follow-up commit must wire real API calls before merge.
6. Register the route in `features/<domain>/routes.tsx`.
7. Run verify steps from workflow step 5.
8. Dispatch wiki-curator to update the relevant `wiki/modules/<feature>.md` with screenshot + page anchor.

## Anti-patterns (instant rewrite)

If you catch yourself doing any of these, stop and reconsider — see spec section 15:

- Putting a feature component in `src/components/` root.
- `useEffect` + `setState` for data fetching.
- Direct `fetch` instead of `lib/api/`.
- Hand-writing types that mirror OpenAPI shapes.
- `HashRouter` or string-pattern path dispatchers.
- Cross-feature imports from one feature's internals into another. If shared, promote to `components/ui/` or `features/shared/`.
- Inline styles for theming.
- Backwards-compat re-export shims during migration.
- God component (>400 lines). Split.

## Output expectations

After any frontend task you should report:

1. What changed and which files (with paths).
2. Where it lives in the canonical structure (which feature folder, why).
3. Verification status (tsc / vitest / e2e / manual).
4. Wiki impact (which docs need updates; whether wiki-curator was dispatched).

That report is the contract — without it the change is not done.
