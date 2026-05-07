---
name: metaldocs-screen-implementation
description: Use this skill when implementing a designed screen from `frontend/apps/web/design-source/<slug>/` into the MetalDocs feature-sliced frontend. Triggers on phrases like "implement screen X", "build the <slug> page from design", "wire up the design at design-source/<slug>", or any task that references a `design-source/<slug>/` directory with HTML + screenshot. ALWAYS run BEFORE writing any TSX or CSS for the screen. Enforces a 7-phase workflow with HARD GATES gated on evidence artifacts (audit log, token coverage report, 3-viewport screenshot diff, behavior trace, reviewer report). Captures lessons from Library screen + novo-documento wizard (CSS audit, primitive drift, status-meta SSOT, semantic HTML, error UX, no hydration flash, mock TODO trail, structure-mirror + style-port fidelity, visual parity vs design HTML). Use alongside `metaldocs-frontend`.
---

# MetalDocs Screen Implementation

This skill drives the implementation of a designed screen from concept to merge-ready code on the first pass.

## The Iron Law

```
NO PHASE PROGRESSION WITHOUT EVIDENCE ARTIFACT
NO INLINE BYPASS OF SUBAGENT PHASES
NO SELF-GRADED VISUAL PARITY
```

Each phase produces a named artifact under `frontend/apps/web/design-source/<slug>/artifacts/`. If the artifact is missing or empty, the phase is **incomplete**, regardless of how the code looks.

This rule exists because the past two screens (Library, novo-documento wizard) shipped with visual gaps despite "feeling done." The pattern in both: main agent executed phases inline, skipped primitive CSS audit, self-graded the screenshot diff, and called Phase 3b complete without artifacts. Result: visual debt that took multiple correction passes to fix. The artifact requirement makes that bypass impossible.

Violating the letter of this skill is violating its spirit.

## Why this skill exists

The Library screen (`/documents`) and the novo-documento wizard (`/documents-v2/new`) both shipped after multiple correction passes. Recurring failures: skipped primitive CSS audits, status-meta sprawl across files, invalid HTML semantics (`<button>` inside `<button>`), error UX bypass (raw `alert`), hydration flash on persisted state, mock data without a TODO trail, primitives drifting from design tokens, and "looks close enough" passing as Phase 3b done. This skill structures the work so those failures get caught before they ship.

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

| Phase | Executor | Gate | Required artifact (in `artifacts/`) |
|---|---|---|---|
| 0 — Audit | Main agent inline | HARD | `phase0-audit.md` (Keep/Cut/Defer table, user signature) |
| 1 — Map | Main agent inline | HARD | `phase1-map.md` (worksheet §1 filled, no open questions) |
| 2 — Pre-flight | Subagent in worktree | HARD | `phase2-preflight.md` (primitive CSS audit + token coverage of reused atoms) |
| 3a — Structure mirror | Subagent in worktree | HARD | `phase3a-structure.md` (DOM diff vs reference, main agent reviewed) |
| 3b — Style port | Subagent in worktree | HARD | `phase3b-style.md` + `screenshots/{1440,1024,375}-{ref,impl}.png` + `token-coverage.txt` |
| 3c — State wiring | Subagent in worktree | soft | checklist in worksheet |
| 4 — Behavior verify | Main agent inline | HARD | `phase4-behavior.md` (tsc, tests, smoke trace) |
| 4.5 — Visual review | `frontend-screen-reviewer` agent | HARD | `phase4-review.md` (Critical/Major/Minor report) |
| 5 — Document | Main agent + `wiki-curator` | soft | wiki diff summary |

**Missing artifact = phase not done.** No exceptions for "it's a small screen" or "I already verified manually."

## Run sequence

1. Read `frontend/apps/web/design-source/<slug>/NOTES.md` and view `<slug>.html` + `<slug>.png`.
2. `mkdir frontend/apps/web/design-source/<slug>/artifacts` (and `artifacts/screenshots`).
3. Copy `templates/IMPLEMENTATION.md` to `frontend/apps/web/design-source/<slug>/IMPLEMENTATION.md`. Fill the header.
4. Run Phase 0 with the user → `artifacts/phase0-audit.md`.
5. Run Phase 1 with the user → `artifacts/phase1-map.md`.
6. Dispatch Phase 2 subagent (`templates/subagent-phase2.md`) → `artifacts/phase2-preflight.md`.
7. Dispatch Phase 3a subagent (`templates/subagent-phase3a.md`) → `artifacts/phase3a-structure.md`. Main agent reviews DOM diff.
8. Dispatch Phase 3b subagent (`templates/subagent-phase3b.md`) → `artifacts/phase3b-style.md` + screenshots + token-coverage. **User approves screenshot triple-diff.**
9. Dispatch Phase 3c subagent (`templates/subagent-phase3c.md`).
10. Run Phase 4 in main session → `artifacts/phase4-behavior.md`.
11. Dispatch `frontend-screen-reviewer` agent → `artifacts/phase4-review.md`. Address Critical + Major before merge.
12. Run Phase 5 doc handoff → dispatch `wiki-curator`.

