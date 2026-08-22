# T11 — MetalDocs Material Interaction Ledger

> **TEMPORARY T11 CANDIDATE WORK / BRANCH-ONLY.** F6 binds every material wireframe control to accepted owner/wire behavior. It is implementation-planning Evidence, not a second Product/API contract.

## 1. Ledger law

Every material action answers:

```text
surface/control
→ semantic owner / admitted browser mechanism
→ exact operationId or AuthN browser route
→ input source
→ wire mechanics
→ authoritative success truth
→ material failure states
→ client query/navigation consequence
→ retry law
→ forbidden inference
```

`Problems` below list the material branches that change the user's safe next action; the complete exact Problem set remains T8-E authority.

## 2. Browser authentication / shell

| Surface / control | Owner | Operation | Input / wire | Success | Material failure | Client consequence / retry | Forbidden |
|---|---|---|---|---|---|---|---|
| WF-00 `Sign in` | Authentication + external IdP mechanism | `GET /auth/login` → `/auth/callback` | intended return route is browser nav context; no Product body | valid callback creates ApplicationSession; subsequent `getSession` returns SessionView | invalid state/code/provider, dependency unavailable | failure returns WF-00; retry starts fresh OIDC transaction | local password; provider claim as Product grant |
| APP entry/session bootstrap | Authentication | `getSession` | HttpOnly cookie | SessionView + current CSRF material | `auth.unauthenticated`, dependency/internal failure | 401→WF-00; otherwise sanitized retry | localStorage session authority |
| WF-01 `Sign out` | Authentication | `endSession` | CSRF; session cookie | 204 + session cookie clear | csrf/auth failure | success clears session query→WF-00; CSRF recovery uses fresh SessionView | client-only logout success while cookie remains valid |

## 3. Library / Document Official controls

| Surface / control | Owner | operationId | Input / wire | Success truth | Material failure | Client consequence / retry | Forbidden |
|---|---|---|---|---|---|---|---|
| WF-02 search/filter/page | Controlled Documents | `listDocuments` | exact q/filter/cursor/limit query; safe read | DocumentPage | 403 history-sensitive status; bad cursor; auth/dependency | replace page/query entry; cursor remains opaque | generic filter DSL; offset/total; client status truth |
| WF-02 row `Open` | Controlled Documents | `getDocument` at destination | returned `document_id` | DocumentOfficialView | 404/non-disclosable | route WF-04 or unavailable | projection row as current truth |
| WF-02 metadata `Filter by this Type/Area/Owner` | Controlled Documents | `listDocuments` | exact returned reference id becomes admitted filter | filtered DocumentPage | normal list Problems | URL query changes + refetch | Admin directory lookup for decoration/selection |
| WF-03 options refresh | Controlled Documents | `getDocumentCreationOptions` | selected area/type ids | current bounded options | 404 selected ref; auth | replace options query; form selection reconciles | options as grant/eligibility authority |
| WF-03 numbering preview | Controlled Documents | `getDocumentTypeNumberingPreview` | type + optional area | NumberingPreviewView `reservation=false` | validation/notfound | show new preview only | reservation/client sequence |
| WF-03 `Create` | Controlled Documents | `createDocument` | form + CSRF + **Idempotency-Key** | CreateDocumentResult | validation, state conflict, content integrity, ambiguous transport, key reused | success→WF-06→Work; ambiguous retry SAME key; changed form=new key | optimistic Document/code; second create on retry |
| WF-04 official route load | Controlled Documents | `getDocument` | path document_id | DocumentOfficialView incl. disclosure refs/candidates | 404/non-disclosable | replace current lens | History/Audit fallback |
| WF-04 official source metadata | Controlled Documents | `getRelease` | release_id from DocumentOfficialView | ReleaseView | 404/denied | viewer mode derives from returned representation | current Release guessed from revision |
| WF-04 `Open exact source` | Controlled Documents | `getReleaseSource` | release_id | exact bytes | content-integrity/dependency/notfound | render only complete verified 200 | provider URL/partial corrupt success |
| WF-04 OfficialRendition viewer | Controlled Documents | `getOfficialRenditionContent` | rendition id from returned Release/Official view | exact PDF bytes | content-integrity/dependency/notfound | render only verified success | viewer output as OfficialRendition authority |
| WF-04 `Open work` | Controlled Documents | destination `getDocument` | current document_id | current open_revision resolver | no disclosed current Work/404 | branch WF-07 or WF-10 from fresh server state | cached open_revision as durable pointer |
| WF-04 `Create revision` | Controlled Documents | `createDocumentRevision` | document_id + CSRF + **Idempotency-Key** | CreateRevisionResult | conflict/notfound/content integrity/ambiguous transport/key reused | success→Work route + fresh getDocument; ambiguous retry SAME key | History lookup; second revision on retry |
| WF-04 `History` | Controlled Documents | destination `getDocumentHistory` | document_id | DocumentHistoryPage | denied/notfound | WF-14 | History used to repair current Work resolution |

