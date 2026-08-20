---
id: t8e-executable-wire-contract-proposal
kind: work
owner: architecture
summary: Active T8-E proposal for the smallest exact 78-operation executable application wire.
---

# T8-E executable wire contract — active proposal

> Temporary / non-authoritative. Implementation remains blocked.

Accepted baseline remains in `../../reference/t8e-checkpoint.md`; Product/API meaning remains in `../../product/journeys.md` + `../../decisions/api-operation-census.md`. Current census: **78 application operations**. Operation 79 is a material Product/T6 reopen.

## 1. Lead outcome

The smallest closed executable contract is:

```text
one OpenAPI 3.0.3 component registry
+
one closed 78-row operation ledger
```

The registry owns reusable wire shapes once. Each operation row owns only:

```text
operationId
method + path
request component / no body
request-header profile
success status + body
success-header profile
exact Problem.code set
exact filters/order when list-like
request/body limit profile
```

The wire does **not** restate Authorization predicates, lifecycle transitions, transaction law, target eligibility, persistence shape, provider state, or future capability. Those remain in their T1→T8-D owners.

---

# 2. Global wire laws

## 2.1 Surface / authentication / routing

Application SSOT:

```text
api/openapi/v1/openapi.yaml
OpenAPI 3.0.3
paths already include /api/v1
no absolute runtime host is durable contract
same-origin browser baseline
no CORS baseline
```

Root security scheme:

```yaml
MetalDocsSession:
  type: apiKey
  in: cookie
  name: __Host-metaldocs_session
```

All 78 operations require it. OIDC `/auth/login` and `/auth/callback` remain outside this SSOT. OpenAPI security scopes never encode MetalDocs AuthZ.

Cookie law remains:

```text
Secure
HttpOnly
SameSite=Lax
Path=/
Domain absent
```

Routes are case-sensitive and exact. No automatic trailing-slash redirect, HEAD, or OPTIONS route is added. Unknown `/api/v1` path -> `404 notfound.resource`; undeclared method on a known path -> `405 request.method_not_allowed` with exact `Allow` from the census.

`Accept`/`Accept-Language` do not create alternative representations or 406 behavior at Launch.

## 2.2 Strict request decoding

Every fixed JSON object uses:

```text
additionalProperties: false
```

The only deliberate map is `DraftUploadAllocation.required_headers`.

Central request handling rejects:

```text
unknown JSON member                         400 request.invalid
duplicate JSON object member name          400 request.invalid
unknown query parameter                    400 request.invalid
duplicate scalar query parameter           400 request.invalid
body on an operation declared bodyless     400 request.invalid
malformed JSON / path / query / header      400 request.invalid
wrong request media type                    415 request.unsupported_media_type
non-identity JSON Content-Encoding          415 request.unsupported_media_type
raw JSON body > 65,536 bytes                413 request.content_too_large
```

JSON is UTF-8. Unsupported content coding returns `Accept-Encoding: identity`. The eventual OAS marks JSON bodies with:

```text
x-metaldocs-max-request-body-bytes: 65536
```

The raw limit fires before JSON decode. Document bytes never travel in an application JSON body.

## 2.3 Presence / composition

```text
required = member must be present
optional = member may be absent
nullable = explicit JSON null is a semantic value
```

Absence and `null` are never interchangeable by convention. PATCH/PUT never acquires implicit “null means clear”. OAS 3.0.3 `nullable: true` is used only where explicit null is semantic; the baseline nullable member is `Page.next_cursor`.

True semantic unions use `oneOf` + required discriminator. Do not use `allOf` inheritance merely to reduce YAML. Request and response shapes are purpose-built instead of hiding overbroad DTOs with `readOnly`/`writeOnly`.

## 2.4 Scalar normalization

```text
Uuid
  input: RFC UUID text, case-insensitive
  output: canonical lowercase 8-4-4-4-12 text

UtcInstant
  RFC3339 UTC; server emits Z

RevisionOrdinal / ByteCount
  integer 0..9007199254740991

Sha256Hex
  ^[0-9a-f]{64}$

SearchQuery
  trim -> nonblank; normalized length <=256

CodeInput
  trim -> uppercase ASCII -> validate ^[A-Z0-9]+$
  normalized length 1..32
  '-' forbidden because Product owns the numbering separator

CodeToken
  canonical response ^[A-Z0-9]+$, length 1..32

DocumentCode
  ^[A-Z0-9]+(?:-[A-Z0-9]+)?-[0-9]{3,}$
  maxLength=85

OpaqueCursor
  nonblank, maxLength=2048

CsrfToken
  opaque/nonblank, maxLength=512

ProviderSubjectRef
  opaque/nonblank, maxLength=2048
  server-resolvable external issuer+subject handle
  client must not parse or treat it as Product identity
  unchanged current binding returns byte-stable ref
```

Other human text is validated nonblank where stated and is bounded by the aggregate JSON body ceiling rather than guessed field-specific maxima.

Document `q` comparison:

```text
code: ASCII case-insensitive against canonical code
title: NFC normalization + Unicode case folding; diacritics preserved
no accent folding, stemming, fuzzy match, body/OCR/vector search
```

Ranking remains T6:

```text
q present: exact code -> code prefix -> title prefix -> title contains -> code -> document_id
q absent:  code -> document_id
```

## 2.5 Strong ETag / conditional law

ETags are strong, opaque, server-verifiable tokens bound to:

```text
resource identity
+ exact concurrency domain
+ owner current version/equivalent
```

They never expose raw database version/generation.

```text
missing/malformed required condition               -> 400 request.invalid
well-formed but wrong resource/domain/tampered tag -> 412 precondition.resource_changed
stale same-domain tag + different desired state    -> 412 precondition.resource_changed
```

Whole-replacement retry exception already accepted by T8-E is now deterministic:

```text
stale authenticated same-domain tag
+ requested whole representation == exact current representation
-> success with current representation/current ETag
-> zero mutation
-> zero duplicate Audit
```

This exception applies to whole-replacement current truths only. DRAFT PATCH with stale ETag **always** returns `412 precondition.draft_changed` even if its desired values happen to match current state.

An ETag-protected representation contains only fields governed by that concurrency token plus stable immutable identifiers. Independently mutable labels are excluded from such views.

`PROFILE_REPLACE`:

```text
existing profile -> If-Match required
absent profile recreation -> If-None-Match: * required
both / neither -> 400
If-None-Match:* while profile exists -> 412
```

## 2.6 CSRF / idempotency

Every unsafe application operation requires `X-CSRF-Token`.

Exactly these 10 semantic POST creations additionally require `Idempotency-Key`:

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

The key is a client-generated UUID **value**; textual UUID case does not create a distinct key.

Scope:

```text
current user id + canonical operationId + Idempotency-Key UUID value
```

Fingerprint uses validated normalized semantic command fields, never raw HTTP bytes.

Completed replay:

```text
current session/CSRF/AuthZ/disclosure rechecked
stored success status/body replayed exactly
historical lifecycle/preconditions not re-run
no second semantic mutation/Audit
no replay-indicator response header
```

`createSubmission` still requires a structurally valid `If-Match`; on a completed idempotency replay the historical DRAFT precondition is not evaluated again.

Same key + different fingerprint -> `422 validation.idempotency_key_reused`.

The 10 replay bodies contain only stable IDs + tiny closed result enums. Freeze operation-local ReplaySnapshot payload maximum:

```text
2,048 bytes
```

No ReplaySnapshot contains erasable UserProfile PII, provider payloads/tokens, title/reason/message/free text, raw request/response bytes, or governed content.

Replay retention remains the already-ratified bounded mechanism guarantee exceeding normal retry windows; no exact duration becomes a public wire promise in T8-E.

