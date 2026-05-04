# documents Module

> **Last verified:** 2026-05-04
> **Scope:** Document instances — create, edit, autosave, checkpoints, finalize, archive, export.
> **Out of scope:** Template authoring (see `modules/templates-v2.md`), approval routes (`modules/approval.md`), PDF fanout (`modules/render-fanout.md`).
> **Key files:**
> - `frontend/apps/web/src/features/documents/v2/DocumentEditorPage.tsx:20` — `DocumentEditorPage` component; `handleRename` at :113 captures previous name for rollback on server error (E9)
> - `frontend/apps/web/src/lib/api/errorMessages.ts:14` — `resolveErrorMessage` — maps API error codes to localised user strings; used by rename handler
> - `frontend/apps/web/src/lib/api/client.ts:5` — `ApiError` class — structured error with `.code` + `.status` fields
> - `frontend/apps/web/src/features/documents/v2/styles/DocumentEditorPage.module.css:1` — wine-brand chrome CSS
> - `frontend/apps/web/src/features/documents/v2/routes.tsx:1` — route parsing/rendering for `/documents-v2/*`
> - `frontend/apps/web/src/features/documents/v2/DocumentCreatePage.tsx:1` — step 1: pick controlled document
> - `frontend/apps/web/src/features/documents/DocumentsHubView.tsx:758` — detail panel with Edit/PDF/Duplicate actions
> - `internal/modules/documents/delivery/http/handler.go:73` — `NewHandlerWithSubmit` — wires db + submitSvc for atomic finalize
> - `internal/modules/documents/delivery/http/handler.go:259` — `finalizeDocument` — resolves approval route, calls SubmitRevisionForReview
> - `internal/modules/documents/application/service.go:218` — `CreateDocument` — calls `ResolveTemplate` pre-INSERT for atomic snapshot
> - `internal/modules/documents/application/service.go:598` — `Archive` — soft-archive via `MarkArchived`; no status change
> - `internal/modules/documents/application/snapshot_service.go:73` — `ResolveTemplate` — returns `TemplateSnapshot` + `[]Placeholder` pre-INSERT
> - `internal/modules/documents/application/snapshot_service.go:48` — `SnapshotFromTemplate` — deprecated; retained for backfill only
> - `internal/modules/documents/repository/repository.go:37` — `CreateDocument` — accepts `requiredPlaceholders`; seeds `document_placeholder_values` atomically
> - `internal/modules/documents/repository/repository.go:186` — `ListDocuments` — filters `archived_at IS NULL` by default
> - `internal/modules/documents/repository/repository.go:880` — `MarkArchived` / `Unarchive` — set/clear `documents.archived_at`
> - `internal/modules/documents/domain/model.go:25` — `Document` struct — `TemplateSnapshot` field; `FinalizedAt` removed
> - `internal/modules/documents/module.go:1` — DI wiring
> - `migrations/0171_drop_finalized_at.sql` — drops `documents.finalized_at`; adds `v_document_finalized` view

## Overview

A **document** is an instance filled from a template version, bound to a controlled document entry.
Documents move through states: `draft → under_review → approved → published`.
Only `draft` documents can be edited in the editor.

**Archive:** soft-hide via `documents.archived_at` timestamp (ADR 0008-soft-archive-via-timestamp). Status is never changed by archive. Default list queries filter `archived_at IS NULL`. `Service.Archive` / `Service.Unarchive` set/clear this column without touching the lifecycle status.

**Finalization timestamp:** `documents.finalized_at` was dropped (migration 0171). The finalization time now derives from `document_state_history` via the `v_document_finalized` view. `archived_at` is kept as a column because it is a hot-path list predicate.

**Backend module path:** `internal/modules/documents/` (renamed from `documents_v2` in migration batch 0167/0168)
**Table:** `public.documents`

## Frontend Routing

