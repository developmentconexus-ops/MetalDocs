# Feature F9.1 — duplicateDocument typed body

> **Milestone:** M9 — HS-5 contract type-gate  ·  **Folder:** `f9.1-duplicate-typed`
> **Status:** Done

## Source

- Milestone spec row: route `duplicateDocument` (handler.go) through generated
  `documentsapi.DocumentCreateResult` instead of a `map[string]string` response literal.
- Governing-spec reference: mission.md §8 H-D class (no untyped map response body on a public route).

## Plan

1. Confirm the OpenAPI op for `POST /documents/{id}/duplicate` already declares `DocumentCreateResult`
   ({document_id, initial_revision_id, session_id} — all uuid). No schema change needed.
2. In `handler.go duplicateDocument`: replace `map[string]string{...}` with
   `documentsapi.DocumentCreateResult{...}`; parse the three service-returned id strings into
   `uuid.UUID` via a single `parseCreateResultUUIDs` helper, 500 on parse failure.
3. TDD: add a wire-contract lock (`TestDocumentCreateResult_WireContract`) pinning the key set; keep the
   existing live `TestDuplicateDocument_InternalError_DoesNotLeakDetail`.
4. Gates: `go build ./...`, `go test ./internal/modules/documents/...`, `go run ./tools/cilint ./...`.

## Execution notes

- Helper centralizes the 3 uuid.Parse calls with field-named errors; keeps the handler body flat.
- Service still returns id strings; conversion is at the delivery boundary (correct layer).
