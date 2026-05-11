# Tech Debt Register — registry

> Companion to [wiki/modules/registry.md](registry.md). Lists known gaps, smells, and missing-ADR items. **Debt only — no fix prescriptions.** Fixes belong in [wiki/backlog/registry-refactor.md](../backlog/registry-refactor.md).

**Last verified:** 2026-05-11 (Plan 5)

## Severity scale

The category names are useful only when paired with concrete triggers. Use the trigger list. When in doubt and the bug is on a regulated path, escalate one level.

### Critical — at least one trigger fires
- Authn/authz bypass: a code path lets a request mutate or read without the capability the spec requires.
- Regulated audit-trail gap: a mutation on an ISO 9001 / QMS / regulated path is not written to the audit sink.
- Multi-tenant data leak: a query path can return rows from a different tenant.
- Data-loss path: a code path can drop / overwrite / silently truncate user data.
- Contract violation that downstream consumers rely on.
- Schema/version drift the boot check is supposed to catch but does not.

### Major — at least one trigger fires
- Defense-in-depth gap: only one layer protects a mutation that the spec calls for multiple layers on.
- Governance / observability sink wired to `nil` on a regulated path.
- Duplicated write surfaces with different semantics for the same use case.
- Documented contract not followed by this module yet (e.g. RFC 9457 envelope on a v1 route).
- Cross-module dependency that blocks another module's clean refactor.

### Minor — code-smell / latent / docs
- Symbol naming collision across packages.
- Missing Go doc comments on exported symbols.
- Latent debt: the surface for the bug exists in code but no caller hits it today.
- Bidirectional dependency that is non-circular today but would be hard to detangle.
- Missing standalone ADR for a rule that is already enforced by code + tests.

## Items

### T-001 · Lifecycle PUTs lack in-module authz; capability mapping unverified — CLOSED 2026-05-11 (Plan 5)
- **Severity:** critical (closed)
- **Surface:** `internal/modules/registry/application/service.go:297` (`Obsolete` passes `string(iamdomain.CapRegistryObsolete)` to `changeStatus`) · `:301` (`Supersede` passes `CapRegistrySupersede`) · `:327` (`changeStatus` calls `authz.Require(ctx, tx, cap, "tenant")` inside a new tx). `migrations/0187_registry_lifecycle_caps_seed.sql` renames `doc.supersede → registry.supersede`, seeds `registry.obsolete`. `apps/api/cmd/metaldocs-api/permissions.go` PATCH method added to taxonomy families + areas routes; obsolete/supersede cases changed to `CapRegistryObsolete`/`CapRegistrySupersede`.
- **Observation (original):** `Obsolete` and `Supersede` handlers + service contained no `authz.Require`, no `CapabilityService` call, no capability-name constant. `migrations/0165_role_capabilities_reseed.sql` seeded only `registry.create`. Any authenticated user could transition the canonical QMS catalog.
- **Evidence:** `_artifacts/02-flow-obsolete.md` §4
- **Linked backlog row:** [`backlog/registry-refactor.md#R-001`](../backlog/registry-refactor.md)
- **Linked ADR:** missing-ADR

### T-002 · Audit-trail gap on Obsolete / Supersede
- **Severity:** critical
- **Surface:** `internal/modules/registry/application/service.go:309-317`
- **Observation:** `changeStatus` performs get + active-guard + UPDATE only. No `s.govLogger.Log(...)` call. The create path emits governance events (`service.go:267-271`); the lifecycle path does not. Transitioning a controlled document from `active` to `obsolete` or `superseded` is a regulated QMS event under ISO 9001 — Critical trigger "Regulated audit-trail gap" fires.
- **Evidence:** `_artifacts/02-flow-obsolete.md` §6
- **Linked backlog row:** [`backlog/registry-refactor.md#R-002`](../backlog/registry-refactor.md)
- **Linked ADR:** missing-ADR

### T-003 · Legacy `{code, message}` error envelope across module
- **Severity:** major
- **Surface:** `internal/platform/httpresponse/response.go:14-15` consumed by all registry routes (`routes.go:43..343`)
- **Observation:** Errors emit JSON object `{"code": "...", "message": "..."}` with default content-type. RFC 9457 Problem Details (`application/problem+json`) is not used. `Content-Type` is never set to `application/problem+json` on any registry error path. Peer modules (documents T-001, audit T-002, templates_v2 T-005, approval T-001) carry the same gap — uniform spec drift.
- **Evidence:** `_artifacts/02-flow-atomic-create.md` §5; `_artifacts/05-industry.md` IP-001
- **Linked backlog row:** [`backlog/registry-refactor.md#R-003`](../backlog/registry-refactor.md)
- **Linked ADR:** missing-ADR

