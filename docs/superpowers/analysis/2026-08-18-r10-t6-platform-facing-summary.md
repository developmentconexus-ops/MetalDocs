# R10-T6 — Canonical API / Frontend Journeys — Platform-Facing Summary

> **Status:** ACTIVE STAGING / NON-AUTHORITATIVE — **C1→C8 INCORPORATED / BOUNDED COHERENCE DELTA NEXT**  
> **Date:** 2026-08-18  
> **Repository:** `developmentconexus-ops/MetalDocs`  
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131  
> **Material adjudication:** OPERATOR-APPROVED  
> **Pre-ratification corrections C1→C8:** OPERATOR-APPROVED  
> **Implementation:** BLOCKED  
> **T7:** NOT OPEN

This summary is the implementation-facing expression of the operator-approved T6 architecture after the pre-ratification Global Coherence Review. It defines the target API/frontend platform without depending on the current MetalDocs runtime shape.

It is still staging, not durable authority. Before operator summary ratification, a bounded coherence delta must prove that the incorporated corrections close the review without contradicting Product Contract REV001, Whole-Product GCR, 4+1 ownership, T1→T5 or the Decision Registry.

---

# 1. Platform model and semantic lenses

MetalDocs Launch is one controlled-document product with exactly four business semantic owners plus Audit:

```text
Authentication
Organization
Authorization
Controlled Documents
Audit (supporting semantic evidence)
```

The public product is **not** a collection of peer products called Approval, Templates, Controlled Documents, Distribution, Taxonomy or Tokens.

Frontend/API surfaces are semantic lenses over the accepted T1→T5 facts:

```text
Library              → current official/effective discovery truth
My Work              → actor-relevant authoring/governance work
Document Official    → stable Document + current EFFECTIVE truth
Document Work        → exact current open Revision DRAFT
Governance Case      → exact GovernanceAttempt + exact immutable Submission
Document History     → Controlled Documents history
Audit                → AuditEvent action evidence
Administration       → Organization / Access / Document Governance
```

A read model may denormalize labels or `allowed_actions` for UX, but it never becomes a second business authority.

---

# 2. HTTP surface classes and contract law

MetalDocs has three distinct HTTP surface classes. They must not be conflated.

## A. Application API

```text
/api/v1
```

This is the generated product/application API.

```text
api/openapi/v1/openapi.yaml = /api/v1 application-contract SSOT
OpenAPI 3.0.3 feature set   = Launch baseline
generated Go boundary       = required
generated TypeScript types  = required
JSON casing                 = snake_case
opaque technical IDs        = UUID
trusted instants            = RFC3339/UTC
```

The pre-launch `/api/v1` contract is rewritten in place from current product semantics. There is no `/api/v2` migration namespace and no compatibility shim for the old pre-launch API.

OAS 3.0.3 remains the Launch feature set because no named T6 consumer requires 3.1 semantics. A later OAS-version change requires a concrete schema/tooling benefit, not modernization by itself.

## B. Browser authentication integration

```text
GET /auth/login
GET /auth/callback
```

These are fixed browser/OIDC integration routes, not JSON product resources and not part of the generated `/api/v1` operation census.

## C. Operations surface

Liveness, readiness, metrics and bounded diagnostics are non-product operational surfaces. Exact route/network layout belongs deployment/observability implementation design.

Operations surfaces:

```text
must never become business authority
must never bypass product access to business data
must expose readiness truthfully
must reflect T4 non-serving restore readiness rather than reporting healthy while ordinary serving is blocked
must support the minimum activated-effect visibility required by T5
```

---

# 3. Authentication, ApplicationSession and browser trust boundary

MetalDocs never receives or owns end-user passwords.

Browser authentication:

```text
GET /auth/login
→ server creates OIDC state/PKCE transaction
→ Keycloak Authorization Code flow
→ GET /auth/callback
→ verify state/code/provider
→ resolve issuer + subject
→ ProviderSubjectBinding
→ enabled MetalDocs User
→ create fresh ApplicationSession
→ set secure session cookie
→ redirect SPA
```

No Direct Grant/ROPC. No MetalDocs password login/change-password route. No JIT User creation. Unknown or unbound provider subjects fail closed.

Session cookie baseline:

```text
__Host-metaldocs_session
Secure
HttpOnly
SameSite=Lax
Path=/
no Domain attribute
```

Application session resource:

```text
GET    /api/v1/session
DELETE /api/v1/session
```

Unsafe same-origin browser API requests require a session-bound CSRF token. SameSite is defense-in-depth, not the sole CSRF control. No permissive cross-origin CORS baseline exists.

Provider roles/groups never become MetalDocs Authorization.

