---
id: t8e-executable-wire-contract-proposal
kind: work
owner: architecture
summary: Active T8-E proposal for the smallest exact 78-operation executable application wire.
---

# T8-E executable wire contract — active proposal

> Temporary / non-authoritative. Implementation remains blocked.

Accepted baseline remains in `../../reference/t8e-checkpoint.md`; Product/API meaning remains in `../../product/journeys.md` + `../../decisions/api-operation-census.md`. Current census: **78 application operations**. Operation 79 is a material Product/T6 reopen.

The two bounded upstream contradictions exposed by this ledger were **operator-approved and durably reconciled on 2026-08-20**:

```text
T8-D  Governance Step label persisted + immutable label_snapshot per GovernanceAttempt Step
T3    unreachable ProviderSubjectBinding-disabled Audit census entry removed
```

They are no longer T8-E blockers.

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

```text
SSOT                         api/openapi/v1/openapi.yaml
OpenAPI                      3.0.3
application prefix           /api/v1
runtime host                 not durable contract
browser baseline             same-origin; no CORS baseline
```

Root security scheme:

```yaml
MetalDocsSession:
  type: apiKey
  in: cookie
  name: __Host-metaldocs_session
```

All 78 operations require it. OIDC `/auth/login` and `/auth/callback` remain outside this SSOT. OpenAPI security scopes never encode MetalDocs AuthZ.

Cookie law remains `Secure; HttpOnly; SameSite=Lax; Path=/; Domain absent`.

Routes are case-sensitive and exact. No automatic trailing-slash redirect, HEAD, or OPTIONS route is added. Unknown `/api/v1` path -> `404 notfound.resource`; undeclared method on a known path -> `405 request.method_not_allowed` with exact `Allow` from the census. `Accept`/`Accept-Language` do not create alternate application representations at Launch.

## 2.2 Strict request decoding / body ceiling

Every fixed request/response/reference/page/union/Problem object has:

```text
additionalProperties: false
```

The only deliberate map is `DraftUploadAllocation.required_headers`.

Central request handling rejects:

```text
unknown JSON member                         400 request.invalid
duplicate JSON object member                400 request.invalid
unknown query parameter                     400 request.invalid
duplicate scalar query parameter            400 request.invalid
body on bodyless operation                  400 request.invalid
malformed path/query/header/JSON             400 request.invalid
wrong request media type                    415 request.unsupported_media_type
non-identity JSON Content-Encoding           415 request.unsupported_media_type
raw application/json body > 65,536 bytes    413 request.content_too_large
```

JSON is UTF-8. Unsupported content coding returns `Accept-Encoding: identity`. The raw 65,536-byte ceiling fires before JSON decoding and is recorded through `x-metaldocs-max-request-body-bytes: 65536`. Document bytes never travel in an application JSON body.

Presence law:

```text
required = member must be present
optional = member may be absent
nullable = explicit JSON null is semantic
```

Absence and `null` are not interchangeable. PATCH/PUT never acquires implicit “null means clear”. OAS 3.0.3 `nullable: true` is used only where explicit null is semantic; baseline nullable member is `Page.next_cursor`.

True closed semantic unions use `oneOf` + required discriminator. Do not use `allOf` inheritance merely to reduce YAML. Request/response shapes are purpose-built rather than hiding overbroad DTOs with `readOnly`/`writeOnly`.

## 2.3 Scalar normalization

```text
Uuid
  input RFC UUID text; output canonical lowercase 8-4-4-4-12

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
  '-' forbidden because Product owns numbering separator

CodeToken
  ^[A-Z0-9]+$, length 1..32

DocumentCode
  ^[A-Z0-9]+(?:-[A-Z0-9]+)?-[0-9]{3,}$

OpaqueCursor
  nonblank, maxLength=2048

CsrfToken
  opaque/nonblank, maxLength=512

ProviderSubjectRef
  opaque/nonblank, maxLength=2048
  server-resolvable anti-corruption handle for exact issuer+subject
  unchanged binding returns byte-stable ref
  client never parses it or treats it as Product identity

EmailAddress
  trim surrounding whitespace
  minLength=3; maxLength=254; OpenAPI format=email
  no case-folding/canonicalization, uniqueness or verification claim
  profile/contact metadata only; never authentication or Authorization identity
```

Other human text is nonblank where stated and is bounded by the aggregate JSON ceiling rather than unrelated guessed per-field maxima.

Document `q` comparison:

```text
code  ASCII case-insensitive against canonical code
title NFC normalization + Unicode case folding; diacritics preserved
no accent folding, stemming, fuzzy match, body/OCR/vector search
```

T6 ranking remains:

```text
q present: exact code -> code prefix -> title prefix -> title contains -> code -> document_id
q absent:  code -> document_id
```

## 2.4 Strong ETag / conditional law

ETags are strong opaque server-verifiable validators bound to:

```text
resource identity + exact concurrency domain + owner current version/equivalent
```

They never expose raw DB version/generation.

```text
missing/malformed required condition               -> 400 request.invalid
well-formed wrong-resource/domain/tampered tag      -> 412 precondition.resource_changed
stale same-domain + different desired state         -> 412 precondition.resource_changed
```

Accepted exact-current whole-replacement retry is deterministic:

```text
stale authenticated same-domain tag
+ requested whole representation == exact current representation
-> ordinary success with current representation/current ETag
-> zero mutation / zero version advance / zero duplicate Audit
```

This applies only to whole-replacement current truths. Stale DRAFT PATCH **always** -> `412 precondition.draft_changed`, even if requested values happen to equal current values.

An ETag-protected representation contains only fields governed by that concurrency token plus immutable identifiers; independently mutable display enrichment is excluded.

`PROFILE_REPLACE`:

```text
existing profile           If-Match required
absent profile recreation  If-None-Match:* required
both / neither             400
If-None-Match:* + existing profile -> 412
```

## 2.5 CSRF / durable idempotency

Every unsafe operation requires `X-CSRF-Token`.

Exactly these 10 POST creations additionally require a client-generated UUID `Idempotency-Key`:

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

Key scope:

```text
current user id + canonical operationId + UUID value
```

UUID textual case does not create a distinct key. Fingerprint is over complete validated normalized semantic command fields, not raw HTTP bytes. `createSubmission` additionally fingerprints the authenticated semantic DRAFT precondition token value; CSRF/session transport material is not fingerprint input.

Completed replay path:

```text
current session + CSRF + AuthZ/disclosure rechecked
-> key/fingerprint lookup
-> exact stored success status/body replay
-> historical lifecycle/preconditions NOT re-run
-> no second semantic mutation/Audit
```

`createSubmission` still requires a structurally valid `If-Match` on every request; after completed replay recognition the historical DRAFT-currentness test is not re-evaluated.

Same key + different fingerprint -> `422 validation.idempotency_key_reused`.

Replay bodies contain only stable IDs + tiny closed result enums. Operation-local `ReplaySnapshot` maximum is **2,048 bytes** and excludes erasable profile PII, provider payloads/tokens, title/reason/message/free text, raw request/response bytes, and governed content.

Exact semantic replay window:

```text
completed_at <= now < completed_at + 24h  -> key remains replay-authoritative
now >= completed_at + 24h                 -> key is semantically expired
```

Physical cleanup may lag, but cannot extend semantic validity. After expiry the same UUID may be treated as a new key; clients therefore generate one UUID per logical command and reuse it only for retries. This closes the upstream `24h` candidate into the smallest window that materially exceeds browser/network retry periods without turning replay into long-lived history.

## 2.6 Request-processing precedence

For coexisting failures:

```text
1. route / method
2. raw envelope guards (size/coding/media)
3. ApplicationSession authentication
4. CSRF for unsafe request
5. structural parse/schema/pure normalization
6. current AuthZ/disclosure
7. completed idempotency replay recognition
8. current precondition/lifecycle/referential checks
9. business effect
```

