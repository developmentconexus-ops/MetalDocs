# Data-flow Trace — POST /api/v1/auth/login

## 1. Entry point

| Layer | Symbol | File:line |
|---|---|---|
| OpenAPI op | n/a — stdlib mux registration | n/a — stdlib mux registration |
| Generated server stub | n/a — stdlib mux registration | n/a — stdlib mux registration |
| Handler | `(*Handler).handleLogin` | `internal/modules/auth/delivery/http/handler.go:42` |
| Route registration | `(*Handler).RegisterRoutes` -> `mux.HandleFunc("/api/v1/auth/login", h.handleLogin)` | `internal/modules/auth/delivery/http/handler.go:35` |

## 2. Call chain

1. `internal/modules/auth/delivery/http/handler.go:42` `(*Handler).handleLogin` — validates method/body, calls auth service, writes cookie + JSON response.
   -> calls: `internal/modules/auth/application/service.go:100` `(*Service).Authenticate`
2. `internal/modules/auth/application/service.go:100` `(*Service).Authenticate` — trims credentials, loads identity, checks lock/inactive, verifies password.
   -> calls: `internal/modules/auth/infrastructure/postgres/repository.go:21` `(*Repository).FindIdentityByIdentifier`
3. `internal/modules/auth/infrastructure/postgres/repository.go:21` `(*Repository).FindIdentityByIdentifier` — builds identifier lookup SQL.
   -> calls: `internal/modules/auth/infrastructure/postgres/repository.go:329` `(*Repository).loadIdentity`
4. `internal/modules/auth/infrastructure/postgres/repository.go:329` `(*Repository).loadIdentity` — executes `QueryRowContext(...).Scan(...)` and maps `sql.ErrNoRows` to `ErrIdentityNotFound`.
   -> returns to: `internal/modules/auth/application/service.go:107` `(*Service).Authenticate`
5. `internal/modules/auth/application/service.go:117` `bcrypt.CompareHashAndPassword` — verifies supplied password against `identity.PasswordHash`.
   -> on mismatch calls: `internal/modules/auth/infrastructure/postgres/repository.go:136` `(*Repository).RecordFailedLogin`
6. `internal/modules/auth/infrastructure/postgres/repository.go:136` `(*Repository).RecordFailedLogin` — updates `failed_login_attempts`, `locked_until`, `updated_at` on `metaldocs.auth_identities`.
   -> returns to: `internal/modules/auth/application/service.go:124` `(*Service).Authenticate`
7. `internal/modules/auth/application/service.go:129` `(*Service).Authenticate` success path — records successful login metadata.
   -> calls: `internal/modules/auth/infrastructure/postgres/repository.go:120` `(*Repository).RecordSuccessfulLogin`
8. `internal/modules/auth/infrastructure/postgres/repository.go:120` `(*Repository).RecordSuccessfulLogin` — sets `last_login_at`, resets lock counters.
   -> returns to: `internal/modules/auth/application/service.go:133` `(*Service).Authenticate`
9. `internal/modules/auth/application/service.go:439` `(*Service).newSessionToken` — generates 32-byte random token, signs via HMAC-SHA256, returns cookie value and SHA-256 token hash.
   -> calls: `internal/modules/auth/application/service.go:461` `(*Service).signToken`
10. `internal/modules/auth/application/service.go:461` `(*Service).signToken` — `hmac.New(sha256.New, []byte(s.cfg.SessionSecret))`, writes token bytes, base64url encodes MAC.
   -> returns to: `internal/modules/auth/application/service.go:447` `(*Service).newSessionToken` -> `hashToken(token)`
11. `internal/modules/auth/application/service.go:467` `hashToken` — computes SHA-256 and hex-encodes for `session_id`.
   -> returns to: `internal/modules/auth/application/service.go:146` `(*Service).Authenticate`
12. `internal/modules/auth/application/service.go:146` `(*Service).Authenticate` creates `authdomain.Session`.
   -> calls: `internal/modules/auth/infrastructure/postgres/repository.go:43` `(*Repository).CreateSession`
