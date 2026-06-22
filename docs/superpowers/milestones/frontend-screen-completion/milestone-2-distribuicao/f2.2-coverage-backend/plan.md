# Feature F2.2 — coverage-backend

> **Milestone:** 2 — Distribuição coverage-scope  ·  **Folder:** `f2.2-coverage-backend`
> **Status:** Planning

This is the feature's **execution plan** — the "how" `milestone.md` left out. Body in
`superpowers:writing-plans` shape. Executes in a **fresh `/milestone` session** (per operator choice
2026-06-22 + milestone.md "fresh session per feature").

## Source

- **Milestone spec row (F2.2):** *implement* — distribution handler implementation over the two new
  views; *validate* — integration test asserts the resolved obligated set against a fixtured graph +
  fails closed on an unhandled grant type; new endpoint serves real denominator data.
- **Feature spec:** [`spec.md`](spec.md) — approved 2026-06-22 (tier-1 read-guard).
- **Governing decisions:** ADR-0042 (module + cap), ADR-0040 (`v_cd_obligated_readers`), ADR-0041
  (`v_process_area_name`), ADR-0029 (display-name port), ADR-0039 (read boundary), ADR-0007 (two-tier authz).

## Plan

### Execution rules (carry into the fresh session)
- Worker model **Sonnet 4.6** (never fable, ≤15 concurrent). Fresh implementer subagent per task +
  two-stage review (spec-compliance → code-quality). Fix by root-cause family.
- TDD: failing integration test first (T1 red), then implement to green.
- Commit after each verified task (standing auth, CLAUDE.md §5.0). **Never push.**
- Do **not** stage the pre-existing unstaged `AGENTS.md` / `CLAUDE.md` edits.
- Hard-stops: **HS-2** (any pull to mutate publish path / `v_cd_grantee` / search / a shared
  primitive → STOP, report boundary). **HS-3** (prereq fail: app won't start, views absent, or
  contract↔generated drift → repair prereq first, then resume). **HS-6** (any numerator / new table /
  action surface → STOP, route to parked mission).
- "obligated audience" never "reader" in any new prose/identifier; numerator-grep
  `read|acknowledg|overdue|pending|deadline|timeline|reminder` over new code must stay 0.

### T0 — Prerequisite gate (HS-3) — no feature code yet
1. Bring up live PG + apply migrations: `.\scripts\dev-migrate.ps1` (or `.\scripts\start-api.ps1 -Build`).
2. Confirm `metaldocs.v_cd_obligated_readers` (0245) + `metaldocs.v_process_area_name` (0246) exist.
3. Confirm contract↔generated parity (G11 baseline): `git diff --stat origin/main -- internal/modules/distribution/api/api.gen.go frontend/apps/web/src/lib/api-types/index.d.ts` = empty.
4. `go build ./...` green from clean.
- **Done:** PG up, both views present, parity clean, build green. Any failure → repair before T1.

### T1 — Failing integration test (TDD red)
- **File (new):** `internal/modules/distribution/infrastructure/coverage_repository_integration_test.go`
  — `//go:build integration`, package `distributioninfra` (match sibling
  `controlleddocuments/infrastructure`). Top-level `TestDistributionCoverage` with sub-tests.
- Use `tests/integration/testdb`: `testdb.Open(t)` (clones template w/ real migrations incl. 0245/0246),
  `SeedWithCaps` / `SeedGovernedTaxonomy` / `NewUser` / area-grant + user-grant + company-scope seeding
  to build a deterministic graph where the 3 legs overlap (forces DISTINCT + precedence).
