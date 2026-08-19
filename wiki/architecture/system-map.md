# System Map — What to Read When

> **Last verified:** 2026-08-19  
> **Active program:** R10 Post-T6 Implementation Readiness  
> **Current stage:** T8-A — Technical Authority & Legacy Census  
> **Implementation:** BLOCKED

## When in doubt

Read:

1. `../../AGENTS.md`
2. `../standards/root-cause-global-maximum-method.md`
3. `../references/current-agent-handoff.md`
4. `r10-technical-architecture.md`
5. Product Contract + T1→T7 authorities named by the router
6. `r10-post-t6-implementation-readiness-program.md`
7. `r10-technical-realization-reconciliation-baseline.md`
8. the current active stage/staging named by the router

Do not route current target work through `cohesive-platform-redesign.md`; it is superseded for active target routing.

## “I’m classifying current technical structures”

Current stage is **T8-A**.

Read the active T8-A bootstrap and inspect current code/schema/API/frontend/runtime/deploy/test evidence. Classify material structures as:

```text
PRESERVE
REFINE
REHOME
REWRITE
DELETE
CURRENT-STATE ONLY
SUPERSEDED
```

Do not use the census to select the replacement topology by stealth.

## “I’m working on historical/source migration truth”

T7 is **CLOSED**.

Durable authority:

`r10-t7-historical-migration-truth-semantic-mapping.md`

Ratified decision:

```text
NO HISTORICAL BUSINESS MIGRATION REQUIRED FOR LAUNCH
```

Current MetalDocs DB/content/history is DEV/test/throwaway and is not business-history migration evidence. T10 still owns technical current→target transition.

## “I’m deciding target backend/package architecture”

Do **not** use current `internal/modules/` or `backend-target-architecture.md` as target defaults.

Current state may be inspected during T8-A through:

```text
wiki/backend/repo-topology.md
backend-blueprint.md
current Go package/import graph
current SQL/table access
```

Target backend/package topology belongs to **T8-B**, after T8-A closes.

## “I’m deciding internal module/owner communication”

T1→T7 define semantic/product constraints. Current cross-package calls, ports, direct SQL and legacy module imports are evidence only.

Target internal owner communication belongs to **T8-C**.

## “I’m deciding target database/schema/transactions”

Current state:

1. `../database/index.md`
2. `db/baseline/`, current migrations/grants/dictionary
3. `data-model.md` only as current-state/legacy evidence

Target physical persistence belongs to **T8-D**. Current tables do not survive by existence. Concrete current→target schema/data transition belongs **T10**.

## “I’m deciding target API/OpenAPI”

T6 owns semantic operation/journey meaning.

Current implementation evidence:

```text
backend-api-structure.md
api/openapi/v1/openapi.yaml
current generated Go packages
current generated frontend types
runtime handlers/validation
```

The exact target executable wire contract belongs **T8-E**. Current module/tag/package structure is not target entitlement.

## “I’m deciding target frontend structure”

T6 owns semantic frontend lenses/journeys.

Current evidence:

```text
frontend-structure.md
frontend/apps/web/src/
current generated API types
current frontend tests/QA
```

Current features such as `approval`, `templates`, `taxonomy`, `iam`, `documents` and `controlled-documents` are current implementation evidence. Target route/feature/query/cache/editor/viewer topology belongs **T8-F**.

## “I’m deciding runtime/jobs/deploy/observability”

Current API/worker/jobs/renderer processes, River wiring, compose/Docker/scripts and observability are evidence only.

Target process/service/runtime/deployment/trust/observability realization belongs **T8-G**.

## “I’m designing end-to-end proof”

Target composed Golden Flows, failure/restart/security/recovery proofs and Validation Baseline belong **T9**.

Existing QA/checks are evidence and safety rails, not automatically the final proof architecture.

## “I’m planning how to transform the current code into the target”

Do not do this before T8/T9 close.

Current→target package/schema/API/frontend/runtime refactor, DEV/test-state disposal/reset, cutover, rollback and legacy deletion belong **T10**.

## “I’m writing implementation tasks”

Implementation decomposition belongs **T11** only after T8→T10 close.

No Writer task may contain a material architecture decision that should have been resolved earlier.

## “I’m reviewing whether implementation may start”

Use **T12 — Adversarial Implementation-Readiness Review**, followed by Integrated Whole-R10 GCR, fresh independent/cold review and explicit operator implementation authorization.

## Current-state technical evidence

The following remain useful for archaeology/current runtime only unless a later T8 decision promotes a property:

```text
backend-blueprint.md
backend-api-structure.md
frontend-structure.md
data-model.md
wiki/backend/repo-topology.md
wiki/modules/*
current code/schema/OpenAPI/deploy/tests
```

`backend-target-architecture.md` and `cohesive-platform-redesign.md` are historical/superseded target artifacts.
