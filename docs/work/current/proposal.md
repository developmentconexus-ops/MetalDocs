---
id: t8e-executable-wire-contract-proposal
kind: work
owner: architecture
summary: Active T8-E proposal containing the smallest exact executable-wire candidate and its remaining proof prerequisites.
---

# T8-E executable wire contract — active proposal

> Temporary / non-authoritative. Implementation remains blocked.

## Accepted baseline

Do not duplicate or reinvent the accepted layers. Read:

- `../../reference/t8e-checkpoint.md`
- `../../product/journeys.md`
- `../../decisions/api-operation-census.md`
- `../../decisions/forward-obligations.md`

Current application census: **78 operations**. Operation 79 requires a material Product/T6 reopen.

## Current question

What is the smallest closed executable ledger for all 78 operations that leaves no material wire choice to an implementation Writer while avoiding duplicate lifecycle/AuthZ authority and speculative capability?

## Lead outcome

Use exactly:

```text
one closed OpenAPI component registry
+
one closed 78-row operation ledger
```

The registry owns reusable wire shapes once. Each operation row owns only the choices a Writer would otherwise have to invent:

```text
operation_id
method + path
request component / no body
request-header profile
success status set
success body component / no body
success-header profile
exact Problem.code set
exact filters/order when list-like
request/body limit profile
```

The ledger does **not** restate business authorization predicates, lifecycle transition law, transaction law, ownership eligibility, or persistence mechanics. Those remain in their T1→T8-D owners. `allowed_actions` is a projection from the same canonical Authorization + Controlled Documents decisions, never a second role/action matrix.

This is the current Global Maximum: removing a column leaves a material wire choice to Writers; adding domain predicates, universal envelopes, generic actions, generic filters, policy DSLs, provider state, or future capability adds accidental complexity or duplicate authority.

---

# 1. Wire laws

## 1.1 Application authentication boundary

The `/api/v1` OpenAPI root security requirement is one cookie scheme:

```yaml
MetalDocsSession:
  type: apiKey
  in: cookie
  name: __Host-metaldocs_session
```

All 78 operations require it. Browser OIDC `/auth/login` and `/auth/callback` remain outside this application SSOT. OAS security scopes are **not** used to encode MetalDocs Authorization; T3 remains the sole permission/scope authority.

Session cookie attributes remain the accepted browser law:

```text
Secure
HttpOnly
SameSite=Lax
Path=/
Domain absent
```

`GET /api/v1/session` bootstraps the session-bound CSRF token. Every unsafe operation requires `X-CSRF-Token` in addition to the cookie.

## 1.2 Closed objects

OpenAPI 3.0.3 defaults `additionalProperties` to `true`. Every fixed MetalDocs request, response, reference, page, union branch and Problem object therefore uses:

```text
additionalProperties: false
```

The only current deliberate map is `DraftUploadAllocation.required_headers`, because temporary direct upload must forward provider-required HTTP headers exactly while provider identity remains mechanism only.

There is no generic metadata/settings/facts/property-bag map in the application contract.

## 1.3 Presence and nullability

```text
required = member must be present
optional = member may be absent
nullable = explicit JSON null is a semantic value
```

OAS 3.0.3 `nullable: true` is used only for explicit null. Absence and null are not interchangeable. PATCH/PUT never acquires implicit “null means delete”. Request and response components are purpose-built rather than using `readOnly`/`writeOnly` to hide an overbroad shared DTO.

The baseline nullable member remains `Page.next_cursor`. All other optional members below are absent when their semantic fact does not exist unless a component explicitly states otherwise.

## 1.4 Composition

Do not use `allOf` inheritance merely to reduce YAML. True closed semantic unions use `oneOf` with a required discriminator; ordinary reuse nests `$ref` components. A generator limitation may change encoding only when semantic exactness remains unchanged.

## 1.5 JSON vocabulary and code normalization

- JSON member names: `snake_case`.
- Semantic enums described upstream in uppercase/PascalCase normalize to lower snake case.
- Canonical Product identifiers keep their accepted spelling; T3 `PermissionCode` remains dot-separated (`document.read_effective`, etc.).
- Unknown enum values are invalid; there is no `other`/future catch-all.

Area/DocumentType code requests use `CodeInput`:

```text
already trimmed ASCII alphanumeric
case-insensitive input
maxLength 32
```

Server normalization is exactly uppercase ASCII. Responses use canonical `CodeToken` (`^[A-Z0-9]+$`). `-` is forbidden inside a token because Product owns it as the numbering separator.

## 1.6 Strong ETag domain law

Each ETag is a strong opaque validator bound to exactly one canonical representation concurrency domain. It does not expose raw generation/version and cannot be transferred between resource/subresource domains.

```text
GET canonical representation -> only canonical source of that domain's ETag
If-Match from another domain/resource -> syntactically valid but false -> 412
stale same-domain tag -> 412
missing/malformed required header -> 400
```

An ETag-protected representation may contain only:

```text
fields governed by that concurrency token
+ stable immutable identifiers/references
```

It must not embed independently mutable display enrichment that could change while the ETag remains unchanged. Therefore, for example:

```text
ResponsibleOwnerView       -> responsible_owner_user_id, not UserProfile display name
EligibleTemplatesView      -> stable DocumentReference only, not current Revision title
```

Display-rich aggregate read models remain allowed when they are not the canonical conditional-write representation.

## 1.7 Aggregate JSON request limit

All `/api/v1` `application/json` request bodies have one raw-body ceiling:

```text
65,536 bytes
```

The eventual OAS carries:

```text
x-metaldocs-max-request-body-bytes: 65536
```

on every JSON request body. Central request handling enforces the raw limit before JSON decoding.

Request compression is absent from Launch. A non-identity `Content-Encoding` on a JSON request returns `415 request.unsupported_media_type`; that response includes `Accept-Encoding: identity`. A 415 caused only by media type does not include `Accept-Encoding`.

Why one ceiling instead of many guessed field maxima:

- API resource-consumption guidance requires bounded incoming payloads;
- document bytes never traverse an application JSON body;
- the largest current commands are bounded configuration/UUID arrays and route steps;
- 64 KiB supplies large operational headroom while removing an unbounded parser/memory surface;
- raising this transport ceiling later does not change semantic ownership.

Reopen trigger: a measured legitimate Launch command cannot be represented within 64 KiB.

## 1.8 Opaque/query bounds

```text
SearchQuery / provider query   maxLength 256
OpaqueCursor                   maxLength 2048
CsrfToken                      maxLength 512
ProviderSubjectRef             maxLength 2048
CodeInput / CodeToken          maxLength 32
```

Server-produced/user-provided UUIDs retain UUID format. Revision ordinals and byte counts use JSON integers with maximum `9007199254740991` so the generated TypeScript boundary remains exact rather than relying on unsafe IEEE-754 integers. The measured document limit will be far smaller than that generic byte-count ceiling.

## 1.9 Document search normalization

`q` is trimmed and must be nonblank when supplied. Matching is:

```text
Document code: ASCII case-insensitive against canonical uppercase code
Revision title: case-insensitive, accent-sensitive
no fuzzy matching
no stemming
no accent folding
```

Accepted ranking remains:

```text
q present:
exact code
-> code prefix
-> title prefix
-> title contains
-> code + document_id tie-break

q absent:
code + document_id
```

Cursor binding uses the normalized query/filter tuple and this ordering.

## 1.10 Exact-byte delivery

Semantic byte GETs return authenticated application-origin `200`, never provider redirect, 206, 304 or transformed content.

Exact success headers:

```text
Content-Type
Content-Length
Content-Disposition: inline; filename="<document_code>-REV<ordinal-min-width-3>.<ext>"
Content-Digest: sha-256=:<base64 exact SHA-256 bytes>:
Cache-Control: private, no-store, no-transform
Accept-Ranges: none
X-Content-Type-Options: nosniff
```

Exact MIME mapping:

```text
docx -> application/vnd.openxmlformats-officedocument.wordprocessingml.document
pdf  -> application/pdf
```

`Content-Encoding` is absent. Filename is server-generated ASCII from stable Document code + Revision ordinal only; title/user/provider text never enters it. Revision ordinal display width is minimum three and expands naturally after 999.

