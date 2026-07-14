# public.approval_stage_instances

> **Source:** `db/baseline/0001_current_schema.sql`
> **Schema:** `public`
> **Owner:** approval

## Purpose
Current curated-baseline table owned by `approval`. See the owning module wiki and runtime repositories for business behavior.

## Columns

| Column | Type | Nullable | Meaning |
|---|---|---|---|
| `id` | `uuid` | no | Baseline column. |
| `approval_instance_id` | `uuid` | no | Baseline column. |
| `stage_order` | `integer` | no | Baseline column. |
| `name_snapshot` | `text` | no | Baseline column. |
| `required_role_snapshot` | `text` | no | Baseline column. |
| `required_capability_snapshot` | `text` | no | Baseline column. |
| `area_code_snapshot` | `text` | no | Baseline column. |
| `quorum_snapshot` | `text` | no | Baseline column. |
| `quorum_m_snapshot` | `integer` | yes | Baseline column. |
| `on_eligibility_drift_snapshot` | `text` | no | Baseline column. |
| `eligible_actor_ids` | `jsonb` | no | Baseline column. |
| `effective_denominator` | `integer` | yes | Baseline column. |
| `status` | `text` | no | Baseline column. |
| `opened_at` | `timestamp with time zone` | yes | Baseline column. |
| `completed_at` | `timestamp with time zone` | yes | Baseline column. |

## Baseline Definition

```sql
CREATE TABLE public.approval_stage_instances (
id uuid DEFAULT gen_random_uuid() NOT NULL,
    approval_instance_id uuid NOT NULL,
    stage_order integer NOT NULL,
    name_snapshot text NOT NULL,
    required_role_snapshot text NOT NULL,
    required_capability_snapshot text NOT NULL,
    area_code_snapshot text NOT NULL,
    quorum_snapshot text NOT NULL,
    quorum_m_snapshot integer,
    on_eligibility_drift_snapshot text NOT NULL,
    eligible_actor_ids jsonb NOT NULL,
    effective_denominator integer,
    status text NOT NULL,
    opened_at timestamp with time zone,
    completed_at timestamp with time zone,
    CONSTRAINT approval_stage_instances_on_eligibility_drift_snapshot_check CHECK ((on_eligibility_drift_snapshot = ANY (ARRAY['reduce_quorum'::text, 'fail_stage'::text, 'keep_snapshot'::text]))),
    CONSTRAINT approval_stage_instances_quorum_snapshot_check CHECK ((quorum_snapshot = ANY (ARRAY['any_1_of'::text, 'all_of'::text, 'm_of_n'::text]))),
    CONSTRAINT approval_stage_instances_stage_order_check CHECK ((stage_order >= 1)),
    CONSTRAINT approval_stage_instances_status_check CHECK ((status = ANY (ARRAY['pending'::text, 'active'::text, 'completed'::text, 'skipped'::text, 'rejected_here'::text, 'cancelled'::text])))
);
```

## Runtime Usage

Use `rg -n "approval_stage_instances" internal apps` and the owning module wiki to verify readers/writers before changing this table.

## Seed or Reference Data

Check `db/reference-data/0001_product_reference_data.sql` and `db/dev-seeds/0001_local_dev_seed.sql` before adding rows.

## Notes and Debt

Retained in `public` because current runtime/baseline truth still uses it. Do not move schemas without an approved migration plan.
