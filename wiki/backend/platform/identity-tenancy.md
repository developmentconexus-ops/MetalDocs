# Platform — Identity, Tenancy & Security Primitives

> **Last verified:** 2026-06-11
> **Scope:** Five platform packages that form the infrastructure layer for request-identity resolution, tenant scoping, perimeter security, and safe SQL construction: `internal/platform/authn`, `internal/platform/security`, `internal/platform/tenant`, `internal/platform/ratelimit`, `internal/platform/sqlescape`. This document covers their public surfaces, logic flows, inter-package dependencies, and known flags. Domain logic (credential verification, session lifecycle) lives in `internal/modules/auth/` and is documented separately.
> **Key files:**
> - `internal/platform/authn/config.go` — auth config loader from env; assembles `authapp.Config`
> - `internal/platform/authn/context.go` — `UserIDFromContext` canonical accessor
> - `internal/platform/security/ratelimit.go` — global fixed-window rate limiter
> - `internal/platform/security/cors.go` — CORS middleware
> - `internal/platform/security/origin_protection.go` — CSRF-class origin/referer guard
> - `internal/platform/security/proxy.go` — proxy-aware IP resolution
> - `internal/platform/security/attachmentsigner.go` — HMAC-SHA256 signed download URLs
> - `internal/platform/tenant/context.go` — session-bound tenant context write/read
> - `internal/platform/tenant/const.go` — dev sentinel UUID
> - `internal/platform/ratelimit/config.go` — per-route quota definitions
> - `internal/platform/ratelimit/middleware.go` — token-bucket per-route rate limiter
> - `internal/platform/sqlescape/like.go` — LIKE/ILIKE pattern escaping

---

## 1. Platform-vs-module responsibility split

A key invariant governs this area: **platform packages are domain-free leaf nodes**. They provide infrastructure primitives; they do not implement business rules.

| Concern | Platform package | Module counterpart |
|---|---|---|
| Auth configuration assembly (env → struct) | `platform/authn` (`config.go`) | `modules/auth/application` — consumes `authapp.Config` |
| Presence-aware user ID accessor | `platform/authn` (`context.go`) | `modules/iam/domain` — provides the underlying `iamdomain.UserIDFromContext` being wrapped |
| Session cookie verification, bcrypt, lockout | — (none) | `modules/auth/application/service.go` — owns all credential logic |
| Tenant context write | `platform/tenant` (`WithTenantID`) | `modules/auth/delivery/http/middleware.go:83` — the **sole production writer** |
| Tenant context read | `platform/tenant` (`FromContext`) | Every domain module HTTP handler — 46 direct call sites across 26 files (non-test) |
| CORS, origin protection, client IP, HMAC attachment signing | `platform/security` | No module counterpart; pure HTTP-perimeter concerns |
| Global fixed-window rate limit | `platform/security` (`RateLimiter`) | No module counterpart |
| Per-route token-bucket rate limit | `platform/ratelimit` (`Middleware`) | No module counterpart; currently not activated in production (see §6) |
| LIKE/ILIKE pattern escaping | `platform/sqlescape` | `modules/audit/infrastructure/postgres/writer.go` — sole consumer |

**REQ-TOP-2 note:** `platform/security/ratelimit.go` currently imports `modules/auth/domain` and `modules/iam/domain` to resolve request identity from context — an upward dependency that violates the platform-as-leaf-node invariant. This is flagged as RF-2. See §6.

---

## 2. Package reference

### 2.1 `platform/authn`

Startup-config and context-accessor tier for the authentication system. Its only responsibilities are (a) loading `authapp.Config` from environment variables at process start and (b) providing a presence-aware `UserIDFromContext` wrapper for use throughout the codebase.

| File | Role |
|---|---|
| `config.go` | Parses all authentication env vars into `authapp.Config`; exposes `Enabled()`, `CacheTTL()`, `LoadRuntimeConfig()`, `DevRoleMap()` |
| `context.go` | `UserIDFromContext(ctx) (string, bool)` — canonical presence-aware accessor wrapping `iamdomain.UserIDFromContext`; returns `("", false)` on empty or whitespace-only value |
| `config_test.go` | Unit tests for `LoadRuntimeConfig` env-var validation edge cases |
| `context_test.go` | Unit tests for `UserIDFromContext` (absent, whitespace, populated) |

