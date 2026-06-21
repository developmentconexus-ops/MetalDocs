# Feature F4.2 — documents search projection read contract

> **Milestone:** 4 — search consumes published visibility contracts  ·  **Folder:** `f4.2-documents-search-projection`
> **Status:** Closed (2026-06-21 — parity GREEN on PG :5434; evidence.md written)

## Source

- Milestone spec row F4.2 + `spec.md` (operator-approved 1:1 passthrough). ADR-0039 D3a/D4.

## Plan

**Contract (from `spec.md`):** one `metaldocs` view, 1 row/document, the 14 columns search reads from
`public.documents`, pure projection — no `WHERE`, no `COALESCE`.

**TDD order:**
1. **RED** — write `internal/modules/documents/repository/document_search_facts_parity_integration_test.go`
   (`//go:build integration`, package `repository`). Seed 4 discriminator docs in one tenant via direct
   inserts (testdb isolated clone): standalone (NULL `controlled_document_id`), CD-linked, archived
   (`archived_at` set), NULL-snapshot (NULL `profile_code_snapshot`/`process_area_code_snapshot`/`code`).
   Assertions:
   - row count: `count(v_document_search_facts WHERE tenant=$1)` == `count(public.documents WHERE tenant=$1 ...)`.
   - per-doc NULL-safe equality: view row's 14 cols == base row's 14 cols (scan into `sql.Null*`).
   - archived row PRESENT in the view (passthrough — no hidden filter).
   Run → FAIL (view doesn't exist).
2. **GREEN** — write `db/migrations/0244_documents_search_projection.sql` creating the view + `COMMENT ON VIEW`
   + `schema_migrations` insert. Rerun → PASS.
3. Extend ADR-0039 `Related code` note: add migration 0244 + `v_document_search_facts`.

**Files touched:**
- NEW `db/migrations/0244_documents_search_projection.sql`
- NEW `internal/modules/documents/repository/document_search_facts_parity_integration_test.go`
- EDIT `wiki/decisions/0039-cross-module-base-table-read-boundary.md` (Related code note only)

**Test strategy:** integration parity on real PG :5434 (`127.0.0.1:5434`, `-tags integration`). The archived +
NULL-snapshot rows are the anti-drift discriminators (prove no hidden filter / no COALESCE). No Go production
code, no search change (F4.3).

**Non-goals (guardrails):** no search change; no CD view (F4.1 done); no documents repository change; no filter
or COALESCE in the view; no columns beyond the 14.

## Execution notes

- Implemented directly in main session (1 migration + 1 integration test). Live DSN: `127.0.0.1:5434`.
- `public.documents` has NOT-NULL `template_version_id`, `form_data_json`, `name`, `status`, `created_by`,
  `tenant_id` — the seed must satisfy these (and any FK). Discover constraints empirically during RED.
