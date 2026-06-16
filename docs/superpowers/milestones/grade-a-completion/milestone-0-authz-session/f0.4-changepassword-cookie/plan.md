# F0.4 — Plan

1. **Failing test first** — add `TestPasswordChangeEmitsExpiredCookie` to
   `tests/unit/auth_password_change_flow_test.go`. Use the existing flow
   (login → change-password) and assert the response carries one
   `Set-Cookie` with name = `cfg.SessionCookieName`, empty value, `MaxAge < 0`.
   Run; expect failure.
2. **Implement** — in `internal/modules/auth/delivery/http/handler.go::handleChangePassword`,
   after `CurrentUser` succeeds and before `WriteJSON`, emit
   `http.SetCookie(w, h.service.ExpiredSessionCookie())`. Mirror of `handleLogout:115`.
   Single line, success path only. No error-path change.
3. **Verify** — rerun the new test (green), rerun
   `TestPasswordChangeRevokesSessionAndClearsMustChangePassword` (still green),
   then `go test ./... -count=1`.
4. **Evidence** — record commands, output snippets, diff in `evidence.md`.
