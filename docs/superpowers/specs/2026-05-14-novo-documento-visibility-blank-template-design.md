# Novo Documento Visibility and Blank Template Design

> Date: 2026-05-14
> Status: approved for implementation planning
> Scope: `novo-documento` visibility model and blank-document creation path
> Out of scope: external sharing links, edit permission changes, broad template-management redesign

## Context

The current `novo-documento` wizard captures visibility in frontend state only. It does not submit visibility to `POST /api/v1/controlled-documents`, and the backend has no persisted visibility model on controlled documents. The current blank-template card is disabled because the documents creation pipeline still expects a real template version for editor bootstrap, storage key, snapshots, and hashing.

This design turns both items into real product behavior without fake data or frontend-only filtering.

## Goals

- Persist visibility at controlled-document creation time.
- Enforce visibility in backend read/list/detail behavior from day one.
- Keep visibility as view-only access; editing remains governed by normal role and capability permissions.
- Replace "Apenas minha area" with an area picker that defaults to the document's own process area and can include additional areas.
- Keep company-wide visibility exclusive.
- Keep external sharing deferred.
- Enable `Em branco` by using a real system-owned immutable blank template version.

## Non-Goals

- Do not implement external sharing, password-protected links, expiry, or watermark rules.
- Do not grant edit access through visibility selections.
- Do not create a native `templateVersionId: null` document path.
- Do not expose the system blank template as a normal editable template.

## Product Behavior

### Visibility Choices

`Toda empresa` is exclusive. If selected, every internal authenticated user in the tenant can view the document. Selected areas and selected people are cleared or ignored.

`Areas selecionadas` is restricted visibility. It starts with the document's own process area selected by default. The author can add other areas. The document's own area should remain selected in the first implementation.

`Pessoas especificas` is additive with selected areas. Selected users receive view access only.

`Compartilhamento externo` remains deferred. It can stay visible as a disabled future option, but it is not part of this implementation.

### Blank Document

`Em branco` becomes selectable in Step 3. Selecting it submits a real `templateVersionId` belonging to a system-owned immutable blank template version.

The created document behaves like any other controlled document: it gets a code, first revision, editor route, content hash, audit trail, and lifecycle.

## API Contract

Atomic create should accept visibility as a structured object:

```ts
type ControlledDocumentVisibilityInput = {
  scope: "company" | "restricted";
  areaCodes: string[];
  userIds: string[];
};
```

Rules:

- `scope: "company"` requires empty `areaCodes` and empty `userIds`.
- `scope: "restricted"` requires at least one `areaCodes` or `userIds` entry.
- The wizard defaults restricted area visibility to `[processAreaCode]`.
- Visibility grants `document.view` only.

Controlled-document detail responses should expose the persisted visibility summary needed by detail and confirmation screens. List responses should include a compact visibility summary only when the existing list UI needs to render it; list filtering must always be server-side.

## Backend Storage

Registry owns visibility because registry owns `controlled_documents`, the atomic create endpoint, and controlled-document identity.

Recommended schema:

- `controlled_documents.visibility_scope`, with values `company` and `restricted`.
- `controlled_document_visibility_areas`, with `tenant_id`, `controlled_document_id`, and `area_code`.
- `controlled_document_visibility_users`, with `tenant_id`, `controlled_document_id`, and `user_id`.

Indexes should support common reads:

- `controlled_documents(tenant_id, profile_code, process_area_code, status)`
- `controlled_document_visibility_areas(tenant_id, area_code, controlled_document_id)`
- `controlled_document_visibility_users(tenant_id, user_id, controlled_document_id)`

The create transaction should insert the controlled document, create the first document revision, and persist visibility rows atomically.

## Visibility Enforcement

Visibility is a backend read filter, not a frontend filter.

List/detail/open queries should return a document only when one of these is true:

- The document has `visibility_scope = 'company'`.
- The document is restricted to at least one active area membership held by the current user.
- The document is restricted to the current user's `user_id`.
- The current user has an existing admin/system bypass according to IAM conventions.

Pagination must run after visibility filtering, so page 1 means the first page of documents the current user can see.

The frontend should never load all documents and remove hidden ones in React.

## System Blank Template

The blank document path uses a system-owned immutable blank template/version.

Rules:

- Created by migration or startup bootstrap, not by user action.
- Hidden from normal template list/manage screens.
- Not editable, archivable, submit-able, approve-able, or publish-able through normal template routes.
- Has one stable published version for the atomic create flow.
- Has a valid empty DOCX/editor object.
- Has an empty placeholder schema.
- Carries metadata such as `systemOwned: true` or an equivalent persisted marker.

The wizard should obtain or receive the blank template version ID through a real backend surface, not hardcoded fake data.

## Frontend Behavior

Step 2:

- Replace "Apenas minha area" with "Areas selecionadas".
- Keep `Toda empresa` exclusive.
- Show an area multi-select for restricted area visibility.
- Default the selected visibility areas to the document's own process area.
- Allow additional areas to be selected.
- Add a specific-people picker only when a real user lookup surface exists for authors.
- Keep external disabled/deferred.

Step 3:

- Make `Em branco` selectable only after the real system blank template version is available.
- Submit its real `templateVersionId`.
- Keep normal template cards unchanged.

Step 4:

- Summarize visibility honestly: company-wide, selected areas, selected users, or blank template when applicable.

## Implementation Split

### PR 1: Visibility Contract and Enforcement

Scope:

- Registry schema, domain, service, repository, API contract, generated backend API, generated frontend types.
- Backend read/list/detail enforcement.
- Wizard Step 2 state and payload.
- Frontend tests and API tests.
- Wiki/module/backlog sync.

Verification:

- Mandatory runtime and contract gates.
- API tests for create/list/detail filtering.
- Frontend tests for wizard state and payload.
- Runtime smoke proving allowed users can see restricted documents and unrelated users cannot.

### PR 2: System Blank Template

Scope:

- Seed or bootstrap immutable system blank template/version.
- Protect it from normal template mutation routes.
- Hide it from normal template lists unless explicitly requested by system code.
- Expose the blank option to the wizard through a real backend surface.
- Make Step 3 submit the real blank template version ID.
- Wiki/module/backlog sync.

Verification:

- Backend tests proving the blank template exists and cannot be mutated.
- Atomic create smoke using the blank template.
- Editor opens the created blank document.
- Frontend test proving the blank card is selectable and submits the real version ID.

## Risks and Constraints

- This is shared contract work, not a screen-local change.
- IAM user lookup for author-facing people selection may need a new narrowed endpoint instead of reusing admin `GET /api/v1/iam/users`.
- Template system ownership must be enforced server-side; hiding in the UI is not enough.
- Existing runtime create `500 INTERNAL_ERROR` from the prior smoke remains a prerequisite to verify or repair before relying on end-to-end create tests.

## Planning Prerequisite

Before implementing the people picker, audit runtime truth for an author-safe user directory endpoint. If none exists, the implementation plan must include a narrowed user-search endpoint rather than reusing admin `GET /api/v1/iam/users`.
