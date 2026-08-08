# metaldocs.tenant_lifecycle_jobs

**Schema:** `metaldocs.tenant_lifecycle_jobs`
**Owner:** iam module (M7 F7.3 Task E — tenant export/erase lifecycle)
**Last verified:** 2026-08-07

## Purpose

The lifecycle ledger for tenant-wide data export and erase requests. A row is
inserted (`status='pending'`) in the same tx that asserts the
`tenant.export`/`tenant.erase` capability and seeds the target tenant's RLS
GUC; a paired River job is enqueued transactionally alongside it
(transactional-outbox pairing, per `tenant_lifecycle_enqueuer.go`). The row is
never deleted by the tenant erase flow itself — it **is** the audit record
that an erase happened — so `TenantDataPort.EraseTenantData` redacts
`requested_by` to `'erased'` instead of deleting the row (see
`tenant_data_port.go:118-134`).

## Columns

| Column         | Type          | Notes |
|----------------|---------------|-------|
| `id`           | `uuid` PK     | Default `gen_random_uuid()`. |
| `tenant_id`    | `uuid`        | Not null. FK → `metaldocs.tenants(id)`. |
| `kind`         | `text`        | Not null. CHECK `kind IN ('export','erase')`. |
| `status`       | `text`        | Default `'pending'`. CHECK `status IN ('pending','running','ready','failed')`. |
| `requested_by` | `text`        | Not null. Redacted to `'erased'` by `EraseTenantData` rather than deleted. |
| `object_key`   | `text`        | Nullable. Export artifact location; empty for erase jobs (no artifact). |
| `error`        | `text`        | Nullable. Set on `status='failed'`. |
| `created_at`   | `timestamptz` | Default `now()`, not null. |
| `completed_at` | `timestamptz` | Nullable. Set when the job reaches `ready` or `failed`. |

## Migrations

Table is present in `db/baseline/0001_current_schema.sql` (folded baseline); no
post-baseline migration alters it as of 2026-08-07. A `BEFORE INSERT` trigger
(`trg_require_cap_asserted` → `public.enforce_capability_asserted()`) enforces
that the inserting tx has already asserted the required capability.

## Key callers

- `internal/modules/iam/infrastructure/postgres/tenant_lifecycle_repository.go::TenantLifecycleRepository.InsertLifecycleJobTx` — inserts the pending row in the caller's tx.
- `internal/modules/iam/infrastructure/postgres/tenant_lifecycle_repository.go::TenantLifecycleRepository.LoadLifecycleJob` — pool read by `id` (River worker, off any tx).
- `internal/modules/iam/infrastructure/postgres/tenant_lifecycle_repository.go::TenantLifecycleRepository.MarkLifecycleJobRunning` / `MarkLifecycleJobReady` / `MarkLifecycleJobFailed` — status transitions from the run-side worker.
- `internal/modules/iam/application/tenant_lifecycle_service.go` — orchestrates the request path (capability assertion + insert + enqueue, in-tx).
- `internal/modules/iam/jobs/tenant_lifecycle_enqueuer.go` — pairs the row insert with the River job enqueue (transactional-outbox).
- `internal/modules/iam/infrastructure/postgres/tenant_data_port.go::TenantDataPort.EraseTenantData` — redacts `requested_by` instead of deleting (this table is the lifecycle ledger, not exportable/erasable data itself).
- `internal/platform/tripwire/arms.go`, `internal/platform/tripwire/render.go` — the capability-assertion tripwire arm for this table's INSERT.

## Tenant scoping

`tenant_id` (uuid, not null, FK → `metaldocs.tenants(id)`) carries tenant
scoping and RLS (`FORCE ROW LEVEL SECURITY`, `tenant_isolation` policy)
restricts rows to the session's `metaldocs.tenant_id` GUC. `LoadLifecycleJob`
and the `MarkLifecycleJob*` methods query by `id` alone (no explicit
`tenant_id` predicate) because they run off the pool on the River worker's
run-side, before any tenant GUC is seeded — the job id is globally unique and
the row's own `tenant_id` is what seeds the subsequent fan-out tx, per the
file header comment.
