# Stage-1 Audit Artifact — Platform: Identity, Tenancy & Security Primitives

> **Area:** platform-identity-tenancy
> **Packages covered:** `internal/platform/authn`, `internal/platform/security`, `internal/platform/tenant`, `internal/platform/ratelimit`, `internal/platform/sqlescape`
> **Produced:** 2026-06-10
> **Read-only audit — no code was modified.**

---

## 1. Identity & Purpose

This area comprises five platform packages that collectively form the infrastructure layer for request-identity resolution, tenant scoping, perimeter security, and safe SQL construction. They are depended upon by every domain module but own no domain logic themselves.

`platform/authn` is the startup-config and context-accessor tier for the authentication system: it loads the runtime `authapp.Config` struct from environment variables, exposes a presence-aware `UserIDFromContext` wrapper that fail-closes on missing or whitespace-only identities, and provides dev-mode role-injection helpers. `platform/security` is the HTTP-middleware package for CORS, origin-protection (CSRF-class defence), proxy-aware client-IP resolution, attachment HMAC signing, and the global fixed-window rate limiter. `platform/tenant` is a three-file context package that writes and reads the session-bound tenant ID, providing the single invariant that handlers must call `tenant.FromContext` rather than trusting a client-supplied header. `platform/ratelimit` is a separate, newer token-bucket rate-limiter with per-route quotas and a background sweeper; it is fully implemented but not yet activated in the production middleware chain. `platform/sqlescape` is a single-function package providing safe LIKE/ILIKE pattern escaping for Postgres.

---

## 2. File Inventory

### internal/platform/authn

| File | Role |
|---|---|
| `config.go` | Parses all authentication env vars into `authapp.Config`; exposes `Enabled()`, `CacheTTL()`, `LoadRuntimeConfig()`, `DevRoleMap()` |
| `context.go` | `UserIDFromContext(ctx) (string, bool)` — canonical presence-aware accessor wrapping `iamdomain.UserIDFromContext` |
| `config_test.go` | Unit tests for `LoadRuntimeConfig` env-var validation edge cases |
| `context_test.go` | Unit tests for `UserIDFromContext` (absent, whitespace, populated) |

### internal/platform/security

| File | Role |
|---|---|
| `attachmentsigner.go` | `AttachmentSigner` — HMAC-SHA256 sign, verify, and URL builder for attachment download links; `SignedURL` value type |
| `attachmentsigner_test.go` | Unit tests for expiry boundary, signature verification, URL construction |
| `cors.go` | `CORS` middleware — validates `Origin` header, sets CORS response headers, handles OPTIONS preflight |
| `origin_protection.go` | `OriginProtection` middleware — CSRF-class origin/referer validation on mutating verbs when a session cookie is present; `normalizeOrigin` helper shared with `cors.go` |
| `proxy.go` | `IsTrustedRemote`, `ClientIP` — proxy-aware IP resolution against a `[]netip.Prefix` allowlist; used by rate limiters and origin protection |
| `ratelimit.go` | `RateLimiter` — global fixed-window per-identity rate limiter; wrapped around the entire mux |

### internal/platform/tenant

| File | Role |
|---|---|
| `const.go` | `DevTenantID = "ffffffff-ffff-ffff-ffff-ffffffffffff"` — sentinel for single-tenant dev/test mode |
| `context.go` | `WithTenantID`, `FromContext`, `ErrTenantMissing` — context write/read for session-bound tenant |
| `context_test.go` | Round-trip, missing, and empty-panic tests |

### internal/platform/ratelimit

| File | Role |
|---|---|
| `config.go` | `RouteKey` enum, `Config` struct, `NewConfig`, `DefaultConfig` — per-route quota definitions (5 routes, req/min) |
| `middleware.go` | `Middleware` — token-bucket per-route rate limiter with `sync.Map`, background sweeper, fail-closed IP fallback |
| `config_test.go` | Tests for `NewConfig` validation, isolation, and zero-quota degradation regression |
| `eviction_test.go` | Tests for sweeper correctness, goroutine lifecycle, race-safe LoadOrStore |
| `export_test.go` | Test-only `NewConfigUnchecked` helper exposing internal constructor without validation |
| `middleware_test.go` | Tests for burst, per-user isolation |

