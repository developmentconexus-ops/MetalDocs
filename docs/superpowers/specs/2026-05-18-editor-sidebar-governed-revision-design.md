# Editor Sidebar Governed Revision Design

> **Date:** 2026-05-18
> **Status:** Approved in-session, written for user review
> **Scope:** Documents editor right sidebar as a real product capability across `documents`, `registry`, `approval`, OpenAPI/generated types, TanStack Query, and database-backed revision metadata.
> **Out of scope:** Right-sidebar visual redesign, published-screen cleanup, template-comments parity, generic comments platform.

## 1. Why this exists

The current editor sidebar is mock-driven. That is not only a screen problem; it exposes a deeper system gap:

- the product does not yet expose a first-class governed revision history surface for a controlled document
- the current sidebar conflates draft editing, effective published state, and approval flow semantics
- approval tracking exists at runtime but is not yet professionalized as a shared contract

We need to turn the sidebar into a real SaaS capability backed by the correct domain layers:

- `documents` owns governed document revisions and editor state
- `document_revisions` remains technical autosave/content lineage
- `registry` owns controlled-document identity and visibility
- `approval` owns approval-chain execution and signoff tracking

The resulting design must improve the system as a whole, not just patch a screen.

## 2. Core design decisions

### 2.1 Governed revision vs technical revision

MetalDocs already has two revision concepts and they must stay separate:

- **Governed revision**: the business/QMS revision in the controlled-document lineage. This is represented by `documents` rows linked by `controlled_document_id` and numbered by `documents.revision_number`.
- **Technical revision**: editor/autosave content lineage inside a governed revision. This is represented by `document_revisions`.

The sidebar must show only governed revisions. It must never list autosave rows from `document_revisions`.

### 2.2 Revision numbering model

The governed revision sequence is the user-facing revision history.

Visual format:

- `revision_number = 0` -> `REV00`
- `revision_number = 1` -> `REV01`
- `revision_number = 2` -> `REV02`

Interpretation:

- first governed revision of a document lineage is `REV00`
- when published, it remains `REV00 · Publicada`
- the next governed draft becomes `REV01 · Draft`
- then `REV01 · Publicada`, followed by `REV02 · Draft`, and so on

The zero-padded `REVxx` string is presentation logic, but it must not hide an offset. Backend/database truth is zero-based `documents.revision_number`; frontend formats `REV00`, `REV01`, etc. directly from that persisted numeric value.

### 2.3 Revision title / reason

Each governed revision needs a formal human-readable reason/title describing why that revision exists.

Canonical field:

- `documents.revision_title`

Semantics:

- short, user-authored summary of the revision reason
- belongs to the governed revision, not to autosave history
- required at formal submission time
- immutable after submission

This keeps autosave friction low while making the audit trail strong.

### 2.4 Capture timing

`revision_title` is required on `POST /api/v1/documents/{id}/finalize`.

Rationale:

- autosave remains technical and silent
- the reason becomes mandatory exactly when the revision enters the formal review workflow
- the field can be frozen transactionally with submit/approval-instance creation

Alternative designs considered and rejected:

- requiring the title at draft creation: too much friction and not aligned with autosave editing
- allowing title edits after submission: weakens governance and audit clarity
- storing title on technical revision rows: wrong domain boundary

### 2.5 Approval tracking in the sidebar

The sidebar should show the full approval chain, not only the next pending stage.

Goal:

- anyone looking at the revision can understand who already approved, who is currently active, and who still remains in the flow

However, this section must only appear when there is a real approval instance relevant to the current governed revision. The UI must never invent approvers for a draft that has not entered review.

## 3. Sidebar behavior by status

### 3.1 Draft

Show:

- `Código`
- `Perfil`
- `Área`
- `Visibilidade`
- governed revision history, with current item shown as `REVxx · Draft`

Hide:

- `Próximos aprovadores` / approval chain
- `Vigência atual`
- `Próx. revisão`

Reason:

- a draft is not yet an active approval flow
- a draft is not necessarily the currently effective/published revision

### 3.2 Under review

Show:

- metadata base (`Código`, `Perfil`, `Área`, `Visibilidade`)
- revision history with current item shown as `REVxx · Em revisão`
- full approval chain for the active approval instance

