# R10-T6 — Canonical API / Frontend Journeys

> **Status:** ACTIVE / OPERATOR-RATIFIED TECHNICAL AUTHORITY  
> **Ratified:** 2026-08-18  
> **Repository:** `developmentconexus-ops/MetalDocs`  
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131  
> **Product authority:** `wiki/architecture/launch-v1-product-contract.md` — REV001  
> **T1→T5:** OPERATOR-RATIFIED  
> **Bounded T3 precision:** `wiki/architecture/r10-t3-d4-responsible-owner-eligibility-amendment.md`  
> **Ratified source snapshot:** T6 Platform Summary REV2, blob `5f3f0ec93bf94f586eafd341d72ec484ef2ec848`  
> **Final exact delta:** D1→D4 `APPROVE / NEW MATERIAL FINDINGS = 0`  
> **Implementation:** BLOCKED

This page is the durable T6 authority promoted from the operator-ratified Platform Summary REV2 after the pre-ratification Global Coherence Review, C1→C8 corrections, L1→L5 refinements, D1→D4 bounded precision corrections and final exact delta review.

The current implementation is evidence only. The target has no compatibility obligation to preserve legacy routes, modules, DTOs, screens, provider integrations or sunk-cost abstractions.

T6 does not reopen T1→T5 except the already-ratified D4 precision recorded in the bounded T3 amendment. Everything not explicitly stated here remains governed by Product Contract REV001, Whole-Product GCR A1→A10, 4+1 ownership, T1→T5 and the Decision Registry.

---

## 1. Platform model

MetalDocs Launch is one controlled-document product with exactly four business semantic owners plus Audit:

```text
Authentication
Organization
Authorization
Controlled Documents
Audit (supporting semantic evidence)
```

Frontend/API surfaces are semantic lenses over those owners, never new semantic owners:

```text
Library              → current official/effective discovery truth
My Work              → actor-relevant authoring/governance work
Document Official    → stable Document + current EFFECTIVE truth
Document Work        → exact current open Revision DRAFT
Governance Case      → exact GovernanceAttempt + exact immutable governed subject
Document History     → Controlled Documents history
Audit                → AuditEvent action evidence
Administration       → Organization / Access / Document Governance
```

Launch has no peer product authority named Approval, Templates, Artifact, Distribution, Taxonomy, Tokens or Notifications.

A read model may denormalize bounded labels/references for UX, but never becomes mutation/current-state authority.

---

## 2. HTTP surface classes

MetalDocs exposes three distinct HTTP surface classes.

### 2.1 Application API

```text
/api/v1
```

`api/openapi/v1/openapi.yaml` is the contract SSOT for the `/api/v1` application API.

Launch contract laws:

```text
OpenAPI feature set          3.0.3
generated Go boundary        required
generated TypeScript types   required
JSON fields                  snake_case
technical identifiers        opaque UUIDs
trusted instants             RFC3339 UTC
```

The pre-launch `/api/v1` contract is rebuilt in place from current semantics. There is no `/api/v2` migration namespace and no compatibility shim for removed legacy routes.

OAS 3.0.3 remains the Launch feature set because no current consumer requires 3.1 semantics. Reopen the OAS version only on a concrete schema/tooling benefit.

### 2.2 Browser AuthN integration

```text
GET /auth/login
GET /auth/callback
```

These are fixed OIDC/browser integration routes, not JSON business resources and not part of the generated `/api/v1` product census.

### 2.3 Operations surface

Liveness, readiness, metrics and bounded diagnostics are operational, not product semantics.

They must:

```text
never become business authority
never bypass product authorization to business data
reflect T4 non-serving restore readiness truthfully
provide the minimum activated-effect observability required by T5
```

Exact operations routes/network exposure remain implementation/deployment design.

---

## 3. Authentication and browser trust boundary

MetalDocs never receives or owns end-user passwords.

Browser authentication:

```text
GET /auth/login
→ create OIDC state + PKCE transaction
→ Keycloak Authorization Code flow
→ GET /auth/callback
→ verify state/code/provider
→ resolve issuer + subject
→ ProviderSubjectBinding
→ require existing ENABLED MetalDocs User
→ create fresh ApplicationSession
→ set secure session cookie
→ redirect SPA
```

No Direct Grant/ROPC, no MetalDocs password login/change-password and no JIT User creation. Unknown/unbound provider subjects fail closed.

Session cookie baseline:

```text
name        __Host-metaldocs_session
Secure      true
HttpOnly    true
SameSite    Lax
Path        /
Domain      absent
```

Application session surface:

```text
GET    /api/v1/session
DELETE /api/v1/session
```

Unsafe same-origin browser application requests require a valid session-bound CSRF token. SameSite is defense-in-depth, not the sole CSRF control. There is no permissive cross-origin CORS baseline.

Provider roles/groups/organizations/claims never become MetalDocs Authorization.

---

## 4. Provider directory, User creation and provider-binding replacement

Provider-directory lookup is external preflight:

```text
GET /api/v1/authentication/provider-subjects?query=...
→ opaque provider_subject_ref + bounded display hints
```

After provider subject resolution, successful User creation is one local business transition:

```text
BEGIN
→ User
→ required UserProfile
→ ProviderSubjectBinding
→ required Audit
COMMIT
```

