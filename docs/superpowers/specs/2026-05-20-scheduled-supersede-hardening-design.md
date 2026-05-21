# Scheduled Supersede Hardening Design

Date: 2026-05-20
Scope: modern documents + approval publish scheduling, canonical sibling CTA behavior, and scheduled supersede lineage safety
Status: proposed

## Goal

Close the remaining lifecycle gap in the modern publish flow so a scheduled publication that replaces an already-published lineage head behaves with the same rigor as an immediate supersede:

- explicit contract truth
- explicit governed persistence
- no ambiguous cutover behavior
- truthful canonical UI navigation
- no legacy fallback

This design extends the earlier canonical publish hardening work and turns the remaining review findings into a shared, production-grade contract.

## Confirmed Findings Driving This Design

1. The current `schedule publish` path does not persist which published document must be superseded at cutover time.
2. The current publish dialog copy implies replacement semantics for scheduled publish, but the backend does not implement that promise.
3. The canonical `/documents/:id` page uses state-aware labels for active sibling revisions, but not state-aware destinations.
4. A regulated workflow cannot safely infer replacement intent later from mutable runtime state.

## Product Rules

### Lifecycle

- `approved` remains a pre-publication state.
- `published` remains an explicit later transition.
- `scheduled` means an approved revision has an explicit future publication intent.
- Future auto-publish policy remains a separate domain seam and is out of scope here.

### Publish Family

- If no currently published lineage head exists, `Publish now` and `Schedule publish` may operate without supersede context.
- If a currently published lineage head exists, `Publish now` must always supersede it explicitly.
- If a currently published lineage head exists, `Schedule publish` must always register a scheduled supersede explicitly.
- The UI must never offer a choice between "replace" and "do not replace" when a published head already exists.

### Scheduled Supersede Invariant

- A scheduled replacement must persist the exact published document it intends to supersede.
- Cutover must only execute if that exact published head is still current at execution time.
- If the published head changed after scheduling, the job must fail explicitly instead of guessing a new target.

### Active Sibling Revision

- The canonical published page must treat any active sibling revision as blocking new revision creation.
- Active sibling states are:
  - `draft`
  - `under_review`
  - `approved`
  - `scheduled`
  - `rejected`
- The canonical page must map both CTA label and CTA destination from sibling state.

## Recommended Architecture

### Contract Boundary

This is a cross-boundary change and must be implemented as:

- `metaldocs-backend-api`
- `metaldocs-database`
- `metaldocs-tanstack-query`

Reason:

- the public HTTP contract changes
- governed persistence changes
- generated frontend types and query/mutation wrappers change

### Persistence Model

Persist scheduled supersede intent on `documents`, not only in approval runtime state.

Recommended governed fields:

- existing scheduled publication timestamp remains on `documents`
- add explicit scheduled supersede reference on `documents`
  - field name: `superseded_document_id`

Why `documents` owns this:

- governed lifecycle truth already lives on `documents`
- lineage history is defined by `documents` under `controlled_document_id`
- the future published result is a governed document transition, not merely approval session metadata
- operational recovery and audit review should be possible from the document aggregate itself

Rejected alternatives:

- storing supersede intent only on `approval_instances`
- resolving supersede target only inside the cutover job
- freezing lineage through operational lockouts instead of explicit contract state

## Backend / API Design

### Route Ownership

The affected public surface remains document-scoped:

- `POST /api/v1/documents/{id}/publish`
- `POST /api/v1/documents/{id}/schedule-publish`

The route truth table must be rebuilt before implementation because the approval module is still partially raw and partially documented.

### Schedule Publish Request

When scheduling a publish for a document that replaces a current published head, the request must carry:

- `effective_from`
- `superseded_document_id`
- `If-Match` precondition for the approved revision being scheduled

If no current published head exists, `superseded_document_id` may be absent.

### Schedule Publish Validation

At schedule time, the backend must validate:

- target document exists in the tenant
- target document is in `approved`
- `If-Match` matches the approved revision version
- if present, `superseded_document_id`:
  - exists in the tenant
  - belongs to the same `controlled_document_id`
  - is currently `published`
  - is not equal to the target document id

Validation failures should map to explicit business conflicts or validation errors, not generic internal failures.

### Schedule Publish Persistence

On success, the backend stores on `documents`:

- `status = scheduled`
- `effective_date`
- `superseded_document_id` when replacement is intended
- revision OCC bump as already required by the publish family

### Cutover Job Behavior

At effective time, the scheduler must:

1. load the scheduled document
2. read its explicit `superseded_document_id`
3. verify the referenced document is still the current published head of the same lineage
4. if true:
   - publish the scheduled document
   - supersede the referenced published document
   - preserve one canonical published head after cutover
5. if false:
   - do not publish
   - record explicit operational conflict outcome
   - preserve current state for human resolution

### Failure Semantics

If the published head changed before cutover, the scheduler must fail explicitly with a domain/business conflict outcome.

This is not a transient infrastructure retry problem. It is a truth mismatch between scheduled intent and current lineage head.

## Frontend Design

### Canonical Published Page

`DocumentPublishedPage` must derive both label and destination from active sibling state.

Recommended mapping:

| State | Label | Destination |
|---|---|---|
| `draft` | `Continuar rascunho` | `/documents/:id/edit` |
| `under_review` | `Acompanhar revisão` | `/documents/:id` |
| `approved` | `Publicar revisão aprovada` | `/documents/:id` |
| `scheduled` | `Ver publicação agendada` | `/documents/:id` |
| `rejected` | `Retomar revisão rejeitada` | `/documents/:id/edit` |

Rationale:

- editable workflow states route to the editor
- non-edit workflow states route to the canonical detail surface
- the canonical detail page remains the truthful operational surface for publish and publish-scheduled contexts

### Publish Dialog

If `publishedDocumentId` exists:

- the dialog offers only timing choice:
  - `Publicar agora`
  - `Agendar publicação`
- it does not offer semantic replacement choice
- schedule mode must send `superseded_document_id`

If `publishedDocumentId` does not exist:

- the dialog may publish or schedule without supersede context

### TanStack Query Impact

Generated API types and feature wrappers must stay contract-first.

Expected frontend updates:

- generated API types for schedule-publish request/response
- approval/document feature API wrapper updates
- mutation invalidation review for:
  - canonical document detail
  - controlled-document active-document lookup
  - approval instance or publish context queries

The implementation should prefer server response + targeted invalidation, not optimistic cache rewriting, because publish transitions are regulated workflow state.

## Error UX

If cutover fails because the current published head no longer matches the stored `superseded_document_id`, the user-facing outcome should be a clear business conflict message in the error-UX style, for example:

- "A publicação agendada não pôde ser concluída porque a versão publicada mudou antes da data agendada."

Requirements:

- no generic "internal error" for this case
- no silent repointing to a new supersede target
- enough wording for an operator to understand that the schedule intent must be reviewed

## Testing Strategy

### Backend

- request validation tests for `superseded_document_id`
- schedule publish tests proving supersede intent is persisted
- cutover scheduler tests proving:
  - happy path publishes scheduled doc and supersedes recorded head
  - mismatch path fails explicitly and does not publish
- OCC/precondition tests for `If-Match`

### Frontend

- dialog tests proving schedule requests include `superseded_document_id`
- canonical page tests proving per-state destination, not only label
- error mapping tests for scheduled cutover conflict where applicable

### Runtime Verification

Required runtime truth verification before completion:

1. schedule approved revision while a published head exists
2. confirm schedule payload/DB state persists explicit supersede target
3. confirm canonical page shows truthful state-aware CTA
4. execute or simulate due cutover
5. verify one published head remains and intended prior head becomes superseded
6. verify mismatch scenario fails explicitly if lineage head changes before cutover

## Wiki / Memory Sync Requirements

If implemented, sync:

- `wiki/modules/approval.md`
- `wiki/modules/documents.md`
- `wiki/modules/registry.md`
- relevant `wiki/database/*` entries for the new governed field

Required updates:

- scheduled supersede contract truth
- governed persistence ownership on `documents`
- canonical CTA label + destination truth
- cutover conflict semantics

## Out of Scope

- system-wide auto-publish settings
- redesigning approval route ownership beyond this targeted slice
- legacy screen fallback
- unrelated editor UX refinements

## Acceptance Criteria

1. A scheduled publish that replaces an already published head persists explicit supersede intent on `documents`.
2. The schedule-publish API contract explicitly carries `superseded_document_id` when replacement applies.
3. Cutover publishes only if the stored supersede target is still the current published head.
4. If the published head changed before cutover, the job fails explicitly and does not publish silently.
5. The canonical published page uses state-aware CTA labels and destinations.
6. Generated frontend API types, wrappers, and query invalidation stay aligned with the new contract.
7. Wiki and database memory are synced to runtime and contract truth.

## Classification

- Scheduled supersede contract gap: `shared contract prerequisite`
- Canonical state-aware destination gap: `screen-local implementation`
- Error mapping for head-changed cutover conflict: `module-local implementation`
- Wiki/database sync after implementation: `wiki-memory drift` if omitted
