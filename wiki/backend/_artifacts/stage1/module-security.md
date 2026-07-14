# Stage-1 Audit Artifact — module-security

**Area:** `internal/modules/security`
**Audit date:** 2026-06-10
**Source commits mapped:** `e408d7578` (PR-7 initial, 2026-06-02) · `aa7ef0416` (snake_case big-bang, 2026-06-08)

---

## 1. Identity & purpose

The `security` module backs the **Sessions & Security** admin tab (PR-7). It exposes
three read-only HTTP endpoints that give a tenant admin a real-time view of their
tenant's security posture: MFA enrollment coverage, currently-locked accounts, and
rule-based anomaly signals (repeated failed logins, lockout spikes, new-device logins,
off-hours admin actions). All data is computed on-demand per HTTP request from short
time-window SQL queries — there is no background aggregation, no persistent signal
store, and no cross-tenant visibility. The module is purely observational; it has no
write path and takes no security enforcement actions of its own.

---

## 2. File inventory

| File | Role |
|------|------|
| `internal/modules/security/domain/model.go` | Value types: `MfaCoverage`, `MfaRoleSlice`, `Lockout`, `Signal` |
| `internal/modules/security/domain/port.go` | `Repository` interface (read-only port); auxiliary types `RecentFailureSummary`, `NewDeviceLogin`, `OffHoursAction` |
| `internal/modules/security/application/service.go` | Orchestration layer: `Service`, `NewService`, `WithClock`, `MfaCoverage`, `ListLockouts`, `ListSignals`; rule constants; `stableID` hash helper |
| `internal/modules/security/application/service_test.go` | Unit tests: tenant-isolation, rule-firing, threshold, deterministic ID |
| `internal/modules/security/delivery/http/handler.go` | HTTP delivery: `Handler`, `RegisterRoutes`, three `handle*` methods, tenant extraction, nil-service fallback, JSON serialisation |
| `internal/modules/security/infrastructure/postgres/repository.go` | Postgres implementation of `domain.Repository`: five SQL queries against `auth_identities`, `auth_sessions`, `iam_users`, `iam_user_roles`, `audit_events` |

Total: 6 files. No sub-packages beyond the four canonical layers (`domain`, `application`, `delivery/http`, `infrastructure/postgres`).

---

## 3. Public surface

### Exported Go types / functions consumed outside the module

| Symbol | Consumed by | Purpose |
|--------|-------------|---------|
| `application.Service` | `apps/api/cmd/metaldocs-api/main.go:257–264` | Wired at startup; passed to `securitydelivery.NewHandler` and adapted to `mfaCoveragePctAdapter` for `iamapp.ObservabilityService` |
| `application.NewService(repo)` | `main.go:260` | Constructor |
| `application.Service.MfaCoverage` | `main.go:934` via `mfaCoveragePctAdapter` | Feeds `iam/kpi` KPI snapshot (`MfaCoveragePct` field) |
| `delivery/http.Handler` | `main.go:261–264, 289–291` | Route registration |
| `delivery/http.NewHandler(svc)` | `main.go:261, 263` | Constructor; accepts `nil` for SQL-less mode |

The module exports no domain types directly to other modules; `mfaCoveragePctAdapter` in `main.go:927–943` narrows `Service` to the single-method `MfaCoveragePctReader` interface defined in `iamapp` so the IAM package does not import the security domain.

### HTTP routes

| Method | Path | Auth requirement | Capability (PEP) |
|--------|------|-----------------|------------------|
| `GET` | `/api/v1/security/mfa-coverage` | Session required | `CapUserView` (`user.view`) |
| `GET` | `/api/v1/security/lockouts` | Session required | `CapUserView` (`user.view`) |
| `GET` | `/api/v1/security/signals` | Session required | `CapUserView` (`user.view`) |

Auth/authz is enforced by the outer middleware chain (`authMiddleware` + `iamMiddleware`) declared in `main.go:602` before the mux. The handler itself only calls `tenant.FromContext` (`handler.go:134`); it does **not** re-check capabilities internally. Permissions declared in `permissions.go:240–242`.

OpenAPI counterparts: `getMfaCoverage`, `listLockouts`, `listSecuritySignals` in `api/openapi/v1/openapi.yaml`.

---

## 4. Logic flows

### Flow 1 — `GET /api/v1/security/signals` (primary business flow)