No successful operation may expose a half-created User/binding combination. Authentication and Organization remain separate semantic owners despite same-transaction composition for this invariant.

Provider-binding replacement is current security-sensitive truth:

```text
GET /api/v1/users/{user_id}/provider-binding
→ strong ETag

PUT /api/v1/users/{user_id}/provider-binding
If-Match: <current ETag>
```

The replacement subject is resolved before the local transaction. The transaction revalidates the expected current binding, replaces it, invalidates existing ApplicationSessions as required and appends Audit.

Stale replacement:

```text
412 precondition.resource_changed
zero mutation
```

---

## 5. Frontend information architecture

Top-level Launch spaces:

```text
Library
My Work
Audit           when authorized
Administration  when authorized
```

Stable route meanings:

```text
/documents
  ordinary official discovery lens

/documents/:document_id
  stable Document official/management lens; current EFFECTIVE truth primary

/documents/:document_id/work
  authorized current open Revision work lens

/documents/:document_id/history
  authorized Controlled Documents history

/work
  actor work workspace

/work/governance/:attempt_id
  exact GovernanceAttempt / exact governed case

/audit
  AuditEvent inspection

/admin/organization
/admin/access
/admin/document-governance
```

A route never silently switches between EFFECTIVE and DRAFT content based on caller identity. Visual components may be reused; semantic route meaning does not change.

No top-level Launch workspace exists for Approvals, Templates, Distribution, Notifications, Taxonomy, Tokens, Sessions or Metrics.

---

## 6. Reader truth, Library, Search and derived status

Ordinary Library discovery defaults to current official/effective truth.

Canonical searchable facts:

```text
Document code
current EFFECTIVE Revision title
Document Type
Area
responsible owner
lens-derived status
```

There is no persisted `Document.current_status` or equivalent second current-state authority.

Status is derived from Revision lifecycle + Release + Obsolescence for the lens that needs it.

`GET /api/v1/documents` defaults to:

```text
status=effective
```

Where the caller has required historical/management authority, bounded explicit catalog modes may include:

```text
effective
obsolete
cancelled
```

Derived status follows current official truth. Example:

```text
REV000 EFFECTIVE
REV001 CANCELLED
→ catalog status = effective
```

DRAFT/SUBMITTED are not ordinary Library official results; their work states belong My Work / Document Work / Governance lenses.

Free text:

```text
q = code + current EFFECTIVE title
```

Deterministic ranking with `q`:

```text
exact code
→ code prefix
→ title prefix
→ title contains
→ code + stable id tie-break
```

Without `q`:

```text
code + stable id
```

Typed filters include Document Type, Area, responsible owner and lens-valid status.

Search infrastructure remains:

```text
materialized Search projection = OFF
search_refresh                  = OFF
full Search rebuild             = N/A
external Search engine          = OFF
```

Full-text body/OCR/vector search is a future material reopen. Search never grants access or establishes effectivity.

---

## 7. Numbering and code authority

Numbering is closed and deliberately non-generic:

```text
numbering_scope = DOCUMENT_TYPE | DOCUMENT_TYPE_AREA
separator       = '-'
sequence        = decimal; minimum display width 3; expands naturally after 999
```

Examples:

```text
PO-001
PO-RH-003
PO-RH-1000
```

No configurable separator, padding width, year/reset token or expression language.

Normalized `DocumentType.code` and `Area.code` use:

```text
trim
uppercase ASCII alphanumeric
'-' forbidden because product owns separator
bounded length; exact API maximum belongs implementation-contract design absent stronger evidence
```

Within one Company:

```text
normalized DocumentType.code = unique
normalized Area.code         = unique
committed Document.code      = unique and never reused
```

Area code is immutable after creation. DocumentType code + numbering scope become immutable after the first committed Document uses that type.

Preview:

```text
GET /api/v1/document-types/{document_type_id}/numbering-preview?area_id=...
→ preview_code
→ reservation=false
```

Preview writes no sequence state. Atomic Document create returns final code authority. Gaps are allowed.

---

## 8. Least-privilege document-creation options

Create/revise UI must not use administrative directories merely to populate selectors.

Purpose-built projection:

```text
GET /api/v1/document-creation/options
    ?area_id=<optional>
    &document_type_id=<optional>
```

The projection is current derived read truth, not a new owner or generic reference-data platform.

Base admission:

```text
caller = ENABLED authenticated User
+ document.create in at least one relevant Company/Area scope
```

It returns only bounded references usable by the actor.

Area references:

```text
Company-scoped document.create
→ current ACTIVE Areas eligible for creation

Area-scoped document.create
→ only matching current ACTIVE Areas
```

DocumentType references return only current ACTIVE types usable for the permitted journey and never the admin write model.

Template option requires all:

```text
current Template role
current eligibility for selected DocumentType
selected Template Revision = current EFFECTIVE
caller has current document.read_effective for Template scope
```

Responsible-owner candidates are returned only when caller has `document.owner.manage` in the selected scope.

Candidate eligibility is the ratified D4 law:

```text
existing MetalDocs User
+ same Company
+ current eligibility = ENABLED
```

Only bounded `UserReference` data needed for selection is exposed, not a general UserProfile/PII directory.

