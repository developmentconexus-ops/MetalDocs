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
- `rebaseline-decision-registry.md` — prior-decision disposition baseline.
- `rebaseline-decision-registry-d4-amendment.md` — D4 reconciliation.
- `rebaseline-decision-registry-t6-amendment.md` — T6 closure reconciliation.
- `rebaseline-decision-registry-post-t6-amendment.md` — **operator-ratified post-T6 stage-ownership reconciliation.**
- `r10-post-t6-implementation-readiness-program.md` — **active post-T6 program authority; T7→T12 decomposition.**
- `r10-technical-architecture.md` — sole current stage/status/next-action router.
- `../references/current-agent-handoff.md` — fresh-session recovery point.

## Current gate

```text
Product Contract / GCR / ownership       APPROVED
T1→T6                                    CLOSED / OPERATOR-RATIFIED
Decision Registry                        CURRENT + D4 + T6 + post-T6 amendments
Post-T6 Stage-Decomposition GCR          RESTRUCTURE NOW / OPERATOR-RATIFIED
TRRB                                     STAGED / OPERATOR REVIEW NEXT
T7→T12                                   NOT OPEN
implementation                           BLOCKED
```

## Current operator review target

`../../docs/superpowers/analysis/2026-08-19-r10-technical-realization-reconciliation-baseline.md`

The Technical Realization Reconciliation Baseline is a census, not target design. It classifies current evidence as:

```text
CURRENT-PROVEN
LAST-REPRODUCED
STALE / SUPERSEDED
UNKNOWN / REMEASURE
```

Its purpose is to prove that backend/package, persistence, API wire, frontend, runtime/deploy, proof and transition decisions have explicit future owners before implementation.

## Corrected post-T6 descent

```text
T7  Historical Migration Truth & Semantic Mapping
T8  Technical Realization Architecture
T9  Golden Flows & Validation Baseline
T10 Transition / Refactor / Migration / Cutover
T11 Implementation Program & Execution Graph
T12 Adversarial Implementation-Readiness Review
```

Then:

```text
Integrated Whole-R10 GCR
→ fresh independent/cold review
→ operator final ratification
→ explicit implementation authorization
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
