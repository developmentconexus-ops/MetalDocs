# R10-T6 — Canonical API / Frontend Journeys — Material Decision Candidate

> **Status:** ACTIVE STAGING / NON-AUTHORITATIVE — **T6 MATERIAL DECISION CANDIDATE / OPERATOR ADJUDICATION NEXT**  
> **Date:** 2026-08-18  
> **Repository:** `developmentconexus-ops/MetalDocs`  
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131  
> **Product Contract:** `wiki/architecture/launch-v1-product-contract.md` — REV001  
> **Decision baseline:** `wiki/architecture/rebaseline-decision-registry.md`  
> **Implementation:** BLOCKED

This candidate derives Launch API and frontend journey architecture from Product Contract REV001 + ratified T1→T5. It does **not** preserve current routes, modules, DTOs, screens, capabilities, editor sessions or current OpenAPI shapes merely because they exist.

Current implementation is evidence only. The target is allowed to delete/rewrite every current public route and screen if that is the smallest sustainable solution.

---

# 0. Authority and evidence posture

Authority order:

1. `AGENTS.md`
2. DevelopmentConexus Engineering Method v1.0.0 mirror
3. current handoff
4. Product Contract REV001
5. Whole-Product GCR
6. ownership topology
7. T1→T5 durable authorities
8. Decision Registry
9. T6 router/bootstrap
10. this candidate
11. current API/frontend/provider evidence only to falsify or validate a specific claim

Binding inversion law for this stage:

> If the current API/frontend did not exist, the candidate below should still follow from the ratified product semantics.

Therefore T6 deliberately rejects “migration cost”, “existing route”, “current module”, “already implemented capability”, “current UI tab” and “current generated type” as reasons to preserve a target shape.

Evidence that survives Structural Inversion and may be preserved as mechanism/pattern evidence:

```text
OpenAPI contract-first discipline
RFC 9457 Problem Details
snake_case wire vocabulary
typed generated backend/frontend HTTP types
feature-sliced SPA organization
TanStack Query as replaceable server-state mechanism
PDF.js-class direct PDF viewer
browser-buffer DOCX editor/viewer class
```

Evidence that does NOT survive as target authority:

```text
separate public ControlledDocument object/module
separate Approval public module/workspace
separate Template lifecycle/routes
Distribution routes/state
Taxonomy/token platforms
old capabilities/roles
Tenant/RLS assumptions
writer-session correctness dependency
scheduled publish
universal PDF/export machinery
legacy document status vocabulary
current DTO fields and route hierarchy
```

---

# 1. Macro approaches

## Option A — preserve current surface and rename/subtract

Shape:

```text
current modules/routes/screens
→ remove obvious legacy concepts
→ rename remaining nouns to T1→T5 vocabulary
```

Rejected.

Root defect: the existing surface was shaped by now-superseded ownership and lifecycle decisions. Subtraction leaves hidden semantic coupling: `Approval` still looks like a peer product, Template still looks like a peer object, one route can switch meaning by actor/status, and old capability/navigation concepts remain the frontend mental model.

This is the Local Maximum.

## Option B — one polymorphic super-workspace / generic command API

Shape:

```text
one document route
one giant context payload
one generic action endpoint
frontend derives mode dynamically
```

Rejected.

It reduces route count but increases semantic ambiguity. A URL must not mean “current official content”, “mutable DRAFT”, or “exact governed Submission” depending on who opens it. Generic action endpoints also hide distinct idempotency, authorization and lifecycle contracts.

## Option C — semantic lenses over one core model — RECOMMENDED

Shape:

```text
Document official lens
Document work lens
Governance case lens
Document history lens
My Work lens
Audit lens
Admin lens
```

The lenses reference the same T1→T5 semantic facts but expose different truthful views. They do not create new owners.

This is the candidate Global Maximum: explicit truth with a small number of stable user/API concepts.

---

# 2. T6-A — Public contract discipline

## Candidate: ACCEPT

MetalDocs Launch exposes one versioned HTTP contract whose source of truth is OpenAPI.

Laws:

```text
OpenAPI public contract first
→ generated backend boundary types
→ generated frontend transport types
→ implementation conforms
```

Preserve as target conventions:

```text
/api/v1 public prefix
JSON fields = snake_case
stable opaque UUID technical ids
UTC RFC3339 timestamps
typed response bodies; no ad-hoc map payloads
same-origin ApplicationSession cookie for the web SPA
provider/JWT claims never Authorization authority
```

