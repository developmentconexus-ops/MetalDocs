# metaldocs.role_capabilities

> **Source:** `db/baseline/0001_current_schema.sql`
> **Schema:** `metaldocs`
> **Owner:** iam

## Purpose
Current curated-baseline table owned by `iam`. See the owning module wiki and runtime repositories for business behavior.

## Columns

| Column | Type | Nullable | Meaning |
|---|---|---|---|
| `role` | `text` | no | Baseline column. |
| `capability` | `text` | no | Baseline column. |
| `description` | `text` | no | Baseline column. |
| `created_at` | `timestamp with time zone` | no | Baseline column. |

## Baseline Definition

```sql
CREATE TABLE metaldocs.role_capabilities (
role text NOT NULL,
    capability text NOT NULL,
    description text DEFAULT ''::text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT ck_cap_format CHECK ((capability ~ '^[a-z][a-z0-9._]*[a-z0-9]$'::text)),
    CONSTRAINT ck_cap_not_legacy CHECK ((capability <> ALL (ARRAY['document.finalize'::text, 'document.archive'::text])))
);
```

## Runtime Usage

Use `rg -n "role_capabilities" internal apps` and the owning module wiki to verify readers/writers before changing this table.

## Seed or Reference Data

Check `db/reference-data/0001_product_reference_data.sql` and `db/dev-seeds/0001_local_dev_seed.sql` before adding rows.

## Notes and Debt

Curated baseline table in the `metaldocs` schema.
