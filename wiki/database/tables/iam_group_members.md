# metaldocs.iam_group_members

> **Source:** `db/baseline/0001_current_schema.sql`
> **Schema:** `metaldocs`
> **Owner:** iam

## Purpose
Current curated-baseline table owned by `iam`. See the owning module wiki and runtime repositories for business behavior.

## Columns

| Column | Type | Nullable | Meaning |
|---|---|---|---|
| `group_id` | `uuid` | no | Baseline column. |
| `user_id` | `text` | no | Baseline column. |
| `tenant_id` | `uuid` | no | Baseline column. |
| `granted_at` | `timestamp with time zone` | no | Baseline column. |
| `granted_by` | `text` | yes | Baseline column. |

## Baseline Definition

```sql
CREATE TABLE metaldocs.iam_group_members (
group_id uuid NOT NULL,
    user_id text NOT NULL,
    tenant_id uuid NOT NULL,
    granted_at timestamp with time zone DEFAULT now() NOT NULL,
    granted_by text
);
```

## Runtime Usage

Use `rg -n "iam_group_members" internal apps` and the owning module wiki to verify readers/writers before changing this table.

## Seed or Reference Data

Check `db/reference-data/0001_product_reference_data.sql` and `db/dev-seeds/0001_local_dev_seed.sql` before adding rows.

## Notes and Debt

Curated baseline table in the `metaldocs` schema.