OpenAPI/generator versions are mechanism details and may be reselected at implementation time; the contract-first law is the durable decision.

The public API is organized by product semantics, never backend package/module names.

No public route is retained because current code depends on it.

---

# 3. T6-B — Frontend information architecture

## Candidate: ACCEPT semantic-lens navigation

Launch top-level product spaces:

```text
Library
My Work
Audit                  when allowed
Administration         when allowed
```

No Launch top-level product space for:

```text
Approvals
Controlled Documents
Templates
Distribution
Notifications
Taxonomy
Tokens/Dictionary
Sessions
Metrics/Usage
```

Templates remain ordinary Documents with a governed Template role/eligibility. Governance work is part of `My Work`. Audit is evidence inspection, not administration/configuration.

Recommended stable frontend route meanings:

```text
/documents
  Library: ordinary current-effective discovery

/documents/:document_id
  stable Document overview; official/effective truth is primary

/documents/:document_id/work
  current open Revision authoring lens; authorized actors only

/work
  actor-relevant author/governance worklist

/work/governance/:attempt_id
  exact GovernanceAttempt + exact Submission decision lens

/documents/:document_id/history
  authorized domain history

/audit
  AuditEvent evidence workspace

/admin/organization
/admin/access
/admin/document-governance
```

Exact component hierarchy remains implementation design.

### Why not one mode-adaptive `/documents/:id`?

Because an address should have stable semantic meaning. Opening `/documents/:id` must not sometimes render mutable DRAFT bytes and sometimes current EFFECTIVE truth based on caller identity or current workflow state.

Visual shell reuse is encouraged; semantic route ambiguity is not.

---

# 4. T6-C — Canonical query/command surface

## Candidate: ACCEPT semantic operations, reject module-shaped API

The API should expose the smallest set of product-semantic operations needed by Launch journeys.

Conceptual query families:

```text
Library / current-effective Documents
Document official overview
Document current open work
Document history
exact Submission
exact GovernanceAttempt/case
My Work
Audit events
Organization/AuthZ/config administration
```

Conceptual command families:

```text
create Document + REV000
create next Revision
update DRAFT state
SUBMIT exact DRAFT generation
withdraw Submission
cancel Revision
record governance ACCEPT / RETURN_FOR_CHANGES
initiate obsolescence
withdraw active human-governed obsolescence request
change responsible owner
Organization/AuthZ/config mutations
managed-content allocation/finalization
```

Recommended resource/command route family examples:

```text
POST /api/v1/documents
GET  /api/v1/documents
GET  /api/v1/documents/{document_id}
GET  /api/v1/documents/{document_id}/work
PATCH /api/v1/documents/{document_id}/work
POST /api/v1/documents/{document_id}/revisions
POST /api/v1/documents/{document_id}/submissions
GET  /api/v1/documents/{document_id}/history

GET  /api/v1/submissions/{submission_id}
POST /api/v1/submissions/{submission_id}/withdraw

POST /api/v1/revisions/{revision_id}/cancel

GET  /api/v1/governance-attempts/{attempt_id}
POST /api/v1/governance-attempts/{attempt_id}/decisions

POST /api/v1/documents/{document_id}/obsolescence-requests
POST /api/v1/obsolescence-requests/{request_id}/withdraw

GET  /api/v1/work
GET  /api/v1/audit/events
```

These examples are candidate canonical route families, not permission to implement them yet.

There is no generic `/actions` endpoint and no arbitrary workflow-subject API.

---

# 5. T6-D — Server-derived UX affordances

## Candidate: ACCEPT

The frontend must not reconstruct T3 Authorization or lifecycle predicates from roles/status strings.

Read models may include bounded server-derived `allowed_actions` semantic affordances, for example:

```text
edit_draft
submit
withdraw_submission
cancel_revision
change_owner
start_obsolescence
withdraw_obsolescence
governance_accept
governance_return
```

Laws:

```text
allowed_actions = UX hint only
allowed_actions != grant
allowed_actions may become stale
command endpoint always rechecks T2/T3 canonical truth
```

Session/bootstrap may similarly expose coarse workspace/navigation affordances rather than raw client-computed role logic.

Do not expose a materialized ACL as a public source of truth.

---

# 6. T6-E — Create journey + numbering

## Candidate: ACCEPT structured numbering; reject generic format grammar

