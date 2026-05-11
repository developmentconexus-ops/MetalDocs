# Registry Persistence Surface

Module: `internal/modules/registry`
Scope source: module code + repo-root `migrations/`
Last verified: 2026-05-10

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
| `override_template_version_id` | `UUID` | nullable FK -> templates_v2_template_version(id) |
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
| `documents` | documents | Read + Write | Backfill read/update (migration.go:30-34, :77-84); active-doc read (routes.go:243-253) |
| `approval_instances` | approval | Read | active-doc query (routes.go:235-238, :307-312) |
| `document_revisions` | documents | Read | content hash fallback (routes.go:223-225) |
| `templates_v2_template_version` | templates_v2 | Read | template state lookup (repository.go:278-280) |
| `templates_v2_template` | templates_v2 | Read | profile-code join (repository.go:279) |
| `document_profiles` | taxonomy | Read | profile lookup (repository.go:307-308) |
| `document_process_areas` | taxonomy | Read | area lookup (repository.go:346-347) |
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
| `public.enforce_capability_asserted()` | function | 0142b_role_capabilities_v2_enforce.sql:67 | reads metaldocs.asserted_caps GUC for tripwire enforcement |
| `trg_require_cap_asserted_instances` | trigger | 0142b_role_capabilities_v2_enforce.sql:201 | capability-assertion tripwire trigger |
| `trg_require_cap_asserted_signoffs` | trigger | 0142b_role_capabilities_v2_enforce.sql:207 | capability-assertion tripwire trigger |

GUC reads in registry module SQL:
- `metaldocs.asserted_caps`: NO evidence in internal/modules/registry.
- `metaldocs.tenant_id`: NO evidence in internal/modules/registry.

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
| `Create` (repository.go:133-146) | NO | (unclear: no registry-local Authz.Require call) | INSERT | `controlled_documents` |
| `CreateTx` (repository.go:137-146) | NO | (unclear: no registry-local Authz.Require call) | INSERT | `controlled_documents` |
| `UpdateStatus` (repository.go:184-187) | NO | (unclear: no registry-local Authz.Require call) | UPDATE | `controlled_documents` |
| `EnsureCounter` (repository.go:208-216) | NO | (unclear: no registry-local Authz.Require call) | INSERT ... ON CONFLICT DO NOTHING | `cd_sequence_counters` |
| `NextAndIncrement` (repository.go:239, 251-254) | NO | (unclear: no registry-local Authz.Require call) | INSERT ... ON CONFLICT DO NOTHING; UPDATE | `cd_sequence_counters` |

VIOLATIONS (NO Authz.Require + mutating verb): Create, CreateTx, UpdateStatus, EnsureCounter, NextAndIncrement — all 5.

Tenant enforcement split:
- SET LOCAL metaldocs.tenant_id before mutation: NO evidence in registry repo SQL.
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

No additional migrations found touching controlled_documents / cd_sequence_counters / profile_sequence_counters beyond the above 7.

---

<!-- summary: 3 tables owned · 5 tripwire violations · 7 migrations -->
