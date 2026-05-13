---
name: metaldocs-screen-implementation
description: Use this skill when implementing a designed screen from `frontend/apps/web/design-source/<slug>/` into the MetalDocs feature-sliced frontend. Triggers on phrases like "implement screen X", "build the <slug> page from design", "wire up the design at design-source/<slug>", or any task that references a `design-source/<slug>/` directory with HTML + screenshot. ALWAYS run BEFORE writing any TSX or CSS for the screen. Tiered workflow (Light / Heavy) gated on evidence artifacts. Captures lessons from Library screen, novo-documento wizard, and templates wizard.
---

# MetalDocs Screen Implementation

Drives a designed screen from concept to merge-ready code on the first pass.

## The Iron Law

```
NO PHASE PROGRESSION WITHOUT EVIDENCE ARTIFACT
NO SELF-GRADED VISUAL PARITY
USER IS THE ONLY VISUAL APPROVER
```

Each phase produces an artifact under `design-source/<slug>/artifacts/`. Missing artifact = phase not done.

## Pre-requisite skill

Load `metaldocs-frontend` first. It owns feature-sliced layout, OpenAPI codegen, CSS Modules + tokens, no `HashRouter`, and no legacy paths. If the screen wires real API calls, query hooks, query keys, invalidation, optimistic updates, polling, prefetching, or freshness policy, also load `.agents/skills/metaldocs-tanstack-query/SKILL.md`.

## Tier classification (FIRST ACTION)

After Phase 0+1, classify the screen. **Default to Light unless any Heavy trigger fires.**

| Tier | Triggers | Workflow |
|---|---|---|
| **Light** | Reuses existing primitives only · ≤100 lines new CSS · no new component placed in `components/ui/` or `features/shared/` · no breakpoint-specific layout (single-column responsive OK) · no form inputs needing leakage probe | Phase 2 + **combined Phase 3** + Phase 4 + Phase 4.5 (1 subagent) |
| **Heavy** | Any: new shared primitive · new responsive layout (`@media` rule) · multiple regions with distinct typography · form-heavy with ≥3 inputs · novel interaction (drag, virtual list, etc.) | Phase 2 + Phase 3a + 3b + 3c + Phase 4 + Phase 4.5 (4 subagents) |

Tier choice goes in `IMPLEMENTATION.md` header. If unsure, ask the user once: "Light tier OK for this screen?" with the trigger checklist.

## Hard rule: ask, don't assume

Stop when: backend shape ambiguous, design element unmapped to state/role, two valid placements, status enum unclear, design vs semantic-HTML conflict, missing token. One topic per pause.

## Workflow (both tiers)

| Phase | Executor | Tier | Artifact |
|---|---|---|---|
| 0 — Audit | Main inline | both | `phase0-audit.md` (Keep/Cut/Defer + user OK) |
| 1 — Map | Main inline | both | `phase1-map.md` |
| 2 — Pre-flight | Subagent worktree | both | `phase2-preflight.md` |
| 3 — Combined (struct+style+state) | Subagent worktree | **Light** | `phase3-combined.md` + `parity-diff.md` + screenshots |
| 3a — Structure mirror | Main inline | **Heavy** | `phase3a-structure.md` (DOM diff) |
| 3b — Style port | Subagent worktree | **Heavy** | `phase3b-style.md` + `parity-diff.md` + `leakage-probe.md` (if forms) + screenshots |
| 3c — State wiring | Subagent worktree | **Heavy** | checklist in worksheet |
| 4 — Behavior verify | Main inline | both | `phase4-behavior.md` (tsc + tests + smoke) |
| 4.5 — Visual review | `frontend-screen-reviewer` | both | `phase4-review.md` |
| 5 — Document | Main + `wiki-curator` | both | wiki diff |

## Run sequence

