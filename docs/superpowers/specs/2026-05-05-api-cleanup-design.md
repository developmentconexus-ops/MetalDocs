# Sub-Project B — API Contract Cleanup (Legacy v1 Removal)

> **Date:** 2026-05-05
> **Scope:** Remove orphaned `/api/v1/documents/*` frontend clients, dead views, dead components, and stale backend authz rules. Stub notifications. Drop dead `locked_at` column. Normalize date formatting (L1).
> **Out of scope:** `documents-v2/` canonical client (survives). Backend v2 handlers. Sub-Project A (J1/J2/M1 — already shipped, commits cb56e1e0..1cebea64). Future frontend rewrite (separate effort).
> **Bug map ref:** `wiki/bugs/audit-2026-05-04.md` — bugs I1, I2, I3, I4, L1, N1, N2, N3.

---

## Context

Frontend split between two API clients:
- **Legacy** `src/api/documents.ts` — uses `request()` + `API_BASE_URL`, prepends `/api/v1`. Backend serves `/api/v2`. Calls 404.
- **Canonical** `src/features/documents/v2/api/documentsV2.ts` — uses `apiFetch` with absolute `/api/v2/*` paths. Live.

Legacy views (`useDocumentsWorkspace`, `DocumentsWorkspaceView`, `DocumentCreateView`, `components/create/*`) consume the dead client. Sidebar still routes to them in some branches. Backend `permissions.go` carries ~15 dead `/api/v1/documents/*` authz rules. `documents.locked_at` column orphaned post edit-lock removal.

User plans full frontend rewrite. Decision: remove all legacy now, ship clean baseline.

## Goals

1. Zero `/api/v1/documents/*` calls from frontend.
2. Zero references to deleted view modules (TS compile catches drift).
3. Zero dead authz rules in `permissions.go`.
4. `documents.locked_at` dropped.
5. Single date formatter (L1).
6. Notifications stubbed (no EventSource, no fetch) until rewrite.

## Architecture

Top-down deletion across three layers, one commit per layer for bisect:

1. **Routing/sidebar** — kill legacy view enum branches in `App.tsx`, `viewFromPath` in `workspaceRoutes.ts`.
2. **Views/state** — delete `useDocumentsWorkspace.ts`, `DocumentsWorkspaceView.tsx`, `DocumentCreateView.tsx`, `components/create/` (full dir incl. `widgets/`).
3. **API clients** — surgical delete in `src/api/documents.ts`, prune `lib.api.ts` aggregator, stub `src/api/notifications.ts`.
4. **Backend authz** — drop dead v1/documents rules from `apps/api/cmd/metaldocs-api/permissions.go`.
5. **Drive-bys** — L1 date util, N1 migration 0181 drop `locked_at`.

`documents-v2/` untouched. `searchDocuments`, `fetchDocumentTypeBundle`, `uploadAttachment`, `getAttachmentDownloadURL`, `renderDocumentContentPdf`, `exportDocumentDocx` survive in `api/documents.ts` — live backends.

## File-Level Kill List

### Delete entire files
- `frontend/apps/web/src/features/documents/useDocumentsWorkspace.ts`
- `frontend/apps/web/src/features/documents/DocumentsWorkspaceView.tsx`
- `frontend/apps/web/src/components/DocumentCreateView.tsx`
- `frontend/apps/web/src/components/create/` (full dir incl. `widgets/`)

### Surgical deletions in `src/api/documents.ts`
Delete: `listDocuments`, `getDocument`, `getDocumentEditorBundle`, `getDocumentBrowserEditorBundle`, `listDocumentTemplates`, `assignDocumentTemplate`, `createDocument` (legacy), `listVersions`, `addVersion`, `getVersionDiff`, `listAttachments`, `heartbeatDocumentPresence`, `listDocumentPresence`, `acquireDocumentEditLock`, `getDocumentEditLock`, `releaseDocumentEditLock`, `saveDocumentContent`, `fetchDocumentEditorBundle`, `saveDocumentContentNative`, `getDocumentContentNative`, `saveDocumentBrowserContent`, `uploadDocumentContentDocx`, `downloadDocumentTemplateDocx`.

