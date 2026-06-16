# F0.4 — Self-service ChangePassword emits expired session cookie

> Milestone: M0 (grade-a-completion).
> Status: **Approved — 2026-06-15** (no contract ambiguity; mirror of `handleLogout` cookie pattern).
> Site: `internal/modules/auth/delivery/http/handler.go:132` (`handleChangePassword`).

## Problem

`auth.Service.ChangePasswordForUser` already revokes **all** sessions for the user
(`internal/modules/auth/application/service.go:494`, CWE-613 fix). The HTTP handler
returns `200 OK` but does **not** instruct the browser to drop the now-revoked
session cookie. The browser keeps a dead cookie until it expires naturally; subsequent
requests get 401 with no client-visible state change. Inconsistent with `handleLogout`
(line 115) which sets an expired cookie after revocation.

## Consumer contract (BEFORE → AFTER)

Caller: any authenticated client posting `POST /api/v1/auth/change-password`.

| Field | Before | After |
|---|---|---|
| HTTP status (success) | `200 OK` | `200 OK` (unchanged) |
| Response body | `{changed, user}` | `{changed, user}` (unchanged) |
| `Set-Cookie` header on success | **absent** | `metaldocs_session=; Path=/; HttpOnly; SameSite=Strict; Max-Age=-1` (one cookie, expired) |
| Server-side session revocation | already done in service tx | unchanged |

**Non-goals:**
- No change to `auth.Service` (revocation already correct).
- No change to error paths — only the success path emits the cookie (mirrors `handleLogout`).
- No new audit event (existing `auth.password.changed` in service tx unchanged).
- No change to FE; FE re-login after 401 is the existing contract.

## Validation gate

Acceptance:
1. After a successful self-service change-password, response carries a `Set-Cookie`
   for the session cookie name with `MaxAge < 0` (expired).
2. The existing sessions-revoked behavior in `TestPasswordChangeRevokesSessionAndClearsMustChangePassword`
   stays green (no regression).
3. `go test ./...` green.

Named test (new):
- `tests/unit/auth_password_change_flow_test.go::TestPasswordChangeEmitsExpiredCookie`
  asserts `Set-Cookie` present with `MaxAge=-1` and empty value.

Proof commands:
- `go test ./tests/unit -run TestPasswordChange -count=1`
- `go test ./... -count=1`

## ADR

None. Mirror of an existing established pattern (`handleLogout`). Not a durable
architectural decision.

## Interview record

No interview needed. Contract is mechanically derivable from the named site, the
existing `handleLogout` mirror, and the mission B4 finding. Operator authorized
M0 features in mission.md.