Rate limiting or a dependency failure may terminate earlier only when that protection/dependency prevents reaching the next stage. No precedence may bypass current authorization to disclose a stored replay.

## 2.7 Pagination / complete bounded projections

Potentially unbounded lists use stateless integrity-protected seek cursors.

```text
first page: operation filters + optional limit 1..100 (default20)
next page: cursor + optional limit only
cursor + any repeated operation filter/query other than limit -> 400
invalid/tampered cursor -> 400
```

Cursor authenticates `operationId + normalized filters + ordering + seek position`; current AuthZ is rechecked every page. Limit may change on a subsequent page and is not cursor authority.

Response invariant:

```text
items count <= effective limit
has_more=true  <=> next_cursor non-null
has_more=false <=> next_cursor null
```

No offset, total count, generic sort DSL, server cursor state, or frozen multi-page snapshot.

`document-creation/options` arrays are complete, never silently truncated:

```text
no filters:
  all actor-usable active Areas + DocumentTypes; templates=[]; owner candidates absent

area_id only:
  exact usable Area + actor-usable types; templates=[];
  owner candidates present iff caller has owner.manage for selected scope

document_type_id only:
  exact usable type + actor-usable Areas;
  templates = current eligible/effective/readable templates for that type;
  owner candidates absent until Area is selected

both:
  exact usable Area/type;
  templates for type;
  owner candidates present iff owner.manage in selected scope
```

Explicit absent/non-disclosable filter ->404; semantically inapplicable valid combination ->422. If real scale makes these complete arrays unsustainable, that evidence reopens T6; T8-E does not truncate or add operation79.

## 2.8 Header profiles

Request profiles:

```text
SAFE_READ            MetalDocsSession
UNSAFE_CSRF          session + X-CSRF-Token
IDEMPOTENT_CREATE    session + CSRF + Idempotency-Key
IF_MATCH_MUTATION    session + CSRF + exactly one strong If-Match; no *, weak, list
SUBMISSION_CREATE    session + CSRF + Idempotency-Key + exact DRAFT If-Match
PROFILE_REPLACE      session + CSRF + profile conditional matrix in §2.4
```

Success profiles:

```text
NO_STORE
  Cache-Control:no-store

JSON_NO_STORE
  Content-Type:application/json
  Cache-Control:no-store

JSON_ETAG / JSON_ETAG_MUTATION
  Content-Type:application/json
  ETag:current/resulting strong tag
  Cache-Control:no-store

SESSION_END
  Cache-Control:no-store
  Set-Cookie clears __Host-metaldocs_session:
    empty; Path=/; Max-Age=0; Secure; HttpOnly; SameSite=Lax; Domain absent
```

No baseline `Location`, replay header, permission snapshot, provider header, or generic metadata header.

Problem responses: `application/problem+json` + `Cache-Control:no-store`; 401 adds `WWW-Authenticate: MetalDocsSession`; 405 adds exact `Allow`; 429 may add truthful non-negative delta-seconds `Retry-After` but never fabricates one.

## 2.9 Exact-byte delivery — verify before commit

Semantic byte resources return authenticated application-origin `200` only.

```text
Content-Type:
  docx application/vnd.openxmlformats-officedocument.wordprocessingml.document
  pdf  application/pdf
Content-Length exact raw byte count
Content-Disposition inline; filename="<document_code>-REV<ordinal-min-width-3>.<ext>"
Content-Digest sha-256=:<base64 exact SHA-256 bytes>:
Cache-Control private,no-store,no-transform
Accept-Ranges none
X-Content-Type-Options nosniff
Content-Encoding absent
```

No provider redirect, Range/206, 304, automatic HEAD, or transformation. `Range` -> `400 request.invalid`.

Correctness law:

```text
load semantic descriptor
-> OpenExact
-> verify actual byte count + SHA-256 + format coherence
-> ONLY THEN commit HTTP 200 headers/body
```

The implementation may use a bounded temporary spool or provider-neutral equivalent proof, but it may not stream unverified provider bytes to the client and discover a hash mismatch after the 200 has begun. Existing visible semantic record with missing/corrupt bytes -> `500 internal.content_integrity` with zero success bytes; temporary content-store dependency failure ->503. Provider checksum/ETag may be optimization evidence but never replaces semantic SHA-256 authority.

## 2.10 Direct DRAFT upload — exact-length create-only capability

Accepted flow remains:

```text
allocate -> direct provider PUT create-only -> bodyless completion -> server derives descriptor -> READY
```

Allocation request adds exactly one non-authoritative transport hint:

```text
StartDraftUploadRequest
  expected_size_bytes: integer 1..DOC_RAW_MAX_BYTES
```

This does **not** become `ExactContentDescriptor.size_bytes`. Server first proves `expected_size_bytes <= DOC_RAW_MAX_BYTES`, then calls the already-ratified T8-C seam as:

```text
PresignCreate(handle, maxBytes=expected_size_bytes, ttl=15 minutes)
```

No T8-C signature or authority reopen is required: the current `maxBytes` contract is consumed with the exact intended body size as its maximum, and the Launch direct-PUT provider profile is allowed to be stricter by binding that value as exact HTTP `Content-Length`.

Allocation response:

```text
DraftUploadAllocation
  upload_id:Uuid
  upload_url:URI capability
  expires_at:UtcInstant = allocation_time + 15 minutes
  max_bytes:DOC_RAW_MAX_BYTES
  required_headers:map<string,string>
```

`upload_url` is opaque capability data. Its syntax may necessarily contain provider mechanism material, but no separate provider/account/bucket/key/version/storage-ETag fields are Product contract and the client must not parse the URL for semantic identity.

Capability/admission law:

```text
create-only (`If-None-Match:*` or provider-equivalent)
+ exact HTTP body length == expected_size_bytes
+ one shared expires_at for provider capability and unconsumed admission claim
```

For browser PUT, `Content-Length` is user-agent generated from the exact Blob/body and is part of the signed/provider-enforced request constraint; it is not placed in `required_headers` as a script-settable header. `required_headers` contains only exact browser-settable provider headers.

Completion independently `Stat/OpenExact`s the object, requires actual size == expected size, derives SHA-256/actual format/actual size, performs structural validation, and only then establishes READY. Client-declared type/hash/name is non-authoritative.

Expiry semantics are closed:

```text
now < expires_at + OPEN                        completion may proceed
now < expires_at + READY/unconsumed            exact completion repeat ->204; DRAFT attach may proceed
now >= expires_at + OPEN or READY/unconsumed   410 state.upload_expired; content reclaimable
consumed attachment                            upload claim is no longer an attachable resource
```

This preserves direct PUT and closes max-size resource consumption without S3 POST or multipart. S3 POST `content-length-range` supplies no additional required property once application validates the global maximum and the PUT capability binds one exact body length.

---

# 3. Shared schema registry

All fixed objects are closed. Unless marked `?`, members are required.

## 3.1 References / enums / unions

```text
UserReference { user_id:Uuid, display_name?:string }
AreaReference { area_id:Uuid, code:CodeToken, name:nonblank string }
DocumentTypeReference { document_type_id:Uuid, code:CodeToken, name:nonblank string }
DocumentReference { document_id:Uuid, code:DocumentCode } // title belongs Revision
RevisionIdentity { revision_id:Uuid, ordinal:RevisionOrdinal }
RevisionReference { revision:RevisionIdentity, title:nonblank string }
ContentSummary { sha256:Sha256Hex, size_bytes:ByteCount, content_format:ContentFormat }
```

Closed enums:

```text
ContentFormat               docx | pdf
RevisionState               draft | submitted | effective | superseded | obsolete | cancelled
OpenRevisionState           draft | submitted
UserEligibilityState        enabled | disabled
AreaLifecycleState          active | retired
NumberingScope              document_type | document_type_area
GovernanceMode              no_human_approval | use_governance_route
GovernanceSelectorKind      named_user | group
GovernanceDecisionOutcome   accept | return_for_changes
GovernanceSubjectKind       submission | obsolescence
GovernanceAttemptState      active | completed | returned | withdrawn | cancelled
GovernanceStepState         pending | active | decided
RepresentationKind          source_only | require_official_rendition
OfficialRenditionFormat     pdf
RoleCode                    governance_admin | area_manager | author | approver | viewer | governance_viewer
RoleAssignmentSubjectKind   user | group
RoleAssignmentScopeKind     company | area
DocumentCatalogStatus       effective | obsolete | cancelled
DocumentOfficialStatus      draft | submitted | effective | obsolete | cancelled
SubmissionCreationState     governance_pending | rendition_pending | released
SubmissionTerminationKind  returned_for_changes | withdrawn | revision_cancelled
ObsolescenceRequestState    active | returned | withdrawn | completed
ObsolescenceCreationState   governance_pending | obsolete
GovernanceCaseAction        accept | return_for_changes | add_feedback
```

`PermissionCode` is exactly the accepted 14-value T3 dot-spelled vocabulary.

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
  label:nonblank string
  selector:GovernanceSelector
  // array order is route order; ordinal is not wire data
RepresentationPolicy
  {kind:source_only}
  {kind:require_official_rendition,format:pdf}
```

Pagination: `Page { next_cursor:OpaqueCursor|null, has_more:boolean }`.

## 3.2 Session / Organization

```text
SessionView { user:UserReference, csrf_token:CsrfToken }
ProviderSubjectOption { provider_subject_ref:ProviderSubjectRef, display_hints:string[] }
  display_hints maxItems3; each nonblank <=256
ProviderSubjectSearchView { items:ProviderSubjectOption[] } // maxItems20, provider order

CompanyView { company_id:Uuid, display_name:nonblank string }
ReplaceCompanyRequest { display_name:nonblank string }

UserProfileInput { display_name:nonblank string, email?:EmailAddress }
CreateUserRequest { provider_subject_ref:ProviderSubjectRef, profile:UserProfileInput }
CreateUserResult { user_id:Uuid }
UserView { user:UserReference, eligibility:UserEligibilityState }
UserPage { items:UserView[], page:Page }
UserProfileView { user_id:Uuid, display_name:nonblank string, email?:EmailAddress }
ReplaceUserProfileRequest { display_name:nonblank string, email?:EmailAddress }
UserProviderBindingView { user_id:Uuid, provider_subject_ref:ProviderSubjectRef }
ReplaceUserProviderBindingRequest { provider_subject_ref:ProviderSubjectRef }
UserEligibilityView { user_id:Uuid, state:UserEligibilityState }
ReplaceUserEligibilityRequest { state:UserEligibilityState }

AreaView { area_id:Uuid, code:CodeToken, name:nonblank string }
AreaSummary { area:AreaReference, state:AreaLifecycleState }
AreaPage { items:AreaSummary[], page:Page }
CreateAreaRequest { code:CodeInput, name:nonblank string }
CreateAreaResult { area_id:Uuid }
ReplaceAreaRequest { name:nonblank string }
AreaLifecycleView { area_id:Uuid, state:AreaLifecycleState }
ReplaceAreaLifecycleRequest { state:AreaLifecycleState }

GroupView { group_id:Uuid, name:nonblank string }
GroupPage { items:GroupView[], page:Page }
CreateGroupRequest / ReplaceGroupRequest { name:nonblank string }
CreateGroupResult { group_id:Uuid }
GroupMemberPage { items:UserReference[], page:Page }
```

CreateUser establishes enabled User + required profile + binding atomically. New Area starts active. Area code is immutable and absent from replacement. Profile DELETE removes the profile subresource; already absent ->404 rather than inventing `ensure absent` semantics.

## 3.3 Authorization

Exact role projection:

```text
governance_admin   scopes[company]       organization.manage,access.manage,document_type.manage,template_use.manage
area_manager       scopes[area]          document.read_effective,document.read_history,document.read_working,document.create,document.edit,document.submit,document.cancel_revision,document.obsolete,document.owner.manage,governance.act
author             scopes[company,area]  document.read_effective,document.read_history,document.read_working,document.create,document.edit,document.submit
approver           scopes[company,area]  document.read_effective,governance.act
viewer             scopes[company,area]  document.read_effective
governance_viewer  scopes[company,area]  document.read_effective,document.read_history,audit.read
```

```text
RoleView { code:RoleCode, permissions:unique PermissionCode[], allowed_scope_kinds:unique RoleAssignmentScopeKind[] }
RoleListView { items in order governance_admin,area_manager,author,approver,viewer,governance_viewer }
RoleAssignmentView { assignment_id:Uuid, subject:RoleAssignmentSubject, role:RoleCode, scope:RoleAssignmentScope }
RoleAssignmentPage { items:RoleAssignmentView[], page:Page }
CreateRoleAssignmentRequest { subject:RoleAssignmentSubject, role:RoleCode, scope:RoleAssignmentScope }
CreateRoleAssignmentResult { assignment_id:Uuid }
```

No editable Role/Permission policy API.

## 3.4 Document Governance

```text
DocumentTypeView { document_type_id:Uuid, code:CodeToken, name:nonblank string, numbering_scope:NumberingScope, active:boolean }
DocumentTypePage { items:DocumentTypeView[], page:Page }
CreateDocumentTypeRequest { code:CodeInput, name:nonblank string, numbering_scope:NumberingScope, active:boolean, governance:GovernancePolicy, representation:RepresentationPolicy }
CreateDocumentTypeResult { document_type_id:Uuid }
ReplaceDocumentTypeRequest { code:CodeInput, name:nonblank string, numbering_scope:NumberingScope, active:boolean }
DocumentTypeGovernanceView / ReplaceDocumentTypeGovernanceRequest { governance:GovernancePolicy, representation:RepresentationPolicy }
EligibleTemplatesView { templates:DocumentReference[] } // code,id order; stable refs only
ReplaceEligibleTemplatesRequest { template_document_ids:unique Uuid[] } // empty valid
NumberingPreviewView { preview_code:DocumentCode, reservation:false }
TemplateConfigurationItem { document:DocumentReference, template_role:boolean, has_effective_revision:boolean, current_effective_title?:nonblank string, eligible_document_type_ids:unique Uuid[] }
TemplateConfigurationPage { items:TemplateConfigurationItem[], page:Page }
```

Create explicitly supplies initial governance/representation because T8-D requires current values and no accepted default exists. Eligible-template set starts empty. Base replacement rejects normalized code/numbering-scope change after first committed Document with409.

## 3.5 Controlled Documents — create/read/work/upload

```text
TemplateCreationOption { document:DocumentReference, effective_revision:RevisionReference }
DocumentCreationOptionsView { areas:AreaReference[], document_types:DocumentTypeReference[], templates:TemplateCreationOption[], default_responsible_owner:UserReference, responsible_owner_candidates?:UserReference[] }
CreateDocumentRequest { document_type_id:Uuid, area_id:Uuid, title:nonblank string, template_document_id?:Uuid, responsible_owner_user_id?:Uuid }
CreateDocumentResult { document_id:Uuid, revision_id:Uuid }
DocumentSummary { document:DocumentReference, document_type:DocumentTypeReference, area:AreaReference, responsible_owner:UserReference, status:DocumentCatalogStatus, official_revision?:RevisionReference }
DocumentPage { items:DocumentSummary[], page:Page }
ReleasedRevisionView { revision:RevisionIdentity, title:nonblank string, release_id:Uuid, released_at:UtcInstant, source:ContentSummary, representation:{kind:source_only}|{kind:official_rendition,official_rendition_id:Uuid,content:ContentSummary} }
DocumentOfficialView { document:DocumentReference, document_type:DocumentTypeReference, area:AreaReference, responsible_owner:UserReference, status:DocumentOfficialStatus, official?:ReleasedRevisionView }
```

`official` is present iff at least one Release exists; obsolete retains last released official. Newer cancelled/open work never replaces older EFFECTIVE official truth. Before first Release, status may be draft/submitted/cancelled and `official` is absent.

```text
ResponsibleOwnerView { document_id:Uuid, responsible_owner_user_id:Uuid }
ReplaceResponsibleOwnerRequest { user_id:Uuid }
TemplateRoleView { document_id:Uuid, is_template:boolean }
ReplaceTemplateRoleRequest { is_template:boolean }
CreateRevisionResult { revision_id:Uuid }
RevisionView { revision:RevisionIdentity, document:DocumentReference, title:nonblank string, state:RevisionState, created_at:UtcInstant, current_submission_id?:Uuid }
DocumentWorkView { document:DocumentReference, revision:RevisionIdentity, title:nonblank string, content:ContentSummary, updated_at:UtcInstant }
UpdateDraftRequest { title?:nonblank string, source_upload_id?:Uuid } // minProperties1; null forbidden; omitted unchanged
StartDraftUploadRequest { expected_size_bytes:integer minimum1 maximum DOC_RAW_MAX_BYTES }
DraftUploadAllocation { upload_id:Uuid, upload_url:URI, expires_at:UtcInstant, max_bytes:ByteCount, required_headers:map<string,string> }
```

Raw WorkingContent generation is never public; ETag is wire OCC authority.

## 3.6 Submission / representation gate

```text
SubmissionCreateResult
  {state:governance_pending,submission_id:Uuid,governance_attempt_id:Uuid}
  {state:rendition_pending,submission_id:Uuid}
  {state:released,submission_id:Uuid,release_id:Uuid}
