---
id: frontend-realization
kind: authority
owner: architecture
summary: Owns T8-F frontend route/lens realization, interaction coverage, generated transport consumption, query/state behavior, read-model consumption, and editor/viewer boundaries.
---

# T8-F — Frontend Realization

> **Status:** OPEN / ACTIVE CANDIDATE. This document is not operator-ratified. Program stage/status and implementation permission remain exclusively in `../roadmap.md`.

Product/API meaning remains in `../product/journeys.md` and `wire-contract.md`. This authority realizes those accepted semantics for the browser frontend; it does not create a new Product architecture, permission model, lifecycle, DTO authority, or API surface.

## 1. Lead outcome

The smallest Frontend Realization contract is:

```text
accepted Product/API authority
→ complete human-interaction coverage
→ semantic route/lens tree
→ per-lens vertical realization
→ one generated transport boundary
→ one server-state model
→ bounded local/navigation state
→ one editor/viewer boundary
```

The derivation is bidirectional:

```text
Product journey / accepted capability
→ semantic owner
→ admitted API operation
→ frontend interaction / UX home

frontend interaction
→ admitted API operation
→ semantic owner
→ accepted Product journey / capability
```

If either direction breaks for a material interaction, stop and reopen only the smallest implicated upstream authority. React code, a client-only business rule, a generic BFF, an ad hoc endpoint, or operation 79 may not silently repair the gap.

The frontend remains a human-operable projection of accepted Product truth. It never becomes a new business owner.

---

## 2. Frontend planning law

T8-F applies this order:

```text
1. recover the bounded accepted authority
2. prove Product/API coverage
3. map accepted human goals and flows
4. derive screens/routes from those flows
5. vertically trace material interactions
6. close state/transport/read-model behavior
7. close editor/viewer boundaries
8. derive package topology last
9. attack for missing authority or speculative mechanism
```

The number of routes, screens, packages, components, stores, and libraries is an output, never a target.

Do not infer a screen from an API noun and do not infer an API from a desired screen.

Generic planning concepts such as external convergence, operational-work ownership, unknown/partial/stale business states, or Approval are applied only where current MetalDocs authority actually contains those semantics. They do not become Launch ontology by methodology reuse.

---

## 3. Accepted human goals / flow inventory

Frontend actors are current ENABLED MetalDocs Users exercising current permissions and relationships; T8-F does not create a second persona/role authority.

| Human goal | Entry lens | Material flow | Exit / handoff |
|---|---|---|---|
| Establish/inspect application session | application shell | resolve session → render authorized navigation; or browser login handoff | requested Product lens |
| End own session | application shell | explicit sign-out → terminate session | browser unauthenticated state |
| Discover official documents | Library | search/filter official truth → inspect result | Document Official |
| Create a Document | Library | creation options → area/type/title/template/owner where allowed → numbering preview → create | Document Work |
| Inspect official/current Document truth | Document Official | inspect stable Document/current official truth → view exact official content | same lens / History / Work |
| Start next Revision | Document Official | request new Revision from current accepted source | Document Work |
| Author current DRAFT | Document Work | inspect draft → edit title and/or replace source → save against DRAFT ETag | Document Work |
| Upload replacement source | Document Work | allocate → exact provider PUT → complete → attach through DRAFT mutation | Document Work |
| Submit DRAFT | Document Work | submit exact current generation → observe governance/rendition/release result | My Work / Governance / Document Official |
| Withdraw/cancel eligible work | Document Work | explicit withdrawal or cancellation | Document Work / History |
| See actor-relevant work | My Work | inspect authoring and governance projections | Document Work / Governance Case |
| Participate in governance | Governance Case | inspect exact immutable subject → feedback and/or Step Decision when allowed | same case / My Work |
| Inspect controlled-document history | Document History | inspect semantic history and authorized exact historical content | referenced detail/context |
| Initiate/manage obsolescence | Document Official | create request; inspect/withdraw when authorized | Governance Case / Document Official |
| Administer Organization identity | Admin / Organization | Company, Users/Profile/provider binding/eligibility, Areas, Group identity | same Admin section |
| Administer effective access | Admin / Access | GroupMemberships + RoleAssignments using fixed roles/scopes | same Admin section |
| Administer document governance | Admin / Document Governance | Document Types, routes, representation, eligible Templates, Template role | same Admin section |
| Inspect Audit | Audit | filter/page meaningful action evidence | referenced Product context |

