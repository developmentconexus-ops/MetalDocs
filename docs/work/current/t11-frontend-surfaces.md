# T11 — MetalDocs Material Frontend Surface Inventory

> **TEMPORARY T11 CANDIDATE WORK / BRANCH-ONLY.** F2 output derived from accepted Product/T6/T8-E/T8-F authority and the closed F1 Coverage Matrix. Surface IDs are implementation-planning identifiers only: they are not Product routes, API resources, semantic owners or durable frontend taxonomy.

## 1. Purpose

Enumerate every **material user decision context** that must receive a Screen Contract before wireframing.

A surface is split only when at least one materially changes:

```text
semantic truth presented
safe user action
owner / application operation receiving a write
target identity needed for navigation
ETag / idempotency / exact-byte mechanics
lifecycle or disclosure context
failure/recovery path
editor/viewer mode
```

A separate surface is **not** created for cosmetic loading variants, visual styling, simple responsive reflow or component boundaries.

One SPA route may contain many material surfaces. Multiple material surfaces may later be shown in one composed wireframe when the interaction remains clear.

## 2. Inventory result

```text
stable SPA Product routes               10
browser AuthN integration routes         2 outside Product tree/census
material frontend surfaces              36
new Product routes                       0
new application operations               0
operation 79                              absent
unresolved Product/T8-F coverage gap      0 at F2 entry
```

The 36 surface IDs below are derived planning handles, not a requirement for 36 separately navigable pages or 36 final wireframe images.

## 3. Application shell / authentication — 2 surfaces

### APP-01 — Session gate / unauthenticated entry

```text
route context      SPA entry / any protected route
read               getSession
browser integration GET /auth/login, GET /auth/callback outside application census
truth              authenticated current User + session-bound CSRF material, or 401
safe actions       begin login; return to application only after accepted callback/session creation
tranche            S1
```

Material variants:

```text
no valid session / 401 → unauthenticated entry + login action
valid session          → hand off to APP-02
OIDC callback/protocol failure → no ApplicationSession; sanitized failure/retry entry
provider unavailable   → login unavailable; never synthesize local credentials
```

Must not become a password/login Product surface, provider-claim authority or local session truth.

### APP-02 — Authenticated application shell / stable navigation

```text
route context      all stable SPA routes
read               getSession cached as server state
write              endSession
truth              SessionView only; no effective-Permission snapshot
safe actions       navigate stable Product spaces; end session
tranche            S1
```

Material variants:

```text
session valid
session lost/expired during navigation → invalidate session presentation and return to APP-01
endSession success → cleared session presentation
```

Navigation presence is UX only; it never proves permission.

## 4. Library — 2 surfaces

### LIB-01 — Official document discovery

```text
route              /documents
read               listDocuments
truth              DocumentPage / DocumentSummary
URL state          q + admitted filters + cursor context
write              none
tranche            S3
```

Material variants:

```text
known empty result
effective catalog rows
explicit obsolete/cancelled query when admitted
cursor continuation
403 on history-sensitive catalog status request
```

Rows navigate only with returned `document_id`. No DRAFT/SUBMITTED row is manufactured as official discovery truth.

### LIB-02 — Create Document interaction

```text
route              /documents
reads              getDocumentCreationOptions; getDocumentTypeNumberingPreview when requested
write              createDocument
truth              options/preview are guidance; create result is authority
state              local creation form + query-backed option/preview reads
wire               CSRF + one Idempotency-Key per logical create
tranche            S3
```

Material variants:

```text
Area/DocumentType selection changes option truth
responsible-owner candidates present/absent exactly per server representation
eligible Template choices empty/non-empty
numbering preview available / validation or non-disclosable failure
ambiguous transport retry → same logical key
successful create → returned document_id/revision_id
```

Preview never reserves or becomes final code authority.

## 5. Document Official — 5 surfaces on one stable route

### OFF-01 — Document Official core overview

```text
route              /documents/:document_id
read               getDocument
truth              DocumentOfficialView
tranche            S3 base; progressively enriched by S4/S5/S6
```

Material server states that the screen contract/wireframe must distinguish:

```text
status = draft | submitted | effective | obsolete | cancelled
official present only for effective|obsolete
open_revision disclosed / absent
active_obsolescence_request_id disclosed / absent
```

Absence of disclosure-safe references never proves semantic non-existence to an unauthorized caller.

### OFF-02 — Official content presentation / exact-byte viewer