## 4. Responsible owner / obsolescence

| Surface / control | Owner | operationId | Input / wire | Success truth | Material failure | Client consequence / retry | Forbidden |
|---|---|---|---|---|---|---|---|
| WF-05 owner candidate selector | Controlled Docs lens composed with Organization/AuthZ | `getDocument` | safe read | optional complete `responsible_owner_candidates` under T8-E-RO | absent member | no inference from absence; selector exists only when returned | Admin `listUsers` as required dependency; client eligibility engine |
| WF-05 load current owner | Controlled Documents | `getDocumentResponsibleOwner` | document_id | ResponsibleOwnerView + strong ETag | denied/notfound | bind form to exact ETag | candidate list used as concurrency token |
| WF-05 `Change owner` | Controlled Documents | `replaceDocumentResponsibleOwner` | selected returned user_id + CSRF + **If-Match** | resulting ResponsibleOwnerView + ETag | `precondition.resource_changed`, target conflict/eligibility race, denied | success replace owner query + refetch getDocument/Work; stale preserves selection + explicit reconcile | automatic overwrite; candidate presence as mutation guarantee |
| WF-05 load active obsolescence | Controlled Documents | `getObsolescenceRequest` | request_id only from current getDocument/create result | ObsolescenceRequestView | 404/denied | show request state or refetch Document | guessed/current request id |
| WF-05 `Start obsolescence` | Controlled Documents | `createObsolescenceRequest` | reason + CSRF + **Idempotency-Key** | governance_pending or obsolete result | validation/conflict/ambiguous transport/key reused | remain OFF; refetch Document/request/History; same-key ambiguous retry | fake approval state; duplicate request |
| WF-05 `Withdraw request` | Controlled Documents | `withdrawObsolescenceRequest` | request_id + CSRF | withdrawal view / exact repeat | conflict/denied/notfound | refetch Document/request/History | generic retry after changed lifecycle |

## 5. Document Work / upload / submission

