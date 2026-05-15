---
name: metaldocs-database
description: Use for any MetalDocs database work touching migrations, bootstrap, curated baseline, reference data, dev seeds, schema ownership, database dictionary, schema_migrations, Postgres extensions, grants, triggers, functions, or runtime DB startup drift.
---

# MetalDocs Database Workflow

Use this skill for database migration/bootstrap/dictionary work.

## Required Reads

- `wiki/database/overview.md`
- `wiki/database/migration-policy.md`
- `wiki/database/dictionary-index.md`
- `docs/superpowers/specs/2026-05-14-curated-database-baseline-and-dictionary-design.md`

## Classification

Classify the task before editing:

- runtime prerequisite
- workflow/tooling gap
- schema baseline change
- product reference data change
- local dev seed change
- post-baseline forward migration
- dictionary/wiki update
- historical migration evidence/recovery

## Rules

- Do not patch historical migrations to hide bootstrap drift.
- Fresh local setup uses curated baseline artifacts, not Docker entrypoint migration replay.
- Product schema, product reference data, and local dev seeds stay separated.
- Local authenticated smoke tests use the optional dev seed account unless a first-boot bootstrap admin flow is intentionally configured.
- New post-baseline migrations must write `public.schema_migrations`.
- Required extensions are declared before use.
- Every baseline table needs a wiki dictionary page or explicit exception.
- Runtime truth wins over migration archaeology when deciding whether an object is still used.
- Stop and surface contradictions involving startup path, schema ownership, migration ledger, seed scope, or verification expectations.

## Verification

Use the smallest applicable gate:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/dev-bootstrap-baseline.ps1 -WithDevSeed
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/check-db-dictionary-coverage.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/start-api.ps1 -Build -NoWorker
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/check-system-runnable.ps1 -TargetRoute /api/v1/controlled-documents
```