**Exported surface:**

| Symbol | Signature | Purpose |
|---|---|---|
| `Enabled` | `func() bool` | Reads `METALDOCS_AUTH_ENABLED`; true by default |
| `CacheTTL` | `func() time.Duration` | Reads `METALDOCS_AUTHZ_CACHE_TTL_SECONDS`; default 30 s |
| `LoadRuntimeConfig` | `func() (authapp.Config, error)` | Validates and assembles full auth config from env |
| `DevRoleMap` | `func() map[string][]iamdomain.Role` | Parses `METALDOCS_DEV_USER_ROLES`; `sync.Once` cached |
| `UserIDFromContext` | `func(context.Context) (string, bool)` | Presence-aware user ID accessor |

**Inbound consumers:** 15 files. `apps/api/cmd/metaldocs-api/main.go` (`LoadRuntimeConfig`, `Enabled`, `CacheTTL`), `apps/api/cmd/metaldocs-e2e-seed/main.go`, `internal/platform/bootstrap/api.go`, and all domain-module delivery handlers for `UserIDFromContext`.

### 2.2 `platform/security`

HTTP-middleware package for CORS, CSRF-class origin protection, proxy-aware IP resolution, attachment URL signing, and the global fixed-window rate limiter. Wired at the composition root by `apps/api/cmd/metaldocs-api/main.go`.

| File | Role |
|---|---|
| `attachmentsigner.go` | `AttachmentSigner` — HMAC-SHA256 sign, verify, URL builder for attachment download links |
| `cors.go` | `CORS` middleware — validates `Origin` header, sets CORS response headers, handles OPTIONS preflight |
| `origin_protection.go` | `OriginProtection` — CSRF-class origin/referer validation on mutating verbs when a session cookie is present |
| `proxy.go` | `IsTrustedRemote`, `ClientIP` — proxy-aware IP resolution against a `[]netip.Prefix` allowlist |
| `ratelimit.go` | `RateLimiter` — global fixed-window per-identity rate limiter; wraps the entire mux |

**Exported surface:**

| Symbol | Signature | Purpose |
|---|---|---|
| `NewAttachmentSigner` | `func(secret string) *AttachmentSigner` | Constructs HMAC signer; panics if secret < 32 bytes (FIPS 198-1) |
| `AttachmentSigner.Sign` | `func(attachmentID string, expiresAt time.Time) string` | HMAC-SHA256 over `attachmentID + "|" + expiresAt.UTC().RFC3339` |
| `AttachmentSigner.Verify` | `func(attachmentID, expiresAtRFC3339, signature string) bool` | Constant-time signature + expiry validation |
| `AttachmentSigner.BuildDownloadURL` | `func(basePath, attachmentID string, expiresAt time.Time) SignedURL` | Constructs signed download URL with query params |
| `SignedURL` | value type `{ URL string; ExpiresAt time.Time }` | Couples URL with expiry |
| `SignedURL.IsExpired` | `func(now time.Time) bool` | True at or after `expiresAt` |
| `NewCORS` | `func(config.CORSConfig) *CORS` | Constructs CORS middleware |
| `CORS.Wrap` | `func(http.Handler) http.Handler` | No-ops if CORS is disabled |
| `NewOriginProtection` | `func(OriginProtectionConfig) *OriginProtection` | Constructs CSRF-class origin middleware |
| `OriginProtection.Wrap` | `func(http.Handler) http.Handler` | Checks mutating verbs when session cookie present |
| `IsTrustedRemote` | `func(*http.Request, []netip.Prefix) bool` | True if RemoteAddr is in trusted CIDR list |
| `ClientIP` | `func(*http.Request, []netip.Prefix) netip.Addr` | Best-effort client IP; honors `X-Forwarded-For` when source is trusted |
| `NewRateLimiter` | `func(config.RateLimitConfig) *RateLimiter` | Global fixed-window per-identity rate limiter |
| `RateLimiter.Wrap` | `func(http.Handler) http.Handler` | HTTP middleware wrapping the entire mux |

**Inbound consumers:** 3 files excluding tests: `apps/api/cmd/metaldocs-api/main.go` wires `NewCORS`, `NewOriginProtection`, `NewRateLimiter`; `internal/modules/auth/application/service.go` calls `ClientIP` for session IP recording; `internal/platform/ratelimit/middleware.go` calls `security.ClientIP`.