Keep: `searchDocuments` (live backend `/api/v1/search/documents`), `fetchDocumentTypeBundle`, `uploadAttachment`, `getAttachmentDownloadURL`, `renderDocumentContentPdf`, `exportDocumentDocx`, `getDocumentContentPdf`, `getDocumentContentDocx`.

### Stub `src/api/notifications.ts`
```ts
export async function listNotifications() { return { items: [] }; }
export async function markNotificationRead(_id: string) { return; }
export async function listAuditEvents() { return { items: [] }; }
export function subscribeOperationsStream() { return () => {}; }
export type OperationsStreamSnapshot = { items: never[] };
```
No EventSource. No fetch.

### Trim `src/lib.api.ts`
Remove imports + `api.*` entries for every deleted function. Keep aggregator pattern (rewrite later).

### Trim `src/App.tsx` + `workspaceRoutes.ts`
Delete view enum branches that mount deleted components. Default route → /documents (v2).

### Backend `apps/api/cmd/metaldocs-api/permissions.go`
Delete every rule whose path matches `/api/v1/documents*`. Keep `/api/v1/search/documents`, `/api/v1/audit/events`, `/api/v1/iam/users`.

### Drive-bys
- **L1:** introduce `src/lib/formatDate.ts` exporting `formatISODate(d)` and `formatDateTime(d)`. Replace ad-hoc `toLocaleDateString` calls in surviving views.
- **N1:** migration `0181_drop_documents_locked_at.sql`:
  ```sql
  -- up
  ALTER TABLE public.documents DROP COLUMN IF EXISTS locked_at;
  -- down
  ALTER TABLE public.documents ADD COLUMN locked_at TIMESTAMPTZ NULL;
  ```

## Commit Sequence

1. `chore(routing): drop legacy document view branches`
2. `chore(frontend): delete legacy documents workspace + create views`
3. `chore(api-client): prune dead /api/v1/documents calls`
4. `chore(notifications): stub client until rewrite`
5. `chore(backend): drop dead v1/documents authz rules`
6. `chore(db): drop unused documents.locked_at column (0181)`
7. `chore(ui): consolidate date formatting helper (L1)`

## Verification

Per commit:
- `pnpm tsc --noEmit` green.
- `pnpm vitest run` green (orphan tests die with their files).
- `go build ./...` + `go test ./internal/modules/documents/... ./apps/api/...` green for backend commits.
- Manual smoke: login → /documents → list → open → edit → submit → approve. Routine A0 from `wiki/tests/system-acceptance-test.md`.
- Network tab: zero `/api/v1/documents/*`, zero `/api/v1/notifications`, zero `/api/v1/operations/stream`.

Pre-N1 check:
```sql
SELECT COUNT(*) FROM documents WHERE locked_at IS NOT NULL;
```
Expect 0. Abort drop if non-zero. Grep `locked_at` in `apps/api/` + `internal/` post Commit 2 — must be zero hits before 0181.

## Rollback

Each commit `git revert`-able. Migration 0181 reversible via down (column data loss acceptable — orphan).

## Risks

- **Hidden caller of deleted func.** Mitigation: tsc catches at compile time. If runtime caller (dynamic dispatch) missed, revert Commit 3.
- **Sidebar route default mismatch.** Mitigation: smoke test post Commit 1.
- **`searchDocuments` accidentally deleted.** Mitigation: explicit keep list in spec; verify post Commit 3 grep.

## Wiki Updates

Post-merge, dispatch `wiki-curator`:
- Update `wiki/modules/documents.md` — remove legacy client refs, bump Last verified.
- Update `wiki/bugs/audit-2026-05-04.md` — mark I1/I2/I3/I4/L1/N1/N2/N3 fixed with commit SHAs.
- Update `wiki/README.md` index entry for audit doc.

## Out of Scope (Defer)

- Frontend rewrite (separate effort).
- `lib.api.ts` god-aggregator restructure (rewrite will replace).
- Backend v1 route handler removal (audit separately if any survive permissions.go prune).
- OTel/Prom instrumentation (not wired in project; do not add).
