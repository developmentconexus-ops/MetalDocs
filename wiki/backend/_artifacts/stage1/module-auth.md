# Stage-1 Audit Artifact — module-auth

> **Generated:** 2026-06-10
> **Branch audited:** qa/iam-area-membership
> **Repo root:** `internal/modules/auth/`
> **Read-only map — no redesign, no fixes.** Claims that cannot be confirmed from static code alone are tagged `[runtime-unverified]`.

---

## 1. Identity & purpose

The `auth` module is the sole owner of user authentication and HTTP session lifecycle inside MetalDocs. It answers the question "who is this request?" by verifying credentials at login, minting and storing opaque session tokens, and injecting a `CurrentUser` value into every request context via middleware. All four auth routes (`/api/v1/auth/{login,logout,me,change-password}`) are registered against the stdlib `net/http` mux without oapi-codegen — a deliberate carve-out documented in ADR 0012.

Beyond session management, the module owns the `auth_identities` and `auth_sessions` tables and supplies the user-administration methods (`CreateUser`, `UpdateUser`, `AdminResetPassword`, `UnlockUser`, `ListUsers`) that the IAM admin surface delegates to. Password hashing (bcrypt cost 12), per-account brute-force lockout (with Postgres advisory-lock serialization), and first-boot admin bootstrap are all service-layer concerns of this module. Authorization — "can this user do X?" — is explicitly out of scope and delegated to `internal/modules/iam/` per ADR 0007.

---

## 2. File inventory

### `domain/`
| File | Role |
|---|---|
| `domain/model.go` | Core value types: `Identity`, `Session`, `OnlineUser`, `ManagedUser`, `CreateUserParams`, `CreateUserInput`, `UpdateUserParams`, `BootstrapAdminParams`, `CurrentUser`, `AuthenticatedSession`; redaction guards on `AuthenticatedSession.String()` / `MarshalJSON()` |
| `domain/port.go` | `Repository` interface (17 methods); `CapabilityProvider` interface; `LoginState` and `LoginTx` interfaces for the advisory-lock critical section |
| `domain/context.go` | `WithCurrentUser` / `CurrentUserFromContext` — typed context accessors for the `CurrentUser` DTO |
| `domain/errors.go` | 13 sentinel errors (`ErrInvalidCredentials`, `ErrSessionNotFound`, `ErrSessionExpired`, `ErrSessionRevoked`, `ErrPasswordPolicy`, `ErrPasswordChangeRequired`, `ErrIdentityLocked`, `ErrIdentityInactive`, `ErrIdentityNotFound`, `ErrUserAlreadyExists`, `ErrTenantNotPermitted`, `ErrTenantClaimRequired`, `ErrTenantNotFound`) |

### `application/`
| File | Role |
|---|---|
| `application/service.go` | `Service` struct and all use cases; `Config` struct with 13 fields; session-token crypto helpers (`newSessionToken`, `tokenHashFromCookieValue`, `signToken`, `hashToken`); constant-time timing oracle mitigations; 779 lines |
| `application/service_test.go` | Unit tests for login, lockout, token verification, CreateUser shared-tx path, bcrypt-cost assertion, constant-time guard |

### `delivery/http/`
| File | Role |
|---|---|
| `delivery/http/handler.go` | `Handler` struct; `RegisterRoutes` (4 routes on stdlib mux); `handleLogin`, `handleLogout`, `handleMe`, `handleChangePassword`; `writeAuthError` error-to-problem mapper; `recordAudit` helper; `hashIdentifier` (SHA-256 of PII before logging) |
| `delivery/http/middleware.go` | `Middleware` struct; `Wrap` — session enforcement on every non-public request; `PublicPathChecker` function type; `defaultPublicPaths`; `isPasswordChangeAllowedPath` |
| `delivery/http/handler_problem_test.go` | Tests verifying RFC 9457 problem+json shape for every error code |
| `delivery/http/middleware_test.go` | Tests for session enforcement, public-path bypass, MustChangePassword gate |

### `infrastructure/postgres/`
| File | Role |
|---|---|
| `infrastructure/postgres/repository.go` | Postgres `Repository` implementing `authdomain.Repository`; all DML against `auth_identities` + `auth_sessions`; `WithinLoginLock` via `pg_advisory_xact_lock`; `CreateUserTx` / `BeginTx` (tx-aware extension for shared-transaction path); `RecordLastLoginContext` (cross-module write to `iam_users`); 667 lines |
| `infrastructure/postgres/sessions_admin.go` | `SessionListItem`, `SessionAdminQuery`, `ListActiveSessions` — admin-facing session listing with tenant-scoped JOIN on `iam_users`; consumed by `iam/delivery/http/sessions_handler.go` |
| `infrastructure/postgres/repository_test.go` | Integration tests requiring a live Postgres connection |
| `infrastructure/postgres/repository_unit_test.go` | sqlmock-based unit tests for `UpdateUser` no-op, `TouchSession` grace-window SQL |

### `infrastructure/memory/`
| File | Role |
|---|---|
| `infrastructure/memory/repository.go` | In-memory `Repository` implementing `authdomain.Repository`; also implements `iamdomain.RoleAdminRepository` (`HasAnyRole`, `UpsertUserAndAssignRole`, `ReplaceUserRoles`, `RolesByUserID`); `WithinLoginLock` via per-identity `sync.Mutex`; `SeedUserTenants` test helper; 540 lines |
| `infrastructure/memory/repository_test.go` | Unit tests for in-memory auth operations |
| `infrastructure/memory/repository_roles_test.go` | Unit tests for in-memory role administration methods |

---

## 3. Public surface

### Exported types consumed outside the module

