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

## 1.1 Closed objects

OpenAPI 3.0.3 defaults `additionalProperties` to `true`. Every fixed MetalDocs request, response, reference, page, union branch and Problem object therefore uses:

```text
additionalProperties: false
```

The only current deliberate map is `DraftUploadAllocation.required_headers`, because temporary direct upload must forward provider-required HTTP headers exactly while provider identity remains mechanism only.

There is no generic metadata/settings/facts/property-bag map in the application contract.

## 1.2 Presence and nullability

```text
required = member must be present
optional = member may be absent
nullable = explicit JSON null is a semantic value
```

OAS 3.0.3 `nullable: true` is used only for explicit null. Absence and null are not interchangeable. PATCH/PUT never acquires implicit “null means delete”. Request and response components are purpose-built rather than using `readOnly`/`writeOnly` to hide an overbroad shared DTO.

The baseline nullable member remains `Page.next_cursor`. All other nullability below is explicit where needed; otherwise absence means the fact does not exist in that projection.

## 1.3 Composition

Do not use `allOf` inheritance merely to reduce YAML. True closed semantic unions use `oneOf` with a required discriminator; ordinary reuse nests `$ref` components. A generator limitation may change encoding only when semantic exactness remains unchanged.

## 1.4 JSON vocabulary

- JSON member names: `snake_case`.
- Semantic enums described upstream in uppercase/PascalCase normalize to lower snake case.
- Canonical product identifiers keep their accepted spelling; T3 `PermissionCode` therefore remains dot-separated (`document.read_effective`, etc.).
- Unknown enum values are invalid; there is no `other`/future catch-all.

## 1.5 Aggregate JSON request limit

All `/api/v1` `application/json` request bodies have one raw-body ceiling:

```text
65,536 bytes
```

The eventual OAS carries one machine-readable extension on each JSON request body:

```text
x-metaldocs-max-request-body-bytes: 65536
```

Central request validation enforces the raw limit before JSON decoding. Request compression is not a Launch capability: a JSON request with a non-identity `Content-Encoding` is rejected as unsupported media.

Why one ceiling instead of many guessed limits:

- OWASP requires bounded incoming payloads/parameters for API resource-consumption safety;
- no application JSON command carries document bytes;
- the largest current semantic commands are bounded configuration/UUID arrays and route steps;
- 64 KiB leaves large operational headroom while preventing an effectively unbounded parser/memory surface;
- increasing this ceiling later does not change semantic ownership or persistence meaning.

Reopen trigger: a measured legitimate Launch command cannot be represented within 64 KiB.

Direct document bytes are outside this JSON ceiling and remain blocked on the separate measured corpus obligation in §9.

## 1.6 Query/opaque transport bounds

```text
SearchQuery / provider directory query  maxLength 256
OpaqueCursor                            maxLength 2048
CsrfToken                               maxLength 512
ProviderSubjectRef                      maxLength 2048
CodeToken                               maxLength 32
```

`CodeToken` is uppercase ASCII alphanumeric only; `-` remains product-owned separator and is forbidden inside Area/DocumentType code tokens.

The 32-character code ceiling is a T8-E contract choice, not inherited implementation. It keeps human business codes bounded, permits materially longer values than current examples, and is additively reopenable if a real code vocabulary exceeds it.

## 1.7 Exact-byte delivery

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

A `Range` request is unsupported at Launch and fails `400 request.invalid`; HEAD is undeclared and follows the 405 router law. Exact-byte missing/corrupt semantic content fails `500 internal.content_integrity`; temporary storage/dependency failure uses `503 dependency.unavailable`.

---

# 2. Request / success header profiles

## 2.1 Request profiles

