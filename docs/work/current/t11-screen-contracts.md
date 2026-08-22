# T11 — MetalDocs Screen Contracts

> **TEMPORARY T11 CANDIDATE WORK / BRANCH-ONLY.** F3 derives implementation-ready interaction contracts for the 36 F2 material surfaces. These contracts reference accepted Product/T6/T8-E/T8-F authority and do not replace it. `READY` means the current accepted backend contract is sufficient for that surface; `BLOCKED` means a material upstream precision is required before wireframing that surface.

## 1. Contract shorthand

Every contract below closes the F3 questions through these fields:

```text
GOAL / ROUTE / TRANCHE
OWNER + READ TRUTH
WRITE CONTROLS
IDENTITY / NAVIGATION
CLIENT STATE
WIRE MECHANICS
MATERIAL STATES / FAILURES
SUCCESS CONSEQUENCE
AUTHZ / DISCLOSURE
PROOF
FORBIDDEN
BACKEND SUFFICIENCY
```

State classification always inherits T8-F:

```text
server truth     → TanStack Query
route/filter     → router/URL
unaccepted input → local form/feature state
ephemeral UI     → local React state
```

All application operations use generated T8-E TypeScript shapes through the one thin transport. UI visibility is never Authorization.

## 2. Application shell / authentication

### APP-01 — Session gate / unauthenticated entry — READY

```text
GOAL/ROUTE       establish session; SPA entry / protected route interception; S1
OWNER/READ       Authentication; getSession → SessionView or auth.unauthenticated
WRITE            none in /api/v1; browser GET /auth/login → external OIDC → /auth/callback
IDENTITY/NAV     callback/session result, never provider claims, returns user to SPA context
CLIENT STATE     getSession query; intended return location URL/router context only
WIRE             same-origin cookie; callback outside app OpenAPI
STATES/FAILURES  valid session; 401; OIDC protocol/provider failure; dependency unavailable
SUCCESS          authoritative SessionView enters APP-02; fresh CSRF material comes only from session read
AUTHZ             authentication gate only; no Product Permission decision here
PROOF             E4 real browser + E5 selected OIDC + E3 getSession
FORBIDDEN         password form; local credential; provider-role Product grant; synthetic session
SUFFICIENCY       accepted authority sufficient
```

### APP-02 — Authenticated shell / stable navigation — READY

```text
GOAL/ROUTE       navigate Product spaces/end session; all SPA routes; S1
OWNER/READ       Authentication; getSession → SessionView
WRITE            endSession
IDENTITY/NAV     stable T8-F route meanings; no client-derived permission routes
CLIENT STATE     SessionView query; local disclosure/focus/nav only
WIRE             DELETE carries CSRF; session cookie cleared by server response
STATES/FAILURES  valid; expired/lost→401; CSRF recovery; end success
SUCCESS          endSession clears/invalidate session presentation and routes APP-01
AUTHZ             SessionView has no permission snapshot; route entry server 403/404 remains authority
PROOF             E4 browser session/logout/direct navigation + E3
FORBIDDEN         permission matrix; route presence as access proof; global authz store
SUFFICIENCY       accepted authority sufficient
```

## 3. Library

### LIB-01 — Official document discovery — READY

```text
GOAL/ROUTE       discover official documents; /documents; S3
OWNER/READ       Controlled Documents; listDocuments → DocumentPage
WRITE            none
IDENTITY/NAV     DocumentSummary.document.document_id → OFF-01 route
CLIENT STATE     query = operationId + normalized filters/cursor; q/filters in URL
WIRE             seek cursor; no offset/total; exact admitted q/filter vocabulary
STATES/FAILURES  empty/non-empty; effective default; admitted obsolete/cancelled; cursor; 403 historical mode; 401/404 as applicable
SUCCESS          authoritative page replaces query entry; row selection routes by returned id
AUTHZ             server filters/discloses; historical catalog modes require current authority
PROOF             E4 list/filter/paging/direct row open + E3; negative DRAFT-as-official
FORBIDDEN         client currentStatus; body/full-text search; generic filters/sort; inferred totals
SUFFICIENCY       accepted authority sufficient
```

### LIB-02 — Create Document interaction — READY

```text
GOAL/ROUTE       create Document; /documents dialog/flow; S3
OWNER/READ       Controlled Documents; getDocumentCreationOptions → DocumentCreationOptionsView; getDocumentTypeNumberingPreview → NumberingPreviewView
WRITE            createDocument → CreateDocumentResult
IDENTITY/NAV     option IDs only from purpose-built projection; result.document_id → OFF-01; result.revision_id is returned fact, not route authority
CLIENT STATE     option/preview queries + local form selections/title
WIRE             CSRF + one Idempotency-Key per logical create; exact query inputs; JSON validation
STATES/FAILURES  changing Area/Type options; candidates/templates present/empty; preview reservation=false; validation/conflict/content integrity; ambiguous retry
SUCCESS          accept CreateDocumentResult; invalidate relevant Library/options; route by returned document_id
AUTHZ             options guidance is filtered current truth; create rechecks all authority/eligibility
PROOF             E4 + E3 + E2 GF2 create path; same-key changed-request and code-race negatives
FORBIDDEN         Admin directories for selectors; preview as reservation; client number generation
SUFFICIENCY       accepted authority sufficient
```

