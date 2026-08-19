# R10-T6 — Canonical API / Frontend Journeys — Platform-Facing Summary

> **Status:** ACTIVE STAGING / NON-AUTHORITATIVE — **OPERATOR SUMMARY RATIFICATION NEXT**  
> **Date:** 2026-08-18  
> **Repository:** `developmentconexus-ops/MetalDocs`  
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131  
> **Material adjudication:** OPERATOR-APPROVED  
> **Implementation:** BLOCKED  
> **T7:** NOT OPEN

This summary is the platform-facing expression of the operator-approved T6 material decisions. It explains the target product/API/frontend system a future implementation team must build **without depending on current MetalDocs runtime shape**.

It is not yet durable authority. The operator must explicitly ratify this summary before T6 is promoted to `wiki/`, reconciled into the Decision Registry and closed.

---

# 1. Platform model

MetalDocs Launch is one controlled-document product with four business semantic owners plus Audit:

```text
Authentication
Organization
Authorization
Controlled Documents
Audit (supporting semantic evidence)
```

The public product is **not** a set of peer applications called Approval, Templates, Controlled Documents, Distribution, Taxonomy or Tokens.

Frontend/API surfaces are semantic lenses over the same accepted T1→T5 facts:

```text
Library              → current official/effective truth
My Work              → actor-relevant author/governance work
Document Official    → stable Document + current EFFECTIVE truth
Document Work        → exact current open Revision DRAFT
Governance Case      → exact GovernanceAttempt + exact immutable Submission
Document History     → Controlled Documents history
Audit                → AuditEvent action evidence
Administration       → Organization / Access / Document Governance
```

A route/read model may denormalize information for UX, but it never becomes a second domain authority.

---

# 2. Public contract law

The pre-launch public API is rewritten in place under:

```text
/api/v1
```

There is no `/api/v2` migration namespace and no compatibility shim for the old pre-launch contract.

Contract discipline:

```text
api/openapi/v1/openapi.yaml = public HTTP contract SSOT
OpenAPI 3.0.3 feature set   = Launch baseline
generated Go boundary       = required
generated TypeScript types  = required
JSON casing                 = snake_case
opaque technical IDs        = UUID
trusted instants            = RFC3339/UTC
```

OAS 3.0.3 is retained because T6 has no named consumer for OAS 3.1 semantics; upgrading the description language/toolchain during a complete contract rewrite would add accidental complexity. Reopen only on a concrete schema/tooling benefit.

The API follows product semantics, not backend package/module names.

---

# 3. Authentication and browser trust boundary

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

No Direct Grant/ROPC. No MetalDocs password login/change-password route. No JIT User creation.

Unknown/unbound provider subject fails closed.

Session cookie baseline:

```text
__Host-metaldocs_session
Secure
HttpOnly
SameSite=Lax
Path=/
no Domain attribute
```

Application session surface:

```text
GET    /api/v1/session
DELETE /api/v1/session
```

Unsafe same-origin browser API requests require a session-bound CSRF token. SameSite is defense-in-depth, not the only CSRF defense.

Provider roles/groups never become MetalDocs Authorization.

Administration may search provider identities and bind an existing provider subject to a local User. Provider-binding replacement is explicit, security-sensitive and invalidates existing ApplicationSessions.

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
  ordinary current-effective Library

/documents/:document_id
  stable Document official/management lens; current EFFECTIVE truth primary

/documents/:document_id/work
  authorized open Revision work lens

/documents/:document_id/history
  authorized Controlled Documents history

/work
  actor work workspace

/work/governance/:attempt_id
  exact GovernanceAttempt + exact Submission case

/audit
  AuditEvent inspection

