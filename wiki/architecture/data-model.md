# Architecture: Data Model

> **Last verified:** 2026-05-02
> **Status:** Stub. Expand with ERD + per-table schema notes when SQL stabilizes.
> **Scope:** Postgres tables, key relationships, snapshot columns, hash columns.
> **Out of scope:** Migration history (see `internal/platform/db/migrations/`).
> **Key files:**
> - `internal/platform/db/migrations/` — source of truth for schema
> - `internal/modules/templates_v2/infrastructure/repo/` — template tables
> - `internal/modules/documents_v2/infrastructure/repo/` — document tables
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

## Snapshot columns (documents_v2)

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

## See also

- [architecture/system-overview.md](system-overview.md)
- [modules/documents-v2.md](../modules/documents-v2.md)
- [modules/templates-v2.md](../modules/templates-v2.md)
- [modules/taxonomy.md](../modules/taxonomy.md)