Create journey:

```text
choose Document Type
→ choose Area
→ enter REV000 title
→ optionally choose eligible current-EFFECTIVE Template
→ responsible owner defaults to actor unless manager selects another eligible User
→ show non-reserving code preview
→ one atomic create command
→ navigate to Document work lens
```

### Numbering configuration

Do not expose a general string/token formatting language.

Use a closed structured configuration:

```text
numbering_scope:
  DOCUMENT_TYPE
  DOCUMENT_TYPE_AREA

sequence_width:
  bounded integer, default 3

separator:
  fixed "-"
```

Result:

```text
DOCUMENT_TYPE
  {TYPE}-{SEQ}
  example: PO-001

DOCUMENT_TYPE_AREA
  {TYPE}-{AREA}-{SEQ}
  example: PO-RH-001
```

No Launch support for:

```text
{YEAR}
monthly/annual reset
free-form literal grammar
random suffix
custom JavaScript/expression numbering
preview reservation
```

If `DOCUMENT_TYPE_AREA` is enabled, Area requires a stable short code usable for future numbering; renaming display text does not rewrite committed Document codes.

### Preview law

```text
preview = best current next-code illustration
preview reserves nothing
another commit may consume the shown sequence first
final committed code is returned by atomic create
committed code never reuses
```

No promise of gap-free numbering.

---

# 7. T6-F — DRAFT title/content mutation and concurrency

## Candidate: ACCEPT one existing T2 OCC law

The post-Fable observation is closed by placing all mutable governed DRAFT editing under the existing WorkingContent generation/CAS mechanism.

Title remains Revision-owned governed metadata; concurrency mechanism does not change ownership.

Conceptual DRAFT mutation:

```text
expected_generation
+ optional new title
+ optional READY managed_content_id / exact descriptor attachment
→ compare current WorkingContent generation
→ atomically update accepted DRAFT fields
→ generation++ exactly once
```

Consequences:

```text
content edit and retitle cannot silently overwrite a newer DRAFT mutation
SUBMIT freezes one coherent generation + title
no second draft_version token is required
```

A title-only change still advances DRAFT generation.

### Conflict UX

Stale generation:

```text
server rejects
frontend stops silent autosave
local unsaved buffer remains local/recoverable
user reloads/reconciles deliberately
```

No silent last-write-wins and no baseline auto-merge.

---

# 8. T6-G — Source upload / T4 admission journey

## Candidate: ACCEPT 4-step browser journey

T4 remains authority. The frontend merely orchestrates mechanism states.

```text
1. allocate upload
   server creates OPEN handle + bounded admission claim bound to intended DRAFT/root

2. browser uploads exact bytes directly to managed-content provider

3. finalize admission
   server independently reads/verifies exact bytes
   → SHA-256 + size + ContentFormat
   → READY

4. attach READY content to current DRAFT
   WorkingContent OCC/CAS transaction
```

Frontend UX states may be:

```text
uploading
verifying
ready
saving
saved
error
```

They are not business lifecycle states.

### Malware boundary

DRAFT upload does not require scanner success on every autosave.

SUBMIT/import governed-boundary preflight performs the required exact-byte malware check server-side before the semantic transaction. Scanner unavailable is a visible retriable dependency error; malicious bytes cannot cross the boundary.

The browser never talks to the malware scanner directly.

### Failure handling

```text
provider upload succeeds + attach fails
→ READY mechanism object remains retry/reclaim candidate
→ no false Submission truth
```

The frontend never invents a “confirmed document” state from upload success.

---

# 9. T6-H — Author work + governance work

## Candidate: ACCEPT unified `My Work`, separate exact governance-case lens

`My Work` aggregates actor-relevant actionable/readable items without becoming lifecycle authority.

Candidate sections:

```text
Drafts / Returned — author action required
Submitted / Waiting — author visibility, no false action
Governance — currently active Steps the actor may act on
```

Area managers may use filters to inspect manageable work within scope.

### Governance case

A governance participant opens the exact attempt, not a generic document cockpit:

```text
GovernanceAttempt
+ active Step
+ exact immutable Submission
+ bounded prior decisions/feedback
+ exact server-derived allowed_actions
```

The primary content displayed is the exact Submission content, never current WorkingContent.

Decision surface:

```text
ACCEPT
RETURN_FOR_CHANGES
```

No generic signoff/quorum/reassign/SLA UI.

### Author lifecycle actions

