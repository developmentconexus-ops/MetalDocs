# F9.2 — comment endpoints typed body

> **Milestone:** M9  ·  **Status:** approved (operator Option-A proceed, 2026-06-20) — code may begin.

## Consumer contract (read from the consumer, before the producer)

Three routes declare `$ref: DocumentCommentResponse` in OpenAPI
(`openapi.yaml:2930/2956/2988`) and the generated `documentsapi.DocumentCommentResponse`
(`api.gen.go:188`) exists:

- `GET  /api/v1/documents/{id}/comments` → **200** `[]DocumentCommentResponse`
- `POST /api/v1/documents/{id}/comments` → **201** `DocumentCommentResponse`
- `PATCH /api/v1/documents/{id}/comments/{library_id}` → **200** `DocumentCommentResponse`

Generated shape (the contract FE codegen consumes):
`{ author, author_id, content: DocumentCommentContentNode[], created_at: date-time, done, id: uuid,
library_comment_id, parent_library_id?, resolved_at?: date-time, updated_at: date-time }` where
`DocumentCommentContentNode = map[string]interface{}` (`api.gen.go:177`).

The package-local `commentResponse` (`handler.go:1217`) diverges: `id` `string` (not uuid), `content`
`json.RawMessage` (not a typed node array), timestamps hand-formatted `string` (not `time.Time`). **The
contract is published; the producer emits a divergent local struct.**

## What to implement

- `toCommentResponse` returns `documentsapi.DocumentCommentResponse`, mapping `domain.Comment`:
  - `Id: c.ID` (already `uuid.UUID`), `Author: c.AuthorDisplay`, `AuthorId: c.AuthorID`,
    `LibraryCommentId`, `ParentLibraryId`, `Done: c.ResolvedAt != nil`, `ResolvedAt`,
    `CreatedAt: c.CreatedAt`, `UpdatedAt: c.UpdatedAt`.
  - `Content`: unmarshal `c.ContentJSON` (a JSON array of nodes) into `[]documentsapi.DocumentCommentContentNode`;
    nil/empty → `[]` (the generated field is non-omitempty → must serialize `[]`, not `null`).
- `listComments`/`createComment`/`updateComment` emit the generated type (slice for list).
- Remove the local `commentResponse` struct once unused.

## Non-goals

- No domain-model change to `domain.Comment`.
- No spec change (schema already correct).
- No change to comment create/update request decoding.

## Validation Gate

- Local `commentResponse` struct removed; the 3 handlers serialize `documentsapi.DocumentCommentResponse`.
- New handler test: a fixed `domain.Comment` (with a 2-node content array) serializes to the generated
  shape — `id` is the uuid string, `content` is the node array (not a raw blob), `done` reflects
  resolution; empty content serializes as `[]` not `null`.
- `go build ./...` + `go test ./internal/modules/documents/...` green; FE codegen drift-clean.
- The widened `noresponsemap` (F9.4) reports these sites clean.
