# Feature F3.4 — Plan (the "how")

> **Milestone:** 3 · **Feature:** `f3.4-documents-pagination-toctou` · **Approach:** B (CTE single-query)
> Contract in `spec.md`. Execution: subagent-driven review discipline (spec-compliance reviewer →
> code-quality reviewer after the implementation, both PASS before close).

## Boundary

Repo→service only. External HTTP/service shape (`items, total, hasMore`) unchanged. No OpenAPI/FE
regen. `repo.CountDocuments` retained (standalone), only its use in the list path removed.

## Steps

1. **Repo TDD (red first).** Update `repository_pagination_test.go` to the 4-value arity and assert the
   emitted query contains `COUNT(*) OVER()` (G1) and that `total` is the grand total carried on the
   rows, identical across pages (G4). Add the `total_count` column to every sqlmock row builder
   (`StatusFilter`, `CursorKeyset` `makeRows`, plus the `list_documents_paginated_test.go` scan test).
   → verify: tests fail to compile/red against the old 3-value signature.
2. **Repo impl.** Rewrite `ListDocumentsPaginated` as a CTE:
   - `buildDocumentFilter` (base WHERE, args `[tenant, …optional]`) unchanged.
   - CTE `filtered` selects the 18 columns + `COUNT(*) OVER() AS total_count` over the base filter
     (NO cursor) — so the count is the grand total, not the post-cursor tail.
   - Outer query selects the 18 cols + `total_count` from `filtered`, applies the cursor predicate
     (`WHERE (updated_at,id) < ($k::timestamptz,$k+1)`) only when a cursor is present, then
     `ORDER BY updated_at DESC, id DESC LIMIT $n` (n = limit+1 probe).
   - Scan 19 cols (18 + `rowTotal`); set `total = rowTotal` per row (identical on each); n+1 trim →
     `hasMore`; empty result → `total = 0` (documented benign edge).
   - All error paths return the new `(nil, 0, false, err)` tuple.
   → verify: G1/G3/G4 repo tests green.
3. **Service.** Interface decl (`Repository.ListDocumentsPaginated`) → 4-value. `Service.ListDocumentsPaginated`
   takes `total` from the single repo call; delete the separate `repo.CountDocuments` call. Update
   `fakeRepo.ListDocumentsPaginated` to the new arity (returns `countReturn` as total). `CountDocuments`
   left in interface + fake (retained capability).
   → verify: `go build ./...` clean.
4. **Verify the boundary.** `go test ./internal/modules/documents/... -count=1` green; redocly lint clean
   (no contract drift). Confirm no other call site of `ListDocumentsPaginated` (handler, module
   wrapper) broke — handler consumes the service-level 4-value signature, unchanged.
5. **Two-stage review.** Dispatch spec-compliance reviewer, then code-quality reviewer. Fix by
   root-cause family if either requests changes; re-review until both pass.

## Risk / rollback

Pure repo-internal query rewrite + a return-arity widen. Rollback = revert the two source files and
three test files. No migration, no data change, no contract change → no deploy coordination.
