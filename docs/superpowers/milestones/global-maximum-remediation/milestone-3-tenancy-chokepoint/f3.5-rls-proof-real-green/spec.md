# F3.5 — negative-RLS proof real-green (HS-4 fix feature)

> **Milestone:** M3 · **Opened by:** operator-mandated REAL run of the F3.2 negative-RLS proof (no defer)
> → the deferred proof was **RED** on first real execution. **Test-construction fix only** — no change to
> the `SeedTxTenant` primitive, RLS policy, or any production code.

## Root cause (real-run RED)
`internal/modules/iam/authz/seed_tx_tenant_rls_integration_test.go` targets `public.documents`. `documents`
carries the M2 capability **write-tripwire** (`trg_require_cap_asserted` → `enforce_capability_asserted`,
`0001_current_schema.sql:3886`): any INSERT/UPDATE with `metaldocs.asserted_caps` unset raises
`P0001 ErrCapabilityNotAsserted` **before** the RLS `tenant_isolation` policy is the deciding control. So:
- `leak_before_no_seed`: the unseeded cross-tenant UPDATE errors P0001 instead of leaking 1 row.
- `insert_or_update_producing_b_row_is_42501`: the retenant UPDATE errors P0001 instead of RLS 42501.
The core subtest `blocked_after_seed_tenant_a/select_update_delete_see_zero_rows` **passed** (0-row
SELECT/UPDATE/DELETE under `SeedTxTenant(A)`) — the RLS backstop itself is sound; the proof table was wrong.

## Consumer contract
Consumer: the milestone-validator + a future maintainer reading the M3 async-RLS-backstop evidence. They
need a **real, green** end-to-end proof that `authz.SeedTxTenant` engages FORCE RLS on a **tenant-scoped,
non-tripwired** table that async fleets actually write — so RLS is the *sole* deciding control.

**Required end-state:**
1. Retarget the proof from `documents` → **`metaldocs.notifications`**: FORCE-RLS `tenant_isolation` table
   (`0001_current_schema.sql:1437,1453`), a **real F3.2 async seed site** (notifications fanout worker),
   and NOT in the `trg_require_cap_asserted` set. Build rows via `testdb.NewNotification(WithTenant(...))`.
2. Assertions, unchanged in shape (contract §2.5), now valid because no capability tripwire intercepts:
   - **leak_before_no_seed** (unseeded tx, NOBYPASSRLS role): B's notification visible; cross-tenant
     `UPDATE ... SET status='SENT' WHERE id=$B` affects **1** row (NULL-permissive leak).
   - **blocked_after_seed_tenant_a** (`SeedTxTenant(A)` first): B invisible; cross-tenant UPDATE and DELETE
     of B's row affect **0** rows; A's own row visible.
   - **retenant_is_42501**: `UPDATE metaldocs.notifications SET tenant_id=$B WHERE id=$A` → SQLSTATE **42501**
     (WITH CHECK / USING violation), not P0001.
3. RLS role `rls_tester_async_tenant_seed`: `GRANT USAGE ON SCHEMA metaldocs` + `GRANT SELECT,UPDATE,INSERT,
   DELETE ON metaldocs.notifications` (schema/table corrected from public.documents).
4. The proof is **run for real, GREEN**, under a NOBYPASSRLS role against live Postgres. Evidence records
   the real run (redacted output). The "live run deferred" label is **removed** from F3.2 evidence + the
   validation-contract §2.5 defer note.

## Non-goals
- No change to `authz.SeedTxTenant`, the chokepoint, RLS policy, or any `.sql`/migration/production `.go`.
- No change to the acceptance shape (§2.5 leak-before / blocked-after / 42501) — only the table it runs on.
- Do not weaken the assertions to force green. If notifications does not reproduce leak-before, STOP.

## Validation gate
- **PG-1:** the test targets `metaldocs.notifications` (no `documents` write remains in it); role grants on
  `metaldocs.notifications`.
- **PG-2 (REAL RUN, binding):** `go test -tags integration -count=1 -run 'SeedTxTenant_RLSBackstop'
  ./internal/modules/iam/authz/...` against live Postgres → **PASS**, all subtests green (leak-before=1 row,
  blocked-after=0 rows ×2, retenant=42501). Redacted real output in evidence.
- **PG-3:** no production code diff — `git diff` touches only the test file + M3 docs (evidence,
  validation-contract defer note, F3.2 evidence label, README).
- **PG-4:** `go build ./...` green; `go vet` clean on the test package.

## Named proof commands
- `pwsh scripts/.../run wrapper` loading `.env` → `METALDOCS_DATABASE_URL` (secret redacted) → the
  `go test -tags integration -run SeedTxTenant_RLSBackstop` above → PASS.
- `git diff --stat` → test file + docs only.

## Interview record
| Q | A |
|---|---|
| Why did the real run fail? | `documents` capability write-tripwire (P0001) fires before RLS; proof conflated two controls. |
| Fix by satisfying asserted_caps on documents, or change table? | Change table — `notifications` is a non-tripwired FORCE-RLS tenant table AND a real async seed site → RLS is the sole control, strictly more representative. |
| Any production/policy change? | None. Test-construction fix only; RLS + SeedTxTenant byte-identical. |
| Keep the defer? | No. Operator mandate: close only after a real green run. Defer label removed. |