The Document work lens exposes only currently allowed actions such as:

```text
submit
withdraw active Submission
cancel open Revision
```

Obsolescence initiation/withdrawal belongs to the stable Document official/management lens, not the DRAFT editor.

---

# 10. T6-I — In-product view vs exact-source download

## Candidate: ACCEPT explicit representation labels

Reader-facing current-effective view:

```text
SourceOnly PDF
  → direct PDF viewer
  → exact source download

SourceOnly DOCX
  → direct read-only DOCX viewer
  → exact source download

RequireOfficialRendition(PDF)
  → Official PDF is primary in-product official view
  → source remains separately available as exact governed source
  → UI labels source vs official rendition explicitly
```

No viewer/preview output becomes semantic OfficialRendition.

Historical/governance views identify exactly which Revision/Submission/Rendition is being inspected.

### Download contract

Frontend asks MetalDocs for an authorized exact-content read/download capability. Provider key/bucket/version is never exposed as product identity.

Implementation may stream through MetalDocs or return a short-lived provider URL; either remains mechanism.

---

# 11. T6-J — DOCX editor/viewer provider strategy

## Candidate: ACCEPT provider-neutral browser-buffer baseline; no EditorSession baseline

Required editor adapter behavior:

```text
load exact DOCX bytes
render/edit in browser
emit complete resulting DOCX bytes
read-only mode for governed inspection
no provider-owned durable document identity
no provider save/callback state as business truth
```

Current candidate evidence:

- EigenPal/docx-editor exposes browser DOCX editing around a document buffer and explicit editing/read-only modes.
- ONLYOFFICE Docs provides high-fidelity server-backed editing but requires a document-storage integration/callback URL for save/status flow.
- PDF.js remains a straightforward direct PDF viewer class.

Architecture consequence:

```text
reference Launch editor/viewer candidate = EigenPal-class browser-buffer adapter
final provider freeze = after representative DOCX fidelity corpus
```

If the corpus proves the browser-buffer candidate unsustainable, ONLYOFFICE-class integration may be selected, but then:

```text
ONLYOFFICE document key/session/callback = mechanism only
callback output must pass T4 admission
WorkingContent OCC remains final DRAFT acceptance authority
provider callback must never overwrite newer accepted generation
```

### EditorSession decision

Launch baseline: **no EditorSession/lease correctness dependency**.

WorkingContent OCC is sufficient for correctness.

A bounded editor lease/session reopens only if the selected integration proves a concrete UX/provider need that cannot be represented through ordinary ApplicationSession + WorkingContent generation.

Realtime coauthoring remains Future/CRDT.

---

# 12. T6-K — Search / Library

## Candidate: ACCEPT canonical PostgreSQL query; do NOT activate materialized Search

Named Launch search facts are canonical:

```text
Document code
current EFFECTIVE Revision title
Document Type
Area
responsible owner
current lifecycle/effectivity status appropriate to the workspace
```

Therefore T6 finds no current derived/expensive searchable fact that justifies T5 materialization.

Decision:

```text
materialized Search projection = NOT ACTIVATED for Launch
search_refresh = NOT ACTIVATED for Launch
full Search rebuild = NOT APPLICABLE in baseline
```

### Ordinary Library truth

`Library` is current-effective discovery.

It never returns DRAFT/SUBMITTED work as equivalent official results.

Candidate free-text query:

```text
q searches code + current-effective title only
```

Candidate deterministic ranking when `q` exists:

```text
exact code
→ code prefix
→ title prefix
→ title contains
→ stable code tie-breaker
```

Filters:

```text
Document Type
Area
responsible owner
```

Obsolete/superseded/historical content belongs history/governance lenses, not default ordinary Library truth.

No Launch full-text body search over DOCX/PDF content.

Full-text body search is an explicit future T5/T6 materialization reopen trigger.

---

# 13. T6-L — History vs Audit

## Candidate: ACCEPT two distinct inspection surfaces

### Domain history

Per-Document history is derived from Controlled Documents facts:

```text
Revisions
Submissions
GovernanceAttempts / decisions
feedback
cancellations
Releases
OfficialRenditions
obsolescence request/result/withdrawal
```

This answers “what happened to this Document?”

### Audit workspace

AuditEvent workspace answers “what meaningful actions occurred across the product?” and supports only authorized T3 visibility.

Candidate filters:

```text
time range
actor
operation family
Area when authorized
resource kind/id
```