1. **Auth middleware** validates session cookie, writes `tenant.Context` and `iamdomain.UserID` into request context (`authMiddleware.Wrap`, `main.go:602`).
2. **IAM middleware** checks `CapUserView` against the resolved role set; rejects with 403 if missing (`iamMiddleware.Wrap`, `main.go:602`).
3. **`handleSignals`** (`handler.go:97`) calls `h.requireTenant` (`handler.go:133`), which reads `tenant.FromContext` (`platform/tenant/context.go`). If missing, returns RFC 9457 problem `AUTH_UNAUTHORIZED 401`.
4. **Nil-service guard** (`handler.go:106`): if `h.service == nil` (no SQL DB, e.g. memory mode), returns `{"items": []}` 200 immediately.
5. **`service.ListSignals`** (`service.go:86`):
   a. Guards empty `tenantID` → `ErrTenantRequired` (`service.go:67, 88`).
   b. Pins wall-clock via `s.now()` (`service.go:90`).
   c. Calls **four repo methods in sequence** (not parallel):
      - `CountRecentFailedLoginsByUser(ctx, tenantID, 900, 5)` (`service.go:93`) — 15-minute window, threshold 5.
      - `CountRecentLockouts(ctx, tenantID, 3600)` (`service.go:108`) — 1-hour window.
      - `ListNewDeviceLogins(ctx, tenantID, 86400, 7776000)` (`service.go:123`) — 24h window, 90-day lookback.
      - `ListOffHoursAdminActions(ctx, tenantID, 86400, AdminRoles, 22, 6)` (`service.go:138`) — 24h window, UTC 22:00–06:00.
   d. Converts each result row into a `domain.Signal` with a deterministic `stableID` (`service.go:179–188`): SHA-1 over NUL-delimited `(kind, tenantID, evidence...)`, truncated to 16 hex chars, prefixed `"sig_"`.
   e. Sorts by `severityRank` desc, then `DetectedAt` desc (`service.go:154–161`).
6. **`handleSignals`** serialises signals to `map[string]any` rows (`handler.go:117–131`), wraps in `{"items": [...]}`, writes with `httpresponse.WriteJSON` (`handler.go:130`).
7. On any repo error, logs with `slog.Error` and returns RFC 9457 problem `INTERNAL_ERROR 500` via `problem.Write` (`handler.go:112–113`).

### Flow 2 — `GET /api/v1/security/mfa-coverage`

1–4: identical auth/tenant path as Flow 1.
5. **Nil-service guard** (`handler.go:44`): returns zero-value JSON (`writeMfaCoverageZero`, `handler.go:166`).
6. **`service.MfaCoverage`** (`service.go:69`) delegates directly to `repo.MfaCoverage`.
7. **`repo.MfaCoverage`** (`repository.go:27`) runs two SQL queries:
   - Scalar `COUNT` over `iam_users` filtered by `deactivated_at IS NULL` + `tenant_id` → totals (`repository.go:28–36`).
   - Grouped by `iam_user_roles.role_code` with JOIN to `iam_users` → per-role slices (`repository.go:39–70`). Pct is computed in Go, not SQL.
8. Handler serialises via `mfaCoverageToJSON` (`handler.go:148`).

### Flow 3 — `GET /api/v1/security/lockouts`

1–4: identical auth/tenant path.
5. **`service.ListLockouts`** → `repo.ListLockouts` (`repository.go:83`).
6. Query JOINs `auth_identities` to `iam_users` on `user_id` — this is how tenant scoping is achieved for `auth_identities`, which has **no `tenant_id` column** of its own (`repository.go:84–86`).
7. Filter: `locked_until IS NOT NULL AND locked_until > NOW()` (active lockouts only), `LIMIT 100` (`repository.go:99–102`).
8. Handler builds `map[string]any` rows, omitting nullable fields when nil (`handler.go:77–94`).

### Flow 4 — `repeated-failed-login` signal SQL detail

`CountRecentFailedLoginsByUser` (`repository.go:128`) uses `auth_identities.failed_login_attempts` as a proxy for "N failures within window" because there is no per-attempt log table. The query matches rows where `failed_login_attempts >= threshold AND last_failed_login_at >= NOW() - window`. False negatives occur when `RecordSuccessfulLogin` resets `failed_login_attempts` mid-window — this is a documented limitation in `repository.go:131–134` and `security-tech-debt.md` item 2.

### Flow 5 — `off-hours-admin-action` signal SQL detail

