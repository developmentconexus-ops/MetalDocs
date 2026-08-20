---
id: t8e-executable-wire-contract-proposal
kind: work
owner: architecture
summary: Active T8-E proposal containing only the unresolved executable-wire decisions after the accepted checkpoint.
---

# T8-E executable wire contract — active proposal

> Temporary / non-authoritative. Implementation remains blocked.

## Accepted baseline

Do not duplicate or reinvent the accepted layers. Read:

- `../../reference/t8e-checkpoint.md`
- `../../product/journeys.md`
- `../../decisions/api-operation-census.md`
- `../../decisions/forward-obligations.md`

Current application census: **78 operations**.

## Current question

What is the smallest closed executable ledger for all 78 operations that leaves no material wire choice to an implementation Writer while avoiding duplicate lifecycle/AuthZ authority and speculative capability?

## Lead decision — ledger shape

Use **one closed operation ledger plus one closed component registry**. Do not create prose-per-operation mini-specifications and do not copy lifecycle/AuthZ predicates into the wire contract.

Each of the 78 operation rows must contain exactly the wire choices a Writer would otherwise have to invent:

```text
operation_id
method + path
request component(s)
request header profile
success status set
success body component or no-body
success header profile
allowed Problem.code set
pagination/filter/order profile when applicable
allowed_actions vocabulary reference when the response exposes it
request/body limit profile when materially applicable
```

The component registry owns exact reusable wire shapes once:

```text
shared scalar formats and bounded references
request objects
success projection objects
required / optional / nullable fields
closed enums / discriminators
pagination page shape
header profiles
Problem base + closed problem catalog
allowed_actions enums by projection
limit profiles
```

An operation row may reference a component/profile but may not rely on prose such as “appropriate fields”, “standard errors”, “usual headers”, “etc.”, or an implementation-defined enum. Conversely, a row does **not** restate business authorization, lifecycle transition law, transaction law, or owner eligibility when those are already owned by T1→T8-D authority.

This is the minimum sustainable closure because removing any ledger column leaves a material wire choice to the Writer, while copying semantic predicates into the ledger would create a second authority.

## Wire modeling laws

### Closed OpenAPI 3.0.3 objects

OpenAPI 3.0.3 defaults `additionalProperties` to `true`. Therefore every fixed MetalDocs request, response, reference, page and Problem object is encoded with:

```text
additionalProperties: false
```

unless the component is deliberately a map. The only currently justified map is the temporary direct-upload capability's provider-required HTTP header map. There is no generic metadata/settings/property-bag map in the Launch application wire.

This is required for the accepted central request-contract validation property: without it, undeclared JSON members remain schema-valid and the “exact request” claim is false.

OAS 3.0.3 `nullable: true` is used only for explicit JSON null; `null` is not modeled as a separate schema type. Request and response components are separate when their legal fields differ. `readOnly`/`writeOnly` annotations are not used as a substitute for purpose-built request/response shapes.

Do not build object inheritance with `allOf` merely to reduce YAML. Closed objects plus OAS 3.0 composition can make inheritance harder to validate/generate correctly. True semantic unions use explicit `oneOf` + discriminator branches; ordinary reuse uses nested `$ref` components.

### Presence and nullability

```text
required   = member must be present on the wire
optional   = member may be absent
nullable   = explicit JSON null is a valid semantic value
```

Absence and `null` are never interchangeable by convention. `nullable: true` is permitted only where current Product/T1→T8-D semantics contain an explicit empty/unknown state that must be distinguishable on the wire. Otherwise use a required concrete value or optional absence according to the owning semantics.

Request update semantics must be explicit per request component. No PATCH/PUT field acquires implicit “null means clear” behavior. The accepted DRAFT PATCH remains the concrete precedent: omitted means unchanged and null has no implicit delete meaning.

### Closed vocabularies

Every semantic discriminator/action/state emitted or accepted by the application contract is a closed OpenAPI enum at Launch. No free-form string is used where current authority already defines a finite vocabulary.

Wire spelling is lower snake case when an owning authority names values only as uppercase/PascalCase semantic prose. A vocabulary that is already a canonical product identifier retains its exact spelling. In particular, T3 `PermissionCode` values remain the canonical dot-separated strings (`document.read_effective`, etc.); T8-E does not create snake-case aliases for them.