/admin/organization
/admin/access
/admin/document-governance
```

A URL does not silently switch between DRAFT and EFFECTIVE content based on the caller. Visual components may be reused; semantic route meaning remains stable.

No top-level Launch workspace for Approvals, Templates, Distribution, Notifications, Taxonomy, Tokens, Sessions or Metrics.

---

# 5. Reader truth and Library

Ordinary Library discovery is current-effective truth only.

Searchable canonical facts:

```text
Document code
current EFFECTIVE Revision title
Document Type
Area
responsible owner
```

Default discovery never presents DRAFT/SUBMITTED work as official.

Launch free text:

```text
q = code + current EFFECTIVE title
```

Deterministic rank when `q` is present:

```text
exact code
→ code prefix
→ title prefix
→ title contains
→ code + stable id tie-break
```

No `q`:

```text
code + stable id
```

Search materialization is explicitly **not activated** for Launch:

```text
materialized Search projection = OFF
search_refresh                  = OFF
full Search rebuild             = N/A
external Search engine          = OFF
```

Full-text DOCX/PDF body search, OCR or semantic/vector search is a future material reopen, not an implementation convenience.

Search never grants access/effectivity; current T3/domain predicates remain authoritative.

---

# 6. Create, numbering and seed semantics

Create journey:

```text
choose Document Type
→ choose Area
→ enter REV000 title
→ optionally choose eligible Template Document/current EFFECTIVE Revision
→ choose/default responsible owner
→ show non-reserving code preview
→ atomic Document + REV000 + WorkingContent create
→ open Document work lens
```

Numbering is deliberately closed:

```text
numbering_scope = DOCUMENT_TYPE | DOCUMENT_TYPE_AREA
separator       = '-'
sequence        = decimal, minimum display width 3, grows naturally after 999
```

Examples:

```text
PO-001
PO-RH-003
PO-RH-1000
```

No configurable separator/width/reset/year token/expression language.

`DocumentType.code` / `Area.code` are trimmed uppercase ASCII alphanumeric, 1..16 chars, with `-` forbidden because the product owns the separator.

Area code is immutable after creation. DocumentType code and numbering scope become immutable after the first committed Document uses that type.

Preview never reserves or mutates sequence state. Atomic create returns the final committed code; gaps are allowed; committed codes never reuse.

### DRAFT seeds

Blank create uses a trusted product-owned blank DOCX mechanism asset. It is not a semantic Template Document.

Template-based create copies the exact released **source** from the selected current EFFECTIVE eligible Template Revision into independent WorkingContent. It never reuses the Template's storage identity and never seeds from OfficialRendition.

Create-next-Revision copies:

```text
current EFFECTIVE Revision title
+ exact winning released source
```

into a new independent WorkingContent handle. It does not rebind to Template and never seeds from OfficialRendition.

---

# 7. DRAFT editing and optimistic concurrency

Revision title remains Revision-owned semantic metadata, but title and source share one T2 DRAFT concurrency authority: WorkingContent generation.

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

At least one is required. Omitted means unchanged. No implicit null-delete semantic.

Server:

```text
revalidate If-Match
→ revalidate upload if present
→ atomically apply title/source
→ increment generation exactly once
→ return updated DocumentWorkView + new strong ETag
```

Stale mutation:

```text
412 precondition.draft_changed
zero mutation
```

No second title version, last-write-wins, silent merge or baseline CRDT.

Autosave does not create business Revision/Submission/history.

---

# 8. Exact-content upload/admission

Frontend orchestrates T4; it never becomes content authority.

```text
POST /api/v1/revisions/{revision_id}/draft/uploads
→ allocate OPEN upload_id + live admission binding + create-only target

browser uploads exact bytes directly to provider

POST /api/v1/revisions/{revision_id}/draft/uploads/{upload_id}/complete
→ server reads exact stored bytes
→ derives SHA-256 + size + ContentFormat
→ structural validation
→ READY

PATCH /api/v1/revisions/{revision_id}/draft + If-Match
→ reference upload_id
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

The client never submits an authoritative SHA/descriptor.

Untrusted external bytes must pass exact-byte malware inspection before immutable governed admission. DRAFT autosave does not rescan every debounce.

```text
scanner unavailable → 503 dependency.malware_inspector_unavailable
malicious bytes      → 422 validation.content_malicious
```

No browser-scanner integration and no business `scan_status` workflow.

---

# 9. Submission, governance and reviewer behavior

SUBMIT freezes the exact accepted DRAFT generation and decision-relevant governed state into an immutable Submission.

A reviewer never judges current WorkingContent. Governance case reads:

```text
GovernanceAttempt
+ exact active Step
+ exact immutable Submission
+ bounded prior decisions/feedback
+ live server-derived allowed_actions
```

The primary review content is the exact Submission source.

Case participation permits inspection and governance action only; it does not grant mutation of WorkingContent.

