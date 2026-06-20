# Feature F8.4 — Evidence (deactivation session enforcement; CWE-613)

> **Milestone:** 8  ·  **Feature:** `f8.4-deactivation-session-enforcement`  ·  **Closed:** 2026-06-20
> **Contract:** `spec.md` (approved 2026-06-20). Plan: `plan.md`. No ADR (mirrors revoke-on-credential-change).
> **Commit:** recorded at commit time below.

## What was implemented (defense-in-depth, two paths)

- **Write path — revoke on deactivate** ([`internal/modules/auth/application/service.go:618`](../../../../../internal/modules/auth/application/service.go)) —
  `UpdateUser` now revokes all of the user's sessions when `params.IsActive` is being set false:
  `deactivating := params.IsActive != nil && !*params.IsActive`. Atomic tx (`UpdateUserTx` +
  `RevokeSessionsByUserIDTx` + `Commit`) when the repo supports it, else the in-memory fallback
  (`UpdateUser` + `RevokeSessionsByUserID`). No authz-recording read inside the tx (**H-PRE-1**) — same
  shape as `AdminResetPassword`. A non-deactivating update keeps the original single-write path.
- **Read path — fail closed at resolve** ([`internal/modules/auth/application/service.go:845`](../../../../../internal/modules/auth/application/service.go)) —
  `buildCurrentUser` returns `authdomain.ErrIdentityInactive` when `!identity.IsActive`, *before* tenant/role
  resolution. Backstop independent of the revoke path: a deactivated identity never yields a `CurrentUser`,
  even if a session row survived. Login (`Authenticate`) already rejects inactive earlier, so this only fires
  on the resolve path.
- **Delivery mapping** ([`internal/modules/auth/delivery/http/middleware.go:75`](../../../../../internal/modules/auth/delivery/http/middleware.go)) —
  `ErrIdentityInactive` added to the `ResolveSession` 401 set, so a deactivated identity is rejected **401**
  (treated as revoked, per the consumer contract), not 500.

## Consumer contract satisfied

After `is_active=false`, any existing session is rejected at the next resolve (401) and a deactivated identity
never yields a `CurrentUser`. Both the revoke-on-write and fail-closed-on-read paths enforce this; either alone
would close the gap, together they are belt-and-suspenders.

## Verification

| Check (spec Validation Gate) | Command / test | Result | Real vs fixture |
|------------------------------|----------------|--------|-----------------|
| Deactivate kills live session | `TestUpdateUser_DeactivateRevokesSessions` (login → ResolveSession ok → `UpdateUser(IsActive=false)` → ResolveSession rejected) | **PASS** (red before impl: "ResolveSession succeeded") | real (in-memory repo, fallback path) |
| Inactive identity fails closed at resolve | `TestResolveSession_FailsClosedWhenIdentityInactive` (repo-side deactivate, session row left live → resolve = `ErrIdentityInactive`) | **PASS** (red before impl: got `<nil>`) | real |
| No over-broad revoke (scope) | `TestUpdateUser_NonDeactivatingEditDoesNotRevoke` (rename + reactivate keep session live) | **PASS** | real |
| Delivery maps inactive → 401 | `TestMiddleware_DeactivatedIdentity_Returns401` (end-to-end real service + middleware) | **PASS** | real |
| Change-password/reset revocation unchanged | `go test ./internal/modules/auth/...` (incl. `TestChangePasswordForUser_RevokesSessions`, AdminReset tests) | all `ok` | real |
| No advisory-lock hazard (H-PRE-1) | review: revoke tx = `UpdateUserTx` + `RevokeSessionsByUserIDTx` only (no authz read) | confirmed | real |
| Static | `go build ./...`; `go vet ./internal/modules/auth/...` | exit 0, no findings | — |

TDD red captured before implementation: deactivate test reported "ResolveSession succeeded"; fail-closed test
reported `got <nil>, want ErrIdentityInactive`. Both flipped green after the two service changes.

## Acceptance vs spec Validation Gate

| Acceptance criterion | Met? | Evidence |
|----------------------|------|----------|
| Deactivate kills live session | yes | `TestUpdateUser_DeactivateRevokesSessions` |
| Inactive identity fails closed at resolve | yes | `TestResolveSession_FailsClosedWhenIdentityInactive` + middleware 401 test |
| Change-password/reset revocation unchanged | yes | full auth suite green |
| No advisory-lock hazard | yes | revoke tx carries no authz-recording read (H-PRE-1) |

## Review disposition

- Spec-compliance review: PASS — both contracted paths implemented; non-goals honored (no token rotation, no
  idle/absolute-TTL change, no new IAM endpoint); `deactivating` guard prevents over-broad revoke.
- Code-quality review: PASS — reuses the established atomic-revoke pattern; idempotent-revoke rationale
  documented; fail-closed placed before tenant/role resolution; sentinel reuses existing `ErrIdentityInactive`.

## Bounded defers

| Defer | Why bounded | Written trigger / owner |
|-------|-------------|-------------------------|
| Force-logout admin endpoint | Explicitly out of scope (PR-7); resolve-time fail-closed already enforces deactivation | Tracked separately as IAM PR-7 |
