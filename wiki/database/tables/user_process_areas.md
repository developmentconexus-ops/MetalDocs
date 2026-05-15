# public.user_process_areas

> **Source:** `db/baseline/0001_current_schema.sql`
> **Schema:** `public`
> **Owner:** iam

## Purpose
Current curated-baseline table owned by `iam`. See the owning module wiki and runtime repositories for business behavior.

## Columns

| Column | Type | Nullable | Meaning |
|---|---|---|---|
| `user_id` | `text` | no | Baseline column. |
| `tenant_id` | `uuid` | no | Baseline column. |
| `area_code` | `text` | no | Baseline column. |
| `role` | `text` | no | Baseline column. |
| `effective_from` | `timestamp with time zone` | no | Baseline column. |
| `effective_to` | `timestamp with time zone` | yes | Baseline column. |
| `granted_by` | `text` | yes | Baseline column. |
| `revoked_by` | `text` | yes | Baseline column. |

## Baseline Definition

```sql
CREATE TABLE public.user_process_areas (
user_id text NOT NULL,
    tenant_id uuid NOT NULL,
    area_code text NOT NULL,
    role text NOT NULL,
    effective_from timestamp with time zone NOT NULL,
    effective_to timestamp with time zone,
    granted_by text,
    revoked_by text,
    CONSTRAINT effective_interval_valid CHECK (((effective_to IS NULL) OR (effective_to > effective_from))),
    CONSTRAINT revoked_by_required_when_revoked CHECK ((((effective_to IS NULL) AND (revoked_by IS NULL)) OR ((effective_to IS NOT NULL) AND (revoked_by IS NOT NULL)))),
    CONSTRAINT user_process_areas_role_check CHECK ((role = ANY (ARRAY['viewer'::text, 'editor'::text, 'reviewer'::text, 'approver'::text, 'author'::text, 'signer'::text, 'area_admin'::text, 'qms_admin'::text])))
);
```

## Runtime Usage

Use `rg -n "user_process_areas" internal apps` and the owning module wiki to verify readers/writers before changing this table.

## Seed or Reference Data

Check `db/reference-data/0001_product_reference_data.sql` and `db/dev-seeds/0001_local_dev_seed.sql` before adding rows.

## Notes and Debt

Retained in `public` because current runtime/baseline truth still uses it. Do not move schemas without an approved migration plan.