Decision is a singleton immutable resource for one Step:

```text
GET /api/v1/governance-attempts/{attempt_id}/steps/{step_id}/decision
PUT /api/v1/governance-attempts/{attempt_id}/steps/{step_id}/decision
```

Body:

```text
outcome = ACCEPT | RETURN_FOR_CHANGES
reason  = required for RETURN_FOR_CHANGES
```

Exact repeat returns the existing result. A different decision after creation fails:

```text
409 state.governance_step_already_decided
```

No generic signoff/quorum/reassign/SLA UI or API.

Submission feedback remains a separate immutable fact and may use POST because multiple feedback records are legitimate.

---

# 10. Lifecycle command transport

Transport follows semantic resource shape rather than blindly applying `Idempotency-Key` to every write.

Natural singleton/idempotent resources include:

```text
User eligibility PUT
Group membership PUT/DELETE
Governance Step Decision PUT
Submission withdrawal PUT
Revision cancellation PUT
obsolescence withdrawal PUT
configuration PUT + If-Match
DRAFT PATCH + If-Match
```

### User eligibility

```text
GET /api/v1/users/{user_id}/eligibility
PUT /api/v1/users/{user_id}/eligibility
If-Match: <eligibility ETag>

state = ENABLED | DISABLED
```

`ENABLED → DISABLED` executes the T3 offboarding transaction: disable User, revoke sessions, remove memberships/direct grants and Audit. `DISABLED → ENABLED` restores eligibility only. Repeating the same desired state is an idempotent no-op and does not create duplicate semantic Audit transitions.

### Idempotency-Key

Require durable replay only for genuinely non-idempotent semantic POST creation where a lost response could create a second fact/resource, including:

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

Fingerprint uses validated semantic command fields, never client hash authority.

Behavior:

```text
missing required key                → 400 request.idempotency_key_required
first key/fingerprint               → execute + store exact status/body
completed same key/fingerprint      → exact replay + Idempotent-Replay: true
same key/fingerprint still running  → 409 conflict.idempotency_in_progress
same key/different fingerprint      → 422 validation.idempotency_key_reused
```

Replay retention is bounded and must exceed ordinary browser/network retry windows. 24h is the first implementation-default candidate, not an architecture invariant. Business correctness after expiry remains T1/T2/T3 responsibility.

---

# 11. Release, source and OfficialRendition presentation

There is **no public publish/release command**. Release remains system-owned under T2.

Semantic exact-byte resources:

```text
GET /api/v1/revisions/{revision_id}/draft/source
GET /api/v1/submissions/{submission_id}/source
GET /api/v1/releases/{release_id}/source
GET /api/v1/official-renditions/{rendition_id}/content
```

The client/API never exposes provider bucket/key/version or `managed_content_id` as product identity.

Baseline byte gateway supports full and Range reads. Implementation may later return an authorization-checked short-lived provider/CDN redirect without changing the semantic URL.

Presentation:

```text
SourceOnly PDF
→ direct PDF view of Release source
→ exact source download

SourceOnly DOCX
→ read-only DOCX adapter on Release source
→ exact source download

RequireOfficialRendition(PDF)
→ OfficialRendition PDF is primary official in-product view
→ exact Release source remains separately available/labeled

Governance
→ exact Submission source is primary decision content
```

Viewer/preview output never becomes OfficialRendition authority.

---

# 12. DOCX editor/viewer strategy

Architecture boundary:

```text
load exact DOCX bytes
render/edit in browser
emit complete resulting DOCX bytes
read-only mode for inspection
no provider-owned durable document identity
no provider callback/save state as business truth
```

First candidate: EigenPal-class browser-buffer adapter because it naturally fits T4 admission + T2 OCC.

It is **not** frozen from feature marketing. Before implementation planning freezes the provider, run one representative fidelity corpus including:

