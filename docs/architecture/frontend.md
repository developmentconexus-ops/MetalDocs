---
id: frontend-realization
kind: authority
owner: architecture
summary: Owns T8-F frontend route/lens realization, interaction coverage, generated transport consumption, query/state behavior, read-model consumption, and editor/viewer boundaries.
---

# T8-F — Frontend Realization

> **Ratification:** OPERATOR-RATIFIED on 2026-08-21 after required CI #1047 and bounded Fable Round 2 CONVERGED. Current program stage, integration status, implementation permission and exact next action remain exclusively in `../roadmap.md`.

Product/API meaning remains in `../product/journeys.md`; the executable wire is `wire-contract.md`. The bounded T8-E precision discovered during T8-F is recorded for provenance in `../decisions/frontend-read-symmetry.md` and is folded into the wire SSOT. This authority realizes those accepted semantics for the browser frontend. It creates no new Product owner, lifecycle, Permission, DTO authority, or API operation.

## 1. Lead outcome

The smallest frontend contract is:

```text
accepted Product/API authority
→ complete human-interaction coverage
→ semantic route/lens tree
→ per-lens vertical realization
→ one generated transport boundary
→ server state + bounded URL/form/UI state
→ one DOCX editor/viewer boundary
→ package topology last
```

Derivation is bidirectional:

```text
Product journey → semantic owner → admitted operation/read model → UX home
frontend interaction → admitted operation → semantic owner → accepted Product journey
```

A material break stops T8-F and reopens only the smallest implicated authority. React business truth, generic BFFs, ad hoc endpoints, parallel DTOs, and operation 79 may not repair a gap silently.

## 2. Planning law

T8-F applies this order:

```text
1. recover bounded accepted authority
2. prove Product/API coverage
3. map accepted human goals and flows
4. derive screens/routes
5. vertically trace material interactions
6. close state/transport/read-model behavior
7. close editor/viewer boundaries
8. derive package topology
9. attack for missing authority or speculative mechanism
```

Generic concepts imported from another planning method — Operational Work ownership, Approval as peer Product, external convergence, generic unknown/partial/stale ontology — exist only if current MetalDocs authority contains them.

## 3. Human goals and flows

Frontend actors are current ENABLED MetalDocs Users exercising accepted Permissions/relationships; T8-F does not invent a second persona or role authority.

| Human goal | UX home | Material flow |
|---|---|---|
| Establish/end session | application shell | `getSession` → render shell; `endSession` → unauthenticated browser state |
| Discover official documents | Library | search/filter official truth → Document Official |
| Create Document | Library | creation options → numbering preview → create → Document Work |
| Inspect official/current Document truth | Document Official | official metadata/content; management affordances only when supported by returned truth |
| Start or enter open Revision | Document Official | `open_revision` decides go-to-work vs create-next-Revision |
| Author DRAFT | Document Work | read draft → edit title/content → save under DRAFT ETag |
| Upload replacement source | Document Work | allocate → provider PUT → complete → DRAFT attach |
| Submit / withdraw / cancel | Document Work | exact current generation → accepted lifecycle command/result |
| See actor-relevant work | My Work | authoring/governance projections → owner lenses |
| Participate in governance | Governance Case | exact immutable subject → feedback/Step Decision when allowed |
| Inspect Document history | Document History | Controlled Documents semantic history; exact referenced facts/content when authorized |
| Initiate/manage obsolescence | Document Official | current active request identity when disclosable → create/inspect/withdraw |
| Administer Organization | Admin / Organization | Company/User/Profile/binding/eligibility/Area/Group identity |
| Administer access | Admin / Access | GroupMembership + RoleAssignment |
| Administer document governance | Admin / Document Governance | DocumentType/governance/representation/Template configuration |
| Inspect Audit | Audit | page meaningful AuditEvent evidence |

No UI is created for a machine/system actor merely because a system transition exists. Release remains system-owned; provider/job/renderer state remains mechanism.

## 4. Stable route meanings

The Launch SPA route meanings remain exactly the accepted T6 set:

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

