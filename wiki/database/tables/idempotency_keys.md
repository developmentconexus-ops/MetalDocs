# metaldocs.idempotency_keys

> **Source:** `db/baseline/0001_current_schema.sql`
> **Schema:** `metaldocs`
> **Owner:** platform/idempotency

## Purpose
Current curated-baseline table owned by the platform idempotency layer (`internal/platform/idempotency`). See the owning module wiki and runtime repositories for business behavior.

## Columns

| Column | Type | Nullable | Meaning |
|---|---|---|---|
| `tenant_id` | `uuid` | no | Baseline column. |
| `actor_user_id` | `text` | no | Baseline column. |
| `route_template` | `text` | no | Baseline column. |
| `key` | `text` | no | Baseline column. |
| `payload_hash` | `text` | no | Baseline column. |
| `response_status` | `integer` | yes | Null while the request is still `in_flight`; required once the row is `completed`. |
| `response_body` | `bytea` | yes | Raw replay bytes; null while the request is still `in_flight`; required once the row is `completed`. |
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
    response_status integer,
    response_body bytea,
    status text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    expires_at timestamp with time zone NOT NULL,
    CONSTRAINT idempotency_keys_actor_nonempty CHECK ((actor_user_id <> ''::text)),
    CONSTRAINT idempotency_keys_completed_has_response CHECK (((status <> 'completed'::text) OR ((response_status IS NOT NULL) AND (response_body IS NOT NULL)))),
    CONSTRAINT idempotency_keys_response_body_size CHECK (((response_body IS NULL) OR (octet_length(response_body) <= 65536))),
    CONSTRAINT idempotency_keys_response_status_range CHECK (((response_status IS NULL) OR ((response_status >= 100) AND (response_status <= 599)))),
    CONSTRAINT idempotency_keys_status_check CHECK ((status = ANY (ARRAY['in_flight'::text, 'completed'::text, 'failed'::text])))
);
```

## Correctness Constraints

- `idempotency_keys_completed_has_response` enforces that completed rows always store both `response_status` and `response_body`.
- `idempotency_keys_response_status_range` restricts stored replay statuses to valid HTTP status codes.
- `idempotency_keys_response_body_size` caps stored replay bodies at 64 KiB.
- `idempotency_keys_actor_nonempty` rejects empty-string actors; service actors remain allowed because the column is still `text`.

## Important Indexes

- `idempotency_keys_pkey` on `(tenant_id, actor_user_id, route_template, key)` scopes replay safety by tenant, actor, and route template.
- `idx_idempotency_keys_completed_expires` supports TTL cleanup for completed rows.
- `idx_idempotency_keys_in_flight_expires` supports orphan detection/sweep for stale `in_flight` rows.

## Runtime Usage

Use `rg -n "idempotency_keys" internal apps` and the owning module wiki to verify readers/writers before changing this table.

## Seed or Reference Data

Check `db/reference-data/0001_product_reference_data.sql` and `db/dev-seeds/0001_local_dev_seed.sql` before adding rows.

## Notes and Debt

Curated baseline table in the `metaldocs` schema.
