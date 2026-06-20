# Feature F7.2 — Auth typed responses (hand-rolled structs; pre-codegen)

> **Milestone:** 7 — HS-2 Contract Completion  ·  **Folder:** `f7.2-auth-typed`
> **Status:** Approved 2026-06-20 — code change may begin.
> **Approved before code:** 2026-06-20 / leandrotca.work — inherited from the M7 Phase-2 operator
> approval (commit `45a03fa6`). No new consumer contract beyond `milestone.md` F7.2.

## Interview record (B1.5)

| # | Question | Answer / source |
|---|----------|----------------|
| 1 | Is auth codegen or pre-codegen? | **Pre-codegen** — no `internal/modules/auth/api/` dir, no `api.gen.go`, no `cfg.yaml`. ADR 0012 legacy posture: hand-rolled typed Go structs are the sanctioned pattern for pre-codegen modules (same as IAM in M6). |
| 2 | Which sites are response literals? | 2: `handler.go:90` (login 200) and `:161` (change-password 200). |
| 3 | Which `map[string]any` are kept? | The 4 `recordAudit(... payload map[string]any)` internal audit-emit uses (`:83,94,112`) + the `recordAudit` param decl (`:189`). These are **audit-emit params**, not response bodies — allowlisted. (`_test.go` uses excluded from the gate.) |
| 4 | What is the login wire shape? | `{user: CurrentUser, expires_at: <RFC3339 string>}` — OpenAPI `AuthLoginResponse` (`openapi.yaml:3805`, required `[user, expires_at]`, `expires_at` is `string/date-time`). `session.CurrentUser` is `authdomain.CurrentUser`. |
| 5 | What is the change-password wire shape? | `{changed: bool, user: CurrentUser}` — OpenAPI `ChangePasswordResponse` (`openapi.yaml:3849`, required `[changed, user]`). `currentUser` from `service.CurrentUser()` is `authdomain.CurrentUser`. |
| 6 | Keep `expires_at` as string or switch to `time.Time`? | **Keep as `string`**, assigned `session.ExpiresAt.UTC().Format(time.RFC3339)` exactly as today → byte-identical wire. The struct is hand-rolled, so the field type is ours to choose; string avoids any marshaling-precision drift. |
| 7 | HS-2 risk — does this imply standing up a codegen pipeline? | **No.** Hand-rolled structs only; no `cfg.yaml`, no `go generate`, no routing rewire. Within ADR 0012 legacy posture (the operator's explicit M7 scope). |

## Consumer contract (FIRST — before any producer)

**Consumers:** FE auth callers (login / change-password) + the OpenAPI `AuthLoginResponse` /
`ChangePasswordResponse` schemas (source of truth).

| Op | Path / method | Status | Body type emitted | Wire keys (unchanged) |
|----|---------------|--------|-------------------|------------------------|
| login | `POST /api/v1/auth/login` | 200 | `authLoginResponse{User authdomain.CurrentUser, ExpiresAt string}` | `{user, expires_at}` |
| change password | `POST /api/v1/auth/change-password` | 200 | `changePasswordResponse{Changed bool, User authdomain.CurrentUser}` | `{changed, user}` |

`user` serializes to the existing `CurrentUser` JSON (`user_id, tenant_id, …, roles, capabilities`) —
unchanged, since the same `authdomain.CurrentUser` value is emitted.

## What this feature implements

1. Define two unexported typed response structs in `auth/delivery/http/handler.go` (next to
   `loginRequest`/`changePasswordRequest`):
   - `authLoginResponse{ User authdomain.CurrentUser \`json:"user"\`; ExpiresAt string \`json:"expires_at"\` }`
   - `changePasswordResponse{ Changed bool \`json:"changed"\`; User authdomain.CurrentUser \`json:"user"\` }`
2. Swap `:90` login emit → `authLoginResponse{User: session.CurrentUser, ExpiresAt: session.ExpiresAt.UTC().Format(time.RFC3339)}`.
3. Swap `:161` change-password emit → `changePasswordResponse{Changed: true, User: currentUser}`.

## Non-goals (mandatory)

- No change to wire keys/values (byte-identical, incl. the `expires_at` RFC3339 string).
- No codegen pipeline standup, no `go generate`, no routing rewire (HS-2 boundary).
- No change to the 4 `recordAudit` audit-emit `map[string]any` payloads (kept — internal, non-response).
- No change to auth/session logic, cookie handling, error mapping, or the `CurrentUser` shape.
- No FE codegen regen (no OpenAPI change — the schemas already declare these shapes).

## Validation Gate (concrete — approved before code)

| # | Acceptance criterion | Named test / proof command | Real vs fixture |
|---|----------------------|----------------------------|------------------|
| 1 | Zero response-literal `map[string]any` in `auth/.../handler.go` — only the 4 `recordAudit` audit-emit uses + param decl survive | `grep -nE 'map\[string\]any' internal/modules/auth/delivery/http/handler.go` → 5 hits, all `recordAudit` (audit-emit/param), 0 response literals | real (grep) — **red→green: 2 response literals → 0** |
| 2 | Login response struct wire keys == OpenAPI `AuthLoginResponse` | NEW `TestAuthLoginResponse_WireContract` — marshal `authLoginResponse`, assert key set `{expires_at, user}` + `user.user_id` round-trip | real |
| 3 | Change-password response struct wire keys == OpenAPI `ChangePasswordResponse` | NEW `TestChangePasswordResponse_WireContract` — assert key set `{changed, user}` | real |
| 4 | Build + existing auth tests green | `go build ./...`; `go test -count=1 ./internal/modules/auth/...` → 0 FAIL | real |

> Note: there is **no pre-existing handler-level login/change-password success test** (success wire is
> exercised at the full-HTTP E2E level the terminal re-audit re-runs; application-layer `Authenticate`
> is unit-tested in `service_test.go`). The struct wire-contract tests are the honest unit guard that
> the emitted typed shapes equal the OpenAPI; the build proves the handler wires them.

## ADR needed?

- [x] No durable decision — F7.2 follows the existing ADR 0012 legacy posture (hand-rolled typed structs
  for pre-codegen modules); no new design.
