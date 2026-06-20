# Feature F7.3 — Evidence (search typed response; pre-codegen hand-rolled envelope)

> **Status:** CLOSED 2026-06-20. Spec `./spec.md` (approved), plan `./plan.md`.

## Summary

Search is pre-codegen (no `api.gen.go`). The item mirror struct `SearchDocumentResponse` already
existed; the only response-literal gap was the envelope `map[string]any{"items": out}` at
`handler.go:134`. Per ADR 0012, defined a hand-rolled `searchDocumentsResponse{Items
[]SearchDocumentResponse}` and swapped the one site. Wire-equivalent: the same already-non-nil
`out` slice is emitted under the same `items` key → byte-identical `{"items":[...]}` (empty →
`{"items":[]}`).

## Acceptance — every gate, real commands + output

| # | Criterion | Command | Result |
|---|-----------|---------|--------|
| 1 | Zero `map[string]any` in search `handler.go` | `grep -nE 'map\[string\]any' internal/modules/search/delivery/http/handler.go` | **0 hits (exit 1)**. red→green: 1 response literal → 0 |
| 2 | No `WriteJSON(...map[string]any)` left | `grep -nE 'WriteJSON.*map\[string\]any' .../handler.go` | 0 (exit 1) |
| 3 | Envelope wire keys == OpenAPI `SearchDocumentsResponse` | `go test -run TestSearchDocumentsResponse_WireContract ./internal/modules/search/delivery/http/` | PASS — keys `{items}`, item `document_id` round-trip |
| 4 | Empty result → `{"items":[]}` not null | `go test -run TestSearchDocumentsResponse_EmptyIsArrayNotNull ...` | PASS |
| 5 | Build + existing search tests green | `go build ./...` exit 0; `go test -count=1 ./internal/modules/search/...` | all `ok` (3 packages w/ tests) |

## TDD note (honest)

The H-D grep is a real red→green (1 response literal → 0). The struct wire-contract test locks the
hand-rolled envelope to the OpenAPI `items` key set; the empty-array test guards the `null`-vs-`[]`
parity invariant. The pre-existing `handler_test.go` (which decodes `payload["items"]`) stays green
across the swap — it is the characterization proof that the wire shape is unchanged.

## Files changed

- `internal/modules/search/delivery/http/handler.go` — 1 typed envelope struct; 1 emit swap.
- `internal/modules/search/delivery/http/handler_typed_response_test.go` — NEW, 2 wire-contract tests.

## Scope / HS discipline

- No codegen pipeline standup, no `go generate`, no routing rewire (HS-2 boundary respected — search
  stays pre-codegen, hand-rolled struct per ADR 0012).
- No OpenAPI change (schema already declared) → no FE codegen regen.
- No change to the item shape, SQL reader, or query mapping.

## Defers

None.
