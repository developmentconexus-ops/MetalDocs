# Onboarding — Day 1

> **Last verified:** 2026-08-12  
> **For:** Any engineer or agent new to the codebase.  
> **Goal:** Get the system running locally, understand the ownership model, and ship a small verified change.

If live guidance is wrong, fix it. Runtime/repository truth wins over remembered instructions.

## 0. Read these first

| Doc | Why |
|---|---|
| [`README.md`](../README.md) | Product overview + canonical commands |
| [`AGENTS.md`](../AGENTS.md) | Model-agnostic routing, truth hierarchy, task boundaries |
| [`CLAUDE.md`](../CLAUDE.md) | MetalDocs system facts and non-negotiable invariants |
| [`docs/engineering/root-cause-global-maximum-method.md`](../docs/engineering/root-cause-global-maximum-method.md) | Required method for non-trivial engineering decisions |
| [`wiki/diagrams/c4-context.md`](diagrams/c4-context.md) | System boundary and actors |

MetalDocs is a multi-tenant SaaS for controlled documents: templates -> drafts -> approval -> frozen artifact -> PDF. Backend is Go, frontend is React, persistence is Postgres/MinIO, and `docx-renderer` is the internal Node rendering service.

## 1. Get it running

Canonical local startup:

```powershell
.\scripts\start-api.ps1
```

Use [`wiki/references/local-dev-startup.md`](references/local-dev-startup.md) as the startup truth boundary. Do not infer runtime state from a stale executable or an ad-hoc command.

Frontend:

```powershell
cd frontend/apps/web
corepack pnpm install
corepack pnpm dev
```

Before trusting an application behavior, prove the relevant runtime is fresh.

## 2. Build the mental model

Read in this order as needed:

1. [`wiki/diagrams/c4-context.md`](diagrams/c4-context.md)
2. [`wiki/diagrams/c4-container-backend.md`](diagrams/c4-container-backend.md)
3. [`wiki/architecture/system-overview.md`](architecture/system-overview.md)
4. [`wiki/architecture/backend-target-architecture.md`](architecture/backend-target-architecture.md)
5. the owning [`wiki/modules/`](modules/) page for the work you will touch

Do not re-map the entire repository for every task. Start from governed architecture, then verify only material premises against source when needed.

## 3. Role-based deep dives

### Backend / API

Read:

1. [`wiki/architecture/backend-api-structure.md`](architecture/backend-api-structure.md)
2. [`wiki/architecture/api-contract.md`](architecture/api-contract.md)
3. [`wiki/architecture/api-design-system.md`](architecture/api-design-system.md)
4. the owning module page under [`wiki/modules/`](modules/)

Routes are contract-first. Change OpenAPI and regenerate; do not hand-edit generated API artifacts.

### Frontend

Read:

1. [`wiki/architecture/frontend-structure.md`](architecture/frontend-structure.md)
2. the relevant page under [`wiki/modules/frontend/`](modules/frontend/)
3. generated frontend API types for any server-state work

Do not create a second handwritten API contract beside the generated one.

### Approval / workflow

Read:

1. [`wiki/modules/approval.md`](modules/approval.md)
2. [`wiki/modules/approval-tech-debt.md`](modules/approval-tech-debt.md)
3. relevant approval ADRs and sequence diagrams

Preserve transactional-outbox and capability-based authorization invariants.

### Database

Read:

1. [`wiki/database/index.md`](database/index.md)
2. relevant migration/schema/ownership pages
3. the owning module docs

Schema constraints/triggers are invariant backstops, not optional documentation.

### QA / close-out

Read:

1. [`wiki/quality/qa-operating-system.md`](quality/qa-operating-system.md)
2. the checklist matching the changed workflow

A green implementation is not closed work until the required evidence is recorded.

## 4. Current workflow helpers

The old repository-local `.agents/skills/` and retired `metaldocs-*` skill trees were removed during the pre-v1 re-baseline and must not be treated as live routing.

Current repository workflows:

- New module/feature pre-design -> [`.claude/skills/developing-new-work/SKILL.md`](../.claude/skills/developing-new-work/SKILL.md)
- Adversarial design/plan/diff review -> [`.claude/skills/adversarial-review/SKILL.md`](../.claude/skills/adversarial-review/SKILL.md)
- Code relationship/impact tracing -> [`.claude/skills/gitnexus/gitnexus-impact-analysis/SKILL.md`](../.claude/skills/gitnexus/gitnexus-impact-analysis/SKILL.md) when needed
- Harness coordination -> [`.claude/skills/harness-hub/SKILL.md`](../.claude/skills/harness-hub/SKILL.md) when needed
- Documentation governance -> [`wiki/standards/documentation-governance.md`](standards/documentation-governance.md)

For ordinary backend/frontend/database work, the wiki architecture/module docs plus `AGENTS.md` are the routing authority. Do not invent a missing local skill as a prerequisite.

## 5. Conventions

- **Root cause before patch.** For non-trivial work, follow `docs/engineering/root-cause-global-maximum-method.md`.
- **Runtime truth wins.** Stale docs are repaired; they are not treated as stronger than runnable/source evidence.
- **Contract-first API.** OpenAPI -> generated backend/frontend surfaces -> handlers/consumers.
- **Authorization uses capabilities, never role reasoning.** See ADR 0022 and `CLAUDE.md`.
- **Cross-module access uses published surfaces.** Never reach into another module's repository/SQL/domain internals.
- **Async side effects use transactional outbox.** No network call shares the state-write transaction.
- **Wiki is governed memory.** Update affected live docs when code/contract truth changes.
- **ADRs are historical decisions.** Supersede with a new ADR rather than silently rewriting decision history.

## 6. Your first PR

1. Pick a bounded change.
2. Identify owner, target property, and verification before editing.
3. Branch using the repository naming convention.
4. Run the checks appropriate to the touched boundary.
5. Include test/verification evidence in the PR.
6. If review exposes a repeated structural problem, stop patching and apply the root-cause/global-maximum method.

## 7. Getting unstuck

- “How does X work?” -> owning `wiki/modules/<name>.md`, then source anchors.
- “Why was X designed this way?” -> search `wiki/decisions/`.
- “Local runtime looks wrong.” -> `wiki/references/local-dev-startup.md` and the canonical scripts.
- “Runtime/spec/generated code disagree.” -> classify as a prerequisite, repair the owning boundary, rerun the failed checkpoint.
- “A reviewer suggested a patch.” -> verify the finding against source and determine the root cause before applying it.

When a local task uncovers redesign-grade work, do not hide it behind a local workaround. Surface the prerequisite or structural decision and resume the feature only after that boundary is trustworthy.