## 4. Document Official

### OFF-01 — Official core overview — READY

```text
GOAL/ROUTE       inspect current/official Document truth; /documents/:document_id; S3 base→S6 enrichment
OWNER/READ       Controlled Documents; getDocument → DocumentOfficialView
WRITE            none directly in core
IDENTITY/NAV     path document_id; returned open_revision/active_obsolescence_request_id only when disclosed; official release/rendition ids from returned truth
CLIENT STATE     getDocument query only; presentation expansion local
WIRE             safe read; no client lifecycle synthesis
STATES/FAILURES  draft|submitted|effective|obsolete|cancelled; official present/absent by schema; 404 non-disclosable
SUCCESS          returned view is the complete current lens input; later child surfaces use returned identities
AUTHZ             disclosure-safe references grant nothing; follow-up operations recheck
PROOF             E4 direct route across status variants + E3; negative DRAFT-as-official/absence inference
FORBIDDEN         currentStatus store; History/Audit resolver; persisted client pointers
SUFFICIENCY       accepted authority sufficient
```

### OFF-02 — Official content presentation — READY

```text
GOAL/ROUTE       inspect exact official content; Document Official region/viewer; S5
OWNER/READ       Controlled Documents; getRelease → ReleaseView; getReleaseSource exact bytes; getOfficialRenditionContent exact PDF bytes
WRITE            none
IDENTITY/NAV     release_id / official_rendition_id only from DocumentOfficialView/ReleaseView
CLIENT STATE     query metadata + ephemeral viewer state; exact bytes never normalized entity truth
WIRE             EXACT_BYTES complete-response integrity; application semantic URLs
STATES/FAILURES  SourceOnly DOCX; SourceOnly PDF; required OfficialRendition PDF; no official; 404/500 content integrity/dependency failure
SUCCESS          render only after exact successful response; OfficialRendition primary only where representation says so
AUTHZ             every exact-byte read rechecks current access/disclosure
PROOF             E4 real DOCX/PDF viewer + E3 + claim-relevant E5; corruption causal negative
FORBIDDEN         provider location identity; viewer-generated official truth; partial corrupt success
SUFFICIENCY       accepted authority sufficient
```

### OFF-03 — Responsible owner management — BLOCKED / F3-F01

```text
GOAL/ROUTE       later responsible-owner replacement; /documents/:document_id; S3
OWNER/READ       Controlled Documents getDocumentResponsibleOwner → ResponsibleOwnerView + ETag
WRITE            replaceDocumentResponsibleOwner(target user_id)
IDENTITY/NAV     current owner id is available; required eligible replacement target identity is NOT completely discoverable from current wire for all accepted states
CLIENT STATE     current owner query + local target selection
WIRE             CSRF + exact If-Match; 412 reconciliation
STATES/FAILURES  current owner; stale ETag; target disabled/offboarding race; denied/non-disclosable
SUCCESS          authoritative ResponsibleOwnerView replaces query; getDocument/affected Work lenses refetch
AUTHZ             document.owner.manage + matching scope + target existing ENABLED same-Company User; server decides
PROOF             E4/E3/E2 owner change + stale/offboarding causal race
FORBIDDEN         manual opaque UUID UX as target discovery; Admin User directory privilege escalation; creation-only options treated as universal truth
SUFFICIENCY       MATERIAL GAP — see §14 F3-F01
```

### OFF-04 — Create / enter open Revision — READY

```text
GOAL/ROUTE       enter existing or create next Revision; OFF-01 → /documents/:document_id/work; S4
OWNER/READ       getDocument.open_revision; target route resolves again through current getDocument then Work reads
WRITE            createDocumentRevision when admitted
IDENTITY/NAV     disclosed open_revision supplies current revision identity; create result supplies new revision_id but route remains document_id/work
CLIENT STATE     no durable client pointer; ephemeral action state
WIRE             create is CSRF + Idempotency-Key
STATES/FAILURES  disclosed existing Work; no disclosed Work; create conflict; 403/404; ambiguous retry
SUCCESS          existing→route; created→refetch official/current truth then route Work
AUTHZ             server current document.edit + relationship/scope/lifecycle authority
PROOF             E4 official→work + E3/E2 create race
FORBIDDEN         History lookup for current Revision; client lifecycle eligibility; dead Work link
SUFFICIENCY       accepted authority sufficient
```

### OFF-05 — Obsolescence management — READY

```text
GOAL/ROUTE       initiate/inspect/withdraw obsolescence; Document Official; S6
OWNER/READ       getDocument active_obsolescence_request_id + getObsolescenceRequest
WRITE            createObsolescenceRequest; withdrawObsolescenceRequest
IDENTITY/NAV     document_id from route; request_id only from create result/current disclosure-safe reference
CLIENT STATE     request query + local reason/confirmation form
WIRE             create CSRF+Idempotency-Key; withdrawal CSRF
STATES/FAILURES  none disclosed; governance_pending active; synchronous obsolete; returned/withdrawn/completed; conflict/404/403
SUCCESS          refetch getDocument + request/history/list lenses; synchronous obsolete updates official truth
AUTHZ             document.obsolete + scope + lifecycle; withdrawal relation predicate; server authority
PROOF             E4/E3/E2 GF5; absence-inference and competing-request negatives
FORBIDDEN         fake approval Step; active-request client pointer; operation79 navigation repair
SUFFICIENCY       accepted authority sufficient
```