```text
SAFE_READ
  authenticated request; no application mutation header

UNSAFE_CSRF
  X-CSRF-Token required

IDEMPOTENT_CREATE
  X-CSRF-Token required
  Idempotency-Key required; client-generated UUID

IF_MATCH_MUTATION
  X-CSRF-Token required
  If-Match required
  exactly one strong entity-tag
  `*`, weak tags and lists forbidden

SUBMISSION_CREATE
  X-CSRF-Token required
  Idempotency-Key required; client-generated UUID
  If-Match required for exact DRAFT generation

PROFILE_REPLACE
  X-CSRF-Token required
  existing profile -> If-Match required
  absent profile recreation -> If-None-Match exactly `*` required
  both conditional headers together -> invalid request
```

Every unsafe `/api/v1` operation uses one unsafe profile even when bodyless. Missing/malformed required conditional or idempotency headers are `request.invalid`; a syntactically valid stale precondition is 412.

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

EXACT_BYTES
  §1.7 exact-byte header set
```

No baseline `Location`, replay-indicator, permission snapshot, provider ID, or generic metadata response header exists.

Problem responses use `Content-Type: application/problem+json` and `Cache-Control: no-store`. `401` additionally requires:

```text
WWW-Authenticate: MetalDocsSession
```

`429` requires `Retry-After` as non-negative delta-seconds. `405` requires the exact `Allow` set for that path.

---

# 3. Shared component registry

All objects below are closed.

## 3.1 Scalars

```text
Uuid
  string; format=uuid

UtcInstant
  string; format=date-time
  server serialization = RFC3339 UTC `Z`

OpaqueCursor
  nonblank string; maxLength=2048

IdempotencyKey
  UUID

CsrfToken
  nonblank string; maxLength=512

ProviderSubjectRef
  nonblank opaque string; maxLength=2048

Sha256Hex
  lowercase hex; pattern ^[0-9a-f]{64}$

NonBlankString
  string whose trimmed semantic value is nonblank

SearchQuery
  NonBlankString; maxLength=256

CodeToken
  pattern ^[A-Z0-9]+$; maxLength=32

DocumentCode
  pattern ^[A-Z0-9]+(?:-[A-Z0-9]+)?-[0-9]{3,}$
  maximum derived from two 32-char code tokens + separators + BIGINT decimal counter = 85 chars

EmailAddress
  string; format=email

RevisionOrdinal
  integer; format=int64; minimum=0

ByteCount
  integer; format=int64; minimum=0
```

Other human text remains bounded by the 64-KiB aggregate request ceiling rather than inventing unrelated per-field maxima. A future requirement may narrow a particular semantic field without changing this law.

## 3.2 References

```text
UserReference
  required: user_id
  optional: display_name
  display_name absence = erasable UserProfile enrichment absent
  never contains email/provider/grant data

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

A Submission never reuses a live `RevisionReference` as historical title authority; it carries its own frozen `title` snapshot beside `RevisionIdentity`.

## 3.3 Closed enums

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

`DocumentType.active` is a boolean, matching the ratified T8-D current truth; no redundant `active|inactive` wire enum is introduced.

No public upload lifecycle enum exists. OPEN/READY/GC_PENDING are mechanism states, not product wire state.

## 3.4 True unions

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

The `GovernanceRouteStep.label` field is Product-owned and currently exposes the bounded T8-D contradiction in §8. It is **not** removed to make persistence convenient.

## 3.5 Pagination

```text
Page
  required: next_cursor, has_more
  next_cursor: OpaqueCursor | null
  has_more: boolean
```

Potentially unbounded lists return `{items,page}`. Query `limit` is integer `1..100`, default `20`; `cursor` is opaque. No offset, total count, generic sort, frozen snapshot, or server cursor state.

Cursor integrity binds canonical operationId + normalized filters + deterministic ordering. Current Authorization is rechecked every page.

---

# 4. RFC 9457 Problem catalog

## 4.1 Exact base law

Every Problem variant is a full closed object, not an open base inherited with `allOf`.

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

`instance` is a fresh `urn:uuid:<uuid>` for that Problem occurrence. `trace_id` is opaque/nonblank correlation text; clients do not parse its provider format.

Only `request.invalid` and validation-family variants may contain optional non-empty `errors[]`:

```text
ProblemError
  required: pointer, detail
  pointer = RFC 6901 pointer rooted at /path, /query, /header, or /body
  rejected sensitive values are never echoed
```

