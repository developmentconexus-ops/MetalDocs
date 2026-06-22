# Feature F2.2 — Evidence

> **Milestone:** 2 — Distribuição coverage-scope  ·  **Feature:** `f2.2-coverage-backend`  ·  **Closed:** 2026-06-22
> **Contract:** `spec.md` (consumer contract + Validation Gate G1–G11 this proves against) · **Plan:** `plan.md`.
> A feature is closed only when every row below is filled with **real, honestly-labeled** output — not
> "done"/"green", and not a fixture passed off as real. All integration proof below is **live Postgres**
> (testdb clones a template DB with real migrations 0245/0246 applied; localhost:5433).

## What was implemented

A new **read-only** delivery + repository layer for the greenfield `distribution` module that implements
the three frozen denominator-only endpoints (F2.1c contract) as a projection over the two published views
+ the ADR-0029 iam display-name port. **Zero** contract/generated/migration/publish-path change.

- **Repository** `internal/modules/distribution/infrastructure/coverage_repository.go` (package `distributioninfra`):
  `NewCoverageRepository(db, names iamdomain.UserDisplayNameReader)`.
  - `Summary` — `COUNT(*)` over `metaldocs.v_cd_obligated_readers` (tenant+doc filtered).
  - `Recipients` — projection ⋈ `metaldocs.v_process_area_name`; keyset cursor on `(area_name NULLS LAST, user_id)`;
    `limit+1` for `has_more`; batch `DisplayNames` post-fetch (no N+1), `COALESCE` fallback to `user_id`;
    fail-closed `validateSource`.
  - `Coverage` — `source='area_grant'` grouped by area ⋈ `v_process_area_name`, ordered by `area_name`; empty slice for company-scope.
- **Transfer types** `internal/modules/distribution/domain/types.go` (package `distributiondomain`):
  `RecipientRow`, `RecipientsPage`, `AreaCoverageRow` — owned by domain so the delivery layer depends on
  domain, never on infra (code-review MAJOR fix).
- **Handler** `internal/modules/distribution/delivery/http/handler.go` + `routes.go` (package `distributionhttp`):
  implements the generated `StrictServerInterface`; maps domain results → generated `*ResponseObject`; errors via
  `httpresponse`/`problem` (`ErrInvalidCursor`→400, missing tenant→404, else 500); `pagination.ClampLimit` at the handler.
