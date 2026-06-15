# Feature F3.4 — Documents list pagination TOCTOU → single-snapshot query

> **Milestone:** 3 · **Feature:** `f3.4-documents-pagination-toctou`
> **Skill:** `metaldocs-backend-api` (+ persistence) · **Approach:** **B (CTE single-query)** —
> operator-approved at the HS-6 gate (2026-06-14), who delegated the engineering call.

## HS-6 reconciliation (why the governing-spec one-liner was wrong)

Governing spec §6/§5.2 said: "derive total from a single windowed query (`COUNT(*) OVER()`) instead of
a separate `SELECT COUNT(*)`." That prescription assumed OFFSET pagination. **The real code is
keyset/cursor pagination** (`repository.go:433` `ListDocumentsPaginated`:
`WHERE (updated_at,id) < (cursor) … ORDER BY updated_at DESC, id DESC LIMIT n+1`, returns `hasMore`).
Bolting `COUNT(*) OVER()` onto *that* query counts only **post-cursor** rows, not the grand total the
contract returns. The correct fix puts the count in a CTE computed **before** the cursor predicate.

## Consumer contract (read from the consumer first)

- **HTTP boundary (`delivery/http/handler.go:256`):** the list handler returns
  `items, total, hasMore`. `total` = **grand total of the base-filtered set** (all pages), shown by
  keyset clients on the first page; navigation uses the cursor, not `total`. **This external shape
  does not change.**
- **Service boundary (`application/service.go:369`):** `Service.ListDocumentsPaginated` returns
  `(items, total, hasMore, err)` — **unchanged** to the handler.
- **The race (today):** `service.go:374` runs the keyset page query, then `service.go:379` runs a
  **separate** `repo.CountDocuments` grand-total query. Two statements on a pooled connection, no
  shared snapshot → concurrent writes between them make `total` inconsistent with the page (TOCTOU).

**Required behavior:** `total` and the page rows come from **one SQL statement** (one MVCC snapshot),
so they are intrinsically consistent. External `(items, total, hasMore)` shape preserved.

## What to implement

Change the **repo→service boundary only**:

1. `repository.ListDocumentsPaginated` → return `(items []*domain.Document, total int64, hasMore bool, err error)`.
   Rewrite the query as a CTE:
   ```sql
   WITH filtered AS (
     SELECT <18 cols…>, COUNT(*) OVER() AS total_count
     FROM documents WHERE <base filter, NO cursor>
   )
   SELECT <18 cols…>, total_count
   FROM filtered
   WHERE <cursor predicate, if cursor>        -- (updated_at,id) < ($k::timestamptz, $k+1)
   ORDER BY updated_at DESC, id DESC
   LIMIT $n                                    -- limit+1 probe, unchanged
   ```
   Arg order preserved: `[base…, (cursorTS, cursorID)?, limit+1]`. Scan the extra `total_count` per
   row (all rows carry the same value); keep the `n+1` probe-row `hasMore` logic. **Empty result
   (cursor past end) → `total = 0`** (keyset clients read `total` on page 1 only — documented edge).
2. `application/service.go`: take `total` from the single repo call; **remove** the separate
   `repo.CountDocuments` call (`:379-382`). Update the repo interface decl (`:34`) signature.
3. Update repo fakes that implement the service's repo interface (`service_test.go` fakeRepo) and the
   repo-level sqlmock tests to the new arity + single CTE query.

## Non-goals

- No change to the external HTTP/service contract (`items, total, hasMore` shape identical).
- No OpenAPI/FE change (response shape unchanged → no regen).
- **`repo.CountDocuments` is retained** as a standalone, tested count capability — only its use in the
  list path is removed. (It becomes unused-by-list; recorded as a bounded note, not deleted, to keep
  the change surgical and preserve a tested public query.)
- No change to `StatsByStatus`/`StatsByArea`, the cursor encoding, filters, or auth scoping.
- No new transaction on the read path (B avoids A's snapshot-tx overhead).

## Validation Gate

| # | Acceptance | Proof |
|---|------------|-------|
| G1 | TDD: `ListDocumentsPaginated` issues **one** query containing `COUNT(*) OVER()` and returns a non-zero `total` consistent with the rows; red before impl (old code has no `COUNT(*) OVER()` in the list query and returns 3 values), green after. | `go test ./internal/modules/documents/repository/ -run TestListDocumentsPaginated -count=1` |
| G2 | Service no longer issues a separate count query: `grep -n CountDocuments service.go` shows no call in `ListDocumentsPaginated`. | read `service.go:369-385` |
| G3 | Keyset behavior preserved: first page `hasMore` via `n+1` probe; cursor page binds `(updated_at,id) < (…)`; invalid cursor still `ErrInvalidCursor`. | existing `repository_pagination_test.go` cases pass (updated for arity/CTE) |
| G4 | Total consistency: a test asserts the returned `total` equals the `COUNT(*) OVER()` value carried on the rows (same snapshot), independent of page. | new assertion in repo test |
| G5 | Build + full documents module tests green. | `go build ./...`; `go test ./internal/modules/documents/... -count=1` |
| G6 | No external contract drift. | `npx @redocly/cli lint api/openapi/v1/openapi.yaml` clean; handler/service signatures to the FE unchanged |

## Interview record

Operator was presented the HS-6 stop (spec prescription wrong for keyset) with three options
(snapshot-tx / CTE / defer) and asked for the senior-engineer recommendation. Recommendation: **B
(CTE)** — correct *and* removes a round-trip (one query vs two), vs A which adds tx overhead while
keeping two queries; defer buys nothing since the correct fix is cheaper at runtime and the churn is
mechanical + fully test-guarded. Operator delegated the call → proceeding with B.
