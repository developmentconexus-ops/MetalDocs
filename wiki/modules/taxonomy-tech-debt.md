# Tech Debt Register — taxonomy

> Companion to `wiki/modules/taxonomy.md`. Lists known gaps, smells, and missing-ADR items. **Debt only — no fix prescriptions.** Fixes belong in `wiki/backlog/taxonomy-refactor.md`.

**Last verified:** 2026-06-12 (Wave 2.12 sync — T-010 fully closed: DBGovernanceLogger deleted, AuditGovernanceAdapter is sole GovernanceLogger; new CI guards nosqltxindomain + nodualmode; prior Wave 2 sync)

## Severity scale

Triggers per `.claude/skills/metaldocs-module-doc/templates/tech-debt-register.md`.

- **Critical** — authn/authz bypass, regulated audit-trail gap, multi-tenant data leak, data-loss path, contract violation downstream consumers rely on, schema/version drift the boot-check misses.
- **Major** — defense-in-depth gap, governance sink wired to `nil` on regulated path, duplicated write surfaces with divergent semantics, documented contract not followed (measurable consumer impact), cross-module dep blocking another module's refactor.
- **Minor** — symbol-naming collision, missing doc comments, latent surface (no caller hits it today), bidirectional non-circular dep, missing standalone ADR for a rule already enforced by code + tests.

Pick highest trigger. Justify the call in `Observation`.

## Items

### T-001 · Tenant header trusted as authoritative on writes — **RESOLVED Plan 3**
- **Severity:** critical → **resolved**
- **Surface:** `internal/modules/taxonomy/delivery/http/routes_profiles.go:230-231` (`tenantIDFromRequest`).
- **Resolution (2026-05-11):** `tenantIDFromRequest` now calls `tenant.FromContext(r.Context())` (Plan 3 module sweep). The `X-Tenant-ID` header is stripped by auth middleware (`auth/delivery/http/middleware.go:87-88`) before reaching taxonomy handlers. Tenant is sourced from the session-bound `auth_sessions.tenant_id` (migration 0184). Cross-tenant write via header forgery is closed. Residual gap: no GUC/RLS row-level enforcement at DB layer (T-006 tracks this).
- **Evidence:** `_artifacts/02-flow-create-profile.md` §4 (trust chain — now stale); `_artifacts/03-deps.md` §1 (`internal/platform/tenant` import); `_artifacts/05-industry.md` IP-008.
- **Linked backlog row:** `backlog/taxonomy-refactor.md#R-001` (can be closed)
- **Linked ADR:** `wiki/architecture/tenant-context.md`

### T-002 · `document_families` globally shared with no ADR
- **Severity:** critical
- **Surface:** `migrations/0023_init_document_family_and_profile_registry.sql:1-7` (no `tenant_id` column); `migrations/0161_grant_families_write_privileges.sql:1` (`metaldocs_app` write granted); `internal/modules/taxonomy/infrastructure/family_repository.go:38-99` (every method lacks tenant predicate); `apps/api/cmd/metaldocs-api/permissions.go:177-181` (cap dispatcher allows any tenant with `taxonomy.manage`)
- **Observation:** `document_families` carries no `tenant_id` column and is a global catalog. `qms_admin` (migration 0169) and `system_admin` (migration 0165) both hold `taxonomy.manage`. A `qms_admin` in tenant A can `POST /families`, `PUT /families/{code}`, or `DELETE /families/{code}` and the mutation is visible to every tenant — blast radius is the entire deployment. No ADR documents the global-by-design choice or the threat model that accepted it. Trigger fired: multi-tenant data leak + cross-tenant write surface on a regulated catalog (Critical).
- **Evidence:** `_artifacts/04-persistence.md` §1, §2 (no `tenant_id` row), §5; `_artifacts/02-flow-deactivate-family.md` §4 ("Cross-tenant blast radius"); `_artifacts/05-industry.md` IP-008.
- **Linked backlog row:** `backlog/taxonomy-refactor.md#R-002`
- **Linked ADR:** missing-ADR

