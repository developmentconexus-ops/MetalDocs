# metaldocs.roles

> **Added:** 2026-08-11, migration `0318` (issue #89/A8.1, ADR 0092 D1 — grant-model unification)
> **Source:** `db/migrations/0318_capability_bindings_schema_backfill.sql`
> **Schema:** `metaldocs`
> **Owner:** iam

## Purpose

FK target for `capability_bindings.role_code` — "one role catalog referenced by FK
(adding a role becomes data, not a migration)" per ADR 0092 D1.

## TRANSITIONAL — labelled per CLAUDE.md's Global Maximum rule

This table's seed is **hand-maintained**, not generated. ADR 0092's Context names **six
role-declaration surfaces** already drifting from each other (`iamtypes.validRoles`,
`iamtypes.areaRoles`, OpenAPI `UserRole`, OpenAPI `AreaRole`, the `user_process_areas`
CHECK of 7, the `iam_user_roles` CHECK of 5) plus `iam_group_roles` with no CHECK at
all. Hand-seeding a *seventh* surface here does not fix that — it is a deliberate,
labelled local maximum: this table exists so `capability_bindings` has an FK target at
all in A8.1.

**Global-maximum structure:** a single Go-registry-driven role catalog, generated the
same way `TripwireArms` generates the tripwire (see
`internal/platform/tripwire/arms.go`'s header for the pattern this should follow).
**Milestone that deletes this local maximum: A8.3** ("generated role catalogs +
tripwire repoint", issue #89/A8). Until then, this seed is the union of every
role-declaration surface known at A8.1 time and must be kept a superset of all six (plus
`iam_group_roles`) by hand.

## Columns

| Column | Type | Nullable | Meaning |
|---|---|---|---|
| `code` | `text` | no | PK. `CHECK (code ~ '^[a-z][a-z0-9_]*$')`. |
| `description` | `text` | no | Default `''`. |
| `created_at` | `timestamptz` | no | Default `now()`. |

## Seed (migration 0318)

`system_admin`, `approver`, `author`, `editor`, `viewer`, `signer`, `area_admin`,
`qms_admin` — the same 8 codes as `user_process_areas`' CHECK (7 area roles) plus
`system_admin` (tenant-wide, never an area membership).

## RLS

None — global catalog, no `tenant_id` column. Mirrors `role_capabilities` (also a global
catalog with no RLS) and the `iam_group_roles` tenant-id-less precedent.

## DB tripwire

Migration `0319` attaches `trg_require_cap_asserted` (BEFORE INSERT OR UPDATE OR
DELETE), gated on `user.manage` — arm #22 in `internal/platform/tripwire/arms.go`. The
trigger sets `v_tenant_id := NULL` (no tenant column to derive it from), the same
pattern as `iam_group_roles`.

## Notes and Debt

Do not add a role code here without confirming it is reflected in every surface ADR 0092
names — until A8.3 repoints this at codegen, that check is manual.
