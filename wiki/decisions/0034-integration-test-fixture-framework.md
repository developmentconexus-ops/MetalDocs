# ADR 0034 — Integration Test Fixture Framework (IntegreSQL Template-DB-Per-Test)

> **Status:** Accepted 2026-06-15
> **Last verified:** 2026-06-15
> **Scope:** The canonical integration-test harness for MetalDocs: IntegreSQL template-DB-per-test
> isolation, the `testdb` factory package, the four discipline rules (R1–R4), and the CI guard that
> enforces them. Does **not** cover unit-test mocking strategy (separate concern) or future
> per-module harness extensions beyond the factory.
> **Key files:**
> - `tests/integration/testdb/factory.go` — exported factory builders (entity + composite + Opt setters)
> - `tests/integration/testdb/db.go` — `Open(t)`, `DeterministicID`, template-DB lifecycle
> - `tests/integration/testdb/fixtures.go` — `SeedWithCaps`, `SetCapsOnTx`, `SetCapsOnDB`, `Qualified`
> - `scripts/check-test-discipline.sh` — CI grep guard (R1–R4, allowlists, scope filter)
> - `.github/workflows/module-boundaries.yml` — workflow that runs the guard on PR→main

## Context

Milestone 4b (legacy-schema teardown) performed a `/systematic-debugging` pass that surfaced
two co-occurring failure classes in the integration test suite:

**Cross-test state leakage (shared-DB pool).**
Multiple test files called `SELECT set_config('metaldocs.asserted_caps', '...', false)` directly.
`is_local=false` sets the GUC at session scope; on a shared pool, a second test's connection
inherits the asserted capabilities from the first test's session, making the second test pass
or fail based on execution order, not its own setup.

**`search_path` resolution drift (bare table names).**
Connections configured with a `metaldocs`-first `search_path` (the dev/test DSN default) resolved
bare `documents` SQL to `metaldocs.documents` — a dead early-editor-era table that lacks
`tenant_id` / `active_session_id` / `controlled_document_id` — producing `column "tenant_id"
does not exist` (SQLSTATE 42703). The root fix (ADR 0032) removed the dead table; the harness fix
is to ensure every test uses a schema-isolated clone and qualifies its SQL.

Three alternatives were evaluated:

| Alternative | Ruling |
|-------------|--------|
| **Shared DB with truncation** (existing state) | Ruled out: does not eliminate session-level GUC leakage; `search_path` on a shared pool is set once at connection time and does not reset between tests |
| **Per-test migration run** | Ruled out: migration run time is O(number of migrations) and already exceeds 4s locally; this compounds with test count; no isolation benefit over cloning |
| **IntegreSQL template-DB-per-test cloning** | Selected: each test gets its own isolated Postgres database cloned from a pre-migrated template; a new connection is opened from scratch; no GUC leakage; `search_path` is the template DB's `search_path`; clone is dropped at `t.Cleanup` |

The root-cause analysis was recorded in the systematic-debugging pass evidence
(`milestone-4-systemic-ports/f4.1b-testdb-search-path-robustness/`) and the M4b teardown
evidence. The fix-not-adapt hard-stop (CLAUDE.md §4) prohibited adapting the `search_path`
in individual tests as a symptom patch.

## Decision

**1. `tests/integration/testdb/` is the canonical home for the integration test harness.**
No test infrastructure is permitted outside this package. The `pgtest` package
(`internal/testsupport/pgtest/`) is retired (zero callers — deleted in M4c F4c.4).

**2. Template-DB-per-test via IntegreSQL is the harness.**
`testdb.Open(t *testing.T) (*sql.DB, string)` opens a fresh clone of the curated baseline
(all migrations applied + reference data + capability tripwire caps) for every test. The
`*sql.DB` points to an isolated database. The clone is dropped when `t.Cleanup` fires.
`INTEGRESQL_URL` must be set in the environment (CI and dev); tests skip gracefully when it
is absent.