## Provider directory and User creation

Provider-directory search is external preflight:

```text
GET /api/v1/authentication/provider-subjects?query=...
→ opaque provider_subject_ref + display hints
```

After the provider subject is resolved, one successful `POST /api/v1/users` creates the intended local product truth as **one local business transition**:

```text
User
+ required UserProfile
+ ProviderSubjectBinding
+ required Audit
```

No successful operation may expose a half-created local User/binding combination. Authentication and Organization retain separate semantic ownership even though one local transaction composes their facts for this invariant.

## Provider-binding replacement

Provider-binding replacement is security-sensitive current truth:

```text
GET /api/v1/users/{user_id}/provider-binding
→ strong ETag

PUT /api/v1/users/{user_id}/provider-binding
If-Match: <current provider-binding ETag>
```

The replacement provider subject is resolved before the local transaction. The local transition revalidates the expected current binding, establishes the new binding, invalidates existing ApplicationSessions as required and appends Audit.

Stale precondition:

```text
412 precondition.resource_changed
zero mutation
```

Provider roles/groups remain irrelevant to T3.

---

# 4. Frontend information architecture

Top-level Launch product spaces:

```text
Library
My Work
Audit           when authorized
Administration  when authorized
```

Stable route meanings:

```text
/documents
  Library / official discovery lens

/documents/:document_id
  stable Document official/management lens; current EFFECTIVE truth primary

/documents/:document_id/work
  authorized current open Revision work lens

/documents/:document_id/history
  authorized Controlled Documents history

/work
  actor work workspace

/work/governance/:attempt_id
  exact GovernanceAttempt + exact immutable Submission case

/audit
  AuditEvent inspection

/admin/organization
/admin/access
/admin/document-governance
```

A URL never silently changes from EFFECTIVE truth to DRAFT truth because a different actor opened it. Visual components may be shared; semantic route meaning stays stable.

No top-level Launch workspace exists for Approvals, Templates, Distribution, Notifications, Taxonomy, Tokens, Sessions or Metrics.

---

# 5. Reader truth, Library, Search and derived status

Ordinary Library discovery defaults to the current official/effective truth.

Searchable canonical facts:

```text
Document code
current EFFECTIVE Revision title
Document Type
Area
responsible owner
lens-derived status
```

No persisted `Document.current_status` or equivalent second current-state authority is introduced. Status is derived from Revision lifecycle + Release + Obsolescence in the lens that needs it.

## Library status semantics

`GET /api/v1/documents` defaults to:

```text
status=effective
```

When the caller has the required historical/management read authority, explicit status filtering may include bounded derived catalog states such as:

```text
effective
obsolete
cancelled
```

Derived status follows current reader truth. Example:

```text
REV000 EFFECTIVE
REV001 CANCELLED
→ Document catalog status = effective
```

because REV000 remains the current official reader truth.

DRAFT/SUBMITTED are **not** ordinary Library official results. Their status/work filtering belongs `My Work` / exact Document Work / Governance lenses.

Document History may filter exact Revision lifecycle states without creating a Document status scalar.

## Search behavior

Launch free text:

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

Search materialization remains explicitly off:

```text
materialized Search projection = OFF
search_refresh                  = OFF
full Search rebuild             = N/A
external Search engine          = OFF
```

Full-text DOCX/PDF body search, OCR or semantic/vector search is a future material reopen. Search never grants access/effectivity; current T3 and Controlled Documents predicates remain authoritative.

---

# 6. Create, numbering and code authority

Create journey:

```text
choose Document Type
→ choose Area
→ enter REV000 title
→ optionally choose eligible current-EFFECTIVE Template
→ choose/default responsible owner
→ show non-reserving code preview
→ prepare any required source copy outside semantic tx
→ atomic Document + REV000 + WorkingContent create after commit-time revalidation
→ open Document Work lens
```

Numbering is deliberately closed:

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

No configurable separator, width, reset, year token or expression language.

Normalized numbering codes use:

```text
trim
uppercase ASCII alphanumeric
'-' forbidden because the product owns the separator
bounded length; exact API maximum is implementation-contract design unless stronger current evidence is produced
```

Within one Company:

```text
normalized DocumentType.code = unique
normalized Area.code         = unique
committed Document.code      = unique and never reused
```

Area code is immutable after creation. DocumentType code and numbering scope become immutable after the first committed Document uses that type. Display names and permitted lifecycle/config fields remain independently mutable.

Numbering preview:

```text
GET /api/v1/document-types/{document_type_id}/numbering-preview?area_id=...
→ preview_code
→ reservation=false
```