| Route | Lens | Meaning |
|---|---|---|
| `/documents` | Library | ordinary official/effective discovery + bounded Document creation |
| `/documents/:document_id` | Document Official | stable Document official/management lens; current official truth primary |
| `/documents/:document_id/work` | Document Work | exact current open Revision work lens |
| `/documents/:document_id/history` | Document History | authorized Controlled Documents history |
| `/work` | My Work | actor-relevant authoring/governance projections |
| `/work/governance/:attempt_id` | Governance Case | exact GovernanceAttempt + governed subject |
| `/audit` | Audit | AuditEvent evidence |
| `/admin/organization` | Admin / Organization | Organization identity/security administration |
| `/admin/access` | Admin / Access | effective access administration |
| `/admin/document-governance` | Admin / Document Governance | DocumentType/governance/Template configuration |

OIDC `/auth/login` and `/auth/callback` are browser integration routes outside this SPA Product tree and outside the 78-operation census.

No Launch route is added for Approvals, Templates, Distribution, Notifications, Taxonomy, Tokens, Sessions, Metrics, Search, Releases, Uploads, or provider/job state. Router library, nested-route implementation and visual navigation presentation are not durable contracts.

## 5. Exact 78-operation coverage

Every T8-E application operation has a concrete frontend consumer; no operation is invented.

```text
Shell/session                         1–2       = 2
Admin / Organization                 3–26      = 24
Admin / Access                       27–33     = 7
Admin / Document Governance          34–43,50–51 = 12
Library / creation                    44–46     = 3
Document Official / management        47–49,52,72–73,75,77 = 8
My Work / History supporting reads    53–56,76  = 5
Document Work                         57–66     = 10
Governance Case                       67–71     = 5
Official rendition                    74        = 1
Audit                                 78        = 1

TOTAL                                           = 78
orphaned accepted operations                    = 0
invented application operations                 = 0
operation 79                                    = absent
```

Primary homes do not transfer authority. Supporting reads may be consumed cross-lens only when current server disclosure permits them.

## 6. Vertical realization contracts

### 6.1 Application shell

```text
READ        getSession -> SessionView
WRITE       endSession
STATE       SessionView in server-state cache; local navigation/presentation only
SECURITY    HttpOnly cookie; csrf_token in memory/server-state cache
AUTHORITY   no frontend permission matrix
```

Launch navigation presence is **not a permission-filtering correctness requirement** because `SessionView` contains no effective-permission snapshot. The shell may present the stable Product spaces; entering a lens relies on current server 403/404 disclosure/authorization. Any future affordance suppression must derive from actual server-returned truth and remains UX only.

### 6.2 Library + creation

```text
READ          listDocuments
              getDocumentCreationOptions
              getDocumentTypeNumberingPreview
WRITE         createDocument
URL STATE     q + admitted catalog filters + cursor context
SERVER STATE  DocumentPage / DocumentCreationOptionsView / NumberingPreviewView
FORM STATE    creation selections/title until accepted
IDEMPOTENCY   one createDocument UUID per logical command; same key on retry
```

Ordinary creation selectors use `DocumentCreationOptionsView`, never Admin directories. Creation options are guidance; create revalidates canonical truth.

### 6.3 Document Official

```text
READ          getDocument -> DocumentOfficialView
              getDocumentResponsibleOwner when management is usable
              getRelease / getReleaseSource / getOfficialRenditionContent as referenced
              getObsolescenceRequest when active_obsolescence_request_id is disclosed
WRITE         replaceDocumentResponsibleOwner
              createDocumentRevision
              createObsolescenceRequest
              withdrawObsolescenceRequest
SERVER STATE  purpose-built official/current representations
```

The operator-approved T8-E precision, now folded into `wire-contract.md`, adds disclosure-safe routing references to `DocumentOfficialView`:

```text
open_revision?: { revision:RevisionIdentity, state:OpenRevisionState }
active_obsolescence_request_id?: Uuid
```

They are derived read truth, not persisted pointers. `open_revision` is present only when one current DRAFT/SUBMITTED Revision exists **and** the caller may receive working-context existence. `active_obsolescence_request_id` is present only when one ACTIVE ObsolescenceRequest exists **and** the caller may receive that request context.

Consequences:

```text
open_revision present  -> /documents/:document_id/work resolves directly
open_revision absent   -> UI may offer create-next-Revision only when current command authority permits
active request present -> inspect/withdraw path can resolve request_id directly
```

The route never silently renders DRAFT content in the official lens.

### 6.4 Document Work

