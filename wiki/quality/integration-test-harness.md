# Integration Test Harness

> Last verified: 2026-07-11 — added canonical local runner `scripts/test-integration.ps1` (derives `DATABASE_URL` from `.env`, fails loud on unreachable DB) + silent-skip warning. | prior: 2026-07-06 — F9.4 doc-truth pass: fixed import path `internal/modules/documents/repository` → `internal/modules/documents/infrastructure` (F9.5 rename) in the code sample. | prior: 2026-07-05 — M7 F7.4 added `testdb.OpenAsCIRole` (§"Other helpers"); rest of `factory.go`
> API unchanged since F4c.5.
> See also: [test-discipline.md](test-discipline.md) (CI guard rules R1–R4), [ADR 0034](../decisions/0034-integration-test-fixture-framework.md), [ADR 0027 Amendment 2026-07-05](../decisions/0027-rls-adoption-sequencing.md) (why a second DB role exists for isolation proofs).

This page is for developers **writing a new integration test**. It explains the harness choice,
how to open a database, which factory builders are available, and how to handle guarded writes.

---

## Why template-DB-per-test (IntegreSQL)

MetalDocs integration tests use **IntegreSQL** to clone a prepared template database for every
test. Each test gets its own isolated Postgres database, receives the curated baseline (migrations
+ reference data + capability tripwire caps), runs its assertions, and the clone is discarded.

**Why not a shared database with truncation?** M4b root-cause analysis identified two failure
classes that shared-DB designs cannot prevent:

- **Cross-test state leakage** — a prior test's uncommitted state or session-level GUC
  (`set_config` with `is_local=false`) bleeds into a later test's connection from the same pool.
- **`search_path` resolution drift** — a wrong `search_path` on a shared connection resolves bare
  table names (e.g. `documents`) to a stale or dead schema. Per-clone isolation removes this: every
  clone starts from the same curated schema state.

**Why not per-test migration?** Migration run time is O(number of migrations), which is already
large (~230+) and grows. Cloning a pre-migrated template is O(1) at test time.

See [ADR 0034](../decisions/0034-integration-test-fixture-framework.md) for the full decision record.

---

## Opening a database in a test

```go
//go:build integration

package mymodule_test

import (
    "testing"
    "metaldocs/tests/integration/testdb"
)

func TestSomething(t *testing.T) {
    db, schema := testdb.Open(t)
    // db is a *sql.DB pointed at a fresh clone of the curated baseline.
    // schema is the Postgres schema name (always "public" in current config).
    // The clone is automatically dropped when t.Cleanup fires.
    _ = schema
}
```

`testdb.Open` skips the test if `DATABASE_URL` or `METALDOCS_DATABASE_URL` is not set — safe on
machines without a local Postgres. **This skip is silent and reads as a green run.** Never verify
integration behavior with a bare `go test -tags=integration` unless you have confirmed the DSN is
set and the DB is reachable.

### Running integration tests locally (canonical)

```powershell
.\scripts\test-integration.ps1                                  # whole repo
.\scripts\test-integration.ps1 -Package ./tests/integration/approval/...
.\scripts\test-integration.ps1 -Run TestGovernancePolicy -Package ./tests/integration/approval/...
```

The script is the single chokepoint for DB access in tests:

1. If `DATABASE_URL`/`METALDOCS_DATABASE_URL` is already set, it uses it.
2. Otherwise it derives `DATABASE_URL` from the `POSTGRES_USER` / `POSTGRES_PASSWORD` /
   `POSTGRES_DB` / `POSTGRES_HOST_PORT` keys in `.env` (host `127.0.0.1`, default port `5433`,
   `sslmode=disable`) — without ever printing the value.
3. It probes the compose `metaldocs-postgres` container with `pg_isready` and **fails loud** if
   Postgres is not accepting connections (pre-empting the second silent-skip path at
   `tests/integration/testdb/db.go:73`).
4. Only then does it run `go test -tags=integration`.

Do NOT hand-craft a DSN or export `DATABASE_URL` manually in a session — that is exactly the
non-canonical drift this script exists to prevent. `.env.example` intentionally ships the
`DATABASE_URL` line commented out; the script derivation is the source of truth.

**When the system-under-test takes `*sql.DB` directly** and must honour session-wide caps:

```go
db, _ := testdb.Open(t)
db.SetMaxOpenConns(1)                          // single connection — makes session-level safe
testdb.SetCapsOnDB(t, db, `[{"cap":"document.edit"}]`)
```

