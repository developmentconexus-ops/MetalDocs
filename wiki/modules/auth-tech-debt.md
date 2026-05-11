# Tech Debt Register — auth

> Companion to `wiki/modules/auth.md`. Lists known gaps, smells, and missing-ADR items. **Debt only — no fix prescriptions.** Fixes belong in `wiki/backlog/auth-refactor.md`.

**Last verified:** 2026-05-11

## Severity scale

Triggers per `templates/tech-debt-register.md`. Authn bypass, regulated audit-trail gap, multi-tenant data leak, data-loss path → Critical. Defense-in-depth gap, contract violation with measurable consumer impact, cross-module dep blocking refactors → Major. Latent code surfaces, missing standalone ADRs, undoc'd exports, bidirectional non-circular deps → Minor.

## Items

### T-001 · LegacyHeaderEnabled X-User-Id authn bypass
- **Severity:** critical
- **Surface:** `internal/modules/auth/delivery/http/middleware.go:58-61`
- **Observation:** When `LegacyHeaderEnabled=true`, middleware accepts unauthenticated `X-User-Id` header and synthesises a current-user context with no proof of identity. Single-flag compromise grants tier-0 bypass on every protected route. Fact, not hypothetical: code path is live; only the env-var gate prevents abuse. Trigger fired: authn bypass.
- **Evidence:** `_artifacts/02-flow-resolve-session.md` §middleware chain; `_artifacts/05-industry.md` IP-004.
- **Linked backlog row:** `backlog/auth-refactor.md#R-001`
- **Linked ADR:** missing-ADR

### T-002 · Audit-trail gap on identity mutations
- **Severity:** critical
- **Surface:** `internal/modules/auth/application/service.go:117-126,279-326,358-397`; `internal/modules/iam/delivery/http/admin_handler.go:259-285`
- **Observation:** Login, logout, password-change, admin password-reset, and `CreateUser` emit no audit-sink record. Only `log.Printf` (`handler.go:56,112`) is wired. `handleCreateUser` does NOT call `recordAudit` even though its sibling `handleReplaceUserRoles` does (`admin_handler.go:398`). Identity is a regulated/QMS surface under ISO 9001 controls. Trigger fired: regulated audit-trail gap.
- **Evidence:** `_artifacts/02-flow-login.md` §audit; `_artifacts/02-flow-create-user.md` §6; `_artifacts/03-deps.md` §5.
- **Linked backlog row:** `backlog/auth-refactor.md#R-002`
- **Linked ADR:** missing-ADR

### T-003 · Legacy error envelope (RFC 9457 drift)
- **Severity:** major
- **Surface:** `internal/modules/auth/delivery/http/handler.go:166`; `internal/modules/auth/delivery/http/middleware.go:65,76,79,83`
- **Observation:** Auth emits `{error:{code,message,details,trace_id}}` instead of `application/problem+json` (RFC 9457). Mirrors documented drift in iam T-006 and documents T-001. Consumer tooling that relies on `type`/`title`/`status`/`detail`/`instance` fields cannot parse auth errors uniformly. Trigger fired: contract violation with measurable consumer impact.
- **Evidence:** `_artifacts/05-industry.md` IP-001.
- **Linked backlog row:** `backlog/auth-refactor.md#R-003`
- **Linked ADR:** missing-ADR

### T-004 · CreateUser two-transaction non-atomicity
- **Severity:** major
- **Surface:** `internal/modules/auth/application/service.go:305,325`; `internal/modules/auth/infrastructure/postgres/repository.go:151-170`; `internal/modules/iam/infrastructure/postgres/role_admin_repository.go:73-112`
- **Observation:** `Service.CreateUser` calls `repo.CreateUser` (own `BeginTx` → INSERT `auth_identities` → COMMIT) then `roleAdmin.ReplaceUserRoles` (own `BeginTx` → UPSERT `iam_users`, DELETE+INSERT `iam_user_roles` → COMMIT). No outer transaction. If TX-B fails after TX-A commits, `auth_identities` row is orphaned with no role binding. Recovery is manual. Trigger fired: data-loss-adjacent path (orphan rows on partial failure).
- **Evidence:** `_artifacts/02-flow-create-user.md` §2 transaction-boundary fact.
- **Linked backlog row:** `backlog/auth-refactor.md#R-004`
- **Linked ADR:** missing-ADR

### T-005 · Login endpoint missing IP-based rate limit
- **Severity:** major
- **Surface:** `internal/modules/auth/application/service.go:117-126`; `internal/modules/auth/delivery/http/handler.go:46-83`
- **Observation:** Auth enforces per-account bcrypt + lockout (`LoginMaxFailedAttempts=5`, `LoginLockDuration=15m`) at identity layer. No upstream IP-based rate limit on `/api/v1/auth/login`. Distributed credential-spray across many accounts is not throttled. Defense-in-depth single-layer (NIST SP 800-95 §4.3). Trigger fired: defense-in-depth gap on a sensitive mutation.
- **Evidence:** `_artifacts/05-industry.md` IP-004; `_artifacts/02-flow-login.md` §lockout policy.
- **Linked backlog row:** `backlog/auth-refactor.md#R-005`
- **Linked ADR:** missing-ADR