Preview never writes the sequence. Only atomic Document creation returns the final committed code authority. Gaps are allowed.

---

# 7. DRAFT seed semantics and external-copy race law

## Blank create

Blank creation seeds WorkingContent from a trusted product-owned blank DOCX mechanism asset. The blank asset is not a semantic Template Document.

The managed copy is prepared through T4 outside the local business transaction. Failed business creation leaves only reclaimable mechanism content.

## Create from Template

The user selects an eligible Template Document + exact current EFFECTIVE Revision.

```text
read selected current EFFECTIVE Template Revision/source
→ create independent T4 managed copy outside semantic tx
→ BEGIN local create transaction
→ revalidate selected Template is still Template-role/eligible
→ revalidate selected Revision is still current EFFECTIVE
→ revalidate copied exact descriptor/provenance matches selected source
→ create independent Document + REV000 + WorkingContent + DocumentOrigin
→ required Audit
→ COMMIT
```

Later Template changes never rebind the derived Document. The derived Document never reuses Template storage identity and never seeds from OfficialRendition.

## Create next Revision

A new Revision starts from:

```text
current EFFECTIVE Revision title
+ exact winning released source
```

The required source copy is external mechanism work and cannot occur inside the business transaction. Therefore:

```text
read candidate current EFFECTIVE Revision + exact released source
→ create independent managed copy outside semantic tx
→ BEGIN Document-serialized local tx
→ revalidate source Revision is still current EFFECTIVE
→ revalidate no T2 conflict now blocks next-Revision creation
→ revalidate copied descriptor/provenance matches selected released source
→ create next Revision + independent WorkingContent
→ required Audit
→ COMMIT
```

If revalidation fails — including a concurrent successful obsolescence or another lifecycle change — no Revision is created. The copied mechanism bytes remain reclaimable.

A next Revision does not rebind to Template and never starts from OfficialRendition. If the source format cannot be edited by the selected editor, the author may replace DRAFT source through the ordinary T4 upload path; MetalDocs does not silently convert semantic source.

---

# 8. DRAFT editing and optimistic concurrency

Revision title remains Revision-owned metadata, but title and source share one T2 DRAFT concurrency authority: WorkingContent generation.

HTTP expression:

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

At least one field is required. Omitted means unchanged. Null has no implicit delete meaning.

Server:

```text
revalidate If-Match generation
→ if source_upload_id present, prove T4 READY + live binding and load server-owned descriptor
→ atomically apply title/source
→ increment generation exactly once
→ return updated DocumentWorkView + new strong ETag
```

Stale:

```text
412 precondition.draft_changed
zero mutation
```

No second title version, last-write-wins, silent merge or baseline CRDT. Autosave never creates a business Revision, Submission or history fact.

---

# 9. Exact-content upload/admission

Frontend orchestrates T4; it never becomes content authority.

```text
POST /api/v1/revisions/{revision_id}/draft/uploads
→ allocate OPEN upload_id + live admission binding + create-only upload target

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
→ attaches to WorkingContent
```

Truth ladder:

```text
provider upload success != READY
READY != WorkingContent
WorkingContent != Submission
Submission != Release
```

The client never supplies authoritative exact-content hash/descriptor.

Untrusted external bytes must pass exact-byte malware inspection before immutable governed admission. DRAFT autosave need not scan every debounce.

```text
scanner unavailable → 503 dependency.malware_inspector_unavailable
malicious bytes      → 422 validation.content_malicious
```

No browser-scanner integration and no business `scan_status` lifecycle.

---

# 10. Complete Launch lifecycle journeys

This section is normative summary completeness. It does not create a new lifecycle; it restates Product Contract + T1/T2/T5 in implementation-facing form.

## Create

```text
blank/template create
→ stable Document
→ REV000 DRAFT
→ WorkingContent
→ not official yet
```

## NoHumanApproval + SourceOnly submit

```text
DRAFT
→ validate expected generation + current config
→ freeze immutable Submission
→ no fake GovernanceAttempt/Step
→ representation gate absent
→ system Release may commit in the same local transaction
→ Revision EFFECTIVE
```

Submission and Release remain separate semantic facts even when one transaction commits both.

## Human-governed submit

```text
DRAFT
→ freeze immutable Submission
→ Revision SUBMITTED
→ freeze one GovernanceAttempt snapshot
→ exactly one Step active at a time
```

The participant judges the exact Submission, never current WorkingContent.

## RETURN_FOR_CHANGES

```text
record immutable Decision/reason
→ terminate current GovernanceAttempt as returned
→ same Revision SUBMITTED → DRAFT
→ old Submission + Decision + feedback remain immutable history
```

A later submit creates a new Submission for the same business Revision.

