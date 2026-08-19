# MetalDocs Agent Bootstrap

> **Scope:** repository-wide. This file is routing/bootstrap only; it is not methodology, stage-status, roadmap or architecture authority.

## Fresh-session read order

Before proposing or changing anything:

1. `AGENTS.md`
2. `docs/engineering/standards/root-cause-global-maximum-method.md`
3. `wiki/references/current-agent-handoff.md`
4. `wiki/architecture/r10-technical-architecture.md` — sole active R10 stage/status/next-action router
5. the durable authorities and current program/stage artifact named by that router
6. only then task-specific current code/schema/API/frontend/runtime evidence

Never infer active authority from conversation memory, Git history, legacy plans, current package names or implementation existence.

## DevelopmentConexus Engineering Method

Canonical authority: `developmentconexus-ops/conexus-methodology/METHOD.md`.

MetalDocs currently consumes version **1.0.0** through the byte-for-byte local mirror at `docs/engineering/standards/root-cause-global-maximum-method.md`. The mirror is availability/context, not a fork.

Repository specialization may operationalize the Method but must never silently weaken it. Surface conflicts inside the Method's scope.

For material independent Fable review, follow the canonical Standard Fable review workflow in `developmentconexus-ops/conexus-methodology/README.md`. Reviewer findings are evidence until operator adjudication/ratification.

## Active R10 authority

Current target architecture is routed by:

```text
wiki/references/current-agent-handoff.md
→ wiki/architecture/r10-technical-architecture.md
→ Product Contract + T1→current closed durable authority
→ wiki/architecture/r10-post-t6-implementation-readiness-program.md
→ current stage/staging named by the router
```

The current post-T6 program deliberately blocks implementation while technical realization, validation, transition and execution planning remain open.

### Do not use as active target authority

Unless the current R10 router explicitly re-promotes a decision, treat these as evidence/history only:

```text
wiki/architecture/cohesive-platform-redesign.md          SUPERSEDED target routing
wiki/architecture/backend-target-architecture.md         HISTORICAL prior target
wiki/architecture/data-model.md                          CURRENT-STATE DB evidence
wiki/architecture/backend-blueprint.md                   CURRENT-STATE composition evidence
wiki/architecture/backend-api-structure.md               CURRENT/legacy API realization evidence
wiki/architecture/frontend-structure.md                  CURRENT/legacy frontend realization evidence
wiki/backend/repo-topology.md                            CURRENT runtime/repository evidence
wiki/modules/*                                            CURRENT implementation evidence
```

Current code, DB, OpenAPI, generated types, deploy files and tests answer **what runs**, not automatically **what the R10 target must be**.

If two active authorities materially contradict, stop and surface the conflict. Do not reconcile silently in code.

## Authority by question

- **Current R10 status / next step / implementation gate:** `wiki/architecture/r10-technical-architecture.md` via `wiki/references/current-agent-handoff.md`.
- **Product and semantic target:** Product Contract REV001 + Whole-Product GCR + 4+1 ownership + closed durable R10 pages.
- **Post-T6 realization/readiness:** `wiki/architecture/r10-post-t6-implementation-readiness-program.md` and the current stage artifact routed from the R10 router.
- **What runs today:** current runtime code, database/schema/migrations, OpenAPI/generated contracts, frontend and deploy artifacts.
- **QA/close-out mechanisms:** `wiki/quality/qa-operating-system.md` where not superseded by a more specific current R10 proof decision.
- **Documentation lifecycle:** `wiki/standards/documentation-governance.md`.
- **New work outside active R10 boundaries:** `.claude/skills/developing-new-work/SKILL.md` only after confirming the active program does not already own the boundary.

## Stable safety rails

- Never expose `.env` secrets, credentials, tokens, PII or private keys.
- Keep existing tenant/security controls, generated-contract workflows, transaction guards and runtime verification intact unless accepted R10 authority explicitly replaces them with an equal-or-stronger property.
- Treat generated files as generated; use the owning generator/contract workflow.
- Do not weaken a verifier/test/guard merely to obtain green.
- Do not restore superseded architecture from Git history by inertia.
- Keep changes scoped; do not reset, revert, stash, clean or delete unrelated user work.
- Operator startup uses canonical repository PowerShell scripts; do not invent ad-hoc `.env` sourcing paths.

## Git workflow

- Work on a scoped branch/PR; never push directly to `main`.
- Never force-push/rewrite shared history without explicit operator authorization.
- Push only with explicit operator authorization.
- Commit only verified task scope.

## Verification

`tools/verify` is the repository verification authority; `.github/workflows/ci.yml` owns required PR gate composition.

Normal local PR gate:

```text
go run ./tools/verify --profile=pr
```

Use targeted checks while iterating but never substitute them for the required gate when claiming repository verification. Runtime/startup work also uses applicable runnable checks. Record command + outcome; if infrastructure is unavailable, report the limitation rather than bypassing the gate.

## Documentation discipline

Follow `wiki/standards/documentation-governance.md`:

```text
wiki/ = durable maintained authority/truth
docs/ = active staging/working evidence unless explicitly promoted
Git history = archive
```

Move/delegate authority instead of duplicating it. Bootstrap files must not become hidden architecture or milestone ledgers.