---
id: repository-roadmap
kind: authority
owner: architecture
summary: Sole mutable MetalDocs stage, gate, implementation-status, and next-action authority.
---

# MetalDocs roadmap

## Current state

```text
REPOSITORY MODE       CLEAN-SLATE / ARCHITECTURE-FIRST
T1 → T10              CLOSED / OPERATOR-RATIFIED / INTEGRATED
T11                   OPEN / ACTIVE
T11 checkpoint        B01-B10 ACCEPTED / INTEGRATED
T11 current block     B11 ACCESS ADMINISTRATION / LOCKED / P8-P10 COMPLETE
METHODOLOGY ADOPTION  ACCEPTED / INTEGRATED
T12                   NOT OPEN
IMPLEMENTATION         BLOCKED
```

Current Product/architecture authority lives in `docs/product/**`, `docs/architecture/**`, and `docs/decisions/**`.

Current system census:

```text
semantic owners                  4 business + 2 supporting
stable SPA routes                11
PermissionCode values            16
application operations           89
Idempotency-Key creations        11
ETag read / mutation domains     13 / 13
exact-byte resources             4
```

`docs/decisions/api-operation-census.md` owns the current application-operation / Idempotency / ETag / exact-byte census.

## Organizational methodology baseline

The exact accepted organizational methodology pin and method-selection route are owned by `AGENTS.md`. Method selection starts at that pin's `ROUTER.md`.

The reusable local Frontend Product Experience Planning Method v2.3 was superseded by the central `FRONTEND-METHOD.md`, whose lineage explicitly consolidates the MetalDocs v2.3 generation. MetalDocs-specific P8 rendered-Evidence handling remains local in `docs/development/engineering-rules.md`.

Adoption impact sweep at integration time:

```text
B01-B09 protected structure / Screen Contracts  UNAFFECTED
Product/backend authority                       UNAFFECTED
89-operation / 11-route census                  UNAFFECTED
exact P8 LOCK Evidence                           UNAFFECTED
FP2 / P11                                        NOT OPEN
B10-B12                                          NOT OPEN at methodology-adoption integration
```

The independent methodology-adoption review converged with `MATERIAL=0` and `IMPORTANT=0`. Methodology adoption did not reopen Product/architecture/frontend LOCKs and did not authorize implementation.

Aggregate verification enforces the concrete bootstrap properties it owns, including the exact AGENTS methodology pin/ROUTER presence, context budget, temporary-work hygiene, known local-method removal, active-router de-reference, and exact Evidence-ref protections. Broader semantic duplicate-authority defects remain hard stops/review concerns rather than falsely claimed as exhaustively grep-detectable.

## Frontend Product Experience Program

```text
FP0  Frontend Foundation                         CLOSED / 89 operations / 11 routes REBASELINED
FP1  Block-by-block Product Experience           ACTIVE
FP2  Integrated Low-Fidelity Product / P11       NOT OPEN
FP3  Whole-Product Adversarial Review / P12      NOT OPEN
FP4  Visual Handoff + Readiness / P13-P14        NOT OPEN
```

## FP1 block status

