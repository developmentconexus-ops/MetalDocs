# Feature F4c.2 — Evidence

> **Milestone:** 4c — Unified test-fixture framework  ·  **Folder:** `f4c.2-migrate-blocker-files`
> **Status:** ✅ CLOSED — all four blocker files migrated onto the `testdb` factory; the HS-2
> trigger prerequisite (operator-approved) is delivered as migration 0241 + ADR 0033.
> Judged against `spec.md` (approved pre-code) and `plan.md`. All proof is **real** (template-cloned
> Postgres under the operator DSN + the live dev DB), not fixture/mock.

## Outcome (TDD step 5 — GREEN)

Env (both): `$env:METALDOCS_DATABASE_URL` = `$env:DATABASE_URL` = operator DSN. The `testdb` template
rebuilds from current `db/` files every `go test` process, so migration 0241 applies automatically.

```
# F4c.2 named blocker tests — repository package (fillin + commit_upload)
go test -tags integration -count=1 -run 'TestFillInRepository|TestCommitUpload' ./internal/modules/documents/repository/...
ok  	metaldocs/internal/modules/documents/repository	161.386s

# approval/repository (5 real-DB tests migrated pgtest → factory)
go test -tags integration -count=1 ./internal/modules/documents/approval/repository/...
ok  	metaldocs/internal/modules/documents/approval/repository	416.569s

# approval/jobs (3 scheduled-publish worker tests migrated pgtest → factory)
go test -tags integration -count=1 ./internal/modules/documents/approval/jobs/...
ok  	metaldocs/internal/modules/documents/approval/jobs	402.404s
```

- **fillin** — GREEN with **no test change**; the only fix was the HS-2-approved trigger
  (`db/migrations/0241_*`, ADR 0033). Applied to the live dev DB too (`BEGIN / CREATE FUNCTION /
  INSERT 0 1 / COMMIT`).
- **commit_upload** (×2) — Family-C seed gap closed with the factory: deleted the local
  `seedTenantRole`; at both call sites seed the `metaldocs.tenants` parent (`NewTenant`) + the
  user/role (`NewUser(WithRole)`).
- **approval/repository** (×5) — swapped `pgtest.OpenAndMigrate` → `testdb.Open`; seed the
  CD/document/route/instance graph via the factory (real tripwire asserted); the raw guarded
  scheduled-write UPDATE runs inside `testdb.SeedWithCaps([{"cap":"document.edit"}])`.
- **approval/jobs** (×3) — deleted `seedScheduledDocument` / `scheduledDocumentSeed`; seed via
  `Scenario.ScheduledRevision`. The worker's own publish UPDATE satisfies the curated-baseline
  tripwire via `authz.BypassSystem` (`metaldocs.bypass_authz = 'scheduler'`) — verified the baseline
  trigger honors that bypass (`db/baseline/0001_current_schema.sql:496-498`), so no manual cap
  assertion is needed.

### Structural note — approval repository test file split (in-boundary)

`postgres_approval_repository_test.go` was `package repository` with **no build tag**, mixing pure
unit tests (`TestMapPgError`, OCC logic, struct shapes — must run under plain `go test`) with the 5
real-DB tests. The factory lives in the `//go:build integration` `testdb` package, so importing it
into the untagged file would break the non-integration build. Resolved by **splitting**: the 5 DB
tests moved to a new `postgres_approval_repository_integration_test.go` (`//go:build integration`,
same `package repository`); the unit tests stay in the original file (untagged; the now-unused
`pgtest` import dropped). Both compile in both tag states — `go vet` and `go vet -tags integration`
clean.

## Validation Gate (all GREEN)