Ordinary authors without `document.owner.manage` receive no owner-candidate list; create defaults the actor as responsible owner.

All selected references/permissions are revalidated during actual create. Options are UX guidance, never grants.

---

## 9. Create journey and seed semantics

Create journey:

```text
load least-privilege creation options
→ choose allowed DocumentType
→ choose allowed Area
→ enter REV000 title
→ optionally choose eligible current-EFFECTIVE Template
→ responsible owner defaults to actor unless document.owner.manage permits another target
→ show non-reserving code preview
→ prepare required content copy outside semantic tx
→ atomic create after commit-time revalidation
→ open Document Work lens
```

Blank creation seeds WorkingContent from a trusted product-owned blank DOCX mechanism asset. The blank asset is not a semantic Template Document.

Template create:

```text
read selected current EFFECTIVE Template Revision/source
→ create independent T4 managed copy outside semantic tx
→ BEGIN
→ revalidate actor create authorization + scope
→ revalidate DocumentType/Area eligibility
→ revalidate Template role + target-type eligibility
→ revalidate selected Template Revision still current EFFECTIVE
→ revalidate copied descriptor/provenance matches selected source
→ create stable Document + REV000 + WorkingContent + DocumentOrigin
→ required Audit
→ COMMIT
```

Derived Document never reuses Template storage identity and never seeds from OfficialRendition. Later Template changes never rebind it.

---

## 10. Create next Revision and external-copy race law

A new Revision seeds from:

```text
current EFFECTIVE Revision title
+ exact winning released source
```

Provider/content copy cannot occur inside semantic transaction:

```text
read candidate current EFFECTIVE Revision + exact released source
→ create independent managed copy outside semantic tx
→ BEGIN Document-serialized local tx
→ revalidate actor document.edit + relationship/scope authority
→ revalidate selected source Revision still current EFFECTIVE
→ revalidate no T2 conflict blocks next-Revision creation
→ revalidate copied descriptor/provenance matches exact released source
→ create next Revision + independent WorkingContent
→ required Audit
→ COMMIT
```

If revalidation fails, no Revision is created; copied mechanism bytes are reclaimable.

The new Revision does not rebind to Template and never starts from OfficialRendition.

---

## 11. Responsible-owner semantics

Responsible owner is a Controlled Documents current relationship, not an access grant.

New assignment eligibility:

```text
existing MetalDocs User
+ same Company
+ current eligibility = ENABLED
```

Assignment:

```text
does NOT grant Role
does NOT grant Permission
does NOT depend on provider roles/groups
does NOT require pre-existing document.edit
does NOT itself grant read_working/edit/submit
```

Actual actions continue to require current T3 grant + scope + Controlled Documents relationship/state predicates.

Create-time deliberate owner selection and later owner replacement require `document.owner.manage` in matching scope.

Owner replacement:

```text
GET /api/v1/documents/{document_id}/responsible-owner
→ strong ETag

PUT /api/v1/documents/{document_id}/responsible-owner
If-Match: <current ETag>
```

Assignment serializes with target User eligibility/offboarding. If offboarding linearizes first, assignment sees DISABLED and fails closed.

---

## 12. DRAFT editing and OCC

Revision title remains Revision-owned metadata. Title and source share one T2 DRAFT race authority: WorkingContent generation.

```text
GET /api/v1/revisions/{revision_id}/draft
→ ETag: "draft-<generation>"

PATCH /api/v1/revisions/{revision_id}/draft
If-Match: "draft-<expected_generation>"
```

Closed mutable fields:

```text
title             optional set-to value
source_upload_id  optional newly READY upload reference
```

At least one is required. Omitted means unchanged. Null has no implicit delete meaning.

Server:

```text
revalidate If-Match
→ if source_upload_id present, prove T4 READY + live binding and load server-owned descriptor
→ atomically apply accepted title/source changes
→ generation++ exactly once
→ return updated DocumentWorkView + new strong ETag
```

Stale:

```text
412 precondition.draft_changed
zero mutation
```

No second title version, silent LWW, auto-merge or baseline CRDT. Autosave creates no business Revision/Submission/history.

---

## 13. Exact-content upload/admission

Frontend orchestrates T4 and never becomes content authority.

```text
POST /api/v1/revisions/{revision_id}/draft/uploads
→ OPEN upload_id + live admission binding + create-only upload target

browser uploads exact bytes directly to provider

POST /api/v1/revisions/{revision_id}/draft/uploads/{upload_id}/complete
→ server independently reads exact stored bytes
→ derives SHA-256 + size + ContentFormat
→ structural validation
→ READY

PATCH /api/v1/revisions/{revision_id}/draft + If-Match
→ references upload_id
→ server proves READY + live binding
→ server loads its own descriptor
→ attach to WorkingContent
```

Truth ladder:

```text
provider upload success != READY
READY != WorkingContent
WorkingContent != Submission
Submission != Release
```

Client never supplies authoritative exact-content descriptor/hash.

Untrusted external bytes must pass exact-byte malware inspection before immutable governed admission. DRAFT autosave need not scan every debounce.

```text
scanner unavailable → 503 dependency.malware_inspector_unavailable
malicious bytes      → 422 validation.content_malicious
```

No business `scan_status` lifecycle exists.

