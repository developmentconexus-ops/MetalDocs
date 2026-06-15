# Feature F4c.2 — Plan (the "how")

> Spec (contract) is `spec.md`, approved pre-code. Build plan only.

## Order (TDD: discard → RED → migrate → GREEN)

1. **Discard WIP** — `git checkout -- <4 files>` (approval repo, jobs, revision_history, authz_bypass).
   Restores HEAD. The first two are then rewritten; the latter two are left at HEAD (F4c.3 scope).
2. **RED capture** — run the four suites as-is on the operator DSN. approval-repo + jobs are still on
   `pgtest` (will pass or fail there — capture); commit_upload + fillin already on `testdb.Open`. The key
   RED is **fillin**: capture its exact error and classify (Family-A 42703 schema-shadow → HS-2; else seed gap).
3. **Add `testdb.SeedWithCaps`** (factory framework, not db.go) — exported wrapper:
   ```go
   func SeedWithCaps(t *testing.T, db *sql.DB, capsJSON string, fn func(tx *sql.Tx) error) {
       seedWithCaps(t, db, capsJSON, fn)
   }
   ```
4. **Migrate `postgres_approval_repository_test.go`** — delete `seedGovernedParents`; swap
   `pgtest.OpenAndMigrate(t)` → `testdb.Open(t)` (drop `pgtest` import, add `testdb`). Per test: seed the
   document graph via the factory on `db`, then `tx, _ := db.BeginTx(ctx,nil); defer tx.Rollback()` and pass
   `tx` to the repo read. Use minted IDs from the returned structs (`tn.ID`, `doc.ID`, …) — no hardcoded
   UUID consts. `TestScheduleGenerationIncrements…`: seed approved doc via factory, run the raw UPDATE inside
   `testdb.SeedWithCaps([{"cap":"document.edit"}])`, assert via a post-commit read on `db`.
5. **Migrate `scheduled_publish_job_test.go`** — delete `seedScheduledDocument` + `scheduledDocumentSeed`;
   swap to `testdb.Open(t)`. Per test: `doc := testdb.Scenario{}.ScheduledRevision(t, db, gen,
   testdb.WithEffectiveFrom(effectiveAt), testdb.WithRevisionVersion(rv))`; build worker args from `doc.TenantID`
   / `doc.ID`; query by the same. If the worker needs a connection `search_path`, set it on the `db` pool with
   `SetMaxOpenConns(1)` (mirroring commit_upload) — decide empirically.
6. **Migrate `repository_commit_upload_integration_test.go`** — delete `seedTenantRole`; replace its two call
   sites with `testdb.NewTenant(t, db, testdb.WithTenant(tenantID))` (Family-C missing parent) +
   `testdb.NewUser(t, db, testdb.WithUserID(userID), testdb.WithTenant(tenantID), testdb.WithRole("system_admin"))`.
   Leave the rest (search_path / `InsertDraftDocument` / asserted-caps GUC) — discipline cleanup is F4c.3/F4c.4.
7. **Resolve `fillin`** — apply the step-2 classification: seed-gap → add the missing factory parent (likely
   `NewTenant(WithTenant(fillInTenantID))` if the dev tenant row is absent on the clean baseline); schema-shadow
   → **STOP (HS-2)**, write the boundary + minimum prerequisite plan into `evidence.md`, do not patch.
8. **GREEN** — run each Validation-Gate command, capture real output → `evidence.md`. Verify empty `db.go`
   diff, helpers-gone grep, `go vet`.

## Design notes / risks

- **Factory commits to the clone** (`seedWithCaps` commits its tx); a later read `tx` or the worker's own
  connection sees the data. Safe because each test owns a private template-clone DB (destroyed after).
- **Reads need no caps** — repo Load*/Validate* are SELECTs; the tripwire fires only on guarded
  INSERT/UPDATE/DELETE. H-PRE-1 not engaged (no authz-recording read inside a lock-holding tx).
- **Shared taxonomy across two CDs** (supersede test): build one `NewTaxonomy`, pass it via `WithTaxonomy(tax)`
  to both `NewControlledDoc` so they share `(tenant, profile, process_area)` — the lineage distinction is the
  `controlled_document_id`, which is what `ValidateScheduledSupersedeTarget` keys on.
- **jobs search_path** — `seedScheduledDocument` committed on `pgtest` with no `SET search_path`; the worker
  read worked because production SQL is schema-qualified. On a `testdb` clone the same should hold; if a
  bare-name resolution fails, pin `search_path` on a 1-conn pool. Empirical.
- **Minimal touch** — do not refactor the non-DB unit tests, do not remove the commit_upload/fillin inline
  `set_config`/`search_path` (F4c.3/F4c.4 own that). Every changed line traces to: helper deletion, harness
  swap, factory seed, or the fillin fix.

## Verify

- All Validation-Gate commands green (or HS-2 reported for fillin).
- `git diff --exit-code tests/integration/testdb/db.go` empty.
- `grep` shows the 3 helpers gone; `go vet -tags integration` clean.
