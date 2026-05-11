# Phase 4 — Persistence map · templates_v2

> Date: 2026-05-10
> Source: `migrations/0120_templates_v2_init.sql`, `internal/modules/templates_v2/repository/postgres.go`, application service files.
> Note: codex-rescue ran read-only/sandboxed on dispatch — main agent verified directly from source and authored this artifact.

## 1. Tables owned (4)

All four tables are created in a single migration: `migrations/0120_templates_v2_init.sql`.

### `templates_v2_template`

| Column | Type | Constraints |
|---|---|---|
| `id` | uuid | PRIMARY KEY |
| `tenant_id` | text | NOT NULL |
| `doc_type_code` | text | NOT NULL |
| `key` | text | NOT NULL; UNIQUE (tenant_id, key) |
| `name` | text | NOT NULL |
| `description` | text | NOT NULL DEFAULT '' |
| `areas` | text[] | NOT NULL DEFAULT '{}' |
| `visibility` | text | NOT NULL (`public` / `internal` / `specific`) |
| `specific_areas` | text[] | NOT NULL DEFAULT '{}' |
| `latest_version` | int | NOT NULL DEFAULT 0 |
| `published_version_id` | uuid | NULL; FK → `templates_v2_template_version(id)` (ALTER ADD CONSTRAINT) |
| `created_by` | text | NOT NULL |
| `created_at` | timestamptz | NOT NULL DEFAULT now() |
| `archived_at` | timestamptz | NULL |

### `templates_v2_template_version`

| Column | Type | Constraints |
|---|---|---|
| `id` | uuid | PRIMARY KEY |
| `template_id` | uuid | NOT NULL; FK → `templates_v2_template(id)` |
| `version_number` | int | NOT NULL; UNIQUE (template_id, version_number) |
| `status` | text | NOT NULL (`draft` / `in_review` / `approved` / `published` / `obsolete`) |
| `docx_storage_key` | text | NOT NULL |
| `content_hash` | text | NOT NULL (empty string sentinel until upload commit) |
| `metadata_schema` | jsonb | NOT NULL |
| `placeholder_schema` | jsonb | NOT NULL |
| `editable_zones` | jsonb | NOT NULL (legacy from zone-purge era; ADR 0002) |
| `author_id` | text | NOT NULL |
| `pending_reviewer_role` | text | NULL |
| `pending_approver_role` | text | NOT NULL DEFAULT '' |
| `reviewer_id` | text | NULL |
| `approver_id` | text | NULL |
| `submitted_at` / `reviewed_at` / `approved_at` / `published_at` / `obsoleted_at` | timestamptz | NULL |
| `created_at` | timestamptz | NOT NULL DEFAULT now() |

### `templates_v2_approval_config`

| Column | Type | Constraints |
|---|---|---|
| `template_id` | uuid | PRIMARY KEY; FK → `templates_v2_template(id)` |
| `reviewer_role` | text | NULL |
| `approver_role` | text | NOT NULL |

### `templates_v2_audit_log`

| Column | Type | Constraints |
|---|---|---|
| `id` | bigserial | PRIMARY KEY |
| `tenant_id` | text | NOT NULL |
| `template_id` | uuid | NOT NULL (no FK declared) |
| `version_id` | uuid | NULL (no FK declared) |
| `actor_id` | text | NOT NULL |
| `action` | text | NOT NULL |
| `details` | jsonb | NOT NULL DEFAULT '{}' |
| `occurred_at` | timestamptz | NOT NULL DEFAULT now() |

## 2. Tables read or written but NOT owned

| Table | Owner | R/W | Operation site |
|---|---|---|---|
| `documents_v2` | documents (W1 scaffold, dropped per `wiki/modules/documents.md`) | R/W | `migrations/0121_documents_v2_link_template_version.sql` adds FK to `templates_v2_template_version(id)` |
| `controlled_documents` | registry | R | `migrations/0124_registry_controlled_documents.sql` adds FK to `templates_v2_template_version(id)` |
| `metaldocs.role_capabilities` | iam | seed-time INSERT | `migrations/0165_role_capabilities_reseed.sql` rows 9–38 seed `template.view/create/edit/submit/approve/publish` |

## 3. Triggers, GUCs, functions

| Object | Kind | Site | Purpose |
|---|---|---|---|
| — | — | — | None. No triggers, no GUC reads/sets, no PL/pgSQL functions installed by this module's migrations. |

Compare: `documents` module installs `enforce_snapshot_on_submit_trg` (per `wiki/modules/documents.md §6`); `templates_v2` installs no equivalent enforcement. All invariants are application-layer only.

## 4. Indexes