## Phase 0 — Audit (HARD GATE)

Goal: every UI element in the design has a real reason to exist. Cut decoration that implies behavior we do not support.

Steps:

1. Open `NOTES.md` if it exists; if not, audit the HTML directly.
2. For every region/component in the HTML, fill `IMPLEMENTATION.md` §0.1: element → maps to (state/role/persona/data) → Keep/Cut/Defer → reason. Cross-ref `wiki/concepts/design-workflow-audit.md`.
3. Show the cut list to the user, get explicit confirmation. Update `NOTES.md` with the confirmed cut list.
4. Write `artifacts/phase0-audit.md` with the Keep/Cut/Defer table and user-confirmation timestamp.

## Phase 1 — Map (HARD GATE)

Steps:

1. **1.1 Reusability scan — backward.** Grep `frontend/apps/web/src/components/ui/` and `frontend/apps/web/src/features/shared/`. For each design element, fill the worksheet table — primitive in use / extension needed / missing.
2. **1.2 Reusability scan — forward.** Classify NEW components with placement decision tree: generic → `components/ui/`, multi-feature → `features/shared/`, domain → `features/<domain>/components/`.
3. **1.3 Decomposition.** Component tree using primitives from 1.1 + new from 1.2.
4. **1.4 Status/enum meta SSOT.** One file: `features/<domain>/lib/<x>Meta.ts`.
5. **1.5 State design.** Server (TanStack Query), local (`useState`), persisted (lazy initializer required), debounced inputs (`lib/hooks/useDebouncedValue`).
6. **1.6 Backend contract.** Existing vs needed endpoints. For "needed" → mock fallback strategy + backlog file `wiki/backlog/<screen>.md`.
7. **1.7 Checkpoint.** User reviews reusability classifications + backend contract. No open Phase-1 questions.
8. Write `artifacts/phase1-map.md` summarizing the worksheet.

## Phase 2 — Pre-flight (subagent, worktree, HARD GATE)

Subagent prompt body: `templates/subagent-phase2.md`. Mechanical given filled worksheet.

**HARD requirement: Primitive CSS audit.** Before assembling the page, the subagent audits each REUSED primitive (from §1.1) against design tokens and against the reference HTML's expectation:

- Read each primitive's CSS Module + style file.
- For every value (color, spacing, radius, font-size, shadow, line-height): is it a `var(--token)`? If not, flag.
- Compare primitive's visual against the design HTML usage. Drift = primitive needs fix BEFORE page assembly.

Subagent commits separately for: codegen, **primitive CSS fixes**, status-meta file, new shared atoms, route stub.

Output to `artifacts/phase2-preflight.md`: list of primitives audited, fixes made, residual drift accepted (with reason + user sign-off), tokens added.

## Phase 3a — Structure mirror (subagent, worktree, HARD GATE)

Subagent prompt body: `templates/subagent-phase3a.md`. Prompt includes the full `<slug>.html` content inline. Output: TSX skeleton + CSS Module skeleton with class names mirroring design HTML class names. No logic.

Main agent reviews `artifacts/phase3a-structure.md` (DOM diff): same tag, same nesting depth, same DOM order. Mismatch → block + send back.

## Phase 3b — Style port (subagent, worktree, HARD GATE)

Subagent prompt body: `templates/subagent-phase3b.md`. Token map first; missing tokens added in a separate commit. CSS Module uses ONLY tokens — no raw hex, no raw px for spacing.

**HARD requirements added:**

1. **Token coverage report** at `artifacts/token-coverage.txt`:
   ```bash
   # Run from frontend/apps/web/
   grep -REn '#[0-9a-fA-F]{3,8}|rgb\(|[0-9]+px' src/features/<domain>/pages/<Page>.module.css | grep -v 'var(--' | grep -vE '\b(0|1)px\b' > artifacts/token-coverage.txt
   ```
   Empty file = pass. Non-empty = subagent fixes before reporting done.