| Symbol | Package | Consumers (verified) |
|---|---|---|
| `Config` | `application` | `platform/authn/config.go` (loads from env), `apps/api/cmd/metaldocs-api/main.go` (wires to `NewService` + `NewMiddleware`) |
| `Secret` | `application` | `platform/authn/config.go`, `apps/api/cmd/metaldocs-api/main.go` |
| `Service` | `application` | `apps/api/cmd/metaldocs-api/main.go:196`, `apps/api/cmd/metaldocs-e2e-seed/main.go`, `iam/application/people_service.go` (via `AuthService` interface), `iam/delivery/http/admin_handler.go` |
| `CurrentUser` | `domain` | `iam/delivery/http/middleware.go`, `iam/delivery/http/people_handler.go`, `platform/observability/http.go`, `platform/security/ratelimit.go`, `apps/api/cmd/metaldocs-api/reauth.go`, `internal/modules/controlleddocuments/application/service.go` |
| `CurrentUserFromContext` | `domain` | `iam/delivery/http/middleware.go`, `platform/observability/http.go:90`, `platform/security/ratelimit.go:182`, `delivery/http/handler.go`, `delivery/http/middleware.go`, `internal/modules/controlleddocuments/application/service.go:558` |
| `WithCurrentUser` | `domain` | `delivery/http/middleware.go:81` |
| `Repository` (interface) | `domain` | `platform/bootstrap/api.go:37`, `apps/api/cmd/metaldocs-api/reauth.go` |
| `ManagedUser`, `OnlineUser`, `UpdateUserParams` | `domain` | `iam/delivery/http/admin_handler.go:13`, `iam/delivery/http/people_handler.go` |
| `CreateUserInput`, `CreateUserParams` | `domain` | `iam/application/people_service.go:272-278` |
| `ErrPasswordPolicy`, `ErrUserAlreadyExists`, `ErrIdentityNotFound`, `ErrInvalidCredentials` | `domain` | `iam/delivery/http/admin_handler.go`, `iam/delivery/http/people_handler.go` |
| `Identity`, `PasswordHash` | `domain` | `apps/api/cmd/metaldocs-api/reauth.go` (reads `.PasswordHash` for re-auth e-signature) |
| `Handler`, `NewHandler` | `delivery/http` | `apps/api/cmd/metaldocs-api/main.go:221` |
| `Middleware`, `NewMiddleware`, `PublicPathChecker`, `Wrap` | `delivery/http` | `apps/api/cmd/metaldocs-api/main.go:237-238`, `apps/api/cmd/metaldocs-api/permissions.go:7` |
| `Repository` (postgres) | `infrastructure/postgres` | `platform/bootstrap/api.go:84`, `apps/api/cmd/metaldocs-api/main.go:255` (second instance for sessions admin) |
| `Repository` (memory) | `infrastructure/memory` | `platform/bootstrap/api.go:128` |
| `SessionListItem`, `SessionAdminQuery`, `ListActiveSessions` | `infrastructure/postgres` | `iam/delivery/http/sessions_handler.go` (via `SessionAdmin` interface) |

### HTTP routes

| Method | Path | Handler | Auth requirement | Notes |
|---|---|---|---|---|
| `POST` | `/api/v1/auth/login` | `Handler.handleLogin` (`handler.go:64`) | Public (no session) | Body: `{identifier, password}`; sets session cookie; returns `{user, expires_at}` |
| `POST` | `/api/v1/auth/logout` | `Handler.handleLogout` (`handler.go:99`) | Public (in `defaultPublicPaths`) | Revokes session if cookie present; returns 204 |
| `GET` | `/api/v1/auth/me` | `Handler.handleMe` (`handler.go:119`) | Session cookie required | Returns `CurrentUser` from context |
| `POST` | `/api/v1/auth/change-password` | `Handler.handleChangePassword` (`handler.go:132`) | Session cookie required; allowed during `MustChangePassword` lock | Body: `{current_password, new_password}` |
| `GET` | `/api/v1/auth/sessions` | `sessions_handler.go:50` | Session + tenant (IAM layer) | Admin session listing; registered by `iam/delivery/http/SessionsHandler` which depends on `authpg.Repository` |
| `GET/DELETE` | `/api/v1/auth/sessions/{id}` | `sessions_handler.go:50` | Session + tenant (IAM layer) | Session detail/revoke; same handler |

All four auth module routes bypass oapi-codegen and are registered directly via `mux.HandleFunc` (`handler.go:57-61`). Routes `/api/v1/auth/sessions*` are owned by `iam/delivery/http/sessions_handler.go` but depend on `authpg.Repository` for the SQL — a cross-boundary dependency.

---

## 4. Logic flows

### 4.1 Login — POST /api/v1/auth/login

1. **Decode request** (`handler.go:70-73`): JSON body `{identifier, password}`; malformed JSON → 400 `VALIDATION_ERROR`.
2. **Blank guard** (`service.go:205-208`): empty identifier or empty password → `ErrInvalidCredentials` without touching DB.
3. **Find identity** (`service.go:210-220`): `repo.FindIdentityByIdentifier` queries `auth_identities` by `LOWER(username) = $1 UNION ALL LOWER(email) = $1 AND username != $1` (`repository.go:31-49`). If `ErrIdentityNotFound`, spend bcrypt-equivalent time against `s.dummyHash` (constant-time timing oracle mitigation: `service.go:215-217`) then return `ErrInvalidCredentials`.
4. **Per-identity advisory lock** (`service.go:229`, `repository.go:341-362`): `repo.WithinLoginLock(userID, fn)` opens a Postgres transaction and acquires `pg_advisory_xact_lock(hashtextextended(userID, 0))`. The `fn` closure runs as `loginTx`, serializing concurrent login attempts on the same identity.
5. **Inside the lock — load state + bcrypt** (`service.go:229-254`): `tx.LoadLoginState` reads `password_hash, is_active, locked_until` from `auth_identities` on the locked transaction. `bcrypt.CompareHashAndPassword` always runs (equalize timing for locked/inactive/bad-password paths). Branch:
   - `locked_until > now()` → `outcome = loginLocked`
   - `!isActive` → `outcome = loginInactive`
   - password mismatch → `tx.RecordFailedLogin` atomically increments `failed_login_attempts` and conditionally sets `locked_until = NOW() + lockDuration` → `outcome = loginInvalid`
   - all pass → `outcome = loginOK`
