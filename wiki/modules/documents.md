# documents Module

> **Last verified:** 2026-05-04
> **Scope:** Document instances — create, edit, autosave, checkpoints, finalize, export.
> **Out of scope:** Template authoring (see `modules/templates-v2.md`), approval routes (`modules/approval.md`), PDF fanout (`modules/render-fanout.md`).
> **Key files:**
> - `frontend/apps/web/src/features/documents/v2/DocumentEditorPage.tsx:126` — `handleFinalize` — catches `ApiError`, calls `resolveErrorMessage` for toast (E3)
> - `frontend/apps/web/src/features/documents/v2/styles/DocumentEditorPage.module.css:1` — wine-brand chrome CSS
> - `frontend/apps/web/src/features/documents/v2/routes.tsx:1` — route parsing/rendering for `/documents-v2/*`
> - `frontend/apps/web/src/features/documents/v2/DocumentCreatePage.tsx:1` — step 1: pick controlled document
> - `frontend/apps/web/src/features/documents/DocumentsHubView.tsx:758` — detail panel with Edit/PDF/Duplicate actions
> - `internal/modules/documents/delivery/http/handler.go:73` — `NewHandlerWithSubmit` — wires db + submitSvc for atomic finalize
> - `internal/modules/documents/delivery/http/handler.go:259` — `finalizeDocument` — resolves approval route, calls SubmitRevisionForReview
> - `internal/modules/documents/application/service.go:1` — domain logic, session management
> - `internal/modules/documents/repository/repository.go:35` — `CreateDocument` INSERT with `MAX(revision_number)+1` auto-increment
> - `internal/modules/documents/module.go:1` — DI wiring

## Overview

A **document** is an instance filled from a template version, bound to a controlled document entry.
Documents move through states: `draft → under_review → approved → published`.
Only `draft` documents can be edited in the editor.

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

Migration 0167 fixed missing columns that were mistakenly added to the now-dropped `public.documents_v2` table (migrations 0126/0129). `public.documents` now has the full schema:

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

Migration 0131's unique index `ux_documents_v2_cd_revision ON documents(controlled_document_id, revision_number)` now resolves correctly because `controlled_document_id` exists. `CreateDocument` at `repository.go:35` auto-computes `MAX(revision_number)+1` — the previous default-to-1 gap is fixed.

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

## See also

- [concepts/error-ux.md](../concepts/error-ux.md) — `apiFetch` wrapper used in `DocumentEditorPage`; `resolveErrorMessage` for finalize error toast (E3)
- [modules/approval.md](approval.md)
- [workflows/approval.md](../workflows/approval.md)