### T-003 · PATCH /api/v1/taxonomy/families/{code} bypasses capability dispatcher — CLOSED 2026-05-11 (Plan 5)
- **Severity:** critical (closed)
- **Surface:** `apps/api/cmd/metaldocs-api/permissions.go` — PATCH method added to the families branch (and areas branch). Plan 5 commit aligns the families dispatcher with the profiles branch which already included `MethodPatch`.
- **Observation (original):** The families capability switch enumerated POST/PUT/DELETE but omitted PATCH. The handler was mounted at PATCH. Capability resolver therefore returned `("", false)` for PATCH `/families/{code}` — path treated as public.
- **Evidence:** `_artifacts/01-surface.md` (handler.go route list); `_artifacts/03-deps.md` §1 (no authz import); direct grep of `permissions.go:177-181`.
- **Linked backlog row:** `backlog/taxonomy-refactor.md#R-003`
- **Linked ADR:** missing-ADR

### T-004 · `FamilyService` has no govLogger — Create/Update/Deactivate emit no governance events — CLOSED 2026-05-11 (Plan 6a)
- **Severity:** critical (closed)
- **Surface (resolved):** `internal/modules/taxonomy/application/family_service.go:13-19` — struct now has `govLogger domain.GovernanceLogger` field; `NewFamilyService` takes `govLogger` param; Create `:46-63`, Update `:103-120`, Deactivate `:161-177` all call `s.govLogger.Log`.
- **Observation:** Mutations on the globally-shared `document_families` (Create, Update, Deactivate) leave no governance-event row. Compared to ProfileService and AreaService, which both hold a `govLogger GovernanceLogger` field. ISO 9001 / QMS controls require traceability of catalog changes; the regulated path on the table with the widest blast radius (T-002) is the one that emits nothing. Trigger fired: regulated audit-trail gap (Critical, per rubric).
- **Evidence:** `_artifacts/02-flow-deactivate-family.md` §6; module-wiring code at `module.go:22-31`.
- **Linked backlog row:** `backlog/taxonomy-refactor.md#R-004`
- **Linked ADR:** missing-ADR

### T-005 · ProfileService.Create/Update + AreaService.Create/Update do not emit governance events — CLOSED 2026-05-11 (Plan 6a)
- **Severity:** critical (closed)
- **Surface (resolved):** `internal/modules/taxonomy/application/profile_service.go:70` (`Create` — calls `s.govLogger.Log`); `:96` (`Update` — same); `internal/modules/taxonomy/application/area_service.go:59` (`Create` — calls `s.govLogger.Log`); `:98` (`Update` — same). All four previously-silent paths now emit.
- **Observation:** ProfileService panics if govLogger is nil at construction (`profile_service.go:14`) but does not call it on Create or Update — only on SetDefaultTemplate and Archive. AreaService follows the same pattern (Archive emits; Create + Update do not). Regulated tenant-scoped catalog mutations therefore leave a partial audit trail: archives + template re-points are observed, but the act of bringing the row into existence and renaming/redefining it is not. Trigger fired: regulated audit-trail gap (Critical, per rubric).
- **Evidence:** `internal/modules/taxonomy/application/profile_service.go:70` (`s.govLogger.Log` call inside `Create`); `:96` (same inside `Update`); `internal/modules/taxonomy/application/area_service.go:59` + `:98` (same for `AreaService`). Note: `_artifacts/02-flow-create-profile.md` §6 is stale — it predates the 2026-05-11 Plan 6a patch and still claims no emission; the code is authoritative.
- **Linked backlog row:** `backlog/taxonomy-refactor.md#R-005`
- **Linked ADR:** missing-ADR

### T-006 · Single-tier defense-in-depth (no `authz.Require`, no DB tripwire) — PARTIALLY CLOSED 2026-05-11 (Plan 5)
- **Severity:** major → **partially resolved** (FamilyRepository + ProfileRepository + AreaRepository write methods now have tier-2; DB tripwire attached to the three owned tables)
- **Surface (resolved):** `internal/modules/taxonomy/infrastructure/family_repository.go:113` (`Create` calls `authz.Require(CapTaxonomyManage)`) · `:135` (`Update` same). `internal/modules/taxonomy/infrastructure/repository.go` — `ProfileRepository.Create` + `Update` and `AreaRepository.Create` + `Update` all call `authz.Require(CapTaxonomyManage)` inside a tx. `migrations/0188_tripwire_extend.sql:211-224` attaches `trg_require_cap_asserted` to `metaldocs.document_profiles`, `metaldocs.document_process_areas`, `metaldocs.document_families`.
- **Surface (residual):** Archive/deactivate paths and read paths still have no tier-2 call. `iam/authz` import is now present in taxonomy infrastructure (`family_repository.go:9`).
- **Observation (original):** Defense relied entirely on tier-1 path-prefix dispatcher. No in-tx `authz.Require(ctx, tx, cap, area)` call; no Postgres tripwire on any of the 3 owned tables. A bug at the tier-1 layer (e.g. T-003 dispatcher gap) was unguarded.
- **Evidence:** `_artifacts/03-deps.md` §1 (`internal/platform/authz` was ABSENT); `_artifacts/04-persistence.md` §3, §5; `_artifacts/05-industry.md` IP-004 (note: IP-004 is a pre-Plan 5 snapshot — its "Tier-2 absent" claim is stale; `family_repository.go:9` now imports `iam/authz` and lines 39, 71, 113, 135, 166, 196, 226, 250 call `authz.Require`; the code is authoritative).
- **Linked backlog row:** `backlog/taxonomy-refactor.md#R-006`
- **Linked ADR:** `wiki/decisions/0007-two-tier-authz.md` (taxonomy partially conformant as of Plan 5)

