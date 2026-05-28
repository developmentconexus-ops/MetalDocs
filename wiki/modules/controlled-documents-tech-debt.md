# Tech Debt Register — controlled-documents

> Companion to [wiki/modules/controlled-documents.md](controlled-documents.md). Lists known gaps, smells, and missing-ADR items. **Debt only — no fix prescriptions.** Fixes belong in [wiki/backlog/controlled-documents-refactor.md](../backlog/controlled-documents-refactor.md).

**Last verified:** 2026-05-25 (backend medium quality-bar sync)

## Severity scale

The category names are useful only when paired with concrete triggers. Use the trigger list. When in doubt and the bug is on a regulated path, escalate one level.

### Critical     at least one trigger fires
- Authn/authz bypass: a code path lets a request mutate or read without the capability the spec requires.
- Regulated audit-trail gap: a mutation on an ISO 9001 / QMS / regulated path is not written to the audit sink.
- Multi-tenant data leak: a query path can return rows from a different tenant.
- Data-loss path: a code path can drop / overwrite / silently truncate user data.
- Contract violation that downstream consumers rely on.
- Schema/version drift the boot check is supposed to catch but does not.

### Major     at least one trigger fires
- Defense-in-depth gap: only one layer protects a mutation that the spec calls for multiple layers on.
- Governance / observability sink wired to `nil` on a regulated path.
- Duplicated write surfaces with different semantics for the same use case.
- Documented contract not followed by this module yet (e.g. RFC 9457 envelope on a v1 route).
- Cross-module dependency that blocks another module's clean refactor.

### Minor     code-smell / latent / docs
- Symbol naming collision across packages.
- Missing Go doc comments on exported symbols.
- Latent debt: the surface for the bug exists in code but no caller hits it today.
- Bidirectional dependency that is non-circular today but would be hard to detangle.
- Missing standalone ADR for a rule that is already enforced by code + tests.

## Items

### T-001    Lifecycle PUTs lack in-module authz; capability mapping unverified     CLOSED 2026-05-11 (Plan 5)
- **Severity:** critical (closed)
- **Surface:** `internal/modules/controlleddocuments/application/service.go:367` (`Obsolete` passes `string(iamdomain.CapControlledDocumentObsolete)` to `changeStatus`)    `:371` (`Supersede` passes `CapControlledDocumentSupersede`)    `:394` (`changeStatus` calls `authz.Require(ctx, tx, cap, "tenant")` inside a new tx). `migrations/0187_registry_lifecycle_caps_seed.sql` (legacy literal migration filename) performs lifecycle-capability reseeding. `apps/api/cmd/metaldocs-api/permissions.go` PATCH method added to taxonomy families + areas routes; obsolete/supersede cases changed to `CapControlledDocumentObsolete`/`CapControlledDocumentSupersede`.
- **Observation (original):** `Obsolete` and `Supersede` handlers + service contained no `authz.Require`, no `CapabilityService` call, no capability-name constant. `migrations/0165_role_capabilities_reseed.sql` seeded only `registry.create` (legacy literal capability key). Any authenticated user could transition the canonical QMS catalog.
- **Evidence:** `_artifacts/02-flow-obsolete.md`   4
- **Linked backlog row:** [`backlog/controlled-documents-refactor.md#R-001`](../backlog/controlled-documents-refactor.md)
- **Linked ADR:** missing-ADR

### T-002    Audit-trail gap on Obsolete / Supersede     CLOSED 2026-05-11 (Plan 6a)
- **Severity:** critical (closed)
- **Surface:** `internal/modules/controlleddocuments/application/service.go:309-317`
- **Observation:** `changeStatus` performs get + active-guard + UPDATE only. No `s.govLogger.Log(...)` call. The create path emits governance events (`service.go:267-271`); the lifecycle path does not. Transitioning a controlled document from `active` to `obsolete` or `superseded` is a regulated QMS event under ISO 9001     Critical trigger "Regulated audit-trail gap" fires.
- **Evidence:** `_artifacts/02-flow-obsolete.md`   6
- **Linked backlog row:** [`backlog/controlled-documents-refactor.md#R-002`](../backlog/controlled-documents-refactor.md)
- **Linked ADR:** missing-ADR

### T-003    Legacy `{code, message}` error envelope across module     CLOSED 2026-05-12 (Plan 7)
- **Severity:** major (closed)
- **Surface (resolved):** `internal/platform/httpresponse/response.go:16-18`     `WriteError` now delegates to `problem.Write(w, problem.New(status, code, message))`. All controlled-documents routes that called `httpresponse.WriteError` inherit RFC 9457 `application/problem+json` output. `internal/modules/controlleddocuments/delivery/http/routes.go:470-471`     `ErrTemplateProfileMismatch` branch directly calls `problem.Write` with 422 `template_invalid`.
- **Observation (original):** Errors emitted JSON object `{"code": "...", "message": "..."}` with default content-type. RFC 9457 Problem Details (`application/problem+json`) was not used. Peer modules (documents T-001, audit T-002, templates T-005, approval T-001) carried the same gap.
- **Evidence:** `_artifacts/02-flow-atomic-create.md`   5; `_artifacts/05-industry.md` IP-001
- **Linked backlog row:** [`backlog/controlled-documents-refactor.md#R-003`](../backlog/controlled-documents-refactor.md) (merged Plan 7 2026-05-11, commit `11589032` + `395b0b24`)
- **Linked ADR:** `wiki/architecture/api-design-system.md`

