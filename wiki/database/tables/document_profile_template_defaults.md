# metaldocs.document_profile_template_defaults

> **Source:** `db/baseline/0001_current_schema.sql`
> **Schema:** `metaldocs`
> **Owner:** taxonomy

## Purpose
Current curated-baseline table owned by `taxonomy`. See the owning module wiki and runtime repositories for business behavior.

## Columns

| Column | Type | Nullable | Meaning |
|---|---|---|---|
| `profile_code` | `text` | no | Baseline column. |
| `template_key` | `text` | no | Baseline column. |
| `template_version` | `integer` | no | Baseline column. |
| `assigned_at` | `timestamp with time zone` | no | Baseline column. |

## Baseline Definition

```sql
CREATE TABLE metaldocs.document_profile_template_defaults (
profile_code text NOT NULL,
    template_key text NOT NULL,
    template_version integer NOT NULL,
    assigned_at timestamp with time zone DEFAULT now() NOT NULL
);
```

## Runtime Usage

Use `rg -n "document_profile_template_defaults" internal apps` and the owning module wiki to verify readers/writers before changing this table.

## Seed or Reference Data

Check `db/reference-data/0001_product_reference_data.sql` and `db/dev-seeds/0001_local_dev_seed.sql` before adding rows.

## Notes and Debt

Curated baseline table in the `metaldocs` schema.
