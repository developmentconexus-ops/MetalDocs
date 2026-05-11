# Phase 0 — Context (auth)

**Date:** 2026-05-10
**Module path:** `internal/modules/auth/`
**Existing wiki stub:** none (only `wiki/references/local-dev-credentials.md` mentions auth credentials).

## Sources read

- `wiki/README.md` — index; auth not yet a module doc.
- `wiki/decisions/0007-two-tier-authz.md` — auth excluded from scope ("Out of scope: Authentication; … see `wiki/modules/iam.md`").
- `wiki/concepts/authz-tiers.md` — same exclusion ("Out of scope: Authentication (login/sessions)").
- `wiki/modules/iam.md` — bidirectional dep called out at §3.2 / §8.4: `auth` imports `iamdomain.Role`, IAM `AdminHandler` consumes `auth.domain.ManagedUser`.
- `wiki/modules/documents.md` §8.1 — consumer of auth context (`authdomain.CurrentUser`) via middleware.
- `wiki/references/local-dev-credentials.md` — admin user `admin` / `AdminMetalDocs123!`.

## Module shape (pre-scan)

- `domain/` — `Identity`, `Session`, `ManagedUser`, `CurrentUser`, `AuthenticatedSession`, `Repository` port; sentinel errors; `WithCurrentUser` context plumbing.
- `application/service.go` — single `Service` exposing: `BootstrapLocalAdmin`, `Authenticate`, `ResolveSession`, `Logout`, `ChangePassword(*)`, `ListUsers`, `ListOnlineUsers`, `CreateUser`, `UpdateUser`, `AdminResetPassword`, `UnlockUser`, `CurrentUser`, cookie helpers, HMAC token signing.
- `delivery/http/handler.go` — 4 routes: `POST /api/v1/auth/login`, `POST /api/v1/auth/logout`, `GET /api/v1/auth/me`, `POST /api/v1/auth/change-password`.
- `delivery/http/middleware.go` — `Wrap` enforces session cookie on non-public routes, injects `authdomain.WithCurrentUser` + `iamdomain.WithAuthContext`.
- `infrastructure/postgres/repository.go` — owns `iam_users` (identities) + `iam_user_sessions` (sessions). NB: identity rows shared with IAM module (which writes display fields via `iam_users` upsert).
- `infrastructure/memory/repository.go` — dev/test in-memory; imports `iamdomain.Role` (auth↔iam).

## Cross-deps (pre-scan, to confirm in Phase 3)

- **auth → iam (OUT)**: `iamdomain.Role`, `iamdomain.RoleProvider`, `iamdomain.RoleAdminRepository`, `iamdomain.WithAuthContext` (4+ sites).
- **iam → auth (IN)**: `iam/delivery/http/admin_handler.go:13` imports `authdomain` for `ManagedUser` listing.
- **documents/templates_v2 → auth (IN)**: consume `authdomain.CurrentUserFromContext` after middleware injection.
- **platform → auth (IN)**: composition root in `apps/api/cmd/metaldocs-api/main.go` wires Service + Middleware + Handler; `internal/platform/authn` likely supplies enabled flag.

Bidirectional auth↔iam circularity is real: auth Identity holds `[]iamdomain.Role`; IAM repository holds `iam_users` rows seeded by auth. Captured for §11.

## ADRs touching auth

- ADR 0007 (two-tier-authz) — auth provides middleware identity used by tier-1; auth itself NOT under tier-1 enforcement (`/api/v1/auth/login` is public).
- No standalone auth ADR. Session-cookie HMAC scheme, bcrypt-cost choice, lockout policy, single-table identity-with-roles model — all undocumented decisions.

## Phase 2 op picks (likely)

1. **Login (write)** — `POST /api/v1/auth/login` → `Service.Authenticate` → password verify, session create, role read, set cookie. Most representative regulated path.
2. **Session resolve (read/middleware)** — `Middleware.Wrap` → `Service.ResolveSession` → token HMAC verify, session lookup, touch, role build, context inject.
3. **Admin user create (write touching ManagedUser)** — `Service.CreateUser` → `Repository.CreateUser` + `RoleAdminRepository.ReplaceUserRoles` (cross-module write).

## Open questions deferred to tech-debt

- **Q1.** Session table: separate `auth_sessions` or shared `iam_user_sessions`? Ownership ambiguity if shared with IAM.
- **Q2.** Audit-trail emission on login / logout / password-change / admin-reset? None observed in source — likely Critical gap (regulated audit trail) given QMS scope.
- **Q3.** `LegacyHeaderEnabled` (X-User-Id bypass) still wired — auth-bypass risk if enabled in prod by misconfig.
- **Q4.** RFC 9457 envelope drift — handler emits `{error:{code,message,details,trace_id}}` (not RFC 9457 `application/problem+json`). Mirrors documents/iam T-001.
- **Q5.** Origin protection (`OriginProtection` config) — referenced in Config but enforcement site not yet located.
- **Q6.** No rate limiting visible on login endpoint beyond per-account lockout (`LoginMaxFailedAttempts`). IP-based throttle absent.
- **Q7.** `ResolveSession` calls `TouchSession` on every authenticated request → write amplification per request.
- **Q8.** Bidirectional auth↔iam coupling: refactor candidate or accepted given QMS identity overlap?

Proceed to Phase 1.
