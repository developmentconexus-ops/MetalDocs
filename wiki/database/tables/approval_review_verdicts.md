# public.approval_review_verdicts

**Schema:** `public.approval_review_verdicts`
**Owner:** approval module
**Last verified:** 2026-08-07

## Purpose

Records one reviewer verdict (`ready` or `request_changes`) per actor per
approval stage instance — the eQMS review/expiry surface's durable record of
who reviewed what and what they decided, including an on-behalf-of delegate
trail and a point-in-time display-name snapshot (`actor_display_name_snapshot`,
migration 0294 / ADR 0079) so historical verdicts remain legible even if the
actor's profile later changes. `InsertVerdict` uses `ON CONFLICT
(stage_instance_id, actor_user_id) DO NOTHING` plus a same-value replay check,
giving idempotent-insert semantics without a dedicated idempotency-key table.
A `BEFORE INSERT` trigger (`trg_review_verdict_sod` →
`public.enforce_approval_sod()`) enforces separation-of-duties at insert time.

## Columns

| Column                         | Type          | Notes |
|--------------------------------|---------------|-------|
| `id`                           | `uuid` PK     | Default `gen_random_uuid()`. |
| `approval_instance_id`         | `uuid`        | Not null. |
| `stage_instance_id`            | `uuid`        | Not null. Part of the `(stage_instance_id, actor_user_id)` unique constraint (`approval_review_verdicts_stage_actor_uq`) — one verdict per actor per stage. |
| `actor_user_id`                | `text`        | Not null. |
| `actor_tenant_id`              | `uuid`        | Not null. Carries tenant scoping (see below) — no declared FK. |
| `verdict`                      | `text`        | Not null. CHECK `verdict IN ('ready','request_changes')`. |
| `comment`                      | `text`        | Nullable. |
| `verdict_at`                   | `timestamptz` | Default `now()`, not null. |
| `actor_display_name_snapshot`  | `text`        | Not null. CHECK `<> ''` (migration 0294 / ADR 0079) — bound unconditionally by `InsertVerdict`, fails closed rather than writing NULL. |
| `on_behalf_of_user_id`         | `text`        | Nullable. Set when the verdict was recorded by a delegate. |

## Migrations

| Version | Description |
|---------|-------------|
| `0294`  | ADR 0079: added NOT NULL + non-empty CHECK on `actor_display_name_snapshot` (eQMS audit-truth guarantee). |

## Key callers

- `internal/modules/approval/infrastructure/postgres_approval_repository.go::postgresApprovalRepository.InsertVerdict` — idempotent insert with same-actor-same-stage conflict → replay check, in-tx.
- `internal/modules/approval/infrastructure/postgres_approval_repository.go::postgresApprovalRepository.loadVerdictByStageActor` — reads the existing verdict for the replay comparison, joined against `approval_instances` for the tenant predicate.
- `internal/modules/approval/application/review_verdict_service.go` — application service that carries the persisted verdict id out of the tx (idempotency envelope).
- `internal/modules/approval/application/signoff_idemp.go` — signoff idempotency path replays through the same verdict id.
- `internal/modules/approval/infrastructure/tenant_data_port.go::TenantDataPort` — tenant export/erase port, keyed on `actor_tenant_id`.

## Tenant scoping

Tenant scoping is carried by `actor_tenant_id` (uuid, not null) — **not**
`tenant_id`; there is no declared FK to `metaldocs.tenants(id)`. RLS
`tenant_isolation` filters on `actor_tenant_id = current_setting('metaldocs.tenant_id')`.
`loadVerdictByStageActor` additionally joins `approval_instances i ON i.id =
v.approval_instance_id` and predicates `i.tenant_id = $3::uuid`, so a
cross-tenant lookup (mismatched `actor_tenant_id` or joined instance tenant)
returns no row, never another tenant's verdict.