Audit never reconstructs current lifecycle truth.

The frontend must not merge domain history and Audit into one synthetic source-of-truth timeline. They may be visually cross-linked while ownership remains explicit.

---

# 14. T6-M — Administration journeys

## Candidate: ACCEPT one Admin Center with three semantic sections

### Organization

```text
Company current display/settings where needed
People: User + erasable UserProfile lifecycle
Areas
Groups
Group memberships inside Group/Person detail
```

### Access

```text
RoleAssignment grant/revoke
subject = User | Group
role = static product role
scope = Company | Area according to T3 matrix
```

No custom Role/Permission editor.

### Document Governance

```text
Document Types
activation/inactivation
numbering configuration + preview
NoHumanApproval | UseGovernanceRoute
sequential route Steps: business label + NAMED_USER | GROUP
SourceOnly | RequireOfficialRendition(PDF)
Template role/eligibility management
```

No browsable generic PolicyVersion platform; in-flight snapshots remain immutable by T2.

### Deliberately absent Admin areas

```text
Sessions administration
Metrics/usage product
Notifications
Distribution
Taxonomy/Dictionary platform
Approval engine admin
generic workflow designer
Records/Hold/Retention
```

Audit is an inspection workspace, not configuration.

---

# 15. T6-N — Public error contract

## Candidate: ACCEPT RFC 9457 Problem Details + closed semantic code families

All public errors use `application/problem+json`.

Minimum shape:

```text
type
title
status
detail
instance
code
errors[] optional field-level details
```

Machine `code` is stable and semantic, never module-named.

Candidate closed families:

```text
request.       malformed transport/request
validation.    well-formed but invalid input

 auth.          unauthenticated/session issue
permission.    authenticated but action not allowed
notfound.      absent or intentionally invisible resource
state.         canonical lifecycle/state disallows command
precondition.  caller precondition is stale
conflict.      concurrent writer / uniqueness conflict
dependency.    storage/scanner/renderer/provider temporarily unavailable
internal.      server fault
```

Typical status mapping:

```text
400 request
401 auth
403 permission
404 notfound
409 state/conflict
412 precondition
422 validation
503 dependency
500 internal
```

Read paths may return 404 rather than reveal the existence of a resource the caller cannot inspect. Commands on an already-visible resource may return 403 for denied action.

No raw provider/storage/scanner error leaks to clients.

---

# 16. T6-O — Public idempotency and concurrency preconditions

## Candidate: ACCEPT semantic rule, not blanket middleware

Idempotency and concurrency solve different problems.

### Idempotency-Key

Require `Idempotency-Key` for POST commands that create a new durable semantic fact and are not naturally idempotent, including at least:

```text
create Document
create next Revision
create Submission
record governance decision
create RevisionCancellation
create obsolescence request
withdraw active obsolescence request
```

Exact retained-key TTL/storage is implementation detail.

Same key + same operation/request replays the original committed result. Same key + materially different request fails explicitly.

### Natural idempotency for DRAFT autosave

High-frequency DRAFT mutation does not require a durable idempotency row per autosave.

Use WorkingContent OCC plus target-handle identity:

```text
first accepted expected_generation + target state → generation advances
exact retry where target is already current → may return the already-accepted result
stale competing mutation → conflict; never overwrite
```

### Mutable admin configuration

Mutable current configuration should expose an opaque HTTP precondition/version where lost-update protection is required (for example ETag/If-Match). Exact database versioning is implementation detail.

Lifecycle commands target exact stable semantic IDs and revalidate canonical eligibility; no generic lifecycle ETag is required merely by convention.

---

# 17. T6-P — Lists, pagination, filters and response shape

## Candidate: ACCEPT simple typed envelopes

Single-resource responses return the resource/read model directly.

Unbounded list responses use:

```text
items: [...]
page:
  next_cursor: opaque|null
  has_more: boolean
```

No mandatory total count.

Cursor is opaque and bound to the query/sort inputs so changing filters restarts pagination.

Contractually bounded/static vocabularies need no pagination.

Filters are explicit typed query parameters; arbitrary `filter[field]`/generic query DSL is rejected.

Sort options are a closed per-workspace vocabulary, not arbitrary SQL field exposure.

---

# 18. T6-Q — Read-model design law

## Candidate: ACCEPT purpose-built semantic read models

Do not publish database tables or internal aggregate shapes directly.