6. **Error dispatch** (`service.go:259-266`): map outcome to typed sentinel; handler maps to RFC 9457 problem via `writeAuthError` (`handler.go:165-184`).
7. **Tenant resolution** (`service.go:270`, `service.go:316-338`): `resolveLoginTenant` queries `SELECT DISTINCT tenant_id FROM iam_user_roles WHERE user_id = $1` (`repository.go:101-122`). If a valid `X-Tenant-ID` claim matches, use it. Single-tenant user: auto-assign. Zero-tenant user + `AllowDevTenantFallback` → `DevTenantID`. Multi-tenant without a claim → `ErrTenantClaimRequired`.
8. **Record success + token** (`service.go:274-313`): `repo.RecordSuccessfulLogin` resets `failed_login_attempts = 0, locked_until = NULL, last_login_at = now`. `repo.RecordLastLoginContext` (best-effort, swallowed on error) updates `iam_users.last_login_ip/user_agent`. `newSessionToken` generates 32 random bytes, base64url-encodes to `token`, computes `HMAC-SHA256(secret, token)` as `sig`, yields cookie value `token.sig` and DB key `SHA-256(token)` (`service.go:725-756`).
9. **Create session + build user** (`service.go:300-313`): `repo.CreateSession` inserts `auth_sessions` row with `tenant_id`. `buildCurrentUser` loads identity + tenant name from `tenants` table + roles from `roleProvider.RolesByUserID` + optional capabilities from `capProvider` (wired in production via `CapabilityService`).
10. **Response** (`handler.go:89-96`): Sets `HttpOnly, SameSite=Strict, Secure=(cfg)` session cookie; returns `{user: CurrentUser, expires_at}` as JSON; emits `auth.login` audit event.

### 4.2 Session resolution — middleware on every non-public request

1. **Public check** (`middleware.go:55-58`): if `isPublic(method, path)` returns true → pass through without cookie verification. The public checker is injected from the composition root (`main.go:237-238`) so the auth middleware and IAM permission layer share one authoritative list.
2. **Cookie extraction** (`middleware.go:60-64`): if cookie missing or blank → 401 `AUTH_UNAUTHORIZED`.
3. **Token verification** (`service.go:340-372`): `tokenHashFromCookieValue` splits on `.`, recomputes `HMAC-SHA256(secret, tokenPart)` and uses `hmac.Equal` (constant-time comparison, `service.go:741`) to validate the signature. Returns `ErrSessionNotFound` on any mismatch.
4. **Session lookup** (`service.go:351-352`): `repo.FindSession` retrieves the row by SHA-256 hash (`repository.go:75-99`). On `ErrSessionNotFound` → 401.
5. **Expiry and revoke checks** (`service.go:355-367`): `RevokedAt != nil` → `ErrSessionRevoked`. `ExpiresAt < now` → `ErrSessionExpired`. `SessionIdleTimeout > 0 && now - LastSeenAt > IdleTimeout` → `ErrSessionExpired` (sliding idle timeout, new in current code).
6. **TouchSession** (`service.go:368`, `repository.go:140-165`): `UPDATE auth_sessions SET last_seen_at = $2 WHERE session_id = $1 AND last_seen_at < ($2 - INTERVAL '30 seconds')`. The 30-second grace window prevents a write on every in-window request. If `rows_affected = 0`, the repository double-checks existence to distinguish "already up-to-date" (no-op, normal) from "row missing" (`ErrSessionNotFound`).
7. **Build user + inject context** (`service.go:371`, `middleware.go:81-87`): `buildCurrentUser` loads identity, tenant name, roles, capabilities. Middleware injects `WithCurrentUser`, `iamdomain.WithAuthContext`, and `platformtenant.WithTenantID` into the request context. `X-Tenant-ID` header is deleted from the cloned request (`middleware.go:86`) so downstream handlers never see it.
8. **MustChangePassword gate** (`middleware.go:76-79`): if `currentUser.MustChangePassword && !isPasswordChangeAllowedPath(path, method)` → 403 `AUTH_PASSWORD_CHANGE_REQUIRED`.

### 4.3 Admin create user — CreateUserWithInput (called from iam/application/people_service.go)

1. **Input normalization** (`service.go:470-515`, `service.go:671-703`): trim whitespace, default `userID` to `username` if blank, default `displayName` to `username`, enforce exactly one valid role, validate password length.
2. **Hash password** (`service.go:706-722`): bcrypt cost 12 via `hashPasswordBytes`.
3. **Shared-transaction path** (`service.go:485-509`): if both `repo` and `roleAdmin` implement the tx-aware interfaces (`createUserTxRepository`, `replaceUserRolesTxRepository`, `beginTxRepository`), begin a single transaction, call `CreateUserTx` (INSERT `auth_identities`) and `ReplaceUserRolesTx` (UPSERT `iam_users` + DELETE/INSERT `iam_user_roles`) in the same tx, commit. This resolves the T-004 atomicity gap on the canonical Postgres path.
4. **Fallback path** (`service.go:511-514`): when either adapter lacks tx-awareness, executes `repo.CreateUser` (own transaction) then `roleAdmin.ReplaceUserRoles` sequentially. The orphan-identity window remains on this path.
5. **Response** emitted by `iam/delivery/http/people_handler.go`; audit emission handled there.

