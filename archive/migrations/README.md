# Legacy migration archive (pre-curated-baseline)

These are the original, pre-baseline historical migrations (`0001_*` …). They are
**archived, not active**. They were moved here from the repo-root `migrations/`
directory when the database moved to the curated-baseline model.

## What is authoritative now

Fresh setup and runtime use the **curated baseline**, not this tree:

- `db/baseline/0001_current_schema.sql` — the schema source of truth
- `db/prerequisites/`, `db/reference-data/`, `db/dev-seeds/`
- `db/migrations/` — the forward migration tail applied after the baseline (0203+)

See `wiki/database/migration-policy.md` and the curated-baseline design spec.

## Why these are kept (not deleted)

Per the migration policy, historical migrations are not deleted without archive
approval — they are evidence of how the schema was built and are still read by a
few historical content-assertion tests and the legacy-replay dev script. They are
**not** applied during normal bootstrap or at runtime.

## Who still references this tree

- `scripts/dev-migrate.ps1` — LEGACY REPLAY MODE (recovery/debugging only)
- `scripts/check-release-v2-names.ps1` — v2-name inventory scan
- a few `tests/integration/**/migration_0NNN_test.go` content-assertion tests

Do not add new migrations here — add forward migrations to `db/migrations/`.