## 2.7 Request processing precedence

For failures that coexist, the boundary applies:

```text
1. exact route/method
2. raw envelope guards (body size / coding / media type)
3. ApplicationSession authentication
4. CSRF for unsafe request
5. structural parse + schema + pure normalization
6. current AuthZ/disclosure
7. completed idempotency replay recognition where applicable
8. current conditional/lifecycle/referential checks
9. business effect
```

Rate limiting or a technical dependency failure may terminate earlier when that protection/dependency is what prevents reaching the next stage. No error precedence may bypass current AuthZ to disclose stored replay results.

## 2.8 Pagination

Potentially unbounded lists use stateless integrity-protected seek cursors.

First page:

```text
filters (only those named for the operation)
+ optional limit 1..100 (default 20)
```

Next page:

```text
cursor
+ optional limit
```

When `cursor` is present, re-supplying any operation filter/query other than `limit` is `400 request.invalid`. The cursor carries and authenticates operationId + normalized filters + ordering. Invalid/tampered cursor -> 400. Current AuthZ is rechecked on every page.

Response:

```text
items count <= effective limit
has_more=true  <=> next_cursor is non-null
has_more=false <=> next_cursor is null
```

No offset, total count, generic sort DSL, server cursor state, or frozen multi-page snapshot.

Purpose-built creation/options arrays are **complete, not truncated**. If real cardinality makes a complete synchronous options response unsustainable, that evidence reopens T6; T8-E does not silently truncate eligible Areas/DocumentTypes/Templates/owner candidates or add operation 79.

## 2.9 Response/header profiles

```text
NO_STORE
  Cache-Control: no-store

JSON_NO_STORE
  Content-Type: application/json
  Cache-Control: no-store

JSON_ETAG
  Content-Type: application/json
  ETag: strong current tag
  Cache-Control: no-store

JSON_ETAG_MUTATION
  Content-Type: application/json
  ETag: resulting/current strong tag
  Cache-Control: no-store

SESSION_END
  Cache-Control: no-store
  Set-Cookie clears __Host-metaldocs_session:
    empty value; Path=/; Max-Age=0; Secure; HttpOnly; SameSite=Lax; Domain absent
```

No baseline `Location`, replay header, permission snapshot, provider header, or generic metadata header.

Problem responses: `application/problem+json` + `Cache-Control: no-store`; 401 adds `WWW-Authenticate: MetalDocsSession`; 405 adds exact `Allow`; 429 may add truthful non-negative delta-seconds `Retry-After` but never fabricates one.

## 2.10 Exact-byte response profile

Semantic byte resources return authenticated application-origin `200` only.

```text
Content-Type:
  docx -> application/vnd.openxmlformats-officedocument.wordprocessingml.document
  pdf  -> application/pdf
Content-Length: exact raw byte count
Content-Disposition: inline; filename="<document_code>-REV<ordinal-min-width-3>.<ext>"
Content-Digest: sha-256=:<base64 exact SHA-256 bytes>:
Cache-Control: private, no-store, no-transform
Accept-Ranges: none
X-Content-Type-Options: nosniff
Content-Encoding: absent
```

No provider redirect, Range/206, 304, automatic HEAD, or transformation. `Range` on these GETs -> `400 request.invalid`. Existing visible semantic record with missing/corrupt bytes -> `500 internal.content_integrity`; temporary storage/dependency failure -> 503.

---

# 3. Shared schema registry

All fixed objects below are closed. Unless marked `?`, members are required.

## 3.1 References / core enums

```text
UserReference
  user_id: Uuid
  display_name?: string
  // display name may disappear after lawful UserProfile erasure

AreaReference
  area_id: Uuid
  code: CodeToken
  name: string

DocumentTypeReference
  document_type_id: Uuid
  code: CodeToken
  name: string

DocumentReference
  document_id: Uuid
  code: DocumentCode
  // title intentionally absent: title belongs Revision

RevisionIdentity
  revision_id: Uuid
  ordinal: RevisionOrdinal

RevisionReference
  revision: RevisionIdentity
  title: nonblank string
  // used where title is stable for the projection; DRAFT uses current ETag view instead

ContentSummary
  sha256: Sha256Hex
  size_bytes: ByteCount
  content_format: ContentFormat
```

Closed enums:

```text
ContentFormat                docx | pdf
RevisionState                draft | submitted | effective | superseded | obsolete | cancelled
OpenRevisionState            draft | submitted
UserEligibilityState         enabled | disabled
AreaLifecycleState           active | retired
NumberingScope               document_type | document_type_area
GovernanceMode               no_human_approval | use_governance_route
GovernanceSelectorKind       named_user | group
GovernanceDecisionOutcome    accept | return_for_changes
GovernanceSubjectKind        submission | obsolescence
GovernanceAttemptState       active | completed | returned | withdrawn | cancelled
GovernanceStepState          pending | active | decided
RepresentationKind           source_only | require_official_rendition
OfficialRenditionFormat      pdf
RoleCode                     governance_admin | area_manager | author | approver | viewer | governance_viewer
RoleAssignmentSubjectKind    user | group
RoleAssignmentScopeKind      company | area
DocumentCatalogStatus        effective | obsolete | cancelled
DocumentOfficialStatus       draft | submitted | effective | obsolete | cancelled
SubmissionCreationState      governance_pending | rendition_pending | released
SubmissionTerminationKind   returned_for_changes | withdrawn | cancelled
ObsolescenceRequestState     active | returned | withdrawn | completed
ObsolescenceCreationState    governance_pending | obsolete
GovernanceCaseAction         accept | return_for_changes | add_feedback
```

`PermissionCode` is exactly the accepted 14-value T3 vocabulary and retains dot spelling.

True unions:

```text
RoleAssignmentSubject
  {kind:user,user_id}
  {kind:group,group_id}

RoleAssignmentScope
  {kind:company}
  {kind:area,area_id}

GovernanceSelector
  {kind:named_user,user_id}
  {kind:group,group_id}

GovernancePolicy
  {mode:no_human_approval}
  {mode:use_governance_route,steps:[GovernanceRouteStep,...]}

GovernanceRouteStep
  label: nonblank string
  selector: GovernanceSelector
  // array order is route order; no ordinal is exposed on the wire

RepresentationPolicy
  {kind:source_only}
  {kind:require_official_rendition,format:pdf}
```

## 3.2 Page / Problem error

```text
Page
  next_cursor: OpaqueCursor | null
  has_more: boolean

ProblemError
  pointer: RFC6901 pointer beginning /path, /query, /header, or /body
  detail: human-readable string
```

Rejected sensitive values are never echoed.

## 3.3 Session / Organization

