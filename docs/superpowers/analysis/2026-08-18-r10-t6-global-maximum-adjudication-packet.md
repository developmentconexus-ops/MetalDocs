# R10-T6 — Canonical API / Frontend Journeys — Corrected Global-Maximum Adjudication Packet

> **Status:** ACTIVE STAGING / NON-AUTHORITATIVE — **OPERATOR MATERIAL ADJUDICATION NEXT**  
> **Date:** 2026-08-18  
> **Repository:** `developmentconexus-ops/MetalDocs`  
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131  
> **Base candidate:** `docs/superpowers/analysis/2026-08-18-r10-t6-canonical-api-frontend-journeys-candidate.md`  
> **Product Contract:** REV001  
> **Implementation:** BLOCKED

This packet does not reopen T1→T5. It tightens the T6 material candidate where the first pass still left enough freedom for two competent implementers to produce materially different systems.

When this packet is more specific than the base candidate, **this packet is the proposed T6 disposition for operator adjudication**. Neither file is durable authority until operator ratification + platform-facing summary ratification + promotion.

---

# 0. Corrections required after the evidence pass

The first candidate had the correct architecture direction but left ambiguity in:

```text
AuthN / browser session / CSRF
exact public operation census
OpenAPI version posture
numbering width/code mutability
HTTP OCC expression for DRAFT
upload attachment authority
exact byte-serving routes
blank/template/revise seed semantics
provider-fidelity freeze gate
Problem type/code authority
Idempotency-Key operation set + replay semantics/TTL
admin lost-update protection
```

Those are corrected below.

---

# 1. T6-A refinement — contract replacement + OpenAPI version

## Candidate disposition: REFINED / ACCEPT

The pre-launch `/api/v1` contract is rebuilt from current authority. There is no `/api/v2` and no compatibility shim.

```text
namespace                    /api/v1
contract SSOT                api/openapi/v1/openapi.yaml
OpenAPI feature set          3.0.3 for Launch
generated Go boundary        YES
generated TypeScript types   YES
JSON casing                  snake_case
```

Why keep OAS 3.0.3 while replacing the contract content completely?

- no T6 consumer requires OAS 3.1 semantics;
- current stable `oapi-codegen` only added initial OAS 3.1 support in July 2026;
- changing the public product contract is already a high-blast-radius change;
- adding a description-language/toolchain migration without a consumer is accidental complexity, not Global Maximum.

This is **not** compatibility preservation: every legacy route may disappear.

Reopen the OAS minor version only on a named schema/tooling benefit.

---

# 2. T6-S — Authentication + ApplicationSession + browser security

## Candidate disposition: ACCEPT

MetalDocs does not authenticate by receiving end-user credentials.

### Browser AuthN

```text
GET /auth/login
→ server creates OIDC state/PKCE transaction
→ Keycloak Authorization Code redirect
→ GET /auth/callback
→ server verifies state/code/provider
→ issuer + subject
→ ProviderSubjectBinding
→ enabled User
→ create fresh ApplicationSession
→ set session cookie
→ redirect SPA
```

No Direct Grant/ROPC. No `/api/v1/auth/login` password POST. No password-change API in MetalDocs.

### Session cookie

```text
name        __Host-metaldocs_session
Secure      true
HttpOnly    true
SameSite    Lax
Path        /
Domain      absent
```

Application API:

```text
GET    /api/v1/session
DELETE /api/v1/session
```

`GET /session` may return current live-derived navigation/UX affordances; they are not persisted permission authority and backend commands always re-evaluate T3.

### CSRF

Every unsafe same-origin browser API request requires:

```text
X-CSRF-Token: <session-bound token>
```

The token is issued by the authenticated session bootstrap and bound to the current ApplicationSession.

```text
invalid/missing session    → 401 auth.session_invalid
invalid/missing CSRF       → 403 permission.csrf_invalid
```

SameSite is defense-in-depth, not the only defense.

No permissive cross-origin CORS baseline.

### Provider directory / User binding

Launch binds **existing** provider identities; it does not provision Keycloak accounts/passwords.

```text
GET /api/v1/authentication/provider-subjects?query=...
→ opaque provider_subject_ref + display hints

POST /api/v1/users
→ provider_subject_ref + local profile
→ server resolves ref to issuer+subject
→ local User/Profile/ProviderSubjectBinding
```

No JIT user creation on first login. An unknown provider subject fails closed.

Provider roles/groups are never imported into T3.

Provider-binding replacement is an explicit admin operation, invalidates existing ApplicationSessions and must prove only one live binding can authenticate the User.

---

# 3. T6-C refinement — closed public operation census

