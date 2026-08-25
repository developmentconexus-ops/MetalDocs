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
T11 acceptance        B11 EXPLICITLY OPENED IN THIS INCREMENT / BOUNDED P8-P9 REOPEN / TWO REVIEW FINDINGS / AWAITING OPERATOR AUTHORIZATION
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

PR #172 restored these local methods after B11 planning had already begun on its feature branch. The restoration changed operating mechanics only; it did not change Product, Authorization, wire, UX or existing frontend truth.

Current bounded impact sweep after final-gate PR #173 review:

```text
B01-B10 protected structure / Screen Contracts  UNAFFECTED
B11 IA / frame / most R6/R7 semantics           PRESERVED
B11 add-member picker membership knowledge      BOUNDED REOPEN
B11 repeated successful grant confirmation      BOUNDED REOPEN
B11 R7 grant User raw-page fidelity             PRESERVED / RE-LOCKED
B11-F1 Access Assignment read precision         UNAFFECTED
B11 P9 bidirectional trace                      REOPEN AFFECTED ROWS ONLY
B11 P10 pattern consolidation                   UNAFFECTED / COMPLETE
B11 R5/R6/R7 exact Evidence                     PRESERVED / PARTIAL, NOT COMPLETE RECONSTRUCTION
89-operation / 11-route census                  UNAFFECTED
FP2 / P11                                       NOT OPEN
B12                                             NOT OPEN
```

Required CI remains intentionally limited to objective repository properties. Global Maximum, UX/architecture quality, evidence sufficiency, repository-reading depth and methodology reasoning are review/method concerns rather than grep-based CI assertions.

## B11 opening / acceptance law

Current accepted `main` before this increment states that B11 remains `NOT OPEN` until explicitly opened by a later acceptance increment. PR #173 is that explicit later increment, but it cannot integrate while material findings remain open.

```text
main before PR #173
  B11 NOT OPEN

PR #173 acceptance candidate
  explicitly opens B11
  + integrates B11-F1
  + consumes operator-operated B11 Evidence
  + must close every material review finding
  + leaves B12/FP2/T12/implementation blocked

integration of PR #173
  is the repository state transition that accepts B11
```

Until integration, current `main` remains authoritative with B11 `NOT OPEN`; this branch is the candidate transition, not a parallel main-roadmap authority.

## PR #173 first pagination finding — closed by R6

The first merge attempt was correctly blocked because R5/P9 did not prove visible continuation for Group members and paginated User/Group/Area selection.

Global-Maximum disposition:

```text
CURRENT STRUCTURE CONFIRMED
+ bounded P8/P9 correction
```

R6 added and the operator re-LOCKED visible continuation for op27/op6/op22/op16. Exact R6 Evidence remains preserved at:

```text
ref      evidence/t11-b11-r6-locks-20260825
commit   6dbcec41a43dc2a74629351e22b748188e5c6dc4
tree     c5054688c68068457a6c46add198c1797cddec0a
P8 blob  26e8905c5c5012aba59280b1001f62529ed4dfd0
```

## PR #173 grant User page-fidelity finding — closed by R7

A later review proved that R6 filtered ENABLED Users before pagination even though op6 exposes an unfiltered `UserPage` ordered by `user_id ASC`.

R7 re-LOCKed:

```text
raw op6 UserPage boundary remains authoritative
→ ENABLED selectable
→ DISABLED visible but unavailable
→ no pre-pagination client filter
→ no hidden all-page crawl
→ no invented op6 state filter/search
```

Exact R7 Evidence remains preserved at:

```text
ref      evidence/t11-b11-r7-amendment-20260825
commit   5c3b407c1bc0e789da823570a27c33e5f8f777c3
tree     077b25ffb9e5460f563ed84f7eedd4ed3a01d52f
P8 blob  3e9130fd7b9e5b6b414b5c8e96faf6c6644cb4df
proof    static + Chromium 12 / 12 PASS; JavaScript PASS
operator RE-LOCK APPROVED
P9      READY / PASS
```

R7 remains valid and is not reopened by the current findings.

## PR #173 final-gate findings — OPEN

Two additional review threads appeared during the final merge gate and were validated against the exact R6 Evidence.

### Finding A — add-member picker claims complete GroupMembership knowledge

R6 currently implements:

```text
op6 User picker
+ Set(all Group members from complete fixture)
→ every existing member disabled as "já membro"
```

But current accepted authority provides:

```text
op27 listGroupMembers
→ paginated GroupMemberPage only
→ no per-User GroupMembership lookup
→ no complete embedded membership projection
```

Therefore preserving complete “já membro” knowledge would require a forbidden hidden all-page crawl or a new lookup not currently accepted.

Smallest correction candidate:

```text
op6 UserPage remains picker source
→ do not claim complete membership knowledge
→ User DISABLED may remain unavailable from User state truth
→ already-member relation may be unknown before mutation
→ op28 PUT remains idempotent:
     201 first add
     204 already exists
→ reconcile result after command
→ optional local guidance may use only membership rows already legitimately loaded
```

No backend reopen is currently justified.

### Finding B — repeated successful grant confirmation violates same-key replay

R6 currently leaves the successful grant dialog/confirm action active and every subsequent click appends another fixture RoleAssignment while the same logical key remains active.

Accepted idempotency law is:

```text
same Idempotency-Key
+ same normalized command fingerprint
→ exact stored success replay
→ zero second semantic mutation
```

Smallest correction candidate:

```text
successful createRoleAssignment
→ command becomes terminal in current dialog
→ close/disable repeat confirmation
OR explicitly model same-key replay
→ repeated confirmation never appends another assignment
```