SubmissionHumanGate { required:boolean, satisfied:boolean }
SubmissionRepresentationGate { required:boolean, satisfied:boolean, attention_required:boolean }
SubmissionView { submission_id:Uuid, revision:RevisionIdentity, title:nonblank string, submitter:UserReference, submitted_at:UtcInstant, content:ContentSummary, governance_mode:GovernanceMode, representation:RepresentationPolicy, human_gate:SubmissionHumanGate, representation_gate:SubmissionRepresentationGate, governance_attempt_id?:Uuid, release_id?:Uuid, termination?:SubmissionTerminationKind }
```

Cross-field law:

```text
governance_attempt_id iff governance route
human_gate.required iff governance route; not-required => satisfied
representation_gate.required iff require_official_rendition; not-required => satisfied
attention_required only while rendition required + unsatisfied + terminal renderer attention exists
release_id and termination mutually exclusive
```

Representation execution:

```text
SourceOnly + DOCX/PDF
  representation gate satisfied by absence

RequireOfficialRendition(PDF) + submitted PDF
  establish OfficialRendition semantic fact over the exact same already-admitted PDF handle/descriptor
  no physical duplicate, renderer job, or provider copy
  gate satisfied synchronously

RequireOfficialRendition(PDF) + submitted DOCX
  durable renderer intent
  gate pending until exact admitted OfficialRendition PDF exists
```

Creation result:

```text
human governance required                  governance_pending
no human + DOCX rendition pending          rendition_pending
all gates synchronously satisfied          released
```

Human-governed DOCX may render concurrently in mechanism space; initial result remains `governance_pending` while `SubmissionView` exposes orthogonal gates.

```text
SubmissionWithdrawalView { submission_id:Uuid, actor:UserReference, withdrawn_at:UtcInstant }
RevisionCancellationRequest { reason:nonblank string }
RevisionCancellationView { revision_id:Uuid, actor:UserReference, reason:nonblank string, cancelled_at:UtcInstant }
```

Cancellation singleton: first201; exact same reason repeat200; different later reason409.

## 3.7 Governance / history / work

```text
GovernanceFeedbackView { feedback_id:Uuid, actor:UserReference, message:nonblank string, created_at:UtcInstant }
GovernanceFeedbackPage { items:GovernanceFeedbackView[], page:Page }
CreateGovernanceFeedbackRequest { message:nonblank string }
CreateGovernanceFeedbackResult { feedback_id:Uuid }
GovernanceDecisionRequest {outcome:accept} | {outcome:return_for_changes,reason:nonblank string}
GovernanceDecisionView {decision_id:Uuid,outcome:accept,actor:UserReference,decided_at:UtcInstant} | {decision_id:Uuid,outcome:return_for_changes,actor:UserReference,decided_at:UtcInstant,reason:nonblank string}
GovernanceStepView {step_id:Uuid,label:nonblank string,state:pending} | {step_id:Uuid,label:nonblank string,state:active} | {step_id:Uuid,label:nonblank string,state:decided,decision:GovernanceDecisionView}
```

Step array order is frozen route order; persistence ordinal/candidate snapshots/Group membership/grants/provider claims are not exposed.

```text
SubmissionGovernanceSubject { kind:submission, submission_id:Uuid, document:DocumentReference, revision:RevisionIdentity, title:nonblank string, submitter:UserReference, submitted_at:UtcInstant, content:ContentSummary }
ObsolescenceGovernanceSubject { kind:obsolescence, request_id:Uuid, document:DocumentReference, target_revision:RevisionReference, initiator:UserReference, reason:nonblank string, requested_at:UtcInstant }
GovernanceCaseView { governance_attempt_id:Uuid, state:GovernanceAttemptState, subject:SubmissionGovernanceSubject|ObsolescenceGovernanceSubject, steps:GovernanceStepView[], feedback:GovernanceFeedbackPage, allowed_actions:unique GovernanceCaseAction[] }
```

Embedded feedback is first20 and any cursor targets `listGovernanceFeedback`. `allowed_actions` canonical order is `accept, return_for_changes, add_feedback`, filtered from the same current T3 + Controlled Documents decisions used by commands; may be empty.

Decision singleton: first201; exact same outcome+reason repeat200; any later different outcome/reason ->409 `state.governance_step_already_decided`.

`DocumentHistoryItem` closed kinds:

```text
revision_created             revision,title,occurred_at
submission_created           submission_id,revision,title,submitter,occurred_at,?governance_attempt_id
governance_decision          decision_id,governance_attempt_id,step_id,actor,outcome,occurred_at,?reason
feedback_added               feedback_id,governance_attempt_id,actor,message,occurred_at
submission_withdrawn         submission_id,actor,occurred_at
revision_cancelled           revision_id,actor,reason,occurred_at
release_completed            release_id,revision_id,submission_id,occurred_at,?predecessor_revision_id
official_rendition_completed official_rendition_id,submission_id,occurred_at
obsolescence_requested       request_id,target_revision_id,initiator,reason,occurred_at,?governance_attempt_id
obsolescence_withdrawn       request_id,actor,occurred_at
obsolescence_completed       request_id,target_revision_id,occurred_at
```

```text
DocumentHistoryPage {items:DocumentHistoryItem[],page:Page}
WorkAuthoringItem {document,revision,title,state:OpenRevisionState,responsible_owner,updated_at}
WorkAuthoringPage {items:WorkAuthoringItem[],page:Page}
WorkGovernanceItem {governance_attempt_id,subject_kind,document,created_at}
WorkGovernancePage {items:WorkGovernanceItem[],page:Page}
```

## 3.8 Release / obsolescence

```text
ReleaseView
  source-only:
    release_id,document,revision,title,submission_id,released_at,?predecessor_revision_id,
    representation:{kind:source_only,source:ContentSummary}
  official-rendition:
    same core,
    representation:{kind:official_rendition,source:ContentSummary,official_rendition_id:Uuid,official_rendition:ContentSummary}