A `Range` request fails `400 request.invalid`; HEAD is undeclared and follows the 405 router law. Exact-byte missing/corrupt semantic content fails `500 internal.content_integrity`; temporary storage/dependency failure uses `503 dependency.unavailable`.

---

# 2. Header profiles

## 2.1 Request profiles

```text
SAFE_READ
  MetalDocsSession only

UNSAFE_CSRF
  MetalDocsSession
  X-CSRF-Token required

IDEMPOTENT_CREATE
  MetalDocsSession
  X-CSRF-Token required
  Idempotency-Key required; client-generated UUID

IF_MATCH_MUTATION
  MetalDocsSession
  X-CSRF-Token required
  If-Match required
  exactly one strong entity-tag
  `*`, weak tags and lists forbidden

SUBMISSION_CREATE
  MetalDocsSession
  X-CSRF-Token required
  Idempotency-Key required; client-generated UUID
  If-Match required from exact DRAFT domain

PROFILE_REPLACE
  MetalDocsSession
  X-CSRF-Token required
  existing profile -> If-Match required
  absent profile recreation -> If-None-Match exactly `*` required
  both conditional headers together -> invalid request
```

Missing/malformed required conditional/idempotency headers are `request.invalid`; a valid stale/wrong-domain precondition is 412.

## 2.2 Success profiles

```text
NO_STORE
  Cache-Control: no-store

JSON_NO_STORE
  Content-Type: application/json
  Cache-Control: no-store

JSON_ETAG
  Content-Type: application/json
  ETag: required current strong opaque entity-tag
  Cache-Control: no-store

JSON_ETAG_MUTATION
  Content-Type: application/json
  ETag: required resulting/current strong opaque entity-tag
  Cache-Control: no-store

SESSION_END
  Cache-Control: no-store
  Set-Cookie clears __Host-metaldocs_session with:
    empty value; Path=/; Max-Age=0; Secure; HttpOnly; SameSite=Lax; Domain absent

EXACT_BYTES
  §1.10 exact-byte header set
```

No baseline `Location`, replay-indicator, permission snapshot, provider ID, or generic metadata response header exists.

Problem responses use `Content-Type: application/problem+json` and `Cache-Control: no-store`.

```text
401 -> WWW-Authenticate: MetalDocsSession
405 -> exact Allow for the known path
429 -> Retry-After MAY be present; when present it is non-negative delta-seconds
415 due unsupported content coding -> Accept-Encoding: identity
```

A rate limiter that cannot truthfully predict a retry time does not fabricate `Retry-After`.

---

# 3. Shared component registry

All fixed objects are closed.

## 3.1 Field type convention

Unless explicitly overridden:

```text
*_id                         Uuid
*_at                         UtcInstant
provider_subject_ref         ProviderSubjectRef
email                        EmailAddress
sha256                       Sha256Hex
size_bytes / max_bytes       ByteCount
revision ordinal             RevisionOrdinal
Area/DocumentType response code  CodeToken
Area/DocumentType request code   CodeInput
Document code                DocumentCode
name/display_name/title/label/reason/message  NonBlankString
```

These conventions are normative schema shorthand, not implementation inference.

## 3.2 Scalars

```text
Uuid
  string; format=uuid

UtcInstant
  string; format=date-time
  canonical server serialization = RFC3339 UTC `Z`

OpaqueCursor
  nonblank string; maxLength=2048

IdempotencyKey
  UUID

CsrfToken
  nonblank string; maxLength=512

ProviderSubjectRef
  nonblank opaque string; maxLength=2048

CorrelationId
  nonblank opaque string

Sha256Hex
  lowercase hex; pattern ^[0-9a-f]{64}$

NonBlankString
  string whose trimmed semantic value is nonblank

SearchQuery
  NonBlankString; maxLength=256

CodeInput
  string; pattern ^[A-Za-z0-9]+$; maxLength=32

CodeToken
  string; pattern ^[A-Z0-9]+$; maxLength=32

DocumentCode
  string; pattern ^[A-Z0-9]+(?:-[A-Z0-9]+)?-[0-9]{3,}$
  maxLength=85

EmailAddress
  string; format=email

RevisionOrdinal
  integer; minimum=0; maximum=9007199254740991

ByteCount
  integer; minimum=0; maximum=9007199254740991
```

Other human text is bounded by the aggregate request ceiling rather than unrelated guessed per-field maxima.

## 3.3 References

```text
UserReference
  required: user_id
  optional: display_name
  display_name absent when erasable UserProfile enrichment is absent
  never includes email/provider/grant data

AreaReference
  required: area_id, code, name

DocumentTypeReference
  required: document_type_id, code, name

DocumentReference
  required: document_id, code
  no title: title belongs Revision

RevisionIdentity
  required: revision_id, ordinal

RevisionReference
  required: revision, title
  revision: RevisionIdentity

ContentSummary
  required: sha256, size_bytes, content_format
```

A Submission never reuses a live `RevisionReference` as historical title authority; it carries frozen `title` beside `RevisionIdentity`.

## 3.4 Closed enums

```text
ContentFormat
  docx | pdf

RevisionState
  draft | submitted | effective | superseded | obsolete | cancelled

OpenRevisionState
  draft | submitted

UserEligibilityState
  enabled | disabled

AreaLifecycleState
  active | retired

NumberingScope
  document_type | document_type_area

GovernanceMode
  no_human_approval | use_governance_route

GovernanceSelectorKind
  named_user | group

GovernanceDecisionOutcome
  accept | return_for_changes

GovernanceSubjectKind
  submission | obsolescence

GovernanceAttemptState
  active | completed | returned | withdrawn | cancelled

GovernanceStepState
  pending | active | decided

RepresentationKind
  source_only | require_official_rendition

OfficialRenditionFormat
  pdf

RoleCode
  governance_admin | area_manager | author | approver | viewer | governance_viewer

PermissionCode
  organization.manage
  access.manage
  document_type.manage
  template_use.manage
  document.read_effective
  document.read_history
  document.read_working
  document.create
  document.edit
  document.submit
  document.cancel_revision
  document.obsolete
  document.owner.manage
  governance.act
  audit.read

RoleAssignmentSubjectKind
  user | group

RoleAssignmentScopeKind
  company | area

DocumentCatalogStatus
  effective | obsolete | cancelled

DocumentOfficialStatus
  draft | submitted | effective | obsolete | cancelled

SubmissionCreationState
  governance_pending | rendition_pending | released

SubmissionTerminationKind
  returned_for_changes | withdrawn | cancelled

ObsolescenceRequestState
  active | returned | withdrawn | completed

ObsolescenceCreationState
  governance_pending | obsolete

GovernanceCaseAction
  accept | return_for_changes | add_feedback
```

`DocumentType.active` is a boolean, matching T8-D. No redundant active/inactive enum is introduced. OPEN/READY/GC_PENDING upload states remain mechanism-only.

## 3.5 True unions

```text
RoleAssignmentSubject
  user  -> { kind:user, user_id }
  group -> { kind:group, group_id }

RoleAssignmentScope
  company -> { kind:company }
  area    -> { kind:area, area_id }

GovernanceSelector
  named_user -> { kind:named_user, user_id }
  group      -> { kind:group, group_id }

GovernancePolicy
  no_human_approval -> { mode:no_human_approval }
  use_governance_route -> { mode:use_governance_route, steps:[GovernanceRouteStep, ...] }

GovernanceRouteStep
  required: label, selector

RepresentationPolicy
  source_only -> { kind:source_only }
  require_official_rendition -> { kind:require_official_rendition, format:pdf }
```

`GovernanceRouteStep.label` exposes the bounded T8-D contradiction in §8; it is not deleted to accommodate persistence.

## 3.6 Pagination

```text
Page
  required: next_cursor, has_more
  next_cursor: OpaqueCursor | null
  has_more: boolean
```

Potentially unbounded lists return `{items,page}`. `limit` is integer `1..100`, default `20`; `cursor` is opaque. No offset, total count, generic sort, frozen snapshot, or server cursor state.

