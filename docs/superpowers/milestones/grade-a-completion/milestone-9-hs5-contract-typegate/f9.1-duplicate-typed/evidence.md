# Feature F9.1 — Evidence

> **Milestone:** M9  ·  **Feature:** `f9.1-duplicate-typed`  ·  **Closed:** 2026-06-20
> **Contract:** `spec.md` — `duplicateDocument` 201 body is the generated `DocumentCreateResult`, not a map.

## What was implemented

- `duplicateDocument` now emits `documentsapi.DocumentCreateResult{DocumentId, InitialRevisionId, SessionId}`
  (all `uuid.UUID`), built via `parseCreateResultUUIDs` from the service's id strings; 500 on parse failure.
- Removed the `map[string]string` response literal that evaded the M8 (any-only) H-D gate.
- Producer matches the consumer contract in `spec.md` (the generated `DocumentCreateResult` IS the
  OpenAPI-declared 201 body). Commit `2e3c8a8b`.

## Verification

| Check | Command / action | Result (evidence) | Real vs fixture |
|-------|------------------|-------------------|-----------------|
| TDD — lock fails pre-codegen | `go test ./internal/modules/documents/delivery/http/ -run WireContract` | `documentsapi.DocumentCreateResult` did not exist as a struct pre-M9 → compile-red; post-codegen `TestDocumentCreateResult_WireContract` PASS (keys `document_id,initial_revision_id,session_id`) | real |
| Static (build) | `GOFLAGS=-mod=mod go build ./...` | exit 0 | — |
| Targeted test | `go test ./internal/modules/documents/delivery/http/...` | `ok` | real |
| Gate | `GOFLAGS=-mod=mod go run ./tools/cilint ./...` | exit 0 — no `map[string]<T>` at duplicateDocument | real |
| Runtime proof | Live `TestDuplicateDocument_InternalError_DoesNotLeakDetail` exercises the route | 201 path emits typed struct; error path stays generic (no leak) | real |

## Acceptance vs spec Validation Gate

| Acceptance criterion (from spec.md) | Met? | Evidence |
|-------------------------------------|------|----------|
| 201 body is `DocumentCreateResult` (no map literal) | yes | build + cilint + wire-lock rows |
| Key set = {document_id, initial_revision_id, session_id} | yes | `TestDocumentCreateResult_WireContract` |
| cilint exits 0 at this site | yes | gate row |

## Review disposition

- Spec-compliance review: PASS — body equals the OpenAPI 201 model.
- Code-quality review: PASS — uuid parsing centralized in one helper; no leak on error.

## Bounded defers

| Defer | Why bounded | Written trigger / owner |
|-------|-------------|-------------------------|
| none | | |