| Surface / control | Owner / mechanism | operationId / action | Input / wire | Success truth | Material failure | Client consequence / retry | Forbidden |
|---|---|---|---|---|---|---|---|
| WF-06 Work route resolve | Controlled Documents | `getDocument` | document_id | disclosure-safe open_revision | absent/404 | DRAFT→WF-07/08; SUBMITTED→WF-10; absent→no Work | History fallback |
| WF-07 load Revision | Controlled Documents | `getRevision` | revision_id from resolver | RevisionView | 404/state drift | branch current state | cached Revision lifecycle authority |
| WF-07 load DRAFT | Controlled Documents | `getRevisionDraft` | current revision_id | DocumentWorkView + DRAFT ETag | notfound/state drift | editor binds exact current ETag | editor session generation authority |
| WF-07/08 load DRAFT source | Controlled Documents | `getRevisionDraftSource` | current revision_id | exact bytes | integrity/dependency/notfound | editable DOCX or read-only PDF only after verified success | provider bytes as authority |
| WF-07 `Save title/source` | Controlled Documents | `updateRevisionDraft` | local accepted fields + CSRF + **If-Match** | DocumentWorkView + new ETag | `precondition.draft_changed`, conflict, upload expired, validation | success replace Work query; stale→WF-09 with local buffer preserved | LWW/auto merge |
| WF-08 `Allocate upload` | Controlled Documents | `startRevisionDraftUpload` | expected_size_bytes + CSRF | DraftUploadAllocation | size/validation/conflict/dependency | retain local bytes; retry new allocation when safe | allocation as semantic content fact |
| WF-08 browser upload | ManagedContentStore capability mechanism | opaque provider PUT | exact local bytes + returned browser-settable required headers; capability URL opaque | provider object may exist only | transport/provider failure | retry according to capability validity; no Product success yet | parsing provider URL; PUT=READY/WorkingContent |
| WF-08 `Complete upload` | Controlled Documents | `completeRevisionDraftUpload` | upload_id + CSRF | 204 means READY completion semantics only | `state.upload_expired`, `validation.content_invalid`, dependency | expired→new allocation + reupload same intended bytes; READY still not attached | reviving old upload id |
| WF-08 attach admitted source | Controlled Documents | `updateRevisionDraft` | `source_upload_id` + current **If-Match** | authoritative DocumentWorkView | draft_changed/upload_expired/conflict | success means attachment; stale→explicit reconcile | READY=attached before PATCH |
| WF-07 `Submit` | Controlled Documents | `createSubmission` | revision_id + CSRF + **Idempotency-Key + DRAFT If-Match** | SubmissionCreateResult | draft_changed, content invalid/malicious, scanner unavailable, conflict, key reused, ambiguous transport | pending→WF-10; released→Official; ambiguous retry SAME key + semantic condition | optimistic SUBMITTED/Release; new key for same ambiguous command |
| WF-10 load Submission | Controlled Documents | `getSubmission` | current_submission_id from current Revision/result | SubmissionView | 404/state/disclosure | render orthogonal gates/termination | client gate/lifecycle synthesis |
| WF-10 load submitted source | Controlled Documents | `getSubmissionSource` | submission_id | exact immutable bytes | integrity/dependency/notfound | read-only viewer | mutation of submitted content |
| WF-10 `Withdraw` | Controlled Documents | `withdrawSubmission` | submission_id + CSRF | SubmissionWithdrawalView/exact repeat | conflict/denied/notfound | refetch Document/Revision/Submission/Work/History; current resolver decides next route | treating repeat as new lifecycle action |
| WF-07/10 `Cancel revision` | Controlled Documents | `cancelRevision` | reason + CSRF | RevisionCancellationView/exact repeat | conflict/denied/notfound/validation | refetch current Document; resolver chooses Work vs Official | client decides cancellation currentness |

## 6. My Work / Governance

| Surface / control | Owner | operationId | Input / wire | Success truth | Material failure | Client consequence / retry | Forbidden |
|---|---|---|---|---|---|---|---|
| WF-11 Authoring lane | Controlled Documents | `listAuthoringWork` | cursor/limit | WorkAuthoringPage | auth/dependency | rows route by document_id; stale target refetch | projection as lifecycle authority |
| WF-11 Governance lane | Controlled Documents | `listGovernanceWork` | cursor/limit | WorkGovernancePage | auth/dependency | rows route by attempt_id | projection as participation grant |
| WF-12/13 case load | Controlled Documents | `getGovernanceAttempt` | attempt_id | GovernanceCaseView | denied/notfound/state | render exact subject/steps/actions | Work projection or role as case authority |
| WF-12 feedback continuation | Controlled Documents | `listGovernanceFeedback` | attempt_id + opaque cursor | GovernanceFeedbackPage | cursor/notfound/denied | append/replace page | offset/total/frozen snapshot |
| WF-12/13 `Add feedback` | Controlled Documents | `createGovernanceFeedback` | message + CSRF + **Idempotency-Key** | feedback id/result | validation/conflict/denied/key reused/ambiguous transport | refetch case/feedback; ambiguous retry SAME key | duplicate feedback on retry |
| WF-12/13 decision read | Controlled Documents | `getGovernanceStepDecision` | attempt_id + step_id | GovernanceDecisionView | 404/denied | display exact singleton | absence interpreted as permission |
| WF-12/13 `Accept` / `Return for changes` | Controlled Documents | `recordGovernanceStepDecision` | closed outcome + required reason; CSRF | GovernanceDecisionView first/exact repeat | `state.governance_step_already_decided`, conflict, denied, validation | remain case; refetch case/Work/Document/History | client publish/Release; editable decided fact |

