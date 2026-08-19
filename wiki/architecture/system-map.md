# System Map — What to Read When

> **Last verified:** 2026-08-19  
> **Active program:** R10 Post-T6 Implementation Readiness  
> **Implementation:** BLOCKED

## When in doubt

Read:

1. `../../AGENTS.md`
2. `../standards/root-cause-global-maximum-method.md`
3. `../references/current-agent-handoff.md`
4. `r10-technical-architecture.md`
5. Product Contract + T1→T6 authorities named by the router
6. `r10-post-t6-implementation-readiness-program.md`
7. the current active stage/staging named by the router

Do not route current target work through `cohesive-platform-redesign.md`; it is superseded for active target routing.

## “I’m deciding target backend/package architecture”

Do **not** use current `internal/modules/` or `backend-target-architecture.md` as target defaults.

Current state may be inspected through:

```text
wiki/backend/repo-topology.md
backend-blueprint.md
current Go package/import graph
current SQL/table access
```

Target backend/package topology belongs to **T8-B**, after the Technical Realization Reconciliation Baseline is accepted and T7 closes.

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

T4/T5 provide binding correctness constraints. Exact target process/deployment/trust/readiness/observability realization belongs **T8-G**.

## “I’m planning historical migration”

The former combined Historical Migration & Cutover T7 was superseded by the post-T6 Stage-Decomposition GCR.

New ownership:

```text
T7  source truth + semantic historical mapping
T10 concrete migration runtime/cutover/rollback/deletion/recovery choreography
```

Do not design concrete migration scripts before T8 has frozen the target physical realization.

## “I’m planning implementation”

Implementation planning is **not open**.

Before implementation:

```text
T7  Historical Migration Truth & Mapping
T8  Technical Realization Architecture
T9  Golden Flows & Validation Baseline
T10 Transition / Refactor / Migration / Cutover
T11 Implementation Program & Execution Graph
T12 Adversarial Implementation-Readiness
→ Integrated Whole-R10 GCR
→ fresh independent/cold review
→ operator final authorization
```

A Writer task may not contain unresolved architecture decisions.

## “I’m inspecting current runtime behavior”

Use current code/schema/OpenAPI/generated types/deploy/tests. Runtime evidence may contradict stale wiki evidence; fix/reroute stale documentation rather than granting runtime shape target authority.

## “I’m reviewing architecture”

Use the DevelopmentConexus Method. Attack duplicate/missing authority, hidden package/SQL coupling, accidental topology inheritance, contract/frontend/runtime mismatch, proof gaps and migration traps. Findings are evidence; reopen only the exact decision materially implicated.

## “I’m closing work”

Current gate is the Technical Realization Reconciliation Baseline:

`../../docs/superpowers/analysis/2026-08-19-r10-technical-realization-reconciliation-baseline.md`

It is an evidence census only. Operator acceptance of its coverage/classification is required before redefined T7 opens.

No implementation plan or product code is authorized.
