# Frontend module: iam

> **Last verified:** 2026-08-12 (routing-only refresh; A2.2 domain content last verified 2026-08-11)
> **Scope:** Admin Center (Overview, People, Roles & Capabilities, Audit, Sessions & Security, Usage) and tenant area-membership administration. Frontend slice of the backend [`iam`](../iam.md) and [`audit`](../audit.md) modules.
> **Owner:** unassigned | **Backend counterpart:** [`wiki/modules/iam.md`](../iam.md)

## 1. Purpose

Tenant operator UI for IAM. Surfaces user lifecycle, role/capability matrix, audit trail, active sessions + MFA coverage + lockouts + risk signals, and usage/seat metrics. Each tab is gated by a tier-1 capability declared on the route handle (defense-in-depth; backend remains the sole authz enforcer — see [`wiki/concepts/authz-tiers.md`](../../concepts/authz-tiers.md)).

## 2. Key files

- [`frontend/apps/web/src/features/iam/routes.tsx:4`](../../../frontend/apps/web/src/features/iam/routes.tsx) — 6-tab IA with per-tab `requiresCapability` / `requiresAnyCapability`.
- [`frontend/apps/web/src/features/iam/pages/AdminCenterPage.tsx`](../../../frontend/apps/web/src/features/iam/pages/AdminCenterPage.tsx) — shell with header + tablist + `<Outlet />`.
- [`frontend/apps/web/src/features/iam/pages/AreaMembershipAdminRoutePage.tsx`](../../../frontend/apps/web/src/features/iam/pages/AreaMembershipAdminRoutePage.tsx) — legacy membership admin (`/admin/memberships`).
- Tab containers — one `.route.tsx` + one `.tsx` per tab:
  - [`tabs/OverviewTab.route.tsx`](../../../frontend/apps/web/src/features/iam/tabs/OverviewTab.route.tsx) → [`OverviewTab.tsx`](../../../frontend/apps/web/src/features/iam/tabs/OverviewTab.tsx)
  - [`tabs/PeopleTab.route.tsx`](../../../frontend/apps/web/src/features/iam/tabs/PeopleTab.route.tsx) → [`PeopleTab.tsx`](../../../frontend/apps/web/src/features/iam/tabs/PeopleTab.tsx)
  - [`tabs/PeopleDetailDrawer.route.tsx`](../../../frontend/apps/web/src/features/iam/tabs/PeopleDetailDrawer.route.tsx) → [`PeopleDetailDrawer.tsx`](../../../frontend/apps/web/src/features/iam/tabs/PeopleDetailDrawer.tsx)
  - [`tabs/RolesCapsTab.route.tsx`](../../../frontend/apps/web/src/features/iam/tabs/RolesCapsTab.route.tsx) → [`RolesCapsTab.tsx`](../../../frontend/apps/web/src/features/iam/tabs/RolesCapsTab.tsx)
  - [`tabs/AuditTab.route.tsx`](../../../frontend/apps/web/src/features/iam/tabs/AuditTab.route.tsx) → [`AuditTab.tsx`](../../../frontend/apps/web/src/features/iam/tabs/AuditTab.tsx)
  - [`tabs/SessionsTab.route.tsx`](../../../frontend/apps/web/src/features/iam/tabs/SessionsTab.route.tsx) → [`SessionsTab.tsx`](../../../frontend/apps/web/src/features/iam/tabs/SessionsTab.tsx)
  - [`tabs/UsageTab.route.tsx`](../../../frontend/apps/web/src/features/iam/tabs/UsageTab.route.tsx) → [`UsageTab.tsx`](../../../frontend/apps/web/src/features/iam/tabs/UsageTab.tsx)
- Components: [`components/KpiStrip.tsx`](../../../frontend/apps/web/src/features/iam/components/KpiStrip.tsx), [`ActivityPanel.tsx`](../../../frontend/apps/web/src/features/iam/components/ActivityPanel.tsx), [`SessionsTable.tsx`](../../../frontend/apps/web/src/features/iam/components/SessionsTable.tsx), [`UserSessionsTable.tsx`](../../../frontend/apps/web/src/features/iam/components/UserSessionsTable.tsx), [`LockoutsTable.tsx`](../../../frontend/apps/web/src/features/iam/components/LockoutsTable.tsx), [`MfaCoverageCard.tsx`](../../../frontend/apps/web/src/features/iam/components/MfaCoverageCard.tsx), [`SecuritySignalsList.tsx`](../../../frontend/apps/web/src/features/iam/components/SecuritySignalsList.tsx), [`PresenceBadge.tsx`](../../../frontend/apps/web/src/features/iam/components/PresenceBadge.tsx).
- Queries (openapi-fetch + TanStack Query): everything under [`features/iam/queries/`](../../../frontend/apps/web/src/features/iam/queries/) (`useKpiQuery`, `useUsersQuery`, `useUserDetailQuery`, `useRolesQuery`, `useCapabilitiesQuery`, `useRoleCapabilitiesQuery`, `useAuditEventsQuery`, `useSessionsQuery`, `useMfaCoverageQuery`, `useLockoutsQuery`, `useSecuritySignalsQuery`, `useUsageQuery`, `useUserMembershipsQuery`, `usePresenceStream`).
- Mutations: [`features/iam/mutations/`](../../../frontend/apps/web/src/features/iam/mutations/) (`useInviteUserMutation`, `usePatchUserMutation`, `useBulkUsersMutation`, `useResetPasswordMutation`, `useRevokeSessionMutation`, `useUnlockUserMutation`, `useExportAuditMutation`).
- Legacy compat (membership area admin only): [`membershipApi.ts`](../../../frontend/apps/web/src/features/iam/membershipApi.ts) — still uses `apiFetch` for `/api/v1/admin/users/{userId}/memberships`.

