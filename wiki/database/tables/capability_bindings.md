# metaldocs.capability_bindings

> **Added:** 2026-08-11, migration `0318` (issue #89/A8.1, ADR 0092 D1 — grant-model unification)
> **Source:** `db/migrations/0318_capability_bindings_schema_backfill.sql`
> **Schema:** `metaldocs`
> **Owner:** iam

## Purpose

The single grant relation ADR 0092 D1 collapses `iam_user_roles`, `user_process_areas`,
and `iam_group_roles` into: `(subject_kind, subject_id, role_code, scope_kind,
scope_ref)` plus effective-interval and grant/revoke provenance. Model precedent:
Kubernetes RBAC — scope lives on the binding, not in a parallel table.

**As of A8.1, nothing reads this table.** It exists (schema, constraints, RLS) and is
backfilled with full history from the three source relations, but tier-1
(`CapabilityService.CanDo`), tier-2 (`authz.Require`), and every other consumer keep
reading `iam_user_roles`/`user_process_areas`/`iam_group_roles` unchanged — those stay
the live path until A8.2 (query builder: `Granted`/`GrantedAnyScope`) lands. Do not wire
a reader against this table without first reading ADR 0092 and confirming A8.2 has
shipped.

**NOT AUTHORITATIVE (TRANSITIONAL — write cutover not landed):** the corollary of "nothing
reads this table" is that nothing writes it either, past the one-time backfill.
`role_admin_repository.go`, `user_area_repository.go`, and `onboard_tenant_service.go` —
every current grant write site — still write exclusively to `iam_user_roles`,
`user_process_areas`, and `iam_group_roles`; none dual-writes into
`capability_bindings`. From the moment 0318 merges, every grant issued afterwards is
invisible here and this table drifts stale by construction. Harmless today because no
read path consults it (above); becomes a live correctness hole the instant one does. Do
not treat row counts or contents here as ground truth, and do not add a reader, until a
write cutover (dual-write or repoint of the three sites above) lands — that cutover is
not yet owned by any slice in the canonical A8.1–A8.4 decomposition and is tracked
separately. Until then `iam_user_roles`, `user_process_areas`, and `iam_group_roles`
remain the sole grant source of record.

`iam_group_members` (group *membership* — who is in a group) is explicitly **not** one
of the source relations folded in here; ADR 0092 D4 keeps membership and grants
orthogonal.

## Named deviation from ADR 0092's literal column list

The ADR names a single logical `subject_id`. Postgres cannot express one FK column
conditionally pointing at two different tables/types (`iam_users.user_id` is `text`;
`iam_groups.id` is `uuid`). `subject_id` is split into physical `subject_user_id` (FK →
`iam_users`) + `subject_group_id` (FK → `iam_groups`), discriminated by `subject_kind`.
This is a stronger reading of the ADR's own requirement — referentially valid,
unrepresentable-by-construction — than a single polymorphic column could deliver: a
dangling subject is a foreign-key violation, not an app-level check.

## Columns

| Column | Type | Nullable | Meaning |
|---|---|---|---|
| `id` | `uuid` | no | Surrogate PK, `gen_random_uuid()`. |
| `tenant_id` | `uuid` | no | FK → `metaldocs.tenants(id)`. Default dev sentinel `ffffffff-...`. |
| `subject_kind` | `text` | no | `'user'` or `'group'`. |
| `subject_user_id` | `text` | yes | Set iff `subject_kind = 'user'`. FK `(tenant_id, subject_user_id)` → `iam_users(tenant_id, user_id)`. |
| `subject_group_id` | `uuid` | yes | Set iff `subject_kind = 'group'`. FK `(tenant_id, subject_group_id)` → `iam_groups(tenant_id, id)`. |
| `role_code` | `text` | no | FK → `metaldocs.roles(code)`. |
| `scope_kind` | `text` | no | `'tenant'` or `'area'`. |
| `scope_ref` | `text` | yes | Set iff `scope_kind = 'area'`. FK `(tenant_id, scope_ref)` → `document_process_areas(tenant_id, code)`. |
| `effective_from` | `timestamptz` | no | Default `now()`. |
| `effective_to` | `timestamptz` | yes | NULL = active (mirrors `user_process_areas`' ADR 0037 temporal model). |
| `granted_by` | `text` | yes | Actor who created the binding. |
| `revoked_by` | `text` | yes | Required whenever `effective_to` is set (`chk_capability_bindings_revoked_by_required`). |
| `source_relation` | `text` | yes | `NULL` = native grant (A8.2+); else one of `'iam_user_roles'`/`'user_process_areas'`/`'iam_group_roles'` — backfill provenance. |

## Constraints (unrepresentable-by-construction, not app-validated)

- `chk_capability_bindings_subject_kind` / `chk_capability_bindings_subject_shape` — exactly
  one physical subject column set, matching `subject_kind`.
- `chk_capability_bindings_scope_kind` / `chk_capability_bindings_scope_shape` — `scope_ref`
  NULL iff `scope_kind = 'tenant'`.
- `chk_capability_bindings_effective_interval`, `chk_capability_bindings_revoked_by_required`
  — byte-identical to `user_process_areas`' ADR 0037 CHECKs.
- `chk_capability_bindings_source_relation` — provenance tag is `NULL` or one of the three
  source relation names.
- `fk_capability_bindings_tenant`, `fk_capability_bindings_role`,
  `fk_capability_bindings_subject_user`, `fk_capability_bindings_subject_group`,
  `fk_capability_bindings_scope_area` — all FKs use Postgres's default `MATCH SIMPLE`, so a
  NULL discriminated-union column correctly skips that FK rather than needing app-level
  branching.
- `ux_capability_bindings_active_identity` (UNIQUE, partial `WHERE effective_to IS NULL`) —
  at most one active binding per `(tenant, subject, role, scope)`. Uses `COALESCE(...,
  '')` on the nullable discriminated-union columns because Postgres unique indexes treat
  NULL as distinct-from-itself, which would otherwise defeat dedup.

## Prerequisite schema changes (same migration)

- `iam_users_tenant_user_uk` — promotes the pre-existing unique **index**
  `ux_iam_users_tenant_user` to a unique **constraint** via `UNIQUE USING INDEX` (a bare
  index cannot be an FK target).
- `iam_groups_tenant_id_id_uk` — new `UNIQUE (tenant_id, id)` on `iam_groups` (no prior
  uniqueness on that pair existed; cheap since `id` alone is already PK-unique).

## RLS

`ENABLE ROW LEVEL SECURITY` + `FORCE ROW LEVEL SECURITY` + a `tenant_isolation` policy
identical to every sibling grant table (`iam_user_roles`, `user_process_areas`,
`iam_groups`, `iam_users`): null-GUC admits all rows (migration/bootstrap escape hatch,
not a leak), else `tenant_id` must match `metaldocs.tenant_id`.

## DB tripwire

Migration `0319` attaches `trg_require_cap_asserted` (BEFORE INSERT OR UPDATE OR DELETE,
`enforce_capability_asserted()`), gated on `user.manage` OR `membership.manage`
(match-one) — arm #21 in `internal/platform/tripwire/arms.go`. The trigger is attached
**after** migration 0318's backfill, since no writer other than that backfill exists at
0318 time and nothing needs asserting yet; every write from this migration forward is
gated.

## Backfill (migration 0318)

Three `INSERT ... SELECT` passes, each preserving full history (not just active rows)
and stamping `source_relation`:

1. `iam_user_roles` → `subject_kind='user'`, `scope_kind='tenant'`.
2. `user_process_areas` → `subject_kind='user'`, `scope_kind='area'`. **All** rows,
   active and already-revoked.
3. `iam_group_roles` (joined to `iam_groups` for `tenant_id`) → `subject_kind='group'`,
   `scope_kind='tenant'`. This source has no timestamp/provenance columns of its own;
   `effective_from` anchors on the owning group's `created_at` and `granted_by` is
   honestly `NULL` rather than fabricated. A pre-check inside the migration raises
   `P0001` if any `iam_group_roles.role` value is absent from `metaldocs.roles` (that
   table's `role` column carries no upstream CHECK, unlike the other two sources).

**Reversibility** (no down-migration tooling exists in this repo — see
`internal/platform/migrate/migrate.go`): valid only while `iam_user_roles` /
`user_process_areas` / `iam_group_roles` remain untouched (true for as long as A8.1 is
the latest landed slice). Both forms below were executed against a throwaway database
on 2026-08-11, not just asserted — see the migration file's REVERSIBILITY header for the
same text.

```sql
-- partial: undo only the backfilled rows. This DELETE is itself tripwire-guarded
-- (0319 arm #21 applies to every write, not just INSERT), so it needs the same
-- capability GUC an application write would carry.
BEGIN;
SELECT set_config('metaldocs.asserted_caps', '[{"cap":"user.manage"}]', true);
DELETE FROM metaldocs.capability_bindings
  WHERE source_relation IN ('iam_user_roles', 'user_process_areas', 'iam_group_roles');
COMMIT;

-- full: also drops the two new tables (their trigger from 0319 goes with them,
-- no separate step needed) and the iam_groups promotion.
DROP TABLE metaldocs.capability_bindings;
DROP TABLE metaldocs.roles;
ALTER TABLE metaldocs.iam_groups DROP CONSTRAINT iam_groups_tenant_id_id_uk;
```

**Named limit, proven not assumed:** `ALTER TABLE metaldocs.iam_users DROP CONSTRAINT
iam_users_tenant_user_uk` fails without `CASCADE`. Four FK constraints that *predate*
this migration — `approval_instances_submitted_by_tenant_fkey`,
`approval_signoffs_actor_tenant_fkey`, `user_process_areas_granted_by_same_tenant`,
`user_process_areas_revoked_by_same_tenant` — already referenced
`iam_users(tenant_id, user_id)` and already depended on the bare index
`ux_iam_users_tenant_user` before A8.1 ever ran; `ADD CONSTRAINT ... UNIQUE USING INDEX`
renamed that same physical index in place rather than creating a new dependency. Dropping
the promoted constraint would CASCADE into four unrelated modules' FKs — out of scope to
fix here and not a regression this migration introduced. Treat the `iam_users` promotion
as effectively permanent; only the `iam_groups` promotion and the two new tables are
freely reversible.

## Notes and Debt

- A8.2 (query builder) is the next slice: `Granted`/`GrantedAnyScope` predicates over
  this table, still with no tier repointed yet.
- A8.3 repoints `metaldocs.roles` at generated catalogs; see that table's page.
- A8.4 removes the `system_admin` tier bypass (ADR 0092 D2) — out of scope here.
