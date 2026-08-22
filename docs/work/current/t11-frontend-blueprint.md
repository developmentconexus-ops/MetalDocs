# T11 — Frontend Implementation Blueprint

> **TEMPORARY T11 CANDIDATE / BRANCH-ONLY WORK.** This is the consolidated F1→F9 frontend implementation-readiness candidate. It derives from accepted Product/T6/T8-E/T8-F authority and the operator-approved responsible-owner selection precision. It creates no frontend semantic owner, Product capability or API operation.

## 1. Goal

Freeze enough frontend behavior **before implementation** that a future implementer does not need to invent:

```text
which screens exist
what each screen means
which backend truth feeds each block
what each button calls
where target ids come from
what route follows a successful action
how concurrency/idempotency/exact-content failures change safe UX
what must never become frontend authority
```

The blueprint follows:

```text
accepted authority
→ coverage
→ material surfaces
→ Screen Contracts
→ navigation/data graph
→ functional wireframes
→ material interaction ledger
→ bidirectional 78-operation trace
→ finding closure
```

Detailed functional wireframes:

```text
docs/work/current/t11-wireframes.md
```

Detailed material control ledger:

```text
docs/work/current/t11-interaction-ledger.md
```

## 2. Fixed frontend envelope

```text
React SPA                              accepted
TanStack Query                         server-state mechanism
generated TypeScript wire projection  required
thin application transport            one boundary
stable Product SPA paths              exact 10
application operations                78
operation 79                          absent
frontend semantic owner               none
frontend Authorization engine         absent
parallel handwritten DTO/API authority absent
parallel global server-truth store     absent
interactive DOCX adapter               one boundary
```

Stable Product paths:

```text
/documents
/documents/:document_id
/documents/:document_id/work
/documents/:document_id/history
/work
/work/governance/:attempt_id
/audit
/admin/organization
/admin/access
/admin/document-governance
```

Browser AuthN integration remains outside the application census:

```text
/auth/login
/auth/callback
```

## 3. Client state authority

Exactly four baseline client state classes remain:

```text
SERVER STATE
  Product/server truth → TanStack Query

NAVIGATION / URL
  route, admitted catalog filters, cursor context, route-local presentation selection

FORM DRAFT
  unaccepted user input / editor buffer

EPHEMERAL UI
  dialog/drawer/focus/selection/disclosure state
```

No Redux/Zustand/global entity mirror is baseline. A route, component or cache entry never becomes lifecycle/Authorization authority.

## 4. Human-goal coverage — 16 / 16

| Accepted human goal | Primary route/surface | Wireframe |
|---|---|---|
| Establish/end session | APP shell | WF-00/WF-01 |
| Discover official documents | `/documents` LIB-01 | WF-02 |
| Create Document | LIB-02 → Work | WF-03 → WF-06/07/08 |
| Inspect official/current Document | OFF-01/OFF-02 | WF-04 |
| Start/enter open Revision | OFF-04 → Work | WF-04 → WF-06/07 |
| Author DRAFT | DW-01 | WF-07/WF-09 |
| Upload replacement source | DW-02 | WF-08/WF-09 |
| Submit/withdraw/cancel | DW-03/DW-04 | WF-07/WF-10 |
| See actor-relevant work | WRK-01/WRK-02 | WF-11 |
| Participate in governance | GOV-01/02/03 | WF-12/WF-13 |
| Inspect Document history | HIS-01 | WF-14 |
| Initiate/manage obsolescence | OFF-05 | WF-05 |
| Administer Organization | ORG-01..08 | WF-16..19 |
| Administer access | ACC-01/02 | WF-20/WF-21 |
| Administer document governance | DGV-01..06 | WF-22..25 |
| Inspect Audit | AUD-01 | WF-15 |

## 5. Material surface inventory — 36

Surface IDs are planning handles, not routes or semantic owners.

```text
Application shell                 APP-01..02       2
Library                           LIB-01..02       2
Document Official                 OFF-01..05       5
History                           HIS-01           1
My Work                           WRK-01..02       2
Document Work                     DW-01..04        4
Governance Case                   GOV-01..03       3
Admin Organization                ORG-01..08       8
Admin Access                      ACC-01..02       2
Admin Document Governance         DGV-01..06       6
Audit                             AUD-01           1
----------------------------------------------------
TOTAL                                              36
```