## 3. Routes

| Path | Component | Required capability |
|---|---|---|
| `/admin` | `AdminCenterPage` | any of `user.view`, `membership.view`, `metrics.view` |
| `/admin/overview` | `OverviewTab` | any of `user.view`, `membership.view`, `metrics.view` |
| `/admin/people` | `PeopleTab` | `user.view` |
| `/admin/people/:userId` | `PeopleDetailDrawer` | `user.view` |
| `/admin/roles` | `RolesCapsTab` | `membership.view` |
| `/admin/audit` | `AuditTab` | `audit.read` |
| `/admin/sessions` | `SessionsTab` | `user.view` |
| `/admin/usage` | `UsageTab` | `metrics.view` |
| `/admin/memberships` | `AreaMembershipAdminRoutePage` | `membership.view` |

> Note: the `AppShell` gate collects every `requiresCapability` / `requiresAnyCapability` along the match chain and requires all of them to pass. Parent + child constraints both enforced (fixed at PR-12 closeout — QA evidence `QA-evidence-admin-center-rebuild.md` removed at the v1 re-baseline, commit `c7f06f2e`).

## 4. TanStack Query

All keys flow through hooks in `features/iam/queries/`. Each hook returns a typed `useQuery` (`paths` from `lib/api-types/`) and a stable key under namespaces like `["iam","users", …]`, `["iam","sessions", …]`, `["iam","audit", …]`, `["iam","usage", …]`. Invalidation on mutation:

- `useInviteUserMutation` / `usePatchUserMutation` / `useBulkUsersMutation` → `["iam","users"]`, `["iam","kpi"]`.
- `useResetPasswordMutation` → `["iam","users", userId]`.
- `useRevokeSessionMutation` → `["iam","sessions"]`, `["iam","kpi"]`.
- `useUnlockUserMutation` → `["iam","lockouts"]`, `["iam","kpi"]`.
- `useExportAuditMutation` → `["iam","audit"]` (re-fetches timeline so the export event appears).

Presence is a WebSocket subscription (`usePresenceStream`) that writes into a dedicated key family so it can be observed by Overview without polling.

## 5. API endpoints consumed

| FE call | Backend route |
|---|---|
| `useKpiQuery` | `GET /api/v1/admin/iam/kpi` |
| `useUsersQuery` / `useUserDetailQuery` | `GET /api/v1/admin/iam/users[?cursor=…]`, `GET /api/v1/admin/iam/users/{userId}` |
| `useInviteUserMutation` | `POST /api/v1/admin/iam/users` |
| `usePatchUserMutation` | `PATCH /api/v1/admin/iam/users/{userId}` |
| `useBulkUsersMutation` | `POST /api/v1/admin/iam/users/bulk` |
| `useResetPasswordMutation` | `POST /api/v1/admin/iam/users/{userId}/reset-password` |
| `useRolesQuery` / `useCapabilitiesQuery` / `useRoleCapabilitiesQuery` | `GET /api/v1/admin/iam/roles`, `/capabilities`, `/role-capabilities` |
| `useAuditEventsQuery` | `GET /api/v1/admin/audit/events` |
| `useExportAuditMutation` | `POST /api/v1/admin/audit/exports` |
| `useSessionsQuery` | `GET /api/v1/admin/iam/sessions` |
| `useRevokeSessionMutation` | `DELETE /api/v1/admin/iam/sessions/{sessionId}` (bulk via body) |
| `useMfaCoverageQuery` | `GET /api/v1/admin/iam/mfa-coverage` |
| `useLockoutsQuery` / `useUnlockUserMutation` | `GET /api/v1/admin/iam/lockouts`, `POST /api/v1/admin/iam/users/{userId}/unlock` |
| `useSecuritySignalsQuery` | `GET /api/v1/admin/iam/signals` |
| `useUsageQuery` | `GET /api/v1/admin/iam/usage` |
| `usePresenceStream` | `WS /api/v1/admin/iam/presence` (PR-9) |
| `fetchMemberships` / `grantMembership` / `revokeMembership` | `/api/v1/admin/users/{userId}/memberships[/{areaCode}]` (legacy `apiFetch`) |

