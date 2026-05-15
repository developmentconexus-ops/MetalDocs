# metaldocs.document_sequences

> **Source:** `db/baseline/0001_current_schema.sql`
> **Schema:** `metaldocs`
> **Owner:** unknown-owner

## Purpose
Current curated-baseline table owned by `unknown-owner`. See the owning module wiki and runtime repositories for business behavior.

## Columns

| Column | Type | Nullable | Meaning |
|---|---|---|---|
| `profile_code` | `text` | no | Baseline column. |
| `next_value` | `integer` | no | Baseline column. |

## Baseline Definition

```sql
CREATE TABLE metaldocs.document_sequences (
profile_code text NOT NULL,
    next_value integer NOT NULL,
    CONSTRAINT document_sequences_next_value_check CHECK ((next_value > 0))
);
```

## Runtime Usage

Use `rg -n "document_sequences" internal apps` and the owning module wiki to verify readers/writers before changing this table.

## Seed or Reference Data

Check `db/reference-data/0001_product_reference_data.sql` and `db/dev-seeds/0001_local_dev_seed.sql` before adding rows.

## Notes and Debt

Curated baseline table in the `metaldocs` schema.
