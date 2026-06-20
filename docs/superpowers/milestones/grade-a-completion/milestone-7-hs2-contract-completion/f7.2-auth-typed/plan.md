# Feature F7.2 — Plan

> Engine: inline. Spec: `./spec.md` (approved).

## Files touched

| File | Change |
|------|--------|
| `internal/modules/auth/delivery/http/handler.go` | Define 2 unexported typed response structs; swap the 2 response-literal emits (`:90`, `:161`). |
| `internal/modules/auth/delivery/http/handler_typed_response_test.go` | NEW — 2 white-box struct wire-contract tests. |

## Test strategy

- **Class:** handler-unit / white-box (`package httpdelivery`, like `handler_problem_test.go`) — needed
  to reference the unexported response structs.
- The structs are hand-rolled (auth is pre-codegen), so the wire-contract test marshals each struct and
  asserts the JSON key set equals the OpenAPI required-field set, plus the embedded `CurrentUser`
  round-trips. Deterministic, no service fixture needed.
- **red→green:** the H-D grep (2 response literals → 0). The struct tests are new contract locks.

## Task order

1. Write `handler_typed_response_test.go` referencing `authLoginResponse`/`changePasswordResponse` —
   compiles only after step 2 defines them (compile-red first).
2. Define the 2 structs; swap the 2 emits.
3. `go test ./internal/modules/auth/...`; grep gate; `go build ./...`.
4. Evidence; commit.

## Risk / rollback

- One file + one new test file. Rollback = `git checkout`.
- Parity risk minimal: same `CurrentUser` value, `expires_at` kept as the identically-formatted string.
