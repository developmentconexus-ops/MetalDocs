# MetalDocs Screen Implementation Skill — Design

> **Last updated:** 2026-05-06
> **Status:** Design (pre-plan)
> **Author:** brainstorming session — Library screen lessons

---

## 1. Goal

Eliminate the iterate-fix-iterate loop when implementing designed screens (`frontend/apps/web/design-source/<slug>/`) into the feature-sliced React codebase. Produce a reusable workflow — a dedicated skill plus per-screen worksheet template — that delivers senior-grade results on the first pass.

## 2. Problem statement

The Library screen (`/documents`) shipped after multiple correction passes. Lessons captured:

| Failure | Root cause |
|---|---|
| Status pill / Avatar lost CSS after refactor | No primitive CSS audit before relying on `components/ui/` atoms |
| Status meta duplicated across 3 files | No status/enum SSOT step |
| `<button>` nested inside `<button>` | No semantic HTML check |
| Errors via raw `alert()` | Error UX pipeline (`ApiError` + `resolveErrorMessage`) bypassed |
| Hydration flash on tab/page-size restore | Synchronous `useState` instead of lazy initializer |
| Mock data shipped without TODO trail | No structured backlog handoff |
| Several iterations to match reference HTML | Phase 3 ran loose — no structured "mirror the design" gate |

Common thread: the worker (human or agent) executed phases out of order or skipped audits, then patched symptoms.

## 3. Solution overview

Two artifacts:

1. **Skill: `metaldocs-screen-implementation`** at `.claude/skills/metaldocs-screen-implementation/SKILL.md`. Sibling to existing `metaldocs-frontend` (which stays as the architecture rulebook). New skill imports the rulebook and adds the screen-specific workflow.

2. **Worksheet template: `IMPLEMENTATION.md`** co-located at `frontend/apps/web/design-source/<slug>/IMPLEMENTATION.md`, next to existing `<slug>.html`, `<slug>.png`, `NOTES.md`. One worksheet per screen. Lives in repo, reviewed in PRs.

Flow: skill loads worksheet → fills Phase 0/1 with user → hard-gate review → dispatches subagent for Phase 2/3 → main agent verifies + documents (Phase 4/5).

## 4. Architecture

### 4.1 Activation (decision: C)

- **Primary:** explicit user trigger — "implement screen `<slug>`" or `/implement-screen <slug>`.
- **Secondary:** auto-suggest — when `metaldocs-frontend` skill detects a task referencing `design-source/<slug>/`, it suggests invoking the screen skill.

### 4.2 Execution model (decision: Z)

| Phase | Executor | Reason |
|---|---|---|
| 0 — Audit | Main agent inline | User confirms cut list — needs full conversation context |
| 1 — Map | Main agent inline | User reviews reusability + backend decisions |
| 2 — Pre-flight | Subagent in worktree | Mechanical given filled worksheet |
| 3 — Page assembly | Subagent in worktree | Mechanical given filled worksheet + design HTML inline |
| 4 — Verify | Main agent inline | tsc/test/manual smoke needs main session |
| 5 — Document | Main agent + wiki-curator | Wiki updates need conversation memory |

Subagents receive design HTML and screenshot **inline in the prompt**, not just paths — eliminates "look it up" loops.

### 4.3 Hard rule: ask, don't assume

Cross-phase. Agent never self-decides:

- Backend endpoint missing or shape ambiguous (Phase 1.6)
- Design element doesn't map to known document state / role / persona (Phase 0)
- Two valid component placements exist (Phase 1.2)
- Status / enum value meaning unclear (Phase 1.4)
- Design HTML conflicts with semantic HTML rules (Phase 3a)
- Token missing with no clear existing match (Phase 3b)
- Mock data fallback would hide unknown behavior

Self-deciding any above = skill failure. Pause, log question in worksheet "Open Questions" table, wait for answer. One topic per pause — no batched dumps.

### 4.4 Phase gates (decision: C — hybrid hard/soft)

