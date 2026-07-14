# Frontend module: auth

> **Last verified:** 2026-06-01 (P2 consolidation: added Failure modes section)
> **Scope:** Login, logout, change-password, session bootstrap, 401 → login redirect bus. Frontend slice of the backend [`auth`](../auth.md) module.
> **Owner:** unassigned | **Backend counterpart:** [`wiki/modules/auth.md`](../auth.md)

## 1. Purpose

Owns the only public route in the SPA (`/login`) plus the session lifecycle helper consumed by the shell. All TanStack Query state is dropped on logout/expiry via `queryClient.clear()`.

## 2. Key files

- [`frontend/apps/web/src/features/auth/routes.tsx:1`](../../../frontend/apps/web/src/features/auth/routes.tsx) — `/auth` placeholder (the public `/login` route is declared at the root in [`AppRouter.tsx:15`](../../../frontend/apps/web/src/app/AppRouter.tsx)).
- [`frontend/apps/web/src/features/auth/pages/LoginPage.tsx:137`](../../../frontend/apps/web/src/features/auth/pages/LoginPage.tsx) — login form, returnTo-aware `<Navigate>` post-success.
- [`frontend/apps/web/src/features/auth/pages/AuthRoutePage.tsx:3`](../../../frontend/apps/web/src/features/auth/pages/AuthRoutePage.tsx).
- [`frontend/apps/web/src/features/auth/api/auth.ts:34`](../../../frontend/apps/web/src/features/auth/api/auth.ts) — `login` (34), `logout` (42), `me` (46), `changePassword` (50).
- [`frontend/apps/web/src/features/auth/useAuthSession.ts:11`](../../../frontend/apps/web/src/features/auth/useAuthSession.ts) — `handleLogin`, `handleLogout` (clears `queryClient`), `handleChangePassword`.
- [`frontend/apps/web/src/features/shell/pages/AppRoot.tsx:28`](../../../frontend/apps/web/src/features/shell/pages/AppRoot.tsx) — guarded bootstrap; calls `me()` on mount.
- [`frontend/apps/web/src/lib/api/`](../../../frontend/apps/web/src/lib/api/) — `apiFetch`, `ApiError`, `authBus` (401 → redirect).
- [`frontend/apps/web/src/store/auth.store.ts`](../../../frontend/apps/web/src/store/auth.store.ts) — global zustand: `user`, `authState`, form drafts.

## 3. Routes

| Path | Component | Notes |
|---|---|---|
| `/login` | `LoginPage` (declared in `AppRouter.tsx:15`) | Public — outside `AppRoot` guard. No Rail, no Toolbar. |
| `/auth` | `AuthRoutePage` | Protected helper page. |

## 4. TanStack Query

No domain queries. Auth state lives in `store/auth.store.ts`. On logout / 401, `queryClient.clear()` is invoked from `useAuthSession.handleLogout` (`useAuthSession.ts:47`) so every cached `QK.*` entry is dropped.

## 5. API endpoints consumed

| FE call | Backend route |
|---|---|
| `login` | `POST /api/v1/auth/login` |
| `logout` | `POST /api/v1/auth/logout` |
| `me` | `GET /api/v1/auth/me` |
| `changePassword` | `POST /api/v1/auth/change-password` |

## 6. Dependencies

**Imports from:** `lib/api/`, `store/auth.store`, `store/ui.store` (error/message), `features/shared/errors` (`asMessage`, `codeOf`).

**Imported by:**
- `features/shell/pages/AppRoot.tsx` — bootstrap + guard.
- `features/shell/components/AppToolbar.tsx` — user menu logout.
- `features/iam/` — reads `auth.store.user.roles` for admin gates.

## 7. Invariants

- Public route surface is exactly `/login`. Nothing else is rendered outside the auth guard.
- `useAuthSession.handleLogout` must call `queryClient.clear()` — protects against PII leak across users.
- 401 on any `apiFetch` triggers the auth bus → redirect to `/login?returnTo=...` (no inline retries).
- `auth.store` holds session metadata only — no document or domain data.

## 8. Known issues / tech-debt

- See backend [`auth-tech-debt.md`](../auth-tech-debt.md).
- No external SSO/IdP — see [`ONBOARDING.md` §7](../../ONBOARDING.md).

## 9. Failure modes

| Failure | Symptom | Detection | Response |
|---|---|---|---|
| Backend `/auth/login` 401 (bad credentials / locked) | Login form shows error message from `asMessage(err)` | `ApiError.code` carries `auth.invalid_credentials` or `auth.account_locked` | User retries; locked accounts require admin unlock via IAM module |
| Network error on `me()` during bootstrap | `AppRoot` shows blocking error state; SPA does not render past guard | `useAuthSession` surface; `apiFetch` rejects with `Error` (no `ApiError.code`) | User refreshes; if persistent, backend `metaldocs-api` is down — check `/healthz` |
| 401 on any post-login `apiFetch` (session expired) | `authBus` 401 listener fires → `<Navigate to="/login?returnTo=...">`; `queryClient.clear()` runs | Any `apiFetch` rejection with `status === 401` | User re-authenticates; returnTo path preserved across the redirect |
| Logout race (multiple tabs) | One tab clears `queryClient` while other holds in-flight mutation | Mutation completes against expired session → 401 → authBus → both tabs redirect | Expected behavior; no recovery needed |
| Change-password rate-limit (backend `auth.too_many_attempts`) | Form shows backend message; submit disabled | `ApiError.code === 'auth.too_many_attempts'` | Wait for backend cooldown window (see backend `auth.md` §10) |

## 10. Cross-links

- Backend module: [`wiki/modules/auth.md`](../auth.md)
- Concept: [`wiki/concepts/error-ux.md`](../../concepts/error-ux.md) — apiFetch / authBus contract.
- Skill: [`metaldocs-frontend`](../../../.agents/skills/metaldocs-frontend/SKILL.md)
