# Migration Archaeology

## Commands

```powershell
$files = Get-ChildItem migrations -Filter *.sql | Sort-Object Name
"total=$($files.Count)"
"ledger_insert_files=$(( $files | Select-String -Pattern 'INSERT INTO public\.schema_migrations|INSERT INTO schema_migrations' -List ).Count)"
"begin_files=$(( $files | Select-String -Pattern '^BEGIN;' -List ).Count)"
rg -n "CREATE TABLE|ALTER TABLE|DROP TABLE|CREATE FUNCTION|CREATE TRIGGER|CREATE EXTENSION|GRANT |REVOKE |INSERT INTO|DELETE FROM|schema_migrations|dev|seed|repair|legacy|destroy|drop" migrations -S
```

## Counts

- `total=181`
- `ledger_insert_files=34`
- `begin_files=77`
- Range present `0001..0201`; missing numbers: `0078-0100`, `0119`, `0178-0179`.

## Era Summary

- Foundation schema era (`0001`-`~0031`): core DDL and grants.
- Domain expansion + seed-heavy era (`~0032`-`~0065`): registry/taxonomy/template expansion and data seeds.
- Ledger/idempotency + repair/defer era (`0112`,`0113`,`0167+`): schema_migrations-centric operations.
- Governance/hardening era (`0180`-`0201`): capabilities, audit/session hardening, tripwire/trigger controls.

## Classification Rules

- Structural DDL: create/alter/drop tables/functions/triggers.
- Privilege/security: grant/revoke changes.
- Product/reference seed data: runtime-required data rows.
- Local dev seed data: local users/passwords/demo workflow conveniences.
- Repair/deferred/destructive historical ops: exclude from curated fresh baseline unless final invariant is required.
- Ledger marker/repair mechanics: keep policy intent, not historical replay mechanics.

## Representative File Classifications

- `migrations/0001_init_documents.sql`: structural foundation.
- `migrations/0005_grant_workflow_audit_privileges.sql`: privilege/security.
- `migrations/0029_seed_metal_nobre_document_registry.sql`: mixed seed/reference (requires split decision).
- `migrations/0064_clean_slate_old_documents.sql`: destructive cleanup historical.
- `migrations/0112_docx_v2_schema_migrations_ledger.sql`: ledger-era bootstrap mechanics.
- `migrations/0167_documents_bridge_and_state_columns.sql`: idempotent bridge + ledger.
- `migrations/0171_drop_finalized_at.sql`: deferred/no-op with ledger insert.
- `migrations/0188_tripwire_extend.sql`: governance hardening and guardrails.
- `migrations/0197_repair_recent_migration_ledger.sql`: repair + ledger correction mechanics.

## Open Questions

- Confirm intent of missing numeric ranges (reserved/archived/lost).
- Confirm canonical ordering policy for duplicate numeric prefixes (`0042`, `0070`, `0130`) in legacy replay.
