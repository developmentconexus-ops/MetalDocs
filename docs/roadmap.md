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
T11 acceptance        B11 EXPLICITLY OPENED IN THIS INCREMENT / LOCKED / R6 BASE + R7 AMENDMENT / P9-P10 COMPLETE / PENDING INTEGRATION
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

Final bounded impact sweep after both PR #173 review findings:

```text
B01-B10 protected structure / Screen Contracts  UNAFFECTED
B11 IA / frame / non-pagination semantics       PRESERVED
B11 R6 member/add-member/Group/Area pagination PRESERVED / RE-LOCKED
B11 grant User picker                           RE-LOCKED BY R7 AMENDMENT
B11-F1 Access Assignment read precision        UNAFFECTED
B11 P9 bidirectional trace                     COMPLETE / R7 SUPERSEDES R6-03 ONLY
B11 P10 pattern consolidation                  UNAFFECTED / COMPLETE
B11 exact LOCK Evidence                        R6 BASE + R7 AMENDMENT CURRENT
89-operation / 11-route census                 UNAFFECTED
FP2 / P11                                      NOT OPEN
B12                                            NOT OPEN
```

Required CI remains intentionally limited to objective repository properties. Global Maximum, UX/architecture quality, evidence sufficiency, repository-reading depth and methodology reasoning are review/method concerns rather than grep-based CI assertions.

## B11 opening / acceptance law

Current accepted `main` before this increment states that B11 remains `NOT OPEN` until explicitly opened by a later acceptance increment. PR #173 is that explicit later increment.

```text
main before PR #173
  B11 NOT OPEN

PR #173 acceptance candidate
  explicitly opens B11
  + integrates B11-F1
  + consumes operator-operated B11 Evidence
  + closes both material review findings through bounded re-LOCKs
  + records P9/P10 closure
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

R6 added and the operator re-LOCKED:

```text
Group member list        op27 visible continuation
Add-member User picker   op6 visible continuation / failure precision
Grant User picker        op6 visible continuation
Grant Group picker       op22 visible continuation
Grant Area picker        op16 visible continuation
```

R6 proof preserved:

```text
P8 R6 Git blob                    26e8905c5c5012aba59280b1001f62529ed4dfd0
R6 structural verification       12 / 12 PASS
R6 Chromium behavior             23 / 23 PASS
R6 JavaScript parse              PASS
R6 operator partial re-LOCK      APPROVED
P9 R6 controls                   5 / 5 READY / PASS at that checkpoint
```

## PR #173 grant User picker page-fidelity finding — closed by R7

A later review found one narrower contradiction in the exact R6 artifact.

R6 formed the grant User picker as:

```text
all Users fixture
→ client filter ENABLED
→ paginate
```

Current op6 authority is:

```text
GET /api/v1/users
listUsers
UserPage
PAGED
user_id ASC
no state filter
```

Therefore the R6 grant User picker changed server page boundaries before rendering. Reproducing that behavior in production would require a hidden crawl/post-filter over op6 pages, which B11 already forbids.

Disposition:

```text
CURRENT STRUCTURE CONFIRMED
+ smallest P8/P9 correction
```

No Product/backend/wire reopen was justified.

R7 re-LOCKed target invariant:

```text
server UserPage boundary remains authoritative
→ render every User returned on the page
→ ENABLED selectable
→ DISABLED visible but unavailable
→ no pre-pagination client filter
→ no hidden all-page crawl
→ no invented op6 state filter/search
```

R7 exact Evidence:

```text
P8 R7 delta path       docs/work/current/t11-b11-grant-user-picker-p8-r7.html
P8 R7 Git blob         3e9130fd7b9e5b6b414b5c8e96faf6c6644cb4df
R7 Evidence ref        evidence/t11-b11-r7-amendment-20260825
R7 Evidence commit     5c3b407c1bc0e789da823570a27c33e5f8f777c3
R7 Evidence tree       077b25ffb9e5460f563ed84f7eedd4ed3a01d52f
static + Chromium      12 / 12 PASS
JavaScript parse       PASS
operator disposition   APPROVED / RE-LOCK
P9 R7                  READY / PASS
```

R7 supersedes only the grant User picker row previously named R6-03. All other R6 pagination and B11 semantics remain protected.

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
B11   Access Administration                        LOCKED / R6 BASE + R7 AMENDMENT / P9-P10 COMPLETE / OPERATOR-RATIFIED / ACCEPTANCE CANDIDATE
       durable Evidence locator                     docs/decisions/t11-b11-lock-evidence.md
       R6 complete base Evidence ref                evidence/t11-b11-r6-locks-20260825
       R6 complete base Evidence commit             6dbcec41a43dc2a74629351e22b748188e5c6dc4
       R6 complete base Evidence tree               c5054688c68068457a6c46add198c1797cddec0a
       R6 full P8 blob                              26e8905c5c5012aba59280b1001f62529ed4dfd0
       R7 User-picker amendment ref                 evidence/t11-b11-r7-amendment-20260825
       R7 User-picker amendment commit              5c3b407c1bc0e789da823570a27c33e5f8f777c3
       R7 User-picker amendment tree                077b25ffb9e5460f563ed84f7eedd4ed3a01d52f
       R7 User-picker delta blob                    3e9130fd7b9e5b6b414b5c8e96faf6c6644cb4df
       B11-F1 Access Assignment Read Precision      OPERATOR-RATIFIED / DURABLE CANDIDATE
       application-operation delta                  +0 / CENSUS REMAINS 89
       first PR #173 finding                        CLOSED / R6 RE-LOCKED
       second PR #173 finding                       CLOSED / R7 RE-LOCKED
       R6 structural verification                   12 / 12 PASS
       R6 Chromium behavior                         23 / 23 PASS
       R7 static + Chromium                         12 / 12 PASS
       R7 JavaScript parse                          PASS
       R7 operator disposition                      RE-LOCK / APPROVED
       P9 original material regions/controls        36 / 36 TRACED subject to R6/R7 supersession on affected rows
       P9 primary B11 operations                    7 / 7 BOUND (ops 27-33)
       P9 supporting Organization reads             ops 6 / 16 / 22
       P9 R6 pagination delta                       4 / 5 PRESERVED; R6-03 SUPERSEDED BY R7
       P9 R7 User picker delta                      READY / PASS
       P10                                          PRESERVED / COMPLETE
       operation 90+ consumed                       0
       unresolved material B11 Findings             0
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

B11 R6 complete base Evidence
docs/decisions/t11-b11-lock-evidence.md
evidence/t11-b11-r6-locks-20260825
→ 6dbcec41a43dc2a74629351e22b748188e5c6dc4

B11 R7 User-picker amendment Evidence
evidence/t11-b11-r7-amendment-20260825
→ 5c3b407c1bc0e789da823570a27c33e5f8f777c3

B11 prior R5 Evidence
evidence/t11-b11-locks-20260825
→ 469a753904041e7800400dc1074510456aa50df8
```

