# R10 Technical Architecture — Active Stage Router

> **Status:** ACTIVE — **T1→T7 CLOSED / OPERATOR-RATIFIED; T8-A ACTIVE / DISPOSITION CANDIDATE / OPERATOR RATIFICATION NEXT; T8-B→T12 NOT OPEN; IMPLEMENTATION BLOCKED**  
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
10. active T8-A staging listed in §5
11. current code/schema/API/frontend/runtime/deploy/test evidence for a concrete T8-A claim

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

T8-A Global Maximum posture:

```text
existing implementation = evidence only
sunk cost / test count / migration convenience = no survival entitlement
PRESERVE must be proved
REWRITE / REHOME / DELETE are valid outcomes
full-greenfield reset is also not justified without evidence
```

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
  T8-A Technical Authority & Legacy Census       ACTIVE / DISPOSITION CANDIDATE / OPERATOR RATIFICATION NEXT
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

## 5. T8-A — DISPOSITION CANDIDATE

Active T8-A staging, read in order:

1. `docs/superpowers/analysis/2026-08-19-r10-t8a-technical-authority-legacy-census-bootstrap.md`
2. `docs/superpowers/analysis/2026-08-19-r10-t8a-technical-census-disposition-matrix.md`
3. `docs/superpowers/analysis/2026-08-19-r10-t8a-technical-document-authority-reconciliation.md`
4. `docs/superpowers/analysis/2026-08-19-r10-t8a-technical-authority-legacy-disposition-candidate.md`
5. `docs/superpowers/analysis/2026-08-19-r10-t8a-candidate-adversarial-challenge.md`
6. `docs/superpowers/analysis/2026-08-19-r10-t8a-platform-facing-summary.md` — **operator ratification target**

All are staging/non-authoritative until explicit operator ratification + durable promotion.

T8-A dispositions current structures using:

```text
PRESERVE
REFINE
REHOME
REWRITE
DELETE
CURRENT-STATE ONLY
SUPERSEDED
```

Evidence continues to use:

```text
CURRENT-PROVEN
LAST-REPRODUCED
STALE / SUPERSEDED
UNKNOWN / REMEASURE
```

### Candidate Global Maximum

```text
CLEAN-SLATE PHYSICAL TARGET FREEDOM
+ SELECTIVE PROOF-BACKED MECHANISM REUSE
- LEGACY SHAPE INHERITANCE
- FULL-GREENFIELD PURITY RESET WITHOUT EVIDENCE
```

High-confidence candidate consequences:

```text
legacy semantic package/module topology          REWRITE / REHOME
current persistent model / table families        REWRITE; many DELETE candidates
current tenant/GUC/RLS mesh                      REWRITE / DELETE current mechanism
current OpenAPI surface                          REWRITE
local-password authentication capability         DELETE / REWRITE AuthN realization
current frontend feature/route topology          REWRITE / REHOME
provider/storage-key semantic references         DELETE / REWRITE storage contract
current jobs/process/provider topology            CURRENT-STATE ONLY / rederive
non-Launch capability implementation             DELETE / DEFER absent named Launch consumer

PostgreSQL product-state substrate               PRESERVE
River durable-job mechanism                      PRESERVE
contract-first + generated Go/TS boundaries      PRESERVE
verification registry / local-CI SSOT model      PRESERVE
runtime DB identity != schema/DDL owner property PRESERVE
deterministic DB bootstrap/proof property        PRESERVE / REFINE
```

Selective reuse is allowed only if all are proven:

```text
named current R10 consumer
+ public contract free of legacy semantic authority
+ dependency direction fits target
+ proof asserts target property rather than legacy shape
+ reuse remains smaller than rewrite after transition cost
```

Exact old Aug-09 counts are remeasured only when load-bearing. Current foreign-SQL leakage is qualitatively proven; exact historical `55/12` reproduction is not required to choose the T8-A disposition strategy.

### Explicitly still undecided

T8-A does **not** choose:

```text
exact target Go packages/module count
exact owner interfaces/dependency graph
exact target tables/constraints/RLS posture
exact OpenAPI operations/schemas
frontend libraries/folder/query/cache realization
interactive DOCX / renderer provider
number of runtime processes/containers
Redis survival
exact deployment topology
current→target transition sequence
```

Those belong to T8-B→T8-G/T10.

### Exact next action

```text
operator reviews/ratifies or rejects the T8-A platform-facing summary
→ if ratified:
   promote T8-A durable authority to wiki/
   → reconcile Decision Registry
   → repair stale technical-document / ADR routing labels
   → remove completed T8-A staging
   → mark T8-A CLOSED
   → only then open T8-B Backend Module & Package Topology
```

No T8-B design or product implementation is authorized before that gate.

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
