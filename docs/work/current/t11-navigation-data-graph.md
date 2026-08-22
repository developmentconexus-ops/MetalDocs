# T11 — MetalDocs Navigation / Data Graph

> **TEMPORARY T11 CANDIDATE WORK / BRANCH-ONLY.** F4 derives the complete browser navigation/data-resolution graph from the closed F1–F3 frontend-readiness artifacts and accepted Product/T6/T8-E/T8-F authority. It creates no Product route, API operation, semantic owner or backend authority.

## 1. Purpose

A frontend link/control is implementation-ready only when the target can be resolved without guessing:

```text
source surface
→ exact server-returned target identity/reference
→ stable browser route / same-route surface
→ exact target initial read
→ current 401/403/404/state behavior
```

Client-only DOM state, History scans, Audit, provider identity, guessed UUIDs and stale projection fields may never substitute for current-resource identity resolution.

## 2. Fixed route envelope

Stable Product paths remain exactly:

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

Browser AuthN integration remains outside the Product tree/census:

```text
/auth/login
/auth/callback
```

No F4 decision adds a Product path or application operation.

## 3. Frontend-only route-local navigation state

Large composed routes need deterministic landing/switching without inventing nested Product routes.

T11 selects these **browser-only presentation parameters**:

```text
/work
  ?lane=authoring|governance
  default authoring

/admin/organization
  ?section=company|users|areas|groups
  default company

/admin/access
  ?section=memberships|roles
  default memberships

/admin/document-governance
  ?section=document-types|templates
  default document-types
```

Laws:

```text
these parameters are SPA presentation/navigation state only
never sent to /api/v1 unless an owning operation independently defines the same name
never become Product state or Authorization input
unknown value → normalize/replace to route default; no Product error is manufactured
entity-row selection inside a section is ephemeral local UI state, not a new route baseline
```

The fixed `/documents` query vocabulary remains the accepted T8-E application read vocabulary:

```text
q
document_type_id
area_id
responsible_owner_user_id
status
cursor
limit
```

No frontend-only generic filter/sort parameter is added.

## 4. Route-entry graph

| Browser route | Landing surface | Required first authoritative read(s) | Resolution law |
|---|---|---|---|
| any SPA entry | APP-01 / APP-02 | `getSession` | 200→authenticated shell; 401→APP-01; never infer session from client storage |
| `/documents` | LIB-01 | `listDocuments` using exact URL filter/cursor semantics | server-filtered DocumentPage is sole catalog truth |
| `/documents/:document_id` | OFF-01 | `getDocument(document_id)` | 404 means absent/non-disclosable; no fallback resolver |
| `/documents/:document_id/work` | DW-01 or DW-04 | `getDocument(document_id)` → disclosed `open_revision` | `state=draft`→DRAFT reads; `state=submitted`→Revision/Submission reads; absent/404 never triggers History search |
| `/documents/:document_id/history` | HIS-01 | `getDocumentHistory(document_id)` | historical lens only; never resolves current Work/obsolescence |
| `/work?lane=authoring` | WRK-01 | `listAuthoringWork` | actor-relevant projection; target revalidates live truth |
| `/work?lane=governance` | WRK-02 | `listGovernanceWork` | actor-relevant projection; target case rereads canonical truth |
| `/work/governance/:attempt_id` | GOV-01 | `getGovernanceAttempt(attempt_id)` | exact case; 404/denial never repaired through Work projection |
| `/audit` | AUD-01 | `listAuditEvents` | evidence page only; no generic current-resource resolver |
| `/admin/organization?section=company` | ORG-01 | `getCompany` | 403/404/server truth controls access |
| `/admin/organization?section=users` | ORG-02 | `listUsers` | selected User ids originate only from returned UserReferences/results |
| `/admin/organization?section=areas` | ORG-06/07 | `listAreas` | selected Area id originates from AreaPage/create result |
| `/admin/organization?section=groups` | ORG-08 | `listGroups` | selected Group id originates from GroupPage/create result |
| `/admin/access?section=memberships` | ACC-01 | `listGroups`; after selection `listGroupMembers` | membership target ids come from admitted User/Group references |
| `/admin/access?section=roles` | ACC-02 | `listRoles` + `listRoleAssignments`; admitted User/Group/Area reads for selectors | no client role/scope inference beyond returned fixed catalog |
| `/admin/document-governance?section=document-types` | DGV-01..05 | `listDocumentTypes` | selected type id originates from list/create result |
| `/admin/document-governance?section=templates` | DGV-06 | `listTemplateConfigurations` | selected Document id originates from bounded template configuration projection |

## 5. Cross-route navigation edges