## 5. History

### HIS-01 — Semantic Document history — READY

```text
GOAL/ROUTE       inspect controlled-document history; /documents/:document_id/history; S3 base→S6 enrichment
OWNER/READ       Controlled Documents; getDocumentHistory → DocumentHistoryPage; explicit supporting reads only from returned ids
WRITE            none by virtue of History
IDENTITY/NAV     document_id path; submission/governance/release/rendition/request ids only from union items
CLIENT STATE     paged query; optional selected item/viewer ephemeral
WIRE             seek cursor; supporting exact-byte operations preserve their own laws
STATES/FAILURES  each closed history kind; empty/page; supporting resource denied/non-disclosable
SUCCESS          append/refetch server history facts; selected historical content remains read-only
AUTHZ             document.read_history/current exact-content access; server disclosure always rechecked
PROOF             E4/E3 history chronology/content links; negative current-resource resolution attempt
FORBIDDEN         current lifecycle reconstruction authority; Audit merge; current Work resolver
SUFFICIENCY       accepted authority sufficient
```

## 6. My Work

### WRK-01 — Authoring work lane — READY

```text
GOAL/ROUTE       find actor-relevant authoring work; /work; S4
OWNER/READ       Controlled Documents; listAuthoringWork → WorkAuthoringPage
WRITE            none
IDENTITY/NAV     item.document.id → /documents/:id/work; current route then re-resolves current Work truth
CLIENT STATE     paged query + local lane selection
WIRE             seek cursor
STATES/FAILURES  empty/page; stale projection target now absent/non-disclosable
SUCCESS          selecting row opens live DW route; projection is never carried as mutation truth
AUTHZ             query already actor-relevant; target rechecks current authorization/relationship
PROOF             E4/E3 projection→live target; stale-row causal case
FORBIDDEN         projection-owned lifecycle; direct mutation from list row using stale projection
SUFFICIENCY       accepted authority sufficient
```

### WRK-02 — Governance work lane — READY

```text
GOAL/ROUTE       find actor-relevant governance work; /work; S5
OWNER/READ       Controlled Documents; listGovernanceWork → WorkGovernancePage
WRITE            none
IDENTITY/NAV     item.governance_attempt_id → /work/governance/:attempt_id
CLIENT STATE     paged query + local lane selection
WIRE             seek cursor
STATES/FAILURES  empty/page; target attempt already resolved/withdrawn/non-disclosable
SUCCESS          row opens GOV-01 which re-reads canonical case
AUTHZ             projection attention only; case operation recomputes current participation/disclosure
PROOF             E4/E3 projection→case; stale row case
FORBIDDEN         Work projection as decision authority; dead target before S5
SUFFICIENCY       accepted authority sufficient
```

## 7. Document Work

### DW-01 — Current DRAFT authoring / inspection — READY

```text
GOAL/ROUTE       author/inspect current DRAFT; /documents/:document_id/work; S4
OWNER/READ       getDocument resolver + getRevision → RevisionView + getRevisionDraft → DocumentWorkView/ETag + getRevisionDraftSource exact bytes
WRITE            updateRevisionDraft title/source attachment as applicable
IDENTITY/NAV     document route → getDocument.open_revision → revision_id; never History
CLIENT STATE     server queries + local title/editor buffer; viewer focus ephemeral
WIRE             exact If-Match on PATCH; exact-byte source read
STATES/FAILURES  DOCX editable; PDF read-only; clean/dirty; save success; 412 draft_changed; 404/current state drift
SUCCESS          authoritative DocumentWorkView + new ETag replaces query; related Work/Official refs refetch if needed
AUTHZ             current read_working/edit + relationship/scope/lifecycle server checks
PROOF             E4/E3/E2 GF3 OCC; stale DRAFT negative
FORBIDDEN         automatic merge/LWW; EditorSession truth; client generation authority
SUFFICIENCY       accepted authority sufficient
```

### DW-02 — Source replacement / upload admission — READY

```text
GOAL/ROUTE       replace DRAFT source; Document Work; S4
OWNER/READ       current DRAFT/ETag; exact resulting Work read
WRITE            startRevisionDraftUpload → provider PUT → completeRevisionDraftUpload → updateRevisionDraft(source_upload_id)
IDENTITY/NAV     revision_id current Work; upload_id only from allocation
CLIENT STATE     intended local bytes + transient progress; server state for final Work
WIRE             CSRF; opaque upload URL/required_headers verbatim; final PATCH If-Match
STATES/FAILURES  allocated/uploading/PUT done/not READY; complete success; 410 expired; content invalid; stale DRAFT before attach; dependency failure
SUCCESS          only authoritative final DocumentWorkView means source attached
AUTHZ             allocation/completion/attach each server-enforced; provider capability grants no Product rights
PROOF             E4+E3+E5/E2 GF3; expired/corrupt/provider-PUT-only negatives
FORBIDDEN         provider success=WorkingContent; revived expired allocation; provider identity parsing
SUFFICIENCY       accepted authority sufficient
```

