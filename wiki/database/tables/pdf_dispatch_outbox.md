# metaldocs.pdf_dispatch_outbox

> **Source:** `db/baseline/0001_current_schema.sql`
> **Schema:** `metaldocs`
> **Owner:** platform/workers

## Purpose
Current curated-baseline table owned by `platform/workers`. See the owning module wiki and runtime repositories for business behavior.

## Columns

| Column | Type | Nullable | Meaning |
|---|---|---|---|
| `id` | `uuid` | no | Baseline column. |
| `tenant_id` | `uuid` | no | Baseline column. |
| `revision_id` | `uuid` | no | Baseline column. |
| `content_hash` | `bytea` | no | Baseline column. |
| `status` | `text` | no | Baseline column. |
| `attempts` | `integer` | no | Baseline column. |
| `last_error` | `text` | yes | Baseline column. |
| `claimed_at` | `timestamp with time zone` | yes | Baseline column. |
| `next_retry_at` | `timestamp with time zone` | no | Baseline column. |
| `created_at` | `timestamp with time zone` | no | Baseline column. |
| `dispatched_at` | `timestamp with time zone` | yes | Baseline column. |

## Baseline Definition

```sql
CREATE TABLE metaldocs.pdf_dispatch_outbox (
id uuid DEFAULT gen_random_uuid() NOT NULL,
    tenant_id uuid NOT NULL,
    revision_id uuid NOT NULL,
    content_hash bytea NOT NULL,
    status text DEFAULT 'pending'::text NOT NULL,
    attempts integer DEFAULT 0 NOT NULL,
    last_error text,
    claimed_at timestamp with time zone,
    next_retry_at timestamp with time zone DEFAULT now() NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    dispatched_at timestamp with time zone,
    CONSTRAINT pdf_dispatch_outbox_status_check CHECK ((status = ANY (ARRAY['pending'::text, 'processing'::text, 'dispatched'::text, 'failed'::text])))
);
```

## Runtime Usage

Use `rg -n "pdf_dispatch_outbox" internal apps` and the owning module wiki to verify readers/writers before changing this table.

## Seed or Reference Data

Check `db/reference-data/0001_product_reference_data.sql` and `db/dev-seeds/0001_local_dev_seed.sql` before adding rows.

## Notes and Debt

Curated baseline table in the `metaldocs` schema.
