# public.approval_route_stage_selectors

**Schema:** `public.approval_route_stage_selectors`
**Owner:** approval module (unit 3.2 / M4, migration 0303)
**Last verified:** 2026-08-07

## Purpose

Holds the ordered actor-resolution selectors for a route stage — who is
eligible to act on that stage, expressed as one of four `kind`s
(`named_user`, `role_in_fixed_area`, `role_in_document_area`,
`submit_choice`), each with its own required/forbidden column combination
enforced by the `approval_route_stage_selectors_fields_consistent` CHECK
constraint. `insertRouteStageSelectors` correlates each selector to its
just-inserted parent `approval_route_stages` row entirely in SQL (a single
`INSERT ... SELECT ... JOIN` keyed on `(route_id, stage_order)`), because the
batched stage insert has no `RETURNING` clause to hand back generated stage
ids.

## Columns

| Column           | Type      | Notes |
|------------------|-----------|-------|
| `id`             | `uuid` PK | Default `gen_random_uuid()`. |
| `tenant_id`      | `uuid`    | Not null. Carries its own tenant scoping independently of the parent stage (migration 0303 design choice — see Tenant scoping). |
| `route_stage_id` | `uuid`    | Not null. FK → `public.approval_route_stages(id)` ON DELETE CASCADE. Part of the `(route_stage_id, selector_order)` unique constraint. |
| `selector_order` | `integer` | Not null. Ordering within the stage; unique per `route_stage_id`. |
| `kind`           | `text`    | Not null. CHECK `kind IN ('named_user','role_in_fixed_area','role_in_document_area','submit_choice')`. |
| `user_id`        | `text`    | Nullable. Required (only) when `kind = 'named_user'`. |
| `role`           | `text`    | Nullable. Required when `kind` is any role-based variant. |
| `area_code`      | `text`    | Nullable. Required when `kind IN ('role_in_fixed_area','submit_choice')`, forbidden otherwise. |

`approval_route_stage_selectors_fields_consistent` CHECK enforces the exact
required/forbidden column combination per `kind` (see migration/DDL for the
full boolean expression).

## Migrations

| Version | Description |
|---------|-------------|
| `0303`  | Unit 3.2 / M4: introduced this table with its own `tenant_id` (called out in `tenant_data_port.go` as carrying tenant scoping independently of the parent, rather than inheriting solely via the `route_stage_id` FK). |

## Key callers

- `internal/modules/approval/application/route_admin_service.go::insertRouteStageSelectors` — batched `INSERT ... SELECT ... JOIN` against `approval_route_stages`, in-tx.
- `internal/modules/approval/application/route_admin_service.go` (~line 1122) — reads selectors for a stage.
- `internal/modules/approval/infrastructure/postgres_approval_repository.go` (~line 2175) — reads route stage selectors for run-side resolution.
- `internal/modules/approval/infrastructure/tenant_data_port.go::TenantDataPort` — tenant export/erase port, keyed on `tenant_id`.
- `internal/test/e2e_seed.go` — E2E test seed helper inserts selector rows directly (test-only).

## Tenant scoping

`tenant_id` (uuid, not null) carries tenant scoping on this table directly —
`tenant_data_port.go` notes this is deliberate (unit 3.2/M4, migration 0303):
the row is not scoped solely via its `route_stage_id` FK chain. There is also
an index `ix_approval_route_stage_selectors_tenant_id` on `tenant_id` and a
separate `ix_approval_route_stage_selectors_route_stage_id` index. Unlike most
tenant tables in this dictionary, this table has **no RLS policy and no
`ENABLE`/`FORCE ROW LEVEL SECURITY`** in the baseline DDL — tenant isolation
here is app-level only (explicit `tenant_id` predicates and the parent FK's
`ON DELETE CASCADE`), not DB-enforced. A cross-tenant lookup returns no rows
only because callers filter by `tenant_id` explicitly; there is no DB tripwire
if a caller omitted that predicate.
