# Feature F8.4 — Plan

> Engine: inline (superpowers:writing-plans). Spec: `./spec.md` (approved 2026-06-20).
> Defense-in-depth for CWE-613: revoke-on-deactivate (write path) + fail-closed resolve (read path).

## Files touched

| File | Change |
|------|--------|
| `internal/modules/auth/application/service.go` | `UpdateUser` (`:618`): when `params.IsActive` is being set false, revoke all of the user's sessions — atomic tx (mirror `AdminResetPassword`/`ChangePasswordForUser`: `UpdateUserTx` + `RevokeSessionsByUserIDTx` + Commit) with the in-memory fallback (`UpdateUser` + `RevokeSessionsByUserID`). No authz-recording read inside the tx (H-PRE-1). `buildCurrentUser` (`:845`): after `FindIdentityByUserID`, return `ErrIdentityInactive` when `!identity.IsActive` (fail closed). |
| `internal/modules/auth/delivery/http/middleware.go` | Add `authdomain.ErrIdentityInactive` to the `ResolveSession` 401 branch (`:75`) so a deactivated identity is rejected as 401, not 500. |
| `internal/modules/auth/application/service_test.go` | New tests (in-memory repo, fallback path). |
| `internal/modules/auth/delivery/http/middleware_test.go` | New/extended test: resolve returning `ErrIdentityInactive` → 401 problem+json. |

## Test strategy

- **Class:** unit (in-memory `memory.NewRepository()`, no DB) — the service revoke + fail-closed logic is
  storage-agnostic; the memory repo hits the non-tx fallback branch, which is the branch under test.
- **red→green:**
  1. `TestUpdateUser_DeactivateRevokesSessions`: bootstrap/create an active identity, `CreateSession`,
     assert `ResolveSession` ok → `UpdateUser(IsActive=false)` → `ResolveSession` now rejected
     (`ErrIdentityInactive` or `ErrSessionRevoked`). Fails before the change (session survives).
  2. `TestBuildCurrentUser_FailsClosedWhenInactive` (via `ResolveSession`): identity `IsActive=false` with a
     still-live (un-revoked) session row → `ResolveSession` returns `ErrIdentityInactive`. Proves the
     read-path backstop independent of revoke.
  3. `TestUpdateUser_ActivateOrProfileEdit_DoesNotRevoke`: `UpdateUser` with `IsActive=true` or only
     display-name change does **not** revoke a live session (no over-broad revoke).
  4. Middleware: `ErrIdentityInactive` → 401 `CodeAuthUnauthorized` (not 500).
- **regression:** existing `ChangePasswordForUser` / `AdminResetPassword` revoke tests stay green
  (`go test ./internal/modules/auth/...`).

## Task order

1. Write failing tests (1–4 above).
2. Implement `buildCurrentUser` fail-closed (smallest change; flips tests 2 & the resolve side of 1).
3. Implement `UpdateUser` revoke-on-deactivate (flips test 1 fully; test 3 guards scope).
4. Add `ErrIdentityInactive` to the middleware 401 set (flips test 4).
5. `go build ./...`; `go test -count=1 ./internal/modules/auth/...`; vet.
6. Evidence + commit.

## Risk / rollback

- **Risk:** over-broad revoke (revoking on a non-deactivating update) would log users out spuriously — guarded
  by test 3 (`deactivating := params.IsActive != nil && !*params.IsActive`). Fail-closed at resolve could lock
  out an active user if `IsActive` were mis-read — guarded by test 2 only asserting on the inactive case and
  the existing active-path tests staying green.
- **H-PRE-1:** the revoke tx contains only `UpdateUserTx` + `RevokeSessionsByUserIDTx` (no authz-recording
  read) — same shape as the audited `AdminResetPassword` path.
- Rollback = `git checkout` the 3 source files. No schema/migration change.

## ADR

- None (mirrors the established revoke-on-credential-change pattern; no new session contract). Per spec §ADR.