| Phase | Gate | Block? |
|---|---|---|
| 0 — Audit | User confirmed Keep/Cut/Defer | HARD |
| 1 — Map | Reusability scan + backend contract complete; no open questions for phase | HARD |
| 2 — Pre-flight | All worksheet items checked | soft |
| 3a — Structure mirror | DOM tree matches design HTML; main agent reviewed | HARD |
| 3b — Style port | Token map filled; user approved screenshot diff | HARD |
| 3c — State wiring | Items checked | soft |
| 4 — Verify | tsc green + tests green + manual smoke logged | HARD |
| 5 — Document | wiki-curator dispatched | soft |

## 5. Worksheet template

File: `frontend/apps/web/design-source/<slug>/IMPLEMENTATION.md`. Skill creates from template at start of run.

```markdown
# <Screen Name> — Implementation Worksheet

> **Slug:** <slug>
> **Owning feature:** features/<domain>
> **Target route:** /<route>
> **Reference:** ./<slug>.html + ./<slug>.png + ./NOTES.md
> **Skill version:** 1.0
> **Started:** YYYY-MM-DD
> **Completed:** YYYY-MM-DD

## Open Questions Log

| # | Phase | Question | User answer | Resolved |
|---|---|---|---|---|

(Append rows as questions surface. Phase cannot pass while open rows exist for that phase.)

---

## Phase 0 — Audit (HARD GATE)

Filled by: main agent reading NOTES.md + design files + wiki personas/RBAC.

### 0.1 Element-by-element audit

| Element (HTML region) | Maps to (state / role / persona / data) | Keep / Cut / Defer | Reason |
|---|---|---|---|

### 0.2 Cut list confirmed by user
- [ ] User reviewed cut list
- [ ] Cuts recorded in NOTES.md

---

## Phase 1 — Map (HARD GATE)

### 1.1 Reusability scan — backward

Grep `frontend/apps/web/src/components/ui/` and `frontend/apps/web/src/features/shared/`.

| Design element | Existing primitive | Path | Action (use / extend) |
|---|---|---|---|

Missing primitives → 1.2.

### 1.2 Reusability scan — forward

For each new component proposed:

| Name | Generic? | Used by 2+ screens? | Placement | Rationale |
|---|---|---|---|---|

Placement rules:
- Generic, no domain knowledge → `components/ui/`
- 2+ features (current or planned) → `features/shared/`
- Domain-specific only → `features/<domain>/components/`

### 1.3 Component decomposition

Tree (ASCII or nested list) showing final TSX structure using primitives from 1.1 + new from 1.2.

### 1.4 Status / enum meta SSOT

| Key | Label (pt-BR) | Pill class / variant | Notes |
|---|---|---|---|

Target file: `features/<domain>/lib/<x>Meta.ts`. Single source — no inlined records elsewhere.

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

Mock fallback strategy — if any endpoint is "needed":
- TODO comment block above mock data, listing required endpoint + shape + backlog ref
- Disabled CTAs with `aria-disabled` + `title="Em breve"`
- Backlog file: `wiki/backlog/<screen>.md`

### 1.7 User review checkpoint
- [ ] Reusability classifications reviewed
- [ ] Backend contract reviewed
- [ ] No open questions for Phase 1

---

## Phase 2 — Pre-flight (advisory)

Subagent in worktree. Mechanical given filled worksheet.

- [ ] OpenAPI codegen run (if backend endpoint added/changed)
- [ ] Primitive fixes/extensions committed (separate commits per primitive)
- [ ] Status-meta file committed: `features/<domain>/lib/<x>Meta.ts`
- [ ] New atoms (from 1.2) committed in correct location
- [ ] Route stub registered in `features/<domain>/routes.tsx`

---

## Phase 3a — Structure mirror (HARD GATE)

Subagent input includes full `<slug>.html` + `<slug>.png` inline in prompt.

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
- [ ] Missing tokens added to `styles/tokens.css` or `@metaldocs/shared-tokens` in separate commit
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

Smoke steps:
1. ...

---

## Phase 5 — Document (advisory)

- [ ] `wiki/modules/<domain>.md` updated — `Last verified` bumped, `Key files:` anchors fixed, new patterns recorded
- [ ] `wiki/backlog/<screen>.md` created if any deferred items from 1.6
- [ ] `wiki-curator` agent dispatched
- [ ] PR description references this worksheet
```