No UI is created for a machine/system actor merely because a system transition exists. Release remains system-owned; renderer/provider/job mechanisms remain outside Product navigation.

---

## 4. Stable route and screen inventory

The durable Launch SPA route meanings are exactly:

```text
/documents
/documents/:document_id
/documents/:document_id/work
/documents/:document_id/history
/work
/work/governance/:attempt_id
/audit
/admin/organization
/admin/access
/admin/document-governance
```

| Route | Screen / lens | Primary meaning |
|---|---|---|
| `/documents` | Library | ordinary official/effective discovery plus entry to bounded Document creation |
| `/documents/:document_id` | Document Official | stable Document + current official/management lens; EFFECTIVE truth primary |
| `/documents/:document_id/work` | Document Work | exact current open Revision work lens |
| `/documents/:document_id/history` | Document History | authorized Controlled Documents semantic history |
| `/work` | My Work | actor-relevant authoring and governance projections |
| `/work/governance/:attempt_id` | Governance Case | exact GovernanceAttempt + exact immutable governed subject |
| `/audit` | Audit | AuditEvent action evidence |
| `/admin/organization` | Admin / Organization | Company/User/Area/Group identity and User security/eligibility administration |
| `/admin/access` | Admin / Access | GroupMembership and RoleAssignment administration |
| `/admin/document-governance` | Admin / Document Governance | DocumentType/governance/representation/Template configuration |

Root/default redirect behavior, router library choice, nested URL implementation, and visual navigation presentation are not durable Product contracts.

OIDC `/auth/login` and `/auth/callback` are browser integration routes outside the SPA Product route tree and outside the 78-operation application census.

No stable Launch route is introduced for Approvals, Templates, Distribution, Notifications, Taxonomy, Tokens, Sessions, Metrics, Search, Releases, Uploads, or provider/job state.

Creation and bounded admin detail/edit interactions do not require new stable routes at Launch. They may be realized as contextual panels/dialogs or in-place work regions inside their owning screen. A future need for durable deep-link identity must be justified by a concrete user journey rather than route symmetry.

---

## 5. Exact 78-operation frontend coverage proof

Every T8-E operation has at least one concrete frontend consumer. An operation may support more than one lens, but it has one primary coverage home below. No operation 79 is required by this derivation.

### 5.1 Shell / session — 2

```text
1  getSession
2  endSession
```

Primary consumer: application shell.

### 5.2 Admin / Organization — 24

```text
3  searchProviderSubjects
4  getCompany
5  replaceCompany
6  listUsers
7  createUser
8  getUser
9  getUserProfile
10 replaceUserProfile
11 deleteUserProfile
12 getUserProviderBinding
13 replaceUserProviderBinding
14 getUserEligibility
15 replaceUserEligibility
16 listAreas
17 createArea
18 getArea
19 replaceArea
20 getAreaLifecycle
21 replaceAreaLifecycle
22 listGroups
23 createGroup
24 getGroup
25 replaceGroup
26 deleteGroup
```

Primary consumer: `/admin/organization`.

### 5.3 Admin / Access — 7

```text
27 listGroupMembers
28 addGroupMember
29 removeGroupMember
30 listRoles
31 listRoleAssignments
32 createRoleAssignment
33 deleteRoleAssignment
```

Primary consumer: `/admin/access`.

Group identity remains Organization-owned even when membership mutation is co-located under Access for effective-authority administration.

### 5.4 Admin / Document Governance — 12

```text
34 listDocumentTypes
35 createDocumentType
36 getDocumentType
37 replaceDocumentType
38 getDocumentTypeGovernance
39 replaceDocumentTypeGovernance
40 getDocumentTypeEligibleTemplates
41 replaceDocumentTypeEligibleTemplates
42 getDocumentTypeNumberingPreview
43 listTemplateConfigurations
50 getDocumentTemplateRole
51 replaceDocumentTemplateRole
```

Primary consumer: `/admin/document-governance`; operation 42 is also consumed by the Document creation flow when displaying the non-reserving code preview.

### 5.5 Library / Document creation — 3

```text
44 getDocumentCreationOptions
45 listDocuments
46 createDocument
```

Primary consumer: `/documents`.

