# Subagent prompt — Phase 2: Data-flow trace

You are a research-only Codex subagent. One operation per subagent. Output FACTS only.

## Task

For operation `resolve-session-middleware` (HTTP `n/a n/a`) in module `internal/modules/auth`, produce an artifact at `wiki/modules/auth/_artifacts/02-flow-resolve-session.md` tracing the call end-to-end.

### 1. Entry point

| Layer | Symbol | File:line |
|---|---|---|
| OpenAPI op | `n/a — middleware, not an HTTP op` | `n/a` |
| Generated server stub | `n/a — middleware, not a generated server op` | `n/a` |
| Handler | `(*Middleware).Wrap` | `internal/modules/auth/delivery/http/middleware.go:47` |

### 2. Call chain

1. `internal/modules/auth/delivery/http/middleware.go:47` `(*Middleware).Wrap` — auth middleware wrapper for non-public paths.
   → calls: `internal/modules/auth/delivery/http/middleware.go:53` `(*Middleware).isPublic`
2. `internal/modules/auth/delivery/http/middleware.go:40` `(*Middleware).isPublic` — checks injected `PublicPathChecker`, else fallback public-path list.
   → calls: `internal/modules/auth/delivery/http/middleware.go:44` `defaultPublicPaths`
3. `internal/modules/auth/delivery/http/middleware.go:96` `defaultPublicPaths` — default unauthenticated list:
   `/api/v1/health/live`, `/api/v1/health/ready`, `POST /api/v1/auth/login`, `POST /api/v1/auth/logout`.
   → calls: `internal/modules/auth/delivery/http/middleware.go:63` `(*http.Request).Cookie`
4. `internal/modules/auth/delivery/http/middleware.go:63` `r.Cookie(m.cfg.SessionCookieName)` — reads session cookie by configured cookie name.
   → calls: `internal/modules/auth/delivery/http/middleware.go:73` `(*Service).ResolveSession`
5. `internal/modules/auth/application/service.go:166` `(*Service).ResolveSession` — trims token and resolves authenticated session.
   → calls: `internal/modules/auth/application/service.go:172` `(*Service).tokenHashFromCookieValue`
6. `internal/modules/auth/application/service.go:450` `(*Service).tokenHashFromCookieValue` — splits `<token>.<sig>`, verifies signature with `hmac.Equal` (constant-time), hashes token payload.
   → calls: `internal/modules/auth/application/service.go:177` `Repository.FindSession`
7. `internal/modules/auth/infrastructure/postgres/repository.go:55` `(*Repository).FindSession` — loads session by `session_id`.
   → calls: `internal/modules/auth/application/service.go:181` revoked/expiry checks
8. `internal/modules/auth/application/service.go:181` revoked/expiry checks — rejects when `RevokedAt != nil` or `ExpiresAt.Before(time.Now().UTC())`.
   → calls: `internal/modules/auth/application/service.go:187` `Repository.TouchSession`
9. `internal/modules/auth/infrastructure/postgres/repository.go:80` `(*Repository).TouchSession` — updates `last_seen_at` for the session.
   → calls: `internal/modules/auth/application/service.go:190` `(*Service).buildCurrentUser`
10. `internal/modules/auth/application/service.go:405` `(*Service).buildCurrentUser` — assembles current user.
    → calls: `internal/modules/auth/application/service.go:406` `Repository.FindIdentityByUserID`
11. `internal/modules/auth/infrastructure/postgres/repository.go:32` `(*Repository).FindIdentityByUserID` — loads auth identity by `user_id`.
    → calls: `internal/modules/auth/application/service.go:410` `RoleProvider.RolesByUserID`
12. `internal/modules/iam/infrastructure/postgres/role_provider.go:19` `(*RoleProvider).RolesByUserID` — loads tenant-scoped roles for user.
    → calls: `internal/modules/auth/delivery/http/middleware.go:87` `authdomain.WithCurrentUser`
