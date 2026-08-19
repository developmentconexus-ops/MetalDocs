# R10 — Technical Realization Reconciliation Baseline

> **Status:** ACTIVE / OPERATOR-RATIFIED RECONCILIATION BASELINE  
> **Ratified:** 2026-08-19  
> **Repository:** `developmentconexus-ops/MetalDocs`  
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131  
> **Program authority:** `wiki/architecture/r10-post-t6-implementation-readiness-program.md`  
> **Ratified source snapshot:** Git blob `2f6e1fa9fa44262e609c828b75ba99360fe8e42b`  
> **Implementation:** BLOCKED

This page is the durable authority created from the operator-ratified Technical Realization Reconciliation Baseline (TRRB).

The exact ratified census is preserved by the source blob above and by Git history. The former staging header/procedural text is superseded by this durable status page; the substantive evidence classifications, coverage census and stage-routing conclusions of that snapshot are ratified.

The TRRB is **not target technical design**. It proves which technical surfaces are already known, which evidence must be remeasured, which legacy documents are evidence only, and which later R10 stage must close each remaining material implementation-readiness decision.

---

## 1. Evidence classification law

Every technical-realization claim consumed after this baseline is classified as one of:

```text
CURRENT-PROVEN
  directly verified against the current branch/current durable authority

LAST-REPRODUCED
  mechanically reproduced on a prior pinned runtime baseline;
  useful as evidence but exact count/fact must be rerun before becoming current

STALE / SUPERSEDED
  contradicted by current authority or explicitly retired as target authority

UNKNOWN / REMEASURE
  not currently proven strongly enough for a target decision
```

Historical audit metrics are never promoted to current truth merely because they are convenient.

The Aug-09 architecture audit used `main@418070bf38a9f358f9131bcc36b7a6bcbc069273`, while this redesign PR is based on `main@7f5b8928cc5a13feb8ee3fa7c8ceb1c7d3655a18`. Load-bearing counts from that audit are therefore `LAST-REPRODUCED` until T8-A remeasures them.

---

## 2. Ratified coverage

The TRRB covers all technical surfaces that must be deliberately resolved before implementation:

```text
repository/runtime topology
backend modules/packages/import graph
semantic-owner physical realization
internal owner-to-owner communication
persistence/table/SQL ownership
constraints/transactions/locking
API/OpenAPI/codegen/runtime conformance
frontend routes/features/query/cache/state
editor/viewer integration boundaries
async/jobs/rendering mechanisms
runtime/process/deployment/trust topology
observability/readiness/recovery
verification/tests/CI/tools/verify
legacy technical-document authority drift
current→target transition/refactor/cutover
implementation decomposition and proof
```

No omitted implementation-readiness class is currently accepted as Writer discretion.

---

## 3. Current-state evidence posture

### Repository/runtime

CURRENT-PROVEN:

```text
one Go module (`metaldocs`)
mixed Go + React/TypeScript + SQL + Node/TypeScript repository
current API / worker / jobs / DOCX-renderer roots exist
tools/verify is the repository verification SSOT
ci.yml delegates PR verification to tools/verify
```

LAST-REPRODUCED counts such as package/module/platform-package totals are T8-A inputs, not target commitments.

### Backend/package topology

Current module names such as:

```text
approval
auth
iam
documents
controlleddocuments
templates
taxonomy
render
search
jobs
...
```

are current implementation evidence only.

Ratified semantic ownership remains:

```text
Authentication
Organization
Authorization
Controlled Documents
Audit — supporting evidence owner
```

The Aug-09 size-9 module SCC, reciprocal edges and platform→module counts remain `LAST-REPRODUCED` until T8-A remeasurement.

### Persistence

`wiki/architecture/data-model.md` remains current-database evidence only; no complete target physical relational model is inherited.

T1→T6 semantic persistence laws remain binding, including Document/Revision/WorkingContent/Submission separation, immutable Submission, system Release/effectivity, Organization/AuthZ current truth, Audit semantics, exact-content identity separation, River durable-job selection/reference and idempotency replay laws.

Aug-09 foreign-SQL counts remain `LAST-REPRODUCED` until remeasurement.

### API/OpenAPI