### 5.6 Document Official / management — 8

```text
47 getDocument
48 getDocumentResponsibleOwner
49 replaceDocumentResponsibleOwner
52 createDocumentRevision
72 getRelease
73 getReleaseSource
75 createObsolescenceRequest
77 withdrawObsolescenceRequest
```

Primary consumer: `/documents/:document_id`.

### 5.7 My Work / Document History supporting reads — 5

```text
53 getDocumentHistory
54 listAuthoringWork
55 listGovernanceWork
56 getRevision
76 getObsolescenceRequest
```

Primary consumers: `/work` and `/documents/:document_id/history`; detail reads may also support cross-links from official/governance lenses.

### 5.8 Document Work — 10

```text
57 getRevisionDraft
58 updateRevisionDraft
59 startRevisionDraftUpload
60 completeRevisionDraftUpload
61 getRevisionDraftSource
62 createSubmission
63 getSubmission
64 getSubmissionSource
65 withdrawSubmission
66 cancelRevision
```

Primary consumer: `/documents/:document_id/work`. Submission reads/source may also appear in authorized history/governance context without transferring Submission authority.

### 5.9 Governance Case — 5

```text
67 getGovernanceAttempt
68 listGovernanceFeedback
69 createGovernanceFeedback
70 getGovernanceStepDecision
71 recordGovernanceStepDecision
```

Primary consumer: `/work/governance/:attempt_id`.

The case also consumes exact governed-subject content through existing Submission/obsolescence reads where the case grants that exact context.

### 5.10 Official rendition presentation — 1

```text
74 getOfficialRenditionContent
```

Primary consumer: Document Official viewer when representation policy requires OfficialRendition(PDF); authorized history may also consume it when the referenced fact is exposed.

### 5.11 Audit — 1

```text
78 listAuditEvents
```

Primary consumer: `/audit`.

Count proof:

```text
2 + 24 + 7 + 12 + 3 + 8 + 5 + 10 + 5 + 1 + 1 = 78
orphaned accepted operations = 0
invented application operations = 0
operation 79 = absent
```

This is a coverage proof, not permission transfer: frontend presence never grants operation admission.

---

## 6. Vertical realization contracts

### 6.1 Application shell

```text
READ             getSession -> SessionView
WRITE            endSession
SERVER STATE     current SessionView only
LOCAL STATE      navigation/presentation only
SECURITY         cookie remains HttpOnly; csrf_token remains in memory/server-state cache
AUTHORITY         no frontend permission matrix
```

A successful `getSession` establishes the current browser application context. The shell may suppress navigation affordances that are not usable, but it never authorizes them. Direct navigation still depends on current API authorization.

No session token or CSRF token is persisted to localStorage/sessionStorage as a correctness baseline.

### 6.2 Library + Document creation

```text
READ             listDocuments
                 getDocumentCreationOptions
                 getDocumentTypeNumberingPreview
WRITE            createDocument
URL STATE        q + admitted catalog filters + cursor navigation context
SERVER STATE     DocumentPage / DocumentCreationOptionsView / NumberingPreviewView
FORM STATE       creation selections + title until accepted
IDEMPOTENCY      one createDocument key per logical create command; same key on transport retry
AUTHORITY         creation options are guidance; create revalidates all truth server-side
```

Library defaults to official/effective truth. DRAFT/SUBMITTED work never appears as ordinary official Library truth merely because the caller can edit it.

The creation UI uses `DocumentCreationOptionsView`; it must not populate ordinary creation selectors from administrative directories. A successful create hands off to Document Work for the created REV000.

### 6.3 Document Official

```text
READ             getDocument
                 getDocumentResponsibleOwner when management affordance is needed
                 getRelease / getReleaseSource as referenced by official truth
                 getOfficialRenditionContent when required
                 getObsolescenceRequest when referenced/currently relevant
WRITE            replaceDocumentResponsibleOwner
                 createDocumentRevision
                 createObsolescenceRequest
                 withdrawObsolescenceRequest
SERVER STATE     purpose-built official/current representations
CONCURRENCY      responsible-owner replacement uses its own strong ETag domain
IDEMPOTENCY      createDocumentRevision/createObsolescenceRequest retain one key per logical command
```

The route never silently switches to DRAFT content based on caller identity. Work remains a separate lens.

### 6.4 Document Work