### 4.4 Password change — POST /api/v1/auth/change-password

1. **Auth gate**: middleware already verified session; `CurrentUserFromContext` must be present (`handler.go:137`).
2. **Decode + validate**: `{current_password, new_password}` JSON; `validatePassword` enforces `PasswordMinLength` (`service.go:641-645`).
3. **Current-password check** (`service.go:395-409`): if `!MustChangePassword`, current password must be non-empty and match bcrypt. If `MustChangePassword` and current password provided, it must still match (allows optional current-pw verification during forced change).
4. **Hash + update**: `hashPassword` (bcrypt cost 12); `repo.UpdateUser` atomically clears `must_change_password = false`, sets new `password_hash`.
5. **Refresh user + respond**: `service.CurrentUser` re-reads identity; handler returns `{changed: true, user: CurrentUser}`.
6. **Audit emission** (`handler.go:162`): `auth.password.changed` event with actorID.

### 4.5 Session token cryptography

Token generation (`service.go:725-756`):
- Generate 32 random bytes via `crypto/rand.Read`.
- `token = base64RawURL(rand32)` — the bearer secret, opaque to the server once stored.
- `sig = base64RawURL(HMAC-SHA256(sessionSecret, token))`.
- Cookie value: `token.sig` (two parts, `.` separator).
- DB key: `SHA-256(token)` as hex string — the `session_id` column never stores a usable token.

Token verification (`service.go:736-745`):
- Split on `.`; recompute expected sig; `hmac.Equal(parts[1], expected_sig)` — constant-time to prevent timing oracle on signature comparison.
- On match: compute `SHA-256(parts[0])` → look up in DB.

---

## 5. Dependencies

### Outbound (auth imports)

| Import | What for |
|---|---|
| `metaldocs/internal/modules/iam/domain` | `Role`, `RoleProvider` (build CurrentUser), `RoleAdminRepository` (create/replace user roles), `WithAuthContext` (inject IAM auth context in middleware), `ErrUserNotFound`, `ErrUserInactive`, `ErrNoRolesAssigned`, `ErrInvalidRole`, `RoleSystemAdmin`, `Capability` |
| `metaldocs/internal/platform/tenant` | `DevTenantID` sentinel, `WithTenantID` (middleware context inject), `FromContext` |
| `metaldocs/internal/platform/httpresponse` | `WriteJSON` |
| `metaldocs/internal/platform/problem` | `problem.Write`, `problem.New` — RFC 9457 responses |
| `metaldocs/internal/platform/requesttrace` | `Resolve` — extract trace ID for audit events |
| `metaldocs/internal/platform/security` | `ClientIP` — extract client IP respecting `TrustedProxyCIDRs` |
| `metaldocs/internal/modules/audit/domain` | `auditdomain.Writer`, `auditdomain.Event` — audit sink wired via `Handler.WithAudit` |
| `golang.org/x/crypto/bcrypt` | Password hashing |
| `github.com/google/uuid` | Audit event ID generation |
| `database/sql` | Shared tx interfaces |
| `github.com/lib/pq` | Postgres unique-violation detection (`pq.Error` code `23505`) |

### Inbound (who imports auth — grep-verified)

| Importer | Import path | What it uses |
|---|---|---|
| `apps/api/cmd/metaldocs-api/main.go` | `authapp`, `authdelivery`, `authpg` | Wires the module at startup |
| `apps/api/cmd/metaldocs-api/permissions.go` | `authdelivery` | `PublicPathChecker` type |
| `apps/api/cmd/metaldocs-api/reauth.go` | `authdomain` | `Repository.FindIdentityByUserID` for e-signature re-auth |
| `apps/api/cmd/metaldocs-e2e-seed/main.go` | `authapp` | Seed binary reuses `NewService` |
| `internal/platform/authn/config.go` | `authapp` | Loads `authapp.Config` from env |
| `internal/platform/bootstrap/api.go` | `authdomain`, `authmemory`, `authpg` | Wires repository into `APIDependencies` |
| `internal/platform/observability/http.go` | `authdomain` | `CurrentUserFromContext` at line 90 for structured log enrichment |
| `internal/platform/security/ratelimit.go` | `authdomain` | `CurrentUserFromContext` at line 182 for identity-keyed rate limiting |
| `internal/modules/iam/delivery/http/admin_handler.go` | `authdomain` | `ManagedUser`, `OnlineUser`, `UpdateUserParams`, error sentinels |
| `internal/modules/iam/delivery/http/middleware.go` | `authdomain` | `CurrentUserFromContext` |
| `internal/modules/iam/delivery/http/people_handler.go` | `authdomain` | `ManagedUser`, error sentinels, `CreateUserInput` types |
| `internal/modules/iam/delivery/http/sessions_handler.go` | `authdomain`, `authpg` | `Session`, `RevokeSession`, `RevokeSessionsByUserID`, `SessionAdmin` via `authpg.Repository` |
| `internal/modules/iam/application/people_service.go` | `authapp`, `authdomain` | `AuthService` interface calling `CreateUserWithInput`, `UpdateUser`, `ListUsers` |
| `internal/modules/controlleddocuments/application/service.go` | `authdomain` | `CurrentUserFromContext` at line 558 for audit actor |

---

## 6. Persistence

### Tables owned

