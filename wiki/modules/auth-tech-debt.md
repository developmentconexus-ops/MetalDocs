# Tech Debt Register — auth

> Companion to `wiki/modules/auth.md`. Lists known gaps, smells, and missing-ADR items. **Debt only — no fix prescriptions.** Fixes belong in `wiki/backlog/auth-refactor.md`.

**Last verified:** 2026-07-02 (T-008 closed by-design — ADR 0027 schema decision + new ADR 0055 binding session/lookup/cross-tenant-probe rules; closes grade-A SEC-10/DEC-03) | **Prior:** 2026-07-01 (Grade-A simplification register reconciliation — T-012 closed as stale; OriginProtection middleware confirmed wired since commit 5d9f1884, register drift per commit fb0250e5) | **Prior:** 2026-06-12 (Wave 2.12 sync — no debt rows opened or closed; loginCtxPort now panics on nil (was nil-guarded); InMemoryAuthFailureRateLimiter deleted — no new debt opened) | **Prior:** 2026-06-11 (Stage-1 adversarial verification pass — T-004 false-open corrected; T-009 behavior + anchor corrected; T-014 line anchors corrected)

## Severity scale

Triggers per `templates/tech-debt-register.md`. Authn bypass, regulated audit-trail gap, multi-tenant data leak, data-loss path → Critical. Defense-in-depth gap, contract violation with measurable consumer impact, cross-module dep blocking refactors → Major. Latent code surfaces, missing standalone ADRs, undoc'd exports, bidirectional non-circular deps → Minor.

## Items

### T-001 · LegacyHeaderEnabled X-User-Id authn bypass — CLOSED 2026-06-10 (Stage-1 backend audit)
- **Severity:** critical (closed)
- **Surface (resolved):** `internal/modules/auth/delivery/http/middleware.go` — no `LegacyHeaderEnabled` field or `X-User-Id` branch exists; `authapp.Config` (`internal/modules/auth/application/service.go:38-50`) has no `LegacyHeaderEnabled` field. `internal/platform/authn/config.go` does not load or set this field.
- **Observation (original):** When `LegacyHeaderEnabled=true`, middleware accepted an unauthenticated `X-User-Id` header and synthesised a current-user context with no proof of identity. Single-flag compromise granted tier-0 bypass on every protected route.
- **Resolution:** Bypass code and config field removed in commit 554c4007d. `Middleware.Wrap` (`middleware.go:49-88`) now handles exactly two code paths: public route pass-through and cookie-based session resolution — no header bypass branch.
- **Evidence:** `internal/modules/auth/delivery/http/middleware.go:49-88`; `internal/modules/auth/application/service.go:38-50`; `internal/platform/authn/config.go` (no `LegacyHeaderEnabled` load site).
- **Linked backlog row:** `backlog/auth-refactor.md#R-001`
- **Linked ADR:** missing-ADR

### T-002 · Audit-trail gap on identity mutations — CLOSED 2026-05-11 (Plan 6a)
- **Severity:** critical (closed)
- **Surface:** `internal/modules/auth/application/service.go:117-126,279-326,358-397`; `internal/modules/iam/delivery/http/admin_handler.go:259-285`
- **Observation:** Login, logout, password-change, admin password-reset, and `CreateUser` emit no audit-sink record. Only `log.Printf` (`handler.go:56,112`) is wired. `handleCreateUser` does NOT call `recordAudit` even though its sibling `handleReplaceUserRoles` does (`admin_handler.go:398`). Identity is a regulated/QMS surface under ISO 9001 controls. Trigger fired: regulated audit-trail gap.
- **Evidence:** `_artifacts/02-flow-login.md` §audit; `_artifacts/02-flow-create-user.md` §6; `_artifacts/03-deps.md` §5.
- **Linked backlog row:** `backlog/auth-refactor.md#R-002`
- **Linked ADR:** missing-ADR