```text
READ             getRevision / getRevisionDraft / getRevisionDraftSource
                 getSubmission / getSubmissionSource when current work references a Submission
WRITE            updateRevisionDraft
                 startRevisionDraftUpload
                 completeRevisionDraftUpload
                 createSubmission
                 withdrawSubmission
                 cancelRevision
SERVER STATE     DocumentWorkView + operation-specific supporting views
FORM STATE       local title/editor buffer bound to the loaded DRAFT ETag
CONCURRENCY      stale DRAFT mutation always 412 precondition.draft_changed
IDEMPOTENCY      createSubmission retains one key + exact semantic DRAFT If-Match for the logical command
```

Draft title and source are one concurrency domain. A 412 preserves the user's unsaved local input, reloads current authoritative DRAFT truth, and requires an explicit human choice to reapply/retry; no silent last-write-wins or client auto-merge is introduced.

Provider upload success is not READY; READY is not WorkingContent. Only successful DRAFT attachment establishes the new WorkingContent source.

### 6.5 My Work

```text
READ             listAuthoringWork
                 listGovernanceWork
WRITE            none
SERVER STATE     WorkAuthoringPage + WorkGovernancePage
AUTHORITY         projections route the actor to owner lenses; My Work owns no lifecycle truth
```

My Work is an actor-relevant workspace, not a new Operational Work semantic owner or generic task engine.

### 6.6 Governance Case

```text
READ             getGovernanceAttempt
                 listGovernanceFeedback
                 getGovernanceStepDecision
                 exact governed subject/source via already-admitted reads
WRITE            createGovernanceFeedback
                 recordGovernanceStepDecision
SERVER STATE     GovernanceCaseView and immutable supporting facts
LOCAL STATE      feedback/decision form input only
IDEMPOTENCY      feedback POST retains one key per logical command
AUTHORITY         allowed_actions are UX hints only
```

The case displays the exact immutable Submission or obsolescence subject. Participation never grants general WorkingContent/history authority and never permits reviewer mutation of WorkingContent.

A Governance Step Decision is not a public publish command. Release remains system-owned.

### 6.7 Document History

```text
READ             getDocumentHistory
                 referenced getRevision/getSubmission/getRelease/getObsolescenceRequest
                 exact source/rendition reads only when current authorization permits
WRITE            none by virtue of History itself
SERVER STATE     DocumentHistoryView/Page + explicitly requested supporting facts
AUTHORITY         Controlled Documents history; Audit remains separate evidence
```

History never reconstructs current truth from Audit.

### 6.8 Admin / Organization

```text
READ/WRITE       operations 3→26 from §5.2
PERMISSION       organization.manage for Company/User/Area/Group identity surfaces;
                 operation-specific current authority still applies
CONCURRENCY      Company, UserProfile, provider binding, eligibility, Area and Group ETag domains remain separate
IDEMPOTENCY      createUser/createArea/createGroup retain one key per logical command
```

User provider-directory search is a bounded selection preflight, not a general provider directory. Provider subject references remain opaque.

Profile, provider binding, eligibility and User identity must not be flattened into one client-side User write model because they have independent semantics/concurrency.

### 6.9 Admin / Access

```text
READ/WRITE       operations 27→33 from §5.3
PERMISSION       access.manage
SERVER STATE     GroupMemberPage, fixed RoleListView, RoleAssignmentPage
AUTHORITY         fixed T3 roles/scopes; no Role/Permission editor
IDEMPOTENCY      createRoleAssignment retains one key per logical command
```

Membership UI may reference Organization-owned Group/User identity, but membership changes effective authority and remains governed by Access permission semantics.

### 6.10 Admin / Document Governance

```text
READ/WRITE       operations 34→43 + 50→51 from §5.4
PERMISSIONS      document_type.manage and template_use.manage according to owning operation
CONCURRENCY      DocumentType base, governance, eligible-template set and Template-role domains remain separate
IDEMPOTENCY      createDocumentType retains one key per logical command
```

Template administration uses bounded `TemplateConfigurationItem`/template-role views. It does not acquire source, WorkingContent, Submission or full history access merely because a User can administer template configuration.

### 6.11 Audit

```text
READ             listAuditEvents
PERMISSION       audit.read
SERVER STATE     AuditEventPage
WRITE            none
AUTHORITY         meaningful action evidence; never current business-state authority
```

