# Reference Data

> **Last verified:** 2026-08-03
> **Scope:** Product reference data (bootstrap stage 3) and its relationship to local-only dev seeds.
> **Out of scope:** Grants/privilege stage — `wiki/database/migration-policy.md` ("Grants stage").
> **Key files:**
> - `db/reference-data/0001_product_reference_data.sql` — stage 3
> - `db/dev-seeds/0001_local_dev_seed.sql` — dev-only, optional

Product reference data is required in every environment and must remain separate from local-only developer seed records.

Reference data currently lives in:

- `db/reference-data/0001_product_reference_data.sql`

Local developer seed data currently lives in:

- `db/dev-seeds/0001_local_dev_seed.sql`

As of the 2026-07-29 baseline fold, this file also carries the seeded `public.schema_migrations` ledger rows for the folded migration range (`0257`–`0315`, minus the never-existing `0261`/`0280`/`0289`/`0291`), so a fresh bootstrap has a full ledger without replaying those files. See `wiki/database/migration-policy.md` for the full stage list and fold procedure.
