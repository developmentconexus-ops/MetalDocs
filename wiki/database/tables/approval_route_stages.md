# public.approval_route_stages

> **Source:** `db/baseline/0001_current_schema.sql`
> **Schema:** `public`
> **Owner:** approval

## Purpose
Current curated-baseline table owned by `approval`. See the owning module wiki and runtime repositories for business behavior.

## Columns

| Column | Type | Nullable | Meaning |
|---|---|---|---|
| `id` | `uuid` | no | Baseline column. |
| `route_id` | `uuid` | no | Baseline column. |
| `stage_order` | `integer` | no | Baseline column. |
| `name` | `text` | no | Baseline column. |
| `required_role` | `text` | no | Baseline column. |
| `required_capability` | `text` | no | Baseline column. |
| `area_code` | `text` | no | Baseline column. |
| `quorum` | `text` | no | Baseline column. |
| `quorum_m` | `integer` | yes | Baseline column. |
| `on_eligibility_drift` | `text` | no | Baseline column. |

## Baseline Definition

```sql
CREATE TABLE public.approval_route_stages (
id uuid DEFAULT gen_random_uuid() NOT NULL,
    route_id uuid NOT NULL,
    stage_order integer NOT NULL,
    name text NOT NULL,
    required_role text NOT NULL,
    required_capability text NOT NULL,
    area_code text NOT NULL,
    quorum text NOT NULL,
    quorum_m integer,
    on_eligibility_drift text NOT NULL,
    CONSTRAINT approval_route_stages_on_eligibility_drift_check CHECK ((on_eligibility_drift = ANY (ARRAY['reduce_quorum'::text, 'fail_stage'::text, 'keep_snapshot'::text]))),
    CONSTRAINT approval_route_stages_quorum_check CHECK ((quorum = ANY (ARRAY['any_1_of'::text, 'all_of'::text, 'm_of_n'::text]))),
    CONSTRAINT approval_route_stages_quorum_m_consistent CHECK ((((quorum = 'm_of_n'::text) AND (quorum_m IS NOT NULL) AND (quorum_m >= 1)) OR ((quorum <> 'm_of_n'::text) AND (quorum_m IS NULL)))),
    CONSTRAINT approval_route_stages_stage_order_check CHECK ((stage_order >= 1))
);
```

## Runtime Usage

Use `rg -n "approval_route_stages" internal apps` and the owning module wiki to verify readers/writers before changing this table.

## Seed or Reference Data

Check `db/reference-data/0001_product_reference_data.sql` and `db/dev-seeds/0001_local_dev_seed.sql` before adding rows.

## Notes and Debt

Retained in `public` because current runtime/baseline truth still uses it. Do not move schemas without an approved migration plan.
