# Feature F4.1b — Plan

## Approach

Root cause = the operator DSN's `search_path` runtime param leaks into every isolated-test-DB
connection and overrides the per-database `ALTER DATABASE` default. Fix at the harness seam so the
test layer is DSN-independent, while preserving the canonical default today's tests rely on.

## Steps

1. **RED** — run `TestCreateDocumentTx_PopulatesAllSnapshotColumns` with the operator DSN that carries
   `search_path=metaldocs,public`; capture the `42703` failure verbatim. (Gate A.)
2. **Fix `tests/integration/testdb/db.go`:**
   - `openDBWithDatabase`: after `cfg.Database = dbName`, delete `cfg.RuntimeParams["search_path"]`
     (guard nil map) so no connection — admin, template, or test — sends a `search_path` startup
     param inherited from the DSN. Comment why (DSN-independence; per-DB `ALTER DATABASE` governs).
   - `Open`: immediately after the `CREATE DATABASE ... TEMPLATE ...` succeeds, run
     `ALTER DATABASE <dbName> SET search_path TO metaldocs, public` (canonical default) **before**
     opening the test pool, so all pool connections inherit it. Quote-ident the db name.
3. **GREEN** — rerun Gate A command → `ok`. (Gate B.)
4. **DSN-independence** — rerun with a DSN that omits `search_path` → `ok`. (Gate C.)
5. **Regression** — `go test -tags integration ./internal/modules/documents/... ./internal/modules/templates/...`
   under operator DSN; plus the 3 no-`search_path` tests (revision-history, authz_bypass,
   reader_visibility). (Gates D, E.)
6. **build + vet** — `go build ./...`; `go vet -tags integration ./tests/integration/... ./internal/modules/documents/...`. (Gate F.)
7. **evidence.md** — record RED→GREEN, DSN-independence, regression, real-vs-fixture labels, defers.
8. **Commit** F4.1b (harness file + F4.1b docs only). Re-dispatch milestone-validator.

## Risk / blast radius

- `openDBWithDatabase` is shared by admin/template/test connections. Admin ops are catalog/qualified;
  template bootstrap is fully schema-qualified (verified — every `CREATE TABLE` names its schema, no
  top-level `SET search_path`) → stripping the param cannot break bootstrap.
- Canonical `ALTER DATABASE ... metaldocs, public` reproduces the exact effective default the DSN
  provided, so no-`search_path` tests are unchanged.
- Per-connection `SET search_path` tests are unaffected (session SET overrides DB default on that conn).
