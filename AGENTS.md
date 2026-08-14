# MetalDocs Agent Routing

## ACTIVE ARCHITECTURE RESET — READ FIRST

MetalDocs is currently in a **design-only Cohesive Platform Redesign**. This is a hard stop before product implementation or continuation of historical plans.

For any work touching product/domain architecture, authentication, organization, authorization, areas, groups, roles, permissions, approval/workflow, documents, controlled documents, templates, taxonomy, release, rendering/renditions, periodic review, distribution, notifications, search, tokens, tenant lifecycle, or their data/API boundaries, read in this exact order:

1. `wiki/standards/root-cause-global-maximum-method.md`
2. `wiki/architecture/cohesive-platform-redesign.md`
3. `docs/superpowers/analysis/2026-08-14-cohesive-platform-redesign-ledger.md`
4. `wiki/references/current-agent-handoff.md`

### Design-only gate

**NO product code, schema, migration, OpenAPI or frontend implementation is authorized by the redesign yet.**

Do not:

- resume an old roadmap/milestone/spec/plan from Git history;
- restore deleted `docs/superpowers` material into the live tree;
- implement PR #113 / historical A8 shapes by inertia;
- treat current module layout, table ownership, route topology or old ADR wording as proof that the same target boundary should survive;
- patch `documents`, `controlleddocuments`, `templates`, `taxonomy`, `approval` or IAM locally when the active ledger classifies the issue as part of the redesign;
- invent compatibility layers merely to preserve concepts that the target intends to delete.

Current runtime/schema/OpenAPI truth remains authoritative for **what runs today**. The cohesive redesign ledger is authoritative for **what the target is becoming**. Never use runtime existence as a premise that a target abstraction must survive.

## Root-Cause / Global-Maximum Engineering Gate

For every non-trivial bug fix, refactor, architecture change, remediation, simplification, new abstraction, new guard, repeated review finding, or cross-boundary change, read `wiki/standards/root-cause-global-maximum-method.md` before implementation.

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

## Documentation Authority

- `wiki/` is durable maintained truth.
- `docs/` is staging only.
- During the active redesign, `wiki/architecture/cohesive-platform-redesign.md` is the canonical program entrypoint.
- The only active detailed staging artifact under `docs/superpowers/` is `analysis/2026-08-14-cohesive-platform-redesign-ledger.md`.
- Deleted `docs/superpowers` plans/specs/milestones/reports remain available in Git history only and are **historical evidence, never forward authority**.
- Wiki pages explicitly marked `LEGACY`, `HISTORICAL`, `SUPERSEDED`, or `CURRENT-STATE REFERENCE` may be consulted to understand the existing system but must not drive target design.
- Durable final decisions from the redesign will be promoted into the owning wiki architecture/ADR pages only after operator approval.

## Truth Hierarchy

Use the hierarchy appropriate to the question.

### What runs today

1. Runtime/code/database truth.
2. OpenAPI/generated contract truth.
3. Current-state wiki documentation.
4. Historical evidence.

### What should exist after the redesign

1. Operator-approved decisions in the active cohesive redesign ledger.
2. `wiki/architecture/cohesive-platform-redesign.md`.
3. Canonical cross-cutting standards.
4. Final ADRs explicitly retained by the redesign.
5. Runtime/schema/module/legacy docs as evidence only.

When current implementation and target design disagree, do **not** “reconcile” the target back toward the current implementation. Record the current shape in the migration/deletion map later.

## Boundary Routing

Use the smallest current source set that matches the task.

- Backend HTTP / OpenAPI / codegen / handler work -> `wiki/architecture/backend-api-structure.md`, `wiki/architecture/api-contract.md`, `wiki/architecture/api-design-system.md`.
- Database / migrations / bootstrap / grants / schema ownership -> `wiki/database/index.md` plus relevant database pages.
- Frontend under `frontend/apps/web/` -> `wiki/architecture/frontend-structure.md` plus the owning frontend page.
- Frontend API / TanStack Query / generated API types -> `wiki/architecture/frontend-structure.md` query/API sections plus generated API types.
- Module or wiki documentation -> `wiki/standards/documentation-governance.md` plus the owning page.
- New feature/module pre-design -> `.claude/skills/developing-new-work/SKILL.md`, **unless the active cohesive redesign already owns the boundary**, in which case continue the redesign ledger instead of opening a parallel design authority.
- Adversarial design/plan/diff review -> `.claude/skills/adversarial-review/SKILL.md`.
- Code relationship / impact tracing -> `.claude/skills/gitnexus/gitnexus-impact-analysis/SKILL.md`.
- Harness coordination -> `.claude/skills/harness-hub/SKILL.md`.
- QA / close-out -> `wiki/quality/qa-operating-system.md` and relevant checklists.

If startup, auth/session, route truth, migrations, runtime/spec/generated/frontend alignment or the redesign authority itself is contradictory, treat that as a prerequisite hard stop. Do not patch around a shared API redesign, cross-module auth/authz redesign, storage/provider redesign, workflow semantic redesign, or other boundary-changing prerequisite.

## Default Workflow

### While Cohesive Platform Redesign gate is closed

1. Read the active redesign stack.
2. State the product/domain question being decided.
3. Census current implementation only as evidence.
4. Compare mature products/standards/libraries where useful.
5. Apply Root-Cause / Global-Maximum + YAGNI.
6. Present the smallest correct target and alternatives/trade-offs.
7. Record operator-approved decision in the active ledger.
8. Continue to the next design dependency.
9. Do **not** implement.

### After the redesign implementation gate is explicitly opened

1. Read the final promoted ADR/spec set and implementation plan.
2. Name owning bounded context, invariant and boundary.
3. Implement only the planned slice.
4. Run static + targeted verification.
5. Perform code review and product QA.
6. Fix root-cause families, not isolated symptoms.
7. Sync durable wiki truth and deletion/migration evidence.
8. Close only with explicit proof and bounded defers.

## Engineering Behavior

- Always simplify code; never simplify correctness.
- Prefer one authority and one path over parallel implementations.
- YAGNI removes speculative capability, not invariants or fail-closed behavior.
- Existing code has no right to survive the architecture reset merely because migration is inconvenient.
- Conversely, do not rewrite a healthy supporting component simply because the core is being redesigned; preserve boundaries that already match the target.
- Delete obsolete compatibility paths once their successor is authoritative.
- Do not create frameworks, policy languages, workflow engines, ReBAC engines or new services without a concrete requirement that simpler structure cannot serve.
- `implemented`, `fixed`, `done`, `green`, and `looks good` are never sufficient closure evidence by themselves.

## Context7

Use Context7 MCP for current documentation when work depends on a library, framework, SDK, API, CLI tool, or cloud service. Prefer it over remembered syntax. Do not invoke it for ordinary repository-local reasoning that does not depend on external API behavior.