**`metaldocs.auth_identities`** (migration `0021`, extended `0036`, `0222`):
- `user_id TEXT PK` — no `tenant_id`; identity is tenant-global.
- `username TEXT NOT NULL`, `email TEXT` (nullable), `display_name TEXT NOT NULL`, `is_active BOOLEAN NOT NULL DEFAULT TRUE`.
- `password_hash TEXT NOT NULL`, `password_algo TEXT NOT NULL`.
- `must_change_password BOOLEAN NOT NULL DEFAULT FALSE`.
- `last_login_at TIMESTAMPTZ`, `failed_login_attempts INT NOT NULL DEFAULT 0`, `locked_until TIMESTAMPTZ`.
- `last_failed_login_at TIMESTAMPTZ`, `last_failed_login_ip TEXT` (added by migration `0222` — PR-7 retro-land).
- `created_at TIMESTAMPTZ`, `updated_at TIMESTAMPTZ`.
- Indexes: `uq_auth_identities_username_ci (LOWER(username))`, `uq_auth_identities_email_ci (LOWER(email)) WHERE email IS NOT NULL`, `idx_auth_identities_locked_until` (migration `0222`).
- **No FK to `iam_users`** — decoupled by migration `0036`; FK to `auth_identities(user_id)` from `auth_sessions`.

**`metaldocs.auth_sessions`** (migration `0021`, extended `0036`, `0184`):
- `session_id TEXT PK` — stores SHA-256 hex of the token; never the raw token.
- `user_id TEXT NOT NULL REFERENCES auth_identities(user_id) ON DELETE CASCADE`.
- `tenant_id TEXT NOT NULL` — added by migration `0184`; backfilled from `iam_user_roles`, revoked on ambiguity (migration `0185`).
- `created_at TIMESTAMPTZ`, `expires_at TIMESTAMPTZ`, `revoked_at TIMESTAMPTZ` (nullable), `last_seen_at TIMESTAMPTZ`.
- `ip_address TEXT`, `user_agent TEXT` (truncated to 128 / 512 bytes at insert).
- Indexes: `idx_auth_sessions_user_id (user_id)`, `idx_auth_sessions_active (user_id, expires_at DESC) WHERE revoked_at IS NULL`, `idx_auth_sessions_tenant_user (tenant_id, user_id)` (migration `0184`).

### Cross-module writes (via injected ports, not direct SQL)

| Table | Who writes | Via |
|---|---|---|
| `metaldocs.iam_users` | `repository.go:RecordLastLoginContext` | Direct SQL on `iam_users` — cross-module write; best-effort, swallowed on error (`service.go:281`) |
| `metaldocs.iam_users`, `metaldocs.iam_user_roles` | `CreateUserWithInput` shared-tx path | `iamdomain.RoleAdminRepository.ReplaceUserRolesTx` |

### Cross-module reads

| Table | Query | File:line |
|---|---|---|
| `metaldocs.iam_user_roles` | `GetUserTenants` — DISTINCT tenant IDs for login tenant resolution | `repository.go:101` |
| `metaldocs.tenants` | `GetTenantByID` — tenant name/slug for `CurrentUser.TenantName` | `repository.go:124` |

### Active migrations

| Migration | Effect |
|---|---|
| `0021` | Creates `auth_identities` + `auth_sessions` (archive only — superseded by live `db/migrations/`) |
| `0036` | Adds `display_name`, `is_active` to `auth_identities`; decouples FKs from `iam_users` → `auth_identities` |
| `0184` | Adds `tenant_id NOT NULL` to `auth_sessions`; backfills from `iam_user_roles` |
| `0185` | Revokes ambiguously-pinned sessions (multi-tenant users assigned to `DevTenantID` by migration `0184`) |
| `0222` | Adds `last_failed_login_at`, `last_failed_login_ip` to `auth_identities`; adds locked index (PR-7 retro-land from archive `migrations/0210`) |

The canonical migration path is `db/migrations/` (numbers 0203+). Migrations 0021–0202 were applied from the legacy `migrations/` archive; the live directory starts at 0203. Migrations `0184`, `0185`, and `0222` exist only in the archive directory and must have been applied to production separately — their DDL effects are visible in the current code (`repository.go` inserts `tenant_id` in `CreateSession`; `RecordFailedLogin` writes `last_failed_login_ip`).

---

## 7. Config & environment

All loaded in `internal/platform/authn/config.go:37-149` via `authn.LoadRuntimeConfig()`.