2. **Three-viewport screenshot triple-diff:**
   - 1440 (desktop), 1024 (tablet), 375 (mobile).
   - Save reference (HTML rendered) + implementation pairs to `artifacts/screenshots/{viewport}-{ref|impl}.png`.
   - Subagent annotates `artifacts/phase3b-style.md` with side-by-side observations per viewport.

3. **User approves the triple-diff.** Subagent does NOT self-mark approved. Main agent shows the user all 6 screenshots and waits for explicit "ok".

Without these three, Phase 3c does not start.

## Phase 3c — State wiring (subagent, worktree)

Subagent prompt body: `templates/subagent-phase3c.md`. Wire query hooks, error UX (`ApiError` + `resolveErrorMessage` + `role="alert"`), disabled CTAs (`aria-disabled` + `title="Em breve"`), all four states (loading/empty/error/success), lazy `useState(() => readStored())`, `useDebouncedValue`. Semantic HTML: no `<button>` in `<button>`; non-button rows use `<div role="button" tabIndex={0} onClick onKeyDown>` with `:focus-visible` outline.

## Phase 4 — Behavior verify (main agent, HARD GATE)

```bash
cd frontend/apps/web
pnpm.cmd tsc --noEmit -p tsconfig.build.json
pnpm test
```

Both must be green. Then run `pnpm dev` and walk the manual smoke steps recorded in §4 of the worksheet — every interactive path, every state, every keyboard route.

Write `artifacts/phase4-behavior.md`: tsc result, test result, smoke trace (step → expected → observed), any console errors.

## Phase 4.5 — Visual review (`frontend-screen-reviewer` agent, HARD GATE)

Dispatch the `frontend-screen-reviewer` agent (`.claude/agents/frontend-screen-reviewer.md`) with:
- Slug
- Implemented page path
- Worksheet path
- Phase 3b screenshots path

Agent returns `artifacts/phase4-review.md` bucketed Critical / Major / Minor. **Resolve every Critical and every Major before merge.** Minor items go to backlog.

If reviewer reports zero issues, that is fine — but the artifact must still exist.

## Phase 5 — Document

1. Update `wiki/modules/<domain>.md` — bump `Last verified`, fix `Key files:` line anchors, record any new patterns introduced.
2. If any item from §1.6 was deferred, create or update `wiki/backlog/<screen>.md`.
3. Dispatch the `wiki-curator` agent.
4. PR description references the worksheet path + reviewer report.

## Red flags — STOP and follow process

If you catch yourself thinking any of these, you are about to bypass the skill:

| Thought | Reality |
|---|---|
| "Small screen, I'll just do it inline" | Inline = no artifact = no audit trail. Use the subagents. |
| "Primitive already exists, skip the audit" | Primitives drift. The drift IS the bug. Audit. |
| "Looks close enough at this viewport" | Visual parity is at 3 viewports, side-by-side. Not "close enough". |
| "I'll mark Phase 3b done, user can review later" | User approval IS the gate. No approval = phase open. |
| "Token doesn't exist, I'll inline `#fafafa`" | Tokens-only. Add the token in a separate commit. |
| "Reviewer agent is overhead, the code is fine" | Self-grading is the failure mode that produced 2 visual-debt screens. Dispatch the reviewer. |
| "Behavior works, that's enough for Phase 4" | Phase 4 is behavior. Phase 4.5 is visual. Both required. |
| "I'll fix Major findings post-merge" | Critical + Major before merge. Minor → backlog. |

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
- **Self-grading screenshot diff** — user is the only Phase 3b approver.
- **Phase 3b artifact missing screenshots at 3 viewports** — phase not done.
- **Skipping Phase 4.5 reviewer** — phase not done.
- **Inline-executing a phase that the workflow assigns to a subagent** — bypasses isolation, no audit artifact.

## Output expectations

After the run, report:

1. Files changed (with paths).
2. Reusability classifications and why each new component landed where it did.
3. Worksheet path + artifacts directory listing.
4. Verify status: tsc, tests, manual smoke (Phase 4), reviewer Critical/Major/Minor counts (Phase 4.5).
5. Wiki impact (which docs updated; backlog file if any; whether `wiki-curator` was dispatched).

## Changelog

- 1.1 (2026-05-07) — Iron Law section. Evidence-artifact requirement per phase. Phase 2 primitive CSS audit becomes hard gate. Phase 3b adds token-coverage report + 3-viewport triple-diff + explicit user-approval gate. Phase 4 splits into 4 (behavior) + 4.5 (visual review via `frontend-screen-reviewer` agent). Red flags / rationalizations table. Anti-patterns expanded. Captures wizard-screen lessons.
- 1.0 (2026-05-06) — initial release. Captures Library screen lessons.
