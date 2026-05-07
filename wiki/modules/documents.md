# documents Module

> **Last verified:** 2026-05-07
> **Scope:** Document instances — library listing, novo-documento wizard (`/documents-v2/new`), edit, autosave, finalize.
> **Out of scope:** Template authoring (see `modules/templates-v2.md`), approval routes (`modules/approval.md`), PDF fanout (`modules/render-fanout.md`).
> **Key files:**
> - `frontend/apps/web/src/features/documents/pages/LibraryPage.tsx:36` — `LibraryPage` — server-side paginated table at `/documents`; lazy `useState` init, debounced search, `resolveErrorMessage` error UX
> - `frontend/apps/web/src/features/documents/api/library.ts:34` — `fetchLibrary` — `GET /api/v2/documents` typed client; `asApiError` wrapper ensures all errors surface as `ApiError`
> - `frontend/apps/web/src/features/documents/api/library.ts:46` — `fetchLibraryStats` — `GET /api/v2/documents/stats` typed client
> - `frontend/apps/web/src/features/documents/api/library.ts:28` — `asApiError` — wraps any 2xx-envelope error into a real `ApiError` so `resolveErrorMessage` works downstream
> - `frontend/apps/web/src/features/documents/lib/libraryStatus.ts:1` — `LIBRARY_STATUSES` + `LibraryFilter` + `filterToStatus()` — single source of truth for backend status ↔ URL filter ↔ pt-BR label
> - `frontend/apps/web/src/features/documents/lib/libraryStatus.ts:29` — `LIBRARY_STATUSES` array (6 entries: draft → obsolete)
> - `frontend/apps/web/src/features/documents/lib/libraryStatus.ts:38` — `filterToStatus()` — maps `LibraryFilter` slug to `DocumentStatus | undefined`
> - `frontend/apps/web/src/features/documents/lib/visibilityMeta.ts:1` — `VISIBILITY_META` + `VISIBILITY_KEYS` + `VisibilityKey` — SSOT for wizard visibility options (area/people/company/external); icon mapping uses existing `IconName` union (option B); visibility is captured client-side only — no backend field today (see `backlog/novo-documento.md#visibility`)
> - `frontend/apps/web/src/features/documents/components/AuthorCell.tsx:30` — `AuthorCell` — Avatar + deterministic hashed color + first-name display for document rows
> - `frontend/apps/web/src/features/documents/components/LibraryFilterTabs.tsx:11` — `LibraryFilterTabs` — iterates `LIBRARY_STATUSES`; prop type tightened to `LibraryFilter`; Filtros/Exportar disabled with `aria-disabled`
> - `frontend/apps/web/src/features/documents/components/LibrarySidebar.tsx:32` — `LibrarySidebar` — iterates `LIBRARY_STATUSES` for status section
> - `frontend/apps/web/src/features/documents/components/Pagination.tsx:9` — `Pagination` — prev/next controls
> - `frontend/apps/web/src/features/documents/components/PageSizeSelector.tsx:9` — `PageSizeSelector` — 10/20/50 dropdown
> - `frontend/apps/web/src/features/documents/queries/useLibraryQuery.ts:15` — `useLibraryQuery` — TanStack Query hook; `placeholderData: keepPreviousData` prevents empty flash on page/filter change
> - `frontend/apps/web/src/features/documents/queries/useLibraryStatsQuery.ts:5` — `useLibraryStatsQuery` — TanStack Query hook; `staleTime: 30_000` avoids refetch on every focus
> - `frontend/apps/web/src/features/documents/routes.tsx:1` — route definitions for `/documents` (Library), `/documents-v2/new` (wizard), and `/documents-v2/:id` (editor)
> - `frontend/apps/web/src/features/documents/routes.tsx:55` — `documents-v2/new` route: lazy-loads `NewDocumentWizardPage`
> - `frontend/apps/web/src/features/documents/pages/NewDocumentWizardPage.tsx:44` — `NewDocumentWizardPage` — 4-step wizard entry; `useReducer(wizardReducer)` for form state; `?step=1..4` URL param; 2-call submit sequence (slot → doc)
> - `frontend/apps/web/src/features/documents/components/wizard/WizardShell.tsx:1` — stepper chrome + layout shell for all 4 steps
> - `frontend/apps/web/src/features/documents/components/wizard/steps/StepProfile.tsx:1` — Step 1: profile radio cards; calls `GET /api/v2/taxonomy/profiles`
> - `frontend/apps/web/src/features/documents/components/wizard/steps/StepAreaCodeVisibility.tsx:1` — Step 2: area picker + title + visibility; calls `GET /api/v2/taxonomy/areas`; shows `{profile}-{area}-???` code preview
> - `frontend/apps/web/src/features/documents/components/wizard/steps/StepTemplate.tsx:1` — Step 3: template selector; calls `GET /api/v2/templates` filtered by profile; only shows templates with `published_version_id`; "Em branco" intentionally disabled
> - `frontend/apps/web/src/features/documents/components/wizard/steps/StepConfirm.tsx:1` — Step 4: read-only summary + consent checkbox + "Criar documento" trigger
> - `frontend/apps/web/src/features/documents/components/wizard/CodePreviewBanner.tsx:1` — inline code preview banner (`{profile}-{area}-???`); server resolves sequence at create time
> - `frontend/apps/web/src/features/documents/pages/DocumentEditorPage.tsx:200` — layout root: `<aside class={styles.rail}>` left rail + `<main class={styles.canvas}>` + `EditorChrome` center/right slots; `handleFinalize` catches `ApiError`, calls `resolveErrorMessage` for toast (E3)
> - `frontend/apps/web/src/features/documents/pages/styles/DocumentEditorPage.module.css:19` — `.rail`, `.railBackBtn`, `.railTip` — design-token-only; mirrors TemplateAuthorPage rail
> - `frontend/apps/web/src/features/documents/pages/DocumentCreatePage.tsx:1` — legacy step 1: pick controlled document (pre-wizard flow)
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
> - `src/api/documents.ts` (legacy v1 client, 496 lines) — deleted. Canonical client: `features/documents/api/documentsV2.ts`.
> - `src/components/DocumentCreateView.tsx`, `DocumentsWorkspaceView`, `useDocumentsWorkspace` — deleted. Replaced by `features/documents/pages/DocumentCreatePage.tsx` + `DocumentEditorPage.tsx`.
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
/documents-v2/new          → NewDocumentWizardPage (4-step create wizard)
/documents-v2/<uuid>       → DocumentEditorPage
```

Routes are defined in `features/documents/routes.tsx`. The `documents-v2/new` registration is at `routes.tsx:55`.

## Library Screen (`/documents`)

`LibraryPage` is the primary document list view. Key design decisions:

- **Two-query pattern:** `GET /api/v2/documents` (LIMIT/OFFSET items) + `GET /api/v2/documents/stats` (GROUP BY counts) run in parallel via TanStack Query. No full-list fetch.
- **Server-side pagination:** page, pageSize (10/20/50, cap=50), status filter, areaCode filter, profileCode filter, and free-text `q` (ILIKE on name) are all sent to the backend. Client never filters/sorts in memory.
- **RBAC scoping:** `document_filler` users are auto-scoped to their own documents server-side via `effectiveUserID` in `parseListOptions`; `system_admin` sees all.
- **pageSize cap:** backend returns 400 if `pageSize > 50` (defense-in-depth; selector is also limited to 10/20/50).
- **Status meta — single source of truth:** `features/documents/lib/libraryStatus.ts` exports `LIBRARY_STATUSES`, `LibraryFilter`, and `filterToStatus()`. Both `LibraryFilterTabs` and `LibrarySidebar` iterate `LIBRARY_STATUSES` — no local STATUS_ITEMS / TABS arrays in component files.
- **Filter tabs:** 6 status tabs (Rascunhos / Em Revisão / Aprovados / Publicados / Rejeitados / Obsoletos) + "Todos" root, rendered by `LibraryFilterTabs`. Filtros and Exportar buttons are `disabled + aria-disabled` until backend endpoints exist.
- **Debounced search:** `searchQuery` state is passed to `useDebouncedValue(q, 300)` before inclusion in the query params; resets `page` to 1 on change.
- **Lazy `useState` init:** `readStoredPageSize()` and `readStoredActivityOpen()` are passed as initializer functions (not expressions) to avoid unnecessary `localStorage` reads on re-renders and prevent hydration flash.
- **Area navigation:** `LibrarySidebar` contains the ÁREAS section (no separate `LibraryAreaTree` component); it fetches process areas via `QK.taxonomy.areas()` and adds `areaCode` to the query on selection.
- **Activity sidebar:** 320px, default-collapsed, toggle persisted to `localStorage`. Placeholder only — `ActivityPanel` and `LibraryStatCards` carry TODO trails for hardcoded mocks; approval inbox integration deferred.
- **AuthorCell:** `features/documents/components/AuthorCell.tsx` renders Avatar + first name per row. Avatar background color is derived deterministically from the username hash (stable across sessions, no server round-trip).
- **Row accessibility:** document rows are `<div role="button">` (valid, keyboard-accessible) with `:focus-visible` outline in CSS.
- **StatusPill:** extended to all 8 Spec 2 states (`draft`, `under_review`, `approved`, `rejected`, `scheduled`, `published`, `superseded`, `obsolete`) with Portuguese labels.
- **Error UX:** `resolveErrorMessage(apiError.code, apiError.message)` used directly in `LibraryPage` to surface query errors inline (not toast). Pattern matches `concepts/error-ux.md`.

## Library Patterns

### Status-meta module (`libraryStatus.ts`)

```typescript
// features/documents/lib/libraryStatus.ts:29
export const LIBRARY_STATUSES: readonly StatusEntry[] = [
  { status: 'draft',        filter: 'rascunhos',  label: 'Rascunhos'  },
  { status: 'under_review', filter: 'em_revisao', label: 'Em Revisão' },
  { status: 'approved',     filter: 'aprovados',  label: 'Aprovados'  },
  { status: 'published',    filter: 'publicados', label: 'Publicados' },
  { status: 'rejected',     filter: 'rejeitados', label: 'Rejeitados' },
  { status: 'obsolete',     filter: 'obsoletos',  label: 'Obsoletos'  },
];
```

`filterToStatus('todos')` returns `undefined` (no status param → server returns all). Components must not redeclare their own status/label lists — import from this module.

### `asApiError` wrapper (`library.ts:28`)

Wraps any error from a successful HTTP response (2xx with error envelope) into a real `ApiError` so that `resolveErrorMessage` works downstream without null-checking. 4xx/5xx are already `ApiError` from `apiFetch`.

### TanStack Query defaults for library hooks

| Hook | Key option | Reason |
|------|-----------|--------|
| `useLibraryQuery` | `placeholderData: keepPreviousData` | Hold previous page while new page/filter loads — no empty-state flash |
| `useLibraryStatsQuery` | `staleTime: 30_000` | Stats are stable; skip refetch on every window focus/mount |

### `useDebouncedValue` (`lib/hooks/useDebouncedValue.ts:10`)

Generic hook for debouncing any value. Used by `LibraryPage` for the search input (300 ms). Lives in `lib/hooks/` because it has no domain dependency — reusable across features.

## Create Flow — Novo-Documento Wizard

The 4-step wizard at `/documents-v2/new` (`NewDocumentWizardPage`) replaced the old `DocumentCreatePage` single-step flow. State lives in `useReducer(wizardReducer)`. Step is driven by `?step=1..4` URL param; refreshing retains the step but resets in-memory form state.

| Step | Component | API call |
|---|---|---|
| 1 — Profile | `StepProfile` | `GET /api/v2/taxonomy/profiles` |
| 2 — Area + Title + Visibility | `StepAreaCodeVisibility` | `GET /api/v2/taxonomy/areas` |
| 3 — Template | `StepTemplate` | `GET /api/v2/templates?profileCode=…` (published only) |
| 4 — Confirm + Create | `StepConfirm` | 2-call submit: `POST /api/v2/controlled-documents` → `POST /api/v2/documents` |

**2-call submit sequence** (in `handleCreate`, `NewDocumentWizardPage.tsx:112`):
1. `POST /api/v2/controlled-documents` (slot reservation) — returns the CD with a server-resolved code (e.g. `PROC-02`). Code preview in steps 2–4 shows `{profile}-{area}-???` because no preview endpoint exists.
2. `POST /api/v2/documents` — clones the selected template version into a new draft document, linked to the CD slot.
3. Redirect to `/documents-v2/${doc.document_id}`.

**Slot-rollback gap:** if step 2 (`POST /api/v2/documents`) fails after step 1 succeeds, the slot remains in the registry and its sequence number is consumed. No automatic compensation today. See `backlog/novo-documento.md#slot-rollback`.

