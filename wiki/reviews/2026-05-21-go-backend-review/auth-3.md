# Module 3 — `internal/modules/auth/`

**Reviewed:** 2026-05-22
**Reviewers (ECC, Sonnet 4.6):** go-reviewer, security-reviewer, silent-failure-hunter, type-design-analyzer, database-reviewer
**LoC (prod):** 1899 across `application` / `delivery` / `domain` / `infrastructure/{memory,postgres}`.

## Severity counts (deduped)

| Critical | High | Medium | Low |
|----------|------|--------|-----|
| 8        | 18   | 22     | 11  |

Highest-impact module so far. Multiple Criticals are auth-bypass-grade.

---

## Critical

### C1 — `application/service.go:147` — Swallowed `RecordFailedLogin` error → unbounded brute-force
**Lenses:** go-reviewer, silent-failure-hunter
**Problem:** `_ = s.repo.RecordFailedLogin(ctx, identity.UserID, attempts, lockedUntil)` discards write failure. If DB transient, lockout state never advances. Attacker brute-forces credentials with no lock sticking.
**Recommend:** Propagate error and fail-closed:
```go
if err := s.repo.RecordFailedLogin(...); err != nil {
    return authdomain.AuthenticatedSession{}, fmt.Errorf("record failed login: %w", err)
}
```
If timing-uniform response required, log structured warn and still return `ErrInvalidCredentials` — never bare `_`.

### C2 — `delivery/http/middleware.go:59` — `X-User-Id` legacy header bypass
**Lens:** security-reviewer
**Problem:** When `LegacyHeaderEnabled=true`, any caller setting non-empty `X-User-Id` bypasses cookie/HMAC/revocation/expiry/tenant. No allowlist. No `CurrentUser` injection — downstream sees `ok=false` and may apply weaker logic. Full auth bypass on any tenant resource.
**Recommend:** Remove legacy bypass. If transitional, resolve header against a real session, gate behind IP allowlist, inject full `CurrentUser` context, hard sunset date logged at startup.

### C3 — `infrastructure/postgres/repository.go:163` + `service.go:140-148` — Lockout-counter TOCTOU
**Lenses:** security-reviewer, database-reviewer
**Problem:** Service reads `FailedLoginAttempts`, increments in Go, writes back. Two concurrent attempts both read same N, both write N+1. Lockout threshold skippable. Distinct from C1 (race vs error path).
**Recommend:** Atomic DB increment with lockout in SQL:
```sql
UPDATE metaldocs.auth_identities
SET failed_login_attempts = failed_login_attempts + 1,
    locked_until = CASE WHEN failed_login_attempts + 1 >= $2
                        THEN NOW() + ($3 * INTERVAL '1 second')
                        ELSE NULL END,
    updated_at = NOW()
WHERE user_id = $1
RETURNING failed_login_attempts, locked_until;
```

### C4 — `infrastructure/postgres/repository.go:120-131` — `RevokeSession`/`TouchSession` silent on missing row
**Lens:** database-reviewer
**Problem:** Neither checks `RowsAffected`. Replay of deleted session token through `Logout` returns nil. Caller cannot tell whether revocation landed.
**Recommend:** `result.RowsAffected()` check → `authdomain.ErrSessionNotFound` when 0. Apply to both methods.

### C5 — `auth_sessions` schema missing `tenant_id` column (or query carrying it)
**Lens:** database-reviewer
**Problem:** Go `Session` struct + `CreateSession` INSERT carry `tenant_id` ($3) but migration `0021` doesn't define the column per agent read. Runtime INSERT fails with "column tenant_id does not exist". **Needs verification** against canonical `migrations/` set — agent may have read stale.
**Recommend:** Verify migration list. If gap real, add `tenant_id TEXT NOT NULL` + index, scope `FindSession` with `WHERE tenant_id = $tenant` so cross-tenant lookup is structurally impossible.

### C6 — `application/service.go:534-538` + `Config.SessionSecret` — Empty-secret HMAC silently degrades
**Lenses:** silent-failure-hunter, type-design-analyzer
**Problem:** `signToken` uses `[]byte(s.cfg.SessionSecret)` directly. Zero-length key → all session tokens forgeable. `NewService` accepts `Config` with no validation. `_, _ = mac.Write(...)` adds noise but real bug is the missing entropy guard.
**Recommend:** Validate `len(SessionSecret) >= 32` in `NewService` returning `(*Service, error)`. Optionally `type SessionSecret string` with constructor.

