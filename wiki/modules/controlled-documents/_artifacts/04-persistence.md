# Controlled Documents Persistence Surface

Module: `internal/modules/controlleddocuments`
Scope source: module code + repo-root `migrations/`
Last verified: 2026-06-11

## 1. Tables owned

| Table | Created in (migration filename) | Notes |
|---|---|---|
| `controlled_documents` | `0124_registry_controlled_documents.sql` | Current registry-owned table. |
| `cd_sequence_counters` | `0182_cd_sequence_per_area.sql` | Current sequence counter table (per tenant + profile + process area). |
| `profile_sequence_counters` | `0124_registry_controlled_documents.sql` | Legacy; dropped in `0182_cd_sequence_per_area.sql`. |

### `controlled_documents`
| Column | Type | Constraints (NOT NULL, FK, default) |
|---|---|---|
| `id` | `UUID` | PRIMARY KEY, DEFAULT gen_random_uuid() |
| `tenant_id` | `UUID` | NOT NULL |
| `profile_code` | `TEXT` | NOT NULL; FK (tenant_id, profile_code) -> document_profiles(tenant_id, code) |
| `process_area_code` | `TEXT` | NOT NULL; FK (tenant_id, process_area_code) -> document_process_areas(tenant_id, code) |
| `department_code` | `TEXT` | nullable |
| `code` | `TEXT` | NOT NULL; UNIQUE (tenant_id, profile_code, code); check length 2..100 |
| `sequence_num` | `INT` | nullable |
| `title` | `TEXT` | NOT NULL |
| `owner_user_id` | `TEXT` | NOT NULL |
| `override_template_version_id` | `UUID` | nullable FK -> templates_template_version(id) |
| `status` | `TEXT` | NOT NULL, default 'active', check in ('active','obsolete','superseded') |
| `created_at` | `TIMESTAMPTZ` | NOT NULL, DEFAULT now() |
| `updated_at` | `TIMESTAMPTZ` | NOT NULL, DEFAULT now() |

`tenant_id`: present.

### `cd_sequence_counters`
| Column | Type | Constraints (NOT NULL, FK, default) |
|---|---|---|
| `tenant_id` | `UUID` | NOT NULL; part of PK/FKs |
| `profile_code` | `TEXT` | NOT NULL; part of PK; FK -> document_profiles(tenant_id, code) |
| `process_area_code` | `TEXT` | NOT NULL; part of PK; FK -> document_process_areas(tenant_id, code) |
| `next_seq` | `INT` | NOT NULL, DEFAULT 1 |

`tenant_id`: present.

### `controlled_document_area_grants`

Created by `archive/0198_controlled_document_visibility.sql`. PK: `(tenant_id, controlled_document_id, area_code)`. FK: `controlled_document_id → controlled_documents(id) ON DELETE CASCADE`; `(tenant_id, area_code) → document_process_areas`. Index: `ix_cd_area_grants_tenant_area (tenant_id, area_code, controlled_document_id)`.

### `controlled_document_user_grants`

Created by `archive/0198_controlled_document_visibility.sql`. PK: `(tenant_id, controlled_document_id, user_id)`. FK: `controlled_document_id → controlled_documents(id) ON DELETE CASCADE`. Index: `ix_cd_user_grants_tenant_user (tenant_id, user_id, controlled_document_id)`.

Also adds `visibility_scope TEXT NOT NULL DEFAULT 'company'` column to `controlled_documents`.

---

### `profile_sequence_counters` (legacy — dropped)
| Column | Type | Constraints (NOT NULL, FK, default) |
|---|---|---|
| `tenant_id` | `UUID` | NOT NULL; part of PK/FK |
| `profile_code` | `TEXT` | NOT NULL; part of PK; FK -> document_profiles(tenant_id, code) |
| `next_seq` | `INT` | NOT NULL, DEFAULT 1 |

`tenant_id`: present. State: dropped in `0182_cd_sequence_per_area.sql`.

---

## 2. Tables read or written but NOT owned

