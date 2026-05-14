# Runbook: Migration Governance

# Runbook: Migration Governance

MetalDocs uses a curated current-state baseline for fresh environments and preserves historical migrations for legacy replay/debugging until archive approval.

Canonical policy lives in `wiki/database/migration-policy.md`.

## Rules

1. Do not use Docker Postgres entrypoint to auto-run historical migrations.
2. Do not patch historical migrations to hide bootstrap drift.
3. Keep product schema, product reference data, and local dev seeds separated.
4. Every post-baseline forward migration must record `public.schema_migrations`.
5. Every baseline table must have a database dictionary page or explicit exception.
6. Historical migration archive/deletion requires explicit review.
