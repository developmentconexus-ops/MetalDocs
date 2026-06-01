# Frontend module: controlled-documents

> **Last verified:** 2026-06-01 (P2 consolidation: added Failure modes section)
> **Scope:** Read-side query/API surface for controlled documents (the regulated wrapper around a document family). No standalone pages — consumed by `documents` and `approval` slices.
> **Owner:** unassigned | **Backend counterpart:** [`wiki/modules/controlled-documents.md`](../controlled-documents.md)

## 1. Purpose

Exposes typed accessors for the controlled-document aggregate (profile + area + manual code + active revision) so other features can read/preview/create without duplicating types. This is a "library" feature — no `routes.tsx`, no pages.

## 2. Key files

- [`frontend/apps/web/src/features/controlled-documents/types.ts:1`](../../../frontend/apps/web/src/features/controlled-documents/types.ts) — `ControlledDocument`, `CreateControlledDocumentRequest`.
- [`frontend/apps/web/src/features/controlled-documents/api/controlledDocuments.ts:7`](../../../frontend/apps/web/src/features/controlled-documents/api/controlledDocuments.ts) — `fetchControlledDocuments` (7), `fetchControlledDocument` (25), `createControlledDocumentAtomic` (32), `createRevision` (47), `previewCode` (65), `obsoleteControlledDocument` (73), `supersedeControlledDocument` (80), `fetchActiveDocumentInstance` (89).
- [`frontend/apps/web/src/features/controlled-documents/api/catalog.ts`](../../../frontend/apps/web/src/features/controlled-documents/api/catalog.ts).
- [`frontend/apps/web/src/features/controlled-documents/queries/useControlledDocumentDetailQuery.ts:6`](../../../frontend/apps/web/src/features/controlled-documents/queries/useControlledDocumentDetailQuery.ts)
- [`frontend/apps/web/src/features/controlled-documents/queries/usePreviewCodeQuery.ts:11`](../../../frontend/apps/web/src/features/controlled-documents/queries/usePreviewCodeQuery.ts)

## 3. Routes

None. The slice is queries + API only.

## 4. TanStack Query

| Key | Source | Owner hook |
|---|---|---|
| `QK.controlledDocuments.list(filter)` | `lib/queryKeys.ts:43` | direct callers |
| `QK.controlledDocuments.detail(id)` | `lib/queryKeys.ts:45` | `useControlledDocumentDetailQuery` |
| `QK.controlledDocuments.activeDocument(id)` | `lib/queryKeys.ts:46` | `useControlledDocumentActiveDocumentQuery` (documents slice) |
| `QK.controlledDocuments.preview(profileCode, areaCode)` | `lib/queryKeys.ts:47` | `usePreviewCodeQuery` |

**Invalidation:** new-document atomic-create invalidates `QK.controlledDocuments.list()`; obsolete/supersede mutations should invalidate `detail(id)` and the relevant `list()` filter. The wizard switches preview keys on profile/area change (`NewDocumentWizardPage.tsx:175`).

## 5. API endpoints consumed

| FE call | Backend route |
|---|---|
| `fetchControlledDocuments` | `GET /api/v1/controlled-documents` |
| `fetchControlledDocument` | `GET /api/v1/controlled-documents/{id}` |
| `createControlledDocumentAtomic` | `POST /api/v1/controlled-documents/atomic` |
| `createRevision` | `POST /api/v1/controlled-documents/{id}/revisions` |
| `previewCode` | `GET /api/v1/controlled-documents/preview-code` |
| `obsoleteControlledDocument` | `POST /api/v1/controlled-documents/{id}/obsolete` |
| `supersedeControlledDocument` | `POST /api/v1/controlled-documents/{id}/supersede` |
| `fetchActiveDocumentInstance` | `GET /api/v1/controlled-documents/{id}/active-document` |

## 6. Dependencies

**Imports from:** `lib/api/`, `lib/api-types/` (generated `ControlledDocument`, `ActiveDocumentResponse`, `CreateAtomicRequest`).

**Imported by:**
- `features/documents/pages/NewDocumentWizardPage.tsx` — atomic create + preview code.
- `features/documents/pages/DocumentPublishedPage.tsx` — recovers publish context via `active-document` lookup.
- `features/approval/api/approvalApi.ts:181` — `getActiveDocumentContext`.

## 7. Invariants

- All shapes come from `lib/api-types/` — no hand-written request/response types (see `types.ts:1`).
- Atomic create is the only supported create path (see backend module). FE must not chain two-step create + revision.
- Preview-code query is `enabled` only when both `profileCode` and `areaCode` are present.

## 8. Known issues / tech-debt

- See backend [`controlled-documents-tech-debt.md`](../controlled-documents-tech-debt.md).
- No dedicated UI surface yet — controlled-document detail is rendered through the published-page / editor layouts.

## 9. Failure modes

| Failure | Symptom | Detection | Response |
|---|---|---|---|
| Backend 5xx on `fetchControlledDocument(id)` | Document detail / published view fails to load; consumer page shows fallback | `useControlledDocumentDetailQuery.error` | Retry via `refetch()`; check backend `controlled-documents` logs |
| 409 sequence collision on `createControlledDocumentAtomic` | Wizard submit fails with `cd.sequence_collision` (rare; retried server-side) | `ApiError.code === 'cd.sequence_collision'` | Generate new `Idempotency-Key` and retry; backend ADR 0011 covers atomic create rollback |
| Idempotency replay (network retry on atomic create) | 201 with prior body; no duplicate CD row | `createControlledDocumentAtomic` returns stored response | Expected — UI continues to success step |
| Preview-code query disabled when profile/area incomplete | Code preview shows `???` placeholder | `usePreviewCodeQuery.enabled === false` | Expected; wizard gates advance until both fields set |
| Stale list cache after obsolete/supersede mutation | Library list still shows obsoleted CD | Mutation `onSuccess` did not invalidate `QK.controlledDocuments.list()` | Add invalidation; tracked alongside backend `controlled-documents-tech-debt` |

## 10. Cross-links

- Backend module: [`wiki/modules/controlled-documents.md`](../controlled-documents.md)
- Decision: [`wiki/decisions/`](../../decisions/) — atomic CD create ADR.
- Skill: [`metaldocs-frontend`](../../../.agents/skills/metaldocs-frontend/SKILL.md), [`metaldocs-tanstack-query`](../../../.agents/skills/metaldocs-tanstack-query/SKILL.md)