```text
SessionView
  user: UserReference
  csrf_token: CsrfToken

ProviderSubjectOption
  provider_subject_ref: ProviderSubjectRef
  display_hints: array<string>, maxItems=3, each nonblank <=256 chars

ProviderSubjectSearchView
  items: array<ProviderSubjectOption>, maxItems=20

CompanyView
  company_id: Uuid
  display_name: nonblank string

ReplaceCompanyRequest
  display_name: nonblank string

UserProfileInput
  display_name: nonblank string
  email?: EmailAddress

CreateUserRequest
  provider_subject_ref: ProviderSubjectRef
  profile: UserProfileInput

CreateUserResult
  user_id: Uuid

UserView
  user: UserReference
  eligibility: UserEligibilityState

UserPage
  items: UserView[]
  page: Page

UserProfileView
  user_id: Uuid
  display_name: nonblank string
  email?: EmailAddress

ReplaceUserProfileRequest
  display_name: nonblank string
  email?: EmailAddress
  // whole replacement: omitted email => no resulting email

UserProviderBindingView
  user_id: Uuid
  provider_subject_ref: ProviderSubjectRef

ReplaceUserProviderBindingRequest
  provider_subject_ref: ProviderSubjectRef

UserEligibilityView
  user_id: Uuid
  state: UserEligibilityState

ReplaceUserEligibilityRequest
  state: UserEligibilityState

AreaView
  area_id: Uuid
  code: CodeToken
  name: nonblank string

AreaSummary
  area: AreaReference
  state: AreaLifecycleState

AreaPage
  items: AreaSummary[]
  page: Page

CreateAreaRequest
  code: CodeInput
  name: nonblank string
  // new Area starts active

CreateAreaResult
  area_id: Uuid

ReplaceAreaRequest
  name: nonblank string
  // Area code is immutable and absent

AreaLifecycleView
  area_id: Uuid
  state: AreaLifecycleState

ReplaceAreaLifecycleRequest
  state: AreaLifecycleState

GroupView
  group_id: Uuid
  name: nonblank string

GroupPage
  items: GroupView[]
  page: Page

CreateGroupRequest / ReplaceGroupRequest
  name: nonblank string

CreateGroupResult
  group_id: Uuid

GroupMemberPage
  items: UserReference[]
  page: Page
```

Profile DELETE removes the profile subresource; if already absent, it is 404 rather than an invented “ensure absent” command.

## 3.4 Authorization

T3-derived role projection is exact:

```text
governance_admin
  scopes: [company]
  permissions: organization.manage, access.manage, document_type.manage, template_use.manage

area_manager
  scopes: [area]
  permissions: document.read_effective, document.read_history, document.read_working,
               document.create, document.edit, document.submit,
               document.cancel_revision, document.obsolete, document.owner.manage,
               governance.act

author
  scopes: [company,area]
  permissions: document.read_effective, document.read_history, document.read_working,
               document.create, document.edit, document.submit

approver
  scopes: [company,area]
  permissions: document.read_effective, governance.act

viewer
  scopes: [company,area]
  permissions: document.read_effective

governance_viewer
  scopes: [company,area]
  permissions: document.read_effective, document.read_history, audit.read
```

```text
RoleView
  code: RoleCode
  permissions: unique PermissionCode[] in the exact bundle above
  allowed_scope_kinds: unique RoleAssignmentScopeKind[] in the exact order above

RoleListView
  items in canonical order:
    governance_admin, area_manager, author, approver, viewer, governance_viewer

RoleAssignmentView
  assignment_id: Uuid
  subject: RoleAssignmentSubject
  role: RoleCode
  scope: RoleAssignmentScope

RoleAssignmentPage
  items: RoleAssignmentView[]
  page: Page

CreateRoleAssignmentRequest
  subject: RoleAssignmentSubject
  role: RoleCode
  scope: RoleAssignmentScope

CreateRoleAssignmentResult
  assignment_id: Uuid
```

No editable Role/Permission policy API exists.

## 3.5 Document Governance

```text
DocumentTypeView
  document_type_id: Uuid
  code: CodeToken
  name: nonblank string
  numbering_scope: NumberingScope
  active: boolean

DocumentTypePage
  items: DocumentTypeView[]
  page: Page

CreateDocumentTypeRequest
  code: CodeInput
  name: nonblank string
  numbering_scope: NumberingScope
  active: boolean
  governance: GovernancePolicy
  representation: RepresentationPolicy
  // eligible-template set starts empty

CreateDocumentTypeResult
  document_type_id: Uuid

ReplaceDocumentTypeRequest
  code: CodeInput
  name: nonblank string
  numbering_scope: NumberingScope
  active: boolean

DocumentTypeGovernanceView / ReplaceDocumentTypeGovernanceRequest
  governance: GovernancePolicy
  representation: RepresentationPolicy

EligibleTemplatesView
  templates: DocumentReference[]
  // stable references only; this is an ETag concurrency representation

ReplaceEligibleTemplatesRequest
  template_document_ids: unique Uuid[]
  // empty array valid

NumberingPreviewView
  preview_code: DocumentCode
  reservation: false

TemplateConfigurationItem
  document: DocumentReference
  template_role: boolean
  has_effective_revision: boolean
  current_effective_title?: nonblank string
  eligible_document_type_ids: unique Uuid[]

TemplateConfigurationPage
  items: TemplateConfigurationItem[]
  page: Page
```

Create explicitly supplies initial governance/representation because T8-D requires current values and no accepted default exists. Base replacement rejects normalized code/numbering-scope changes after first committed Document with `409 state.conflict`.

## 3.6 Controlled Documents — create/read/work

```text
TemplateCreationOption
  document: DocumentReference
  effective_revision: RevisionReference

DocumentCreationOptionsView
  areas: AreaReference[]
  document_types: DocumentTypeReference[]
  templates: TemplateCreationOption[]
  default_responsible_owner: UserReference
  responsible_owner_candidates?: UserReference[]
```

Options are complete and ordered:

```text
areas: code, area_id
document_types: code, document_type_id
templates: document.code, document_id
responsible_owner_candidates: user_id
```

Absence of `responsible_owner_candidates` means caller lacks `document.owner.manage`; present empty means capability exists but no alternate eligible target.

```text
CreateDocumentRequest
  document_type_id: Uuid
  area_id: Uuid
  title: nonblank string
  template_document_id?: Uuid
  responsible_owner_user_id?: Uuid

CreateDocumentResult
  document_id: Uuid
  revision_id: Uuid

DocumentSummary
  document: DocumentReference
  document_type: DocumentTypeReference
  area: AreaReference
  responsible_owner: UserReference
  status: DocumentCatalogStatus
  official_revision?: RevisionReference

DocumentPage
  items: DocumentSummary[]
  page: Page

ReleasedRevisionView
  revision: RevisionIdentity
  title: nonblank string
  release_id: Uuid
  released_at: UtcInstant
  source: ContentSummary
  representation:
    {kind:source_only}
    {kind:official_rendition,official_rendition_id:Uuid,content:ContentSummary}

DocumentOfficialView
  document: DocumentReference
  document_type: DocumentTypeReference
  area: AreaReference
  responsible_owner: UserReference
  status: DocumentOfficialStatus
  official?: ReleasedRevisionView
```

`official` is present iff at least one Release exists; obsolete keeps the last released official view. A newer cancelled/open Revision never replaces an older EFFECTIVE official truth. Before first Release, status is the open DRAFT/SUBMITTED state or `cancelled` and `official` is absent.

```text
ResponsibleOwnerView
  document_id: Uuid
  responsible_owner_user_id: Uuid

ReplaceResponsibleOwnerRequest
  user_id: Uuid

TemplateRoleView
  document_id: Uuid
  is_template: boolean

ReplaceTemplateRoleRequest
  is_template: boolean

CreateRevisionResult
  revision_id: Uuid

RevisionView
  revision: RevisionIdentity
  document: DocumentReference
  title: nonblank string
  state: RevisionState
  created_at: UtcInstant
  current_submission_id?: Uuid
  // present iff state=submitted

DocumentWorkView
  document: DocumentReference
  revision: RevisionIdentity
  title: nonblank string
  content: ContentSummary
  updated_at: UtcInstant
  // raw generation is absent; ETag is wire OCC authority

UpdateDraftRequest
  title?: nonblank string
  source_upload_id?: Uuid
  minProperties=1; null forbidden; omitted=unchanged

DraftUploadAllocation
  upload_id: Uuid
  upload_url: URI
  expires_at: UtcInstant
  max_bytes: ByteCount // unresolved until §7 measurement
  required_headers: map<string,string>
```

