# F3.4 — idempotency_keys RLS-truth reconciliation (HS-4 fix feature)

> **Milestone:** M3 · **Opened by:** milestone-validator FAIL verdict (`../qa/milestone-qa.md`, finding F-1).
> **Approval:** operator approved the full fix incl. the HS-7 contract re-open (2026-07-03). **No behavior/
> policy change** — pure runtime-truth reconciliation across durable records.

## Consumer contract (who consumes what)

Consumer: a future maintainer/reviewer + the milestone-validator reading the M3 enforcement record. The
record must state the **true** RLS surface. Finding F-1: the durable docs asserted `idempotency_keys` is a
"system table with no `tenant_id` column / RLS structurally N/A." Source
(`db/baseline/0001_current_schema.sql:1330,1347`) shows it **has** `tenant_id uuid NOT NULL` + FORCE ROW
LEVEL SECURITY + the `tenant_isolation` policy — it is 1 of the 33 FORCE tables. Its idempotency-janitor TTL
`DELETE … WHERE expires_at < now()` (`internal/modules/jobs/idempotency_janitor/job.go:34`) is a sanctioned
cross-tenant system-maintenance sweep run GUC-unset under the NULL-permissive hatch — same class as the
audit-integrity scan. `job_leases` genuinely has no `tenant_id` (that claim was correct).

**Required end-state:**
1. The false "`idempotency_keys` has no `tenant_id`" claim corrected in **all three** durable places:
   `validation-contract.md` §0.3 + §2.4 + §4 (HS-7 re-open, operator-approved, corrected in place + dated
   erratum at the contract head); ADR `0027` amendment (§4 residual list + per-binary table + amendment
   erratum note); `scripts/api-lint/async-tenant-tables.txt` comment.
2. `idempotency_keys` reclassified honestly as a `tenant_id`-bearing FORCE-RLS table whose janitor sweep is
   a sanctioned cross-tenant NULL-permissive maintenance `DELETE`; `job_leases` unchanged (correctly no-RLS).
3. `async-tenant-tables.txt` decision recorded + applied: **add** `idempotency_keys` so the list is the
   honest 33-table FORCE mirror. The idempotency-janitor is outside the `ASYNC-TENANT-SEED` scanned handler
   roots → no lint allowlist entry needed and no false trip.

## Non-goals (mandatory)
- **No behavior/policy change** — RLS stays byte-identical (no `.sql`/migration/policy edit); no seeding
  added or removed; the janitor keeps its GUC-unset cross-tenant sweep (correct by design).
- **No acceptance-bar change** — the §3.3/PG-2 "docs match runtime truth" bar is unchanged; this fixes the
  docs to meet it. The HS-7 re-open corrects a false *premise*, it does not weaken a criterion.
- **No new lint scope** — do not pull the idempotency-janitor into the async handler roots.

## Validation gate
- **PG-1:** no durable record (contract §0.3/§2.4/§4, ADR 0027 amendment, async-tenant-tables.txt) states or
  implies `idempotency_keys` has no `tenant_id` / RLS N/A. `job_leases` claim intact.
- **PG-2:** `async-tenant-tables.txt` (non-comment entries) == the 33 FORCE-RLS tables in
  `0001_current_schema.sql`, exactly (set-equal).
- **PG-3:** `go run ./scripts/api-lint -strict api/openapi/v1/openapi.yaml .` → **0 violations**; lint unit
  tests green (no hardcoded-32 assumption broken).
- **PG-4:** no `.go`/`.sql`/migration behavior file changed (only the `.txt` data/comment + the two docs +
  the contract). RLS policy byte-identical.

## Named proof commands
- `diff <(grep FORCE 0001_current_schema.sql | table-normalize | sort -u) <(grep -v '^#' async-tenant-tables.txt | sort -u)` → identical (33=33).
- `go run ./scripts/api-lint -strict api/openapi/v1/openapi.yaml .` → 0.
- `git diff --stat` → only the contract, ADR 0027, async-tenant-tables.txt, and the f3.4 folder.

## Interview record

| Q | A |
|---|---|
| Edit the committed contract in place or erratum-only? | Operator approved editing §0.3/§2.4/§4 in place under an HS-7 re-open + a dated erratum note at the contract head (auditable). |
| Add `idempotency_keys` to `async-tenant-tables.txt`? | Yes — make the list the honest 33-table FORCE mirror. Janitor is outside scanned roots → stays green, no allowlist entry. |
| Any behavior/policy change? | None. Truth reconciliation only; RLS byte-identical; janitor sweep unchanged. |
| Is `job_leases` also wrong? | No — genuinely no `tenant_id`, no RLS. Left as-is. |