### C7 — `domain/model.go:14` — `PasswordHash string` vs plaintext indistinguishable
**Lens:** type-design-analyzer
**Problem:** `PasswordHash string` and plaintext passwords both pass through bare `string` parameters. `CreateUserParams.PasswordHash`, `BootstrapAdminParams.PasswordHash` accept any string — compiler cannot catch a plaintext stored as a hash.
**Recommend:** `type PasswordHash string` with unexported constructor `newPasswordHash(raw []byte) PasswordHash`. Same treatment for `PasswordAlgo` with `const AlgoBcrypt PasswordAlgo = "bcrypt"`. Apply across `CreateUserParams`, `BootstrapAdminParams`, `Identity`.

### C8 — `infrastructure/postgres/repository.go:77,380,396,411` — `err == sql.ErrNoRows` direct equality
**Lens:** go-reviewer
**Problem:** Direct comparison bypasses wrapping. Future driver/middleware that wraps errors silently breaks `FindSession`, `loadIdentity`, `ensureUniqueIdentity` (both branches).
**Recommend:** `errors.Is(err, sql.ErrNoRows)` at all four sites.

---

## High

| ID | Location | Problem | Recommend |
|----|----------|---------|-----------|
| H1 | `service.go:504-509` | `bcrypt.DefaultCost` (10) — below 2024+ bar for regulatory system | Raise to 12; store algo+cost on `password_algo` for future migrations |
| H2 | `service.go:125-126` (+ matching trim in `ChangePasswordForUser`/`CreateUser`) | `strings.TrimSpace(password)` before bcrypt → `" secret "` and `"secret"` collide; password surface silently widened | Never mutate raw password; trim only for policy validation |
| H3 | `handler.go:64` (+ `service.go:63-66,119-121`) | `log.Printf("auth login failed for %q: %v", req.Identifier, err)` writes PII + DB error string → log oracle + PII leak | Log only opaque event code + classified category; no raw identifier; no raw err |
| H4 | `postgres/repository.go:332-361` | `BootstrapAdmin` always returns `true` (ON CONFLICT DO UPDATE overwrites hash on every restart). Diverges from memory adapter | `INSERT ... ON CONFLICT DO NOTHING`; check `RowsAffected==1` for `created` |
| H5 | `service.go:152` | `X-Tenant-ID` header client-controlled at login; `resolveLoginTenant` validates membership but threat model unspecified | Document; consider post-auth tenant selection step so CSRF/origin protections apply |
| H6 | `service.go:534-538` | `signToken` HMAC `_, _ = mac.Write(...)` (related to C6 — keeping High for the discard + missing startup secret guard distinct from min-length) | Drop `_, _`; add `cfg.SessionSecret != ""` guard in `NewService` (paired with C6 length check) |
| H7 | `postgres/repository.go:183,291,337` + `service.go:381` | `defer func() { _ = tx.Rollback() }()` — rollback failure on real error path invisible | Log when `rbErr != nil && !errors.Is(rbErr, sql.ErrTxDone)` |
| H8 | `handler.go:85-89` | `Logout` error swallowed; audit only on success → revocation failure undetected, no audit of attempt | Log error; audit the *attempt* regardless of outcome |
| H9 | `service.go:373-398` | `CreateUser` silent non-atomic fallback when type-assertions fail. Memory adapter permanently non-tx | Require `TransactionalUserCreator` interface in `NewService`; fail fast if Postgres path lacks tx capability; mark memory path test-only |
| H10 | `delivery/http/handler.go:20-21` | `Handler` holds concrete `*authapp.Service` — delivery pinned to application impl; untestable | Extract `AuthService` interface (mirrors `audit.Writer` pattern) |
| H11 | `service.go:259-261` (`ChangePassword`) | Double `strings.TrimSpace(userID)` (dead) + `MustChangePassword` always false on the wrapper → forced-reset semantics broken | Either drop the wrapper or populate `MustChangePassword` from loaded identity before delegating |
| H12 | `domain/model.go:10,27,29` | `UserID`, `SessionID`, `TenantID` raw `string`; 14-method `Repository` interface transposable by position | `type UserID string` / `SessionID` / `TenantID`; ripple through `port.go` and `service.go` signatures |
| H13 | `domain/model.go:9-24` | `Identity` zero value = active, unlocked, no-password user — usable if uninitialized through error path | Unexported `identity` + `NewIdentity(...)` constructor returning `(Identity, error)` |
| H14 | `service.go:28` (`Config.BootstrapAdminPassword`) | Plaintext password in long-lived struct; `%+v` leaks; no zeroing after hash | `String()`/`MarshalJSON` redactors; zero `cfg.BootstrapAdminPassword = ""` after `hashPassword` |
| H15 | `service.go:329` | `CreateUser` 9 positional `string` params — transposition silent | `CreateUserInput` struct with named fields |
| H16 | `postgres/repository.go:25-33` | `FindIdentityByIdentifier` uses `lower(COALESCE(i.email,''))` — partial index `uq_auth_identities_email_ci WHERE email IS NOT NULL` not sargable; hot login path is seq scan | Rewrite WHERE: `lower(username)=lower($1) OR (email IS NOT NULL AND lower(email)=lower($1))` |
| H17 | `postgres/repository.go:163,148` | `RecordFailedLogin` uses `NOW()` (DB-side) but `RecordSuccessfulLogin` uses `$2` (Go-side). Inconsistent time authority under clock skew | Standardize on Go-side `$timestamp` parameter |
| H18 | `postgres/repository.go:210-216` + `service.go:303-319` | `ListUsers` no `tenant_id` filter, no `deleted_at`/`is_active` — cross-tenant identity leak (usernames, emails, lockout state); plus N+1 role lookup | Add `tenant_id` param + `iam_user_roles` join; batch role lookup via `RolesByUserIDs` |
| H19 | `postgres/repository.go:388-418,286-330` | `ensureUniqueIdentity` SELECT-then-INSERT TOCTOU + redundant w/ unique indexes; `UpdateUser` runs two independent UPDATEs sharing the TOCTOU | Drop `ensureUniqueIdentity`; catch `pq.Error.Code == "23505"` → `ErrUserAlreadyExists`; collapse `UpdateUser` to single COALESCE/CASE update |