```text
READ          getRevision / getRevisionDraft / getRevisionDraftSource
              getSubmission / getSubmissionSource when current work references Submission
WRITE         updateRevisionDraft
              startRevisionDraftUpload
              completeRevisionDraftUpload
              createSubmission
              withdrawSubmission
              cancelRevision
SERVER STATE  DocumentWorkView + supporting views
FORM STATE    local title/editor buffer bound to loaded DRAFT ETag
CONCURRENCY   stale DRAFT -> 412 precondition.draft_changed
IDEMPOTENCY   createSubmission preserves key + semantic DRAFT If-Match on retry
```

A 412 keeps unsaved local input, refetches authoritative DRAFT truth and requires explicit reconciliation. No silent LWW, automatic merge or client lifecycle truth.

Upload truth remains:

```text
provider PUT success != READY
READY != WorkingContent
WorkingContent != Submission
```

If `state.upload_expired` occurs during completion or attach, preserve local bytes and restart at a **new** `startRevisionDraftUpload` allocation; upload the same intended bytes again, complete the new allocation, then attach under the current DRAFT ETag. Never reuse or revive the expired allocation.

### 6.5 My Work

```text
READ       listAuthoringWork + listGovernanceWork
WRITE      none
STATE      WorkAuthoringPage + WorkGovernancePage
AUTHORITY  projections route to owner lenses; My Work owns no lifecycle truth
```

### 6.6 Governance Case

```text
READ          getGovernanceAttempt
              listGovernanceFeedback
              getGovernanceStepDecision
              exact governed subject/source through admitted reads
WRITE         createGovernanceFeedback
              recordGovernanceStepDecision
LOCAL STATE   feedback/decision form only
AUTHORITY     allowed_actions are hints only
```

The case renders exact immutable Submission/obsolescence subject content. Participation does not grant general WorkingContent/history authority. Governance Decision is not publish; Release remains system-owned.

### 6.7 History

```text
READ       getDocumentHistory + explicitly referenced supporting facts/content
WRITE      none by virtue of History
AUTHORITY  Controlled Documents semantic history; Audit remains separate
```

History is not used as a current-resource resolver for Document Work or active obsolescence.

### 6.8 Admin / Organization

Operations 3–26 are realized under their current authority. Company, UserProfile, provider binding, eligibility, Area and Group ETag domains remain independent; User concerns are not flattened into one client write model.

### 6.9 Admin / Access

Operations 27–33 realize GroupMembership and fixed-role RoleAssignment administration under `access.manage`. There is no custom Role/Permission editor.

### 6.10 Admin / Document Governance

Operations 34–43 + 50–51 realize DocumentType/governance/eligible-Template/Template-role administration. Configuration access does not imply governed content/history access.

### 6.11 Audit

```text
READ       listAuditEvents
WRITE      none
STATE      AuditEventPage
AUTHORITY  action evidence; never current business-state authority
```

Launch Audit is inspection/paging only. T8-F does not invent client-side or server-side filters absent from operation 78.

## 7. Frontend state authority

Launch baseline state classes:

```text
SERVER STATE       Product/server truth -> TanStack Query
NAVIGATION / URL   route/filter/navigation context -> router/URL
FORM DRAFT         unaccepted user input -> local feature/form state
EPHEMERAL UI       disclosure/focus/selection/dialog -> local React state
```

A fifth durable/global class requires a real consumer, protected property and proof the four classes are insufficient. No Redux/Zustand/global entity store is baseline. Transport loading/error is not Product lifecycle truth.

## 8. TanStack Query behavior

```text
query identity = operationId + canonical path/query semantic inputs
```

Different read models remain distinct query entries; no normalized universal Document entity becomes authority.

Pagination follows T8-E cursors exactly. No offset/page-number emulation, total-count requirement, generic filter/sort DSL or frozen client snapshot is introduced.

Mutation success:

```text
1. accept returned authoritative result
2. replace exact affected query representation when returned
3. invalidate/refetch only semantic lenses that can have changed
4. never fabricate lifecycle truth optimistically
```

There is no generic automatic mutation retry. Retry preserves the same logical command, Idempotency-Key and applicable conditional semantics. A materially changed or new deliberate command gets a new key.

## 9. Generated transport consumption

```text
api/openapi/v1/openapi.yaml
→ openapi-typescript paths/components
→ one thin lib/api transport
→ feature query/command functions
→ React lenses
```

