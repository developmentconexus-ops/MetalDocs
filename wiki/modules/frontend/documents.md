# Frontend module: documents

> **Last verified:** 2026-06-01 (P2 consolidation: added Failure modes section)
> **Scope:** Library, document detail, eigenpal-based editor, distribution, new-document wizard. Frontend slice of the backend [`documents`](../documents.md) module.
> **Owner:** unassigned | **Backend counterpart:** [`wiki/modules/documents.md`](../documents.md)

## 1. Purpose

Owns the document lifecycle UI: discoverability (Library), authored revision (Editor), the published artifact view, and the create-new flow. Wraps the eigenpal editor through the local ACL (see [`editor-ui-eigenpal.md`](../editor-ui-eigenpal.md)).

## 2. Key files

- [`frontend/apps/web/src/features/documents/routes.tsx:1`](../../../frontend/apps/web/src/features/documents/routes.tsx) — route table; legacy `/documents/*` paths redirect to `/documents`.
- [`frontend/apps/web/src/features/documents/pages/LibraryPage.tsx:36`](../../../frontend/apps/web/src/features/documents/pages/LibraryPage.tsx) — Library shell.
- [`frontend/apps/web/src/features/documents/pages/DocumentDetailLayout.tsx`](../../../frontend/apps/web/src/features/documents/pages/DocumentDetailLayout.tsx) — outlet for `/documents/:id` and `/distribution`.
- [`frontend/apps/web/src/features/documents/pages/DocumentPublishedPage.tsx:121`](../../../frontend/apps/web/src/features/documents/pages/DocumentPublishedPage.tsx) — canonical published view; reads approval-instance for settled metadata.
- [`frontend/apps/web/src/features/documents/pages/DocumentEditorPage.tsx:44`](../../../frontend/apps/web/src/features/documents/pages/DocumentEditorPage.tsx) — eigenpal-hosted editor; submit invalidates approval + document caches at lines 288–300.
- [`frontend/apps/web/src/features/documents/pages/NewDocumentWizardPage.tsx:69`](../../../frontend/apps/web/src/features/documents/pages/NewDocumentWizardPage.tsx) — create-new wizard; uses `WizardShell`.
- [`frontend/apps/web/src/features/documents/api/documents.ts`](../../../frontend/apps/web/src/features/documents/api/documents.ts), [`library.ts`](../../../frontend/apps/web/src/features/documents/api/library.ts), [`exports.ts`](../../../frontend/apps/web/src/features/documents/api/exports.ts).
- Query hooks under [`frontend/apps/web/src/features/documents/queries/`](../../../frontend/apps/web/src/features/documents/queries/).

## 3. Routes

| Path | Component | Handle |
|---|---|---|
| `/documents` | `LibraryPage` | `workspaceView: 'library'` |
| `/documents/new` | `NewDocumentWizardPage` (routes.tsx:42) | `workspaceView: 'document-editor'` |
| `/documents/:documentId` | `DocumentDetailLayout` → `DocumentPublishedPage` (index) | `workspaceView: 'library'` |
| `/documents/:documentId/distribution` | `DocumentDistributionPage` (routes.tsx:34) | — |
| `/documents/:documentId/edit` | `DocumentEditorRoutePage` (routes.tsx:22) | `workspaceView: 'document-editor'` |
| `/documents/all`, `/area/:code`, `/type/:code`, `/doc/:id`, `/mine`, `/recent` | `<Navigate to="/documents" replace />` | legacy aliases |

## 4. TanStack Query

| Key | Source | Owner hook |
|---|---|---|
| `QK.documents.list(params)` | `lib/queryKeys.ts:30` | `useLibraryQuery` (`useLibraryQuery.ts:17`) |
| `QK.documents.stats()` | `lib/queryKeys.ts:31` | `useLibraryStatsQuery` |
| `QK.documents.detail(id)` | `lib/queryKeys.ts:32` | `useDocumentDetailQuery` |
| `QK.documents.revisionHistory(id)` | `lib/queryKeys.ts:33` | `useDocumentRevisionHistoryQuery` |
| `QK.documents.comments(id)` | `lib/queryKeys.ts:34` | `useDocumentCommentsQuery` |
| `QK.approval.instance(id)` | `lib/queryKeys.ts:61` | `useApprovalInstanceQuery` |
| `QK.controlledDocuments.activeDocument(id)` | `lib/queryKeys.ts:46` | `useControlledDocumentActiveDocumentQuery` |
| `QK.templates.blank()` / `QK.templates.byProfile(code)` | `lib/queryKeys.ts:56–57` | `useBlankTemplateQuery`, `useTemplatesByProfileQuery` |

**Invalidation rules:** editor commit/submit updates `QK.documents.detail`, `QK.documents.revisionHistory`, `QK.approval.instance` (`DocumentEditorPage.tsx:288–300`). Wizard preview invalidates `QK.controlledDocuments.preview(profileCode, areaCode)` on profile/area change (`NewDocumentWizardPage.tsx:175`).