`required_headers` contains only exact browser-settable headers required by the create-only provider PUT. No provider account/bucket/key/version/storage ETag is exposed.

Upload completion is bodyless/naturally idempotent. A completed READY handle is recognized before OPEN-expiry handling for a repeat; an uncompleted known expired handle returns 410. Server independently derives descriptor; completion never accepts or returns a client-authored authoritative hash/size/format.

## 3.7 Submission / Governance / history

```text
SubmissionCreateResult
  {state:governance_pending,submission_id:Uuid,governance_attempt_id:Uuid}
  {state:rendition_pending,submission_id:Uuid}
  {state:released,submission_id:Uuid,release_id:Uuid}

SubmissionHumanGate
  required: boolean
  satisfied: boolean

SubmissionRepresentationGate
  required: boolean
  satisfied: boolean
  attention_required: boolean

SubmissionView
  submission_id: Uuid
  revision: RevisionIdentity
  title: nonblank string                 // frozen Submission title snapshot
  submitter: UserReference
  submitted_at: UtcInstant
  content: ContentSummary
  governance_mode: GovernanceMode
  representation: RepresentationPolicy
  human_gate: SubmissionHumanGate
  representation_gate: SubmissionRepresentationGate
  governance_attempt_id?: Uuid
  release_id?: Uuid
  termination?: SubmissionTerminationKind
```

Cross-field law:

```text
governance_attempt_id present iff governance_mode=use_governance_route
human_gate.required iff governance_mode=use_governance_route
not-required human gate is satisfied
representation_gate.required iff representation=require_official_rendition
not-required representation gate is satisfied
attention_required only when rendition required, unsatisfied, and terminal renderer attention exists
release_id and termination mutually exclusive
```

```text
SubmissionWithdrawalView
  submission_id: Uuid
  actor: UserReference
  withdrawn_at: UtcInstant

RevisionCancellationRequest
  reason: nonblank string

RevisionCancellationView
  revision_id: Uuid
  actor: UserReference
  reason: nonblank string
  cancelled_at: UtcInstant
```

Cancellation singleton: first -> 201; exact same reason repeat -> existing 200; later different reason -> `409 state.conflict`.

```text
GovernanceFeedbackView
  feedback_id: Uuid
  actor: UserReference
  message: nonblank string
  created_at: UtcInstant

GovernanceFeedbackPage
  items: GovernanceFeedbackView[]
  page: Page

CreateGovernanceFeedbackRequest
  message: nonblank string

CreateGovernanceFeedbackResult
  feedback_id: Uuid

GovernanceDecisionRequest
  {outcome:accept}
  {outcome:return_for_changes,reason:nonblank string}

GovernanceDecisionView
  {decision_id:Uuid,outcome:accept,actor:UserReference,decided_at:UtcInstant}
  {decision_id:Uuid,outcome:return_for_changes,actor:UserReference,decided_at:UtcInstant,reason:nonblank string}

GovernanceStepView
  {step_id:Uuid,label:string,state:pending}
  {step_id:Uuid,label:string,state:active}
  {step_id:Uuid,label:string,state:decided,decision:GovernanceDecisionView}
```

Step array order is authoritative route order; no persistence ordinal is exposed. Candidate users, Group membership, selector internals, grants, and provider claims are absent.

```text
SubmissionGovernanceSubject
  kind: submission
  submission_id: Uuid
  document: DocumentReference
  revision: RevisionIdentity
  title: nonblank string
  submitter: UserReference
  submitted_at: UtcInstant
  content: ContentSummary

ObsolescenceGovernanceSubject
  kind: obsolescence
  request_id: Uuid
  document: DocumentReference
  target_revision: RevisionReference
  initiator: UserReference
  reason: nonblank string
  requested_at: UtcInstant

GovernanceCaseView
  governance_attempt_id: Uuid
  state: GovernanceAttemptState
  subject: SubmissionGovernanceSubject | ObsolescenceGovernanceSubject
  steps: GovernanceStepView[] in route order
  feedback: GovernanceFeedbackPage // first 20
  allowed_actions: unique GovernanceCaseAction[]
```

Embedded feedback cursor, when non-null, is minted for `listGovernanceFeedback`. `allowed_actions` canonical order is `accept`, `return_for_changes`, `add_feedback` filtered to current truth; array may be empty. Every command rechecks canonical AuthZ/lifecycle.

Decision singleton: first -> 201; exact same outcome + same required reason -> existing 200; any later different outcome/reason -> `409 state.governance_step_already_decided`.

`DocumentHistoryItem` is a closed `kind`-discriminated union:

```text
revision_created           revision,title,occurred_at
submission_created         submission_id,revision,title,submitter,occurred_at,?governance_attempt_id
governance_decision        decision_id,governance_attempt_id,step_id,actor,outcome,occurred_at,?reason
feedback_added             feedback_id,governance_attempt_id,actor,message,occurred_at
submission_withdrawn       submission_id,actor,occurred_at
revision_cancelled         revision_id,actor,reason,occurred_at
release_completed          release_id,revision_id,submission_id,occurred_at,?predecessor_revision_id
official_rendition_completed official_rendition_id,submission_id,occurred_at
obsolescence_requested     request_id,target_revision_id,initiator,reason,occurred_at,?governance_attempt_id
obsolescence_withdrawn     request_id,actor,occurred_at
obsolescence_completed     request_id,target_revision_id,occurred_at
```

```text
DocumentHistoryPage
  items: DocumentHistoryItem[]
  page: Page

WorkAuthoringItem
  document: DocumentReference
  revision: RevisionIdentity
  title: nonblank string
  state: OpenRevisionState
  responsible_owner: UserReference
  updated_at: UtcInstant

WorkAuthoringPage
  items: WorkAuthoringItem[]
  page: Page

WorkGovernanceItem
  governance_attempt_id: Uuid
  subject_kind: GovernanceSubjectKind
  document: DocumentReference
  created_at: UtcInstant

WorkGovernancePage
  items: WorkGovernanceItem[]
  page: Page
```

## 3.8 Release / obsolescence

```text
ReleaseView
  source-only:
    release_id,document,revision,title,submission_id,released_at,
    ?predecessor_revision_id,
    representation:{kind:source_only,source:ContentSummary}

  official-rendition:
    same core,
    representation:{kind:official_rendition,source:ContentSummary,
                    official_rendition_id:Uuid,official_rendition:ContentSummary}

ObsolescenceRequestCreateRequest
  reason: nonblank string

ObsolescenceCreateResult
  {state:governance_pending,request_id:Uuid,governance_attempt_id:Uuid}
  {state:obsolete,request_id:Uuid}
```

`ObsolescenceRequestView` is a closed state union:

```text
active
  request_id,document,target_revision,initiator,reason,state=active,requested_at,governance_attempt_id

returned / withdrawn / completed-human
  same core + matching state + governance_attempt_id + ended_at

completed-no-human
  same core + state=completed + ended_at
  governance_attempt_id absent
```

```text
ObsolescenceWithdrawalView
  request_id: Uuid
  actor: UserReference
  withdrawn_at: UtcInstant
```

Release remains immutable/system-owned; no publish mutation exists.

---

# 4. Audit wire

Audit wire projects T3 evidence; it never exports raw `audit.events.facts` JSONB or operational correlation metadata.

```text
AuditActor
  {kind:user,user_id:Uuid}
  {kind:system,system_actor_code:metaldocs}

AuditVisibility
  {kind:company}
  {kind:area,area_id:Uuid}
```

Current wire `AuditOperationCode` contains the **37 reachable** Launch codes:

