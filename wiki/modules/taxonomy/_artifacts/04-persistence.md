# Phase 4 — Persistence map (taxonomy)

3 tables owned · 6 mutating repo methods with no `authz.Require` (all VIOLATIONs vs tripwire pattern) · 19 migrations.

## 1. Tables owned

| Table | Created | Notes |
|---|---|---|
| `metaldocs.document_families` | `0023_init_document_family_and_profile_registry.sql:1-7` | Global (no tenant_id). Write GRANT added `0161_grant_families_write_privileges.sql:1`. |
| `metaldocs.document_profiles` | `0023:9-17` | Extended by `0035` (alias) and `0122` (tenant_id, default_template_version_id, archived_at, code CHECK, immutability trigger). |
| `metaldocs.document_process_areas` | `0025_init_document_taxonomy.sql:1-7` | Extended by `0123` (tenant_id, parent_code self-FK, archived_at, code CHECK, immutability trigger). |

### `document_families` columns

| Column | Type | Constraints |
|---|---|---|
| code | TEXT | PK (`0023:2`) |
| name | TEXT | NOT NULL (`0023:3`) |
| description | TEXT | NOT NULL DEFAULT '' (`0023:4`) |
| is_active | BOOLEAN | NOT NULL DEFAULT TRUE (`0023:5`) |
| created_at | TIMESTAMPTZ | NOT NULL DEFAULT NOW() (`0023:6`) |

### `document_profiles` columns

| Column | Type | Constraints |
|---|---|---|
| code | TEXT | PK (`0023:10`); CHECK `^[a-z][a-z0-9_-]{1,63}$` (`0122:16-18`); immutable via trigger |
| family_code | TEXT | NOT NULL; FK → document_families(code) (`0023:11`) |
| name | TEXT | NOT NULL (`0023:12`) |
| description | TEXT | NOT NULL DEFAULT '' (`0023:13`) |
| review_interval_days | INT | NOT NULL CHECK > 0 (`0023:14`) |
| is_active | BOOLEAN | NOT NULL DEFAULT TRUE (`0023:15`) — superseded operationally by `archived_at` |
| created_at | TIMESTAMPTZ | NOT NULL DEFAULT NOW() |
| alias | TEXT | NOT NULL, length CHECK (`0035`) |
| tenant_id | UUID | NOT NULL DEFAULT `'ffffffff-...'` (DevTenantID) (`0122:4-6`) |
| default_template_version_id | UUID | FK → templates_template_version(id) (`0122:8-10`) |
| owner_user_id | TEXT | (`0122:11`) |
| editable_by_role | TEXT | NOT NULL DEFAULT 'admin' (`0122:12`) |
| archived_at | TIMESTAMPTZ | nullable (`0122:13`) |

### `document_process_areas` columns

| Column | Type | Constraints |
|---|---|---|
| code | TEXT | PK (`0025:2`); CHECK `^[a-z][a-z0-9_-]{1,63}$` (`0123:15-17`); immutable via trigger |
| name | TEXT | NOT NULL (`0025:3`) |
| description | TEXT | NOT NULL DEFAULT '' (`0025:4`) |
| is_active | BOOLEAN | NOT NULL DEFAULT TRUE (`0025:5`) |
| created_at | TIMESTAMPTZ | NOT NULL DEFAULT NOW() (`0025:6`) |
| tenant_id | UUID | NOT NULL DEFAULT DevTenantID (`0123:3-5`) |
| parent_code | TEXT | self-FK (tenant_id,parent_code) → (tenant_id,code) (`0123:10-13`) |
| owner_user_id | TEXT | (`0123:6`) |
| default_approver_role | TEXT | (`0123:7`) |
| archived_at | TIMESTAMPTZ | nullable (`0123:8`) |

## 2. Tables read or written but NOT owned

| Table | Owner | R/W | Use |
|---|---|---|---|
| `templates_template_version` | templates (`0120_templates_init.sql:19`) | READ | `template_version_checker.go:14-17` joins to verify `IsPublished` + return owning profile_code |
| `templates_template` | templates | READ | same join (`template_version_checker.go:14-17`) |

`HasActiveProfiles` (`family_repository.go:91-99`) queries only `document_profiles` (no joins).

## 3. Triggers, GUCs, functions

| Object | Kind | Source | Purpose |
|---|---|---|---|
| `reject_code_update()` | function | `0122_taxonomy_extend_document_profiles.sql:25-31` | RAISE EXCEPTION if `NEW.code <> OLD.code` |
| `trg_document_profiles_code_immutable` | trigger BEFORE UPDATE | `0122:33-39` | enforces immutable profile code |
| `reject_code_update()` (re-declared) | function | `0123_taxonomy_extend_process_areas.sql:23-31` | same — area variant |
| `trg_process_areas_code_immutable` | trigger BEFORE UPDATE | `0123:33-37` | enforces immutable area code |
| `0175` snapshot trigger? | — | none found in `0175_documents_area_name_snapshot.sql:1-18` (one-shot UPDATE only, not a trigger) | documents module reads `process_areas.name` LIVE at create time (`internal/modules/documents/repository/repository.go:94-96`) |

