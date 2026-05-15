# metaldocs.document_process_areas

> **Source:** `db/baseline/0001_current_schema.sql`
> **Schema:** `metaldocs`
> **Owner:** taxonomy

## Purpose
Current curated-baseline table owned by `taxonomy`. See the owning module wiki and runtime repositories for business behavior.

## Columns

| Column | Type | Nullable | Meaning |
|---|---|---|---|
| `code` | `text` | no | Baseline column. |
| `name` | `text` | no | Baseline column. |
| `description` | `text` | no | Baseline column. |
| `is_active` | `boolean` | no | Baseline column. |
| `created_at` | `timestamp with time zone` | no | Baseline column. |
| `tenant_id` | `uuid` | no | Baseline column. |
| `parent_code` | `text` | yes | Baseline column. |
| `owner_user_id` | `text` | yes | Baseline column. |
| `default_approver_role` | `text` | yes | Baseline column. |
| `archived_at` | `timestamp with time zone` | yes | Baseline column. |

## Baseline Definition

```sql
CREATE TABLE metaldocs.document_process_areas (
code text NOT NULL,
    name text NOT NULL,
    description text DEFAULT ''::text NOT NULL,
    is_active boolean DEFAULT true NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    tenant_id uuid DEFAULT 'ffffffff-ffff-ffff-ffff-ffffffffffff'::uuid NOT NULL,
    parent_code text,
    owner_user_id text,
    default_approver_role text,
    archived_at timestamp with time zone,
    CONSTRAINT area_code_format CHECK ((code ~ '^[a-z][a-z0-9_-]{1,63}$'::text))
);
```

## Runtime Usage

Use `rg -n "document_process_areas" internal apps` and the owning module wiki to verify readers/writers before changing this table.

## Seed or Reference Data

Check `db/reference-data/0001_product_reference_data.sql` and `db/dev-seeds/0001_local_dev_seed.sql` before adding rows.

## Notes and Debt

Curated baseline table in the `metaldocs` schema.
