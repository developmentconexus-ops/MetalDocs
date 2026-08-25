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
T11 acceptance        B11 EXPLICITLY OPENED IN THIS INCREMENT / BOUNDED P8-P9 REOPEN / R6 CANDIDATE
LOCAL METHODS         RESTORED / ENGINEERING v1.0.0 + FRONTEND v2.3
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

## Local methodology baseline

MetalDocs uses the repository-local shared method files:

```text
docs/development/engineering-method.md
  DevelopmentConexus Engineering Method v1.0.0

docs/development/frontend-product-experience-planning-method.md
  Frontend Product Experience Planning Method v2.3
```

Both files are the unchanged accepted methods also used by the other DevelopmentConexus product repositories. There is no external methodology router/pin in the active operating path.

PR #172 restored these local methods after B11 planning had already begun on its feature branch. The restoration changed operating mechanics only; it did not change Product, Authorization, wire, UX or existing frontend truth. B11 was therefore revalidated against the restored local methods before this acceptance candidate was reconciled with current `main`.

Current bounded impact sweep after the PR #173 review finding:

```text
B01-B10 protected structure / Screen Contracts  UNAFFECTED
B11 IA / frame / non-pagination R5 semantics     PRESERVED
B11 R5 pagination surfaces                       BOUNDED REOPEN
B11-F1 Access Assignment read precision          UNAFFECTED
B11 P9 bidirectional trace                       BOUNDED REOPEN FOR op6/op16/op22/op27 TRAVERSAL ONLY
B11 P10 pattern consolidation                    UNAFFECTED
B11 prior exact LOCK Evidence                     PRESERVED AS HISTORICAL ACCEPTED INPUT
89-operation / 11-route census                   UNAFFECTED
operating-method routing                         REBASELINED TO LOCAL v1.0 / v2.3
FP2 / P11                                        NOT OPEN
B12                                              NOT OPEN
```

Required CI remains intentionally limited to objective repository properties. Global Maximum, UX/architecture quality, evidence sufficiency, repository-reading depth and methodology reasoning are review/method concerns rather than grep-based CI assertions.

## B11 opening / acceptance law

Current accepted `main` before this increment states that B11 remains `NOT OPEN` until explicitly opened by a later acceptance increment.

This B11 acceptance increment is that explicit later increment.

```text
main before PR #173
  B11 NOT OPEN

PR #173 acceptance candidate
  explicitly opens B11
  + integrates the bounded B11-F1 read precision
  + consumes operator-operated B11 Evidence
  + must close all material B11 findings before integration
  + leaves B12/FP2/T12/implementation blocked

integration of PR #173
  is the repository state transition that accepts B11
```

A separate open-only PR is not required by the adopted Engineering Method or Frontend Product Experience Planning Method and would protect no additional invariant. Until this acceptance increment is integrated, current `main` remains authoritative with B11 `NOT OPEN`; this branch is the candidate transition, not a parallel main-roadmap authority.

## PR #173 pagination finding

The first merge attempt was correctly blocked by an unresolved review conversation. The review finding was validated against the exact R5/P9 Evidence and is material but bounded.

Root contradiction:

```text
op27 listGroupMembers is paginated
  but R5 Group member list had no continuation surface

op6 listUsers / op22 listGroups / op16 listAreas are paginated
  but the R5 grant composer represented those identity collections as complete plain selects

P9 marked those surfaces READY
  without proving how later pages were reached
```

Leaving that contradiction unresolved would force one of two invalid implementation outcomes:

```text
omit valid later-page identities/members
OR
perform a hidden full-collection crawl that B11 explicitly forbids
```

Global-Maximum disposition:

```text
CURRENT STRUCTURE CONFIRMED
+ bounded P8/P9 correction
```

No Product/backend/wire reopen is justified because the current operations already expose the required paginated truth.

Affected only:

```text
Group member list        op27 visible continuation
Add-member User picker   op6 continuation/failure precision
Grant User picker        op6 visible continuation
Grant Group picker       op22 visible continuation
Grant Area picker        op16 visible continuation
```

Unaffected and preserved:

```text
Por Área / Grupos / Funções IA
R4/R5 low-fi frame
Group multi-Area / Company footprint semantics
Area-specific vs Company-wide separation
fixed Role meaning
membership consequence
contextual Area/Group grant entry
Subject × Role × Scope final review
exact revoke
ambiguous Idempotency-Key retry
B11-F1 op31 precision
Authorization authority
89-operation census
P10 consolidation
```

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
B11   Access Administration                        BOUNDED P8-P9 REOPEN / R6 CANDIDATE / PARTIAL RE-LOCK REQUIRED
       durable prior Evidence locator               docs/decisions/t11-b11-lock-evidence.md
       prior Evidence ref                           evidence/t11-b11-locks-20260825
       prior Evidence exact commit                  469a753904041e7800400dc1074510456aa50df8
       prior Evidence exact tree                    c4f04b75c3676dcde00caa07279824b3c653c7f3
       prior R5 Git blob                            96094773435a88c357e308779639415d9853b327
       B11-F1 Access Assignment Read Precision      OPERATOR-RATIFIED / DURABLE CANDIDATE
       B11-F1 authority                             docs/decisions/access-assignment-read.md
       op31 listRoleAssignments                     REFINED / FILTERED + HUMAN-RECOGNIZABLE READ
       application-operation delta                  +0 / CENSUS REMAINS 89
       PR #173 review finding                       VALIDATED / MATERIAL / BOUNDED
       review finding Evidence                      docs/work/current/t11-b11-p8-r6-review-finding.md
       R5 preserved scope                           ALL EXCEPT AFFECTED PAGINATION / IDENTITY-PICKER TRAVERSAL
       P8 R6 artifact                               docs/work/current/t11-b11-access-administration-p8-r6.html
       P8 R6 Git blob                               26e8905c5c5012aba59280b1001f62529ed4dfd0
       P8 R6 structural verification                12 / 12 PASS
       P8 R6 Chromium behavior                      23 / 23 PASS
       P8 R6 JavaScript parse                       PASS
       P8 R6 operator disposition                   AWAITING PARTIAL RE-LOCK
       P9 R6 pagination delta                       docs/work/current/t11-b11-screen-contract-r6-delta.md
       P9 R6 disposition                            CANDIDATE / AWAITING OPERATOR PARTIAL RE-LOCK
       P10                                          PRESERVED / UNAFFECTED
       operation 90+ consumed                       0
       unresolved material B11 Findings             1
B12   Document Governance Administration           NOT OPEN
```

Exact prior locked frontend Evidence remains recoverable through:

```text
B01-B09
docs/decisions/t11-b01-b09-lock-evidence.md
evidence/t11-pr162-b01-b09-locks-20260824
→ adf58e448bc5bd3a20cae5b7228d729c031f94ac

B10
docs/decisions/t11-b10-lock-evidence.md
evidence/t11-pr170-b10-locks-20260824
→ b8c607cbd30d61d6bcf6ec1ea734ed1653d2569e

B11 prior R5 Evidence
docs/decisions/t11-b11-lock-evidence.md
evidence/t11-b11-locks-20260825
→ 469a753904041e7800400dc1074510456aa50df8
```

The prior B11 Evidence remains useful and protected for all unaffected semantics, but it no longer proves the reopened pagination surfaces after the material PR review finding. A new exact Evidence checkpoint is required after operator partial re-LOCK of R6.

## Exact next action

```text
1. Operator operates exact P8 R6 blob 26e8905c5c5012aba59280b1001f62529ed4dfd0; judge only the bounded pagination delta.
2. In Grupos → Aprovadores Financeiro, page the member list to a later page and verify a later-page member (fixture: Sofia Barros) is inspectable/removable.
3. Exercise continuation failure for the member list; loaded rows must remain visible and must not be presented as the complete membership.
4. In Conceder acesso, page User subjects to a later page and select/review Mariana Costa.
5. Switch the grant subject kind to Group, page to the later page and select/review Segurança Operacional.
6. Choose Area scope, page Areas to the later page and select/review Compras.
7. Exercise a supporting-read continuation failure; the loaded page/draft must remain intact.
8. Confirm Group-context grant still preselects Aprovadores Financeiro and Area-context grant still preselects Comercial.
9. If the operator explicitly re-LOCKs this bounded delta, close only P9 R6-01..R6-05 as READY/PASS; preserve all unaffected R5/P9/P10 decisions.
10. Create a new exact B11 R6 Evidence checkpoint and update the durable B11 Evidence locator without rewriting the prior R5 Evidence meaning.
11. Remove `docs/work/**` from the merge candidate, reply to the PR #173 review thread with exact proof, and resolve the thread only after the re-LOCK/P9 closure exists.
12. Mark PR #173 Ready only after the cleaned candidate is mergeable and `required` is green.
13. Because the accepted artifact changed after the earlier merge authorization, obtain fresh explicit operator merge authorization before squash merge.
14. B12 and FP2/P11 remain NOT OPEN. T12 and Product implementation remain BLOCKED.
```

## Hard stops

```text
no Product code/schema/OpenAPI/runtime/deploy implementation
no T12 work
no B12 work
no FP2/P11 work
no B01-B10 reopen without material Evidence
no widening of the B11 reopen beyond the validated pagination finding without new Evidence
no application operation 90+ for B11 without a new lawful bounded reopen
no global User/Group/Area search invented for UI convenience
no client-side crawl/post-filter over incomplete pages presented as complete
no frontend effective-permission authority or inferred inherited-access matrix
no Group single-Area ownership inferred from access scope
no custom Role/Permission editor
no ad-hoc rewrite of the shared local method files
no unapproved shell/sidebar/header/local-nav topology redesign
no production-like visual design authority inferred from P8
no assistant/reviewer LOCK
no declaring P9 complete while R6 lacks operator partial re-LOCK
no resolving the material PR #173 review thread before exact re-LOCK/P9 proof exists
no merge while the material finding remains open
no merge without fresh explicit operator authorization after the changed accepted artifact
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

Accepted Product/R10/frontend LOCK decisions reopen only on material Evidence under the DevelopmentConexus Engineering Method; the current reopen is bounded to the specific pagination surfaces falsified by PR #173 review.
