# Architecture

> **Last verified:** 2026-08-19  
> **Scope:** Durable architecture truth and active target-design routing.

## Active authority — read first

- `launch-v1-product-contract.md` — Product Contract REV001.
- `whole-product-alignment-review.md` — Whole-Product GCR A1–A10.
- `launch-v1-ownership-topology.md` — operator-approved 4+1 semantic ownership.
- `r10-t1-semantic-state-invariants.md` — T1 operator-ratified.
- `r10-t2-governance-effectivity-transactions.md` — T2 operator-ratified.
- `r10-t3-authorization-audit-enforcement.md` + D4 amendment — T3 operator-ratified.
- `r10-t4-exact-content-storage-integrity-restore.md` — T4 operator-ratified.
- `r10-t5-durable-async-search-external-effects.md` — T5 operator-ratified.
- `r10-t6-canonical-api-frontend-journeys.md` — T6 operator-ratified.
- `r10-t7-historical-migration-truth-semantic-mapping.md` — T7 operator-ratified.
- `r10-t8a-technical-authority-legacy-disposition.md` — **T8-A operator-ratified technical-inheritance/disposition authority.**
- `rebaseline-decision-registry.md` + D4/T6/post-T6/T7/T8-A amendments — current Registry chain.
- `r10-post-t6-implementation-readiness-program.md` — operator-ratified T7→T12 program authority.
- `r10-technical-realization-reconciliation-baseline.md` — operator-ratified technical census/routing baseline.
- `r10-technical-architecture.md` — **sole current stage/status/next-action router.**
- `../references/current-agent-handoff.md` — fresh-session recovery point.

## Current gate

```text
Product Contract / GCR / ownership       APPROVED
T1→T8-A                                  CLOSED / OPERATOR-RATIFIED
Decision Registry                        CURRENT + D4 + T6 + post-T6 + T7 + T8-A amendments
TRRB                                     CLOSED / OPERATOR-RATIFIED / PROMOTED
T8-B                                     ACTIVE / BACKEND MODULE & PACKAGE TOPOLOGY
T8-C→T12                                 NOT OPEN
implementation                           BLOCKED
```

## T8-A durable ruling

```text
CLEAN-SLATE PHYSICAL TARGET FREEDOM
+ SELECTIVE PROOF-BACKED MECHANISM REUSE
- LEGACY SHAPE INHERITANCE
- FULL-GREENFIELD PURITY RESET WITHOUT EVIDENCE
```

Current implementation is evidence only. Legacy package/table/API/frontend/runtime shapes receive no survival entitlement from existence, sunk cost, test volume, migration convenience or historical ADR status.

## Current T8-B stage

T8-B derives the **target backend module/package topology** from ratified semantic ownership and invariants.

It must answer only questions such as:

```text
what backend package/module boundaries exist
which package owns which semantic authority
allowed dependency direction between target packages
where composition roots live
which concerns are domain vs supporting mechanism
which legacy packages are structurally reusable vs replaced
```

It must not silently decide:

```text
exact inter-owner communication contracts        → T8-C
exact target tables/constraints                  → T8-D
exact OpenAPI operations/schemas                 → T8-E
frontend realization                             → T8-F
process/deployment topology                      → T8-G
current→target transition                        → T10
```

## Current-state technical references

These remain **evidence about the existing implementation, not target R10 authority**:

```text
backend-blueprint.md
backend-api-structure.md
frontend-structure.md
data-model.md
wiki/backend/repo-topology.md
wiki/modules/*
current code/schema/OpenAPI/deploy/tests
```

`backend-target-architecture.md` and `cohesive-platform-redesign.md` are superseded historical target artifacts.

## Remaining descent

```text
T8-B Backend Module & Package Topology            ACTIVE
T8-C Internal Communication Contracts             NOT OPEN
T8-D Persistence Realization                      NOT OPEN
T8-E Executable Wire Contract                     NOT OPEN
T8-F Frontend Realization                         NOT OPEN
T8-G Runtime / Process / Deployment               NOT OPEN
T8-H Whole-T8 Global Coherence Review             NOT OPEN
T9   Golden Flows & Validation Baseline           NOT OPEN
T10  Transition / Refactor / Migration / Cutover  NOT OPEN
T11  Implementation Program & Execution Graph     NOT OPEN
T12  Adversarial Implementation-Readiness         NOT OPEN
```

No implementation plan or product code is authorized.