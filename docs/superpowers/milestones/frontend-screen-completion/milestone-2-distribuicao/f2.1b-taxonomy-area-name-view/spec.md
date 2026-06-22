# Feature F2.1b — Spec (taxonomy-area-name-view)

> **Milestone:** 2 — Distribuição coverage-scope  ·  **Folder:** `f2.1b-taxonomy-area-name-view`
> **Status:** Approved (pre-code) — 2026-06-21 (re-decomposition gate, operator).
> **Approved before code:** 2026-06-21 / operator (leandrotca) — *publish minimal area-name view from taxonomy; ADR-0041 to record.*

> This is the feature's **contract**, written and approved **before any code**. It defines what the
> feature must do and how it will be proven — not how it will be built (that is `plan.md`). The
> milestone-validator judges the feature against *this* file (C1).

## Interview record (fail-closed gate)

| # | Question | Answer |
|---|----------|--------|
| 1 | Why does taxonomy need to publish anything? | F2.1c's `DistributionRecipient` and `DistributionAreaCoverage` schemas carry `area_name` (human label). The base table holding it is taxonomy-owned (`document_process_areas` or equivalent — confirm during plan.md recon). ADR-0039 / `hgcrossmodule` forbids the new distribution module from reading taxonomy's base table raw → taxonomy must publish a read contract. |
| 2 | Why a new view vs reusing something existing? | Recon (subagent report, this session) enumerated all `metaldocs.v_*` views: none expose area_name. No prior published taxonomy view exists. |
| 3 | View name + shape? | `metaldocs.v_process_area_name (tenant_id uuid, area_code text, area_name text)`. Minimal — just the human label keyed by `(tenant_id, area_code)`. Naming follows `v_<publisher_domain>_<shape>` convention. |
| 4 | One row per `(tenant_id, area_code)`? | Yes. Process areas are unique by `(tenant_id, area_code)` in the base table; the view is a 1:1 projection. |
| 5 | Include `is_active` or any other taxonomy attribute? | **No.** YAGNI — F2.1c only consumes `area_name` for labels. Adding columns now would be speculative + grow the published contract beyond its consumer's need. Additive later if a real consumer surfaces. |
| 6 | Forward-only? | Yes — plain `CREATE VIEW` in a new migration (same posture as F2.1a). |
| 7 | ADR-0039 inventory update? | Yes — add a row: view `v_process_area_name`, owner `taxonomy`, reader `distribution` (M2/F2.1b). |
| 8 | Owner-module location for the migration? | `db/migrations/0246_taxonomy_process_area_name_view.sql` (sequential after F2.1a's 0245). The migration is owned by the taxonomy module conceptually; the file lives in the shared `db/migrations/` per repo policy. |

## Consumer contract (FIRST — before any producer)

- **Consumer:** the new `internal/modules/distribution` read module (F2.2). Distribution joins this
  view to F2.1a's `v_cd_obligated_readers` on `(tenant_id, area_code)` to populate the
  `DistributionRecipient.area_name` and `DistributionAreaCoverage.area_name` fields in F2.1c's
  contract.
- **Contract (the new published view):**

  ```
  metaldocs.v_process_area_name (
    tenant_id  uuid not null,
    area_code  text not null,
    area_name  text not null
  )
  -- One row per (tenant_id, area_code). 1:1 projection over taxonomy's process-area base table.
  ```

- **Source of truth for the contract:** the new migration file
  `db/migrations/0246_taxonomy_process_area_name_view.sql` (owner: `taxonomy`). ADR-0041 documents the
  decision; ADR-0039 inventory adds the row.

## What this feature implements

A single forward-only migration creating `metaldocs.v_process_area_name`, plus ADR-0041 + the
ADR-0039 inventory row. **No Go code.**

## Non-goals (mandatory)

- **No additional taxonomy columns** (`is_active`, `parent_code`, `description`, etc.) — minimal
  contract, additive later if a consumer needs them.
- **No change to any taxonomy base table.**
- **No reverse migration** (forward-only).
- **No Go code** — F2.2 is the only consumer.

## Validation Gate (concrete — approved before code)

| Acceptance criterion | Named test / proof command | Real vs fixture |
|----------------------|----------------------------|-----------------|
| Migration applies cleanly + idempotent | `.\scripts\start-api.ps1 -Build`; `SELECT version FROM public.schema_migrations WHERE version='0246'` returns one row | real |
| View shape matches the contract | `\d+ metaldocs.v_process_area_name` shows exactly 3 columns with the declared types + nullability | real |
| 1:1 row count with the taxonomy base table per tenant | Integration test seeds N process areas in one tenant, asserts `SELECT count(*) FROM metaldocs.v_process_area_name WHERE tenant_id=$1` = N. Test file: `internal/modules/taxonomy/.../v_process_area_name_integration_test.go` | real (live PG) |
| ADR-0041 recorded + ADR-0039 inventory updated | `ls wiki/decisions/0041-*.md`; `grep -n "v_process_area_name" wiki/decisions/0039-cross-module-base-table-read-boundary.md` ≥ 1 hit | real |
| `hgcrossmodule` analyzer green | `go run ./tools/cilint/...` over taxonomy module = 0 H-G | real |

## ADR needed?

- [x] **Durable decision made** → ADR-0041 (taxonomy publishes `v_process_area_name`): records the
  minimal published shape + the rationale for excluding additional columns now + the inventory row
  added to ADR-0039.
