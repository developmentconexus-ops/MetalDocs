# Feature F9.2 — comment responses typed

> **Milestone:** M9 — HS-5 contract type-gate  ·  **Folder:** `f9.2-comments-typed`
> **Status:** Done

## Source

- Milestone spec row: route comment list/create/update through generated
  `documentsapi.DocumentCommentResponse` instead of a hand-rolled map / local struct.
- Governing-spec reference: mission.md §8 H-D class.

## Plan

1. Replace the local `commentResponse` struct + map building in `toCommentResponse` with
   `documentsapi.DocumentCommentResponse` (Id uuid, Author, AuthorId, Content typed nodes, Done,
   CreatedAt/UpdatedAt time.Time, ResolvedAt *time.Time).
2. Add `decodeCommentContent` → `[]documentsapi.DocumentCommentContentNode`, normalizing nil/empty/invalid
   JSON to `[]` (never `null`) to preserve the prior wire.
3. `listComments` slice → `make([]documentsapi.DocumentCommentResponse, 0, n)`.
4. TDD: wire-contract lock for unresolved (8 keys) and resolved (adds resolved_at, parent_library_id).
5. Gates: build, documents tests, cilint.

## Execution notes

- `Done` derived from `ResolvedAt != nil`; `resolved_at`/`parent_library_id` stay omitempty pointers to
  match the generated optionals.
- Empty-content normalization to `[]` is the wire-neutrality guarantee for FE list rendering.
