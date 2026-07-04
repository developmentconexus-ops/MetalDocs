# F3.4 plan — idempotency_keys RLS-truth reconciliation

> Contract: `../validation-contract.md` (re-opened §0.3/§2.4/§4 under HS-7) + `../qa/milestone-qa.md` F-1.
> Docs + lint-data only. No behavior/policy change.

## Task list (ordered)

### T1 — Committed contract (HS-7 re-open, operator-approved)
- Dated **erratum** note at the contract head recording the re-open + the corrected fact.
- §0.3 async-fleet row: `idempotency_keys` → "tenant_id-bearing FORCE-RLS table; cross-tenant TTL sweep
  under NULL-permissive, same class as audit scan"; `job_leases` → "genuinely no tenant_id column".
- §2.4 allowlist: split the old "system tables with no tenant_id" bullet into `job_leases` (no tenant_id,
  RLS N/A) and `idempotency_keys` (tenant_id-bearing FORCE-RLS; sanctioned cross-tenant NULL-permissive
  janitor sweep; janitor outside ASYNC-TENANT-SEED roots → no allowlist entry).
- §4 janitor row: same correction inline.

### T2 — ADR 0027 amendment (F3.3 durable record)
- §4 residual-surface list: split job_leases / idempotency_keys as above.
- Per-binary janitor row: correct the last cell.
- Amendment header: dated F3.4 erratum note.

### T3 — Lint data file
- `scripts/api-lint/async-tenant-tables.txt`: rewrite the NOTE comment (32→33; honest classification);
  **add** `idempotency_keys` to the entry list so it mirrors the 33 FORCE tables exactly.

### T4 — Gate + evidence
- PG-2 set-equality (33==33); PG-3 api-lint 0 + lint unit tests green; PG-4 no code/SQL/migration diff.

## Files expected to change
- `docs/.../milestone-3-tenancy-chokepoint/validation-contract.md`
- `wiki/decisions/0027-rls-adoption-sequencing.md`
- `scripts/api-lint/async-tenant-tables.txt`
- `f3.4-idempotency-keys-rls-truth/{spec,plan,evidence}.md`

## Risk / ordering notes
- Adding `idempotency_keys` to the table set could false-trip `ASYNC-TENANT-SEED` if any scanned async
  handler root wrote it unseeded. Verified: the only two write sites (`idempotency_janitor/job.go`,
  `platform/idempotency/postgres_store.go`) are **both outside** the scanned roots → stays green.
- Any `.go`/`.sql` behavior diff here is a scope breach — stop. This is truth reconciliation only.
