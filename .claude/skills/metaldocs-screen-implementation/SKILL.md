---
name: metaldocs-screen-implementation
description: Use this skill when implementing a designed screen from `frontend/apps/web/design-source/<slug>/` into the MetalDocs feature-sliced frontend. Triggers on phrases like "implement screen X", "build the <slug> page from design", "wire up the design at design-source/<slug>", or any task that references a `design-source/<slug>/` directory with HTML + screenshot. ALWAYS run BEFORE writing any TSX or CSS for the screen. It enforces a 6-phase workflow with hard gates that captures lessons from the Library screen (CSS audit, status-meta SSOT, semantic HTML, error UX, no hydration flash, mock TODO trail, structure-mirror + style-port fidelity). Use alongside `metaldocs-frontend`, which remains the architecture rulebook.
---

# MetalDocs Screen Implementation

This skill drives the implementation of a designed screen from concept to merge-ready code on the first pass.

## Why this skill exists

The Library screen (`/documents`) shipped after multiple correction passes. Recurring failures: missing primitive CSS audits, status-meta sprawl across files, invalid HTML semantics (`<button>` inside `<button>`), error UX bypass (raw `alert`), hydration flash on persisted state, mock data without a TODO trail, and several iterations to match the reference HTML. This skill structures the work so those failures get caught before they ship.

Spec: `docs/superpowers/specs/2026-05-06-screen-implementation-skill-design.md`

## Pre-requisite skill

`metaldocs-frontend` (`.claude/skills/metaldocs-frontend/SKILL.md`) is the architecture rulebook — feature-sliced layout, TanStack Query for server state, OpenAPI codegen for types, CSS Modules + tokens, no `HashRouter`, no legacy paths. This skill builds on top. Load it first.

## Hard rule: ask, don't assume

Stop and ask the user when:

- Backend endpoint missing or shape ambiguous (Phase 1.6)
- Design element doesn't map to a known document state, role, or persona (Phase 0)
- Two valid component placements exist (Phase 1.2)
- Status / enum value meaning unclear (Phase 1.4)
- Design HTML conflicts with semantic HTML rules (Phase 3a)
- Token missing with no clear existing match (Phase 3b)
- Mock data fallback would hide unknown behavior

Self-deciding any of the above = skill failure. Pause, append a row to the worksheet `Open Questions Log`, wait for the user. One topic per pause — no batched dumps.

## Workflow

| Phase | Executor | Gate |
|---|---|---|
| 0 — Audit | Main agent inline | HARD — user confirms cut list |
| 1 — Map | Main agent inline | HARD — reusability + backend done, no open Phase-1 questions |
| 2 — Pre-flight | Subagent in worktree | soft — checklist |
| 3a — Structure mirror | Subagent in worktree | HARD — DOM matches design, main agent reviewed |
| 3b — Style port | Subagent in worktree | HARD — token map filled, user approved screenshot diff |
| 3c — State wiring | Subagent in worktree | soft — checklist |
| 4 — Verify | Main agent inline | HARD — tsc + tests + manual smoke green |
| 5 — Document | Main agent + wiki-curator | soft |

## Run sequence

1. Read `frontend/apps/web/design-source/<slug>/NOTES.md` and view `<slug>.html` + `<slug>.png`.
2. Copy `templates/IMPLEMENTATION.md` to `frontend/apps/web/design-source/<slug>/IMPLEMENTATION.md`. Fill the header (slug, owning feature, target route).
3. Run Phase 0 with the user.
4. Run Phase 1 with the user.
5. Dispatch Phase 2 subagent using `templates/subagent-phase2.md` as prompt body.
6. Dispatch Phase 3a subagent using `templates/subagent-phase3a.md`. Review structure diff.
7. Dispatch Phase 3b subagent using `templates/subagent-phase3b.md`. User reviews screenshot.
8. Dispatch Phase 3c subagent using `templates/subagent-phase3c.md`.
9. Run Phase 4 verify in main session.
10. Run Phase 5 doc handoff. Dispatch `wiki-curator`.

## Phase 0 — Audit

Goal: every UI element in the design has a real reason to exist (real document state, real role, real persona, real data). Cut decoration that implies behavior we do not support.

Steps:

1. Open `NOTES.md` if it exists; if not, audit the HTML directly.
2. For every region/component in the HTML, fill `IMPLEMENTATION.md` §0.1: element → maps to (state/role/persona/data) → Keep/Cut/Defer → reason. Cross-ref `wiki/concepts/design-workflow-audit.md`.
3. Show the cut list to the user and get explicit confirmation. Update `NOTES.md` with the confirmed cut list.