---

## 7. Frontend state authority

Launch uses exactly these baseline state classes:

```text
SERVER STATE
  Product/server-owned truth
  -> TanStack Query

NAVIGATION / URL STATE
  route identity, admitted filters/search, navigation context
  -> router/URL

FORM DRAFT
  user-authored input not yet accepted by Product owner
  -> local feature/form state

EPHEMERAL UI STATE
  disclosure, focus, local selection, panel/dialog state
  -> local React state
```

A fifth durable/global state class requires a concrete consumer, a protected property, and evidence that the four accepted classes are insufficient.

No Redux/Zustand/global entity store is Launch baseline. No server truth is copied to a global client store merely for convenience.

Transport loading/error is not a Product lifecycle or business knowledge state. Generic `unknown`, `partial`, `stale`, `accepted`, `converged`, or similar methodology vocabulary is not invented unless an accepted MetalDocs read model explicitly carries that meaning.

---

## 8. TanStack Query behavior

Query identity follows the admitted operation, not an invented frontend entity cache:

```text
query identity = operationId + canonical path/query semantic inputs
```

Different read models over the same semantic object remain distinct query entries when their Product meanings differ. Document Official, Document Work, History and Governance Case do not share a universal normalized `Document` entity authority.

Potentially unbounded lists consume T8-E cursors exactly:

```text
first page: admitted filters + optional limit
next page: cursor + optional limit only
```

No offset/page-number emulation, total-count requirement, frozen multi-page snapshot, or client generic sort/filter DSL is introduced.

Mutation success handling:

```text
1. accept the returned authoritative representation/result
2. replace the exact affected query representation when the operation returns it
3. invalidate/refetch only semantic lenses that can have changed
4. never fabricate a lifecycle transition optimistically
```

Optimistic presentation state such as a disabled button/spinner is allowed; optimistic semantic truth is not.

There is no generic automatic mutation retry. A retry is operation-aware and must preserve the exact logical command identity, including the same Idempotency-Key where required and the same conditional semantics where applicable. A materially changed command gets a new key/request.

Read retry policy is an implementation tuning choice and is never a correctness dependency; semantic Problems are not converted into blind transport retries.

---

## 9. Generated transport consumption

The application transport boundary is:

```text
api/openapi/v1/openapi.yaml
→ openapi-typescript generated paths/components
→ one thin lib/api transport
→ feature query/command functions
→ React lenses
```

Generated application shapes remain the wire type authority. Features do not hand-maintain parallel DTOs, enums, Problem families, route registries, or request header matrices.

The thin transport owns only cross-cutting browser wire mechanics:

```text
same-origin cookie credentials
session CSRF header injection for unsafe requests
Idempotency-Key carriage supplied by the logical command scope
If-Match / If-None-Match carriage supplied by the concurrency scope
JSON vs exact-byte response handling
RFC9457 Problem decoding
response ETag preservation outside the generated body value
```

It does not own Authorization, lifecycle predicates, Product normalization, target eligibility, or business retries.

No universal response envelope, client-side schema authority, provider DTO, or generated-client-specific Product field is introduced.

---

## 10. ETag, idempotency and Problem behavior

### 10.1 ETag

An ETag-protected frontend representation preserves:

```text
exact generated body value
+ exact response ETag transport metadata
```

The tag is not embedded into the Product DTO and is never derived from database fields/client generation counters.

A form editing an ETag-protected truth binds to the tag from which the form began. On `412 precondition.resource_changed`, keep local user input, refetch authoritative state, and require explicit reconciliation/retry.

DRAFT conflict uses the stricter `412 precondition.draft_changed` law and never takes the whole-replacement exact-current exception.

### 10.2 Idempotency-Key

For each of the accepted 10 idempotent POST creations, the frontend generates one UUID for one logical command and retains it in the in-memory mutation scope while resolving/retrying that exact command.

```text
same logical command / ambiguous transport outcome -> same key
materially changed command                        -> new key
new deliberate command                            -> new key
```

The key is never business identity and needs no durable browser storage baseline.

### 10.3 Problems

Frontend branching uses canonical `Problem.code`, never localized `detail` text.

At minimum:

```text
401 auth.*             -> invalidate current session presentation / offer browser login flow
403 permission.*       -> denied action/context; never reinterpret as authorization success
404 notfound.*         -> absent/non-disclosable according to server contract
409 state./conflict.   -> operation-specific current-state conflict UX
412 precondition.*     -> explicit stale/current-truth reconciliation
422 validation.*       -> input/business validation; field pointers may bind to form fields
503 dependency.*       -> dependency-unavailable UX; never expose raw provider error
500 internal.*         -> generic failure + trace context; never leak provider/storage internals
```

A generic error toast may present non-material failures, but it may not erase a Problem state that changes the user's safe next action.

---

## 11. Read-model consumption

Purpose-built T6/T8-E read models are consumed directly by their owning lenses, including:

```text
SessionView
DocumentSummary / DocumentPage
DocumentOfficialView
DocumentWorkView
DocumentCreationOptionsView
SubmissionView
GovernanceCaseView
DocumentHistoryView / Page
WorkAuthoringItem / Page
WorkGovernanceItem / Page
AuditEventView / Page
GroupMemberPage / UserReference
TemplateConfigurationItem / Page
Admin configuration views
```

The frontend does not create a second durable `DocumentEntity`, `ApprovalDTO`, `ArtifactViewModel`, generic `ReferenceData` model, normalized lifecycle entity store, or persisted `currentStatus` authority.

`allowed_actions` are display/action hints derived from canonical server authorization/predicates. They may hide/disable controls for usability; every command still rechecks server authority.

A read model may contain bounded display references without transferring mutation/current-state ownership to the consuming feature.

---

## 12. Editor / viewer boundary

Launch has one shared interactive DOCX adapter boundary and separate read-only PDF presentation. The adapter is mechanism, never Product authority.

| Context | Exact content authority | Frontend mode | Save/effect path |
|---|---|---|---|
| DRAFT DOCX | current WorkingContent source | interactive editable DOCX | export complete DOCX bytes → upload/admit → DRAFT PATCH + If-Match |
| DRAFT PDF | current WorkingContent source | read-only inspect; source may be replaced | replacement exact bytes → upload/admit → DRAFT PATCH + If-Match |
| SourceOnly official DOCX | exact Release source | interactive adapter read-only | none |
| SourceOnly official PDF | exact Release source | read-only PDF | none |
| RequireOfficialRendition(PDF) | OfficialRendition PDF primary; exact Release source separately labeled/available | read-only PDF primary | none |
| Governance submission | exact immutable Submission source | read-only decision content | governance action never mutates content |

DOCX adapter law:

```text
load exact DOCX bytes
render/edit in browser
emit complete resulting DOCX bytes
support read-only inspection
hold no provider-owned durable semantic identity
provider callback/save state is not business truth
```

Draft save law:

```text
editor complete bytes
→ startRevisionDraftUpload(expected_size_bytes)
→ browser applies returned required_headers verbatim and uploads fixed Blob/File body
→ completeRevisionDraftUpload
→ updateRevisionDraft(source_upload_id, If-Match=current draft ETag)
→ authoritative DocumentWorkView + new ETag
```

The client never authors SHA-256/semantic size/content-format truth. Completion remains server-derived.

No second interactive editor, generic EditorSession, edit lease, collaborative/CRDT baseline, viewer-generated OfficialRendition, or provider callback as semantic save authority is introduced.

The concrete interactive DOCX provider remains subject to the already-ratified fidelity/security corpus before provider freeze; T8-F closes the consumer boundary, not a speculative provider dependency.

---

## 13. Feature/package topology

Package topology is derived after the interaction/route proof:

```text
src/
  app/
    router/
    shell/

  features/
    library/
    document-official/
    document-work/
    governance-work/
    history/
    audit/
    admin/
      organization/
      access/
      document-governance/

  lib/
    api/
      generated/
      transport/

  shared/
    ui/
    editor/
    viewer/
```

This is a semantic-lens organization, not a Go-package mirror.

`/work` composes `document-work` and `governance-work` projections rather than creating a generic Work semantic owner/package platform. Admin co-location does not merge Organization, Access and Document Governance permission ownership.

Do not add baseline top-level frontend layers named `entities`, `domain`, `repositories`, `services`, `approvals`, `templates`, `workflows`, or `stores` merely because an architectural style commonly contains them. A new package must have a current named consumer and a distinct responsibility not already owned above.

