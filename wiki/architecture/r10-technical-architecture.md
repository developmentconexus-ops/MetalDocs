# R10 Technical Architecture — Active Stage Router

> **Status:** ACTIVE — **T1→T6 CLOSED / OPERATOR-RATIFIED; POST-T6 PROGRAM RESTRUCTURED; TRRB OPERATOR REVIEW NEXT; T7→T12 NOT OPEN; IMPLEMENTATION BLOCKED**  
> **Rebaselined:** 2026-08-19  
> **Repository:** `developmentconexus-ops/MetalDocs`  
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131

This file is the sole R10 stage/status/next-action router. Detailed meaning lives in the durable authorities it routes to.

## 1. Binding authority chain

Read in order:

1. `AGENTS.md`
2. `docs/engineering/standards/root-cause-global-maximum-method.md`
3. `wiki/references/current-agent-handoff.md`
4. `wiki/architecture/launch-v1-product-contract.md` — REV001
5. `wiki/architecture/whole-product-alignment-review.md`
6. `wiki/architecture/launch-v1-ownership-topology.md`
7. `wiki/architecture/r10-t1-semantic-state-invariants.md`
8. `wiki/architecture/r10-t2-governance-effectivity-transactions.md`
9. `wiki/architecture/r10-t3-authorization-audit-enforcement.md`
10. `wiki/architecture/r10-t3-d4-responsible-owner-eligibility-amendment.md`
11. `wiki/architecture/r10-t4-exact-content-storage-integrity-restore.md`
12. `wiki/architecture/r10-t5-durable-async-search-external-effects.md`
13. `wiki/architecture/r10-t6-canonical-api-frontend-journeys.md`
14. `wiki/architecture/rebaseline-decision-registry.md`
15. `wiki/architecture/rebaseline-decision-registry-d4-amendment.md`
16. `wiki/architecture/rebaseline-decision-registry-t6-amendment.md`
17. `wiki/architecture/r10-post-t6-implementation-readiness-program.md`
18. this router
19. `docs/superpowers/analysis/2026-08-19-r10-technical-realization-reconciliation-baseline.md`
20. current code/schema/API/frontend/runtime evidence only for a concrete census/design claim

Legacy implementation and legacy technical design documents are evidence only unless the active R10 authority explicitly promotes their meaning.

## 2. Binding Method laws

```text
smallest sustainable solution
one semantic authority per meaning
mechanism != authority
proof before implementation
Structural Inversion
unknown remains unknown
revalidation does not mean reinvention
prepare the seam, not dormant future capability
```

Program-specific implementation law:

> **No Writer task may contain a material architecture decision that should have been decided before execution.**

## 3. Current descent

```text
Product Contract REV001                          CLOSED / OPERATOR-APPROVED
Whole-Product GCR A1→A10                         CLOSED / OPERATOR-APPROVED
Launch ownership topology                        CLOSED / OPERATOR-APPROVED / 4+1
T1 — Semantic State & Invariants                 CLOSED / OPERATOR-RATIFIED
T2 — Governance / Effectivity / Tx               CLOSED / OPERATOR-RATIFIED
T3 — Authorization & Audit                       CLOSED / OPERATOR-RATIFIED + D4 amendment
T4 — Exact Content / Storage / Restore           CLOSED / OPERATOR-RATIFIED
T5 — Durable Async / Search / Effects            CLOSED / OPERATOR-RATIFIED
T6 — Canonical API / Frontend Journeys           CLOSED / OPERATOR-RATIFIED / PROMOTED
Decision Registry                                CURRENT + D4 + T6 amendments

Post-T6 Stage-Decomposition GCR                  RESTRUCTURE NOW / OPERATOR-RATIFIED
Technical Realization Reconciliation Baseline   STAGED / OPERATOR REVIEW NEXT

T7 — Historical Migration Truth & Mapping        NOT OPEN
T8 — Technical Realization Architecture          NOT OPEN
T9 — Golden Flows & Validation Baseline          NOT OPEN
T10 — Transition / Refactor / Migration/Cutover  NOT OPEN
T11 — Implementation Program & Execution Graph   NOT OPEN
T12 — Adversarial Implementation-Readiness       NOT OPEN

implementation                                    BLOCKED
```

## 4. Post-T6 Global Coherence finding

The previously accepted sequence from T7 directly to final review/implementation planning was challenged after T6 closure.

Finding:

```text
semantic architecture T1→T6 = coherent
physical technical realization = not yet completely designed
old post-T6 sequence could delegate material realization choices to implementation planning
```

Method outcome:

```text
RESTRUCTURE NOW
```

Operator ratified the corrected program on 2026-08-19.

Durable program authority:

`wiki/architecture/r10-post-t6-implementation-readiness-program.md`

T1→T6 remain preserved. The restructure changes the descent after T6, not their accepted meaning.

## 5. Current gate — TRRB

Current staging/evidence target:

`docs/superpowers/analysis/2026-08-19-r10-technical-realization-reconciliation-baseline.md`

The TRRB is a **census**, not a target architecture. It classifies current evidence as:

```text
CURRENT-PROVEN
LAST-REPRODUCED
STALE / SUPERSEDED
UNKNOWN / REMEASURE
```

It covers:

```text
repository/runtime topology
backend modules/packages/import graph
persistence/table/SQL ownership
API/OpenAPI/codegen
frontend topology
async/jobs/rendering
runtime/deploy/operations
verification/tests/CI
technical-document authority drift
future-stage decision ownership
```

Exact next action:

```text
operator reviews/adjudicates TRRB coverage + evidence classifications
→ correct any material census gap/misclassification
→ complete technical-document authority reconciliation
→ only then open redefined T7
```

Do **not** perform substantive T7 design before this gate closes.

## 6. Future stage boundaries

```text
T7  = source truth + semantic historical mapping only
T8  = backend/package/internal contracts/persistence/wire/frontend/runtime realization
T9  = Golden Flows + falsifiable Validation Baseline
T10 = current→target refactor/data/API/frontend/runtime migration + cutover/rollback
T11 = bounded implementation Execution Graph; no hidden architecture decisions
T12 = fresh adversarial implementation-readiness challenge
```

Detailed subgates and reopen laws belong to the post-T6 program authority.

## 7. Final implementation gate

Implementation remains blocked until:

```text
T7→T12 CLOSED / OPERATOR-RATIFIED
→ Integrated Whole-R10 Global Coherence Review PASS
→ fresh independent/cold review converged
→ operator explicitly authorizes implementation
```

Existing runtime safety controls remain binding until deliberately replaced by accepted target realization.