### T-004 · Tier-3 Postgres tripwire absent for registry-owned tables — CLOSED 2026-05-11 (Plan 5)
- **Severity:** major (closed)
- **Surface (resolved):** `internal/modules/registry/infrastructure/repository.go:135` (`Create` now opens tx + `authz.Require(CapRegistryCreate)` at `:142`) · `:151` (`CreateTx` calls `authz.Require` at `:155`). `migrations/0188_tripwire_extend.sql:201-208` attaches `trg_require_cap_asserted` to `public.controlled_documents` (INSERT + UPDATE, with OR-logic for `registry.obsolete|registry.supersede` on UPDATE) and `public.cd_sequence_counters` (line 206).
- **Observation (original):** 5 mutator methods (`Create`, `CreateTx`, `UpdateStatus`, `EnsureCounter`, `NextAndIncrement`) executed INSERT/UPDATE without preceding `authz.Require(...)`. None set `metaldocs.asserted_caps`. The `enforce_capability_asserted` trigger installed by 0142b covered `approval_instances` and `signoffs`; `controlled_documents` and `cd_sequence_counters` were NOT in the protected set.
- **Evidence:** `_artifacts/04-persistence.md` §5 (5 violations); `_artifacts/05-industry.md` IP-004
- **Linked backlog row:** [`backlog/registry-refactor.md#R-004`](../backlog/registry-refactor.md)
- **Linked ADR:** [wiki/decisions/0007-two-tier-authz.md](../decisions/0007-two-tier-authz.md)

### T-005 · Tenant scoping via query arg only — no GUC + RLS backstop
- **Severity:** major
- **Surface:** `internal/modules/registry/infrastructure/repository.go:26`, `:36`, `:46`, `:57`, `:137`, `:184`, `:208`, `:239`
- **Observation:** Every WHERE clause includes `tenant_id = $...` from the request context (sourced via `tenant.FromContext` — Plan 3 removed the `X-Tenant-ID` header source). No `SET LOCAL metaldocs.tenant_id` GUC is issued before the query; no RLS policy on `controlled_documents` / `cd_sequence_counters`. A repository method that forgets the `tenant_id` predicate has no DB-level backstop. Defense-in-depth gap on a multi-tenant table.
- **Evidence:** `_artifacts/04-persistence.md` §5; `_artifacts/05-industry.md` IP-008
- **Linked backlog row:** [`backlog/registry-refactor.md#R-005`](../backlog/registry-refactor.md)
- **Linked ADR:** missing-ADR

### T-006 · GetActiveDocument: no authz
- **Severity:** major
- **Surface:** `internal/modules/registry/delivery/http/routes.go:232-326`
- **Observation:** No `authz.Require` call; no `metaldocs.assert_caps`. Document content hashes, approval state, and published-revision IDs are returned to any authenticated caller. **Plan 3 resolved the header-trust sub-issue** — tenant is now sourced from `tenant.FromContext` via `injectTenant` middleware (`handler.go:48`); the `X-Tenant-ID` header is stripped by auth middleware. The outstanding gap is the missing read-policy authz enforcement. Pending a centralized read-policy ADR, defense-in-depth gap.
- **Evidence:** `_artifacts/02-flow-get-active.md` §2, §4
- **Linked backlog row:** [`backlog/registry-refactor.md#R-006`](../backlog/registry-refactor.md)
- **Linked ADR:** missing-ADR

### T-007 · OpenAPI spec/handler drift on 422 `template_invalid`
- **Severity:** major
- **Surface:** `api/openapi/v1/partials/registry.yaml:73`; `internal/modules/registry/delivery/http/routes.go:410-444`
- **Observation:** Spec declares 422 `template_invalid` on `POST /controlled-documents`; handler's `writeDomainError` switch has no branch mapping any error to that code. Downstream OpenAPI clients (generated from the spec) include a 422 case that the server never emits. Contract drift — Major per "Documented contract not followed by this module yet".
- **Evidence:** `_artifacts/02-flow-atomic-create.md` §5
- **Linked backlog row:** [`backlog/registry-refactor.md#R-007`](../backlog/registry-refactor.md)
- **Linked ADR:** [wiki/decisions/0012-contract-first-api.md](../decisions/0012-contract-first-api.md)