### 2.3 `platform/tenant`

Three-file context package. Writes and reads the session-bound tenant ID. The single invariant: handlers must call `tenant.FromContext` rather than trusting any client-supplied header. The `X-Tenant-ID` header is actively deleted from the request in the auth middleware (`modules/auth/delivery/http/middleware.go:85-86`).

| File | Role |
|---|---|
| `const.go` | `DevTenantID = "ffffffff-ffff-ffff-ffff-ffffffffffff"` — sentinel for single-tenant dev/test mode |
| `context.go` | `WithTenantID`, `FromContext`, `ErrTenantMissing` — context write/read |
| `context_test.go` | Round-trip, missing, and empty-panic tests |

**Exported surface:**

| Symbol | Signature | Purpose |
|---|---|---|
| `DevTenantID` | `const string` | Dev sentinel UUID |
| `WithTenantID` | `func(context.Context, string) context.Context` | Writes tenant ID; panics on empty (programming error, not user error) |
| `FromContext` | `func(context.Context) (string, error)` | Reads tenant ID; returns `ErrTenantMissing` on absence |
| `ErrTenantMissing` | `var error` | Sentinel; handlers must treat absence as internal invariant violation, not a 400 |

**Inbound consumers:** 46 direct call sites across 26 non-test files — the largest consumer set of any platform package in scope. Callers span HTTP handlers (delivery layer), application services, and infrastructure repositories across audit, controlled-documents, documents, iam, search, security, taxonomy, and templates modules. The single `WithTenantID` production writer is `internal/modules/auth/delivery/http/middleware.go:83`.

### 2.4 `platform/ratelimit`

A separate, newer token-bucket rate limiter with per-route quotas and a background sweeper. Fully implemented and tested but **not activated in the production middleware chain** (see §6, flag RL-2).

| File | Role |
|---|---|
| `config.go` | `RouteKey` enum (5 constants), `Config` struct, `NewConfig`, `DefaultConfig` — per-route quota definitions |
| `middleware.go` | `Middleware` — token-bucket per-route limiter using `sync.Map`, background sweeper, fail-closed IP fallback |
| `config_test.go` | Tests for `NewConfig` validation, isolation, zero-quota regression |
| `eviction_test.go` | Sweeper correctness, goroutine lifecycle, race-safe `LoadOrStore` |
| `export_test.go` | Test-only `NewConfigUnchecked` helper |
| `middleware_test.go` | Burst and per-user isolation tests |

**Exported surface:**

| Symbol | Signature | Purpose |
|---|---|---|
| `RouteKey` | `type string` + 5 constants | Typed route identifiers for the quota map |
| `Config` | struct | Per-route quotas + sweeper config + trusted CIDR list |
| `NewConfig` | `func(map[RouteKey]int) (Config, error)` | Validated constructor; all quotas must be ≥ 1 |
| `DefaultConfig` | `func() Config` | Static defaults: `uploads_presign` 60, `autosave_presign` 60, `autosave_commit` 30, `documents_render` 30, `export_pdf` 20 req/min (`config.go:68-74`) |
| `Config.QuotaFor` | `func(RouteKey) (int, bool)` | Reads one route quota |
| `New` | `func(context.Context, Config) *Middleware` | Constructs middleware and starts sweeper goroutine |
| `Middleware.Limit` | `func(RouteKey, func(*http.Request) string, http.Handler) http.Handler` | Per-route token-bucket wrapper |
| `Middleware.Wait` | `func()` | Blocks until sweeper goroutine exits (tests) |
| `Middleware.SweepNow` | `func()` | Synchronous sweep for tests |
| `Middleware.Size` | `func() int` | Current entry count for tests |

**Inbound consumers:** `internal/modules/documents/delivery/http/handler.go`, `export_handler.go`, `module.go` (all via `RegisterRoutesWithRateLimit`), plus test files.

### 2.5 `platform/sqlescape`

Single-function utility for safe LIKE/ILIKE pattern construction. Has exactly one non-test consumer.

