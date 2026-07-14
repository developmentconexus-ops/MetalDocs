# ADR 0040 — `controlleddocuments` publishes `metaldocs.v_cd_obligated_readers` obligated-reader view

> **Status:** Accepted 2026-06-21
> **Last verified:** 2026-06-21
> **Deciders:** leandrotca.work (operator), MetalDocs backend
> **Context window:** Mission `frontend-screen-completion` · Milestone M2 (Distribuição) · Feature F2.1a.
> **Supersedes:** none.
> **Related ADRs:** [0039 — Cross-module read boundary (H-G)](./0039-cross-module-base-table-read-boundary.md) (D3a/D4 published-view exemption — this view is added to the inventory); [0037 — Membership temporal model](./0037-membership-temporal-model.md) (D1: active-now ⟺ `effective_to IS NULL` — inherited via `v_active_user_areas`); [0029 — `UserDisplayNameReader`](./0029-user-display-name-reader-port.md) (iam-owned display-name read-port — this view deliberately omits display names).
> **Related code (Last verified 2026-06-21):**
> - `db/migrations/0245_cd_obligated_readers_view.sql` — publishes `metaldocs.v_cd_obligated_readers` (this ADR).
> - `internal/modules/controlleddocuments/infrastructure/v_cd_obligated_readers_integration_test.go` — integration test pinning the three-leg union, source precedence, deterministic area_code tiebreak, and the info_schema `is_nullable=YES` note.
> - `db/migrations/0243_cd_search_visibility_contract.sql` — defines `metaldocs.v_cd_grantee` with `WHERE visibility_scope='restricted'` (the search-semantic contract that makes `internal/modules/search/infrastructure/v2documents/reader.go`'s EXISTS predicate correct-by-construction; this ADR explicitly does **not** mutate it).
> - `db/migrations/0242_iam_v_active_user_areas_view.sql` — `metaldocs.v_active_user_areas` (iam-published active-membership view; ADR 0037 D1).
> - `wiki/database/migration-policy.md` — forward-only migration policy (this ADR's D5 rationale).
> - Spec: `docs/superpowers/milestones/frontend-screen-completion/milestone-2-distribuicao/f2.1a-cd-obligated-view/spec.md`.

## Context

Milestone M2 (Distribuição) of the `frontend-screen-completion` mission builds a `distribution` module that
exposes per-CD coverage on the Distribuição screen. Coverage is `read / obligated` — the denominator is the
**set of users obligated to read a given CD**, derived from three independent rules:

1. Direct user grant (`controlled_document_user_grants`).
2. Area grant (`controlled_document_area_grants`) joined to active area membership (`v_active_user_areas`).
3. Company-scope CDs × every active tenant user.

ADR-0039 forbids non-owner modules from reading CD or iam base tables (D1 — H-G violation). The distribution
module is a non-owner of CD and iam, so it cannot reach into `controlled_document_*_grants` /
`user_process_areas` directly; it must consume an owner-published view (D3a) or read-port (D3b).

CD already publishes `metaldocs.v_cd_grantee` (migration 0243), but that view is **restricted-only by
design**: the `WHERE visibility_scope='restricted'` gate is the *search-semantic contract* making
`internal/modules/search/infrastructure/v2documents/reader.go`'s EXISTS predicate
correct-by-construction. Mutating that view to also serve distribution would force search to carry
distribution-domain knowledge — a module-boundary leak in the opposite direction.

Two further facts pin the shape:

- **Zero `DROP VIEW` / `CREATE OR REPLACE VIEW` precedent across 244 prior migrations.**
  `wiki/database/migration-policy.md` is forward-only; the migration runner provides idempotency via
  `schema_migrations`, not view DDL.
- **PostgreSQL `information_schema.columns.is_nullable` is conservatively `YES` for every column of a
  `UNION ALL` view** regardless of upstream `NOT NULL` constraints. Runtime non-nullability is enforced by
  base-table constraints, not view metadata.

## Decision

1. **Publish a new sibling view `metaldocs.v_cd_obligated_readers`** (do **not** extend `v_cd_grantee`).
   The new view carries the three-rule obligated-reader semantics; `v_cd_grantee` keeps its restricted-only
   search semantics untouched.

2. **Three legs `UNION ALL`** in the view body:
   - (a) `controlled_document_user_grants` → `source='user_grant'`, `area_code=NULL`.
   - (b) `controlled_document_area_grants ⋈ metaldocs.v_active_user_areas` on `(tenant_id, area_code)`
     → `source='area_grant'`, `area_code=upa.area_code`.
   - (c) `metaldocs.v_cd_search_facts WHERE is_company` × `DISTINCT (tenant_id, user_id) FROM
     metaldocs.v_active_user_areas` → `source='company_scope'`, `area_code=NULL`.
     The company-scope leg consumes CD's own `v_cd_search_facts.is_company` (1:1 over
     `controlled_documents`) instead of a hardcoded scope literal.

3. **`DISTINCT BY (tenant_id, controlled_document_id, user_id)`** with source precedence
   `user_grant > area_grant > company_scope` (most-specific wins). For area-grant rows on a user with
   multiple granting areas, the **lowest `area_code`** wins (deterministic). Encoded as:

   ```sql
   SELECT DISTINCT ON (tenant_id, controlled_document_id, user_id)
          tenant_id, controlled_document_id, user_id, area_code, source
     FROM legs
    ORDER BY tenant_id, controlled_document_id, user_id, source_rank, area_code NULLS LAST;
   ```

4. **No `area_name` / `display_name` columns on this view.** Taxonomy and iam own those surfaces:
   F2.1b publishes `metaldocs.v_process_area_name` (taxonomy-owned area label), and ADR-0029's
   `UserDisplayNameReader` read-port supplies user display names. Embedding either here would re-introduce
   the same H-G class this view exists to close.

5. **Forward-only DDL.** The view is created with `CREATE VIEW` (not `CREATE OR REPLACE VIEW`); idempotency
   is the migration runner's `schema_migrations` ledger, consistent with the project's forward-only
   migration policy. Decided at the F2.1a spec interview, Q7.

## Consequences

### Positive
- The distribution module reads **only**: `metaldocs.v_cd_obligated_readers` (this view) +
  F2.1b's `metaldocs.v_process_area_name` + ADR-0029's iam display-name read-port. It never reads CD,
  taxonomy, or iam base tables. H-G stays at 0 under both ADR-0039 readings.
- `metaldocs.v_cd_grantee` is preserved verbatim — search semantics remain correct-by-construction; the
  restricted-only gate is not perturbed.
- The active-user definition is **inherited** from `v_active_user_areas` (ADR-0037 D1: `effective_to IS
  NULL`). The company-scope leg therefore agrees with every other active-now read in the system; a tenant
  user with zero active area memberships is not in the company-scope leg by design (the operator-confirmed
  behavior, since "active tenant user" is encoded through area membership in this schema).
- The company-scope leg consumes `v_cd_search_facts.is_company` rather than a string literal, so the
  scope vocabulary lives in one place (CD-owned).

### Negative / cost
- One additional published view in the `metaldocs.*` namespace. Bounded by the published-view inventory
  in ADR-0039 (now adds one row).
- `information_schema.columns.is_nullable` is `YES` for every column on this view (conservative
  PostgreSQL default for UNION-ALL views) regardless of upstream `NOT NULL`s. Runtime non-nullability is
  enforced by base-table constraints, not view metadata. The F2.1a integration test documents this
  explicitly so a reader does not mistake the metadata for a contract.

### Neutral
- No behavior change to any existing query. The view is additive; nothing reads it yet — the F2.1c
  distribution-module read path (built later in M2) is the first consumer.

## Verification

- Migration `db/migrations/0245_cd_obligated_readers_view.sql` is applied via the runner; the
  `schema_migrations` row at version `0245` is the idempotency boundary.
- The integration test
  `internal/modules/controlleddocuments/infrastructure/v_cd_obligated_readers_integration_test.go`
  pins the three-leg union semantics, the source precedence (`user_grant > area_grant > company_scope`),
  the deterministic lowest-`area_code` tiebreak for multi-area users, and the `is_nullable=YES`
  metadata note.
- ADR-0039's published-view inventory is updated in the same commit to add this view (D3a/D4
  exemption row).

## References

- ADR-0039 — Cross-module read boundary (H-G). D3a (published-view exemption) + D4 (active-now
  membership-view contract). Inventory updated alongside this ADR.
- ADR-0037 D1 — `effective_to IS NULL` = active-now membership (inherited via `v_active_user_areas`).
- ADR-0029 — iam-owned `UserDisplayNameReader` read-port (the path distribution uses for display names;
  this view deliberately omits them).
- Migration: `db/migrations/0245_cd_obligated_readers_view.sql`.
- Integration test: `internal/modules/controlleddocuments/infrastructure/v_cd_obligated_readers_integration_test.go`.
- Spec: `docs/superpowers/milestones/frontend-screen-completion/milestone-2-distribuicao/f2.1a-cd-obligated-view/spec.md`.
- Mission / milestone: `docs/superpowers/milestones/frontend-screen-completion/` — M2 (Distribuição) / F2.1a.
