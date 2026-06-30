# Post-baseline folded migrations (squash 2026-06-30)

These 54 files were the live `db/migrations/` tail (versions 0203..0256, inclusive of
`0224` — which is a file here but self-registered no `schema_migrations` row, so the
seeded ledger has 53 rows, not 54) that ran **after** the curated baseline. On 2026-06-30 their cumulative schema was folded into
`db/baseline/0001_current_schema.sql` (regenerated from a reference DB that applied the
old baseline + every one of these migrations in order), and their ledger rows were
seeded into `db/reference-data/0001_product_reference_data.sql` so a fresh bootstrap
ledger-skips them. They are kept here only as historical record.

## NOT part of the legacy-replay chain

`archive/migrations/` (root) holds the **pre-baseline** lineage 0001..0211, replayed as
one contiguous name-sorted chain by `scripts/dev-migrate.ps1` (non-recursive) for legacy
recovery. This fold range overlaps that lineage's numeric prefixes (0203..0211) with
*different* files, so it is isolated in this subfolder. `dev-migrate.ps1` is non-recursive
and therefore never replays these; `scripts/check-release-v2-names.ps1 -Recurse` still
inventories them for the naming-convention check (harmless).

Do **not** move these back into `db/migrations/` or into the archive root. A fresh
bootstrap builds from the baseline alone; re-introducing them would double-apply schema.

## Equivalence proof

The fold was gated by `scripts/check-baseline-equivalence.ps1`, which proved the squashed
baseline (`prerequisites → baseline → reference-data`) is runtime-equivalent to the full
migration chain across columns, constraints, indexes, triggers, functions, and extensions,
and that the `schema_migrations` ledgers are identical (54 versions: the
`baseline-2026-05-14` marker + 53 folded versions; `0224` never self-registered upstream
and is intentionally absent from both ledgers).
