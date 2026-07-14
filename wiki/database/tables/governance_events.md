# public.governance_events

> **Source:** `db/baseline/0001_current_schema.sql`
> **Schema:** `public`
> **Owner:** audit

## Purpose
Current curated-baseline table owned by `audit`. See the owning module wiki and runtime repositories for business behavior.

## Columns

| Column | Type | Nullable | Meaning |
|---|---|---|---|
| `id` | `uuid` | no | Baseline column. |
| `tenant_id` | `uuid` | no | Baseline column. |
| `event_type` | `text` | no | Baseline column. |
| `actor_user_id` | `text` | no | Baseline column. |
| `resource_type` | `text` | no | Baseline column. |
| `resource_id` | `text` | no | Baseline column. |
| `reason` | `text` | yes | Baseline column. |
| `payload_json` | `jsonb` | no | Baseline column. |
| `created_at` | `timestamp with time zone` | no | Baseline column. |
| `dedupe_key` | `text` | yes | Baseline column. |
| `correlation_id` | `text` | yes | Baseline column. |

## Baseline Definition

```sql
CREATE TABLE public.governance_events (
id uuid DEFAULT gen_random_uuid() NOT NULL,
    tenant_id uuid NOT NULL,
    event_type text NOT NULL,
    actor_user_id text NOT NULL,
    resource_type text NOT NULL,
    resource_id text NOT NULL,
    reason text,
    payload_json jsonb NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    dedupe_key text,
    correlation_id text
);
```

## Runtime Usage

Use `rg -n "governance_events" internal apps` and the owning module wiki to verify readers/writers before changing this table.

## Seed or Reference Data

Check `db/reference-data/0001_product_reference_data.sql` and `db/dev-seeds/0001_local_dev_seed.sql` before adding rows.

## Notes and Debt

Retained in `public` because current runtime/baseline truth still uses it. Do not move schemas without an approved migration plan.