### N01 — Shell → Library

```text
APP-02
→ /documents
→ listDocuments
```

Navigation presence is not proof of `document.read_effective`. Server response owns disclosure/permission behavior.

### N02 — Shell → My Work

```text
APP-02
→ /work?lane=authoring (default)
→ listAuthoringWork
```

User may switch to governance lane; `listGovernanceWork` supplies its own projection.

### N03 — Shell → Audit

```text
APP-02
→ /audit
→ listAuditEvents
```

Because SessionView contains no Permission snapshot, shell visibility is UX only. A 403 remains authoritative.

### N04 — Shell → Administration sections

```text
APP-02
→ one of the three accepted Admin stable routes
→ selected section's first read
```

No frontend permission matrix decides admission. Server 403/404 is final.

### N05 — Library row → Document Official

```text
LIB-01 DocumentSummary.document.document_id
→ /documents/:document_id
→ getDocument(document_id)
```

The list row is a navigation reference only. Current official truth is reread.

### N06 — Successful Document create → Document Work

```text
LIB-02 createDocument
→ CreateDocumentResult.document_id + revision_id
→ /documents/:document_id/work
→ getDocument(document_id)
→ disclosed open_revision must resolve current Work
→ getRevision + DRAFT reads
```

**F4 correction:** the accepted T6 create journey opens Document Work after creation. An earlier F3 shorthand that pointed successful create at OFF-01 is superseded by this graph. `revision_id` from the result is useful evidence but does not create a second route resolver; direct/reload behavior remains `document_id → getDocument.open_revision`.

If the post-create current read cannot disclose/resolve Work, server truth wins; client does not force the returned revision into an editor.

### N07 — Document Official → existing Document Work

```text
OFF-01 getDocument.open_revision present
→ /documents/:document_id/work
→ getDocument again
→ branch on current open_revision.state
```

The reference can become stale between source and target. Target reread controls.

### N08 — Document Official → create next Revision → Document Work

```text
OFF-04 no disclosed open_revision + admitted action
→ createDocumentRevision(document_id)
→ CreateRevisionResult
→ /documents/:document_id/work
→ getDocument reread
```

No History lookup, client lifecycle predicate or returned-id shortcut replaces the route resolver.

### N09 — Document Official → Document History

```text
current document_id
→ /documents/:document_id/history
→ getDocumentHistory(document_id)
```

History is an explicit user destination, never a background current-state resolver.

### N10 — My Work authoring row → Document Work

```text
WRK-01 WorkAuthoringItem.document.document_id
→ /documents/:document_id/work
→ getDocument
→ current open_revision resolution
```

A stale projection row may land on 404/no-current-work; that is truthful and triggers Work projection refetch, not mutation from stale row data.

### N11 — My Work governance row → Governance Case

```text
WRK-02 WorkGovernanceItem.governance_attempt_id
→ /work/governance/:attempt_id
→ getGovernanceAttempt
```

Stale actor relevance never grants case access.

### N12 — Governance Case → Document Official

Both governance subjects carry `document:DocumentReference`:

```text
GOV-01 subject.document.document_id
→ /documents/:document_id
→ getDocument
```

This is optional context navigation. It never changes the immutable governed subject. For an obsolescence case, current Document Official may have changed after the case ended; GOV-01 remains the case authority.

### N13 — Document History item → Governance Case when exact id exists

Only history union branches that actually expose `governance_attempt_id` may offer this link:

```text
history item.governance_attempt_id
→ /work/governance/:attempt_id
→ getGovernanceAttempt
```

A later 404/non-disclosure is shown honestly. History never invents an attempt id.

### N14 — Document History item → bounded inline historical inspection

History item ids may drive admitted supporting reads **inside the History lens**, without manufacturing routes:

```text
revision_id              → getRevision
submission_id            → getSubmission / getSubmissionSource when authorized
release_id               → getRelease / getReleaseSource when authorized
official_rendition_id    → getOfficialRenditionContent when authorized
obsolescence request_id  → getObsolescenceRequest when authorized
```

These are historical inspection actions, not current-resource resolution.

## 6. Same-route material data-resolution edges

### OFF-02 — Official content

```text
OFF-01 DocumentOfficialView.official.release_id
→ getRelease
→ ReleaseView representation
→ getReleaseSource
→ optional official_rendition_id → getOfficialRenditionContent
```

Provider location is never used as a resolver.

### OFF-03 — Responsible-owner replacement

The operator-approved T8-E-RO precision closes candidate discovery:

```text
OFF-01 getDocument
→ DocumentOfficialView.responsible_owner_candidates?   # complete D4-eligible refs when current owner.manage ALLOW

OFF-03 concurrency
→ getDocumentResponsibleOwner
→ ResponsibleOwnerView + strong ETag

selection
→ candidate UserReference.user_id
→ replaceDocumentResponsibleOwner(..., If-Match=current owner ETag)
```

Candidate projection and concurrency representation deliberately remain separate.

### OFF-05 — Active obsolescence

```text
OFF-01 active_obsolescence_request_id present
→ getObsolescenceRequest(request_id)
```

Creation result may return `request_id` and, for governance-pending outcome, `governance_attempt_id`. Baseline stays on Document Official and refetches current truth; it does **not** auto-navigate the initiator into a Governance Case whose current participation may differ. Actor-relevant approvers reach the case through WRK-02.

### DW route resolver

Direct/reload law:

```text
getDocument.open_revision absent
  → no current Work subject may be synthesized

open_revision.state=draft
  → getRevision
  → getRevisionDraft + ETag
  → getRevisionDraftSource
  → DW-01/02/03

open_revision.state=submitted
  → getRevision
  → current_submission_id must exist
  → getSubmission
  → getSubmissionSource as needed
  → DW-04
```

The client never calls `getRevisionDraft` merely because a previously cached Revision was DRAFT.

### Governance Case subject content

```text
subject.kind=submission
  → subject.submission_id
  → getSubmissionSource when exact governed content is needed/authorized

subject.kind=obsolescence
  → immutable case subject already carries document + exact target_revision + reason
  → optional current Document context may use N12; current Document never rewrites the case subject
```

### Admin Organization selections

```text
ORG-02 User selection
  listUsers/CreateUserResult → user_id
  → getUser
  → ORG-03 getUserProfile
  → ORG-04 getUserProviderBinding
  → ORG-05 getUserEligibility

ORG-06/07 Area selection
  listAreas/CreateAreaResult → area_id
  → getArea and independent getAreaLifecycle

ORG-08 Group selection
  listGroups/CreateGroupResult → group_id
  → getGroup
```

Independent ETag domains remain separate even when one panel composes them.

### Admin Access selectors

Current static role bundles make the accepted governance-admin route capable of using Organization reference reads, but the frontend does not encode Permission implication.

```text
ACC-01
  listGroups → group_id
  listUsers → candidate user_id references when adding a member
  listGroupMembers(group_id) → current members

ACC-02
  listRoles → exact RoleCode + allowed_scope_kinds
  listUsers → User subject refs
  listGroups → Group subject refs
  listAreas → Area scope refs
  listRoleAssignments → current assignment ids/state
```

Every write still revalidates target existence/eligibility and current access server-side.

### Document Governance selectors

```text
DGV-01/02/03/04/05
  listDocumentTypes/CreateDocumentTypeResult → document_type_id

DGV-03 governance route named-user selector
  listUsers → UserReference

DGV-03 governance route group selector
  listGroups → GroupReference identity

DGV-04 eligible Template selector
  listTemplateConfigurations → DocumentReference candidates/current eligibility projection

DGV-05 numbering preview
  selected document_type_id + selected/required area_id → getDocumentTypeNumberingPreview

DGV-06 template-role management
  listTemplateConfigurations → document_id
  → getDocumentTemplateRole + ETag
  → replaceDocumentTemplateRole
```

No general reference-data endpoint is created.

## 7. Library filter identity law

`listDocuments` admits typed ids for Area, DocumentType and responsible owner, but Launch has no general low-privilege reference-directory API for ordinary readers.

F4 therefore deliberately does **not** invent global dropdown directories.

Accepted smallest browser behavior:

```text
q + status
  explicit ordinary Library controls

area/document-type/responsible-owner id filters
  may be activated from UserReference/AreaReference/DocumentTypeReference already disclosed in visible DocumentSummary/DocumentOfficial context
  remain in URL once activated
  may always be cleared
```

Example:

```text
click visible responsible-owner reference on a Library row
→ set responsible_owner_user_id=<returned user_id>
→ rerun listDocuments
```

A filter URL loaded without a locally-known display label still executes using the exact opaque id; the UI may label it generically as an active owner filter until returned rows supply a disclosed display reference. It must not call Admin `listUsers` merely to decorate the filter.

This is intentionally less general than inventing a reference-data platform. If real user evidence proves arbitrary owner/type/area selection is required independently of already-disclosed references, that is a new concrete read consumer and triggers the smallest T6/T8-E decision rather than frontend guessing.

## 8. Mutation successor graph

