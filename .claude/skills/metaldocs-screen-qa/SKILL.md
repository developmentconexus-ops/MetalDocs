---
name: metaldocs-screen-qa
description: Use this skill to run an autonomous, closed-loop QA pass on ONE MetalDocs frontend screen, driven through the built-in Preview browser as a real user. Triggers on phrases like "QA screen X", "run screen QA for <slug>", "autonomous QA of the <screen> page", or any per-screen QA branch (`qa/<slug>`). ALWAYS the vehicle for the frontend-screen-qa-roadmap. Grounds every run in the canonical QA operating system (5 truths / 7 gates / closed loop / classification + severity / hard-stop) and produces an evidence-backed, classified per-screen QA report. Prime risk hunted: frontend ↔ contract ↔ runtime ↔ wiki drift from the Go backend quality-bar refactor.
---

# MetalDocs Screen QA

Runs a single screen through the canonical closed-loop QA model and returns an evidence-backed, classified report. One screen per run. One branch per screen (`qa/<slug>`). Designed to run in a **fresh spawned session** with no prior context.

## The Iron Law

```
NO CLOSURE WITHOUT EVIDENCE
"LOOKS GOOD" / "IMPLEMENTED" / "GREEN" IS NOT CLOSURE
DRIVE THE SCREEN AS A USER VIA PREVIEW — DO NOT INFER BEHAVIOR FROM CODE ALONE
ROOT-CAUSE BY FAMILY — NO SYMPTOM PATCHING
HARD-STOP ON REDESIGN-GRADE FIXES — REPORT, DON'T PATCH
```

## Mandatory reading (in order, do not re-grep what the wiki states)

1. `CLAUDE.md` — read order, skill routing, mandatory gates, evidence rule, hard-stop rule
2. `wiki/quality/qa-operating-system.md` — 5 truths, 7 gates, severity, classification, closed loop, hard-stops
3. `wiki/quality/screen-qa-checklist.md` — the default per-screen QA pass
4. The screen's owning module: `wiki/modules/<module>.md` + `<module>-tech-debt.md`
5. `.claude/PRPs/plans/frontend-screen-qa-roadmap.plan.md` — screen inventory, route/module/QA-class mapping, ordering, branch names
6. Conditionally:
   - `wiki/quality/workflow-async-qa-checklist.md` — approval / distribution / render-fanout screens
   - `wiki/architecture/frontend-structure.md` + `wiki/architecture/api-contract.md` — if you touch FE or contract
   - `wiki/quality/backend-api-qa-checklist.md` — if the screen exposes route/contract drift

## Prerequisite skills (route by boundary touched)

- `metaldocs-frontend` — any FE change
- `metaldocs-tanstack-query` — query keys / cache / invalidation / generated FE types
- `metaldocs-backend-api` — contract / route drift
- `runtime-contract-prereq` — when a prerequisite boundary (startup, auth/session, route, shared contract) fails

## Startup truth (every run)

- API: `.\scripts\start-api.ps1` (port :8081). Use `-Build` to force a fresh binary. Never `source .env`.
- Frontend: vite preview/dev on :4173.
- Login (dev): `admin / AdminMetalDocs123!`.
- Package manager: `corepack pnpm` — `pnpm` is NOT on PATH and bash cannot see `pnpm.cmd`. Type-check: `corepack pnpm tsc --noEmit -p tsconfig.build.json` from `frontend/apps/web`. Tests: `corepack pnpm exec vitest run <path>`.

## Browser policy (NON-NEGOTIABLE)

Gate 3 product QA is driven through the built-in **Preview** tools (`preview_*`). Do **not** use Claude-in-Chrome MCP or raw Playwright for screen QA — the system mandate is Preview for running/verifying the app.

Canonical Preview recipe per screen:

