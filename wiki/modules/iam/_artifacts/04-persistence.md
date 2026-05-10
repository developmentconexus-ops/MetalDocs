## 1) Tables owned

| Table | Created in (migration filename) | Notes |
|---|---|---|
| metaldocs.iam_users | migrations/0002_init_iam_rbac.sql | tenant/deactivation columns added in 0130 |

| Column | Type | Constraints (NOT NULL, FK, default) |
|---|---|---|
| user_id | TEXT | PK |
| display_name | TEXT | NOT NULL |
| is_active | BOOLEAN | NOT NULL, DEFAULT TRUE |
| created_at | TIMESTAMPTZ | NOT NULL, DEFAULT NOW() |
| updated_at | TIMESTAMPTZ | NOT NULL, DEFAULT NOW() |
| tenant_id | UUID | NOT NULL, DEFAULT 'ffffffff-ffff-ffff-ffff-ffffffffffff' (added 0130) |
| deactivated_at | TIMESTAMPTZ | nullable; CHECK deactivated_at IS NULL OR deactivated_at >= created_at (0130) |

| Table | Created in (migration filename) | Notes |
|---|---|---|
| metaldocs.iam_user_roles | migrations/0002_init_iam_rbac.sql | tenant_id added in 0162; role constraint rewritten in 0166 |

| Column | Type | Constraints (NOT NULL, FK, default) |
|---|---|---|
| user_id | TEXT | NOT NULL, FK -> metaldocs.iam_users(user_id) ON DELETE CASCADE |
| role_code | TEXT | NOT NULL, CHECK (0003 then 0166 set) |
| assigned_at | TIMESTAMPTZ | NOT NULL, DEFAULT NOW() |
| assigned_by | TEXT | nullable |
| tenant_id | UUID | NOT NULL, DEFAULT 'ffffffff-ffff-ffff-ffff-ffffffffffff' (0162) |

| Table | Created in (migration filename) | Notes |
|---|---|---|
| public.user_process_areas | migrations/0125_registry_iam_user_process_areas_governance_events.sql | hardened in 0136; role allowlist widened in 0158 |

| Column | Type | Constraints (NOT NULL, FK, default) |
|---|---|---|
| user_id | TEXT | NOT NULL |
| tenant_id | UUID | NOT NULL; FK (tenant_id, area_code) -> metaldocs.document_process_areas(tenant_id, code) |
| area_code | TEXT | NOT NULL |
| role | TEXT | NOT NULL; CHECK role IN (...) |
| effective_from | TIMESTAMPTZ | NOT NULL |
| effective_to | TIMESTAMPTZ | nullable |
| granted_by | TEXT | nullable; FK (tenant_id, granted_by) -> metaldocs.iam_users(tenant_id, user_id) (0136) |
| revoked_by | TEXT | nullable; CHECK with effective_to + FK (tenant_id, revoked_by) -> metaldocs.iam_users(tenant_id, user_id) (0136) |

| Table | Created in (migration filename) | Notes |
|---|---|---|
| metaldocs.role_capabilities | migrations/0142a_role_capabilities_v2_additive.sql | reseeded 0165; backfilled 0169 |

| Column | Type | Constraints (NOT NULL, FK, default) |
|---|---|---|
| role | TEXT | NOT NULL |
| capability | TEXT | NOT NULL; CHECK not legacy + format (0142b) |
| description | TEXT | NOT NULL, DEFAULT '' |
| created_at | TIMESTAMPTZ | NOT NULL, DEFAULT now() |

| Table | Created in (migration filename) | Notes |
|---|---|---|
| metaldocs.iam_groups | migrations/0163_iam_groups.sql | |

| Column | Type | Constraints (NOT NULL, FK, default) |
|---|---|---|
| id | UUID | PK |
| tenant_id | UUID | NOT NULL, DEFAULT 'ffffffff-ffff-ffff-ffff-ffffffffffff' |
| name | TEXT | NOT NULL; UNIQUE (tenant_id, name) |
| description | TEXT | NOT NULL, DEFAULT '' |
| created_at | TIMESTAMPTZ | NOT NULL, DEFAULT now() |