ObsolescenceRequestCreateRequest { reason:nonblank string }
ObsolescenceCreateResult {state:governance_pending,request_id:Uuid,governance_attempt_id:Uuid} | {state:obsolete,request_id:Uuid}
```

`ObsolescenceRequestView` state union:

```text
active
  request_id,document,target_revision,initiator,reason,state=active,requested_at,governance_attempt_id
returned / withdrawn / completed-human
  same core + matching state + governance_attempt_id + ended_at
completed-no-human
  same core + state=completed + ended_at; governance_attempt_id absent
```

`ObsolescenceWithdrawalView { request_id:Uuid, actor:UserReference, withdrawn_at:UtcInstant }`.

Release remains immutable/system-owned; no publish mutation.

---

# 4. Audit wire

Audit projects T3 evidence only; raw JSONB facts and operational correlation metadata are not exposed.

```text
AuditActor {kind:user,user_id:Uuid} | {kind:system,system_actor_code:metaldocs}
AuditVisibility {kind:company} | {kind:area,area_id:Uuid}
```

Reachable Launch `AuditOperationCode` set:

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

T3 bounded precision is reconciled: there is no `provider_binding.disabled` Launch transition or Audit operation. `document.created` is atomic Document+REV000; `revision.created` means later Revision only.

Resource kinds:

```text
provider_binding | user | user_profile | area | group | role_assignment |
document_type | document | revision | submission | governance_decision |
official_rendition | release | obsolescence_request
```

GroupMembership has no invented UUID/resource kind: membership events use `resource_kind=group`, `resource_id=group_id`.

Typed wire facts only when operation/resource identity is insufficient:

```text
GroupMembershipAuditFacts { user_id:Uuid }
RoleAssignmentAuditFacts { subject:RoleAssignmentSubject, role:RoleCode, scope:RoleAssignmentScope }
GovernanceDecisionAuditFacts { governance_attempt_id:Uuid, step_id:Uuid, subject_kind:GovernanceSubjectKind, subject_id:Uuid, outcome:GovernanceDecisionOutcome }
ReleaseAuditFacts { document_id:Uuid, revision_id:Uuid, submission_id:Uuid, predecessor_revision_id?:Uuid }
RevisionCancellationAuditFacts { document_id:Uuid, revision_id:Uuid }
ObsolescenceAuditFacts { document_id:Uuid, target_revision_id:Uuid }
```

`resource_id` supplies group/assignment/decision/release/obsolescence evidence identity; duplicate IDs are not repeated inside facts.

`AuditEventView` is a closed `operation_code`-discriminated union with common `event_id,occurred_at,actor,operation_code,resource_kind,resource_id,visibility`. Simple branches have no `facts`; typed branches require the matching facts. No free-form feedback/reason/profile/provider payload.

`AuditEventPage={items:AuditEventView[],page:Page}` ordered `occurred_at DESC,event_id DESC`. `GET /audit/events` accepts only cursor/limit; `audit.read` historical visibility filtering occurs before pagination. No inferred Audit filter.

---

# 5. RFC 9457 Problem contract

Each code is its own full closed schema; no open inherited base and no `default` response.

Required: `type,title,status,detail,instance,code,trace_id`. `instance=urn:uuid:<fresh UUID>` per occurrence. `trace_id` opaque/nonblank. Optional non-empty `errors[]` is allowed only on `request.invalid` and validation-family variants; each error uses an RFC6901 pointer rooted at `/path`, `/query`, `/header`, or `/body` and never echoes sensitive rejected values.

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

`type=https://errors.conexus.fun/metaldocs/{code}`.

Disclosure:

```text
no valid session                                      401
visible request/action but missing permission/trust   403
absent or non-disclosable item                        404
```

Default/effective collections are AuthZ-filtered and may be empty. `listDocuments status=obsolete|cancelled` requires current `document.read_history` in at least one relevant scope; otherwise403.

Textual exact-set macros for ledger only; final OAS expands them:

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

All path `*_id` parameters are required `Uuid`. `PAGED` means §2.7. JSON request rows inherit the 65,536-byte ceiling and `J` where shown.

## 6.1 Operations 1→43

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
|11|`deleteUserProfile`|`DELETE /api/v1/users/{user_id}/profile`|no body / `UNSAFE_CSRF`|`204` first delete|`NO_STORE`|absent profile->404|`U + N`|
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
|26|`deleteGroup`|`DELETE /api/v1/groups/{group_id}`|no body / `UNSAFE_CSRF`|`204` first delete|`NO_STORE`|absent->404; dependency->409|`U + N + S`|
|27|`listGroupMembers`|`GET /api/v1/groups/{group_id}/members`|`SAFE_READ`|`200 GroupMemberPage`|`JSON_NO_STORE`|`PAGED`; user_id ASC|`A + N`|
|28|`addGroupMember`|`PUT /api/v1/groups/{group_id}/members/{user_id}`|no body / `UNSAFE_CSRF`|`201` first; `204` existing relation|`NO_STORE`|none|`U + N + S`|
|29|`removeGroupMember`|`DELETE /api/v1/groups/{group_id}/members/{user_id}`|no body / `UNSAFE_CSRF`|`204`, including absent relation when parent exists|`NO_STORE`|absent/non-disclosable Group->404|`U + N`|
|30|`listRoles`|`GET /api/v1/roles`|`SAFE_READ`|`200 RoleListView`|`JSON_NO_STORE`|fixed T3 role order|`A`|
|31|`listRoleAssignments`|`GET /api/v1/role-assignments`|`SAFE_READ`|`200 RoleAssignmentPage`|`JSON_NO_STORE`|`PAGED`; assignment_id ASC|`A`|
|32|`createRoleAssignment`|`POST /api/v1/role-assignments`|`CreateRoleAssignmentRequest` / `IDEMPOTENT_CREATE`|`201 CreateRoleAssignmentResult`|`JSON_NO_STORE`|none|`U + J + I + S`|
|33|`deleteRoleAssignment`|`DELETE /api/v1/role-assignments/{assignment_id}`|no body / `UNSAFE_CSRF`|`204` first revoke|`NO_STORE`|absent->404|`U + N`|
|34|`listDocumentTypes`|`GET /api/v1/document-types`|`SAFE_READ`|`200 DocumentTypePage`|`JSON_NO_STORE`|`PAGED`; document_type_id ASC|`A`|
|35|`createDocumentType`|`POST /api/v1/document-types`|`CreateDocumentTypeRequest` / `IDEMPOTENT_CREATE`|`201 CreateDocumentTypeResult`|`JSON_NO_STORE`|none|`U + J + I + S`|
|36|`getDocumentType`|`GET /api/v1/document-types/{document_type_id}`|`SAFE_READ`|`200 DocumentTypeView`|`JSON_ETAG`|none|`A + N`|
|37|`replaceDocumentType`|`PUT /api/v1/document-types/{document_type_id}`|`ReplaceDocumentTypeRequest` / `IF_MATCH_MUTATION`|`200 DocumentTypeView`|`JSON_ETAG_MUTATION`|none|`U + J + N + P + S`|
|38|`getDocumentTypeGovernance`|`GET /api/v1/document-types/{document_type_id}/governance`|`SAFE_READ`|`200 DocumentTypeGovernanceView`|`JSON_ETAG`|none|`A + N`|
|39|`replaceDocumentTypeGovernance`|`PUT /api/v1/document-types/{document_type_id}/governance`|`ReplaceDocumentTypeGovernanceRequest` / `IF_MATCH_MUTATION`|`200 DocumentTypeGovernanceView`|`JSON_ETAG_MUTATION`|none|`U + J + N + P + S`|
|40|`getDocumentTypeEligibleTemplates`|`GET /api/v1/document-types/{document_type_id}/eligible-templates`|`SAFE_READ`|`200 EligibleTemplatesView`|`JSON_ETAG`|document.code,document_id|`A + N`|
|41|`replaceDocumentTypeEligibleTemplates`|`PUT /api/v1/document-types/{document_type_id}/eligible-templates`|`ReplaceEligibleTemplatesRequest` / `IF_MATCH_MUTATION`|`200 EligibleTemplatesView`|`JSON_ETAG_MUTATION`|none|`U + J + N + P + S`|
|42|`getDocumentTypeNumberingPreview`|`GET /api/v1/document-types/{document_type_id}/numbering-preview`|`SAFE_READ`|`200 NumberingPreviewView`|`JSON_NO_STORE`|optional area_id|`A + N + validation.failed`|
|43|`listTemplateConfigurations`|`GET /api/v1/document-governance/templates`|`SAFE_READ`|`200 TemplateConfigurationPage`|`JSON_NO_STORE`|`PAGED`; document.code,document_id|`A`|

