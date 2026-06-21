# Feature F2.3 — Plan — taxonomy area-catalog read-port

> Input: `spec.md` (approved). Engine: subagent-driven-development run inline (TDD).

## Plan

### Files touched
- **New** `internal/modules/taxonomy/domain/area_catalog_reader_port.go` — `AreaCatalogReader`
  interface (`AreaName`, `AreaExists`, each taking `db.DB`) + `NoopAreaCatalogReader`.
- **New** `internal/modules/taxonomy/infrastructure/area_catalog_reader.go` — stateless
  `AreaCatalogReaderPG`; `NewAreaCatalogReaderPG()`; identical SQL on the passed exec.
- **New** `internal/modules/taxonomy/infrastructure/area_catalog_reader_parity_integration_test.go`
  — `TestAreaCatalogReader_AreaNameParityWithRaw` (in-tx) + `TestAreaCatalogReader_AreaExistsParityWithRaw`.
- **Edit** `documents/repository/repository.go` — `Repository` gains `areaCatalog
  taxonomydomain.AreaCatalogReader`; `New(db, displayName, cdRead, areaCatalog)` (nil→Noop guard);
  B7 inline SELECT → `r.areaCatalog.AreaName(ctx, tx, ...)` mapped to the NullString.
- **Edit** `documents/module.go` — `Dependencies.AreaCatalogReader` (nil→Noop); pass to `repository.New`.
- **Edit** `iam/infrastructure/postgres/area_catalog_reader.go` — `ProcessAreaCatalog` gains the
  taxonomy port; `NewProcessAreaCatalog(db, areaCatalog)`; `AreaCodeExists` delegates to
  `areaCatalog.AreaExists(ctx, c.db, ...)`; raw EXISTS deleted.
- **Edit** `apps/api/cmd/metaldocs-api/main.go` — construct `taxonomyinfra.NewAreaCatalogReaderPG()`
  once; inject into `documents.Dependencies.AreaCatalogReader` and `NewProcessAreaCatalog(...)`.
- **Edit** test call sites of `documents/repository.New` and `NewProcessAreaCatalog` — pass the real
  adapter (integration) or `NoopAreaCatalogReader{}` (unit).
- **Edit** `tools/cilint/internal/analyzers/hgcrossmodule.go` — remove B7 + B8 entries.

### Ordering (parity-before-delete, D6)
1. Add `AreaCatalogReader` (+ Noop) to `taxonomy/domain`.
2. Implement `AreaCatalogReaderPG` (identical SQL, both methods, on the passed exec).
3. Write both parity tests (raw baseline vs port; B7 baseline run **inside a tx**). RED→GREEN.
4. Wire the port into documents repo + module + iam catalog + main; map results 1:1.
5. **Parity green** → delete the raw B7 `SELECT name` and B8 `EXISTS` SQL.
6. Remove B7 + B8 from `hgPendingRemediation`; `cilint` exit 0.
7. `go build ./...`; targeted tests; `grep` proofs; `evidence.md`.

### Test strategy
- Real PG (:5434). B7 parity: seed taxonomy area; in a tx, compare raw `SELECT name` vs
  `port.AreaName(tx,...)` for present and absent codes (absent → found=false, raw → ErrNoRows). B8
  parity: present / absent / wrong-tenant existence equality.
- Unit call sites that don't exercise the area read get `NoopAreaCatalogReader{}`.

### Import-cycle / boundary check
- `taxonomy/domain` imports only stdlib + `platform/db` — must NOT import documents/iam.
- New edges: `documents/repository → taxonomy/domain`, `iam/infrastructure/postgres → taxonomy/domain`
  (consumer→owner interface). Adapters injected from main (`taxonomy/infrastructure`). Verify `go build`.

### Risks
- **In-tx non-recording (HS-PRE-1).** B7 runs in the lock-holding create tx; the adapter must issue a
  plain SELECT with no authz GUC. Mitigated: port is a bare reader, parity test runs it in a tx.
- **Tenancy predicate parity.** Both methods keep `tenant_id=$1::uuid AND code=$2`. B8 wrong-tenant
  case locks tenant isolation.