## 5. API endpoints consumed

- `GET /api/v1/documents` (library list), `GET /api/v1/documents/{id}`, `GET /api/v1/documents/{id}/revisions`, `GET /api/v1/documents/{id}/comments`.
- `POST /api/v1/documents/{id}/edit/presign` + `POST /api/v1/documents/{id}/edit/commit` — direct-to-MinIO autosave (see [`sequence-edit-autosave.md`](../../diagrams/sequence-edit-autosave.md)).
- `POST /api/v1/documents/{id}/finalize` — transitions draft → under_review (approval module owner).
- `GET /api/v1/documents/{id}/approval-instance` — approval metadata (see [`frontend/approval.md`](approval.md)).
- `POST /api/v1/documents/{id}/export/pdf` — async PDF (see [`sequence-pdf-export.md`](../../diagrams/sequence-pdf-export.md)).
- Wizard: `POST /api/v1/controlled-documents/atomic`, `GET /api/v1/controlled-documents/preview-code`.

## 6. Dependencies

**Imports from:**
- `features/shared/components/editor-chrome/` — `EditorChrome`, `VersionBadge`, `AutosaveStatus`.
- `features/shared/components/wizard/` — `WizardShell`, `WizardFooter`.
- `features/controlled-documents/` — preview-code query, atomic-create API.
- `editor-adapters/` — eigenpal ACL glue.

**Imported by:**
- `features/dashboard/` — recent-documents tile reuses `QK.documents.list`.
- `features/approval/` — inbox links navigate to `/documents/:id/edit` when `under_review`.

## 7. Invariants

- All server state via TanStack Query; no `useEffect` data-fetching.
- Eigenpal interaction never leaks past `editor-adapters/` — see [`wiki/modules/editor-ui-eigenpal.md`](../editor-ui-eigenpal.md) for the ACL boundary.
- Autosave uses presigned-PUT direct to MinIO; the server only receives the `commit` call after upload.
- Editor uses optimistic concurrency on `revision_version` (If-Match header); stale ⇒ 409 surfaced inline.
- Legacy paths (`/documents/all`, `/area/:code`, etc.) only redirect; no business logic remains.

## 8. Known issues / tech-debt

- See backend [`documents-tech-debt.md`](../documents-tech-debt.md).
- Editor relies on eigenpal native placeholders; legacy MetalDocs tokens gap tracked in [`wiki/concepts/placeholders.md`](../../concepts/placeholders.md).

## 9. Failure modes

| Failure | Symptom | Detection | Response |
|---|---|---|---|
| MinIO presigned-PUT fails mid-autosave | Editor `AutosaveStatus` flips to error; commit not called | `presignAutosave` returned 200 but the subsequent direct PUT to MinIO rejected (CORS / 5xx); surfaces as fetch error in editor adapter | User retries (autosave re-runs on next mutation); if persistent, MinIO healthcheck + bucket CORS |
| `commit` 409 stale `If-Match` (concurrent revision write) | Editor surfaces `state.stale_revision` inline; revision not bumped | `documents.ts` commit returns `ApiError.code === 'state.stale_revision'` | Refetch `useDocumentDetailQuery`; user resolves divergence then retries |
| Backend 5xx on `useLibraryQuery` | Library shell shows error state | `useLibraryQuery.error` | Retry; check `metaldocs-api` |
| 401 during editor session | `authBus` redirect; in-flight commit lost | Editor mutation rejects with 401 → authBus | Re-login; returnTo brings user back to `/documents/:id/edit` |
| Approval-instance cache stale after submit | Editor sidebar shows `draft` while backend already opened `under_review` | `useApprovalInstanceQuery` not invalidated | `DocumentEditorPage.tsx:288–300` invalidates `QK.approval.instance(id)` + `QK.documents.detail(id)` on submit success |
| Export PDF outbox still pending | Distribution / published view shows `pdf_status=pending` longer than expected | Backend `pdf_outbox_repository.ReadState` surfaces status | Worker drains async; if `failed`, check worker logs (`render-fanout` module) |

## 10. Cross-links

- Backend module: [`wiki/modules/documents.md`](../documents.md)
- Sequences: [`create-document`](../../diagrams/sequence-create-document.md), [`edit-autosave`](../../diagrams/sequence-edit-autosave.md), [`pdf-export`](../../diagrams/sequence-pdf-export.md)
- Editor ACL: [`wiki/modules/editor-ui-eigenpal.md`](../editor-ui-eigenpal.md)
- Skill: [`metaldocs-frontend`](../../../.agents/skills/metaldocs-frontend/SKILL.md), [`metaldocs-tanstack-query`](../../../.agents/skills/metaldocs-tanstack-query/SKILL.md)
