# documents Module

> **Last verified:** 2026-05-06
> **Scope:** Document instances — library listing, create, edit, autosave, checkpoints, finalize, export.
> **Out of scope:** Template authoring (see `modules/templates-v2.md`), approval routes (`modules/approval.md`), PDF fanout (`modules/render-fanout.md`).
> **Key files:**
> - `frontend/apps/web/src/features/documents/pages/LibraryPage.tsx:45` — `LibraryPage` — server-side paginated table at `/documents`
> - `frontend/apps/web/src/features/documents/api/library.ts:23` — `fetchLibrary` — `GET /api/v2/documents` typed client
> - `frontend/apps/web/src/features/documents/api/library.ts:41` — `fetchLibraryStats` — `GET /api/v2/documents/stats` typed client
> - `frontend/apps/web/src/features/documents/queries/useLibraryQuery.ts:15` — `useLibraryQuery` — TanStack Query hook for paginated list
> - `frontend/apps/web/src/features/documents/queries/useLibraryStatsQuery.ts:5` — `useLibraryStatsQuery` — TanStack Query hook for stats
> - `frontend/apps/web/src/features/documents/components/LibraryFilterTabs.tsx:22` — `LibraryFilterTabs` — 7-tab filter strip mapped to real Spec 2 states
> - `frontend/apps/web/src/features/documents/components/LibraryAreaTree.tsx:12` — `LibraryAreaTree` — SectionPanel area-tree navigation
> - `frontend/apps/web/src/features/documents/components/Pagination.tsx:9` — `Pagination` — prev/next controls
> - `frontend/apps/web/src/features/documents/components/PageSizeSelector.tsx:9` — `PageSizeSelector` — 10/20/50 dropdown
> - `frontend/apps/web/src/features/documents/routes.tsx:1` — route definitions for `/documents` (Library) and `/documents-v2/*` (Editor)
> - `frontend/apps/web/src/features/documents/pages/DocumentEditorPage.tsx:1` — editor page; `handleFinalize` catches `ApiError`, calls `resolveErrorMessage` for toast (E3)
> - `frontend/apps/web/src/features/documents/pages/DocumentCreatePage.tsx:1` — step 1: pick controlled document
> - `internal/modules/documents/delivery/http/handler.go:76` — `NewHandlerWithSubmit` — wires db + submitSvc for atomic finalize
> - `internal/modules/documents/delivery/http/handler.go:189` — `listDocuments` — paginated `GET /api/v2/documents`; filler auto-scoped server-side
> - `internal/modules/documents/delivery/http/handler.go:219` — `documentStats` — `GET /api/v2/documents/stats`
> - `internal/modules/documents/delivery/http/handler.go:244` — `parseListOptions` — shared query-param parser; pageSize cap=50 returns 400
> - `internal/modules/documents/application/service.go:26` — `Repository` interface; includes `ListDocumentsPaginated`, `CountDocuments`, `StatsByStatus`, `StatsByArea`
> - `internal/modules/documents/application/list_options.go:1` — `ListOptions` type alias (alias for `repository.ListOptions`)
> - `internal/modules/documents/repository/repository.go:253` — `ListOptions` struct with `IncludeArchived`
> - `internal/modules/documents/repository/repository.go:284` — `buildDocumentFilter` — shared WHERE builder (list, count, stats)
> - `internal/modules/documents/repository/repository.go:315` — `ListDocumentsPaginated` — LIMIT/OFFSET query
> - `internal/modules/documents/repository/repository.go:348` — `CountDocuments` — COUNT(*) with same filter
> - `internal/modules/documents/repository/repository.go:358` — `StatsByStatus` — GROUP BY status
> - `internal/modules/documents/repository/repository.go:379` — `StatsByArea` — GROUP BY area
> - `internal/modules/documents/repository/repository.go:39` — `CreateDocument` INSERT with `MAX(revision_number)+1` auto-increment
> - `internal/modules/documents/module.go:1` — DI wiring

## Overview