```
/documents-v2/new          → DocumentCreatePage (pick controlled document)
/documents-v2/<uuid>       → DocumentEditorPage
```

`viewFromPath` in `workspaceRoutes.ts` maps both to `activeView = "documents-v2"`.
`docsRouteFromPath` in `v2/routes.tsx` distinguishes `{ kind: 'create' }` vs `{ kind: 'editor', documentID }`.

## Create Flow

1. `DocumentCreatePage` lists active controlled documents (fetched from `GET /api/v2/registry/controlled-documents?status=active`).
2. User picks a controlled document + enters a name → `POST /api/v2/documents` → returns `{ document_id }`.
3. On success: navigate to `/documents-v2/<uuid>` → `DocumentEditorPage`.

Backend `Service.CreateDocument` calls `SnapshotService.ResolveTemplate` **before** the INSERT so that `TemplateSnapshot` + `requiredPlaceholders` are written atomically in the same transaction (fixes audit C2/C4). `SnapshotService.SnapshotFromTemplate` is deprecated — retained only for backfill scripts.

## Edit Flow (Draft Documents)

Entry points:
- Hub detail panel: "Editar" button (visible only for `status === "DRAFT"`) → `navigate('/documents-v2/<uuid>')`.
- "Ir para o editor" in duplicate confirmation modal.

`DocumentEditorPage` lifecycle:
1. Acquires writer session (`POST /api/v2/documents/:id/sessions`).
   - If another user holds the session → falls back to `readonly` mode.
2. Fetches signed URL for current revision DOCX → loads buffer into `MetalDocsEditor`.
3. On change: debounced autosave via `useDocumentAutosave` (`PUT /api/v2/documents/:id/revisions`).
4. "Finalizar" button: flushes autosave → `POST /api/v2/documents/:id/finalize` → atomically creates approval instance + transitions document to `under_review` → returns `{"instanceId":"<uuid>"}` (HTTP 201) → releases session → navigates away.

## Rename — optimistic update + rollback (E9 — fixed b14c7753)

`handleRename` at `DocumentEditorPage.tsx:113` implements the canonical optimistic-UI pattern for document renames:

1. Captures `prev = documentName` before the state update.
2. Calls `setDocumentName(name)` immediately (optimistic).
3. Fires `renameDocument(documentID, name)` (`PUT /api/v2/documents/:id/name`) asynchronously.
4. On server error: restores `setDocumentName(prev)` (rollback) and shows a toast via `resolveErrorMessage(code, 'Falha ao renomear documento.')`.

`resolveErrorMessage` (`lib/api/errorMessages.ts:14`) maps structured `ApiError.code` values to localised strings, falling back to the provided literal. `ApiError` (`lib/api/client.ts:5`) is the typed error primitive thrown by the API client layer.

---

## Session Model

- Writer sessions are exclusive (one writer at a time).
- `useDocumentSession` acquires on mount, heartbeats every 30s, releases on unmount.
- Stale/lost sessions surface as toasts; editor switches to `readonly` mode.

## Checkpoints

Checkpoints are manual snapshots. "Checkpoints" button opens `CheckpointsDialog`.
Restoring a checkpoint re-fetches the revision buffer and reloads the editor.

## Editor Chrome

`DocumentEditorPage` uses the same chrome pattern as `TemplateAuthorPage`:
- Left rail (48 px) with branded back button.
- `<main className={styles.canvas}>` → `<div className={styles.editorWrapper}>` contains `MetalDocsEditor`.
- `overlayTitle`: centered doc name + code + state badge (absolute, pointer-events: none).
- `overlayRight`: autosave status + Checkpoints + Export + Finalizar buttons (absolute, z-index 100).
- Eigenpal overrides in CSS tint the formatting bar with wine brand color.