```text
route              /documents/:document_id
reads              getRelease; getReleaseSource; getOfficialRenditionContent as referenced
truth              ReleasedRevisionView / ReleaseView + exact semantic byte resources
tranche            S5
```

Material viewer modes:

```text
SourceOnly + DOCX → read-only DOCX adapter on exact Release source
SourceOnly + PDF  → read-only PDF on exact Release source
RequireOfficialRendition(PDF) → OfficialRendition PDF primary + exact Release source separately available/labeled
no official release → viewer absent, not placeholder Product truth
```

Provider location is never displayed/parsed as Product identity.

### OFF-03 — Responsible owner management

```text
route              /documents/:document_id
reads              getDocumentResponsibleOwner
write              replaceDocumentResponsibleOwner
truth              ResponsibleOwnerView + exact response ETag
state              local replacement selection bound to loaded ETag
tranche            S3
```

Material variants:

```text
read/manage usable
server-denied/non-disclosable
stale ETag → preserve selection, refetch current truth, explicit reconciliation
eligible-target conflict/race
```

Relationship change never grants permission by itself.

### OFF-04 — Create or enter open Revision

```text
route source       /documents/:document_id
read source        getDocument.open_revision
write              createDocumentRevision only when no disclosed current Work target and command is admitted
route target       /documents/:document_id/work
wire               idempotent create when creation occurs
tranche            S4
```

Material variants:

```text
open_revision disclosed → enter existing Work
open_revision absent + create allowed → idempotent create-next then enter Work
open_revision absent + action unavailable/denied → no frontend inference
concurrent state change → server conflict/current truth wins
```

History is never used to discover the current open Revision.

### OFF-05 — Obsolescence management

```text
route              /documents/:document_id
read source        getDocument.active_obsolescence_request_id; getObsolescenceRequest when disclosed
writes             createObsolescenceRequest; withdrawObsolescenceRequest
truth              current official target + ObsolescenceRequestView/create/withdraw results
wire               idempotent creation; CSRF on withdrawal
tranche            S6
```

Material variants:

```text
no disclosed active request + admitted create
governance_pending active request
synchronous obsolete result under NoHumanApproval
returned / withdrawn / completed request
withdrawal available vs unavailable
stale/non-disclosable active routing reference
```

No fake human Step is shown for NoHumanApproval.

## 6. Document History — 1 surface

### HIS-01 — Semantic Document history

```text
route              /documents/:document_id/history
read               getDocumentHistory
supporting reads   only explicitly referenced admitted facts/content
truth              DocumentHistoryItem closed union
write              none by virtue of History
tranche            S3 base; regression/enrichment S4→S6
```

Material item classes include revision, submission, governance decision/feedback, withdrawal/cancellation, release/rendition and obsolescence facts as they become reachable.

Historical content inspection may reuse the accepted read-only exact-content viewer mechanism, but History never becomes a current-resource resolver or a second lifecycle state machine.

## 7. My Work — 2 material lanes on one stable route

### WRK-01 — Authoring work lane

```text
route              /work
read               listAuthoringWork
truth              WorkAuthoringPage projection
write              none
route target       /documents/:document_id/work
tranche            S4
```

Material variants: empty / paged actor-relevant work; stale row whose current target is no longer disclosable/active must resolve truthfully at target read, not be treated as authority.

### WRK-02 — Governance work lane

```text
route              /work
read               listGovernanceWork
truth              WorkGovernancePage projection
write              none
route target       /work/governance/:attempt_id
tranche            S5
```

This lane appears only when its Governance Case target is implemented. Projection state never owns governance lifecycle.

## 8. Document Work — 4 surfaces on one stable route

### DW-01 — Current DRAFT authoring / inspection

```text
route              /documents/:document_id/work
resolver           getDocument → disclosed open_revision on route entry/direct reload
reads              getRevision; getRevisionDraft; getRevisionDraftSource
write              updateRevisionDraft
truth              RevisionView + DocumentWorkView + exact DRAFT ETag + exact source bytes
tranche            S4
```

Material editor modes:

```text
DRAFT DOCX → editable DOCX adapter
DRAFT PDF  → read-only inspect/replace
```

Material variants:

```text
clean local form/editor
locally dirty buffer
save success with new authoritative view/ETag
412 precondition.draft_changed → preserve local input + refetch + explicit reconciliation
open Revision no longer DRAFT → current server state replaces editor authority
```