**Deferred items:** visibility enforcement, blank-template path, profile counts. See `wiki/backlog/novo-documento.md`.

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

Checkpoints are manual snapshots stored on the server. The `CheckpointsDialog` component and checkpoint API endpoints exist but the dialog is **not currently mounted** from `DocumentEditorPage` — the "Revisões" button was removed in the 2026-05-06 editor layout pass. Restoring a checkpoint would re-fetch the revision buffer and reload the editor when re-enabled.

## Editor Chrome

`DocumentEditorPage` uses `<EditorChrome>` (see `modules/editor-chrome.md`) — the shared toolbar overlay primitive used by both the document editor and `TemplateAuthorPage`. The intermediate `EditorDocBar.tsx` + `EditorDocBar.module.css` were deleted when this page was migrated.

**Layout:** The page shell is `<div.page> → <div.body>` which holds three siblings:
1. `<aside class={styles.rail}>` — 48px left rail with the back ("Voltar") button + tooltip. The `EditorChrome` `left` slot is **not used** here; the rail is an independent `<aside>` element to avoid collision with eigenpal's own toolbar.
2. `<main class={styles.canvas}>` — flex:1 host for `EditorChrome`.
3. `<EditorMetaSidebar>` — collapsible right metadata panel (300px when open).

Slot assignment inside `EditorChrome`:
- `left` — **unused** (back button is in the `<aside>` rail, not in the chrome overlay)
- `center` — `CodeChip` (doc code) + document name + `VersionBadge` (revision) + `StatusPill`
- `right` — `AutosaveStatus` + "Submeter para revisão" button only

**Removed from right slot (intentional):** `CheckpointsDialog` mount + `checkpointsOpen` state + `handleRestored` callback were deleted as orphans. `ExportMenuButton` and the "Revisões" button were also removed. These components still exist as standalone files but are not currently mounted from the editor page.

Eigenpal CSS overrides (wine formatting bar, compact title bar, gradient scrollbar) live in `EditorChrome.module.css`; `DocumentEditorPage.module.css` covers only page-level rail + canvas layout (`.rail`, `.railBackBtn`, `.railTip`, `.canvas`).

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
// features/documents/api/documentsV2.ts  ← canonical API client (legacy api/documents.ts deleted in b0158354)
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
