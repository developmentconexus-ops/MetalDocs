# F0.4 — Evidence

## Diff

`internal/modules/auth/delivery/http/handler.go` (handleChangePassword) — one added
line after `CurrentUser` success, before `WriteJSON`:

```go
// F0.4: service already revoked all sessions (CWE-613). Mirror handleLogout
// and expire the cookie client-side so the browser drops the dead value.
http.SetCookie(w, h.service.ExpiredSessionCookie())
```

`tests/unit/auth_password_change_flow_test.go` — new test
`TestPasswordChangeEmitsExpiredCookie` asserts the response carries a `Set-Cookie`
for `cfg.SessionCookieName` with `MaxAge < 0` and empty value.

## TDD proof

1. **Red** (before fix):

   ```
   $ go test ./tests/unit -run TestPasswordChangeEmitsExpiredCookie -count=1
   --- FAIL: TestPasswordChangeEmitsExpiredCookie (0.76s)
       auth_password_change_flow_test.go:129: cookie metaldocs_session not found
   FAIL    metaldocs/tests/unit    1.059s
   ```

2. **Green** (after fix):

   ```
   $ go test ./tests/unit -run TestPasswordChange -count=1
   ok      metaldocs/tests/unit    2.215s
   ```

   Covers both the new `TestPasswordChangeEmitsExpiredCookie` and the existing
   `TestPasswordChangeRevokesSessionAndClearsMustChangePassword` (regression intact).

## Whole-repo regression

```
$ go test ./... -count=1
ok      metaldocs/tests/unit    4.670s
... (all packages ok; no FAIL lines)
```

## Acceptance map (vs. spec.md "Validation gate")

| Criterion | Result |
|---|---|
| 1. Success `Set-Cookie` present with `MaxAge<0` | PASS — `TestPasswordChangeEmitsExpiredCookie` asserts `MaxAge >= 0` failure and empty value |
| 2. Existing revocation test green | PASS — `TestPasswordChangeRevokesSessionAndClearsMustChangePassword` in the same `-run` batch |
| 3. `go test ./...` green | PASS — full repo run, no FAIL lines |

## Quality-bar / root cause

Fix at the named site (`handleChangePassword`), single line, mirrors the
established `handleLogout` cookie-clear pattern. No service-layer change required —
session revocation already correct (line 494, CWE-613). Not a symptom patch:
the missing browser-side cookie clear was the literal defect (mission §5 B4).

## Bounded defers

None.

## Review/QA disposition

- Code review (self): change is mechanical mirror of `handleLogout:115`; no
  alternative path needed.
- Workflow-class QA (backend-api, authz lens): no new grant, no widened access,
  deny-by-default unchanged.
