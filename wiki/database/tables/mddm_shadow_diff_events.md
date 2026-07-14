# metaldocs.mddm_shadow_diff_events

> **Source:** `db/baseline/0001_current_schema.sql`
> **Schema:** `metaldocs`
> **Owner:** documents

## Purpose
Current curated-baseline table owned by `documents`. See the owning module wiki and runtime repositories for business behavior.

## Columns

| Column | Type | Nullable | Meaning |
|---|---|---|---|
| `id` | `bigint` | no | Baseline column. |
| `document_id` | `character varying(64)` | no | Baseline column. |
| `version_number` | `integer` | no | Baseline column. |
| `user_id_hash` | `character varying(64)` | no | Baseline column. |
| `current_xml_hash` | `character varying(64)` | no | Baseline column. |
| `shadow_xml_hash` | `character varying(64)` | no | Baseline column. |
| `diff_summary` | `jsonb` | no | Baseline column. |
| `current_duration_ms` | `integer` | no | Baseline column. |
| `shadow_duration_ms` | `integer` | no | Baseline column. |
| `shadow_error` | `text` | yes | Baseline column. |
| `recorded_at` | `timestamp with time zone` | no | Baseline column. |
| `trace_id` | `character varying(64)` | yes | Baseline column. |

## Baseline Definition

```sql
CREATE TABLE metaldocs.mddm_shadow_diff_events (
id bigint NOT NULL,
    document_id character varying(64) NOT NULL,
    version_number integer NOT NULL,
    user_id_hash character varying(64) NOT NULL,
    current_xml_hash character varying(64) NOT NULL,
    shadow_xml_hash character varying(64) NOT NULL,
    diff_summary jsonb DEFAULT '{}'::jsonb NOT NULL,
    current_duration_ms integer DEFAULT 0 NOT NULL,
    shadow_duration_ms integer DEFAULT 0 NOT NULL,
    shadow_error text,
    recorded_at timestamp with time zone DEFAULT now() NOT NULL,
    trace_id character varying(64)
);
```

## Runtime Usage

Use `rg -n "mddm_shadow_diff_events" internal apps` and the owning module wiki to verify readers/writers before changing this table.

## Seed or Reference Data

Check `db/reference-data/0001_product_reference_data.sql` and `db/dev-seeds/0001_local_dev_seed.sql` before adding rows.

## Notes and Debt

Curated baseline table in the `metaldocs` schema.