> **Cleanup note (2026-05-05, branch chore/api-cleanup-sub-project-b):**
> - `src/api/documents.ts` (legacy v1 client, 496 lines) — deleted. Canonical client: `features/documents/v2/api/documentsV2.ts`.
> - `src/components/DocumentCreateView.tsx`, `DocumentsWorkspaceView`, `useDocumentsWorkspace` — deleted. Replaced by `features/documents/v2/DocumentCreatePage.tsx` + `DocumentEditorPage.tsx`.
> - `api/notifications.ts` — stubbed pending backend rewrite (commit d136449a).
> - Edit-lock (`/api/v1/documents/:id/lock`) and presence endpoints — dropped from backend + client (I4, commit b0158354).
> - `documents.locked_at` column — dropped (migration 0181, commit c866de8a).

A **document** is an instance filled from a template version, bound to a controlled document entry.
Documents move through states: `draft → under_review → approved → published`.
Only `draft` documents can be edited in the editor.

**Backend module path:** `internal/modules/documents/` (renamed from `documents_v2` in migration batch 0167/0168)
**Table:** `public.documents`

## Frontend Routing

```
/documents                 → LibraryPage (server-side paginated document list)
/documents-v2/new          → DocumentCreatePage (pick controlled document)
/documents-v2/<uuid>       → DocumentEditorPage
```

Routes are defined in `features/documents/routes.tsx`.

## Library Screen (`/documents`)

`LibraryPage` is the primary document list view. Key design decisions:

- **Two-query pattern:** `GET /api/v2/documents` (LIMIT/OFFSET items) + `GET /api/v2/documents/stats` (GROUP BY counts) run in parallel via TanStack Query. No full-list fetch.
- **Server-side pagination:** page, pageSize (10/20/50, cap=50), status filter, areaCode filter, profileCode filter, and free-text `q` (ILIKE on name) are all sent to the backend. Client never filters/sorts in memory.
- **RBAC scoping:** `document_filler` users are auto-scoped to their own documents server-side via `effectiveUserID` in `parseListOptions`; `system_admin` sees all.
- **pageSize cap:** backend returns 400 if `pageSize > 50` (defense-in-depth; selector is also limited to 10/20/50).
- **Filter tabs:** 7 tabs (Todos / Meus / Rascunhos / Em revisão / Publicados / Rejeitados / Obsoletos) mapped to Spec 2 8-state model. "Meus" triggers server-side user scope, not a status filter.
- **SectionPanel:** `LibraryAreaTree` on the left (224px) lists process areas from the taxonomy API; selecting an area adds `areaCode` to the query.
- **Activity sidebar:** 320px, default-collapsed, toggle persisted to `localStorage`. Placeholder only in Phase 4 — approval inbox integration deferred.
- **StatusPill:** extended to all 8 Spec 2 states (`draft`, `under_review`, `approved`, `rejected`, `scheduled`, `published`, `superseded`, `obsolete`) with Portuguese labels.

See `wiki/implementation/plan-library.md` for full implementation plan and phase-by-phase detail.

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
| GET | `/api/v2/documents` | Paginated list; params: `page`, `pageSize` (max 50), `status` (comma-sep or repeated), `areaCode`, `profileCode`, `q`, `includeArchived`. Returns `{items, page, pageSize, total}`. |
| GET | `/api/v2/documents/stats` | Counts grouped by `{byStatus, byArea}`. Same filter params (minus pagination). |
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
| `content_hash_at_submit` | TEXT | Hash at submission time |
| `status` | TEXT | Extended CHECK; includes `draft`, `under_review`, `approved`, `published`, etc. |

Migration 0131's unique index `ux_documents_v2_cd_revision ON documents(controlled_document_id, revision_number)` now resolves correctly because `controlled_document_id` exists. `CreateDocument` at `repository.go:35` auto-computes `MAX(revision_number)+1` — the previous default-to-1 gap is fixed.

## Key Types

```typescript
// features/documents/v2/api/documentsV2.ts  ← canonical API client (legacy api/documents.ts deleted in b0158354)
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
