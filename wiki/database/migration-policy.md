# Migration Policy

> **Last verified:** 2026-08-03
> **Scope:** Fresh-bootstrap stage order, forward-migration rules, and the baseline regeneration procedure.
> **Out of scope:** Per-table schema truth (`wiki/database/dictionary-index.md`), reference-data content (`wiki/database/reference-data.md`).
> **Key files:**
> - `db/prerequisites/0001_extensions.sql` — stage 1
> - `db/baseline/0001_current_schema.sql` — stage 2
> - `db/reference-data/0001_product_reference_data.sql` — stage 3
> - `db/grants/0001_role_grants.sql` — stage 4
> - `db/migrations/` — forward tail (currently only `README.md`, empty)
> - `internal/platform/migrate/migrate.go:32` — `Apply` (forward tail, ledgered)
> - `internal/platform/migrate/migrate.go:115` — `ApplyGrants` (grants stage, unledgered, unconditional)
> - `scripts/export-schema-baseline.ps1` — baseline regeneration
> - `scripts/check-baseline-equivalence.ps1` — regeneration equivalence gate

## Baseline Model

MetalDocs fresh bootstrap is **four stages**, then an empty forward tail:

1. `db/prerequisites/0001_extensions.sql`
2. `db/baseline/0001_current_schema.sql`
3. `db/reference-data/0001_product_reference_data.sql`
4. `db/grants/0001_role_grants.sql` — privilege/role effects `pg_dump --no-privileges` cannot carry (see "Grants stage" below); optionally followed by `db/dev-seeds/0001_local_dev_seed.sql` (dev-only)
5. `db/migrations/*.sql` (post-baseline forward tail) — **currently empty** except `README.md`; migrations `0257`–`0315` were folded into the baseline on 2026-07-29 (`0261`/`0280`/`0289`/`0291` never existed) and archived to `archive/migrations/post-baseline-2026-07-fold/` with a `README.md` explaining the fold. `internal/platform/migrate.Apply` tolerates an empty `db/migrations/` (`os.ReadDir` only errors if the directory is missing).

Applied by, in order: compose `initdb.d` (numbered scripts), `scripts/dev-bootstrap-baseline.ps1`, and `tests/integration/testdb` curated bundle paths.

## Grants stage (`db/grants/`)

`db/grants/0001_role_grants.sql` re-homes privilege effects a schema-only `pg_dump --no-owner --no-privileges` cannot carry: `audit_events` REVOKEs (from the 2026-07 fold's `0266`), the `metaldocs_ci` CI role + grants (from `0284`), and the outbox-retention `GRANT` (from `0314`). It is:

- **idempotent** — safe to re-run
- applied at fresh bootstrap (same three call sites as the other stages)
- applied **unconditionally at every `metaldocs-api` startup** via `migrate.ApplyGrants` (`apps/api/cmd/metaldocs-api/main.go:242`), under the same migration advisory lock as `migrate.Apply`, immediately before it
- **not** ledgered — writes no `schema_migrations` rows
- fail-closed — missing or empty `METALDOCS_GRANTS_DIR` (default `db/grants`) aborts startup

Privilege changes are edited directly in this file. The startup path (not a forward migration) is how privilege changes reach already-provisioned volumes — there is no hand-sync with `db/migrations/`.

## Baseline regeneration procedure

Executed 2026-07-29 to fold migrations `0257`–`0315`:

1. Build a reference DB: prerequisites → previous baseline → reference-data → full replay of every folded migration in lexical order.
2. `pg_dump --schema-only --no-owner --no-privileges` that reference DB.
3. Curate the dump (UTF-8, `BEGIN`/`COMMIT` wrap, fold header) into `db/baseline/0001_current_schema.sql` — the canonical output path for `scripts/export-schema-baseline.ps1`.
4. Append the folded migrations' ledger rows to `db/reference-data/0001_product_reference_data.sql` (fresh bootstrap seeds a full ledger without replaying files).
5. Gate the candidate baseline with `scripts/check-baseline-equivalence.ps1` — zero-diff across all 11 object classes is mandatory before the regenerated baseline is accepted.
6. Archive the folded migration files to `archive/migrations/post-baseline-2026-07-fold/` with a `README.md`.

Ledger note: fresh bootstrap seeds 108 `schema_migrations` rows (sentinel + `0203`–`0256` minus `0224` + `0257`–`0315` minus the never-existing `0261`/`0280`/`0289`/`0291`). `0309` never self-registered a ledger row pre-fold either, so pre-fold volumes may show a benign one-row delta versus a fresh bootstrap — see the archive `README.md`.

Tripwire vocabulary golden lives at `internal/platform/tripwire/golden/0301_tripwire_template_review_retired.sql` (checked by the `TRIPWIRE-ARM-PARITY` api-lint and the `gen-tripwire` target). Future tripwire vocabulary changes ship as **new forward migrations**, not baseline edits.

`migrations_baseline/` (an orphaned stale copy) was deleted in the same fold.

## Post-bootstrap ops

Rotate the `metaldocs_ci` password on non-dev environments — see the checklist in `ops/DEPLOY.md`.

## Rules

- Do not patch historical migrations to hide bootstrap drift.
- Do not delete or move historical migrations without explicit archive approval.
- Do not use historical migration replay as the normal fresh bootstrap path.
- Post-baseline migrations must be forward-only and idempotent.
- Every post-baseline migration must write one `public.schema_migrations` row.
- Runtime and contract truth take precedence over migration archaeology when contradictions appear.