## 7. History / Audit

| Surface / control | Owner | operationId | Input / wire | Success truth | Material failure | Client consequence | Forbidden |
|---|---|---|---|---|---|---|---|
| WF-14 history page/load more | Controlled Documents | `getDocumentHistory` | document_id + cursor | DocumentHistoryPage | denied/notfound/cursor | timeline append/replace | history as current-state resolver |
| WF-14 inspect historical Revision | Controlled Documents | `getRevision` | revision_id from exact history item | RevisionView | denied/notfound | inline read-only detail | reconstructing missing ids |
| WF-14 inspect historical Submission/source | Controlled Documents | `getSubmission` / `getSubmissionSource` | submission_id from history item | immutable metadata/exact bytes | denied/notfound/integrity | inline drawer/viewer | content mutation |
| WF-14 inspect Release/source | Controlled Documents | `getRelease` / `getReleaseSource` | release_id from history item | ReleaseView/exact bytes | denied/notfound/integrity | inline viewer | current official inference |
| WF-14 inspect OfficialRendition | Controlled Documents | `getOfficialRenditionContent` | rendition id from history item | exact PDF | denied/notfound/integrity | inline viewer | generated preview as authority |
| WF-14 inspect ObsolescenceRequest | Controlled Documents | `getObsolescenceRequest` | request id from history item | request view | denied/notfound | inline detail | using it as current active request resolver |
| WF-14 `Open case` | Controlled Documents | destination `getGovernanceAttempt` | exact governance_attempt_id from history item | GovernanceCaseView | denied/notfound | navigate GOV route | inferring attempt id from chronology |
| WF-15 audit page/load more | Audit | `listAuditEvents` | cursor/limit only | AuditEventPage | denied/cursor | evidence table append/replace | generic filter/search/export/current-state lookup |

## 8. Admin / Organization

