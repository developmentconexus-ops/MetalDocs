# metaldocs.document_types

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
| `review_interval_days` | `integer` | no | Baseline column. |
| `is_active` | `boolean` | no | Baseline column. |
| `created_at` | `timestamp with time zone` | no | Baseline column. |
| `type_key` | `text` | yes | Baseline column. |
| `family_key` | `text` | no | Baseline column. |
| `active_version` | `integer` | no | Baseline column. |

## Baseline Definition

```sql
CREATE TABLE metaldocs.document_types (
code text NOT NULL,
    name text NOT NULL,
    description text DEFAULT ''::text NOT NULL,
    review_interval_days integer NOT NULL,
    is_active boolean DEFAULT true NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    type_key text,
    family_key text DEFAULT ''::text NOT NULL,
    active_version integer DEFAULT 1 NOT NULL,
    CONSTRAINT document_types_review_interval_days_check CHECK ((review_interval_days > 0))
);
```

## Runtime Usage

Use `rg -n "document_types" internal apps` and the owning module wiki to verify readers/writers before changing this table.

## Seed or Reference Data

Check `db/reference-data/0001_product_reference_data.sql` and `db/dev-seeds/0001_local_dev_seed.sql` before adding rows.

## Notes and Debt

Curated baseline table in the `metaldocs` schema.