| File | Role |
|---|---|
| `like.go` | `LikeEscape(s string) string` — escapes `%`, `_`, `\` for Postgres LIKE/ILIKE patterns |
| `like_test.go` | Table-driven tests for escape edge cases |

**Exported surface:** `LikeEscape(s string) string` — `like.go:11`.

**Sole consumer:** `internal/modules/audit/infrastructure/postgres/writer.go`.

---

## 3. Logic flows

### 3.1 Startup — loading `authapp.Config`

Called once at process start from `apps/api/cmd/metaldocs-api/main.go:171`.

1. `authn.LoadRuntimeConfig()` — `internal/platform/authn/config.go:37`.
2. Reads `METALDOCS_AUTH_ENABLED` via `authn.Enabled()` — `config.go:17`; rejects `false` unless `APP_ENV=local`.
3. Reads `METALDOCS_AUTH_SESSION_SECRET`; enforces non-empty when auth enabled — `config.go:48`.
4. Reads `METALDOCS_AUTH_SESSION_TTL_HOURS` (default 12), `METALDOCS_AUTH_SESSION_IDLE_MINUTES` (default 0, disabled), `METALDOCS_AUTH_PASSWORD_MIN_LENGTH` (default 8), `METALDOCS_AUTH_LOGIN_MAX_FAILED_ATTEMPTS` (default 5), `METALDOCS_AUTH_LOGIN_LOCK_MINUTES` (default 15) — `config.go:58-104`.
5. Calls `config.LoadTrustedProxyCIDRs()` — `internal/platform/config/trusted_proxy.go:13`.
6. Assembles and returns `authapp.Config` — `config.go:125-149`; returns error on any invalid value.
7. In `main.go`, the resulting `authCfg` is passed to `authapp.NewService`, `authdelivery.NewMiddleware`, and `security.NewOriginProtection`. `authn.Enabled()` and `authn.CacheTTL()` are also called independently at `main.go:231, 237, 239`.

### 3.2 Per-request tenant binding

The only production `WithTenantID` caller is `internal/modules/auth/delivery/http/middleware.go:83`.

1. Auth middleware resolves session cookie — `middleware.go:60-73`.
2. On success, builds three context values in sequence — `middleware.go:81-83`:
   - `authdomain.WithCurrentUser(ctx, currentUser)` — full `CurrentUser` struct including `TenantID`
   - `iamdomain.WithAuthContext(ctx, currentUser.UserID, currentUser.Roles)` — IAM-layer user ID + roles
   - `platformtenant.WithTenantID(ctx, currentUser.TenantID)` — **platform tenant context**
3. Strips `X-Tenant-ID` header from request clone — `middleware.go:85-86`; no downstream code can read a client-supplied tenant.
4. Handlers call `tenant.FromContext(r.Context())` — `internal/platform/tenant/context.go:27`. Returns `ErrTenantMissing` on absence, which is an internal invariant violation — not a 400.
5. `tenant.DevTenantID` is referenced only by `internal/modules/auth/application/service.go:155, 187, 335, 483` when `AllowDevTenantFallback=true` (dev only).

### 3.3 Global rate-limit check (`security.RateLimiter`)

Runs inside `httpObs.Wrap(rateLimiter.Wrap(mux))` at `main.go:598`, after CORS, origin-protection, and both auth middlewares.

1. `RateLimiter.Wrap` checks `shouldSkipRateLimit(path)` — `ratelimit.go:177`; skips `/api/v1/health/live` and `/api/v1/health/ready`.
2. `requestIdentity(req)` — `ratelimit.go:181` — resolves identity in preference order:
   - `authdomain.CurrentUserFromContext` → `"user:<userID>"` if non-empty
   - `iamdomain.UserIDFromContext` → `"user:<userID>"` as fallback
   - `security.ClientIP(r, trustedCIDRs)` → `"ip:<addr>"` (via `proxy.go:29`)
   - Falls through to `"ip:unknown"` if all paths fail
3. `allow(identity)` — `ratelimit.go:111` — acquires `r.mu` lock, sweeps expired entries, checks fixed window.
4. On rejection: writes HTTP 429 with `Retry-After` header and RFC 9457 `application/problem+json` body via `problem.Write` — `ratelimit.go:103`.

### 3.4 Attachment URL signing and verification

Used by document download endpoints (consumer: `internal/modules/documents/delivery/http/handler.go`).

1. `security.NewAttachmentSigner(secret)` — `attachmentsigner.go:39`; panics if secret length < `MinAttachmentSecretLength` (32 bytes, NIST SP 800-107/FIPS 198-1).
2. `signer.BuildDownloadURL(basePath, attachmentID, expiresAt)` — `attachmentsigner.go:72`:
   - Calls `Sign(attachmentID, expiresAt)` — `attachmentsigner.go:54` — HMAC-SHA256 over `attachmentID + "|" + expiresAt.UTC().RFC3339`.
   - Appends `expiresAt` and `signature` as query params; returns `SignedURL{URL, ExpiresAt}`.
3. On download request, handler calls `signer.Verify(attachmentID, expiresAtRFC3339, signature)` — `attachmentsigner.go:60`:
   - Parses `expiresAtRFC3339`; returns `false` on parse failure.
   - Checks `now < expiresAt`; returns `false` if expired.
   - Recomputes expected signature; uses `hmac.Equal` for constant-time comparison.

### 3.5 LIKE pattern injection via `sqlescape`

1. Caller in `internal/modules/audit/infrastructure/postgres/writer.go` has a user-supplied search string for a Postgres LIKE/ILIKE clause.
2. `sqlescape.LikeEscape(s)` — `like.go:11` — replaces `\` → `\\`, `%` → `\%`, `_` → `\_` via `strings.NewReplacer`.
3. Caller appends `ESCAPE '\\'` to the surrounding LIKE clause so Postgres interprets the backslash escapes.
4. The value is still passed as a parameterized query argument; `sqlescape` only escapes LIKE metacharacters within the already-safe parameter value.

---

## 4. Configuration reference

All environment variables read by this area. Each package reads its own vars via `os.Getenv`; there is no single centralized config parse.

| Env var | Default | Read by | Effect |
|---|---|---|---|
| `METALDOCS_AUTH_ENABLED` | `true` | `authn.Enabled()` `config.go:18` | Enable/disable auth middleware globally |
| `APP_ENV` | `""` (treated as `local`) | `authn.LoadRuntimeConfig` `config.go:38` | Governs whether `AUTH_ENABLED=false` is permitted |
| `METALDOCS_AUTH_SESSION_SECRET` | required when enabled | `config.go:48` | HMAC signing key for session tokens; min 32 chars |
| `METALDOCS_AUTH_SESSION_COOKIE_NAME` | `metaldocs_session` | `config.go:53` | Session cookie name |
| `METALDOCS_AUTH_SESSION_TTL_HOURS` | `12` | `config.go:58` | Absolute session TTL |
| `METALDOCS_AUTH_SESSION_IDLE_MINUTES` | `0` (disabled) | `config.go:71` | Sliding idle timeout; 0 = disabled |
| `METALDOCS_AUTH_PASSWORD_MIN_LENGTH` | `8` | `config.go:79` | Min password length |
| `METALDOCS_AUTH_LOGIN_MAX_FAILED_ATTEMPTS` | `5` | `config.go:88` | Failed attempts before lockout |
| `METALDOCS_AUTH_LOGIN_LOCK_MINUTES` | `15` | `config.go:97` | Lockout duration |
| `METALDOCS_AUTHZ_CACHE_TTL_SECONDS` | `30` | `authn.CacheTTL()` `config.go:25` | Role-cache TTL for `CachedRoleProvider` |
| `METALDOCS_BOOTSTRAP_ADMIN_ENABLED` | `true` when `APP_ENV=local` | `config.go:106` | Enable admin bootstrap |
| `METALDOCS_BOOTSTRAP_ADMIN_USER_ID` | `admin-local` | `config.go:107` | Bootstrap admin user ID |
| `METALDOCS_BOOTSTRAP_ADMIN_USERNAME` | `admin` | `config.go:111` | Bootstrap admin username |
| `METALDOCS_BOOTSTRAP_ADMIN_DISPLAY_NAME` | `Administrator` | `config.go:114` | Bootstrap admin display name |
| `METALDOCS_BOOTSTRAP_ADMIN_EMAIL` | `""` | `config.go:136` | Bootstrap admin email |
| `METALDOCS_BOOTSTRAP_ADMIN_PASSWORD` | required when bootstrap enabled | `config.go:137` | Bootstrap admin password (cleared after use) |
| `METALDOCS_AUTH_COOKIE_SECURE` | `true` when not local | `config.go:139` | Cookie `Secure` flag |
| `METALDOCS_AUTH_TRUSTED_ORIGINS` | `""` | `config.go:140` | CSV of trusted cross-origins for auth |
| `METALDOCS_AUTH_ORIGIN_PROTECTION_ENABLED` | `true` when auth enabled | `config.go:141` | Enable CSRF-class origin protection |
| `METALDOCS_DEV_USER_ROLES` | `admin-local:system_admin` | `authn.DevRoleMap()` `config.go:154` | Dev-mode role overrides (once-parsed) |
| `METALDOCS_TRUSTED_PROXY_CIDRS` | `""` (none trusted) | `config/trusted_proxy.go:13` | CSV of trusted upstream proxy CIDRs; shared across auth, rate limiters, origin protection |
| `METALDOCS_CORS_ENABLED` | `false` | `config/cors.go:20` | Enable CORS middleware |
| `METALDOCS_CORS_ALLOWED_ORIGINS` | `""` | `config/cors.go:22` | CORS allowed origins |
| `METALDOCS_CORS_ALLOWED_METHODS` | `GET,POST,PUT,OPTIONS` | `config/cors.go:24` | CORS allowed methods |
| `METALDOCS_CORS_ALLOWED_HEADERS` | `Content-Type,X-Trace-Id` | `config/cors.go:26` | CORS allowed headers |
| `METALDOCS_CORS_EXPOSED_HEADERS` | `""` | `config/cors.go:31` | CORS exposed headers |
| `METALDOCS_CORS_ALLOW_CREDENTIALS` | `false` | `config/cors.go:32` | CORS credentials flag |
| `METALDOCS_CORS_MAX_AGE_SECONDS` | `300` | `config/cors.go:35` | CORS preflight max-age |
| `METALDOCS_RATE_LIMIT_ENABLED` | `false` | `config/ratelimit.go:22` | Enable global rate limiter |
| `METALDOCS_RATE_LIMIT_WINDOW_SECONDS` | `60` | `config/ratelimit.go:25` | Global rate limiter window |
| `METALDOCS_RATE_LIMIT_MAX_REQUESTS` | `120` | `config/ratelimit.go:33` | Global rate limiter request cap per window |

---

## 5. Concurrency & async

**`platform/ratelimit.Middleware`** — starts one background goroutine per `New(ctx, cfg)` call (`middleware.go:71`):
- `sweepLoop(ctx)` runs `time.NewTicker(SweepInterval)`; on each tick calls `sweep()` which walks `sync.Map` and deletes entries with `lastAccess.Load() < cutoff`.
- Goroutine exits when the constructor-provided `ctx` is cancelled; `Middleware.Wait()` exposes the internal `sync.WaitGroup` for test join.
- `size` is an `atomic.Int64` for the `MaxEntries` fail-closed cap.
- `limiters` is a `sync.Map`; race-safe first-hit via `LoadOrStore` — `middleware.go:245`.
- `lastAccess` on each `limiterEntry` is an `atomic.Int64` (unix nanoseconds), updated lock-free on every admission — `middleware.go:199`.

**`platform/security.RateLimiter`** — uses a single `sync.Mutex` (`r.mu`) protecting `byIdentity map[string]windowCounter` — `ratelimit.go:37`. Per-call sweep (`sweepExpiredLocked`) runs inside the lock — `ratelimit.go:161`. No background goroutines.

All other packages in scope are stateless with no goroutines.

---

## 6. Legacy & open flags

> See also [`../legacy-register.md`](../legacy-register.md) for the cross-cutting register.

**RL-1 — Duplicate `splitCSV` helper**
Two identical implementations of a CSV-split function exist: `internal/platform/authn/config.go:222` and `internal/platform/config/cors.go:63`. The `authn` package also has a `parseBoolEnv` at `authn/config.go:213` that duplicates the same `strings.EqualFold(raw, "true")` pattern inline in `platform/config/cors.go`. Neither is incorrect; the duplication will drift if trimming/splitting logic changes. RF-7 hygiene candidate.

**RL-2 — `platform/security.RateLimiter` imports two domain packages (layering violation)**
`platform/security/ratelimit.go:12-13` imports `modules/auth/domain` and `modules/iam/domain` to resolve request identity from context. Platform packages must be domain-free leaf nodes (REQ-TOP-2). The newer `platform/ratelimit.Middleware` avoids this by accepting a `userExtractor func(*http.Request) string` callback — the correct inversion of control. RF-2 / middleware chain refactor surface.

**RL-3 — Dead fallback in `security.RateLimiter.requestIdentity`**
`ratelimit.go:182-186` checks `authdomain.CurrentUserFromContext` first, then `iamdomain.UserIDFromContext` as a second fallback for the same "authenticated user ID". Both are written by the same auth middleware block in the same request (`auth/delivery/http/middleware.go:81-82`), so both are always set or both absent simultaneously. The fallback is dead code in practice. `platform/authn.UserIDFromContext` already provides the canonical single accessor.

**RL-4 — `platform/ratelimit.Middleware` never activated in production**
`platform/ratelimit.Middleware` and the `RegisterRoutesWithRateLimit` entry points on the documents module are fully implemented and tested, but the production startup path calls `docMod.RegisterRoutes(mux)` (nil limiter path) — `apps/api/cmd/metaldocs-api/main.go:501`. `internal/modules/documents/module.go:118-119` passes `nil, nil` to `buildLegacyMux`. Two-line activation exists but is unmerged. Intentional deferral per `wiki/architecture/rate-limiting.md §2.2`. Creates a misleading impression that per-route limiting is in effect.

**RL-5 — `authn.DevRoleMap` uses `sync.Once` with package-level vars; not test-resettable**
`devRoleMapOnce` and `devRoleMapCached` are package-level singletons (`authn/config.go:160-163`). Once set, `t.Setenv` in tests will not invalidate the cache. Tests calling this function with different `METALDOCS_DEV_USER_ROLES` values will see the first-call result. Minor testing-in-production leak.

**RL-6 — `httpObs` sits inside auth in the middleware chain; 401/CORS rejections are not counted in RED metrics**
Chain order: `cors → originProtection → authMiddleware → iamMiddleware → httpObs → rateLimiter → mux` (`main.go:598, 602`). Requests rejected by CORS, origin-protection, or auth middleware bypass `httpObs`. This means 401 responses are not counted in the observability layer, violating REQ-MW-4. Flagged as RF-2.

**RL-7 — Wiki drift: `wiki/architecture/tenant-context.md §5` describes a removed `LegacyHeaderEnabled` X-Tenant-ID fallback**
The doc (line 126-138) describes a fallback in `iam/delivery/http/middleware.go` that reads `X-Tenant-ID` when `tenant.FromContext` fails. The actual code (`internal/modules/iam/delivery/http/middleware.go:92-98`) responds 401 immediately on `ErrTenantMissing` with no fallback. The `LegacyHeaderEnabled` field does not exist.

**RL-8 — Wiki drift: `wiki/architecture/rate-limiting.md §2.1` line-number anchors are stale**
That external doc cites `main.go:209` for `security.NewRateLimiter` and `main.go:471` for the handler chain. Verified actual line numbers: `main.go:276` (`security.NewRateLimiter`) and `main.go:598` (handler chain `httpObs.Wrap(rateLimiter.Wrap(mux))`). This document's own anchors (§3.3) use the correct line numbers; the stale references exist only in `wiki/architecture/rate-limiting.md` and require a separate correction to that file.

**Open questions:**

- [runtime-unverified] Does `slog.Default()` output connect to an OpenTelemetry log bridge in any deployment? The rate-limiter packages log sweep/cap warnings via `slog`; if no bridge is configured, these warnings are invisible in production dashboards.
- [runtime-unverified] `METALDOCS_AUTH_SESSION_IDLE_MINUTES` defaults to `0` (disabled). Whether any deployed environment has this set to the 30-minute target recorded in operating notes cannot be confirmed without a live environment.
- [runtime-unverified] `METALDOCS_RATE_LIMIT_ENABLED` defaults to `false`. Whether the global rate limiter is currently enabled in any production environment is unknown. If not, the only active request-rate guard is CORS and origin-protection middleware.
- [runtime-unverified] `platform/ratelimit.DefaultConfig` quotas (presign 60 req/min, commit 30, export 20) have not been validated against real traffic baselines.

---

## Sources

Stage-1 audit artifact: `wiki/backend/_artifacts/stage1/platform-identity-tenancy.md`

Related: [`../architecture/backend-blueprint.md`](../../architecture/backend-blueprint.md) (current-state grades) · [`../architecture/backend-target-architecture.md`](../../architecture/backend-target-architecture.md) (REQ-*/RF-* register)