### T-004    Tier-3 Postgres tripwire absent for controlled-documents-owned tables     CLOSED 2026-05-11 (Plan 5)
- **Severity:** major (closed)
- **Surface (resolved):** `internal/modules/controlleddocuments/infrastructure/repository.go:317` (`Create` now opens tx + `authz.Require(CapControlledDocumentCreate)` at `:317`)    `:330` (`CreateTx` calls `authz.Require` at `:330`). `migrations/0188_tripwire_extend.sql:201-208` (legacy literal migration filename) attaches `trg_require_cap_asserted` to `public.controlled_documents` (INSERT + UPDATE, with OR-logic for `controlled_documents.obsolete|controlled_documents.supersede` on UPDATE) and `public.cd_sequence_counters` (line 206).
- **Observation (original):** 5 mutator methods (`Create`, `CreateTx`, `UpdateStatus`, `EnsureCounter`, `NextAndIncrement`) executed INSERT/UPDATE without preceding `authz.Require(...)`. None set `metaldocs.asserted_caps`. The `enforce_capability_asserted` trigger installed by 0142b covered `approval_instances` and `signoffs`; `controlled_documents` and `cd_sequence_counters` were NOT in the protected set.
- **Evidence:** `_artifacts/04-persistence.md`   5 (5 violations); `_artifacts/05-industry.md` IP-004
- **Linked backlog row:** [`backlog/controlled-documents-refactor.md#R-004`](../backlog/controlled-documents-refactor.md)
- **Linked ADR:** [wiki/decisions/0007-two-tier-authz.md](../decisions/0007-two-tier-authz.md)

### T-005    Tenant scoping via query arg only     no GUC + RLS backstop
- **Severity:** major
- **Surface:** `internal/modules/controlleddocuments/infrastructure/repository.go:26`, `:36`, `:46`, `:57`, `:137`, `:184`, `:208`, `:239`
- **Observation:** Every WHERE clause includes `tenant_id = $...` from the request context (sourced via `tenant.FromContext`     Plan 3 removed the `X-Tenant-ID` header source). No `SET LOCAL metaldocs.tenant_id` GUC is issued before the query; no RLS policy on `controlled_documents` / `cd_sequence_counters`. A repository method that forgets the `tenant_id` predicate has no DB-level backstop. Defense-in-depth gap on a multi-tenant table.
- **Evidence:** `_artifacts/04-persistence.md`   5; `_artifacts/05-industry.md` IP-008
- **Linked backlog row:** [`backlog/controlled-documents-refactor.md#R-005`](../backlog/controlled-documents-refactor.md)
- **Linked ADR:** missing-ADR

### T-006    GetActiveDocument: no authz
- **Severity:** major
- **Surface:** `internal/modules/controlleddocuments/delivery/http/routes.go:232-326`
- **Observation:** No `authz.Require` call; no `metaldocs.assert_caps`. Document content hashes, approval state, and published-revision IDs are returned to any authenticated caller. **Plan 3 resolved the header-trust sub-issue**     tenant is now sourced from `tenant.FromContext` via `injectTenant` middleware (`handler.go:48`); the `X-Tenant-ID` header is stripped by auth middleware. The outstanding gap is the missing read-policy authz enforcement. Pending a centralized read-policy ADR, defense-in-depth gap.
- **Evidence:** `_artifacts/02-flow-get-active.md`   2,   4
- **Linked backlog row:** [`backlog/controlled-documents-refactor.md#R-006`](../backlog/controlled-documents-refactor.md)
- **Linked ADR:** missing-ADR

### T-007    OpenAPI spec/handler drift on 422 `template_invalid`     CLOSED 2026-05-12 (Plan 7)
- **Severity:** major (closed)
- **Surface (resolved):** `internal/modules/controlleddocuments/delivery/http/routes.go:470-471`     `writeDomainError` now has a branch: `errors.Is(err, controlleddocumentsdomain.ErrTemplateProfileMismatch)`     `problem.Write(w, problem.New(http.StatusUnprocessableEntity, "template_invalid", "template version does not match the document profile"))`. Spec 422 case and runtime are now aligned.
- **Observation (original):** Spec declared 422 `template_invalid` on `POST /controlled-documents`; handler's `writeDomainError` switch had no branch mapping any error to that code. Contract drift     downstream OpenAPI clients included a 422 case the server never emitted.
- **Evidence:** `_artifacts/02-flow-atomic-create.md`   5
- **Linked backlog row:** [`backlog/controlled-documents-refactor.md#R-007`](../backlog/controlled-documents-refactor.md) (merged Plan 7 2026-05-11, commit `395b0b24`)
- **Linked ADR:** [wiki/decisions/0012-contract-first-api.md](../decisions/0012-contract-first-api.md)