Approval chain rendering semantics:

- completed earlier stages appear as completed
- active stage is highlighted
- later pending stages appear as waiting

### 3.3 Approved / scheduled

Show:

- metadata base
- governed revision history
- full recorded approval chain for the revision

Do not introduce `Vigência atual` until the product/runtime distinction between approved and effective/published is explicit in the data contract.

### 3.4 Published

Show:

- metadata base
- governed revision history with current item shown as `REVxx · Publicada`
- full approval chain for that revision
- `Vigência atual` once the runtime/source field is truthful

`Próx. revisão` should only appear once we have a real effective-date anchor plus the review interval rule.

### 3.5 Rejected

Show:

- metadata base
- governed revision history keeping the current revision visible in the lineage

For first delivery, do not render approval tracking when no active approval instance exists. Historical rejected-flow visualization can be added later once the read model is explicit.

## 4. Data model changes

### 4.1 `documents`

Add:

- `revision_title text null`

Rules:

- nullable only to support legacy rows and drafts not yet submitted
- required by application flow at finalize time for new governed submissions
- immutable after successful submission/finalize

Why `documents` and not `document_revisions`:

- `documents` is the governed revision record
- `document_revisions` is autosave/content lineage inside that governed revision
- storing the title on `document_revisions` would conflate product audit metadata with editor technical state

### 4.2 No sidebar dependence on `document_revisions`

No new sidebar semantics will be sourced from `document_revisions`. That table remains responsible for:

- autosave revisions
- content hashes
- storage keys
- content restore lineage
- checkpoint restoration support

### 4.3 No new governed-revision table

We will not introduce a new `document_revision_history` table.

Reason:

- `documents` already expresses the governed revision lineage via `controlled_document_id` + `revision_number`
- creating a third revision abstraction would add cost without improving the domain model enough to justify it

## 5. Backend/API surfaces

### 5.1 Extend `GET /api/v1/documents/{id}`

Keep it as the main editor anchor and add:

- `RevisionTitle`

Retain existing useful fields:

- `ControlledDocumentID`
- `ProfileCodeSnapshot`
- `ProcessAreaCodeSnapshot`
- `RevisionVersion`
- `Status`
- `Code`

Do **not** add formatted `REV00` strings here. `revisionNumber` remains the contract truth for governed revision history; frontend formats the badge without applying an offset.

### 5.2 Add governed revision-history endpoint

Add a new read surface owned by `documents`:

- `GET /api/v1/documents/{id}/revision-history`

Purpose:

- return the governed lineage for the current document's `controlled_document_id`
- provide everything the sidebar needs without leaking autosave internals

Response shape:

- array ordered newest first
- each item contains:
  - `documentId`
  - `revisionVersion`
  - `revisionTitle`
  - `status`
  - `createdAt`
  - `submittedAt` optional
  - `approvedAt` optional
  - `publishedAt` optional
  - `isCurrent`

Implementation note:

- this is a new domain read model over `documents`, not over `document_revisions`

### 5.3 Professionalize approval-instance contract

Do not create a second approval endpoint for the sidebar if the current runtime surface is sufficient.

Use:

- `GET /api/v1/documents/{id}/approval-instance`

But first repair it as a shared contract:

- add/complete the OpenAPI operation and response schema
- regenerate backend and frontend contract types
- replace the handwritten frontend wrapper status guesses with generated types
- align runtime values such as `in_progress`, `active`, `passed`, `failed`, `cancelled`

Sidebar use of approval tracking must wait for this contract repair.

### 5.4 Extend finalize request

`POST /api/v1/documents/{id}/finalize` must accept `revisionTitle`.

Behavior:

- request validates that `revisionTitle` is present and non-empty
- the submit/finalize transaction writes `documents.revision_title`
- the value is thereafter treated as frozen metadata for that governed revision

This keeps the revision reason aligned with the formal review handoff.

## 6. Frontend / TanStack Query design

### 6.1 Editor detail query

Continue using the document detail query as the editor anchor:

- `QK.documents.detail(id)`
- `useDocumentDetailQuery(id)`

Sidebar metadata should consume the same detail payload for:

- code
- revision version
- revision title
- status
- controlled document id
- profile/area snapshots