## Submission withdrawal

```text
active pre-Release Submission attempt
→ terminate attempt WITHDRAWN
→ same Revision → DRAFT
→ preserve old Submission/decisions/feedback
→ no fake RETURN_FOR_CHANGES
```

## Revision cancellation

```text
DRAFT or eligible pre-Release SUBMITTED Revision
→ terminate any live attempt without fake participant verdict
→ immutable RevisionCancellation
→ Revision CANCELLED terminally
```

If an older Revision is EFFECTIVE, it stays EFFECTIVE. Cancelled ordinal is never reused.

## Required OfficialRendition pending

```text
human gate satisfied or absent
+ frozen policy requires OfficialRendition
+ rendition not yet established
→ Revision remains SUBMITTED
→ Release absent
```

Renderer/job state is mechanism only. The product does not invent `RENDERING` or `RENDER_FAILED` Revision lifecycle states. Terminal renderer failure is visible operationally while semantic state remains truthfully SUBMITTED until corrected through an existing authorized path.

Late output for a returned/withdrawn/cancelled candidate is semantic no-op/reclaimable mechanism output under T5.

## First Release

```text
REV000 SUBMITTED
+ all required gates satisfied
→ system Release
→ REV000 EFFECTIVE
```

There is no user `publish latest file` command.

## Replacement Release

```text
prior current EFFECTIVE Revision
+ successor winning Submission with all gates satisfied
→ one Document-serialized transaction
→ predecessor SUPERSEDED
→ successor EFFECTIVE
→ Release binds exact winning Submission + predecessor
```

No externally successful state may expose two EFFECTIVE Revisions.

## Human-governed obsolescence

```text
current EFFECTIVE target
+ reason
+ no open replacement
+ no competing active request
→ initiate request + GovernanceAttempt
→ target stays EFFECTIVE while governance is active
```

Final ACCEPT revalidates all canonical eligibility and atomically changes current EFFECTIVE Revision to OBSOLETE with no successor.

RETURN terminates the request and leaves target EFFECTIVE.

Authorized request withdrawal terminates the active human-governed request as WITHDRAWN, leaves target EFFECTIVE and creates no fake participant RETURN.

## NoHumanApproval obsolescence

```text
authorized initiation
+ reason
+ all eligibility/conflict checks
+ NoHumanApproval
→ zero human Step
→ no fake System approver
→ same local transaction changes current EFFECTIVE Revision to OBSOLETE
```

## Reader consequence

Release changes official reader truth. A newer DRAFT/SUBMITTED Revision never changes the ordinary reader's EFFECTIVE title/content. Successful obsolescence removes the Document from ordinary current-effective discovery while authorized history remains available.

---

# 11. Governance case, feedback and Step Decision

Governance case reads:

```text
GovernanceAttempt
+ exact active Step
+ exact immutable Submission or exact obsolescence subject
+ bounded prior decisions/feedback
+ live server-derived allowed_actions
```

Case participation grants only the exact governed context needed to act; it never grants general WorkingContent/history access and never permits reviewer mutation of WorkingContent.

For Submission governance, primary review content is the exact Submission source.

A Step has at most one immutable Decision:

```text
GET /api/v1/governance-attempts/{attempt_id}/steps/{step_id}/decision
PUT /api/v1/governance-attempts/{attempt_id}/steps/{step_id}/decision
```

Body:

```text
outcome = ACCEPT | RETURN_FOR_CHANGES
reason  = required for RETURN_FOR_CHANGES
```

Exact repeat returns the existing result. A different attempted outcome after Decision creation fails:

```text
409 state.governance_step_already_decided
```

No durable Idempotency-Key replay row is required for this singleton resource.

Submission feedback remains a separate immutable multi-record fact and may use POST.

No generic signoff, quorum, reassignment, delegation, SLA or arbitrary workflow API exists.

---

# 12. Natural HTTP idempotency and crash-consistent Idempotency-Key

Transport follows semantic resource shape rather than adding replay storage to every write.

Natural singleton/idempotent resources include:

```text
User eligibility PUT
Group membership PUT/DELETE
Governance Step Decision PUT
Submission withdrawal PUT
Revision cancellation PUT
obsolescence withdrawal PUT
current configuration replacement PUT + If-Match
DRAFT PATCH + If-Match
```

## User eligibility

```text
GET /api/v1/users/{user_id}/eligibility
PUT /api/v1/users/{user_id}/eligibility
If-Match: <current eligibility ETag>

state = ENABLED | DISABLED
```

`ENABLED → DISABLED` executes the T3 offboarding transaction: disable User, revoke ApplicationSessions, remove current GroupMemberships/direct User RoleAssignments and append required Audit.