### DW-02 — Replace source / upload-admission workflow

```text
route              /documents/:document_id/work
write sequence     startRevisionDraftUpload
                   → opaque provider PUT with required headers
                   → completeRevisionDraftUpload
                   → updateRevisionDraft(source_upload_id, current If-Match)
truth              server admission + resulting DocumentWorkView
local state         intended local bytes / transient upload progress
tranche             S4
```

Material variants:

```text
allocation obtained
provider upload in progress / transport failure
provider PUT success but not yet READY
completion/admission success
410 upload expired → preserve intended bytes, obtain new allocation, re-upload
422 validation.content_invalid
DRAFT ETag changed before attachment → explicit reconciliation; never revive old authority
```

Provider success never means WorkingContent attached.

### DW-03 — Submit DRAFT command

```text
route              /documents/:document_id/work
write              createSubmission
reads              current DRAFT/ETag before command; resulting Submission via returned identity/current Revision path
wire               CSRF + Idempotency-Key + exact semantic DRAFT If-Match
truth              SubmissionCreateResult
tranche            S4
```

Material success outcomes:

```text
governance_pending
rendition_pending
released
```

Material failures include stale DRAFT, malicious/invalid content, scanner dependency unavailable, semantic conflict and changed-request same-key rejection. Ambiguous transport retry preserves the same logical key/condition.

### DW-04 — Submitted Revision / Submission state and termination

```text
route              /documents/:document_id/work
reads              getRevision.current_submission_id; getSubmission; getSubmissionSource
writes             withdrawSubmission; cancelRevision when admitted
truth              SubmissionView orthogonal human/representation gates + termination/release identity
tranche            S4, with later server-driven Release outcome visible through current reads
```

Material variants:

```text
human gate pending/satisfied/not required
representation gate pending/satisfied/not required
release_id present
termination = returned_for_changes | withdrawn | revision_cancelled
withdraw available/unavailable
cancel available/unavailable
```

Submitted content is immutable/read-only; Governance never edits it.

## 9. Governance Case — 3 surfaces on one stable route

### GOV-01 — Governance Case overview / exact subject

```text
route              /work/governance/:attempt_id
read               getGovernanceAttempt; listGovernanceFeedback continuation
truth              GovernanceCaseView
subject             immutable Submission or Obsolescence subject
content             exact admitted Submission source when subject kind requires it
tranche            S5
```

Material variants:

```text
attempt active | completed | returned | withdrawn | cancelled
subject submission | obsolescence
Step pending | active | decided
allowed_actions empty or admitted subset
feedback first page + cursor continuation
```

`allowed_actions` are hints only.

### GOV-02 — Add governance feedback

```text
parent             GOV-01
write              createGovernanceFeedback
wire               idempotent creation
state              local message form
truth              returned feedback identity + refreshed case/feedback read
tranche            S5
```

Material variants include accepted feedback, validation failure, conflict because case changed, denied action and ambiguous transport retry with same key.

### GOV-03 — Governance Step decision

```text
parent             GOV-01
reads              getGovernanceStepDecision as needed/current case Step truth
write              recordGovernanceStepDecision
input variants      accept | return_for_changes(reason required)
truth              GovernanceDecisionView
tranche            S5
```

Material variants include first decision, exact repeat, `state.governance_step_already_decided`, current AuthZ loss and case-state conflict. Decision is never Release/publish.

## 10. Admin / Organization — 8 material surfaces on one stable route

All live under `/admin/organization`; surface IDs do **not** create nested durable Product routes.

### ORG-01 — Company settings

```text
read/write         getCompany / replaceCompany
concurrency        independent Company ETag
tranche            S1
```

### ORG-02 — User directory + atomic User creation

```text
reads              listUsers; searchProviderSubjects for creation preflight
write              createUser
truth              UserPage + bounded provider subject options
wire               idempotent create
tranche            S1
```

Creation atomically establishes enabled User + profile + binding. Provider subject refs remain opaque.

### ORG-03 — User profile subresource

```text
reads/writes       getUserProfile; replaceUserProfile; deleteUserProfile
concurrency        profile-specific If-Match / If-None-Match:* matrix
tranche            S1
```

Material variants: present/absent profile, replace, recreate, erase, stale/current conflict. Profile contact data is never login identity.

### ORG-04 — Provider subject binding

