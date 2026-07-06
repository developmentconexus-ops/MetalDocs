# public.approval_signoffs

> **Source:** `db/baseline/0001_current_schema.sql`
> **Schema:** `public`
> **Owner:** approval
> **Last verified:** 2026-07-05 (M7 F7.4: FORCE RLS + `tenant_isolation` policy added, migration `db/migrations/0285_approval_signoffs_rls.sql`)

## Purpose
Current curated-baseline table owned by `approval`. See the owning module wiki and runtime repositories for business behavior.

## Tenant isolation (RLS)

Tenant column is `actor_tenant_id` (not `tenant_id`) — the signer's own tenant, FK-anchored to
`iam_users(tenant_id, user_id)`. This table was missed by the Wave Z (`0237`) and M3 tenant-scoped-table
census because both were driven by a `tenant_id`-column search. Migration `0285_approval_signoffs_rls.sql`
closes the gap: `ENABLE`/`FORCE ROW LEVEL SECURITY` + a `tenant_isolation` policy on `actor_tenant_id`,
same idiom (NULL-GUC escape hatch included) as the rest of the FORCE-RLS set. See
`wiki/decisions/0027-rls-adoption-sequencing.md` Amendment 2026-07-05 for the full rationale and the
`metaldocs_ci` non-owner role that proves this policy is genuinely enforced (not silently bypassed by a
superuser/owner connection).

## Columns

| Column | Type | Nullable | Meaning |
|---|---|---|---|
| `id` | `uuid` | no | Baseline column. |
| `approval_instance_id` | `uuid` | no | Baseline column. |
| `stage_instance_id` | `uuid` | no | Baseline column. |
| `actor_user_id` | `text` | no | Baseline column. |
| `actor_tenant_id` | `uuid` | no | Baseline column. |
| `decision` | `text` | no | Baseline column. |
| `comment` | `text` | yes | Baseline column. |
| `signed_at` | `timestamp with time zone` | no | Baseline column. |
| `signature_method` | `text` | no | Baseline column. |
| `signature_payload` | `jsonb` | no | Baseline column. |
| `content_hash` | `text` | no | Baseline column. |
| `actor_display_name_snapshot` | `text` | yes | Baseline column. |

## Baseline Definition

```sql
CREATE TABLE public.approval_signoffs (
id uuid DEFAULT gen_random_uuid() NOT NULL,
    approval_instance_id uuid NOT NULL,
    stage_instance_id uuid NOT NULL,
    actor_user_id text NOT NULL,
    actor_tenant_id uuid NOT NULL,
    decision text NOT NULL,
    comment text,
    signed_at timestamp with time zone DEFAULT now() NOT NULL,
    signature_method text NOT NULL,
    signature_payload jsonb NOT NULL,
    content_hash text NOT NULL,
    actor_display_name_snapshot text,
    CONSTRAINT approval_signoffs_decision_check CHECK ((decision = ANY (ARRAY['approve'::text, 'reject'::text])))
);
```

## Runtime Usage

Use `rg -n "approval_signoffs" internal apps` and the owning module wiki to verify readers/writers before changing this table.

## Seed or Reference Data

Check `db/reference-data/0001_product_reference_data.sql` and `db/dev-seeds/0001_local_dev_seed.sql` before adding rows.

## Notes and Debt

Retained in `public` because current runtime/baseline truth still uses it. Do not move schemas without an approved migration plan.
