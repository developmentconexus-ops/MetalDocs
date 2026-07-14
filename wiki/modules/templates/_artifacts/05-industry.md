# Phase 5 — Industry comparison · templates

> Date: 2026-05-10
> Composer: main agent (Opus 4.7)
> Source-of-truth: `.claude/skills/metaldocs-module-doc/references/industry-patterns-index.md`

Each citation maps a pre-vetted pattern to a specific MetalDocs file:line where the gap (or alignment) lives. No new patterns added this session.

## IP-001 — RFC 9457 Problem Details

> "A problem details object can be extended with additional members." — RFC 9457 (2023-07)

**Applies to:** `internal/modules/templates/delivery/http/handler.go:95-102` (`writeErr`).

**Status: NOT applied.** templates emits the legacy MetalDocs envelope `{"error":{"code":"...","message":"..."}}`. Plan 2 (`wiki/architecture/api-design-system.md`, commits ae1229e8..c84215f7) introduced `internal/platform/problem` for RFC 9457; templates has not been migrated. Same shape as `documents` T-001 and `auth` T-003 — module-by-module rollout debt.

## IP-002 — Idempotency-Key

> "Keys are eligible to be removed from the system after they're at least 24 hours old." — Stripe API docs (accessed 2026-05-10)

**Applies to:** `internal/modules/templates/application/lifecycle.go:265` (`PublishTemplateVersion`), `application/create.go:30` (`CreateTemplate`).

**Status: NOT applied.** No `Idempotency-Key` header parsing in templates routes; no lookup against `internal/platform/idempotency/`. Replays of `POST /api/v1/templates` and `POST /publish` either succeed twice (creating duplicate audit rows) or fail with `ErrInvalidStateTransition` after the first transition lands. Compare: `documents` module §6.6 (per `wiki/modules/documents.md`) records the same gap as T-006.

## IP-003 — Cursor pagination

> "Pagination should be done with a forward-only cursor." — Relay Connections (2021)

**Applies to:** `repository/postgres.go:88` (`ListTemplates`) — currently returns full filtered set with no LIMIT/OFFSET or cursor primitive.

**Status: NOT applied.** Result-set is unbounded by tenant. Plan 2 cursor primitive (`feat(pagination): cursor primitive with sort + filter_hash validation`, commit 7effa430) is wired in approval/inbox path but not consumed here.

## IP-004 — Defense-in-depth authz

> "Multiple layers of access control reduce single-point bypass risk." — NIST SP 800-95 (2007)

**Applies to:** `delivery/http/handler.go:24-29` (constructor + nil-authz fallback) and the entire `repository/postgres.go` mutation surface.

**Status: NOT applied — single tier missing entirely.** The architecture (`wiki/decisions/0007-two-tier-authz.md`) specifies tier-1 (`CapabilityService.CanDo` middleware) + tier-2 (`authz.Require` in-tx) + Postgres tripwire. templates has zero of the three:
1. Tier-1: `AuthzFunc` arg accepted but `apps/api/cmd/metaldocs-api/main.go:329` passes `nil`; constructor falls through to no-op (handler.go:25-27).
2. Tier-2: no `internal/platform/authz.Require` import or call in `application/**`.
3. Tripwire: no `metaldocs.asserted_caps` GUC; no SQL trigger checking it on INSERT/UPDATE/DELETE of `templates_*` tables (per Phase 4 §3 — zero triggers/GUCs/functions installed).

Capabilities `template.view/create/edit/submit/approve/publish` are seeded in `migrations/0165_role_capabilities_reseed.sql` but never asserted. The seed-without-enforce shape is identical to the iam T-001 dual-namespace debt.

## IP-005 — OpenAPI as source-of-truth

> "The OpenAPI Specification … is the standard for HTTP APIs." — OAI 3.0.3 (2020)

**Applies to:** `internal/modules/templates/api/api.gen.go` (oapi-codegen output) + hand-rolled handlers in `delivery/http/handler.go:48-61`.

**Status: PARTIAL.** Eight routes are generated (`api.gen.go:954-961`). Twelve hand-rolled routes (handler.go:48-61) are NOT in the spec — `POST /api/v1/templates/{id}/versions`, `PUT /api/v1/templates/{id}/versions/{n}/schema`, `POST /api/v1/templates/{id}/versions/{n}/{submit,review,approve}`, `GET /api/v1/templates/{id}/audit`, `GET /api/v1/templates/v2/placeholder-catalog`, etc. Same drift class as `documents` T-002 (spec/handler drift Critical).

## IP-006 — Forward-only migrations

> "Each change to the database is described by a migration script." — Fowler (2016)

**Applies to:** `migrations/0101..0165` (templates lineage).

**Status: APPLIED.** All schema changes are forward-only files in `migrations/`. The `0109_docx_v2_templates_w2_noop.sql` placeholder maintains numbering integrity. `0157_drop_editable_zones.sql` (per Phase 4 §6) appears to be a deferred drop that did not run (column still present in current DDL inheritance) — flag as drift to verify in Phase 6.

## IP-007 — Observability correlation id

> "Correlate spans across boundaries with a single id." — accessed 2026-05-10

**Applies to:** n/a — observability stack not yet wired in MetalDocs (per index note). No suggestion added; absence is repo-wide, not templates-specific.

## IP-008 — Multi-tenant row-level tenant_id

> "Add tenant_id to every multi-tenant table and index it first." — Crunchy Data (accessed 2026-05-10)

**Applies to:** `migrations/0120_templates_init.sql` schema.

**Status: PARTIAL.** `templates_template` and `templates_audit_log` carry `tenant_id NOT NULL`. `templates_template_version` and `templates_approval_config` do NOT carry `tenant_id` — they inherit tenant scope only via FK to `templates_template`. Repo methods `GetVersion(template_id, version_number)` and `GetVersionByID(version_id)` (`postgres.go`) accept no tenant arg — cross-tenant version-id lookups will succeed if the version_id is known (no constraint to prevent it). Compare IP-008 quote: "Add tenant_id to every multi-tenant table." — gap.

Tenant ID source itself (`handler.go:84-89`) trusts the `X-Tenant-ID` header with `tenant.DevTenantID` fallback — verifiable forgery vector in any non-dev environment.

## Summary

| Pattern | Status |
|---|---|
| IP-001 RFC 9457 envelope | NOT applied |
| IP-002 Idempotency-Key | NOT applied |
| IP-003 Cursor pagination | NOT applied |
| IP-004 Defense-in-depth authz | NOT applied — three-tier zero |
| IP-005 OpenAPI source-of-truth | PARTIAL — 12 of 20 routes off-spec |
| IP-006 Forward-only migrations | APPLIED |
| IP-007 Observability | n/a |
| IP-008 Tenant id every table | PARTIAL — 2 of 4 tables, header-trusted source |