---

## 14. Complete Launch lifecycle journeys

### Blank/template create

```text
stable Document
→ REV000 DRAFT
→ WorkingContent
→ not official
```

### NoHumanApproval + SourceOnly SUBMIT

```text
DRAFT
→ validate expected generation/current config
→ freeze immutable Submission
→ no GovernanceAttempt / no fake Step
→ representation gate absent
→ system Release may commit in same local transaction
→ Revision EFFECTIVE
```

Submission and Release remain separate semantic facts even when committed together.

### Human-governed SUBMIT

```text
DRAFT
→ freeze immutable Submission
→ Revision SUBMITTED
→ freeze GovernanceAttempt route snapshot
→ exactly one active Step at a time
```

Reviewer judges exact Submission, never current WorkingContent.

### RETURN_FOR_CHANGES

```text
immutable Decision/reason
→ terminate attempt as returned
→ same Revision SUBMITTED → DRAFT
→ old Submission/Decision/feedback remain immutable
```

### Submission withdrawal

```text
active pre-Release attempt
→ WITHDRAWN
→ same Revision → DRAFT
→ old history preserved
→ no fake RETURN
```

### Revision cancellation

```text
eligible DRAFT or pre-Release SUBMITTED Revision
→ terminate live attempt without fake participant verdict
→ immutable RevisionCancellation
→ Revision CANCELLED terminally
```

An older EFFECTIVE Revision remains EFFECTIVE; cancelled ordinal never reuses.

### Required OfficialRendition pending

```text
human gate satisfied/absent
+ frozen policy requires OfficialRendition
+ rendition absent
→ Revision remains SUBMITTED
→ Release absent
```

Renderer/job state is mechanism only. No `RENDERING`/`RENDER_FAILED` business lifecycle states exist.

Late renderer output for returned/withdrawn/cancelled candidates is semantic no-op/reclaimable mechanism output.

### First Release

```text
REV000 SUBMITTED + all gates satisfied
→ system Release
→ REV000 EFFECTIVE
```

No user publish command exists.

### Replacement Release

```text
prior EFFECTIVE + eligible successor winning Submission
→ one Document-serialized transaction
→ predecessor SUPERSEDED
→ successor EFFECTIVE
→ Release binds exact winning Submission + predecessor
```

No successful observable state has two EFFECTIVE Revisions.

### Human-governed obsolescence

```text
current EFFECTIVE target
+ mandatory reason
+ no open replacement
+ no competing request
→ active GovernanceAttempt
→ target remains EFFECTIVE during governance
```

Final ACCEPT revalidates and changes target to OBSOLETE. RETURN or authorized request withdrawal ends the request while target remains EFFECTIVE; withdrawal creates no fake RETURN.

### NoHumanApproval obsolescence

```text
authorized initiation + reason + eligibility/conflict checks
→ zero human Step
→ no fake System approver
→ same transaction changes current EFFECTIVE target → OBSOLETE
```

Successful obsolescence removes the Document from ordinary current-effective Library while authorized history remains available.

---

## 15. Governance case, feedback and Step Decision

Governance Case reads exactly:

```text
GovernanceAttempt
+ exact active Step
+ exact immutable Submission or exact obsolescence subject
+ bounded prior decisions/feedback
+ live server-derived allowed_actions
```

Case participation opens only the exact governed context required for action. It never grants general WorkingContent/history authority and never permits reviewer WorkingContent mutation.

A Step has at most one immutable Decision:

```text
GET /api/v1/governance-attempts/{attempt_id}/steps/{step_id}/decision
PUT /api/v1/governance-attempts/{attempt_id}/steps/{step_id}/decision
```

```text
outcome = ACCEPT | RETURN_FOR_CHANGES
reason  = required for RETURN_FOR_CHANGES
```

Exact repeat returns existing result. Different later outcome:

```text
409 state.governance_step_already_decided
```

No Idempotency-Key replay row is required for this singleton fact.

Submission feedback remains separate immutable multi-record fact and may use POST.

No generic signoff/quorum/reassign/delegation/SLA/workflow API exists.

---

## 16. User eligibility/offboarding

```text
GET /api/v1/users/{user_id}/eligibility
PUT /api/v1/users/{user_id}/eligibility
If-Match: <current eligibility ETag>

state = ENABLED | DISABLED
```

`ENABLED → DISABLED` executes the T3 offboarding transaction:

```text
disable User
+ revoke all live ApplicationSessions
+ remove current GroupMemberships
+ remove direct User RoleAssignments
+ required Audit
```

`DISABLED → ENABLED` restores eligibility only; old sessions/memberships/grants never resurrect.

Same desired state is an idempotent no-op and creates no duplicate semantic Audit transition.

---

## 17. Access administration and GroupMembership read

Group identity remains Organization-owned. GroupMembership mutation is protected by `access.manage` because membership changes effective authority.

Current read:

```text
GET /api/v1/groups/{group_id}/members?cursor=<opaque>&limit=<n>
```

Authorization:

```text
current ENABLED User
+ access.manage @ Company
```

Response is cursor-paginated bounded current membership/User reference information sufficient to administer access. It is not a general UserProfile directory and creates no GroupMembership history resource.

Mutation:

```text
PUT    /api/v1/groups/{group_id}/members/{user_id}
DELETE /api/v1/groups/{group_id}/members/{user_id}
```

Current membership read does not transfer GroupMembership ownership from Organization to Authorization.

---

## 18. Natural HTTP idempotency and crash-consistent Idempotency-Key

Prefer semantic resource identity/natural HTTP idempotency before replay machinery.

Natural singleton/idempotent surfaces include:

```text
User eligibility PUT
Group membership PUT/DELETE
Governance Step Decision PUT
Submission withdrawal PUT
Revision cancellation PUT
obsolescence withdrawal PUT
current config replacement PUT + If-Match
DRAFT PATCH + If-Match
```

Durable `Idempotency-Key` is required only for genuinely non-idempotent semantic POST creation where a lost response could create another semantic fact/resource:

```text
POST /users
POST /areas
POST /groups
POST /role-assignments
POST /document-types
POST /documents
POST /documents/{document_id}/revisions
POST /revisions/{revision_id}/submissions
POST /governance-attempts/{attempt_id}/feedback
POST /documents/{document_id}/obsolescence-requests
```

Scope:

```text
current User id + canonical operation id + Idempotency-Key
```

Fingerprint uses validated semantic command fields.

Crash-consistent law:

```text
BEGIN local business transaction
→ insert/lock scoped Idempotency-Key
→ compare fingerprint if key exists
→ execute/revalidate semantic transition
→ append required Audit / transaction-coupled durable intents
→ persist completed replay result sufficient for exact status/body replay
→ COMMIT
```

```text
semantic fact commits ⇔ completed replay result commits
rollback → neither completed fact nor completed replay result
same key + different fingerprint → 422 validation.idempotency_key_reused
```

Concurrent same-key requests serialize on the key; after winner commit, the other may replay. No baseline public/durable `IN_PROGRESS` replay state exists.

Replay retention is bounded and exceeds ordinary browser/network retry windows. `24h` is only an initial implementation-default candidate, not architecture invariant.

---

## 19. Idempotency replay authorization

A completed replay record is transport mechanism, never access authority.

Every unsafe replay request still passes current request trust boundary:

```text
valid current ApplicationSession
+ valid session-bound CSRF
+ current T3 permission/scope sufficient to receive the operation/result
+ minimum current resource-visibility predicate needed to disclose stored response
```

Only after current access authorization succeeds may completed replay be disclosed.

```text
current access/disclosure authorization = RECHECK
historical mutation eligibility/lifecycle preconditions = DO NOT RE-EXECUTE
business transition = DO NOT RE-EXECUTE
```

The replay path must not manufacture a second business action or reject a historically successful command merely because original lifecycle preconditions are no longer true after completion.

Replay storage must not become an unintended retention root for erasable UserProfile PII; response persistence/redaction must remain compatible with T3/T4 privacy/restore laws.

---

## 20. Release source, exact-byte gateway and OfficialRendition presentation

There is no public Release/publish mutation. Release is system-owned.

Semantic byte resources:

```text
GET /api/v1/revisions/{revision_id}/draft/source
GET /api/v1/submissions/{submission_id}/source
GET /api/v1/releases/{release_id}/source
GET /api/v1/official-renditions/{rendition_id}/content
```

Application contract never exposes provider bucket/key/version or `managed_content_id` as product identity.

Required baseline:

```text
exact full semantic read
current authorization recheck
semantic URL independent of provider location
```

Range read is optional mechanism activated only on real viewer/file-size evidence.

Implementation may stream or use an authorization-checked short-lived provider/CDN redirect without changing semantic URL/authority.

Presentation:

```text
SourceOnly PDF
→ direct PDF view of exact Release source

SourceOnly DOCX
→ read-only interactive DOCX adapter on exact Release source

RequireOfficialRendition(PDF)
→ OfficialRendition PDF primary official in-product view
→ exact Release source separately available/labeled

Governance
→ exact Submission source primary decision content
```

Viewer/preview output never becomes OfficialRendition authority.

---

## 21. Interactive DOCX adapter and separate rendition renderer

Launch selects exactly one interactive DOCX editor/viewer adapter, not one universal renderer.

Boundary:

```text
load exact DOCX bytes
render/edit in browser
emit complete resulting DOCX bytes
read-only inspection mode
no provider-owned durable semantic identity
no provider callback/save state as business truth
```

First interactive candidate is EigenPal-class browser-buffer integration because it naturally fits T4 admission + T2 OCC.

Before provider freeze, representative fidelity corpus must include at least:

```text
styles/paragraph hierarchy
fonts/emphasis
numbered/bulleted lists
complex/merged tables
headers/footers
images
page/section breaks
links
page settings/margins
multi-page documents
save/reopen in Microsoft Word or LibreOffice
```

Hard fail includes silent OOXML loss/corruption, material ordinary-document layout loss, destructive rewrite of unsupported constructs or invalid/unopenable round-trip.

If browser-buffer candidate materially fails, evaluate ONLYOFFICE-class integration against same corpus and prove callback/server state cannot bypass T4 or overwrite newer OCC generation.

No dual interactive editor runtime. No EditorSession/lease correctness dependency baseline.

T5 server-side OfficialRendition renderer/converter is a separate mechanism and may use another product.

