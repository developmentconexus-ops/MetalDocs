# metaldocs.document_profile_governance

> **Source:** `db/baseline/0001_current_schema.sql`
> **Schema:** `metaldocs`
> **Owner:** taxonomy

## Purpose
Current curated-baseline table owned by `taxonomy`. See the owning module wiki and runtime repositories for business behavior.

## Columns

| Column | Type | Nullable | Meaning |
|---|---|---|---|
| `profile_code` | `text` | no | Baseline column. |
| `workflow_profile` | `text` | no | Baseline column. |
| `review_interval_days` | `integer` | no | Baseline column. |
| `approval_required` | `boolean` | no | Baseline column. |
| `retention_days` | `integer` | no | Baseline column. |
| `validity_days` | `integer` | no | Baseline column. |
| `created_at` | `timestamp with time zone` | no | Baseline column. |

## Baseline Definition

```sql
CREATE TABLE metaldocs.document_profile_governance (
profile_code text NOT NULL,
    workflow_profile text DEFAULT 'standard_approval'::text NOT NULL,
    review_interval_days integer NOT NULL,
    approval_required boolean DEFAULT true NOT NULL,
    retention_days integer DEFAULT 0 NOT NULL,
    validity_days integer DEFAULT 0 NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT document_profile_governance_retention_days_check CHECK ((retention_days >= 0)),
    CONSTRAINT document_profile_governance_review_interval_days_check CHECK ((review_interval_days > 0)),
    CONSTRAINT document_profile_governance_validity_days_check CHECK ((validity_days >= 0))
);
```

## Runtime Usage

Use `rg -n "document_profile_governance" internal apps` and the owning module wiki to verify readers/writers before changing this table.

## Seed or Reference Data

Check `db/reference-data/0001_product_reference_data.sql` and `db/dev-seeds/0001_local_dev_seed.sql` before adding rows.

## Notes and Debt

Curated baseline table in the `metaldocs` schema.