## API Endpoints (Backend)

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v2/documents` | List documents |
| POST | `/api/v2/documents` | Create document from CD |
| GET | `/api/v2/documents/:id` | Get document metadata |
| PUT | `/api/v2/documents/:id/name` | Rename |
| POST | `/api/v2/documents/:id/sessions` | Acquire writer session |
| DELETE | `/api/v2/documents/:id/sessions/:sid` | Release session |
| GET | `/api/v2/documents/:id/revisions/:rid/signed-url` | Signed URL for DOCX |
| PUT | `/api/v2/documents/:id/revisions` | Save new revision (autosave) |
| POST | `/api/v2/documents/:id/finalize` | Finalize: atomically draft → under_review + create approval instance (HTTP 201, body `{"instanceId":"<uuid>"}`) |
| POST | `/api/v2/documents/:id/duplicate` | Duplicate document |
| GET | `/api/v2/documents/:id/view` | Signed PDF view URL |
| GET | `/api/v2/documents/:id/checkpoints` | List checkpoints |
| POST | `/api/v2/documents/:id/checkpoints` | Create checkpoint |
| POST | `/api/v2/documents/:id/checkpoints/:cid/restore` | Restore checkpoint |

## public.documents Schema

Migration 0167 fixed missing columns that were mistakenly added to the now-dropped `public.documents_v2` table (migrations 0126/0129). Migration 0171 dropped `finalized_at` and added `v_document_finalized` view. `public.documents` current schema:

| Column | Type | Notes |
|--------|------|-------|
| `controlled_document_id` | UUID | FK to `controlled_documents` |
| `profile_code_snapshot` | TEXT | Profile code at creation time |
| `process_area_code_snapshot` | TEXT | Area code at creation time |
| `code` | TEXT | Document code (e.g. `DC-001`) |
| `revision_number` | INT | Auto-incremented per CD (`MAX+1`) |
| `revision_version` | INT | Version within a revision |
| `effective_from` / `effective_to` | timestamptz | Effective date range |
| `locked_at` | timestamptz | Set on freeze |
| `content_hash_at_submit` | TEXT | Hash at submission time |
| `status` | TEXT | Extended CHECK; includes `draft`, `under_review`, `approved`, `published`, etc. |
| `archived_at` | timestamptz | NULL = visible; non-NULL = soft-archived. Set by `MarkArchived`, cleared by `Unarchive`. |

`finalized_at` column was **removed** by migration 0171. Use `v_document_finalized` view (derives from `document_state_history`) for the finalization timestamp.

Migration 0131's unique index `ux_documents_v2_cd_revision ON documents(controlled_document_id, revision_number)` resolves correctly. `CreateDocument` at `repository.go:37` auto-computes `MAX(revision_number)+1` and seeds `document_placeholder_values` in the same transaction.

## Key Types

```typescript
// v2/api/documentsV2.ts
type DocumentResponse = {
  ID?: string; id?: string;           // UUID
  Name?: string; name?: string;
  Status?: string; status?: string;   // "draft" | "under_review" | "approved" | ...
  Code?: string; code?: string;       // document code e.g. "DC-001"
  CurrentRevisionID?: string; current_revision_id?: string;
  CreatedBy?: string; created_by?: string;
  FormDataJSON?: Record<string, unknown>; form_data?: Record<string, unknown>;
};
```

Note: backend returns both camelCase and snake_case fields depending on endpoint version — always check both.

## Common Mistakes

- **Navigating to `/documents-v2/<uuid>` from library views:** Works correctly — `viewFromPath` maps this to `"documents-v2"` activeView, which renders `renderDocumentsV2View`. No extra wiring needed.
- **Checking `doc.status` for edit eligibility:** Status from `SearchDocumentItem` (hub list) is uppercase `"DRAFT"`. Status from `DocumentResponse` (editor API) can be lowercase `"draft"`. Normalize before comparing.
- **204 responses on PUT endpoints:** Vite dev proxy aborts 204 with no body. Backend must return 200 + `{}` body for all mutating endpoints in dev.
