# metaldocs.document_profile_schema_versions

> **Source:** `db/baseline/0001_current_schema.sql`
> **Schema:** `metaldocs`
> **Owner:** taxonomy

## Purpose
Current curated-baseline table owned by `taxonomy`. See the owning module wiki and runtime repositories for business behavior.

## Columns

| Column | Type | Nullable | Meaning |
|---|---|---|---|
| `profile_code` | `text` | no | Baseline column. |
| `version` | `integer` | no | Baseline column. |
| `metadata_rules_json` | `jsonb` | no | Baseline column. |
| `is_active` | `boolean` | no | Baseline column. |
| `created_at` | `timestamp with time zone` | no | Baseline column. |
| `content_schema_json` | `jsonb` | no | Baseline column. |

## Baseline Definition

```sql
CREATE TABLE metaldocs.document_profile_schema_versions (
profile_code text NOT NULL,
    version integer NOT NULL,
    metadata_rules_json jsonb DEFAULT '[]'::jsonb NOT NULL,
    is_active boolean DEFAULT false NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    content_schema_json jsonb DEFAULT '{}'::jsonb NOT NULL,
    CONSTRAINT document_profile_schema_versions_version_check CHECK ((version > 0))
);
```

## Runtime Usage

Use `rg -n "document_profile_schema_versions" internal apps` and the owning module wiki to verify readers/writers before changing this table.

## Seed or Reference Data

Check `db/reference-data/0001_product_reference_data.sql` and `db/dev-seeds/0001_local_dev_seed.sql` before adding rows.

## Notes and Debt

Curated baseline table in the `metaldocs` schema.