## Candidate disposition: REFINED / ACCEPT

The following is the candidate **closed Launch public operation set**. New route families require a named Product Contract journey or explicit T6 reopen.

## Authentication/session

```text
GET    /api/v1/session
DELETE /api/v1/session
GET    /api/v1/authentication/provider-subjects?query=...
```

OIDC redirect/callback remain web-integration routes outside the generated JSON API:

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
PUT    /api/v1/users/{user_id}/provider-binding
POST   /api/v1/users/{user_id}/offboarding
POST   /api/v1/users/{user_id}/reenablement

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

`PUT /areas/{id}/lifecycle` accepts the closed current state `ACTIVE | RETIRED`; repeat of the same desired state is an idempotent no-op. It does not fabricate duplicate Audit transitions.

## Authorization

```text
GET    /api/v1/roles
GET    /api/v1/role-assignments
POST   /api/v1/role-assignments
DELETE /api/v1/role-assignments/{assignment_id}
```

No custom role/permission CRUD.

## Document governance configuration

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
```

`governance` is one whole-replacement governance+representation configuration. No public PolicyVersion resource.

## Controlled Documents / work

```text
GET  /api/v1/documents
POST /api/v1/documents
GET  /api/v1/documents/{document_id}
PUT  /api/v1/documents/{document_id}/responsible-owner
PUT  /api/v1/documents/{document_id}/template-role
POST /api/v1/documents/{document_id}/revisions
GET  /api/v1/documents/{document_id}/history

GET /api/v1/work/authoring
GET /api/v1/work/governance

GET /api/v1/revisions/{revision_id}
GET /api/v1/revisions/{revision_id}/draft
PUT /api/v1/revisions/{revision_id}/draft

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
POST /api/v1/governance-attempts/{attempt_id}/steps/{step_id}/decisions

GET /api/v1/releases/{release_id}
GET /api/v1/releases/{release_id}/source
GET /api/v1/official-renditions/{rendition_id}/content

POST /api/v1/documents/{document_id}/obsolescence-requests
GET  /api/v1/obsolescence-requests/{request_id}
PUT  /api/v1/obsolescence-requests/{request_id}/withdrawal
```

Why use PUT for `withdrawal`/`cancellation`?

They are singleton facts for one exact subject. Natural HTTP idempotency is cheaper and stronger than creating a durable replay row for a command that can be addressed as one resource.

If a repeated PUT has a materially different immutable reason/payload after the fact already exists, return explicit conflict; never rewrite the evidence.

## Audit

```text
GET /api/v1/audit/events
```

No release/publish mutation endpoint. No generic `/actions`. No generic Search endpoint. Library search is `GET /documents`.

---

# 4. T6-E refinement — numbering is closed, not configurable padding

## Candidate disposition: REFINED / ACCEPT

The first candidate retained `sequence_width` configuration without a named consumer. Remove it.

Exactly:

```text
numbering_scope = DOCUMENT_TYPE | DOCUMENT_TYPE_AREA
separator       = "-"
sequence        = decimal; minimum display width 3; expands naturally after 999
```

Examples:

```text
PO-001
PO-RH-003
PO-RH-1000
```

No configurable width, separator, reset, date token or formatting language.

Stable code scalar:

```text
DocumentType.code / Area.code
trim
uppercase ASCII alphanumeric only
1..16 chars
'-' forbidden because product owns separator
```

Area code is immutable after creation.

DocumentType code + numbering scope may be corrected only before the first committed Document of that type; after first use they are immutable. Display name and active status remain independently mutable.

Preview:

```text
GET .../numbering-preview
→ preview_code
→ reservation=false
```

It never writes the counter. Only atomic Document create returns the final code authority.

---

# 5. T6-F refinement — HTTP expression of T2 DRAFT OCC

## Candidate disposition: REFINED / ACCEPT

One existing WorkingContent generation is the single DRAFT edit token for both source and Revision title.

```text
GET /revisions/{id}/draft
→ ETag: "draft-<generation>"

PUT /revisions/{id}/draft
If-Match: "draft-<expected-generation>"
```

`PUT` body is a full accepted DRAFT-edit representation for the fields being governed by that token. It may reference a READY `upload_id`/managed-content claim; **the client never supplies the authoritative ExactContentDescriptor**.

Server re-derives/reloads descriptor from T4 mechanism state before attachment.

Successful title-only or source change increments generation once.

Stale:

```text
412 precondition.draft_changed
```

Not 409.

No second title version token, no auto-merge, no LWW.

---

# 6. T6-G refinement — exact upload/admission routes and authority

## Candidate disposition: REFINED / ACCEPT

```text
POST /revisions/{revision_id}/draft/uploads
→ upload_id + create-only upload target + expiry/constraints

