# template_audit_log

This dictionary page covers same-name tables in multiple schemas. Keep schema qualification explicit in runtime SQL.

## metaldocs.template_audit_log

> **Source:** `db/baseline/0001_current_schema.sql`
> **Schema:** `metaldocs`
> **Owner:** templates

## Purpose
Current curated-baseline table owned by `templates`. See the owning module wiki and runtime repositories for business behavior.

## Columns

| Column | Type | Nullable | Meaning |
|---|---|---|---|
| `id` | `uuid` | no | Baseline column. |
| `template_key` | `text` | no | Baseline column. |
| `version` | `integer` | yes | Baseline column. |
| `action` | `text` | no | Baseline column. |
| `actor_id` | `text` | no | Baseline column. |
| `diff_summary` | `text` | yes | Baseline column. |
| `trace_id` | `text` | yes | Baseline column. |
| `created_at` | `timestamp with time zone` | no | Baseline column. |

## Baseline Definition

```sql
CREATE TABLE metaldocs.template_audit_log (
id uuid DEFAULT gen_random_uuid() NOT NULL,
    template_key text NOT NULL,
    version integer,
    action text NOT NULL,
    actor_id text NOT NULL,
    diff_summary text,
    trace_id text,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);
```

## Runtime Usage

Use `rg -n "template_audit_log" internal apps` and the owning module wiki to verify readers/writers before changing this table.

## Seed or Reference Data

Check `db/reference-data/0001_product_reference_data.sql` and `db/dev-seeds/0001_local_dev_seed.sql` before adding rows.

## Notes and Debt

Curated baseline table in the `metaldocs` schema.

## public.template_audit_log

> **Source:** `db/baseline/0001_current_schema.sql`
> **Schema:** `public`
> **Owner:** templates

## Purpose
Current curated-baseline table owned by `templates`. See the owning module wiki and runtime repositories for business behavior.

## Columns

| Column | Type | Nullable | Meaning |
|---|---|---|---|
| `id` | `uuid` | no | Baseline column. |
| `tenant_id` | `uuid` | no | Baseline column. |
| `template_id` | `uuid` | yes | Baseline column. |
| `template_version_id` | `uuid` | yes | Baseline column. |
| `document_id` | `uuid` | yes | Baseline column. |
| `action` | `text` | no | Baseline column. |
| `actor_user_id` | `text` | no | Baseline column. |
| `metadata_json` | `jsonb` | yes | Baseline column. |
| `created_at` | `timestamp with time zone` | no | Baseline column. |

## Baseline Definition

```sql
CREATE TABLE public.template_audit_log (
id uuid DEFAULT gen_random_uuid() NOT NULL,
    tenant_id uuid NOT NULL,
    template_id uuid,
    template_version_id uuid,
    document_id uuid,
    action text NOT NULL,
    actor_user_id text NOT NULL,
    metadata_json jsonb,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);
```

## Runtime Usage

Use `rg -n "template_audit_log" internal apps` and the owning module wiki to verify readers/writers before changing this table.

## Seed or Reference Data

Check `db/reference-data/0001_product_reference_data.sql` and `db/dev-seeds/0001_local_dev_seed.sql` before adding rows.

## Notes and Debt

Retained in `public` because current runtime/baseline truth still uses it. Do not move schemas without an approved migration plan.