Cursor integrity binds operationId + normalized filters + deterministic ordering. Current Authorization is rechecked on every page.

---

# 4. RFC 9457 Problem catalog

## 4.1 Exact base law

Every Problem code is its own full closed schema; there is no open base inherited with `allOf`.

Common required members:

```text
type
title
status
detail
instance
code
trace_id
```

`instance` is a fresh `urn:uuid:<uuid>` per occurrence. `trace_id` is opaque/nonblank correlation text; clients do not parse provider format.

Only `request.invalid` and validation-family variants may contain optional non-empty `errors[]`:

```text
ProblemError
  required: pointer, detail
  pointer pattern begins /path, /query, /header, or /body
  RFC 6901 escaping applies
  rejected sensitive values are never echoed
```

Each variant freezes `type`, `title`, `status`, and `code` with single-value enums where OAS 3.0.3 lacks `const`.

## 4.2 Closed catalog

| code | status | title |
|---|---:|---|
| `request.invalid` | 400 | Invalid request |
| `auth.unauthenticated` | 401 | Authentication required |
| `permission.denied` | 403 | Permission denied |
| `permission.csrf_failed` | 403 | Request trust verification failed |
| `notfound.resource` | 404 | Resource not found |
| `request.method_not_allowed` | 405 | Method not allowed |
| `state.conflict` | 409 | Resource state conflict |
| `state.governance_step_already_decided` | 409 | Governance step already decided |
| `state.upload_expired` | 410 | Upload expired |
| `precondition.resource_changed` | 412 | Resource changed |
| `precondition.draft_changed` | 412 | Draft changed |
| `request.content_too_large` | 413 | Request content too large |
| `request.unsupported_media_type` | 415 | Unsupported media type |
| `validation.failed` | 422 | Validation failed |
| `validation.idempotency_key_reused` | 422 | Idempotency key reused |
| `validation.content_invalid` | 422 | Invalid document content |
| `validation.content_malicious` | 422 | Malicious content rejected |
| `ratelimit.exceeded` | 429 | Too many requests |
| `internal.failure` | 500 | Internal server error |
| `internal.content_integrity` | 500 | Content integrity failure |
| `dependency.unavailable` | 503 | Service dependency unavailable |
| `dependency.malware_inspector_unavailable` | 503 | Malware inspector unavailable |

Type is mechanically:

```text
https://errors.conexus.fun/metaldocs/{code}
```

Existing upstream exact codes are preserved: `precondition.resource_changed`, `precondition.draft_changed`, `validation.idempotency_key_reused`, `validation.content_malicious`, `dependency.malware_inspector_unavailable`, and `state.governance_step_already_decided`.

No module/provider/storage/database/scanner error text escapes.

## 4.3 Disclosure law

```text
no valid ApplicationSession                         -> 401
valid session, visible route/action but no permission/trust -> 403
item absent OR caller may not learn item existence  -> 404
```

Collection/read-model operations may return an authorized empty result instead of manufacturing 403 when the Product lens is defined by per-item visibility. A requested management catalog mode that itself requires missing authority may return 403.

## 4.4 Ledger shorthand — exact expansion only

These are textual exact-set macros. The final OAS expands every operation to concrete Problem variants; there is no runtime/common-error inheritance.

```text
B = request.invalid
  + auth.unauthenticated
  + ratelimit.exceeded
  + internal.failure
  + dependency.unavailable

A = B + permission.denied
C = B + permission.csrf_failed
U = A + permission.csrf_failed

J = request.unsupported_media_type
  + request.content_too_large
  + validation.failed

I = validation.idempotency_key_reused
N = notfound.resource
S = state.conflict
P = precondition.resource_changed
D = precondition.draft_changed
X = internal.content_integrity
```

429 is allowed cross-cutting because it is already in the accepted T8-E baseline and resource-consumption protection is transport safety, not business authority. T8-E freezes response shape, not rate policy.

Router-level laws outside the 78 operations:

```text
unknown /api/v1 path -> 404 notfound.resource
undeclared method on known path -> 405 request.method_not_allowed + Allow
```

---

# 5. Components — Session / Organization / Access / Document Governance

## 5.1 Session / AuthN

```text
SessionView
  required: user, csrf_token
  user: UserReference

ProviderSubjectOption
  required: provider_subject_ref, display_hints
  display_hints: string[]; maxItems=3; each maxLength=256

ProviderSubjectSearchView
  required: items
  items: ProviderSubjectOption[]; maxItems=20
```

Provider search accepts required `query: SearchQuery`; provider result order is preserved. It is a bounded external preflight, not a product catalog and not cursor-paginated.

## 5.2 Organization

```text
CompanyView
  required: company_id, display_name

ReplaceCompanyRequest
  required: display_name

UserProfileInput
  required: display_name
  optional: email

CreateUserRequest
  required: provider_subject_ref, profile
  successful create establishes ENABLED User + required profile + binding atomically

CreateUserResult
  required: user_id

UserView
  required: user, eligibility
  aggregate read only; eligibility subresource remains canonical ETag source

UserPage
  required: items, page
  items: UserView[]

UserProfileView
  required: user_id, display_name
  optional: email

ReplaceUserProfileRequest
  required: display_name
  optional: email
  whole replacement: omitted email => resulting profile has no email

UserProviderBindingView
  required: user_id, provider_subject_ref

ReplaceUserProviderBindingRequest
  required: provider_subject_ref

UserEligibilityView
  required: user_id, state

ReplaceUserEligibilityRequest
  required: state

AreaView
  required: area_id, code, name

AreaSummary
  required: area, state

AreaPage
  required: items, page

CreateAreaRequest
  required: code, name
  code: CodeInput
  new Area starts active; creating an already-retired Area is absent

CreateAreaResult
  required: area_id

ReplaceAreaRequest
  required: name
  Area code is immutable and absent from replacement

AreaLifecycleView
  required: area_id, state

ReplaceAreaLifecycleRequest
  required: state

GroupView
  required: group_id, name

GroupPage
  required: items, page

CreateGroupRequest / ReplaceGroupRequest
  required: name

CreateGroupResult
  required: group_id

GroupMemberPage
  required: items, page
  items: UserReference[]
```

Profile erasure DELETE is “ensure enrichment absent”, naturally idempotent, and does not require If-Match. Concurrent recreation/replacement still serializes through owner truth.

## 5.3 Authorization

```text
RoleView
  required: code, permissions, allowed_scope_kinds
  permissions: unique PermissionCode[] in canonical T3 permission order
  allowed_scope_kinds: unique scope kinds; company before area

RoleListView
  required: items
  items: RoleView[] in canonical T3 role order

RoleAssignmentView
  required: assignment_id, subject, role, scope

RoleAssignmentPage
  required: items, page

CreateRoleAssignmentRequest
  required: subject, role, scope

CreateRoleAssignmentResult
  required: assignment_id
```

Roles/bundles are static Product truth; this is not an editable role/policy API.

## 5.4 Document Governance

```text
DocumentTypeView
  required: document_type_id, code, name, numbering_scope, active

DocumentTypePage
  required: items, page

CreateDocumentTypeRequest
  required: code, name, numbering_scope, active, governance, representation
  code: CodeInput
  governance: GovernancePolicy
  representation: RepresentationPolicy
  eligible-template set starts empty

CreateDocumentTypeResult
  required: document_type_id

ReplaceDocumentTypeRequest
  required: code, name, numbering_scope, active
  code: CodeInput
  normalized code/numbering_scope change after first committed Document -> state.conflict

DocumentTypeGovernanceView
  required: governance, representation

ReplaceDocumentTypeGovernanceRequest
  required: governance, representation

EligibleTemplatesView
  required: templates
  templates: DocumentReference[] ordered by document code, id
  only stable references are present because this is an ETag concurrency representation

ReplaceEligibleTemplatesRequest
  required: template_document_ids
  unique UUID array; empty array valid

NumberingPreviewView
  required: preview_code, reservation
  preview_code: DocumentCode
  reservation is constant false

TemplateConfigurationItem
  required: document, template_role, has_effective_revision, eligible_document_type_ids
  optional: current_effective_title

TemplateConfigurationPage
  required: items, page
```

