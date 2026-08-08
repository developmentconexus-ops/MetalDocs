# metaldocs.audit_export_jobs

**Schema:** `metaldocs.audit_export_jobs`
**Owner:** audit module (PR-6)
**Last verified:** 2026-08-07

## Purpose

Tracks async audit-log export jobs (CSV/JSONL) requested by a tenant actor. A
row records the export request (`format`, `filter_json`), its lifecycle
(`status`: `pending | running | ready | failed`), and — per PR-6 — the
rendered blob stored inline in `payload` rather than in object storage. A
`download_token` gates one-time/token-scoped download without requiring the
requester's session (`GetByDownloadToken`).

## Columns

| Column           | Type          | Notes |
|------------------|---------------|-------|
| `id`             | `text` PK     | Export job id, caller-supplied (no upsert path — `Save` errors if reused). |
| `tenant_id`      | `uuid`        | Not null. No declared FK to `metaldocs.tenants`; scoping is enforced by app query predicate + RLS. |
| `actor_id`       | `text`        | Not null. Requesting user. |
| `format`         | `text`        | Not null. CHECK `format IN ('csv','jsonl')`. |
| `filter_json`    | `jsonb`       | Not null. Export filter criteria as submitted. |
| `status`         | `text`        | Default `'pending'`. CHECK `status IN ('pending','running','ready','failed')`. |
| `object_key`     | `text`        | Nullable. Reserved for a future MinIO/S3-backed path (comment in `exports.go`); unused while `payload` is inline. |
| `download_token` | `text`        | Nullable. Matched by `GetByDownloadToken` for token-only lookup. |
| `expires_at`     | `timestamptz` | Nullable. |
| `error_message`  | `text`        | Nullable. Populated on `status = 'failed'`. |
| `estimated_rows` | `bigint`      | Default 0, not null. |
| `actual_rows`    | `bigint`      | Default 0, not null. |
| `payload`        | `bytea`       | Nullable. The rendered export blob (PR-6 inline-storage choice). |
| `created_at`     | `timestamptz` | Default `now()`, not null. |
| `completed_at`   | `timestamptz` | Nullable. |

## Migrations

Table is present in `db/baseline/0001_current_schema.sql` (folded baseline); no
post-baseline migration alters it as of 2026-08-07.

## Key callers

- `internal/modules/audit/infrastructure/postgres/exports.go::ExportJobRepository.Save` — inserts a new job.
- `internal/modules/audit/infrastructure/postgres/exports.go::ExportJobRepository.Get` — reads by `id` + `tenant_id`.
- `internal/modules/audit/infrastructure/postgres/exports.go::ExportJobRepository.GetByDownloadToken` — reads by `id` + `download_token`.
- `internal/modules/audit/infrastructure/postgres/tenant_data_port.go::TenantDataPort.ExportTenantData` — full-row tenant export (F7.3).
- `internal/modules/audit/infrastructure/postgres/tenant_data_port.go::TenantDataPort.EraseTenantData` — full-row tenant erase (F7.3).

## Tenant scoping

`tenant_id` (uuid, not null) carries tenant scoping; there is no declared FK to
`metaldocs.tenants(id)` for this column. `Get` and `GetByDownloadToken` both
filter by `tenant_id`/`id` pair (or token) — a cross-tenant `id` lookup returns
`domain.ErrExportJobNotFound`, never another tenant's row. RLS `tenant_isolation`
policy (`FORCE ROW LEVEL SECURITY`) provides a DB-level backstop on top of the
app-level predicate.