13. `internal/modules/auth/infrastructure/postgres/repository.go:43` `(*Repository).CreateSession` — inserts row into `metaldocs.auth_sessions`.
   -> returns to: `internal/modules/auth/application/service.go:154` `(*Service).Authenticate`
14. `internal/modules/auth/application/service.go:405` `(*Service).buildCurrentUser` — loads identity + roles for response payload.
   -> calls: `internal/modules/auth/infrastructure/postgres/repository.go:32` `(*Repository).FindIdentityByUserID`
15. `internal/modules/auth/infrastructure/postgres/repository.go:32` `(*Repository).FindIdentityByUserID` — builds user-id lookup SQL.
   -> calls: `internal/modules/auth/infrastructure/postgres/repository.go:329` `(*Repository).loadIdentity`
16. `internal/modules/auth/application/service.go:410` `s.roleProvider.RolesByUserID` — resolves roles for tenant.
   -> returns to: `internal/modules/auth/delivery/http/handler.go:61` `(*Handler).handleLogin`

## 3. State changes

| Entity | From | To | Trigger | Capability required |
|---|---|---|---|---|
| `metaldocs.auth_sessions` row | absent | inserted | successful login path in `Service.Authenticate` -> `repo.CreateSession` | n/a |
| `metaldocs.auth_identities.failed_login_attempts` | `N` | `N+1` (or reset to `0` on success) | password mismatch (`RecordFailedLogin`) / success (`RecordSuccessfulLogin`) | n/a |
| `metaldocs.auth_identities.locked_until` | existing value | `lock timestamp` on threshold, `NULL` on success | `RecordFailedLogin` / `RecordSuccessfulLogin` | n/a |

## 4. SQL touched

| File:line | Verb | Table(s) | Auth-area arg (if any) |
|---|---|---|---|
| `internal/modules/auth/infrastructure/postgres/repository.go:22` | SELECT | `metaldocs.auth_identities` | none |
| `internal/modules/auth/infrastructure/postgres/repository.go:331` | SELECT (`QueryRowContext` execution in `loadIdentity`) | `metaldocs.auth_identities` (query passed from caller) | none |
| `internal/modules/auth/infrastructure/postgres/repository.go:121` | UPDATE | `metaldocs.auth_identities` | none |
| `internal/modules/auth/infrastructure/postgres/repository.go:137` | UPDATE | `metaldocs.auth_identities` | none |
| `internal/modules/auth/infrastructure/postgres/repository.go:44` | INSERT | `metaldocs.auth_sessions` | none |
| `internal/modules/auth/infrastructure/postgres/repository.go:33` | SELECT | `metaldocs.auth_identities` | none |

Tripwire pairing: N/A — login is unauthenticated; auth tables not in tripwire scope.

## 5. Response shape

- Success response body from `internal/modules/auth/delivery/http/handler.go:62`:
  - `user`: `session.CurrentUser`
  - `expiresAt`: `session.ExpiresAt.UTC().Format(time.RFC3339)`
- Cookie written on success at `internal/modules/auth/delivery/http/handler.go:61` via `Service.SessionCookie`; cookie name source is `s.cfg.SessionCookieName` at `internal/modules/auth/application/service.go:373`.
- Error mapping in `internal/modules/auth/delivery/http/handler.go:131` `writeAuthError`:
  - `401` for `ErrInvalidCredentials` and `ErrIdentityNotFound`
  - `403` for `ErrIdentityLocked` and `ErrIdentityInactive`
  - `400` for `ErrPasswordPolicy`
  - `500` default branch
- Error envelope writer is `writeAPIError` at `internal/modules/auth/delivery/http/handler.go:166`, payload type `apiErrorEnvelope`/`apiError` (`error.code`, `error.message`, `error.details`, `error.trace_id`); this is legacy envelope, not RFC 9457.

## 6. Cross-references

- Idempotency: no.
- Pagination: no.
- Audit log emission: NO — only `log.Printf` in login failure path at `internal/modules/auth/delivery/http/handler.go:56`.