1. `preview_start` — start/attach to the running frontend (:4173). Confirm it serves.
2. Authenticate: `preview_navigate`/`preview_eval` to `/login`, `preview_fill` identifier+password, `preview_click` submit. Confirm post-login landing.
3. Navigate to the target route. `preview_snapshot` the initial load state.
4. Drive the checklist: `preview_click` / `preview_fill` / `preview_resize` to exercise happy path, empty, validation, loading/disabled/submitting, failure, stale/refresh/return-nav, authz per role.
5. Capture proof: `preview_screenshot` (visual), `preview_console_logs` (client errors), `preview_network` (API request/response truth), `preview_snapshot` (DOM/content truth).
6. For async screens, also capture the persisted/async-owner truth (poll the API or re-navigate) — see proof split below.

## The closed loop (run every gate)

### Gate 0 — Scope truth
- Confirm route ownership, owning module wiki, contract surface.
- Establish startup + auth truth (API up, login works, target route reachable).
- **Stop** if startup is unreliable, route ownership is ambiguous, or the FE expects behavior the backend does not own → route to `runtime-contract-prereq`.

### Gate 1 — Implementation truth
- `corepack pnpm tsc --noEmit -p tsconfig.build.json` → zero errors.
- `corepack pnpm exec vitest run <screen test>` → green.
- No knowingly broken path hidden behind a partial patch.

### Gate 3 — Product QA truth (Preview-driven)
Drive the `screen-qa-checklist` as the acting role(s):
- initial load honest · happy path final state correct · empty state intentional · validation messages right/at-right-time · loading/disabled/submitting prevent misleading interaction · server/network failure visible+recoverable · stale/refresh/return-nav correct · authz per role correct · optimistic→persisted settle · copy/badges don't overstate success · linked-nav destinations correct · refresh re-entry doesn't corrupt workflow · basic a11y of touched interactions.

For async / workflow-owned screens, split proof into FOUR:
1. request accepted
2. state persisted
3. async owner executed (worker/outbox/render-fanout)
4. user-visible final truth updated

### Gate 2 + Gate 4 — Review + Root cause
- Review: correct behavior, correct boundary, contract aligned, fix is local not hiding a shared problem, tests prove the real invariant.
- Classify EVERY finding before fixing:
  - family: `runtime prerequisite` · `shared contract prerequisite` · `module-local implementation` · `screen-local implementation` · `wiki-memory drift` · `workflow/tooling gap` · `architecture contradiction` · `defer`
  - severity: `critical` · `high` · `medium` · `low`
- Fix by family at the owning boundary, not the first visible symptom.

### Hard-stop rule
Stop and report (do not patch) when a fix requires: public API redesign affecting multiple consumers · cross-module auth/authz model change · storage/provider architecture redesign · worker/outbox semantic redesign outside the local boundary · migration framework/policy redesign · large cross-screen/FE-BE coordinated rewrite. Report: why it's a hard stop, which boundary is wrong, what's locally fixable vs redesign-grade, the minimum prerequisite plan.

### Gate 5 — Regression truth
- Targeted regression for touched surface (tsc + vitest + the exercised Preview paths).
- Broader affected-system regression when the change crossed boundaries.

### Gate 6 — Evidence truth
- Record validation commands, runtime Preview proof, persisted/API proof, classified findings + disposition, explicit bounded defers linked to the right wiki location.
- Update owning `wiki/modules/<module>.md` `Last verified:` when code truth changed. Dispatch `wiki-curator` for structural change.

## Pipeline mode (optional — subagent delegation)

Default flow is single-session: the main thread runs every gate itself. **Pipeline mode** is opt-in for screens with a large investigation/contract surface, to keep the main context clean. The main session always stays the **orchestrator** — it owns Gate 3 (the live Preview drive cannot be delegated), final classification, and closure. Subagents do bounded, returnable work only.

**When to use pipeline vs single-session (lite):**
- Lite (default): small screen, narrow surface, ≤2 files likely touched.
- Pipeline: broad code/contract surface, multi-module suspicion, or Gate 0 contract truth is non-obvious.