`ListOffHoursAdminActions` (`repository.go:235`) JOINs `audit_events` to `iam_user_roles` on `(user_id, tenant_id)`, filters `role_code = ANY($3)` (pg array), and uses `EXTRACT(HOUR FROM occurred_at AT TIME ZONE 'UTC')` for the time predicate (`repository.go:259`). A `GROUP BY` on `e.id` with `MIN(ur.role_code)` collapses multiple role memberships to one row per event (`repository.go:248–251`). `LIMIT 50` (`repository.go:264`).

---

## 5. Dependencies

### Outbound — packages imported by the security module

| Package | Import path | Why |
|---------|-------------|-----|
| `security/domain` | `metaldocs/internal/modules/security/domain` | Value types and Repository port |
| `platform/httpresponse` | `metaldocs/internal/platform/httpresponse` | `WriteJSON` helper |
| `platform/problem` | `metaldocs/internal/platform/problem` | RFC 9457 problem writer |
| `platform/tenant` | `metaldocs/internal/platform/tenant` | `tenant.FromContext` |
| `database/sql` | stdlib | Postgres driver types (`sql.DB`, `sql.NullTime`) |
| `github.com/lib/pq` | vendor | `pq.Array` for Postgres `ANY($n)` parameter |
| `crypto/sha1`, `encoding/hex` | stdlib | `stableID` deterministic hash |
| `log/slog` | stdlib | Structured error logging |

The module does **not** import any other business module (`auth`, `iam`, `audit`, `documents`, etc.). Cross-module data access goes directly through shared database tables.

### Inbound — who imports the security module (verified with grep)

| Importer | Import path | What it uses |
|----------|-------------|--------------|
| `apps/api/cmd/metaldocs-api/main.go:61–63` | `securityapp`, `securitydelivery`, `securitypg` | Full wiring: constructs repo, service, handler |
| `main.go:270` via `mfaCoveragePctAdapter` | `securityapp.Service` | Provides `MfaCoveragePct` to `iamapp.ObservabilityService` |

No other Go files import the security module (verified: `grep -r "modules/security"` returns only the 5 files within the module itself plus `main.go`).

---

## 6. Persistence

### Tables read (never written)

| Table | Access method | Security module query |
|-------|--------------|----------------------|
| `metaldocs.iam_users` | `QueryRowContext` / `QueryContext` | MFA coverage totals + per-role slices; JOIN anchor for lockout, failed-login, new-device tenant scoping |
| `metaldocs.iam_user_roles` | `QueryContext` | Per-role MFA slice; JOIN for off-hours admin role filter |
| `metaldocs.auth_identities` | `QueryContext` / `QueryRowContext` | Lockouts (`locked_until`), repeated-failed-login (`failed_login_attempts`, `last_failed_login_at`), recent lockout count |
| `metaldocs.auth_sessions` | `QueryContext` | New-device login detection (anti-self-join on `(user_id, user_agent, tenant_id)`) |
| `metaldocs.audit_events` | `QueryContext` | Off-hours admin action detection |

The module has **no write path** to any table.

### Migration files that own the columns this module reads

| Migration | Canonical path | What it adds |
|-----------|---------------|--------------|
| `0222_iam_mfa_and_failed_login_metadata.sql` | `db/migrations/0222_iam_mfa_and_failed_login_metadata.sql` | `iam_users.mfa_enabled`, `iam_users.mfa_enrolled_at`, `auth_identities.last_failed_login_at`, `auth_identities.last_failed_login_ip`; index on `auth_identities(locked_until)` |
| `0223_iam_last_failed_login_index.sql` | `db/migrations/0223_iam_last_failed_login_index.sql` | Partial index `auth_identities(last_failed_login_at)` for `CountRecentFailedLoginsByUser` range scan |

Columns `failed_login_attempts` and `locked_until` on `auth_identities`, and `auth_sessions` schema, are owned by earlier migrations (0184/0185 range) in the `auth` module.

### Query patterns

- All five queries are `SELECT`-only; no DML.
- All queries pass `tenantID` as the first positional parameter (`$1::uuid` or `$1`).
- `ListNewDeviceLogins` uses a correlated anti-join (`NOT EXISTS` subquery on `auth_sessions`).
- `ListOffHoursAdminActions` uses `pq.Array` for the admin-roles parameter.
- Hard `LIMIT`s: lockouts 100, new-device logins 50, off-hours actions 50, failed-login 100.

---

## 7. Config & environment

The security module consumes **no environment variables directly**. All configuration is implicit:

- **`deps.SQLDB`** (`main.go:258`): if `nil`, `securityapp.Service` is not constructed and `securitydelivery.NewHandler(nil)` is called, activating the nil-service fallback in every handler.
- **Rule tunables** (`application/service.go:28–41`): `RepeatedFailedLoginWindowSec`, `RepeatedFailedLoginThreshold`, `LockoutSpikeWindowSec`, `LockoutSpikeThreshold`, `NewDeviceLoginWindowSec`, `NewDeviceLoginLookbackSec`, `OffHoursWindowSec`, `OffHoursStartHourUTC`, `OffHoursEndHourUTC` — all compile-time constants, not runtime-configurable.
- **`AdminRoles`** (`service.go:46`): package-level `var`, not runtime-configurable; mirrors `iamdomain.Role*` constants.

---

## 8. Concurrency & async

The module is **entirely synchronous and stateless**. There are no goroutines, channels, timers, outbox writes, or background jobs within the module boundary.

All five repository queries execute sequentially inside `service.ListSignals` (`service.go:93–149`). They share the same `context.Context` from the HTTP request; cancellation propagates to all in-flight queries if the client disconnects.

The `Service` struct carries only `repo` (interface) and `now` (function); both are assigned once at construction and not mutated after that, making `Service` safe for concurrent use from multiple goroutines.

---

## 9. Error handling & observability

### Error handling

- Tenant guard: `strings.TrimSpace(tenantID) == ""` returns sentinel `ErrTenantRequired` (`service.go:67`). Checked in all three service methods (`service.go:70, 78, 88`).
- Repository errors are wrapped with `fmt.Errorf("<rule>: %w", err)` in the service (`service.go:95, 109, 124, 139`) and again in the repo (`repository.go:37, 54, 68, etc.`) — two-level wrapping chain.
- Handler converts any non-nil error to `problem.New(500, "INTERNAL_ERROR", ...)` via `h.writeProblem` (`handler.go:51, 73, 112`). No error variant mapping; all repository failures collapse to 500.
- Method guard: non-GET requests return bare `405` with no body (`handler.go:37–39, 58–60, 98–100`).
- Tenant extraction failure returns RFC 9457 problem `AUTH_UNAUTHORIZED 401` (`handler.go:136–138`).

### RFC 9457 usage

`problem.New` + `problem.Write` are used for 401 and 500 responses. The 405 responses bypass the problem writer (bare `w.WriteHeader(http.StatusMethodNotAllowed)`).

### Logging

- `slog.Error("security: mfa coverage failed", "err", err)` — `handler.go:50`
- `slog.Error("security: lockouts failed", "err", err)` — `handler.go:73`
- `slog.Error("security: signals failed", "err", err)` — `handler.go:112`
- `slog.Warn("security: write response failed", "err", err)` — `handler.go:144` (response-write failure)

No metrics, tracing, or structured audit events are emitted by this module.

### JSON serialisation

The handler serialises domain types manually to `map[string]any` rather than using struct tags or a generated codec (`handler.go:76–94, 117–131, 148–163`). The `writeJSON` package-level var (`handler.go:19`) aliases `httpresponse.WriteJSON` — this indirection exists to allow test overrides.

---

## 10. Legacy / duplication / smell flags

- **`map[string]any` manual serialisation** — `handler.go:76–131, 148–163`: All three handlers hand-assemble `map[string]any` row-by-row instead of using struct-tag JSON marshalling or a generated codec. The same pattern is inconsistent with sibling modules (e.g. `iam`, `audit`) that use typed response structs or codegen'd types. Risk: field name drift between the map keys and the OpenAPI schema is invisible to the compiler.

- **No OpenAPI codegen wiring** — `delivery/http/handler.go` entire file: Unlike `iam`, `documents`, `audit`, and `templates` modules, `security` is not wired through `oapi-codegen`. The handler is hand-written and the request/response shapes are manually serialised, meaning the OpenAPI contract (`SecuritySignal`, `MfaCoverage`, `ListLockoutsResponse`) is not enforced by generated types at compile time. This is the primary structural divergence from the backend-canon pattern.

- **Nil-service silent fallback returns 200 with empty data** — `handler.go:44–47, 66–68, 106–109`: When `h.service == nil` (no SQL DB), the MFA-coverage handler calls `writeMfaCoverageZero` (returns 200 with `total_users: 0, mfa_enabled: 0`) and the lockouts/signals handlers return `{"items": []}`. This is silent — no indication to the caller that the data is absent because the backend is unconfigured, not because the tenant has no events. The tech-debt doc (item 10) describes this as "returns 501" which is incorrect — it is 200 with empty data.