| Check | Command | Result |
|-------|---------|--------|
| Named blocker tests pass | `go test -tags integration -run 'TestFillInRepository\|TestCommitUpload' .../repository/...` | ✅ ok 161s |
| approval/repository pass | `go test -tags integration .../approval/repository/...` | ✅ ok 416s |
| approval/jobs pass | `go test -tags integration .../approval/jobs/...` | ✅ ok 402s |
| `db.go` untouched (F4.1a Gate #5) | `git diff --stat tests/integration/testdb/db.go` | ✅ empty |
| baseline frozen | `git diff --stat db/baseline/0001_current_schema.sql` | ✅ empty |
| copy-paste helpers gone | `git grep -E 'func seedGovernedParents\|func seedScheduledDocument\|func seedTenantRole'` | ✅ no matches |
| no lingering `pgtest` in migrated files | `git grep testsupport/pgtest -- <4 files>` | ✅ no matches |
| both tag states compile | `go vet` + `go vet -tags integration ./internal/modules/documents/...` | ✅ clean |

## Bounded defers

- **`repository_revision_history_integration_test.go` → F4c.3.** Surfaced by the package-level run:
  `TestListRevisionHistory_ReturnsGovernedDocumentsNotAutosaveRows` fails on the curated baseline
  (`controlled_documents.create` not asserted, P0001) at line 108. **Pre-existing, out of F4c.2
  scope** — committed in the milestone-4 era (`1b04c11f` / `8c0c8340`), it raw-inserts a
  `public.controlled_documents` row while asserting only `document.create` / `document.edit`, so it
  has always tripped the curated tripwire. Not caused by this feature (the package compiles; nothing
  referenced the deleted `seedTenantRole`). Same defect class as commit_upload — **trigger:** F4c.3
  (migrate remaining raw-seed test files onto the factory) must migrate this file with the same
  `NewControlledDoc` / `NewDocument` pattern.

---

_(RED capture below retained for the record.)_

## RED capture (TDD step 2 — clean testdb baseline)

Env (both): `$env:METALDOCS_DATABASE_URL` = `$env:DATABASE_URL` = operator DSN.
Command: `go test -tags integration -count=1 -run 'TestFillInRepository|TestCommitUpload' ./internal/modules/documents/repository/...`

```
--- FAIL: TestFillInRepository_UpsertValueAndListValues (144.62s)
    fillin_repository_integration_test.go:54: upsert first: ERROR: document af17439a-ed43-43bb-b29c-140d1d9baa0d not found (SQLSTATE 23503)
--- FAIL: TestCommitUpload_PersistsRevisionAndFormDataSnapshot (2.41s)
    repository_commit_upload_integration_test.go:43: insert iam_users: ERROR: insert or update on table "iam_users" violates foreign key constraint "iam_users_tenant_id_fkey" (SQLSTATE 23503)
--- FAIL: TestCommitUpload_IdempotentReplayReturnsExistingMetadata (3.32s)
    repository_commit_upload_integration_test.go:146: insert iam_users: ERROR: insert or update on table "iam_users" violates foreign key constraint "iam_users_tenant_id_fkey" (SQLSTATE 23503)
```

### Classification (spec Q5)

| File | SQLSTATE | Class | Disposition |
|------|----------|-------|-------------|
| commit_upload (×2) | 23503 `iam_users_tenant_id_fkey` | **Family-C seed gap** — missing `metaldocs.tenants` parent for the `DeterministicID` tenant (`documents.tenant_id` has no FK, so `InsertDraftDocument` succeeds; `iam_users.tenant_id` does) | Fix with the factory (`NewTenant(WithTenant(tenantID))` before `seedTenantRole`→`NewUser`). In-boundary. |
| fillin | 23503 `document % not found` (custom trigger RAISE) | **Production-schema defect** (NOT a 42703 schema-shadow, NOT a seed gap) | **HS-2 STOP — see below.** |

approval/repository + approval/jobs RED not yet captured (they are still on `pgtest`; migration deferred pending the HS-2 decision so the feature is closed in one consistent pass).

## HS-2 finding — broken tenant-consistency trigger `enforce_placeholder_value_tenant_consistent()`

**The fillin RED is not a seed gap.** The seed (`InsertDraftDocument`) is correct: it creates a document
and an initial `document_revisions` row, and passes that revision's id as `RevisionID`. The failure is a
genuine bug in a **production-schema trigger**.

### Root cause

`document_placeholder_values.revision_id` is a **`document_revisions.id`**:
- FK (live dev DB + baseline): `document_placeholder_values_revision_id_fkey → document_revisions`
  (re-pointed by archived migration `0191_document_placeholder_values_revision_fk.sql`).
- The repo (`FillInRepository.UpsertValue`) and the test both pass `documents.current_revision_id` (a real revision id).

But the BEFORE-INSERT/UPDATE trigger `enforce_placeholder_value_tenant_consistent()` resolves it against the
**wrong table** — verified identical in `db/baseline/0001_current_schema.sql:615-630` **and** the live dev DB
(`docker exec metaldocs-postgres psql … pg_get_functiondef`):

```sql
SELECT tenant_id INTO doc_tenant FROM documents WHERE id = NEW.revision_id;   -- ← revision_id looked up in documents.id
IF doc_tenant IS NULL THEN
    RAISE EXCEPTION 'document % not found', NEW.revision_id USING ERRCODE = 'foreign_key_violation';
END IF;
```

A revision id never matches a `documents.id` → `doc_tenant` is always NULL → **every** insert/update raises
`document % not found (23503)`. The trigger (added in archived `0153_placeholder_values_tenant_consistency.sql`)
was never updated when `0191` re-pointed the FK at `document_revisions`. The **same trigger function** is also
attached to `document_editable_zone_content` (0153), so both write paths are broken on the canonical schema.

### Impact

- The tenant-isolation check it implements is effectively dead: it **fails closed** (raises on every write), so
  there is no cross-tenant leak — but no placeholder-value or editable-zone-content write can succeed against the
  canonical schema. The fillin feature is broken (or running on a hand-patched schema) in any environment built
  from the canonical migrations.

### Why HS-2 (not patch-in-place)

Fixing it is a **production-schema migration** to a **tenant-isolation security trigger** — outside the F4c.2
test-fixture-migration boundary. The F4c.2 spec pre-committed this exact path (Q5; non-goal: "No production-source
change — except an **HS-2-escalated, operator-approved** Family-A schema fix"). It needs an ADR + the
`metaldocs-database` skill + operator sign-off, not a silent in-feature patch.

### Minimum prerequisite plan (for operator approval)

1. New forward migration under `db/migrations/` + matching `db/baseline/0001_current_schema.sql` edit
   (`metaldocs-database` skill; migration policy).
2. Correct the resolver to walk revision → document:
   ```sql
   CREATE OR REPLACE FUNCTION public.enforce_placeholder_value_tenant_consistent() RETURNS trigger AS $$
   DECLARE doc_tenant UUID;
   BEGIN
       SELECT d.tenant_id INTO doc_tenant
         FROM document_revisions r JOIN documents d ON d.id = r.document_id
        WHERE r.id = NEW.revision_id;
       IF doc_tenant IS NULL THEN
           RAISE EXCEPTION 'revision % not found', NEW.revision_id USING ERRCODE = 'foreign_key_violation';
       END IF;
       IF doc_tenant <> NEW.tenant_id THEN
           RAISE EXCEPTION 'tenant mismatch: document=% value=%', doc_tenant, NEW.tenant_id USING ERRCODE = 'check_violation';
       END IF;
       RETURN NEW;
   END;
   $$ LANGUAGE plpgsql;
   ```
   (Confirm `document_editable_zone_content.revision_id` has the same revision semantics before reusing the fn there.)
3. ADR under `wiki/decisions/` recording the defect + fix (cross-tenant security trigger).
4. Then resume F4c.2: migrate fillin onto the factory and run its Validation-Gate command GREEN.

**Status: STOPPED at HS-2. No trigger edit, no fillin migration, until operator approves the prerequisite.**