Creating a DocumentType carries initial governance/representation explicitly because T8-D requires those values and no accepted default exists. This avoids hidden Writer-selected `NoHumanApproval`/`SourceOnly` defaults while preserving separate later ETags.

---

# 6. Components — Controlled Documents / Work

## 6.1 Creation / official projections

```text
TemplateCreationOption
  required: document, effective_revision
  effective_revision: RevisionReference

DocumentCreationOptionsView
  required: areas, document_types, templates, default_responsible_owner
  optional: responsible_owner_candidates
  absence of responsible_owner_candidates = caller lacks owner-manage selection capability
  present empty array = capability exists but no alternate eligible target

CreateDocumentRequest
  required: document_type_id, area_id, title
  optional: template_document_id, responsible_owner_user_id
  absence of template_document_id = trusted blank DOCX seed

CreateDocumentResult
  required: document_id, revision_id
  excludes code/title/free text from replay result

DocumentSummary
  required: document, document_type, area, responsible_owner, status
  optional: official_revision
  official_revision: RevisionReference

DocumentPage
  required: items, page

ReleasedRevisionView
  required: revision, title, release_id, released_at, source, representation
  source: ContentSummary
  representation:
    source_only -> { kind:source_only }
    official_rendition -> { kind:official_rendition, official_rendition_id, content:ContentSummary }

DocumentOfficialView
  required: document, document_type, area, responsible_owner, status
  optional: official
  official: ReleasedRevisionView
```

Cross-field laws:

```text
official present iff at least one Release exists
obsolete may retain the last released `official` view
newer cancelled/open work never replaces older EFFECTIVE official truth
pre-first-release draft/submitted/cancelled has no `official`
```

Library status query is only `effective|obsolete|cancelled`, default `effective`; DRAFT/SUBMITTED remain Work lenses.

## 6.2 Current relationships / work

```text
ResponsibleOwnerView
  required: document_id, responsible_owner_user_id

ReplaceResponsibleOwnerRequest
  required: user_id

TemplateRoleView
  required: document_id, is_template

ReplaceTemplateRoleRequest
  required: is_template

CreateRevisionResult
  required: revision_id

RevisionView
  required: revision, document, title, state, created_at
  optional: current_submission_id
  current_submission_id present iff state=submitted

DocumentWorkView
  required: document, revision, title, content, updated_at
  generation absent: opaque strong ETag is wire concurrency token

UpdateDraftRequest
  optional: title, source_upload_id
  at least one member required
  null forbidden
  omitted = unchanged

DraftUploadAllocation
  required: upload_id, upload_url, expires_at, max_bytes, required_headers
  upload_url: temporary URI capability
  max_bytes: ByteCount; value blocked on §9 measurement
  required_headers: map<string,string>
  expires_at <= allocation time + 15 minutes
```

Allocation exposes no provider account/bucket/key/version/ETag. Completion is bodyless, naturally idempotent, returns no authoritative descriptor, and server independently derives descriptor before READY.

## 6.3 Submission

```text
SubmissionCreateResult
  governance_pending -> { state:governance_pending, submission_id, governance_attempt_id }
  rendition_pending -> { state:rendition_pending, submission_id }
  released -> { state:released, submission_id, release_id }

SubmissionHumanGate
  required: required, satisfied

SubmissionRepresentationGate
  required: required, satisfied, attention_required

SubmissionView
  required:
    submission_id
    revision
    title
    submitter
    submitted_at
    content
    governance_mode
    representation
    human_gate
    representation_gate
  optional:
    governance_attempt_id
    release_id
    termination
```

Cross-field laws:

```text
title = immutable Submission title snapshot
governance_attempt_id present iff governance_mode=use_governance_route
human_gate.required iff governance_mode=use_governance_route
human_gate.required=false -> human_gate.satisfied=true
representation_gate.required iff representation=require_official_rendition
representation_gate.required=false -> representation_gate.satisfied=true
attention_required=true only when representation required + not satisfied + terminal renderer attention exists
release_id and termination are mutually exclusive
release_id present only when this Submission won Release
```

No renderer/River/provider job identity/state appears.

```text
SubmissionWithdrawalView
  required: submission_id, actor, withdrawn_at

RevisionCancellationRequest
  required: reason

RevisionCancellationView
  required: revision_id, actor, reason, cancelled_at
```

Cancellation singleton law:

```text
first cancellation -> 201
same reason exact repeat -> existing 200
later different reason -> 409 state.conflict
```

## 6.4 Governance / feedback

```text
GovernanceFeedbackView
  required: feedback_id, actor, message, created_at

GovernanceFeedbackPage
  required: items, page

CreateGovernanceFeedbackRequest
  required: message

CreateGovernanceFeedbackResult
  required: feedback_id

GovernanceDecisionRequest
  accept -> { outcome:accept }
  return -> { outcome:return_for_changes, reason }

GovernanceDecisionView
  accept -> { decision_id, outcome:accept, actor, decided_at }
  return -> { decision_id, outcome:return_for_changes, actor, decided_at, reason }

GovernanceStepView
  pending -> { step_id, ordinal, label, state:pending }
  active  -> { step_id, ordinal, label, state:active }
  decided -> { step_id, ordinal, label, state:decided, decision:GovernanceDecisionView }
```

No selector candidate snapshot, Group membership, RoleAssignment, grant expansion, or provider claim appears in a case.

```text
SubmissionGovernanceSubject
  required: kind=submission, submission_id, document, revision, title, submitter, submitted_at, content

ObsolescenceGovernanceSubject
  required: kind=obsolescence, request_id, document, target_revision, initiator, reason, requested_at

GovernanceCaseView
  required: governance_attempt_id, state, subject, steps, feedback, allowed_actions
  subject: SubmissionGovernanceSubject | ObsolescenceGovernanceSubject
  steps ordered by ordinal
  feedback = same first 20 items as listGovernanceFeedback with no cursor
  embedded feedback.page.next_cursor, when present, is minted for listGovernanceFeedback
  allowed_actions: unique GovernanceCaseAction[]; may be empty
```

`allowed_actions` is exactly:

```text
accept
return_for_changes
add_feedback
```

It derives from canonical current T3 + Controlled Documents truth; every command rechecks truth.

Governance decision singleton law:

```text
first decision -> 201
same outcome + same required reason exact repeat -> existing 200
any later different outcome/reason -> 409 state.governance_step_already_decided
```

## 6.5 History / work lists

Every `DocumentHistoryItem` branch has a required `kind` discriminator matching the branch:

```text
revision_created
  revision, title, occurred_at

submission_created
  submission_id, revision, title, submitter, occurred_at
  optional governance_attempt_id

governance_decision
  decision_id, governance_attempt_id, step_id, actor, outcome, occurred_at
  optional reason iff outcome=return_for_changes

feedback_added
  feedback_id, governance_attempt_id, actor, message, occurred_at

submission_withdrawn
  submission_id, actor, occurred_at

revision_cancelled
  revision_id, actor, reason, occurred_at

release_completed
  release_id, revision_id, submission_id, occurred_at
  optional predecessor_revision_id

official_rendition_completed
  official_rendition_id, submission_id, occurred_at

obsolescence_requested
  request_id, target_revision_id, initiator, reason, occurred_at
  optional governance_attempt_id

obsolescence_withdrawn
  request_id, actor, occurred_at

obsolescence_completed
  request_id, target_revision_id, occurred_at
```

```text
DocumentHistoryPage
  required: items, page
  order: occurred_at ASC + kind + semantic stable id

WorkAuthoringItem
  required: document, revision, title, state, responsible_owner, updated_at
  state: OpenRevisionState

WorkAuthoringPage
  required: items, page
  order: document.code ASC, document_id ASC

WorkGovernanceItem
  required: governance_attempt_id, subject_kind, document, created_at

WorkGovernancePage
  required: items, page
  order: governance_attempt_id ASC
```

No generic work/action DTO or priority/SLA/ranking is invented.

## 6.6 Release / rendition / obsolescence