```text
B01   App Shell + Global IA + Home                 LOCKED / OPERATOR-RATIFIED
B01N  Notification global chrome + Quick Inbox     LOCKED / OPERATOR-RATIFIED
B02   Library / Discovery                          LOCKED / OPERATOR-RATIFIED
B03   Document Official                            LOCKED / P8-P10 COMPLETE
B04   Document Work                                LOCKED / P8-P10 COMPLETE
B05   My Work                                      LOCKED / P8 R2-P10 COMPLETE
B06   Governance Case                              LOCKED / P8-P10 COMPLETE
       B06-F1 deadline projection                  CLOSED / OPERATOR-RATIFIED
       B06-F2 Governance Review Layer seam         CLOSED / FUTURE-SEAM
B07   Document History                             LOCKED / P8-P10 COMPLETE
       B07-F1 recognition read                     CLOSED / OPERATOR-RATIFIED
B08   Notifications Full Inbox                     LOCKED / P8-P10 COMPLETE
       B08-F1 recognition read                     CLOSED / OPERATOR-RATIFIED
B09   Audit                                        LOCKED / P8-P10 COMPLETE
       B09-F1 Audit query/evidence capability      CLOSED / OPERATOR-RATIFIED
       op78 + op87-op89 package                    OPERATOR-RATIFIED / DURABLE
       unresolved BLOCKING / IMPORTANT             0 / 0
B10   Organization Administration                  LOCKED / P8-P10 COMPLETE / OPERATOR-RATIFIED / INTEGRATED
       locked P8 blob                              1d1cc7d5cb42e034ab9ee71a21c96918cdcf691d
       structural browser verification             32 / 32 PASS
       P9 material regions/controls                 34 / 34 TRACED
       P9 accepted B10 operations                   24 / 24 BOUND (ops 3-26)
       operation 27+ consumed                       0
       B10-A1 paginated-browse sufficiency          VALIDATED FOR CURRENT LAUNCH P8
       unresolved material B10 Findings             0
B11   Access Administration                        LOCKED / P8-P10 COMPLETE / OPERATOR-RATIFIED
       durable Evidence locator                     docs/decisions/t11-b11-lock-evidence.md
       Evidence ref                                evidence/t11-b11-locks-20260825
       Evidence exact commit                       469a753904041e7800400dc1074510456aa50df8
       Evidence exact tree                         c4f04b75c3676dcde00caa07279824b3c653c7f3
       locked P8 Git blob                          96094773435a88c357e308779639415d9853b327
       B11-F1 Access Assignment Read Precision      CLOSED / OPERATOR-RATIFIED / DURABLE
       B11-F1 authority                             docs/decisions/access-assignment-read.md
       op31 listRoleAssignments                     REFINED / FILTERED + HUMAN-RECOGNIZABLE READ
       application-operation delta                  +0 / CENSUS REMAINS 89
       P9 material regions/controls                 36 / 36 TRACED
       P9 primary B11 operations                    7 / 7 BOUND (ops 27-33)
       P9 supporting Organization reads             ops 6 / 16 / 22
       operation 90+ consumed                       0
       P10 existing shared behaviors reused         4
       P10 new shared semantic patterns             0
       P10 B11-local protected patterns             9
       B11-R2-A Group footprint comprehension       VALIDATED FOR CURRENT LAUNCH P8
       B11-R2-B Area configuration comprehension    VALIDATED FOR CURRENT LAUNCH P8
       B11-R2-C membership consequence              VALIDATED FOR CURRENT LAUNCH P8
       B11-R2-D no per-User troubleshooter          SUFFICIENT / VALIDATED FOR CURRENT LAUNCH P8
       B11-R2-E filtered pagination sufficiency     VALIDATED FOR CURRENT LAUNCH P8
       unresolved material B11 Findings             0
B12   Document Governance Administration           NOT OPEN
```

Exact locked frontend Evidence remains recoverable through:

```text
B01-B09
docs/decisions/t11-b01-b09-lock-evidence.md
evidence/t11-pr162-b01-b09-locks-20260824
→ adf58e448bc5bd3a20cae5b7228d729c031f94ac

B10
docs/decisions/t11-b10-lock-evidence.md
evidence/t11-pr170-b10-locks-20260824
→ b8c607cbd30d61d6bcf6ec1ea734ed1653d2569e

B11
docs/decisions/t11-b11-lock-evidence.md
evidence/t11-b11-locks-20260825
→ 469a753904041e7800400dc1074510456aa50df8
```

All Evidence refs are non-authoritative provenance and must remain reachable while current T11/P11/P13/P14 proof depends on them.

## Exact next action

```text
1. Ensure the B11 merge candidate contains no docs/work/** after exact Evidence preservation.
2. Revalidate the cleaned B11 acceptance candidate against AGENTS.md, docs/index.md, engineering rules, B11-F1, the 89-operation census and repository conformance checks.
3. Prepare one B11 acceptance-increment PR against main; Draft is the default while final review/conformance is incomplete.
4. Use independent adversarial review if required by the pinned Method/repository governance for the final acceptance claim; reviewer output remains Evidence, never authority.
5. Do not open B12 merely because B11 is LOCKED. B12 requires an explicit later opening/acceptance increment after B11 integration.
6. FP2/P11 remains NOT OPEN until the block program reaches the accepted assembly boundary.
7. T12 and Product implementation remain BLOCKED.
```

## Hard stops

```text
no Product code/schema/OpenAPI/runtime/deploy implementation
no T12 work
no B12 work
no FP2/P11 work
no B01-B10 reopen without material Evidence
no B11 reopen without material operator/Product/integration Evidence
no application operation 90+ for B11 without a new lawful bounded reopen
no global User/Group/Area search invented for UI convenience
no client-side crawl/post-filter over incomplete pages presented as complete
no frontend effective-permission authority or inferred inherited-access matrix
no Group single-Area ownership inferred from access scope
no custom Role/Permission editor
no unapproved shell/sidebar/header/local-nav topology redesign
no production-like visual design authority inferred from P8
no assistant/reviewer LOCK
no local reusable methodology fork
no floating methodology main as normative authority
no methodology sync bot/submodule/generated copies/framework
no merge without explicit operator authorization
```

## Implementation gate

Implementation remains blocked until:

```text
T11 CLOSED / OPERATOR-RATIFIED
T12 CLOSED / OPERATOR-RATIFIED
Integrated Whole-R10 coherence = PASS
fresh independent challenge = converged
operator implementation authorization = explicit
```

Accepted Product/R10/frontend LOCK decisions reopen only on material Evidence under the pinned DevelopmentConexus Engineering Method; methodology-migration ceremony alone is not a reopen trigger.
