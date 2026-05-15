# metaldocs.idempotency_keys

> **Source:** `db/baseline/0001_current_schema.sql`
> **Schema:** `metaldocs`
> **Owner:** registry

## Purpose
Current curated-baseline table owned by `registry`. See the owning module wiki and runtime repositories for business behavior.

## Columns

| Column | Type | Nullable | Meaning |
|---|---|---|---|
| `tenant_id` | `uuid` | no | Baseline column. |
| `actor_user_id` | `text` | no | Baseline column. |
| `route_template` | `text` | no | Baseline column. |
| `key` | `text` | no | Baseline column. |
| `payload_hash` | `text` | no | Baseline column. |
| `response_status` | `integer` | no | Baseline column. |
| `response_body` | `jsonb` | no | Baseline column. |
| `status` | `text` | no | Baseline column. |
| `created_at` | `timestamp with time zone` | no | Baseline column. |
| `expires_at` | `timestamp with time zone` | no | Baseline column. |

## Baseline Definition

```sql
CREATE TABLE metaldocs.idempotency_keys (
tenant_id uuid NOT NULL,
    actor_user_id text NOT NULL,
    route_template text NOT NULL,
    key text NOT NULL,
    payload_hash text NOT NULL,
    response_status integer NOT NULL,
    response_body jsonb NOT NULL,
    status text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    expires_at timestamp with time zone NOT NULL,
    CONSTRAINT idempotency_keys_status_check CHECK ((status = ANY (ARRAY['in_flight'::text, 'completed'::text, 'failed'::text])))
);
```

## Runtime Usage

Use `rg -n "idempotency_keys" internal apps` and the owning module wiki to verify readers/writers before changing this table.

## Seed or Reference Data

Check `db/reference-data/0001_product_reference_data.sql` and `db/dev-seeds/0001_local_dev_seed.sql` before adding rows.

## Notes and Debt

Curated baseline table in the `metaldocs` schema.