Each variant freezes `type`, `title`, `status`, and `code` using single-value enums where OAS 3.0.3 lacks `const`.

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

No module/provider/storage/database/scanner error string escapes.

## 4.3 Ledger shorthand — exact expansion only

To keep the candidate ledger readable, the following are **textual exact-set macros**. They do not exist as runtime inheritance and the final OAS expands every operation to concrete Problem response variants.

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

`ratelimit.exceeded` is allowed cross-cutting because 429 is already in the accepted T8-E baseline and resource-consumption protection is transport safety, not business authority. T8-E freezes only the response shape; any threshold/policy is not invented here.

`request.method_not_allowed` is router-level: an undeclared method on a declared path returns its 405 Problem + exact `Allow`; it is not attached to one of the 78 declared operations.

---

# 5. Components — Session / Organization / Access / Document Governance

## 5.1 Session / AuthN

```text
SessionView
  required: user, csrf_token
  user: UserReference

ProviderSubjectOption
  required: provider_subject_ref, display_hints
  display_hints: string[]

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
  law: successful create establishes ENABLED User + required profile + binding atomically

CreateUserResult
  required: user_id

UserView
  required: user, eligibility
  law: aggregate read only; eligibility subresource remains canonical ETag source

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
  law: new Area starts `active`; creating an already-retired Area is not a Launch capability

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
  governance: GovernancePolicy
  representation: RepresentationPolicy
  law: eligible-template set starts empty

CreateDocumentTypeResult
  required: document_type_id

ReplaceDocumentTypeRequest
  required: code, name, numbering_scope, active
  code/numbering_scope change after first committed Document -> state.conflict

DocumentTypeGovernanceView
  required: governance, representation

ReplaceDocumentTypeGovernanceRequest
  required: governance, representation

EligibleTemplateItem
  required: document
  optional: current_effective_title

EligibleTemplatesView
  required: templates
  templates ordered by document code, id

ReplaceEligibleTemplatesRequest
  required: template_document_ids
  unique UUID array; empty array valid

NumberingPreviewView
  required: preview_code, reservation
  reservation is constant false

TemplateConfigurationItem
  required: document, template_role, has_effective_revision, eligible_document_type_ids
  optional: current_effective_title

TemplateConfigurationPage
  required: items, page
```

Creating a DocumentType carries its initial governance/representation explicitly because T8-D requires those current values and no accepted default exists. This avoids a hidden Writer-selected `NoHumanApproval`/`SourceOnly` default while retaining the independent later replacement resources/ETags.

---

# 6. Components — Controlled Documents / Work

## 6.1 Creation / official projections

```text
TemplateCreationOption
  required: document, effective_revision
  effective_revision: RevisionReference

DocumentCreationOptionsView
  required:
    areas
    document_types
    templates
    default_responsible_owner
  optional:
    responsible_owner_candidates
  law:
    absence of responsible_owner_candidates = caller lacks owner-manage selection capability
    present empty array = capability exists but no alternate eligible target

CreateDocumentRequest
  required: document_type_id, area_id, title
  optional: template_document_id, responsible_owner_user_id
  absence of template_document_id = trusted blank DOCX seed

CreateDocumentResult
  required: document_id, revision_id
  deliberately excludes code/title/free text from durable replay result

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
  law:
    official present for current/last released truth (including obsolete)
    absent before any Release
    an older EFFECTIVE revision keeps status effective even if a newer open revision is cancelled
```

Library `status` query vocabulary is only `effective|obsolete|cancelled`, default `effective`. DRAFT/SUBMITTED remain Work lenses.

## 6.2 Current relationships / work