`DISABLED → ENABLED` restores eligibility only. Old sessions/grants/memberships never resurrect. Repeating the same desired state is an idempotent no-op and does not fabricate duplicate semantic Audit transitions.

## Idempotency-Key required set

Require durable replay only for genuinely non-idempotent semantic POST creation where a lost response could otherwise create another semantic fact/resource:

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

Upload allocation/completion use `upload_id` mechanism identity/state and do not require semantic replay rows when exact retry is naturally recognizable by that identity.

Scope:

```text
current User id + canonical operation id + Idempotency-Key
```

Fingerprint uses validated semantic command fields, never client hash authority or raw transport bytes as semantic identity.

## Crash-consistent replay law

The replay result is local PostgreSQL mechanism state and must commit atomically with the semantic transition it protects:

```text
BEGIN local business transaction
→ insert/lock scoped Idempotency-Key
→ if existing, compare semantic fingerprint
→ execute/revalidate semantic transition
→ append required Audit and transaction-coupled durable intents
→ store durable replay result sufficient for exact HTTP status/body replay
→ COMMIT
```

Consequences:

```text
semantic fact commits ⇔ completed replay result commits
rollback → neither completed semantic fact nor completed replay result
same key + same fingerprint after commit → exact replay + Idempotent-Replay: true
same key + different fingerprint         → 422 validation.idempotency_key_reused
```

Concurrent exact same-key requests serialize/wait on the key. After the winner commits, the other request returns the stored replay. **No baseline public/durable IN_PROGRESS idempotency state or `409 conflict.idempotency_in_progress` contract is required.**

External preflight may duplicate transient mechanism work before the local transaction; such output remains reclaimable and never creates a second semantic fact.

Replay retention is bounded and must safely exceed ordinary browser/network retry windows. `24h` remains a first implementation-default candidate, not an architectural invariant. Post-expiry correctness remains T1/T2/T3 responsibility.

---

# 13. Release source, exact-byte gateway and OfficialRendition presentation

There is no public publish/release mutation. Release remains system-owned.

Semantic exact-byte resources:

```text
GET /api/v1/revisions/{revision_id}/draft/source
GET /api/v1/submissions/{submission_id}/source
GET /api/v1/releases/{release_id}/source
GET /api/v1/official-renditions/{rendition_id}/content
```

The application contract never exposes provider bucket/key/version or `managed_content_id` as product identity.

Required baseline:

```text
exact full semantic read
current authorization recheck
semantic URL independent of provider location
```

HTTP Range support is **optional mechanism**, activated only when selected PDF/viewer/file-size evidence proves benefit. Adding Range later does not change semantic URLs or authorization meaning.

Implementation may serve bytes directly or use an authorization-checked short-lived provider/CDN redirect while preserving semantic resource identity.

Presentation:

```text
SourceOnly PDF
→ direct PDF view of Release source
→ exact source download

SourceOnly DOCX
→ selected read-only interactive DOCX adapter on Release source
→ exact source download

RequireOfficialRendition(PDF)
→ OfficialRendition PDF is primary official in-product view
→ exact Release source remains separately available/labeled

Governance
→ exact Submission source is primary decision content
```

Viewer/preview output never becomes OfficialRendition authority.

---

# 14. Interactive DOCX provider and separate OfficialRendition renderer

Interactive DOCX adapter boundary:

```text
load exact DOCX bytes
render/edit in browser
emit complete resulting DOCX bytes
read-only mode for inspection
no provider-owned durable document identity
no provider callback/save state as business truth
```

Launch selects **exactly one interactive DOCX editor/viewer adapter**, not one universal renderer for every representation mechanism.

First interactive candidate: EigenPal-class browser-buffer adapter because it naturally fits T4 admission + T2 OCC.

Before implementation planning freezes the interactive provider, run one representative fidelity corpus covering at least:

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

Hard fail includes silent OOXML loss/corruption, material layout loss on ordinary MetalDocs documents, destructive rewrite of unsupported constructs or invalid/unopenable round-trip.

If the browser-buffer candidate materially fails, evaluate an ONLYOFFICE-class server-backed candidate against the same corpus and prove callback/document-server state cannot bypass T4 or overwrite a newer accepted OCC generation.

No dual interactive editor runtime exists in Launch. No EditorSession/lease correctness dependency exists by default; reopen only if the selected integration proves a concrete provider/UX requirement not safely represented by ApplicationSession + WorkingContent generation.

T5's **server-side OfficialRendition renderer/converter remains a separate mechanism** and may use a different product from the interactive editor/viewer. The interactive-provider rule does not collapse the rendition pipeline.

