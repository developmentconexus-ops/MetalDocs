> ⛔ **SUPERSEDED 2026-06-15 — DO NOT IMPLEMENT.** A `/systematic-debugging` pass proved the harness
> `search_path` adaptation specced below is a **symptom patch**. Root cause: a dead legacy
> `metaldocs.documents` duplicate (anchoring a 7-table dead FK cluster) + a dead `template_audit_log`
> duplicate that shadow the real `public.*` tables under metaldocs-first `search_path`. Operator chose
> the root-cause fix: **milestone 4b** (`milestone-4b-legacy-schema-teardown`) drops the dead cluster;
> `tests/integration/testdb/db.go` stays at **HEAD** (no harness change — that empty diff is the
> fix-not-adapt proof). The Non-goal below ("Not deleting the dead `metaldocs.documents` legacy
> duplicate") is exactly what 4b reverses. See `evidence.md` and the README HS-6 entry 2026-06-15.

# Feature F4.1b — testdb search-path robustness (HS-4 fix feature) — SUPERSEDED

> **Milestone:** 4 (Systemic Ports)  ·  **Feature:** `f4.1b-testdb-search-path-robustness`
> **Origin:** HS-4 — the M4 milestone-validator (`qa/milestone-qa.md`, HEAD 30503533) returned
> **FAIL** on a single isolated gate: **F4.1a Gate #5**
> (`TestCreateDocumentTx_PopulatesAllSnapshotColumns`) is **environment-coupled**. The H-G
> class-zero bar itself PASSED; this feature repairs only the test-harness defect the validator named.
> **Approved:** 2026-06-15 (operator standing authorization; named fix feature per HS-4).

## Problem (root cause)

The integration test harness `tests/integration/testdb/db.go` clones a per-test isolated database
from a curated-baseline template. The baseline contains **two** `documents` tables:
`public.documents` (the real one, has `tenant_id` / `controlled_document_id`) and
`metaldocs.documents` (a dead legacy duplicate lacking those columns).
`TestCreateDocumentTx_PopulatesAllSnapshotColumns` resolves bare `documents` and therefore must pin
`search_path` to `public, metaldocs`. It does so with
`ALTER DATABASE <db> SET search_path TO public, metaldocs` + idle-connection eviction.

`testdb.openDBWithDatabase` builds every connection from the operator DSN via `pgx.ParseConfig`,
overriding only `cfg.Database`. The operator's dev DSN carries `?search_path=metaldocs,public`, which
`ParseConfig` stores in `cfg.RuntimeParams["search_path"]` and pgx sends as a **connection startup
parameter**. A connection-level startup param **overrides** the per-database `ALTER DATABASE` default.
So with that DSN, every test-DB connection resolves bare `documents` → `metaldocs.documents` (legacy),
and the snapshot INSERT fails:

```
column "tenant_id" of relation "documents" does not exist (SQLSTATE 42703)
```

The test passes **only** when the DSN omits `search_path` — i.e. it is coupled to operator
environment, not deterministic. F4.1a's evidence reported an unconditional `ok`; under the operator
DSN the gate fails. (C2: environment-coupled is not green; C6 forbidden-list.)

## Consumer contract

The **consumer** is the integration-test layer (`tests/integration/testdb.Open`). Required, DSN-independent behavior:

1. **Bare-name resolution is governed by the test DB's own default, never by the operator DSN's
   `search_path` query parameter.** Two operator DSNs that differ only in a `search_path` param MUST
   produce identical bare-name resolution for every isolated test database.
2. **Canonical default unchanged:** absent any per-test override, a freshly `Open`ed isolated database
   resolves bare names with `search_path = metaldocs, public` — the effective default today's passing
   tests rely on. (No regression for the 3 tests that set no `search_path` of their own:
   `repository_revision_history_integration_test.go`, `iam/authz/authz_bypass_test.go`,
   `search/.../reader_visibility_integration_test.go`.)
3. **Per-test override actually takes effect:** a test that runs
   `ALTER DATABASE <db> SET search_path TO public, metaldocs` (+ idle-conn eviction) gets
   `public, metaldocs` on subsequent pool connections — because no startup param overrides it anymore.
4. The template-bootstrap and admin paths are unaffected (baseline is fully schema-qualified; verified
   — every `CREATE TABLE` names its schema, no top-level `SET search_path`).

## Non-goals

- **Not** deleting the dead `metaldocs.documents` legacy duplicate (separate DB-ownership concern;
  out of M4 H-G class — would be a `metaldocs-database` task).
- **Not** changing any production wiring, the iam ports, or any non-test runtime code.
- **Not** re-litigating the F4.1 value proof's *content* — only making its gate deterministic.
- **Not** rewriting the per-connection `SET search_path` idiom used by other passing tests.

## Validation Gate

| # | Acceptance | Proof |
|---|-----------|-------|
| A | RED first: snapshot test fails with `42703` under operator DSN **with** `search_path=metaldocs,public`, pre-fix | `go test -tags integration -run TestCreateDocumentTx_PopulatesAllSnapshotColumns ./internal/modules/documents/application/` with the operator DSN → FAIL (captured) |
| B | GREEN: same command, same operator DSN, post-fix → `ok` | rerun → PASS |
| C | DSN-independence: the same test passes with a DSN that **omits** `search_path` too | rerun with stripped DSN → PASS |
| D | No regression — full documents/application + documents/repository + templates infra integration green under operator DSN | `go test -tags integration ./internal/modules/documents/... ./internal/modules/templates/...` |
| E | No regression — the 3 no-`search_path` tests still pass under operator DSN | run revision-history, authz_bypass, reader_visibility |
| F | build + vet clean | `go build ./...`; `go vet -tags integration ./tests/integration/... ./internal/modules/documents/...` |

## Interview record

No operator interview needed — the validator named the exact fix feature and the contract is read
directly from the failing consumer (the snapshot test) + the harness. Fix approach (strip the DSN
`search_path` runtime param in `openDBWithDatabase`; set the canonical default via `ALTER DATABASE` in
`Open`) is the minimal root-cause change; recorded in `plan.md`.