Generated shapes are the wire type authority. Features do not maintain parallel DTOs, enums, Problem registries, route registries or header matrices.

The thin transport owns only browser wire mechanics:

```text
same-origin cookie credentials
CSRF carriage on unsafe requests
Idempotency-Key carriage from logical-command scope
If-Match / If-None-Match carriage from concurrency scope
JSON vs exact-byte handling
Problem decoding
response ETag preservation outside body DTO
```

It does not own Authorization, lifecycle, Product normalization, target eligibility or business retries. Provider capability URLs are opaque and are never parsed as Product identity.

## 10. ETag, idempotency and Problem behavior

### 10.1 ETag

An ETag-protected representation stores generated body + exact response ETag metadata. Forms bind to the tag they began from. `precondition.resource_changed` refetches authoritative state while preserving local input; `precondition.draft_changed` remains strictly DRAFT-specific.

### 10.2 Idempotency-Key

Each accepted idempotent POST gets one UUID per logical command:

```text
same logical command / ambiguous transport outcome -> same key
materially changed command                        -> new key
new deliberate command                            -> new key
```

Key is transport mechanism, never business identity.

### 10.3 Problems

Frontend branches on `Problem.code`, never localized `detail`.

```text
401 auth.unauthenticated
  -> invalidate session presentation / offer browser login

403 permission.denied
  -> denied context/action; server remains authority

403 permission.csrf_failed
  -> re-bootstrap `getSession`/csrf_token, then retry the SAME logical unsafe command only when safe;
     preserve the same Idempotency-Key and conditional semantics

404 notfound.*
  -> absent/non-disclosable according to server contract

409 state./conflict.*
  -> operation-specific state-conflict UX

410 state.upload_expired
  -> preserve intended local bytes; allocate a fresh upload capability and restart upload/admission

412 precondition.*
  -> explicit stale/current-truth reconciliation

422 validation.*
  -> input/business validation; field pointers may bind form fields

503 dependency.*
  -> dependency unavailable; never leak raw provider errors

500 internal.*
  -> generic failure + trace context; never leak provider/storage internals
```

A generic toast may present non-material failures but may not erase a state that changes the user's safe next action.

## 11. Read-model consumption

Purpose-built T6/T8-E models are consumed directly by their lenses, including SessionView, DocumentSummary/Page, DocumentOfficialView, DocumentWorkView, DocumentCreationOptionsView, SubmissionView, GovernanceCaseView, DocumentHistoryView/Page, WorkAuthoringItem/Page, WorkGovernanceItem/Page, AuditEventView/Page, GroupMemberPage/UserReference, TemplateConfigurationItem/Page and Admin configuration views.

The frontend creates no durable `DocumentEntity`, ApprovalDTO, ArtifactViewModel, generic ReferenceData model, lifecycle entity store or persisted currentStatus authority.

`allowed_actions` are UX hints only; commands recheck server truth.

## 12. Editor / viewer boundary

| Context | Exact authority | Frontend mode | Save/effect path |
|---|---|---|---|
| DRAFT DOCX | WorkingContent source | editable DOCX | complete bytes → upload/admit → DRAFT PATCH + If-Match |
| DRAFT PDF | WorkingContent source | read-only inspect/replace | replacement bytes → upload/admit → DRAFT PATCH + If-Match |
| SourceOnly official DOCX | Release source | DOCX adapter read-only | none |
| SourceOnly official PDF | Release source | read-only PDF | none |
| RequireOfficialRendition(PDF) | OfficialRendition primary; Release source separately available | read-only PDF | none |
| DRAFT PNG/JPEG | WorkingContent source | inline read-only image / replace | replacement bytes → upload/admit → DRAFT PATCH + If-Match |
| DRAFT XLSX/PPTX/TXT/CSV | WorkingContent source | download / replace — no in-app viewer | replacement bytes → upload/admit → DRAFT PATCH + If-Match |
| SourceOnly official PNG/JPEG | Release source | inline read-only image | none |
| SourceOnly official XLSX/PPTX/TXT/CSV | Release source | download — no in-app viewer | none |
| Governance | immutable Submission source | read-only decision content | governance never mutates content |

DOCX adapter law:

```text
load exact DOCX bytes
render/edit in browser
emit complete resulting DOCX bytes
support read-only mode
no provider-owned semantic identity
provider callback/save state != business truth
```