Public read models should be shaped around a user lens, for example:

```text
DocumentSummary
DocumentOfficialView
DocumentWorkView
SubmissionView
GovernanceCaseView
DocumentHistoryView
WorkItem
AuditEventView
```

They may denormalize labels/reference display data for UX, but canonical identifiers remain stable and the read model never becomes mutation authority.

A query DTO may include server-derived `allowed_actions` as T6-D specifies.

No universal `ArtifactViewModel`, `ApprovalInstanceDTO`, or generic polymorphic resource envelope is required.

---

# 19. T6-R — Frontend technical organization

## Candidate: PRESERVE mechanism pattern, replace semantic feature taxonomy

The feature-sliced SPA pattern survives inversion:

```text
app composition
features/<semantic-lens-or-owner>
lib API/client/generated types
components/ui domain-agnostic primitives
TanStack Query server state
local React state for local UI state
```

But the target feature vocabulary should follow the new product surface, not legacy modules.

Candidate target feature groups:

```text
library
document-work
governance-work
history
audit
admin
auth/shell
shared editor/viewer primitives
```

Do not create frontend features merely to mirror backend package names.

Generated OpenAPI types remain the transport boundary; feature-specific view models may adapt them for UI where genuinely useful.

---

# 20. Integrated user journeys

## Create → author → submit

```text
Library/My Work → New Document
→ type/area/title/template/owner
→ non-reserving code preview
→ atomic create
→ /documents/:id/work
→ upload/editor autosave under WorkingContent generation
→ submit
→ exact immutable Submission
```

## Return → change → resubmit

```text
RETURN_FOR_CHANGES
→ My Work shows Returned
→ /documents/:id/work loads same Revision DRAFT
→ old Submission/history remains immutable
→ author edits under generation OCC
→ new submit creates new Submission
```

## Governance

```text
My Work / Governance
→ exact attempt
→ exact immutable Submission viewer
→ ACCEPT | RETURN_FOR_CHANGES
→ command revalidates T3/T2
```

## Reader

```text
Library
→ current-effective search
→ /documents/:id
→ official view
→ exact source and/or OfficialRendition clearly labeled
```

## Revise

```text
current EFFECTIVE Document
→ authorized create-next-Revision
→ /documents/:id/work
→ prior effective remains reader truth
→ new Release atomically supersedes predecessor
```

## Obsolete

```text
/documents/:id management actions
→ reason + initiate governed obsolescence
→ human route if configured
→ initiator/manager may withdraw while active
→ success makes current Revision OBSOLETE
→ Library no longer presents as current effective
```

## History/Audit

```text
Document history
→ domain truth by Revision/Submission/Release

Audit
→ action evidence under T3 visibility
```

---

# 21. Structural-Inversion / subtraction results

If current implementation were deleted today, this candidate would still choose:

```text
semantic-lens routes instead of module routes
OpenAPI contract-first
Problem Details
WorkingContent OCC for DRAFT title+content
server-derived affordances
structured numbering
T4 upload/admission flow
exact Submission governance lens
current-effective Library
canonical PostgreSQL Search
separate domain history vs Audit
minimal Admin Center
no EditorSession baseline
```

Subtractive deletions implied by the target include, without promising implementation timing:

```text
separate Approvals product workspace/module contract
separate Templates lifecycle contract
public ControlledDocument peer object
Distribution UI/routes in Launch
Taxonomy/token/dictionary admin platform absent promoted requirement
Sessions/metrics admin screens
legacy role/capability-based frontend gate vocabulary
legacy mode-adaptive route ambiguity
writer-session correctness dependency
materialized Search without consumer
scheduled publish UI/API
```

The existence of any of these today is not an argument for retaining them.

---

# 22. Material candidate decisions for operator adjudication