### T-003 · Legacy error envelope (RFC 9457 drift) — CLOSED 2026-05-12 (Plan 7)
- **Severity:** major (closed)
- **Surface (resolved):** `internal/modules/auth/delivery/http/handler.go:141-158` (`writeAuthError`) — every branch calls `problem.Write(w, problem.New(...))`. `handler.go:58,102,115,121,126,131` — inline error paths use `problem.Write` directly. `internal/modules/auth/delivery/http/middleware.go:66,73,76,80` — all error branches use `problem.Write`. `writeAuthError` takes no `traceID` parameter (`handler.go:165`); the parameter was removed in Plan 7 (was a noop).
- **Observation (original):** Auth emitted `{error:{code,message,details,trace_id}}` instead of `application/problem+json`.
- **Evidence:** `_artifacts/05-industry.md` IP-001.
- **Linked backlog row:** `backlog/auth-refactor.md#R-003` (merged Plan 7 2026-05-11, commit `95ebedfc`)
- **Linked ADR:** `wiki/architecture/api-design-system.md`

### T-004 · CreateUser two-transaction non-atomicity — CLOSED 2026-05-13 (Plan 9r)
- **Severity:** major (closed)
- **Surface (resolved):** `internal/modules/auth/application/service.go:485-508` (shared-tx path); `internal/modules/auth/application/service.go:114-124` (interface definitions `createUserTxRepository`, `replaceUserRolesTxRepository`, `beginTxRepository`); `internal/modules/auth/infrastructure/postgres/repository.go:396` (`CreateUserTx` implementation); `internal/modules/iam/infrastructure/postgres/role_admin_repository.go:20-21,94` (`BeginTx` + `ReplaceUserRolesTx` implementations).
- **Observation (original):** `Service.CreateUser` called `repo.CreateUser` (own `BeginTx` → INSERT `auth_identities` → COMMIT) then `roleAdmin.ReplaceUserRoles` (own `BeginTx` → UPSERT `iam_users`, DELETE+INSERT `iam_user_roles` → COMMIT). No outer transaction. If TX-B failed after TX-A committed, `auth_identities` row was orphaned with no role binding. Recovery was manual. Trigger fired: data-loss-adjacent path (orphan rows on partial failure).
- **Resolution:** `Service.CreateUserWithInput` (`service.go:470`) now asserts three interfaces at runtime. When all three are satisfied — `createUserTxRepository` (auth postgres repo), `replaceUserRolesTxRepository` (IAM postgres repo), and `beginTxRepository` (auth postgres repo) — a single `*sql.Tx` is opened, both `CreateUserTx` and `ReplaceUserRolesTx` execute inside it, and a single `Commit` closes it (`service.go:488-508`). Both postgres repositories implement the required interfaces, making this the canonical production path. The two-tx fallback at `service.go:511-514` is retained for test/in-memory implementations that do not satisfy the Tx interfaces.
- **Evidence:** `internal/modules/auth/application/service.go:114-124` (interface defs); `internal/modules/auth/application/service.go:485-508` (shared-tx path); `internal/modules/auth/infrastructure/postgres/repository.go:396` (`CreateUserTx`); `internal/modules/iam/infrastructure/postgres/role_admin_repository.go:20-21,94` (`BeginTx`, `ReplaceUserRolesTx`). Fix merged commit `58a71b5aa` 2026-05-13.
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
- **Surface:** `internal/modules/auth/infrastructure/postgres/repository.go:103`
- **Observation:** Every authenticated request issues an `UPDATE auth_sessions SET last_seen_at = now()` via `TouchSession`. At sustained QPS this is one row-write per request on the hot session row. Latent perf concern; no caller currently observes pressure. Trigger fired: latent (surface exists, not yet a bug).
- **Evidence:** `_artifacts/02-flow-resolve-session.md` §middleware chain; `_artifacts/04-persistence.md` §auth_sessions.
- **Linked backlog row:** `backlog/auth-refactor.md#R-006`
- **Phase 8 status:** CLOSED 2026-05-25. `TouchSession` now updates `auth_sessions.last_seen_at` only outside a 30-second grace window and treats in-window rows as successful no-ops; missing sessions still return `ErrSessionNotFound`.
- **Linked ADR:** `wiki/decisions/0007-two-tier-authz.md`