### T-007 · TOCTOU race + missing tx + cross-tenant SELECT in `FamilyService.Deactivate` — **RESOLVED**
- **Severity:** major → **resolved**
- **Surface (resolved):** `internal/modules/taxonomy/application/family_service.go:124-179` — `Deactivate` now uses `BeginTx` + `GetByCodeForUpdate` (SELECT FOR UPDATE) + `HasActiveProfilesTx` (same tx, takes `tenantID string`) + `UpdateTx` + Commit. `domain/port.go:69-72` — `FamilyRepository` interface exposes `HasActiveProfilesTx(ctx, tx, tenantID, familyCode)`. `family_repository.go:170-173` — query is `WHERE tenant_id=$1 AND family_code=$2`.
- **Observation (original):** `Deactivate` ran `GetByCode` + `HasActiveProfiles` + `Update` on three discrete `sql.DB` connections with no row lock or enclosing tx (TOCTOU), and `HasActiveProfiles` lacked a `tenant_id` predicate (cross-tenant probe). Both defects resolved in the same implementation change.
- **Evidence:** `_artifacts/02-flow-deactivate-family.md` §2, §4 (original evidence — stale post-resolution); `family_service.go:124-179`; `domain/port.go:69-72`; `family_repository.go:153-178`.
- **Linked backlog row:** `backlog/taxonomy-refactor.md#R-007`
- **Linked ADR:** missing-ADR

### T-008 · Legacy error envelope (RFC 9457 drift) — CLOSED 2026-05-12 (Plan 7, alias cascade)
- **Severity:** major (closed)
- **Surface (resolved):** `internal/modules/taxonomy/delivery/http/routes_profiles.go:19` — `writeError = httpresponse.WriteError` (package-level alias). `internal/platform/httpresponse/response.go:16-18` — `WriteError` now calls `problem.Write(w, problem.New(status, code, message))`. All taxonomy error paths (`writeFamilyError`, `writeProfileError`, `writeAreaError`, and inline `writeError` call sites) inherit RFC 9457 `application/problem+json` via this cascade. No taxonomy handler file required direct edits; verified by taxonomy contract tests (`commit f0bb64c0`).
- **Observation (original):** Every error path returned `{"code":"...","message":"..."}` instead of `application/problem+json`. The gap was codebase-wide across audit T-002, auth T-003, iam T-006, documents T-001.
- **Evidence:** `internal/platform/httpresponse/response.go:16-18` (`WriteError` calls `problem.Write(w, problem.New(status, code, message))` — RFC 9457 path confirmed); `_artifacts/05-industry.md` IP-001. Note: `_artifacts/02-flow-list-families.md` §5, `_artifacts/02-flow-create-profile.md` §5, and `_artifacts/02-flow-deactivate-family.md` §5 are stale — they predate the 2026-05-12 Plan 7 cascade and still claim non-RFC9457 envelopes; the code is authoritative.
- **Linked backlog row:** `backlog/taxonomy-refactor.md#R-008` (merged Plan 7 2026-05-11, commit `11589032` cascade + `f0bb64c0` test fix)
- **Linked ADR:** `wiki/architecture/api-design-system.md`