---

# 15. Administration, permission ownership and lost-update protection

One Admin Center has three semantic sections.

## Organization

```text
Company current settings
Users + erasable UserProfile
Areas
Groups + memberships
provider subject binding
User eligibility/offboarding
```

Minimum UserProfile:

```text
display_name required while profile exists
email optional/contact enrichment
```

## Access

```text
RoleAssignment grant/revoke
subject = USER | GROUP
role = static T3 role code
scope = COMPANY | AREA(area_id)
```

No custom role/permission editor.

## Document Governance

```text
Document Types
active/inactive
numbering configuration/preview
NoHumanApproval | UseGovernanceRoute
ordered sequential Steps
NAMED_USER | GROUP selector
SourceOnly | RequireOfficialRendition(PDF)
Template role/eligibility
```

No public generic PolicyVersion/workflow platform.

Permission ownership remains exactly T3:

```text
DocumentType route/representation config → document_type.manage
Template role/eligibility                → template_use.manage
Group identity                           → organization.manage
Group membership                         → access.manage
RoleAssignment                           → access.manage
```

UI placement under one Admin Center does not merge those permissions.

## Bounded template administration read model

A governance admin with `template_use.manage` must be able to administer Template role/eligibility without receiving implicit document content/history access.

The application API therefore exposes a bounded **Document Governance template-configuration projection** containing only the minimum selection metadata, e.g.:

```text
Document id
stable code
bounded display identity label
current Template-role flag
current effective/eligibility indicator needed for configuration
```

It does **not** expose source bytes, full effective content, WorkingContent, Submission/history or general document read authority.

Exact source/history remain protected by `document.read_effective`, `document.read_history` or exact governance-case authorization. No new permission is introduced.

## Strong ETags on authority-bearing current resources

Whole-replacement mutable current resources that can suffer lost update return strong ETags and require `If-Match` where appropriate, including at least:

```text
Company
UserProfile
User eligibility
User ProviderSubjectBinding
Area metadata/lifecycle
Group metadata
DocumentType base
DocumentType governance configuration
DocumentType eligible-template set
Document responsible owner
Document Template role
```

Specific singleton reads/replacements include:

```text
GET /api/v1/users/{user_id}/provider-binding
PUT /api/v1/users/{user_id}/provider-binding

GET /api/v1/documents/{document_id}/responsible-owner
PUT /api/v1/documents/{document_id}/responsible-owner

GET /api/v1/documents/{document_id}/template-role
PUT /api/v1/documents/{document_id}/template-role
```

Stale precondition:

```text
412 precondition.resource_changed
zero mutation
```

Exact repeat of the already-current desired value may be an idempotent no-op and must not fabricate duplicate semantic Audit transitions.

---

# 16. Public errors

All `/api/v1` application failures use RFC 9457 Problem Details:

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

`type` is mechanically derived and never independently registered.

Closed top-level families:

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

No module-named public error family. Frontend branches on `code`, never localized detail text. Raw provider/storage/scanner errors never leak.

---

# 17. Lists, purpose-built read models and `allowed_actions`

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

No mandatory total count, offset pagination or generic filter/sort DSL. Cursor is opaque and bound to filter/order semantics. Static roles and other small closed vocabularies may be unpaginated.

Purpose-built read models include:

```text
SessionView
DocumentSummary
DocumentOfficialView
DocumentWorkView
SubmissionView
GovernanceCaseView
DocumentHistoryView
WorkAuthoringItem
WorkGovernanceItem
AuditEventView
Admin configuration views
TemplateConfigurationItem
```

No table dumps, universal polymorphic envelopes, `ArtifactViewModel` or `ApprovalInstanceDTO` become target public contracts.

Read models may carry live-derived `allowed_actions` for UX. They are hints, never grants.

`allowed_actions` must be derived through the **same canonical T3 permission/scope components and Controlled Documents relationship/state/governance predicates used by command authorization**, or a provably shared equivalent. It must never be computed by a parallel frontend/backend role matrix.

A stale hint may cause harmless UX mismatch. Every command rechecks canonical current truth.

---

# 18. History versus Audit

Document History answers:

> What happened to this controlled Document?

It is constructed from Controlled Documents semantic facts:

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

Audit answers:

> What meaningful actions occurred across the product, who performed them, and when?

Audit filters may include time, actor, operation family, Area where authorized and resource id/kind.

The frontend may cross-link History and Audit visually, but it must never merge them into one synthetic lifecycle authority. Audit cannot reconstruct current lifecycle state.

---

# 19. Frontend technical organization

Preserve mechanism patterns that survive Structural Inversion:

