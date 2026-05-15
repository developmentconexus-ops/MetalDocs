# metaldocs.iam_user_roles

> **Source:** `db/baseline/0001_current_schema.sql`
> **Schema:** `metaldocs`
> **Owner:** iam

## Purpose
Current curated-baseline table owned by `iam`. See the owning module wiki and runtime repositories for business behavior.

## Columns

| Column | Type | Nullable | Meaning |
|---|---|---|---|
| `user_id` | `text` | no | Baseline column. |
| `role_code` | `text` | no | Baseline column. |
| `assigned_at` | `timestamp with time zone` | no | Baseline column. |
| `assigned_by` | `text` | yes | Baseline column. |
| `tenant_id` | `uuid` | no | Baseline column. |

## Baseline Definition

```sql
CREATE TABLE metaldocs.iam_user_roles (
user_id text NOT NULL,
    role_code text NOT NULL,
    assigned_at timestamp with time zone DEFAULT now() NOT NULL,
    assigned_by text,
    tenant_id uuid DEFAULT 'ffffffff-ffff-ffff-ffff-ffffffffffff'::uuid NOT NULL,
    CONSTRAINT chk_iam_user_roles_role_code CHECK ((role_code = ANY (ARRAY['system_admin'::text, 'approver'::text, 'author'::text, 'editor'::text, 'viewer'::text])))
);
```

## Runtime Usage

Use `rg -n "iam_user_roles" internal apps` and the owning module wiki to verify readers/writers before changing this table.

## Seed or Reference Data

Check `db/reference-data/0001_product_reference_data.sql` and `db/dev-seeds/0001_local_dev_seed.sql` before adding rows.

## Notes and Debt

Curated baseline table in the `metaldocs` schema.