| Config field | Env var | Default | Notes |
|---|---|---|---|
| `SessionSecret` (required) | `METALDOCS_AUTH_SESSION_SECRET` | none | HMAC key; rejected if empty and auth enabled; no key-id, rotation invalidates all sessions |
| `SessionCookieName` | `METALDOCS_AUTH_SESSION_COOKIE_NAME` | `metaldocs_session` | |
| `SessionTTL` | `METALDOCS_AUTH_SESSION_TTL_HOURS` | `12h` | |
| `SessionIdleTimeout` | `METALDOCS_AUTH_SESSION_IDLE_MINUTES` | `0` (disabled) | `0` preserves legacy no-idle-timeout behavior in dev/test |
| `PasswordMinLength` | `METALDOCS_AUTH_PASSWORD_MIN_LENGTH` | `8` | min value enforced as 8 |
| `LoginMaxFailedAttempts` | `METALDOCS_AUTH_LOGIN_MAX_FAILED_ATTEMPTS` | `5` | min 3 |
| `LoginLockDuration` | `METALDOCS_AUTH_LOGIN_LOCK_MINUTES` | `15m` | min 1 minute |
| `BootstrapAdminEnabled` | `METALDOCS_BOOTSTRAP_ADMIN_ENABLED` | `APP_ENV == local` | |
| `BootstrapAdminPassword` (required when enabled) | `METALDOCS_BOOTSTRAP_ADMIN_PASSWORD` | none | |
| `BootstrapAdminUserID` | `METALDOCS_BOOTSTRAP_ADMIN_USER_ID` | `admin-local` | |
| `BootstrapAdminUsername` | `METALDOCS_BOOTSTRAP_ADMIN_USERNAME` | `admin` | |
| `BootstrapAdminEmail` | `METALDOCS_BOOTSTRAP_ADMIN_EMAIL` | empty | |
| `BootstrapAdminName` | `METALDOCS_BOOTSTRAP_ADMIN_DISPLAY_NAME` | `Administrator` | |
| `CookieSecure` | `METALDOCS_AUTH_COOKIE_SECURE` | `APP_ENV != local` | |
| `OriginProtection` | `METALDOCS_AUTH_ORIGIN_PROTECTION_ENABLED` | `authn.Enabled()` | Loaded into `authCfg`; consumed by `security.NewOriginProtection` in `main.go:241-245`, NOT by auth middleware directly |
| `TrustedOrigins` | `METALDOCS_AUTH_TRUSTED_ORIGINS` | empty (CSV) | Same — passed to `security.OriginProtectionConfig`, not read in `delivery/http/middleware.go` |
| `TrustedProxyCIDRs` | loaded via `config.LoadTrustedProxyCIDRs()` | empty | Used in `service.remoteIP` to extract real client IP |
| `AllowDevTenantFallback` | not exposed as env var | `false` | Hardcoded false in config loader; only toggleable in memory/test mode |
| `METALDOCS_AUTH_ENABLED` | — | `true` | `authn.Enabled()` in `authn/config.go:17-23`; disabling only permitted when `APP_ENV=local` |

---

## 8. Concurrency & async

- **Per-identity login lock** (`repository.go:341-362`): each call to `WithinLoginLock` opens a dedicated Postgres transaction, acquires `SELECT pg_advisory_xact_lock(hashtextextended(userID, 0))`, runs the `LoginTx` critical section, then commits. The lock is held for the duration of the bcrypt comparison + the `RecordFailedLogin` write. High concurrency on the same identity serializes; no global lock. In-memory equivalent uses a per-`userID` `sync.Mutex` (`memory/repository.go:180-188`).
- **TouchSession 30-second grace window** (`repository.go:141-165`): the SQL `WHERE last_seen_at < ($2 - INTERVAL '30 seconds')` avoids a write on every request. After-grace no-ops perform a secondary existence SELECT to distinguish "up-to-date" from "missing row".
- **No goroutines, channels, or outbox writes in this module.** All operations are synchronous request-path database calls. No River jobs, no timers.
- **`dummyHash` precomputed at startup** (`service.go:131-134`): one `bcrypt.GenerateFromPassword` call in `NewService` to initialize the constant-time unknown-identity guard. No runtime goroutine.
- **Memory repository global mutex** (`memory/repository.go:17`): single `sync.Mutex` guards all map operations in the in-memory adapter; the `loginLocksMu` is a separate mutex protecting the per-identity lock map to avoid deadlock.

---

## 9. Error handling & observability

### Error handling

- All handler error returns map to RFC 9457 `application/problem+json` via `writeProblem` → `problem.Write` (`handler.go:215-218`). The `writeAuthError` switch (`handler.go:165-184`) covers 7 domain-error cases plus a catch-all 500.
- Middleware uses `problem.Write` directly (`middleware.go:62,69,73,77`); the return value is discarded with `_ =` — write failures silently drop the response body.
- `Logout` malformed-cookie error (`service.go:379-383`): `tokenHashFromCookieValue` error is now propagated and handled in `handleLogout` (`handler.go:105-111`); session is still cleared via `ExpiredSessionCookie`.
- `RecordLastLoginContext` failure is explicitly swallowed (`service.go:281`): best-effort update of `iam_users` governance columns at login.

### Logging

- `log.Printf` (stdlib, unstructured) used for: login failure (`handler.go:81`), logout failure (`handler.go:106`), change-password failure (`handler.go:149`), audit marshal failure (`handler.go:192`), audit write failure (`handler.go:211`), problem-write failure (`handler.go:217`), session-resolve failure (`middleware.go:72`).
- No `slog` usage in the auth module. No structured keys, no `request_id`, no `tenant_id` in log lines. This is a measurable observability gap.
- Identifier is hashed (SHA-256) before logging at `handler.go:80` to avoid PII leakage in logs.

### Audit

- `Handler` has an optional `audit auditdomain.Writer` injected via `WithAudit` (`handler.go:52-55`). In production this is wired (`main.go:221`).
- Audit events emitted: `auth.login` (success, `handler.go:94`), `auth.login.failed` (failure, `handler.go:83`), `auth.logout` (success, `handler.go:112`), `auth.password.changed` (`handler.go:162`).
- TraceID is derived from `requesttrace.Resolve(r.Context())` (`handler.go:195`) — a server-generated UUID that does not echo caller-supplied `X-Trace-Id`.
- TenantID for audit events is read from `CurrentUserFromContext` when available; may be empty on login-failure events where no session exists yet.
- **No metrics, no distributed traces, no OpenTelemetry spans** emitted by this module. Observability is limited to unstructured log lines and the audit event stream.

---

## 10. Legacy / duplication / smell flags