```text
provider_binding.accepted
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

`provider_binding.disabled` is deliberately **not** exposed while the bounded T3 contradiction in §8.2 is unresolved.

`document.created` means the atomic Document+REV000 creation action. `revision.created` is a later Revision creation, not a duplicate REV000 event.

Resource kinds:

```text
provider_binding | user | user_profile | area | group | role_assignment |
document_type | document | revision | submission | governance_decision |
official_rendition | release | obsolescence_request
```

GroupMembership has no invented UUID/resource kind: membership events use `resource_kind=group`, `resource_id=group_id`, plus typed membership facts.

Typed facts exist only where the current Audit consumer needs facts beyond operation/resource identity:

```text
GroupMembershipAuditFacts
  group_id: Uuid
  user_id: Uuid

RoleAssignmentAuditFacts
  assignment_id: Uuid
  subject: RoleAssignmentSubject
  role: RoleCode
  scope: RoleAssignmentScope

GovernanceDecisionAuditFacts
  governance_attempt_id: Uuid
  step_id: Uuid
  decision_id: Uuid
  subject_kind: GovernanceSubjectKind
  subject_id: Uuid
  outcome: GovernanceDecisionOutcome

ReleaseAuditFacts
  document_id: Uuid
  revision_id: Uuid
  submission_id: Uuid
  predecessor_revision_id?: Uuid

RevisionCancellationAuditFacts
  document_id: Uuid
  revision_id: Uuid

ObsolescenceAuditFacts
  document_id: Uuid
  target_revision_id: Uuid
