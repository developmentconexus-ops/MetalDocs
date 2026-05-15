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
- Post-baseline migration runners must not derive ordering from a mixed ledger high-water mark when `schema_migrations` contains non-numeric markers such as a baseline marker.
- Normal API startup must not perform legacy data backfill or historical repair unless that path is explicitly enabled as governed recovery/maintenance.
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
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/check-system-runnable.ps1 -StartApi -TargetRoute /api/v1/controlled-documents
```

Use broader DB completion gates when the task changes runtime policy, migration runners, or baseline governance:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/check-db-bootstrap.ps1 -WithDevSeed
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/check-db-bootstrap.ps1
go test ./internal/platform/migrate/... ./internal/modules/registry/... -count=1
```

## Definition of Done

Do not consider the database foundation complete until all of the following are true:

- fresh curated bootstrap with dev seed passes
- product-schema bootstrap without dev seed passes
- post-baseline forward migrations can still apply after the baseline ledger marker
- normal API startup does not depend on historical replay or unconditional legacy repair
- dictionary coverage passes and table pages contain real schema-qualified content
- runtime auth/session and target route gates pass against the supported `/api/v1` contract