Shared UI/editor/viewer primitives remain domain-agnostic; semantic orchestration belongs to the feature lens.

---

## 14. Navigation and authorization presentation

Keep distinct:

```text
authenticated session
!= permission to enter/use a Product interaction
!= relationship/lifecycle eligibility
!= governance participation
!= command executable now
```

The frontend may omit unavailable navigation/actions for usability, but visibility is never authorization.

There is no browser-maintained role/permission matrix. Static role bundles may be displayed in Admin Access from `listRoles`; they are not reimplemented as client policy.

Server 403/404 behavior remains authoritative for direct navigation and current disclosure. A stale UI permission hint never permits a command.

---

## 15. Targeted reopen / falsification law

T8-F may falsify earlier architecture, but only material evidence justifies a reopen.

Examples of legitimate triggers:

```text
a required accepted human journey cannot be completed with operations 1..78
an admitted operation lacks the read data required to present its accepted decision safely
an accepted Product distinction cannot be represented from current read models
an exact-content/editor requirement cannot satisfy the ratified fidelity/security/OCC boundary
current complete creation/options or cursor semantics create demonstrated unusable scale/UX failure
an accepted admin/governance journey requires data that current bounded projections deliberately omit
```

Response:

```text
identify smallest owning upstream authority
→ bounded targeted reopen
→ update Product/API/wire authority and proof
→ return to T8-F derivation
```

Never repair such a gap through hardcoded business labels, hidden database reads, provider DTO leakage, a generic BFF, screen-shaped endpoint, duplicate client truth, or operation 79 by convenience.

Current derivation result:

```text
material upstream falsifier found = none
accepted operations covered       = 78 / 78
operation 79 required              = no
```

---

## 16. Structural inversion / subtraction checkpoint

This realization remains valid if the removed legacy frontend had the opposite folder/component/router shape because it derives from Product lenses and the ratified wire, not legacy code.

Explicitly not introduced:

```text
new Product route or operation
frontend Authorization engine
legacy capability-driven frontend taxonomy
Redux/Zustand/global server-truth mirror
generic normalized Product entity store
manual parallel DTO/schema/Problem authority
generic workflow/Approval frontend
screen-shaped BFF
provider identity/state as Product ontology
SSR correctness dependency
service-worker/offline-first correctness dependency
micro-frontends
design-system platform decision
router-framework decision
dual DOCX editor runtime
EditorSession/lease baseline
optimistic lifecycle authority
generic Search frontend/API
operation 79
```

---

## 17. Concrete T8-G consumers

T8-F supplies T8-G only the runtime consumers that are now concrete:

```text
one React SPA delivery surface
same-origin /api/v1 application access
OIDC browser redirect/callback integration outside application OpenAPI
HttpOnly session cookie + in-memory synchronizer CSRF bootstrap
exact-byte application-origin reads for DOCX/PDF
browser direct-upload capability whose provider CORS permits returned browser-settable headers
one interactive DOCX adapter loaded by the SPA, with provider choice constrained by the ratified fidelity/security proof
read-only PDF presentation
```

T8-F does not choose binaries/processes, runtime host, deployment topology, CDN, container layout, provider CORS configuration, secrets, worker topology, startup/readiness, observability, recovery profile, or server buffering strategy. Those remain T8-G concerns.

No SSR, service worker, BFF, second frontend process, websocket, realtime channel, or editor session service is a concrete T8-G consumer at Launch.

---

## 18. T8-F candidate gate

T8-F remains OPEN / ACTIVE until the candidate has survived repository verification, adversarial coherence review as required by the current Repository Standard process, and explicit operator ratification.

Candidate acceptance must preserve at minimum:

```text
stable Product route meanings       exact accepted T6 set
accepted application operations     78 / 78 covered
operation 79                        absent
frontend semantic owners            none added
server-state authority              TanStack Query / server truth
parallel global server store        absent
manual DTO/API authority            absent
generic frontend AuthZ matrix       absent
DRAFT OCC                           strong ETag; stale -> explicit 412 reconciliation
idempotent logical command retry    same key
exact-content descriptor authority  server-owned
DOCX interactive runtime            one adapter boundary
Governance content                  exact immutable subject / read-only
T8-G                                NOT OPEN
Product implementation              BLOCKED
```

Opening this candidate does not authorize T8-G or Product code.