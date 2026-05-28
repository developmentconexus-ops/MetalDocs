# MetalDocs AI Operating System

> **Last verified:** 2026-05-27
> **Scope:** Path-stable compatibility bridge for safe agentic work in MetalDocs.
> **Out of scope:** Canonical QA policy, detailed module docs, API contract mechanics, and startup script implementation details.
> **Key files:**
> - `docs/superpowers/specs/2026-05-13-metaldocs-ai-operating-system-design.md` - canonical design for the operating model
> - `scripts/start-api.ps1` - startup script truth boundary
> - `scripts/check-system-runnable.ps1` - runnable preflight checkpoint
> - `wiki/quality/qa-operating-system.md` - canonical QA loop, hard-stop policy, evidence rules, and close-out contract

## What this is

This page stays path-stable because repo instructions and handoff docs still point here directly. It is a compact compatibility bridge, not a second canonical operating system.

Use it to understand truth hierarchy, classification, and prerequisite gates. For the mandatory QA loop, hard-stop behavior, and closure evidence rules, follow `wiki/quality/qa-operating-system.md`.

## The four truths

- `runtime truth` - what actually runs now: mounted routes, auth/session behavior, DB schema, and live handler wiring.
- `contract truth` - what shared consumers should rely on: OpenAPI, generated backend surfaces, and generated frontend API types.
- `wiki truth` - governed technical memory: module docs, debt registers, backlogs, ADRs, and route truth tables.
- `execution truth` - what must be run before work is trusted: scripts, preflight checks, verification commands, and skill gates.

A common failure mode is treating wiki truth as if it proves runtime truth. It does not. Wiki memory helps us work faster, but runnable evidence wins when they disagree.

## The eight classifications

- `runtime prerequisite` - startup, migration, auth/session, binary freshness, or dependency drift that makes local runtime untrustworthy.
- `shared contract prerequisite` - runtime, spec, generated code, or frontend wrapper drift that affects more than the task in front of us.
- `module-local implementation` - a bounded change that stays inside one module contract and does not change shared expectations.
- `screen-local implementation` - a bounded screen change that does not require shared runtime or contract repair.
- `wiki-memory drift` - docs or governed memory need updating after code truth changed.
- `workflow/tooling gap` - the failure exposed a missing or weak script, skill gate, or verification rule.
- `architecture contradiction` - the local task uncovered redesign-grade work that must stop the current implementation lane.
- `defer` - the task found a larger product or architecture gap that should be captured, not silently implemented.

## The five hard gates

- `Startup Gate` - start from canonical scripts, prove fresh build truth, and do not trust ad hoc startup commands.
- `Contract Gate` - compare runtime behavior, OpenAPI, generated backend surfaces, generated frontend types, and feature wrappers before changing shared behavior.
- `Screen Gate` - do not begin screen work until startup, auth/session, target route, and contract truth are all trustworthy.
- `Wiki Sync Gate` - no silent omissions; every affected module must be updated or explicitly skipped with a reason.
- `Prerequisite Exit Gate` - after a prerequisite repair, rerun the failed checkpoint and update workflow guidance if the incident exposed a gap.

## Skills we use

These are the core MetalDocs skills. Pick the smallest set that matches the task.

- `metaldocs-backend-api` - use for backend HTTP routes, OpenAPI, codegen, handler wiring, route migrations, and public API contract work.
- `metaldocs-frontend` - use for frontend implementation under `frontend/apps/web/`.
- `metaldocs-tanstack-query` - use when frontend work touches API wrappers, query hooks, generated frontend API types, cache invalidation, or server-state behavior.
- `metaldocs-screen-integration-audit` - use before real screen finalization when a visual/backlog may include mock-era widgets, missing backend capability, legacy API wrappers, or deferred behavior.
- `metaldocs-screen-implementation` - use on top of frontend workflow for designed screens under `frontend/apps/web/design-source/`.
- `metaldocs-module-doc` - use for full module wiki creation, maturity promotion, or rebuilds.
- `metaldocs-module-doc-sync` - use after implementation to sync affected module docs from a concrete change context.
- `runtime-contract-prereq` - use when startup, auth, route truth, migrations, or runtime/spec/generated/frontend-wrapper alignment is unreliable and feature work must stop.

If a task spans multiple boundaries, compose the skills rather than forcing one skill to do everything.

## How to choose a workflow

- If the task changes public HTTP behavior, start with `metaldocs-backend-api`.
- If the task changes frontend screens or components, start with `metaldocs-frontend`.
- If the frontend task also touches API calls or query state, add `metaldocs-tanstack-query`.
- If the task is a designed screen from `design-source/`, add `metaldocs-screen-integration-audit` when real capability mapping is needed, then add `metaldocs-screen-implementation` and pass the Screen Gate.
- If implementation exposed startup or contract drift, stop feature work and switch to `runtime-contract-prereq`.
- If the code change touched an already documented module, finish with `metaldocs-module-doc-sync`.
- If the module wiki is missing, stale beyond repair, or needs full structure, use `metaldocs-module-doc` instead of sync.

## Default delivery loop

This reference page is path-stable for agent instructions, but the canonical QA close-out policy now lives under `wiki/quality/`.

For every non-trivial task, the default loop is:

1. implement inside the bounded task
2. run static and targeted verification for the touched slice
3. perform code review
4. perform product QA
5. classify findings by root cause
6. fix by family
7. rerun targeted review, QA, and regression
8. rerun broader regression when the change crossed boundaries
9. close only with evidence and explicit bounded defers

This loop is mandatory by default for autonomous work. `implemented`, `fixed`, `done`, `green`, or `looks good` are not sufficient close-out states without evidence.

Default reusable checklists:

- `wiki/quality/screen-qa-checklist.md`
- `wiki/quality/backend-api-qa-checklist.md`
- `wiki/quality/workflow-async-qa-checklist.md`
- `wiki/quality/release-closeout-checklist.md`

## How to start work safely

1. Read the relevant wiki docs and the required skill guidance for the area you are touching.
2. Treat startup scripts as authoritative. This is the repo's `script-truth` policy: canonical scripts are supported, stale binaries are not trusted, and script output beats remembered commands.
3. Run the runnable preflight before screen work. A screen task starts only after login, session, and target route checks succeed.
4. Compare runtime truth and contract truth early if the task touches an HTTP surface or frontend API wrapper.
5. Stop on critical contradictions such as route ownership drift, startup instruction drift, or conflicting contract expectations.

Example: a stale binary can make route evidence lie. If an old `metaldocs-api.exe` still answers on `:8081`, you might think a route rename failed or a handler still exists. In reality, you may just be looking at yesterday's binary. Under the operating system, that is a `runtime prerequisite`: restart from the canonical script, rebuild or prove freshness, then re-check the route.

## Workflow recipes

Use these as defaults.

- Backend/API change:
  1. Read the module wiki and API architecture docs.
  2. Use `metaldocs-backend-api`.
  3. Compare runtime, OpenAPI, and generated surfaces before editing.
  4. Implement and verify.
  5. Run `metaldocs-module-doc-sync` if the module is already documented.

- Frontend screen change:
  1. Use `metaldocs-frontend`.
  2. If the screen has mock-era widgets, missing capability questions, legacy wrappers, or deferred backlog items, run `metaldocs-screen-integration-audit`.
  3. If the screen comes from `design-source/`, add `metaldocs-screen-implementation`.
  4. Pass the Startup Gate and Screen Gate before page assembly.
  5. If API/query state changes are involved, add `metaldocs-tanstack-query`.
  6. Stop and classify if runtime or contract drift appears.

- Runtime or contract drift:
  1. Stop the feature task.
  2. Use `runtime-contract-prereq`.
  3. Classify the issue.
  4. Repair only the failing boundary.
  5. Rerun the failed checkpoint before returning to feature work.

- Wiki maintenance after implementation:
  1. Name the exact change context.
  2. Use `metaldocs-module-doc-sync`.
  3. Update every affected module or explicitly report why not.
  4. Escalate to `metaldocs-module-doc` if the module needs a rebuild, not a sync.

## How to know when to stop

Keep going only when the mismatch is local to the current task boundary. Stop when the mismatch changes shared runtime or contract behavior.

Stop immediately when the required fix is redesign-grade rather than local, including:

- shared API redesign affecting multiple consumers
- cross-module auth/authz model change
- storage or provider architecture redesign
- worker or workflow semantic redesign outside the assigned boundary
- large cross-screen or frontend-backend coordinated rewrite not included in the task

When stopped, report the wrong boundary, what remains locally fixable, and the minimum prerequisite or redesign plan needed before resuming.

Example: runtime route exists but the frontend wrapper and generated types still reflect an older contract. Even if the backend endpoint responds, this is not a screen-local fix. The runtime/spec/frontend mismatch is a `shared contract prerequisite` because other callers could be wrong too. Stop the screen task, repair the shared contract surfaces, then resume feature work.

Example: during a screen task, you discover the backend endpoint the design expects does not exist at all. Do not stub around it and pretend the screen is done. Classify it as a prerequisite or `defer`, capture the missing backend work, and stop the screen implementation unless the assignment explicitly includes that backend slice.

## How skills chain together

The normal sequence is:

1. choose the task skill
2. pass the relevant gate
3. implement inside the correct boundary
4. verify with scripts and tests
5. run code review and QA using the canonical quality loop
6. classify and fix findings by family
7. rerun targeted and broader regression as required
8. sync the wiki if code truth changed

Typical chains:

- backend route change -> `metaldocs-backend-api` -> verification -> `metaldocs-module-doc-sync`
- designed screen -> `metaldocs-frontend` + `metaldocs-screen-integration-audit` when needed + `metaldocs-screen-implementation` + `metaldocs-tanstack-query` when needed -> verification
- blocked screen or API work -> `runtime-contract-prereq` -> exit gate -> return to original task

## How wiki sync works now

Wiki sync is no longer best-effort. No silent omissions are allowed.

Every sync must name the exact change context, list every affected module, and either update each one or explicitly say why it was skipped. This matters most for cross-cutting changes.

Example: if a contract change touches `templates` but also changes a shared route prefix or a dependency flow consumed by `documents`, module-doc sync must update more than one affected module. A sync that edits only `templates.md` and leaves `documents.md` stale is incomplete even if the original code change started in one module.

## How to resume feature work after a prerequisite repair

1. Write the root cause in plain language.
2. Keep the repair boundary bounded. Do not quietly roll unrelated cleanup into the prerequisite fix.
3. Rerun the exact failed checkpoint. If startup drift blocked the task, rerun startup and runnable preflight. If contract drift blocked it, rerun the contract comparison.
4. Confirm there is no hidden drift left in that repaired boundary.
5. If the incident exposed a workflow gap, update the script, skill, or runbook before resuming feature work.

Once the failed checkpoint passes again, re-enter the original task at the same boundary instead of reopening discovery from scratch.

## Explicit classification of the current split

- `wiki/quality/qa-operating-system.md` is `canonical wiki content` for the QA loop, evidence rules, and close-out contract.
- `wiki/references/ai-operating-system.md` is a `reference/archive content` + `compatibility bridge` page used because repo instructions still point here directly.
- This split is intentional for path stability right now. It should not be treated as two competing operating systems, and this page must defer to the canonical QA operating system whenever close-out behavior is in question.