| Surface / control | Owner | operationId | Input / wire | Success truth | Material failure | Client consequence / retry | Forbidden |
|---|---|---|---|---|---|---|---|
| WF-16 company load | Organization | `getCompany` | safe read | CompanyView + ETag | denied | form binds ETag | local settings truth |
| WF-16 `Save company` | Organization | `replaceCompany` | form + CSRF + **If-Match** | CompanyView + ETag | stale/validation/denied | replace query; stale preserves input + reconcile | silent overwrite |
| WF-17 Users list | Organization | `listUsers` | cursor/limit | UserPage | denied | current directory | treating profile as identity |
| WF-17 provider subject search | Authentication | `searchProviderSubjects` | bounded SearchQuery | ProviderSubjectSearchView | denied/dependency | selection preflight only | provider role/group authority |
| WF-17 `Create user` | Organization + Authentication composition | `createUser` | provider ref + profile + CSRF + **Idempotency-Key** | CreateUserResult | validation/conflict/key reused/dependency/ambiguous | add/refetch user; ambiguous SAME key | half-created User/Profile/Binding |
| WF-18 load stable User | Organization | `getUser` | user_id from list/create | UserView | denied/notfound | drawer identity state | profile display as stable identity |
| WF-18 profile load | Organization | `getUserProfile` | user_id | UserProfileView + ETag or 404 absent | denied/notfound | bind present/absent state | missing profile=missing User |
| WF-18 `Save/recreate profile` | Organization | `replaceUserProfile` | form + CSRF + **If-Match** or **If-None-Match:\*** exact matrix | UserProfileView + ETag | stale/validation/denied | replace profile; explicit conflict reconcile | both/neither conditions; client null-clear convention |
| WF-18 `Erase profile` | Organization | `deleteUserProfile` | CSRF | 204 | absent/notfound/denied | remove profile query only | deleting stable User identity |
| WF-18 binding load | Authentication | `getUserProviderBinding` | user_id | binding + ETag | denied/notfound | bind exact current ETag | provider ref as Product User identity |
| WF-18 `Replace binding` | Authentication | `replaceUserProviderBinding` | selected provider ref + CSRF + **If-Match** | binding + ETag | stale/conflict/dependency/denied | replace binding; session consequences handled server-side; current browser may lose session | client attempts session revocation itself |
| WF-18 eligibility load | Organization | `getUserEligibility` | user_id | eligibility + ETag | denied/notfound | bind state | role list as eligibility |
| WF-18 `Disable/Re-enable` | Organization | `replaceUserEligibility` | desired state + CSRF + **If-Match** | eligibility + ETag | stale/denied | replace state; refetch Users/access; own session may be revoked if actor affected | restoring removed grants/memberships client-side |
| WF-19 Areas list | Organization | `listAreas` | cursor/limit | AreaPage | denied | list refs | Admin list as Product scope authority elsewhere |
| WF-19 `Create area` | Organization | `createArea` | code/name + CSRF + **Idempotency-Key** | CreateAreaResult | validation/conflict/key reused/ambiguous | add/refetch; ambiguous SAME key | client code normalization authority beyond contract |
| WF-19 Area identity load/save | Organization | `getArea` / `replaceArea` | area_id; save CSRF + **If-Match** | AreaView + ETag | stale/validation/denied | replace exact domain | editing immutable code after creation |
| WF-19 Area lifecycle load/save | Organization | `getAreaLifecycle` / `replaceAreaLifecycle` | area_id; CSRF + **If-Match** | AreaLifecycleView + ETag | stale/denied | replace independent lifecycle domain | flattening metadata+lifecycle ETags |
| WF-19 Groups list | Organization | `listGroups` | cursor/limit | GroupPage | denied | list refs | group as provider group |
| WF-19 `Create group` | Organization | `createGroup` | name + CSRF + **Idempotency-Key** | CreateGroupResult | validation/key reused/ambiguous | add/refetch; ambiguous SAME key | duplicate group on retry |
| WF-19 Group load/save | Organization | `getGroup` / `replaceGroup` | group_id; save CSRF + **If-Match** | GroupView + ETag | stale/denied | replace exact domain | membership embedded in Group ETag |
| WF-19 `Delete group` | Organization | `deleteGroup` | group_id + CSRF | 204 | live dependency conflict/denied/notfound | refetch Groups/access/config | deleting despite live dependency |

## 9. Admin / Access

| Surface / control | Owner | operationId | Input / wire | Success truth | Material failure | Client consequence / retry | Forbidden |
|---|---|---|---|---|---|---|---|
| WF-20 Group selection | Organization | `listGroups`/`getGroup` | returned refs | current Group | denied/notfound | select only returned id | provider group identity |
| WF-20 members page | Organization | `listGroupMembers` | group_id + cursor | GroupMemberPage | denied/notfound | current member list | membership history resource |
| WF-20 candidate User refs | Organization | `listUsers` | current Admin access context | UserPage | denied | selector only | frontend infers access.manage→organization.manage as authorization rule |
| WF-20 `Add member` | Organization state protected by access | `addGroupMember` | group_id/user_id + CSRF | 201 first / 204 existing relation | conflict/denied/notfound | refetch members/affected access | client grants authority itself |
| WF-20 `Remove member` | Organization state protected by access | `removeGroupMember` | group_id/user_id + CSRF | 204 incl. absent relation when parent valid | denied/notfound | refetch members/affected access | treating absent relation as error when wire says success |
| WF-21 Role catalog | Authorization | `listRoles` | safe read | RoleListView fixed vocabulary | denied | display only | role/Permission editor |
| WF-21 Assignments | Authorization | `listRoleAssignments` | cursor | RoleAssignmentPage | denied | current assignments | client-expanded ACL authority |
| WF-21 `Grant` | Authorization | `createRoleAssignment` | admitted refs + role/scope + CSRF + **Idempotency-Key** | assignment id | conflict/validation/key reused/ambiguous | add/refetch; ambiguous SAME key | frontend role hierarchy or inferred scope legality |
| WF-21 `Revoke` | Authorization | `deleteRoleAssignment` | assignment_id + CSRF | 204 | denied/notfound | remove/refetch | resurrecting assignment locally |

