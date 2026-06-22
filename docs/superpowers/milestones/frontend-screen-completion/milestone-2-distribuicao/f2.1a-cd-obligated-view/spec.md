# Feature F2.1a — Spec (cd-obligated-view)

> **Milestone:** 2 — Distribuição coverage-scope  ·  **Folder:** `f2.1a-cd-obligated-view`
> **Status:** Approved (pre-code) — 2026-06-21 (re-decomposition gate, operator).
> **Approved before code:** 2026-06-21 / operator (leandrotca) — *publish new sibling view; do not extend `v_cd_grantee`; ADR-0040 to record.*

> This is the feature's **contract**, written and approved **before any code**. It defines what the
> feature must do and how it will be proven — not how it will be built (that is `plan.md`). The
> milestone-validator judges the feature against *this* file (C1).

## Interview record (fail-closed gate)

Discovered via evidence-based subagent recon (commit body of the re-decomposition commit holds the
full report; key citations below). The genuine contract questions and how each resolved:

| # | Question | Answer |
|---|----------|--------|
| 1 | Why a new view instead of extending `metaldocs.v_cd_grantee`? | `0243.sql:22-24,65,72` + view COMMENT: the `WHERE visibility_scope='restricted'` gate is the **semantic contract** making search's EXISTS predicate (`reader.go:99-106`) correct-by-construction. Extending the view in place forces search to carry a distribution-domain `source` discriminator → module-boundary leak. Plus: zero DROP/ALTER VIEW precedent across 245 migrations; `wiki/database/migration-policy.md` forward-only. New sibling = clean. |
| 2 | What rows does distribution actually need that `v_cd_grantee` cannot give? | (a) `area_code` per row (which area granted access), (b) `source` discriminator (`user_grant`/`area_grant`/`company_scope`), (c) company-scope rows (`v_cd_grantee` is restricted-only by design). |
| 3 | View name? | `metaldocs.v_cd_obligated_readers` — publisher prefix `v_cd_*` matches existing convention (`v_cd_search_facts`, `v_cd_grantee`); `obligated_readers` is shape-qualified (the *obligated* reader set, not just `grantee` which already means restricted-edges). |
| 4 | Source-precedence when a user appears via multiple legs (e.g. direct user-grant AND active area-grant)? | DISTINCT BY `user_id` per `(tenant_id, cd_id)`. Precedence `user_grant` > `area_grant` > `company_scope` (most-specific wins). For `area_grant` rows on multi-area users, deterministically choose the lowest `area_code` (matches the spec's existing distinct rule). Encoded in the view body via `SELECT DISTINCT ON (tenant_id, cd_id, user_id) … ORDER BY tenant_id, cd_id, user_id, source_rank, area_code`. |
| 5 | Where does the company-scope leg get its user set? | Cross-join of `v_cd_search_facts` (where `is_company=true`) × the per-tenant active user set. The active user set itself is reachable via `metaldocs.v_active_user_areas` (DISTINCT `user_id`) — that view already encodes "active membership" (effective_to IS NULL) per ADR-0037 D1, so it doubles as the canonical "active tenant user" projection. If a tenant user has zero area memberships, they would not appear; that matches the IAM definition of "active" (no active areas → not an active tenant participant). |
| 6 | Does the view need `area_name`? | **No.** `area_name` is taxonomy-owned and is published separately by F2.1b (`v_process_area_name`). Distribution joins the two views in F2.2. Keeps each owner's published contract minimal. |
| 7 | Forward-only? Is `CREATE VIEW … IF NOT EXISTS` available? | PG 14+: `CREATE OR REPLACE VIEW` is available, but the policy is **forward-only additive new view** — use plain `CREATE VIEW` in a new migration; idempotency comes from the migration runner's `schema_migrations` ledger, not from view DDL. |
| 8 | ADR-0039 inventory update required? | Yes — add a row: view `v_cd_obligated_readers`, owner `controlleddocuments`, reader `distribution` (M2/F2.1a). |

## Consumer contract (FIRST — before any producer)

- **Consumer:** the new `internal/modules/distribution` read module (built in F2.2), which composes
  the obligated-reader set for the three `/api/v1/documents/:id/distribution*` endpoints declared in
  F2.1c. Distribution must read **only this view + F2.1b's `v_process_area_name` + the ADR-0029 iam
  display-name read-port** — never CD/taxonomy/iam base tables (ADR-0039 / `hgcrossmodule`).
- **Contract (the new published view):**

  ```
  metaldocs.v_cd_obligated_readers (
    tenant_id              uuid     not null,
    controlled_document_id uuid     not null,
    user_id                uuid     not null,
    area_code              text         null,  -- null when source = 'user_grant' or 'company_scope'
    source                 text     not null   -- enum string: 'user_grant' | 'area_grant' | 'company_scope'
  )
  -- DISTINCT BY (tenant_id, controlled_document_id, user_id)
  -- Source precedence: user_grant > area_grant > company_scope
  -- For area_grant rows on a user with multiple granting areas: lowest area_code wins (deterministic).
  ```

  **Three legs (UNION):**
  1. **user-grant leg** — `public.controlled_document_user_grants` joined to its CD; `source='user_grant'`,
     `area_code=NULL`.
  2. **area-grant leg** — `public.controlled_document_area_grants` ⋈ `metaldocs.v_active_user_areas`
     on `(tenant_id, area_code)`; `source='area_grant'`, `area_code=upa.area_code`.
  3. **company-scope leg** — `metaldocs.v_cd_search_facts` where `is_company=true`, cross-joined to
     DISTINCT `user_id` from `metaldocs.v_active_user_areas` per tenant; `source='company_scope'`,
     `area_code=NULL`.

  The view body applies the DISTINCT ON rule above to enforce one row per user per CD with the
  precedence-winning leg's `source`/`area_code`.

- **Source of truth for the contract:** the new migration file
  `db/migrations/0245_cd_obligated_readers_view.sql` (owner: `controlleddocuments`). ADR-0040 documents
  the decision; ADR-0039 inventory adds the row.

## What this feature implements

A single forward-only migration creating `metaldocs.v_cd_obligated_readers`, plus ADR-0040 + the
ADR-0039 inventory row. **No Go code.** F2.1c declares the OpenAPI; F2.2 implements the handlers that
read this view.

## Non-goals (mandatory)

- **No modification of `metaldocs.v_cd_grantee`** (search's contract; mutating it is HS-2).
- **No change to any base table** in `public.*`.
- **No reverse migration** (forward-only per policy).
- **No `area_name` column** on this view (taxonomy-owned; F2.1b publishes it).
- **No `display_name`/`name`** column (iam-owned; ADR-0029 port resolves it in F2.2).
- **No index on the view** (views can't carry indexes; if F2.2's integration test surfaces a real
  latency problem, that's a separate decision, not this feature).
- **No Go code, no handler, no OpenAPI op** — those are F2.1c/F2.2.

## Validation Gate (concrete — approved before code)

| Acceptance criterion | Named test / proof command | Real vs fixture |
|----------------------|----------------------------|-----------------|
| Migration applies cleanly on a fresh DB and is idempotent on re-run | `.\scripts\start-api.ps1 -Build` runs all migrations green; `SELECT version FROM public.schema_migrations WHERE version='0245'` returns one row | real |
| View shape matches the contract | `\d+ metaldocs.v_cd_obligated_readers` shows exactly 5 columns with the declared types + nullability | real |
| Three-leg semantics correct + distinct-with-precedence | A new integration test seeded with: 1 restricted CD with a direct user-grant for U1 + area-grant on area A1 (member U2) + same user U1 also member of A1 (overlap), AND 1 company-scope CD in the same tenant with active users {U1,U2,U3} → expect 3 distinct user rows for restricted CD with U1.source='user_grant'/area_code=NULL, U2.source='area_grant'/area_code='A1', and 3 distinct user rows for company-scope CD all with source='company_scope'/area_code=NULL. Test file: `internal/modules/controlleddocuments/.../v_cd_obligated_readers_integration_test.go` | real (live PG) |
| Search untouched | `git diff db/migrations/0243*` = empty; `git diff internal/modules/search` = empty | real |
| ADR-0040 recorded + ADR-0039 inventory updated | `ls wiki/decisions/0040-*.md`; `grep -n "v_cd_obligated_readers" wiki/decisions/0039-cross-module-base-table-read-boundary.md` ≥ 1 hit | real |
| `hgcrossmodule` analyzer green (the view publishes from CD's own base tables + iam's published view — compliant per ADR-0039 D3a) | `go run ./tools/cilint/...` over distribution + cd modules = 0 H-G | real |

## ADR needed?

- [x] **Durable decision made** → ADR-0040 (controlleddocuments publishes `v_cd_obligated_readers`):
  records (a) the new view + its semantics, (b) why a new sibling vs extending `v_cd_grantee`
  (search-semantic contract preservation + zero DROP/ALTER precedent), (c) the source-precedence rule,
  (d) the company-scope leg's active-user definition, (e) inventory row added to ADR-0039.