| Table | Owner module | Read / Write | Operations using it |
|---|---|---|---|
| `documents` | documents | Read + Write | Backfill read/update (`archive/migrations/0183_documents_name_not_empty.sql:19-23`); active-doc read (`delivery/http/routes.go:345-356`) |
| `approval_instances` | approval | Read | active-doc approval enrichment (`delivery/http/routes.go:408-415`) |
| `document_revisions` | documents | Read | content hash fallback in COALESCE (`delivery/http/routes.go:339-341`) |
| `templates_template_version` | templates | Read | template state lookup (`infrastructure/repository.go:603-609`) |
| `templates_template` | templates | Read | profile-code join inside template state lookup (`infrastructure/repository.go:606`) |
| `document_profiles` | taxonomy | Read | profile lookup (`infrastructure/repository.go:631-641`) |
| `document_process_areas` | taxonomy | Read | area lookup (`infrastructure/repository.go:671-681`) |
| `schema_migrations` | migration framework | Read + Write | idempotency guards in migrations 0167, 0182, 0183 |
| `user_process_areas` | IAM/taxonomy authz | (unclear: grant-only evidence) | grant in 0128_grants_new_tables.sql:5-6 |
| `governance_events` | taxonomy governance | (unclear: direct SQL not in registry module) | grant in 0128_grants_new_tables.sql:7 |
| `idempotency_keys` | platform/idempotency | (unclear: indirect via middleware) | idempotency.Require(...) in handler.go:80-82 |

---

## 3. Triggers, GUCs, functions

| Object | Kind | File:line | Purpose |
|---|---|---|---|
| `reject_code_update()` | function | 0124_registry_controlled_documents.sql:47 | reject updates to controlled_documents.code |
| `trg_controlled_documents_code_immutable` | trigger | 0124_registry_controlled_documents.sql:59 | invoke reject_code_update() before update |
| `check_document_tenant_consistency()` | function | 0127_documents_v2_tenant_consistency_trigger.sql:3 | cross-table tenant consistency check against controlled_documents |
| `trg_documents_v2_tenant_consistency` | trigger | 0127_documents_v2_tenant_consistency_trigger.sql:28 | calls tenant consistency function on documents_v2 writes |
| `public.enforce_capability_asserted()` | function | 0231_db_hardening_tripwire_and_dead_schema.sql | fail-close tripwire; reads `metaldocs.asserted_caps` GUC |
| `trg_require_cap_asserted` on `controlled_documents` | trigger | 0231_db_hardening_tripwire_and_dead_schema.sql | INSERT → requires `controlled_documents.create`; UPDATE → requires `controlled_documents.obsolete` OR `controlled_documents.supersede` |
| `trg_require_cap_asserted` on `cd_sequence_counters` | trigger | 0231_db_hardening_tripwire_and_dead_schema.sql | INSERT → requires `controlled_documents.create` |

GUC writes in controlleddocuments module SQL:
- `metaldocs.tenant_id`: set via `setAuthzGUC` in `application/service.go` before in-tx mutations.
- `metaldocs.actor_id`: set via `setAuthzGUC` in `application/service.go` before in-tx mutations.
- `metaldocs.asserted_caps`: written by `authz.Require` (platform package) inside the transaction.

---

## 4. Indexes

| Index | Table | Columns | Unique? | Purpose |
|---|---|---|---|---|
| `ix_controlled_documents_tenant_area` | `controlled_documents` | (tenant_id, process_area_code) | No | tenant/area filtering |
| `ix_controlled_documents_tenant_profile` | `controlled_documents` | (tenant_id, profile_code) | No | tenant/profile filtering |
| unique constraint index | `controlled_documents` | (tenant_id, profile_code, code) | Yes | per-profile code uniqueness |
| PK index | `controlled_documents` | (id) | Yes | row identity |
| PK index | `profile_sequence_counters` | (tenant_id, profile_code) | Yes | legacy sequence key |
| PK index | `cd_sequence_counters` | (tenant_id, profile_code, process_area_code) | Yes | current sequence key |

---

## 5. Tripwire pairing audit

