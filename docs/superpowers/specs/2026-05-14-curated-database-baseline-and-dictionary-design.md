# Curated Database Baseline and Dictionary Design

## Status

Proposed for review.

## Problem

MetalDocs has accumulated 181 SQL migration files across several product eras. The current local database lifecycle is fragile because historical migrations, schema repairs, grants, extension setup, production reference data, local dev seeds, Docker entrypoint bootstrap, script replay, API startup migrations, and module startup backfills all participate in database creation or mutation.

The previous baseline effort correctly identified the need for a fresh-install baseline, but a raw `pg_dump` baseline is not enough. It would preserve the current database shape, including historical accidents, unused objects, add/drop churn, mixed seed data, and unclear schema ownership. MetalDocs needs a curated current-state baseline that represents how the SaaS database should work now, while preserving legacy migration evidence until upgrade and archive decisions are explicit.

## Goals

- Produce a clean current-state database install path for fresh local and ephemeral environments.
- Study the existing migration history and runtime code before deciding what belongs in the baseline.
- Keep only runtime-used and product-justified database objects in the curated baseline.
- Separate product schema, product reference data, local dev seeds, test/demo data, grants, extensions, and repair-only history.
- Create a database dictionary in the wiki so every surviving table has documented purpose, ownership, relationships, lifecycle rules, and runtime usage.
- Define concise workflow updates for `AGENTS.md`, `CLAUDE.md`, and MetalDocs skills so future DB work follows the same reality-first standards as backend/API, frontend, TanStack Query, and module wiki work.
- Preserve existing migration history as evidence and upgrade/debug material until an explicit archive decision is made.

## Non-Goals

- Blindly compressing all 181 migrations into one raw baseline file.
- Continuing to patch historical migrations just to make local startup pass.
- Rewriting production/shared upgrade history without a reviewed migration and rollback strategy.
- Moving every legacy `public` object into `metaldocs` immediately.
- Redesigning product data models unrelated to current runtime truth.
- Creating fake runtime data or mocked backend behavior to satisfy verification.

## Classification

This work is both:

- `runtime prerequisite`: local startup and migration truth are not trustworthy enough for feature work.
- `workflow/tooling gap`: database evolution lacks a canonical workflow, skill gate, dictionary, and verification standard.

It is not module-local feature work.

## Current Reality

Database changes currently happen through multiple paths:

1. Docker Postgres entrypoint mounts `migrations/` into `/docker-entrypoint-initdb.d`, so fresh volumes auto-run historical SQL.
2. `scripts/dev-migrate.ps1` can replay the same migration directory and now contains guard logic to skip if the entrypoint already initialized the DB.
3. API startup calls `internal/platform/migrate.Apply` unless `METALDOCS_SKIP_STARTUP_MIGRATIONS=true`.
4. Registry startup runs a module-level backfill through `RunStartupMigrations`.

The migration ledger is partial. The repository has 181 SQL files, but only a small subset records `public.schema_migrations` rows. This means the ledger cannot currently be treated as a full historical truth table.

The schema is mixed between `metaldocs` and `public`, and many migration files use unqualified names. Some migrations contain product schema, some contain grants, some contain dev users, some contain repair rows, and some contain destructive cleanup from previous eras.

## Design Decision

MetalDocs will build a curated current-state database baseline rather than using a raw schema dump as the final source of truth.

The curated baseline is written from evidence:

- live schema after trusted legacy replay
- runtime repository/query usage
- migration archaeology
- seed/reference data classification
- module ownership and wiki truth

Every database object included in the curated baseline must have a documented reason to exist. If a table, column, trigger, function, index, role grant, or seed cannot be tied to runtime usage, product-required behavior, or an explicit deferred debt note, it does not silently enter the clean baseline.

## Target Repository Model

The exact paths may be adjusted during planning, but the target shape should be:

```text
db/
  baseline/
    0001_current_schema.sql
  reference-data/
    0001_product_reference_data.sql
  dev-seeds/
    0001_local_dev_users.sql
  migrations/
    0002_forward_change_after_baseline.sql
  prerequisites/
    0001_extensions.sql

migrations/
  legacy historical chain, preserved until archive approval

wiki/database/
  overview.md
  schemas.md
  relationships.md
  reference-data.md
  migration-policy.md
  dictionary-index.md
  tables/
    <table>.md
```

`migrations/` may remain in place during the first rollout, but it must stop being the hidden fresh-install default.

## Database Dictionary

The database dictionary is a required deliverable of the baseline refactor.

Each surviving table should have a wiki entry with:

- table name and schema
- owning module or platform area
- product purpose
- row lifecycle
- column meanings
- primary keys, foreign keys, unique constraints, and checks
- important indexes and why they exist
- triggers/functions involved
- runtime readers and writers
- seed/reference data expectations
- tenant/audit/security implications
- known debt or migration notes

The dictionary should be hybrid-generated and human-reviewed. Scripts may extract mechanical facts from Postgres, but ownership, purpose, lifecycle, and product meaning must be reviewed by humans/Codex.

Baseline rule: a table that survives into the curated baseline must have a dictionary entry or an explicit documented exception.

## Bootstrap Model

Fresh local and ephemeral environments should use the curated baseline path:

1. Start empty Postgres.
2. Apply prerequisites, including required extensions such as `pgcrypto`.
3. Apply curated product schema baseline.
4. Apply required product reference data.
5. Apply local dev seeds only when explicitly requested by a local/dev script.
6. Apply post-baseline forward migrations.
7. Run runtime verification gates.

Existing databases should use the legacy upgrade path until a separate reviewed upgrade strategy replaces it.

Docker Postgres entrypoint must stop auto-running the historical `migrations/` directory. Postgres should provide infrastructure; MetalDocs scripts should own database bootstrap mode selection.

## Startup Migration Policy

API startup migrations may remain only as a controlled forward-migration mechanism after the curated baseline cutoff.

They should not be responsible for:

- bootstrapping a fresh database from historical migrations
- applying dev seeds
- repairing arbitrary historical drift
- masking startup prerequisites

Local development should prefer explicit scripts. API startup should either validate DB readiness or apply a narrow post-baseline tail, depending on the environment mode chosen in the implementation plan.

Module startup backfills should be treated as product migrations unless they are genuinely runtime-maintenance jobs. Any schema/data backfill that changes durable product state should move into a governed migration or documented job with a clear idempotency contract.

## Ledger Policy

`public.schema_migrations` should become a trustworthy ledger from the curated baseline cutoff forward.

Minimum target behavior:

- records a baseline marker
- records every post-baseline forward migration
- records applied timestamps and descriptions
- does not pretend every historical file was applied unless a legacy replay actually ran
- can distinguish curated baseline state from legacy-replay state

Historical ledger gaps should be documented, not papered over by ad hoc repair rows.

## Seeds and Reference Data

Data must be classified before it enters the new baseline path:

- product reference data: roles, capabilities, system records, and invariants required in every environment
- local dev seed data: users, demo tenants, and convenience records for local workflows
- test data: fixtures used by automated tests
- historical repair data: one-time corrections from previous eras

Production schema migrations must not create local-only users or demo workflows.

## Schema Ownership

New baseline files should schema-qualify all objects.

Preferred default:

- application-owned business/platform objects live in `metaldocs`
- Postgres extension objects may remain in their normal extension schema
- legacy `public` objects remain only when moving them would be a separate product migration risk

The dictionary must mark each legacy `public` object as one of:

- intentionally public for now
- candidate for future move
- historical/unused candidate for exclusion

## Research Phase

Before implementation edits, perform a read-only study. This phase is parallel-safe because each agent produces evidence, not code changes.

Recommended subagent split:

1. Runtime Usage Agent: inventory Go repositories, SQL strings, functions, tables, columns, and module ownership.
2. Migration Archaeology Agent: classify all 181 SQL files as current schema, superseded, repair-only, destructive, grant/security, product reference data, dev seed, or unknown.
3. Live Schema Agent: inspect current DB objects after trusted legacy replay: tables, columns, constraints, indexes, triggers, functions, extensions, grants.
4. Seed/Auth Agent: classify auth, IAM, tenant, capability, system template, and approval seed data.
5. Governance Agent: propose concise updates for `AGENTS.md`, `CLAUDE.md`, skills, runbooks, and verification gates.

No research agent may edit files during the study.

## Verification Gates

The curated baseline is not accepted until these pass:

- legacy replay can still be run intentionally for evidence/debugging
- fresh baseline bootstrap starts from empty Postgres with no Docker entrypoint migration auto-run
- baseline plus product reference data plus selected dev seeds starts API successfully
- auth/session runtime gate passes
- representative module routes pass
- baseline schema has dictionary coverage for every included table
- schema comparison detects unexpected drift between curated baseline and runtime-used objects from trusted reference DB
- dev seed path can be skipped without breaking product schema initialization

## Documentation and Skill Updates

The implementation plan should include concise updates to:

- `AGENTS.md`: add DB work to the core skill map and mismatch classification.
- `CLAUDE.md`: mirror the same short database workflow pointer.
- `.agents/skills/`: add or bridge a MetalDocs database migration/bootstrap skill.
- `.claude/skills/`: canonical source for the database workflow skill.
- `wiki/database/`: database dictionary and migration policy.
- `docs/runbooks/`: operator commands for fresh baseline, legacy replay, seed handling, and verification.

These updates should avoid duplicating detailed rules across files. `AGENTS.md` and `CLAUDE.md` should point to the canonical skill and wiki documents.

## Risks and Mitigations

Risk: excluding an object that runtime still needs.

Mitigation: runtime usage inventory, live schema comparison, and startup/auth/module route gates.

Risk: curated baseline becomes a hand-written fantasy instead of runtime truth.

Mitigation: derive it from trusted replay, code usage, and dictionary review.

Risk: legacy upgrade path is lost too early.

Mitigation: preserve historical migrations until an explicit archive/upgrade decision.

Risk: documentation drifts immediately.

Mitigation: make dictionary coverage part of the baseline acceptance gate and future DB workflow skill.

## Success Criteria

This effort succeeds when:

- fresh local setup uses a clean curated baseline rather than replaying historical migration archaeology
- historical migrations are preserved as legacy evidence, not hidden bootstrap machinery
- product schema, reference data, dev seeds, and tests are separated
- `schema_migrations` is trustworthy from the baseline cutoff forward
- every surviving table has a database dictionary entry or explicit exception
- startup/runtime gates pass from an empty database through the official bootstrap path
- future DB changes have a concise skill/runbook workflow matching MetalDocs' existing architecture discipline