browser PUT bytes directly to provider

POST /revisions/{revision_id}/draft/uploads/{upload_id}/complete
→ server exact read
→ derive SHA256/size/ContentFormat
→ validate
→ READY

PUT /revisions/{revision_id}/draft + If-Match
→ body references upload_id
→ server proves READY + live binding
→ stores server-owned descriptor in WorkingContent
```

`upload_id` is not semantic content identity. It is a bounded mechanism handle to one intended DRAFT/root.

Provider PUT success is not READY; READY is not WorkingContent; WorkingContent is not Submission.

SUBMIT performs required malware preflight server-side before semantic transaction.

```text
scanner unavailable → 503 dependency.malware_inspector_unavailable
malicious           → 422 validation.content_malicious
```

No browser scanner call and no business `scan_status` workflow.

---

# 7. T6-V — DRAFT seed semantics for blank/template/revise

## Candidate disposition: ACCEPT

This is required to prevent implementation from inventing a second Template or copying OfficialRendition instead of source.

### Blank create

`Create blank` seeds WorkingContent from a **product-owned trusted blank DOCX mechanism asset**, not from a semantic Template Document.

The seed is admitted/copied through T4 mechanisms before/around the T2 create transaction so no provider call is required inside the local business transaction. A failed create leaves only reclaimable mechanism output.

### Create from Template

The user selects an eligible Template Document/current EFFECTIVE Revision. The new Document receives an independent managed copy of the exact **released source**, not the source Template's storage identity and not an OfficialRendition.

At commit T2 revalidates that the selected template Revision is still current EFFECTIVE/eligible, then pins DocumentOrigin provenance.

### Create next Revision

A later Revision starts from:

```text
current EFFECTIVE Revision title
+ exact winning released source content
```

copied into a new independent WorkingContent handle.

It does not start from OfficialRendition and does not rebind to a Template.

If the source format is not editable by the selected editor, the author may replace the DRAFT source through the normal T4 upload journey; product semantics do not silently convert it.

---

# 8. T6-I refinement — semantic byte routes and representation truth

## Candidate disposition: REFINED / ACCEPT

Stable semantic byte resources:

```text
GET /revisions/{revision_id}/draft/source
GET /submissions/{submission_id}/source
GET /releases/{release_id}/source
GET /official-renditions/{rendition_id}/content
```

The API/client never exposes `managed_content_id`, bucket, key, provider version or provider URL as product identity.

Baseline contract is an authorized semantic byte gateway supporting full/Range read. Implementation may later answer via an authorization-checked short-lived provider/CDN redirect without changing these semantic URLs.

Viewer rules remain:

```text
SourceOnly PDF  → direct PDF view of Release source
SourceOnly DOCX → selected read-only DOCX adapter on Release source
RequireOfficialRendition(PDF) → OfficialRendition is primary current-effective view; Release source separately exact/downloadable
Governance → exact Submission source is primary decision content
```

Immutable byte HTTP ETag may be derived from semantic SHA-256 for caching, but semantic ExactContentDescriptor remains authority.

---

# 9. T6-J refinement — deterministic DOCX provider proof gate

## Candidate disposition: REFINED / ACCEPT

Do not freeze a vendor from feature marketing.

First candidate = EigenPal/docx-editor-class browser-buffer adapter because the integration naturally matches T4/OCC.

Before implementation plan freezes provider, run one representative fidelity corpus covering at least:

```text
styles/paragraph hierarchy
fonts/emphasis
numbered/bulleted lists
complex + merged tables
headers/footers
images
page + section breaks
links
page settings/margins
multi-page docs
save/reopen in Microsoft Word or LibreOffice
```

Hard fail:

```text
silent OOXML loss/corruption
material layout loss on normal MetalDocs documents
destructive rewrite of unsupported constructs during unrelated edit
invalid/unopenable round-trip document
```

Algorithm:

```text
EigenPal-class candidate passes
→ select one browser DOCX adapter for DRAFT edit + SourceOnly read

materially fails
→ evaluate ONLYOFFICE-class candidate against same corpus
→ prove callback/document-server state cannot bypass T4 or overwrite newer OCC generation

