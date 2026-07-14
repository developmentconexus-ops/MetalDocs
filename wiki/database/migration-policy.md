# Migration Policy

## Baseline Model

MetalDocs fresh bootstrap uses:

1. `db/prerequisites/0001_extensions.sql`
2. `db/baseline/0001_current_schema.sql`
3. `db/reference-data/0001_product_reference_data.sql`
4. `db/dev-seeds/0001_local_dev_seed.sql` (optional)
5. `db/migrations/*.sql` (post-baseline forward tail)

## Rules

- Do not patch historical migrations to hide bootstrap drift.
- Do not delete or move historical migrations without explicit archive approval.
- Do not use historical migration replay as the normal fresh bootstrap path.
- Post-baseline migrations must be forward-only and idempotent.
- Every post-baseline migration must write one `public.schema_migrations` row.
- Runtime and contract truth take precedence over migration archaeology when contradictions appear.
