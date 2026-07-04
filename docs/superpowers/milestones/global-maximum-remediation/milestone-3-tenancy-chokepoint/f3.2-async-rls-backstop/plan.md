# F3.2 plan — async RLS backstop

> Engine: `superpowers:subagent-driven-development` (TDD). Contract: `../validation-contract.md` §2.
> Runs AFTER F3.1 is committed (shares `scripts/api-lint` registration + reads the F3.1 chokepoint).
> No RLS policy change; no new migration for the primitive; integration test is Go-only.

## Task list (ordered)

### T1 — `SeedTxTenant` primitive (`internal/modules/iam/authz/context.go`)
- Add `func SeedTxTenant(ctx context.Context, tx *sql.Tx, tenantID string) error` beside `SeedTxIdentity`:
  trims tenantID; empty → `ErrTenantContextMissing`; runs
  `SELECT set_config('metaldocs.tenant_id', $1, true)` (tenant only, NO actor).
- **TDD (PG-1):** unit test — sets tenant GUC; empty → error; does NOT touch `metaldocs.actor_id`.

### T2 — Seed the 5 single-tenant processing txs (contract §2.2)
Seed `SeedTxTenant(ctx, tx, <row/payload tenant_id>)` at tx start, before first lock/write:
1. `internal/platform/worker/materialize_job_runner.go` — existing `BeginTx` tx; seed with `payload.TenantID`.
2. `internal/platform/worker/pdf_job_runner.go` — **wrap** the raw `ExecContext` write(s) in `BeginTx`
   (or the shared `runner.Do`); seed with payload tenant; commit.
3. `internal/modules/documents/approval/jobs/scheduled_publish_job.go` — existing `runner.Do`; seed with
   River-args tenant BEFORE the `FOR UPDATE` (H-PRE-1; add NO `authz.Require`).
4. `internal/modules/notifications/infrastructure/fanout_worker.go` — **wrap** the insert loop in a tx;
   seed with event-args tenant.
5. `internal/modules/render/fanout/staging_outbox_worker.go` — processing step per claimed row; seed with
   `OutboxRow.TenantID` in the processing tx. **Claim** step (`staging_outbox.go`) stays GUC-unset.
- **Constraint check:** each seeded tx must touch exactly ONE tenant's rows (ADR 0054 rule 2). If any
  handler mixes tenants in one tx → STOP, report `HS-2` (do not workaround).
- **TDD (PG-2):** targeted handler tests compile+green; assert seed call precedes first write/lock.

### T3 — `ASYNC-TENANT-SEED` blocking lint (`scripts/api-lint`)
- New rule (mirror `tripwire_arm_rules.go`; register in `RunCodeRules` beside F3.1's `SEED-CHOKEPOINT`).
  AST-scan worker/jobs handler packages: a DB write (`Exec`/`ExecContext` INSERT/UPDATE/DELETE) against a
  **tenant-scoped table** (the RLS-covered set) must be in a function that also calls
  `SeedTxTenant`/`SeedTxIdentity`, unless the site is on `scripts/api-lint/async-tenant-seed-allowlist.txt`.
  Violation → file:line + table. Handler-local only (documented limitation; not call-graph).
- **Tenant-table set:** derive from the RLS-covered table list (the 33 FORCE-RLS tables in
  `db/baseline/0001_current_schema.sql`) — encode as a checked-in list the rule reads (mirror how
  tripwire rules read their registry). Keep it a data file, not hardcoded, so it can't silently drift.
- **TDD (PG-3):** unit test — clean fixture (seeded handler + allowlisted claim) = 0; negative fixture
  (unseeded tenant-scoped worker write) = 1 violation naming file:line+table.

### T4 — Sanctioned allowlist (`scripts/api-lint/async-tenant-seed-allowlist.txt`)
- Enumerate (mirror `tripwire-allowlist.txt`): outbox claim (`.../messaging/outbox/postgres/consumer.go`),
  staging-outbox claim (`render/fanout/staging_outbox.go`), stuck-watchdog list, `idempotency_keys` janitor,
  `job_leases` lease-reaper, audit-integrity scan — each with `path:line  # category + reason`.

### T5 — Negative RLS integration proof (`//go:build integration`, testdb factory)
- New drive mirroring existing tenant-isolation tests (find them: `grep -rl "tenant_isolation\|Isolation" --include=*_test.go`).
  Two tenants A/B each with a `documents` row. **Leak-before:** GUC-unset tx cross-reads/updates B's row →
  succeeds (capture). **Backstop-after:** after `SeedTxTenant(ctx, tx, A)` — SELECT/UPDATE/DELETE of B's row
  → 0 rows; INSERT/UPDATE producing a B-row → error `new row violates row-level security policy` (`42501`).
  Label real-DB (testdb), not sqlmock.
- **Run scope:** `go test -tags integration -run 'RLS|Tenant|Isolation' ./...` ONLY. If box can't run
  `-tags integration`, author the drive and record a **bounded defer** with run-trigger (M1/M2 precedent).

### T6 — Gate run + evidence
- `go build ./...`; targeted handler + authz + api-lint tests; `go run ./scripts/api-lint -strict ...` (0);
  PG-3 RED→GREEN captured; PG-4 leak→blocked captured (or authored+deferred); scheduled-publish + fanout
  targeted drives green.

## Files expected to change
- `internal/modules/iam/authz/context.go` (+ test) — SeedTxTenant
- 5 async handler files (+ targeted tests); pdf + fanout gain a tx wrapper
- `scripts/api-lint/` new rule + test + testdata; `async-tenant-seed-allowlist.txt`; tenant-table list file
- new `//go:build integration` negative-proof test file

## Risk / ordering notes
- pdf + notifications-fanout wrapping raw `ExecContext` in a tx is the behavior-shape change — verify the
  wrapped writes still commit and the handler's error/retry path is unchanged (idempotent consumer).
- H-PRE-1: seed is `SET LOCAL` only; seed before `FOR UPDATE`; never add `authz.Require` to a locked tx.
- Tenant-mixing discovery in any handler = HS-2 stop, not a silent per-row workaround.