A separate surface exists only when semantic truth, safe action, owner/write, target identity, OCC/idempotency/exact-byte behavior, lifecycle/disclosure, recovery path or editor/viewer mode materially changes.

## 6. Screen Contract matrix

| Surface | Primary truth/read | Material write/control | Material safe-state obligation |
|---|---|---|---|
| APP-01 | `getSession`/401 | `/auth/login` browser flow | no local password/provider-role authority |
| APP-02 | `getSession` | `endSession` | shell navigation presence != permission |
| LIB-01 | `listDocuments` | filters/page/row open | official discovery only; no DRAFT-as-official |
| LIB-02 | creation options + numbering preview | `createDocument` | preview non-reserving; create→real Work target |
| OFF-01 | `getDocument` | route/read composition | server-derived status/refs only |
| OFF-02 | `getRelease` + exact bytes | source/viewer open | SourceOnly vs OfficialRendition exact distinction |
| OFF-03 | T8-E-RO candidates + owner ETag | `replaceDocumentResponsibleOwner` | candidate projection != concurrency/eligibility guarantee |
| OFF-04 | `getDocument.open_revision` | `createDocumentRevision` or enter Work | current resolver reread; no History fallback |
| OFF-05 | active request ref + request read | create/withdraw obsolescence | NoHumanApproval has no fake human Step |
| HIS-01 | `getDocumentHistory` | supporting read-only inspection | History never current resolver |
| WRK-01 | `listAuthoringWork` | navigate to Work | projection stale-safe/current target rechecks |
| WRK-02 | `listGovernanceWork` | navigate Case | projection never participation authority |
| DW-01 | Revision + DRAFT + ETag + exact source | `updateRevisionDraft` | DOCX editable/PDF read-only; explicit 412 reconciliation |
| DW-02 | current DRAFT + intended local bytes | allocate→PUT→complete→attach | provider success != READY/WorkingContent; expiry uses new allocation |
| DW-03 | current DRAFT ETag | `createSubmission` | same-key ambiguous retry; no optimistic Submission/Release |
| DW-04 | Revision.current_submission + Submission | withdraw/cancel | immutable submitted content + orthogonal gates |
| GOV-01 | `getGovernanceAttempt` | case navigation/inspection | exact immutable subject; allowed_actions hints only |
| GOV-02 | case/feedback | `createGovernanceFeedback` | idempotent immutable feedback |
| GOV-03 | Step/Decision | `recordGovernanceStepDecision` | singleton; Decision != Release |
| ORG-01 | Company + ETag | replace Company | independent OCC domain |
| ORG-02 | Users + provider preflight | create User | atomic User/Profile/Binding; opaque provider ref |
| ORG-03 | Profile + profile ETag/absence | replace/recreate/erase | profile != identity; exact If-Match/If-None-Match matrix |
| ORG-04 | Provider Binding + ETag | replace binding | replacement session consequence server-owned |
| ORG-05 | Eligibility + ETag | disable/re-enable | offboarding teardown; re-enable does not resurrect access |
| ORG-06 | Areas + Area metadata ETag | create/rename | code immutable after create |
| ORG-07 | Area lifecycle + ETag | retire/re-enable | independent from Area metadata OCC |
| ORG-08 | Groups + Group ETag | create/rename/delete | deletion live-dependency conflict distinct |
| ACC-01 | Groups/members/admitted User refs | add/remove member | membership Organization-owned, access-sensitive |
| ACC-02 | fixed Role catalog + assignments | grant/revoke | no custom Role/Permission editor/hierarchy |
| DGV-01 | DocumentTypes | create type | initial governance/representation required |
| DGV-02 | DocumentType base + ETag | replace base | accepted immutability/conflict preserved |
| DGV-03 | governance/representation + ETag | replace policy | ordered closed route, not workflow builder |
| DGV-04 | eligible Template set + ETag | replace set | set semantics; empty valid |
| DGV-05 | numbering preview | read only | reservation=false visibly communicated |
| DGV-06 | Template config + Template-role ETag | replace Template role | config privilege != content/history access |
| AUD-01 | AuditEventPage | read/page only | evidence != current state; no generic search/export |

