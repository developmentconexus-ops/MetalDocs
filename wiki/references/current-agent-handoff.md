# Current Agent Handoff

> **Last verified:** 2026-08-19  
> **Status:** ACTIVE — **T1→T7 CLOSED / OPERATOR-RATIFIED; T8-A ACTIVE / DISPOSITION CANDIDATE / OPERATOR RATIFICATION NEXT; T8-B→T12 NOT OPEN; IMPLEMENTATION BLOCKED**  
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131

## Fresh-session route

Read in this order:

1. `AGENTS.md`
2. `docs/engineering/standards/root-cause-global-maximum-method.md`
3. this file
4. Product Contract REV001 + Whole-Product GCR + 4+1 ownership
5. T1→T7 durable authorities
6. Decision Registry + D4/T6/post-T6/T7 amendments
7. `wiki/architecture/r10-post-t6-implementation-readiness-program.md`
8. `wiki/architecture/r10-technical-realization-reconciliation-baseline.md`
9. `wiki/architecture/r10-technical-architecture.md`
10. active T8-A staging listed below
11. current source/schema/OpenAPI/frontend/runtime/deploy/test evidence only for a concrete claim

Do not route target design through superseded/historical architecture or current package/module existence.

## Current checkpoint

```text
Product Contract                         REV001 / OPERATOR-APPROVED
Whole-Product GCR / 4+1 ownership       CLOSED / OPERATOR-APPROVED
T1→T7                                    CLOSED / OPERATOR-RATIFIED
Decision Registry                        CURRENT + D4 + T6 + post-T6 + T7 amendments
Post-T6 Stage-Decomposition GCR          RESTRUCTURE NOW / OPERATOR-RATIFIED
TRRB                                     CLOSED / OPERATOR-RATIFIED / PROMOTED

T8-A                                     ACTIVE / DISPOSITION CANDIDATE / OPERATOR RATIFICATION NEXT
T8-B                                     NOT OPEN
T8-C                                     NOT OPEN
T8-D                                     NOT OPEN
T8-E                                     NOT OPEN
T8-F                                     NOT OPEN
T8-G                                     NOT OPEN
T8-H                                     NOT OPEN
T9                                       NOT OPEN
T10                                      NOT OPEN
T11                                      NOT OPEN
T12                                      NOT OPEN
implementation                           BLOCKED
```

## Binding post-T6 execution law

> **No Writer task may contain a material architecture decision that should have been decided before execution.**

## Binding T8-A posture

The operator explicitly reaffirmed on 2026-08-19:

```text
refactor/rewrite from zero when needed
never protect an implemented local maximum
seek the Global Maximum from Method + accepted decisions + evidence
never assume current implementation is good
```

Therefore:

```text
current implementation = evidence only
PRESERVE must be proved
sunk cost / test count / migration convenience = no survival entitlement
```

## T7 — CLOSED

Durable authority:

`wiki/architecture/r10-t7-historical-migration-truth-semantic-mapping.md`

Registry amendment:

`wiki/architecture/rebaseline-decision-registry-t7-amendment.md`

Ratified:

```text
NO HISTORICAL BUSINESS MIGRATION REQUIRED FOR LAUNCH
```

Binding facts:

```text
current MetalDocs DB/content/history = DEV / TEST / THROWAWAY
current MetalDocs business history   = NONE
Launch pre-existing corpus import    = NO
```

Consequences include no DEV/test compatibility consumer for T8 and T10 ownership of technical reset/cutover.

## T8-A — CANDIDATE READY

Read active staging in order:

1. `docs/superpowers/analysis/2026-08-19-r10-t8a-technical-authority-legacy-census-bootstrap.md`
2. `docs/superpowers/analysis/2026-08-19-r10-t8a-technical-census-disposition-matrix.md`
3. `docs/superpowers/analysis/2026-08-19-r10-t8a-technical-document-authority-reconciliation.md`
4. `docs/superpowers/analysis/2026-08-19-r10-t8a-technical-authority-legacy-disposition-candidate.md`
5. `docs/superpowers/analysis/2026-08-19-r10-t8a-candidate-adversarial-challenge.md`
6. `docs/superpowers/analysis/2026-08-19-r10-t8a-platform-facing-summary.md` — **current operator ratification target**

All are non-authoritative until operator ratification + promotion.

### Candidate Global Maximum

```text
CLEAN-SLATE PHYSICAL TARGET FREEDOM
+ SELECTIVE PROOF-BACKED MECHANISM REUSE
- LEGACY SHAPE INHERITANCE
- FULL-GREENFIELD PURITY RESET WITHOUT EVIDENCE
```

Core candidate disposition:

```text
legacy semantic module topology             REWRITE / REHOME
current persistent schema/data model        REWRITE; many DELETE candidates
tenant/GUC/RLS mesh                         REWRITE / DELETE current mechanism
current OpenAPI surface                     REWRITE
local-password capability                   DELETE / AuthN REWRITE
frontend feature/route topology             REWRITE / REHOME
provider-key semantic storage contract      DELETE / REWRITE
old jobs/non-Launch capability wiring       DELETE / DEFER / REWRITE
legacy architecture-specific guard policy   REFINE / REWRITE
legacy technical docs                       CURRENT-STATE ONLY / SUPERSEDED routing

PostgreSQL                                  PRESERVE
River                                       PRESERVE
contract-first + generated Go/TS boundary  PRESERVE
verification registry/local-CI SSOT model  PRESERVE
runtime DB identity != DDL owner property  PRESERVE
deterministic DB bootstrap/proof property  PRESERVE / REFINE
```

Selective reuse is allowed only when all five are proven:

```text
named current R10 consumer
+ public contract contains no legacy semantic authority
+ dependency direction fits target
+ proof asserts target property, not legacy shape
+ reuse is smaller than rewrite after transition cost
```

### Important evidence result

Current foreign-SQL leakage is directly proven by current `tools/cilint/baseline.json`; exact reproduction of the old Aug-09 `55 reads / 12 writes` count is not load-bearing to T8-A's target-disposition choice. Remeasure exact magnitude later only when a material T10 transition/proof decision needs it.

### Still NOT decided

```text
target package/module count
target dependency graph / internal contracts
target tables/constraints/RLS
target exact OpenAPI operations/schemas
frontend libraries/query/cache/folder topology
editor/renderer provider
runtime process/container count
Redis survival
deployment topology
transition sequence
```

These remain T8-B→T8-G/T10.

## Exact next action

```text
operator reviews and ratifies/rejects
`docs/superpowers/analysis/2026-08-19-r10-t8a-platform-facing-summary.md`

if ratified:
→ promote T8-A durable authority
→ reconcile Decision Registry
→ repair stale technical-document / ADR routing labels
→ remove completed T8-A staging
→ mark T8-A CLOSED
→ only then open T8-B Backend Module & Package Topology
```

Do not begin T8-B or product implementation before this gate closes.

Implementation remains **BLOCKED**.
