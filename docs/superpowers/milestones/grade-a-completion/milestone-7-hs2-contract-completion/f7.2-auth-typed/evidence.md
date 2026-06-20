# Feature F7.2 — Evidence (auth typed responses; pre-codegen hand-rolled structs)

> **Status:** CLOSED 2026-06-20. Spec `./spec.md` (approved), plan `./plan.md`.

## Summary

Auth is pre-codegen (no `api.gen.go`). Per ADR 0012, defined two hand-rolled typed response structs —
`authLoginResponse{User authdomain.CurrentUser, ExpiresAt string}` and
`changePasswordResponse{Changed bool, User authdomain.CurrentUser}` — mirroring the OpenAPI
`AuthLoginResponse` / `ChangePasswordResponse` schemas, and swapped the 2 response-literal
`map[string]any` emits (`handler.go` login + change-password) for them. Wire-equivalent: the same
`authdomain.CurrentUser` value is emitted and `expires_at` stays the identically-formatted RFC3339
string.

## Acceptance — every gate, real commands + output

| # | Criterion | Command | Result |
|---|-----------|---------|--------|
| 1 | Zero response-literal `map[string]any` in auth `handler.go` | `grep -nE 'map\[string\]any' internal/modules/auth/delivery/http/handler.go` | **4 hits, all `recordAudit`** (`:98,:109,:127` audit-emit payloads + `:204` param decl) — non-response, allowlisted. **red→green: 2 response literals → 0** |
| 2 | No `WriteJSON(...map[string]any)` left | `grep -nE 'WriteJSON.*map\[string\]any' .../handler.go` | 0 (exit 1) |
| 3 | Login struct wire keys == OpenAPI `AuthLoginResponse` | `go test -run TestAuthLoginResponse_WireContract ./internal/modules/auth/delivery/http/` | PASS — keys `{expires_at, user}`, `user.user_id` + `expires_at` round-trip |
| 4 | Change-password struct wire keys == OpenAPI `ChangePasswordResponse` | `go test -run TestChangePasswordResponse_WireContract ...` | PASS — keys `{changed, user}` |
| 5 | Build + existing auth tests green | `go build ./...` exit 0; `go test -count=1 ./internal/modules/auth/...` | all `ok` (5 packages) |

## TDD note (honest)

The struct wire-contract tests are compile-red first (the structs don't exist until the
implementation defines them), then green — a genuine compile-driven red→green. They lock the
hand-rolled structs to the OpenAPI key set. There is **no pre-existing handler-level login /
change-password success test** (success wire is exercised at the full-HTTP E2E level the terminal
re-audit re-runs; `Authenticate` is unit-tested in the application layer). The build proves the
handler wires the new structs; the contract tests prove their wire shape. No fabricated red.

## Files changed

- `internal/modules/auth/delivery/http/handler.go` — 2 typed response structs; 2 emit swaps.
- `internal/modules/auth/delivery/http/handler_typed_response_test.go` — NEW, 2 wire-contract tests.

## Scope / HS discipline

- No codegen pipeline standup, no `go generate`, no routing rewire (HS-2 boundary respected — auth
  stays pre-codegen, hand-rolled structs per ADR 0012).
- No OpenAPI change (schemas already declared) → no FE codegen regen.
- The 4 `recordAudit` `map[string]any` audit-emit uses kept (internal, non-response).
- No auth/session/cookie/error-mapping logic touched.

## Defers

None.
