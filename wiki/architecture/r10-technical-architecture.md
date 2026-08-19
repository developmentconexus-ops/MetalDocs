# R10 Technical Architecture — Active Stage Router

> **Status:** ACTIVE — **T1→T7 CLOSED / OPERATOR-RATIFIED; T8-A ACTIVE / CURRENT TECHNICAL CENSUS + REMEASUREMENT NEXT; T8-B→T12 NOT OPEN; IMPLEMENTATION BLOCKED**  
> **Rebaselined:** 2026-08-19  
> **Repository:** `developmentconexus-ops/MetalDocs`  
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131

This file is the sole R10 stage/status/next-action router. Detailed meaning lives in the durable authorities it routes to.

## 1. Binding authority chain

Read in order:

1. `AGENTS.md`
2. `docs/engineering/standards/root-cause-global-maximum-method.md`
3. `wiki/references/current-agent-handoff.md`
4. Product Contract REV001 + Whole-Product GCR + 4+1 ownership
5. T1→T7 durable R10 authorities
6. Decision Registry + D4/T6/post-T6/T7 amendments
7. `wiki/architecture/r10-post-t6-implementation-readiness-program.md`
8. `wiki/architecture/r10-technical-realization-reconciliation-baseline.md`
9. this router
10. `docs/superpowers/analysis/2026-08-19-r10-t8a-technical-authority-legacy-census-bootstrap.md`
11. current code/schema/API/frontend/runtime/deploy/test evidence for each concrete T8-A claim

Legacy implementation proves what exists, not what survives.

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
T7 — Historical Migration Truth & Mapping        CLOSED / OPERATOR-RATIFIED / PROMOTED
Decision Registry                                CURRENT + D4 + T6 + post-T6 + T7 amendments
Post-T6 Stage-Decomposition GCR                  RESTRUCTURE NOW / OPERATOR-RATIFIED
Technical Realization Reconciliation Baseline   CLOSED / OPERATOR-RATIFIED / PROMOTED

T8 — Technical Realization Architecture          ACTIVE
  T8-A Technical Authority & Legacy Census       ACTIVE / CURRENT TECHNICAL CENSUS + REMEASUREMENT NEXT
  T8-B Backend Module & Package Topology         NOT OPEN
  T8-C Internal Communication Contracts          NOT OPEN
  T8-D Persistence Realization                   NOT OPEN
  T8-E Executable Wire Contract                  NOT OPEN
  T8-F Frontend Realization                      NOT OPEN
  T8-G Runtime / Process / Deployment            NOT OPEN
  T8-H Whole-T8 Global Coherence Review          NOT OPEN

T9 — Golden Flows & Validation Baseline          NOT OPEN
T10 — Transition / Refactor / Migration/Cutover  NOT OPEN
T11 — Implementation Program & Execution Graph   NOT OPEN
T12 — Adversarial Implementation-Readiness       NOT OPEN

implementation                                    BLOCKED
```

## 4. T7 closure

Durable T7 authority:

`wiki/architecture/r10-t7-historical-migration-truth-semantic-mapping.md`

Registry reconciliation:

`wiki/architecture/rebaseline-decision-registry-t7-amendment.md`

Ratified decision:

```text
NO HISTORICAL BUSINESS MIGRATION REQUIRED FOR LAUNCH
```

Binding consequences:

```text
current MetalDocs business history = NONE
current DB/content/history = DEV / TEST / THROWAWAY
no historical-data compatibility consumer exists for T8
R10 business history begins natively at/after cutover
T10 remains mandatory for technical current→target transition
```

Completed T7 staging was removed from the live tree. Git history is provenance.

## 5. T8-A — ACTIVE

Active bootstrap:

`docs/superpowers/analysis/2026-08-19-r10-t8a-technical-authority-legacy-census-bootstrap.md`

T8-A classifies current technical structures using:

```text
PRESERVE
REFINE
REHOME
REWRITE
DELETE
CURRENT-STATE ONLY
SUPERSEDED
```

Evidence continues to use the TRRB classes:

```text
CURRENT-PROVEN
LAST-REPRODUCED
STALE / SUPERSEDED
UNKNOWN / REMEASURE
```

Old Aug-09 mechanical counts are not current facts until remeasured when load-bearing.

### T8-A required surfaces

```text
backend modules/packages/import graph/composition
SQL/table/view/function/trigger ownership and foreign access
OpenAPI/codegen/runtime contract mechanisms
frontend routes/features/query/cache/state/editor/viewer boundaries
async/jobs/rendering
binaries/processes/deploy/config/trust/observability/recovery
verification/tests/CI/tools/verify/architecture guards
technical-document and ADR authority status
```

### Exact next action

```text
fresh current technical census
→ remeasure load-bearing legacy metrics against current source
→ map each material structure to a ratified target property, if any
→ preliminary PRESERVE / REFINE / REHOME / REWRITE / DELETE / CURRENT-STATE ONLY / SUPERSEDED disposition
→ identify only material unknowns/disagreements
→ T8-A disposition candidate + operator adjudication
```

T8-A may classify `REWRITE` without selecting the replacement topology. Final package/database/API/frontend/runtime designs belong T8-B→T8-G.

## 6. Future stage boundaries

```text
T8-B = target backend/package topology
T8-C = target internal owner communication contracts
T8-D = target persistence realization
T8-E = exact executable OpenAPI/wire contract
T8-F = target frontend realization
T8-G = target runtime/process/deployment realization
T8-H = Whole-T8 Global Coherence Review
T9   = Golden Flows + falsifiable Validation Baseline
T10  = current→target technical transition/cutover/rollback
T11  = bounded implementation Execution Graph; no hidden architecture decisions
T12  = fresh adversarial implementation-readiness challenge
```

## 7. Final implementation gate

Implementation remains blocked until:

```text
T8→T12 CLOSED / OPERATOR-RATIFIED
→ Integrated Whole-R10 Global Coherence Review PASS
→ fresh independent/cold review converged
→ operator explicitly authorizes implementation
```

Existing runtime safety controls remain binding until deliberately replaced by accepted target realization.