| Mutation | Immediate authoritative successor |
|---|---|
| `endSession` | clear session presentation → APP-01 |
| `createDocument` | N06 → Document Work, not Document Official |
| `replaceDocumentResponsibleOwner` | replace owner query from response; refetch `getDocument` + affected Work projections; remain OFF route |
| `createDocumentRevision` | N08 → Document Work after current resolver reread |
| `createObsolescenceRequest` | remain OFF-05; refetch Document/request/current History; synchronous obsolete updates OFF truth |
| `withdrawObsolescenceRequest` | remain OFF-05; refetch Document/request/History |
| `updateRevisionDraft` | replace DocumentWorkView/ETag; remain DW-01 |
| upload completion only | no Product navigation; still not WorkingContent attachment |
| DRAFT source attach | replace DocumentWorkView/ETag; remain DW-01 |
| `createSubmission` governance/rendition pending | remain Work; refetch Revision/Submission/Official/My Work/History → DW-04 |
| `createSubmission` released | refetch `getDocument`; navigate Document Official because no open Work is assumed client-side |
| `withdrawSubmission` / `cancelRevision` | refetch `getDocument`; if current open_revision disclosed stay/re-enter Work; otherwise navigate Official |
| `createGovernanceFeedback` | remain GOV-01; replace/refetch feedback/case |
| `recordGovernanceStepDecision` | remain GOV-01 read-only/current; refetch case/Work/Document/History as relevant; no client publish transition |
| Organization/Admin replacements | remain owning section; replace exact returned ETag representation + narrowly refetch affected lists/dependent security reads |
| create User/Area/Group/DocumentType/RoleAssignment | remain owning section; use returned id to select/fetch new item; never fabricate entity body |
| Group member add/remove | remain ACC-01; refetch member page/security-dependent reads |
| Template/governance configuration replacements | remain DGV section; replace exact ETag domain + narrowly refetch affected config/creation options where semantic truth changed |

No mutation success invents lifecycle truth optimistically.

## 9. Authentication / permission / stale-navigation law

Cross-cutting response behavior:

```text
401 auth.unauthenticated
→ invalidate session presentation
→ APP-01
→ optional intended return route stays browser navigation context only

403 permission.denied
→ stay at destination context and show denied action/lens
→ do not translate to 404 or hide as success

404 notfound.resource
→ absent or non-disclosable
→ do not infer which

412 precondition.resource_changed
→ stay on editing surface
→ preserve local form selection/input
→ refetch exact current ETag representation
→ explicit reconcile

412 precondition.draft_changed
→ same but DRAFT-specific; never automatic merge/LWW

stale projection navigation
→ destination current read wins
→ source projection is invalidated/refetched as needed
```

## 10. F4 findings and adjudication

### F4-F01 — Create successor target — T11 correction / CLOSED

```text
finding
  earlier LIB-02 Screen Contract shorthand said successful create routes by document_id to OFF-01

accepted Product journey
  createDocument → open Document Work

correction
  N06 now routes create success to /documents/:document_id/work
  current resolver remains getDocument.open_revision

upstream reopen
  none
```

### F4-P01 — Library arbitrary filter directory — DELIBERATELY ABSENT

No accepted requirement currently demands an ordinary-reader global User/Area/DocumentType reference directory merely to populate catalog filters. F4 uses already-disclosed references and preserves the existing exact id filters.

Reopen only on concrete UX evidence that independent arbitrary selection is required and row/detail-driven filter activation is materially insufficient.

### Upstream material result

```text
new Product/T6/T8-E blocker discovered by F4     0
operation 79 required                              no
new stable SPA route required                      no
screen-shaped API required                         no
```

The only F4 correction is T11-local N06 alignment.

## 11. F4 closure proof

```text
F2 material surfaces                    36 / 36 attached to a route/data context
stable Product routes                   10 / 10 preserved
browser AuthN routes                     2 / 2 preserved outside census
cross-route material edges              14 explicitly resolved
same-route material resolver families    covered
mutation successor paths                 covered
Admin selector identities                backend-traceable
Document Work direct reload              current-server-resolved
History current-resource fallback        absent
Audit current-resource fallback          absent
provider identity as Product resolver    absent
new application operations               0
application operations                   78
operation 79                              absent
unresolved MATERIAL F4 finding            0
F4 status                                COMPLETE
```

## 12. Handoff to F5

Wireframes may now begin, but only from the closed F1→F4 pack.

Each F5 wireframe must visibly realize:

```text
one or more F2 surface ids
its F3 Screen Contract
its F4 route/data resolver and successor paths
all material safe-action state variants that change the user's next action
```

A wireframe that needs a data field, identity, route or write not traceable through F3/F4 is a new finding and stops that wireframe before any API invention.