| Table | Created in (migration filename) | Notes |
|---|---|---|
| metaldocs.iam_group_members | migrations/0163_iam_groups.sql | |

| Column | Type | Constraints (NOT NULL, FK, default) |
|---|---|---|
| group_id | UUID | NOT NULL, FK -> metaldocs.iam_groups(id) ON DELETE CASCADE |
| user_id | TEXT | NOT NULL |
| tenant_id | UUID | NOT NULL |
| granted_at | TIMESTAMPTZ | NOT NULL, DEFAULT now() |
| granted_by | TEXT | nullable |

| Table | Created in (migration filename) | Notes |
|---|---|---|
| metaldocs.iam_group_roles | migrations/0163_iam_groups.sql | |

| Column | Type | Constraints (NOT NULL, FK, default) |
|---|---|---|
| group_id | UUID | NOT NULL, FK -> metaldocs.iam_groups(id) ON DELETE CASCADE |
| role | TEXT | NOT NULL |

| Table | Created in (migration filename) | Notes |
|---|---|---|
| metaldocs.iam_membership_audit | (unclear: no CREATE TABLE match found by grep in migrations/) | referenced in module guidance only |

## 2) Tables read or written but NOT owned

| Table | Owner module | Read / Write | Operations using it |
|---|---|---|---|
| governance_events (public/metaldocs) | (unclear: shared; created in 0125 with registry+IAM naming) | Read + Write | `internal/modules/iam/application/startup.go` reads MAX version and inserts version bump event |
| metaldocs.document_process_areas | taxonomy (by name) | Read (FK enforcement path) | FK target for `user_process_areas` in 0125 |
| public.approval_instances | documents/approval (by name) | Trigger-enforced write surface (tripwire target) | tripwire trigger attached in 0142b |
| public.approval_signoffs | documents/approval (by name) | Trigger-enforced write surface (tripwire target) | tripwire trigger attached in 0142b |

## 3) Triggers, GUCs, functions

| Object | Kind (trigger / function / GUC) | File:line | Purpose |
|---|---|---|---|
| public.enforce_capability_asserted() | function | migrations/0142b_role_capabilities_v2_enforce.sql:67-179 | tripwire function that checks asserted capability before INSERT on guarded tables |
| trg_require_cap_asserted_instances on public.approval_instances | trigger | migrations/0142b_role_capabilities_v2_enforce.sql:200-203 | runs tripwire before INSERT on `approval_instances` |
| trg_require_cap_asserted_signoffs on public.approval_signoffs | trigger | migrations/0142b_role_capabilities_v2_enforce.sql:206-209 | runs tripwire before INSERT on `approval_signoffs` |
| metaldocs.asserted_caps | GUC read | migrations/0142b_role_capabilities_v2_enforce.sql:139 | read by tripwire for asserted capabilities JSON array |
| metaldocs.actor_id | GUC read | migrations/0137_db_roles_security_definer.sql:68,111 | read by legacy/public membership SECURITY DEFINER functions |
| metaldocs.tenant_id | GUC read | internal/modules/iam/authz/context.go:36 | read by `MustTenantID`; no migration `current_setting('metaldocs.tenant_id', true)` match found |

Tripwire attachment check for requested tables:
- `approval_instances`: attached (0142b:200-203)
- `approval_signoffs`: attached (0142b:206-209)
- `iam_user_roles`: no `enforce_capability_asserted` trigger attachment match
- `user_process_areas`: no `enforce_capability_asserted` trigger attachment match

## 4) Indexes