- **Sequential rule evaluation with no short-circuit** — `application/service.go:93–149`: The four signal-rule repo calls execute in series. If the first query is slow, the latency compounds. There is no context-cancellation-aware early exit for the no-op case (empty tenant). This is a correctness-compliant but suboptimal layout for a latency-sensitive admin tab.

- **SHA-1 used for `stableID`** — `application/service.go:179`: `crypto/sha1` is used to generate signal IDs. SHA-1 is not cryptographically strong; however, the use here is content-addressing for UI deduplication, not security. The risk is cosmetic (may trigger static analysis warnings) rather than a real vulnerability.

- **`auth_identities` cross-tenant scoping via JOIN only** — `repository.go:84–100`: `auth_identities` has no `tenant_id` column; tenant isolation for lockout queries is achieved solely by JOINing to `iam_users` on `user_id`. If a future query omits the JOIN or joins incorrectly, data from other tenants leaks. Comment at `repository.go:84–87` documents this, but the structural dependency is fragile.

- **Rule constants are public but undocumented in godoc** — `application/service.go:28–41`: The exported constants (`RepeatedFailedLoginWindowSec` etc.) carry only a brief inline comment. They are labelled "public so tests can override" but no test in `service_test.go` actually overrides them — tests rely on default values. The "configurable" claim in the comment is aspirational; there is no injection mechanism.

- **`time` import kept alive with blank var** — `application/service_test.go:144`: `var _ = time.Second` exists solely to keep the `time` import alive "in case future tests use it." This is dead code.

- **Missing 403 response on method-not-allowed** — `handler.go:37–39, 58–60, 98–100`: Non-GET requests receive a bare 405 with no body, bypassing `problem.Write`. The OpenAPI spec does not define a 405 response for these routes. Minor inconsistency vs. the RFC 9457 usage elsewhere in the handler.

---

## 11. Wiki drift

- **`security-tech-debt.md:27` cites `migrations/0210_iam_mfa_and_failed_login_metadata.sql`** as the migration that added stub MFA columns. That file lives in `archive/migrations/0210_iam_mfa_and_failed_login_metadata.sql` and is not replayed at runtime. The canonical, actively-applied migration is `db/migrations/0222_iam_mfa_and_failed_login_metadata.sql`. The wiki link and version number are stale.

- **`security-tech-debt.md` item 10 says "returns `501 Not Implemented`" when `deps.SQLDB == nil`**. The actual code (`handler.go:44–47, 66–68, 106–109`) returns `200 OK` with zero/empty data, not 501. The handler never emits 501 anywhere.

---

## 12. Open questions

- **`auth_sessions.user_agent` reliability** [runtime-unverified]: The new-device login signal relies on `user_agent` being populated and stable. If clients send varying or absent user agents, the `NOT EXISTS` anti-join produces false positives (every login looks new) or misses genuine new-device logins. The query filters `user_agent IS NOT NULL AND user_agent <> ''` (`repository.go:198–200`), but the quality of user-agent data in production is unverifiable without a live query.

- **`ListOffHoursAdminActions` GROUP BY correctness** [runtime-unverified]: The query uses `MIN(ur.role_code)` to pick one role per event when an actor holds multiple roles. If an actor is both `area_admin` and `qms_admin`, only the alphabetically-first role code appears in `ActorRole`. Whether this is the intended UI display behaviour is not documented.

- **`audit_events.tenant_id` type** [runtime-unverified]: `CountRecentFailedLoginsByUser` and `ListLockouts` cast `$1` as `$1::uuid` (`repository.go:32, 97`), but `ListOffHoursAdminActions` passes `strings.TrimSpace(tenantID)` without a `::uuid` cast (`repository.go:266`) and uses `e.tenant_id = $1` while the JOIN uses `ur.tenant_id = e.tenant_id::uuid` (`repository.go:252`). Whether `audit_events.tenant_id` is `TEXT` (explaining the no-cast) vs `UUID` (matching other tables) cannot be confirmed without inspecting the table DDL under Docker, which is currently unavailable.

- **Signal `DetectedAt` is always `s.now()`** [design question]: All signals generated in a single `ListSignals` call share the same `DetectedAt` timestamp (the wall-clock value captured at `service.go:90`). This means the sort's secondary key (by `DetectedAt`) is always equal within a single request and provides no meaningful ordering among signals of the same severity. Whether this is intentional (simplicity) or an oversight is undocumented.