### Tripwire / authz-GUC absence

- No `assert_caps`, `tripwire`, or `set_local_tenant_id` trigger installed on any taxonomy table (grep of all migrations referencing the 3 tables — none match). Compare to approval tables (`0142b_role_capabilities_v2_enforce.sql:200-209`).
- `set_local_tenant_id` symbol not found anywhere in `internal/` (rg empty). Tenant scoping is application-layer only.
- `internal/platform/tenant/const.go:1-4` defines `DevTenantID` constant; no per-request GUC propagation.

## 4. Indexes

| Table | Index | Columns | Unique | Source |
|---|---|---|---|---|
| `document_families` | PK | code | yes | `0023:2` |
| `document_profiles` | PK | code | yes | `0023:10` |
| `document_profiles` | `ux_document_profiles_tenant_code` | (tenant_id, code) | yes | `0122:21-22` |
| `document_process_areas` | PK | code | yes | `0025:2` |
| `document_process_areas` | `ux_process_areas_tenant_code` | (tenant_id, code) | yes | `0123:19-20` |

Note: profile + area carry PK on `code` alone (cross-tenant unique) AND a `(tenant_id, code)` unique index — the broader PK is redundant; cross-tenant code collisions are impossible by PK and the per-tenant index adds no new constraint, only an index for tenant-filtered lookups.

## 5. Tripwire pairing audit

Every mutating repo method violates the tier-2 `authz.Require` pattern documented in `wiki/decisions/0007-two-tier-authz.md` (which approval, documents, iam follow).

| Method | authz.Require? | Cap+area | Verb | Table |
|---|---|---|---|---|
| `FamilyRepository.Create` (`infrastructure/family_repository.go:67`) | NO — VIOLATION | n/a | INSERT (`:69`) | `document_families` |
| `FamilyRepository.Update` (`family_repository.go:75`) | NO — VIOLATION | n/a | UPDATE (`:77`) | `document_families` |
| `ProfileRepository.Create` (`infrastructure/repository.go:102`) | NO — VIOLATION | n/a | INSERT (`:104`) | `document_profiles` |
| `ProfileRepository.Update` (`repository.go:127`) | NO — VIOLATION | n/a | UPDATE (`:129`) | `document_profiles` |
| `AreaRepository.Create` (`repository.go:253`) | NO — VIOLATION | n/a | INSERT (`:255`) | `document_process_areas` |
| `AreaRepository.Update` (`repository.go:275`) | NO — VIOLATION | n/a | UPDATE (`:277`) | `document_process_areas` |

Upstream gate: tier-1 capability check via path-prefix dispatcher (`apps/api/cmd/metaldocs-api/permissions.go:158-180`) — `taxonomy.manage` for writes, `doc.view` for reads. Defense-in-depth is single-layer, vs the two-tier + DB-tripwire model in sibling modules.

## 6. Migration history

Sequence order (dates not in filenames):

| # | Filename | Verb summary |
|---|---|---|
| 1 | 0023_init_document_family_and_profile_registry.sql | CREATE families + profiles |
| 2 | 0024_grant_document_registry_privileges.sql | GRANT |
| 3 | 0025_init_document_taxonomy.sql | CREATE process_areas |
| 4 | 0026_grant_document_taxonomy_privileges.sql | GRANT |
| 5 | 0027_init_document_profile_schema_and_governance.sql | FK from external table |
| 6 | 0029_seed_metal_nobre_document_registry.sql | seed |
| 7 | 0032_deactivate_legacy_document_registry.sql | UPDATE |
| 8 | 0034_rename_metal_nobre_document_labels.sql | UPDATE |
| 9 | 0035_add_document_profile_alias.sql | ALTER + UPDATE |
| 10 | 0046_add_document_code_sequence.sql | FK from external table |
| 11 | 0075_create_template_drafts_and_audit.sql | FK from external table |
| 12 | 0122_taxonomy_extend_document_profiles.sql | ALTER (tenant_id, default_template_version_id, archived_at, code CHECK, immutability trigger) |
| 13 | 0123_taxonomy_extend_process_areas.sql | ALTER (tenant_id, parent_code, archived_at, code CHECK, immutability trigger) |
| 14 | 0124_registry_controlled_documents.sql | FK from registry |
| 15 | 0125_registry_iam_user_process_areas_governance_events.sql | FK from iam (user_process_areas) |
| 16 | 0134_approval_routes.sql | FK from approval |
| 17 | 0161_grant_families_write_privileges.sql | GRANT (metaldocs_app write) |
| 18 | 0175_documents_area_name_snapshot.sql | one-shot snapshot UPDATE; reads process_areas.name |
| 19 | 0182_cd_sequence_per_area.sql | FK from registry |

(unclear: dates not in filenames; commit-history derivation skipped.)