| Method (file:line) | Authz.Require called? | Cap + area arg | SQL verb | Table |
|---|---|---|---|---|
| `Create` (repository.go:333) | YES — `repository.go:341` | `CapControlledDocumentCreate`, `doc.ProcessAreaCode` | INSERT | `controlled_documents` |
| `CreateTx` (repository.go:353) | YES — `repository.go:362` | `CapControlledDocumentCreate`, `doc.ProcessAreaCode` | INSERT | `controlled_documents` |
| `UpdateStatus` (repository.go:432) | NO | n/a — relies on tier-2 `authz.Require` called in `service.go:changeStatus` before `UpdateStatusTx` | UPDATE | `controlled_documents` |
| `EnsureCounter` (via `ensureCounterViaExec` in service) | NO direct call — sequence counter lazily initialized inside caller-owned tx that has already asserted `authz.Require` | n/a | INSERT ON CONFLICT DO NOTHING | `cd_sequence_counters` |
| `NextAndIncrement` (repository.go:559) | NO direct call — runs inside caller-owned tx that has already asserted `authz.Require` (`service.go`) | n/a | UPDATE … RETURNING | `cd_sequence_counters` |

Resolved violations (T-004 closed Plan 5): `Create` and `CreateTx` now call `authz.Require` at `repository.go:341` and `repository.go:362` respectively. `UpdateStatus`, `EnsureCounter`, and `NextAndIncrement` remain without direct `authz.Require` at the repository layer; all are guarded by the tier-2 check in `application/service.go` before the enclosing tx executes them. Tier-3 Postgres `trg_require_cap_asserted` (migration 0231) provides DB-level backstop.

Tenant enforcement split:
- `SET LOCAL metaldocs.tenant_id` / `metaldocs.actor_id`: YES — set by `setAuthzGUC` in `application/service.go` before in-tx mutations.
- tenant_id passed as query arg: YES — all mutators include tenant_id in SQL predicates/values.

---

## 6. Migration history

| Order | Filename | Verb summary | Date (from filename or git log) |
|---|---|---|---|
| 1 | `0124_registry_controlled_documents.sql` | CREATE controlled_documents; CREATE legacy profile_sequence_counters; indexes; immutability function + trigger | (unclear: no embedded date; git log not queried) |
| 2 | `0126_documents_v2_bridge_columns.sql` | ADD bridge columns + index on documents_v2 with FK to controlled_documents | (unclear: no embedded date; git log not queried) |
| 3 | `0127_documents_v2_tenant_consistency_trigger.sql` | CREATE tenant consistency function + trigger | (unclear: no embedded date; git log not queried) |
| 4 | `0128_grants_new_tables.sql` | GRANT on new tables | (unclear: no embedded date; git log not queried) |
| 5 | `0167_documents_bridge_and_state_columns.sql` | bridge/state repair on documents with FK to controlled_documents | (unclear: no embedded date; git log not queried) |
| 6 | `0182_cd_sequence_per_area.sql` | DROP profile_sequence_counters; TRUNCATE controlled_documents; CREATE cd_sequence_counters | (unclear: no embedded date; git log not queried) |
| 7 | `0183_documents_name_not_empty.sql` | backfill documents.name from controlled_documents.title; enforce non-empty name | (unclear: no embedded date; git log not queried) |

Additional migrations discovered during Stage-1 audit (2026-06-10):

| Order | Filename | Verb summary | Date |
|---|---|---|---|
| 8 | `archive/0188_tripwire_extend.sql` | Attaches `trg_require_cap_asserted` to `controlled_documents` (INSERT + UPDATE) and `cd_sequence_counters` (legacy capability names `registry.*`) | — |
| 9 | `archive/0198_controlled_document_visibility.sql` | ADD `visibility_scope` column; CREATE `controlled_document_area_grants`, `controlled_document_user_grants` | — |
| 10 | `db/migrations/0203_rename_templates_v2_objects.sql` | Renames template objects referenced by FK | — |
| 11 | `db/migrations/0210_controlled_documents_capability_namespace.sql` | Renames capability values from `registry.*` to `controlled_documents.*` | — |
| 12 | `db/migrations/0225_authz_p2_document_lifecycle_grants.sql` | Grants lifecycle capabilities to `area_admin`, `qms_admin` roles | — |
| 13 | `db/migrations/0231_db_hardening_tripwire_and_dead_schema.sql` | Rewrites `enforce_capability_asserted` with fail-close logic; maps `controlled_documents` INSERT → `controlled_documents.create`, UPDATE → `controlled_documents.obsolete|supersede`; `cd_sequence_counters` → `controlled_documents.create` | — |

---

<!-- summary: 5 tables owned (including 2 visibility grant tables) · Create+CreateTx violations closed (T-004) · 13 migrations -->
