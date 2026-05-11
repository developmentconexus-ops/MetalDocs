# Phase 5 — Industry comparison (taxonomy)

Patterns drawn from `.claude/skills/metaldocs-module-doc/references/industry-patterns-index.md`. No new patterns added.

## Applicable

### IP-001 — RFC 9457 Problem Details (errors)

- Source: https://www.rfc-editor.org/rfc/rfc9457.html (RFC 9457, 2023-07) — "A problem details object can be extended with additional members."
- MetalDocs anchor: `internal/platform/httpresponse/response.go:14-16` defines `{code, message}` envelope; taxonomy returns this shape from every handler (`routes_families.go:98-115`, `routes_profiles.go:177-193`, `routes_areas.go` error helpers).
- Gap: legacy shape, not `application/problem+json`, no `type/title/status/detail/instance` fields. Sibling modules (auth, iam, approval, documents, audit) flagged the same gap; no module has migrated yet — cross-module ADR work, not taxonomy-local.
- Severity: Minor (codebase-wide consistency; module conforms to MetalDocs house style).

### IP-004 — Defense-in-depth authz (edge + in-tx + DB constraint)

- Source: NIST SP 800-95 §4.3 (2007) — "Multiple layers of access control reduce single-point bypass risk."
- MetalDocs anchor: `wiki/decisions/0007-two-tier-authz.md` documents tier-1 `CapabilityService` (HTTP) + tier-2 `authz.Require` (in-tx) + Postgres tripwire (`metaldocs.asserted_caps` GUC, `assert_caps()` trigger fn).
- Conformance check (Phase 4 §5): taxonomy is **single-tier**.
  - Tier-1 present: `apps/api/cmd/metaldocs-api/permissions.go:158-180` path-prefix dispatcher maps `/profiles*`, `/families*`, `/areas*` to `taxonomy.manage` / `doc.view`.
  - Tier-2 absent: `authz.Require` is not imported anywhere under `internal/modules/taxonomy/` (Phase 3 OUT-edge audit).
  - DB tripwire absent: no `assert_caps` trigger on any of the 3 owned tables; no `set_local_tenant_id` GUC propagation anywhere in `internal/` (Phase 4 §3).
- Comparison: approval (`0142b_role_capabilities_v2_enforce.sql:200-209`), iam, documents follow all three layers.
- Severity: Critical — taxonomy writes are protected by a single bypass-able layer; mutating ops on globally-shared `document_families` ⇒ cross-tenant blast.

### IP-006 — Forward-only migrations

- Source: https://martinfowler.com/articles/evodb.html (Fowler, 2016) — "Each change to the database is described by a migration script."
- MetalDocs anchor: 19 migrations in Phase 4 §6, all numbered, none edit in place. Code additions (tenant_id, archived_at, parent_code) added by ALTER in `0122` and `0123`, not by editing `0023`/`0025`.
- Conformance: **conformant**. No deviation flagged.

### IP-008 — Row-level tenant_id + scoped indexes

- Source: https://www.crunchydata.com/blog/designing-your-postgres-database-for-multi-tenancy (accessed 2026-05-10) — "Add tenant_id to every multi-tenant table and index it first."
- MetalDocs anchor:
  - `document_profiles.tenant_id` (UUID NOT NULL, `0122:4-6`) + `ux_document_profiles_tenant_code (tenant_id, code)` (`0122:21-22`) — conformant.
  - `document_process_areas.tenant_id` (`0123:3-5`) + `ux_process_areas_tenant_code` (`0123:19-20`) — conformant.
  - `document_families` — **no tenant_id by design** (`0023:1-7`). Global catalog, write GRANT to `metaldocs_app` (`0161:1`). No ADR justifies global scope. Phase 4 §5 race + cross-tenant blast both flow from this design without explicit acceptance.
- Gap: missing-ADR for "families are global" decision. Tenant_id absence on the table where mutation has the widest blast radius is the inverse of the pattern's intent.
- Severity: Critical (missing-ADR + cross-tenant blast).

## Not applicable

| id | Reason |
|---|---|
| IP-002 (Idempotency-Key) | Taxonomy POST endpoints have no `Idempotency-Key` handling; module is not on the idempotency roadmap (low write volume, manual catalog edits). Cross-module gap, not taxonomy-local. |
| IP-003 (cursor pagination) | All taxonomy list ops return full ordered slices (`family_repository.go:38`, `repository.go:42`, `repository.go:185`). Catalog cardinality expected to stay small (< 1k rows per tenant). Flag as latent risk in tech-debt, not a Critical/Major. |
| IP-005 (OpenAPI as source-of-truth) | Taxonomy uses raw `net/http` `ServeMux` (`handler.go:51-69`), no OpenAPI spec, no oapi-codegen stubs. Sibling modules (documents_v2, registry) ship `*.gen.go`. Gap: house pattern divergence. Severity: Major. |
| IP-007 (structured-log correlation) | Observability not wired in MetalDocs broadly; cross-module gap. Not taxonomy-local. |

## Citation summary

| Pattern | Verdict | Severity (if gap) |
|---|---|---|
| IP-001 RFC 9457 | gap (codebase-wide) | Minor |
| IP-004 defense-in-depth | gap | Critical |
| IP-005 OpenAPI codegen | gap (not in index for taxonomy; cited via N/A) | Major |
| IP-006 forward-only migrations | conformant | — |
| IP-008 row-level tenant_id | partial (families table excluded; no ADR) | Critical |
