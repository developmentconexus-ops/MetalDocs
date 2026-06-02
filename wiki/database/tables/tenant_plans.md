# tenant_plans

**Schema:** `metaldocs.tenant_plans`
**Owner:** IAM Admin Center observability (PR-8)
**Last verified:** 2026-06-02

## Purpose

Stores the per-tenant plan envelope that backs the read-only **Usage** card on the Admin Center Overview tab. The columns describe *quotas* (seats, storage, API calls) and the plan *label* (`free | pro | enterprise`).

This table is **read-only at Tier-B** (tenant-admin observability). Mutation flows — upgrade, downgrade, billing, overage — belong to the Tier-A platform-owner surface and are tracked in [`wiki/modules/iam-tech-debt.md`](../../modules/iam-tech-debt.md).

## Columns

| Column                     | Type        | Notes |
|----------------------------|-------------|-------|
| `tenant_id`                | `uuid` PK   | FK → `metaldocs.tenants(id)` ON DELETE CASCADE. |
| `plan_tier`                | `text`      | `free | pro | enterprise` (CHECK constraint). |
| `seats_allocated`          | `integer`   | Default 100. ≥ 0. |
| `storage_allocated_bytes`  | `bigint`    | Default 50 GiB (53 687 091 200). ≥ 0. |
| `api_calls_allocated`      | `integer`   | Default 1 000 000 per billing window. ≥ 0. |
| `created_at` / `updated_at`| `timestamptz` | Default `now()`. |

## Migrations

| Version | Description |
|---------|-------------|
| `0221`  | PR-8: create table + backfill existing tenants with Pro defaults. Forward-only, idempotent. |

## Key callers

- `internal/modules/iam/infrastructure/postgres/observability_repository.go::GetTenantPlan` — only reader (PR-8).
- `internal/modules/iam/application/observability_service.go::GetUsage` — composes seats/storage/api-call quotas into the Usage snapshot.

## Tenant scoping

PK is `tenant_id` (uuid). `GetTenantPlan(ctx, tenantID)` returns `ErrTenantPlanNotFound` when the row is absent — the observability service treats that as "unknown allocation" and returns zero/empty fields, never another tenant's plan.

## Tech-debt

- Plan **mutation** is not implemented. The Tier-A platform-owner surface owns upgrade/downgrade, billing, overage.
- `storage_used_bytes` aggregation is not wired — `StorageUsedBytes()` returns `-1` until a tenant-scoped blob aggregation lands.
- `api_calls` consumption is not wired — there is no `http.request.*` audit action today, so `CountAuditEventsByActionPrefix(..., "http.request.", ...)` returns 0.

See [`wiki/modules/iam-tech-debt.md`](../../modules/iam-tech-debt.md) for the full backlog.