### T-007 · Auth↔IAM bidirectional dependency — Role slice CLOSED 2026-07-02 (ARC-06), remainder OPEN
- **Severity:** minor
- **Surface:** `internal/modules/auth/{application/service.go, delivery/http/middleware.go, domain/model.go, infrastructure/memory/repository.go, infrastructure/postgres/repository.go}`; `internal/modules/iam/{application/people_service.go, delivery/http/admin_handler.go, delivery/http/middleware.go, delivery/http/people_handler.go, delivery/http/sessions_handler.go}`
- **Observation (original):** auth imports `iamdomain` (Role, RoleProvider, RoleAdminRepository, ErrUserNotFound, etc.). iam imports `authdomain` (ManagedUser, OnlineUser, UpdateUserParams, CurrentUserFromContext, error sentinels). Non-circular today (different sub-packages on each side) but detangling either side requires the other to absorb a contract. Trigger fired: bidirectional dep that is non-circular today but blocks clean refactor.
- **ARC-06 disposition (2026-07-02):** re-grepped before acting — confirmed current and broader than documented (iam-side spread across 5 files, not just `admin_handler.go`; auth-side across 8 files counting tests). `Role` (the enum + its constants + Role-only helpers `IsValidRole`/`IsAreaRole`/`AreaRoles`) moved to a new neutral leaf package `internal/platform/iamtypes`; `iam/domain.Role` is now a type alias (`type Role = iamtypes.Role`) so all ~65 repo-wide call sites keep compiling unchanged (alias chosen over full mechanical rename — 78+ call sites across 14 modules was too large a surface for this bounded fix; alias is byte-identical, not a workaround, since Go aliases are the same type). `internal/modules/auth/**` no longer imports `iam/domain` for `Role`/`RoleSystemAdmin`/`RoleViewer`/`RoleEditor`/`IsValidRole` — those are `iamtypes.*` now. Grep proof: zero `iamdomain\.Role` / `iamdomain\.RoleXxx` matches remain under `internal/modules/auth/**`.
- **Residual (still OPEN, NOT closed by ARC-06):** auth still legitimately imports `iam/domain` for `Capability`, `RoleProvider`, `RoleAdminRepository`, `LoginContextPort`, `UserTenantReader`/`NoopUserTenantReader`, `WithAuthContext`, and 4 error sentinels (`ErrInvalidRole`, `ErrUserNotFound`, `ErrUserInactive`, `ErrNoRolesAssigned`) — these are IAM application ports/contracts, not the Role enum, and moving them is a materially larger refactor outside ARC-06's stated action ("move the Role enum"). iam still imports `auth/domain` (`ManagedUser`, `OnlineUser`, etc.) unchanged. The bidirectional module-to-module edge therefore still exists; only the Role-enum slice of it (T-010's exact scope) is closed. A follow-on finding should be opened if the remaining port/capability coupling is to be addressed.
- **Evidence:** `_artifacts/03-deps.md` §1, §2 (original); ARC-06 grep proof 2026-07-02 (this closure).
- **Linked ADR:** `wiki/decisions/0007-two-tier-authz.md`
- **Linked backlog row:** `backlog/auth-refactor.md#R-007`

### T-008 · auth_identities lacks tenant_id (latent multi-tenant identity sharing) — CLOSED 2026-07-02 (by-design, ADR 0027 + ADR 0055)
- **Severity:** minor (closed)
- **Surface (resolved):** `db/baseline/0001_current_schema.sql:966-982` (`auth_identities`, no `tenant_id`); `db/baseline/0001_current_schema.sql:989-996,3246-3249,4045-4049` (`auth_sessions.tenant_id` + FK + index, added by migration `0184_auth_sessions_tenant_id.sql`); `internal/modules/auth/application/service.go:344-366` (`resolveLoginTenant` fixes tenant at login from verified `iam_users` membership); `internal/modules/auth/application/service.go:368-400` (`ResolveSession` reads tenancy off the session row only).
- **Observation (original):** `auth_identities` carries no `tenant_id` column — identity is tenant-global. **Partially resolved (2026-05-11, Plan 3):** `auth_sessions.tenant_id` was added by migration `0184_auth_sessions_tenant_id.sql`; session-bound tenant is now the authoritative source for all downstream handlers. The remaining gap was that `auth_identities` itself had no standalone ADR binding the session/lookup/cross-tenant-probe behavior that makes the global-identity model safe.
- **Resolution:** ADR 0027 already decided (2026-06-13) that `auth_identities` is tenant-global by design — one identity, N tenant memberships via `iam_users.tenant_id`. ADR 0055 (2026-07-02) converts that narrative decision into binding numbered rules verified against current runtime code: sessions MUST be tenant-scoped (rule 1), identity lookups never bypass tenant checks on resources they unlock (rule 2), cross-tenant probes still 404 (rule 3), any future per-tenant identity requirement is a new ADR (rule 4). No schema or code change was required — both ADRs document verified existing behavior.
- **Evidence:** `_artifacts/04-persistence.md` §3 columns; `_artifacts/05-industry.md` IP-008; `wiki/decisions/0027-rls-adoption-sequencing.md`; `wiki/decisions/0055-global-auth-identities.md`.
- **Linked backlog row:** `backlog/auth-refactor.md#R-008` (can be closed)
- **Linked ADR:** `wiki/decisions/0027-rls-adoption-sequencing.md` (schema decision), `wiki/decisions/0055-global-auth-identities.md` (binding rules)