## 6.2 Operations 44→77

|#|operationId|Method + path|Request/profile|Success|Headers|Query/order|Problems|
|---:|---|---|---|---|---|---|---|
|44|`getDocumentCreationOptions`|`GET /api/v1/document-creation/options`|`SAFE_READ`|`200 DocumentCreationOptionsView`|`JSON_NO_STORE`|optional area_id,document_type_id; §2.7|`A + N + validation.failed`|
|45|`listDocuments`|`GET /api/v1/documents`|`SAFE_READ`|`200 DocumentPage`|`JSON_NO_STORE`|first page q,document_type_id,area_id,responsible_owner_user_id,status,limit; §2.3/2.7|`A + validation.failed`|
|46|`createDocument`|`POST /api/v1/documents`|`CreateDocumentRequest` / `IDEMPOTENT_CREATE`|`201 CreateDocumentResult`|`JSON_NO_STORE`|none|`U + J + I + S + X`|
|47|`getDocument`|`GET /api/v1/documents/{document_id}`|`SAFE_READ`|`200 DocumentOfficialView`|`JSON_NO_STORE`|none|`B + N`|
|48|`getDocumentResponsibleOwner`|`GET /api/v1/documents/{document_id}/responsible-owner`|`SAFE_READ`|`200 ResponsibleOwnerView`|`JSON_ETAG`|none|`A + N`|
|49|`replaceDocumentResponsibleOwner`|`PUT /api/v1/documents/{document_id}/responsible-owner`|`ReplaceResponsibleOwnerRequest` / `IF_MATCH_MUTATION`|`200 ResponsibleOwnerView`|`JSON_ETAG_MUTATION`|none|`U + J + N + P + S`|
|50|`getDocumentTemplateRole`|`GET /api/v1/documents/{document_id}/template-role`|`SAFE_READ`|`200 TemplateRoleView`|`JSON_ETAG`|none|`A + N`|
|51|`replaceDocumentTemplateRole`|`PUT /api/v1/documents/{document_id}/template-role`|`ReplaceTemplateRoleRequest` / `IF_MATCH_MUTATION`|`200 TemplateRoleView`|`JSON_ETAG_MUTATION`|none|`U + J + N + P + S`|
|52|`createDocumentRevision`|`POST /api/v1/documents/{document_id}/revisions`|no body / `IDEMPOTENT_CREATE`|`201 CreateRevisionResult`|`JSON_NO_STORE`|none|`U + N + I + S + X`|
|53|`getDocumentHistory`|`GET /api/v1/documents/{document_id}/history`|`SAFE_READ`|`200 DocumentHistoryPage`|`JSON_NO_STORE`|`PAGED`; occurred_at ASC,kind,semantic id|`A + N`|
|54|`listAuthoringWork`|`GET /api/v1/work/authoring`|`SAFE_READ`|`200 WorkAuthoringPage`|`JSON_NO_STORE`|`PAGED`; document.code,document_id|`B`|
|55|`listGovernanceWork`|`GET /api/v1/work/governance`|`SAFE_READ`|`200 WorkGovernancePage`|`JSON_NO_STORE`|`PAGED`; governance_attempt_id|`B`|
|56|`getRevision`|`GET /api/v1/revisions/{revision_id}`|`SAFE_READ`|`200 RevisionView`|`JSON_NO_STORE`|none|`A + N`|
|57|`getRevisionDraft`|`GET /api/v1/revisions/{revision_id}/draft`|`SAFE_READ`|`200 DocumentWorkView`|`JSON_ETAG`|none|`A + N`|
|58|`updateRevisionDraft`|`PATCH /api/v1/revisions/{revision_id}/draft`|`UpdateDraftRequest` / `IF_MATCH_MUTATION`|`200 DocumentWorkView`|`JSON_ETAG_MUTATION`|none|`U + J + N + D + S + state.upload_expired`|
|59|`startRevisionDraftUpload`|`POST /api/v1/revisions/{revision_id}/draft/uploads`|`StartDraftUploadRequest` / `UNSAFE_CSRF`|`201 DraftUploadAllocation`|`JSON_NO_STORE`|none|`U + J + N + S`|
|60|`completeRevisionDraftUpload`|`POST /api/v1/revisions/{revision_id}/draft/uploads/{upload_id}/complete`|no body / `UNSAFE_CSRF`|`204` live READY repeat|`NO_STORE`|none|`U + N + S + state.upload_expired + validation.content_invalid`|
|61|`getRevisionDraftSource`|`GET /api/v1/revisions/{revision_id}/draft/source`|`SAFE_READ`|`200 exact bytes`|`EXACT_BYTES`|none|`A + N + X`|
|62|`createSubmission`|`POST /api/v1/revisions/{revision_id}/submissions`|no body / `SUBMISSION_CREATE`|`201 SubmissionCreateResult`|`JSON_NO_STORE`|none|`U + N + I + D + S + X + validation.failed + validation.content_malicious + dependency.malware_inspector_unavailable`|
|63|`getSubmission`|`GET /api/v1/submissions/{submission_id}`|`SAFE_READ`|`200 SubmissionView`|`JSON_NO_STORE`|none|`A + N`|
|64|`getSubmissionSource`|`GET /api/v1/submissions/{submission_id}/source`|`SAFE_READ`|`200 exact bytes`|`EXACT_BYTES`|none|`A + N + X`|
|65|`withdrawSubmission`|`PUT /api/v1/submissions/{submission_id}/withdrawal`|no body / `UNSAFE_CSRF`|`201 SubmissionWithdrawalView` first; `200` exact repeat|`JSON_NO_STORE`|none|`U + N + S`|
|66|`cancelRevision`|`PUT /api/v1/revisions/{revision_id}/cancellation`|`RevisionCancellationRequest` / `UNSAFE_CSRF`|`201 RevisionCancellationView` first; `200` exact repeat|`JSON_NO_STORE`|none|`U + J + N + S`|
|67|`getGovernanceAttempt`|`GET /api/v1/governance-attempts/{attempt_id}`|`SAFE_READ`|`200 GovernanceCaseView`|`JSON_NO_STORE`|embedded first feedback page; route-order steps|`A + N`|
|68|`listGovernanceFeedback`|`GET /api/v1/governance-attempts/{attempt_id}/feedback`|`SAFE_READ`|`200 GovernanceFeedbackPage`|`JSON_NO_STORE`|`PAGED`; created_at,feedback_id|`A + N`|
|69|`createGovernanceFeedback`|`POST /api/v1/governance-attempts/{attempt_id}/feedback`|`CreateGovernanceFeedbackRequest` / `IDEMPOTENT_CREATE`|`201 CreateGovernanceFeedbackResult`|`JSON_NO_STORE`|none|`U + J + N + I + S`|
|70|`getGovernanceStepDecision`|`GET /api/v1/governance-attempts/{attempt_id}/steps/{step_id}/decision`|`SAFE_READ`|`200 GovernanceDecisionView`|`JSON_NO_STORE`|none|`A + N`|
|71|`recordGovernanceStepDecision`|`PUT /api/v1/governance-attempts/{attempt_id}/steps/{step_id}/decision`|`GovernanceDecisionRequest` / `UNSAFE_CSRF`|`201 GovernanceDecisionView` first; `200` exact repeat|`JSON_NO_STORE`|none|`U + J + N + S + state.governance_step_already_decided`|
|72|`getRelease`|`GET /api/v1/releases/{release_id}`|`SAFE_READ`|`200 ReleaseView`|`JSON_NO_STORE`|none|`A + N`|
|73|`getReleaseSource`|`GET /api/v1/releases/{release_id}/source`|`SAFE_READ`|`200 exact bytes`|`EXACT_BYTES`|none|`A + N + X`|
|74|`getOfficialRenditionContent`|`GET /api/v1/official-renditions/{rendition_id}/content`|`SAFE_READ`|`200 exact PDF bytes`|`EXACT_BYTES`|none|`A + N + X`|
|75|`createObsolescenceRequest`|`POST /api/v1/documents/{document_id}/obsolescence-requests`|`ObsolescenceRequestCreateRequest` / `IDEMPOTENT_CREATE`|`201 ObsolescenceCreateResult`|`JSON_NO_STORE`|none|`U + J + N + I + S`|
|76|`getObsolescenceRequest`|`GET /api/v1/obsolescence-requests/{request_id}`|`SAFE_READ`|`200 ObsolescenceRequestView`|`JSON_NO_STORE`|none|`A + N`|
|77|`withdrawObsolescenceRequest`|`PUT /api/v1/obsolescence-requests/{request_id}/withdrawal`|no body / `UNSAFE_CSRF`|`201 ObsolescenceWithdrawalView` first; `200` exact repeat|`JSON_NO_STORE`|none|`U + N + S`|