### internal/platform/sqlescape

| File | Role |
|---|---|
| `like.go` | `LikeEscape(s string) string` — escapes `%`, `_`, `\` for Postgres LIKE/ILIKE patterns |
| `like_test.go` | Table-driven tests for escape edge cases |

---

## 3. Public Surface

### platform/authn — exported symbols

| Symbol | Signature | Purpose |
|---|---|---|
| `Enabled` | `func() bool` | Reads `METALDOCS_AUTH_ENABLED`; true by default |
| `CacheTTL` | `func() time.Duration` | Reads `METALDOCS_AUTHZ_CACHE_TTL_SECONDS`; default 30s |
| `LoadRuntimeConfig` | `func() (authapp.Config, error)` | Validates and assembles full auth config from env |
| `DevRoleMap` | `func() map[string][]iamdomain.Role` | Parses `METALDOCS_DEV_USER_ROLES`; `sync.Once` cached |
| `UserIDFromContext` | `func(context.Context) (string, bool)` | Presence-aware user ID accessor; returns `("", false)` on empty/whitespace |

### platform/security — exported symbols

| Symbol | Signature | Purpose |
|---|---|---|
| `NewAttachmentSigner` | `func(secret string) *AttachmentSigner` | Constructs HMAC signer; panics if secret < 32 bytes |
| `AttachmentSigner.Sign` | `func(attachmentID string, expiresAt time.Time) string` | Returns hex HMAC signature |
| `AttachmentSigner.Verify` | `func(attachmentID, expiresAtRFC3339, signature string) bool` | Validates signature and expiry |
| `AttachmentSigner.BuildDownloadURL` | `func(basePath, attachmentID string, expiresAt time.Time) SignedURL` | Constructs signed download URL |
| `SignedURL` | value type `{ URL string; ExpiresAt time.Time }` | Couples URL with expiry |
| `SignedURL.IsExpired` | `func(now time.Time) bool` | Expiry test; expired at or after `expiresAt` |
| `NewCORS` | `func(config.CORSConfig) *CORS` | Constructs CORS middleware |
| `CORS.Wrap` | `func(http.Handler) http.Handler` | HTTP middleware; no-ops if disabled |
| `NewOriginProtection` | `func(OriginProtectionConfig) *OriginProtection` | Constructs origin/CSRF middleware |
| `OriginProtection.Wrap` | `func(http.Handler) http.Handler` | HTTP middleware; checks mutating verbs when session cookie present |
| `IsTrustedRemote` | `func(*http.Request, []netip.Prefix) bool` | Returns true if RemoteAddr is in trusted CIDR list |
| `ClientIP` | `func(*http.Request, []netip.Prefix) netip.Addr` | Best-effort client IP; honors XFF when trusted |
| `NewRateLimiter` | `func(config.RateLimitConfig) *RateLimiter` | Global fixed-window per-identity rate limiter |
| `RateLimiter.Wrap` | `func(http.Handler) http.Handler` | HTTP middleware wrapping the mux |

### platform/tenant — exported symbols

| Symbol | Signature | Purpose |
|---|---|---|
| `DevTenantID` | `const string` | Dev sentinel UUID |
| `WithTenantID` | `func(context.Context, string) context.Context` | Writes tenant ID into context; panics on empty |
| `FromContext` | `func(context.Context) (string, error)` | Reads tenant ID; returns `ErrTenantMissing` on absence |
| `ErrTenantMissing` | `var error` | Sentinel error; handlers must treat as internal invariant violation |

### platform/ratelimit — exported symbols

| Symbol | Signature | Purpose |
|---|---|---|
| `RouteKey` | `type string` + 5 constants | Typed route identifiers for the quota map |
| `Config` | struct | Per-route quotas + sweeper config + trusted CIDR list |
| `NewConfig` | `func(map[RouteKey]int) (Config, error)` | Validated constructor; requires all quotas ≥ 1 |
| `DefaultConfig` | `func() Config` | Static defaults (presign 60, commit 30, export 20 req/min) |
| `Config.QuotaFor` | `func(RouteKey) (int, bool)` | Reads one route quota |
| `New` | `func(context.Context, Config) *Middleware` | Constructs middleware and starts sweeper goroutine |
| `Middleware.Limit` | `func(RouteKey, func(*http.Request) string, http.Handler) http.Handler` | Per-route token-bucket wrapper |
| `Middleware.Wait` | `func()` | Blocks until sweeper goroutine exits |
| `Middleware.SweepNow` | `func()` | Synchronous sweep for tests |
| `Middleware.Size` | `func() int` | Current entry count for tests |

### platform/sqlescape — exported symbols

| Symbol | Signature | Purpose |
|---|---|---|
| `LikeEscape` | `func(s string) string` | Escapes `%`, `_`, `\` for Postgres LIKE patterns |

---

## 4. Logic Flows

### Flow 1: Startup — loading auth configuration into `authapp.Config`

Called once at process start from `apps/api/cmd/metaldocs-api/main.go:171`.

1. `authn.LoadRuntimeConfig()` — `internal/platform/authn/config.go:37`
2. Reads `METALDOCS_AUTH_ENABLED` via `authn.Enabled()` — `config.go:17`; rejects `false` unless `APP_ENV=local`.
3. Reads `METALDOCS_AUTH_SESSION_SECRET`; enforces non-empty when auth enabled — `config.go:48`.
4. Reads optional `METALDOCS_AUTH_SESSION_TTL_HOURS` (default 12), `METALDOCS_AUTH_SESSION_IDLE_MINUTES` (default 0 = disabled), `METALDOCS_AUTH_PASSWORD_MIN_LENGTH` (default 8), `METALDOCS_AUTH_LOGIN_MAX_FAILED_ATTEMPTS` (default 5), `METALDOCS_AUTH_LOGIN_LOCK_MINUTES` (default 15) — `config.go:58–104`.
5. Calls `config.LoadTrustedProxyCIDRs()` — `internal/platform/config/trusted_proxy.go:13` — parses `METALDOCS_TRUSTED_PROXY_CIDRS`.
6. Assembles and returns `authapp.Config` — `config.go:125–149`; returns error on any invalid value.
7. In `main.go`, the resulting `authCfg` is passed to `authapp.NewService`, `authdelivery.NewMiddleware`, and `security.NewOriginProtection`. `authn.Enabled()` and `authn.CacheTTL()` are also called independently at `main.go:231, 237, 239`.

### Flow 2: Per-request tenant binding (auth middleware → context → handler)

The only production `WithTenantID` caller is `internal/modules/auth/delivery/http/middleware.go:83`.

1. Auth middleware resolves session cookie — `middleware.go:60–73`.
2. On success, builds three context values in sequence — `middleware.go:81–83`:
   - `authdomain.WithCurrentUser(ctx, currentUser)` — full `CurrentUser` struct including `TenantID`
   - `iamdomain.WithAuthContext(ctx, currentUser.UserID, currentUser.Roles)` — IAM-layer user ID + roles
   - `platformtenant.WithTenantID(ctx, currentUser.TenantID)` — **platform tenant context**
3. Strips `X-Tenant-ID` header from request clone — `middleware.go:85–86`; no downstream code can read a client-supplied tenant.
4. Handlers call `tenant.FromContext(r.Context())` — `internal/platform/tenant/context.go:27` — to retrieve the tenant ID. Returns `ErrTenantMissing` on absence (invariant violation, not a 400).
5. `tenant.DevTenantID` is referenced only by `internal/modules/auth/application/service.go:155, 187, 335, 483` when `AllowDevTenantFallback=true` (dev only).

### Flow 3: Global rate-limit check (security.RateLimiter)

Runs inside `httpObs.Wrap(rateLimiter.Wrap(mux))` — `main.go:598` — applied after CORS, origin-protection, and both auth middlewares.

1. `RateLimiter.Wrap` checks `shouldSkipRateLimit(path)` — `ratelimit.go:177` — skips `/api/v1/health/live` and `/api/v1/health/ready`.
2. `requestIdentity(req)` — `ratelimit.go:181` — resolves identity in preference order:
   a. `authdomain.CurrentUserFromContext` → `"user:<userID>"` if non-empty
   b. `iamdomain.UserIDFromContext` → `"user:<userID>"` as fallback
   c. `security.ClientIP(r, trustedCIDRs)` → `"ip:<addr>"` (uses `proxy.go:29`)
   d. Falls through to `"ip:unknown"` if all paths fail
3. `allow(identity)` — `ratelimit.go:111` — acquires `r.mu` lock, sweeps expired entries, checks fixed window.
4. On rejection: writes `HTTP 429` with `Retry-After` header and RFC 9457 `application/problem+json` body via `problem.Write` — `ratelimit.go:103`.

### Flow 4: Attachment URL signing and verification

Used by document download endpoints (inbound: `internal/modules/documents/delivery/http/handler.go`).

1. `security.NewAttachmentSigner(secret)` — `attachmentsigner.go:39` — panics if secret length < `MinAttachmentSecretLength` (32 bytes, NIST SP 800-107/FIPS 198-1).
2. `signer.BuildDownloadURL(basePath, attachmentID, expiresAt)` — `attachmentsigner.go:72`:
   - Calls `Sign(attachmentID, expiresAt)` — `attachmentsigner.go:54` — HMAC-SHA256 over `attachmentID + "|" + expiresAt.UTC().RFC3339`.
   - Appends `expiresAt` and `signature` as query params; returns `SignedURL{URL, ExpiresAt}`.
3. On download request, handler calls `signer.Verify(attachmentID, expiresAtRFC3339, signature)` — `attachmentsigner.go:60`:
   - Parses `expiresAtRFC3339`; returns `false` on parse failure.
   - Checks `now < expiresAt` (expired at boundary); returns `false` if expired.
   - Recomputes expected signature; uses `hmac.Equal` for constant-time comparison.

### Flow 5: LIKE pattern injection via sqlescape

Single callsite: `internal/modules/audit/infrastructure/postgres/writer.go` (the only non-test consumer).

1. Caller has a user-supplied search string that will be embedded in a Postgres LIKE/ILIKE clause.
2. `sqlescape.LikeEscape(s)` — `like.go:11` — replaces `\` → `\\`, `%` → `\%`, `_` → `\_` via `strings.NewReplacer`.
3. Caller appends `ESCAPE '\\'` to the surrounding LIKE clause so Postgres interprets the backslash escapes.
4. The value (not the identifier) is still passed as a parameterized query argument; `sqlescape` only escapes the LIKE metacharacters within the already-safe parameter value.

---

## 5. Dependencies

### Outbound imports

**platform/authn**
- `metaldocs/internal/modules/auth/application` — imports `authapp.Config`, `authapp.Secret`; this is the primary purpose of the package (assembles that Config from env)
- `metaldocs/internal/modules/iam/domain` — imports `iamdomain.Role`, `iamdomain.UserIDFromContext`, `iamdomain.RoleSystemAdmin` for `DevRoleMap` and `UserIDFromContext`
- `metaldocs/internal/platform/config` — imports `config.LoadTrustedProxyCIDRs`
- stdlib: `context`, `errors`, `fmt`, `os`, `strconv`, `strings`, `sync`, `time`

**platform/security**
- `metaldocs/internal/modules/auth/domain` — imports `authdomain.CurrentUserFromContext` in `ratelimit.go:182` for identity resolution
- `metaldocs/internal/modules/iam/domain` — imports `iamdomain.UserIDFromContext` in `ratelimit.go:185` as secondary identity fallback
- `metaldocs/internal/platform/config` — imports `config.RateLimitConfig`, `config.CORSConfig`
- `metaldocs/internal/platform/problem` — imports `problem.Write`, `problem.New` for error responses
- stdlib: `crypto/hmac`, `crypto/sha256`, `encoding/hex`, `fmt`, `log/slog`, `net`, `net/http`, `net/netip`, `net/url`, `strconv`, `strings`, `sync`, `time`

**platform/tenant**
- stdlib only: `context`, `errors`, `strings`

**platform/ratelimit**
- `metaldocs/internal/platform/problem` — RFC 9457 error body on 429 responses
- `metaldocs/internal/platform/security` — imports `security.ClientIP` for IP-fallback identity resolution
- `golang.org/x/time/rate` — `rate.Limiter` token bucket
- stdlib: `context`, `fmt`, `log/slog`, `net/http`, `net/netip`, `strconv`, `sync`, `sync/atomic`, `time`

**platform/sqlescape**
- stdlib only: `strings`

### Inbound consumers (verified by grep)

**platform/authn** — 15 files import it:
- `apps/api/cmd/metaldocs-api/main.go` — `LoadRuntimeConfig`, `Enabled`, `CacheTTL`
- `apps/api/cmd/metaldocs-e2e-seed/main.go` — `LoadRuntimeConfig`
- `internal/platform/bootstrap/api.go` — `Enabled`, `DevRoleMap`
- All delivery HTTP handlers (modules: taxonomy, iam, controlleddocuments, audit, documents/approval, search) — `UserIDFromContext`

**platform/security** — 9 files (excluding tests and worktrees):
- `apps/api/cmd/metaldocs-api/main.go` — `NewCORS`, `NewOriginProtection`, `NewRateLimiter`
- `internal/modules/auth/application/service.go` — imports for `ClientIP` (proxy-aware session IP recording)
- `internal/platform/ratelimit/middleware.go` — imports `security.ClientIP`
- Test files: `tests/unit/`

**platform/tenant** — 92 files import it across the entire module surface; the largest consumer set of any platform package in scope. Every domain-module HTTP handler and many application-service files call `tenant.FromContext`. The single `WithTenantID` production writer is `internal/modules/auth/delivery/http/middleware.go:83`.

**platform/ratelimit** — 8 files:
- `internal/modules/documents/delivery/http/handler.go` — `RegisterRoutesWithRateLimit`
- `internal/modules/documents/delivery/http/export_handler.go` — `RegisterRoutesWithRateLimit`
- `internal/modules/documents/module.go` — `RegisterRoutesWithRateLimit`, `buildLegacyMux`
- Test files under `tests/unit/` and `internal/platform/ratelimit/`

**platform/sqlescape** — 1 file:
- `internal/modules/audit/infrastructure/postgres/writer.go`

---

## 6. Persistence

**platform/authn, platform/security, platform/tenant, platform/ratelimit, platform/sqlescape** — stateless. None of these packages touch a database directly. All rate-limiter state is held in process memory (`sync.Map` in `ratelimit.Middleware`, `map[string]windowCounter` under `sync.Mutex` in `security.RateLimiter`). State is lost on process restart.

The attachment signer (`security.AttachmentSigner`) derives signatures from a shared secret configured at startup; no DB rows are written or read.

---

## 7. Config & Environment

All environment variables are read via `os.Getenv`; there is no centralized config struct parsed at one point — each package reads its own vars. The full list for this area:

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
| `METALDOCS_CORS_MAX_AGE_SECONDS` | `300` | `config/cors.go:35` | CORS max-age for preflight |
| `METALDOCS_RATE_LIMIT_ENABLED` | `false` | `config/ratelimit.go:22` | Enable global rate limiter |
| `METALDOCS_RATE_LIMIT_WINDOW_SECONDS` | `60` | `config/ratelimit.go:25` | Global rate limiter window |
| `METALDOCS_RATE_LIMIT_MAX_REQUESTS` | `120` | `config/ratelimit.go:33` | Global rate limiter request cap per window |
| `APP_PORT` | `8080` | `main.go:605` | HTTP listen port |

---

## 8. Concurrency & Async

**platform/ratelimit.Middleware** starts one background goroutine per `New(ctx, cfg)` call — `middleware.go:71`:
- Goroutine: `sweepLoop(ctx)` — runs `time.NewTicker(SweepInterval)`; on each tick calls `sweep()` which walks `sync.Map` and deletes entries with `lastAccess.Load() < cutoff`.
- Lifetime: goroutine exits when the constructor-provided `ctx` is cancelled.
- Join: `Middleware.Wait()` exposes the internal `sync.WaitGroup`; tests use it to assert goroutine termination.
- `size` is an `atomic.Int64` incremented on insert and decremented on sweep-eviction; used to enforce the `MaxEntries` fail-closed cap.
- `limiters` is a `sync.Map`; race-safe first-hit uses `LoadOrStore` — `middleware.go:245`.
- `lastAccess` on each `limiterEntry` is an `atomic.Int64` (unix nanoseconds), updated lock-free on every admission — `middleware.go:199`.

**platform/security.RateLimiter** uses a single `sync.Mutex` (`r.mu`) protecting `byIdentity map[string]windowCounter` — `ratelimit.go:37`. Per-call sweep (`sweepExpiredLocked`) runs inside the lock — `ratelimit.go:161`. No background goroutines.

All other packages in scope are stateless with no goroutines.

---

## 9. Error Handling & Observability

**Error patterns:**

- `platform/authn.LoadRuntimeConfig` returns typed `error` values with env-var names in messages — `config.go:51, 74, 83, 92, 101`; callers in `main.go` call `log.Fatalf` on error (process exit at startup).
- `platform/tenant.FromContext` returns `ErrTenantMissing` (sentinel) on absence — `context.go:14`; handler callers are documented to treat this as an internal invariant violation.
- `platform/tenant.WithTenantID` panics on empty tenant ID — `context.go:20`; panic at the write site is intentional (programming error, not user error).
- `platform/security.NewAttachmentSigner` panics if secret is shorter than 32 bytes — `attachmentsigner.go:41`; enforces FIPS 198-1 key-length invariant at construction time.
- `platform/security.CORS` and `OriginProtection` write RFC 9457 `application/problem+json` `403 Forbidden` responses via `problem.Write` — `cors.go:64`, `origin_protection.go:155`.
- `platform/security.RateLimiter` writes RFC 9457 `429 Too Many Requests` with `Retry-After` header — `ratelimit.go:102–103`.
- `platform/ratelimit.Middleware` writes RFC 9457 `429 Too Many Requests` with `Retry-After` header — `middleware.go:258–265`.

**Observability:**

- `platform/security.RateLimiter` logs via `slog.Default()`: `slog.Warn` on sweep eviction (`evicted`, `remaining`) and map-cap overflow (`map_size`, `max_entries`) — `ratelimit.go:162–174, 122–127`.
- `platform/ratelimit.Middleware` logs via injected `*slog.Logger` (defaults to `slog.Default()`): `slog.Warn` on sweep (`evicted`, `remaining`) and map-cap overflow; `slog.DebugContext` on empty-user IP fallback; `slog.WarnContext` on no-identity fail-closed; `slog.Error` on invalid quota (zero) — `middleware.go:147–151, 195–198, 169–170, 178–183`.
- No metrics or tracing calls in any of the five packages. [runtime-unverified: whether `slog` output is consumed by an OpenTelemetry log bridge]
- `platform/authn` has no log calls.
- `platform/tenant`, `platform/sqlescape` have no log calls.

---

## 10. Legacy / Duplication / Smell Flags

- **Duplicate `splitCSV` helper** — WHAT: identical implementation of a CSV-split function; WHERE: `internal/platform/authn/config.go:222` and `internal/platform/config/cors.go:63`; WHY: two independent copies with the same semantics suggest `authn` was written without reusing the `config` package helper. Neither is incorrect but the duplication will drift if trimming/splitting logic ever changes. (RF-7 hygiene candidate.)

- **Two parallel rate-limiter implementations** — WHAT: `platform/security.RateLimiter` (fixed window, global, identity-from-context) and `platform/ratelimit.Middleware` (token bucket, per-route, identity from injected function) coexist; WHERE: `internal/platform/security/ratelimit.go` and `internal/platform/ratelimit/middleware.go`; WHY: known intentional deferral documented in `wiki/architecture/rate-limiting.md §4`. The per-route limiter is the better long-term substrate (per that doc). The global limiter will eventually be retired. Flag retained here for completeness as a known legacy item.

- **Per-route rate limiter never activated in production** — WHAT: `platform/ratelimit.Middleware` and the `RegisterRoutesWithRateLimit` entry points on the documents module are fully implemented but the production startup path calls `docMod.RegisterRoutes(mux)` (nil limiter path) — `apps/api/cmd/metaldocs-api/main.go:501`; WHERE: `internal/modules/documents/module.go:118–119` (passes `nil, nil` to `buildLegacyMux`); WHY: intentional deferral per rate-limiting.md §2.2 (behavior change — clients see 429s on autosave/export). Two-line activation exists but is unmerged. Creates a misleading impression that per-route limiting is in effect.

- **`security.RateLimiter` reaches into two domain packages** — WHAT: `platform/security/ratelimit.go` imports both `authdomain` and `iamdomain` to resolve request identity from the session context; WHERE: `ratelimit.go:12–13`, `ratelimit.go:182–186`; WHY: platform packages should be leaf nodes with no upward imports into domain modules. This is an inverted dependency: a platform utility imports from application domain layers. The `platform/ratelimit.Middleware` avoids this by accepting a `userExtractor func(*http.Request) string` callback — a cleaner inversion of control. (RF-2 / middleware chain refactor surface.)

- **Double identity check in `security.RateLimiter.requestIdentity`** — WHAT: checks `authdomain.CurrentUserFromContext` first, then `iamdomain.UserIDFromContext` as a second fallback for the same "authenticated user ID"; WHERE: `ratelimit.go:182–186`; WHY: `authdomain.CurrentUser` and `iamdomain.UserIDFromContext` are written by the same auth middleware block (`auth/delivery/http/middleware.go:81–82`) in the same request, so both are always set together. The fallback is dead code in practice. The `platform/authn.UserIDFromContext` wrapper already provides the canonical single accessor.

- **`splitCSV` is duplicated and private, not a shared utility** — (same as the first bullet, called out separately because the `authn.splitCSV` also has a sister `parseBoolEnv` that duplicates `platform/config`'s equivalent pattern) — WHAT: `parseBoolEnv` at `authn/config.go:213` — WHERE: `authn/config.go:213–220`; WHY: `platform/config/cors.go` has the same `strings.EqualFold(raw, "true")` pattern inline; not extracted. Minor inconsistency.

- **`authn.DevRoleMap` uses `sync.Once` with package-level vars** — WHAT: `devRoleMapOnce` and `devRoleMapCached` are package-level singletons; WHERE: `authn/config.go:160–163`; WHY: env var is read once at first call and cached forever; cannot be reset between test cases without process restart. `t.Setenv` in tests will not invalidate the cache once set. This is a testing-in-production leak if the function is ever called in tests that set different `METALDOCS_DEV_USER_ROLES` values.

- **`httpObs` sits inside auth in the middleware chain** — WHAT: the middleware chain is `cors → originProtection → authMiddleware → iamMiddleware → httpObs → rateLimiter → mux`; WHERE: `apps/api/cmd/metaldocs-api/main.go:598, 602`; WHY: placing `httpObs` (observability) inside `authMiddleware` means 401/CORS-reject responses are not counted in RED metrics, violating `REQ-MW-4`. This is an existing gap in the middleware ordering, flagged as `RF-2` in the target architecture register.

---

## 11. Wiki Drift

- **`wiki/architecture/tenant-context.md §5` ("IAM legacy fallback pattern") no longer matches code** — The doc (line 126–138) describes a fallback in `iam/delivery/http/middleware.go` that reads `X-Tenant-ID` header when `tenant.FromContext` fails and falls back to `DevTenantID`. The actual code at `internal/modules/iam/delivery/http/middleware.go:92–98` (comment `C7`) has removed this fallback entirely: when `tenant.FromContext` returns an error, the IAM middleware now responds `401 AUTH_UNAUTHORIZED` immediately with no header or DevTenantID fallback. The doc's claim that "this fallback activates only when `LegacyHeaderEnabled=true`" is also outdated — there is no `LegacyHeaderEnabled` field visible in the current middleware code.

- **`wiki/architecture/backend-blueprint.md:151`** references `reauth*.go` files alongside `platform/authn`: the text says "re-auth flow (`apps/api/cmd/metaldocs-api/reauth*.go`)" as part of the authn description. `reauth.go` exists but is an HTTP middleware helper for signoff re-authentication, not a platform/authn file. The framing conflates the module-layer reauth flow with the platform-layer authn primitives. Minor framing drift, not an incorrect fact.

- **`wiki/architecture/rate-limiting.md §2.1` line-number anchor is stale** — The doc cites `apps/api/cmd/metaldocs-api/main.go:209` for `security.NewRateLimiter` and `main.go:471` for the handler chain. The actual line numbers at audit time are `main.go:276` (NewRateLimiter) and `main.go:598–602` (handler chain). Anchors have drifted since the doc was written at a smaller `main.go` line count.

---

## 12. Open Questions

- **[runtime-unverified]** Does `slog.Default()` output connect to an OpenTelemetry log bridge in any deployment? None of the five packages configure an exporter or set a handler; the log output depends on what the process root installs. If the answer is no, rate-limiter sweep/cap warnings are invisible in production dashboards.

- **[runtime-unverified]** `METALDOCS_AUTH_SESSION_IDLE_MINUTES` defaults to `0` (idle timeout disabled) at the platform config layer. The memory notes record "30m sliding idle timeout" as the production target value. Is the idle timeout currently enforced in any deployed environment, or is it still `0` everywhere?

- **[runtime-unverified]** `platform/ratelimit.Middleware` is constructed and unit-tested but never instantiated in the production binary (`ratelimit.New(ctx, cfg)` is called only in tests). The `DefaultConfig` quotas (presign 60 req/min, commit 30, export 20) have not been validated against real traffic baselines. Are those values calibrated for production load?

- **[runtime-unverified]** `security.RateLimiter` is controlled by `METALDOCS_RATE_LIMIT_ENABLED` which defaults to `false`. Is the global rate limiter currently enabled in any production environment? If not, the only active request-rate guard is the CORS and origin-protection middleware.

- **Genuine unknown:** `platform/sqlescape.LikeEscape` has exactly one non-test consumer (`audit/infrastructure/postgres/writer.go`). Whether all other query paths that interpolate search strings use parameterized queries exclusively (making `sqlescape` unnecessary for those paths) is asserted by the target architecture (`REQ-DATA-3`) but was not exhaustively verified in this audit. A targeted grep of `LIKE` / `ILIKE` clauses across all repository files would confirm.
