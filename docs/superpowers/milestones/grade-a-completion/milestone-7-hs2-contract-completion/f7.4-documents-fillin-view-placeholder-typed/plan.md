# Feature F7.4 — Plan

> Engine: inline. Spec: `./spec.md` (approved). Contract-first order: OpenAPI → BE `api.gen.go` → handlers.

## Files touched

| File | Change |
|------|--------|
| `api/openapi/v1/openapi.yaml` | +4 component schemas; wire each into its op's `200` content. |
| `internal/modules/documents/api/api.gen.go` | Regenerated (`go generate`) — committed. |
| `internal/modules/documents/delivery/http/fillin_handler.go` | Emit generated models at `:58` + `:116`. |
| `internal/modules/documents/delivery/http/view_handler.go` | Emit `ViewDocumentResponse` at `:46-51`. |
| `internal/modules/documents/delivery/http/placeholder_options_handler.go` | Emit `PlaceholderOptionsResponse` at `:67,74`; box helper results into `[]any`. |
| `internal/modules/documents/delivery/http/pdf_webhook_handler.go` | Hand-rolled `pdfCompleteResponse` + swap `:113` (off-spec, no OpenAPI). |
| NEW `internal/modules/documents/delivery/http/typed_response_test.go` | White-box wire-parity tests for all 5 sites. |

## Task order

1. Edit `openapi.yaml`: add `DocumentFillInSchemaResponse`, `PutPlaceholderValueResponse`,
   `PlaceholderOptionsResponse`, `ViewDocumentResponse` components; `$ref` from the 4 ops' 200.
2. `GOFLAGS=-mod=mod go generate ./internal/modules/documents/api/...`; read `api.gen.go` for the exact
   generated type + field names (oapi-codegen naming).
3. Edit the 4 codegen handlers to emit the generated models (box rich/polymorphic items into `[]any`).
4. Add hand-rolled `pdfCompleteResponse`; swap pdf_webhook `:113`.
5. Write `typed_response_test.go` (wire-parity per route).
6. Gates: grep (0 map literals), `git diff --exit-code` on api.gen.go, `go build ./...`,
   `go test ./internal/modules/documents/...`.
7. Evidence; commit (OpenAPI + regen + handlers + tests + spec/plan/evidence together).

## Test strategy

- **Class:** handler-unit / white-box (`package http`) — reference unexported `pdfCompleteResponse` +
  the generated models.
- Wire-parity teeth: marshal each emitted body and assert (a) the envelope key set equals the OpenAPI
  declaration, (b) byte-parity on representative payloads — `updated_at` second precision, empty
  `placeholder_schema`/`options` → `[]`, view conditional `signed_url`/`pdf_url` presence/absence,
  pdf_webhook `{document_id, final_pdf_s3_key}`.
- **red→green:** the H-D grep (5 response literals → 0).

## Risk / rollback

- Biggest risk: generated type/field names differ from assumption → step 2 reads them before step 3.
- `time.Time` RFC3339Nano drift on `updated_at` → mitigated by `.Truncate(time.Second)` + a parity test.
- Empty-slice `null` vs `[]` → handler boxes into `make([]any, len(...))` (non-nil) + a parity test.
- Rollback = `git checkout` of the listed files (OpenAPI + regen revert together).