- **F-01 — `log.Printf` instead of `slog` across all handler and middleware error paths**
  - WHERE: `handler.go:81,106,149,192,211,217`; `middleware.go:72`
  - WHY: rest of the API uses `slog` with structured fields; auth uses stdlib `log.Printf` with no `request_id`, `tenant_id`, or other context keys. Unstructured log lines cannot be correlated with request traces and miss the observability contract. Not a functional bug; a gap in structured observability.

- **F-02 — N+1 role queries in `ListUsers`**
  - WHERE: `application/service.go:432-451` (`ListUsers`)
  - WHY: for each user in the full `auth_identities` result set, one `roleProvider.RolesByUserID(ctx, userID, tenantID)` call is issued. A TODO comment at `service.go:432` acknowledges this. At low user counts this is harmless; at production scale this is a performance hazard. No batch IN-clause alternative exists in `RoleProvider`.

- **F-03 — `auth_identities` lacks `tenant_id` (latent multi-tenant identity sharing)**
  - WHERE: `migrations/0021` (archive), `repository.go:379-408`
  - WHY: identity row is tenant-global; a single username can authenticate across tenants. Session-bound tenant (migration `0184`) mitigates the session layer but the identity row has no per-tenant isolation. In a true multi-tenant deployment where tenants should not share identity namespaces, this is a structural gap. Currently latent (single-tenant install). Tracked as T-008.

- **F-04 — No key-id in session cookie; HMAC key rotation invalidates all sessions**
  - WHERE: `application/service.go:725-756` (`newSessionToken`, `signToken`)
  - WHY: the session cookie format (`token.sig`) carries no key identifier. When `SessionSecret` is rotated, every active session becomes invalid simultaneously. A two-key rolling window would allow graceful rotation. Tracked as T-010.

- **F-05 — `OriginProtection` / `TrustedOrigins` declared in `authapp.Config` but never read by auth middleware**
  - WHERE: `application/service.go:50,54`; `platform/authn/config.go:140-141`; `delivery/http/middleware.go`
  - WHY: `OriginProtection` and `TrustedOrigins` are loaded into `authapp.Config` and stored there, but `delivery/http/middleware.go` never reads them to enforce origin validation on login/change-password. The enforcement is in `platform/security/origin_protection.go`, wired at `main.go:241-244` using the same config values. The fields on `authapp.Config` are therefore vestigial — they carry data but the auth module itself does not use them. Tracked as T-012.

- **F-06 — `AllowDevTenantFallback` not configurable via env var**
  - WHERE: `application/service.go:53`; `platform/authn/config.go` (no load site for this field)
  - WHY: `AllowDevTenantFallback` exists in `authapp.Config` but `authn.LoadRuntimeConfig()` never sets it — it defaults to `false`. The only way to enable it is to wire it manually in tests or the memory-mode bootstrap. This means the field cannot be turned on via env even in dev mode without code changes. The lack of an env-var load site makes this behavior invisible to operators. Latent/minor but the asymmetry with every other flag is confusing.

- **F-07 — `memory/repository.go` implements both `authdomain.Repository` and `iamdomain.RoleAdminRepository` in one struct (multi-interface fat adapter)**
  - WHERE: `infrastructure/memory/repository.go:1-540`
  - WHY: the in-memory adapter satisfies two distinct domain ports. The struct is 540 lines and exports methods from two different bounded contexts. This is correct for dev/test convenience (it avoids a second in-memory impl), but it creates a structural coupling between auth and IAM infrastructure in test code. Any change to `iamdomain.RoleAdminRepository` interface requires updating this file.

- **F-08 — `RecordFailedLogin` is a concrete method on `*postgres.Repository` but is NOT part of `authdomain.Repository`**
  - WHERE: `infrastructure/postgres/repository.go:365-377`
  - WHY: after the advisory-lock refactor, `RecordFailedLogin` was removed from the `authdomain.Repository` interface (production writes now go through `LoginTx.RecordFailedLogin` inside the lock). The method is retained as a concrete method on `*postgres.Repository` for tests that call it directly on the concrete type. This creates a surface that is callable from outside the port but not through it — a latent confusion between the port contract and the concrete implementation's test API.

- **F-09 — `TouchSession` secondary existence SELECT on every no-op grace-window hit is a potential hot path**
  - WHERE: `infrastructure/postgres/repository.go:141-165` (TODO comment at line 141)
  - WHY: when `last_seen_at` is within the 30-second window, the UPDATE affects 0 rows, triggering a second `SELECT EXISTS` to distinguish "in-window" (normal) from "missing row" (error). At high QPS with many in-window requests this doubles the DB round-trips. The TODO at line 141 acknowledges this. The grace window already reduces the problem, but the secondary check is structural overhead.

- **F-10 — `TODO` in `domain/model.go:31` referencing migration history**
  - WHERE: `domain/model.go:31-34`
  - WHY: `// TODO: migration 0021 originally missed display_name/is_active in early environments; keep later auth migrations applied before relying on them.` This is a deployment warning embedded in a domain model file — belongs in migration notes or a runbook, not a Go source comment.

- **F-11 — All 98 exported symbols lack Go doc comments**
  - WHERE: `application/service.go`, `delivery/http/handler.go`, `delivery/http/middleware.go`, `domain/*.go`, `infrastructure/*/repository.go`
  - WHY: `go doc` and IDE tooling yield no documentation for any exported symbol. `godoc` would produce empty pages. Tracked as T-011.

---

## 11. Wiki drift

The existing wiki doc (`wiki/modules/auth.md`, last verified `2026-06-08`) is largely accurate. The following specific claims have diverged from the current code state:

1. **T-001 status (wiki claims `LegacyHeaderEnabled` bypass is live; code shows it was removed)**
   - `wiki/modules/auth.md` §2 states: "`LegacyHeaderEnabled` — when true, requests with `X-User-Id` header bypass session enforcement entirely (`middleware.go:58-61`); single-flag compromise vector (T-001)."
   - `wiki/modules/auth-tech-debt.md` T-001 status is listed as open / critical.
   - Current `delivery/http/middleware.go` has NO `LegacyHeaderEnabled` check and NO `X-User-Id` branch. The field does not appear in `authapp.Config`. Commit `554c4007d` ("remove dead code, legacy flag") removed it.
   - Also: `wiki/modules/auth.md` §8.7 config table lists `LegacyHeaderEnabled` field with its env var — this field no longer exists in `authapp.Config`.
   - **Drift:** T-001 is resolved and closed. The critical severity flag should be removed or marked closed.

2. **T-002 Audit-trail gap — wiki says CLOSED but `auth-tech-debt.md` header still says open**
   - `wiki/modules/auth-tech-debt.md` T-002 header says "CLOSED 2026-05-11 (Plan 6a)" but the module wiki §11 summary still lists it as Critical (count: 2 critical).
   - Current `handler.go` emits `auth.login`, `auth.login.failed`, `auth.logout`, `auth.password.changed` audit events. Login and logout were gaps previously; they are now wired.
   - **Drift:** summary counts at wiki §11 and tech-debt register are inconsistent; T-002 is closed in the register but the module doc counts have not been adjusted.

3. **T-003 Legacy error envelope — wiki says CLOSED but `auth.md` §6.4 still claims RFC 9457 "not used"**
   - `wiki/modules/auth.md` §6.4 Failure modes table includes the row: "RFC 9457 Problem envelope: **not used** (T-003)."
   - `wiki/modules/auth-tech-debt.md` T-003 is marked "CLOSED 2026-05-12 (Plan 7)."
   - Current `handler.go` and `middleware.go` use `problem.Write` / `problem.New` throughout; the legacy `writeAPIError` function with `{error:{code,...}}` envelope is gone.
   - **Drift:** §6.4 failure modes table in `auth.md` still carries the stale "not used" note.

4. **T-006 TouchSession write amplification — wiki says CLOSED but code shows the secondary SELECT still runs**
   - `wiki/modules/auth-tech-debt.md` T-006 is marked "CLOSED 2026-05-25" with a note that TouchSession now uses a 30-second grace window.
   - The grace window is implemented. However, a `TODO` comment at `repository.go:141` remains: "consider updating only expired/grace-window-stale sessions to reduce write amplification further," and the secondary existence SELECT at `repository.go:157-163` still executes on every grace-window no-op.
   - **Minor drift:** the bulk of T-006 is resolved but the secondary SELECT overhead was not eliminated.

5. **Wiki `service.go` line anchors for `bcrypt.DefaultCost` are stale**
   - `wiki/modules/auth.md` §1.2 Quality Goals states: "source: `application/service.go:431` (`bcrypt.DefaultCost`)."
   - Current `service.go:29` uses `bcryptCost = 12` (not `bcrypt.DefaultCost = 10`). The cost was intentionally increased to 12. The line number reference is stale and the `DefaultCost` claim is incorrect.

6. **Wiki Key files section references `service.go:255` for `ResolveSession` — line number shifted**
   - `wiki/modules/auth.md` Key files section: "`service.go:255` (`ResolveSession`)". Current code has `ResolveSession` at line 340.

7. **Wiki lists `domain/port.go:23` for `GetUserTenants` as "new Repository method"**
   - The current `domain/port.go` has `GetUserTenants` at line 67, and also includes additional newer methods: `GetTenantByID`, `WithinLoginLock`, `RecordLastLoginContext`, `LoginState`, `LoginTx`, and `CapabilityProvider`. The port surface has grown materially since that note was written.

---

## 12. Open questions

1. **`[runtime-unverified]` Session idle timeout in production**: `METALDOCS_AUTH_SESSION_IDLE_MINUTES` defaults to `0` (disabled). Whether production deployments have this set cannot be confirmed without a live environment. The code path exists and is tested but the operational value is unknown.

2. **`[runtime-unverified]` `AllowDevTenantFallback` wiring**: The field is never set by `authn.LoadRuntimeConfig()` so it defaults to `false` in all environments including local. Whether there is a secondary config path that sets it (e.g. in the e2e-seed binary or test fixtures) outside this module's scope has not been exhaustively checked.

3. **`[runtime-unverified]` `auth_identities` / `auth_sessions` out-of-scope from `enforce_capability_asserted` tripwire**: `wiki/modules/auth.md` §2 states these tables are explicitly excluded from the Postgres tripwire (migration `0188`). This cannot be confirmed without reading migration `0188` (which is above the current `db/migrations/` range visible in this audit window — the highest migration is `0233`). The claim appears credible given the module's design, but verification requires reading the migration.

4. **`sessions_admin.go` / `SessionsHandler` ownership**: `infrastructure/postgres/sessions_admin.go` and its consumer `iam/delivery/http/sessions_handler.go` register routes at `/api/v1/auth/sessions*` but are owned and registered by the IAM module. The auth module provides the data contract (`authpg.SessionListItem`, `authpg.SessionAdminQuery`) but the route handler and business logic sit in IAM. This cross-boundary pattern (IAM handler + auth persistence type) is not formally documented as an ADR decision.

5. **`RecordLastLoginContext` cross-module SQL write**: `repository.go:211-235` writes directly to `metaldocs.iam_users` from the auth postgres repository. This cross-module SQL is best-effort and swallowed on error, but it means `auth` holds SQL dependency on the `iam_users` table schema without a formal port abstraction. A schema change to `iam_users.last_login_ip` columns requires updating auth's repository. Not tracked in T-007.
