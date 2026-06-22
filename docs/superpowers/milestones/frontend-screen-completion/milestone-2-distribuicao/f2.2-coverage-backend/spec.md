# Feature F2.2 — Spec (coverage-backend)

> **Milestone:** 2 — Distribuição coverage-scope  ·  **Folder:** `f2.2-coverage-backend`
> **Status:** Approved (pre-code) — 2026-06-22 (operator).
> **Approved before code:** 2026-06-22 / operator (leandrotca) — *tier-1 middleware read-guard (`CapDistributionRead`, `VisibilityPermissionGuarded`) per ADR-0007 / `authz-tiers.md`, correcting the milestone.md write-pattern mandate; producer over `v_cd_obligated_readers` + `v_process_area_name` + ADR-0029 port; zero migration; contract frozen by F2.1c. Heavy TDD executes in a fresh `/milestone` session.*

> This is the feature's **contract**, written and approved **before any code**. It defines what the
> feature must do and how it will be proven — not how it will be built (that is `plan.md`). The
> milestone-validator judges the feature against *this* file (C1).

> **Engine note:** the consumer contract for F2.2 is already **locked** — F2.1c authored the OpenAPI
> operations and committed the generated Go server interface
> (`internal/modules/distribution/api/api.gen.go`) + FE types. F2.2 is the **producer** that satisfies
> that frozen contract. The interview below therefore records contract questions as *resolved-by-recon*
> implementation-contract decisions (the consumer shape itself needed no re-litigation).

## Interview record (fail-closed gate)