```

`AuditEventView` is a closed operation-code-discriminated union with common:

```text
event_id: Uuid
occurred_at: UtcInstant
actor: AuditActor
operation_code: AuditOperationCode
resource_kind
resource_id: Uuid
visibility: AuditVisibility
```

Simple operation branches have no `facts` member. Membership/RoleAssignment/governance decision/Release/revision cancellation/obsolescence branches require the matching typed facts above. No free-form feedback/reason/profile/provider payload is added by wire convenience.

`AuditEventPage = {items: AuditEventView[], page: Page}` ordered `occurred_at DESC, event_id DESC`. `GET /audit/events` accepts only cursor/limit; `audit.read` historical visibility filtering happens before pagination. No Audit filter is invented by inference.

---

# 5. RFC 9457 closed Problem profile

Each code is its own full closed schema; there is no open inherited base and no `default` response.

Required:

```text
type
title
status
detail
instance
code
trace_id
```

`instance` is `urn:uuid:<fresh UUID>` per occurrence. `trace_id` is opaque/nonblank. `errors[]`, when present, is non-empty `ProblemError[]` and is allowed only on `request.invalid` and validation-family variants.

Catalog:

| code | status | fixed title |
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

Disclosure law:

```text
no valid session -> 401
visible request/action but missing permission/trust -> 403
absent or non-disclosable item -> 404
```

Collections may return authorized empty results rather than manufacture 403. `listDocuments status=obsolete|cancelled` is the exception: those management catalog modes require current `document.read_history` in at least one relevant scope; otherwise 403. Default/effective catalog is authorization-filtered and may be empty.

Ledger macros below are textual exact-set abbreviations only; final OAS expands them:

```text
B = request.invalid + auth.unauthenticated + ratelimit.exceeded + internal.failure + dependency.unavailable
A = B + permission.denied
C = B + permission.csrf_failed
U = A + permission.csrf_failed
J = request.unsupported_media_type + request.content_too_large + validation.failed
I = validation.idempotency_key_reused
N = notfound.resource
S = state.conflict
P = precondition.resource_changed
D = precondition.draft_changed
X = internal.content_integrity
```

---

# 6. Closed 78-operation ledger

All path `*_id` parameters are required `Uuid`. `PAGED` means the §2.8 cursor/limit law. JSON request rows inherit the 65,536-byte limit and `J` where shown.

## 6.1 Session / Organization / Authorization / Document Governance — 1→43

|#|operationId|Method + path|Request/profile|Success|Headers|Query/order|Problems|
|---:|---|---|---|---|---|---|---|
|1|`getSession`|`GET /api/v1/session`|`SAFE_READ`|`200 SessionView`|`JSON_NO_STORE`|none|`B`|
|2|`endSession`|`DELETE /api/v1/session`|no body / `UNSAFE_CSRF`|`204`|`SESSION_END`|none|`C`|
|3|`searchProviderSubjects`|`GET /api/v1/authentication/provider-subjects`|`SAFE_READ`|`200 ProviderSubjectSearchView`|`JSON_NO_STORE`|required `query:SearchQuery`; provider order|`A + validation.failed`|
|4|`getCompany`|`GET /api/v1/company`|`SAFE_READ`|`200 CompanyView`|`JSON_ETAG`|none|`A`|
|5|`replaceCompany`|`PUT /api/v1/company`|`ReplaceCompanyRequest` / `IF_MATCH_MUTATION`|`200 CompanyView`|`JSON_ETAG_MUTATION`|none|`U + J + P`|
|6|`listUsers`|`GET /api/v1/users`|`SAFE_READ`|`200 UserPage`|`JSON_NO_STORE`|`PAGED`; user_id ASC|`A`|
|7|`createUser`|`POST /api/v1/users`|`CreateUserRequest` / `IDEMPOTENT_CREATE`|`201 CreateUserResult`|`JSON_NO_STORE`|none|`U + J + I + S`|
|8|`getUser`|`GET /api/v1/users/{user_id}`|`SAFE_READ`|`200 UserView`|`JSON_NO_STORE`|none|`A + N`|
|9|`getUserProfile`|`GET /api/v1/users/{user_id}/profile`|`SAFE_READ`|`200 UserProfileView`|`JSON_ETAG`|none|`A + N`|
|10|`replaceUserProfile`|`PUT /api/v1/users/{user_id}/profile`|`ReplaceUserProfileRequest` / `PROFILE_REPLACE`|`200` replace or `201` recreate, `UserProfileView`|`JSON_ETAG_MUTATION`|none|`U + J + N + P`|
|11|`deleteUserProfile`|`DELETE /api/v1/users/{user_id}/profile`|no body / `UNSAFE_CSRF`|`204` first delete|`NO_STORE`|absent profile ->404|`U + N`|
|12|`getUserProviderBinding`|`GET /api/v1/users/{user_id}/provider-binding`|`SAFE_READ`|`200 UserProviderBindingView`|`JSON_ETAG`|none|`A + N`|
|13|`replaceUserProviderBinding`|`PUT /api/v1/users/{user_id}/provider-binding`|`ReplaceUserProviderBindingRequest` / `IF_MATCH_MUTATION`|`200 UserProviderBindingView`|`JSON_ETAG_MUTATION`|none|`U + J + N + P + S`|
|14|`getUserEligibility`|`GET /api/v1/users/{user_id}/eligibility`|`SAFE_READ`|`200 UserEligibilityView`|`JSON_ETAG`|none|`A + N`|
|15|`replaceUserEligibility`|`PUT /api/v1/users/{user_id}/eligibility`|`ReplaceUserEligibilityRequest` / `IF_MATCH_MUTATION`|`200 UserEligibilityView`|`JSON_ETAG_MUTATION`|none|`U + J + N + P`|
|16|`listAreas`|`GET /api/v1/areas`|`SAFE_READ`|`200 AreaPage`|`JSON_NO_STORE`|`PAGED`; code ASC,area_id ASC|`A`|
|17|`createArea`|`POST /api/v1/areas`|`CreateAreaRequest` / `IDEMPOTENT_CREATE`|`201 CreateAreaResult`|`JSON_NO_STORE`|none|`U + J + I + S`|
|18|`getArea`|`GET /api/v1/areas/{area_id}`|`SAFE_READ`|`200 AreaView`|`JSON_ETAG`|none|`A + N`|
|19|`replaceArea`|`PUT /api/v1/areas/{area_id}`|`ReplaceAreaRequest` / `IF_MATCH_MUTATION`|`200 AreaView`|`JSON_ETAG_MUTATION`|none|`U + J + N + P`|
|20|`getAreaLifecycle`|`GET /api/v1/areas/{area_id}/lifecycle`|`SAFE_READ`|`200 AreaLifecycleView`|`JSON_ETAG`|none|`A + N`|
|21|`replaceAreaLifecycle`|`PUT /api/v1/areas/{area_id}/lifecycle`|`ReplaceAreaLifecycleRequest` / `IF_MATCH_MUTATION`|`200 AreaLifecycleView`|`JSON_ETAG_MUTATION`|none|`U + J + N + P`|
|22|`listGroups`|`GET /api/v1/groups`|`SAFE_READ`|`200 GroupPage`|`JSON_NO_STORE`|`PAGED`; group_id ASC|`A`|
|23|`createGroup`|`POST /api/v1/groups`|`CreateGroupRequest` / `IDEMPOTENT_CREATE`|`201 CreateGroupResult`|`JSON_NO_STORE`|none|`U + J + I`|
|24|`getGroup`|`GET /api/v1/groups/{group_id}`|`SAFE_READ`|`200 GroupView`|`JSON_ETAG`|none|`A + N`|
|25|`replaceGroup`|`PUT /api/v1/groups/{group_id}`|`ReplaceGroupRequest` / `IF_MATCH_MUTATION`|`200 GroupView`|`JSON_ETAG_MUTATION`|none|`U + J + N + P`|
|26|`deleteGroup`|`DELETE /api/v1/groups/{group_id}`|no body / `UNSAFE_CSRF`|`204` first delete|`NO_STORE`|absent ->404; live dependency ->409|`U + N + S`|
|27|`listGroupMembers`|`GET /api/v1/groups/{group_id}/members`|`SAFE_READ`|`200 GroupMemberPage`|`JSON_NO_STORE`|`PAGED`; user_id ASC|`A + N`|
|28|`addGroupMember`|`PUT /api/v1/groups/{group_id}/members/{user_id}`|no body / `UNSAFE_CSRF`|`201` first; `204` exact existing relation|`NO_STORE`|none|`U + N + S`|
|29|`removeGroupMember`|`DELETE /api/v1/groups/{group_id}/members/{user_id}`|no body / `UNSAFE_CSRF`|`204` including absent relation when parent Group exists|`NO_STORE`|absent/non-disclosable parent Group ->404|`U + N`|
|30|`listRoles`|`GET /api/v1/roles`|`SAFE_READ`|`200 RoleListView`|`JSON_NO_STORE`|fixed T3 role order|`A`|
|31|`listRoleAssignments`|`GET /api/v1/role-assignments`|`SAFE_READ`|`200 RoleAssignmentPage`|`JSON_NO_STORE`|`PAGED`; assignment_id ASC|`A`|
|32|`createRoleAssignment`|`POST /api/v1/role-assignments`|`CreateRoleAssignmentRequest` / `IDEMPOTENT_CREATE`|`201 CreateRoleAssignmentResult`|`JSON_NO_STORE`|none|`U + J + I + S`|
|33|`deleteRoleAssignment`|`DELETE /api/v1/role-assignments/{assignment_id}`|no body / `UNSAFE_CSRF`|`204` first revoke|`NO_STORE`|absent ->404|`U + N`|
|34|`listDocumentTypes`|`GET /api/v1/document-types`|`SAFE_READ`|`200 DocumentTypePage`|`JSON_NO_STORE`|`PAGED`; document_type_id ASC|`A`|
|35|`createDocumentType`|`POST /api/v1/document-types`|`CreateDocumentTypeRequest` / `IDEMPOTENT_CREATE`|`201 CreateDocumentTypeResult`|`JSON_NO_STORE`|none|`U + J + I + S`|
|36|`getDocumentType`|`GET /api/v1/document-types/{document_type_id}`|`SAFE_READ`|`200 DocumentTypeView`|`JSON_ETAG`|none|`A + N`|
|37|`replaceDocumentType`|`PUT /api/v1/document-types/{document_type_id}`|`ReplaceDocumentTypeRequest` / `IF_MATCH_MUTATION`|`200 DocumentTypeView`|`JSON_ETAG_MUTATION`|none|`U + J + N + P + S`|
|38|`getDocumentTypeGovernance`|`GET /api/v1/document-types/{document_type_id}/governance`|`SAFE_READ`|`200 DocumentTypeGovernanceView`|`JSON_ETAG`|none|`A + N`|
|39|`replaceDocumentTypeGovernance`|`PUT /api/v1/document-types/{document_type_id}/governance`|`ReplaceDocumentTypeGovernanceRequest` / `IF_MATCH_MUTATION`|`200 DocumentTypeGovernanceView`|`JSON_ETAG_MUTATION`|none|`U + J + N + P + S`|
|40|`getDocumentTypeEligibleTemplates`|`GET /api/v1/document-types/{document_type_id}/eligible-templates`|`SAFE_READ`|`200 EligibleTemplatesView`|`JSON_ETAG`|document.code ASC,document_id ASC|`A + N`|
|41|`replaceDocumentTypeEligibleTemplates`|`PUT /api/v1/document-types/{document_type_id}/eligible-templates`|`ReplaceEligibleTemplatesRequest` / `IF_MATCH_MUTATION`|`200 EligibleTemplatesView`|`JSON_ETAG_MUTATION`|none|`U + J + N + P + S`|
|42|`getDocumentTypeNumberingPreview`|`GET /api/v1/document-types/{document_type_id}/numbering-preview`|`SAFE_READ`|`200 NumberingPreviewView`|`JSON_NO_STORE`|optional `area_id:Uuid`|`A + N + validation.failed`|
|43|`listTemplateConfigurations`|`GET /api/v1/document-governance/templates`|`SAFE_READ`|`200 TemplateConfigurationPage`|`JSON_NO_STORE`|`PAGED`; document.code ASC,document_id ASC|`A`|

Rows 35/38/39 are promotion-blocked by §8.1 until Step-label persistence is reconciled.

## 6.2 Controlled Documents / Work — 44→77

|#|operationId|Method + path|Request/profile|Success|Headers|Query/order|Problems|
|---:|---|---|---|---|---|---|---|
|44|`getDocumentCreationOptions`|`GET /api/v1/document-creation/options`|`SAFE_READ`|`200 DocumentCreationOptionsView`|`JSON_NO_STORE`|optional area_id,document_type_id; complete arrays per §2.8|`A + validation.failed`|
|45|`listDocuments`|`GET /api/v1/documents`|`SAFE_READ`|`200 DocumentPage`|`JSON_NO_STORE`|first page: q,document_type_id,area_id,responsible_owner_user_id,status,cursor absent,limit; ranking §2.4; status default effective|`A + validation.failed`|
|46|`createDocument`|`POST /api/v1/documents`|`CreateDocumentRequest` / `IDEMPOTENT_CREATE`|`201 CreateDocumentResult`|`JSON_NO_STORE`|none|`U + J + I + S`|
|47|`getDocument`|`GET /api/v1/documents/{document_id}`|`SAFE_READ`|`200 DocumentOfficialView`|`JSON_NO_STORE`|none|`B + N`|
|48|`getDocumentResponsibleOwner`|`GET /api/v1/documents/{document_id}/responsible-owner`|`SAFE_READ`|`200 ResponsibleOwnerView`|`JSON_ETAG`|none|`A + N`|
|49|`replaceDocumentResponsibleOwner`|`PUT /api/v1/documents/{document_id}/responsible-owner`|`ReplaceResponsibleOwnerRequest` / `IF_MATCH_MUTATION`|`200 ResponsibleOwnerView`|`JSON_ETAG_MUTATION`|none|`U + J + N + P + S`|
|50|`getDocumentTemplateRole`|`GET /api/v1/documents/{document_id}/template-role`|`SAFE_READ`|`200 TemplateRoleView`|`JSON_ETAG`|none|`A + N`|
|51|`replaceDocumentTemplateRole`|`PUT /api/v1/documents/{document_id}/template-role`|`ReplaceTemplateRoleRequest` / `IF_MATCH_MUTATION`|`200 TemplateRoleView`|`JSON_ETAG_MUTATION`|none|`U + J + N + P + S`|
|52|`createDocumentRevision`|`POST /api/v1/documents/{document_id}/revisions`|no body / `IDEMPOTENT_CREATE`|`201 CreateRevisionResult`|`JSON_NO_STORE`|none|`U + N + I + S`|
|53|`getDocumentHistory`|`GET /api/v1/documents/{document_id}/history`|`SAFE_READ`|`200 DocumentHistoryPage`|`JSON_NO_STORE`|`PAGED`; occurred_at ASC,kind,semantic stable id|`A + N`|
|54|`listAuthoringWork`|`GET /api/v1/work/authoring`|`SAFE_READ`|`200 WorkAuthoringPage`|`JSON_NO_STORE`|`PAGED`; document.code ASC,document_id ASC|`B`|
|55|`listGovernanceWork`|`GET /api/v1/work/governance`|`SAFE_READ`|`200 WorkGovernancePage`|`JSON_NO_STORE`|`PAGED`; governance_attempt_id ASC|`B`|
|56|`getRevision`|`GET /api/v1/revisions/{revision_id}`|`SAFE_READ`|`200 RevisionView`|`JSON_NO_STORE`|none|`A + N`|
|57|`getRevisionDraft`|`GET /api/v1/revisions/{revision_id}/draft`|`SAFE_READ`|`200 DocumentWorkView`|`JSON_ETAG`|none|`A + N`|
|58|`updateRevisionDraft`|`PATCH /api/v1/revisions/{revision_id}/draft`|`UpdateDraftRequest` / `IF_MATCH_MUTATION`|`200 DocumentWorkView`|`JSON_ETAG_MUTATION`|none|`U + J + N + D + S + state.upload_expired`|
|59|`startRevisionDraftUpload`|`POST /api/v1/revisions/{revision_id}/draft/uploads`|no body / `UNSAFE_CSRF`|`201 DraftUploadAllocation`|`JSON_NO_STORE`|none|`U + N + S`|
|60|`completeRevisionDraftUpload`|`POST /api/v1/revisions/{revision_id}/draft/uploads/{upload_id}/complete`|no body / `UNSAFE_CSRF`|`204`, including recognized READY repeat|`NO_STORE`|none|`U + N + S + state.upload_expired + validation.content_invalid`|
|61|`getRevisionDraftSource`|`GET /api/v1/revisions/{revision_id}/draft/source`|`SAFE_READ`|`200 exact bytes`|`EXACT_BYTES`|none|`A + N + X`|
|62|`createSubmission`|`POST /api/v1/revisions/{revision_id}/submissions`|no body / `SUBMISSION_CREATE`|`201 SubmissionCreateResult`|`JSON_NO_STORE`|none|`U + N + I + D + S + validation.failed + validation.content_malicious + dependency.malware_inspector_unavailable`|
|63|`getSubmission`|`GET /api/v1/submissions/{submission_id}`|`SAFE_READ`|`200 SubmissionView`|`JSON_NO_STORE`|none|`A + N`|
|64|`getSubmissionSource`|`GET /api/v1/submissions/{submission_id}/source`|`SAFE_READ`|`200 exact bytes`|`EXACT_BYTES`|none|`A + N + X`|
|65|`withdrawSubmission`|`PUT /api/v1/submissions/{submission_id}/withdrawal`|no body / `UNSAFE_CSRF`|`201 SubmissionWithdrawalView` first; `200` exact repeat|`JSON_NO_STORE`|none|`U + N + S`|
|66|`cancelRevision`|`PUT /api/v1/revisions/{revision_id}/cancellation`|`RevisionCancellationRequest` / `UNSAFE_CSRF`|`201 RevisionCancellationView` first; `200` exact repeat|`JSON_NO_STORE`|none|`U + J + N + S`|
|67|`getGovernanceAttempt`|`GET /api/v1/governance-attempts/{attempt_id}`|`SAFE_READ`|`200 GovernanceCaseView`|`JSON_NO_STORE`|embedded first feedback page; steps route order|`A + N`|
|68|`listGovernanceFeedback`|`GET /api/v1/governance-attempts/{attempt_id}/feedback`|`SAFE_READ`|`200 GovernanceFeedbackPage`|`JSON_NO_STORE`|`PAGED`; created_at ASC,feedback_id ASC|`A + N`|
|69|`createGovernanceFeedback`|`POST /api/v1/governance-attempts/{attempt_id}/feedback`|`CreateGovernanceFeedbackRequest` / `IDEMPOTENT_CREATE`|`201 CreateGovernanceFeedbackResult`|`JSON_NO_STORE`|none|`U + J + N + I + S`|
|70|`getGovernanceStepDecision`|`GET /api/v1/governance-attempts/{attempt_id}/steps/{step_id}/decision`|`SAFE_READ`|`200 GovernanceDecisionView`|`JSON_NO_STORE`|none|`A + N`|
|71|`recordGovernanceStepDecision`|`PUT /api/v1/governance-attempts/{attempt_id}/steps/{step_id}/decision`|`GovernanceDecisionRequest` / `UNSAFE_CSRF`|`201 GovernanceDecisionView` first; `200` exact repeat|`JSON_NO_STORE`|none|`U + J + N + S + state.governance_step_already_decided`|
|72|`getRelease`|`GET /api/v1/releases/{release_id}`|`SAFE_READ`|`200 ReleaseView`|`JSON_NO_STORE`|none|`A + N`|
|73|`getReleaseSource`|`GET /api/v1/releases/{release_id}/source`|`SAFE_READ`|`200 exact bytes`|`EXACT_BYTES`|none|`A + N + X`|
|74|`getOfficialRenditionContent`|`GET /api/v1/official-renditions/{rendition_id}/content`|`SAFE_READ`|`200 exact PDF bytes`|`EXACT_BYTES`|none|`A + N + X`|
|75|`createObsolescenceRequest`|`POST /api/v1/documents/{document_id}/obsolescence-requests`|`ObsolescenceRequestCreateRequest` / `IDEMPOTENT_CREATE`|`201 ObsolescenceCreateResult`|`JSON_NO_STORE`|none|`U + J + N + I + S`|
|76|`getObsolescenceRequest`|`GET /api/v1/obsolescence-requests/{request_id}`|`SAFE_READ`|`200 ObsolescenceRequestView`|`JSON_NO_STORE`|none|`A + N`|
|77|`withdrawObsolescenceRequest`|`PUT /api/v1/obsolescence-requests/{request_id}/withdrawal`|no body / `UNSAFE_CSRF`|`201 ObsolescenceWithdrawalView` first; `200` exact repeat|`JSON_NO_STORE`|none|`U + N + S`|

Row 67 is promotion-blocked by §8.1 because GovernanceCase Step labels require a durable frozen label.

## 6.3 Audit — 78

|#|operationId|Method + path|Request/profile|Success|Headers|Query/order|Problems|
|---:|---|---|---|---|---|---|---|
|78|`listAuditEvents`|`GET /api/v1/audit/events`|`SAFE_READ`|`200 AuditEventPage`|`JSON_NO_STORE`|`PAGED`; occurred_at DESC,event_id DESC|`A`|

Count proof:

```text
Session/AuthN                         3
Organization                         26
Authorization                         4
Document Governance                  10
Controlled Documents / Work          34
Audit                                 1
TOTAL                                78
```

---

# 7. Document admission limits — measured prerequisite

Do not substitute internet defaults for Product corpus evidence.

Current repository search contains no representative `.docx` or `.pdf` binary corpus. Before T8-E promotion, measure representative controlled-document files and freeze:

```text
DOC_RAW_MAX_BYTES
DOCX_EXPANDED_MAX_BYTES
DOCX_MAX_ZIP_ENTRIES
```

Structural DOCX validation expands only the top-level OOXML ZIP; it does **not** recursively unpack embedded archives. Therefore no generic nested-archive depth engine/limit is added. Malware inspection remains a separate exact-byte governed-boundary control.

No separate compression-ratio limit is required when raw bytes + total expanded bytes + entry count are all bounded and enforced while streaming expansion.

Measurement protocol:

```text
representative accepted DOCX/PDF set
-> raw byte count by format
-> DOCX top-level entry count
-> DOCX total streamed expanded bytes
-> editor save/reopen corpus constraints
-> server-side rendition/converter constraints
-> select smallest ceiling that admits required corpus + explicit operating headroom
-> prove one fixture just below and one above each ceiling
```

`DraftUploadAllocation.max_bytes` is exactly `DOC_RAW_MAX_BYTES`. Multipart remains absent unless measured accepted files make single create-only PUT materially inadequate.

---

# 8. Bounded upstream contradictions exposed by T8-E

These are evidence-triggered bounded reopens, not preference-based redesign.

## 8.1 T8-D — Governance Step label preservation

Product/T6 requires human-facing configured Step labels and GovernanceAttempt freezes an exact route snapshot. T8-D persists current and frozen steps without any label field/snapshot. Reading a historical Step label from current configuration would allow later config rename to rewrite old governed context.

Method classification:

```text
KNOWN     Product Step label exists
KNOWN     attempt freezes route snapshot
KNOWN     T8-D current/frozen step shape omits label
MATERIAL  historical governed context / persistent meaning
OUTCOME   bounded T8-D completeness reopen only
```

Smallest correction if operator ratifies:

```text
controlled_docs.document_type_governance_steps
  + label TEXT NOT NULL