No backend reopen is justified.

### Finding disposition

```text
CURRENT STRUCTURE CONFIRMED
+ two smallest frontend P8/P9 corrections
```

The corrections are **not yet authorized**. Only the findings and candidate smallest resolutions are recorded.

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
B11   Access Administration                        BOUNDED P8/P9 REOPEN / TWO MATERIAL FINDINGS / AWAITING OPERATOR AUTHORIZATION
       durable Evidence locator                     docs/decisions/t11-b11-lock-evidence.md
       R6 preserved Evidence ref                    evidence/t11-b11-r6-locks-20260825
       R6 preserved Evidence commit                 6dbcec41a43dc2a74629351e22b748188e5c6dc4
       R6 full P8 blob                              26e8905c5c5012aba59280b1001f62529ed4dfd0
       R7 preserved amendment ref                   evidence/t11-b11-r7-amendment-20260825
       R7 preserved amendment commit                5c3b407c1bc0e789da823570a27c33e5f8f777c3
       R7 User-picker delta blob                    3e9130fd7b9e5b6b414b5c8e96faf6c6644cb4df
       B11-F1 Access Assignment Read Precision      OPERATOR-RATIFIED / DURABLE CANDIDATE
       application-operation delta                  +0 / CENSUS REMAINS 89
       first PR #173 finding                        CLOSED / R6 RE-LOCKED
       second PR #173 finding                       CLOSED / R7 RE-LOCKED
       third PR #173 finding                        OPEN / ADD-MEMBER COMPLETE-MEMBERSHIP KNOWLEDGE
       fourth PR #173 finding                       OPEN / REPEATED GRANT SAME-KEY SEMANTIC DUPLICATE
       P9 original material regions/controls        PRESERVED EXCEPT AFFECTED ROWS
       P9 primary B11 operations                    7 / 7 BOUND (ops 27-33)
       P9 supporting Organization reads             ops 6 / 16 / 22
       P10                                          PRESERVED / COMPLETE
       operation 90+ consumed                       0
       unresolved material B11 Findings             2
B12   Document Governance Administration           NOT OPEN
```

## Exact locked frontend Evidence

```text
B01-B09
docs/decisions/t11-b01-b09-lock-evidence.md
evidence/t11-pr162-b01-b09-locks-20260824
→ adf58e448bc5bd3a20cae5b7228d729c031f94ac

B10
docs/decisions/t11-b10-lock-evidence.md
evidence/t11-pr170-b10-locks-20260824
→ b8c607cbd30d61d6bcf6ec1ea734ed1653d2569e

B11 R6 preserved base Evidence
evidence/t11-b11-r6-locks-20260825
→ 6dbcec41a43dc2a74629351e22b748188e5c6dc4

B11 R7 preserved User-picker amendment Evidence
evidence/t11-b11-r7-amendment-20260825
→ 5c3b407c1bc0e789da823570a27c33e5f8f777c3

B11 prior R5 Evidence
evidence/t11-b11-locks-20260825
→ 469a753904041e7800400dc1074510456aa50df8
```

R5/R6/R7 Evidence remains valid for unaffected semantics but must not be treated as a fully realizable reconstruction package while the two current findings remain unresolved.

## Exact next action

```text
1. Operator ratifies, rejects, or restructures the two candidate smallest corrections above.
2. If authorized, create one bounded P8 delta covering only:
     A. add-member picker without complete hidden membership knowledge;
     B. terminal/replay-safe successful grant confirmation.
3. Prove Finding A against op6 + paginated op27 + idempotent op28 201/204 without hidden crawl or new operation.
4. Prove Finding B against existing createRoleAssignment Idempotency-Key law: repeated same command must produce zero second semantic mutation.
5. Preserve all R6/R7 unaffected semantics and do not redesign Access Administration.
6. Only after operator use/re-LOCK close the corresponding P9 rows and preserve new exact amendment Evidence.
7. Remove any temporary `docs/work/**`, reply to both review threads with exact proof, and resolve them only after re-LOCK.
8. Return PR #173 to Ready only after the cleaned candidate is mergeable and no material thread remains open.
9. Run/inspect final `required` on the final Ready head.
10. Obtain fresh explicit operator merge authorization for that final head before squash merge.
11. B12 and FP2/P11 remain NOT OPEN. T12 and Product implementation remain BLOCKED.
```

## Hard stops

```text
no Product code/schema/OpenAPI/runtime/deploy implementation
no T12 work
no B12 work
no FP2/P11 work
no B01-B10 reopen without material Evidence
no widening of B11 beyond the two validated current findings without new Evidence
no implementing the two candidate corrections before operator authorization
no application operation 90+ for B11 without a new lawful bounded reopen
no invented GroupMembership lookup/search for UI convenience
no hidden op27 all-page crawl to manufacture complete membership knowledge
no second semantic grant mutation for same Idempotency-Key + same command
no invented op6 state filter/search
no pre-pagination client filter that changes server page boundaries
no client-side crawl/post-filter over incomplete pages presented as complete
no frontend effective-permission authority or inferred inherited-access matrix
no Group single-Area ownership inferred from access scope
no custom Role/Permission editor
no ad-hoc rewrite of the shared local method files
no assistant/reviewer LOCK
no resolving the two current material review threads before exact operator re-LOCK/P9 proof exists
no merge while either current finding remains open
no merge without fresh explicit operator authorization on the final accepted head
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

Accepted Product/R10/frontend LOCK decisions reopen only on material Evidence under the DevelopmentConexus Engineering Method. The current reopen is bounded solely to the two final-gate frontend contradictions validated from PR #173 review.