### Response shape

Purpose-built success schemas remain the rule. There is no universal `{data, meta}` envelope and no generic action result.

Potentially unbounded collection responses use the accepted page shape:

```json
{
  "items": [],
  "page": {
    "next_cursor": null,
    "has_more": false
  }
}
```

`next_cursor` is the only baseline nullable pagination member: it is `null` when no next page exists. `items`, `page`, `next_cursor`, and `has_more` are present on every paginated success response.

### Request header profiles

```text
SAFE_READ
  no application header beyond ordinary authenticated request context

UNSAFE_CSRF
  X-CSRF-Token required

IDEMPOTENT_CREATE
  X-CSRF-Token required
  Idempotency-Key required; client-generated UUID

IF_MATCH_MUTATION
  X-CSRF-Token required
  If-Match required; exactly one strong RFC entity-tag
  wildcard, weak tag, or list forbidden

SUBMISSION_CREATE
  X-CSRF-Token required
  Idempotency-Key required; client-generated UUID
  If-Match required for the exact DRAFT generation

PROFILE_REPLACE
  X-CSRF-Token required
  existing profile replacement -> If-Match required
  absent-profile recreation -> If-None-Match exactly `*` required
  both headers together -> invalid request
```

Every unsafe `/api/v1` operation uses one of the unsafe profiles above even where its body is empty.

### Success header profiles

```text
NO_STORE
  Cache-Control: no-store

JSON_NO_STORE
  Content-Type: application/json
  Cache-Control: no-store

JSON_ETAG
  Content-Type: application/json
  ETag: required strong opaque entity tag
  Cache-Control: no-store

JSON_ETAG_MUTATION
  Content-Type: application/json
  ETag: required current/new strong opaque entity tag
  Cache-Control: no-store

EXACT_BYTES
  Content-Type
  Content-Length
  Content-Disposition: server-generated ASCII filename
  Content-Digest: RFC 9530 dictionary containing exactly SHA-256
  Cache-Control: private, no-store, no-transform
  Accept-Ranges: none
  X-Content-Type-Options: nosniff
```

For exact bytes, the `Content-Digest` wire spelling is the RFC 9530 structured-field form:

```text
sha-256=:<base64 SHA-256 bytes>:
```

not a hex digest copied into an HTTP header.

`WWW-Authenticate` belongs to the 401 Problem response profile, not ordinary success responses. Problem responses also carry `Cache-Control: no-store`.

### Problem closure

The accepted RFC 9457 catalog remains one global code → type/title/status authority. The operation ledger owns only the **allowed subset of catalog codes for that operation**.

No operation has a `default` response and no operation inherits an undocumented “common errors” bucket. Shared response components are allowed only when they preserve an explicit closed code set; tooling convenience may not widen the public contract.

`401`, `403`, `404`, `409`, `410`, `412`, `413`, `415`, `422`, `429`, `500`, and `503` therefore appear only on operations for which the corresponding failure is reachable under the accepted semantics. `405` remains HTTP method handling for the declared path surface and is not evidence of an extra application operation.

### Filters and ordering

There is no generic filter or sort DSL. Each list operation names its exact accepted query members and one deterministic ordering. Cursor integrity binds the normalized filter/order tuple.

The already-ratified document catalog rule is preserved:

```text
q present:
exact code
→ code prefix
→ title prefix
→ title contains
→ code + stable id tie-break

q absent:
code + stable id
```

Other lists use the smallest stable ordering sufficient for deterministic seek pagination. Do not add locale/collation semantics merely for aesthetics when a stable identifier ordering is sufficient.

### `allowed_actions`

`allowed_actions` is emitted only by projections whose owning journey names it as a UX hint. Each such projection references its own closed enum. There is no product-wide action vocabulary and no generic `/actions` operation.

The current authority explicitly requires it for `GovernanceCaseView`; no second projection has yet justified an `allowed_actions` member. Its Launch vocabulary is exactly:

```text
accept
return_for_changes
add_feedback
```

The array may be empty. These are hints projected from the canonical current Authorization + Controlled Documents predicate evaluation. Every command rechecks canonical truth.

### Limits

Do not invent round-number limits to make the schema look complete.