```text
ResponsibleOwnerView
  required: document_id, responsible_owner

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
  current_submission_id present only while state=submitted

DocumentWorkView
  required: document, revision, title, content, updated_at
  generation is deliberately absent: opaque strong ETag is the wire concurrency token

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

Upload allocation contains no provider account/bucket/key/version/ETag. `required_headers` contains only exact headers needed for provider create-only PUT.

Upload completion is bodyless and returns no client-authored/returned authoritative descriptor. The server independently derives descriptor; subsequent DRAFT PATCH references only `upload_id`.

## 6.3 Submission

```text
SubmissionCreateResult
  governance_pending ->
    { state:governance_pending, submission_id, governance_attempt_id }
  rendition_pending ->
    { state:rendition_pending, submission_id }
  released ->
    { state:released, submission_id, release_id }

SubmissionHumanGate
  required: required, satisfied

SubmissionRepresentationGate
  required: required, satisfied, attention_required
  attention_required is only a derived terminal-renderer attention hint; never job/lifecycle state

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

`title` is the immutable Submission title snapshot. No renderer/River/provider job identity or state appears.

```text
SubmissionWithdrawalView
  required: submission_id, actor, withdrawn_at

RevisionCancellationRequest
  required: reason

RevisionCancellationView
  required: revision_id, actor, reason, cancelled_at
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
  steps: ordered by ordinal
  feedback: first page (20) in created_at ASC, feedback_id ASC
  feedback.page.next_cursor, when present, is a continuation cursor for listGovernanceFeedback
  allowed_actions: unique GovernanceCaseAction[]; may be empty
```

`allowed_actions` vocabulary is exactly:

```text
accept
return_for_changes
add_feedback
```

It is computed from the same current T3 decisions + Controlled Documents facts used by command authorization. Every command rechecks truth.

## 6.5 History / work lists

`DocumentHistoryItem` is a closed union. Route `document_id` supplies stable Document context.

```text
revision_created
  revision, title, occurred_at

submission_created
  submission_id, revision, title, submitter, occurred_at
  optional governance_attempt_id

governance_decision
  decision_id, governance_attempt_id, step_id, actor, outcome, occurred_at
  optional reason only for return_for_changes

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
      optional predecessor_revision_id, representation:{kind:source_only, source:ContentSummary} }

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

ObsolescenceRequestView
  required: request_id, document, target_revision, initiator, reason, state, requested_at
  optional: ended_at, governance_attempt_id

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
  group_membership
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

Audit does not expose its JSONB storage shape as an open object. Public facts are a closed union:

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

Audit event variants constrain `operation_code` to the matching facts family; invalid code/facts combinations are not schema-valid. Config/user/group/provider events whose resource identity + operation code are sufficient use `none`; Audit never copies free-form governed reasons/content/PII into the wire merely because storage has bounded facts.

```text
AuditEventView
  required: event_id, occurred_at, actor, operation_code,
            resource_kind, resource_id, visibility, facts

AuditEventPage
  required: items, page
  order: occurred_at DESC, event_id DESC
```

No Audit filter is added by inference. Launch `GET /audit/events` accepts only cursor/limit; current `audit.read` scope filtering occurs before pagination per T3/T8-C. A named auditor filter requirement can be added later without a new owner or route family.

---

# 8. Material contradiction — bounded T8-D reopen prerequisite

## 8.1 Evidence

Product Contract/T6 makes Governance Route Step a product concept with human-facing labels such as:

```text
Revisão técnica
Gestor
Qualidade
```

The executable wire therefore correctly needs `GovernanceRouteStep.label`, and Governance Case must preserve the exact frozen label shown for each Step.

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

Neither shape preserves the product Step label. Deriving a historical case from the **current** DocumentType label would allow a later configuration rename to rewrite the label of an earlier frozen GovernanceAttempt.

That contradicts the already-ratified rule that each attempt freezes one coherent route snapshot.

## 8.2 Method classification

```text
KNOWN          Product requires human-facing Step label meaning
KNOWN          GovernanceAttempt freezes route snapshot
KNOWN          T8-D current + snapshot persistence omit label
MATERIAL       yes: historical governed context + persistent meaning
CAUSE          bounded T8-D completeness omission exposed by executable wire
NOT A REASON   preference, legacy shape, generator convenience
OUTCOME        STOP / SPLIT PREREQUISITE for label-dependent promotion only
```

Unrelated T8-E closure continues.

## 8.3 Smallest correction if operator ratifies the bounded reopen

Exactly two persistence additions; no new table/owner/lifecycle/API operation:

```text
controlled_docs.document_type_governance_steps
  + label TEXT NOT NULL

