# Phase 5 — Industry Comparison (iam)

> Patterns drawn from `.claude/skills/metaldocs-module-doc/references/industry-patterns-index.md`. One pattern per row; verbatim quote ≤ 30 words; anchored to a MetalDocs file:line.

## Admitted patterns

### IP-004 · Defense-in-depth authz (edge + in-tx + DB constraint)

- **Source:** NIST SP 800-95 §4.3 (2007) — `https://csrc.nist.gov/publications/detail/sp/800-95/final`
- **Accessed:** 2026-05-10
- **Quote:** "Multiple layers of access control reduce single-point bypass risk."
- **Maps to:**
  - Tier 1 (edge): `internal/modules/iam/application/capability_service.go:31` — `(*CapabilityService).CanDo`
  - Tier 2 (in-tx): `internal/modules/iam/authz/authz.go:44` — `Require(ctx, tx, capability, areaCode)`
  - DB tripwire: `migrations/0142b_role_capabilities_v2_enforce.sql:67-179` — `enforce_capability_asserted()`
- **MetalDocs adherence — partial:** IAM defines all three layers, but the tripwire trigger is attached ONLY to `public.approval_instances` and `public.approval_signoffs` (`0142b:200-209`). IAM's OWN mutating tables (`iam_user_roles`, `user_process_areas`, `iam_users`, `iam_groups*`) sit behind tier-1 middleware alone — no tier-2 `authz.Require` call in `RoleAdminRepository.UpsertUserAndAssignRole`/`ReplaceUserRoles`, `UserAreaRepository.GrantAtomic`, `area_membership.Grant/Revoke`. Persistence audit (artifact 04 §5) lists 0 violations because no trigger exists on those tables; from a defense-in-depth view this is a documented gap, not a bug.

### IP-008 · Row-level `tenant_id` + scoped indexes

- **Source:** `https://www.crunchydata.com/blog/designing-your-postgres-database-for-multi-tenancy`
- **Accessed:** 2026-05-10
- **Quote:** "Add tenant_id to every multi-tenant table and index it first."
- **Maps to:**
  - `metaldocs.iam_users.tenant_id` — migration 0130; UNIQUE (`tenant_id, user_id`) — `ux_iam_users_tenant_user`
  - `metaldocs.iam_user_roles.tenant_id` — migration 0162; UNIQUE (`tenant_id, user_id`) constraint added by 0166
  - `metaldocs.iam_groups.tenant_id` — migration 0163; UNIQUE (`tenant_id, name`)
  - `public.user_process_areas.tenant_id` — migration 0125; UNIQUE active-row (`tenant_id, user_id, area_code, role`) — `ux_user_process_areas_single_active`
- **MetalDocs adherence — full:** every IAM-owned table carries `tenant_id` and at least one tenant-scoped unique index. `RolesByUserID` (`infrastructure/postgres/role_provider.go:19`) and `RoleAdminRepository.*` (`role_admin_repository.go:20,33,72`) all filter by `tenant_id` after Group B fix (2026-05-03). Sentinel `DevTenantID = ffffffff-…` (`internal/platform/tenant/const.go:4`) covers single-tenant dev path.

### IP-001 · RFC 9457 Problem Details JSON envelope

- **Source:** RFC 9457 — `https://www.rfc-editor.org/rfc/rfc9457.html`
- **Version:** 2023-07
- **Quote:** "A problem details object can be extended with additional members."
- **Maps to:** `api/openapi/v1/openapi.yaml` — `Problem` schema (referenced from `wiki/architecture/api-design-system.md`)
- **MetalDocs adherence — gap:** IAM middleware emits `{error:{code,message,details,trace_id}}` (`internal/modules/iam/delivery/http/middleware.go:129`); membership handler emits `{code,message}` (`routes_memberships.go:137`). Neither matches RFC 9457 `type` / `title` / `status` / `detail` shape. No `Problem` `type` URI on IAM responses (artifact 02-flow-list-memberships §5; 02-flow-grant-membership §5). `wiki/architecture/api-design-system.md` names RFC 9457 as the canonical envelope — IAM is not on it yet.

### IP-005 · OpenAPI as source-of-truth (oapi-codegen)

- **Source:** OAI Best Practices — `https://learn.openapis.org/best-practices.html`
- **Version:** OpenAPI 3.0.3 (2020)
- **Quote:** "The OpenAPI Specification … is the standard for HTTP APIs."
- **Maps to:** `api/openapi/v1/openapi.yaml`, generated `*.gen.go` per `wiki/architecture/api-contract.md`
- **MetalDocs adherence — gap:** IAM routes are wired on a hand-written `*http.ServeMux` (`delivery/http/routes_memberships.go:30`, `delivery/http/admin_handler.go:82`). `listAreaMemberships`, `grantAreaMembership`, `revokeAreaMembership` are NOT declared in `openapi.yaml` (artifact 02 §1 across all 3 traces). Admin POST `/api/v1/iam/users/{userId}/roles` has request + response schemas in `openapi.yaml:5043,5054` but no `operationId` and no `*.gen.go` stub — partial contract-first. Documents module bootstrap is codegen-enabled per ADR 0012; IAM is not.

### IP-006 · Forward-only migrations

- **Source:** Fowler — `https://martinfowler.com/articles/evodb.html`
- **Version:** Fowler (2016)
- **Quote:** "Each change to the database is described by a migration script."
- **Maps to:** `migrations/` — 17 IAM-affecting migrations from `0002_init_iam_rbac.sql` through `0170_dev_approver_role_correction.sql` (artifact 04 §6)
- **MetalDocs adherence — full:** every IAM change ships as a numbered `migrations/0NNN_*.sql` file. 0166 (rename) used `UPDATE`/`DELETE` rather than DROP-RECREATE. 0170 corrects a 0166 over-rename without rewinding 0166. Aligns with the index pattern verbatim.

## Not-applicable (admitted but rejected for this module)

- **IP-002 idempotency** — not-applicable: IAM admin role upsert is naturally idempotent at the SQL layer (DELETE-INSERT under unique key), but no `Idempotency-Key` header surface exists, and replay risk is low (admin-only, no money semantics).
- **IP-003 cursor pagination** — not-applicable: `listAreaMemberships` returns the active set for a single `(userID, tenantID)` (artifact 02-flow-list-memberships §2 step 6). Bounded cardinality; pagination not implied.
- **IP-007 observability (request-scoped correlation id)** — not-applicable here but flagged as missing-ADR in tech-debt: IAM middleware writes `trace_id` into its error body (`middleware.go:129`) but no module-wide structured-logging contract exists; tracking belongs in a future cross-cutting ADR, not IAM-local debt.

## Summary

- 5 patterns admitted (IP-001, IP-004, IP-005, IP-006, IP-008)
- 2 patterns rated full adherence (IP-006, IP-008)
- 3 patterns rated partial / gap (IP-001, IP-004, IP-005) → drive tech-debt rows
- 0 new patterns proposed; no additions to `references/industry-patterns-index.md`