- Pagination is already closed at default `20`, maximum `100`.
- Direct-upload capability lifetime is already bounded to at most `15 minutes`, with exact `expires_at` returned.
- Raw document bytes, structurally expanded DOCX bytes, ZIP entry count/depth, and any document-format-specific admission limits remain **Unknown pending the required measurement evidence**.
- JSON/string/list maxima are frozen only where Product/T1→T8-D semantics or concrete abuse/tooling evidence supplies a real bound. A runtime/framework default is not Product contract evidence.

A 413 catalog entry may exist before every operation uses it; operation rows may reference it only after a concrete application request-size ceiling is frozen for that request profile.

## Closed shared component registry — layer A

The following is the minimum shared vocabulary. Field-specific schemas may further constrain these scalars; they may not weaken them.

### Scalars

```text
Uuid
  string; format=uuid

UtcInstant
  string; format=date-time; canonical server serialization is RFC3339 UTC with `Z`

OpaqueCursor
  string; nonblank; client treats as opaque

IdempotencyKey
  string; format=uuid

CsrfToken
  string; nonblank; opaque

ProviderSubjectRef
  string; nonblank; opaque

Sha256Hex
  string; pattern `^[0-9a-f]{64}$`

NonBlankString
  string; semantic value after trimming must be nonblank

CodeToken
  string; pattern `^[A-Z0-9]+$`; `-` forbidden

DocumentCode
  string; pattern `^[A-Z0-9]+(?:-[A-Z0-9]+)?-[0-9]{3,}$`

EmailAddress
  string; format=email

RevisionOrdinal
  integer; minimum=0
```

Exact maximum lengths remain unset until an aggregate JSON/request-limit decision or stronger semantic evidence makes them material. This is deliberate absence of a field-specific limit, not a Writer-selected maximum.

### Bounded references

All are closed objects.

```text
UserReference
  required: user_id
  optional: display_name
  law: display_name may be absent after erasable UserProfile removal; no email/contact data

AreaReference
  required: area_id, code, name

DocumentTypeReference
  required: document_type_id, code, name

DocumentReference
  required: document_id, code
  law: stable Document reference does not smuggle Revision-owned title

RevisionReference
  required: revision_id, ordinal, title

ContentSummary
  required: sha256, size_bytes, content_format
  size_bytes: integer >= 0
```

### Core closed enums

```text
ContentFormat
  docx | pdf

RevisionState
  draft | submitted | effective | superseded | obsolete | cancelled

UserEligibilityState
  enabled | disabled

AreaLifecycleState
  active | inactive

DocumentTypeState
  active | inactive

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

SubmissionCreationState
  governance_pending | rendition_pending | released
```

No upload lifecycle enum is public: successful allocation means the handle is usable for its temporary create-only upload, and successful completion means it is READY for the accepted attachment flow. OPEN/READY remain bounded mechanism semantics rather than a generic application state vocabulary.

### True unions

Each union is `oneOf` with a required discriminator and closed branches.

```text
RoleAssignmentSubject
  { kind: user,  user_id }
  { kind: group, group_id }

RoleAssignmentScope
  { kind: company }
  { kind: area, area_id }

GovernanceSelector
  { kind: named_user, user_id }
  { kind: group, group_id }

GovernancePolicy
  { mode: no_human_approval }
  { mode: use_governance_route, steps: GovernanceRouteStep[1..] }

GovernanceRouteStep
  required: label, selector

RepresentationPolicy
  { kind: source_only }
  { kind: require_official_rendition, format: pdf }
```

No branch contains provider IDs, Role predicates, candidate-user snapshots, job state, or a generic extension bag.

### Pagination and Problem base

```text
Page
  required: next_cursor, has_more
  next_cursor: OpaqueCursor | null
  has_more: boolean

ProblemError
  required: pointer, detail
  pointer: RFC 6901 pointer rooted only at /path, /query, /header, or /body
  detail: human-readable; rejected sensitive value never echoed

Problem
  required: type, title, status, detail, instance, code, trace_id
  optional: errors
  errors when present: non-empty ProblemError[]
  additionalProperties: false
```

The Problem profile is intentionally stricter than RFC 9457's extensibility allowance because MetalDocs has accepted a closed application contract. The RFC requires clients to ignore unknown extensions; it does not require this server contract to emit arbitrary extensions.

## First operation families — exact request/success component closure