## 6. Dependencies

**Imports from:** `lib/api` (typed openapi-fetch client `api.GET/POST/PATCH/DELETE`), `lib/api-types/` (generated paths), `store/auth.store` (current user capabilities), `components/ui/` primitives, `lib/queryKeys.ts` (audit + sessions + usage namespaces).

**Imported by:** `features/shell/components/AppShell.tsx` (capability gate read), router root (`routes.tsx` registration).

## 7. Invariants

- Every IAM admin API call uses the typed `api.*` client; only the legacy `membershipApi.ts` still uses `apiFetch`. Direct `fetch()` is banned in this slice.
- Sessions filters live in the URL (`?usuario=…&ip=…&mfa=…`) so the tab is shareable; cache key includes the filter signature.
- Bulk-row-action mutations use a single bulk endpoint, not N parallel single-id calls.
- All KPI cards have a `refetch` action; failed cards render an inline retry, not a full-page error.
- The 6-tab IA is the only public surface — no legacy `AdminCenterView` / `useAdminCenter` / `useManagedUsers` / `state/admin.store.ts` remains (verified at PR-12 closeout).

## 8. Known issues / tech-debt

- **`UserRole` carries `admin` + `reviewer` literals beyond canonical 8** in `lib/types/index.ts:19-29`. Phase 1 left these to avoid touching unrelated call sites (templates/documents/taxonomy/iam-membership). Cleanup tracked for follow-up PR.
- Membership area admin (`/admin/memberships`) still uses `apiFetch` — migration to `api.*` deferred.
- Backend IAM tech-debt remains tracked at [`iam-tech-debt.md`](../iam-tech-debt.md).

## 9. Failure modes

| Failure | Symptom | Detection | Response |
|---|---|---|---|
| User without IAM caps reaches `/admin/*` | First tab cap denied → `<Navigate to="/" />`; tabs without caps hidden from tablist | `AppShell.tsx` reads `auth.store.user.capabilities`; backend returns 403 on data calls | Expected; tighten gate (see §8) |
| 403 on `usePatchUserMutation` / `useBulkUsersMutation` | Inline error banner; row stays selected, no optimistic flicker | `ApiError.code === 'authz.capability_denied'` | Operator lacks `user.manage` |
| 409 on `useInviteUserMutation` (`iam.user_exists`) | Toast with backend message; drawer stays open with field highlighted | `ApiError.code` | Operator chooses different identifier or re-uses existing |
| Bulk revoke partial failure | Per-session error rendered in row; successes invalidate the list | `useRevokeSessionMutation` returns `{succeeded, failed[]}` | Operator retries failed rows |
| Audit export returns 202 | Toast "Exportação solicitada"; new `audit_export.requested` event appears at top of timeline | `useExportAuditMutation` resolves with `exportId` | Operator picks up downloaded file from audit row when status flips to `done` |
| WebSocket presence drops | `Quem está online` shows last known snapshot with reconnect banner | `usePresenceStream` reconnect logic | Auto-retry; manual refresh as fallback |
| `useInfiniteQuery` (Audit timeline) cursor desync | "Carregar mais" returns duplicate or empty page | Cursor pagination contract drift | See ADR [`0028-audit-events-cursor-shape.md`](../../decisions/0028-audit-events-cursor-shape.md) |
| 401 anywhere | `authBus` redirect → `/login`; admin context lost | Standard `apiFetch` / openapi-fetch error path | Re-login; returnTo restores last `/admin/*` tab |

## 10. Cross-links

- Backend module: [`wiki/modules/iam.md`](../iam.md), [`wiki/modules/audit.md`](../audit.md)
- Concept: [`wiki/concepts/authz-tiers.md`](../../concepts/authz-tiers.md)
- ADRs: [`0019-cap-audit-read-and-session-manage.md`](../../decisions/0019-cap-audit-read-and-session-manage.md), [`0020-admin-center-six-tab-ia.md`](../../decisions/0020-admin-center-six-tab-ia.md), [`0021-tenant-vs-platform-admin-separation.md`](../../decisions/0021-tenant-vs-platform-admin-separation.md), [`0028-audit-events-cursor-shape.md`](../../decisions/0028-audit-events-cursor-shape.md)
- Workflow: [`AGENTS.md`](../../../AGENTS.md) + [`wiki/architecture/frontend-structure.md`](../../architecture/frontend-structure.md); use the canonical engineering method for non-trivial changes.
- QA evidence: `QA-evidence-admin-center-rebuild.md` (removed at the v1 re-baseline, commit `c7f06f2e`)