---

## 22. Administration and permission ownership

One Admin Center contains three semantic sections.

Organization:

```text
Company current settings
Users + erasable UserProfile
Areas
Groups + memberships
provider subject binding
User eligibility/offboarding
```

Access:

```text
GroupMembership administration
RoleAssignment grant/revoke
subject = USER | GROUP
role = static T3 role
scope = COMPANY | AREA(area_id)
```

Document Governance:

```text
Document Types
active/inactive
numbering/preview
NoHumanApproval | UseGovernanceRoute
ordered sequential Steps
NAMED_USER | GROUP selector
SourceOnly | RequireOfficialRendition(PDF)
Template role/eligibility
```

No custom Role/Permission editor and no generic PolicyVersion/workflow platform.

Permission ownership remains exactly T3:

```text
Company/User/Area/Group identity         → organization.manage
GroupMembership + RoleAssignment         → access.manage
DocumentType route/representation config → document_type.manage
Template role/eligibility                → template_use.manage
```

UI co-location never merges permission authority.

---

## 23. Bounded Template administration read

A governance admin with `template_use.manage` can administer Template role/eligibility without implicit document content/history access.

Bounded configuration projection exposes only minimum selection metadata such as:

```text
Document id
stable code
bounded current display label
Template-role flag
current effective/eligibility indicator needed for config
```

It does not expose source bytes, WorkingContent, full effective content, Submission/history or general document read authority.

Exact content/history remain protected by `document.read_effective`, `document.read_history` or exact governance-case authorization.

No new permission is introduced.

---

## 24. Lost-update protection

Whole-replacement current truth that can suffer meaningful lost update returns strong ETag and requires `If-Match`.

At minimum:

```text
Company
UserProfile
User eligibility
User ProviderSubjectBinding
Area metadata/lifecycle
Group metadata
DocumentType base
DocumentType governance
DocumentType eligible-template set
Document responsible owner
Document Template role
```

Specific singleton APIs include:

```text
GET/PUT /api/v1/users/{user_id}/provider-binding
GET/PUT /api/v1/documents/{document_id}/responsible-owner
GET/PUT /api/v1/documents/{document_id}/template-role
```

Stale:

```text
412 precondition.resource_changed
zero mutation
```

Exact already-current repeat may be no-op and must not fabricate duplicate semantic Audit.

---

## 25. Public error contract

All `/api/v1` application failures use RFC 9457 Problem Details.

Minimum:

```text
type
title
status
detail
instance
code
trace_id
errors[] optional
```

One machine authority:

```text
code = canonical MetalDocs problem id
type = https://errors.metaldocs.io/{code}
```

`type` is mechanically derived.

Closed families:

```text
request.       400
validation.    422
auth.          401
permission.    403
notfound.      404
state.         409
precondition.  412
conflict.      409
dependency.    503
ratelimit.     429
internal.      500
```

No module-named public error family. Frontend branches on `code`, never localized `detail`. Raw provider/scanner/storage errors never leak.

---

## 26. Lists, read models and `allowed_actions`

Potentially unbounded lists:

```text
?cursor=<opaque>&limit=<n>
limit default 20
limit max 100
```

Response:

```json
{
  "items": [],
  "page": {
    "next_cursor": null,
    "has_more": false
  }
}
```

No mandatory total count, offset pagination or generic filter/sort DSL. Cursor binds filter/order semantics.

Purpose-built read models include:

```text
SessionView
DocumentSummary
DocumentOfficialView
DocumentWorkView
DocumentCreationOptionsView
SubmissionView
GovernanceCaseView
DocumentHistoryView
WorkAuthoringItem
WorkGovernanceItem
AuditEventView
GroupMembershipView/UserReference
TemplateConfigurationItem
Admin configuration views
```

No table dumps, universal polymorphic envelope, ArtifactViewModel or ApprovalInstanceDTO becomes target public contract.

`allowed_actions` are UX hints only. They must derive from the same canonical T3 permission/scope components plus Controlled Documents predicates used by command authorization, or a provably shared equivalent. Never maintain a parallel role matrix.

Every command rechecks current canonical truth.

---

## 27. History versus Audit

Document History answers what happened to one controlled Document. It derives from Controlled Documents semantic facts:

```text
Revisions
Submissions
GovernanceAttempts/Steps/Decisions
feedback
withdrawals
cancellations
Releases
OfficialRenditions
obsolescence request/result/withdrawal
```

Audit answers what meaningful actions occurred across the product, who performed them and when.

The UI may cross-link both but never merges them into synthetic lifecycle authority. Audit cannot reconstruct current business state.

---

## 28. Frontend technical organization

Preserve mechanism patterns that survived Structural Inversion:

```text
React SPA
feature-sliced organization
TanStack Query for server state
OpenAPI-generated application types
lib API/client infrastructure
local React state for local UI state
shared domain-agnostic UI/editor/viewer primitives
```

Replace legacy module taxonomy. Target feature vocabulary follows semantic lenses/owners, e.g.:

```text
library
document-work
governance-work
history
audit
admin
auth/shell
shared editor/viewer
```

Frontend feature folders do not mirror Go package names by default.

---

## 29. Closed Launch `/api/v1` operation census