This section freezes request/success wire choices for the first **43** census operations. Problem-code subsets are not stamped into these rows until the closed catalog pass immediately following all request/success components; that prevents inventing error codes family-by-family and later discovering inconsistent meanings.

### Session / AuthN components

```text
SessionView
  required: user, csrf_token
  user: UserReference

ProviderSubjectOption
  required: provider_subject_ref, display_hints
  display_hints: string[]; bounded provider-neutral presentation hints

ProviderSubjectSearchView
  required: items
  items: ProviderSubjectOption[]
```

Provider-subject search preserves provider result order; it is a bounded external selection preflight, not a product catalog and not cursor-paginated. `query` is required and nonblank. Provider ordering has no Product semantic meaning.

### Organization components

```text
CompanyView
  required: company_id, name

ReplaceCompanyRequest
  required: name

UserProfileInput
  required: display_name
  optional: email

CreateUserRequest
  required: provider_subject_ref, profile
  profile: UserProfileInput

CreateUserResult
  required: user_id

UserView
  required: user_id
  law: stable User identity only; profile/binding/eligibility stay on their canonical subresources

UserSummary
  required: user, eligibility
  user: UserReference

UserPage
  required: items, page

UserProfileView
  required: user_id, display_name
  optional: email

ReplaceUserProfileRequest
  required: display_name
  optional: email
  law: whole replacement; omitted email means resulting profile has no email

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
  law: new Area starts active; create-time inactive capability is absent

CreateAreaResult
  required: area_id

ReplaceAreaRequest
  required: name
  law: Area code is immutable after creation and therefore absent from replacement

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

`DELETE /users/{user_id}/profile` means “ensure erasable profile enrichment is absent”; it is naturally idempotent and does not require `If-Match`. Concurrent profile recreation/update serializes according to canonical owner truth; the delete command does not claim to delete a specific stale field version.

### Authorization components

```text
RoleView
  required: code, permissions, allowed_scope_kinds
  permissions: unique PermissionCode[]
  allowed_scope_kinds: unique RoleAssignmentScopeKind[]

RoleListView
  required: items
  items: RoleView[] in the canonical T3 role order

RoleAssignmentView
  required: assignment_id, subject, role, scope

RoleAssignmentPage
  required: items, page

CreateRoleAssignmentRequest
  required: subject, role, scope

CreateRoleAssignmentResult
  required: assignment_id
```

The static Role endpoint exposes canonical product-owned role bundles; it does not expose an editable policy model.

### Document Governance components

```text
DocumentTypeView
  required: document_type_id, code, name, numbering_scope, state

DocumentTypePage
  required: items, page

CreateDocumentTypeRequest
  required: code, name, numbering_scope
  law: new DocumentType starts active

CreateDocumentTypeResult
  required: document_type_id

ReplaceDocumentTypeRequest
  required: code, name, numbering_scope, state
  law: server rejects code/numbering_scope changes once the ratified first-use immutability condition applies

DocumentTypeGovernanceView
  required: governance, representation

ReplaceDocumentTypeGovernanceRequest
  required: governance, representation
  governance: GovernancePolicy
  representation: RepresentationPolicy

EligibleTemplatesView
  required: templates
  templates: DocumentReference[] ordered by code, document_id

ReplaceEligibleTemplatesRequest
  required: template_document_ids
  template_document_ids: unique Uuid[]; empty array valid

NumberingPreviewView
  required: preview_code, reservation
  reservation: boolean constrained to false

TemplateConfigurationItem
  required: document, template_role, has_effective_revision, eligible_document_type_ids
  optional: current_effective_title
  document: DocumentReference
  eligible_document_type_ids: unique Uuid[]

TemplateConfigurationPage
  required: items, page
