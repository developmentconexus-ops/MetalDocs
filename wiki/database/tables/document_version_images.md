# metaldocs.document_version_images

> **Source:** `db/baseline/0001_current_schema.sql`
> **Schema:** `metaldocs`
> **Owner:** documents

## Purpose
Current curated-baseline table owned by `documents`. See the owning module wiki and runtime repositories for business behavior.

## Columns

| Column | Type | Nullable | Meaning |
|---|---|---|---|
| `document_version_id` | `uuid` | no | Baseline column. |
| `image_id` | `uuid` | no | Baseline column. |

## Baseline Definition

```sql
CREATE TABLE metaldocs.document_version_images (
document_version_id uuid NOT NULL,
    image_id uuid NOT NULL
);
```

## Runtime Usage

Use `rg -n "document_version_images" internal apps` and the owning module wiki to verify readers/writers before changing this table.

## Seed or Reference Data

Check `db/reference-data/0001_product_reference_data.sql` and `db/dev-seeds/0001_local_dev_seed.sql` before adding rows.

## Notes and Debt

Curated baseline table in the `metaldocs` schema.