```text
React SPA
feature-sliced organization
TanStack Query for server state
OpenAPI-generated application transport types
lib API/client infrastructure
local React state for local UI state
shared domain-agnostic UI/editor/viewer primitives
```

Replace the legacy frontend feature taxonomy. Target feature vocabulary follows semantic lenses/owners, for example:

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

Frontend feature boundaries never mirror Go package names by default.

---

# 20. Closed Launch application operation census

The following is the T6 Launch `/api/v1` application-operation family. Minor path spelling/operationId details may be normalized during OpenAPI authoring only when semantic meaning remains identical; adding a new product route family requires a named Product Contract journey or explicit T6 reopen.

## Session / Authentication integration support

```text
GET    /api/v1/session
DELETE /api/v1/session
GET    /api/v1/authentication/provider-subjects?query=...
```

Browser OIDC routes remain outside `/api/v1`:

```text
GET /auth/login
GET /auth/callback
```

## Organization

```text
GET    /api/v1/company
PUT    /api/v1/company

GET    /api/v1/users
POST   /api/v1/users
GET    /api/v1/users/{user_id}
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
PUT    /api/v1/areas/{area_id}/lifecycle

GET    /api/v1/groups
POST   /api/v1/groups
GET    /api/v1/groups/{group_id}
PUT    /api/v1/groups/{group_id}
DELETE /api/v1/groups/{group_id}
PUT    /api/v1/groups/{group_id}/members/{user_id}
DELETE /api/v1/groups/{group_id}/members/{user_id}
```

## Authorization

```text
GET    /api/v1/roles
GET    /api/v1/role-assignments
POST   /api/v1/role-assignments
DELETE /api/v1/role-assignments/{assignment_id}
```

No custom Role/Permission CRUD.

## Document Governance configuration

```text
GET  /api/v1/document-types
POST /api/v1/document-types
GET  /api/v1/document-types/{document_type_id}
PUT  /api/v1/document-types/{document_type_id}

GET /api/v1/document-types/{document_type_id}/governance
PUT /api/v1/document-types/{document_type_id}/governance

GET /api/v1/document-types/{document_type_id}/eligible-templates
PUT /api/v1/document-types/{document_type_id}/eligible-templates

GET /api/v1/document-types/{document_type_id}/numbering-preview?area_id=...

GET /api/v1/document-governance/templates
```

`GET /document-governance/templates` is the bounded `template_use.manage` metadata projection, not a separate Template lifecycle API.

## Controlled Documents / Work

```text
GET  /api/v1/documents
POST /api/v1/documents
GET  /api/v1/documents/{document_id}
GET  /api/v1/documents/{document_id}/responsible-owner
PUT  /api/v1/documents/{document_id}/responsible-owner
GET  /api/v1/documents/{document_id}/template-role
PUT  /api/v1/documents/{document_id}/template-role
POST /api/v1/documents/{document_id}/revisions
GET  /api/v1/documents/{document_id}/history

GET /api/v1/work/authoring
GET /api/v1/work/governance

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

PUT /api/v1/revisions/{revision_id}/cancellation

GET  /api/v1/governance-attempts/{attempt_id}
GET  /api/v1/governance-attempts/{attempt_id}/feedback
POST /api/v1/governance-attempts/{attempt_id}/feedback
GET  /api/v1/governance-attempts/{attempt_id}/steps/{step_id}/decision
PUT  /api/v1/governance-attempts/{attempt_id}/steps/{step_id}/decision

GET /api/v1/releases/{release_id}
GET /api/v1/releases/{release_id}/source
GET /api/v1/official-renditions/{rendition_id}/content

POST /api/v1/documents/{document_id}/obsolescence-requests
GET  /api/v1/obsolescence-requests/{request_id}
PUT  /api/v1/obsolescence-requests/{request_id}/withdrawal
```

No public Release/publish mutation endpoint, no generic `/actions`, no generic Search endpoint, no separate Approval lifecycle API and no separate Template lifecycle API.

## Audit

```text
GET /api/v1/audit/events
```

Operations endpoints are explicitly outside this product-operation census.

---

# 21. Future-evolution seam verification

T6 remains additive-compatible with every named future horizon without dormant implementation:

```text
Distribution / Read & Acknowledge
→ attaches to Release + User/Group
→ never becomes effectivity authority

Periodic Review
→ attaches to stable Document + exact current EFFECTIVE Revision

Dossier
→ references stable Document identity

Evidence
→ may own independent lifecycle + T4 exact-content seam without restoring Artifact

Retention / Legal Hold / Disposition
→ attaches to stable governed identities/history; provider storage remains enforcement only

Governed Export
→ consumes semantic IDs + exact descriptors/byte routes without becoming source authority

External Repository IMPORT/PUBLISH
→ uses target-owner + T4 copy/admission seams; provider identity remains external mechanism

Training/LMS
→ consumes released/effective truth and future Distribution obligations

Multi-document Change Control
→ may orchestrate stable Document/Revision transitions without taking their lifecycle authority

Pooled tenancy
→ may reopen substrate/API context around stable Company identity without rewriting governed history

CRDT/realtime collaboration
→ may replace DRAFT OCC/editor mechanism without changing Revision or immutable Submission meaning
```

No future capability justifies a dormant Launch module, table, permission, worker or generic framework.

---

# 22. Explicitly removed from the Launch target

Current implementation may be refactored or deleted where it represents superseded scope/ownership. T6 deliberately does not preserve by sunk cost:

```text
Tenant/RLS universal product ontology
local password login/change-password
separate Approval product/workspace/API
separate Template lifecycle/API
public ControlledDocument peer object
Distribution Launch routes/state
Notifications/inbox Launch routes/state
Taxonomy/Dictionary/Tokens platforms absent promoted requirement
legacy role/capability-driven frontend authority
writer/editor session correctness dependency
scheduled publish
universal PDF/export paths
reviewer mutation of WorkingContent
materialized Search without a consumer
legacy API compatibility layer
```

A low-level mechanism may remain only when independently justified by current authority; old naming/topology is never entitlement.

---

# 23. Implementation-proof contract inherited from T6

Later implementation design/tests must make at least these claims falsifiable:

```text
OIDC Authorization Code works without MetalDocs password input
unbound provider subject cannot login
unsafe browser command fails closed without valid session-bound CSRF
User creation cannot leave half-created User/Profile/Binding truth
provider-binding replacement is If-Match protected and invalidates required sessions
allowed_actions cannot authorize a backend command and cannot use a parallel role matrix
legacy removed route families are absent from target application contract
there is no /api/v2 compatibility layer
operations readiness reflects T4 non-serving restore state truthfully
Library status filtering is derived and no Document.currentStatus authority exists
DRAFT title/source share exactly one ETag generation
stale DRAFT If-Match yields 412 and zero mutation
client-provided descriptor/hash cannot become exact-content authority
provider upload success alone cannot attach WorkingContent
blank seed is not a semantic Template
Template-based create revalidates current EFFECTIVE/eligibility after external copy
create-next-Revision revalidates current EFFECTIVE source after external copy and before commit
review case displays exact Submission bytes rather than current WorkingContent
NoHumanApproval + SourceOnly may Release atomically without fake governance
RETURN/withdraw/cancel preserve immutable prior history
required-rendition pending remains SUBMITTED without fake renderer lifecycle state
replacement Release cannot expose two EFFECTIVE Revisions
active obsolescence leaves target EFFECTIVE until successful completion
NoHumanApproval obsolescence creates zero fake human Step
repeated singleton withdrawal/cancellation/eligibility/Step-decision cannot duplicate evidence
Idempotency-Key semantic fact and replay result commit atomically
concurrent same-key request serializes/replays rather than requiring public IN_PROGRESS state
numbering preview never reserves
normalized DocumentType.code and Area.code are Company-unique
committed Document code never reuses
provider-binding/responsible-owner/template-role lost update is blocked by strong If-Match
Template administration does not leak document content/history to governance_admin
Search projection/jobs are absent in Launch baseline
ordinary Library never presents DRAFT/SUBMITTED as official
Domain history can be reconstructed without Audit
Audit cannot become current lifecycle authority
interactive DOCX provider passes fidelity corpus and cannot bypass OCC/T4
interactive DOCX provider choice does not constrain separate OfficialRendition renderer
Range read is optional, not required for semantic correctness
no EditorSession is needed for correctness in baseline
future capability seams attach without duplicate current authority or dormant implementation
```

---

# 24. Current gate

```text
T6 material decisions             OPERATOR-APPROVED
C1→C8 bounded corrections         OPERATOR-APPROVED / INCORPORATED
L1→L5 refinements                 INCORPORATED
this corrected platform summary   BOUNDED COHERENCE DELTA NEXT
T6 durable promotion              NOT YET
Decision Registry reconciliation  NOT YET
T6 staging cleanup                NOT YET
T7                                NOT OPEN
implementation                    BLOCKED
```

Next:

```text
bounded coherence delta against Product Contract REV001 + Whole-Product GCR + 4+1 + T1→T5 + Registry
→ if clean: operator platform-summary ratification
→ durable T6 promotion only after explicit summary ratification
```

No implementation plan or product code is authorized by this summary.