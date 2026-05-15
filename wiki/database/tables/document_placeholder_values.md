# public.document_placeholder_values

> **Source:** `db/baseline/0001_current_schema.sql`
> **Schema:** `public`
> **Owner:** documents

## Purpose
Current curated-baseline table owned by `documents`. See the owning module wiki and runtime repositories for business behavior.

## Columns

| Column | Type | Nullable | Meaning |
|---|---|---|---|
| `tenant_id` | `uuid` | no | Baseline column. |
| `revision_id` | `uuid` | no | Baseline column. |
| `placeholder_id` | `text` | no | Baseline column. |
| `value_text` | `text` | yes | Baseline column. |
| `value_typed` | `jsonb` | yes | Baseline column. |
| `source` | `text` | no | Baseline column. |
| `computed_from` | `text` | yes | Baseline column. |
| `resolver_version` | `integer` | yes | Baseline column. |
| `inputs_hash` | `bytea` | yes | Baseline column. |
| `validated_at` | `timestamp with time zone` | yes | Baseline column. |
| `created_at` | `timestamp with time zone` | no | Baseline column. |
| `updated_at` | `timestamp with time zone` | no | Baseline column. |

## Baseline Definition

```sql
CREATE TABLE public.document_placeholder_values (
tenant_id uuid NOT NULL,
    revision_id uuid NOT NULL,
    placeholder_id text NOT NULL,
    value_text text,
    value_typed jsonb,
    source text NOT NULL,
    computed_from text,
    resolver_version integer,
    inputs_hash bytea,
    validated_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT document_placeholder_values_inputs_hash_check CHECK (((inputs_hash IS NULL) OR (octet_length(inputs_hash) = 32))),
    CONSTRAINT document_placeholder_values_source_check CHECK ((source = ANY (ARRAY['user'::text, 'computed'::text, 'default'::text])))
);
```

## Runtime Usage

Use `rg -n "document_placeholder_values" internal apps` and the owning module wiki to verify readers/writers before changing this table.

## Seed or Reference Data

Check `db/reference-data/0001_product_reference_data.sql` and `db/dev-seeds/0001_local_dev_seed.sql` before adding rows.

## Notes and Debt

Retained in `public` because current runtime/baseline truth still uses it. Do not move schemas without an approved migration plan.
