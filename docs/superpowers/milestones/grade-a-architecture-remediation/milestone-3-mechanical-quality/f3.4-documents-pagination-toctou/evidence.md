# Feature F3.4 — Evidence

> **Milestone:** 3 · **Feature:** `f3.4-documents-pagination-toctou` · **Closed:** 2026-06-14
> **Contract:** `spec.md` (Approach B — CTE single-query). Consumer = the list handler's
> `(items, total, hasMore)` boundary, shape unchanged.

## What was implemented

`repository.ListDocumentsPaginated` rewritten from two statements (keyset page query + separate
`CountDocuments`) to **one** CTE query. The CTE `filtered` computes `COUNT(*) OVER() AS total_count`
over the **base-filtered** set (before the cursor predicate); the outer query applies the cursor +
`ORDER BY … LIMIT n+1`. Page rows and `total` now come from one statement → one MVCC snapshot → no
list/count TOCTOU. Return widened to `(items, total, hasMore, err)`. `Service.ListDocumentsPaginated`
takes `total` from the single repo call and **drops** the separate `repo.CountDocuments` call;
`CountDocuments` itself is retained as a standalone tested capability (interface + impl + its own test).

Files: `repository/repository.go` (`ListDocumentsPaginated`), `application/service.go` (interface +
service body), `application/service_test.go` (fakeRepo arity),
`repository/repository_pagination_test.go` + `repository/list_documents_paginated_test.go` (arity +
`total_count` column).

**Net external effect:** none — `(items, total, hasMore)` identical to the handler/FE. This is a
correctness fix (total↔page consistency) plus a round-trip removed (two queries → one).

## Verification

| Gate | Command | Result (real output) |
|------|---------|----------------------|
| G1 (TDD: one query w/ `COUNT(*) OVER()`, non-zero consistent total) | `go test ./internal/modules/documents/repository/ -run TestListDocumentsPaginated -count=1` | `ok metaldocs/internal/modules/documents/repository 2.071s`. RED before impl: old 3-value signature + no `COUNT(*) OVER()` in the list query → tests fail to compile against the new arity and the `COUNT\(\*\) OVER\(\)` `ExpectQuery` regex would not match the old SQL. GREEN after impl. |
| G2 (no separate count in list path) | read `service.go:369-382` | `Service.ListDocumentsPaginated` makes one repo call; the `repo.CountDocuments(...)` call is removed. `grep` shows `CountDocuments` only in the interface decl + retained impl/test, not in the list service path. |
| G3 (keyset preserved) | `repository_pagination_test.go` `TestListDocumentsPaginated_CursorKeyset` / `_InvalidCursor` | green: page-1 `hasMore` via n+1 probe (11→trim 10, hasMore=true); page-2 cursor binds `(updated_at,id) < ($2::timestamptz,$3)`, 5 rows → hasMore=false; invalid cursor → `ErrInvalidCursor`. |
| G4 (total consistency, page-independent) | `TestListDocumentsPaginated_CursorKeyset` | `total1 == 25` (page 1) and `total2 == 25` (page 2, different cursor) — grand total identical across pages, not a per-page count. |
| G5 (build + module tests) | `go build ./...`; `go test ./internal/modules/documents/... -count=1` | build clean; all documents packages `ok` (application, delivery/http, repository, domain, approval/*). |
| G6 (no contract drift) | `npx @redocly/cli lint api/openapi/v1/openapi.yaml` | `Your API description is valid. 🎉` — no OpenAPI/FE edit; handler/service signatures to the FE unchanged. |

Real vs fixture: repo tests use **sqlmock** (no Postgres) — labeled fixture. The query-shape and
arg-binding assertions are real against the production query string; the row values (incl.
`total_count`) are author-supplied, so the G4 "consistency across pages" assertion guards the
**code path** (trim-to-limit still returns the scanned grand total, not zero or a per-page count), not
the SQL semantics of `COUNT(*) OVER()` itself — see bounded defer.

## Acceptance vs spec Validation Gate

All six gates (G1–G6) met. External `(items, total, hasMore)` shape preserved → **HS-2 did not trip**
(no shared-contract redesign; repo→service boundary only).

## Review disposition (two-stage, subagent-driven)

- **Spec-compliance reviewer → PASS.** No missing requirement, no scope creep. Confirmed: cursor
  predicate is in the **outer** query (NOT inside the CTE WHERE — the exact bug the spec warns
  against); `CountDocuments` retained; arg numbering correct; StatsByStatus/StatsByArea/cursor
  encoding/filters/auth/OpenAPI untouched; empty-result `total=0` edge handled.
- **Code-quality reviewer → APPROVE.** Zero critical, zero important *blockers*. SQL valid; CTE
  `COUNT(*) OVER()` correctly computes the pre-cursor grand total; column lists (CTE / outer / Scan)
  aligned 19-wide; `$N` placeholders correct; all error paths return the 4-tuple; all call sites
  (handler, module wrapper) consume the unchanged service-level signature. Non-blocking notes folded
  into defers below.

## Bounded defers

| Defer | Why bounded | Written trigger / owner |
|-------|-------------|-------------------------|
| No real-DB integration test proves the CTE-cursor placement (a regression putting the cursor inside the CTE would make `COUNT(*) OVER()` count only the post-cursor tail). sqlmock returns author-supplied `total_count`, so the unit tier cannot catch this. | sqlmock cannot execute SQL; the placement is guarded today by the spec-compliance review + the `COUNT\(\*\) OVER\(\)` shape assertion. Accepted limitation of the unit tier, not a code defect. | Add a Postgres-backed integration test (known 2-page dataset, assert page-1 and page-2 `total` equal the true grand total) when the documents integration-test harness lands; owner: backend agent. Trigger: M4/integration-test milestone or first real-DB documents test. |
| `total = 0` on a cursor-past-end page is documented benign only because keyset clients read `total` on page 1. The convention lives in a repo comment, not the handler/OpenAPI response doc. | Handler (`handler.go`) drives has-more from the `hasMore` bool, not `total` → no current bug. | Annotate the list response `total` semantics at the OpenAPI/handler level (or an ADR note) before v1; owner: backend agent. |

## Memory / cross-refs

TOCTOU/consistency fix on the documents list path. CTE read-only, no authz inside → does not touch the
`[[advisory-lock-deadlock-constraint]]` (H-PRE-1) tx boundary. Related:
`[[backend-target-architecture-governs-reviews]]`.