### DW-03 — Submit DRAFT command — READY

```text
GOAL/ROUTE       submit exact DRAFT generation; Document Work; S4
OWNER/READ       current DRAFT/ETag; resulting submission identities/status
WRITE            createSubmission
IDENTITY/NAV     revision_id from Work; governance_attempt/release identity only from SubmissionCreateResult/current server reads
CLIENT STATE     local confirmation only; logical-command key scope
WIRE             CSRF + Idempotency-Key + semantic DRAFT If-Match
STATES/FAILURES  governance_pending|rendition_pending|released; 412 draft_changed; content invalid/malicious; scanner unavailable; conflict; key reused
SUCCESS          accept result, invalidate Work/Official/MyWork/History; fetch resulting Submission/current truth as needed
AUTHZ             current document.submit + relationship/state; submission revalidates exact content
PROOF             E4/E3/E2/E5 GF3/GF4 precursor; stale/malware/key negatives
FORBIDDEN         optimistic SUBMITTED lifecycle; new key for same ambiguous retry
SUFFICIENCY       accepted authority sufficient
```

### DW-04 — Submitted Revision / Submission state + termination — READY

```text
GOAL/ROUTE       inspect submitted work and withdraw/cancel when admitted; Document Work; S4
OWNER/READ       getRevision.current_submission_id → getSubmission → SubmissionView; getSubmissionSource exact bytes
WRITE            withdrawSubmission; cancelRevision
IDENTITY/NAV     submission_id only from current Revision/result; release/governance ids only from SubmissionView/result
CLIENT STATE     server queries; local reason/confirm forms
WIRE             CSRF; cancellation JSON; exact source bytes
STATES/FAILURES  human gate required/satisfied; representation gate required/satisfied; released; returned/withdrawn/cancelled; conflicts/403/404
SUCCESS          refetch Revision/Submission/Official/MyWork/History; immutable submitted content stays unchanged
AUTHZ             withdraw relationship + submit permission; cancel permission/lifecycle; server current truth
PROOF             E4/E3/E2; immutable subject/termination/repeat conflict negatives
FORBIDDEN         mutation of submitted bytes; gates collapsed into one generic status
SUFFICIENCY       accepted authority sufficient
```

## 8. Governance Case

### GOV-01 — Case overview / immutable subject — READY

```text
GOAL/ROUTE       inspect exact governed case; /work/governance/:attempt_id; S5
OWNER/READ       getGovernanceAttempt → GovernanceCaseView; listGovernanceFeedback continuation; submission source when subject kind=submission
WRITE            none in overview
IDENTITY/NAV     attempt_id route; subject ids/Step ids from case; no inferred resource
CLIENT STATE     case/feedback queries + selected Step/viewer ephemeral
WIRE             cursor continuation; exact-byte subject read when required
STATES/FAILURES  attempt active/completed/returned/withdrawn/cancelled; subject submission|obsolescence; Steps pending/active/decided; allowed_actions subset/empty
SUCCESS          server case is sole display authority
AUTHZ             exact participation/disclosure; allowed_actions hints only
PROOF             E4/E3/E5 where exact source; nonparticipant/disclosure negative
FORBIDDEN         general WorkingContent access; client action matrix; mutable governed subject
SUFFICIENCY       accepted authority sufficient
```

### GOV-02 — Feedback composer — READY

```text
GOAL/ROUTE       add case feedback; GOV-01 dialog/region; S5
OWNER/READ       current GovernanceCaseView/feedback
WRITE            createGovernanceFeedback → CreateGovernanceFeedbackResult
IDENTITY/NAV     attempt_id from route; returned feedback_id only evidence/refetch handle
CLIENT STATE     local message + logical command key
WIRE             CSRF + Idempotency-Key
STATES/FAILURES  accepted; validation; current case/action loss; conflict; key reuse/ambiguous retry
SUCCESS          invalidate/refetch case/feedback/history as semantically affected
AUTHZ             current governance participation/action server check
PROOF             E4/E3/E2 idempotency/authorization negative
FORBIDDEN         feedback as decision/lifecycle state
SUFFICIENCY       accepted authority sufficient
```

### GOV-03 — Step decision interaction — READY

```text
GOAL/ROUTE       ACCEPT or RETURN_FOR_CHANGES active Step; GOV-01 dialog/region; S5
OWNER/READ       current case Step; getGovernanceStepDecision when exact read needed
WRITE            recordGovernanceStepDecision
IDENTITY/NAV     attempt_id route + step_id from exact case
CLIENT STATE     local outcome/reason form only
WIRE             CSRF; closed decision union
STATES/FAILURES  accept; return(reason); exact repeat; already_decided; current AuthZ loss; state conflict
SUCCESS          replace/refetch case/decision + Work/Submission/Official/History according to server result
AUTHZ             governance.act + exact active candidate snapshot + domain predicates; server only
PROOF             E4/E3/E2 GF4; concurrent decision/authorization negatives
FORBIDDEN         decision=publish; client determines next lifecycle
SUFFICIENCY       accepted authority sufficient
```

