# public.approval_delegations

**Schema:** `public.approval_delegations`
**Owner:** approval module (F9, ADR 0077)
**Last verified:** 2026-08-07

## Purpose

Records an active grant letting `delegate_id` act on `delegator_id`'s behalf
for approval review during `[starts_at, ends_at)`. Overlapping active grants
for the same delegator are explicitly allowed — the domain uses union
semantics, not last-writer-wins (spec.md Interview #7). This table is also
the tripwire-last-line backstop for the delegation-window and no-self-delegate
invariants (also enforced by the `approval_delegations_window_chk` and
`approval_delegations_no_self` CHECK constraints at the DB layer).

## Columns

| Column         | Type          | Notes |
|----------------|---------------|-------|
| `id`           | `uuid` PK     | Default `gen_random_uuid()`. |
| `tenant_id`    | `uuid`        | Not null. No declared FK to `metaldocs.tenants` (confirmed independent — see `tenant_data_port.go`). |
| `delegator_id` | `text`        | Not null. The user granting the delegation. |
| `delegate_id`  | `text`        | Not null. The user receiving it. CHECK `delegator_id <> delegate_id` (no self-delegation). |
| `starts_at`    | `timestamptz` | Not null. |
| `ends_at`      | `timestamptz` | Not null. CHECK `ends_at > starts_at`. |
| `reason`       | `text`        | Not null. |
| `created_by`   | `text`        | Not null. |
| `created_at`   | `timestamptz` | Default `now()`, not null. |

## Migrations

Table is present in `db/baseline/0001_current_schema.sql` (folded baseline); no
post-baseline migration alters it as of 2026-08-07.

## Key callers

- `internal/modules/approval/infrastructure/postgres_approval_repository.go::postgresApprovalRepository.InsertDelegation` — plain insert, in-tx.
- `internal/modules/approval/infrastructure/postgres_approval_repository.go::postgresApprovalRepository.DeleteDelegation` — atomic ownership-checked delete (WHERE-clause OCC idiom: `delegator_id = caller OR callerIsOversee`).
- `internal/modules/approval/infrastructure/postgres_approval_repository.go::postgresApprovalRepository.LoadActiveDelegationsFor` — in-tx read of active grants covering `asOf`, always at actual use-time, never cached.
- `internal/modules/approval/application/delegation_service.go` — application service orchestrating grant/revoke.
- `internal/modules/approval/domain/sod.go` — separation-of-duties logic consults active delegations for the same actor-substitution question.
- `internal/modules/approval/infrastructure/tenant_data_port.go::TenantDataPort` — tenant export/erase port for this table.

## Tenant scoping

`tenant_id` (uuid, not null) carries tenant scoping with no declared FK to
`metaldocs.tenants(id)` — it is an independent tenant-tagged table (per the
tenant-data-port comment: "no FK, independent"). All reads/writes filter
explicitly by `tenant_id` (`LoadActiveDelegationsFor`, `DeleteDelegation`), and
RLS `tenant_isolation` (`FORCE ROW LEVEL SECURITY`) backstops the app-level
predicate — a cross-tenant lookup returns no matching row.