1. Read `design-source/<slug>/NOTES.md`, view `<slug>.html` + `<slug>.png`.
2. `mkdir design-source/<slug>/artifacts/screenshots`.
3. Copy `templates/IMPLEMENTATION.md` → `design-source/<slug>/IMPLEMENTATION.md`. Fill header (slug, tier, date).
4. Phase 0 with user → `phase0-audit.md`.
5. Phase 1 with user → `phase1-map.md`. **Classify tier here.**
6. Dispatch Phase 2 subagent (`templates/subagent-phase2.md`) → `phase2-preflight.md`.
7. **If Light:** dispatch combined Phase 3 (`templates/subagent-phase3-combined.md`). User approves screenshots + parity-diff. Skip 3a/3b/3c.
8. **If Heavy:** Phase 3a inline (mirror DOM in TSX skeleton), then Phase 3b subagent, then Phase 3c subagent.
9. Phase 4 main session → `phase4-behavior.md`.
10. Phase 4.5 `frontend-screen-reviewer` → `phase4-review.md`. Resolve Critical+Major before merge.
11. Phase 5: wiki update + dispatch `wiki-curator`.

## Phase 0 — Audit

Every UI element maps to (state/role/persona/data) → Keep/Cut/Defer. Show cut list to user. Cross-ref `wiki/concepts/design-workflow-audit.md`.

## Phase 1 — Map

1.1 backward primitive scan (grep `components/ui/`, `features/shared/`).
1.2 forward placement decision tree (generic→`components/ui/`, multi-feature→`features/shared/`, domain→`features/<domain>/components/`).
1.3 component tree.
1.4 status/enum SSOT (`features/<domain>/lib/<x>Meta.ts`).
1.5 state design (server=TanStack via `metaldocs-tanstack-query`, persisted=lazy initializer, debounced=`useDebouncedValue`).
1.6 backend contract (existing vs needed; needed→mock + `wiki/backlog/<screen>.md` row).
1.7 **tier classification.**
1.8 user checkpoint.

## Phase 2 — Pre-flight (subagent)

`templates/subagent-phase2.md`. Codegen + status-meta + new shared atoms + route stub.

**Primitive audit cache:** if a primitive was audited in another screen within the last 14 days (check `wiki/modules/frontend-primitives.md` `Last verified` stamps), skip re-audit and link to the prior `phase2-preflight.md`. Otherwise audit per primitive: tokens-only, drift vs design HTML.

No tsc in Phase 2 (saves ~30s × subagent boots; tsc runs in Phase 4).

## Phase 3 (Light, combined subagent)

`templates/subagent-phase3-combined.md`. Single subagent does:

1. Mirror design HTML structure into TSX (same tags, same nesting, same DOM order).
2. Port styles to CSS Module — tokens only (`token-coverage.txt` empty).
3. Wire state (queries, error UX with `role="alert"`, four states, lazy `useState(() => readStored())`, semantic HTML — no `<button>` in `<button>`).
4. **Parity-diff at single viewport (1440)** — Pixel Parity Playbook §1 snapshot on impl + reference; numerical diff to `parity-diff.md`. Empty deltas = pass.
5. **Skip leakage-probe unless any `<input>`/`<select>`/`<textarea>`/`<label>` rendered.** If forms present: run §2 probe → `leakage-probe.md`.
6. **Skip multi-viewport screenshots unless any `@media` rule needed.** Single 1440 ref+impl pair otherwise.

User approves parity-diff (and screenshots). No self-grading.

## Phase 3a (Heavy only) — Structure mirror, INLINE

Main agent now writes the TSX skeleton + CSS Module skeleton with class names mirroring design HTML. No subagent boot. Compare DOM tree to reference HTML; mismatch = fix. Append DOM diff to `phase3a-structure.md`.

## Phase 3b (Heavy) — Style port (subagent)

`templates/subagent-phase3b.md`. Tokens-only, missing tokens added in separate commit.

**HARD requirements (Heavy):**
- `token-coverage.txt` empty (grep raw hex / px from page CSS Module).
- `parity-diff.md` numerical region-by-region — empty deltas.
- 3-viewport screenshots (1440, 1024, 375) — only because Heavy implies media queries.
- `leakage-probe.md` — only if form inputs rendered.
- User approves.

## Phase 3c (Heavy) — State wiring (subagent)

