# Document Revision Artifact Metadata Design

## Status

Proposed for review.

## Problem

The editor sidebar should show real document artifact metadata, including page count and DOCX size. Keeping this only in React state would be useful visually but weak as SaaS/product truth: future recovery, audit investigation, support, and database-level inspection would not know what artifact metadata belonged to a saved head revision.

EigenPal can calculate the rendered page count in the browser, but EigenPal is client-side only. The server owns durable metadata and must persist artifact facts in the same transaction that commits the autosaved DOCX revision.

## Goals

- Persist page count and DOCX size as database metadata for the saved current document artifact.
- Keep governed revision history sourced from `public.documents` lineage by `controlled_document_id`.
- Keep `public.document_revisions` classified as technical/autosave revision storage, not governed business history.
- Avoid a second metadata-only HTTP round trip after autosave.
- Keep page count provenance explicit because the first source is EigenPal client pagination, not a server renderer.
- Expose the metadata through the documents API so the editor sidebar can render real data without mocks.

## Non-Goals

- Do not redesign governed revision history.
- Do not move artifact metadata into `metaldocs.document_versions`; current runtime editor flow uses `public.documents` and `public.document_revisions`.
- Do not create a separate endpoint just to patch page count after autosave.
- Do not make `document_revisions` appear as business revision history in the UI.
- Do not require server-side DOCX rendering in this slice.

## Classification

- Database migration: `post-baseline forward migration`
- API/OpenAPI change: `shared contract prerequisite`
- Backend autosave persistence: `module-local implementation`
- Editor wrapper/page count collection: `screen-local implementation`
- Server-side page count recomputation: `defer`

## Current Truth

Runtime truth:

- Autosave commit currently verifies the uploaded object hash server-side, inserts a new `public.document_revisions` row, updates `public.documents.current_revision_id`, updates the editor session acknowledgment, and marks the pending upload consumed in one transaction.
- `public.document_revisions` stores technical artifact rows with `storage_key`, `content_hash`, and `form_data_snapshot`.
- `public.documents.current_revision_id` points to the current saved artifact row.
- EigenPal public `DocxEditorRef` exposes `getTotalPages()`.
- The MetalDocs wrapper `packages/editor-ui` is the anti-corruption layer; app screens must not import EigenPal directly.

Contract/wiki truth:

- `document_revisions` is not governed history.
- Governed revision display comes from `documents` lineage by `controlled_document_id`.
- The current API contract does not expose page count or file size on document detail or autosave commit.
- Post-baseline DB changes belong in `db/migrations` and must write `public.schema_migrations`.

## Design

Add artifact metadata to `public.document_revisions`:

- `file_size_bytes bigint`
- `page_count integer`
- `page_count_source text`

The metadata belongs to the saved DOCX artifact represented by that technical revision row. The current document detail can expose the metadata by joining `public.documents.current_revision_id` to `public.document_revisions.id`.

`file_size_bytes` is server-authoritative. The backend should derive it from object storage metadata or the existing server-side object inspection path during autosave commit.

`page_count` is initially client-observed from EigenPal pagination. The frontend should get it through `MetalDocsEditorRef.getPageCount()` in `packages/editor-ui`, and the documents page should include it in the autosave commit request. The backend validates it as a positive integer and stores it with `page_count_source = 'eigenpal_client'`.

Future server-side page count can be added later as a renderer/freeze enhancement by introducing another source, for example `server_renderer`, and updating rows when the server has stronger layout truth.

## Autosave Efficiency

The efficient path is to persist metadata in the existing autosave commit transaction.

Every non-idempotent autosave commit already creates a new `document_revisions` row. Adding metadata columns to the existing `INSERT` has low marginal cost and keeps artifact bytes, hash, size, page count, and current-head pointer aligned atomically.

The implementation should not call a separate metadata endpoint after autosave. It should also not issue a separate conditional `UPDATE` merely to avoid writing unchanged values, because a new content hash already means a new artifact row. Idempotent replay should return the already committed revision metadata without inserting or rewriting.

## API Contract

Extend `POST /api/v1/documents/{id}/autosave/commit` request with:

- `page_count`: optional integer, minimum `1`

Extend its success response with:

- `file_size_bytes`: integer, nullable if unavailable during transitional rows
- `page_count`: integer, nullable
- `page_count_source`: string, nullable

Extend document detail response with current artifact metadata:

- `currentRevisionFileSizeBytes`
- `currentRevisionPageCount`
- `currentRevisionPageCountSource`

Names should follow the existing generated API naming conventions after OpenAPI codegen. Frontend code must consume generated types rather than hand-written response mirrors.

## Database Migration

Create the next post-baseline migration after the current tail:

- expected next version: `0206`
- path: `db/migrations/0206_document_revision_artifact_metadata.sql`

Migration requirements:

- Add nullable columns to `public.document_revisions`.
- Add check constraints:
  - `file_size_bytes IS NULL OR file_size_bytes >= 0`
  - `page_count IS NULL OR page_count > 0`
  - `page_count_source IS NULL OR page_count_source IN ('eigenpal_client', 'server_renderer')`
- Insert one `public.schema_migrations` row.
- Do not backfill historical rows unless runtime has a reliable object-store/stat path for every stored revision.

Dictionary impact:

- Update `wiki/database/tables/document_revisions.md`.
- Update related documents module docs if API/sidebar behavior changes.

## Frontend Flow

`packages/editor-ui` should expose page count through the wrapper ref:

- `MetalDocsEditorRef.getPageCount(): number | null`
- implementation delegates to EigenPal `inner.current.getTotalPages()`
- return `null` when editor/layout is not ready or the value is invalid

`DocumentEditorPage` should collect page count at autosave commit time, not through direct EigenPal imports.

The editor sidebar should render pages/size from API-backed current artifact metadata. It may use local runtime metadata immediately after autosave success if the autosave response updates the local detail/cache consistently.

## Error Handling

If the client omits `page_count`, the backend should still commit the autosave and store `NULL` page metadata. Missing page count should not block authoring because server-authoritative bytes and hash are more important than client pagination.

If the client sends invalid `page_count`, the backend should reject the commit with a validation error before inserting the revision. This protects DB integrity and avoids storing impossible metadata.

If object storage size lookup fails after hash verification succeeds, the safer first implementation is to fail the commit rather than store a revision with unknown server artifact facts, unless investigation proves the hash stream already provides byte count reliably.

## Verification

Minimum implementation gates:

- DB migration applies through the post-baseline runner and records `0206`.
- Dictionary coverage passes.
- OpenAPI lints and backend codegen succeeds.
- Documents module tests cover autosave commit with metadata.
- Frontend generated API types refresh successfully.
- `packages/editor-ui` tests cover `getPageCount()` wrapper behavior.
- Editor page tests prove autosave sends page count through the wrapper, not EigenPal direct imports.
- Browser E2E verifies sidebar pages/size match API/runtime truth on the editor.

## Deferred Work

- Server-renderer page count computation for stronger server-side layout truth.
- Historical backfill of old `document_revisions` rows.
- Persisting paper size/orientation if product later needs it as governed metadata.