```text
ReleaseView
  source_only ->
    { release_id, document, revision, title, submission_id, released_at,
      optional predecessor_revision_id,
      representation:{kind:source_only, source:ContentSummary} }

  official_rendition ->
    { release_id, document, revision, title, submission_id, released_at,
      optional predecessor_revision_id,
      representation:{kind:official_rendition, source:ContentSummary,
                      official_rendition_id, official_rendition:ContentSummary} }

ObsolescenceRequestCreateRequest
  required: reason

ObsolescenceCreateResult
  governance_pending -> { state:governance_pending, request_id, governance_attempt_id }
  obsolete -> { state:obsolete, request_id }
```

`ObsolescenceRequestView` is a state-discriminated union:

```text
active
  required: request_id, document, target_revision, initiator, reason,
            state=active, requested_at, governance_attempt_id
  ended_at absent

returned
  same core fields + state=returned + governance_attempt_id + ended_at

withdrawn
  same core fields + state=withdrawn + governance_attempt_id + ended_at

completed-human
  same core fields + state=completed + governance_attempt_id + ended_at

completed-no-human
  same core fields + state=completed + ended_at
  governance_attempt_id absent
```

```text
ObsolescenceWithdrawalView
  required: request_id, actor, withdrawn_at
```

Release is immutable/system-owned; no public publish mutation exists.

---

# 7. Audit wire

## 7.1 Closed audit vocabularies

```text
AuditActor
  user   -> { kind:user, user_id }
  system -> { kind:system, system_actor_code:metaldocs }

AuditVisibility
  company -> { kind:company }
  area    -> { kind:area, area_id }

AuditResourceKind
  provider_binding
  user
  user_profile
  area
  group
  role_assignment
  document_type
  document
  revision
  submission
  governance_decision
  official_rendition
  release
  obsolescence_request
```

There is intentionally no `group_membership` resource kind because T8-D GroupMembership has composite `(group_id,user_id)` identity and no invented UUID. Membership events use stable Group resource identity and typed membership facts.

`AuditOperationCode` is the closed T3 census projection:

```text
provider_binding.accepted
provider_binding.disabled
provider_binding.replaced
user.created
user.offboarded
user.reenabled
user_profile.erased
area.created
area.renamed
area.retired
area.reenabled
group.created
group.renamed
group.deleted
group_membership.added
group_membership.removed
role_assignment.granted
role_assignment.revoked
document_type.created
document_type.reconfigured
document_type.activated
document_type.inactivated
document_governance.changed
template_eligibility.changed
document.responsible_owner_changed
document.template_role_changed
document.created
revision.created
submission.created
governance.accepted
governance.returned_for_changes
submission.withdrawn
revision.cancelled
official_rendition.completed
release.completed
obsolescence.requested
obsolescence.withdrawn
obsolescence.completed
```

## 7.2 Typed public facts

Audit storage may retain additional bounded owner-authored facts permitted by T3. The Launch wire projects only the typed facts required by current Audit inspection; it never exposes raw JSONB.

```text
none
  { kind:none }

group_membership
  { kind:group_membership, user_id, group_id }

role_assignment
  { kind:role_assignment, assignment_id, subject, role, scope }

governance_decision
  { kind:governance_decision, governance_attempt_id, step_id, decision_id,
    subject_kind, subject_id, outcome }

release
  { kind:release, document_id, revision_id, submission_id,
    optional predecessor_revision_id }

revision_evidence
  { kind:revision_evidence, document_id, revision_id, evidence_id }
```

Exact operation -> resource/facts mapping:

```text
provider_binding.*                         -> provider_binding / none
user.*                                     -> user / none
user_profile.erased                        -> user_profile (resource_id=user_id) / none
area.*                                     -> area / none
group.created|renamed|deleted              -> group / none
group_membership.added|removed             -> group (resource_id=group_id) / group_membership
role_assignment.granted|revoked            -> role_assignment / role_assignment
document_type.*                            -> document_type / none
document_governance.changed                -> document_type / none
template_eligibility.changed               -> document_type / none
document.responsible_owner_changed         -> document / none
document.template_role_changed             -> document / none
document.created                           -> document / none
revision.created                            -> revision / none
submission.created|withdrawn               -> submission / none
governance.accepted|returned_for_changes   -> governance_decision / governance_decision
official_rendition.completed               -> official_rendition / none
release.completed                           -> release / release
revision.cancelled                          -> revision / revision_evidence
obsolescence.*                              -> obsolescence_request / revision_evidence
```

`AuditEventView` is a closed union whose operation-code branch fixes the matching resource/facts family above:

```text
required: event_id, occurred_at, actor, operation_code,
          resource_kind, resource_id, visibility, facts
optional: correlation_id
```

`correlation_id` is non-semantic correlation metadata only.

```text
AuditEventPage
  required: items, page
  order: occurred_at DESC, event_id DESC
```

No Audit filter is added by inference. `GET /audit/events` accepts only cursor/limit; `audit.read` historical visibility filtering happens before pagination.

---

# 8. Material contradiction — bounded T8-D reopen prerequisite

## 8.1 Evidence

Product Contract/T6 makes Governance Route Step a human-facing product concept with labels such as:

```text
Revisão técnica
Gestor
Qualidade
```

The executable wire therefore needs `GovernanceRouteStep.label`, and Governance Case must preserve the exact frozen label shown for each Step.

Ratified T8-D currently persists current route steps as:

```text
document_type_id
ordinal
selector_kind
named_user_id / group_id
```

and frozen attempt steps as:

```text
id
attempt_id
ordinal
selector_kind
named_user_id / group_id snapshot
state
activated_at
```

Neither preserves the Product Step label. Deriving a historical case from the current DocumentType label would let a later config rename rewrite an earlier frozen GovernanceAttempt, contradicting the ratified coherent-route snapshot law.

## 8.2 Method classification

```text
KNOWN          Product requires Step label meaning
KNOWN          GovernanceAttempt freezes route snapshot
KNOWN          T8-D current + snapshot persistence omit label
MATERIAL       historical governed context + persistent meaning
CAUSE          bounded T8-D completeness omission exposed by executable wire
NOT A REASON   preference, legacy shape, generator convenience
OUTCOME        STOP / SPLIT PREREQUISITE for label-dependent promotion only
```

Unrelated T8-E closure continues.

## 8.3 Smallest correction if operator ratifies bounded reopen

Exactly two persistence additions; zero new tables/owners/lifecycles/API operations:

```text
controlled_docs.document_type_governance_steps
  + label TEXT NOT NULL

controlled_docs.governance_attempt_steps
  + label_snapshot TEXT NOT NULL
```

Attempt creation copies the exact configured label into `label_snapshot` in the coherent route-snapshot transition. Frozen snapshot label is immutable.

No candidate list, Role semantics, workflow metadata, localization framework, or PolicyVersion aggregate is added.

T8-E may specify the intended wire, but cannot be ratified/promoted until this contradiction is explicitly adjudicated and T8-D is reconciled.

---

# 9. Limits still requiring evidence

## 9.1 Idempotency replay snapshot — CLOSED

The 10 accepted Idempotency-Key POSTs now have an exact success census. Their durable replay bodies contain only UUIDs plus small closed state enums.

Current largest compact JSON success body is under 160 bytes; no replay result contains UserProfile PII, title, reason, feedback, provider data, or arbitrary text.

Freeze:

```text
ReplaySnapshot payload maximum = 2,048 bytes
```

This leaves >10x current body headroom for a small versioned self-contained status/body encoding without making replay storage an open blob. If a future material wire change needs more, that change already reopens the affected contract.

## 9.2 Document admission — OPEN / measured prerequisite

Do not guess raw DOCX/PDF ceilings from internet defaults.

Before T8-E close, measure a representative controlled-document corpus and tooling behavior, then freeze:

```text
maximum raw document bytes
maximum structurally expanded DOCX bytes
maximum ZIP entry count
maximum ZIP nesting/depth if parser follows nested archives
```

Security evidence establishes that upload size and archive expansion must be bounded; it does **not** establish MetalDocs' correct business ceiling.

The current repository authority pack contains no named representative binary corpus. Therefore these numbers remain **Unknown**, and `DraftUploadAllocation.max_bytes` remains unresolved until measurement evidence exists.