controlled_docs.governance_attempt_steps
  + label_snapshot TEXT NOT NULL
```

Attempt creation copies the configured label into immutable `label_snapshot` in the coherent route-snapshot transition. No new table, owner, lifecycle, selector, candidate list, workflow platform, or API operation.

## 8.2 T3 — unreachable `ProviderSubjectBinding disabled` Audit event

T3's required Audit census names:

```text
ProviderSubjectBinding accepted / disabled / replaced
```

But current Product/T6 has only creation + replacement of the binding, and T8-D explicitly ratifies that User offboarding **preserves the current ProviderSubjectBinding** while disabling User eligibility and sessions/access. There is no Launch binding-disable command or state transition.

Therefore exposing `provider_binding.disabled` in T8-E would be a dormant impossible event and would contradict current persistence/lifecycle truth.

Method classification:

```text
KNOWN     T3 text names disabled Audit event
KNOWN     no application operation disables binding
KNOWN     T8-D offboarding preserves current binding
MATERIAL  public Audit vocabulary must contain only reachable semantic events
OUTCOME   bounded T3 precision: remove `disabled` from required binding Audit census
```

No Product operation, persistence field, or replacement behavior changes. The current T8-E wire intentionally exposes only `provider_binding.accepted|replaced` until operator adjudication.

T8-E cannot be promoted while §8.1/§8.2 remain unresolved against their owning durable authorities.

---

# 9. Generation / conformance proof

Current disposable probe pins (evidence only, not implementation authorization):

```text
Go          oapi-codegen v2.8.0 strict-server
TypeScript  openapi-typescript 7.13.0 paths/components
```

The pair is minimal because it supplies the accepted typed Go request/response boundary + TS paths/components boundary without adding a generated runtime SDK or second DTO authority.

Probe must prove:

```text
additionalProperties:false does not widen fixed objects
required/optional/nullable stay distinct
closed enums stay finite
oneOf unions avoid any/untyped escape
safe-integer bounds survive TS generation
multiple success statuses generate a closed response set
per-operation Problem variants require no default response
strict-server unexpected errors can route through canonical RFC9457 500 serializer
incoming request validation is demonstrated separately from strict generation
custom 65,536-byte request extension is enforced by central boundary
unknown JSON/query/bodyless-body cases fail as §2 states
one Go wire package + one TS paths/components output are the only generated authorities
no generator/provider field leaks into Product wire
```

Runtime proof shape:

```text
raw HTTP
-> route/envelope/session/CSRF
-> central OpenAPI + strict-request validation
-> generated typed request boundary
-> semantic application handler
-> generated typed response boundary
-> HTTP
-> contract fixture validates exact status + headers + body + Problem variant
```

Required negative fixtures additionally cover:

```text
wrong-domain/stale ETags + exact-current PUT exception
stale DRAFT always 412
PROFILE_REPLACE conditional matrix
Idempotency-Key same/different fingerprint + completed replay
cursor filter replay/tamper rules
options arrays never truncated
role bundles/scope matrix exact
Audit operation/resource/facts combinations exact
router 404/405 without implicit HEAD/OPTIONS
exact-byte Range/redirect/206/304/compression rejection
Content-Digest equals exact body SHA-256
ReplaySnapshot <=2,048 bytes
all measured document ceilings
```

No generic production response-buffer validator is added; generated typed output + contract tests remain the accepted minimum.

External evidence checked in this pass: OpenAPI 3.0.3 schema semantics, RFC 9110 conditional/idempotent HTTP semantics, RFC 9457, RFC 9530, RFC 6585, OWASP API resource/file-upload guidance, current `oapi-codegen`/`openapi-typescript` release documentation.

---

# 10. Structural Inversion / subtractive checkpoint

Current candidate remains true if legacy API/schema shape were opposite:

```text
78 semantic operations still derive from Product/T6
ETag/idempotency/CSRF/pagination still derive from accepted invariants
exact-content wire still derives from T4
component registry + operation ledger still close Writer choices
```

Not introduced:

```text
universal response envelope
generic /actions
generic filter/sort DSL
generic public facts/metadata bag
provider/job state
persisted permission snapshot
editable Role/Permission/policy engine
separate Approval API
operation 79
server-side cursor state
multipart without measured need
Range/HEAD/304 baseline
arbitrary Problem extensions/default response
generator-specific Product fields
nested archive framework
dormant future capability
```

The two upstream findings in §8 are narrow completeness contradictions; neither justifies whole-architecture reopen.

---

# 11. Remaining closure gate

```text
A. operator adjudication + durable reconciliation of §8.1 T8-D Step label
B. operator adjudication + durable reconciliation of §8.2 T3 binding-disabled Audit census
C. representative DOCX/PDF measurement -> freeze §7 numeric ceilings
D. disposable pinned Go+TS generation/compile/type probe
E. exact contract fixtures over all 78 rows
F. final whole-candidate Structural Inversion / YAGNI / overengineering / global-coherence pass
G. only then create isolated review/t8e-fable from exact candidate HEAD
H. Lead adjudication
I. explicit operator ratification
```

Until A→F converge:

```text
T8-E ACTIVE
T8-F NOT OPEN
implementation BLOCKED
Fable NOT STARTED
```