## 9. Admin / Organization

### ORG-01 — Company settings — READY

```text
GOAL/ROUTE       manage Company display settings; /admin/organization; S1
OWNER/READ       Organization; getCompany → CompanyView + ETag
WRITE            replaceCompany
IDENTITY/NAV     singleton company_id returned; no tenant selector
CLIENT STATE     query + local form bound to ETag
WIRE             CSRF + If-Match
STATES/FAILURES  current; dirty; 412 reconcile; denied
SUCCESS          authoritative CompanyView/ETag replaces query
AUTHZ             organization.manage server check
PROOF             E4/E3/E2 stale negative
FORBIDDEN         multi-company selector/pooled tenancy
SUFFICIENCY       accepted authority sufficient
```

### ORG-02 — User directory + atomic create — READY

```text
GOAL/ROUTE       inspect/create Users; /admin/organization; S1
OWNER/READ       Organization listUsers → UserPage; Authentication searchProviderSubjects → ProviderSubjectSearchView
WRITE            createUser → CreateUserResult
IDENTITY/NAV     user_id returned/listed; provider_subject_ref opaque search result
CLIENT STATE     paged users query + provider search query + local profile/create form
WIRE             create CSRF+Idempotency-Key; provider query SearchQuery; seek users cursor
STATES/FAILURES  empty/page; provider search refine; create conflict/key reuse/dependency unavailable
SUCCESS          fetch/refetch new User/profile/binding context by returned user_id
AUTHZ             organization.manage; provider search is bounded preflight, not directory authority
PROOF             E4/E3/E2/E5; unbound/duplicate provider causal cases
FORBIDDEN         raw issuer/subject parsing; password; provider role grant
SUFFICIENCY       accepted authority sufficient
```

### ORG-03 — User Profile — READY

```text
GOAL/ROUTE       view/replace/recreate/erase profile; /admin/organization selected User; S1
OWNER/READ       Organization; getUserProfile → UserProfileView + ETag or 404
WRITE            replaceUserProfile; deleteUserProfile
IDENTITY/NAV     selected user_id from UserPage/CreateUserResult
CLIENT STATE     profile query + local profile form
WIRE             replace uses If-Match when present or If-None-Match:* when absent; delete CSRF
STATES/FAILURES  present/absent; replace/recreate; erase; 412 matrix; validation/404
SUCCESS          replace current view/ETag or invalidate to absent state
AUTHZ             organization.manage; profile PII never identity authority
PROOF             E4/E3/E2 conditional matrix/erase negative
FORBIDDEN         email as auth identity; client-generated version
SUFFICIENCY       accepted authority sufficient
```

### ORG-04 — Provider binding — READY

```text
GOAL/ROUTE       inspect/replace binding; /admin/organization selected User; S1
OWNER/READ       Authentication getUserProviderBinding + searchProviderSubjects
WRITE            replaceUserProviderBinding
IDENTITY/NAV     user_id selected; provider_subject_ref from opaque search result
CLIENT STATE     binding query + search query + local selection
WIRE             CSRF + binding If-Match
STATES/FAILURES  current; search no/multiple hints; 412; provider conflict/unavailable; post-replacement current-session loss where actor affected
SUCCESS          authoritative binding/ETag; invalidate affected User sessions presentation as server dictates
AUTHZ             organization.manage/current binding rules; provider data never grants Product authority
PROOF             E4/E3/E2/E5; stale/conflict/session-revocation negatives
FORBIDDEN         raw provider identity editing; provider group mapping
SUFFICIENCY       accepted authority sufficient
```

### ORG-05 — User eligibility / offboarding / re-enable — READY

```text
GOAL/ROUTE       disable/re-enable User; /admin/organization selected User; S1
OWNER/READ       Organization getUserEligibility → UserEligibilityView + ETag
WRITE            replaceUserEligibility
IDENTITY/NAV     user_id from directory
CLIENT STATE     eligibility query + local explicit confirmation
WIRE             CSRF + If-Match
STATES/FAILURES  enabled|disabled; 412; security-state conflict; actor session may become invalid
SUCCESS          authoritative eligibility; invalidate/refetch users/memberships/assignments/session-relevant queries
AUTHZ             organization.manage; server transaction owns session/membership/direct-grant teardown
PROOF             E4/E3/E2 GF1 offboarding atomicity/revocation
FORBIDDEN         UI claim that re-enable restores old grants/memberships/sessions
SUFFICIENCY       accepted authority sufficient
```

### ORG-06 — Area identity directory + create/edit — READY

```text
GOAL/ROUTE       list/create/rename Areas; /admin/organization; S1
OWNER/READ       listAreas → AreaPage; getArea → AreaView + ETag
WRITE            createArea; replaceArea
IDENTITY/NAV     area_id from page/create result
CLIENT STATE     paged query + local create/edit forms
WIRE             create CSRF+Idempotency-Key; edit CSRF+If-Match
STATES/FAILURES  active/retired summary visible; create conflict; 412 edit; code immutable
SUCCESS          refetch Area list/detail/options affected
AUTHZ             organization.manage
PROOF             E4/E3/E2 uniqueness/OCC negative
FORBIDDEN         editable Area code after creation
SUFFICIENCY       accepted authority sufficient
```