### T-006 · TouchSession write amplification per request
- **Severity:** minor
- **Surface:** `internal/modules/auth/infrastructure/postgres/repository.go:80`
- **Observation:** Every authenticated request issues an `UPDATE auth_sessions SET last_seen_at = now()` via `TouchSession`. At sustained QPS this is one row-write per request on the hot session row. Latent perf concern; no caller currently observes pressure. Trigger fired: latent (surface exists, not yet a bug).
- **Evidence:** `_artifacts/02-flow-resolve-session.md` §middleware chain; `_artifacts/04-persistence.md` §auth_sessions.
- **Linked backlog row:** `backlog/auth-refactor.md#R-006`
- **Linked ADR:** `wiki/decisions/0007-two-tier-authz.md`

### T-007 · Auth↔IAM bidirectional dependency
- **Severity:** minor
- **Surface:** `internal/modules/auth/{application/service.go:18, delivery/http/middleware.go:10, domain/model.go:6, infrastructure/memory/repository.go:10}`; `internal/modules/iam/delivery/http/admin_handler.go:13`; `internal/modules/iam/delivery/http/middleware.go:8`
- **Observation:** auth imports `iamdomain` (Role, RoleProvider, RoleAdminRepository, ErrUserNotFound, etc.). iam imports `authdomain` (ManagedUser, OnlineUser, UpdateUserParams, CurrentUserFromContext, error sentinels). Non-circular today (different sub-packages on each side) but detangling either side requires the other to absorb a contract. Trigger fired: bidirectional dep that is non-circular today but blocks clean refactor.
- **Evidence:** `_artifacts/03-deps.md` §1, §2.
- **Linked ADR:** `wiki/decisions/0007-two-tier-authz.md`
- **Linked backlog row:** `backlog/auth-refactor.md#R-007`

### T-008 · Auth tables lack tenant_id (latent multi-tenant identity sharing)
- **Severity:** minor
- **Surface:** `migrations/0021_init_auth_identities_and_sessions.sql`; `migrations/0036_decouple_auth_identity_from_iam_user_tables.sql`
- **Observation:** `auth_identities` and `auth_sessions` carry no `tenant_id` column. Identity is tenant-global. IAM tables added `tenant_id` in 0130/0162 to enforce row-level isolation; auth did not. Latent because today every tenant is on the same identity pool by design (single-tenant install). Becomes a multi-tenant data-leak vector if multi-tenancy is enabled without backfill. Trigger fired: latent (surface exists, no caller hits it under current single-tenant deployment).
- **Evidence:** `_artifacts/04-persistence.md` §3 columns; `_artifacts/05-industry.md` IP-008.
- **Linked backlog row:** `backlog/auth-refactor.md#R-008`
- **Linked ADR:** missing-ADR

### T-009 · Logout swallows malformed-cookie error silently
- **Severity:** minor
- **Surface:** `internal/modules/auth/application/service.go:198-201`
- **Observation:** `Service.Logout` discards the error from `parseAndVerifyToken` when the cookie is malformed and returns nil. Caller cannot distinguish "no session existed" from "cookie was tampered". No log emission. Trigger fired: latent (no current caller relies on the distinction).
- **Evidence:** `_artifacts/02-flow-resolve-session.md` §logout sub-flow.
- **Linked backlog row:** `backlog/auth-refactor.md#R-009`
- **Linked ADR:** n/a

### T-010 · Missing standalone ADR for session-cookie + bcrypt + lockout policy
- **Severity:** minor
- **Surface:** `internal/modules/auth/application/service.go:117-126,431-432`; `internal/platform/authn/config.go:101-116`
- **Observation:** Session-cookie format (`<base64url(rand32)>.<base64url(HMAC-SHA256(secret,token))>` with `SHA-256(token)` stored as `session_id`), `bcrypt.DefaultCost`, and per-account lockout policy are enforced by code + tests but no standalone ADR captures the choice. ADR 0007 covers tier split, not credential mechanics. Trigger fired: missing standalone ADR for an enforced rule.
- **Evidence:** `_artifacts/02-flow-login.md` §token mint + verify; `_artifacts/03-deps.md` §4 config surface.
- **Linked backlog row:** `backlog/auth-refactor.md#R-010`
- **Linked ADR:** missing-ADR

### T-011 · Exported symbols undocumented
- **Severity:** minor
- **Surface:** `internal/modules/auth/{application,delivery/http,domain,infrastructure/{memory,postgres}}/*.go`
- **Observation:** All 98 exported symbols enumerated in `_artifacts/01-surface.md` lack Go doc comments. Trigger fired: missing Go doc comments on exported symbols.
- **Evidence:** `_artifacts/01-surface.md` §exports.
- **Linked backlog row:** `backlog/auth-refactor.md#R-011`
- **Linked ADR:** n/a

### T-012 · OriginProtection config field unwired
- **Severity:** minor
- **Surface:** `internal/platform/authn/config.go:116`; `internal/modules/auth/delivery/http/middleware.go`
- **Observation:** `OriginProtection` and `TrustedOrigins` are loaded into `authapp.Config` but no middleware path reads them to enforce same-origin / CSRF protection on state-changing routes. Latent: surface exists, no enforcement. Trigger fired: latent (surface exists, no caller hits it).
- **Evidence:** `_artifacts/03-deps.md` §4 config surface.
- **Linked backlog row:** `backlog/auth-refactor.md#R-012`
- **Linked ADR:** missing-ADR

---

## Coverage stats (computed at compose time)

- Public symbols undocumented: 98 / 98
- Operations missing C4 placement: 0 / 4
- Cross-deps missing in §5/§8: 0 / 16
- State transitions missing in §6: 0 / 3
- Decisions without ADR link: 8
