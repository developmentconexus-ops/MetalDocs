# Feature F4.1 — CD search visibility + projection read contract

> **Milestone:** 4 — search consumes published visibility contracts  ·  **Folder:** `f4.1-cd-visibility-contract`
> **Status:** Closed (2026-06-21 — parity GREEN on PG :5434; evidence.md written)

## Source

- Milestone spec row F4.1: *Implement* — controlleddocuments publishes a versioned visibility read
  contract (view) for the search consumer (company/restricted/owner/area-grant/user-grant), sourced
  from CD's own grant tables joined to iam's `v_active_user_areas` (D3a). Covers C4b/C4c/C4e, removes
  the need for search to inline CD's `visibility_scope` literals. *Validate* — migration applies; parity
  query proves view decision == inline-predicate decision across all scopes incl. revoked + ungranted;
  ADR-0039 references the views.
- Governing-spec reference: `mission.md` §7 M4/F4.1; ADR-0039 D3a/D4. Contract shape = Shape 2
  (operator-ratified, see `spec.md` interview).

## Plan

**Contract (from `spec.md`):** two `metaldocs` views.
- `v_cd_search_facts(tenant_id, controlled_document_id, code, department_code, profile_code, sequence_num, is_company, owner_user_id)` — 1 row per CD, over `public.controlled_documents` only (own table).
- `v_cd_grantee(tenant_id, controlled_document_id, grantee_user_id)` — bounded edges: area-grant members (via `controlled_document_area_grants` ⋈ `metaldocs.v_active_user_areas`) UNION direct `controlled_document_user_grants`. Restricted CDs only by construction (grants only exist on restricted CDs; the area/user grant tables are the source).

**TDD order:**
1. **RED** — write `internal/modules/controlleddocuments/infrastructure/cd_visibility_contract_parity_integration_test.go` (`//go:build integration`). Reuse the M3 `seedCDVisibility` scenario shape (company + restricted CDs; owner/areaMember/revokedMem/userGrant/none actors). Assertions:
   - facts-view: 1 row per CD; `is_company`/`owner_user_id`/projection cols == base-table values.
   - grantee-view: set for restricted CD == {areaMember, userGrant}; revokedMem ∉, none ∉.
   - composed decision `is_company OR owner=$13 OR EXISTS(grantee=$13)` == verbatim raw copy of `reader.go:89-118` predicate, for all 5 actors × {company, restricted}.
   Run → FAIL (views don't exist).
2. **GREEN** — write `db/migrations/0243_cd_search_visibility_contract.sql` creating both views + `COMMENT ON VIEW` + `schema_migrations` insert. Rerun integration test → PASS.
3. Extend ADR-0039 `Related code` note: add migration 0243 + the two view names (mirrors the M3/F3.1 `v_active_user_areas` note).

**Files touched:**
- NEW `db/migrations/0243_cd_search_visibility_contract.sql`
- NEW `internal/modules/controlleddocuments/infrastructure/cd_visibility_contract_parity_integration_test.go`
- EDIT `wiki/decisions/0039-cross-module-base-table-read-boundary.md` (Related code note only)

**Test strategy:** integration parity on real PG :5434 (`METALDOCS_DATABASE_URL=postgres://metaldocs:metaldocs@127.0.0.1:5434/metaldocs?sslmode=disable`, `-tags integration`). The revoked-member + ungranted-user rows are the anti-drift discriminators. No unit/Go production code in F4.1 (views are pure SQL; search consume is F4.3).

**Non-goals (guardrails):** no search change; no documents view (F4.2); no change to CD's own repository; no (cd×actor) cross-product; no new authz semantics.

## Execution notes

- Implemented directly in main session (contained: 1 migration + 1 integration test). Parity test is
  the correctness guard; no subagent dispatch needed.
- Live test DSN resolved during HS-3: `metaldocs-test-pg` container (was Exited 255; started). Use
  `127.0.0.1:5434` (IPv4) — `localhost` resolved `::1` first earlier.
