# MetalDocs AI Operating System

> **Last verified:** 2026-08-12  
> **Scope:** Path-stable compatibility bridge for safe agentic work in MetalDocs.  
> **Out of scope:** Re-defining the canonical engineering doctrine, QA policy, module rules, or startup implementation.

## Canonical sources

- Engineering doctrine: `docs/engineering/root-cause-global-maximum-method.md`
- Agent routing: `AGENTS.md`
- MetalDocs invariants/system facts: `CLAUDE.md`
- QA/close-out: `wiki/quality/qa-operating-system.md`
- Startup truth: `wiki/references/local-dev-startup.md`
- Target architecture: `wiki/architecture/backend-target-architecture.md`

This page is a compatibility bridge, not a second operating system. When definitions conflict, the canonical source above wins.

## Truth hierarchy

1. **Runtime truth** — what actually runs now: routes, auth/session behavior, DB schema, binaries, workers, and live wiring.
2. **Contract truth** — OpenAPI and generated backend/frontend surfaces shared by consumers.
3. **Wiki truth** — governed technical memory: modules, debt, roadmap, ADRs, architecture, and operational guidance.
4. **Execution truth** — supported scripts, preflight checks, verification commands, and CI gates.

Wiki memory accelerates work; it does not prove runtime behavior. If they disagree, verify runtime/repository truth and repair stale guidance.

## Classifications

Use these when a mismatch or finding appears:

- `runtime prerequisite`
- `shared contract prerequisite`
- `module-local implementation`
- `screen-local implementation`
- `wiki-memory drift`
- `workflow/tooling gap`
- `architecture contradiction`
- `defer`

A prerequisite or architecture contradiction stops local feature work until its owning boundary is repaired or deliberately deferred.

## Current workflows

The retired repository-local `metaldocs-*` and `.agents/skills/*` trees are not part of the current operating model.

Use:

- **Canonical engineering method** -> `docs/engineering/root-cause-global-maximum-method.md`
- **New feature/module pre-design** -> `.claude/skills/developing-new-work/SKILL.md`
- **Adversarial design/plan/diff review** -> `.claude/skills/adversarial-review/SKILL.md`
- **Code impact tracing when needed** -> `.claude/skills/gitnexus/SKILL.md`
- **Harness coordination when needed** -> `.claude/skills/harness-hub/SKILL.md`
- **Backend/API rules** -> `wiki/architecture/backend-api-structure.md`, `wiki/architecture/api-contract.md`, `wiki/architecture/api-design-system.md`
- **Frontend rules** -> `wiki/architecture/frontend-structure.md`
- **Database rules** -> `wiki/database/index.md` plus relevant database pages
- **Documentation governance** -> `wiki/standards/documentation-governance.md`
- **QA/close-out** -> `wiki/quality/qa-operating-system.md` plus the relevant checklist

## Default delivery loop

For non-trivial work:

1. Apply the canonical root-cause/global-maximum method.
2. Name the owning module(s), target property, authority, and boundary.
3. Read only the relevant architecture/module docs.
4. Pass startup/contract/prerequisite gates before local implementation.
5. Implement inside the correct boundary.
6. Run targeted static checks/tests.
7. Perform code review and product QA.
8. Classify findings by root-cause family.
9. Fix the owning family, not only the first visible symptom.
10. Rerun targeted verification; broaden regression only when boundaries were crossed.
11. Sync governed wiki truth when runtime or contract truth changed.
12. Close only with evidence and explicit bounded defers.

`implemented`, `fixed`, `done`, `green`, and `looks good` are not sufficient closure evidence by themselves.

## Hard gates

- **Startup Gate** — use canonical startup scripts and prove you are observing fresh binaries/runtime.
- **Contract Gate** — compare runtime behavior, OpenAPI, generated backend surfaces, generated frontend types, and wrappers before shared HTTP behavior changes.
- **Screen Gate** — do not finalize a screen while startup/auth/session/target-route/contract truth is unreliable.
- **Wiki Sync Gate** — update every affected governed module/doc or explicitly record why it is unchanged.
- **Prerequisite Exit Gate** — after repairing a prerequisite, rerun the exact failed checkpoint before resuming the original task.

## Prerequisite handling

When startup, auth/session, route truth, migrations, runtime/spec/generated alignment, or module ownership is unreliable:

1. stop the feature slice;
2. classify the mismatch;
3. identify root cause and owning boundary;
4. repair only that boundary;
5. rerun the failed checkpoint;
6. return to the original task without reopening settled discovery unless new material evidence requires it.

Do not stub around a missing backend capability, hide contract drift in a frontend wrapper, or patch a local symptom around an architecture contradiction.

## QA companions

Use as applicable:

- `wiki/quality/screen-qa-checklist.md`
- `wiki/quality/backend-api-qa-checklist.md`
- `wiki/quality/workflow-async-qa-checklist.md`
- `wiki/quality/release-closeout-checklist.md`

## Stop rule

Keep going only while findings are local to the current boundary. Stop when the correct fix changes shared ownership, contracts, authorization model, storage/provider architecture, workflow semantics, or another redesign-grade foundation.

The global-maximum rule is not permission for endless analysis. Reopen a settled decision only when new material evidence or a changed constraint invalidates it.
