# public.approval_instances

> **Source:** `db/baseline/0001_current_schema.sql`
> **Schema:** `public`
> **Owner:** approval

## Purpose
Current curated-baseline table owned by `approval`. See the owning module wiki and runtime repositories for business behavior.

## Columns

| Column | Type | Nullable | Meaning |
|---|---|---|---|
| `id` | `uuid` | no | Baseline column. |
| `tenant_id` | `uuid` | no | Baseline column. |
| `document_id` | `uuid` | no | Baseline column. |
| `route_id` | `uuid` | no | Baseline column. |
| `route_version_snapshot` | `integer` | no | Baseline column. |
| `status` | `text` | no | Baseline column. |
| `submitted_by` | `text` | no | Baseline column. |
| `submitted_at` | `timestamp with time zone` | no | Baseline column. |
| `completed_at` | `timestamp with time zone` | yes | Baseline column. |
| `content_hash_at_submit` | `text` | no | Baseline column. |
| `idempotency_key` | `text` | no | Baseline column. |

## Baseline Definition

```sql
CREATE TABLE public.approval_instances (
id uuid DEFAULT gen_random_uuid() NOT NULL,
    tenant_id uuid NOT NULL,
    document_id uuid NOT NULL,
    route_id uuid NOT NULL,
    route_version_snapshot integer NOT NULL,
    status text NOT NULL,
    submitted_by text NOT NULL,
    submitted_at timestamp with time zone DEFAULT now() NOT NULL,
    completed_at timestamp with time zone,
    content_hash_at_submit text NOT NULL,
    idempotency_key text NOT NULL,
    CONSTRAINT approval_instances_status_check CHECK ((status = ANY (ARRAY['in_progress'::text, 'approved'::text, 'rejected'::text, 'cancelled'::text])))
);
```

## Runtime Usage

Use `rg -n "approval_instances" internal apps` and the owning module wiki to verify readers/writers before changing this table.

## Seed or Reference Data

Check `db/reference-data/0001_product_reference_data.sql` and `db/dev-seeds/0001_local_dev_seed.sql` before adding rows.

## Notes and Debt

Retained in `public` because current runtime/baseline truth still uses it. Do not move schemas without an approved migration plan.