Row59 is fully expressible through existing T8-C after §2.10 consumer precision.

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

# 7. Document admission limits — measured Launch candidate

The Launch candidate freezes exactly three resource ceilings:

```text
DOC_RAW_MAX_BYTES        = 104857600  // 100 MiB; DOCX and PDF
DOCX_EXPANDED_MAX_BYTES  = 268435456  // 256 MiB; streamed top-level OPC expansion
DOCX_MAX_ZIP_ENTRIES     = 4096       // top-level ZIP entries
```

These are application admission limits, not claims about the theoretical maximum accepted by Microsoft Word, S3, a DAM, or another provider.

Measured real corpus supplied during T8-E:

```text
ForgeFlow_Arquitetura_Base_v01.docx
  raw bytes              22,863
  ZIP entries            24
  file entries           20
  expanded bytes         284,172
  embedded media entries 0
  expanded/raw           12.43x

PO-05-04 Projeto e Desenvolvimento.pdf
  raw bytes              445,131
  pages                  11
  encrypted              no
```

The largest measured real sample therefore sits below the candidate ceilings by approximately:

```text
raw bytes       235x
DOCX expansion  944x
ZIP entries     170x
```

The deliberately large headroom is necessary because the supplied DOCX contains tables/formatting but no embedded media, while ordinary future controlled documents may contain images, headers/footers, page/section breaks and richer OOXML parts. The limits are still materially below generic DAM/large-media scales.

Adversarial disposable probes demonstrate that each control protects a different resource:

```text
expanded_above_256m.docx
  raw       306,139 bytes
  expanded  314,572,846 bytes
  -> reject by DOCX_EXPANDED_MAX_BYTES

many_entries.docx
  raw       628,056 bytes
  entries   5,002
  expanded  320,040 bytes
  -> reject by DOCX_MAX_ZIP_ENTRIES

duplicate_parts.docx
  duplicate canonical ZIP part name
  -> reject structurally

traversal.docx
  parent-traversal ZIP path
  -> reject structurally

expanded_bomb.docx
  raw       130,838 bytes
  expanded  134,217,774 bytes
  -> high compression ratio alone is NOT invalid because actual resource use remains inside the explicit expanded-byte budget
```

Industry comparables are sanity bounds only, never Product authority: controlled-document products commonly admit individual files around the 100 MB class, while DAM/large-media products allow materially larger files. MetalDocs therefore keeps a conservative controlled-document Launch ceiling rather than importing DAM-scale behavior.

Closed structural laws:

```text
DOCX = valid top-level OOXML/OPC ZIP with WordprocessingML main document
duplicate canonical ZIP part names rejected
absolute/parent-traversal paths rejected
symlink entry extraction rejected
no recursive expansion of embedded archives
stream expansion while enforcing cumulative expanded-byte + entry-count ceilings
encrypted/password-protected or macro-enabled/non-DOCX Office packages rejected as DOCX
PDF = structurally parseable PDF; encrypted/password-protected PDF rejected at Launch
client filename/Content-Type never decides actual ContentFormat
```

Validation does not recursively unpack arbitrary embedded archives, so application archive depth is exactly one. No generic nested-archive framework or compression-ratio threshold is added. Malware inspection remains a separate exact-byte governed-boundary control.

Boundary behavior is exact:

```text
expected_size_bytes > DOC_RAW_MAX_BYTES
  -> allocation rejected before provider capability exists

actual provider bytes != expected_size_bytes
  -> completion rejected; READY not established

actual DOCX expanded bytes > DOCX_EXPANDED_MAX_BYTES
  -> 422 validation.content_invalid

actual DOCX ZIP entries > DOCX_MAX_ZIP_ENTRIES
  -> 422 validation.content_invalid

structural path/duplicate/encryption/package violation
  -> 422 validation.content_invalid
```

`DraftUploadAllocation.max_bytes = DOC_RAW_MAX_BYTES`; `StartDraftUploadRequest.expected_size_bytes <= DOC_RAW_MAX_BYTES`.

Multipart, recursive archive inspection, compression-ratio thresholds and DAM-scale upload machinery remain absent. A later measured ordinary controlled document that cannot fit these ceilings is a bounded admission-limit reopen, not permission to raise limits silently.

---

# 8. Bounded upstream findings exposed by T8-E — RESOLVED

Both findings below were evidence-triggered bounded corrections, explicitly operator-approved on 2026-08-20 and reconciled in their owning durable authorities. Neither reopened Product, lifecycle, ownership, topology, or the 78-operation census.

## 8.1 T8-D — Governance Step label preservation — RESOLVED

Product/T6 requires human-facing configured Step labels and GovernanceAttempt freezes exact route snapshot. T8-D originally persisted current/frozen steps without any label field, allowing a later relabel to rewrite historical governed context.

Operator-approved correction now durable in T8-D:

```text
controlled_docs.document_type_governance_steps
  + label TEXT NOT NULL
controlled_docs.governance_attempt_steps
  + label_snapshot TEXT NOT NULL
```

Attempt creation copies immutable `label_snapshot`. Zero new tables/owners/lifecycles/selectors/candidate APIs/workflow platform/operations.

## 8.2 T3 — unreachable ProviderSubjectBinding disabled Audit event — RESOLVED

T3 originally named `ProviderSubjectBinding accepted / disabled / replaced`, while Product/T6 has only creation + replacement and T8-D explicitly preserves the current binding on User offboarding.

Operator-approved bounded precision now durable in T3:

```text
ProviderSubjectBinding accepted / replaced
```

No Product operation, persistence field, replacement behavior, or offboarding behavior changed. The impossible Audit event was removed rather than manufacturing a binding-disable transition.

## 8.3 Subtractive result — no T8-C reopen for direct PUT

The Lead initially treated exact PUT length as a possible T8-C signature gap. Structural inversion removed that reopen:

```text
current seam: PresignCreate(handle,maxBytes,ttl)
consumer call: maxBytes = expected_size_bytes
provider PUT profile: exact Content-Length = maxBytes
```

A capability that allows exactly `maxBytes` bytes is a valid stricter realization of an at-most-`maxBytes` contract. No new parameter or durable authority is needed.

