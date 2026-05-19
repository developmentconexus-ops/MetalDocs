# Editor Sidebar Revision Title and Visual Density Design

> **Date:** 2026-05-19
> **Status:** Draft for user review
> **Scope:** Focused amendment to the editor sidebar governed-revision work after runtime E2E review.
> **Depends on:** `docs/superpowers/specs/2026-05-18-editor-sidebar-governed-revision-design.md`
> **Out of scope:** Approval route/profile creator repair, published-page redesign, effective-date/vigency semantics, and non-editor registry screens.

## 1. Why this amendment exists

The first sidebar implementation replaced mock data with real governed revision, registry visibility, and approval-instance data. Browser review showed the domain wiring moving in the right direction, but also surfaced two product/design corrections before the feature can be considered professional SaaS quality:

- `REV00` should not ask every user to type a revision title. The first governed revision is always the document creation event.
- The sidebar's current visual density is too sparse and leaves an unattractive blank panel area. It needs to feel like a compact operational side panel, not a large white placeholder.

This amendment updates the previously approved design without changing the core truth hierarchy:

- governed revision history still comes from `documents` lineage by `controlled_document_id`
- `document_revisions` remains technical/autosave-only
- registry remains the source for controlled-document visibility
- approval remains the source for approval-chain state

## 2. Product decisions

### 2.1 `REV00` title is automatic

For the first governed revision in a controlled-document lineage:

- `documents.revision_number = 0`
- display code is `REV00`
- revision title defaults to `Criacao do documento`
- no user input is required to submit/finalize `REV00`

Reason:

- every document has the same business reason for its first governed revision: creation
- asking the user to type a title for `REV00` creates noise and inconsistent audit text
- the governed history becomes more reliable if the system owns this first title

### 2.2 Revision title remains required after `REV00`

For governed revisions after creation:

- `revision_number > 0` must still require `revisionTitle` before formal submission/finalize
- the title is still stored on `documents.revision_title`
- the title still freezes after submission/finalize

Reason:

- later revisions need a human-readable reason
- audit quality matters most after a controlled document already exists

### 2.3 Sidebar revision rows do not show document status inline

The revision row must not display `Draft`, `under_review`, or equivalent status text next to the revision code.

Target row model:

- left/content: `REV00`
- adjacent/main text: revision title, for example `Criacao do documento`
- right/aligned date: short date such as `19/05/2026`

Rejected row model:

- `REV00 · Draft`
- `REV00 · Em revisao`

Reason:

- the document status is already represented elsewhere in the editor chrome/status pill
- mixing revision identity and workflow state makes the history noisy
- the revision history should read as a business history, not as a state-machine dump

## 3. Sidebar visual design

### 3.1 Target feel

The sidebar should feel like a compact enterprise metadata rail:

- narrow enough to read as a supporting panel
- dense enough that sections feel intentional
- visually bounded so it does not become a large blank white column
- organized with section rules, small uppercase labels, and card-like rows for dynamic lists

The user-provided screenshot is a visual reference for density and panel rhythm, not a source of truth for exact layout, colors, labels, or content.

### 3.2 Layout rules

The sidebar should:

- keep a fixed panel width consistent with the editor layout
- use compact vertical padding per section
- use subtle horizontal separators inside the panel
- avoid a full-height white void below short content
- let the panel content define its visual card height while preserving scroll if content grows

Implementation can use a panel background plus an inner content stack/card so the empty area does not look like unfinished UI.

### 3.3 Revision history compaction

For small histories:

- show all revision rows when there are at most 3 items

For larger histories:

- show the current revision plus the two most recent relevant rows by default
- provide an explicit expand/collapse affordance such as `Ver todas as revisoes`
- when expanded, show the full governed lineage in the sidebar scroll area

The collapsed state must never hide the current revision.

## 4. Behavior by status

### 4.1 Draft

Show:

- real metadata: code, profile, area, visibility
- governed revision row: `REV00 Criacao do documento <date>` for first revision
- no approvers

Do not require a revision title modal for `REV00`.

### 4.2 Under review

Show:

- real metadata: code, profile, area, visibility
- governed revision history rows without inline workflow status
- full approval chain from the real approval instance

The approval section remains gated to `under_review` for this slice.

### 4.3 Later revisions

When `revision_number > 0` and the document is submitted/finalized:

- require user-authored `revisionTitle`
- render the row as `REV01 <user title> <date>`
- do not append the workflow status inside the revision row

## 5. Data and contract impact

### 5.1 Backend/application

Finalize/submission behavior must distinguish first governed revision from later governed revisions:

- if `documents.revision_number = 0` and no `revisionTitle` is provided, use `Criacao do documento`
- if `documents.revision_number > 0`, reject missing/blank `revisionTitle`
- preserve existing freeze semantics after submission/finalize

This is a module-local implementation change in `documents` unless OpenAPI request semantics need a description update.

### 5.2 API contract

The existing finalize request can keep `revisionTitle` as optional at transport level if application validation is conditional.

Contract documentation should state:

- omitted/blank `revisionTitle` is accepted only for `REV00`
- later governed revisions require it

### 5.3 Frontend

The editor page should:

- skip the revision-title dialog when current governed revision is `REV00`
- call finalize with either omitted `revisionTitle` or the canonical automatic value, depending on backend contract choice
- keep the dialog for `revision_number > 0`
- render revision rows as code/title/date, not code/status/date

Preferred boundary:

- backend owns the canonical default title
- frontend may mirror the rule for UX flow, but backend remains authoritative

## 6. Error handling

If finalize fails because `revisionTitle` is missing for a later revision:

- show the existing user-facing modal/error path
- keep the user in the editor
- do not silently submit a later revision with a generated title

If the revision history query returns an empty list unexpectedly:

- show no fake history
- keep metadata visible
- let normal query/error handling surface the problem

## 7. Verification expectations

Focused checks:

- unit/page test: `REV00` submit does not open the revision-title-required error path
- unit/page test: `REV01` or later still requires title
- sidebar component test: revision row renders `REV00`, title, and date without status text
- sidebar component test: draft does not render approvers
- sidebar component test: history collapses only when item count exceeds the threshold

Runtime/browser checks:

- draft `REV00` displays `Criacao do documento` in history
- draft `REV00` submit path does not require manual title
- under_review displays revision rows without workflow status text
- approval chain still displays only for `under_review`
- screenshots cover draft, title behavior, and under_review sidebar

## 8. Legitimate defers

- Full approval route/profile creator repair remains deferred until after editor completion.
- Effective published/vigency dates remain deferred until the runtime contract exposes truthful fields.
- Rich long-history filtering/search remains deferred; this slice only needs collapse/expand.

