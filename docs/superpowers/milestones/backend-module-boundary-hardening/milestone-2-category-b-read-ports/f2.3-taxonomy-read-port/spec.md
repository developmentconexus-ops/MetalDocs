# Feature F2.3 — Spec — taxonomy area-catalog read-port

> **Milestone:** M2 (category-b-read-ports) · **Feature:** `f2.3-taxonomy-read-port`
> **Census sites:** **B7** (`documents/repository/repository.go:159` — in-tx `SELECT name FROM
> metaldocs.document_process_areas WHERE tenant_id=$1 AND code=$2`), **B8**
> (`iam/infrastructure/postgres/area_catalog_reader.go:24` — off-tx `SELECT EXISTS(... FROM
> metaldocs.document_process_areas WHERE tenant_id=$1 AND code=$2)`).
> **Owner:** taxonomy owns `metaldocs.document_process_areas`. ADR-0039 D1 + D3(b).

## Problem

Two foreign modules read taxonomy's `document_process_areas` base table with raw SQL:

- **B7** — `documents` create-document path, **inside the create tx** (after the revision-lock
  acquire), reads the area *name* for the `(tenant, process_area_code_snapshot)` to denormalize
  `area_name_snapshot` onto the new `documents` row. `sql.ErrNoRows`-tolerant (missing → NULL name).
- **B8** — `iam` `PeopleService` pre-invite validation, **off-tx**, checks an area *exists* for
  `(tenant, areaCode)` before burning an auth identity on an invite. Mirrors the
  `user_process_areas (tenant_id, area_code)` FK as a clean boundary validation.

Both are plain, **non-recording** SELECTs (no authz GUC / CapTaxonomyView). They must stay
non-recording — B7 in particular runs inside a lock-holding tx (**HS-PRE-1**: no authz-recording
read inside a lock-holding tx).

## Consumer contract (defined before the producer)

Taxonomy publishes one narrow read-port. The interface lives in `taxonomy/domain`; a **stateless**
Postgres adapter lives in `taxonomy/infrastructure`. Each method takes a `db.DB` executor so one
adapter serves both the in-tx (B7: caller passes its `*sql.Tx`) and off-tx (B8: caller passes its
`*sql.DB`) callers — `*sql.Tx` and `*sql.DB` both satisfy `db.DB` structurally
(`internal/platform/db/tx.go`), exactly the F2.1 `CDFieldReader` pattern.

```go
package domain // taxonomy/domain

type AreaCatalogReader interface {
    // AreaName returns document_process_areas.name for (tenantID, code).
    // found is false when no such area row exists in that tenant — matching the
    // ErrNoRows-tolerant B7 call site (missing → NULL area_name_snapshot).
    AreaName(ctx context.Context, exec db.DB, tenantID, code string) (name string, found bool, err error)

    // AreaExists reports whether an area row exists for (tenantID, code).
    // The off-tx B8 boundary existence check; the iam caller passes its own pool.
    AreaExists(ctx context.Context, exec db.DB, tenantID, code string) (bool, error)
}
```

- **B7 consumer** (`documents/repository.Repository`): gains an `areaCatalog
  taxonomydomain.AreaCatalogReader` collaborator (mirrors the injected `displayName`/`cdRead`
  ports); the inline `document_process_areas` SELECT is replaced by `r.areaCatalog.AreaName(ctx, tx,
  d.TenantID, code)`; `found → areaName sql.NullString{Valid:true}`, `!found → NULL` (identical to
  today's ErrNoRows path). Passed `tx` ⇒ runs in the existing create tx, non-recording.
- **B8 consumer** (`iam/.../ProcessAreaCatalog`): gains the same port; `AreaCodeExists` delegates to
  `areaCatalog.AreaExists(ctx, c.db, tenantID, areaCode)`; the raw `EXISTS` SQL is deleted. iam
  depends only on `taxonomy/domain` (interface); main injects the `taxonomy/infrastructure` adapter.

## Non-goals

- **No authz change.** The port stays a plain non-recording reader — no GUC, no CapTaxonomyView
  (the canonical `NewTaxonomyAreaReader` enforces those for the *governed* read paths; this is the
  denormalization/boundary-validation seam and must preserve current behavior). HS-PRE-1 preserved.
- **No behavior/visibility/result change.** Same rows, same NULL-on-missing semantics, same tenancy
  predicate. Seam only.
- **No new owner table or view.** Mechanical ADR-0039 D3(b) read-port over the existing table.
- **No tx semantics change for B8.** It stays off-tx (caller passes pool).

## Interview record (consumer-contract discovery)

| Q | A |
|---|---|
| One port for both sites, or two? | One `AreaCatalogReader` with two methods — same owner table, same tenancy predicate; B7 wants `name`, B8 wants existence. Mirrors keeping a module's read surface cohesive. |
| Tx-aware how? | `db.DB` executor param per method (F2.1 pattern). B7 passes `*sql.Tx` (in create tx); B8 passes `*sql.DB`. One stateless adapter serves both. |
| Does the port enforce authz? | **No.** Both sites are non-recording today; B7 is inside a lock-holding tx (HS-PRE-1). Enforcing here would change behavior and risk deadlock. Plain SELECT only. |
| B7 missing-area semantics? | `found=false` → caller writes NULL `area_name_snapshot`, identical to the current `sql.ErrNoRows` branch. |
| Adapter home? | `taxonomy/infrastructure` (alongside `repository.go`, `family_code_resolver.go`); stateless `NewAreaCatalogReaderPG()`. |

## Validation Gate (concrete — approved before code)

| Acceptance criterion | Named test / proof command | Real vs fixture |
|----------------------|----------------------------|-----------------|
| B7 port path returns the same `(name, found)` as the raw in-tx `SELECT name` across present-area / absent-area, run **inside a tx** | `TestAreaCatalogReader_AreaNameParityWithRaw` (integration, :5434) | real (PG) |
| B8 port path returns the same existence bool as the raw `EXISTS` across present / absent / wrong-tenant | `TestAreaCatalogReader_AreaExistsParityWithRaw` (integration, :5434) | real (PG) |
| Whole tree builds; targeted tests pass | `go build ./...`; targeted module suites | real |
| Guard clears B7 + B8 | `go run ./tools/cilint ./...` exit 0 with `{documents/repository/repository.go, document_process_areas}` and the iam area_catalog site removed from `hgPendingRemediation` | real |
| 0 raw `document_process_areas` reads remain in `documents/` and `iam/` (non-test) | `grep` shows none outside taxonomy | real |
| B7 stays in-tx & non-recording (HS-PRE-1) | review: `AreaName` called with `tx`; adapter issues a plain SELECT, no authz GUC | real (review) |

> TDD: write the failing parity tests first, port to green, then delete the raw reads (parity green
> **before** deletion — D6). PG unavailable ⇒ mark integration steps **not-run (HS-3)**, never
> false-green.

## ADR needed?

- [x] **No new ADR.** Mechanical application of **ADR-0039** D1 + D3(b). Port shape (one reader, two
  methods, tx-aware via `db.DB`, non-recording) recorded here + in `evidence.md`. No novel
  cross-module contract.