---

# 9. Generation / provider feasibility and runtime conformance proof

Disposable probe pins are evidence only; they do not pre-authorize T8-G runtime/toolchain choices:

```text
Go          oapi-codegen v2.8.0 strict-server
TypeScript  openapi-typescript 7.13.0 paths/components
S3 probe    AWS SDK for Go v2 service/s3 v1.106.2
```

## 9.1 Generated boundary feasibility — PASS

A disposable OpenAPI 3.0.3 probe exercised the actual T8-E encoding patterns rather than a scalar-only toy schema:

```text
additionalProperties:false
required nullable member
optional non-nullable member
closed string enum
safe-integer maximum 9007199254740991
oneOf + discriminator union
multiple success responses 200 + 201
operation-specific RFC9457 response schemas
strict Go server response objects
TypeScript paths/components
```

Execution evidence:

```text
oapi-codegen v2.8.0
  generator built successfully
  runner Go 1.24.13 automatically acquired Go 1.26.7 because generator requires Go >=1.25
  generated Go compiled
  generated Go tests PASS

openapi-typescript 7.13.0
  generated declarations successfully
  TypeScript 5.9.2 strict noEmit probe PASS
```

Observed generated properties:

```text
Go
  fixed objects gained no AdditionalProperties field
  required nullable stayed non-omitempty
  optional member stayed omitempty
  required nullable Page.next_cursor serialized as explicit null
  strict response set contained distinct 200 and 201 JSON response objects

TypeScript
  closed enum -> "active" | "retired"
  next_cursor -> string | null
  required_nullable -> required string | null
  optional_nonnullable -> optional string
  200 and 201 response keys both present
  negative compile probes rejected undeclared enum, missing required nullable,
    extra fixed-object member and unknown discriminator
```

`oapi-codegen` represents `oneOf` internally with private `json.RawMessage` union storage plus typed conversion helpers. That is generated encoding mechanism, not a public `any`/map, provider identifier or second DTO authority.

The generator's build-time Go requirement is **not** a MetalDocs runtime Go-version decision; T8-G remains free to select the compatible runtime/toolchain floor.

## 9.2 Direct S3 PUT constraint feasibility — PASS

A disposable AWS SDK Go v2 presign probe used:

```text
PutObjectInput:
  ContentLength = 12345
  IfNoneMatch   = "*"
```

and produced:

```text
Content-Length=["12345"]
If-None-Match=["*"]
X-Amz-SignedHeaders=content-length;host;if-none-match
```

Therefore the concrete reference provider can bind both exact body length and create-only precondition in the signed PUT request without POST-form/multipart.

Browser feasibility also closes the public contract:

```text
If-None-Match   browser-settable request header; not Fetch-forbidden
Content-Length  script-forbidden but user-agent generated from the fixed Blob/File body
cross-origin PUT uses normal provider CORS preflight for returned browser-settable headers
```

Exact provider CORS configuration belongs T8-G. T8-E requires only that `required_headers` be returned and applied verbatim and that the upload body have known fixed length.

## 9.3 Closed ledger fixture proof — PASS

A mechanical Lead proof compared the candidate ledger against the canonical Product/T6 census and special profiles:

```text
ledger rows                         78
row numbers                         exact 1..78
method+path census                  exact match; zero missing/extra
operationId                         78 unique
family partition                    3 / 26 / 4 / 10 / 34 / 1
Idempotency-Key POST creations      exact accepted 10
ETag concurrency domains            13 GET/mutation domains
exact-byte application resources    exact accepted 4
operation 79                         absent
```

The proof also checks that unsafe application operations use a CSRF-bearing request profile and that JSON-body operations carry the closed structural/media/size validation family where applicable.

The final OpenAPI must expand proposal macros (`A`, `B`, `U`, `J`, etc.) into explicit per-operation response schemas; macros never survive as executable ambiguity.

## 9.4 Runtime conformance contract

T8-E closes the proof architecture; it does not fabricate an application runtime while implementation is blocked:

```text
raw HTTP
-> route/raw envelope limit
-> ApplicationSession
-> CSRF for unsafe request
-> central OpenAPI + strict request validation
-> generated typed request boundary
-> semantic application
-> generated typed response boundary
-> HTTP
-> contract fixture validates exact status + headers + body/Problem
```

Required negative/edge fixture classes:

```text
all 78 rows and no 79th
unknown path ->404; undeclared method ->405 exact Allow
unknown JSON/query member and duplicate scalar/member rejection
bodyless operation rejects a body
wrong media/content coding and 65,536-byte JSON ceiling
role bundles/scope matrix
wrong-domain/tampered/stale ETags + exact-current PUT exception
stale DRAFT always412
PROFILE_REPLACE If-Match/If-None-Match matrix
Idempotency-Key replay/different fingerprint/24h expiry/current-AuthZ recheck
cursor tamper/filter replay/ordering
complete creation/options arrays
upload exact Content-Length/create-once/shared15min expiry + completion size re-proof
100 MiB raw / 256 MiB expanded / 4096-entry admission boundaries
duplicate/traversal/encrypted/invalid package rejection
Governance Step historical label snapshot survives current-route relabel
Audit operation/resource/facts combinations exclude provider_binding.disabled
exact bytes verified before response commit
Range/redirect/206/304/compression absent
Content-Digest == exact body SHA-256
corrupt semantic bytes ->500 internal.content_integrity with zero success bytes
semantic byte-copy mutations may emit only their declared internal.content_integrity path
ReplaySnapshot <=2048
PDF-source RequireOfficialRendition creates no duplicate bytes/job
```

No generic production response-buffer validator is added. Generated typed output plus targeted contract tests remain the accepted minimum. Actual runtime execution of these fixtures belongs to the later validation/implementation program once a runtime exists.

External evidence checked during T8-E includes OpenAPI 3.0.3, RFC9110, RFC9457, RFC9530, Fetch forbidden-header behavior, OWASP archive/upload resource controls, current AWS S3 PutObject/presigning behavior, controlled-document upload limits, Stripe/Adyen idempotency practice, and current `oapi-codegen` / `openapi-typescript` behavior.

---

# 10. Structural Inversion / subtractive checkpoint

Current candidate remains true if legacy API/schema shape were opposite:

```text
78 semantic operations derive from Product/T6
ETag/idempotency/CSRF/pagination derive from accepted invariants
exact-content wire derives from T4
component registry + operation ledger close Writer choices
```

Subtracted / not introduced:

```text
universal response envelope
generic /actions
generic filter/sort DSL
generic public facts/metadata bag
provider/job state
persisted permission snapshot
editable Role/Permission/policy engine
separate Approval API
operation79
server-side cursor state
new T8-C presign signature
S3 POST-form/multipart without measured need
Range/HEAD/304 baseline
arbitrary Problem extension/default response
generator-specific Product fields
recursive archive framework
compression-ratio knob redundant with raw/expanded/entry ceilings
duplicate PDF rendition bytes for PDF source
dormant future capability
```

The two upstream material findings were resolved by the smallest owner-local corrections: two persistence fields for truthful Step-label history and one impossible Audit event removed.

---

# 11. Remaining closure gate

The measurement, generated-boundary feasibility, provider presign feasibility and ledger-census fixture obligations are closed at candidate level.

Remaining Lead gate:

```text
A. run one final whole-candidate Structural Inversion / YAGNI / overengineering / global-coherence attack
B. close every surviving Lead finding without speculative capability
C. revalidate exact candidate HEAD + intended 5-file durable/work diff + required CI
D. only if A→C converge, create review/t8e-fable from that exact candidate HEAD
E. independent Fable challenge
F. Lead adjudication of Fable evidence
G. explicit operator ratification
```

Until A→C converge:

```text
T8-E ACTIVE
T8-F NOT OPEN
implementation BLOCKED
Fable NOT STARTED
```
