# Frontend module: approval

> **Last verified:** 2026-08-12 (workflow-routing refresh; Unit 4.2 subject-generic inbox and prior F5.1 cockpit content unchanged)
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
- [`frontend/apps/web/src/features/approval/components/InboxStack.tsx`](../../../frontend/apps/web/src/features/approval/components/InboxStack.tsx), [`InboxApprovalCard.tsx`](../../../frontend/apps/web/src/features/approval/components/InboxApprovalCard.tsx), [`InboxTimeline.tsx`](../../../frontend/apps/web/src/features/approval/components/InboxTimeline.tsx), [`ApprovalTimelinePanel.tsx`](../../../frontend/apps/web/src/features/approval/components/ApprovalTimelinePanel.tsx).
- [`frontend/apps/web/src/features/approval/pages/DetailheSignoffPage.tsx`](../../../frontend/apps/web/src/features/approval/pages/DetailheSignoffPage.tsx) — Detalhe Signoff cockpit (F5.1). Mounts `MetalDocsEditor` in review/suggesting mode for inline redline viewing.

## 2.1 Detalhe Signoff Cockpit (F5.1)

The cockpit (`/approvals/:documentId`) unifies review, redline, and signoff in a single page:

- **Document mount:** `MetalDocsEditor` renders the under-review DOCX in review mode, mapping to Eigenpal suggesting mode.
- **Sidebar:** Timeline of prior approvals, comments, and decision buttons.
- **Save-on-decision flow:** on approve/reject, `SignoffDialog` collects the decision + optional password reauth; the cockpit then saves the review buffer and commits signoff metadata in the backend flow.
- **Tracked changes:** inline redlines live in the DOCX and are rendered by Eigenpal; there is no fabricated two-buffer diff.

## 3. Routes

| Path | Component | Handle | Auth |
|---|---|---|---|
| `/approvals` | `InboxPage` | `workspaceView: 'approvals'` | tier-1 session |
| `/approval-routes` | `RouteAdminPage` | `requiresAdmin: true` | client UX gate + backend `route.admin` capability |

## 4. TanStack Query

| Key | Source | Notes |
|---|---|---|
| `QK.inbox(params)` | `lib/queryKeys.ts:36` | Used by `useInboxQuery` and dashboard inbox query. |
| `QK.approval.instance(documentId)` | `lib/queryKeys.ts:61` | Read by approval/document surfaces; invalidated after relevant mutations. |

Signoff/publish/cancel mutations should invalidate `QK.inbox()`, `QK.approval.instance(documentId)`, and `QK.documents.detail(documentId)` as appropriate.

## 5. API endpoints consumed

Frontend wrappers live in `approvalApi.ts` and target approval/document routes including:

- `GET /api/v1/approval/inbox`
- `GET /api/v1/documents/{id}/approval-instance`
- `POST /api/v1/documents/{id}/submit`
- `POST /api/v1/documents/{id}/signoff`
- publication/scheduling/lifecycle routes
- `/api/v1/approval/routes[/{id}]`

The backend approval module is authoritative for these workflows.

## 6. Dependencies

**Imports from:** generated API types, `lib/api/`, shared error handling, UI store/surfaces.

**Imported by:** dashboard inbox surfaces and document editor/published views.

## 7. Invariants

- Server state goes through TanStack Query and centralized `QK` keys.
- Mutating endpoints use the established mutation/ETag path where required.
- Signoff uses the idempotency and optimistic-concurrency contract defined by the backend/API surface.
- Client route gating is defense-in-depth/UX only; backend capability authorization is authoritative.
- Subject-generic inbox rows must preserve document-vs-template behavior rather than assuming every approval subject is a document.

## 8. Known issues / tech-debt

See [`approval-tech-debt.md`](../approval-tech-debt.md) and current roadmap/backlog entries. Where OpenAPI/runtime/generated surfaces disagree, treat it as a shared contract prerequisite rather than adding another handwritten source of truth.

## 9. Failure modes

| Failure | Symptom | Detection | Response |
|---|---|---|---|
| Backend failure on inbox query | Inbox error state/toast | query error / network response | Retry; inspect backend approval path if persistent |
| 401 during inbox/signoff | auth redirect | auth bus | Re-authenticate; session/query state is cleared |
| Stale `If-Match` | signoff conflict | mutation response | Refresh detail/approval state and retry against fresh revision |
| Unresolved comments | business conflict | backend problem response | Resolve comments then retry |
| Idempotency replay | success with replay semantics | response contract | Treat as intended retry behavior, not duplicate mutation |
| Inbox snapshot drift | paged totals disagree | pagination/count behavior | Refresh; tracked as backend persistence/read-model debt |

## 10. Cross-links

- Backend module: [`wiki/modules/approval.md`](../approval.md)
- Sequence: [`wiki/diagrams/sequence-signoff-freeze.md`](../../diagrams/sequence-signoff-freeze.md)
- Concepts: [`wiki/concepts/authz-tiers.md`](../../concepts/authz-tiers.md), [`wiki/concepts/error-ux.md`](../../concepts/error-ux.md)
- Workflow: [`AGENTS.md`](../../../AGENTS.md) + [`wiki/architecture/frontend-structure.md`](../../architecture/frontend-structure.md); use generated API types and the canonical engineering method for non-trivial changes.