controlled_docs.governance_attempt_steps
  + label_snapshot TEXT NOT NULL
```

Attempt creation copies the exact current configured label into `label_snapshot` in the same coherent route-snapshot transition. Frozen snapshot label is immutable thereafter.

No candidate-user list, Role semantics, generic workflow metadata, localization framework, or versioned Policy aggregate is added.

T8-E may specify the intended wire now, but **T8-E cannot be ratified/promoted until this contradiction is explicitly adjudicated and the affected T8-D authority is reconciled**.

---

# 9. Document admission limits — measured evidence prerequisite

Do not guess raw DOCX/PDF ceilings from internet defaults.

Before T8-E close, measure a representative MetalDocs/controlled-document corpus and tooling behavior, then freeze:

```text
maximum raw document bytes
maximum structurally expanded DOCX bytes
maximum ZIP entry count
maximum ZIP nesting/depth if the chosen structural parser follows nested archives
```

Security evidence already establishes the direction: uploads must have size limits, actual content type/signature must be validated, archive expansion must be bounded, and ZIP/XML bomb behavior must fail closed. It does **not** establish MetalDocs' correct business ceiling.

The current repository authority pack contains no named representative binary corpus. Therefore these numbers remain **Unknown**, and `DraftUploadAllocation.max_bytes` remains unresolved until the measurement evidence exists.

Multipart upload remains absent unless the measured accepted raw size proves a real need.

---

# 10. 78-operation executable ledger

## 10.1 Session / Organization / Authorization / Document Governance — 1→43

`JSON` in Request means 64-KiB JSON body + `J` problem additions. `PAGED` means only `cursor`,`limit` unless stated.

| # | operationId | Method + path | Request/profile | Success | Headers | Query/order | Problems |
|---:|---|---|---|---|---|---|---|
|1|`getSession`|`GET /api/v1/session`|`SAFE_READ`|`200 SessionView`|`JSON_NO_STORE`|none|`B`|
|2|`endSession`|`DELETE /api/v1/session`|no body / `UNSAFE_CSRF`|`204`|`NO_STORE`|none|`C`|
|3|`searchProviderSubjects`|`GET /api/v1/authentication/provider-subjects`|`SAFE_READ`|`200 ProviderSubjectSearchView`|`JSON_NO_STORE`|required `query`; provider order|`A`|
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
|16|`listAreas`|`GET /api/v1/areas`|`SAFE_READ`|`200 AreaPage`|`JSON_NO_STORE`|`PAGED`; code ASC, area_id ASC|`A`|
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
|28|`addGroupMember`|`PUT /api/v1/groups/{group_id}/members/{user_id}`|no body / `UNSAFE_CSRF`|`201` first creation; `204` exact repeat|`NO_STORE`|none|`U + N + S`|
|29|`removeGroupMember`|`DELETE /api/v1/groups/{group_id}/members/{user_id}`|no body / `UNSAFE_CSRF`|`204`, including absent repeat|`NO_STORE`|none|`U`|
|30|`listRoles`|`GET /api/v1/roles`|`SAFE_READ`|`200 RoleListView`|`JSON_NO_STORE`|fixed T3 role order; not paged|`A`|
|31|`listRoleAssignments`|`GET /api/v1/role-assignments`|`SAFE_READ`|`200 RoleAssignmentPage`|`JSON_NO_STORE`|`PAGED`; assignment_id ASC|`A`|
|32|`createRoleAssignment`|`POST /api/v1/role-assignments`|`CreateRoleAssignmentRequest` JSON / `IDEMPOTENT_CREATE`|`201 CreateRoleAssignmentResult`|`JSON_NO_STORE`|none|`U + J + I + S`|
|33|`deleteRoleAssignment`|`DELETE /api/v1/role-assignments/{assignment_id}`|no body / `UNSAFE_CSRF`|`204`, including absent repeat|`NO_STORE`|none|`U`|
|34|`listDocumentTypes`|`GET /api/v1/document-types`|`SAFE_READ`|`200 DocumentTypePage`|`JSON_NO_STORE`|`PAGED`; code ASC, id ASC|`A`|
|35|`createDocumentType`|`POST /api/v1/document-types`|`CreateDocumentTypeRequest` JSON / `IDEMPOTENT_CREATE`|`201 CreateDocumentTypeResult`|`JSON_NO_STORE`|none|`U + J + I + S`|
|36|`getDocumentType`|`GET /api/v1/document-types/{document_type_id}`|`SAFE_READ`|`200 DocumentTypeView`|`JSON_ETAG`|none|`A + N`|
|37|`replaceDocumentType`|`PUT /api/v1/document-types/{document_type_id}`|`ReplaceDocumentTypeRequest` JSON / `IF_MATCH_MUTATION`|`200 DocumentTypeView`|`JSON_ETAG_MUTATION`|none|`U + J + N + P + S`|
|38|`getDocumentTypeGovernance`|`GET /api/v1/document-types/{document_type_id}/governance`|`SAFE_READ`|`200 DocumentTypeGovernanceView`|`JSON_ETAG`|none|`A + N`|
|39|`replaceDocumentTypeGovernance`|`PUT /api/v1/document-types/{document_type_id}/governance`|`ReplaceDocumentTypeGovernanceRequest` JSON / `IF_MATCH_MUTATION`|`200 DocumentTypeGovernanceView`|`JSON_ETAG_MUTATION`|none|`U + J + N + P + S`|
|40|`getDocumentTypeEligibleTemplates`|`GET /api/v1/document-types/{document_type_id}/eligible-templates`|`SAFE_READ`|`200 EligibleTemplatesView`|`JSON_ETAG`|code ASC,id ASC|`A + N`|
|41|`replaceDocumentTypeEligibleTemplates`|`PUT /api/v1/document-types/{document_type_id}/eligible-templates`|`ReplaceEligibleTemplatesRequest` JSON / `IF_MATCH_MUTATION`|`200 EligibleTemplatesView`|`JSON_ETAG_MUTATION`|none|`U + J + N + P + S`|
|42|`getDocumentTypeNumberingPreview`|`GET /api/v1/document-types/{document_type_id}/numbering-preview`|`SAFE_READ`|`200 NumberingPreviewView`|`JSON_NO_STORE`|optional area_id|`A + N + validation.failed`|
|43|`listTemplateConfigurations`|`GET /api/v1/document-governance/templates`|`SAFE_READ`|`200 TemplateConfigurationPage`|`JSON_NO_STORE`|`PAGED`; document.code ASC,id ASC|`A`|

Rows 35/38/39 are wire-defined but promotion-blocked by §8 until Step labels become persistence-complete.

## 10.2 Controlled Documents / Work — 44→77

| # | operationId | Method + path | Request/profile | Success | Headers | Query/order | Problems |
|---:|---|---|---|---|---|---|---|
|44|`getDocumentCreationOptions`|`GET /api/v1/document-creation/options`|`SAFE_READ`|`200 DocumentCreationOptionsView`|`JSON_NO_STORE`|optional area_id,document_type_id; arrays code/id, candidates user_id|`A + validation.failed`|
|45|`listDocuments`|`GET /api/v1/documents`|`SAFE_READ`|`200 DocumentPage`|`JSON_NO_STORE`|q,document_type_id,area_id,responsible_owner_user_id,status,cursor,limit; accepted T6 ranking|`B + permission.denied` only when requested catalog mode itself is forbidden|
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

Row 67's label-dependent Step projection is promotion-blocked by §8. No other Controlled Documents operation requires the bounded T8-D correction.

## 10.3 Audit — 78

| # | operationId | Method + path | Request/profile | Success | Headers | Query/order | Problems |
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

Feasibility needs concrete probe generators, not a generic “OpenAPI can generate types” claim.

Current probe pair:

```text
Go          oapi-codegen strict-server generation
TypeScript  openapi-typescript paths/components generation
```

Why this is the smallest evidence-backed pair:

```text
oapi-codegen
  typed per-operation request objects
  typed closed response-object sets
  oneOf/discriminator support
  customizable strict ResponseErrorHandlerFunc
  does not pretend strict generation itself validates incoming requests

