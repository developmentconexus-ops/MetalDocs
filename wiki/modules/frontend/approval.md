# Frontend module: approval

> **Last verified:** 2026-07-14 (Unit 4.2 slice 2: the inbox is now subject-generic and shows TEMPLATE reviews alongside document reviews. `InboxItem` (`lib/api-types/`) dropped `document_id`/`document_title`, replaced by `subject_kind`/`subject_key`/`subject_title`/`subject_ref` + required-and-nullable `controlled_document_id`. `InboxPage.openDocument` (`pages/InboxPage.tsx:99`) branches nav on `subject_kind`: document rows go to `/documents/:subject_ref`, template rows to `/templates/:subject_ref/approval`. `InboxPage.openDecisionFlow` (`:112`) is document-only — gated on `subject_kind === 'document' && controlled_document_id != null`; template rows never enter the quick-decision flow. `InboxApprovalCard.tsx` and `InboxStack.tsx` render `subject_title`/`subject_ref` and hide document-only fields/actions (approve/reject/"Abrir documento") for template rows, showing a single "Abrir revisão do modelo" action instead. See backend [`approval.md`](../approval.md) §header/§6.3 for the read-path detail. Prior: 2026-06-23 (F5.1 cockpit closed: editor mount + save-on-decision flow; review redline via @eigenpal/docx-editor-react@1.9.0 suggesting mode; prior: 2026-06-02)
> **Scope:** Inbox (caixa-de-aprovação), signoff dialog, approval-route admin, detalhe signoff cockpit. Frontend slice of the backend [`approval`](../approval.md) module.
> **Owner:** unassigned | **Backend counterpart:** [`wiki/modules/approval.md`](../approval.md)

## 1. Purpose

Renders the approver inbox, drives the signoff flow with `If-Match` revision precondition, and offers a route-admin surface for `route.admin` capability holders.

## 2. Key files

- [`frontend/apps/web/src/features/approval/routes.tsx:1`](../../../frontend/apps/web/src/features/approval/routes.tsx) — route table.
- [`frontend/apps/web/src/features/approval/pages/InboxPage.tsx:23`](../../../frontend/apps/web/src/features/approval/pages/InboxPage.tsx) — `InboxPage` route entry.
- [`frontend/apps/web/src/features/approval/pages/route-admin/RouteAdminPage.tsx:17`](../../../frontend/apps/web/src/features/approval/pages/route-admin/RouteAdminPage.tsx) — `RouteAdminPage` (admin only; PR-4 canonical rewrite split from monolithic file).
- [`frontend/apps/web/src/features/approval/queries/useInboxQuery.ts:6`](../../../frontend/apps/web/src/features/approval/queries/useInboxQuery.ts) — `useInboxQuery`.
- [`frontend/apps/web/src/features/approval/api/approvalApi.ts:43`](../../../frontend/apps/web/src/features/approval/api/approvalApi.ts) — `getInstance`, `listInbox` (line 53), `submit` (74), `signoff` (85), `publish` (96), `schedulePublish` (107), `supersede` (118), `obsolete` (129), `cancel` (140), route CRUD (151–169).
- [`frontend/apps/web/src/features/approval/api/mutationClient.ts:31`](../../../frontend/apps/web/src/features/approval/api/mutationClient.ts) — `If-Match` + ETag wrapper.
- [`frontend/apps/web/src/features/approval/api/etagCache.ts:3`](../../../frontend/apps/web/src/features/approval/api/etagCache.ts) — per-resource ETag cache.
- [`frontend/apps/web/src/features/approval/components/SignoffDialog.tsx`](../../../frontend/apps/web/src/features/approval/components/SignoffDialog.tsx) — approve/reject with password reauth.
- [`frontend/apps/web/src/features/approval/components/InboxStack.tsx`](../../../frontend/apps/web/src/features/approval/components/InboxStack.tsx), [`InboxApprovalCard.tsx`](../../../frontend/apps/web/src/features/approval/components/InboxApprovalCard.tsx) (Unit 4.2: renders `subject_title`/`subject_ref`, gates document-only fields/actions on `subject_kind`), [`InboxTimeline.tsx`](../../../frontend/apps/web/src/features/approval/components/InboxTimeline.tsx), [`ApprovalTimelinePanel.tsx`](../../../frontend/apps/web/src/features/approval/components/ApprovalTimelinePanel.tsx).
- [`frontend/apps/web/src/features/approval/pages/DetailheSignoffPage.tsx`](../../../frontend/apps/web/src/features/approval/pages/DetailheSignoffPage.tsx) — Detalhe Signoff cockpit (F5.1, 2026-06-23). Mounts `MetalDocsEditor` in `review`-mode (Eigenpal `suggesting`) for inline redline viewing. On user decision (approve/reject via `SignoffDialog`), saves the document buffer and commits metadata in a single transaction.

## 2.1 Detalhe Signoff Cockpit (F5.1)

The cockpit (`/approvals/:documentId`) unifies review, redline, and signoff in a single page:

- **Document mount:** `MetalDocsEditor` (from `packages/editor-ui/`) renders the under-review DOCX in `review` mode, which maps to Eigenpal's `suggesting` mode. This shows tracked changes (additions, deletions, formatting shifts) as inline redlines, replacing the prior two-tab diff design. The document is read-only to the reviewer.
- **Sidebar:** Timeline of prior approvals, comments (unresolved counted), and decision buttons.
- **Save-on-decision flow:** When reviewer clicks approve or reject, `SignoffDialog` collects the decision + optional password reauth, then the cockpit saves the reviewer's own edits (if any) via the captured `documentBuffer` from the editor, and commits the signoff metadata in a single backend transaction. This enforces "buffer state at signoff time must match what the reviewer saw."
- **Tracked changes:** Inline redlines live in the document itself (sourced from Eigenpal). A future enhancement (deferred item in `wiki/backlog/detalhe-signoff.md`) would extract and summarize them via Eigenpal's `extractTrackedChanges` API.

**Source of truth for redlines:** When a document enters review, the backend (documents module) returns DOCX with tracked changes embedded. Eigenpal renders these as visible redlines in suggesting mode. No fabricated two-buffer diff; all changes come from the DOCX itself.

## 3. Routes

| Path | Component | Handle | Auth |
|---|---|---|---|
| `/approvals` | `InboxPage` (routes.tsx:5) | `workspaceView: 'approvals'` | tier-1 session |
| `/approval-routes` | `RouteAdminPage` (routes.tsx:10) | `requiresAdmin: true` | `system_admin` role (UX gate) + backend `route.admin` |

## 4. TanStack Query

| Key | Source | Notes |
|---|---|---|
| `QK.inbox(params)` | `lib/queryKeys.ts:36` | Used by `useInboxQuery` and `useDashboardInboxQuery` (`features/dashboard/queries/useDashboardInboxQuery.ts:7`). Params: `page`, `areaFilter`, `onlyOverdue`, `limit`. |
| `QK.approval.instance(documentId)` | `lib/queryKeys.ts:61` | Read by `useApprovalInstanceQuery` (`features/documents/queries/useApprovalInstanceQuery.ts:5`); editor invalidates after submit/signoff. |

**Invalidation:** signoff/publish/cancel mutations should invalidate `QK.inbox()`, `QK.approval.instance(documentId)`, and `QK.documents.detail(documentId)`. Editor performs this on submit success (`DocumentEditorPage.tsx:288–300`).

## 5. API endpoints consumed

Backend owner: `internal/modules/approval/http/router.go` (now the top-level 15th module, [ADR 0082](../../decisions/0082-approval-kernel-extraction.md) — see backend [`approval.md` §5.3](../approval.md)). Frontend wrappers in `approvalApi.ts`:

| FE call | Backend route |
|---|---|
| `listInbox` | `GET /api/v1/approval/inbox` |
| `getInstance` | `GET /api/v1/documents/{id}/approval-instance` |
| `submit` | `POST /api/v1/documents/{id}/submit` |
| `signoff` | `POST /api/v1/documents/{id}/signoff` (with `Idempotency-Key`, `If-Match`) |
| `publish` / `schedulePublish` / `supersede` / `obsolete` / `cancel` | doc-scoped POST routes |
| `listRoutes` / `createRoute` / `updateRoute` / `deactivateRoute` | `/api/v1/approval/routes[/{id}]` |

## 6. Dependencies

**Imports from:**
- `lib/api/` — `apiFetch`, `ApiError`.
- `lib/api-types/` — generated `ApprovalInstance`, `InboxItem`, `ActiveDocumentResponse`.
- `store/ui.store` — toast/error surface (via `features/shared/errors`).

**Imported by:**
- `features/dashboard/` — `useDashboardInboxQuery` reuses `QK.inbox` with `limit: 6`.
- `features/documents/pages/DocumentEditorPage.tsx` — invalidates `QK.approval.instance` after submit.
- `features/documents/pages/DocumentPublishedPage.tsx` — reads approval-instance for settled metadata.

## 7. Invariants

- All server state through TanStack Query keyed via `QK`. No inline string arrays.
- Mutating endpoints go through `mutationClient` so `If-Match` is set from `etagCache` automatically (resource id = revision id).
- Signoff requires `Idempotency-Key` header; replay surfaces `was_replay=true` on the response (handled inside `signoff` in `approvalApi.ts:85`).
- `/approval-routes` is gated client-side via `handle.requiresAdmin`; backend `route.admin` capability is authoritative.

## 8. Known issues / tech-debt

- Backend doc-scoped signoff/cancel routes are not yet in OpenAPI — FE hand-rolls types (see backend tech-debt T-002 in [`approval-tech-debt.md`](../approval-tech-debt.md)).
- Inbox uses two-query LIMIT/OFFSET + COUNT with snapshot drift — backend T-005.
- Frontend backlog: [`wiki/backlog/caixa-aprovacao.md`](../../backlog/caixa-aprovacao.md).

## 9. Failure modes

| Failure | Symptom | Detection | Response |
|---|---|---|---|
| Backend 5xx on `listInbox` | Inbox empty state with toast `Falha ao carregar caixa de aprovação` | `useInboxQuery.error instanceof ApiError`; browser network tab shows 500 on `GET /api/v1/approval/inbox` | Retry via `refetch()`; if persistent, check backend `approval` logs and `metaldocs-api` health |
| 401 on any inbox/signoff call | Forced redirect to `/login?returnTo=/approvals` | `authBus` 401 listener in `lib/api/` fires before component sees error | User re-authenticates; cached queries cleared on next login (`queryClient.clear()` from auth slice) |
| Stale `If-Match` on signoff (concurrent edit) | Signoff dialog shows 409 conflict; row not signed | `mutationClient.ts:31` strips ETag from `etagCache` on 412/409; surfaces `state.stale_revision` from backend | Invalidate `QK.approval.instance(documentId)` + `QK.documents.detail(documentId)`; reopen dialog with fresh ETag |
| Unresolved comments block final approval | 409 `approval.unresolved_comments` shown inline in `SignoffDialog` | Backend conflict mapped by `SignoffDialog.tsx` business-conflict branch | Resolve outstanding comments on the document then retry signoff |
| Idempotency replay (network retry) | 200 with `was_replay=true`; row not double-inserted | `approvalApi.ts:85` propagates `was_replay` flag; no client error | None — replay is intended behavior; UI shows success without double-toast |
| Inbox snapshot drift (T-005) | List page total disagrees with count on adjacent page boundary | Manual: compare `total` vs paged item count | Page refresh; backend tech-debt T-005 tracks the two-query LIMIT/OFFSET+COUNT race |

## 10. Cross-links

- Backend module: [`wiki/modules/approval.md`](../approval.md)
- Sequence: [`wiki/diagrams/sequence-signoff-freeze.md`](../../diagrams/sequence-signoff-freeze.md)
- Concept: [`wiki/concepts/authz-tiers.md`](../../concepts/authz-tiers.md), [`wiki/concepts/error-ux.md`](../../concepts/error-ux.md)
- Skill: [`metaldocs-frontend`](../../../.agents/skills/metaldocs-frontend/SKILL.md), [`metaldocs-tanstack-query`](../../../.agents/skills/metaldocs-tanstack-query/SKILL.md)