```text
reads              getUserProviderBinding; searchProviderSubjects
write              replaceUserProviderBinding
concurrency        binding-specific ETag
security effect    replacement revokes required existing sessions per accepted authority
tranche            S1
```

Material variants: current binding, provider search/refinement, stale ETag, subject conflict/non-disclosable, post-replacement session consequence.

### ORG-05 — User eligibility / offboarding / re-enable

```text
read/write         getUserEligibility / replaceUserEligibility
concurrency        eligibility-specific ETag
states             enabled | disabled
tranche            S1
```

Disabling is a high-impact atomic security transition: sessions, memberships and direct assignments are torn down as accepted. Re-enable never resurrects them. Confirmation/presentation must not imply otherwise.

### ORG-06 — Area identity directory + create/edit

```text
reads              listAreas; getArea
writes             createArea; replaceArea
truth              immutable code + mutable name
wire               idempotent create; Area ETag for edit
tranche            S1
```

### ORG-07 — Area lifecycle

```text
read/write         getAreaLifecycle / replaceAreaLifecycle
states             active | retired
concurrency        independent lifecycle ETag
tranche            S1
```

This remains separate from Area identity editing because T8-E deliberately gives it a distinct concurrency domain.

### ORG-08 — Group identity directory + create/edit/delete

```text
reads              listGroups; getGroup
writes             createGroup; replaceGroup; deleteGroup
wire               idempotent create; Group ETag for replacement
tranche            S1
```

Deletion conflict when live dependency exists must be distinguished from ordinary validation/not-found. Membership administration is not here; it is ACC-01 because it changes effective access.

## 11. Admin / Access — 2 material surfaces on one stable route

### ACC-01 — Group membership administration

```text
route              /admin/access
reads              listGroups/getGroup as admitted references; listGroupMembers; admitted User references
writes             addGroupMember; removeGroupMember
truth              Organization-owned membership with access-sensitive effect
tranche            S1
```

Material variants: paged members, add existing relation success semantics, remove absent relation idempotent-like success, parent non-disclosable, security-state conflict.

### ACC-02 — Role catalog + RoleAssignment administration

```text
route              /admin/access
reads              listRoles; listRoleAssignments; admitted User/Group/Area references
writes             createRoleAssignment; deleteRoleAssignment
truth              fixed product-owned roles/Permission bundles + current assignments
wire               idempotent grant creation
tranche            S1
```

Material variants: subject User|Group; scope Company|Area constrained by selected role; grant conflict; revoke absent/not-found. Role bundles are inspectable but not editable.

## 12. Admin / Document Governance — 6 material surfaces on one stable route

All live under `/admin/document-governance`; no new Product route family is introduced.

### DGV-01 — DocumentType directory + create

```text
reads              listDocumentTypes
write              createDocumentType
create input        base + initial GovernancePolicy + RepresentationPolicy
wire               idempotent create
tranche            S2
```

Creation cannot defer governance/representation to an implicit default because none is accepted.

### DGV-02 — Existing DocumentType base configuration

```text
read/write         getDocumentType / replaceDocumentType
concurrency        base DocumentType ETag
tranche            S2
```

Material conflict: code/numbering-scope change after first committed Document can be rejected as state conflict.

### DGV-03 — Governance + representation policy configuration

```text
read/write         getDocumentTypeGovernance / replaceDocumentTypeGovernance
concurrency        independent governance ETag
modes              no_human_approval | use_governance_route
representation     source_only | require_official_rendition(pdf)
tranche            S2
```

Route Step order is semantic; no generic workflow builder is inferred.

### DGV-04 — Eligible Template set configuration

```text
read/write         getDocumentTypeEligibleTemplates / replaceDocumentTypeEligibleTemplates
concurrency        independent eligibility ETag
truth              set of stable Document references
tranche            S2
```

Empty is valid. Template documents remain Controlled Documents, not a peer Template product.

### DGV-05 — Numbering preview interaction

```text
read               getDocumentTypeNumberingPreview
input              DocumentType + optional Area when numbering scope requires context
truth              NumberingPreviewView {reservation:false}
write              none
tranche            S2; also consumed by LIB-02 creation context
```

The UI must explicitly avoid communicating reservation/finality.

### DGV-06 — Template configuration / Document template-role administration

```text
reads              listTemplateConfigurations; getDocumentTemplateRole
write              replaceDocumentTemplateRole
truth              configuration projection + concrete Document template-role ETag
tranche            S2 list/projection base; S3 concrete Document role enrichment after Documents exist
```

