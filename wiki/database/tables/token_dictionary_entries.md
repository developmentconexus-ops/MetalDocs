# metaldocs.token_dictionary_entries

**Schema:** `metaldocs.token_dictionary_entries`
**Owner:** tokens module
**Last verified:** 2026-08-07

## Purpose

Stores per-tenant named placeholder tokens (`name` → `value`) used by the
template/document authoring surfaces (§3.2 business/authoring data). Each
entry has a user-facing `label` and optional `description` alongside the
machine `name`/`value` pair. The tokens module's Postgres repository touches
only this table (per the package doc: "The repo touches ONLY
token_dictionary_entries").

## Columns

| Column        | Type          | Notes |
|---------------|---------------|-------|
| `id`          | `uuid` PK     | Default `gen_random_uuid()`. |
| `tenant_id`   | `uuid`        | Not null. No declared FK to `metaldocs.tenants`. |
| `name`        | `text`        | Not null. CHECK matches `^[A-Za-z0-9_]+$`, length 1-64. Unique per tenant (`uq_token_dictionary_tenant_name` on `(tenant_id, name)`). |
| `value`       | `text`        | Not null. CHECK length 1-4096. |
| `label`       | `text`        | Not null. CHECK length 1-256. |
| `description` | `text`        | Nullable. CHECK length ≤ 1024 when present. |
| `created_by`  | `text`        | Not null. |
| `updated_by`  | `text`        | Not null. |
| `created_at`  | `timestamptz` | Default `now()`, not null. |
| `updated_at`  | `timestamptz` | Default `now()`, not null. Bumped explicitly by `Update`. |

## Migrations

Table is present in `db/baseline/0001_current_schema.sql` (folded baseline); no
post-baseline migration alters it as of 2026-08-07.

## Key callers

- `internal/modules/tokens/infrastructure/repository.go::PostgresRepository.Create` — inserts a new entry, in-tx.
- `internal/modules/tokens/infrastructure/repository.go::PostgresRepository.Update` — updates `value`/`label`/`description`/`updated_by`/`updated_at`, scoped by `tenant_id` + `id`.
- `internal/modules/tokens/infrastructure/repository.go` — `Delete`, `Get` (by id), `GetByName`, `List` (all scoped by `tenant_id`, ordered by `name`).
- `internal/modules/tokens/infrastructure/tenant_data_port.go` — tenant export (`ExportTable`) and erase (`EraseTable`) port, both keyed on `tenant_id`.
- `internal/modules/tokens/domain/port.go` — declares the `domain.Repository` interface this adapter implements.

## Tenant scoping

`tenant_id` (uuid, not null) carries tenant scoping; there is no declared FK to
`metaldocs.tenants(id)`. Every repository method (`Get`, `GetByName`, `List`,
`Update`, `Delete`) filters explicitly by `tenant_id`, and the `(tenant_id,
name)` unique index means the same `name` may exist independently per tenant.
A cross-tenant `id`/`name` lookup returns `domain.ErrNotFound` (or an empty
list), never another tenant's entry. RLS `tenant_isolation`
(`FORCE ROW LEVEL SECURITY`) backstops the app-level predicate.