---

## Factory builders

All builders live in `tests/integration/testdb/factory.go`. Each takes `(t *testing.T, db *sql.DB,
opts ...Opt)` and returns a struct carrying the IDs the test can assert on.

### Entity builders

| Builder | Returns | What it seeds |
|---------|---------|---------------|
| `NewTenant(t, db, opts...)` | `Tenant{ID}` | One row in `metaldocs.tenants` |
| `NewUser(t, db, opts...)` | `User{ID, TenantID, Role}` | One row in `metaldocs.iam_users` + optional role in `metaldocs.iam_user_roles`; also calls `NewTenant` if no `WithTenant` given |
| `NewTaxonomy(t, db, opts...)` | `Taxonomy{ProcessAreaCode, ProfileCode, FamilyCode, TemplateID, TemplateVersionID}` | Process area + document profile + document family + template + template version; per-test unique codes |
| `NewControlledDoc(t, db, opts...)` | `ControlledDoc{ID, TenantID, ProfileCode, OwnerUserID, TemplateVersionID}` | Controlled document row seeded via curated bootstrap; depends on taxonomy existing |
| `NewDocument(t, db, opts...)` | `Document{ID, TenantID, RevisionID, ScheduleGeneration}` | One document + initial revision in the correct state; uses `WithStatus`, `WithRevisionVersion`, `WithScheduleGen` to vary state |
| `NewApprovalRoute(t, db, opts...)` | `ApprovalRoute{ID, TenantID, ControlledDocID, ProfileCode}` | Approval route record for a controlled document |
| `NewApprovalInstance(t, db, opts...)` | `ApprovalInstance{ID, TenantID, DocumentID}` | Approval instance (in-progress approval) for a document |

### Option setters (`Opt` values)

Pass zero or more `Opt` values to override defaults:

| Option | Default if omitted |
|--------|--------------------|
| `WithTenant(id string)` | New random UUID |
| `WithUserID(id string)` | New random UUID |
| `WithDisplayName(name string)` | `""` |
| `WithRole(role string)` | `"contributor"` |
| `WithTaxonomy(tax Taxonomy)` | Fresh `NewTaxonomy` |
| `WithControlledDoc(cd ControlledDoc)` | Fresh `NewControlledDoc` |
| `WithDocument(d Document)` | Fresh `NewDocument` |
| `WithRoute(r ApprovalRoute)` | Fresh `NewApprovalRoute` |
| `WithOwner(userID string)` | First user created for the tenant |
| `WithTemplateVersionID(id string)` | From taxonomy |
| `WithName(name string)` | Auto-generated |
| `WithStatus(status string)` | `"draft"` (for documents) |
| `WithCode(code string)` | Random per-test unique |
| `WithProfile(code string)` | From taxonomy |
| `WithRevisionNumber(n int)` | `1` |
| `WithRevisionVersion(n int)` | `1` |
| `WithScheduleGen(g int64)` | `0` |
| `WithEffectiveFrom(at time.Time)` | `time.Now()` |

### Scenario composites

`Scenario` bundles multiple builders into one call for common test setups:

| Composite | Returns | What it seeds |
|-----------|---------|---------------|
| `Scenario{}.PublishedDocument(t, db, opts...)` | `Document` | Full published-document graph: tenant + user + taxonomy + controlled doc + document in `published` state |
| `Scenario{}.ScheduledRevision(t, db, gen, opts...)` | `Document` | Document with a scheduled supersede revision at the given schedule generation |

---

## Handling guarded writes

MetalDocs uses a capability tripwire: certain tables (`documents`, `controlled_documents`, etc.)
require `metaldocs.asserted_caps` to be set for the connection before writes are accepted. Tests
that exercise write paths must satisfy this tripwire through one of the sanctioned helpers.

**Never use raw `set_config` SQL in test files** — that is R1/R2 in `test-discipline.md`.

### Which helper to use

| Situation | Helper |
|-----------|--------|
| SUT takes `*sql.Tx` (caller-managed tx) | `testdb.SeedWithCaps(t, db, capsJSON, func(tx *sql.Tx) error { ... })` |
| SUT takes `*sql.Tx` you already have open | `testdb.SetCapsOnTx(t, tx, capsJSON)` |
| SUT takes `*sql.DB` directly (no tx), `MaxOpenConns=1`, isolated clone | `testdb.SetCapsOnDB(t, db, capsJSON)` |

