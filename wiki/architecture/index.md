# Architecture

> **Last verified:** 2026-08-19  
> **Scope:** Durable architecture truth and active target-design routing.

## Active authority — read first

- `launch-v1-product-contract.md` — Product Contract REV001.
- `whole-product-alignment-review.md` — Whole-Product GCR A1–A10.
- `launch-v1-ownership-topology.md` — operator-approved 4+1 ownership.
- `r10-t1-semantic-state-invariants.md` — T1 operator-ratified.
- `r10-t2-governance-effectivity-transactions.md` — T2 operator-ratified.
- `r10-t3-authorization-audit-enforcement.md` — T3 operator-ratified.
- `r10-t3-d4-responsible-owner-eligibility-amendment.md` — bounded T3 D4 authority.
- `r10-t4-exact-content-storage-integrity-restore.md` — T4 operator-ratified.
- `r10-t5-durable-async-search-external-effects.md` — T5 operator-ratified.
- `r10-t6-canonical-api-frontend-journeys.md` — T6 operator-ratified.
- `r10-t7-historical-migration-truth-semantic-mapping.md` — **T7 operator-ratified: no historical business migration required for Launch.**
- `rebaseline-decision-registry.md` — prior-decision disposition baseline.
- `rebaseline-decision-registry-d4-amendment.md` — D4 reconciliation.
- `rebaseline-decision-registry-t6-amendment.md` — T6 closure reconciliation.
- `rebaseline-decision-registry-post-t6-amendment.md` — post-T6 stage-ownership reconciliation.
- `rebaseline-decision-registry-t7-amendment.md` — **T7 closure reconciliation.**
- `r10-post-t6-implementation-readiness-program.md` — operator-ratified T7→T12 program authority.
- `r10-technical-realization-reconciliation-baseline.md` — operator-ratified technical census/routing baseline.
- `r10-technical-architecture.md` — sole current stage/status/next-action router.
- `../references/current-agent-handoff.md` — fresh-session recovery point.

## Current gate

```text
Product Contract / GCR / ownership       APPROVED
T1→T7                                    CLOSED / OPERATOR-RATIFIED
Decision Registry                        CURRENT + D4 + T6 + post-T6 + T7 amendments
TRRB                                     CLOSED / OPERATOR-RATIFIED / PROMOTED
T8-A                                     ACTIVE / CURRENT TECHNICAL CENSUS + REMEASUREMENT NEXT
T8-B→T12                                 NOT OPEN
implementation                           BLOCKED
```

## T7 closure

```text
NO HISTORICAL BUSINESS MIGRATION REQUIRED FOR LAUNCH
```

Current MetalDocs DB/content/history is DEV/test/throwaway and is not a business-history migration source. No historical-data compatibility consumer constrains T8. T10 still owns technical current→target transition and DEV/test-state disposal/reset.

## Active T8-A staging

`../../docs/superpowers/analysis/2026-08-19-r10-t8a-technical-authority-legacy-census-bootstrap.md`

T8-A classifies current technical structures as:

```text
PRESERVE
REFINE
REHOME
REWRITE
DELETE
CURRENT-STATE ONLY
SUPERSEDED
```

It remeasures load-bearing current evidence across backend/package topology, DB/SQL ownership, API/codegen, frontend, runtime/deploy, verification and technical documentation. It does not yet freeze the replacement topology.

## Corrected post-T6 descent

```text
T7  Historical Migration Truth & Semantic Mapping       CLOSED
T8  Technical Realization Architecture                   ACTIVE
  A Technical Authority & Legacy Census                 ACTIVE
  B Backend Module & Package Topology                    NOT OPEN
  C Internal Communication Contracts                     NOT OPEN
  D Persistence Realization                              NOT OPEN
  E Executable Wire Contract                             NOT OPEN
  F Frontend Realization                                 NOT OPEN
  G Runtime / Process / Deployment                       NOT OPEN
  H Whole-T8 Global Coherence Review                     NOT OPEN
T9  Golden Flows & Validation Baseline                   NOT OPEN
T10 Transition / Refactor / Migration / Cutover          NOT OPEN
T11 Implementation Program & Execution Graph             NOT OPEN
T12 Adversarial Implementation-Readiness Review          NOT OPEN
```

## Current-state technical references

Until T8 explicitly promotes/replaces them, these are **evidence about the existing implementation, not target R10 authority**:

```text
backend-blueprint.md
backend-api-structure.md
frontend-structure.md
data-model.md
wiki/backend/repo-topology.md
current module pages
current code/schema/OpenAPI/deploy/tests
```

`backend-target-architecture.md` and `cohesive-platform-redesign.md` are historical/superseded target artifacts and must not route new target decisions.

No implementation plan or product code is authorized.
