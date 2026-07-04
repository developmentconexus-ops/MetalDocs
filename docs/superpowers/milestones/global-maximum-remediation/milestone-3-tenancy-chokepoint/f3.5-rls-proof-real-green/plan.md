# F3.5 plan — negative-RLS proof real-green

> Contract: `spec.md`. Test-construction fix only. Bug caught by operator-mandated REAL run.

## Tasks (ordered)
### T1 — Retarget the proof to `metaldocs.notifications`
File: `internal/modules/iam/authz/seed_tx_tenant_rls_integration_test.go`.
- Replace the two `testdb.NewDocument(WithTenant(...))` builders with
  `testdb.NewNotification(testdb.WithTenant(tntA.ID))` / `(tntB.ID)` → `docA`/`docB` become `notifA`/`notifB`.
- Role grants: `GRANT USAGE ON SCHEMA metaldocs TO <role>` + `GRANT SELECT,UPDATE,INSERT,DELETE ON
  metaldocs.notifications TO <role>` (was `public.documents`).
- All DML in the subtests: `public.documents` → `metaldocs.notifications`; the mutating column becomes
  `status` (`SET status='SENT'` / `'READ'`), retenant stays `SET tenant_id=$B`. Keep `WHERE id=$1::uuid`.
- Update the doc comment: table rationale (notifications = non-tripwired FORCE-RLS tenant table + real F3.2
  async seed site) and note WHY not `documents` (M2 capability tripwire would mask RLS).

### T2 — Run it GREEN for real (binding gate PG-2)
- Use the `.env`-loading PowerShell wrapper (sets `METALDOCS_DATABASE_URL`, redacts the password) →
  `go test -tags integration -count=1 -v -run 'SeedTxTenant_RLSBackstop' ./internal/modules/iam/authz/...`.
- Postgres already up: container `metaldocs-postgres`, host port from `.env` `POSTGRES_HOST_PORT`.
- Expect: leak_before=1 row; blocked_after 0/0; retenant=42501. If leak-before does NOT reproduce, STOP
  (investigate — do not weaken assertions).

### T3 — Drop the defer + record real evidence
- F3.5 `evidence.md`: real redacted run output, PG-1..4, root cause.
- `f3.2-async-rls-backstop/evidence.md`: change "negative RLS proof authored + run-deferred" → "authored,
  retargeted in F3.5, RUN GREEN for real (see f3.5)"; remove the defer.
- `validation-contract.md` §2.5: remove/annotate the "live run deferred" note → real green (F3.5), dated.

### T4 — Confirm no production drift
- `git diff --stat`: only the test file + M3 docs. `go build ./...` green.

## Files
- `internal/modules/iam/authz/seed_tx_tenant_rls_integration_test.go` (test only)
- `docs/.../f3.5-rls-proof-real-green/{spec,plan,evidence}.md`
- `docs/.../f3.2-async-rls-backstop/evidence.md` (defer label removed)
- `docs/.../validation-contract.md` (§2.5 defer note → real green)
- `docs/.../README.md` (status)

## Risk
- Notifications RLS policy identical shape to documents' `tenant_isolation`? Verify the policy exists on
  notifications (it's in the 33 FORCE set) — the USING doubling as WITH CHECK gives the 42501 on retenant.
- `recipient_user_id` FK: factory auto-creates a tenant user; fine. Two tenants' rows independent.