**Gate → agent mapping (each subagent returns a report; main verifies before acting):**

| Gate / phase | Agent | Scope handed over | Returns |
|---|---|---|---|
| 0 scope + locate | `cavecrew-investigator` (or `Explore`) | route ownership, owning page/components, contract surface — read-only | file:line table |
| 0/4 contract + boundary | `ecc:architect` / `ecc:code-architect` | "which boundary owns this behavior; is the fix local or shared?" — read-only | boundary verdict + minimum-fix shape |
| 2 review | `ecc:code-reviewer` (+ `ecc:typescript-reviewer`/`react-reviewer`/`go-reviewer` by language) | the diff on the QA branch | severity-tagged findings |
| 4 root fix (bounded) | `cavecrew-builder` | a **single, fully-specified** edit with file:line + exact change | caveman diff receipt |
| 5 regression | `ecc:e2e-runner` only if a Playwright suite is in scope; else main runs tsc+vitest | — | pass/fail |

**Non-delegable (main session only):**
- Gate 3 Preview drive (`preview_*`) — live browser truth must be exercised by the orchestrator.
- Final finding classification (family + severity) and disposition.
- Closure / evidence sign-off and the wiki `Last verified:` bump.

**Trust-but-verify (hard rule):** a subagent report states intent, not ground truth. Before acting on any subagent claim — especially "X is unwired / wired / already fixed / drift exists" — re-verify with a direct read or grep. A compacted or delegated summary has produced false wiki-drift claims before; never propagate one into the wiki. Builder diffs are verified by reading the actual change, not the receipt.

**Hand-over rule:** never delegate understanding. Give a builder the exact file:line and change, not "fix the bug". Give an investigator the question, not prescribed steps.

## Branch + commit convention

- Branch `qa/<screen-slug>` from `main`. One screen per branch.
- Commit findings + fixes on that branch. Open a PR per screen back to `main`.
- CI note: GitHub Actions may be billing-blocked (jobs fail ~3s). When so, local evidence (tsc + vitest + Preview runtime proof) is the gate per the QA operating system's runtime-wins + evidence rules. State this explicitly in the report.

## Per-screen QA report (the deliverable)

Return this structure. Closure requires evidence, not confidence.

```markdown
# QA Report — <Screen> (qa/<slug>)
- Route(s): <paths>   Page: <component>   Owning module: <module>   QA class: <class>
- Acting role(s): <roles exercised>   Date: <date>   CI: <green | billing-blocked → local-evidence gate>

## Gate results
- Gate 0 scope truth: <pass/fail + note>
- Gate 1 impl truth: tsc <result>, vitest <result>
- Gate 3 product QA: <checklist outcome; async 4-part split if applicable>
- Gate 5 regression: <result>

## Findings (severity-ordered)
| # | Severity | Family | Finding | Disposition (fixed / hard-stop / defer+wiki-link) |
|---|---|---|---|---|

## Evidence
- Commands run: <list>
- Runtime proof: <Preview screenshots / snapshot / console / network>
- Persisted/API proof: <endpoint + observed truth>

## Hard-stops / Bounded defers
- <boundary + minimum prerequisite plan, or "none">
```

## Self-check before claiming done

- [ ] Screen driven through Preview as a user (not inferred from code)
- [ ] All checklist paths exercised for the acting role(s)
- [ ] Every finding classified (family + severity) with disposition
- [ ] Root-cause fixes at the owning boundary, no symptom patches
- [ ] Hard-stops reported, not patched
- [ ] Evidence recorded (commands + runtime + persisted/API)
- [ ] Wiki `Last verified:` bumped where code truth changed
- [ ] (pipeline mode) Every subagent claim re-verified by direct read/grep before acting; no delegated "fact" propagated unchecked

Stop if `wiki/quality/qa-operating-system.md` or `wiki/quality/screen-qa-checklist.md` is missing — those are the canonical grounding.