```

`current_effective_title` is absent when no EFFECTIVE Revision exists; `null` is not emitted. No source bytes, WorkingContent, Submission/history, provider identity, or content permission snapshot appears in this admin projection.

### Operation rows 1–43

`PAGED` means query accepts only `cursor` and `limit` (`1..100`, default `20`) unless another query member is named.

| # | operationId | Method + path | Request / request profile | Success | Success headers | Query / deterministic order |
|---:|---|---|---|---|---|---|
| 1 | `getSession` | `GET /api/v1/session` | `SAFE_READ` | `200 SessionView` | `JSON_NO_STORE` | none |
| 2 | `endSession` | `DELETE /api/v1/session` | no body / `UNSAFE_CSRF` | `204` | `NO_STORE` | none |
| 3 | `searchProviderSubjects` | `GET /api/v1/authentication/provider-subjects` | `SAFE_READ` | `200 ProviderSubjectSearchView` | `JSON_NO_STORE` | required `query`; provider order preserved |
| 4 | `getCompany` | `GET /api/v1/company` | `SAFE_READ` | `200 CompanyView` | `JSON_ETAG` | none |
| 5 | `replaceCompany` | `PUT /api/v1/company` | `ReplaceCompanyRequest` / `IF_MATCH_MUTATION` | `200 CompanyView` | `JSON_ETAG_MUTATION` | none |
| 6 | `listUsers` | `GET /api/v1/users` | `SAFE_READ` | `200 UserPage` | `JSON_NO_STORE` | `PAGED`; `user_id ASC` |
| 7 | `createUser` | `POST /api/v1/users` | `CreateUserRequest` / `IDEMPOTENT_CREATE` | `201 CreateUserResult` | `JSON_NO_STORE` | none |
| 8 | `getUser` | `GET /api/v1/users/{user_id}` | `SAFE_READ` | `200 UserView` | `JSON_NO_STORE` | none |
| 9 | `getUserProfile` | `GET /api/v1/users/{user_id}/profile` | `SAFE_READ` | `200 UserProfileView` | `JSON_ETAG` | none |
| 10 | `replaceUserProfile` | `PUT /api/v1/users/{user_id}/profile` | `ReplaceUserProfileRequest` / `PROFILE_REPLACE` | `200` update or `201` recreation, both `UserProfileView` | `JSON_ETAG_MUTATION` | none |
| 11 | `deleteUserProfile` | `DELETE /api/v1/users/{user_id}/profile` | no body / `UNSAFE_CSRF` | `204` including exact repeat | `NO_STORE` | none |
| 12 | `getUserProviderBinding` | `GET /api/v1/users/{user_id}/provider-binding` | `SAFE_READ` | `200 UserProviderBindingView` | `JSON_ETAG` | none |
| 13 | `replaceUserProviderBinding` | `PUT /api/v1/users/{user_id}/provider-binding` | `ReplaceUserProviderBindingRequest` / `IF_MATCH_MUTATION` | `200 UserProviderBindingView` | `JSON_ETAG_MUTATION` | none |
| 14 | `getUserEligibility` | `GET /api/v1/users/{user_id}/eligibility` | `SAFE_READ` | `200 UserEligibilityView` | `JSON_ETAG` | none |
| 15 | `replaceUserEligibility` | `PUT /api/v1/users/{user_id}/eligibility` | `ReplaceUserEligibilityRequest` / `IF_MATCH_MUTATION` | `200 UserEligibilityView` including exact-current no-op | `JSON_ETAG_MUTATION` | none |
| 16 | `listAreas` | `GET /api/v1/areas` | `SAFE_READ` | `200 AreaPage` | `JSON_NO_STORE` | `PAGED`; `code ASC, area_id ASC` |
| 17 | `createArea` | `POST /api/v1/areas` | `CreateAreaRequest` / `IDEMPOTENT_CREATE` | `201 CreateAreaResult` | `JSON_NO_STORE` | none |
| 18 | `getArea` | `GET /api/v1/areas/{area_id}` | `SAFE_READ` | `200 AreaView` | `JSON_ETAG` | none |
| 19 | `replaceArea` | `PUT /api/v1/areas/{area_id}` | `ReplaceAreaRequest` / `IF_MATCH_MUTATION` | `200 AreaView` | `JSON_ETAG_MUTATION` | none |
| 20 | `getAreaLifecycle` | `GET /api/v1/areas/{area_id}/lifecycle` | `SAFE_READ` | `200 AreaLifecycleView` | `JSON_ETAG` | none |
| 21 | `replaceAreaLifecycle` | `PUT /api/v1/areas/{area_id}/lifecycle` | `ReplaceAreaLifecycleRequest` / `IF_MATCH_MUTATION` | `200 AreaLifecycleView` | `JSON_ETAG_MUTATION` | none |
| 22 | `listGroups` | `GET /api/v1/groups` | `SAFE_READ` | `200 GroupPage` | `JSON_NO_STORE` | `PAGED`; `group_id ASC` |
| 23 | `createGroup` | `POST /api/v1/groups` | `CreateGroupRequest` / `IDEMPOTENT_CREATE` | `201 CreateGroupResult` | `JSON_NO_STORE` | none |
| 24 | `getGroup` | `GET /api/v1/groups/{group_id}` | `SAFE_READ` | `200 GroupView` | `JSON_ETAG` | none |
| 25 | `replaceGroup` | `PUT /api/v1/groups/{group_id}` | `ReplaceGroupRequest` / `IF_MATCH_MUTATION` | `200 GroupView` | `JSON_ETAG_MUTATION` | none |
| 26 | `deleteGroup` | `DELETE /api/v1/groups/{group_id}` | no body / `UNSAFE_CSRF` | `204` | `NO_STORE` | none |
| 27 | `listGroupMembers` | `GET /api/v1/groups/{group_id}/members` | `SAFE_READ` | `200 GroupMemberPage` | `JSON_NO_STORE` | `PAGED`; `user_id ASC` |
| 28 | `addGroupMember` | `PUT /api/v1/groups/{group_id}/members/{user_id}` | no body / `UNSAFE_CSRF` | `201` first creation; `204` exact repeat | `NO_STORE` | none |
| 29 | `removeGroupMember` | `DELETE /api/v1/groups/{group_id}/members/{user_id}` | no body / `UNSAFE_CSRF` | `204` including exact repeat | `NO_STORE` | none |
| 30 | `listRoles` | `GET /api/v1/roles` | `SAFE_READ` | `200 RoleListView` | `JSON_NO_STORE` | fixed canonical role order; not paginated |
| 31 | `listRoleAssignments` | `GET /api/v1/role-assignments` | `SAFE_READ` | `200 RoleAssignmentPage` | `JSON_NO_STORE` | `PAGED`; `assignment_id ASC` |
| 32 | `createRoleAssignment` | `POST /api/v1/role-assignments` | `CreateRoleAssignmentRequest` / `IDEMPOTENT_CREATE` | `201 CreateRoleAssignmentResult` | `JSON_NO_STORE` | none |
| 33 | `deleteRoleAssignment` | `DELETE /api/v1/role-assignments/{assignment_id}` | no body / `UNSAFE_CSRF` | `204` including exact repeat | `NO_STORE` | none |
| 34 | `listDocumentTypes` | `GET /api/v1/document-types` | `SAFE_READ` | `200 DocumentTypePage` | `JSON_NO_STORE` | `PAGED`; `code ASC, document_type_id ASC` |
| 35 | `createDocumentType` | `POST /api/v1/document-types` | `CreateDocumentTypeRequest` / `IDEMPOTENT_CREATE` | `201 CreateDocumentTypeResult` | `JSON_NO_STORE` | none |
| 36 | `getDocumentType` | `GET /api/v1/document-types/{document_type_id}` | `SAFE_READ` | `200 DocumentTypeView` | `JSON_ETAG` | none |
| 37 | `replaceDocumentType` | `PUT /api/v1/document-types/{document_type_id}` | `ReplaceDocumentTypeRequest` / `IF_MATCH_MUTATION` | `200 DocumentTypeView` | `JSON_ETAG_MUTATION` | none |
| 38 | `getDocumentTypeGovernance` | `GET /api/v1/document-types/{document_type_id}/governance` | `SAFE_READ` | `200 DocumentTypeGovernanceView` | `JSON_ETAG` | none |
| 39 | `replaceDocumentTypeGovernance` | `PUT /api/v1/document-types/{document_type_id}/governance` | `ReplaceDocumentTypeGovernanceRequest` / `IF_MATCH_MUTATION` | `200 DocumentTypeGovernanceView` | `JSON_ETAG_MUTATION` | none |
| 40 | `getDocumentTypeEligibleTemplates` | `GET /api/v1/document-types/{document_type_id}/eligible-templates` | `SAFE_READ` | `200 EligibleTemplatesView` | `JSON_ETAG` | templates ordered `code ASC, document_id ASC` |
| 41 | `replaceDocumentTypeEligibleTemplates` | `PUT /api/v1/document-types/{document_type_id}/eligible-templates` | `ReplaceEligibleTemplatesRequest` / `IF_MATCH_MUTATION` | `200 EligibleTemplatesView` | `JSON_ETAG_MUTATION` | none |
| 42 | `getDocumentTypeNumberingPreview` | `GET /api/v1/document-types/{document_type_id}/numbering-preview` | `SAFE_READ` | `200 NumberingPreviewView` | `JSON_NO_STORE` | optional `area_id`; semantic applicability revalidated |
| 43 | `listTemplateConfigurations` | `GET /api/v1/document-governance/templates` | `SAFE_READ` | `200 TemplateConfigurationPage` | `JSON_NO_STORE` | `PAGED`; `document.code ASC, document_id ASC` |

The use of stable-id ordering for Users, Groups, memberships and RoleAssignments is deliberate: no Product authority names a user-facing search/ranking contract for those administration lists, and inventing locale/case-folding rules solely for pagination would add accidental contract complexity. A later named UX consumer can reopen a list filter/order without changing semantic ownership.

## Direct-upload capability component exception

The accepted direct-upload allocation is the one current wire shape that deliberately contains a map because provider-required upload headers are mechanism-specific and must be forwarded exactly without making provider identity Product truth.

```text
DraftUploadAllocation
  required: upload_id, upload_url, expires_at, required_headers
  upload_id: Uuid
  upload_url: URI capability; temporary provider target
  expires_at: UtcInstant; <= 15 minutes from allocation
  required_headers: map<string,string>
  law: contains only exact headers required for the create-only provider PUT