Save path:

```text
complete DOCX bytes
→ startRevisionDraftUpload(expected_size_bytes)
→ apply returned required_headers verbatim + PUT fixed body
→ completeRevisionDraftUpload
→ updateRevisionDraft(source_upload_id, If-Match=current ETag)
→ authoritative DocumentWorkView + new ETag
```

A format without an in-app viewer loses no governance: identity, status, revision, dates, history, discussion, audit and access render identically; only the content is downloaded (`../decisions/content-format-vocabulary.md` §5A).

No second editor, EditorSession, lease, CRDT baseline, viewer-generated OfficialRendition, third-party Office viewer, conversion-for-preview service or provider callback authority is introduced. Concrete editor/provider freeze remains subject to the already-ratified fidelity/security evidence and later runtime realization.

## 13. Feature/package topology

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

This follows semantic lenses, not Go packages/tables/endpoints. `/work` composes authoring/governance features and creates no Work owner. Admin co-location does not merge Organization/Access/Document Governance authority.

Do not add generic top-level `entities`, `domain`, `repositories`, `services`, `approvals`, `templates`, `workflows` or `stores` without a current distinct responsibility/protected property.

## 14. Navigation and authorization

Keep distinct:

```text
authenticated
!= permission
!= relationship/lifecycle eligibility
!= governance participation
!= command executable now
```

UI visibility is never authorization. There is no browser-maintained role/permission matrix. `listRoles` may display static role bundles only in Admin Access. Server 403/404 remains authoritative for direct navigation/disclosure.

## 15. Targeted reopen law

Legitimate triggers include:

```text
accepted human journey cannot be completed with operations 1..78
admitted operation lacks read data required to present/route an accepted decision safely
accepted Product distinction cannot be represented by current read models
editor requirement cannot satisfy exact-byte/OCC/fidelity/security boundary
current cursor/options semantics create demonstrated UX/scale failure
```

Response:

```text
identify smallest owner
→ bounded targeted reopen
→ update authority/proof
→ return to T8-F
```

Round-1 produced exactly one bounded upstream package: `DocumentOfficialView` read-symmetry precision for current open Revision and active obsolescence request identity. It adds no Product capability or operation.

Current corrected derivation:

```text
accepted operations covered       = 78 / 78
operation 79 required              = no
Product/T6 scope reopen            = no
bounded T8-E schema precision      = yes / operator-approved
```

## 16. Structural inversion / subtraction

Not introduced:

```text
new Product route/operation
frontend Authorization engine
legacy capability taxonomy
Redux/Zustand/global server-truth mirror
normalized Product entity authority
manual DTO/schema/Problem authority
generic workflow/Approval frontend
screen-shaped BFF
provider state as Product ontology
SSR/service-worker/offline correctness dependency
micro-frontends
design-system/router decision
dual DOCX editor
EditorSession/lease
optimistic lifecycle authority
generic Search API/frontend
operation 79
```

## 17. Concrete T8-G consumers

T8-F hands T8-G only concrete consumers:

```text
one React SPA delivery surface
same-origin /api/v1 access
OIDC redirect/callback integration outside application OpenAPI
HttpOnly session cookie + in-memory synchronizer CSRF bootstrap
exact-byte application-origin DOCX/PDF reads
browser direct-upload capability needing provider CORS for returned browser-settable headers
one interactive DOCX adapter boundary
read-only PDF presentation
```

T8-F does not choose binaries/processes, runtime host, deployment topology, CDN, container layout, provider CORS configuration, secrets, workers, readiness, observability, recovery profiles or buffering strategy. No SSR, service worker, BFF, websocket/realtime channel or editor-session service is a Launch runtime consumer.

## 18. Ratification record

T8-F was explicitly operator-ratified on 2026-08-21 after repository verification, Fable Round-1 adjudication and bounded Fable Round 2 **CONVERGED**. This is an immutable closure record, not current program-status authority.

Minimum preserved properties:

```text
stable T6 route meanings          exact
application operations            78 / 78
operation 79                      absent
frontend semantic owner           none added
server truth                      TanStack Query / server authority
parallel global server store      absent
manual DTO/API authority          absent
frontend AuthZ matrix             absent
Document Work resolver            disclosure-safe open_revision read reference
active obsolescence resolver      disclosure-safe request id read reference
DRAFT OCC                         strong ETag; stale -> explicit reconciliation
logical idempotent retry          same key
exact-content descriptor          server-owned
DOCX interactive runtime          one adapter boundary
Governance content                immutable/read-only
```

