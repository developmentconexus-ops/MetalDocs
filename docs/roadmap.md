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
T11 acceptance        B11 EXPLICITLY OPENED IN THIS INCREMENT / LOCKED / P8 R6-P10 COMPLETE / PENDING INTEGRATION
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

Final bounded impact sweep after the PR #173 review finding and R6 partial re-LOCK:

```text
B01-B10 protected structure / Screen Contracts  UNAFFECTED
B11 IA / frame / non-pagination R5 semantics     PRESERVED
B11 pagination surfaces                          RE-LOCKED IN R6
B11-F1 Access Assignment read precision          UNAFFECTED
B11 P9 bidirectional trace                       COMPLETE + R6 PAGINATION DELTA PASS
B11 P10 pattern consolidation                    UNAFFECTED / COMPLETE
B11 exact LOCK Evidence                          R6 CURRENT + R5 HISTORICAL PRESERVED
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
  + consumes operator-operated B11 R5 + R6 Evidence
  + closes the material PR review finding through bounded partial re-LOCK
  + records P9/P10 closure
  + leaves B12/FP2/T12/implementation blocked

integration of PR #173
  is the repository state transition that accepts B11
```

A separate open-only PR is not required by the adopted Engineering Method or Frontend Product Experience Planning Method and would protect no additional invariant. Until this acceptance increment is integrated, current `main` remains authoritative with B11 `NOT OPEN`; this branch is the candidate transition, not a parallel main-roadmap authority.

## PR #173 pagination finding — closed

The first merge attempt was correctly blocked by an unresolved review conversation. The review finding was validated against the exact R5/P9 Evidence and was material but bounded.

Root contradiction:

```text
op27 listGroupMembers is paginated
  but R5 Group member list had no continuation surface

op6 listUsers / op22 listGroups / op16 listAreas are paginated
  but the R5 grant composer represented those identity collections as complete plain selects

P9 marked those surfaces READY
  without proving how later pages were reached
```

Leaving that contradiction unresolved would have forced one of two invalid implementation outcomes:

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

No Product/backend/wire reopen was required because the current operations already expose the necessary paginated truth.

R6 corrected and the operator re-LOCKED only:

```text
Group member list        op27 visible continuation
Add-member User picker   op6 visible continuation / failure precision
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

Closure proof:

```text
P8 R6 Git blob                    26e8905c5c5012aba59280b1001f62529ed4dfd0
R6 structural verification       12 / 12 PASS
R6 Chromium behavior             23 / 23 PASS
R6 JavaScript parse              PASS
R6 operator partial re-LOCK      APPROVED
P9 R6 controls                   5 / 5 READY
P9 R6 verdict                    PASS
unresolved material finding      0
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
B11   Access Administration                        LOCKED / P8 R6-P10 COMPLETE / OPERATOR-RATIFIED / ACCEPTANCE CANDIDATE
       durable Evidence locator                     docs/decisions/t11-b11-lock-evidence.md
       current Evidence ref                         evidence/t11-b11-r6-locks-20260825
       current Evidence exact commit                6dbcec41a43dc2a74629351e22b748188e5c6dc4
       current Evidence exact tree                  c5054688c68068457a6c46add198c1797cddec0a
       canonical P8 R6 Git blob                     26e8905c5c5012aba59280b1001f62529ed4dfd0
       prior R5 Evidence ref                        evidence/t11-b11-locks-20260825
       prior R5 Evidence exact commit               469a753904041e7800400dc1074510456aa50df8
       prior R5 Git blob                            96094773435a88c357e308779639415d9853b327
       B11-F1 Access Assignment Read Precision      OPERATOR-RATIFIED / DURABLE CANDIDATE
       B11-F1 authority                             docs/decisions/access-assignment-read.md
       op31 listRoleAssignments                     REFINED / FILTERED + HUMAN-RECOGNIZABLE READ
       application-operation delta                  +0 / CENSUS REMAINS 89
       PR #173 review finding                       CLOSED / VALIDATED + CORRECTED + RE-LOCKED
       R6 structural verification                   12 / 12 PASS
       R6 Chromium behavior                         23 / 23 PASS
       R6 JavaScript parse                          PASS
       R6 operator disposition                      PARTIAL RE-LOCK / APPROVED
       P9 original material regions/controls        36 / 36 TRACED subject to R6 supersession on affected rows
       P9 primary B11 operations                    7 / 7 BOUND (ops 27-33)
       P9 supporting Organization reads             ops 6 / 16 / 22
       P9 R6 pagination delta                       5 / 5 READY / PASS (ops 6 / 16 / 22 / 27)
       P10 existing shared behaviors reused         4
       P10 new shared semantic patterns             0
       P10 B11-local protected patterns             9
       operation 90+ consumed                       0
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

B11 current R6 Evidence
docs/decisions/t11-b11-lock-evidence.md
evidence/t11-b11-r6-locks-20260825
→ 6dbcec41a43dc2a74629351e22b748188e5c6dc4

B11 prior R5 Evidence
evidence/t11-b11-locks-20260825
→ 469a753904041e7800400dc1074510456aa50df8
```

The Evidence refs are non-authoritative recovery/provenance inputs. R6 is the current complete B11 reconstruction artifact; R5 remains historical Evidence for the pre-review LOCK and unaffected semantics.

## Exact next action

```text
1. Remove `docs/work/**` from the PR #173 merge candidate now that exact R6 Evidence is preserved.
2. Reply to the PR #173 review thread with the exact R6 blob, operator partial re-LOCK and P9 R6 PASS; resolve the thread because its material finding is now closed.
3. Revalidate PR #173 against current main; the branch must remain mergeable with no unresolved material review conversation.
4. Mark PR #173 Ready only after the cleaned candidate contains no `docs/work/**`.
5. Run/inspect the final objective aggregate check `required` on the Ready candidate.
6. Because the accepted artifact changed after the earlier merge authorization, obtain fresh explicit operator merge authorization before squash merge.
7. B12 and FP2/P11 remain NOT OPEN. T12 and Product implementation remain BLOCKED.
```

## Hard stops

```text
no Product code/schema/OpenAPI/runtime/deploy implementation
no T12 work
no B12 work
no FP2/P11 work
no B01-B10 reopen without material Evidence
no widening of B11 without new material Evidence
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
no declaring B11/P9 closed without current R6 operator partial re-LOCK Evidence
no resolving a material review thread before exact re-LOCK/P9 proof exists
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

Accepted Product/R10/frontend LOCK decisions reopen only on material Evidence under the DevelopmentConexus Engineering Method. The PR #173 pagination finding was closed by the bounded R6 correction and operator partial re-LOCK; no broader B11 authority was reopened.
