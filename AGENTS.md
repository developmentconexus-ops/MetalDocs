# MetalDocs Agent Bootstrap

> **Scope:** repository-wide. This file is a routing/bootstrap document, not a methodology, status, roadmap, or architecture authority.

## Fresh-session read order

Start every fresh session here, then read:

1. `docs/engineering/standards/root-cause-global-maximum-method.md` — local mirror of the organizational engineering method.
2. `wiki/references/current-agent-handoff.md` — current status/router; never infer the current stage from this file.
3. The current architecture/program authority named by the router. For the active cohesive redesign this is `wiki/architecture/cohesive-platform-redesign.md`.
4. The frozen product/domain decision artifact named by that authority when the current work derives from those decisions.
5. The active stage authority named by the router/program authority for the current design stage.
6. Only then read the owner documents for the task surface.

Do not use conversation memory, historical plans, current folder shape, or implementation existence as authority when a current owner document exists.

## DevelopmentConexus Engineering Method

Canonical authority: `developmentconexus-ops/conexus-methodology/METHOD.md`.

This repository currently consumes **version `1.0.0`** through the byte-for-byte local mirror at `docs/engineering/standards/root-cause-global-maximum-method.md`. The mirror exists only for agent availability/context; it is **not a fork or second authority**.

Updates are manual: when the organization deliberately adopts a newer Method, replace the local mirror with the chosen canonical bytes and update the consumed version here. Do not add automatic sync, submodules, packages, bots, CI synchronization, registries, or other distribution machinery without a demonstrated failure class.

MetalDocs may specialize or operationalize the Method for its product/repository, but MUST NOT silently redefine or weaken it. Surface any conflict inside the Method's scope instead of locally reinterpreting it.

## Authority and routing

Use the owner for the question being answered:

- **Current status / next step / implementation gate:** `wiki/references/current-agent-handoff.md`, routed to the active program/stage authority.
- **Target architecture / active design decisions:** `wiki/architecture/cohesive-platform-redesign.md` + the frozen product/domain decision artifact + the active stage authority named by the router while that program is active; durable decisions live under `wiki/` when promoted.
- **What runs today:** runtime code, database/schema/migrations, OpenAPI/generated contracts; current-state wiki is supporting memory.
- **Backend/API contracts:** `wiki/architecture/backend-api-structure.md`, `wiki/architecture/api-contract.md`, `wiki/architecture/api-design-system.md`.
- **Database / tenant isolation:** `wiki/database/index.md` and owning database/ADR pages.
- **Frontend:** `wiki/architecture/frontend-structure.md` and owning frontend pages.
- **QA / feature close-out:** `wiki/quality/qa-operating-system.md`. This is a repo-specific QA specialization, not a second organizational engineering method.
- **Documentation lifecycle/authority:** `wiki/standards/documentation-governance.md`.
- **New feature/module pre-design:** `.claude/skills/developing-new-work/SKILL.md` when the active architecture program does not already own the boundary. This is a repo-specific specialization of the organizational Method.
- **Impact tracing / adversarial review:** use the applicable skills under `.claude/skills/`; do not create parallel authority documents from their outputs.

If two active authorities materially contradict each other, stop and surface the conflict. Do not reconcile it silently in code or documentation.

## Stable repo safety rails

- Never read out, print, document, commit, or expose `.env` secrets or credentials.
- Keep tenant isolation/RLS, generated-contract workflows, transaction boundaries, and other existing safety controls intact unless their owning authority explicitly changes them. Never weaken verification merely to make a check pass.
- Treat generated outputs as generated; follow the owning contract/generator workflow instead of hand-maintaining parallel shapes.
- Current implementation is evidence for current state, not automatic target-architecture entitlement.
- Keep changes scoped. Do not revert unrelated user work, rewrite shared history, or restore superseded material from Git history by inertia.
- Operator startup uses the repository PowerShell scripts; do not source `.env` into a shell as an ad-hoc startup path.

## Git workflow

- Work on a scoped branch/PR; do not push directly to `main`.
- Commit only the verified task scope. Do not force-push/rewrite history unless the operator explicitly requests it and the branch is safe to rewrite.
- Push only with explicit operator authorization.

## Verification

`tools/verify` is the repository verification authority; `.github/workflows/ci.yml` owns the required PR gate composition.

Normal local PR gate:

```text
go run ./tools/verify --profile=pr
```

Use targeted checks while iterating, but do not substitute a hand-picked subset for the required gate when claiming repository verification. For runtime/startup work also use the applicable PowerShell runnable checks (for example `./scripts/check-system-runnable.ps1` from PowerShell); for frontend-only targeted tests, `make test` runs Vitest from the correct app directory.

Record the command and outcome; never claim green without evidence. If required infrastructure/tooling is unavailable, report that limitation instead of weakening or bypassing the gate.

## Documentation discipline

Follow `wiki/standards/documentation-governance.md`:

- `wiki/` holds durable maintained product/repository truth;
- `docs/` holds active staging/working material unless an owner explicitly says otherwise;
- Git history is the archive.

Move/delegate authority instead of duplicating it. Do not embed current milestones, handoffs, volatile decision lists, or temporary task instructions in bootstrap files.