```text
styles and paragraph hierarchy
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

If the browser-buffer candidate materially fails, evaluate an ONLYOFFICE-class server-backed candidate against the same corpus and prove its callback/document-server state cannot bypass T4 or overwrite a newer OCC generation.

Exactly one DOCX provider is selected for Launch. No dual-editor runtime.

No EditorSession/lease correctness dependency baseline. Reopen only if the actually selected integration proves a concrete provider/UX need that ordinary ApplicationSession + WorkingContent generation cannot satisfy.

---

# 13. Administration

One Admin Center, three semantic sections.

### Organization

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

### Access

```text
RoleAssignment grant/revoke
subject = USER | GROUP
role = static T3 role code
scope = COMPANY | AREA(area_id)
```

No custom role/permission editor.

### Document Governance

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

Mutable admin representations that can cause lost update return a strong ETag and require `If-Match` on PUT, including at least Company, UserProfile, Area metadata/lifecycle, Group metadata, DocumentType base, governance configuration and eligible-template set.

Stale admin mutation = 412.

---

# 14. Errors

All public failures use RFC 9457 Problem Details:

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

`type` is mechanically derived, never independently registered.

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

No module-named public problem families. Frontend branches on `code`, never localized detail text. Provider/storage/scanner raw errors never leak.

---

# 15. Lists and read models

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

No mandatory total count, offset pagination or generic filter/sort DSL. Cursor binds filter/order semantics. Static roles and other small closed vocabularies may be unpaginated.

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
```

No table dumps, universal polymorphic envelopes, `ArtifactViewModel` or `ApprovalInstanceDTO` as target public contracts.

Read models may carry live-derived `allowed_actions` for UX. They are hints, not grants; every command rechecks canonical T2/T3 truth.

---

# 16. History versus Audit

Document History answers:

> What happened to this controlled Document?

It is built from Controlled Documents semantic facts:

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

> What meaningful actions occurred across the product and who performed them?

Audit filters may include time, actor, operation family, Area where authorized and resource id/kind.

The frontend may cross-link them visually, but it must never merge them into a synthetic lifecycle source of truth. Audit does not reconstruct current state.

---

# 17. Frontend technical organization

Preserve mechanism patterns that survive Structural Inversion:

```text
React SPA
feature-sliced organization
TanStack Query for server state
OpenAPI-generated transport types
lib API/client infrastructure
local React state for local UI state
shared domain-agnostic UI/editor/viewer primitives
```

Replace legacy feature taxonomy.

Target feature vocabulary follows semantic lenses/owners, for example:

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

Frontend feature boundaries do not mirror Go package names.

---

# 18. Explicitly removed from the Launch target

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

The target implementation may retain a low-level mechanism only when it independently satisfies current authority; old naming/topology is never entitlement.

---

# 19. Implementation-proof contract inherited from T6

Later implementation design/tests must make at least these claims falsifiable:

```text
OIDC Authorization Code path works without MetalDocs password input
unbound provider subject cannot login
unsafe browser command fails closed without valid session-bound CSRF
session/read-model allowed_actions cannot authorize a backend command
legacy removed route families are absent from target contract
there is no /api/v2 compatibility layer
DRAFT title/source share exactly one ETag generation
stale If-Match yields 412 and zero mutation
client-provided descriptor/hash cannot become exact-content authority
provider upload success alone cannot attach WorkingContent
blank seed is not a semantic Template
Template/revise seed copies released source, never OfficialRendition
review case displays exact Submission bytes rather than current WorkingContent
repeated singleton withdrawal/cancellation/eligibility/Step-decision cannot create duplicate evidence
Idempotency-Key exact retry replays while changed-key request rejects
numbering preview never reserves
committed Document code never reuses
Area/type code immutability holds after prescribed boundary
Search projection/jobs are absent in Launch baseline
ordinary Library never presents DRAFT as official
Domain history can be reconstructed without Audit
Audit cannot become current lifecycle authority
selected DOCX provider passes fidelity corpus and cannot bypass OCC/T4
no EditorSession is needed for correctness in baseline
```

---

# 20. Current gate

```text
T6 material decisions             OPERATOR-APPROVED
this platform-facing summary      OPERATOR RATIFICATION NEXT
T6 durable promotion              NOT YET
Decision Registry reconciliation  NOT YET
T6 staging cleanup                NOT YET
T7                                NOT OPEN
implementation                    BLOCKED
```

If this summary is ratified, the next operation is mechanical governance:

```text
promote T6 durable authority to wiki/
→ reconcile Decision Registry
→ update router/handoff/index/PR
→ remove completed T6 staging from live tree (Git history archive)
→ only then open T7 Historical Migration & Cutover
```

No implementation plan or product code is authorized by this summary.