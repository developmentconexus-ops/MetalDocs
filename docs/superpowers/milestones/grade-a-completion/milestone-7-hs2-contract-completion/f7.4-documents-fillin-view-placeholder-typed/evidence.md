# Feature F7.4 — Evidence (documents typed 200 bodies; codegen-first + 1 off-spec hand-rolled)

> **Status:** CLOSED 2026-06-20. Spec `./spec.md` (approved), plan `./plan.md`.

## Summary

Documents has a codegen pipeline. Declared 200 JSON body schemas for the 4 previously-undeclared ops
(`getDocumentFillInSchema`, `putDocumentPlaceholderValue`, `getDocumentPlaceholderOptions`,
`viewDocument`), regenerated `api.gen.go`, and emitted the generated models at the 4 response sites.
Rich/polymorphic payloads (templates-domain placeholders; polymorphic placeholder options) use a
**typed envelope with opaque `[]interface{}` items**: the handler boxes the existing typed values via a
generic `toAnySlice` helper, so each element marshals via its own json tags — byte-identical wire, zero
conversion structs, no FE regression (items were untyped before). The **HS-6** 5th site
(`pdf_webhook_handler.go:113`, unwired + off-spec) got a hand-rolled `pdfCompleteResponse` struct per
ADR 0012, no OpenAPI change.

## Acceptance — every gate, real commands + output

| # | Criterion | Command | Result |
|---|-----------|---------|--------|
| 1 | OpenAPI declares 200 JSON schemas for all 4 codegen ops | 4 component schemas + `$ref`s added in `openapi.yaml` | DONE — `DocumentFillInSchemaResponse`, `PutPlaceholderValueResponse`, `PlaceholderOptionsResponse`, `ViewDocumentResponse` |
| 2 | BE codegen fresh — regen stable | `GOFLAGS=-mod=mod go generate ./internal/modules/documents/api/...` (run twice; 2nd produced no further diff) | regenerated `api.gen.go` committed in this feature |
| 3 | Zero response-literal `map[string]any` (incl. via `writeFillInJSON`) in the 4 files | `grep -nE 'map\[string\]any' fillin_handler.go placeholder_options_handler.go view_handler.go pdf_webhook_handler.go` | **0 hits (exit 1)**. red→green: 5 response literals → 0 |
| 4 | Handlers emit generated models / hand-rolled struct; build green | `GOFLAGS=-mod=mod go build ./...` | exit 0 |
| 5 | Wire JSON unchanged — new byte-parity tests | `go test -run 'TestFillInSchemaResponse_EnvelopeAndEmptyParity\|TestPutPlaceholderValueResponse_SecondPrecision\|TestViewDocumentResponse_ConditionalKeys\|TestPlaceholderOptionsResponse_BothBranchesParity\|TestPDFCompleteResponse_WireContract' ./internal/modules/documents/delivery/http/` | PASS (5/5) |
| 6 | Existing documents tests (characterization of wire) green | `go test -count=1 ./internal/modules/documents/...` | all `ok` |

## Wire-parity invariants proven

- **Empty `placeholder_schema` / `options` → `[]` not `null`** — handler boxes a non-nil `make()`-d slice
  (`toAnySlice`), test `TestFillInSchemaResponse_EnvelopeAndEmptyParity`.
- **Boxing is wire-neutral** — boxed `[]templatesdomain.Placeholder` marshals byte-identically to the
  domain slice emitted directly (same test); both option branches likewise
  (`TestPlaceholderOptionsResponse_BothBranchesParity`).
- **`updated_at` second-precision** — generated `time.Time` field assigned `…Truncate(time.Second)`
  marshals `2026-06-20T12:00:00Z` (no fractional part), matching the prior `Format(time.RFC3339)`
  (`TestPutPlaceholderValueResponse_SecondPrecision`).
- **View conditional keys** — `pdf_status` always present; `signed_url`/`pdf_url` present only when ready
  (omitempty `*string`) (`TestViewDocumentResponse_ConditionalKeys` + existing `view_handler_test.go`).
- **pdf_webhook** — `{document_id, final_pdf_s3_key}` unchanged (`TestPDFCompleteResponse_WireContract`
  + existing `pdf_webhook_handler_test.go`).

## TDD note (honest)

The H-D grep is a real red→green (5 response literals → 0). The pre-existing handler-level tests
(`fillin_handler_test.go`, `placeholder_options_handler_test.go`, `view_handler_test.go`,
`pdf_webhook_handler_test.go`) are the characterization proof: they drive the real handlers and assert
the wire keys, and stayed green across the swap → wire unchanged. The new `typed_response_test.go` adds
byte-parity teeth for the subtle invariants those tests don't cover (second precision, empty → `[]`,
boxing neutrality, envelope key sets).

## Files changed

- `api/openapi/v1/openapi.yaml` — +4 component schemas; wired into the 4 ops' `200`.
- `internal/modules/documents/api/api.gen.go` — regenerated (new model types).
- `internal/modules/documents/delivery/http/fillin_handler.go` — import + `toAnySlice` helper; 2 emits.
- `internal/modules/documents/delivery/http/view_handler.go` — import; 1 emit.
- `internal/modules/documents/delivery/http/placeholder_options_handler.go` — import; 2 emits.
- `internal/modules/documents/delivery/http/pdf_webhook_handler.go` — hand-rolled struct; 1 emit.
- `internal/modules/documents/delivery/http/typed_response_test.go` — NEW, 5 wire-parity tests.

## Scope / HS discipline

- **HS-6 resolution** (operator "convert it", 2026-06-20): the off-plan 5th site `pdf_webhook:113` folded
  into F7.4 as a hand-rolled struct; milestone.md F7.4 row updated to record the extension.
- No routing rewire through generated `ServerInterface`/`NewStrictHandler`; no new pipeline (HS-2 held).
- No OpenAPI entry for the off-spec pdf_webhook route (Phase C wont-fix preserved).
- "Typed envelope, opaque items" is a deliberate proportionality call (spec Q8) — no faithful re-modeling
  of `Placeholder`/`VisibilityCondition`/`UserOptionView`; that would be gold-plating beyond the H-D defect.
- No service/domain/authz/error-mapping change.

## Defers

None. (FE codegen regen for these documents schema additions is F7.5, per the contract-first sequencing.)