`templates/subagent-phase3c.md`. Query hooks follow `.agents/skills/metaldocs-tanstack-query/SKILL.md`: QK key factories, generated API types, targeted invalidation, conservative optimism, `ApiError` + `resolveErrorMessage`, `aria-disabled` + `title="Em breve"`, four states, `:focus-visible` outline.

## Phase 4 — Behavior verify

```bash
cd frontend/apps/web
pnpm.cmd tsc --noEmit -p tsconfig.build.json
pnpm test
```

Both green. `pnpm dev` → walk smoke steps. Write `phase4-behavior.md` (tsc + tests + smoke trace + console errors).

## Phase 4.5 — Visual review

Dispatch `frontend-screen-reviewer` with slug, page path, worksheet path, screenshots path. Resolve every Critical+Major before merge. Minor → backlog.

## Phase 5 — Document

Update `wiki/modules/<domain>.md` (bump `Last verified`, fix anchors). Dispatch `wiki-curator`.

## Subagent prompt diet

- Cap subagent prompt at ~800 tokens. No inline HTML — point to file path.
- Subagent reads design HTML itself.
- Phase 2/3 subagents do NOT run tsc (saves boots × tsc cost).
- Artifact body: ≤30 lines, bullets, no prose paragraphs.

## Red flags — STOP

| Thought | Reality |
|---|---|
| "I'll bump to Heavy because I'm not sure" | Default Light; trigger checklist is binary. Pick by trigger fire, not by feel. |
| "Skip parity-diff, screenshot looks fine" | Eyeballs lie. Numbers don't. Run §1 snapshot. |
| "Primitive was just audited, audit again" | Cache rule: 14-day fresh skip. Read prior `phase2-preflight.md`. |
| "Inline 3a even on Heavy is faster" | Heavy 3a IS inline now. Heavy uses subagent only for 3b+3c. |
| "Run tsc in subagent for safety" | tsc only in Phase 4. Subagent tsc duplicates work. |
| "Mock data without TODO" | TODO + `wiki/backlog/<screen>.md` row. Always. |
| "Self-grade Phase 3 visual" | User is sole approver. |

## Anti-patterns (instant rewrite)

- Skipping Phase 0 audit.
- Building a primitive that already exists.
- Status meta in 2+ files.
- `<button>` in `<button>`.
- Raw `alert()` for errors.
- Synchronous `useState(initial)` from `localStorage` (hydration flash) — must be lazy.
- Mock data without TODO + backlog row.
- Restructuring design HTML in TSX.
- Raw hex / px in CSS Module.
- Self-grading screenshot diff.
- Iterating fix without re-running §1 snapshot.
- Skipping 4.5 reviewer.
- Heavy-tier ceremony on a Light screen.

## Output expectations

After run, report: files changed · reusability classifications · tier · worksheet + artifacts listing · tsc/tests/smoke status · reviewer Critical/Major/Minor counts · wiki impact.

## Changelog

- 1.3 (2026-05-09) — **Tier system (Light / Heavy).** Captures templates wizard lessons: 4 subagent boots × small Light-tier screen wasted ~50% tokens. Light tier collapses Phase 3a/3b/3c into 1 combined subagent (`subagent-phase3-combined.md`). Heavy tier moves Phase 3a inline (no subagent boot for trivial DOM mirror). Conditional artifacts: parity-diff at single viewport for Light, multi-viewport only when `@media` rule present, leakage-probe only when form inputs rendered. Primitive audit cache: 14-day skip via `Last verified` stamp. Subagent prompt diet (≤800 tok, no inline HTML, no tsc). Artifact diet (≤30 lines, bullets). Target: ~85 tool calls + 1 subagent boot for Light tier (down from ~175 + 4).
- 1.2 (2026-05-07) — Visual-parity-loop refactor. `parity-diff.md` + `leakage-probe.md` HARD artifacts. Pixel Parity Playbook.
- 1.1 (2026-05-07) — Iron Law. Phase 4.5 reviewer split.
- 1.0 (2026-05-06) — initial release.
