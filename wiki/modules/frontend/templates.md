# Frontend module: templates

> **Last verified:** 2026-06-01 (P2 consolidation: added Failure modes section)
> **Scope:** Template list, create-new wizard, eigenpal-based template editor with draft/review/approve lifecycle. Frontend slice of the backend [`templates`](../templates.md) module.
> **Owner:** unassigned | **Backend counterpart:** [`wiki/modules/templates.md`](../templates.md)

## 1. Purpose

Lets template authors create, edit, submit for review, and approve template versions. The editor is the eigenpal canvas plus MetalDocs chrome (outline, placeholders, version actions).

## 2. Key files

- [`frontend/apps/web/src/features/templates/routes.tsx:1`](../../../frontend/apps/web/src/features/templates/routes.tsx)
- [`frontend/apps/web/src/features/templates/pages/TemplatesListRoutePage.tsx:4`](../../../frontend/apps/web/src/features/templates/pages/TemplatesListRoutePage.tsx) → [`TemplatesListPage.tsx`](../../../frontend/apps/web/src/features/templates/TemplatesListPage.tsx).
- [`frontend/apps/web/src/features/templates/pages/TemplateWizardPage.tsx:44`](../../../frontend/apps/web/src/features/templates/pages/TemplateWizardPage.tsx) — create-new wizard (shared `WizardShell`).
- [`frontend/apps/web/src/features/templates/pages/TemplateEditorRoutePage.tsx`](../../../frontend/apps/web/src/features/templates/pages/TemplateEditorRoutePage.tsx) → [`TemplateEditorPage.tsx:53`](../../../frontend/apps/web/src/features/templates/pages/TemplateEditorPage.tsx).
- [`frontend/apps/web/src/features/templates/api/templates.ts`](../../../frontend/apps/web/src/features/templates/api/templates.ts) — full template surface (`createTemplate`:118, `listTemplates`:144, `getTemplate`:176, `getVersion`:182, `presignAutosave`:188, `commitAutosave`:201, `importTemplateDocx`:220, `saveDraft`:255, `publishVersion`:269, `submitForReview`:299, `reviewVersion`:307, `approveVersion`:336, `getTemplateSchemas`:414, `StaleLockVersionError`:406).
- [`frontend/apps/web/src/features/templates/queries/useTemplatesQuery.ts:5`](../../../frontend/apps/web/src/features/templates/queries/useTemplatesQuery.ts)
- [`frontend/apps/web/src/features/templates/TemplateOutlinePanel.tsx`](../../../frontend/apps/web/src/features/templates/TemplateOutlinePanel.tsx), [`PlaceholderCatalogPanel.tsx`](../../../frontend/apps/web/src/features/templates/PlaceholderCatalogPanel.tsx), [`VersionActionPanel.tsx`](../../../frontend/apps/web/src/features/templates/VersionActionPanel.tsx).

## 3. Routes

| Path | Component | Handle |
|---|---|---|
| `/templates` | `TemplatesListRoutePage` | `workspaceView: 'templates'` |
| `/templates/new` | `TemplateWizardPage` | `workspaceView: 'templates'` |
| `/templates/:templateId/versions/:versionNum` | `TemplateEditorRoutePage` | `workspaceView: 'templates', editMode: true` |

## 4. TanStack Query

| Key | Source | Owner hook |
|---|---|---|
| `QK.templates.list()` | `lib/queryKeys.ts:55` | `useTemplatesQuery` |
| `QK.templates.blank()` | `lib/queryKeys.ts:56` | `useBlankTemplateQuery` (documents wizard) |
| `QK.templates.byProfile(code)` | `lib/queryKeys.ts:57` | `useTemplatesByProfileQuery` |

**Invalidation:** publish / approve / submit-for-review mutations must invalidate `QK.templates.list()` and `QK.templates.byProfile(profileCode)`. Approve response returns `next_draft` (see [`templates.ts:326–336`](../../../frontend/apps/web/src/features/templates/api/templates.ts)) so the editor navigates `vN → vN+1` without a list refetch.

## 5. API endpoints consumed

Backend routes under `/api/v1/templates` and `/api/v1/templates/{id}/versions`. Notable:

| FE call | Backend route |
|---|---|
| `createTemplate` | `POST /api/v1/templates` |
| `listTemplates` | `GET /api/v1/templates` |
| `getTemplate` | `GET /api/v1/templates/{id}` |
| `getVersion` | `GET /api/v1/templates/{id}/versions/{n}` |
| `presignAutosave` + `commitAutosave` | autosave direct-to-MinIO |
| `importTemplateDocx`, `presignDocxUpload`, `presignSchemaUpload` | initial docx ingestion |
| `saveDraft` | `PUT /api/v1/templates/{id}/versions/{n}` |
| `submitForReview` | `POST .../submit` |
| `reviewVersion` | `POST .../review` |
| `approveVersion` | `POST .../approve` — returns `next_draft` |
| `publishVersion` | `POST .../publish` |
| `getDocxURL` | resolves storage URL |
| `getTemplateSchemas` | placeholder catalog |

## 6. Dependencies

**Imports from:** `features/shared/components/wizard/` (WizardShell/WizardFooter), `features/shared/components/editor-chrome/`, `editor-adapters/` (eigenpal ACL), `lib/api/`, `lib/api-types/`.

**Imported by:** `features/documents/` — blank/by-profile template queries feed the new-document wizard.

## 7. Invariants

- Server state via TanStack Query keyed by `QK.templates.*`.
- Optimistic concurrency on version edits — `StaleLockVersionError` (`templates.ts:406`) surfaces 409 inline; no silent overwrite.
- Approve gates on `template.approve` capability (backend). FE button visibility derives from server state, never role string.
- Autosave uses presigned-PUT to MinIO before `commitAutosave`.
- Editor wizard chrome is the shared `WizardShell` — do not fork.

## 8. Known issues / tech-debt

- See backend [`templates-tech-debt.md`](../templates-tech-debt.md).
- Placeholder catalog migration tracked in [`wiki/concepts/placeholders.md`](../../concepts/placeholders.md) — fixed 7-token catalog gap between eigenpal native and MetalDocs legacy.
- Recent fixes on this branch: PR #36 friendly errors + idempotency, PR #41 next_draft return for approve, PR #42 `canPublish` gating on `template.approve`.

## 9. Failure modes

| Failure | Symptom | Detection | Response |
|---|---|---|---|
| Schema PUT 412 `stale_lock_version` (multi-tab edit) | Editor surfaces conflict inline; schema not saved | `templates.ts:406` throws `StaleLockVersionError`; `useTemplateSchemas` surfaces `staleConflict` flag | User refetches via the prompted action; resolves divergence then retries |
| 403 on `publishVersion` / `approveVersion` (missing `template.publish` / `template.approve`) | Action button hidden; if reached, backend returns 403 with role-binding code | FE button visibility derives from server `canPublish` / `canApprove` (per backend Tier-2 binding) | Operator escalation; never bypass FE gate |
| MinIO presigned-PUT fails during template autosave | Editor autosave indicator turns red; `commitAutosave` not invoked | `presignAutosave` returns OK but PUT to MinIO rejects | User retries; MinIO healthcheck / CORS audit |
| `importTemplateDocx` 4xx (invalid docx) | Wizard surfaces friendly error from `resolveErrorMessage` | `ApiError.code` from backend templates module | User fixes source docx; PR #36 wired friendly errors here |
| Idempotency replay on submit/review/approve | 200 with prior body; lifecycle state unchanged | Backend `Idempotency-Key` replay | Expected on network retry; no double-transition |
| Approve response missing `next_draft` | Editor stays on vN instead of jumping to vN+1 | `approveVersion` response lacks `nextDraft` field | Backend regression — see PR #41 / `templates.ts:326–336`; refetch list and navigate manually |

## 10. Cross-links

- Backend module: [`wiki/modules/templates.md`](../templates.md)
- Editor ACL: [`wiki/modules/editor-ui-eigenpal.md`](../editor-ui-eigenpal.md), [`wiki/modules/editor-chrome.md`](../editor-chrome.md)
- Concept: [`wiki/concepts/placeholders.md`](../../concepts/placeholders.md)
- Skill: [`metaldocs-frontend`](../../../.agents/skills/metaldocs-frontend/SKILL.md)
