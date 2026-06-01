# Frontend module: iam

> **Last verified:** 2026-06-01 (P2 consolidation: added Failure modes section)
> **Scope:** Admin Center (users, roles, locks, password resets) and area-membership administration. Frontend slice of the backend [`iam`](../iam.md) module.
> **Owner:** unassigned | **Backend counterpart:** [`wiki/modules/iam.md`](../iam.md)

## 1. Purpose

`system_admin` operator UI for user management and tenant area membership. Gated by `handle.requiresAdmin` client-side (defense-in-depth); backend tier-1 capability middleware is authoritative.

## 2. Key files

- [`frontend/apps/web/src/features/iam/routes.tsx:1`](../../../frontend/apps/web/src/features/iam/routes.tsx)
- [`frontend/apps/web/src/features/iam/pages/AdminCenterPage.tsx:3`](../../../frontend/apps/web/src/features/iam/pages/AdminCenterPage.tsx) → [`AdminCenterView.tsx`](../../../frontend/apps/web/src/features/iam/AdminCenterView.tsx).
- [`frontend/apps/web/src/features/iam/pages/AreaMembershipAdminRoutePage.tsx:3`](../../../frontend/apps/web/src/features/iam/pages/AreaMembershipAdminRoutePage.tsx) → [`AreaMembershipAdminPage.tsx`](../../../frontend/apps/web/src/features/iam/AreaMembershipAdminPage.tsx).
- [`frontend/apps/web/src/features/iam/MembershipGrantDialog.tsx`](../../../frontend/apps/web/src/features/iam/MembershipGrantDialog.tsx)
- [`frontend/apps/web/src/features/iam/api/iam.ts:52`](../../../frontend/apps/web/src/features/iam/api/iam.ts) — `listUsers` (52), `getAdminOverview` (57), `createUser` (66), `updateUser` (70), `assignRole` (77), `replaceUserRoles` (84), `adminResetPassword` (91), `unlockUser` (98).
- [`frontend/apps/web/src/features/iam/membershipApi.ts:21`](../../../frontend/apps/web/src/features/iam/membershipApi.ts) — `fetchMemberships` (21), `grantMembership` (25), `revokeMembership` (33).
- [`frontend/apps/web/src/features/iam/useAdminCenter.ts:7`](../../../frontend/apps/web/src/features/iam/useAdminCenter.ts), [`useManagedUsers.ts`](../../../frontend/apps/web/src/features/iam/useManagedUsers.ts).

## 3. Routes

| Path | Component | Handle |
|---|---|---|
| `/admin` | `AdminCenterPage` | `workspaceView: 'admin', requiresAdmin: true` |
| `/admin/memberships` | `AreaMembershipAdminRoutePage` | `workspaceView: 'iam-memberships', requiresAdmin: true` |

## 4. TanStack Query

The IAM slice currently uses ad-hoc keys inside `useAdminCenter` / `useManagedUsers` hooks (no central `QK.iam.*` entry yet). Server-state moves through these hooks — server data never lives in `auth.store`.

**Invalidation rules:** role / lock / password mutations should refetch the admin overview and managed-users list. Membership grant/revoke should refetch `fetchMemberships(userId)` for the affected user.

## 5. API endpoints consumed

| FE call | Backend route |
|---|---|
| `getAdminOverview` | `GET /api/v1/admin/overview` |
| `listUsers` | `GET /api/v1/admin/users` |
| `createUser` | `POST /api/v1/admin/users` |
| `updateUser` | `PUT /api/v1/admin/users/{id}` |
| `assignRole` / `replaceUserRoles` | `POST/PUT /api/v1/admin/users/{id}/roles` |
| `adminResetPassword` | `POST /api/v1/admin/users/{id}/reset-password` |
| `unlockUser` | `POST /api/v1/admin/users/{id}/unlock` |
| `fetchMemberships` / `grantMembership` / `revokeMembership` | `/api/v1/admin/users/{userId}/memberships[/{areaCode}]` |

## 6. Dependencies

**Imports from:** `lib/api/`, `store/auth.store` (read-only — current user), `components/ui/` (Avatar with hashed color, StatusPill).

**Imported by:** `features/shell/components/AppShell.tsx` — reads `auth.store.user.roles` to honor `handle.requiresAdmin` redirect; no IAM API imports from outside this feature.

## 7. Invariants

- `handle.requiresAdmin` redirects non-admins from `/admin*` to `/` (UX gate only; backend authoritative).
- All mutations go through `lib/api/apiFetch`; no direct `fetch`.
- A future migration should add `QK.iam.*` to `lib/queryKeys.ts` — current ad-hoc keys are tracked as drift.

## 8. Known issues / tech-debt

- See backend [`iam-tech-debt.md`](../iam-tech-debt.md).
- Centralized `QK.iam.*` keys missing — invalidation depends on hook-local strings.

## 9. Failure modes

| Failure | Symptom | Detection | Response |
|---|---|---|---|
| Non-admin user reaches `/admin*` directly | `handle.requiresAdmin` redirects to `/` (UX gate); backend tier-1 still enforced | `AppShell.tsx` reads `auth.store.user.roles`; backend returns 403 if FE bypassed | Expected behavior; if backend 403 still leaks data, check `iam` capability middleware |
| 403 on `assignRole` / `replaceUserRoles` (operator missing `system_admin`) | Toast with backend `ApiError.message`; mutation fails | `ApiError.code === 'authz.forbidden'` | Operator escalates; no client retry |
| Duplicate `grantMembership` for `(userId, areaCode)` | Backend 409 `iam.membership_exists` | `ApiError.code` | Refetch `fetchMemberships(userId)` to reconcile UI state |
| Backend 5xx on admin overview load | Admin Center shows error state; managed users list empty | `useAdminCenter.error` | Retry; check `metaldocs-api` IAM logs |
| Stale managed-users list after mutation | New user not visible until manual refresh | Ad-hoc query keys not invalidated (tracked in §8) | Manual refetch; migrate to `QK.iam.*` invalidation |
| 401 on any admin call | `authBus` redirect → `/login`; admin context lost | Standard `apiFetch` 401 path | Re-login; returnTo restores `/admin` |

## 10. Cross-links

- Backend module: [`wiki/modules/iam.md`](../iam.md)
- Concept: [`wiki/concepts/authz-tiers.md`](../../concepts/authz-tiers.md)
- Skill: [`metaldocs-frontend`](../../../.agents/skills/metaldocs-frontend/SKILL.md)