```

The application never exposes bucket/key/version/storage ETag/provider account identity. A successful completion request has **no request body** and the minimum success is `204`; the subsequent DRAFT PATCH references only `upload_id`, while the server uses its own derived descriptor.

## Generation feasibility — evidence-backed candidate pair

T8-E has not selected implementation dependencies, but feasibility needs concrete generators rather than generic claims. The first probe pair is:

```text
Go          oapi-codegen strict-server generation
TypeScript  openapi-typescript paths/components generation
```

Why these are the current Global-Maximum probe candidates:

```text
oapi-codegen
  generates typed per-operation request objects
  generates typed finite response-object unions/statuses
  supports OpenAPI oneOf/discriminator patterns
  supports a customizable strict ResponseErrorHandlerFunc
  keeps incoming request validation as a separate concern

openapi-typescript
  generates the exact paths/components type boundary requested by T6/T8-E
  adds no generated runtime SDK requirement
  preserves required vs optional and nullable type distinctions
  supports oneOf-oriented discriminated type modeling
  rewards explicit additionalProperties modeling instead of silently widening objects
```

This pair matches the accepted architecture better than selecting a larger generated SDK/runtime merely because it can also generate types. It remains a **probe target, not implementation authorization**.

Material probe conditions:

```text
1. fixed object with additionalProperties:false rejects/does not type arbitrary members
2. required vs optional vs nullable remains distinguishable in Go and TypeScript
3. every closed enum remains finite in generated types
4. GovernancePolicy / GovernanceSelector / RepresentationPolicy unions generate without `any`/untyped escape
5. operation with multiple declared success statuses generates a closed typed response set
6. operation-specific Problem responses do not require a `default` response
7. strict-server unexpected errors can be routed through the canonical RFC 9457 500 serializer
8. request validation is demonstrably provided by the separate central OpenAPI validator, not assumed from strict generation
9. generated package remains one Go wire boundary and one TS paths/components boundary
10. no generator-specific field/provider identity enters the public contract
```

A generator limitation may change schema **encoding** only when semantics remain exact. It is not authority to add fields, widen enums, collapse nullability, add a default response, or replace a true union with an untyped bag.

Evidence checked during this pass:

```text
OpenAPI 3.0.3 specification
  additionalProperties defaults true
  nullable is the OAS 3.0 null mechanism
  oneOf/discriminator are supported Schema/Object mechanisms

