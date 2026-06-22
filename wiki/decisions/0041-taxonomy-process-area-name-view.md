# ADR 0041 — taxonomy publishes `metaldocs.v_process_area_name` (per-area label read contract)

> **Status:** Accepted 2026-06-21
> **Last verified:** 2026-06-21
> **Deciders:** leandrotca.work (operator), MetalDocs backend
> **Context window:** Mission `frontend-screen-completion` · Milestone M2 (Distribuição coverage-scope) · Feature F2.1b (re-decomposition under HS-6 path A).
> **Related ADRs:** [0039 — Cross-module read boundary](./0039-cross-module-base-table-read-boundary.md) (this view is a D3(a) exemption — the compliant mechanism distribution uses to read area names); [0040 — `v_cd_obligated_readers`](./0040-cd-obligated-readers-view.md) (sibling published view F2.2 joins on `(tenant_id, area_code)`); ADR-0042 (distribution module + denominator-only contract, F2.1c).
> **Related code (Last verified 2026-06-21):**
> - `db/migrations/0246_taxonomy_process_area_name_view.sql` — this view.
> - `internal/modules/taxonomy/infrastructure/area_catalog_reader.go:32` — the existing taxonomy port reading the same base table with the same (no-filter) semantic.
> - `db/migrations/0245_cd_obligated_readers_view.sql` — sibling F2.1a view the F2.2 consumer joins on `(tenant_id, area_code)`.

## Context

M2/F2.1c's `DistributionRecipient` and `DistributionAreaCoverage` schemas carry an `area_name` field (the human label rendered next to each recipient + per-area total on the Distribuição screen). The base table holding it is `metaldocs.document_process_areas` (`name text NOT NULL`), owned by the taxonomy module.

ADR-0039 forbids the new `distribution` module from reading taxonomy's base table raw (`hgcrossmodule` H-G violation). Taxonomy must therefore publish a read contract.

Recon (this session, HEAD post-F2.1a): no existing `metaldocs.v_*` view exposes `area_name`. The taxonomy module has only one published port today — the in-Go `AreaCatalogReader` port (ADR-0039 D3b) — which serves single-area lookups inside the documents create tx and the iam pre-invite check; it is not appropriate for the distribution module's set-shaped `JOIN` use case (distribution joins thousands of obligated-reader rows to area names in a single query, not one lookup at a time).

## Decision

### D1 — Publish a minimal sibling view

Migration 0246 creates `metaldocs.v_process_area_name` as a plain `CREATE VIEW` over `metaldocs.document_process_areas`. No new base table; no Go code; no port change. The taxonomy in-Go port (`AreaCatalogReader`) is untouched — it continues to serve its existing two callers (B7 in-tx documents create, B8 off-tx iam pre-invite).

### D2 — Shape

```
metaldocs.v_process_area_name (
  tenant_id  uuid not null,
  area_code  text not null,
  area_name  text not null
)
-- One row per (tenant_id, area_code). 1:1 projection of
-- metaldocs.document_process_areas.
```

Renames: `code → area_code`, `name → area_name`. The `area_code` rename aligns the shape with F2.1a's `v_cd_obligated_readers.area_code` so the F2.2 join is natural; the `area_name` rename disambiguates "name" at the consumer (distribution may project area names alongside user names).

### D3 — No `is_active` / `archived_at` filter

The existing taxonomy port (`internal/modules/taxonomy/infrastructure/area_catalog_reader.go:32`) reads the base table without any active/archived filter — names resolve for archived areas too. The view preserves that semantic so label rendering remains consistent across both reader paths. Adding a filter now would change behavior at one of two readers and is YAGNI (spec.md Q5: minimal contract, additive later if a real consumer surfaces).

### D4 — No additional columns

`description`, `parent_code`, `owner_user_id`, `default_approver_role`, `archived_at`, `is_active`, `created_at` are deliberately omitted. F2.1c's contract consumes only `area_name`. Adding columns now would grow the published surface beyond its consumer's need and complicate future additive evolution.

### D5 — Security posture

Plain `CREATE VIEW` (no `security_invoker`) — identical to `v_active_user_areas` (0242), `v_cd_grantee` (0243), `v_document_search_facts` (0244), and `v_cd_obligated_readers` (0245). RLS over the base table continues to apply unchanged.

## Consequences

- The `distribution` module (F2.1c/F2.2) is ADR-0039 D3a compliant by construction — it reads only published views (`v_cd_obligated_readers`, `v_process_area_name`) + the ADR-0029 iam display-name port; `hgcrossmodule` analyzer holds.
- The taxonomy in-Go port (`AreaCatalogReader`) is unchanged — no risk to the existing B7 in-tx + B8 off-tx callers.
- The contract is **forward-compatible** for the parked `document-distribution-mission`: that mission may add columns to the consumer DTO without altering the published view shape.

## Verification

- Migration applies cleanly + idempotent on re-run (the `schema_migrations` ledger enforces this).
- `internal/modules/taxonomy/infrastructure/v_process_area_name_integration_test.go` asserts view shape + 1:1 per-tenant projection + cross-tenant isolation against a fixtured graph.
- `git diff -- internal/modules/taxonomy` consists only of the new test file (no runtime Go code).
- `go run ./tools/cilint/...` = 0 H-G under both `taxonomy` (publishes) and `distribution` (reads, once F2.2 lands).