| Index | Table | Columns | Unique? | Purpose |
|---|---|---|---|---|
| `templates_v2_template_pkey` | `templates_v2_template` | `(id)` | yes | PK |
| `templates_v2_template_tenant_id_key_key` | `templates_v2_template` | `(tenant_id, key)` | yes | unique key per tenant |
| `idx_templates_v2_template_tenant_doctype` | `templates_v2_template` | `(tenant_id, doc_type_code)` | no | list filter |
| `templates_v2_template_version_pkey` | `templates_v2_template_version` | `(id)` | yes | PK |
| `templates_v2_template_version_template_id_version_number_key` | `templates_v2_template_version` | `(template_id, version_number)` | yes | version uniqueness |
| `idx_templates_v2_version_template_status` | `templates_v2_template_version` | `(template_id, status)` | no | status lookup |
| `templates_v2_approval_config_pkey` | `templates_v2_approval_config` | `(template_id)` | yes | PK |
| `templates_v2_audit_log_pkey` | `templates_v2_audit_log` | `(id)` | yes | PK |
| `idx_templates_v2_audit_template_time` | `templates_v2_audit_log` | `(template_id, occurred_at DESC)` | no | recent-events scan |

## 5. Tripwire pairing audit

Repo mutations (verified at `internal/modules/templates_v2/repository/postgres.go`):

| Method (file:line) | Authz.Require called? | Cap + area arg | SQL verb | Table |
|---|---|---|---|---|
| `Repository.CreateTemplate` (postgres.go:29) | **NO** | (none) | INSERT | `templates_v2_template` — VIOLATION |
| `Repository.UpdateTemplate` (postgres.go:134) | **NO** | (none) | UPDATE | `templates_v2_template` — VIOLATION |
| `Repository.CreateVersion` (postgres.go:169) | **NO** | (none) | INSERT | `templates_v2_template_version` — VIOLATION |
| `Repository.UpdateVersion` (postgres.go:236) | **NO** | (none) | UPDATE | `templates_v2_template_version` — VIOLATION |
| `Repository.ObsoletePreviousPublished` (postgres.go:271) | **NO** | (none) | UPDATE | `templates_v2_template_version` — VIOLATION |
| `Repository.UpsertApprovalConfig` (postgres.go:302) | **NO** | (none) | INSERT/UPDATE | `templates_v2_approval_config` — VIOLATION |
| `Repository.AppendAudit` (postgres.go:318) | **NO** | (none) | INSERT | `templates_v2_audit_log` — VIOLATION (audit) |

Total tripwire violations: **7**.

Root cause: the module's HTTP handler accepts an `AuthzFunc` argument but `apps/api/cmd/metaldocs-api/main.go:329` wires `nil`, and the `New(...)` constructor falls through to a no-op `func(*http.Request, string, string, string) error { return nil }` (handler.go:25–27). No `internal/platform/authz.Require` call exists anywhere in the module. There is no Postgres `metaldocs.asserted_caps` GUC tripwire installed for these tables — purely advisory `template.*` capabilities seeded in iam migration 0165 are never asserted.

## 6. Migration history (chronological)

| # | Filename | Verb summary | Notes |
|---|---|---|---|
| 1 | `0101_docx_v2_templates.sql` | CREATE TABLE `templates` (V2 lineage initial) | DOCX-V2 scaffold |
| 2 | `0102_docx_v2_template_versions.sql` | CREATE TABLE `template_versions`; ALTER `templates` | DOCX-V2 scaffold |
| 3 | `0108_docx_v2_template_audit_log.sql` | CREATE TABLE `template_audit_log` | DOCX-V2 scaffold |
| 4 | `0109_docx_v2_templates_w2_noop.sql` | no-op | placeholder for cutover sequencing |
| 5 | `0120_templates_v2_init.sql` | CREATE 4 owned tables (`templates_v2_*`) + 3 indexes + FK | **canonical templates_v2 schema** |
| 6 | `0121_documents_v2_link_template_version.sql` | ALTER `documents_v2` → FK to `templates_v2_template_version` | downstream coupling |
| 7 | `0124_registry_controlled_documents.sql` | CREATE `controlled_documents` w/ FK to `templates_v2_template_version` | registry coupling |
| 8 | `0126_documents_v2_bridge_columns.sql` | ALTER `documents_v2` (bridge) | downstream cleanup |
| 9 | `0129_documents_v2_bridge_not_null.sql` | ALTER `documents_v2` NOT NULL | downstream cleanup |
| 10 | `0130_documents_drop_old_template_version_fk.sql` | ALTER `documents` drop legacy FK | retires DOCX-V2 lineage FK |
| 11 | `0157_drop_editable_zones.sql` | (unclear: confirm — `editable_zones` jsonb still present in current DDL per init at 0120; this migration name suggests a schema-level drop that was either deferred or applies to a different column path) | possible drift — see tech-debt |
| 12 | `0165_role_capabilities_reseed.sql` | TRUNCATE/INSERT `metaldocs.role_capabilities` | seeds `template.view/create/edit/submit/approve/publish` (currently unenforced — see §5) |

Total migrations: **12**.

## Counts

- Tables owned: **4**
- Tripwire violations: **7**
- Migrations: **12**
