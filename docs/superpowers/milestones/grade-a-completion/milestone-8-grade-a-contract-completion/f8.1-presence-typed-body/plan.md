# Feature F8.1 — Plan

> Engine: inline (superpowers:writing-plans structure). Spec: `./spec.md` (approved 2026-06-20).
> Mirrors the operator-ratified M7 F7.1 typed-parity refactor pattern.

## Files touched

| File | Change |
|------|--------|
| `internal/modules/iam/presence/handler.go` | Import `iamapi "metaldocs/internal/modules/iam/api"`; in `handleSnapshot`, map `[]presence.Item` → `[]iamapi.OnlinePresenceItem` and emit `iamapi.PresenceSnapshotResponse` via `httpresponse.WriteJSON`, replacing the `map[string]any{"items":…}` literal. Drop the now-unused `encoding/json` import if no longer referenced elsewhere in the file (it is still used by `writeJSON` for the WS path — keep). |
| `internal/modules/iam/presence/handler_typed_test.go` | NEW — strict-decode parity-lock test: decode the snapshot body into `iamapi.PresenceSnapshotResponse` with `DisallowUnknownFields`, assert items/field-values + `status` always present + content-type/200. |

## Test strategy

- **Class:** handler-unit (`package presence`, `httptest` + the existing `fakeRepo`) — the canonical pattern
  already used by `presence_test.go`. No testdb (not a DB integration test).
- New test strict-decodes the 200 body into the generated `iamapi.PresenceSnapshotResponse` and asserts
  field values + `status` present + `Content-Type: application/json`. It is a **parity-lock characterization
  test** — green before and after the swap (the wire is intentionally key-set-equivalent); its teeth are
  catching any accidental wire drift from the typed swap (e.g. a dropped or renamed field, a lost status).
- **The feature's red→green is the H-D grep** (Gate #1): 1 response-literal `map[string]any` → 0.
- Existing `TestHandler_Snapshot_TenantScoped` / `TestHandler_Snapshot_MissingTenantReturns500` /
  `TestStream_SnapshotMatchesHTTP` pass **unmodified** (they decode into `map[string][]Item`; same keys).

## Task order (TDD-adjacent for a parity refactor)

1. Write `handler_typed_test.go` (strict-decode parity-lock). Run → green (generated types already match the
   wire key-set; confirms they compile against the snapshot).
2. Capture baseline grep (`handler.go:83` — 1 response literal) — the red side of the H-D gate.
3. Refactor `handleSnapshot`: map items → `[]iamapi.OnlinePresenceItem` (set `Status` pointer always),
   emit `iamapi.PresenceSnapshotResponse` via `httpresponse.WriteJSON(w, http.StatusOK, …)`.
4. Run presence tests → green (parity held); run grep → 0 response literals.
5. `go build ./...`; `go test -count=1 ./internal/modules/iam/...`.
6. Write `evidence.md`; commit.

## Risk / rollback

- Single source file + one new test file. Rollback = `git checkout` the two files.
- Primary risk = the `Status` pointer left nil (would drop `status` for omitempty) → mitigated by always
  setting it from `it.Status`; the parity-lock test asserts `status` present on every item.
- No OpenAPI change, no codegen regen (types already generated; the generated model was dead code — this
  feature makes it live).