Multipart upload remains absent unless measured accepted size proves a real need.

---

# 10. 78-operation executable ledger

`JSON` means 64-KiB application/json body + `J`. `PAGED` means only `cursor`,`limit` unless stated.

## 10.1 Session / Organization / Authorization / Document Governance — 1→43

|#|operationId|Method + path|Request/profile|Success|Headers|Query/order|Problems|
|---:|---|---|---|---|---|---|---|
|1|`getSession`|`GET /api/v1/session`|`SAFE_READ`|`200 SessionView`|`JSON_NO_STORE`|none|`B`|
|2|`endSession`|`DELETE /api/v1/session`|no body / `UNSAFE_CSRF`|`204`|`SESSION_END`|none|`C`|
|3|`searchProviderSubjects`|`GET /api/v1/authentication/provider-subjects`|`SAFE_READ`|`200 ProviderSubjectSearchView`|`JSON_NO_STORE`|required query; provider order|`A`|
|4|`getCompany`|`GET /api/v1/company`|`SAFE_READ`|`200 CompanyView`|`JSON_ETAG`|none|`A`|
|5|`replaceCompany`|`PUT /api/v1/company`|`ReplaceCompanyRequest` JSON / `IF_MATCH_MUTATION`|`200 CompanyView`|`JSON_ETAG_MUTATION`|none|`U + J + P`|
|6|`listUsers`|`GET /api/v1/users`|`SAFE_READ`|`200 UserPage`|`JSON_NO_STORE`|`PAGED`; user_id ASC|`A`|
|7|`createUser`|`POST /api/v1/users`|`CreateUserRequest` JSON / `IDEMPOTENT_CREATE`|`201 CreateUserResult`|`JSON_NO_STORE`|none|`U + J + I + S`|
|8|`getUser`|`GET /api/v1/users/{user_id}`|`SAFE_READ`|`200 UserView`|`JSON_NO_STORE`|none|`A + N`|
|9|`getUserProfile`|`GET /api/v1/users/{user_id}/profile`|`SAFE_READ`|`200 UserProfileView`|`JSON_ETAG`|none|`A + N`|
|10|`replaceUserProfile`|`PUT /api/v1/users/{user_id}/profile`|`ReplaceUserProfileRequest` JSON / `PROFILE_REPLACE`|`200` replacement or `201` recreation, `UserProfileView`|`JSON_ETAG_MUTATION`|none|`U + J + N + P`|
|11|`deleteUserProfile`|`DELETE /api/v1/users/{user_id}/profile`|no body / `UNSAFE_CSRF`|`204`, including absent repeat|`NO_STORE`|none|`U`|
|12|`getUserProviderBinding`|`GET /api/v1/users/{user_id}/provider-binding`|`SAFE_READ`|`200 UserProviderBindingView`|`JSON_ETAG`|none|`A + N`|
|13|`replaceUserProviderBinding`|`PUT /api/v1/users/{user_id}/provider-binding`|`ReplaceUserProviderBindingRequest` JSON / `IF_MATCH_MUTATION`|`200 UserProviderBindingView`|`JSON_ETAG_MUTATION`|none|`U + J + N + P + S`|
|14|`getUserEligibility`|`GET /api/v1/users/{user_id}/eligibility`|`SAFE_READ`|`200 UserEligibilityView`|`JSON_ETAG`|none|`A + N`|
|15|`replaceUserEligibility`|`PUT /api/v1/users/{user_id}/eligibility`|`ReplaceUserEligibilityRequest` JSON / `IF_MATCH_MUTATION`|`200 UserEligibilityView`, including exact no-op|`JSON_ETAG_MUTATION`|none|`U + J + N + P`|
|16|`listAreas`|`GET /api/v1/areas`|`SAFE_READ`|`200 AreaPage`|`JSON_NO_STORE`|`PAGED`; code ASC,area_id ASC|`A`|
|17|`createArea`|`POST /api/v1/areas`|`CreateAreaRequest` JSON / `IDEMPOTENT_CREATE`|`201 CreateAreaResult`|`JSON_NO_STORE`|none|`U + J + I + S`|
|18|`getArea`|`GET /api/v1/areas/{area_id}`|`SAFE_READ`|`200 AreaView`|`JSON_ETAG`|none|`A + N`|
|19|`replaceArea`|`PUT /api/v1/areas/{area_id}`|`ReplaceAreaRequest` JSON / `IF_MATCH_MUTATION`|`200 AreaView`|`JSON_ETAG_MUTATION`|none|`U + J + N + P`|
|20|`getAreaLifecycle`|`GET /api/v1/areas/{area_id}/lifecycle`|`SAFE_READ`|`200 AreaLifecycleView`|`JSON_ETAG`|none|`A + N`|
|21|`replaceAreaLifecycle`|`PUT /api/v1/areas/{area_id}/lifecycle`|`ReplaceAreaLifecycleRequest` JSON / `IF_MATCH_MUTATION`|`200 AreaLifecycleView`|`JSON_ETAG_MUTATION`|none|`U + J + N + P`|
|22|`listGroups`|`GET /api/v1/groups`|`SAFE_READ`|`200 GroupPage`|`JSON_NO_STORE`|`PAGED`; group_id ASC|`A`|
|23|`createGroup`|`POST /api/v1/groups`|`CreateGroupRequest` JSON / `IDEMPOTENT_CREATE`|`201 CreateGroupResult`|`JSON_NO_STORE`|none|`U + J + I`|
|24|`getGroup`|`GET /api/v1/groups/{group_id}`|`SAFE_READ`|`200 GroupView`|`JSON_ETAG`|none|`A + N`|
|25|`replaceGroup`|`PUT /api/v1/groups/{group_id}`|`ReplaceGroupRequest` JSON / `IF_MATCH_MUTATION`|`200 GroupView`|`JSON_ETAG_MUTATION`|none|`U + J + N + P`|
|26|`deleteGroup`|`DELETE /api/v1/groups/{group_id}`|no body / `UNSAFE_CSRF`|`204`, including absent repeat|`NO_STORE`|none|`U + S`|
|27|`listGroupMembers`|`GET /api/v1/groups/{group_id}/members`|`SAFE_READ`|`200 GroupMemberPage`|`JSON_NO_STORE`|`PAGED`; user_id ASC|`A + N`|
|28|`addGroupMember`|`PUT /api/v1/groups/{group_id}/members/{user_id}`|no body / `UNSAFE_CSRF`|`201` first; `204` exact repeat|`NO_STORE`|none|`U + N + S`|
|29|`removeGroupMember`|`DELETE /api/v1/groups/{group_id}/members/{user_id}`|no body / `UNSAFE_CSRF`|`204`, including absent repeat|`NO_STORE`|none|`U`|
|30|`listRoles`|`GET /api/v1/roles`|`SAFE_READ`|`200 RoleListView`|`JSON_NO_STORE`|fixed T3 role order|`A`|
|31|`listRoleAssignments`|`GET /api/v1/role-assignments`|`SAFE_READ`|`200 RoleAssignmentPage`|`JSON_NO_STORE`|`PAGED`; assignment_id ASC|`A`|
|32|`createRoleAssignment`|`POST /api/v1/role-assignments`|`CreateRoleAssignmentRequest` JSON / `IDEMPOTENT_CREATE`|`201 CreateRoleAssignmentResult`|`JSON_NO_STORE`|none|`U + J + I + S`|
|33|`deleteRoleAssignment`|`DELETE /api/v1/role-assignments/{assignment_id}`|no body / `UNSAFE_CSRF`|`204`, including absent repeat|`NO_STORE`|none|`U`|
|34|`listDocumentTypes`|`GET /api/v1/document-types`|`SAFE_READ`|`200 DocumentTypePage`|`JSON_NO_STORE`|`PAGED`; document_type_id ASC|`A`|
|35|`createDocumentType`|`POST /api/v1/document-types`|`CreateDocumentTypeRequest` JSON / `IDEMPOTENT_CREATE`|`201 CreateDocumentTypeResult`|`JSON_NO_STORE`|none|`U + J + I + S`|
|36|`getDocumentType`|`GET /api/v1/document-types/{document_type_id}`|`SAFE_READ`|`200 DocumentTypeView`|`JSON_ETAG`|none|`A + N`|
|37|`replaceDocumentType`|`PUT /api/v1/document-types/{document_type_id}`|`ReplaceDocumentTypeRequest` JSON / `IF_MATCH_MUTATION`|`200 DocumentTypeView`|`JSON_ETAG_MUTATION`|none|`U + J + N + P + S`|
|38|`getDocumentTypeGovernance`|`GET /api/v1/document-types/{document_type_id}/governance`|`SAFE_READ`|`200 DocumentTypeGovernanceView`|`JSON_ETAG`|none|`A + N`|
|39|`replaceDocumentTypeGovernance`|`PUT /api/v1/document-types/{document_type_id}/governance`|`ReplaceDocumentTypeGovernanceRequest` JSON / `IF_MATCH_MUTATION`|`200 DocumentTypeGovernanceView`|`JSON_ETAG_MUTATION`|none|`U + J + N + P + S`|
|40|`getDocumentTypeEligibleTemplates`|`GET /api/v1/document-types/{document_type_id}/eligible-templates`|`SAFE_READ`|`200 EligibleTemplatesView`|`JSON_ETAG`|document.code ASC,id ASC|`A + N`|
|41|`replaceDocumentTypeEligibleTemplates`|`PUT /api/v1/document-types/{document_type_id}/eligible-templates`|`ReplaceEligibleTemplatesRequest` JSON / `IF_MATCH_MUTATION`|`200 EligibleTemplatesView`|`JSON_ETAG_MUTATION`|none|`U + J + N + P + S`|
|42|`getDocumentTypeNumberingPreview`|`GET /api/v1/document-types/{document_type_id}/numbering-preview`|`SAFE_READ`|`200 NumberingPreviewView`|`JSON_NO_STORE`|optional area_id|`A + N + validation.failed`|
|43|`listTemplateConfigurations`|`GET /api/v1/document-governance/templates`|`SAFE_READ`|`200 TemplateConfigurationPage`|`JSON_NO_STORE`|`PAGED`; document.code ASC,id ASC|`A`|

