# public.editor_sessions

> **Source:** `db/baseline/0001_current_schema.sql`
> **Schema:** `public`
> **Owner:** documents

## Purpose
Current curated-baseline table owned by `documents`. See the owning module wiki and runtime repositories for business behavior.

## Columns

| Column | Type | Nullable | Meaning |
|---|---|---|---|
| `id` | `uuid` | no | Baseline column. |
| `document_id` | `uuid` | no | Baseline column. |
| `user_id` | `text` | no | Baseline column. |
| `acquired_at` | `timestamp with time zone` | no | Baseline column. |
| `expires_at` | `timestamp with time zone` | no | Baseline column. |
| `released_at` | `timestamp with time zone` | yes | Baseline column. |
| `last_acknowledged_revision_id` | `uuid` | no | Baseline column. |
| `status` | `text` | no | Baseline column. |

## Baseline Definition

```sql
CREATE TABLE public.editor_sessions (
id uuid DEFAULT gen_random_uuid() NOT NULL,
    document_id uuid NOT NULL,
    user_id text NOT NULL,
    acquired_at timestamp with time zone DEFAULT now() NOT NULL,
    expires_at timestamp with time zone NOT NULL,
    released_at timestamp with time zone,
    last_acknowledged_revision_id uuid NOT NULL,
    status text NOT NULL,
    CONSTRAINT editor_sessions_status_check CHECK ((status = ANY (ARRAY['active'::text, 'expired'::text, 'released'::text, 'force_released'::text])))
);
```

## Runtime Usage

Use `rg -n "editor_sessions" internal apps` and the owning module wiki to verify readers/writers before changing this table.

## Seed or Reference Data

Check `db/reference-data/0001_product_reference_data.sql` and `db/dev-seeds/0001_local_dev_seed.sql` before adding rows.

## Notes and Debt

Retained in `public` because current runtime/baseline truth still uses it. Do not move schemas without an approved migration plan.
