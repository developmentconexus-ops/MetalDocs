# F9.1 — duplicateDocument typed body

> **Milestone:** M9  ·  **Status:** approved (operator Option-A proceed, 2026-06-20) — code may begin.

## Consumer contract (read from the consumer, before the producer)

`POST /api/v1/documents/{id}/duplicate` → **201** with body matching the generated
`documentsapi.DocumentCreateResult`:

```json
{ "document_id": "<uuid>", "initial_revision_id": "<uuid>", "session_id": "<uuid>" }
```

OpenAPI (`api/openapi/v1/openapi.yaml:2576`) already declares `$ref: DocumentCreateResult` for this 201.
The generated Go type `documentsapi.DocumentCreateResult` (`api.gen.go:208`) already exists with
`DocumentId`/`InitialRevisionId`/`SessionId` as `openapi_types.UUID` (= `uuid.UUID`). FE codegen already
consumes the declared schema. **The contract is already published; the producer is non-compliant.**

## What to implement

Replace the `map[string]string{...}` literal at `handler.go:674` with `documentsapi.DocumentCreateResult`,
parsing `res.DocumentID`/`res.InitialRevisionID`/`res.SessionID` (strings) into `uuid.UUID`. On a parse
failure (server-generated UUIDs — should never happen), `slog.Error` + 500 problem+json.

## Non-goals

- No change to the service layer / `DuplicateDocument` signature.
- No spec change (the schema is already correct).
- No status-code change (201 stays).

## Validation Gate

- `handler.go` `duplicateDocument` constructs no response `map[string]string`.
- New handler test: response is 201 and JSON keys/values are byte-identical to the prior map output for a
  fixed result (`document_id`/`initial_revision_id`/`session_id`).
- `GOFLAGS=-mod=mod go build ./...` + `go test ./internal/modules/documents/...` green.
- The widened `noresponsemap` (F9.4) reports this site clean.