Rows 35/38/39 are wire-defined but promotion-blocked by §8 until Step labels become persistence-complete.

## 10.2 Controlled Documents / Work — 44→77

|#|operationId|Method + path|Request/profile|Success|Headers|Query/order|Problems|
|---:|---|---|---|---|---|---|---|
|44|`getDocumentCreationOptions`|`GET /api/v1/document-creation/options`|`SAFE_READ`|`200 DocumentCreationOptionsView`|`JSON_NO_STORE`|optional area_id,document_type_id; arrays code/id,candidates user_id|`A + validation.failed`|
|45|`listDocuments`|`GET /api/v1/documents`|`SAFE_READ`|`200 DocumentPage`|`JSON_NO_STORE`|q,document_type_id,area_id,responsible_owner_user_id,status,cursor,limit; §1.9 ranking|`A`|
|46|`createDocument`|`POST /api/v1/documents`|`CreateDocumentRequest` JSON / `IDEMPOTENT_CREATE`|`201 CreateDocumentResult`|`JSON_NO_STORE`|none|`U + J + I + S`|
|47|`getDocument`|`GET /api/v1/documents/{document_id}`|`SAFE_READ`|`200 DocumentOfficialView`|`JSON_NO_STORE`|none|`B + N`|
|48|`getDocumentResponsibleOwner`|`GET /api/v1/documents/{document_id}/responsible-owner`|`SAFE_READ`|`200 ResponsibleOwnerView`|`JSON_ETAG`|none|`A + N`|
|49|`replaceDocumentResponsibleOwner`|`PUT /api/v1/documents/{document_id}/responsible-owner`|`ReplaceResponsibleOwnerRequest` JSON / `IF_MATCH_MUTATION`|`200 ResponsibleOwnerView`|`JSON_ETAG_MUTATION`|none|`U + J + N + P + S`|
|50|`getDocumentTemplateRole`|`GET /api/v1/documents/{document_id}/template-role`|`SAFE_READ`|`200 TemplateRoleView`|`JSON_ETAG`|none|`A + N`|
|51|`replaceDocumentTemplateRole`|`PUT /api/v1/documents/{document_id}/template-role`|`ReplaceTemplateRoleRequest` JSON / `IF_MATCH_MUTATION`|`200 TemplateRoleView`|`JSON_ETAG_MUTATION`|none|`U + J + N + P + S`|
|52|`createDocumentRevision`|`POST /api/v1/documents/{document_id}/revisions`|no body / `IDEMPOTENT_CREATE`|`201 CreateRevisionResult`|`JSON_NO_STORE`|none|`U + N + I + S`|
|53|`getDocumentHistory`|`GET /api/v1/documents/{document_id}/history`|`SAFE_READ`|`200 DocumentHistoryPage`|`JSON_NO_STORE`|`PAGED`; occurred_at ASC,kind,semantic id|`A + N`|
|54|`listAuthoringWork`|`GET /api/v1/work/authoring`|`SAFE_READ`|`200 WorkAuthoringPage`|`JSON_NO_STORE`|`PAGED`; document.code ASC,id ASC|`B`|
|55|`listGovernanceWork`|`GET /api/v1/work/governance`|`SAFE_READ`|`200 WorkGovernancePage`|`JSON_NO_STORE`|`PAGED`; governance_attempt_id ASC|`B`|
|56|`getRevision`|`GET /api/v1/revisions/{revision_id}`|`SAFE_READ`|`200 RevisionView`|`JSON_NO_STORE`|none|`A + N`|
|57|`getRevisionDraft`|`GET /api/v1/revisions/{revision_id}/draft`|`SAFE_READ`|`200 DocumentWorkView`|`JSON_ETAG`|none|`A + N`|
|58|`updateRevisionDraft`|`PATCH /api/v1/revisions/{revision_id}/draft`|`UpdateDraftRequest` JSON / `IF_MATCH_MUTATION`|`200 DocumentWorkView`|`JSON_ETAG_MUTATION`|none|`U + J + N + D + S + state.upload_expired`|
|59|`startRevisionDraftUpload`|`POST /api/v1/revisions/{revision_id}/draft/uploads`|no body / `UNSAFE_CSRF`|`201 DraftUploadAllocation`|`JSON_NO_STORE`|none|`U + N + S`|
|60|`completeRevisionDraftUpload`|`POST /api/v1/revisions/{revision_id}/draft/uploads/{upload_id}/complete`|no body / `UNSAFE_CSRF`|`204`, including READY repeat|`NO_STORE`|none|`U + N + S + state.upload_expired + validation.content_invalid`|
|61|`getRevisionDraftSource`|`GET /api/v1/revisions/{revision_id}/draft/source`|`SAFE_READ`|`200 exact bytes`|`EXACT_BYTES`|none|`A + N + X`|
|62|`createSubmission`|`POST /api/v1/revisions/{revision_id}/submissions`|no body / `SUBMISSION_CREATE`|`201 SubmissionCreateResult`|`JSON_NO_STORE`|none|`U + N + I + D + S + validation.failed + validation.content_malicious + dependency.malware_inspector_unavailable`|
|63|`getSubmission`|`GET /api/v1/submissions/{submission_id}`|`SAFE_READ`|`200 SubmissionView`|`JSON_NO_STORE`|none|`A + N`|
|64|`getSubmissionSource`|`GET /api/v1/submissions/{submission_id}/source`|`SAFE_READ`|`200 exact bytes`|`EXACT_BYTES`|none|`A + N + X`|
|65|`withdrawSubmission`|`PUT /api/v1/submissions/{submission_id}/withdrawal`|no body / `UNSAFE_CSRF`|`201 SubmissionWithdrawalView` first; `200` exact repeat|`JSON_NO_STORE`|none|`U + N + S`|
|66|`cancelRevision`|`PUT /api/v1/revisions/{revision_id}/cancellation`|`RevisionCancellationRequest` JSON / `UNSAFE_CSRF`|`201 RevisionCancellationView` first; `200` exact repeat|`JSON_NO_STORE`|none|`U + J + N + S`|
|67|`getGovernanceAttempt`|`GET /api/v1/governance-attempts/{attempt_id}`|`SAFE_READ`|`200 GovernanceCaseView`|`JSON_NO_STORE`|embedded first feedback page; ordered steps|`A + N`|
|68|`listGovernanceFeedback`|`GET /api/v1/governance-attempts/{attempt_id}/feedback`|`SAFE_READ`|`200 GovernanceFeedbackPage`|`JSON_NO_STORE`|`PAGED`; created_at ASC,id ASC|`A + N`|
|69|`createGovernanceFeedback`|`POST /api/v1/governance-attempts/{attempt_id}/feedback`|`CreateGovernanceFeedbackRequest` JSON / `IDEMPOTENT_CREATE`|`201 CreateGovernanceFeedbackResult`|`JSON_NO_STORE`|none|`U + J + N + I + S`|
|70|`getGovernanceStepDecision`|`GET /api/v1/governance-attempts/{attempt_id}/steps/{step_id}/decision`|`SAFE_READ`|`200 GovernanceDecisionView`|`JSON_NO_STORE`|none|`A + N`|
|71|`recordGovernanceStepDecision`|`PUT /api/v1/governance-attempts/{attempt_id}/steps/{step_id}/decision`|`GovernanceDecisionRequest` JSON / `UNSAFE_CSRF`|`201 GovernanceDecisionView` first; `200` exact repeat|`JSON_NO_STORE`|none|`U + J + N + S + state.governance_step_already_decided`|
|72|`getRelease`|`GET /api/v1/releases/{release_id}`|`SAFE_READ`|`200 ReleaseView`|`JSON_NO_STORE`|none|`A + N`|
|73|`getReleaseSource`|`GET /api/v1/releases/{release_id}/source`|`SAFE_READ`|`200 exact bytes`|`EXACT_BYTES`|none|`A + N + X`|
|74|`getOfficialRenditionContent`|`GET /api/v1/official-renditions/{rendition_id}/content`|`SAFE_READ`|`200 exact PDF bytes`|`EXACT_BYTES`|none|`A + N + X`|
|75|`createObsolescenceRequest`|`POST /api/v1/documents/{document_id}/obsolescence-requests`|`ObsolescenceRequestCreateRequest` JSON / `IDEMPOTENT_CREATE`|`201 ObsolescenceCreateResult`|`JSON_NO_STORE`|none|`U + J + N + I + S`|
|76|`getObsolescenceRequest`|`GET /api/v1/obsolescence-requests/{request_id}`|`SAFE_READ`|`200 ObsolescenceRequestView`|`JSON_NO_STORE`|none|`A + N`|
|77|`withdrawObsolescenceRequest`|`PUT /api/v1/obsolescence-requests/{request_id}/withdrawal`|no body / `UNSAFE_CSRF`|`201 ObsolescenceWithdrawalView` first; `200` exact repeat|`JSON_NO_STORE`|none|`U + N + S`|

