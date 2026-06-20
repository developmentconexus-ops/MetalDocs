# Feature F7.3 — Plan

> Engine: inline. Spec: `./spec.md` (approved).

## Files touched

| File | Change |
|------|--------|
| `internal/modules/search/delivery/http/handler.go` | Define 1 unexported envelope struct; swap the response-literal emit (`:134`). |
| `internal/modules/search/delivery/http/handler_typed_response_test.go` | NEW — white-box envelope wire-contract + empty-array tests. |

## Test strategy

- **Class:** handler-unit / white-box (`package httpdelivery`, like `handler_test.go`) — needed to
  reference the unexported envelope struct.
- The struct is hand-rolled (search is pre-codegen), so the wire-contract test marshals it and asserts
  the JSON key set equals the OpenAPI required-field set `{items}`, plus the item round-trips and the
  empty case marshals to `[]` not `null`. Deterministic, no service fixture needed.
- **red→green:** the H-D grep (1 response literal → 0). The struct tests are new contract locks.

## Task order

1. Define the struct; swap the emit.
2. Write `handler_typed_response_test.go` referencing `searchDocumentsResponse`.
3. `go test ./internal/modules/search/...`; grep gate; `go build ./...`.
4. Evidence; commit.

## Risk / rollback

- One file + one new test file. Rollback = `git checkout`.
- Parity risk minimal: same already-non-nil `out` slice, same `items` key.
