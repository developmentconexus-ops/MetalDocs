# Post-baseline folded migrations (squash 2026-07-29)

These 55 files were the entire live `db/migrations/` tail (versions 0257..0315 —
`0261`, `0280`, `0289` and `0291` never existed) that ran **after** the curated baseline.
On 2026-07-29 their cumulative schema was folded into
`db/baseline/0001_current_schema.sql` (regenerated wholesale from a reference DB that
applied prerequisites → the previous baseline → reference-data → every one of these
migrations in lexical order), and their ledger rows were seeded into
`db/reference-data/0001_product_reference_data.sql` so a fresh bootstrap ledger-skips
them. They are kept here only as historical record.

`0309_pdf_dispatch_outbox_final_docx_key.sql` is a file here but self-registered **no**
`schema_migrations` row upstream, so the seeded ledger has **54 rows, not 55** — the same
class of upstream omission as `0224` in the 2026-06 fold. (Side effect worth knowing: for
as long as 0309 lived in `db/migrations/`, every `metaldocs-api` start re-applied it,
because `internal/platform/migrate.Apply` skips by ledger row and 0309 never wrote one.
The fold ends that re-apply loop.)

Four of these files — `0260`, `0306`, `0307`, `0308` — had already been folded into the
baseline by the 2026-07-16 squash (ROADMAP unit 4.5) but were deliberately left in
`db/migrations/` as verified no-op replays. This fold archives them with the rest.

`db/migrations/` is now empty except for its `README.md`, which is intentional and
tolerated end-to-end: `internal/platform/migrate.Apply` does `os.ReadDir` on the
directory (errors only if it is **missing**, not if it is empty) and
`tests/integration/testdb.curatedBundlePaths` appends an empty file list.

## Privileges and roles were NOT folded into the baseline

The baseline is regenerated with `pg_dump --schema-only --no-owner --no-privileges`,
which carries no ACLs, and role creation is cluster-global so `pg_dump` never emits it.
Three files in this range carried privilege/role effects that a schema-only fold would
have silently dropped:

| file | effect |
|---|---|
| `0266_audit_events_hardening.sql` (part a) | `REVOKE UPDATE, DELETE, TRUNCATE` on `metaldocs.audit_events` from the app role |
| `0284_ci_rls_role.sql` | create non-owner `NOSUPERUSER`+`NOBYPASSRLS` `metaldocs_ci` role + DML grants + default privileges |
| `0314_outbox_events_retention_grant.sql` | `GRANT DELETE` on `metaldocs.outbox_events` to the app role |

They were re-homed **verbatim, in guarded/idempotent form** into the new bootstrap stage
`db/grants/0001_role_grants.sql`, applied after `db/reference-data/`. This mirrors how the
2026-06 fold re-homed extension ownership into `db/prerequisites/0001_extensions.sql`.
Dropping them would have broken `tests/integration/testdb/ci_role.go` (RLS-truth tests
would false-green under the owner role) and re-opened the outbox-retention purge and
audit-mutation holes on any deployment where `metaldocs_app` is not the table owner.

## NOT part of the legacy-replay chain

`archive/migrations/` (root) holds the **pre-baseline** lineage 0001..0211, replayed as one
contiguous name-sorted chain by `scripts/dev-migrate.ps1` (non-recursive) for legacy
recovery. `dev-migrate.ps1` is non-recursive and therefore never replays these;
`scripts/check-release-v2-names.ps1 -Recurse` still inventories them for the
naming-convention check (harmless).

Do **not** move these back into `db/migrations/` or into the archive root. A fresh bootstrap
builds from the baseline alone; re-introducing them would double-apply schema.

## Equivalence proof

The fold was gated by `scripts/check-baseline-equivalence.ps1`, which proved the squashed
bundle (`prerequisites → baseline → reference-data → grants`) is runtime-equivalent to the
full migration chain across **columns, constraints, indexes, triggers, functions,
extensions, tables, RLS flags, policies, views and sequences** — zero diff in all eleven
object classes. The `schema_migrations` ledgers are identical (108 versions: the
`baseline-2026-05-14` marker + 53 rows from the 2026-06 fold + 54 rows from this one).
ACLs, which the gate deliberately does not compare, were verified separately (67 tables +
4 default-ACL entries + both schema ACLs match).