---

## Medium

| ID | Location | Problem | Recommend |
|----|----------|---------|-----------|
| M1 | `service.go:322-327` | Nil-receiver guard on value method → hides wiring bug | Remove guard; fix construction |
| M2 | `domain/errors.go:18-21` | Two tenant errors use `"auth: ..."` with em-dash; rest use `"auth <noun>"` | Normalize to package style |
| M3 | `service.go:329-398` | `CreateUser` 70 lines, mixed levels | Extract `normaliseCreateUserInput` + `createUserTx` helpers |
| M4 | `handler.go:162-188` | `recordAudit` only on login success; `auth.login.failed` unaudited → brute-force trail empty | Audit failure with hashed identifier + category |
| M5 | `postgres/repository.go:286-330` | `UpdateUser` opens tx even for single-branch updates | Conditional tx, or single combined UPDATE |
| M6 | `service.go:140-148` (race noted in C3); separate note: synchronous critical-section structure | (Same root cause as C3) | (See C3) |
| M7 | `memory/repository.go:285-315` | `GetUserTenants` empty unless `SeedUserTenants` → `AllowDevTenantFallback` always silently taken in tests; dev/prod parity gap | Document; integration tests must exercise Postgres path |
| M8 | `service.go:443-454` | `SameSite: Lax` cookie — Strict gives stronger CSRF for app w/o cross-site embed | `SameSiteStrictMode` unless documented req |
| M9 | `service.go:80-120` | `Config.BootstrapAdminPassword` plaintext in memory for process lifetime (heap dump / stack trace leak) — paired with H14 | Zero after hash; secrets wrapper |
| M10 | `handler.go:162-188` | Client `X-Trace-Id` written raw to audit `TraceID` → log-search poisoning | Validate `[A-Za-z0-9_-]{1,64}`; fall back to server UUID |
| M11 | `handler.go:166` | `raw, _ := json.Marshal(payload)` discards err → audit record with empty payload looks valid | Log err; substitute `{"error":"marshal_failed"}` |
| M12 | `handler.go:57-60,102,115,121` (and similar) | `_ = problem.Write(...)` everywhere — broken writer invisible | Log write errors at warn |
| M13 | `domain/model.go:103-107` | `AuthenticatedSession.RawToken` exported; `%+v` leaks | `String()` + `MarshalJSON` redactors |
| M14 | `domain/model.go:93` (`CurrentUser`) | Principal in context has exported fields, no redactor → `%+v` leaks email/UserID | `String()` redactor |
| M15 | `domain/model.go:60-71,87` | `CreateUserParams.PasswordHash` / `BootstrapAdminParams.PasswordHash` accept any `string` — paired with C7 | Use proposed `PasswordHash` named type |
| M16 | `domain/port.go:8` (whole interface) | All method IDs raw `string` — paired with H12 | Named ID types ripple here |
| M17 | `handler.go:25-33` | `loginRequest`/`changePasswordRequest` plaintext password fields — no redactor → structured logger leak | `String()` redactor per request type |
| M18 | `postgres/repository.go:85-105` | `GetUserTenants` no `ORDER BY` → non-deterministic; auto-tenant-select drifts under plan changes | `ORDER BY tenant_id` |
| M19 | `postgres/repository.go:249-283` | `ListOnlineUsers` mixes Go `activeSince` ($1) with `NOW()` for `expires_at` → clock-skew inconsistency | Pass second `now` param |
| M20 | `postgres/repository.go:107-117` | `TouchSession` fires on every auth request → write amplification on hot table | Touch grace window (e.g. 30s) gated in `ResolveSession` |
| M21 | Migration `0021` vs `repository.go:27,201` | `auth_identities` agent-observed missing `display_name`/`is_active` cols. Verify canonical migration | If gap real: add cols w/ defaults |
| M22 | `service.go:303-319` (N+1, paired with H18) | Per-user `RolesByUserID` loop | Batch `RolesByUserIDs(ctx, []UserID, TenantID)` |