Minor path spelling/operationId normalization during OpenAPI authoring is allowed only when semantic meaning is unchanged. New product route families require a named Product Contract journey or explicit T6 reopen.

A bounded read-symmetry precision accepted during T8-E analysis on 2026-08-20 is durably recorded in `docs/decisions/api-operation-census.md` and adds exactly these two reads to the original 76-operation census:

```text
GET /api/v1/users/{user_id}/profile       getUserProfile
GET /api/v1/areas/{area_id}/lifecycle     getAreaLifecycle
```

These reads expose the canonical current representations already mutated by the existing profile/lifecycle surfaces so clients can obtain strong ETags without deriving or guessing state. This precision introduces no new semantic owner, capability, or route family.

**Current closed application census: 78 operations.** Operation 79 requires a new material Product/T6 reopen; it cannot be inferred from this precision.

Session/AuthN support:

```text
GET    /api/v1/session
DELETE /api/v1/session
GET    /api/v1/authentication/provider-subjects?query=...
```

Organization:

```text
GET    /api/v1/company
PUT    /api/v1/company
GET    /api/v1/users
POST   /api/v1/users
GET    /api/v1/users/{user_id}
GET    /api/v1/users/{user_id}/profile
PUT    /api/v1/users/{user_id}/profile
DELETE /api/v1/users/{user_id}/profile
GET    /api/v1/users/{user_id}/provider-binding
PUT    /api/v1/users/{user_id}/provider-binding
GET    /api/v1/users/{user_id}/eligibility
PUT    /api/v1/users/{user_id}/eligibility
GET    /api/v1/areas
POST   /api/v1/areas
GET    /api/v1/areas/{area_id}
PUT    /api/v1/areas/{area_id}
GET    /api/v1/areas/{area_id}/lifecycle
PUT    /api/v1/areas/{area_id}/lifecycle
GET    /api/v1/groups
POST   /api/v1/groups
GET    /api/v1/groups/{group_id}
PUT    /api/v1/groups/{group_id}
DELETE /api/v1/groups/{group_id}
GET    /api/v1/groups/{group_id}/members
PUT    /api/v1/groups/{group_id}/members/{user_id}
DELETE /api/v1/groups/{group_id}/members/{user_id}
```

Authorization:

```text
GET    /api/v1/roles
GET    /api/v1/role-assignments
POST   /api/v1/role-assignments
DELETE /api/v1/role-assignments/{assignment_id}
```

Document Governance config:

```text
GET  /api/v1/document-types
POST /api/v1/document-types
GET  /api/v1/document-types/{document_type_id}
PUT  /api/v1/document-types/{document_type_id}
GET  /api/v1/document-types/{document_type_id}/governance
PUT  /api/v1/document-types/{document_type_id}/governance
GET  /api/v1/document-types/{document_type_id}/eligible-templates
PUT  /api/v1/document-types/{document_type_id}/eligible-templates
GET  /api/v1/document-types/{document_type_id}/numbering-preview?area_id=...
GET  /api/v1/document-governance/templates
```

Controlled Documents / Work:

```text
GET  /api/v1/document-creation/options
GET  /api/v1/documents
POST /api/v1/documents
GET  /api/v1/documents/{document_id}
GET  /api/v1/documents/{document_id}/responsible-owner
PUT  /api/v1/documents/{document_id}/responsible-owner
GET  /api/v1/documents/{document_id}/template-role
PUT  /api/v1/documents/{document_id}/template-role
POST /api/v1/documents/{document_id}/revisions
GET  /api/v1/documents/{document_id}/history
GET  /api/v1/work/authoring
GET  /api/v1/work/governance
GET   /api/v1/revisions/{revision_id}
GET   /api/v1/revisions/{revision_id}/draft
PATCH /api/v1/revisions/{revision_id}/draft
POST /api/v1/revisions/{revision_id}/draft/uploads
POST /api/v1/revisions/{revision_id}/draft/uploads/{upload_id}/complete
GET  /api/v1/revisions/{revision_id}/draft/source
POST /api/v1/revisions/{revision_id}/submissions
GET  /api/v1/submissions/{submission_id}
GET  /api/v1/submissions/{submission_id}/source
PUT  /api/v1/submissions/{submission_id}/withdrawal
PUT  /api/v1/revisions/{revision_id}/cancellation
GET  /api/v1/governance-attempts/{attempt_id}
GET  /api/v1/governance-attempts/{attempt_id}/feedback
POST /api/v1/governance-attempts/{attempt_id}/feedback
GET  /api/v1/governance-attempts/{attempt_id}/steps/{step_id}/decision
PUT  /api/v1/governance-attempts/{attempt_id}/steps/{step_id}/decision
GET  /api/v1/releases/{release_id}
GET  /api/v1/releases/{release_id}/source
GET  /api/v1/official-renditions/{rendition_id}/content
POST /api/v1/documents/{document_id}/obsolescence-requests
GET  /api/v1/obsolescence-requests/{request_id}
PUT  /api/v1/obsolescence-requests/{request_id}/withdrawal
```

Audit:

```text
GET /api/v1/audit/events
```

No public Release mutation, generic `/actions`, generic Search endpoint, separate Approval lifecycle API or separate Template lifecycle API.

OIDC and operations endpoints remain outside this application census.