### ORG-07 — Area lifecycle — READY

```text
GOAL/ROUTE       retire/re-enable Area; /admin/organization selected Area; S1
OWNER/READ       getAreaLifecycle → AreaLifecycleView + independent ETag
WRITE            replaceAreaLifecycle
IDENTITY/NAV     selected area_id
CLIENT STATE     lifecycle query + confirmation
WIRE             CSRF + lifecycle If-Match
STATES/FAILURES  active|retired; 412; create-vs-retire race
SUCCESS          replace lifecycle query; invalidate Areas/creation options as needed
AUTHZ             organization.manage
PROOF             E4/E3/E2 serialization negative
FORBIDDEN         merge lifecycle ETag with Area identity ETag
SUFFICIENCY       accepted authority sufficient
```

### ORG-08 — Group identity directory + create/edit/delete — READY

```text
GOAL/ROUTE       manage Group identity; /admin/organization; S1
OWNER/READ       listGroups → GroupPage; getGroup → GroupView + ETag
WRITE            createGroup; replaceGroup; deleteGroup
IDENTITY/NAV     group_id from list/create result
CLIENT STATE     paged query + local create/edit/delete confirmation
WIRE             create CSRF+Idempotency-Key; edit If-Match; delete CSRF
STATES/FAILURES  create/edit; 412; delete 409 live dependency; 404
SUCCESS          invalidate Group lists/detail and access/governance reference queries
AUTHZ             organization.manage for identity/delete; group membership remains ACC-01
PROOF             E4/E3/E2 dependency conflict/OCC negatives
FORBIDDEN         nested/dynamic/provider-mirrored group
SUFFICIENCY       accepted authority sufficient
```

## 10. Admin / Access

### ACC-01 — Group membership administration — READY

```text
GOAL/ROUTE       inspect/add/remove GroupMembership; /admin/access; S1
OWNER/READ       Organization listGroupMembers; supporting listGroups/listUsers bounded references
WRITE            addGroupMember; removeGroupMember
IDENTITY/NAV     group_id/user_id from accepted reads
CLIENT STATE     member query + supporting reference queries + local selection
WIRE             add PUT CSRF; remove DELETE CSRF; pagination
STATES/FAILURES  member/nonmember; add existing 204; remove absent relation 204; group 404; state conflict
SUCCESS          refetch members + access-sensitive views; current server access changes immediately
AUTHZ             access.manage; supporting Organization reads are available to current governance_admin bundle; server remains authority
PROOF             E4/E3/E2 GF1 revocation/access change
FORBIDDEN         membership as mere UI label; client-effective permission cache
SUFFICIENCY       accepted authority sufficient under fixed role bundle
```

### ACC-02 — Role catalog + RoleAssignment administration — READY

```text
GOAL/ROUTE       inspect fixed roles and grant/revoke assignments; /admin/access; S1
OWNER/READ       Authorization listRoles → RoleListView; listRoleAssignments → RoleAssignmentPage; supporting User/Group/Area refs
WRITE            createRoleAssignment; deleteRoleAssignment
IDENTITY/NAV     subject/scope ids from accepted reads; assignment_id from page/create result
CLIENT STATE     queries + local grant form/confirmation
WIRE             create CSRF+Idempotency-Key; revoke CSRF; seek cursor
STATES/FAILURES  subject user|group; scope company|area; role scope compatibility; grant conflict/key reuse; revoke 404
SUCCESS          invalidate assignments and access-sensitive current views
AUTHZ             access.manage; static role bundle remains server/code authority
PROOF             E4/E3/E2 current grant/revoke + next-request effect
FORBIDDEN         custom role/permission editor; hierarchical roles; client evaluator
SUFFICIENCY       accepted authority sufficient under fixed governance_admin bundle
```

## 11. Admin / Document Governance

### DGV-01 — DocumentType directory + create — READY

```text
GOAL/ROUTE       list/create DocumentTypes; /admin/document-governance; S2
OWNER/READ       Controlled Documents listDocumentTypes → DocumentTypePage; supporting Organization refs for governance selectors where required
WRITE            createDocumentType including initial GovernancePolicy + RepresentationPolicy
IDENTITY/NAV     document_type_id from list/create result
CLIENT STATE     paged query + local composite create form
WIRE             CSRF + Idempotency-Key; closed governance/representation unions
STATES/FAILURES  active/inactive create values; no-human vs route; source-only vs required PDF; validation/conflict/key reuse
SUCCESS          refetch types/governance/creation-options relevant projections
AUTHZ             document_type.manage; fixed governance_admin bundle supplies needed Org selector reads
PROOF             E4/E3/E2 configuration atomicity/validation
FORBIDDEN         implicit default governance; generic workflow builder
SUFFICIENCY       accepted authority sufficient
```

### DGV-02 — Existing DocumentType base configuration — READY