## 6. Skill content (`SKILL.md`)

File: `.claude/skills/metaldocs-screen-implementation/SKILL.md`.

Contents (outline):

1. **Why this skill exists** — references this design doc + Library lessons.
2. **Pre-requisite skill** — `metaldocs-frontend` rulebook MUST be loaded first; this skill imports its decision rules.
3. **Hard rule: ask, don't assume** — full text from §4.3.
4. **Workflow** — 6 phases with hard/soft gates table from §4.4.
5. **Phase 0 instructions** — audit checklist; how to read NOTES.md + wiki personas; where to record cut list.
6. **Phase 1 instructions** — reusability grep targets, placement decision tree, status-meta SSOT location, state design checklist, backend contract format.
7. **Phase 2 subagent prompt template** — exact prompt body for pre-flight subagent (worktree, isolated, given worksheet).
8. **Phase 3a/3b/3c subagent prompt templates** — design HTML + screenshot inline; explicit "mirror, don't restructure" instruction; token map first; state wiring last.
9. **Phase 4 verify commands** — exact `pnpm.cmd tsc --noEmit -p tsconfig.build.json`, `pnpm test`, smoke template.
10. **Phase 5 doc handoff** — wiki-curator dispatch.
11. **Anti-patterns** — copied from `metaldocs-frontend` plus screen-specific (skipping audit, building local component when shared exists, status meta in two places, etc.).
12. **Output expectations** — final report format: changed files, placement justifications, verify status, wiki impact.

## 7. File layout

New:
```
.claude/skills/metaldocs-screen-implementation/
  SKILL.md
  templates/
    IMPLEMENTATION.md         # the worksheet template
    subagent-phase2.md        # pre-flight prompt body
    subagent-phase3a.md       # structure-mirror prompt body
    subagent-phase3b.md       # style-port prompt body
    subagent-phase3c.md       # state-wiring prompt body
```

Per-screen (created at run time):
```
frontend/apps/web/design-source/<slug>/
  <slug>.html       (existing)
  <slug>.png        (existing)
  NOTES.md          (existing)
  IMPLEMENTATION.md (NEW — worksheet copy)
```

Wiki updates:
- `wiki/architecture/frontend-structure.md` — link to skill
- `wiki/concepts/design-workflow-audit.md` — note that audit is now Phase 0 of skill
- `wiki/README.md` — index entry

## 8. Out of scope

- Backend module implementation (this skill calls out backend gaps; resolution is a separate task).
- Brand-new design system primitives (handled in pre-flight as separate commits).
- Cross-screen refactors (skill is one screen at a time).
- Backfilling worksheets for already-shipped screens.

## 9. Open risks

- **Worksheet maintenance burden:** if too long, agents skip sections. Mitigation: hard gates only on the items that bit us; rest advisory.
- **Subagent context limits:** Phase 3 prompts include full HTML + screenshot. If a screen's HTML is huge, may need to split. Mitigation: advise reducing design complexity before implementation, or split into sub-screens.
- **Wiki drift on the skill itself:** skill version bumps when worksheet template changes. Skill contains a `## Changelog` section.

## 10. Success criteria

A future screen implemented with this skill produces:

1. A filled `IMPLEMENTATION.md` in `design-source/<slug>/`.
2. No more than one round of corrections after Phase 4 verify.
3. Zero raw hex/spacing in CSS Modules.
4. Zero `useEffect`-based fetching.
5. Zero `alert()` calls; all errors via `ApiError` + `resolveErrorMessage`.
6. Status meta in exactly one file.
7. All deferred items captured in `wiki/backlog/<screen>.md`.
8. `wiki/modules/<domain>.md` updated with new patterns.

## 11. Next step

Run `nexus:writing-plans` to produce task-by-task implementation plan that builds:
- The skill `SKILL.md`
- The worksheet template
- The 4 subagent prompt templates
- Wiki cross-links