openapi-typescript
  directly generates the requested paths/components boundary
  no generated runtime SDK requirement
  preserves required/optional/nullable distinctions
  supports oneOf-oriented discriminated modeling
  benefits from explicit additionalProperties closure
```

A larger generated SDK/runtime is not selected merely because it also emits types.

Executable disposable probe must prove:

```text
1. additionalProperties:false does not generate arbitrary object bags
2. required/optional/nullable remain distinguishable in Go + TS
3. every enum remains finite
4. RoleAssignment/Governance/Representation/Submission/Audit unions avoid any/untyped escape
5. multiple declared success statuses generate a closed response set
6. per-operation Problem variants require no default response
7. strict-server unexpected errors route through canonical RFC9457 500 serializer
8. central OpenAPI request validation is a separate demonstrated control
9. 64-KiB OAS extension is enforced by the central request boundary
10. unknown JSON/query members and undeclared request bodies are rejected
11. one Go wire package + one TS paths/components boundary remain the only generated authorities
12. no generator/provider field leaks into public contract
```

Failure of a generator may adjust schema encoding only; it cannot widen Product semantics or add operation 79.

The probe is architectural/tooling evidence, not Product implementation.

---

# 12. Runtime contract-conformance proof

Freeze the proof obligation, not runtime realization:

```text
request:
raw HTTP
→ raw request/body/header limit checks
→ central OpenAPI request validation
→ generated typed request boundary
→ application semantic handler