Material variants: no effective revision vs effective revision title present; role on/off; stale ETag; eligible-document-type references. Configuration access does not imply content/history access.

## 13. Audit — 1 surface

### AUD-01 — Audit evidence list

```text
route              /audit
read               listAuditEvents
truth              AuditEventPage / closed AuditEventView union
URL/server paging  cursor + limit only
write              none
tranche            S6
```

Material variants: empty/paged visible event set; continuation; denied access. No search/filter DSL is inferred. Audit never becomes current business-state authority.

## 14. Cross-cutting material state patterns — attached, not standalone surfaces

These patterns require explicit representation in the Screen Contracts/wireframes that can reach them, but do not justify separate Product routes or global state owners.

### X-01 — Authentication loss

```text
401 auth.unauthenticated → invalidate session presentation → APP-01
```

### X-02 — Permission/disclosure

```text
403 permission.denied → denied action/context
404 notfound.resource → absent or non-disclosable; client must not infer which
```

### X-03 — CSRF recovery

```text
permission.csrf_failed
→ re-bootstrap getSession/csrf_token
→ retry same logical unsafe command only when accepted-safe
→ preserve same idempotency/conditional semantics
```

### X-04 — OCC reconciliation

```text
precondition.resource_changed → preserve local input + current refetch + explicit reconcile
precondition.draft_changed    → same, strictly DRAFT-specific; never automatic merge/LWW
```

### X-05 — Idempotent ambiguous retry

```text
same logical command / ambiguous transport outcome → same Idempotency-Key
changed/new semantic command → new key
validation.idempotency_key_reused → do not reinterpret as ordinary validation retry
```

### X-06 — State conflict

`state.conflict` and `state.governance_step_already_decided` are operation-specific state changes, not generic form errors.

### X-07 — Dependency failure

`dependency.*` is sanitized mechanism unavailability. Parent surface decides whether current user can safely retry/continue; raw provider errors are never shown.

### X-08 — Exact-content integrity failure

`internal.content_integrity` prevents successful content presentation; viewer must not display partial/corrupt success as authoritative content.

### X-09 — Cursor continuation

Seek cursor is opaque navigation/server-query context. No offset/page-number/total-count authority is manufactured.

## 15. Surface-to-tranche summary

```text
S1
  APP-01 APP-02
  ORG-01..ORG-08
  ACC-01 ACC-02

S2
  DGV-01..DGV-05
  DGV-06 base projection

S3
  LIB-01 LIB-02
  OFF-01 OFF-03
  HIS-01 base
  DGV-06 concrete Document role enrichment

S4
  OFF-04
  WRK-01
  DW-01..DW-04
  HIS-01 authoring/submission enrichment

S5
  OFF-02
  WRK-02
  GOV-01 GOV-02 GOV-03
  HIS-01 governance/release enrichment

S6
  OFF-05
  AUD-01
  HIS-01 obsolescence enrichment
```

Every actionable surface has its real backend target in the same or an already-closed tranche.

## 16. F2 completeness challenge

### Human-goal coverage

All 16 F1 human goals have at least one material surface.

### Operation-consumer coverage

F2 does not change the accepted T8-F consumer reconciliation. The 78-operation mapping remains exactly F1/T8-F; this inventory names the finer interaction contexts within those homes.

### Subtraction test

Deliberately **not** separate surfaces:

```text
loading spinner per query
mobile vs desktop layout
success toast
ordinary validation message
cursor page N
visual tab/component boundaries
provider-job diagnostics
queue/retry administration
permission-denied screen per Permission
one screen per lifecycle enum value
```

They become state variants/patterns only when the safe user action materially changes.

### Authority challenge

No surface requires:

```text
new Product route
new application operation
operation 79
new semantic owner
frontend Authorization matrix
parallel DTO/read-model authority
History/Audit as current resolver
provider identity as Product identity
generic retry-jobs screen
```

## 17. F2 result

```text
material surfaces inventoried           36 / 36 current set
accepted human goals with surface       16 / 16
stable SPA routes changed               0
application operations changed          0
operation 79                            absent
new Product/T8-F material gap found     0
cosmetic over-fragmentation detected    none after subtraction pass
F2 status                               COMPLETE CANDIDATE
```

F2 is ready to feed F3 Screen Contracts. It is not a wireframe approval and does not authorize drawing until F3/F4 are coherent under the roadmap sequence.