```go
// Example: SeedWithCaps wrapping a repo write
testdb.SeedWithCaps(t, db, `[{"cap":"document.create"}]`, func(tx *sql.Tx) error {
    return repo.Create(ctx, tx, params)
})

// Example: SetCapsOnDB for pool-level (use only with MaxOpenConns=1)
db.SetMaxOpenConns(1)
testdb.SetCapsOnDB(t, db, `[{"cap":"document.create"},{"cap":"document.edit"}]`)
_, err := repo.CommitUpload(ctx, ...)
```

---

## Qualified table names

When your test issues raw SQL against a module-owned table, use `testdb.Qualified` to prefix the
table name with the schema returned from `testdb.Open`:

```go
db, schema := testdb.Open(t)

var count int
db.QueryRowContext(ctx,
    `SELECT count(*) FROM `+testdb.Qualified(schema, "documents")+` WHERE tenant_id=$1::uuid`,
    tenantID,
).Scan(&count)
```

This prevents the R4 bare-table anti-pattern and ensures the query hits the correct clone schema.

---

## Minimal working example

```go
//go:build integration

package mymodule_test

import (
    "context"
    "testing"

    "metaldocs/internal/modules/documents/infrastructure"
    iamdomain "metaldocs/internal/modules/iam/domain"
    "metaldocs/tests/integration/testdb"
)

func TestCreateDocument_StoresNameCorrectly(t *testing.T) {
    ctx := context.Background()
    db, _ := testdb.Open(t)

    // Seed a full published-document context via Scenario.
    doc := testdb.Scenario{}.PublishedDocument(t, db)

    // Verify via the repository under test.
    repo := repository.New(db, iamdomain.NoopUserDisplayNameReader{})
    got, err := repo.LoadDocument(ctx, doc.TenantID, doc.ID)
    if err != nil {
        t.Fatalf("LoadDocument: %v", err)
    }
    if got.Status != "published" {
        t.Errorf("status = %q, want published", got.Status)
    }
}
```

---

## Other helpers

| Helper | Location | Purpose |
|--------|----------|---------|
| `testdb.DeterministicID(t, suffix)` | `db.go` | Stable UUID derived from the test name + suffix — useful for seeding IDs that must survive an idempotency replay without collision |
| `testdb.InsertDraftDocument(t, db, schema, tenantID)` | `fixtures.go` | Low-level draft document seeder (prefer `NewDocument` / `Scenario` for new tests) |
| `testdb.SeedWithCaps` / `SetCapsOnTx` / `SetCapsOnDB` | `fixtures.go` | Guarded-write helpers (see above) |
| `testdb.OpenAsCIRole(t, dbName)` | `ci_role.go:38` | Opens a second connection to the same per-test clone as `metaldocs_ci` — a dedicated non-owner, `NOSUPERUSER`+`NOBYPASSRLS` role (migration `0284_ci_rls_role.sql`). Use **only** for RLS isolation-proof reads; keep schema setup and row seeding on the owner handle from `testdb.Open` (`metaldocs_app`), because `metaldocs_ci` holds DML-only grants (no DDL). `metaldocs_app` is SUPERUSER+BYPASSRLS+owner, so RLS is inert on that connection — real isolation proofs must go through `OpenAsCIRole`. See [ADR 0027 Amendment 2026-07-05](../decisions/0027-rls-adoption-sequencing.md) and `tests/integration/security/rls_truth_test.go`. |

---

## What NOT to do

| Anti-pattern | Rule | Sanctioned replacement |
|--------------|------|------------------------|
| `db.ExecContext(ctx, "SELECT set_config('metaldocs.asserted_caps'...")` | R1 / R2 | `testdb.SeedWithCaps` / `SetCapsOnTx` / `SetCapsOnDB` |
| `tenantID := "ffffffff-ffff-ffff-ffff-ffffffffffff"` (literal) | R3 | `testdb.NewTenant(t, db).ID` or `tenant.DevTenantID` const |
| `FROM documents WHERE ...` (bare table name in SQL) | R4 | `FROM `+testdb.Qualified(schema, "documents")+` WHERE ...` |
| `db.ExecContext(ctx, "SET search_path TO metaldocs, public")` | — | Not needed: `testdb.Open` sets up the correct `search_path` on the clone |

See [test-discipline.md](test-discipline.md) for the full per-rule CI enforcement reference.