### T-009 · Logout cannot distinguish "no session" from "tampered cookie"
- **Severity:** minor
- **Surface:** `internal/modules/auth/application/service.go:374-384`; `internal/modules/auth/application/service.go:736-744`
- **Observation:** `Service.Logout` (`service.go:374-384`) calls `tokenHashFromCookieValue` (`service.go:736-744`) and propagates its error via `return err` (`service.go:381`). However, `tokenHashFromCookieValue` returns `authdomain.ErrSessionNotFound` for both a structurally malformed cookie (wrong number of parts, empty part — line 739) and an HMAC-mismatched (tampered) cookie (line 742). The error is not discarded — it is returned — but both failure modes surface as the same `ErrSessionNotFound` sentinel. The caller cannot distinguish "session never existed" from "cookie was tampered". No log emission on the tampered-cookie path. Trigger fired: latent (no current caller relies on the distinction).
- **Evidence:** `_artifacts/02-flow-resolve-session.md` §logout sub-flow.
- **Linked backlog row:** `backlog/auth-refactor.md#R-009`
- **Linked ADR:** n/a

### T-010 · Missing standalone ADR for session-cookie + bcrypt + lockout policy
- **Severity:** minor
- **Surface:** `internal/modules/auth/application/service.go:117-126,431-432`; `internal/platform/authn/config.go:101-116`
- **Observation:** Session-cookie format (`<base64url(rand32)>.<base64url(HMAC-SHA256(secret,token))>` with `SHA-256(token)` stored as `session_id`), bcrypt cost 12 (`bcryptCost = 12` at `service.go:29`), and per-account lockout policy are enforced by code + tests but no standalone ADR captures the choice. ADR 0007 covers tier split, not credential mechanics. Trigger fired: missing standalone ADR for an enforced rule.
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

### T-012 · OriginProtection config field unwired — CLOSED 2026-07-01 (register drift — already fixed by commit 5d9f1884)
- **Severity:** minor (closed)
- **Surface (resolved):** `internal/platform/authn/config.go:140-141` — `TrustedOrigins`/`OriginProtection` loaded from env into `authapp.Config`. `internal/platform/security/origin_protection.go:12-45` — `NewOriginProtection`/`OriginProtection.Wrap` implement same-origin enforcement gated on `cfg.Enabled` + session cookie presence. `apps/api/cmd/metaldocs-api/main.go:237-242` — `originProtection := security.NewOriginProtection(...)` wired with `authCfg.OriginProtection`/`authCfg.TrustedOrigins`/`authCfg.TrustedProxyCIDRs`; `main.go:643` — `originProtection.Wrap` mounted in the request middleware chain (`apiChain(...)`, between `cors.Wrap` and `loginRateLimit`).
- **Observation (original):** `OriginProtection` and `TrustedOrigins` were loaded into `authapp.Config` but no middleware path read them to enforce same-origin / CSRF protection on state-changing routes.
- **Resolution:** Middleware wiring landed in commit `5d9f1884` ("fix(api): Phase D review fixes — missed rate limiter, canonical origin code, RFC-correct Problem, harder lint", 2026-06-07) — `security/origin_protection.go` predates that commit and was already mounted at `main.go:643` at that point. This register row was stale; confirmed via commit `fb0250e5` (SEC-07 note: "middleware was already wired at main.go:643 since 5d9f1884; auth-tech-debt T-012 is register drift") and independently re-verified against current code 2026-07-01.
- **Evidence:** `internal/platform/security/origin_protection.go:1-45`; `apps/api/cmd/metaldocs-api/main.go:237-242,643`; commits `5d9f1884`, `fb0250e5`.
- **Linked backlog row:** `backlog/auth-refactor.md#R-012` (can be closed)
- **Linked ADR:** missing-ADR