RFC 9530
  Content-Digest is an HTTP Structured Fields dictionary
  SHA-256 value is a byte sequence, producing `sha-256=:base64:` syntax

current oapi-codegen documentation/release line
  strict server typed request/response boundary
  custom ResponseErrorHandlerFunc
  incoming request validation is not implied by strict server generation

current openapi-typescript documentation
  generated paths/components boundary
  precise optional/nullable typing
  oneOf guidance
  explicit additionalProperties guidance
```

## Runtime contract-conformance proof design

T8-E freezes the proof obligation, not the runtime implementation:

```text
request path:
raw HTTP request
→ central OpenAPI request validation
→ generated typed boundary
→ semantic handler

response path:
semantic result
→ generated typed response boundary
→ HTTP response
→ contract tests validate status + headers + body against OpenAPI
```

Required negative proof classes include at least:

```text
undeclared JSON member is rejected on every fixed request object
malformed request member or header is rejected
missing required If-Match / Idempotency-Key / CSRF is rejected where applicable
weak/list/wildcard If-Match is rejected where forbidden
PROFILE_REPLACE rejects missing conditional header and rejects both conditional headers together
undeclared enum value is rejected
success cannot omit a required response member/header
operation cannot emit a Problem.code not declared for that operation
paginated response/cursor contract is exact
exact-byte response cannot silently become redirect/range/compressed/provider URL behavior
Content-Digest syntax and digest bytes match the exact response body
```

Do not add a generic production response-buffer validator merely to prove the contract; generated typed output plus contract tests is the accepted minimum.

## Closed operation-census partition

The ledger must contain exactly the current authority's 78 operations and no more:

```text
Session/AuthN support               3
Organization                       26
Authorization                       4
Document Governance config         10
Controlled Documents / Work        34
Audit                               1
TOTAL                              78
```

`getUserProfile` and `getAreaLifecycle` are the two bounded read-symmetry operations already ratified in `api-operation-census.md`. Operation 79 is a material Product/T6 reopen.

## Ledger authoring order

Close the ledger in dependency order, not repository/module order:

```text
1. shared scalar/reference/problem/header/page components       CLOSED in candidate
2. exact read projections and their enums                       PARTIAL: operations 1–43
3. exact command/request objects and presence/nullability       PARTIAL: operations 1–43
4. 78 operation rows: request → success → problems              request/success 1–43 closed; problems open
5. list filters/order/cursor binding                            1–43 closed
6. projection-specific allowed_actions vocabularies             governance case vocabulary closed
7. request/admission limits from evidence                       document corpus measurement open
8. generated Go + TypeScript feasibility probe                  candidate pair/proof frozen; executable probe open
9. runtime conformance proof design                             CLOSED in candidate
10. Structural Inversion / subtractive / global-coherence pass  open
```

This ordering lets command schemas reuse already-closed representations without making generated DTOs or handlers semantic authority.

## Still-open evidence / authority work

The remaining material work is now narrower:

1. Close Controlled Documents / Work request/success components and operation rows **44–77**.
2. Close Audit operation **78**.
3. Build the single RFC 9457 Problem catalog and stamp an exact allowed-code subset onto every row 1–78.
4. Freeze Controlled Documents list filters/order and the exact `GovernanceCaseView` shape using the already-closed `allowed_actions` vocabulary.
5. Measure representative raw DOCX/PDF corpus plus structurally expanded DOCX/ZIP behavior; do not substitute internet defaults for the required product corpus evidence.
6. Execute the disposable Go + TypeScript generation/compile/type probe against the candidate schema encoding before promotion.
7. Run final Structural Inversion / YAGNI / overengineering / global-coherence challenge.

The repository currently supplies no named representative binary document corpus in the current authority pack. The measurement obligation therefore remains **Unknown**, not converted into a guessed upload ceiling. Because the checkpoint explicitly requires measurement, locating/obtaining the representative corpus is a real closure prerequisite rather than optional polish.

## Laws

```text
accepted checkpoint decisions are not reopened by preference
no implementation code/schema/OpenAPI authoring in T8-E
no restored legacy wire becomes target authority
no generic response envelope/action API/error dialect/ACL/event bus
no field/enum/nullability/problem-code choice is deferred to Writers
unknown remains unknown until evidence closes it
prepare seams; do not build dormant capability
```

## Completion gate

```text
closed 78-operation executable ledger
+ exact schema/component closure
+ generation/conformance feasibility proof
+ measured document-admission evidence
+ subtractive/global-coherence pass
+ isolated final Fable review branch
+ Lead adjudication
+ explicit operator ratification
```

Only after T8-E ratification may T8-F open.