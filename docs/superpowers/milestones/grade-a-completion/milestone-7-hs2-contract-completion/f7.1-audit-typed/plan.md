# Feature F7.1 — Plan

> Engine: inline (superpowers:writing-plans structure). Spec: `./spec.md` (approved).

## Files touched

| File | Change |
|------|--------|
| `internal/modules/audit/delivery/http/handler.go` | Import `auditapi`; swap 4 response-literal maps → generated types; retarget `buildEventResponses`; remove `EventResponse` type. |
| `internal/modules/audit/delivery/http/handler_typed_test.go` | NEW — 3 strict-decode parity-lock tests. |

## Test strategy

- **Class:** handler-unit (`package httpdelivery_test`, `httptest` + fakes / memory export svc) — the
  canonical pattern already used by `handler_test.go` / `handler_export_test.go`. No testdb (not a DB
  integration test).
- New tests strict-decode (`json.Decoder` + `DisallowUnknownFields`) each 200/202 body into the generated
  `auditapi.*` type and assert key values. They are **parity-lock characterization tests** — green before
  and after the swap (the wire is intentionally unchanged); their teeth are catching any accidental wire
  drift introduced by the typed swap.
- **The feature's red→green is the H-D grep** (Gate #1): 5 `map[string]any` hits → 1 (decode buffer only).

## Task order (TDD-adjacent for a parity refactor)

1. Write `handler_typed_test.go` (3 strict-decode tests). Run → green (types exist, wire already matches)
   — this characterizes the current contract and confirms the generated types compile against the wire.
2. Capture the baseline grep (5 hits) — the red side of the H-D gate.
3. Refactor `handler.go`: the 4 swaps + `buildEventResponses` retarget + remove `EventResponse`.
4. Run audit tests → green (parity held); run grep → 1 hit (green side).
5. `go build ./...`; `go test -count=1 ./...`.
6. Drive-by §7 #11–13 (405 Allow): assess; the GET-only `handleEvents` already sets `Allow: GET` (:81).
   Record decision in evidence.
7. Write `evidence.md`; commit.

## Risk / rollback

- Single file + one new test file. Rollback = `git checkout` the two files.
- Primary risk = timestamp precision drift → mitigated by `Truncate(time.Second)` (Q5). The strict-decode
  tests + existing tests guard it.
