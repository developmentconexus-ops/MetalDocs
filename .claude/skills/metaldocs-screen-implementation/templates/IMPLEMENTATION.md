# <Screen Name> — Implementation Worksheet

> **Slug:** <slug>
> **Owning feature:** features/<domain>
> **Target route:** /<route>
> **Reference:** ./<slug>.html + ./<slug>.png + ./NOTES.md
> **Skill version:** 1.0
> **Started:** YYYY-MM-DD
> **Completed:** YYYY-MM-DD

---

## Open Questions Log

Append a row whenever you must pause for user input. Phase cannot pass while open rows for that phase exist.

| # | Phase | Question | User answer | Resolved |
|---|---|---|---|---|

---

## Phase 0 — Audit (HARD GATE)

Filled by main agent reading NOTES.md + design files + wiki personas/RBAC.

### 0.1 Element-by-element audit

| Element (HTML region) | Maps to (state / role / persona / data) | Keep / Cut / Defer | Reason |
|---|---|---|---|

### 0.2 Cut list confirmed

- [ ] User reviewed cut list
- [ ] Cuts recorded in NOTES.md

---

## Phase 1 — Map (HARD GATE)

### 1.1 Reusability scan — backward

Grep targets:
- `frontend/apps/web/src/components/ui/`
- `frontend/apps/web/src/features/shared/`

| Design element | Existing primitive | Path | Action (use / extend / missing) |
|---|---|---|---|

### 1.2 Reusability scan — forward

For each new component proposed:

| Name | Generic? | Used by 2+ screens? | Placement | Rationale |
|---|---|---|---|---|

Placement rules:
- Generic, no domain knowledge → `components/ui/`
- Used by 2+ features (current or planned) → `features/shared/`
- Domain-specific only → `features/<domain>/components/`

### 1.3 Component decomposition

(ASCII tree or nested list of final TSX structure using primitives from 1.1 + new from 1.2.)

### 1.4 Status / enum meta SSOT

Target file: `features/<domain>/lib/<x>Meta.ts`. Single source — no inlined records elsewhere.

| Key | Label (pt-BR) | Pill class / variant | Notes |
|---|---|---|---|

### 1.5 State design

| Type | Item | Notes |
|---|---|---|
| Server state | useXxxQuery hooks | path under features/<domain>/queries/ |
| Local state | useState/useReducer | per-component |
| Persisted | localStorage keys | lazy `useState(() => readStored())` required |
| Cross-cutting | store/ui.store.ts usage | only if truly global |
| Debounced inputs | which + ms | use `lib/hooks/useDebouncedValue` |

### 1.6 Backend contract

| Endpoint | Path | Status (existing/needed) | Shape (if needed) | Backlog issue |
|---|---|---|---|---|

Mock fallback strategy (if any "needed"):
- TODO comment block above mock data, listing endpoint + shape + backlog ref
- Disabled CTAs with `aria-disabled` + `title="Em breve"`
- Backlog file: `wiki/backlog/<screen>.md`

### 1.7 User review checkpoint

- [ ] Reusability classifications reviewed
- [ ] Backend contract reviewed
- [ ] No open Phase-1 questions

---

## Phase 2 — Pre-flight (advisory)

Subagent in worktree.

- [ ] OpenAPI codegen run (if backend endpoint added/changed)
- [ ] Primitive fixes/extensions committed (separate commits per primitive)
- [ ] Status-meta file committed: `features/<domain>/lib/<x>Meta.ts`
- [ ] New atoms (from 1.2) committed in correct location
- [ ] Route stub registered in `features/<domain>/routes.tsx`

---

## Phase 3a — Structure mirror (HARD GATE)

Subagent receives full `<slug>.html` + `<slug>.png` inline.

- [ ] DOM tree mirrors design HTML — same tag, same nesting depth, same order
- [ ] CSS Module class names = direct rename of design HTML class names (no invention)
- [ ] No logic yet — TSX skeleton only
- [ ] Main agent diffed structure vs design HTML — match confirmed

---

## Phase 3b — Style port (HARD GATE)

### 3b.1 Token map

| Design value (px / hex / rem) | Existing token | New token (if needed) |
|---|---|---|

- [ ] All design values mapped
- [ ] Missing tokens added to `frontend/apps/web/src/styles/tokens.css` or `@metaldocs/shared-tokens` in separate commit
- [ ] CSS Module uses ONLY tokens — no raw hex, no raw px for spacing
- [ ] `pnpm dev` running — visual diff vs `<slug>.png` taken
- [ ] User approved screenshot diff

---

## Phase 3c — State wiring (advisory)

- [ ] Query hooks wired per 1.5
- [ ] Error UX wired: `ApiError` + `resolveErrorMessage(code, msg)` + `role="alert"` rendering
- [ ] Disabled CTAs: `disabled aria-disabled="true" title="Em breve"`
- [ ] All four states rendered: loading, empty, error, success
- [ ] Lazy `useState(() => readStored())` for persisted values — no hydration flash
- [ ] `useDebouncedValue` for search/filter inputs per 1.5
- [ ] Semantic HTML check: no `<button>` inside `<button>`; non-button rows use `<div role="button" tabIndex={0} onClick onKeyDown>` with `:focus-visible` outline

---

## Phase 4 — Verify (HARD GATE)

```bash
cd frontend/apps/web
pnpm.cmd tsc --noEmit -p tsconfig.build.json
pnpm test
```

- [ ] tsc green
- [ ] vitest green
- [ ] Manual smoke (steps recorded below)
- [ ] Screenshot diff vs `<slug>.png` final review

Smoke steps (filled per screen):

1. ...

---

## Phase 5 — Document (advisory)

- [ ] `wiki/modules/<domain>.md` updated — `Last verified` bumped, `Key files:` anchors fixed, new patterns recorded
- [ ] `wiki/backlog/<screen>.md` created if any deferred items from 1.6
- [ ] `wiki-curator` agent dispatched
- [ ] PR description references this worksheet path