### T-008 · Cross-module audit sink — taxonomy logger reused
- **Severity:** major
- **Surface:** `internal/modules/registry/module.go:31`; `internal/modules/taxonomy/application/governance_logger.go:18`
- **Observation:** Registry wires its governance logger from `taxonomyapp.NewDBGovernanceLogger(deps.DB)`. Registry's audit emissions land in `governance_events` via taxonomy's adapter. Two coupled concerns: (a) the audit sink semantics belong to a platform-owned `internal/audit` per the audit module's port/adapter contract (see `wiki/modules/audit.md`), not to a sibling business module; (b) registry's debt items + retention policies become coupled to taxonomy refactors. Cross-module dependency that blocks audit's clean port adoption.
- **Evidence:** `_artifacts/03-deps.md` §1 (OUT-edges)
- **Linked backlog row:** [`backlog/registry-refactor.md#R-008`](../backlog/registry-refactor.md)
- **Linked ADR:** missing-ADR

### T-009 · Documents DI cycle resolved via post-construction setter
- **Severity:** minor
- **Surface:** `apps/api/cmd/metaldocs-api/main.go:203`, `:325`; `internal/modules/registry/application/service.go:99`
- **Observation:** `registry.New(...)` constructs `RegistryService` with the 8th argument `nil` (the `DocumentInitializer`). `main.go:325` calls `registryModule.Service().WithDocumentInitializer(docapp.NewCDDocumentInitializer(docMod.Service))` after the documents module is built. Cycle break is intentional (registry owns the port; documents implements). Order-of-construction is now a hidden contract — if a future caller forgets the setter, `RegistryService.Create` will nil-panic on the port call (`service.go:247`). Latent.
- **Evidence:** `_artifacts/03-deps.md` §3
- **Linked backlog row:** [`backlog/registry-refactor.md#R-009`](../backlog/registry-refactor.md)
- **Linked ADR:** [wiki/decisions/0011-cd-atomic-create.md](../decisions/0011-cd-atomic-create.md)

### T-010 · Parallel repository instance constructed outside module boundary
- **Severity:** minor
- **Surface:** `apps/api/cmd/metaldocs-api/main.go:224`
- **Observation:** A second `PostgresControlledDocumentRepository` is instantiated standalone at `main.go:224` for search/resolver wiring. The module exposes `Module.Service()` but not its internal repository, so consumers reach in via `registryinfra.NewPostgresControlledDocumentRepository`. Module-boundary leak; ties external code to the internal infrastructure package layout.
- **Evidence:** `_artifacts/03-deps.md` §3
- **Linked backlog row:** [`backlog/registry-refactor.md#R-010`](../backlog/registry-refactor.md)
- **Linked ADR:** missing-ADR

### T-011 · OpenAPI partial at `v1/` despite `/api/v2/` HTTP path
- **Severity:** minor
- **Surface:** `api/openapi/v1/partials/registry.yaml`; routes at `/api/v2/controlled-documents/*`
- **Observation:** The spec partial lives under `api/openapi/v1/` while the HTTP path prefix is `/api/v2/`. Generated server stubs (`internal/modules/registry/api/api.gen.go`) and clients consequently encode the `v2` path strings without a matching `v2/` spec tree. Cosmetic but confusing for new contributors; potential breakage on future spec restructuring (e.g. when other modules grow a real `v2/` tree).
- **Evidence:** `_artifacts/02-flow-atomic-create.md` §1
- **Linked backlog row:** [`backlog/registry-refactor.md#R-011`](../backlog/registry-refactor.md)
- **Linked ADR:** missing-ADR

### T-012 · Exported symbols mostly without Go doc comments
- **Severity:** minor
- **Surface:** `internal/modules/registry/**/*.go`
- **Observation:** Surface scan recorded 90 exported symbols. Doc comments present on ~11 (notably `CreateResult` `service.go:63`, `WithDocumentInitializer` `:99`, `PreviewCode` `:279`, `PeekSeq` `:289`, `CreateRevision` `:330`, `CloneTemplateRequest` `document_initializer.go:11`, `DocumentRef` `:20`, `DocumentInitializer` `:30`, `EnsureCounter` `repository.go:208`, `Peek` `:224`, `NextAndIncrement` `:239`). 79 exports lack doc comments. Latent — readability + IDE tooling impact.
- **Evidence:** `_artifacts/01-surface.md` §2
- **Linked backlog row:** [`backlog/registry-refactor.md#R-012`](../backlog/registry-refactor.md)
- **Linked ADR:** missing-ADR

---

## Coverage stats (computed at compose time)

- Public symbols undocumented: 79 / 90
- Operations missing C4 placement: 0 / 8
- Cross-deps missing in §5/§8: 0 / 19
- State transitions missing in §6: 0 / 2
- Decisions without ADR link: 9 / 12
