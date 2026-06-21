# Feature F2.3 — Evidence — taxonomy area-catalog read-port

> **Milestone:** M2 (category-b-read-ports) · **Feature:** `f2.3-taxonomy-read-port` · **Closed:** 2026-06-21
> **Contract:** `spec.md` (consumers = documents create-tx area-name read + iam pre-invite existence
> check; ADR-0039 D3(b) taxonomy-owned read-port).
> Census sites closed: **B7** (`documents/repository/repository.go` → in-tx `document_process_areas`
> name read), **B8** (`iam/infrastructure/postgres/area_catalog_reader.go` → off-tx existence read).

## What was implemented

By outcome:

- **Owner publishes one narrow port.** `taxonomy/domain` gains `AreaCatalogReader` (`AreaName`,
  `AreaExists`, each taking a `db.DB` executor) + `NoopAreaCatalogReader` —
  `internal/modules/taxonomy/domain/area_catalog_reader_port.go`. Consumers depend on the interface;
  they never name `document_process_areas` in their own SQL.
- **Stateless owner-side adapter.** `taxonomy/infrastructure.AreaCatalogReaderPG`
  (`internal/modules/taxonomy/infrastructure/area_catalog_reader.go`) runs each plain, non-recording
  SELECT on the caller-supplied executor. One instance serves both the in-tx B7 caller (pass `*sql.Tx`)
  and the off-tx B8 caller (pass `*sql.DB`) — the F2.1 `CDFieldReader` pattern. No authz GUC.
- **B7 consumer rewired, raw SQL deleted.** `documents/repository.Repository` gains `areaCatalog`;
  `New(db, displayName, cdRead, areaCatalog)` (nil→Noop). The inline `SELECT name FROM
  metaldocs.document_process_areas …` is replaced by `r.areaCatalog.AreaName(ctx, tx, …)`, run with the
  create `tx` (in-tx, non-recording — **HS-PRE-1** preserved); `found==false` reproduces the prior
  `sql.ErrNoRows` branch (NULL `area_name_snapshot`).
- **B8 consumer rewired, raw SQL deleted.** `iam` `ProcessAreaCatalog` gains the taxonomy port;
  `NewProcessAreaCatalog(db, areaCatalog)`; `AreaCodeExists` delegates to `areaCatalog.AreaExists(ctx,
  c.db, …)`. The raw `EXISTS` SQL is removed. iam depends only on `taxonomy/domain`.
- **Composition root.** `apps/api/cmd/metaldocs-api/main.go` constructs
  `taxonomyinfra.NewAreaCatalogReaderPG()` and injects it into both `documents.Dependencies` and
  `NewProcessAreaCatalog`.
- **Ledger drained.** B7 + B8 removed from `hgPendingRemediation`.

> Producer matches the consumer contract in `spec.md`: B7 needed `(name, found)` with ErrNoRows→
> found=false; B8 needed an existence bool with tenant scoping — the port provides exactly those.
>
> No import cycle: new edges are `documents/repository → taxonomy/domain` and
> `iam/infrastructure/postgres → taxonomy/domain` (consumer→owner interface); adapters injected from
> `main`. Confirmed by `go build ./...`.

## Verification

| Check | Command / action | Result (evidence) | Real vs fixture |
|-------|------------------|-------------------|-----------------|
| TDD — parity tests first, green before raw-read deletion (D6) | `go test -tags integration -run 'TestAreaCatalogReader_Area' ./internal/modules/taxonomy/infrastructure/` | `PASS` — `AreaNameParityWithRaw` (present_area, absent_area; **run in a tx**) + `AreaExistsParityWithRaw` (present, absent_code, wrong_tenant) | real (PG :5434) |
| Static — build | `go build ./...` | clean (exit 0) | — |
| Static — guard | `go run ./tools/cilint ./...` | `EXIT=0` with B7 + B8 removed from `hgPendingRemediation` | real |
| Targeted tests — documents + iam-postgres + taxonomy (unit) | `go test ./internal/modules/documents/... ./internal/modules/iam/infrastructure/postgres/ ./internal/modules/taxonomy/...` | all `ok` (incl. B8 sqlmock `TestProcessAreaCatalog_AreaCodeExists` — delegation still issues the expected EXISTS SQL) | real |
| Targeted tests — integration | `go test -tags integration ./internal/modules/documents/repository/ ./internal/modules/taxonomy/infrastructure/` | both `ok` — B7 create-document path succeeds via the port | real (PG) |
| Runtime proof — port == raw | parity tests run the verbatim pre-port B7/B8 SQL baselines and the port, asserting equality across present/absent/wrong-tenant | identical results all cases | real (PG) |
| B7 stays in-tx & non-recording (HS-PRE-1) | review: `AreaName` invoked with the create `tx`; adapter issues a bare `SELECT name …`, no authz GUC/CapTaxonomyView | confirmed | real (review) |
| 0 raw `document_process_areas` reads in `documents/` + `iam/` (non-test) | `grep -rn document_process_areas` | only comment references remain; no SQL | real |

## Acceptance vs spec Validation Gate

| Acceptance criterion (from spec.md) | Met? | Evidence |
|-------------------------------------|------|----------|
| B7 port path == raw in-tx `SELECT name` (present/absent), run inside a tx | yes | `AreaNameParityWithRaw` 2/2 PASS |
| B8 port path == raw `EXISTS` (present/absent/wrong-tenant) | yes | `AreaExistsParityWithRaw` 3/3 PASS |
| Whole tree builds; targeted tests pass | yes | `go build ./...` clean; module suites `ok` |
| Guard clears B7 + B8 | yes | cilint `EXIT=0`, ledger entries removed |
| 0 raw `document_process_areas` reads outside taxonomy | yes | grep clean (non-test) |
| B7 in-tx & non-recording (HS-PRE-1) | yes | review + tx-run parity test |

## Review disposition

- **Spec-compliance review:** PASS. One cohesive owner port; tx-aware via `db.DB`; non-recording (no
  authz) — exactly the contract. No parallel cross-module reader introduced.
- **Code-quality review:** PASS. Adapter SQL is byte-for-byte the historical B7/B8 reads (parity tests
  lock equivalence). Stateless adapter + nil→Noop guards keep unit call sites fail-closed. B8 sqlmock
  test retained and green (delegation preserves the observable query).

## Bounded defers

| Defer | Why bounded | Written trigger / owner |
|-------|-------------|-------------------------|
| `TestSequenceAllocatorNextAndIncrement_Concurrent` (controlleddocuments/domain) env FAIL | Pre-existing / raw-base-DSN schema gap, unrelated to F2.3 (documented under F2.2 defers). HS-3 class — not false-greened. | Tracked in F2.2 evidence; env owner |
| Whole-tree `go test ./...` not run green end-to-end | Known pre-existing breaks elsewhere (documented in F2.1), unrelated to this seam | Tracked in F2.1 defers |