---

## 30. Future-evolution seam verification

T6 remains additive-compatible without dormant implementation:

```text
Distribution / Read & Acknowledge → Release + User/Group
Periodic Review                    → stable Document + exact current EFFECTIVE Revision
Dossier                            → stable Document reference
Evidence                           → independent future lifecycle + T4 exact-content seam
Retention / Legal Hold / Disposition → stable governed identities/history; storage stays mechanism
Governed Export                    → semantic IDs + exact descriptors/byte resources
External Repository IMPORT/PUBLISH → target-owner + T4 copy/admission seam
Training/LMS                       → released/effective truth + future Distribution
Multi-document Change Control      → stable Document/Revision transitions
Pooled tenancy                     → stable Company identity, reopenable substrate
CRDT/realtime                      → replace DRAFT mechanism without changing Revision/Submission meaning
```

No future capability justifies dormant Launch modules/tables/permissions/workers/frameworks.

---

## 31. Explicitly removed from Launch target

```text
Tenant/RLS universal product ontology
local password login/change-password
separate Approval product/workspace/API
separate Template lifecycle/API
public ControlledDocument peer object
Artifact semantic owner
Distribution Launch routes/state
Notifications/inbox Launch state
Taxonomy/Dictionary/Tokens platform absent promoted requirement
legacy role/capability-driven frontend authority
writer/editor session correctness dependency
scheduled publish
universal PDF/export paths
reviewer mutation of WorkingContent
materialized Search without consumer
legacy API compatibility layer
```

Old mechanisms survive only when independently justified by current authority; sunk cost is never entitlement.

---

## 32. Implementation proof obligations

Implementation planning/tests must make at least these claims falsifiable:

```text
OIDC auth-code works without MetalDocs password input
unbound provider subject cannot login
unsafe browser operation fails without valid session-bound CSRF
User creation cannot leave half-created User/Profile/Binding
provider-binding replacement is If-Match protected and revokes required sessions
Library status is derived; no Document.currentStatus authority exists
ordinary Library never presents DRAFT/SUBMITTED as official
creation/options exposes only actor-usable bounded references
creation/options never uses admin APIs as authorization bypass
Template option requires effective-read + role/eligibility/current-EFFECTIVE truth
responsible-owner candidate requires owner.manage + D4 ENABLED/same-Company target
responsible-owner relationship does not grant permissions
owner assignment loses race to offboarding when target becomes DISABLED first
GroupMembership list requires access.manage and does not expose general PII
DRAFT title/source share one ETag generation
stale DRAFT mutation = 412 + zero mutation
client descriptor/hash never becomes exact-content authority
provider upload success alone cannot attach WorkingContent
Template create revalidates exact source/eligibility after external copy
next-Revision revalidates current EFFECTIVE source after external copy
review case displays exact immutable Submission content
reviewer case participation never mutates WorkingContent
NoHumanApproval + SourceOnly may Release atomically without fake governance
RETURN/withdraw/cancel preserve immutable prior history
required-rendition pending remains SUBMITTED with mechanism failure separate
replacement Release cannot expose two EFFECTIVE Revisions
active obsolescence leaves target EFFECTIVE until success
NoHumanApproval obsolescence creates no fake human Step
Idempotency semantic fact and replay result commit atomically
same-key concurrent requests serialize/replay without public IN_PROGRESS state
completed Idempotency replay rechecks current access before response disclosure
replay never re-executes historical mutation lifecycle preconditions
replay persistence does not become an unintended UserProfile PII retention root
numbering preview never reserves
DocumentType.code and Area.code are Company-unique after normalization
committed Document code never reuses
provider-binding/responsible-owner/template-role lost update is If-Match protected
Template administration does not leak content/history to governance_admin
allowed_actions uses shared canonical authorization components
Search projection/jobs absent in Launch
Domain History is reconstructible without Audit
Audit never becomes current lifecycle authority
interactive DOCX provider passes fidelity corpus and cannot bypass OCC/T4
interactive editor selection does not constrain separate OfficialRendition renderer
Range support remains optional
no EditorSession correctness dependency baseline
future seams attach without duplicate current authority
```

---

## 33. Reopen triggers

Reopen only the implicated T6 seam on material evidence that:

```text
a real external client requires a transport incompatible with current same-origin web baseline
numbering requires promoted year/reset/custom grammar
browser-buffer DOCX editing fails the representative fidelity/security/scale corpus
selected editor proves a bounded session/lease is required
content-body/full-text/semantic Search becomes promoted Launch scope
canonical PostgreSQL Search cannot meet measured scale/ranking/language needs
compliance requires different source/OfficialRendition presentation
Audit/history separation blocks a named investigation journey
public commands require a different idempotency guarantee
cursor pagination creates demonstrated UX/operability failure
current operation census cannot serve a promoted Product Contract journey
```

Legacy implementation shape is never itself a reopen trigger.

---

## 34. Gate

T6 is closed and operator-ratified.

```text
T6 Canonical API / Frontend Journeys = CLOSED / OPERATOR-RATIFIED
T7 Historical Migration / Cutover    = NEXT STAGE AFTER T6 PROMOTION/CLEANUP
implementation                        = BLOCKED
```

No product code or implementation plan is authorized by T6 closure.