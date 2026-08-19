# R10 Technical Architecture — Active Stage Router

> **Status:** ACTIVE — **T1→T6 CLOSED / OPERATOR-RATIFIED; POST-T6 PROGRAM RESTRUCTURED; TRRB OPERATOR-RATIFIED; T7 ACTIVE / BUSINESS SOURCE CORPUS IDENTIFICATION NEXT; T8→T12 NOT OPEN; IMPLEMENTATION BLOCKED**  
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
7. T1→T6 durable R10 authorities
8. Decision Registry + D4/T6/post-T6 amendments
9. `wiki/architecture/r10-post-t6-implementation-readiness-program.md`
10. `wiki/architecture/r10-technical-realization-reconciliation-baseline.md`
11. this router
12. `docs/superpowers/analysis/2026-08-19-r10-t7-source-corpus-operator-clarification.md`
13. `docs/superpowers/analysis/2026-08-19-r10-t7-historical-migration-truth-semantic-mapping-bootstrap.md`
14. actual business-source evidence only after the source is identified

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
Decision Registry                                CURRENT + D4 + T6 + post-T6 amendments

Post-T6 Stage-Decomposition GCR                  RESTRUCTURE NOW / OPERATOR-RATIFIED
Technical Realization Reconciliation Baseline   CLOSED / OPERATOR-RATIFIED / PROMOTED

T7 — Historical Migration Truth & Mapping        ACTIVE / BUSINESS SOURCE CORPUS IDENTIFICATION NEXT
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

Operator ratified the corrected program on 2026-08-19. T1→T6 remain preserved.

Durable program authority:

`wiki/architecture/r10-post-t6-implementation-readiness-program.md`

Registry stage-ownership reconciliation:

`wiki/architecture/rebaseline-decision-registry-post-t6-amendment.md`

## 5. TRRB closure

Durable TRRB authority:

`wiki/architecture/r10-technical-realization-reconciliation-baseline.md`

The operator ratified the census/coverage/routing baseline. It does **not** choose target packages, tables, wire schemas, frontend topology, runtime topology, transition implementation or execution sequence.

Ratified evidence classes remain:

```text
CURRENT-PROVEN
LAST-REPRODUCED
STALE / SUPERSEDED
UNKNOWN / REMEASURE
```

Old exact audit metrics remain `LAST-REPRODUCED` until a later stage remeasures them when load-bearing.

## 6. T7 — ACTIVE

Binding source-corpus clarification:

`docs/superpowers/analysis/2026-08-19-r10-t7-source-corpus-operator-clarification.md`

Active bootstrap:

`docs/superpowers/analysis/2026-08-19-r10-t7-historical-migration-truth-semantic-mapping-bootstrap.md`

The operator explicitly established:

```text
current MetalDocs DB/content/history = DEV / TEST / THROWAWAY
current MetalDocs business history   = NONE
current MetalDocs product data       = NOT A HISTORICAL MIGRATION SOURCE
```

Current MetalDocs schema/code/data remain technical legacy evidence for T8/T10 only.

Current T7 source state:

```text
CURRENT METALDOCS AS BUSINESS SOURCE = EXCLUDED
ACTUAL BUSINESS SOURCE CORPUS        = NOT IDENTIFIED
HISTORICAL MIGRATION REQUIRED?       = NOT YET DECIDED
```

T7 owns only:

```text
whether Launch requires any pre-existing business corpus
actual business source/location/authority if one exists
PROVEN / INFERABLE WITH EXPLICIT RULE / UNKNOWN
smallest justified migration-mode set when applicable
imported target-owned facts vs provenance-only evidence
source document/revision/content/actor/governance truth
semantic migration unit
truthful handling of partial/ambiguous/unknown history
```

T7 explicitly does **not** own backend/package realization, target relational schema, exact executable OpenAPI, frontend realization, runtime/deploy realization, concrete migration tooling, production cutover, rollback, deletion or restore choreography. Those belong to T8/T10.

### Exact next action

```text
business-corpus necessity gate
→ if no pre-existing business documents are required: derive bounded no-historical-migration T7 candidate
→ if they are required: identify the actual source corpus/location/authority
→ inspect that source directly
→ classify material source facts PROVEN / INFERABLE / UNKNOWN
→ only then compare migration-truth approaches
```

Do not inspect current MetalDocs DEV rows as business migration evidence.

## 7. Future stage boundaries

```text
T8  = backend/package/internal contracts/persistence/wire/frontend/runtime realization
T9  = Golden Flows + falsifiable Validation Baseline
T10 = current→target technical refactor/data/API/frontend/runtime migration + cutover/rollback
T11 = bounded implementation Execution Graph; no hidden architecture decisions
T12 = fresh adversarial implementation-readiness challenge
```

## 8. Final implementation gate

Implementation remains blocked until:

```text
T7→T12 CLOSED / OPERATOR-RATIFIED
→ Integrated Whole-R10 Global Coherence Review PASS
→ fresh independent/cold review converged
→ operator explicitly authorizes implementation
```

Existing runtime safety controls remain binding until deliberately replaced by accepted target realization.
