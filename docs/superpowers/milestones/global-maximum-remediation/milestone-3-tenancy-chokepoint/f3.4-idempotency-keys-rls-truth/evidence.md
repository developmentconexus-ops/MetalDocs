# F3.4 evidence — idempotency_keys RLS-truth reconciliation (HS-4 fix)

> Closes milestone-validator finding **F-1** (`../qa/milestone-qa.md`). Operator approved the full fix incl.
> the HS-7 contract re-open (2026-07-03). **No behavior/policy change** — durable-record truth only.

## Root cause (F-1)
Three durable records asserted `idempotency_keys` is a "system table with no `tenant_id` column / RLS
structurally N/A." **Source truth:** `db/baseline/0001_current_schema.sql:1330` — `tenant_id uuid NOT NULL`;
`:1347` FORCE ROW LEVEL SECURITY; `:4621` `tenant_isolation` policy. It is **1 of the 33 FORCE tables**. The
idempotency-janitor's `DELETE FROM metaldocs.idempotency_keys WHERE expires_at < now()`
(`internal/modules/jobs/idempotency_janitor/job.go:34`) is a **sanctioned cross-tenant system-maintenance
sweep** run GUC-unset under the NULL-permissive hatch — same class as the audit-integrity scan, **not** a
table where RLS "cannot apply." `job_leases` genuinely has no `tenant_id` (that claim was correct).

## What changed (docs + lint-data only)

| File | Change |
|---|---|
| `validation-contract.md` (re-opened, HS-7) | Dated erratum note at head; §0.3 async-fleet row, §2.4 allowlist bullet (split job_leases / idempotency_keys), §4 janitor row — all corrected in place |
| `wiki/decisions/0027-rls-adoption-sequencing.md` (ADR 0027 amendment) | §4 residual-surface list split; per-binary janitor row corrected; amendment header dated F3.4 erratum |
| `scripts/api-lint/async-tenant-tables.txt` | NOTE comment rewritten (32→33; honest classification); **added `idempotency_keys`** so the set mirrors the 33 FORCE tables exactly |

**Reclassification (honest):** `idempotency_keys` = `tenant_id`-bearing FORCE-RLS table; janitor TTL DELETE
= sanctioned cross-tenant NULL-permissive maintenance sweep (audit-scan class); janitor package is outside
the `ASYNC-TENANT-SEED` scanned handler roots → no lint allowlist entry required, no false trip.

## Gates (captured)

| Gate | Command | Result |
|---|---|---|
| PG-1 — no residual false claim | grep `idempotency_keys` across the 3 records for "no tenant_id / N/A / cannot apply" | 0 (the only "no tenant_id" occurrences now correctly attribute to `job_leases`) |
| PG-2 — list == 33 FORCE tables | `diff <(FORCE-tables normalized, sort -u) <(non-comment entries, sort -u)` | **IDENTICAL** (33 == 33) |
| PG-3 — lint green | `go run ./scripts/api-lint -strict api/openapi/v1/openapi.yaml .` | **0 violation(s)**, exit 0 |
| PG-3 — lint unit tests | `go test ./scripts/api-lint/ -run 'SeedChokepoint\|AsyncTenantSeed'` | `ok` (no hardcoded-32 assumption broken) |
| PG-4 — no behavior diff | `git diff --stat` | only contract + ADR 0027 + async-tenant-tables.txt + f3.4 folder; **no `.go`/`.sql`/migration** |

## Scope discipline
- **No behavior/policy change:** RLS byte-identical (no SQL/migration/policy diff); no seed added/removed;
  the idempotency-janitor keeps its GUC-unset cross-tenant sweep (correct by design).
- **HS-7 honored:** the committed contract was re-opened **with operator approval**, corrected in place, and
  the re-open recorded as a dated erratum — not a silent edit, and not an acceptance-bar change (the false
  *premise* was fixed; the §3.3/PG-2 bar is unchanged).
- **Verified adding `idempotency_keys` stays green:** its only two write sites (janitor + sync idempotency
  store) are both outside the `ASYNC-TENANT-SEED` scanned roots (`internal/modules/jobs/idempotency_janitor`,
  `internal/platform/idempotency` — neither in `asyncHandlerRoots`).

## Contract conformance (spec §PG-1..4)
False claim corrected in all 3 records ✓ · idempotency_keys reclassified, job_leases intact ✓ · list == 33
FORCE mirror ✓ · api-lint 0 + unit tests green ✓ · no code/SQL/migration diff, RLS byte-identical ✓.