13. `internal/modules/auth/delivery/http/middleware.go:87` `authdomain.WithCurrentUser` — injects current user into context.
    → calls: `internal/modules/auth/delivery/http/middleware.go:88` `iamdomain.WithAuthContext`
14. `internal/modules/auth/delivery/http/middleware.go:88` `iamdomain.WithAuthContext` — injects IAM auth context into request context and forwards to next handler.

Facts in this flow:
- Legacy-header bypass: `internal/modules/auth/delivery/http/middleware.go:58` checks `m.cfg.LegacyHeaderEnabled && strings.TrimSpace(r.Header.Get("X-User-Id")) != ""`; when true it skips session auth and calls `next.ServeHTTP` (`:59`).
- Public-path checker injection exists: `internal/modules/auth/delivery/http/middleware.go:35` `WithPublicPathChecker`; if injected, `isPublic` uses it (`:41-42`) instead of `defaultPublicPaths`.

### 3. State changes

| Entity | From | To | Trigger | Capability required |
|---|---|---|---|---|
| `metaldocs.auth_sessions.last_seen_at` | previous timestamp | current request timestamp (`seenAt`) | `ResolveSession` path calls `TouchSession` | `n/a` |

### 4. SQL touched

| File:line | Verb | Table(s) | Auth-area arg (if any) |
|---|---|---|---|
| `internal/modules/auth/infrastructure/postgres/repository.go:56` | SELECT | `metaldocs.auth_sessions` | `n/a` |
| `internal/modules/auth/infrastructure/postgres/repository.go:82` | UPDATE | `metaldocs.auth_sessions` | `n/a` |
| `internal/modules/auth/infrastructure/postgres/repository.go:33` | SELECT | `metaldocs.auth_identities` | `n/a` |
| `internal/modules/iam/infrastructure/postgres/role_provider.go:21` | SELECT | `metaldocs.iam_users` | `tenantID` not used in this query |
| `internal/modules/iam/infrastructure/postgres/role_provider.go:37` | SELECT | `metaldocs.iam_user_roles` | `tenantID` used as `$2::uuid` (`:40`) |

Tripwire pairing:
- `N/A` (no authz.Require/authz.RequireAll call pair in this middleware/session-resolve path).

Write amplification fact:
- `TouchSession` runs `UPDATE metaldocs.auth_sessions SET last_seen_at = $2 WHERE session_id = $1` on each successful authenticated request (`internal/modules/auth/application/service.go:187` → `internal/modules/auth/infrastructure/postgres/repository.go:82`).

### 5. Response shape

- Unauthorized failures write legacy error envelope via `writeAPIError`:
  - `401 AUTH_UNAUTHORIZED` at `internal/modules/auth/delivery/http/middleware.go:65` and `:76`.
- Internal failure writes:
  - `500 INTERNAL_ERROR` at `internal/modules/auth/delivery/http/middleware.go:79`.
- Password-change gate writes:
  - `403 AUTH_PASSWORD_CHANGE_REQUIRED` at `internal/modules/auth/delivery/http/middleware.go:83` when `currentUser.MustChangePassword` (`:82`) and path is not allowed by `isPasswordChangeAllowedPath` (`:109`).
- Error envelope structure:
  - `apiErrorEnvelope{ error: { code, message, details, trace_id } }` at `internal/modules/auth/delivery/http/handler.go:148`, `:152`, `:166-167`.
- Success path:
  - middleware sets context and passes request through with `next.ServeHTTP` at `internal/modules/auth/delivery/http/middleware.go:89`.

### 6. Cross-references

- Idempotency: `no`.
- Pagination: `no`.
- Audit log emission: `no`.

## Constraints

- Read-only. No edits.
- No "should". No prescriptions. If a layer is missing (e.g. no service layer), write `n/a — handler calls repo directly`.
- Mark `(unclear: <why>)` instead of guessing.
- Cap artifact at 250 lines.

## Output

Write the single artifact to `wiki/modules/auth/_artifacts/02-flow-resolve-session.md` and print: operation id · layer count in §2 · tripwire pairing OK / VIOLATION / N/A.

Model: `--model gpt-5.3-codex`.