| Index | Table | Columns | Unique? | Purpose |
|---|---|---|---|---|
| idx_iam_user_roles_role_code | metaldocs.iam_user_roles | role_code | NO | role lookup |
| ux_iam_users_tenant_user | metaldocs.iam_users | tenant_id, user_id | YES | tenant+user uniqueness |
| ux_iam_users_tenant_user_active | metaldocs.iam_users | tenant_id, user_id WHERE deactivated_at IS NULL | YES | active-user uniqueness |
| ix_user_process_areas_active | public.user_process_areas | user_id, area_code WHERE effective_to IS NULL | NO | active membership lookup |
| ux_user_process_areas_one_active | public.user_process_areas | user_id, tenant_id, area_code WHERE effective_to IS NULL | YES | one active row per user+tenant+area |
| ux_user_process_areas_single_active | public.user_process_areas | tenant_id, user_id, area_code, role WHERE effective_to IS NULL | YES | one active row per user+tenant+area+role |
| metaldocs.role_capabilities_pkey | metaldocs.role_capabilities | role, capability | YES | PK |
| metaldocs.iam_groups_pkey | metaldocs.iam_groups | id | YES | PK |
| metaldocs.iam_groups_tenant_id_name_key | metaldocs.iam_groups | tenant_id, name | YES | unique group name per tenant |
| metaldocs.iam_group_members_pkey | metaldocs.iam_group_members | group_id, user_id | YES | PK |
| metaldocs.iam_group_roles_pkey | metaldocs.iam_group_roles | group_id, role | YES | PK |

## 5) Tripwire pairing audit

| Method (file:line) | Authz.Require called? | Cap + area arg | SQL verb | Table |
|---|---|---|---|---|
| RoleAdminRepository.UpsertUserAndAssignRole (internal/modules/iam/infrastructure/postgres/role_admin_repository.go:33) | NO | N/A | INSERT | metaldocs.iam_users |
| RoleAdminRepository.UpsertUserAndAssignRole (internal/modules/iam/infrastructure/postgres/role_admin_repository.go:33) | NO | N/A | DELETE | metaldocs.iam_user_roles |
| RoleAdminRepository.UpsertUserAndAssignRole (internal/modules/iam/infrastructure/postgres/role_admin_repository.go:33) | NO | N/A | INSERT | metaldocs.iam_user_roles |
| RoleAdminRepository.UpsertUserAndAssignRole (internal/modules/iam/infrastructure/postgres/role_admin_repository.go:33) | tripwire pairing: N/A (tier-1 only, no trigger on table) | N/A | INSERT/DELETE/INSERT | metaldocs.iam_users, metaldocs.iam_user_roles |
| RoleAdminRepository.ReplaceUserRoles (internal/modules/iam/infrastructure/postgres/role_admin_repository.go:72) | NO | N/A | INSERT | metaldocs.iam_users |
| RoleAdminRepository.ReplaceUserRoles (internal/modules/iam/infrastructure/postgres/role_admin_repository.go:72) | NO | N/A | DELETE | metaldocs.iam_user_roles |
| RoleAdminRepository.ReplaceUserRoles (internal/modules/iam/infrastructure/postgres/role_admin_repository.go:72) | NO | N/A | INSERT | metaldocs.iam_user_roles |
| RoleAdminRepository.ReplaceUserRoles (internal/modules/iam/infrastructure/postgres/role_admin_repository.go:72) | tripwire pairing: N/A (tier-1 only, no trigger on table) | N/A | INSERT/DELETE/INSERT | metaldocs.iam_users, metaldocs.iam_user_roles |
| UserAreaRepository.Insert (internal/modules/iam/infrastructure/postgres/user_area_repository.go:51) | NO | N/A | INSERT | public.user_process_areas |
| UserAreaRepository.Insert (internal/modules/iam/infrastructure/postgres/user_area_repository.go:51) | tripwire pairing: N/A (tier-1 only, no trigger on table) | N/A | INSERT | public.user_process_areas |
| UserAreaRepository.CloseActive (internal/modules/iam/infrastructure/postgres/user_area_repository.go:75) | NO | N/A | UPDATE | public.user_process_areas |
| UserAreaRepository.CloseActive (internal/modules/iam/infrastructure/postgres/user_area_repository.go:75) | tripwire pairing: N/A (tier-1 only, no trigger on table) | N/A | UPDATE | public.user_process_areas |
| UserAreaRepository.GrantAtomic (internal/modules/iam/infrastructure/postgres/user_area_repository.go:90) | NO | N/A | UPDATE | public.user_process_areas |
| UserAreaRepository.GrantAtomic (internal/modules/iam/infrastructure/postgres/user_area_repository.go:90) | NO | N/A | INSERT | public.user_process_areas |
| UserAreaRepository.GrantAtomic (internal/modules/iam/infrastructure/postgres/user_area_repository.go:90) | tripwire pairing: N/A (tier-1 only, no trigger on table) | N/A | UPDATE+INSERT | public.user_process_areas |
| area_membership.Grant (internal/modules/iam/area_membership/area_membership.go:53) | NO | N/A | SELECT (calls metaldocs.grant_area_membership) | function call path mutates public.user_process_areas + metaldocs.governance_events |
| area_membership.Grant (internal/modules/iam/area_membership/area_membership.go:53) | tripwire pairing: N/A (tier-1 only, no trigger on table) | N/A | function-mediated write | public.user_process_areas |
| area_membership.Revoke (internal/modules/iam/area_membership/area_membership.go:65) | NO | N/A | SELECT (calls metaldocs.revoke_area_membership) | function call path mutates public.user_process_areas + metaldocs.governance_events |
| area_membership.Revoke (internal/modules/iam/area_membership/area_membership.go:65) | tripwire pairing: N/A (tier-1 only, no trigger on table) | N/A | function-mediated write | public.user_process_areas |

