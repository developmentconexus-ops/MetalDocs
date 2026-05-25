# Sync log — auth

> Append-only log of `metaldocs-module-doc-sync` runs against this module. Newest at top.

## 2026-05-25 - Phase 8 auth mediums

- **Context:** uncommitted diff on `fix/phase8-auth-mediums` - Worker 8A fixes for auth-M4/M8/M10/M11/M13/M17/M18/M20.
- **Mode:** structural refresh, bounded to auth audit/session/cookie/redaction behavior.
- **Affected surface scan:** `internal/modules/auth/delivery/http/handler.go`, `application/service.go`, `domain/model.go`, `domain/port.go`, `infrastructure/postgres/repository.go`, `infrastructure/memory/repository.go`, auth tests, auth module doc/debt/backlog/login and resolve-session artifacts.
- **Facts updated:** failed login audit emits `auth.login.failed` with hashed identifier; audit trace IDs are validated with UUID fallback; audit payload marshal failure uses sentinel JSON; session cookies use `SameSite=Strict`; `AuthenticatedSession` redacts raw tokens; `GetUserTenants` orders tenant IDs; `TouchSession` uses a 30-second grace window.
- **T/R rows touched:** T-006 marked closed by Phase 8 note; R-006 noted as merged by Phase 8.
- **Preflight/tally:** preflight attempted; Git Bash tally failed before doc edits with Windows `CreateFileMapping` error 5.
- **Patched files:** `wiki/modules/auth.md`, `wiki/modules/auth-tech-debt.md`, `wiki/backlog/auth-refactor.md`, `wiki/modules/auth/_artifacts/02-flow-login.md`, `wiki/modules/auth/_artifacts/02-flow-resolve-session.md`, `wiki/modules/auth/_artifacts/sync-log.md`.

## 2026-05-11 · Plan 6a — close T-002

- **Context:** Plan 6a (commits 27c19011 + f27529e8) · wire audit writer in auth handler; emit on login/logout/password-change/createUser
- **Anchors moved:** 0
- **Symbols renamed:** 0
- **T-NNN closed:** T-002 · evidence: handler.go now has WithAudit setter + recordAudit helper; emit calls added in handleLogin, handleLogout, handleChangePassword; handleCreateUser emits via iam admin handler
- **R-NNN updated:** R-002 → merged · commits 27c19011 + f27529e8
- **§11 counts after:** Critical=2 Major=3 Minor=7 (unchanged)
- **Tally gate:** PASS
- **Patched files:** wiki/modules/auth-tech-debt.md · wiki/backlog/auth-refactor.md
- **Structural changes noted (sweep needed):** WithAudit setter added to Handler; new auditdomain OUT-edge from auth handler — §5 Key Files + §8 cross-deps not yet updated
