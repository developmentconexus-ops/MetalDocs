# Architecture: Data Model

> **Last verified:** 2026-05-15
> **Status:** Stub. Expand with ERD + per-table schema notes when SQL stabilizes.
> **Scope:** Postgres tables, key relationships, snapshot columns, hash columns.
> **Out of scope:** Migration archaeology (see `docs/db-research/` and retained historical `migrations/` evidence).
> **Key files:**
> - `db/prerequisites/0001_extensions.sql` - required extensions before schema objects
> - `db/baseline/0001_current_schema.sql` - curated current-state schema for fresh environments
> - `db/reference-data/0001_product_reference_data.sql` - product reference data required at runtime
> - `db/dev-seeds/0001_local_dev_seed.sql` - optional local-only developer accounts/data
> - `db/migrations/` - post-baseline forward migrations
> - `internal/modules/templates/infrastructure/repo/` — template tables
> - `internal/modules/documents/repository/repository.go:37` — document tables; `CreateDocument` INSERT (accepts `requiredPlaceholders`; seeds `document_placeholder_values` atomically)
> - `internal/modules/taxonomy/infrastructure/family_repository.go:11` — document_families SQL impl
> - `internal/modules/taxonomy/infrastructure/repo/` — profiles, areas
> - `internal/modules/approval/infrastructure/repo/` — routes, signoffs

## Core entities (high-level)

```
document_families (global, no tenant_id)
  └── document_profiles (tenant-scoped, family_code FK)
       └── controlled_documents (CDs)  <─┐
                            │            │
areas ───────────────────────┤           │
  └── controlled_documents               │
                                         │
templates                                │
  └── template_versions (draft|in_review|approved|published)
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

## Snapshot columns (documents)

Populated at document-version creation by `application.SnapshotService`. Trigger `enforce_snapshot_on_submit_trg` blocks `draft → under_review` if any are NULL.

- `placeholder_schema_snapshot` — fixed 7-token catalog at creation time
- (more — fill in when verified)

## Hash columns (freeze)

- `content_hash` — hash of the post-substitution DOCX
- `values_hash` — hash of resolved token values
- `schema_hash` — hash of the placeholder schema snapshot

See [concepts/freeze-and-hashing.md](../concepts/freeze-and-hashing.md).

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

`ListDocumentsPaginated` and its sibling queries (`CountDocuments`, `StatsByStatus`, `StatsByArea`) all share a single `buildDocumentFilter` helper at `internal/modules/documents/repository/repository.go:284`.

The **two-query pattern** used by the Library screen:
1. `SELECT … FROM documents WHERE <filter> ORDER BY updated_at DESC LIMIT $N OFFSET $M` — returns page items.
2. `SELECT COUNT(*) FROM documents WHERE <filter>` — returns total for pagination controls.

Both queries execute the same WHERE clause (same args, same parameterized bindings) to keep item count and total in sync. Stats queries add `GROUP BY status` or `GROUP BY process_area_code_snapshot` on the same filter.

`ListOptions` struct lives in `repository.go:253`; it is re-exported as a type alias at `application/list_options.go:1` so handlers depend only on the `application` package.

Index note: `(tenant_id, status)` and `(tenant_id, process_area_code_snapshot)` may benefit from a composite index under real-data load — tracked as a backend follow-up in `wiki/implementation/plan-library.md`.

## See also

- [architecture/system-overview.md](system-overview.md)
- [modules/documents.md](../modules/documents.md)
- [modules/templates.md](../modules/templates.md)
- [modules/taxonomy.md](../modules/taxonomy.md)
