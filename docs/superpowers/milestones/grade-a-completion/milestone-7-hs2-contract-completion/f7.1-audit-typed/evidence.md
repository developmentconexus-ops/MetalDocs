# Feature F7.1 — Evidence (audit typed responses; confirmed Major closed)

> **Status:** CLOSED 2026-06-20. Spec `./spec.md` (approved), plan `./plan.md`.
> **Commit:** (recorded at commit time below)

## Summary

Swapped the 4 response-literal `map[string]any` emits in
`internal/modules/audit/delivery/http/handler.go` for the already-generated `auditapi.*` types
(`ListAuditEventsResponse`/`CursorPage`/`AuditEventItem`, `AuditExportResponse`,
`AuditExportStatusResponse`). The export-status site was **the milestone's one confirmed Major**
(generated `AuditExportStatusResponse` existed but was unused). Retired the now-redundant hand-rolled
`EventResponse` type (its allowlisted `Payload` map survives as `AuditEventItem.Payload`). Drive-by:
added RFC 7231 `Allow` headers to the two export 405 sites.

## Wire-equivalence (not behavior change)

The swap is **wire-equivalent**: identical keys and values to the prior output. Two honest caveats:
- **Key ordering** shifts from map-alphabetical to struct-field order — semantically irrelevant per
  JSON (RFC 8259) and the OpenAPI contract; invisible to every key-based decoder (all existing audit
  tests pass unmodified).
- **Timestamps** (`occurred_at`, `expires_at`) are built `…UTC().Truncate(time.Second)` so the
  `time.Time` marshals to the same second-precision `…Z` string the prior `Format(time.RFC3339)`
  produced. `TestAuditHandler_ListEventsTypedShape` asserts `occurred_at` round-trips exactly
  (`.Equal(now)`).

## Acceptance — every gate, real commands + output

| # | Criterion | Command | Result |
|---|-----------|---------|--------|
| 1 | Zero response-literal `map[string]any` in `handler.go` — only the decode buffer survives | `grep -nE 'map\[string\]any' internal/modules/audit/delivery/http/handler.go` | **1 hit** — `:404 payload := map[string]any{}` (decode buffer feeding `AuditEventItem.Payload`, non-response). Baseline was 6 (4 response literals + `EventResponse.Payload` field + buffer). **red→green: 4 response literals → 0** |
| 2 | Grep-A blind-spot site closed | `grep -nE '(WriteJSON\|writeJSON).*map\[string\]any' internal/modules/audit/delivery/http/handler.go` | **0 hits** (exit 1) |
| 3 | Confirmed Major closed — export-status emits `AuditExportStatusResponse` | `go test -run TestAuditHandler_ExportStatusTypedShape ./internal/modules/audit/delivery/http/` | PASS (strict-decode, `DisallowUnknownFields`) |
| 4 | export-POST + events-list emit generated types | `go test -run 'TestAuditHandler_ExportPOSTTypedShape\|TestAuditHandler_ListEventsTypedShape' ...` | PASS |
| 5 | No wire drift — all existing audit tests pass unmodified | `go test -count=1 ./internal/modules/audit/...` | `ok` (all 4 packages with tests) |
| 6 | Build green | `go build ./...` | exit 0 |

Drive-by (§7 #11–13): `Allow: POST` on `handleExport` 405; `Allow: GET` on `handleExportSubresource`
405. New tests `TestHandleExport_405_Allow`, `TestHandleExportSubresource_405_Allow` — PASS.
`handleEvents` 405-Allow was already present (M-prior).

## TDD note (honest)

F7.1 is a typed-parity refactor — the wire is intentionally unchanged, so the generated types already
matched the wire and the strict-decode tests are **parity-lock characterization tests** (green before
and after; their teeth = catching any accidental drift the swap might introduce). The genuine red→green
for this feature is the **H-D Part-B grep** (4 response literals → 0). The drive-by `Allow` tests are
genuine red→green (the export-site `Allow` headers did not exist before this feature). No fabricated red.

## Files changed

- `internal/modules/audit/delivery/http/handler.go` — import `auditapi`; 4 typed swaps; `buildEventResponses`
  retargeted to `[]auditapi.AuditEventItem`; `EventResponse` type removed; 2 drive-by `Allow` headers.
- `internal/modules/audit/delivery/http/handler_typed_test.go` — NEW, 3 strict-decode parity-lock tests.
- `internal/modules/audit/delivery/http/handler_allow_test.go` — +2 drive-by 405-Allow tests.

## Scope / HS discipline

- No routing rewire, no `NewStrictHandler`, no new codegen pipeline (HS-2 boundary respected — audit
  stays a hand-rolled mux; only response bodies retyped).
- No OpenAPI change, no codegen regen (types already generated).
- Allowlisted non-response `map[string]any` kept: the `:404` decode buffer (feeds `AuditEventItem.Payload`).
- No authz/tenant/cursor/pagination logic touched.

## Defers

None.
