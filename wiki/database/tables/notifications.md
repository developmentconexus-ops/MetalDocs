# metaldocs.notifications

> **Source:** `db/baseline/0001_current_schema.sql`
> **Schema:** `metaldocs`
> **Owner:** platform/workers

## Purpose
Current curated-baseline table owned by `platform/workers`. See the owning module wiki and runtime repositories for business behavior.

## Columns

| Column | Type | Nullable | Meaning |
|---|---|---|---|
| `id` | `text` | no | Baseline column. |
| `recipient_user_id` | `text` | no | Baseline column. |
| `event_type` | `text` | no | Baseline column. |
| `resource_type` | `text` | no | Baseline column. |
| `resource_id` | `text` | no | Baseline column. |
| `title` | `text` | no | Baseline column. |
| `message` | `text` | no | Baseline column. |
| `status` | `text` | no | Baseline column. |
| `idempotency_key` | `text` | no | Baseline column. |
| `created_at` | `timestamp with time zone` | no | Baseline column. |
| `read_at` | `timestamp with time zone` | yes | Baseline column. |

## Baseline Definition

```sql
CREATE TABLE metaldocs.notifications (
id text NOT NULL,
    recipient_user_id text NOT NULL,
    event_type text NOT NULL,
    resource_type text NOT NULL,
    resource_id text NOT NULL,
    title text NOT NULL,
    message text NOT NULL,
    status text NOT NULL,
    idempotency_key text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    read_at timestamp with time zone,
    CONSTRAINT notifications_status_check CHECK ((status = ANY (ARRAY['PENDING'::text, 'SENT'::text, 'READ'::text])))
);
```

## Runtime Usage

Use `rg -n "notifications" internal apps` and the owning module wiki to verify readers/writers before changing this table.

## Seed or Reference Data

Check `db/reference-data/0001_product_reference_data.sql` and `db/dev-seeds/0001_local_dev_seed.sql` before adding rows.

## Notes and Debt

Curated baseline table in the `metaldocs` schema.