```text
GOAL/ROUTE       inspect/edit base type; /admin/document-governance; S2
OWNER/READ       getDocumentType → DocumentTypeView + ETag
WRITE            replaceDocumentType
IDENTITY/NAV     selected document_type_id
CLIENT STATE     query + local form
WIRE             CSRF + If-Match
STATES/FAILURES  active flag/name/code/numbering values; 412; 409 immutable-after-use restrictions
SUCCESS          authoritative view/ETag; invalidate type lists/options
AUTHZ             document_type.manage
PROOF             E4/E3/E2 OCC/state conflict
FORBIDDEN         client assumes code/numbering always mutable
SUFFICIENCY       accepted authority sufficient
```

### DGV-03 — Governance + representation policy — READY

```text
GOAL/ROUTE       edit governance/representation; /admin/document-governance selected type; S2
OWNER/READ       getDocumentTypeGovernance → DocumentTypeGovernanceView + ETag; supporting User/Group refs for selectors
WRITE            replaceDocumentTypeGovernance
IDENTITY/NAV     document_type_id selected; selector ids accepted references
CLIENT STATE     query + local ordered route/policy form
WIRE             CSRF + If-Match; closed unions; Step order semantic
STATES/FAILURES  no_human_approval; use_governance_route ordered steps; source_only; require_official_rendition(pdf); 412/validation/conflict
SUCCESS          authoritative policy/ETag; invalidate dependent creation/governance config views
AUTHZ             document_type.manage
PROOF             E4/E3/E2 OCC/order/current-reference negatives
FORBIDDEN         drag/drop semantics beyond ordered accepted steps; generic policy language
SUFFICIENCY       accepted authority sufficient
```

### DGV-04 — Eligible Template set — READY

```text
GOAL/ROUTE       configure eligible Templates for type; /admin/document-governance; S2
OWNER/READ       getDocumentTypeEligibleTemplates → EligibleTemplatesView + ETag; listTemplateConfigurations supplies bounded selectable Template config projection
WRITE            replaceDocumentTypeEligibleTemplates
IDENTITY/NAV     Document ids from TemplateConfigurationItem/selected set
CLIENT STATE     selected-set query + candidate projection query + local set edits
WIRE             CSRF + If-Match; replacement set unique/order-insensitive
STATES/FAILURES  empty/nonempty; stale ETag; invalid/ineligible reference conflict
SUCCESS          authoritative selected set/ETag; invalidate template/config/create-options views
AUTHZ             document_type.manage/template configuration authority; no implied content read
PROOF             E4/E3/E2 OCC/disclosure negative
FORBIDDEN         Template peer lifecycle; content/history read inferred from config
SUFFICIENCY       accepted authority sufficient
```

### DGV-05 — Numbering preview — READY

```text
GOAL/ROUTE       inspect non-reserving next-code preview; admin config and LIB-02 supporting context; S2/S3 consume same op
OWNER/READ       getDocumentTypeNumberingPreview → NumberingPreviewView
WRITE            none
IDENTITY/NAV     document_type_id + optional area_id from accepted selected context
CLIENT STATE     query only
WIRE             safe read; exact optional query
STATES/FAILURES  preview result reservation=false; validation/404/403
SUCCESS          display guidance only; no mutation/invalidation
AUTHZ             server validates scope/context
PROOF             E4/E3 plus causal concurrent create proving preview not reservation
FORBIDDEN         local sequence/finality claim
SUFFICIENCY       accepted authority sufficient
```

### DGV-06 — Template configuration / template-role — READY

```text
GOAL/ROUTE       inspect Template config and set concrete Document template role; /admin/document-governance; S2 base→S3 enrichment
OWNER/READ       listTemplateConfigurations → TemplateConfigurationPage; getDocumentTemplateRole → TemplateRoleView + ETag
WRITE            replaceDocumentTemplateRole
IDENTITY/NAV     document_id from TemplateConfigurationItem
CLIENT STATE     paged config query + selected role query + local toggle
WIRE             template-role CSRF+If-Match; seek cursor
STATES/FAILURES  template_role true/false; has_effective_revision; title presence; eligible type ids; 412/conflict/404
SUCCESS          authoritative TemplateRoleView/ETag; invalidate config/eligible-template/creation-option views
AUTHZ             template_use.manage; config access does not imply content/history
PROOF             E4/E3/E2 OCC/disclosure negative
FORBIDDEN         Template as separate semantic owner; admin content viewer without permission
SUFFICIENCY       accepted authority sufficient
```

## 12. Audit

### AUD-01 — Audit evidence list — READY

```text
GOAL/ROUTE       inspect meaningful action evidence; /audit; S6
OWNER/READ       Audit; listAuditEvents → AuditEventPage closed union
WRITE            none
IDENTITY/NAV     event/resource ids are evidence identifiers only; no current-resource resolver promise
CLIENT STATE     paged query only
WIRE             cursor+limit; no filter/search DSL
STATES/FAILURES  empty/page; 403/401; event union variants
SUCCESS          next page by opaque cursor
AUTHZ             audit.read historical visibility filtering before pagination
PROOF             E4/E3 disclosure/paging; negative current-truth inference
FORBIDDEN         Audit-driven current lifecycle; generic facts bag/filter builder
SUFFICIENCY       accepted authority sufficient
```

## 13. Cross-cutting Screen Contract rules

Every READY surface inherits these material branches where reachable:

```text
401 auth.unauthenticated
  → invalidate session presentation → APP-01

403 permission.denied
  → denied action/context; no client permission inference

404 notfound.resource
  → absent OR non-disclosable; never tell which unless another accepted truth does

permission.csrf_failed
  → re-bootstrap getSession/CSRF then retry SAME logical unsafe command only when safe;
    preserve Idempotency-Key and conditional semantics

precondition.resource_changed
  → preserve local input + refetch + explicit reconcile

precondition.draft_changed
  → same, DRAFT-specific; never automatic merge/LWW

validation.idempotency_key_reused
  → logical-command misuse/state; not ordinary field correction

dependency.*
  → sanitized dependency unavailable; no raw provider internals

internal.content_integrity
  → no successful exact-content presentation
```

Mutation success follows T8-F:

```text
accept returned authoritative result
→ replace exact affected query when returned
→ invalidate/refetch only semantic lenses that can have changed
→ never fabricate lifecycle truth optimistically
```

## 14. Material finding F3-F01 — Responsible-owner candidate discovery

### Evidence

Accepted owner replacement requires:

```text
document.owner.manage
+ matching scope
+ target existing MetalDocs User
+ same Company
+ current target eligibility = ENABLED
+ current ResponsibleOwner If-Match
```

The normal holder is `area_manager`, whose static bundle includes `document.owner.manage` but not `organization.manage`.

Current wire provides:

```text
getDocumentResponsibleOwner
  → current responsible_owner_user_id + ETag

replaceDocumentResponsibleOwner
  → consumes target user_id

DocumentCreationOptionsView.responsible_owner_candidates
  → candidates only inside the creation-options contract
  → base admission requires document.create
  → Area/DocumentType choices are current creation-eligible/ACTIVE truth

listUsers
  → Organization administration surface, not the least-privilege document owner-management selector
```

Create authorization requires active Area/DocumentType. Later responsible-owner replacement does **not**: T3/D4 requires matching scope + target enabled same-Company User, with no active Area/DocumentType predicate. Area retirement can linearize against **creation** without erasing existing Document truth.

Therefore creation options are not a complete candidate source for every valid owner-replacement state, and an opaque manual UUID field is not an acceptable realization of the accepted human management journey.

### Rejected repairs

```text
Use Admin listUsers
  REJECT — couples area_manager journey to organization.manage/admin directory and violates least-privilege selector design.

Stuff candidates into ResponsibleOwnerView
  REJECT — that response is ETag-protected; independently mutable User eligibility would pollute one concurrency domain.

Treat DocumentCreationOptions as universal owner-management truth
  REJECT — its semantics intentionally filter for creation-eligible active context.

Add operation 79
  NOT JUSTIFIED — a smaller existing-read precision is available.
```

### Recommended smallest precision

Boundedly enrich the non-ETag `DocumentOfficialView` returned by existing operation 47 `getDocument`:

```text
responsible_owner_candidates?: UserReference[]
```

Presence law:

```text
present iff caller may receive the Document context
         AND current document.owner.manage authorization is ALLOW for this Document

when present:
  complete bounded set of existing ENABLED Users in the same Company eligible under D4
  deterministic user_id ASC
  current responsible owner may remain present only if currently ENABLED; disabled current owner remains truthful in responsible_owner but is not newly selectable

absent:
  caller cannot infer whether owner-management authority or candidate existence is missing
```

The field is derived read truth only. It grants no Permission and remains revalidated on `replaceDocumentResponsibleOwner`.

Internal composition needs the matching bounded Organization enumeration/fact seam; Organization still owns User existence/eligibility, Controlled Documents still owns responsibility, Authorization still owns ALLOW/default-DENY, and application composes.

This mirrors the already-ratified T8-F read-symmetry pattern that added `open_revision` and `active_obsolescence_request_id` to `DocumentOfficialView` without adding an operation.

### Implicated authority

Smallest bounded correction appears to touch:

```text
T6 / product journeys       precision of later owner-management presentation only
T8-C interfaces             bounded Organization eligible-owner candidate enumeration/composition seam
T8-E wire contract          DocumentOfficialView optional field + deterministic presence/order fixtures
T8-F frontend               OFF-03 consumes returned candidates; no new route/owner/state authority
T9 validation               existing V3/V6/GF2 lanes gain a bounded candidate-disclosure/eligibility falsifier; no new GF/V class
T10                         operation count/invariants unchanged; no cutover semantic change
T11                         no operation reassignment/count change
```

No new Product capability, Permission, semantic owner, route or application operation is required.

### Status

```text
severity                    MATERIAL implementation-readiness gap
operation 79                still absent
recommended outcome         bounded upstream precision, not new capability
operator approval           REQUIRED before accepted authority is edited
F3 overall status           BLOCKED pending adjudication of F3-F01
wireframing                 NOT AUTHORIZED while F3 remains blocked
```

## 15. F3 progress summary

```text
F2 material surfaces                       36
Screen Contracts derived                   36 / 36
READY from current authority                35
BLOCKED by material gap                      1 — OFF-03 / F3-F01
new operation required                       0
operation 79                                absent
unresolved MATERIAL finding                  1
F3 status                                   NOT COMPLETE / STOP at bounded reopen gate
F4/F5                                       NOT OPEN while F3-F01 unresolved
```
