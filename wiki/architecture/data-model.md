# Architecture: Data Model

> **Last verified:** 2026-07-12 (M3 approval-kernel-extraction path fix: `internal/modules/documents/approval/infrastructure/` → `internal/modules/approval/infrastructure/`, approval promoted to top-level 15th module per ADR 0082; no other Key files entry touched) | **Prior:** 2026-07-06 (F9.4 doc-truth pass: Key-files header block fully re-verified — all `repository/`→`infrastructure/` F9.5 renames applied; corrected two independently-stale entries found on inspection — `documents/repository/repository.go:37` was wrong symbol name (`CreateDocument` doesn't exist, corrected to `CreateDocumentTx:125`) and `approval/infrastructure/repo/` was missing `documents/` prefix + nonexistent `/repo/` subdir; also fixed `templates/infrastructure/repo/` and `taxonomy/infrastructure/repo/` (no such `/repo/` subdir), and `buildDocumentFilter`/`ListOptions` body anchors) | **Prior:** 2026-07-03 (DB-01: legacy template family `public.templates`/`public.template_versions` retired via `db/migrations/0268_drop_legacy_template_family.sql`; `public.templates_template`/`public.templates_template_version` is now the only template store — `internal/platform/docgenv2` reads canonical only, `FanoutTemplateReader` and the legacy fallback reader deleted in the same change-set) | **Prior:** 2026-07-02 (DB-06: templates audit sink is fully canonical — `public.templates_audit_log` retired via `db/migrations/0262_drop_templates_audit_log.sql`; both `AppendAudit`/`AppendAuditTx` writes and `ListAudit` reads go through `metaldocs.audit_events`) | **Prior:** 2026-07-01 (DOC-02 drift fix: template version status `in_review` renamed `under_review` per migration 0257) | **Prior-2:** 2026-06-08 (Phase F F4: ListDocumentsPaginated cursor migration; repository.go line anchors updated)
> **Status:** Stub. Expand with ERD + per-table schema notes when SQL stabilizes.
> **Scope:** Postgres tables, key relationships, snapshot columns, hash columns.
> **Out of scope:** Migration archaeology (the legacy DB-research notes were removed at the v1 re-baseline, commit `c7f06f2e`; retained historical `migrations/` evidence remains).
> **Key files:**
> - `db/prerequisites/0001_extensions.sql` - required extensions before schema objects
> - `db/baseline/0001_current_schema.sql` - curated current-state schema for fresh environments
> - `db/reference-data/0001_product_reference_data.sql` - product reference data required at runtime
> - `db/dev-seeds/0001_local_dev_seed.sql` - optional local-only developer accounts/data
> - `db/migrations/` - post-baseline forward migrations
> - `internal/modules/templates/infrastructure/` — template tables (**verify** — was `.../infrastructure/repo/`, nonexistent `/repo/` subdir; no such subdir in current tree)
> - `internal/modules/documents/infrastructure/repository.go:125` — document tables; `CreateDocumentTx` INSERT (accepts `requiredPlaceholders`; seeds `document_placeholder_values` atomically) (**verify** — dir renamed repository/→infrastructure/ per F9.5; symbol was `CreateDocument`, no such func found in current tree, corrected to `CreateDocumentTx` per grep, line 125 per prior verification in wiki/modules/documents.md)
> - `internal/modules/taxonomy/infrastructure/family_repository.go:11` — document_families SQL impl
> - `internal/modules/taxonomy/infrastructure/` — profiles, areas (**verify** — was `.../infrastructure/repo/`, nonexistent `/repo/` subdir; no such subdir in current tree)
> - `internal/modules/approval/infrastructure/` — routes, signoffs (M3 2026-07-12: approval promoted to top-level 15th module, [ADR 0082](../decisions/0082-approval-kernel-extraction.md); path was `internal/modules/documents/approval/infrastructure/` pre-M3, now stale; corrected path confirmed to exist via directory listing)

## Core entities (high-level)

```
document_families (global, no tenant_id)
  └── document_profiles (tenant-scoped, family_code FK)
       └── controlled_documents (CDs)  <─┐
                            │            │
areas ───────────────────────┤           │
  └── controlled_documents               │
                                         │
templates_template                       │
  └── templates_template_version (draft|under_review|approved|published)
       └── content (eigenpal JSON)

controlled_documents
  └── document_versions (draft|under_review|approved|frozen)
       ├── content (eigenpal JSON)
       ├── placeholder_schema_snapshot (JSON, fixed at create)
       ├── content_hash, values_hash, schema_hash
       └── values_frozen_at, frozen_docx_s3_key, frozen_pdf_s3_key

approval_routes ── stages ── steps ── signoffs

users ── roles ── role_capabilities
       └── area_grants (area-scoped permissions)

outbox_events (docgen_v2_pdf, etc.)
```

## document_families table

Globally scoped — **no `tenant_id`**. All tenants share the same family catalog.

| Column | Type | Notes |
|--------|------|-------|
| `code` | text PK | immutable after creation |
| `name` | text | display name |
| `description` | text | optional |
| `is_active` | boolean | deactivation flag; differs from profiles/areas which use `archived_at` |
| `created_at` | timestamptz | set on insert |

**FK:** `document_profiles.family_code REFERENCES document_families(code)` — a profile must reference a valid family.

**Deactivation guard:** `FamilyService.Deactivate` calls `HasActiveProfiles` (`infrastructure/family_repository.go:91`) before setting `is_active = false`. Blocked if any profile with `archived_at IS NULL` references the family.

**Migration:** `migrations/0161_grant_families_write_privileges.sql` — grants `SELECT, INSERT, UPDATE, DELETE` on `document_families` to the `metaldocs_app` runtime user.

See [modules/taxonomy.md](../modules/taxonomy.md) for full entity detail and API routes.

## public.documents_v2 — dropped (migration 0168)

`public.documents_v2` was the W1 scaffold table created in migration 0103. Migration 0110 replaced it with `public.documents` (W3 schema). The table was orphaned — no Go code wrote to it after migration 0110. Migration 0168 drops it. Do not reference this table in new code or queries.

## public.templates / public.template_versions — dropped (migration 0268)

The legacy template family (`public.templates`, `public.template_versions`) was superseded by the canonical `public.templates_template` / `public.templates_template_version` pair at the ARC-01 canonical-first cutover (commit `4160018e`). Since ARC-01, `internal/platform/docgenv2.FanoutTemplateReader` tried the canonical reader first and fell back to the legacy pair only on `sql.ErrNoRows`. `db/migrations/0268_drop_legacy_template_family.sql` drops both legacy tables (plus a safety-net `DROP TABLE IF EXISTS public.signoffs`, a no-op — that table was absent from the canonical baseline), gated on runtime proof that `docgenv2.LegacyTemplateReadCount()` observed zero legacy fallback reads across a full QA run window. The `docgenv2` legacy fallback reader code (`template_reader.go`, the `FanoutTemplateReader` indirection, `LegacyTemplateReadCount()`) is deleted in the same change-set as this migration's promotion out of `db/migrations/_pending/`. `internal/platform/docgenv2` now reads the canonical template tables only. Do not reference `public.templates`/`public.template_versions` in new code or queries — see `wiki/database/tables/templates.md` and `wiki/database/tables/template_versions.md` (both RETIRED) and `wiki/modules/templates-tech-debt.md` DB-01 (closed).

## Snapshot columns (documents)

Populated at document-version creation by `application.SnapshotService`. Trigger `enforce_snapshot_on_submit_trg` blocks `draft → under_review` if any are NULL.

- `placeholder_schema_snapshot` — fixed 7-token catalog at creation time
- (more — fill in when verified)

## Hash columns (freeze)

- `content_hash` — hash of the post-substitution DOCX
- `values_hash` — hash of resolved token values
- `schema_hash` — hash of the placeholder schema snapshot

See [concepts/freeze-and-hashing.md](../concepts/freeze-and-hashing.md).

## Audit sink (templates)

Template audit history has one sink, not two: `metaldocs.audit_events` — a tamper-evident, hash-chained append-only log (`prev_hash`/`row_hash` via `metaldocs.audit_event_row_hash`, ordered by `audit_sequence`). `internal/modules/templates/infrastructure/postgres.go` `AppendAudit`/`AppendAuditTx` (:604-624) write to it through `auditdomain.Writer`; `ListAudit` (:660-706) reads from it filtered to `resource_type = 'template'`. The former module-local `public.templates_audit_log` table (write path closed 2026-05-11 Plan 6a; read path closed earlier in Wave 1.8/F-07-sub-split) was dropped by `db/migrations/0262_drop_templates_audit_log.sql`, which backfilled any pre-cutover rows into the hash chain first. See `wiki/modules/templates-tech-debt.md` T-013 (closed).

## Runtime DB user permissions (metaldocs_app)

Migration `0160` (`migrations/0160_grant_metaldocs_app_schema_objects.sql`) grants the runtime DB user `metaldocs_app` the permissions it needs:

| Object | Grants |
|--------|--------|
| schema `metaldocs` | USAGE |
| `metaldocs.job_leases` | SELECT, INSERT, UPDATE, DELETE |
| `metaldocs.acquire_lease`, `heartbeat_lease`, `release_lease`, `assert_lease_epoch` | EXECUTE |
| `metaldocs.idempotency_keys` | SELECT, INSERT, UPDATE |

Migration `0161` additionally grants `SELECT, INSERT, UPDATE, DELETE` on `document_families`.

Without `0160`: the background scheduler cannot acquire job leases, and `PostgresSignoffIdempStore` returns Postgres permission errors on signoff replay checks.
Without `0161`: all family API endpoints fail with Postgres permission errors.

## List/stats query pattern (Library)

`ListDocumentsPaginated` and its sibling queries (`CountDocuments`, `StatsByStatus`, `StatsByArea`) all share a single `buildDocumentFilter` helper at `internal/modules/documents/infrastructure/repository.go:429` (**verify** — line not re-confirmed this pass).

The **keyset cursor pattern** used by the Library screen (Phase F F4):
- `ListDocumentsPaginated` returns `(items, hasMore)` using an opaque cursor from `internal/platform/pagination` (`EncodeCursor`/`DecodeCursor`/`ClampLimit`).
- The service layer re-exports `hasMore` and builds `next_cursor` in the handler envelope `{items, page:{next_cursor, has_more}}`.
- A `total` count query still runs via `CountDocuments` for pagination controls where needed.

`ListOptions` struct lives in `internal/modules/documents/infrastructure/repository.go:434` (**verify** — was `repository.go:412`, dir renamed + line re-grepped this pass); it is re-exported as a type alias at `application/list_options.go:1` so handlers depend only on the `application` package.

Index note: `(tenant_id, status)` and `(tenant_id, process_area_code_snapshot)` may benefit from a composite index under real-data load — tracked as a backend follow-up in `wiki/implementation/plan-library.md`.

## See also

- [architecture/system-overview.md](system-overview.md)
- [modules/documents.md](../modules/documents.md)
- [modules/templates.md](../modules/templates.md)
- [modules/taxonomy.md](../modules/taxonomy.md)