| # | Question | Answer |
|---|----------|--------|
| 1 | Does the consumer response shape need any (re)discovery? | **none needed** — the shape is frozen by F2.1c's `api.gen.go` (`DistributionSummaryResponse`, `DistributionRecipient`, `DistributionRecipientsResponse`, `DistributionAreaCoverage`) and the regenerated FE types. F2.2 implements *to* it; changing the shape would re-open F2.1c (and break ADR-0042's additive-only commitment). |
| 2 | authz: tier-1 middleware (like `audit`) or tier-2 in-tx `authz.Require` (like `controlleddocuments`)? | **tier-1 middleware row** — canonical for a tenant-grade GET read. **Resolved 2026-06-22 (operator)**, correcting the milestone.md F2.2 row which mandated the *write* pattern. Evidence: (a) `v_cd_obligated_readers` is **security-definer** (migration `0245` header: "no security_invoker") → caller RLS/GUC does **not** gate the read; no RLS reason to force a tx. (b) `trg_require_cap_asserted` fires on **writes only** (12 governed tables, `authz-tiers.md`) — distribution reads touch none, so the tripwire pairing is vacuous on a read. (c) Canonical doctrine (`wiki/concepts/authz-tiers.md`, ADR-0007): "Route guards (entry into HTTP handler): **tier 1**"; tier-2 is "area-scoped **writes** inside DB tx". `CapDistributionRead` is **tenant-grade** (`ScopeTenant`). (d) Precedent: `CapAuditRead` GET guarded by a tier-1 `permissions.go` row (`apps/api/cmd/metaldocs-api/permissions.go:232`, `VisibilityPermissionGuarded`). **Decision:** add one tier-1 row mapping `GET /api/v1/documents/{id}/distribution*` → `iamdomain.CapDistributionRead`, `VisibilityPermissionGuarded`, satisfying tier-1 authoring rules 1 (GET→view-grade cap) + 3. **No tx ceremony** — the handler reads on the pool with an explicit `WHERE tenant_id = $1` from auth context; **no** in-tx `authz.Require`, **no** `trg_require_cap_asserted` reference. (Tier-2 defense-in-depth rejected: no tx exists for any other reason, so it would be ceremony with no DB-enforcement backing.) |
| 3 | How is the recipient `name` resolved without a cross-module base-table read? | Via the **ADR-0029 iam display-name read-port** (`iamdomain.UserDisplayNameReader`), batch method `DisplayNames(ctx, tenantID, userIDs)`, concrete `iampg.NewUserDisplayNameRepository(db)`. **H-PRE-1:** the port reads `iam_users` on the **pool** (consistent with the tier-1 / no-tx posture of row #2). Fallback `COALESCE(name[user_id], user_id)`. One batch call per recipients page — no N+1. |
| 4 | Keyset order + the nullable `area_name`? | Keyset `(area_name, name, user_id)` ascending, opaque `CursorPage` cursor via `internal/platform/pagination` (`EncodeCursor`/`DecodeCursor`, base64 RawURL). `area_name` is **null** for `user_grant`/`company_scope` rows → ordering must use a **deterministic NULL policy** (`ORDER BY area_name NULLS LAST` with a matching keyset predicate, or `COALESCE` into a sort key) so the cursor is stable and total-ordering holds. The exact NULL encoding is a `plan.md` detail; the *contract* requirement is: deterministic, stable, no row skipped or duplicated across pages. |
| 5 | source precedence + distinct rule? | **inherited, not re-implemented** — `v_cd_obligated_readers` (F2.1a) already emits DISTINCT-by-user rows with precedence `user_grant > area_grant > company_scope`. F2.2 projects the view; it must NOT re-derive the union from base tables (ADR-0039 / `hgcrossmodule`). |
| 6 | coverage semantics (Σ ≠ total)? | By-area `coverage` counts `source='area_grant'` rows of `v_cd_obligated_readers` grouped by `area_code`, `area_name` resolved via `v_process_area_name`; ordered by `area_name`. Empty array for company-scope documents. `Σ coverage.total ≠ total_targets` by design (documented in the OpenAPI schema already). |
| 7 | fail-closed on an unhandled grant type? | The integration test asserts the handler/query **fails closed** (does not silently emit a row with an unknown `source`) — only the three contracted `source` values are surfaced; an unexpected value is an error, not a passthrough. |

## Consumer contract (FIRST — before any producer)

- **Consumer(s):** the frontend Distribuição screen (`DocumentDistributionPage.tsx` + child cards),
  wired in **F2.3** via TanStack Query hooks consuming the **generated** FE types
  (`frontend/apps/web/src/lib/api-types/index.d.ts`). The contract is exercised through the three
  HTTP endpoints below, served under `/api/v1`, tag `distribution`, gated by `CapDistributionRead`.
- **Contract** (frozen by F2.1c — F2.2 must match byte-for-byte against the generated types):

  **`GET /documents/{id}/distribution`** → `200 DistributionSummaryResponse { total_targets: integer ≥ 0 }`
  — distinct count of obligated users for the controlled document.

  **`GET /documents/{id}/distribution/recipients?cursor={opaque}&limit={1..100}`**
  → `200 DistributionRecipientsResponse { items: DistributionRecipient[], page: CursorPage }`,
  keyset order `(area_name, name, user_id)`.
  `DistributionRecipient { user_id: string, name: string, area_code: string|null, area_name: string|null, source: "area_grant"|"user_grant"|"company_scope" }`.
  `400` on a malformed cursor (`pagination.ErrInvalidCursor`).

  **`GET /documents/{id}/distribution/coverage`** → `200 DistributionAreaCoverage[]`
  (`{ area_code: string, area_name: string, total: integer ≥ 0 }`), order `area_name`; **empty array**
  for company-scope documents.

  **Errors (all paths):** RFC 9457 problem+json — `401` unauthenticated, `403` missing
  `CapDistributionRead`, `404` unknown/unauthorized document, `500` internal. Mapped via
  `internal/platform/problem` codes + `internal/platform/httpresponse`, matching the generated
  `*ApplicationProblemPlusJSONResponse` response types.
- **Source of truth for the contract:** the OpenAPI operations in `api/openapi/v1/openapi.yaml`
  (tag `distribution`) and their generated artifacts — Go: `internal/modules/distribution/api/api.gen.go`;
  FE: `frontend/apps/web/src/lib/api-types/index.d.ts`. F2.1c authored these; **F2.2 changes none of
  them** (regen drift would be HS-3).

## What this feature implements

A new **read-only** delivery layer for the `distribution` module
(`internal/modules/distribution/delivery/http/` + a supporting repository/query layer) that implements
the three contracted endpoints as a **projection** over the two published views:

- `v_cd_obligated_readers` (F2.1a, migration `0245`) — the obligated-set source (3 legs, DISTINCT-by-user, `source` precedence already encoded).
- `v_process_area_name` (F2.1b, migration `0246`) — area-label resolution.
- the **ADR-0029 iam display-name read-port** for recipient `name` (off-tx, batch).

`CapDistributionRead` is enforced by a **tier-1 middleware row** in
`apps/api/cmd/metaldocs-api/permissions.go` (one row per GET path/prefix, `VisibilityPermissionGuarded`)
— the canonical guard for a tenant-grade read (interview #2); no in-tx `authz.Require`. Handlers read
on the pool with an explicit `tenant_id` filter from auth context. Recipients paginate with the
canonical `CursorPage` keyset cursor. The handler set is registered into the API mux via
`distributionapi.HandlerWithOptions(h, StdHTTPServerOptions{BaseRouter: mux, BaseURL: "/api/v1"})`,
mirroring the `controlleddocuments` wiring.

## Non-goals (mandatory)

- **No numerator** — no read/acknowledge/overdue/pending/deadline/timeline/reminder data, field, query,
  or endpoint. (Parked mission; ADR-0042 additive-only.)
- **No new table, no migration** — F2.1a/F2.1b own the only new DDL; F2.2 adds **zero** `db/migrations` files.
- **No change to `PublishApproved()` / the publish path** — snapshot-at-publish is refused here (HS-2).
- **No modification of `v_cd_grantee`, migration `0243`, or any search-owned code** (HS-2).
- **No raw cross-module base-table reads** — distribution reads only `metaldocs.v_*` published views + the
  iam display-name port (`hgcrossmodule` = 0). Reading `iam_users`/CD/taxonomy base tables directly is forbidden.
- **No role-grant of `CapDistributionRead`** — it stays in `deferredCaps`; the operator grants it separately.
- **No FE work** — wiring the screen is F2.3.
- **No re-derivation of the obligated set** from base tables — the union/precedence lives in the F2.1a view only.
- **No new shared primitive / OpenAPI shape change** — the contract is frozen.

## Validation Gate (concrete — approved before code)

> TDD: write the failing integration test first, then implement to green. Fixture-backed proof against
> a live Postgres is **real** (the testdb framework applies real migrations incl. `0245`/`0246` and
> clones a real DB per test); a unit test with a stubbed repo is **fixture** and labeled as such.

| # | Acceptance criterion | Named test / proof command | Real vs fixture |
|---|----------------------|----------------------------|-----------------|
| G1 | The obligated set resolves correctly across all three legs (user_grant ∪ active area members ∪ company-scope), DISTINCT-by-user with `source` precedence, against a fixtured grant/area/company-scope graph | `go test -tags=integration -run TestDistributionCoverage ./internal/modules/distribution/...` (live PG via `tests/integration/testdb`) | real |
| G2 | `GET …/distribution` returns `{total_targets}` = the distinct obligated count for the fixtured doc | same integration suite (summary case) | real |
| G3 | `GET …/distribution/recipients` returns the contracted `DistributionRecipient[]` + `CursorPage`; keyset `(area_name, name, user_id)` paginates with **no row skipped or duplicated** across pages; `name` resolved via the ADR-0029 port; `400` on a malformed cursor | integration suite (recipients + pagination + bad-cursor cases) | real |
| G4 | `GET …/distribution/coverage` returns by-area `area_grant` totals ordered by `area_name`; **empty array** for a company-scope document | integration suite (coverage + company-scope cases) | real |
| G5 | **Fail-closed** — an unhandled/unexpected grant `source` is an error, not a silently-passed row | integration suite (negative case) | real |
| G6 | authz enforced via the **tier-1 middleware row** — request without `CapDistributionRead` → `403`; unauthenticated → `401`; unknown/unauthorized doc → `404`. The `permissions.go` GET row(s) exist for all three paths and point at `CapDistributionRead` (view-grade), passing `TestPermissionsTable_NoMethodlessWriteShadowing` + tier-1 authoring rules 1/3 | integration/handler test asserting status + problem code; `go test ./apps/api/cmd/metaldocs-api/...` (permissions table) | real |
| G7 | `api-lint -strict` over the live tree = **0** violations (contract unchanged; registry parity holds) | `./scripts/api-lint/api-lint.exe -strict api/openapi/v1/openapi.yaml .` → exit 0 | real |
| G8 | **`hgcrossmodule` = 0** — distribution reads only `metaldocs.v_*` views + the iam port, no base-table reads | `go run ./tools/cilint/... ./internal/modules/distribution/...` → exit 0 | real |
| G9 | Module-boundary + build/vet/test green | `./scripts/check-module-boundaries.ps1`; `go build ./...`; `go vet ./...`; `go test ./...` | real |
| G10 | **Boundary untouched** — no migration, no publish-path, no search, no `v_cd_grantee` change | `git diff --stat origin/main -- db/migrations` (only `0245`/`0246` from F2.1a/b, none from F2.2); `git diff …/approval/application/publish_service.go` = empty; `git diff internal/modules/search` = empty; `git diff db/migrations/0243*` = empty | real |
| G11 | Contract↔generated-types parity holds (HS-3) — `api.gen.go` + FE `index.d.ts` unchanged by F2.2 | `git diff --stat origin/main -- internal/modules/distribution/api/api.gen.go frontend/apps/web/src/lib/api-types/index.d.ts` = empty | real |

> **HS-3 prerequisite (run first in the execution session):** a live Postgres with migrations applied
> (`0245`/`0246` present). Bring up PG + apply migrations (`.\scripts\dev-migrate.ps1` or
> `.\scripts\start-api.ps1 -Build`), set `DATABASE_URL`, confirm the two views exist, *then* write the
> failing integration test. If the contract↔generated-types drift or the app won't start, repair the
> prerequisite before feature work.

## ADR needed?

- [x] **No new durable decision** — F2.2 implements within the decisions already recorded:
  **ADR-0042** (distribution module + `CapDistributionRead` + denominator-only contract),
  **ADR-0040** (`v_cd_obligated_readers`), **ADR-0041** (`v_process_area_name`),
  **ADR-0029** (`UserDisplayNameReader` port), **ADR-0039** (cross-module read boundary),
  **ADR-0007** (two-tier authz — the tier-1 read-guard decision in interview #2).
  If implementation surfaces a genuinely new durable decision (e.g. a NULL-keyset encoding worth
  recording), add an ADR at that point and link it here.
