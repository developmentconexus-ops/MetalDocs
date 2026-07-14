# metaldocs.document_edit_locks

> **Source:** `db/baseline/0001_current_schema.sql`
> **Schema:** `metaldocs`
> **Owner:** documents

## Purpose
Current curated-baseline table owned by `documents`. See the owning module wiki and runtime repositories for business behavior.

## Columns

| Column | Type | Nullable | Meaning |
|---|---|---|---|
| `document_id` | `text` | no | Baseline column. |
| `locked_by` | `text` | no | Baseline column. |
| `display_name` | `text` | no | Baseline column. |
| `lock_reason` | `text` | no | Baseline column. |
| `acquired_at` | `timestamp with time zone` | no | Baseline column. |
| `expires_at` | `timestamp with time zone` | no | Baseline column. |
| `created_at` | `timestamp with time zone` | no | Baseline column. |
| `updated_at` | `timestamp with time zone` | no | Baseline column. |

## Baseline Definition

```sql
CREATE TABLE metaldocs.document_edit_locks (
document_id text NOT NULL,
    locked_by text NOT NULL,
    display_name text NOT NULL,
    lock_reason text DEFAULT ''::text NOT NULL,
    acquired_at timestamp with time zone NOT NULL,
    expires_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT chk_document_edit_locks_expiry CHECK ((expires_at > acquired_at))
);
```

## Runtime Usage

Use `rg -n "document_edit_locks" internal apps` and the owning module wiki to verify readers/writers before changing this table.

## Seed or Reference Data

Check `db/reference-data/0001_product_reference_data.sql` and `db/dev-seeds/0001_local_dev_seed.sql` before adding rows.

## Notes and Debt

Curated baseline table in the `metaldocs` schema.