All 36 are READY after the approved responsible-owner read precision.

## 7. Operator-approved bounded precision — T8-E-RO

Durable precision record:

```text
docs/decisions/responsible-owner-selection-read.md
```

Existing operation 47 gains one optional derived projection member:

```text
DocumentOfficialView.responsible_owner_candidates?: UserReference[]
```

Presence/completeness law:

```text
present iff current canonical document.owner.manage = ALLOW for the exact Document
contents = complete existing + same-Company + ENABLED UserReference set
order = user_id ASC
absence discloses neither candidate existence nor reason
```

Mutation concurrency remains:

```text
getDocumentResponsibleOwner → ResponsibleOwnerView + ETag
→ replaceDocumentResponsibleOwner(target user_id, If-Match)
→ recheck current AuthZ + target eligibility/offboarding serialization
```

No operation, Permission, owner or ETag domain is added.

## 8. Navigation / data graph

### Route-local browser presentation state

No nested Product route is invented. Large composed routes use closed browser-only query state:

```text
/work?lane=authoring|governance
/admin/organization?section=company|users|areas|groups
/admin/access?section=memberships|roles
/admin/document-governance?section=document-types|templates
```

Unknown value normalizes to route default. These values never become `/api/v1` semantics.

### Primary cross-route edges

```text
Shell → Library
  /documents → listDocuments

Shell → My Work
  /work → listAuthoringWork / listGovernanceWork by lane

Shell → Audit/Admin
  destination read decides 403/404; shell has no permission snapshot

Library row
  returned document_id → /documents/:id → getDocument

Create Document
  CreateDocumentResult.document_id
  → /documents/:id/work
  → getDocument.open_revision current resolver

Document Official → Work
  current getDocument.open_revision
  → /documents/:id/work
  → getDocument again

Document Official create-next
  createDocumentRevision
  → Work route
  → current getDocument resolver

Document Official → History
  document_id → /documents/:id/history → getDocumentHistory

Authoring Work row
  returned document_id → Work route → current getDocument resolver

Governance Work row
  returned attempt_id → /work/governance/:attempt_id → getGovernanceAttempt

Governance Case subject
  returned document_id → Document Official optional context navigation

History item
  only exact returned ids may open inline Revision/Submission/Release/Rendition/Request reads
  governance_attempt_id may navigate exact Case
```

No History/Audit/provider/client-only fallback resolves current resources.

### Direct Work reload

```text
getDocument.open_revision absent
  → no current Work synthesized

state=draft
  → getRevision + getRevisionDraft + exact DRAFT source

state=submitted
  → getRevision.current_submission_id
  → getSubmission + submitted source as needed
```

## 9. Library filter identity rule

Ordinary Library does not receive a new low-privilege reference-directory platform merely to populate filters.

```text
q + status
  ordinary explicit controls

area/type/responsible_owner ids
  may be activated from already-disclosed references in DocumentSummary/DocumentOfficial context
  persist in URL
  are always clearable
```

Real UX evidence that arbitrary independent selector discovery is required is a concrete future read-consumer reopen trigger. Until then, YAGNI rejects a generic directory/screen-shaped endpoint.

## 10. Mutation successor law

```text
endSession
→ APP-01

createDocument
→ Document Work (never intermediate dead Official target)

replace responsible owner
→ remain Document Official; replace owner result + refetch composed lenses

createDocumentRevision
→ Work current resolver

create/withdraw obsolescence
→ remain Document Official; refetch request/Document/History

update DRAFT / attach source
→ remain Work; authoritative DocumentWorkView + new ETag

createSubmission governance_pending|rendition_pending
→ remain Work in submitted-state view

createSubmission released
→ refetch/navigate Document Official

withdraw/cancel Revision
→ refetch getDocument; current open_revision decides Work vs Official

feedback/Step Decision
→ remain Governance Case and refetch canonical case

Admin mutations
→ remain owning route/section; replace exact returned ETag representation + narrow refetch
```