```text
T6-A  OpenAPI remains single public HTTP contract; semantic routes, typed generated boundaries.
T6-B  UI uses semantic lenses: Library / My Work / Audit / Administration + explicit Document/work/governance/history routes.
T6-C  API exposes semantic query/command families; no generic action endpoint/module-shaped contract.
T6-D  Server-derived allowed_actions drive UX; frontend never reimplements T3 rules.
T6-E  Numbering = closed structured TYPE or TYPE_AREA + fixed '-' + bounded sequence width; preview never reserves.
T6-F  DRAFT title + content mutations share WorkingContent generation OCC.
T6-G  Upload = allocate OPEN → direct upload → server finalize READY → OCC attach; malware only at governed boundary.
T6-H  My Work unifies author/governance work; exact governance case is separate lens over exact Submission.
T6-I  Reader/viewer explicitly distinguishes source vs OfficialRendition; direct PDF/DOCX viewing as T5 allows.
T6-J  Provider-neutral browser-buffer DOCX adapter baseline; EigenPal-class candidate; no EditorSession baseline; ONLYOFFICE-class fallback only on fidelity evidence.
T6-K  Search materialization NOT activated; Launch q = code/title + canonical filters/ranking.
T6-L  Domain history and Audit remain separate read surfaces.
T6-M  Admin Center = Organization / Access / Document Governance; no custom-role/workflow/platform admin surfaces.
T6-N  RFC9457 Problem + closed semantic code families incl. dependency 503.
T6-O  Idempotency-Key only for new durable-fact POST commands; OCC/natural replay for autosave; preconditions for mutable config.
T6-P  Typed list envelope + opaque cursor for unbounded lists; no generic filtering DSL or mandatory totals.
T6-Q  Purpose-built semantic read models; no DB/module DTO leakage.
T6-R  Preserve feature-sliced SPA/TanStack Query mechanism pattern, replace legacy frontend feature taxonomy.
```

Everything outside these T6 decisions remains frozen unless a material counterexample explicitly reopens it.

---

# 23. Proof obligations before T6 closure

A T6 design cannot close until it can falsifiably demonstrate:

```text
ordinary reader URL cannot accidentally show DRAFT/SUBMITTED truth as official
exact governance link cannot drift to current WorkingContent
frontend allowed_actions cannot grant authority
DRAFT retitle and content edit cannot silently overwrite each other
preview code consumes no sequence and committed code never reuses
upload success alone cannot become governed truth
stale admission handle cannot attach cross-root
scanner/provider outage produces truthful retriable error, not fake lifecycle state
late/dead rendition cannot appear as official
SourceOnly viewing never manufactures OfficialRendition
current-effective search cannot surface DRAFT as official
T6 requires no materialized Search consumer
Audit UI never reconstructs lifecycle state
frontend never computes access from raw roles as canonical truth
idempotent command retry cannot duplicate Submission/Decision/Cancellation/Obsolescence facts
autosave retry cannot require one durable idempotency record per keystroke/save
EditorSession can be absent without violating DRAFT correctness
current implementation may be deleted/replaced without changing any target invariant above
```

---

# 24. Explicit non-decisions

T6 candidate does not yet freeze:

```text
exact OpenAPI schema text
operationId names
all endpoint path spellings beyond candidate families
exact Problem code catalog
exact cursor encoding
exact idempotency retention TTL/table
ETag database representation
exact React component tree
visual styling/design tokens
EigenPal vs another browser-buffer editor final vendor freeze
ONLYOFFICE version/deployment
PDF.js version
renderer/converter product
SQL/indexes
Go package/module layout
Historical Migration API/execution
implementation plan
```

Provider final selection requires the already-ratified representative DOCX fidelity corpus before implementation.

---

# 25. Reopen triggers

Reopen only the implicated T6 decision if material evidence proves:

```text
a real external API client requires a transport incompatible with same-origin ApplicationSession-cookie web baseline
one canonical Document route can safely remain polymorphic without semantic ambiguity and materially reduces user cost
real numbering requirements need year/reset/custom grammar
browser-buffer DOCX editing fails representative fidelity/security/scale requirements
selected editor requires a bounded lease/session for correctness at the integration boundary
content-body full-text Search is a promoted Launch requirement
canonical PostgreSQL Search cannot meet measured Launch needs
reader needs source hidden when OfficialRendition exists for a concrete compliance requirement
Audit/domain-history separation prevents a named investigation journey
public commands need a different idempotency guarantee
cursor pagination creates a demonstrated UX/operability regression
```

No legacy implementation shape is itself a reopen trigger.

---

# 26. Current gate

```text
T6 bootstrap                    OPEN
T6 evidence/inversion pass      COMPLETE ENOUGH FOR CANDIDATE
T6 material candidate           THIS FILE / NON-AUTHORITATIVE
operator material adjudication NEXT
platform-facing T6 summary      NOT YET
T6 durable promotion            NOT YET
T7                              NOT OPEN
implementation                  BLOCKED
```

Do not implement from this candidate. Operator adjudication comes first.