---

## Low

| ID | Location | Note |
|----|----------|------|
| L1 | `service.go:555-559` | `truncate` slices `string` by byte → invalid UTF-8 on multibyte `UserAgent`. Use `[]rune` |
| L2 | `postgres/repository.go:59-83` | `FindSession` does not filter `revoked_at IS NULL AND expires_at > NOW()` at DB layer → revoked-session lookup costs full row | Index-friendly DB-side filter |
| L3 | `middleware.go:97-108` | `POST /api/v1/auth/logout` in `defaultPublicPaths` — wider surface than necessary (defensive at service layer) | Note hardening opportunity |
| L4 | `service.go:249-252` | `Logout` swallows token parse error silently — no signal on cookie tampering | Debug log |
| L5 | `domain/model.go:26` | `Session` has no `IsRevoked()`/`IsExpired()` helpers — repeated logic | Add methods |
| L6 | `handler.go:20-23` | `Handler.audit` nil-default vs no-op writer | Inject no-op default |
| L7 | `domain/context.go:13-18` | `CurrentUserFromContext` returns zero value on miss; ignored `bool` → silent zero principal | `RequireCurrentUser(ctx) (CurrentUser, error)` variant |
| L8 | `postgres/repository.go:120-131` | `RevokeSession` allows double-revocation overwriting timestamp | Add `AND revoked_at IS NULL` (matches `RevokeSessionsByUserID`) |
| L9 | `postgres/repository.go:332-360` | `BootstrapAdmin` wraps single statement in tx | Drop wrapper |
| L10 | Migration `0021` | `idx_auth_sessions_active` (user_id, expires_at DESC) WHERE revoked_at IS NULL — scans wider than needed for bulk-revoke | Add `(user_id) WHERE revoked_at IS NULL` partial index |
| L11 | Memory vs Postgres adapter parity | `RevokeSession`/`RecordFailedLogin` diverge on not-found semantics — tests against memory miss real postgres behavior | Align contract; postgres returns errors on zero rows affected |

---

## Notes / fix-branch reservations

- **8 Criticals** — auth-bypass-grade or data-corruption-grade. All must land before pushing further feature work through this module.
- **Cluster fix branches:**
  - `fix/auth-3-c1-c3-c6` — error swallow + lockout race + HMAC secret (service.go credential pipeline)
  - `fix/auth-3-c2` — middleware legacy bypass removal (standalone, deploy-gated)
  - `fix/auth-3-c4-c5-c8` — repository row-affected + tenant_id + errors.Is hardening
  - `fix/auth-3-c7-types` — PasswordHash / UserID / SessionID / TenantID named types (ripple through port.go + service.go signatures)
- **Land order:** `c2` first (single biggest exposure, smallest diff), then `c1-c3-c6`, then `c4-c5-c8`, then `c7-types`.
- **C5 verification gate:** before committing the fix-branch, run `gh search` / `grep` against `migrations/` to confirm whether `auth_sessions.tenant_id` already exists or is genuinely missing. Agent may have read worktree-stale state.
