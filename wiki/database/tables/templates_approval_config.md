# public.templates_approval_config

> **Source:** `db/baseline/0001_current_schema.sql`
> **Schema:** `public`
> **Owner:** templates

## Purpose
**DROPPED** by forward migration `db/migrations/0302_drop_templates_approval_config.sql` (ROADMAP unit 3.1a, [ADR 0082](../../decisions/0082-approval-kernel-extraction.md) phase c — template legacy-approval retirement). This table still appears in the curated baseline snapshot (`db/baseline/0001_current_schema.sql`) below, but that snapshot predates migration 0302; a database migrated forward past 0302 no longer has this table. See the owning module wiki (`wiki/modules/templates.md`) for the retirement history.

## Columns

| Column | Type | Nullable | Meaning |
|---|---|---|---|
| `template_id` | `uuid` | no | Baseline column. |
| `reviewer_role` | `text` | yes | Baseline column. |
| `approver_role` | `text` | no | Baseline column. |

## Baseline Definition

```sql
CREATE TABLE public.templates_approval_config (
template_id uuid NOT NULL,
    reviewer_role text,
    approver_role text NOT NULL
);
```

## Runtime Usage

Use `rg -n "templates_approval_config" internal apps` and the owning module wiki to verify readers/writers before changing this table.

## Seed or Reference Data

Check `db/reference-data/0001_product_reference_data.sql` and `db/dev-seeds/0001_local_dev_seed.sql` before adding rows.

## Notes and Debt

**Dropped 2026-07-13** by `db/migrations/0302_drop_templates_approval_config.sql` (ROADMAP unit 3.1a S5, ADR 0082), behind a pre-drop emptiness assert (migration fails loud if any row remains rather than silently destroying data). Rationale: the write path was already gone by this point — 3.1a S1 removed `CreateTemplate`'s role-seeding/config-upsert, and 3.1a S4 deleted the 4 legacy approval routes/handlers, the legacy `Service` methods, the domain role symbols, and the repository config methods (`GetApprovalConfig`/`UpsertApprovalConfig`/`UpsertApprovalConfigTx`), leaving zero readers/writers. The templates module's `TenantDataPort.EraseTenantData` DELETE against this table was removed in the same change-set that added the migration, so tenant erasure does not break. The table remains in the curated baseline snapshot (`db/baseline/0001_current_schema.sql`) as a pre-0302 artifact — do not resurrect it without reversing the ADR 0082 retirement decision.