### T-009 · No OpenAPI spec; raw `http.ServeMux` instead of oapi-codegen — **CLOSED**
- **Severity:** major → **resolved**
- **Surface (resolved):** `internal/modules/taxonomy/api/` — `api.gen.go`, `cfg.yaml`, `gen.go`; `internal/modules/taxonomy/delivery/http/handler.go:42-51` calls `taxonomyapi.HandlerWithOptions`; `internal/modules/taxonomy/delivery/http/routes_generated.go:10` — compile-time `var _ taxonomyapi.ServerInterface = (*Handler)(nil)` assertion.
- **Observation (original):** All 16 routes were mounted on raw `net/http.ServeMux` with no operationId or OpenAPI spec; ADR 0012 contract-first commitment was unmet; taxonomy was the residual unmigrated module.
- **Evidence:** `_artifacts/01-surface.md` (original evidence — stale post-resolution); `handler.go:42-51`; `routes_generated.go:10`.
- **Linked backlog row:** `backlog/taxonomy-refactor.md#R-009`
- **Linked ADR:** `wiki/decisions/0012-contract-first-api.md` (taxonomy is the residual unmigrated module)

### T-010 · `DBGovernanceLogger` is a module-local parallel audit sink — FULLY CLOSED 2026-06-12 (Wave 2.12)
- **Severity:** major (fully closed)
- **Surface (resolved):** `internal/modules/taxonomy/application/governance_logger.go` (file DELETED). `application/audit_governance_adapter.go` is now the sole `GovernanceLogger` implementation; it writes to `metaldocs.audit_events` via `auditdomain.Writer`. `internal/modules/controlleddocuments/module.go:35` now uses `taxonomyapp.NewAuditGovernanceAdapter(deps.AuditWriter)` — the `NewDBGovernanceLogger` fallback is removed and `AuditWriter` is required (panics on nil).
- **Observation (original):** Taxonomy shipped its own audit sink (`governance_events` table) instead of consuming `auditdomain.Writer`. Controlled-documents re-exported the taxonomy logger rather than wiring its own audit writer. Both dual-sink paths closed by Wave 2.12.
- **Evidence:** `internal/modules/taxonomy/application/audit_governance_adapter.go` (sole GovernanceLogger impl); `internal/modules/controlleddocuments/module.go:27-35` (AuditWriter nil-panic guard + NewAuditGovernanceAdapter call).
- **Linked backlog row:** `backlog/taxonomy-refactor.md#R-010`
- **Linked ADR:** missing-ADR

### T-011 · No idempotency on write routes
- **Severity:** minor
- **Surface:** `internal/modules/taxonomy/delivery/http/routes_profiles.go:70-112` (createProfile — no `Idempotency-Key` parse); same shape on `createFamily`, `createArea`, `updateProfile`, `updateArea`, `updateFamily`
- **Observation:** Duplicate POST `/profiles` (or `/families`, `/areas`) with the same code returns PG `23505` (PK violation). `writeProfileError` (`routes_profiles.go:177-193`) has no mapping for `23505` → falls to `INTERNAL_ERROR 500`. Catalog volume is small; a determined retry storm is the only meaningful trigger. Trigger fired: latent (surface exists, no caller hits it observably today).
- **Evidence:** `_artifacts/02-flow-create-profile.md` §6 (idempotency=no).
- **Linked backlog row:** `backlog/taxonomy-refactor.md#R-011`
- **Linked ADR:** missing-ADR

### T-012 · No pagination on list endpoints
- **Severity:** minor
- **Surface:** `internal/modules/taxonomy/delivery/http/routes_families.go:20-31` (listFamilies returns `{"items":[...]}` unbounded); same shape on listProfiles, listAreas; `family_repository.go:38-45` (SELECT ORDER BY code, no LIMIT)
- **Observation:** Full ordered slice returned. Cardinality expected to stay small (profiles + areas under ~50 per tenant; families ~10 global). No current driver to paginate, but the surface scales linearly with catalog growth. Trigger fired: latent (no observable caller impact today).
- **Evidence:** `_artifacts/02-flow-list-families.md` §5 ("No `next_cursor` / pagination fields"); `_artifacts/05-industry.md` IP-003 "not applicable".
- **Linked backlog row:** `backlog/taxonomy-refactor.md#R-012`
- **Linked ADR:** missing-ADR