### 6.2 Taxonomy lookups

Do not require backend duplication for display names we can already resolve correctly.

Use:

- taxonomy profiles query to resolve `ProfileCodeSnapshot -> profile.name`
- taxonomy areas query to resolve `ProcessAreaCodeSnapshot -> area.name`

This is a screen-local integration improvement, not a new backend capability.

### 6.3 Controlled-document visibility

Use the existing controlled-document contract as the source of truth for visibility.

Needed frontend cleanup:

- registry wrapper/types must stop omitting `visibility`
- use generated contract types for the controlled-document detail read

Suggested query key:

- `QK.controlledDocuments.detail(id)`

### 6.4 Revision-history query

Add a dedicated documents query hook:

- `QK.documents.revisionHistory(id)`
- `useDocumentRevisionHistoryQuery(id)`

The query should be scoped to the current editor document id and return the governed lineage read model.

### 6.5 Approval-instance query

Reuse the existing query slot once the contract is fixed:

- `QK.approval.instance(documentId)`
- `useApprovalInstanceQuery(documentId)`

UI rule:

- show approval section only when current document status and returned instance make the flow real
- never synthesize approvers for `draft`

## 7. Open questions resolved in this design

These decisions were made in-session and are now explicit design truth:

- revision reason/title is required at finalize time, not at draft creation
- revision reason/title freezes at submission
- sidebar shows the current governed revision even while still in draft
- visual revision numbering uses `REV00`, `REV01`, `REV02`, ...
- sidebar history is governed lineage only, never autosave rows
- approval section should show the full chain, not only the next pending stage

## 8. Remaining prerequisites and deferred items

### 8.1 Shared contract prerequisite

Before implementation can claim approval tracking is complete in the sidebar:

- the `approval-instance` route must become a trustworthy OpenAPI/generated/frontend contract

Classification:

- `shared contract prerequisite`

### 8.2 Deferred for later phase

These are intentionally not part of the first implementation slice:

- `Vigência atual` when the current editor document is draft but not effective
- `Próx. revisão` until an explicit effective-date anchor exists
- historical visualization of rejected approval chains when no active instance exists
- published-screen cleanup and parity work outside the editor

## 9. Testing and verification expectations

### 9.1 Database / backend

- migration adds `documents.revision_title`
- integration test: finalize requires `revisionTitle`
- integration test: successful finalize persists `revision_title`
- integration test: revision-history query returns governed lineage ordered correctly
- integration test: revision-history excludes autosave `document_revisions` noise
- approval-instance contract verification after OpenAPI/codegen update

### 9.2 Frontend

- detail query test covering `RevisionTitle`
- revision-history query test with generated types
- editor page/sidebar test for:
  - draft hides approval chain
  - under_review shows approval chain
  - profile/area names resolve from taxonomy queries
  - visibility resolves from controlled-document query
  - revision list shows `REV00`, `REV01`, ... and `revisionTitle`

### 9.3 Contract gates

- `npx @redocly/cli lint api/openapi/v1/openapi.yaml`
- `go generate ./internal/modules/documents/api/...`
- regenerate affected approval/documents API packages as needed
- `pnpm gen:api`
- `pnpm.cmd tsc --noEmit -p tsconfig.build.json`

## 10. Recommended implementation strategy

Implement in this order:

1. Repair shared contract truth for `approval-instance`
2. Add `revision_title` storage + finalize write path
3. Add governed `revision-history` endpoint
4. Normalize registry frontend wrapper for `visibility`
5. Wire sidebar to real data through TanStack Query
6. Keep deferred items out until their semantics are truthful

This order keeps the implementation aligned with the MetalDocs truth hierarchy:

- runtime truth
- contract truth
- wiki truth
- execution truth

## 11. Why this is the recommended approach

This design intentionally improves the system, not only the screen:

- it preserves the architectural separation between governed revisions and editor autosave revisions
- it adds revision-title semantics at the correct domain level
- it uses existing taxonomy and registry capabilities where they are already truthful
- it avoids inventing UI semantics for draft approval state or effective dates
- it turns approval tracking into a shared contract rather than a screen-local guess

That gives us a right sidebar that is useful, auditable, and consistent with a professional regulated SaaS.