No mutation response is expanded into client lifecycle authority.

## 11. Material failure UX

Wireframe overlays define:

```text
401
→ session presentation invalid → sign-in gate

403 permission.denied
→ denied destination/action; server remains authority

404
→ "not available"; absent vs non-disclosable not distinguished

permission.csrf_failed
→ re-bootstrap getSession/CSRF
→ retry same logical unsafe command only where accepted-safe

412 resource_changed
→ preserve local form
→ refetch current ETag representation
→ explicit review/re-submit

412 draft_changed
→ preserve local editor/form
→ fresh DRAFT + explicit manual reconciliation
→ no automatic merge/LWW

ambiguous idempotent outcome
→ explicit "Retry same request" using same key

upload_expired
→ retain intended local bytes
→ new allocation + reupload

dependency.*
→ sanitized unavailable state; raw provider error absent

internal.content_integrity
→ no partial viewer success
```

## 12. Material Interaction Ledger closure

Detailed ledger:

```text
docs/work/current/t11-interaction-ledger.md
```

Every material write/read/navigation control is bound to:

```text
owner
operationId/browser AuthN mechanism
input identity source
CSRF/If-Match/Idempotency-Key/exact-byte/cursor mechanics
success truth
material failure
cache/refetch/navigation consequence
retry law
forbidden inference
```

No wireframe contains an unbound material button.

## 13. Exact operation/tranche trace — 78 / 78

Final implementation assignment:

```text
S1  operations 1–33                                33
S2  operations 34–43                               10
S3  operations 44–54 except 55 + 56–66             22
S4  operation 55 + operations 67–74                  9
S5  operations 75–78                                 4
------------------------------------------------------
TOTAL                                                78
```

Primary frontend homes:

```text
1–2       APP-01/02                  WF-00/01
3         ORG-02/04                  WF-17/18
4–5       ORG-01                     WF-16
6–15      ORG-02..05 + Admin refs    WF-17/18/20/21/23
16–21     ORG-06/07 + Admin refs     WF-19/21/23
22–26     ORG-08 + Admin refs        WF-19/20/21/23
27–29     ACC-01                     WF-20
30–33     ACC-02                     WF-21
34–43     DGV-01..06                 WF-22..25
44–46     LIB-02/LIB-01              WF-03/02
47        OFF-01 + Work resolver     WF-04/05/06
48–49     OFF-03                     WF-05
50–51     DGV-06                     WF-25
52        OFF-04→Work                WF-04/06/07
53        HIS-01                     WF-14
54        WRK-01                     WF-11
55        WRK-02                     WF-11
56–61     DW-01/02                   WF-07/08/09
62–66     DW-03/04                   WF-07/10
67–71     GOV-01/02/03               WF-12/13
72–74     OFF-02/History             WF-04/14
75–77     OFF-05                     WF-05
78        AUD-01                     WF-15
```

```text
unassigned operations       0
multiply-owned operations   0
invented operations         0
operation 79                absent
```

## 14. Cross-cutting census reconciliation

### Idempotency — exact 10

```text
createUser
createArea
createGroup
createRoleAssignment
createDocumentType
createDocument
createDocumentRevision
createSubmission
createGovernanceFeedback
createObsolescenceRequest
```

All 10 have explicit same-logical-command retry UX. No extra action is keyed.

### OCC/ETag — 13 / 13

```text
Company
UserProfile
UserProviderBinding
UserEligibility
Area metadata
Area lifecycle
Group metadata
DocumentType base
DocumentType governance
DocumentType eligible Templates
Document responsible owner
Document Template role
DRAFT WorkingContent generation
```

All bind user edits to the loaded current validator and have explicit stale reconciliation. Responsible-owner candidate enrichment is not an OCC domain.

### Exact bytes — 4 / 4

```text
getRevisionDraftSource
getSubmissionSource
getReleaseSource
getOfficialRenditionContent
```

Each has an exact editor/viewer home. No provider semantic URL or Range authority is added.

