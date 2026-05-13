# Phase 0 — Context for module taxonomy

## Existing wiki state

- Stub: `wiki/modules/taxonomy.md` (Last verified 2026-05-02 — STALE)
- Sibling shipped Arc42 docs: `iam.md`, `auth.md`, `documents.md`, `templates.md`, `approval.md`, `audit.md`, `registry.md`
- `wiki/modules/registry.md` cross-refs taxonomy as the source of profiles + areas; CD code derives from `{profile_code}-{area_code}-{seq}`.

## Boundary vs registry (confirmed)

- **taxonomy** = flat code-keyed catalog of three entities (DocumentFamily, DocumentProfile, ProcessArea). NOT hierarchical despite the name — only `process_areas` carries an optional `parent_code` with cycle-prevention.
- **registry** = `controlled_documents` catalog. Binds (profile, area) → CD slot.
- Boundary is clean: registry imports taxonomy for FK validation; taxonomy never imports registry.

## Known stale anchors in existing stub

1. Stub claims sentinel `ErrFamilyCodeImmutable` exists at `domain/family.go:16`. **Wrong** — only `ErrFamilyNotFound`, `ErrFamilyAlreadyInactive`, `ErrFamilyHasProfiles`. Family code immutability NOT enforced in domain — only in service `Update()` by ignoring body code.
2. Stub error code list omits some — `routes_families.go:97` does not match new error mapping at `writeFamilyError` (lines 98–115).
3. Stub says "Family deactivation uses `is_active`; Profile/Area uses `archived_at`" — confirmed.

## Critical pre-flight observations (raw, no judgement yet)

Surfaced now so Phase 1/2/4/6 do not miss them:

- **No tenant guard on family routes.** `listFamilies/createFamily/getFamily/updateFamily/deactivateFamily` never read `X-Tenant-ID`. Families are global by design — but no Postgres tripwire equivalent (audit.md / iam.md gold standard) exists for `document_families`. T-debt candidate.
- **Profile/area tenant scoping is header-trusted.** `tenantIDFromRequest` (routes_profiles.go:197) reads `X-Tenant-ID` header and falls back to `tenant.DevTenantID` if empty. **No tripwire-style enforcement at DB layer.** A misrouted request with no header writes/reads dev tenant data.
- **Path-prefix capability dispatcher.** `apps/api/cmd/metaldocs-api/permissions.go:158-180` uses `strings.HasPrefix` + method switch. Families branch (line 174) lists `POST/PUT/DELETE` but **omits PATCH** — yet `routes_families.go:67` exposes `PATCH /families/{code}`. Tier-1 capability check missed by routing string match. (Cross-ref: documents-refactor backlog R-008 cap namespace, approval T-012 unwired AuthorizationService.)
- **Families `code` immutability** advertised in stub but only enforced by `UpdateFamily` handler overwriting `Code: r.PathValue("code")`. No domain-level sentinel.
- **`HasActiveProfiles` deactivate guard** is application-layer, not DB-trigger. A racing INSERT of a profile pointing at the family would succeed during the window between check and update.
- **Capability `taxonomy.manage`** seeded for `system_admin` (migration 0165:40) and `qms_admin` (migration 0169:28). Two roles can mutate global family catalog. Cross-tenant blast radius.
- **Audit-trail gap.** `FamilyService` writes nothing to audit; `ProfileService` + `AreaService` write to `GovernanceLogger` (legacy module-local sink, NOT `audit.Writer` — same shape as audit-tech-debt T-007). Cross-cut: `application/governance_logger.go`.
- **Postgres tripwire status** for `document_families` / `document_profiles` / `process_areas` — TBD in Phase 4.

## ADR cross-refs (existing)

- `wiki/decisions/0007-two-tier-authz.md` — tier-1 cap check vs tier-2 in-tx area authz. Taxonomy uses tier-1 only.
- `wiki/decisions/0011-cd-atomic-create.md` — registry depends on profile + area existing.
- No ADR for "family catalog is globally shared, not tenant-scoped" — gap.
- No ADR for "process area hierarchy permits self-referential trees with cycle prevention only at SetParent" — gap.

## Out-edges (Phase 3 will confirm)

- `registry` — reads profile + area for CD creation (FK).
- `documents` — `documents_area_name_snapshot` (migration 0175) suggests area name is snapshotted into documents row at creation.
- `templates` — profile holds `default_template_version_id`; taxonomy imports `TemplateVersionChecker` adapter wired from templates.
- `iam` — capability namespace `taxonomy.manage` (capabilities.go:16); `process_areas` cross-cuts with `user_process_areas` membership table.
- `approval` — `process_area.default_approver_role` informs approval routing.

## In-edges (Phase 3 will confirm)

- `apps/api/cmd/metaldocs-api/main.go:197` — module wiring.
- `apps/api/cmd/metaldocs-api/main.go:225` — `profileRepo` re-instantiated for documents-v2 adapter (`profileDefaultsAdapter`).
- `permissions.go:158-180` — path-prefix capability dispatcher.

## Open questions deferred to tech-debt

None blocking the doc. All concerns above are candidates for `taxonomy-tech-debt.md` rows; severity to be applied in Phase 6 per rubric.

## Skip log (gates intentionally bypassed)

None.