### T-014 · FE 401 interceptor conflated session-expiry with domain-401 — CLOSED 2026-05-28 (qa/fe-401-interceptor)
- **Severity:** major (closed)
- **Surface:** `frontend/apps/web/src/lib/api/client.ts` (`assertApiResponse`); consumer `frontend/apps/web/src/features/auth/useAuthSession.ts`
- **Observation (original):** `assertApiResponse` treated **every** 401 as session-expiry — it called `dispatchAuthExpired(...)` and threw `ApiError.fromLegacy("authn.expired", 401, "Sessão expirada")` **before** parsing the `application/problem+json` body, discarding the RFC 9457 `code`. Backend already discriminates: wrong current password → `AUTH_INVALID_CREDENTIALS` (`handler.go:165` via `writeAuthError`), genuine unauthenticated/expired session → `AUTH_UNAUTHORIZED` (`middleware.go:62,69`, `handler.go:126,139`). The FE collapse meant a wrong-current-password during forced change surfaced "Sessão expirada" instead of a credential error (found in qa/auth-password-change as F-A; the consumer-level patch there mapped bare status 401 and masked the real interceptor root). Trigger fired: contract violation with measurable consumer impact.
- **Fix:** interceptor now parses the problem first and only dispatches `authExpired` + throws `authn.expired` when `res.status === 401 && (!problem || problem.code === "AUTH_UNAUTHORIZED")`. Domain 401s keep their problem `code`. Consumers (`handleLogin`, `handleChangePassword`) map by `codeOf(err) === 'AUTH_INVALID_CREDENTIALS'` instead of bare status. Regression: `src/lib/api/client.test.ts` (3 cases), `src/features/auth/useAuthSession.test.tsx` (code-based + session-expiry fall-through). Live Preview proof: forced change w/ wrong current pw → `[role=alert]` "Senha atual incorreta.", stays on form, no logout.
- **Evidence:** backend `internal/modules/auth/delivery/http/{handler.go:165` (`writeAuthError` definition), `handler.go:126,139` (`AUTH_UNAUTHORIZED` write sites in `handleMe` and `handleChangePassword`), `handler.go:168,170` (`AUTH_INVALID_CREDENTIALS` write sites inside `writeAuthError`), `middleware.go:62,69}`; FE `src/lib/api/client.ts:51-65` (`assertApiResponse` — problem parsed at line 53, session-expiry branch at lines 58-61, domain error at 63-65).
- **Linked ADR:** `wiki/architecture/api-design-system.md` (RFC 9457); `wiki/concepts/error-ux.md`

### T-013 · Login form `noValidate` makes `required` inert (FE)
- **Severity:** minor (low)
- **Surface:** `frontend/apps/web/src/features/auth/pages/LoginPage.tsx` (`<form noValidate>`, inputs carry inert `required`)
- **Observation:** The login form sets `noValidate`, so the `required` attributes on identifier/password never trigger native empty-field validation. Submitting empty fields posts to the server, which correctly rejects with a generic, non-enumerating message — so there is no security gap, only a missing client-side early-exit. Found during FE login screen QA (qa/auth-login).
- **Evidence:** Preview-driven QA on `/login`; empty submit reaches `POST /auth/login`.
- **Linked ADR:** n/a

---

## Coverage stats (computed at compose time)

- Public symbols undocumented: 98 / 98
- Operations missing C4 placement: 0 / 4
- Cross-deps missing in §5/§8: 0 / 16
- State transitions missing in §6: 0 / 3
- Decisions without ADR link: 7
