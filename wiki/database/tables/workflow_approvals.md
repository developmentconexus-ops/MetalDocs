# metaldocs.workflow_approvals

> **Source:** `db/baseline/0001_current_schema.sql`
> **Schema:** `metaldocs`
> **Owner:** approval

## Purpose
Current curated-baseline table owned by `approval`. See the owning module wiki and runtime repositories for business behavior.

## Columns

| Column | Type | Nullable | Meaning |
|---|---|---|---|
| `id` | `text` | no | Baseline column. |
| `document_id` | `text` | no | Baseline column. |
| `requested_by` | `text` | no | Baseline column. |
| `assigned_reviewer` | `text` | no | Baseline column. |
| `decision_by` | `text` | yes | Baseline column. |
| `status` | `text` | no | Baseline column. |
| `request_reason` | `text` | no | Baseline column. |
| `decision_reason` | `text` | yes | Baseline column. |
| `requested_at` | `timestamp with time zone` | no | Baseline column. |
| `decided_at` | `timestamp with time zone` | yes | Baseline column. |

## Baseline Definition

```sql
CREATE TABLE metaldocs.workflow_approvals (
id text NOT NULL,
    document_id text NOT NULL,
    requested_by text NOT NULL,
    assigned_reviewer text NOT NULL,
    decision_by text,
    status text NOT NULL,
    request_reason text DEFAULT ''::text NOT NULL,
    decision_reason text,
    requested_at timestamp with time zone NOT NULL,
    decided_at timestamp with time zone,
    CONSTRAINT workflow_approvals_status_check CHECK ((status = ANY (ARRAY['PENDING'::text, 'APPROVED'::text, 'REJECTED'::text])))
);
```

## Runtime Usage

Use `rg -n "workflow_approvals" internal apps` and the owning module wiki to verify readers/writers before changing this table.

## Seed or Reference Data

Check `db/reference-data/0001_product_reference_data.sql` and `db/dev-seeds/0001_local_dev_seed.sql` before adding rows.

## Notes and Debt

Curated baseline table in the `metaldocs` schema.