T6 owns semantic application operations and wire laws. Current per-module generated-package/tag topology is legacy/current realization evidence, not target entitlement.

Exact executable request/response schemas and target codegen/runtime wiring belong to T8-E.

### Frontend

Current feature folders such as `approval`, `templates`, `taxonomy`, `iam`, `documents` and `controlled-documents` are evidence only.

T6 target semantic lenses remain:

```text
Library
My Work
Document Official
Document Work
Governance Case
Document History
Audit
Administration
```

Exact frontend realization belongs to T8-F.

### Async/runtime/deploy

T4/T5/T6 correctness constraints remain binding. Current worker/jobs/render/deploy topology is evidence only.

Exact internal async contracts belong to T8-C; exact processes/deployment/trust/observability belong to T8-G; composed operational proofs belong to T9; concrete transition/cutover belongs to T10.

### Verification

Existing firing safety controls remain binding until a later ratified stage replaces their protected property.

T9 owns the target Validation Baseline and proof matrix. T12 owns final implementation-readiness challenge.

---

## 4. Technical-document authority disposition

The following are not current R10 target authority by inheritance:

```text
wiki/architecture/cohesive-platform-redesign.md   SUPERSEDED active target routing
wiki/architecture/backend-target-architecture.md  HISTORICAL prior target
wiki/architecture/data-model.md                   CURRENT-STATE DB evidence
wiki/architecture/backend-blueprint.md            CURRENT-STATE composition evidence
wiki/architecture/backend-api-structure.md        CURRENT/legacy API realization evidence
wiki/architecture/frontend-structure.md           CURRENT/legacy frontend realization evidence
wiki/backend/repo-topology.md                     CURRENT runtime/repository evidence
wiki/modules/*                                     CURRENT implementation evidence unless re-promoted
```

Fresh Actors route through `AGENTS.md` → Method → current handoff/router → Product Contract/T1→T6 → post-T6 program → current active stage.

T8-A later decides which technical documents are rewritten, promoted, replaced or deleted after target physical realization is accepted.

---

## 5. Decision-gap routing

| Missing material decision class | Owning stage |
|---|---|
| truthful historical source semantics, modes and provenance | **T7** |
| legacy technical disposition + fresh realization census | **T8-A** |
| backend package/module topology | **T8-B** |
| internal owner communication realization | **T8-C** |
| target relational schema/constraints/transactions/locks | **T8-D** |
| exact executable OpenAPI wire contract/codegen conformance | **T8-E** |
| target frontend route/feature/query/cache/editor/viewer topology | **T8-F** |
| target processes/jobs/deploy/trust/observability | **T8-G** |
| whole physical realization coherence | **T8-H** |
| Golden Flows + composed proof matrix | **T9** |
| code/schema/API/frontend/runtime transition and cutover | **T10** |
| bounded implementation dependency/execution graph | **T11** |
| independent implementation-readiness challenge | **T12** |

---

## 6. Ratified coverage verdict

```text
semantic target T1→T6                     CLOSED / PRESERVED
post-T6 stage decomposition                OPERATOR-RATIFIED
technical evidence coverage                SUFFICIENT TO ROUTE NEXT DESIGN
old exact audit metrics                     LAST-REPRODUCED / REMEASURE WHEN MATERIAL
legacy technical target docs                NOT TARGET AUTHORITY BY INHERITANCE
hidden Writer architecture decisions        PROHIBITED
TRRB                                        OPERATOR-RATIFIED
implementation                              BLOCKED
```

No package topology, table design, exact executable OpenAPI schema, frontend feature tree, runtime topology, cutover implementation or implementation sequence is selected by the TRRB.

---

## 7. Stage consequence

TRRB ratification authorizes opening **T7 — Historical Migration Truth & Semantic Mapping** only.

T7 may decide historical/source truth and semantic mapping. It may not decide physical target realization or concrete cutover implementation.

```text
TRRB ratified
→ open T7
→ actual source evidence census
→ PROVEN / INFERABLE WITH EXPLICIT RULE / UNKNOWN
→ only then migration-truth alternatives and candidate design
```

T8→T12 and product implementation remain blocked until their own gates are reached and ratified.