Current B11 reconstruction law is R6 complete base + R7 amendment. R5 remains historical Evidence.

## Exact next action

```text
1. Remove `docs/work/**` from the PR #173 merge candidate now that exact R7 amendment Evidence is preserved.
2. Reply to the second PR #173 review thread with the exact R7 blob, operator re-LOCK, P9 R7 PASS and amendment Evidence ref; resolve the thread because its material finding is now closed.
3. Revalidate PR #173 against current main; the branch must remain mergeable with no unresolved material review conversation.
4. Update the PR description to the final R6-base + R7-amendment state and mark PR #173 Ready only after the cleaned candidate contains no `docs/work/**`.
5. Run/inspect the final objective aggregate check `required` on the Ready candidate.
6. Because accepted Evidence changed again after prior merge authorization, obtain fresh explicit operator merge authorization before squash merge.
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
no invented op6 state filter/search for UI convenience
no pre-pagination client filter that changes server page boundaries
no client-side crawl/post-filter over incomplete pages presented as complete
no frontend effective-permission authority or inferred inherited-access matrix
no Group single-Area ownership inferred from access scope
no custom Role/Permission editor
no ad-hoc rewrite of the shared local method files
no assistant/reviewer LOCK
no merge while any material review thread remains unresolved
no merge without fresh explicit operator authorization after the changed accepted Evidence
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

Accepted Product/R10/frontend LOCK decisions reopen only on material Evidence under the DevelopmentConexus Engineering Method. Both PR #173 findings were closed by bounded frontend corrections and operator re-LOCKs; no broader B11 authority was reopened.
