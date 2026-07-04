# F3.5 evidence — negative-RLS proof real-green

## Root cause (confirmed, not re-litigated)
`internal/modules/iam/authz/seed_tx_tenant_rls_integration_test.go` targeted `public.documents`.
`documents` carries the M2 capability write-tripwire (`trg_require_cap_asserted` →
`enforce_capability_asserted`, `0001_current_schema.sql:3886`): any INSERT/UPDATE with
`metaldocs.asserted_caps` unset raises `P0001 ErrCapabilityNotAsserted` **before** the RLS
`tenant_isolation` policy becomes the deciding control. On the operator-mandated real run this made
`leak_before_no_seed` and the retenant `42501` assertion both observe P0001 instead of an RLS outcome.
The core subtest `blocked_after_seed_tenant_a/select_update_delete_see_zero_rows` already passed — the
RLS backstop itself was sound; only the proof's target table was wrong.

## The fix (T1)
Retargeted the proof from `public.documents` to `metaldocs.notifications`:
- `metaldocs.notifications` is a FORCE-RLS `tenant_isolation` table
  (`db/baseline/0001_current_schema.sql:1437,1453,4632-4635`).
- It is a real F3.2 async seed site (notifications fanout worker writes).
- It carries **no** capability write-tripwire, so RLS is the sole deciding control end-to-end.

Edits to `internal/modules/iam/authz/seed_tx_tenant_rls_integration_test.go`:
- Row builders: `testdb.NewDocument(t, db, testdb.WithTenant(tntA.ID/tntB.ID))` →
  `testdb.NewNotification(t, db, testdb.WithTenant(tntA.ID/tntB.ID))`; `docA`/`docB` renamed to
  `notifA`/`notifB` throughout.
- Role grants: added `GRANT USAGE ON SCHEMA metaldocs TO rls_tester_async_tenant_seed` (kept the existing
  `GRANT USAGE ON SCHEMA public` — harmless) and replaced
  `GRANT SELECT, UPDATE, INSERT, DELETE ON public.documents` with the same grant on
  `metaldocs.notifications`. Role unchanged: `rls_tester_async_tenant_seed`, `NOLOGIN NOBYPASSRLS`.
- DML: every `public.documents` reference → `metaldocs.notifications`. Mutating column changed from
  `name` to `status` (the only mutable column notifications has that isn't the tenant/id key):
  `SET status = 'SENT'` for the leak-before demonstration, `SET status = 'READ'` for the blocked-after
  demonstration (both satisfy `notifications_status_check` = `PENDING|SENT|READ`). `WHERE id = $1::uuid`
  kept identical. Retenant assertion:
  `UPDATE metaldocs.notifications SET tenant_id = $2::uuid WHERE id = $1::uuid` with
  `notifA.ID, tntB.ID` — asserted via `strings.Contains(err.Error(), "42501")`, unchanged.
- Header doc comment rewritten to state the table is `metaldocs.notifications` and explain why not
  `documents` (M2 capability write-tripwire would fire first and mask the RLS-only outcome).
- Three-subtest shape kept identical: `leak_before_no_seed` (1-row leak) →
  `blocked_after_seed_tenant_a/select_update_delete_see_zero_rows` (0/0, A visible) →
  `blocked_after_seed_tenant_a/insert_or_update_producing_b_row_is_42501` (SQLSTATE 42501).

No change to `authz.SeedTxTenant`, any RLS policy, any migration, or any other production `.go`/`.sql`.

## Real run (T2 — binding, not skipped, not faked)
Command executed via the `.env`-loading, secret-redacting PowerShell wrapper:

```
& "...\scratchpad\run-rls-proof.ps1"
```

which runs:

```
go test -tags integration -count=1 -v -run 'SeedTxTenant_RLSBackstop' ./internal/modules/iam/authz/...
```

against live Postgres (container `metaldocs-postgres`, host port from `.env` `POSTGRES_HOST_PORT`).

### Verbatim redacted output

```
Postgres: localhost:5433 db=metaldocs user=metaldocs_app (password redacted)
Running F3.2 negative-RLS integration proof...
=== RUN   TestSeedTxTenant_RLSBackstop_LeakBeforeBlockedAfter
=== RUN   TestSeedTxTenant_RLSBackstop_LeakBeforeBlockedAfter/leak_before_no_seed
=== RUN   TestSeedTxTenant_RLSBackstop_LeakBeforeBlockedAfter/blocked_after_seed_tenant_a
=== RUN   TestSeedTxTenant_RLSBackstop_LeakBeforeBlockedAfter/blocked_after_seed_tenant_a/select_update_delete_see_zero_rows
=== RUN   TestSeedTxTenant_RLSBackstop_LeakBeforeBlockedAfter/blocked_after_seed_tenant_a/insert_or_update_producing_b_row_is_42501
--- PASS: TestSeedTxTenant_RLSBackstop_LeakBeforeBlockedAfter (91.38s)
    --- PASS: TestSeedTxTenant_RLSBackstop_LeakBeforeBlockedAfter/leak_before_no_seed (0.01s)
    --- PASS: TestSeedTxTenant_RLSBackstop_LeakBeforeBlockedAfter/blocked_after_seed_tenant_a (0.01s)
        --- PASS: TestSeedTxTenant_RLSBackstop_LeakBeforeBlockedAfter/blocked_after_seed_tenant_a/select_update_delete_see_zero_rows (0.00s)
        --- PASS: TestSeedTxTenant_RLSBackstop_LeakBeforeBlockedAfter/blocked_after_seed_tenant_a/insert_or_update_producing_b_row_is_42501 (0.00s)
PASS
ok  	metaldocs/internal/modules/iam/authz	94.216s

EXITCODE=0
```

leak_before_no_seed reproduced the leak (1-row unseeded cross-tenant UPDATE succeeded) exactly as
required — the backstop is not being engaged some other way; the RLS policy on `metaldocs.notifications`
is the sole control observed, matching the design.

## Other verification
- `go build ./...` → clean, no output (exit 0).
- `go vet -tags integration ./internal/modules/iam/authz/...` → clean, no output (exit 0).

## PG-1..4 disposition
- **PG-1 (table + grants retargeted):** PASS. No `public.documents` write remains in the test (only the
  header comment's rationale mentions `documents` by name, explaining why it is *not* used). Role grants
  are on `metaldocs.notifications` (+ `GRANT USAGE ON SCHEMA metaldocs`).
- **PG-2 (real run, binding):** PASS. See verbatim output above — `ok`, all subtests green, leak-before=1
  row, blocked-after=0/0 with A visible, retenant=42501.
- **PG-3 (no production code diff):** PASS. `git diff --stat` for this task's edit touches only
  `internal/modules/iam/authz/seed_tx_tenant_rls_integration_test.go` plus this feature's own
  `docs/.../f3.5-rls-proof-real-green/` folder. (A pre-existing, out-of-scope
  `docs/superpowers/milestones/global-maximum-remediation/README.md` diff was already present before this
  task started and belongs to the main session, per task instructions — not touched here.)
- **PG-4 (build/vet green):** PASS. `go build ./...` and `go vet -tags integration
  ./internal/modules/iam/authz/...` both clean.

## Self-assessment
**PASS.** Fix-feature F3.5 is test-construction-only, real-run GREEN, no production/RLS/migration diff.
