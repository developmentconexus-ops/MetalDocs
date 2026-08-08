# metaldocs.materialize_dispatch_outbox

**Schema:** `metaldocs.materialize_dispatch_outbox`
**Owner:** render/fanout (materialize dispatch path)
**Last verified:** 2026-08-07

## Purpose

Transactional-outbox staging table for the docx materialize dispatch path. A
row is inserted in the same business transaction as the freeze/approval-decision
write that produces a new revision's resolved values (`freeze_service.go`,
`decision_service.go`); a paired River `MaterializeDispatchArgs` job is only
enqueued when the insert actually lands a new row (the `(tenant_id,
revision_id, COALESCE(release_generation_id, <nil-uuid>))` unique index dedups
repeat enqueues via `ON CONFLICT DO NOTHING`, migration 0310). The
`values_hash` column carries the resolved-placeholder values hash pinned at
freeze time (renamed from the former `content_hash` misnomer in migration
0312, F-QA4-10) — it is not comparable to `pdf_dispatch_outbox.frozen_docx_hash`.

## Columns

| Column                   | Type          | Notes |
|--------------------------|---------------|-------|
| `id`                     | `uuid` PK     | Default `gen_random_uuid()`. |
| `tenant_id`               | `uuid`        | Not null. No declared FK to `metaldocs.tenants`. |
| `revision_id`             | `uuid`        | Not null. |
| `values_hash`             | `bytea`       | Not null. Resolved-placeholder values hash (renamed from `content_hash`, migration 0312). |
| `status`                  | `text`        | Default `'pending'`. CHECK `status IN ('pending','processing','dispatched','failed')`. |
| `attempts`                | `integer`     | Default 0, not null. |
| `last_error`              | `text`        | Nullable. |
| `claimed_at`              | `timestamptz` | Nullable. |
| `next_retry_at`           | `timestamptz` | Default `now()`, not null. |
| `created_at`              | `timestamptz` | Default `now()`, not null. |
| `dispatched_at`           | `timestamptz` | Nullable. |
| `dead_lettered_at`        | `timestamptz` | Nullable. |
| `release_generation_id`   | `uuid`        | Nullable. Part of the dedup unique index, COALESCEd to a nil-uuid sentinel so legacy generation-less rows keep revision-only dedup (migration 0310). |

## Migrations

| Version | Description |
|---------|-------------|
| `0310`  | Added generation-aware dedup unique index `ux_materialize_dispatch_outbox_generation` on `(tenant_id, revision_id, COALESCE(release_generation_id, '00000000-0000-0000-0000-000000000000'::uuid))`. |
| `0312`  | Renamed `content_hash` to `values_hash` (F-QA4-10 misnomer fix). |

## Key callers

- `internal/modules/render/fanout/dispatchjobs/enqueuer.go::Enqueuer.EnqueueMaterializeTx` — inserts + conditionally enqueues the paired River job, in-tx.
- `internal/modules/render/fanout/staging_outbox.go::StagingOutboxRepository.Enqueue` — generic outbox insert (allowlisted to this table and `pdf_dispatch_outbox`).
- `internal/modules/render/fanout/dispatchjobs/workers.go::MaterializeDispatchWorker.Work` — River worker consumer for the paired job (row lifecycle to `dispatched`/`failed` happens via the repo, not shown by table name directly in this file).
- `internal/modules/documents/application/freeze_service.go` — writer at freeze time (in-tx with the revision write).
- `internal/modules/approval/application/decision_service.go` — writer at approval-decision time (in-tx with the decision write).
- `internal/modules/render/fanout/tenant_data_port.go` — tenant export/erase port.

## Tenant scoping

`tenant_id` (uuid, not null) carries tenant scoping and is part of the dedup
unique index together with `revision_id` and `release_generation_id`. There is
no declared FK to `metaldocs.tenants(id)`. RLS `tenant_isolation`
(`FORCE ROW LEVEL SECURITY`) restricts row visibility to the session's
`metaldocs.tenant_id` GUC; a cross-tenant read returns no rows rather than
another tenant's outbox entry.