response:
semantic result
→ generated typed response boundary
→ HTTP
→ contract test validates exact status + headers + body + Problem variant
```

Required negative proof classes:

```text
undeclared JSON member rejected
undeclared query member rejected
body supplied to bodyless operation rejected
JSON > 65,536 bytes rejected with 413
compressed JSON request rejected
missing/malformed CSRF, Idempotency-Key, If-Match rejected
weak/list/wildcard If-Match rejected
PROFILE_REPLACE rejects missing conditional and both conditionals together
stale valid ETag returns the exact 412 code
undeclared enum value rejected
invalid oneOf branch/field combination rejected
success cannot omit required member/header
operation cannot emit undeclared Problem.code
router unknown method returns closed 405 + Allow
pagination cursor is operation/filter/order bound
exact-byte Range fails 400 and response cannot become redirect/206/304/compressed/provider URL
Content-Digest bytes equal SHA-256 of exact response body
```

No generic production response-buffer validator is added. Generated typed output + contract tests remain the accepted minimum.

---

# 13. Structural Inversion / subtractive checkpoint

Current candidate survives the first Lead pass:

```text
if legacy API shape were the opposite:
  78 semantic operations still follow Product/T6
  ETag/idempotency/CSRF/pagination still follow accepted invariants
  exact-content wire still follows T4
  component-registry + operation-ledger shape still closes Writer decisions

removed/not introduced:
  universal response envelope
  generic action endpoint
  generic filter/sort DSL
  generic metadata/facts bag
  provider/job state
  persisted permission snapshot
  editable role/policy engine
  separate Approval API
  multipart upload without measured need
  Range/HEAD/304 baseline
  arbitrary Problem extensions/default response
  dormant future capability
```

The one newly exposed structural defect is **not** a T8-E abstraction problem; it is the bounded T8-D Step-label omission in §8.

---

# 14. Remaining closure prerequisites

```text
A. operator adjudication of the bounded T8-D Step-label contradiction
B. representative DOCX/PDF corpus measurement and exact raw/expanded/ZIP limits
C. disposable oapi-codegen + openapi-typescript generation/compile/type probe
D. exact contract fixtures proving the 78-row Problem/status/header matrix
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