- **authz (tier-1)** `apps/api/cmd/metaldocs-api/permissions.go`: three GET rows
  `pathPrefix:/api/v1/documents` + `pathSuffix:/distribution|/distribution/recipients|/distribution/coverage`
  → `CapDistributionRead`, `VisibilityPermissionGuarded` (interview #2 / ADR-0007, correcting milestone.md's write-pattern mandate). No in-tx `authz.Require`, no `trg_require_cap_asserted`. Non-tx pool reads.
- **DI** `apps/api/cmd/metaldocs-api/main.go`: reuse the existing `displayNameRepo`, construct repo+handler, `RegisterRoutes(h, mux)`.

### Commit list (chronological)
```
1c63bdb3 test(M2/F2.2): add integration test + CoverageRepository for distribution coverage
c4b15d2e feat(M2/F2.2): wire distribution HTTP handler + tier-1 permissions
7c295b5d test(M2/F2.2): make fail_closed sub-test exercise the unknown-source guard (G5)
808ffc02 fix(M2/F2.2): correct NULL-area-name keyset cursor (null sentinel) + refactor transfer types to domain
a8479841 chore(M2/F2.2): normalize distribution pathPrefix to /api/v1/documents (no trailing slash)
```

## Verification (real output)

| # | Acceptance criterion | Command | Result | Real vs fixture |
|---|----------------------|---------|--------|-----------------|
| G1 | Obligated set across 3 legs, DISTINCT-by-user, `source` precedence | `go test -tags=integration -run TestDistributionCoverage ./internal/modules/distribution/...` | PASS — `recipients_three_legs_precedence` + `summary_total` | real (live PG) |
| G2 | `…/distribution` → `{total_targets}` distinct count | same suite (`summary_total`) | PASS | real |
| G3 | `…/recipients` contracted shape + keyset paginates no-skip/no-dup; 400 on bad cursor; name via ADR-0029 port | same suite (`recipients_keyset_pagination`, `recipients_pagination_null_bucket`, `recipients_bad_cursor`) | PASS (4 sub-tests) | real |
| G4 | `…/coverage` by-area `area_grant` totals ordered by `area_name`; empty for company-scope | same suite (`coverage_by_area`, `coverage_company_scope_empty`) | PASS | real |
| G5 | **Fail-closed** — unhandled `source` is an error, not a passed-through row | same suite (`fail_closed` — view-swapped to emit `future_grant`, asserts error) | PASS; **proven**: reverting the sentinel/guard makes the test go RED (see below) | real |
| G6 | authz via tier-1 row; permissions table holds | `go test ./apps/api/cmd/metaldocs-api/...` | `ok 7.767s`, exit 0 | real |
| G7 | `api-lint -strict` = 0 | `./scripts/api-lint/api-lint.exe -strict api/openapi/v1/openapi.yaml .` | `0 violation(s)`, exit 0 | real |
| G8 | `hgcrossmodule` = 0 — distribution reads only `metaldocs.v_*` + iam port | `go run ./tools/cilint/... ./internal/modules/distribution/...` | exit 0 | real |
| G9 | module-boundaries (distribution) + build/vet/test green | `check-module-boundaries.ps1`; `go build ./...`; `go vet ./...`; full integration suite | distribution **0**; build/vet exit 0; 9/9 sub-tests PASS | real |
| G10 | Boundary untouched — no migration/publish/search/`v_cd_grantee` from F2.2 | `git show --stat 1c63bdb3 c4b15d2e 7c295b5d 808ffc02 a8479841` | none list `db/migrations/*`, `publish_service.go`, `internal/modules/search/*`, `db/migrations/0243*` | real |
| G11 | Contract↔generated parity — `api.gen.go` + FE `index.d.ts` unchanged by F2.2 | same `git show --stat` (HEAD-relative; origin/main is behind — F2.1c committed-unpushed) | no F2.2 commit touches `api.gen.go` or `frontend/.../api-types/index.d.ts` | real |

> **G10/G11 measurement note:** the plan's literal `git diff origin/main` is wrong here — origin/main is behind
> (F2.1c's `api.gen.go` + FE types are committed-but-unpushed; the program never pushes). The honest check is
> **F2.2's own commits touch none of the frozen paths**, verified by `git show --stat` over the five SHAs.

### Critical-bug RED proof (independent, by controller)
The code-quality review found a latent bug: NULL-`area_name` recipients (`user_grant`/`company_scope`) encoded a
cursor as `EncodeCursor("", userID)`, which `DecodeCursor` rejects (`parts[0]==""` → `ErrInvalidCursor`), breaking
pagination inside the NULL bucket. Fixed with a `nullAreaCursorSentinel = "\x00"` routing token (non-empty → decodes;
never a SQL value). New sub-test `recipients_pagination_null_bucket` guards it. **Proof the test catches the bug:**
temporarily reverting the encode to `""` produced —
```
coverage_repository_integration_test.go:261: recipients_pagination_null_bucket: page 1 with cursor "fDI0...": invalid cursor
--- FAIL: TestDistributionCoverage/recipients_pagination_null_bucket (2.27s)
```
— then the fix was restored (`git checkout`) and the full suite re-run **9/9 PASS** on the clean committed tree.

## Acceptance vs spec Validation Gate

All eleven criteria (G1–G11) in `spec.md §Validation Gate` are met with real, live-PG evidence (table above).
Denominator-only confirmed: numerator grep (`read|acknowledg|overdue|pending|deadline|timeline|reminder`) over the
new files returns only benign prose/generated hits — **zero** numerator domain fields. Non-goals held: no migration,
no publish-path change, no `v_cd_grantee`/search/`0243` change, no role-grant of the cap, no FE work, no contract change.

## Review disposition

- **Spec-compliance review (subagent, independent):** initial **FAIL** — the `fail_closed` sub-test exercised the
  wrong path (unknown tenant/CD → empty, not unknown-`source` → error). Fixed (`7c295b5d`): renamed the old assertion
  to `unknown_document_empty`; new `fail_closed` swaps the view to emit `future_grant` and asserts the guard fires.
  Re-verified live green. All other 7 checks PASS on first read.
- **Code-quality review (subagent, independent):** **REQUEST CHANGES** — 1 Critical (NULL-bucket cursor),
  1 Major (infra DTO leak into delivery), 4 Minor (`errors.Is` dedup, DDL sync comment, permissions prefix style,
  handler `ClampLimit`). All addressed (`808ffc02`, `a8479841`). Critical independently RED-proven by the controller.

## Bounded defers

| Defer | Why bounded | Written trigger / owner |
|-------|-------------|-------------------------|
| Name-sorted recipient pages | Display name is resolved post-fetch via the iam port → cannot be the SQL keyset sort key; keyset is `(area_name NULLS LAST, user_id)` — deterministic/stable, satisfies the contract. UI sorts the rendered page if needed. | If product requires server-side name ordering, file a follow-up to add a name column to the published view or a name-keyset variant; operator owns. |
| FE wiring of the screen | Explicit non-goal — F2.3. | F2.3 spec + plan. |
| Operator role-grant of `CapDistributionRead` | `deferredCaps` is intentional; op grants to roles separately. | Operator action at deploy; no agent pre-grant. |

## Execution note (honest)

`testdb` teardown occasionally logs `drop isolated test database … timeout: context deadline exceeded` — a slow
`DROP DATABASE` on the degraded C: drive (known machine note), **not** a test failure; the suite reports PASS / exit 0
regardless. The integration suite is ~90–200s wall (template DB build dominates). The `distribution/domain` directory
existed untracked at session start (a prior session's scaffold); it is now the committed home of the transfer types.