## 19. Graduated realization laws — T11 frontend Evidence

Graduated from repeated operator-LOCKED behavior proved by at least two independent blocks and routed by `../decisions/t11-b11-lock-evidence.md` and `../decisions/t11-b12-lock-evidence.md` (with earlier occurrences across the `../decisions/t11-b01-b09-lock-evidence.md` / `../decisions/t11-b10-lock-evidence.md` packages). These are durable class-level realization laws for production frontend work, not a component framework, checker or new semantic owner.

### 19.1 Honest bounded-collection law

Every cursor-paginated (`PAGED`) collection surface:

```text
render exactly the returned page in the server's returned order
+ show whether continuation exists (server-owned cursor / has_more only)
+ continuation uses the returned cursor; first-page filters are never repeated on continuation
+ a transient/retryable continuation failure preserves the current page and offers explicit
  retry; an authorization/disclosure failure on continuation (the cursor contract rechecks
  AuthZ on every page) follows its owning Problem behavior instead — protected content is
  replaced by the denial presentation, never left visible behind a retry
+ no invented total, offset, reverse-cursor assumption or hidden complete-universe claim
+ server-side filter/search executes before pagination; changing any filter starts a new
  first-page identity
+ an empty page under a valid filter identity is an ordinary result, never an error and
  never proof of non-existence
```

Proved by B11 (op6/op22/op31 traversal and op31 server-side filter identity) and re-proved by B12 (op34/op43 traversal and the ratified op43 filter identity). Shared low-level rendering may exist at implementation time only while each operation's exact first-page/continuation law remains explicit; no universal pager abstraction is implied.

Deliberately complete bounded reads (e.g. selection preflights such as `ProviderSubjectSearchView`, or the `document-creation/options` arrays) are not paginated collections: they render the complete returned set and must not invent continuation state or misrepresent the bounded response.

### 19.2 Idempotent-creation recovery law

Every semantic `IDEMPOTENT_CREATE` command surface:

```text
one logical intention = one client-generated Idempotency-Key
+ the composed input is frozen while an outcome is ambiguous
+ an ambiguous transport outcome retries the SAME normalized command with the SAME key
+ within the wire contract's semantic replay window, a committed retry recovers the exact
  stored result (same status/body/identity); semantic mutation count stays 1 → 1
+ once that window may have expired, the same UUID is no longer replay-authoritative
  (`wire-contract.md` §2.5): recovery goes through read reconciliation, never another
  blind create with the expired key
+ a second intention is a new key; a silent duplicate command never occurs inside the
  replay-authoritative window
+ semantic conflicts (e.g. duplicate code) surface the server's named cause with zero mutation
```

Proved by B11 (op32 grant creation, completed-replay and ambiguous recovery) and re-proved by B12 (op35 type creation with duplicate-code 409 and same-key replay). The durable Idempotency-Key wire law remains owned by `wire-contract.md`; this law binds its user-facing realization.

### 19.3 Fixture-truthfulness law

Planning prototypes and, later, tests/storybook-class harnesses:

```text
a fixture may only simulate truth the accepted contracts actually supply
+ a state the current read models cannot express is never simulated through an invented flag
  or precomputed client inference; during planning, a materially needed missing truth is first
  an upstream finding under the frontend method (§3.10A/§13) — only after the Product decision
  owner adjudicates that accepted authority intentionally owns the failure/absence path
  (as with B12-F1) does the honest failure/absence presentation become the accepted realization
+ prototype fixtures, simulated servers, fake cursors and mutation counters are Evidence
  mechanics only — they never become Product state, read-model authority or production code
+ production realizes LOCKED semantics through the accepted architecture; it does not port
  the simulator or its data
```

Proved by B11 (raw op6 eligibility truth, 201/204 reconciliation without client precomputation) and re-proved by B12 (B12-F1: code/scope immutability presented via the honest op37 409 path after the operator rejected an invented in-use flag).

A violation of any law above in later planning, review or implementation is a defect against accepted architecture, not a stylistic preference.

Current integration, stage progression, implementation permission and exact next action are owned exclusively by `../roadmap.md`.