**3. Factory builders are the canonical seeding mechanism.**
`factory.go` exports: `NewTenant`, `NewUser`, `NewTaxonomy`, `NewControlledDoc`, `NewDocument`,
`NewApprovalRoute`, `NewApprovalInstance`, plus composites `Scenario{}.PublishedDocument` and
`Scenario{}.ScheduledRevision`. Tests must use these (or `DeterministicID` / raw SQL scoped via
`Qualified`) to populate the clone. No test may replicate factory logic inline.

**4. Four discipline rules (R1–R4) apply to all integration test files.**
Scope: any `*_test.go` file whose first line is `//go:build integration`, excluding `tests/integration/testdb/**`.

- **R1:** No inline `set_config('metaldocs.asserted_caps', ...)` SQL in test files.
- **R2:** No `set_config(...)` call with `is_local=false` in test files.
- **R3:** No hardcoded `DevTenantID` literal (`"ffffffff-ffff-ffff-ffff-ffffffffffff"`) in test files.
  Use `factory.NewTenant(t, db)` or the `tenant.DevTenantID` constant.
- **R4:** No bare unqualified `documents` table reference in test SQL.
  Use `testdb.Qualified(schema, "documents")` with the schema returned from `testdb.Open(t)`.

Sanctioned guarded-write patterns:
- `testdb.SeedWithCaps(t, db, capsJSON, func(tx *sql.Tx) error {...})` — wraps a tx with `is_local=true`
- `testdb.SetCapsOnTx(t, tx, capsJSON)` — on an already-open tx
- `testdb.SetCapsOnDB(t, db, capsJSON)` — pool-level; safe only with `MaxOpenConns=1` on an isolated clone

**5. The CI guard `scripts/check-test-discipline.sh` enforces R1–R4 on PR→main.**
The guard runs as the second step in `.github/workflows/module-boundaries.yml`. A violation
fails the `conformance` job; the PR cannot merge until the violation is fixed or added to the
allowlist with operator approval. The allowlist can only shrink.

**6. Pre-F4c.4 allowlist files (5 entries) are tracked debt.**
Five files contain pre-existing R3/R4 violations committed before this framework was established.
They are tracked in the guard script's allowlist and must be cleaned up on next structural touch
(see `wiki/quality/test-discipline.md` for the full table). No new entries may be added without
operator approval.

## Consequences

**Positive:**
- Every integration test starts from a clean, schema-correct clone; cross-test state leakage
  is structurally impossible.
- Parallel test execution is safe (`-parallel N`): each test has its own database.
- `search_path` resolution drift is eliminated: the clone's `search_path` matches the template's,
  which is correct by construction.
- Factory builders eliminate one-off seed logic and guard against schema drift: update the
  factory, tests follow.
- CI guard prevents regression: a developer adding raw `set_config` or a bare `documents`
  reference gets an immediate CI failure, not a flaky test discovered in production.

**Negative / Constraints:**
- `INTEGRESQL_URL` is a new runtime dependency for the test environment. CI must provision an
  IntegreSQL daemon. Tests skip (not fail) without it, which may mask coverage gaps in minimal
  CI configurations.
- Template DB creation is one-time per test-runner session; if the curated baseline is updated,
  IntegreSQL must be signaled to regenerate the template (or the daemon restarted).
- `testdb.Open` is not usable in unit tests or without a Postgres connection. Tests requiring
  real DB must carry the `//go:build integration` tag and cannot be run offline.

**Deferred:**
- Allowlist cleanup for 5 pre-F4c.4 files (trigger: next structural touch).
- CI wiring of full integration test suite (with operator DSN) — distinct from the guard step.

## References

- [ADR 0032](0032-drop-legacy-mddm-document-cluster.md) — root-cause fix (dead table removal) that motivated the harness decision
- [wiki/quality/integration-test-harness.md](../quality/integration-test-harness.md) — how-to guide for developers writing new integration tests
- [wiki/quality/test-discipline.md](../quality/test-discipline.md) — R1–R4 rules, allowlist, CI usage reference
- M4c F4c.1 evidence — factory.go shipped; M4c F4c.4 evidence — CI guard shipped