## 10. Admin / Document Governance

| Surface / control | Owner | operationId | Input / wire | Success truth | Material failure | Client consequence / retry | Forbidden |
|---|---|---|---|---|---|---|---|
| WF-22 Types list | Controlled Documents | `listDocumentTypes` | cursor | DocumentTypePage | denied | current config list | generic config repository |
| WF-22 `Create type` | Controlled Documents | `createDocumentType` | base + governance + representation + CSRF + **Idempotency-Key** | type id | validation/conflict/key reused/ambiguous | select/refetch new type; ambiguous SAME key | implicit governance default |
| WF-22 type base load/save | Controlled Documents | `getDocumentType` / `replaceDocumentType` | id; save CSRF + **If-Match** | DocumentTypeView + ETag | stale/conflict/validation | replace exact base domain | mutating frozen code/scope by client assumption |
| WF-23 governance load/save | Controlled Documents | `getDocumentTypeGovernance` / `replaceDocumentTypeGovernance` | closed policy; CSRF + **If-Match** | governance view + ETag | stale/conflict/validation | replace governance domain; reconcile route order explicitly | generic workflow/policy language |
| WF-24 eligible Templates load/save | Controlled Documents | `getDocumentTypeEligibleTemplates` / `replaceDocumentTypeEligibleTemplates` | closed set + CSRF + **If-Match** | EligibleTemplatesView + ETag | stale/conflict/validation | replace set; refetch template config/options | array order as semantic set order |
| WF-22 numbering preview | Controlled Documents | `getDocumentTypeNumberingPreview` | type + optional area | preview reservation=false | validation/notfound | display only | reservation/finality |
| WF-25 Template config list | Controlled Documents | `listTemplateConfigurations` | cursor | TemplateConfigurationPage | denied | selection refs/current config | content/history access inference |
| WF-25 template-role load/save | Controlled Documents | `getDocumentTemplateRole` / `replaceDocumentTemplateRole` | document_id; CSRF + **If-Match** | TemplateRoleView + ETag | stale/conflict/denied | replace exact role; refetch template config/options | Template peer lifecycle |

## 11. Idempotency census reconciliation

Exactly the accepted 10 creations carry client-generated logical-command keys:

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

```text
ledger coverage     10 / 10
extra keyed action   0
missing keyed action 0
```

Same logical command after ambiguous transport outcome reuses the same key. Changed/new semantic command receives a new key.

## 12. ETag/concurrency reconciliation

Exactly the accepted 13 frontend-reachable read/mutation domains retain distinct current validators:

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

```text
ledger domains covered  13 / 13
merged client domains     0
invented client domains   0
```

Responsible-owner **candidate projection is not an ETag domain**.

## 13. Exact-byte reconciliation

Exactly the accepted four application byte resources appear in material viewer/editor interactions:

```text
getRevisionDraftSource
getSubmissionSource
getReleaseSource
getOfficialRenditionContent
```

```text
exact-byte resources covered  4 / 4
provider semantic URLs added  0
Range/partial authority added 0
```

## 14. F6 closure

```text
material wireframe controls/actions    traced
idempotent creations                    10 / 10
ETag domains                            13 / 13
exact-byte resources                     4 / 4
client lifecycle authority               0
frontend Authorization engine            0
screen-shaped application operation      0
application operations                  78
operation 79                            absent
unresolved MATERIAL F6 finding            0
F6 status                               COMPLETE CANDIDATE
next                                    F7 bidirectional trace
```