### T-008    Cross-module audit sink     taxonomy logger reused     CLOSED 2026-05-11 (Plan 6a)
- **Severity:** major (closed)
- **Surface:** `internal/modules/controlleddocuments/module.go:31`; `internal/modules/taxonomy/application/governance_logger.go:18`
- **Observation:** Controlled-documents wires its governance logger from `taxonomyapp.NewDBGovernanceLogger(deps.DB)`. Controlled-documents audit emissions land in `governance_events` via taxonomy's adapter. Two coupled concerns: (a) the audit sink semantics belong to a platform-owned `internal/audit` per the audit module's port/adapter contract (see `wiki/modules/audit.md`), not to a sibling business module; (b) controlled-documents debt items + retention policies become coupled to taxonomy refactors. Cross-module dependency that blocks audit's clean port adoption.
- **Evidence:** `_artifacts/03-deps.md`   1 (OUT-edges)
- **Linked backlog row:** [`backlog/controlled-documents-refactor.md#R-008`](../backlog/controlled-documents-refactor.md)
- **Linked ADR:** missing-ADR

### T-009    Documents DI cycle resolved via post-construction setter
- **Severity:** minor
- **Surface:** `apps/api/cmd/metaldocs-api/main.go:224`, `:343`; `internal/modules/controlleddocuments/application/service.go:99`
- **Observation:** `controlledDocumentsModule.New(...)` constructs `ControlledDocumentService` with the 8th argument `nil` (the `DocumentInitializer`). `main.go:343` calls `controlledDocumentsModule.Service().WithDocumentInitializer(docapp.NewCDDocumentInitializer(docMod.Service))` after the documents module is built. Cycle break is intentional (controlled-documents owns the port; documents implements). Order-of-construction is now a hidden contract     if a future caller forgets the setter, `Create` on `ControlledDocumentService` will nil-panic on the port call (`service.go:247`). Latent.
- **Evidence:** `_artifacts/03-deps.md`   3
- **Linked backlog row:** [`backlog/controlled-documents-refactor.md#R-009`](../backlog/controlled-documents-refactor.md)
- **Linked ADR:** [wiki/decisions/0011-cd-atomic-create.md](../decisions/0011-cd-atomic-create.md)

### T-010    Parallel repository instance constructed outside module boundary
- **Severity:** minor
- **Surface:** `apps/api/cmd/metaldocs-api/main.go:224`
- **Observation:** A second `PostgresControlledDocumentRepository` is instantiated standalone at `main.go:224` for search/resolver wiring. The module exposes `Module.Service()` but not its internal repository, so consumers reach in via `controlleddocumentsinfra.NewPostgresControlledDocumentRepository`. Module-boundary leak; ties external code to the internal infrastructure package layout.
- **Evidence:** `_artifacts/03-deps.md`   3
- **Linked backlog row:** [`backlog/controlled-documents-refactor.md#R-010`](../backlog/controlled-documents-refactor.md)
- **Linked ADR:** missing-ADR

### T-011    OpenAPI partial at `v1/` despite `/api/v1/` HTTP path
- **Severity:** minor
- **Surface:** canonical spec partial `api/openapi/v1/partials/controlled-documents.yaml`; canonical public API remains `/api/v1/controlled-documents/*`.
- **Observation:** The spec partial lives under `api/openapi/v1/` while the HTTP path prefix is `/api/v1/`. Generated server stubs (`internal/modules/controlleddocuments/api/api.gen.go`) and clients encode `/api/v1/...` routes from that partial; the naming/layout mismatch is cosmetic but confusing for new contributors and can complicate future spec-tree restructuring.
- **Evidence:** `_artifacts/02-flow-atomic-create.md`   1
- **Linked backlog row:** [`backlog/controlled-documents-refactor.md#R-011`](../backlog/controlled-documents-refactor.md)
- **Linked ADR:** missing-ADR

### T-012    Exported symbols mostly without Go doc comments
- **Severity:** minor
- **Surface:** `internal/modules/controlleddocuments/**/*.go`
- **Observation:** Surface scan recorded 94 exported symbols after the 2026-05-25 constructor additions. Doc comments present on ~15 (notably `CreateResult`, `WithDocumentInitializer`, `PreviewCode`, `PeekSeq`, `CreateRevision`, `CloneTemplateRequest`, `NewCloneTemplateRequest`, `DocumentRef`, `NewDocumentRef`, `DocumentInitializer`, `EnsureCounter`, `Peek`, `NextAndIncrement`). 79 exports lack doc comments. Latent     readability + IDE tooling impact.
- **Evidence:** `_artifacts/01-surface.md`   2
- **Linked backlog row:** [`backlog/controlled-documents-refactor.md#R-012`](../backlog/controlled-documents-refactor.md)
- **Linked ADR:** missing-ADR

---

## Coverage stats (computed at compose time)

- Public symbols undocumented: 79 / 94
- Operations missing C4 placement: 0 / 8
- Cross-deps missing in   5/  8: 0 / 19
- State transitions missing in   6: 0 / 2
- Decisions without ADR link: 9 / 12


