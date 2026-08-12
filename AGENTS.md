# MetalDocs Agent Routing

## Root-Cause / Global-Maximum Engineering Gate

For every non-trivial bug fix, refactor, architecture change, remediation, simplification, new abstraction, new guard, repeated review finding, or cross-boundary change, read `docs/engineering/root-cause-global-maximum-method.md` before implementation.

Before implementation, record at least:

- symptom;
- root cause;
- target property;
- authority and boundary;
- local-maximum candidate;
- global-maximum candidate;
- chosen outcome;
- enforcement layer;
- proof strategy;
- transitional exit when applicable.

Do not optimize inside a known patch or workaround. Do not remove enforcement merely to reduce code. Do not add a guard when the invalid state can instead be made unrepresentable at a stronger reasonable boundary.

## Boundary Routing

Use the smallest current source set that matches the task. Repository truth and wiki truth are authoritative; do not resurrect deleted `.agents/skills/` or retired `metaldocs-*` skill trees.

- Backend HTTP / OpenAPI / codegen / handler work -> `wiki/architecture/backend-api-structure.md`, `wiki/architecture/api-contract.md`, `wiki/architecture/api-design-system.md`.
- Database / migrations / bootstrap / grants / schema ownership -> `wiki/database/index.md` plus the relevant database pages.
- Frontend under `frontend/apps/web/` -> `wiki/architecture/frontend-structure.md` plus the owning frontend/module page.
- Frontend API / TanStack Query / generated API types -> `wiki/architecture/frontend-structure.md` query/API sections plus generated API types.
- Module or wiki documentation -> `wiki/standards/documentation-governance.md` plus the owning module docs.
- New feature/module pre-design -> `.claude/skills/developing-new-work/SKILL.md`.
- Adversarial design/plan/diff review -> `.claude/skills/adversarial-review/SKILL.md`.
- Code relationship / impact tracing when needed -> `.claude/skills/gitnexus/gitnexus-impact-analysis/SKILL.md`.
- Harness coordination when needed -> `.claude/skills/harness-hub/SKILL.md`.
- QA / close-out -> `wiki/quality/qa-operating-system.md` and the relevant `wiki/quality/*-checklist.md`.

If startup, auth/session, route truth, migrations, or runtime/spec/generated/frontend alignment is not trustworthy, treat that as a prerequisite hard stop. Repair the owning boundary and rerun the failed checkpoint before returning to feature work.

## MetalDocs AI Operating System

Use the MetalDocs operating model for all non-trivial work.

Canonical references:

- Engineering doctrine: `docs/engineering/root-cause-global-maximum-method.md`.
- QA/close-out policy: `wiki/quality/qa-operating-system.md`.
- Path-stable operating bridge: `wiki/references/ai-operating-system.md`.
- Target architecture: `wiki/architecture/backend-target-architecture.md`.
- Ordered program queue: `docs/superpowers/ROADMAP.md`.

Truth hierarchy:

1. **Runtime truth** — what actually runs now.
2. **Contract truth** — OpenAPI and generated backend/frontend surfaces.
3. **Wiki truth** — governed technical memory: module docs, debt, backlog, roadmap, ADRs.
4. **Execution truth** — scripts, preflight checks, verification commands, and gates.

Required mismatch classifications:

- runtime prerequisite;
- shared contract prerequisite;
- module-local implementation;
- screen-local implementation;
- wiki-memory drift;
- workflow/tooling gap;
- architecture contradiction;
- defer.

Default mismatch rule:

1. detect the mismatch;
2. classify it;
3. continue only when it is local to the current task boundary;
4. otherwise stop and surface the prerequisite or redesign first.

Critical contradiction stop rule: stop when contradictions affect route ownership/prefix, plan or prerequisite status, startup instructions, module ownership, API contract expectations, verification expectations, or when the correct fix is redesign-grade rather than local. Do not patch around a shared API redesign, cross-module auth/authz model change, storage/provider redesign, workflow semantic redesign, or other boundary-changing prerequisite.

## Default Workflow

1. Read `docs/engineering/root-cause-global-maximum-method.md` when the engineering gate applies.
2. Read only the wiki/architecture docs required for the task boundary.
3. Name the owning module(s), target invariant, authority, and boundary.
4. Pass the relevant prerequisite gate before implementation.
5. Implement inside the correct boundary.
6. Run static and targeted verification for the touched slice.
7. Perform code review and product QA using the relevant `wiki/quality/` checklist.
8. Classify findings by root-cause family.
9. Fix the owning family, not only the first visible symptom.
10. Rerun targeted review, QA, and regression; broaden only when boundaries were crossed.
11. Close only with evidence and explicit bounded defers.
12. Sync governed wiki truth when code or contract truth changed.

For non-trivial close-out, use as applicable:

- `wiki/quality/screen-qa-checklist.md`;
- `wiki/quality/backend-api-qa-checklist.md`;
- `wiki/quality/workflow-async-qa-checklist.md`;
- `wiki/quality/release-closeout-checklist.md`.

## Engineering Behavior

- Runtime truth beats documentation when they disagree; repair the stale documentation rather than coding against it.
- Keep changes scoped to the root cause and target property. Do not refactor unrelated adjacent code.
- Prefer one authority and one path over parallel implementations.
- YAGNI removes speculative capability and accidental complexity; it does not remove invariants, fail-closed boundaries, or proof for reachable states.
- Do not create a framework or custom guard unless the canonical engineering method justifies that enforcement layer.
- `implemented`, `fixed`, `done`, `green`, and `looks good` are not closure states by themselves. Record commands, outcomes, review disposition, QA evidence, and bounded defers.

## Context7

Use Context7 MCP for current documentation when work depends on a library, framework, SDK, API, CLI tool, or cloud service. Prefer it over remembered syntax. Do not invoke it for ordinary business-logic debugging, repository-local refactoring, or code review that does not depend on external API behavior.