Row 67's label-dependent Step projection is promotion-blocked by §8. No other Controlled Documents operation requires that correction.

## 10.3 Audit — 78

|#|operationId|Method + path|Request/profile|Success|Headers|Query/order|Problems|
|---:|---|---|---|---|---|---|---|
|78|`listAuditEvents`|`GET /api/v1/audit/events`|`SAFE_READ`|`200 AuditEventPage`|`JSON_NO_STORE`|`PAGED`; occurred_at DESC,event_id DESC|`A`|

## 10.4 Count proof

```text
Session/AuthN support                3
Organization                        26
Authorization                        4
Document Governance config          10
Controlled Documents / Work         34
Audit                                1
TOTAL                               78
```

---

# 11. Generation feasibility

Reproducible probe versions as of 2026-08-20:

```text
Go generator     oapi-codegen v2.8.0, strict-server
TS generator     openapi-typescript 7.13.0, paths/components only
```

These pins are **probe evidence**, not implementation dependency authorization.

Why this is the smallest pair:

```text
oapi-codegen
  typed per-operation request objects
  typed closed response-object sets
  oneOf/discriminator support
  customizable strict ResponseErrorHandlerFunc
  strict generation does not pretend to validate incoming requests

openapi-typescript
  directly generates requested paths/components boundary
  no generated runtime SDK requirement
  preserves required/optional/nullable distinctions
  supports oneOf-oriented modeling
  benefits from explicit additionalProperties closure
```

Executable disposable probe must prove:

```text
1. additionalProperties:false does not generate arbitrary object bags
2. required/optional/nullable remain distinguishable in Go + TS
3. every enum remains finite
4. unions avoid any/untyped escape
5. JS-visible integers remain within the safe integer contract
6. multiple success statuses generate a closed response set
7. per-operation Problem variants require no default response
8. strict-server unexpected errors route through canonical RFC9457 500 serializer
9. central request validation is a separate demonstrated control
10. 64-KiB OAS extension is enforced at request boundary
11. unknown JSON/query members and undeclared request bodies are rejected
12. one Go wire package + one TS paths/components boundary remain sole generated authorities
13. no generator/provider field enters public contract
```

The oapi-codegen tool's own current Go build requirement does not freeze the MetalDocs runtime Go version; generated-code compatibility is part of the probe.

---

# 12. Runtime contract-conformance proof

```text
request:
raw HTTP
-> session/trust + raw request limits
-> central OpenAPI request validation
-> generated typed request boundary
-> semantic handler

response:
semantic result
-> generated typed response boundary
-> HTTP
-> contract test validates exact status + headers + body + Problem variant
```

Required negative proof classes:

```text
missing/invalid session -> 401 Problem + challenge
logout clears the exact __Host cookie
undeclared JSON member rejected
undeclared query member rejected
body on bodyless operation rejected
JSON >65,536 bytes -> 413
compressed JSON -> 415 + Accept-Encoding identity
missing/malformed CSRF, Idempotency-Key, If-Match rejected
weak/list/wildcard If-Match rejected
wrong-domain/stale ETag returns exact 412 code
PROFILE_REPLACE rejects missing conditional and both conditionals together
undeclared enum/invalid oneOf rejected
ETag-protected response cannot change through an independently mutable embedded label
success cannot omit required member/header
operation cannot emit undeclared Problem.code
router unknown path/method returns closed 404/405 behavior
pagination cursor is operation/filter/order bound
ReplaySnapshot fixture remains <=2,048 bytes
exact-byte Range ->400 and response cannot become redirect/206/304/compressed/provider URL
Content-Digest equals SHA-256 of exact body
```

No generic production response-buffer validator is added. Generated typed output + contract tests remain the accepted minimum.

---

# 13. Structural Inversion / subtractive checkpoint

Current candidate survives the Lead pass with the corrections above:

```text
if legacy API shape were opposite:
  78 semantic operations still follow Product/T6
  ETag/idempotency/CSRF/pagination still follow accepted invariants
  exact-content wire still follows T4
  component-registry + operation-ledger still closes Writer decisions
```

Removed/not introduced:

```text
universal response envelope
generic action endpoint
generic filter/sort DSL
generic public JSON facts bag
provider/job state
persisted permission snapshot
editable role/policy engine
separate Approval API
multipart without measured need
Range/HEAD/304 baseline
arbitrary Problem extensions/default response
generator-specific product fields
dormant future capability
```

The one material structural defect exposed is the bounded T8-D Step-label omission in §8, not a reason to redesign lifecycle or governance.

---

# 14. Remaining closure prerequisites

```text
A. operator adjudication of bounded T8-D Step-label contradiction
B. representative DOCX/PDF corpus measurement + raw/expanded/ZIP limits
C. disposable pinned oapi-codegen + openapi-typescript generation/compile/type probe
D. exact contract fixtures proving 78-row status/header/Problem matrix
E. final whole-candidate Structural Inversion / YAGNI / overengineering / global-coherence pass
F. isolated final Fable branch only after A→E converge
G. Lead adjudication + explicit operator ratification
```

Until A→E converge:

```text
T8-E ACTIVE
T8-F NOT OPEN
implementation BLOCKED
```

No Fable review branch is created yet.