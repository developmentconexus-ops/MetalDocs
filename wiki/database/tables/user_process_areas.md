# public.user_process_areas

> **Last verified:** 2026-06-04
> **Source:** `db/baseline/0001_current_schema.sql`
> **Schema:** `public`
> **Owner:** iam

## Purpose
Current curated-baseline table owned by `iam`. See the owning module wiki and runtime repositories for business behavior.

## Temporal model (active vs revoked) — ADR 0037

This table uses **soft-delete with a current marker**, not a future-dated validity interval:

- A membership is **active ⟺ `effective_to IS NULL`**. This is enforced by the partial indexes
  `ux_user_process_areas_single_active` / `ux_user_process_areas_one_active` /
  `ix_user_process_areas_active`, all `WHERE effective_to IS NULL` (the UNIQUE one guarantees at most
  one active row per `(tenant_id,user_id,area_code,role)`).
- **Revoke** stamps `effective_to = now()` (a past tombstone) + `revoked_by`; the row is retained for
  history. No code path writes a **future** `effective_to`, and `Grant` exposes no end-date argument.
- **Active-now reads** therefore filter `effective_to IS NULL` (authz `Require`, CD/search visibility,
  membership directory). **As-of / history reads** that take a point-in-time parameter `$t` use
  `(effective_to IS NULL OR effective_to > $t)` — a *different question*, correctly answered by the
  interval form. Both are correct; this is not drift.
- Do **not** change the active-now predicate to the interval form: it would grant nothing (no
  future-dated rows exist), regress authz off the partial indexes, and contradict the unique index.
  See **ADR 0037** (`wiki/decisions/0037-membership-temporal-model.md`), which refutes re-audit
  2026-06-16 Major #1 on this point. Time-bounded/scheduled memberships (Model B) are gated behind a
  successor ADR.

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
| `revoked_by` | `text` | yes | Actor who revoked the membership. Must be non-NULL whenever `effective_to` is set (`revoked_by_required_when_revoked` CHECK). Writers: `UserAreaRepository.CloseActive` and `GrantAtomic` set this alongside `effective_to`. |

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
    CONSTRAINT user_process_areas_role_check CHECK ((role = ANY (ARRAY['viewer'::text, 'editor'::text, 'approver'::text, 'author'::text, 'signer'::text, 'area_admin'::text, 'qms_admin'::text])))
);
```

> **Role set (7 canonical area roles).** `system_admin` is intentionally excluded — it is a tenant-wide tier-1 role that bypasses tier-2 and is never an area membership. The legacy `reviewer` role was **decommissioned** (ADR 0022, migration `0230`): it granted zero capabilities and was absent from the Go registry, seed, and OpenAPI. This CHECK is the single source of truth the Go `iamdomain.IsAreaRole` set and the approval-stage `required_role` validation bind against.

## Runtime Usage

Use `rg -n "user_process_areas" internal apps` and the owning module wiki to verify readers/writers before changing this table.

## Seed or Reference Data

Check `db/reference-data/0001_product_reference_data.sql` and `db/dev-seeds/0001_local_dev_seed.sql` before adding rows. As of 2026-06-03 the dev seed defines three process areas (`rh`, `qualidade`, `producao`) with memberships seeded in `rh` and `qualidade`; `producao` is intentionally empty so QA can exercise both grant paths (role-change via `GrantAtomic` and new-pair via `Insert`).

## Notes and Debt

Retained in `public` because current runtime/baseline truth still uses it. Do not move schemas without an approved migration plan.
