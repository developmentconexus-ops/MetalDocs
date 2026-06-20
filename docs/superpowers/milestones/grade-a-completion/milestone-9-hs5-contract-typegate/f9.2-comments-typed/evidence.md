# Feature F9.2 — Evidence

> **Milestone:** M9  ·  **Feature:** `f9.2-comments-typed`  ·  **Closed:** 2026-06-20
> **Contract:** `spec.md` — comment list/create/update bodies are `DocumentCommentResponse`, not a map.

## What was implemented

- `toCommentResponse` now returns `documentsapi.DocumentCommentResponse`; removed the local
  `commentResponse` struct and map building.
- `decodeCommentContent` returns `[]documentsapi.DocumentCommentContentNode`, normalizing
  nil/empty/invalid → `[]` (wire-neutral, never `null`).
- `listComments` builds `[]documentsapi.DocumentCommentResponse`.
- Producer matches the generated OpenAPI model (the consumer contract). Commit `2e3c8a8b`.

## Verification

| Check | Command / action | Result (evidence) | Real vs fixture |
|-------|------------------|-------------------|-----------------|
| TDD — lock | `go test ./internal/modules/documents/delivery/http/ -run WireContract` | `TestDocumentCommentResponse_WireContract` PASS — unresolved keys `author,author_id,content,created_at,done,id,library_comment_id,updated_at`; resolved adds `parent_library_id,resolved_at` | real |
| Static (build) | `GOFLAGS=-mod=mod go build ./...` | exit 0 | — |
| Targeted test | `go test ./internal/modules/documents/delivery/http/...` (incl. `handler_comments_test.go`) | `ok` | real |
| Gate | `GOFLAGS=-mod=mod go run ./tools/cilint ./...` | exit 0 — no map literal at comment routes | real |
| Runtime proof | `handler_comments_test.go` exercises list/create/update through the live mux | typed bodies returned; empty content serializes `[]` | real |

## Acceptance vs spec Validation Gate

| Acceptance criterion (from spec.md) | Met? | Evidence |
|-------------------------------------|------|----------|
| Comment bodies are `DocumentCommentResponse` | yes | build + cilint + lock |
| Empty content = `[]` not `null` | yes | `decodeCommentContent` + lock |
| cilint exits 0 at comment routes | yes | gate row |

## Review disposition

- Spec-compliance review: PASS — bodies equal the generated model; optionals preserved.
- Code-quality review: PASS — single conversion helper; no dead local struct left.

## Bounded defers

| Defer | Why bounded | Written trigger / owner |
|-------|-------------|-------------------------|
| none | | |