## 15. Golden Flow / validation linkage

```text
GF1 → S1 → session/admin/access wireframes
GF2 → S2+S3 → governance config + create→Work
GF3 → S3 → DRAFT/upload/Submission
GF4 → S3+S4 → Submission→Governance→Release/Official
GF5 → S3+S5 → Library/Official/obsolescence/history
GF6 → P3/P4 → runtime recovery; frontend only truthful failure states
```

T9 V linkage:

```text
V1  P1/P5 + exact 78/78 generated consumer trace
V2  P1/P5 closed-world imports; no frontend owner
V3  S1 + server 401/403/404/CSRF behavior; no client evaluator
V4  P2 + every F6 semantic mutation path
V5  exact 10 idempotent controls
V6  exact 13 OCC domains
V7  upload/admission + 4 exact-byte resources
V8  S4/P4 durable work; queue state absent from UI
V9  P3/P4 dependency/resource/readiness laws
V10 P4 restore/security readiness; no Product restore UI
```

## 16. Finding history / reopen routing

Findings produced by coverage-first planning:

```text
F1
  5 implementation-graph/dead-target findings
  all corrected within T11

F3-F01
  responsible-owner candidate discovery was a real accepted-journey read asymmetry
  operator approved bounded T6/T8-E/T8-F precision T8-E-RO
  operation count unchanged

F4
  createDocument successor corrected to Document Work
  no upstream semantic reopen

F4-P01
  global ordinary-reader reference directory deliberately absent under YAGNI
  concrete independent-selection UX failure is reopen trigger

F5/F6
  no new material findings

self-review
  original S3 Document-core / S4 Work split produced a dead human-flow seam
  corrected by merging complete Document create+authoring+Submission into final S3
```

Reopen law:

```text
presentation/layout issue only → correct blueprint
missing trace but accepted truth exists → correct blueprint/ledger
accepted human goal not representable → smallest Product/T6/T8-E/T8-F owner
operation79 proposal → hard STOP / material explicit reopen
```

## 17. Frontend implementation-readiness closure

```text
F0 accepted authority baseline             COMPLETE
F1 coverage                                COMPLETE / 16 goals / 78 ops
F2 material surfaces                       COMPLETE / 36
F3 Screen Contracts                        COMPLETE / 36 READY
F4 Navigation/Data Graph                   COMPLETE
F5 functional wireframes                   COMPLETE CANDIDATE
F6 Material Interaction Ledger             COMPLETE CANDIDATE
F7 bidirectional trace                     COMPLETE / 78 / 78
F8 findings                                COMPLETE / 0 unresolved MATERIAL
F9 frontend readiness                      COMPLETE CANDIDATE

stable Product paths                       10 / 10
idempotent creations                       10 / 10
ETag domains                               13 / 13
exact-byte resources                        4 / 4
frontend semantic owner                     0
frontend Authorization engine               0
parallel server-truth store                  0
screen-shaped API                            0
operation 79                                absent
```

**Frontend Implementation Readiness = COMPLETE CANDIDATE.**

This does not authorize Product implementation and does not begin T12.

## 18. Implementation-node linkage

Final node closure is owned by `t11-implementation-program.md`:

```text
S1 → WF-00/01 + WF-16..21
S2 → WF-22..25 configuration base
S3 → WF-02..10 + WF-11 authoring + WF-14 + concrete Template role
S4 → WF-11 governance + WF-12/13 + WF-04 release enrichment + WF-14 enrichment
S5 → WF-05 obsolescence + WF-13/14 enrichment + WF-15
```

No S node closes as "backend complete, frontend later".

## 19. Review pack

Default bounded T11 review route after operator approval:

```text
docs/roadmap.md
→ docs/work/current/t11-implementation-program.md
→ docs/work/current/t11-frontend-blueprint.md
→ docs/decisions/responsible-owner-selection-read.md
```

Only when a concrete frontend question needs detail:

```text
docs/work/current/t11-wireframes.md
docs/work/current/t11-interaction-ledger.md
```

This keeps the default authority pack bounded while retaining exact implementation detail.