neither passes
→ reopen provider mechanism only
```

No dual-editor Launch runtime.

No EditorSession/lease baseline. Reopen only on an actual selected-provider or measured UX need.

---

# 10. T6-K disposition — Search materialization explicitly rejected for Launch

## Candidate disposition: ACCEPT

No T6 journey names body text/full-text/OCR/vector search.

Launch:

```text
GET /documents
q = code + current EFFECTIVE title
filters = DocumentType + Area + responsible owner
ordinary default = current EFFECTIVE only
```

Deterministic ranking when q:

```text
exact code
→ code prefix
→ title prefix
→ title contains
→ code + stable id tie-break
```

No q:

```text
code + stable id
```

Historical discovery is an authorized canonical query mode and must not mix into ordinary Library by default.

```text
materialized Search = OFF
search_refresh       = OFF
Search rebuild       = N/A
external engine      = OFF
```

Full-text/body/semantic search is a future material reopen, not an implementation convenience.

---

# 11. T6-N refinement — one machine problem authority

## Candidate disposition: REFINED / ACCEPT

RFC 9457 remains the transport model.

To avoid two independent machine identifiers:

```text
code = canonical MetalDocs problem identifier
type = mechanically derived: https://errors.metaldocs.io/{code}
```

`type` is not independently registered/configured.

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

`errors[]` field detail = `{path, code, message}`.

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

No public error family named after implementation modules.

Frontend branches on `code`, never localized `detail` text.

---

# 12. T6-O refinement — idempotency exact contract

## Candidate disposition: REFINED / ACCEPT

Use natural HTTP idempotency first; durable Idempotency-Key only where retry uncertainty can duplicate a semantic creation/decision.

### No Idempotency-Key

```text
GET/HEAD
PUT/DELETE resources whose intended effect is naturally idempotent
DRAFT PUT guarded by If-Match
singleton Submission withdrawal PUT
singleton Revision cancellation PUT
singleton obsolescence-withdrawal PUT
Group membership PUT/DELETE
configuration PUT with If-Match
```

### Idempotency-Key required

Every remaining non-idempotent semantic POST, including:

```text
POST /users
POST /users/{id}/offboarding
POST /users/{id}/reenablement
POST /areas
POST /groups
POST /role-assignments
POST /document-types
POST /documents
POST /documents/{id}/revisions
POST /revisions/{id}/submissions
POST /governance-attempts/{id}/feedback
POST /governance-attempts/{id}/steps/{step_id}/decisions
POST /documents/{id}/obsolescence-requests
```

Upload allocation/completion are mechanism operations and use their own stable upload identity/state; they do not need a durable semantic Idempotency-Key row if exact retry is naturally recognized by `upload_id`.

Server scope:

```text
current User id + canonical operation id + Idempotency-Key
```

Server fingerprints validated semantic command fields, not client hash authority/raw transport bytes.

Behavior:

```text
key missing when required                   → 400 request.idempotency_key_required
first key+fingerprint                       → execute/store exact status+body
completed same key+fingerprint              → exact replay + Idempotent-Replay: true
same key+fingerprint executing              → 409 conflict.idempotency_in_progress
same key + different semantic fingerprint   → 422 validation.idempotency_key_reused
```

Baseline replay retention = **24 hours**.

This is transport fault tolerance, not permanent business dedupe. After expiry, T1/T2/T3 domain uniqueness/eligibility remains authority.

Reopen retention on a concrete offline/long-retry client requirement.

---

# 13. T6-P refinement — pagination

## Candidate disposition: REFINED / ACCEPT

Potentially unbounded lists use exactly:

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

No mandatory total count.

Cursor is bound to filter/order semantics; changing filter/order restarts traversal.

Static roles/small closed product vocabularies may be unpaginated.

No offset pagination and no generic filtering/sorting DSL.

---

# 14. T6-M refinement — admin lost-update and identity semantics

## Candidate disposition: REFINED / ACCEPT

Minimum local UserProfile:

```text
display_name required while profile exists
email optional/contact enrichment
```

Profile can be lawfully erased without deleting User identity/history.

Area:

```text
code immutable
name mutable
lifecycle ACTIVE | RETIRED
```

RoleAssignment exactly:

```text
subject USER | GROUP
role one static T3 code
scope COMPANY | AREA(area_id)
```

DocumentType:

```text
base: name, code, active, numbering_scope
governance whole-subresource: mode + ordered steps + representation
eligible-templates whole-set subresource
```

Whole-replacement mutable admin representations return a strong ETag and require `If-Match` on PUT where concurrent lost update could change authority/governance/current configuration.

At minimum:

```text
company
UserProfile
Area metadata/lifecycle
Group metadata
DocumentType base
DocumentType governance
DocumentType eligible-template set
```

Exact persistence token is implementation design; HTTP stale precondition = 412.

---

# 15. T6-C/T6-Q — create/read DTO semantic rules

## Candidate disposition: ACCEPT

Public read DTOs are purpose-built lenses, not table dumps.

Required read-model families:

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

Read models may contain `allowed_actions` computed live as UX hints.

```text
allowed_actions != grant
backend command always rechecks T2/T3
```

### Create Document command minimum

```text
document_type_id
area_id
title
responsible_user_id optional when actor may manage owner
template_document_id optional
template_revision_id required when template selected
```

The client never supplies final Document code or sequence.

### Create next Revision

No arbitrary content body is required. Server seeds from exact current EFFECTIVE source/title per T6-V, then author edits normally.

---

# 16. Integrated proof matrix added to candidate

Before T6 durable promotion, the design must make later implementation tests unambiguous for:

```text
OIDC auth-code path; no MetalDocs password input
unbound provider subject login denied
CSRF fail-closed on unsafe browser commands
session-derived UI affordance cannot authorize command
old API route families absent
no /api/v2 compatibility layer
DRAFT title+source share one ETag generation
If-Match stale = 412 without mutation
client descriptor cannot become content authority
blank seed != semantic Template
new Revision copies released source, not OfficialRendition
reviewer exact Submission bytes != WorkingContent
natural-idempotent singleton cancellation/withdraw repeats do not create duplicate evidence
Idempotency-Key same request replays; changed request rejects
numbering preview never reserves
DocumentType/Area code immutability after prescribed boundary
Search projection/jobs absent
Domain history reconstructs without Audit
Audit cannot reconstruct current lifecycle
selected DOCX provider passes corpus and cannot bypass OCC/T4
```

---

# 17. Corrected material decision slate

The operator is asked to adjudicate the combined base candidate + corrections as:

```text
T6-A   REFINED ACCEPT — rewrite pre-launch /api/v1 from semantics; keep OAS 3.0.3 for Launch; no compatibility surface.
T6-B   ACCEPT — semantic frontend lenses; My Work includes governance list; exact governance case has separate stable route.
T6-C   REFINED ACCEPT — closed semantic operation census in this packet; no module-shaped/generic action API.
T6-D   ACCEPT — server-derived allowed_actions are UX hints only.
T6-E   REFINED ACCEPT — numbering TYPE|TYPE_AREA, fixed '-', fixed min width 3, no custom grammar; codes become immutable at defined boundary.
T6-F   REFINED ACCEPT — one strong DRAFT ETag/If-Match token covers Revision title + WorkingContent; stale = 412.
T6-G   REFINED ACCEPT — bound upload_id OPEN→READY→If-Match attachment; client never owns descriptor; malware at governed boundary.
T6-H   ACCEPT — My Work author/governance projection; reviewer operates only on exact Submission.
T6-I   REFINED ACCEPT — semantic byte gateway/routes; SourceOnly vs OfficialRendition view law explicit.
T6-J   REFINED ACCEPT — fidelity-gated EigenPal-class first candidate, ONLYOFFICE fallback, one provider, no EditorSession baseline.
T6-K   ACCEPT — Search materialization explicitly OFF; canonical code/title/filter Search only.
T6-L   ACCEPT — domain history and Audit separate.
T6-M   REFINED ACCEPT — minimal Admin Center + ETag lost-update protection + provider-bound existing Users only.
T6-N   REFINED ACCEPT — RFC9457; code canonical/type mechanically derived; closed families incl. dependency/ratelimit.
T6-O   REFINED ACCEPT — natural HTTP idempotency first; Idempotency-Key exact POST set + 24h replay.
T6-P   REFINED ACCEPT — cursor default20/max100; no totals/offset/generic DSL.
T6-Q   REFINED ACCEPT — closed purpose-built read models, no DB/module DTO leakage.
T6-R   ACCEPT — preserve feature-sliced React/TanStack mechanism pattern, replace legacy feature taxonomy.
T6-S   ACCEPT — Keycloak Authorization Code + MetalDocs ApplicationSession + CSRF; no local credentials/JIT user.
T6-V   ACCEPT — blank seed is product mechanism; Template/revise copy exact released SOURCE, never OfficialRendition.
```

Everything else in Product Contract REV001 + T1→T5 + Decision Registry remains frozen.

---

# 18. Current gate

```text
T6 base candidate                 STAGED
T6 corrected adjudication packet THIS FILE / STAGED
operator material adjudication   NEXT
platform-facing T6 summary        NOT YET
T6 durable authority              NOT YET
T7                                NOT OPEN
implementation                    BLOCKED
```

No code/implementation plan may be written from this packet before operator material adjudication and later explicit platform-facing summary ratification.
