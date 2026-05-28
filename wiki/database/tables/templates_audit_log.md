# public.templates_audit_log

> **Source:** `db/baseline/0001_current_schema.sql` + `db/migrations/0213_templates_tenant_id_uuid.sql`
> **Schema:** `public`
> **Owner:** templates

## Purpose
Current curated-baseline table owned by `templates`. See the owning module wiki and runtime repositories for business behavior.

## Columns

| Column | Type | Nullable | Meaning |
|---|---|---|---|
| `id` | `bigint` | no | Baseline column. |
| `tenant_id` | `uuid` | no | Tenant scope for template audit history; converted from `text` by forward migration `0213_templates_tenant_id_uuid.sql`. |
| `template_id` | `uuid` | no | Baseline column. |
| `version_id` | `uuid` | yes | Baseline column. |
| `actor_id` | `text` | no | Baseline column. |
| `action` | `text` | no | Baseline column. |
| `details` | `jsonb` | no | Baseline column. |
| `occurred_at` | `timestamp with time zone` | no | Baseline column. |

## Baseline Definition

```sql
CREATE TABLE public.templates_audit_log (
id bigint NOT NULL,
    tenant_id uuid NOT NULL,
    template_id uuid NOT NULL,
    version_id uuid,
    actor_id text NOT NULL,
    action text NOT NULL,
    details jsonb DEFAULT '{}'::jsonb NOT NULL,
    occurred_at timestamp with time zone DEFAULT now() NOT NULL
);
```

## Runtime Usage

Use `rg -n "templates_audit_log" internal apps` and the owning module wiki to verify readers/writers before changing this table.

## Seed or Reference Data

Check `db/reference-data/0001_product_reference_data.sql` and `db/dev-seeds/0001_local_dev_seed.sql` before adding rows.

## Notes and Debt

Retained in `public` because current runtime/baseline truth still uses it. Do not move schemas without an approved migration plan.

Forward migration `0213_templates_tenant_id_uuid.sql` converts `tenant_id` from legacy `text` to `uuid` so audit reads stay type-aligned with template ownership.
