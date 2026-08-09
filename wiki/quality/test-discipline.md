# Integration Test Discipline Rules

> Last verified: 2026-06-15 — F4c.4 (CI grep-guard landed, script + workflow step live).
> Enforced by: `scripts/check-test-discipline.sh` (registry check `test-conventions`, run by `tools/verify` in `ci.yml` on PR→main since the 2026-08 CI restructure).

Four rules govern integration test files (`//go:build integration` first line) in MetalDocs. They
were established in Milestone 4c (F4c.4) after the unified `testdb` factory framework (F4c.1) made
the sanctioned patterns available. The guard fails CI on any new violation; the allowlist (in the
script) tracks pre-F4c.4 debt — it can only shrink.

> **Related policy:** when a *legacy* test breaks or blocks a change, triage it via
> [legacy-test-policy.md](legacy-test-policy.md) (repair-class vs delete-class) before repairing
> or deleting anything.

---

## The Rules

### R1 — No inline `set_config('metaldocs.asserted_caps')` in test files

**Forbidden in `*_test.go` (outside `tests/integration/testdb/`):**
```sql
SELECT set_config('metaldocs.asserted_caps', '[{"cap":"document.create"}]', false)
```

**Why:** the literal SQL string in a test file is an anti-pattern — it bypasses the framework and
is easy to misuse (wrong `is_local` value, wrong caps JSON, leaks session state).

**Sanctioned patterns:**
```go
// When SUT takes *sql.Tx — wrap with:
testdb.SeedWithCaps(t, db, `[{"cap":"document.create"}]`, func(tx *sql.Tx) error {
    return repo.SomeMethod(ctx, tx)
})

// When SUT takes *sql.Tx you already own:
testdb.SetCapsOnTx(t, tx, `[{"cap":"document.create"}]`)

// When SUT takes *sql.DB directly (must use MaxOpenConns=1, isolated DB):
testdb.SetCapsOnDB(t, db, `[{"cap":"document.create"}]`)
```

**Exception zone:** `tests/integration/testdb/**` — the framework itself uses the literal.

---

### R2 — No `is_local=false` in a `set_config` call

**Forbidden in `*_test.go` (outside `tests/integration/testdb/`):**
```sql
SELECT set_config('metaldocs.asserted_caps', '...', false)
```

**Why:** `is_local=false` sets capabilities at session level. On a pool with multiple connections
this leaks across tests. The correct path (`SeedWithCaps`) uses `is_local=true` to discard the
assertion on tx commit. Session-level is safe only when `MaxOpenConns=1` and the DB is dropped
after the test — use `testdb.SetCapsOnDB` which documents this invariant at the call-site.

**Exception zone:** same as R1.

---

### R3 — No hardcoded `DevTenantID` literal

**Forbidden in `*_test.go` (outside `tests/integration/testdb/`):**
```go
tenantID := "ffffffff-ffff-ffff-ffff-ffffffffffff"   // hardcoded literal
```

**Why:** hardcoding the UUID value means the test depends on the dev-sentinel tenant, which may
not exist in a fresh test-clone DB. Tests should either:
- Use `factory.NewTenant(t, db)` for a random, factory-seeded tenant.
- Use the Go constant `tenant.DevTenantID` (from `internal/platform/tenant`) if the sentinel is
  genuinely required, so the value is one place.

**Sanctioned patterns:**
```go
tn := testdb.NewTenant(t, db)           // random per-test tenant
tenantID := tenant.DevTenantID          // sentinel const — not the literal
```

---

### R4 — No bare unqualified `documents` table reference in test SQL

**Forbidden in `*_test.go` (outside `tests/integration/testdb/`):**
```sql
FROM documents WHERE id=$1
JOIN documents d ON d.id = ...
```

**Why:** after M4b legacy-schema teardown, a bare `documents` identifier may resolve to the dead
legacy `metaldocs.documents` table under the wrong `search_path`, causing subtle failures. Tests
must qualify the table via the schema returned from `testdb.Open`:
```go
db, schema := testdb.Open(t)
// ...
db.QueryRowContext(ctx,
    `SELECT id FROM `+testdb.Qualified(schema, "documents")+` WHERE id=$1::uuid`, id)
```

Or use factory-returned IDs (no raw SQL against `documents` needed for most tests).

---

## Allow-list (pre-F4c.4 debt)

The following files contain violations that pre-date the F4c.4 guard. They are tracked in the
script's `R3_ALLOWLIST` / `R4_ALLOWLIST` arrays and must be cleaned up before the corresponding
entry is removed from the list. The list **must only shrink**; additions require operator approval.

| File | Rule | Reason |
|------|------|--------|
| `auth/infrastructure/postgres/sessions_admin_integration_test.go` | R3 | F4.4 migrated port; test still uses literal tenant UUIDs. |
| `iam/infrastructure/postgres/role_provider_integration_test.go` | R3 | Not in F4c.3 cluster scope. |
| `tests/integration/approval/eligibility_test.go` | R3, R4 | M4b post-teardown debt (tests/integration/scenarios/). |
| `iam/integration_test.go` | R4 | Not in F4c.3 cluster scope. |
| `platform/migrate/revision_number_zero_based_integration_test.go` | R4 | Migration test — fixture pattern, not module test. |

---

## Running locally

```bash
bash scripts/check-test-discipline.sh
```

Exit 0 = clean. Exit 1 = violations printed as `path:line: <rule>: <text>`.

Works from repo root via git-bash on Windows or any POSIX shell. No external deps.

---

## CI integration

The guard runs as the second step in `.github/workflows/module-boundaries.yml` on every PR→main.
A violation fails the `conformance` job; the PR cannot merge until the violation is fixed or
(rarely, with operator approval) added to the allowlist.