Notes for listed non-mutators:
- RoleAdminRepository.HasAnyRole (read-only)
- RoleProvider.RolesByUserID (read-only)
- UserAreaRepository.ListActive (read-only)
- UserAreaRepository.GetActiveByUserAndArea (read-only)
- area_membership.List (read-only)

Tripwire `VIOLATION` rows found: none.

## 6) Migration history

| Order | Filename | Verb summary | Date (from filename or commit) |
|---|---|---|---|
| 1 | 0002_init_iam_rbac.sql | CREATE TABLE iam_users, iam_user_roles; CREATE INDEX iam_user_roles(role_code) | 0002 |
| 2 | 0003_iam_role_code_constraint.sql | ALTER TABLE iam_user_roles; ADD role_code CHECK | 0003 |
| 3 | 0125_registry_iam_user_process_areas_governance_events.sql | CREATE TABLE user_process_areas + indexes; CREATE TABLE governance_events + indexes | 0125 |
| 4 | 0130_iam_users_tenant_deactivated.sql | ALTER TABLE iam_users add tenant_id/deactivated_at; add checks/indexes | 0130 |
| 5 | 0136_user_process_areas_hardening.sql | ALTER TABLE user_process_areas; add constraints/index; create no-delete/update triggers | 0136 |
| 6 | 0142a_role_capabilities_v2_additive.sql | CREATE TABLE role_capabilities; seed capability rows | 0142a |
| 7 | 0142b_role_capabilities_v2_enforce.sql | DELETE legacy caps; add checks; CREATE tripwire function + triggers | 0142b |
| 8 | 0158_fix_process_area_role_constraint.sql | ALTER TABLE user_process_areas role CHECK; UPDATE/INSERT user_process_areas; INSERT role_capabilities | 0158 |
| 9 | 0162_iam_user_roles_tenant_id.sql | ALTER TABLE iam_user_roles add tenant_id | 0162 |
| 10 | 0163_iam_groups.sql | CREATE TABLE iam_groups, iam_group_members, iam_group_roles | 0163 |
| 11 | 0164_documents_v2_visibility.sql | ALTER TABLE documents_v2 add visibility | 0164 |
| 12 | 0165_role_capabilities_reseed.sql | TRUNCATE role_capabilities; INSERT reseed rows | 0165 |
| 13 | 0166_role_rename_reviewer_migration.sql | UPDATE/DELETE iam_user_roles; ALTER constraints | 0166 |
| 14 | 0167_documents_bridge_and_state_columns.sql | no IAM table/capability/user_process_areas match by grep | 0167 |
| 15 | 0168_drop_documents_v2_orphan.sql | no IAM table/capability/user_process_areas match by grep | 0168 |
| 16 | 0169_role_capabilities_process_areas.sql | INSERT role_capabilities backfill for signer/area_admin/qms_admin | 0169 |
| 17 | 0170_dev_approver_role_correction.sql | UPDATE iam_user_roles for dev approver correction | 0170 |