### T-013 · Family code immutability is handler-side only — CLOSED 2026-05-11 (Plan 5)
- **Severity:** minor (closed)
- **Surface:** `migrations/0188_tripwire_extend.sql:239-257` — `reject_families_code_update()` BEFORE-UPDATE trigger added to `metaldocs.document_families`, mirroring the `0122`/`0123` pattern for profiles + areas. Handler-side override of `body.code` with path-param `code` remains as an additional safe guard.
- **Observation (original):** Code immutability on profiles + areas was DB-enforced (`reject_code_update()` BEFORE-UPDATE trigger). Family code immutability was enforced only by the handler — a single layer that a bypass path (T-003 PATCH gap) could defeat.
- **Evidence:** `_artifacts/04-persistence.md` §1 (no trigger row for `document_families`); `_artifacts/04-persistence.md` §3 (trigger inventory).
- **Linked backlog row:** `backlog/taxonomy-refactor.md#R-013`
- **Linked ADR:** missing-ADR

### T-014 · No Go doc comments on exported symbols
- **Severity:** minor
- **Surface:** every file under `internal/modules/taxonomy/` — all 80 exported symbols from `_artifacts/01-surface.md` lack doc comments (verified during Phase 1)
- **Observation:** `DocumentFamily`, `DocumentProfile`, `ProcessArea`, sentinel errors, all service + repository methods, the `Module` composition root — none have `// Description` Go doc comments. `go doc` returns signatures only. Trigger fired: missing doc comments on exported symbols.
- **Evidence:** `_artifacts/01-surface.md` (every symbol row missing a doc-comment column).
- **Linked backlog row:** `backlog/taxonomy-refactor.md#R-014`
- **Linked ADR:** missing-ADR

### T-015 · Redundant PK on `code` alongside per-tenant unique index
- **Severity:** minor
- **Surface:** `migrations/0023_init_document_family_and_profile_registry.sql:10` (PK on `document_profiles.code`); `migrations/0122_taxonomy_extend_document_profiles.sql:21-22` (`ux_document_profiles_tenant_code (tenant_id, code)`); same pattern on `document_process_areas` (`0025:2` + `0123:19-20`)
- **Observation:** `document_profiles` and `document_process_areas` carry a PRIMARY KEY on `code` alone (cross-tenant unique) plus a UNIQUE index on `(tenant_id, code)`. The PK on `code` alone makes cross-tenant code collisions impossible, so the `(tenant_id, code)` UNIQUE constraint adds no new uniqueness — it only adds an index for tenant-filtered lookups. The redundancy is a structural artefact of the 0122/0123 tenant-extension migrations that did not drop or re-key the original PK. Trigger fired: latent (correctness is unaffected; cost is one extra index per table + confused readers).
- **Evidence:** `_artifacts/04-persistence.md` §4 ("Note: profile + area carry PK on `code` alone … the broader PK is redundant").
- **Linked backlog row:** `backlog/taxonomy-refactor.md#R-015`
- **Linked ADR:** missing-ADR

### T-016 · Missing-ADR for cycle-prevention rule + area hierarchy shape
- **Severity:** minor
- **Surface:** `internal/modules/taxonomy/application/area_service.go` (`SetParent` invokes `ListAncestors` to reject `parent_code` cycles); `internal/modules/taxonomy/infrastructure/repository.go` (`ListAncestors` recursive walk); `migrations/0123:10-13` (self-FK `(tenant_id, parent_code) → (tenant_id, code)`)
- **Observation:** Area hierarchy uses a self-FK + application-layer acyclicity check. Defensible (recursive CTE would also work; trigger could enforce it at the DB) but undocumented as an ADR. Future refactors (e.g. moving cycle check into a Postgres `assert_no_cycle` function, or switching to closure table) have no design context to evaluate against. Trigger fired: missing standalone ADR for a rule already enforced by code + tests.
- **Evidence:** `_artifacts/03-deps.md` §5 (`application/area_service_test.go` covers archive rules); `_artifacts/04-persistence.md` §1 (`parent_code` self-FK).
- **Linked backlog row:** `backlog/taxonomy-refactor.md#R-016`
- **Linked ADR:** missing-ADR

---

## Coverage stats (computed at compose time)

- Public symbols undocumented: 80 / 80 (all exported symbols lack Go doc comments — see T-014)
- Operations missing C4 placement: 0 / 16 (all 16 routes appear in §5.3 + at least one is traced in §6)
- Cross-deps missing in §5/§8: 0 / 10 (all OUT/IN edges from `_artifacts/03-deps.md` are placed)
- State transitions missing in §6: 0 / 1 (deactivateFamily covered)
- Decisions without ADR link: 14 / 16 (T-001..T-005, T-007, T-008, T-010..T-016 = 14 missing-ADR; T-006 + T-009 link existing ADRs as residuals)