Hard gate: user confirmed. No TSX before this.

## Phase 1 — Map

Steps:

1. **1.1 Reusability scan — backward.** Grep `frontend/apps/web/src/components/ui/` and `frontend/apps/web/src/features/shared/`. For each design element, fill the worksheet table — primitive in use / extension needed / missing.
2. **1.2 Reusability scan — forward.** For every NEW component, classify with the placement decision tree:
   - Generic, no domain knowledge → `components/ui/`
   - Used by 2+ features (current or planned) → `features/shared/`
   - Domain-specific only → `features/<domain>/components/`
3. **1.3 Decomposition.** Component tree using primitives from 1.1 + new from 1.2.
4. **1.4 Status/enum meta SSOT.** One file: `features/<domain>/lib/<x>Meta.ts`.
5. **1.5 State design.** Server (TanStack Query hooks), local (`useState`), persisted (lazy initializer required), debounced inputs (`lib/hooks/useDebouncedValue`).
6. **1.6 Backend contract.** Existing vs needed endpoints. For "needed" → mock fallback strategy + backlog file `wiki/backlog/<screen>.md`.
7. **1.7 Checkpoint.** User reviews reusability classifications + backend contract. No open Phase-1 questions.

Hard gate: 1.7 done.

## Phase 2 — Pre-flight (subagent, worktree)

Mechanical given filled worksheet. Subagent prompt body: `templates/subagent-phase2.md`. Subagent commits separately for: codegen, primitive fixes, status-meta file, new shared atoms, route stub.

## Phase 3a — Structure mirror (subagent, worktree, HARD GATE)

Subagent prompt body: `templates/subagent-phase3a.md`. Prompt includes the full `<slug>.html` content inline. Output: TSX skeleton + CSS Module skeleton with class names mirroring design HTML class names. No logic.

Main agent reviews diff: same tag, same nesting depth, same DOM order. Mismatch → block + send back.

## Phase 3b — Style port (subagent, worktree, HARD GATE)

Subagent prompt body: `templates/subagent-phase3b.md`. Token map first; missing tokens added in a separate commit. CSS Module uses ONLY tokens — no raw hex, no raw px for spacing. User reviews screenshot vs reference before unblocking 3c.

## Phase 3c — State wiring (subagent, worktree)

Subagent prompt body: `templates/subagent-phase3c.md`. Wire query hooks, error UX (`ApiError` + `resolveErrorMessage` + `role="alert"`), disabled CTAs (`aria-disabled` + `title="Em breve"`), all four states (loading/empty/error/success), lazy `useState(() => readStored())`, `useDebouncedValue`. Semantic HTML: no `<button>` in `<button>`; non-button rows use `<div role="button" tabIndex={0} onClick onKeyDown>` with `:focus-visible` outline.

## Phase 4 — Verify (main agent, HARD GATE)

```bash
cd frontend/apps/web
pnpm.cmd tsc --noEmit -p tsconfig.build.json
pnpm test
```

Both must be green. Then run `pnpm dev` and walk the manual smoke steps recorded in §4 of the worksheet. Final screenshot diff vs `<slug>.png`.

## Phase 5 — Document

1. Update `wiki/modules/<domain>.md` — bump `Last verified`, fix `Key files:` line anchors, record any new patterns introduced.
2. If any item from §1.6 was deferred, create or update `wiki/backlog/<screen>.md` listing endpoint, shape needed, frontend wiring steps.
3. Dispatch the `wiki-curator` agent to refresh anchors and the index.
4. PR description references the worksheet path.

## Anti-patterns (instant rewrite)

- Skipping Phase 0 audit because the design "looks fine".
- Building a local component that already exists in `components/ui/` or `features/shared/`.
- Status meta inlined in two or more files.
- `<button>` inside `<button>` to add a row click target.
- Raw `alert()` for errors.
- Synchronous `useState(initial)` reading from `localStorage` (causes hydration flash) — must use lazy `useState(() => readStored())`.
- Mock data without a TODO comment block + matching `wiki/backlog/<screen>.md` row.
- Restructuring the design HTML in TSX ("I think this nesting is cleaner").
- Raw hex / px spacing in CSS Module — must use tokens.

## Output expectations

After the run, report:

1. Which files changed (with paths).
2. Reusability classifications and why each new component landed where it did.
3. Worksheet path for review.
4. Verify status (tsc, tests, manual smoke, screenshot diff).
5. Wiki impact (which docs updated; backlog file if any; whether `wiki-curator` was dispatched).

## Changelog

- 1.0 (2026-05-06) — initial release. Captures Library screen lessons.