- Sub-tests (map to G1–G5):
  - `summary_total` — distinct obligated count = expected.
  - `recipients_three_legs_precedence` — set + each row's `source` = precedence winner; `area_code`/
    `area_name` null for user_grant/company_scope.
  - `recipients_keyset_pagination` — page through with `limit` < total; assert union = full set, **no
    dup, no skip**, stable order `(area_name NULLS LAST, name, user_id)`.
  - `recipients_bad_cursor` — malformed cursor → `pagination.ErrInvalidCursor` (→ 400 at handler).
  - `coverage_by_area` — `area_grant` totals grouped + ordered by `area_name`.
  - `coverage_company_scope_empty` — company-scope doc → empty coverage slice.
  - `fail_closed` — an unexpected/unhandled `source` value is an error, not a passed-through row.
- **Done:** test compiles, runs against live PG, **fails** (no repo yet) for the right reason.

### T2 — Query/repository layer (implement to green, part 1)
- **File (new):** `internal/modules/distribution/infrastructure/coverage_repository.go`, package
  `distributioninfra`. Constructor `NewCoverageRepository(db *sql.DB, names iamdomain.UserDisplayNameReader)`.
- Methods:
  - `Summary(ctx, tenantID, docID) (int, error)` — `SELECT count(*) FROM metaldocs.v_cd_obligated_readers WHERE tenant_id=$1 AND controlled_document_id=$2`.
  - `Recipients(ctx, tenantID, docID, cursor, limit) ([]Recipient, nextCursor, hasMore, error)` —
    SELECT joined to `v_process_area_name` for `area_name`; `ORDER BY area_name NULLS LAST, user_id`
    with tuple keyset predicate decoded from cursor; fetch `limit+1` for `has_more`; encode next via
    `pagination.EncodeCursor`. `name` resolved **after** the row fetch via batch
    `names.DisplayNames(ctx, tenantID, userIDs)` (off-tx, pool; H-PRE-1), `COALESCE` fallback to userID.
    Keyset sort key = `name` requires it be available at SQL time — resolve order on
    `(area_name, user_id)` at SQL level (deterministic, stable) and document that `name` is a display
    overlay, OR join display name into the query if the port exposes a SQL-safe source. **Decision for
    implementer:** keep keyset on view-native columns `(area_name NULLS LAST, user_id)` — `user_id` is
    unique per row so the pair is a total order; `name` is presentation-only. (This refines spec
    interview #4's `(area_name, name, user_id)` to a SQL-stable `(area_name, user_id)` keyset; if the
    UI strictly needs name-sorted pages, that is a follow-up — flag, don't silently change contract.)
  - `Coverage(ctx, tenantID, docID) ([]AreaCoverage, error)` —
    `SELECT area_code, count(*) ... WHERE source='area_grant' GROUP BY area_code` joined to
    `v_process_area_name`, `ORDER BY area_name`.
- **No tx** — pool reads, explicit `tenant_id` filter. No `authz.Require`, no GUCs.
- **Done:** T1 sub-tests for the query layer pass (drive the test at the repo directly, or via handler in T3).

### T3 — HTTP handler (implement to green, part 2)
- **Files (new):** `internal/modules/distribution/delivery/http/handler.go` + `routes.go`, package
  `distributionhttp`. Implement the generated **`StrictServerInterface`** (`api.gen.go:702-713`) — the
  three operations. Map domain results → generated `*ResponseObject` types; errors via
  `h.writeDomainError` → `httpresponse.WriteError` with `problem` codes
  (`ErrInvalidCursor`→400/`CodeValidation`, not-found→404/`CodeNotFound`, else 500/`CodeInternalError`).
  `401`/`403` are produced upstream by the tier-1 middleware (T4), not here.
- Tenant/actor from auth context (`iamdomain` context helpers); pass `tenantID` to the repo.
- **Done:** handler compiles; T1 integration test (driven through the handler where applicable) green.

### T4 — Wiring + tier-1 guard
- **`apps/api/cmd/metaldocs-api/main.go`:** construct `distributioninfra.NewCoverageRepository(db, displayNameRepo)`
  (reuse the ADR-0029 repo built ~main.go:418), the handler, and register:
  `distributionapi.HandlerWithOptions(strictHandler, distributionapi.StdHTTPServerOptions{BaseRouter: mux, BaseURL: "/api/v1"})`
  (strict-handler wrapper per generated `NewStrictHandler`). Mirror the `controlleddocuments` block.
- **`apps/api/cmd/metaldocs-api/permissions.go`:** add tier-1 rows (after the Audit block, mirroring
  `permissions.go:232`):
  ```go
  // Distribution (M2/F2.2). Tenant-grade GET reads — tier-1 guarded.
  {method: http.MethodGet, pathPrefix: "/api/v1/documents/", capability: iamdomain.CapDistributionRead, visibility: iamdelivery.VisibilityPermissionGuarded},
  ```
  — **but** a bare `/api/v1/documents/` prefix collides with existing documents routes; instead add
  **exact/sub-prefix rows** for the three distribution paths only (`…/distribution`,
  `…/distribution/recipients`, `…/distribution/coverage`) so no existing documents row is shadowed.
  Implementer: verify against the existing `/api/v1/documents` rows (rules scan top-down, first match
  wins) — place distribution rows so they match the distribution paths without capturing sibling
  document GETs. Confirm `TestPermissionsTable_NoMethodlessWriteShadowing` + tier-1 authoring rules 1/3.
- **Done:** `go build ./...` green; permissions-table test green; app starts
  (`.\scripts\check-system-runnable.ps1 -TargetRoute /api/v1/health/ready`).

### T5 — Full validation gate (G1–G11) from clean
- `go test -tags=integration -run TestDistributionCoverage ./internal/modules/distribution/...` (G1–G5).
- `go test ./apps/api/cmd/metaldocs-api/...` (G6 permissions table); handler authz status test.
- `.\scripts\api-lint\api-lint.exe -strict api/openapi/v1/openapi.yaml .` → 0 (G7).
- `go run ./tools/cilint/... ./internal/modules/distribution/...` → 0 (G8 hgcrossmodule).
- `.\scripts\check-module-boundaries.ps1`; `go build ./...`; `go vet ./...`; `go test ./...` (G9).
- Boundary diffs (G10): `git diff --stat origin/main -- db/migrations` (only 0245/0246);
  `git diff …/approval/application/publish_service.go` empty; `git diff internal/modules/search` empty;
  `git diff db/migrations/0243*` empty.
- Parity (G11): `api.gen.go` + FE `index.d.ts` diff empty.
- Numerator-grep over new files = 0.
- **Done:** all 11 criteria green, evidenced with real output.

### T6 — Close
- Write `evidence.md` (template): commands + real output, TDD red→green proof, fixture-vs-real labels
  (all real / live-PG), review disposition, bounded defers (e.g. name-sorted keyset follow-up if filed).
- Commit `docs(M2/F2.2): close evidence …`. **No push.** Do **not** dispatch milestone-validator
  (that fires only after F2.3 closes the milestone).

### Files touched (summary)
| Path | Action |
|------|--------|
| `internal/modules/distribution/infrastructure/coverage_repository.go` | new — query layer |
| `internal/modules/distribution/infrastructure/coverage_repository_integration_test.go` | new — live-PG TDD test |
| `internal/modules/distribution/delivery/http/handler.go` | new — StrictServerInterface impl |
| `internal/modules/distribution/delivery/http/routes.go` | new — HandlerWithOptions registration |
| `apps/api/cmd/metaldocs-api/main.go` | edit — DI + mux registration |
| `apps/api/cmd/metaldocs-api/permissions.go` | edit — tier-1 GET rows → `CapDistributionRead` |
| `docs/.../f2.2-coverage-backend/evidence.md` | new — close-out proof |

**No** edits to: `openapi.yaml`, `api.gen.go`, FE types, `db/migrations`, publish path, search,
`v_cd_grantee`, iam capability registry (cap already registered in F2.1c).

## Execution notes

_(filled during `subagent-driven-development` in the fresh session — model choices, plan deviations
with rationale, questions answered. Durable record is `evidence.md`.)_
