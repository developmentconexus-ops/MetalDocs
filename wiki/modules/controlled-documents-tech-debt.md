# Tech Debt Register — controlled-documents

> Companion to [wiki/modules/controlled-documents.md](controlled-documents.md). Lists known gaps, smells, and missing-ADR items. **Debt only — no fix prescriptions.** Fixes belong in [wiki/backlog/controlled-documents-refactor.md](../backlog/controlled-documents-refactor.md).

**Last verified:** 2026-07-02 (TST-09 closed — tenant_isolation_test.go's six unconditional-skip stubs replaced with real integration-tagged tests on the canonical testdb factory; compile-verified, main session runs against local DB) | **Prior:** 2026-07-01 (Grade-A simplification register reconciliation — T-005 fully closed, cd_sequence_counters RLS confirmed via migration 0237/commit ad70f641 folded into baseline by fa5b6fd9; T-006 re-verified still open, no authz.Require on GetActiveDocument read path) | **Prior:** 2026-06-12 (Wave 2.12 sync — db==nil authz-bypass class-B branch in Create deleted (authz now unconditional); DBTX replaced by db.Tx; sequence.go no longer imports database/sql; orphan document_subjects column+index noted as deferred; no existing debt rows opened or closed) | **Prior:** 2026-06-12 (Wave 2 sync — T-005 partially addressed by RLS migration 0234) | **Prior:** 2026-06-11

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
- **Surface:** `internal/modules/controlleddocuments/application/service.go:451` (`Obsolete` passes `string(iamdomain.CapControlledDocumentObsolete)` to `changeStatus`)    `:455` (`Supersede` passes `CapControlledDocumentSupersede`)    `:491` (`changeStatus` defined)    `:534` (`changeStatus` calls `authz.Require(ctx, tx, cap, areaCode)` — area-scoped, not tenant-wide). `migrations/0187_registry_lifecycle_caps_seed.sql` (legacy literal migration filename) performs lifecycle-capability reseeding. `apps/api/cmd/metaldocs-api/permissions.go:187-188` — PUT method rows map `/api/v1/controlled-documents/{id}/obsolete` → `CapControlledDocumentObsolete` and `/api/v1/controlled-documents/{id}/supersede` → `CapControlledDocumentSupersede` at tier-1.
- **Observation (original):** `Obsolete` and `Supersede` handlers + service contained no `authz.Require`, no `CapabilityService` call, no capability-name constant. `migrations/0165_role_capabilities_reseed.sql` seeded only `registry.create` (legacy literal capability key). Any authenticated user could transition the canonical QMS catalog.
- **Evidence:** `controlled-documents/_artifacts/02-flow-obsolete.md`   4
- **Linked backlog row:** [`backlog/controlled-documents-refactor.md#R-001`](../backlog/controlled-documents-refactor.md)
- **Linked ADR:** missing-ADR

### T-002    Audit-trail gap on Obsolete / Supersede     CLOSED 2026-05-11 (Plan 6a)
- **Severity:** critical (closed)
- **Surface:** `internal/modules/controlleddocuments/application/service.go:491-570` (`changeStatus` definition through govLogger emission)
- **Observation:** `changeStatus` originally performed get + active-guard + UPDATE only, with no `s.govLogger.Log(...)` call. The create path emits governance events at `service.go:347-351` (after document commit); the lifecycle path did not. Transitioning a controlled document from `active` to `obsolete` or `superseded` is a regulated QMS event under ISO 9001     Critical trigger "Regulated audit-trail gap" fires. Fixed: `changeStatus` now emits at `service.go:561-570` post-commit.
- **Evidence:** `controlled-documents/_artifacts/02-flow-obsolete.md`   6
- **Linked backlog row:** [`backlog/controlled-documents-refactor.md#R-002`](../backlog/controlled-documents-refactor.md)
- **Linked ADR:** missing-ADR

### T-003    Legacy `{code, message}` error envelope across module     CLOSED 2026-05-12 (Plan 7)
- **Severity:** major (closed)
- **Surface (resolved):** `internal/platform/httpresponse/response.go:16-18`     `WriteError` now delegates to `problem.Write(w, problem.New(status, code, message))`. All controlled-documents routes that called `httpresponse.WriteError` inherit RFC 9457 `application/problem+json` output. `internal/modules/controlleddocuments/delivery/http/routes.go:560-561`     `ErrTemplateProfileMismatch` branch directly calls `problem.Write` with 422 `template_invalid`.
- **Observation (original):** Errors emitted JSON object `{"code": "...", "message": "..."}` with default content-type. RFC 9457 Problem Details (`application/problem+json`) was not used. Peer modules (documents T-001, audit T-002, templates T-005, approval T-001) carried the same gap.
- **Evidence:** `controlled-documents/_artifacts/02-flow-atomic-create.md`   5; `controlled-documents/_artifacts/05-industry.md` IP-001
- **Linked backlog row:** [`backlog/controlled-documents-refactor.md#R-003`](../backlog/controlled-documents-refactor.md) (merged Plan 7 2026-05-11, commit `11589032` + `395b0b24`)
- **Linked ADR:** `wiki/architecture/api-design-system.md`

### T-004    Tier-3 Postgres tripwire absent for controlled-documents-owned tables     CLOSED 2026-05-11 (Plan 5)
- **Severity:** major (closed)
- **Surface (resolved):** `internal/modules/controlleddocuments/infrastructure/repository.go:333-350` (`Create` opens tx + calls `authz.Require(CapControlledDocumentCreate)` at `:341`)    `:353-365` (`CreateTx` calls `authz.Require` at `:362`). `migrations/0188_tripwire_extend.sql:201-208` (legacy literal migration filename) attaches `trg_require_cap_asserted` to `public.controlled_documents` (INSERT + UPDATE, with OR-logic for `controlled_documents.obsolete|controlled_documents.supersede` on UPDATE) and `public.cd_sequence_counters` (line 206).
- **Observation (original):** 5 mutator methods (`Create`, `CreateTx`, `UpdateStatus`, `EnsureCounter`, `NextAndIncrement`) executed INSERT/UPDATE without preceding `authz.Require(...)`. None set `metaldocs.asserted_caps`. The `enforce_capability_asserted` trigger installed by 0142b covered `approval_instances` and `signoffs`; `controlled_documents` and `cd_sequence_counters` were NOT in the protected set.
- **Evidence:** `controlled-documents/_artifacts/04-persistence.md`   5 (5 violations); `controlled-documents/_artifacts/05-industry.md` IP-004
- **Linked backlog row:** [`backlog/controlled-documents-refactor.md#R-004`](../backlog/controlled-documents-refactor.md)
- **Linked ADR:** [wiki/decisions/0007-two-tier-authz.md](../decisions/0007-two-tier-authz.md)

### T-005    Tenant scoping via query arg only — no GUC + RLS backstop — CLOSED 2026-07-01 (Wave Z, migration 0237)
- **Severity:** major (closed)
- **Surface:** `internal/modules/controlleddocuments/infrastructure/repository.go` — `GetByID`, `GetByCode`, `CodeExists`, `List`, `CreateTx`, `UpdateStatus`, `NextAndIncrement` (all include `tenant_id = $...` predicate)
- **Wave 2 progress:** Migration 0234 (`db/migrations/0234_rls_controlled_documents_audit_events.sql`) applied ENABLE + FORCE ROW LEVEL SECURITY + NULL-permissive `tenant_isolation` policy to `public.controlled_documents`.
- **Resolution (Wave Z):** `cd_sequence_counters` residual gap closed by migration 0237 ("rls_all_tenant_tables", commit `ad70f641` — "Z-2/Z-3 RLS on all 27 remaining tenant tables + idempotency tenant FK (F-12 tail, RF-6, REQ-TEN-1, F-09d, ADR 0027 executed in full)"), archived at `archive/migrations/post-baseline-2026-06-fold/0237_rls_all_tenant_tables.sql:179-183` (`ALTER TABLE public.cd_sequence_counters ENABLE ROW LEVEL SECURITY; ... FORCE ROW LEVEL SECURITY;` + `tenant_isolation` policy). Folded into the curated baseline by commit `fa5b6fd9`; confirmed live in `db/baseline/0001_current_schema.sql:4687` (`ENABLE ROW LEVEL SECURITY`) and `:4791` (`CREATE POLICY tenant_isolation ON public.cd_sequence_counters ...`). RLS is effective for NOSUPERUSER+NOBYPASSRLS production roles (dev Docker superuser bypasses).
- **Observation (original):** Every WHERE clause includes `tenant_id = $...` from the request context (sourced via `tenant.FromContext` — Plan 3 removed the `X-Tenant-ID` header source). `cd_sequence_counters` had query-arg-only scoping with no RLS or GUC backstop; now closed.
- **Evidence:** `db/baseline/0001_current_schema.sql:4687,4791`; `archive/migrations/post-baseline-2026-06-fold/0237_rls_all_tenant_tables.sql:179-183`; commits `ad70f641`, `fa5b6fd9`; `controlled-documents/_artifacts/04-persistence.md` §5; `controlled-documents/_artifacts/05-industry.md` IP-008 (stale — still shows T-005 open, needs separate refresh by wiki-curator).
- **Linked backlog row:** [`backlog/controlled-documents-refactor.md#R-005`](../backlog/controlled-documents-refactor.md) (can be closed)
- **Linked ADR:** missing-ADR

### T-006    GetActiveDocument: no authz     CLOSED 2026-07-01 (SEC-03)
- **Severity:** major (closed)
- **Surface (resolved):** tier-2 check lives in the application layer, not the handler: `Service.GetActiveInstance` (`internal/modules/controlleddocuments/application/service.go:643`) runs in-tx `authz.Require(CapDocumentView, "tenant")` after `SeedTxIdentity` (commit `b113ba51`, with deny/allow integration tests). The HTTP handler (`routes.go:276`) delegates to that service method, so the read path is capability-gated end-to-end.
- **Observation (original):** No `authz.Require` call; no `metaldocs.assert_caps`. Document content hashes, approval state, and published-revision IDs were returned to any authenticated caller. Plan 3 had already resolved the header-trust sub-issue (tenant from `tenant.FromContext`); the residual read-policy gap is what SEC-03 closed.
- **Evidence:** `controlled-documents/_artifacts/02-flow-get-active.md`   2,   4 (pre-fix trace)
- **Linked backlog row:** [`backlog/controlled-documents-refactor.md#R-006`](../backlog/controlled-documents-refactor.md) (can be closed)
- **Linked ADR:** ADR 0022 (two-tier PDP)

### T-007    OpenAPI spec/handler drift on 422 `template_invalid`     CLOSED 2026-05-12 (Plan 7)
- **Severity:** major (closed)
- **Surface (resolved):** `internal/modules/controlleddocuments/delivery/http/routes.go:560-561`     `writeDomainError` now has a branch: `errors.Is(err, controlleddocumentsdomain.ErrTemplateProfileMismatch)`     `problem.Write(w, problem.New(http.StatusUnprocessableEntity, "template_invalid", "template version does not match the document profile"))`. Spec 422 case and runtime are now aligned.
- **Observation (original):** Spec declared 422 `template_invalid` on `POST /controlled-documents`; handler's `writeDomainError` switch had no branch mapping any error to that code. Contract drift     downstream OpenAPI clients included a 422 case the server never emitted.
- **Evidence:** `controlled-documents/_artifacts/02-flow-atomic-create.md`   5
- **Linked backlog row:** [`backlog/controlled-documents-refactor.md#R-007`](../backlog/controlled-documents-refactor.md) (merged Plan 7 2026-05-11, commit `395b0b24`)
- **Linked ADR:** [wiki/decisions/0012-contract-first-api.md](../decisions/0012-contract-first-api.md)

### T-008    Cross-module audit sink — taxonomy logger reused — FULLY CLOSED 2026-06-12 (Wave 2.12)
- **Severity:** major (fully closed)
- **Surface (resolved):** `internal/modules/controlleddocuments/module.go:35` — `govLogger := taxonomyapp.NewAuditGovernanceAdapter(deps.AuditWriter)`. The `NewDBGovernanceLogger(deps.DB)` call and `governance_logger.go` are deleted. `AuditWriter` is required (`module.go:27-29` panics on nil). Controlled-documents audit emissions now go to `metaldocs.audit_events` via the canonical `auditdomain.Writer` — no more dual-sink split.
- **Observation (original):** Controlled-documents used `taxonomyapp.NewDBGovernanceLogger(deps.DB)` as its governance adapter, causing emissions to land in `governance_events` instead of the canonical audit sink.
- **Evidence:** `internal/modules/controlleddocuments/module.go:27-35`.
- **Linked backlog row:** [`backlog/controlled-documents-refactor.md#R-008`](../backlog/controlled-documents-refactor.md)
- **Linked ADR:** missing-ADR

### T-010    Orphan `documents.subject_code` column + index (Wave 2.12 deferred)
- **Severity:** minor
- **Surface:** `db/migrations/0236` (migration that dropped `document_subjects` table with CASCADE, which dropped the FK from `controlled_documents`). The `documents.subject_code` column and its index remain as orphans.
- **Observation:** Migration 0236 dropped `document_subjects` and its FK CASCADE deleted the `document_subjects_document_subject_code_fkey` constraint on `controlled_documents`. However the `subject_code` column itself and its index remain. They reference a non-existent FK target (table gone), making them dead schema. Next-touch trigger: any migration that touches the `controlled_documents` table or schema.
- **Linked backlog row:** none yet
- **Linked ADR:** missing-ADR

### T-009    Documents DI cycle resolved via post-construction setter
- **Severity:** minor
- **Surface:** `apps/api/cmd/metaldocs-api/main.go:224`, `:343`; `internal/modules/controlleddocuments/application/service.go:99`
- **Observation:** `controlledDocumentsModule.New(...)` constructs `ControlledDocumentService` with the 8th argument `nil` (the `DocumentInitializer`). `main.go:343` calls `controlledDocumentsModule.Service().WithDocumentInitializer(docapp.NewCDDocumentInitializer(docMod.Service))` after the documents module is built. Cycle break is intentional (controlled-documents owns the port; documents implements). Order-of-construction is now a hidden contract     if a future caller forgets the setter, `Create` on `ControlledDocumentService` will nil-panic on the port call (`service.go:247`). Latent.
- **Evidence:** `controlled-documents/_artifacts/03-deps.md`   3
- **Linked backlog row:** [`backlog/controlled-documents-refactor.md#R-009`](../backlog/controlled-documents-refactor.md)
- **Linked ADR:** [wiki/decisions/0011-cd-atomic-create.md](../decisions/0011-cd-atomic-create.md)

### T-010    Parallel repository instance constructed outside module boundary
- **Severity:** minor
- **Surface:** `apps/api/cmd/metaldocs-api/main.go:224`
- **Observation:** A second `PostgresControlledDocumentRepository` is instantiated standalone at `main.go:224` for search/resolver wiring. The module exposes `Module.Service()` but not its internal repository, so consumers reach in via `controlleddocumentsinfra.NewPostgresControlledDocumentRepository`. Module-boundary leak; ties external code to the internal infrastructure package layout.
- **Evidence:** `controlled-documents/_artifacts/03-deps.md`   3
- **Linked backlog row:** [`backlog/controlled-documents-refactor.md#R-010`](../backlog/controlled-documents-refactor.md)
- **Linked ADR:** missing-ADR

### T-011    OpenAPI partial at `v1/` despite `/api/v1/` HTTP path
- **Severity:** minor
- **Surface:** canonical spec partial `api/openapi/v1/partials/controlled-documents.yaml`; canonical public API remains `/api/v1/controlled-documents/*`.
- **Observation:** The spec partial lives under `api/openapi/v1/` while the HTTP path prefix is `/api/v1/`. Generated server stubs (`internal/modules/controlleddocuments/api/api.gen.go`) and clients encode `/api/v1/...` routes from that partial; the naming/layout mismatch is cosmetic but confusing for new contributors and can complicate future spec-tree restructuring.
- **Evidence:** `controlled-documents/_artifacts/02-flow-atomic-create.md`   1
- **Linked backlog row:** [`backlog/controlled-documents-refactor.md#R-011`](../backlog/controlled-documents-refactor.md)
- **Linked ADR:** missing-ADR

### TST-09    Tenant-isolation coverage was unconditional-skip stubs     CLOSED 2026-07-02 (grade-A simplification program)
- **Severity:** major (closed)
- **Surface (resolved):** `internal/modules/controlleddocuments/application/tenant_isolation_test.go` — six real `//go:build integration` tests on the canonical testdb factory (ADR 0034): `TestTenantIsolation_GetByID_CrossTenant_NotFound`, `TestTenantIsolation_GetByCode_CrossTenant_NotFound` (+ `CodeExists` parity), `TestTenantIsolation_ListControlledDocuments_CrossTenant`, `TestTenantIsolation_VisibilityGrants_CrossTenant` (CanRead + GetByID against a foreign-tenant actor), `TestTenantIsolation_SequenceCounters_CrossTenant` (`cd_sequence_counters` keyed independently per tenant under a deliberately-shared profile/area code), `TestTenantIsolation_CreateCD_CrossTenantProfile_NotFound` (service-level `Create` against a profile/area that exists only in another tenant, through the real `taxonomyinfra` repositories wrapped by the canonical `controlleddocumentsinfrastructure.TaxonomyProfileReader`/`TaxonomyAreaReader` adapters — not fakes).
- **Observation (original):** All six tests in `tenant_isolation_test.go` were unconditional `t.Skip("requires live DB")` stubs (no live-DB assertion ever ran). Multi-tenant pooled isolation is a non-negotiable invariant (tenant_id everywhere, tx-local GUCs, cross-tenant URL -> 404); this module's tenant-scoped SQL (`repository.go` GetByID/GetByCode/List/CanRead, `cd_sequence_counters`) was guarded only by the DB tripwire suite (capability assertion, not tenant scoping) and RLS (T-005, schema-level only) — no test proved the query-level `tenant_id = $1` predicates actually exclude a foreign tenant's rows.
- **Resolution rationale:** verified no equivalent coverage existed elsewhere (grep for `memberships`/`CrossTenantFK`/`DocumentsTrigger` across the module returned only the stub file itself) — implemented rather than deleted. Two stub names were retargeted to concepts the module actually owns: `ListMemberships_CrossTenant` (iam group/role memberships, not a controlled-documents concept) became `VisibilityGrants_CrossTenant` (the CD-owned `controlled_document_area_grants`/`controlled_document_user_grants` tables via `CanRead`); `CreateCD_CrossTenantProfile_Returns404` became `..._NotFound` (404 is an HTTP-layer concept — the service-level test pins the underlying `taxonomydomain.ErrProfileNotFound` the handler maps to 404 `PROFILE_NOT_FOUND`).
- **Evidence:** `go build -tags integration ./internal/modules/controlleddocuments/...` and `go vet -tags integration ./internal/modules/controlleddocuments/...` both clean; `scripts/check-test-discipline.sh` reports zero R1–R4 violations in the new file.
- **Linked backlog row:** none (test-only closure)
- **Linked ADR:** [wiki/decisions/0034-integration-test-fixture-framework.md](../decisions/0034-integration-test-fixture-framework.md)

### T-012    Exported symbols mostly without Go doc comments
- **Severity:** minor
- **Surface:** `internal/modules/controlleddocuments/**/*.go`
- **Observation:** Surface scan recorded 94 exported symbols after the 2026-05-25 constructor additions. Doc comments present on ~15 (notably `CreateResult`, `WithDocumentInitializer`, `PreviewCode`, `PeekSeq`, `CreateRevision`, `CloneTemplateRequest`, `NewCloneTemplateRequest`, `DocumentRef`, `NewDocumentRef`, `DocumentInitializer`, `EnsureCounter`, `Peek`, `NextAndIncrement`). 79 exports lack doc comments. Latent     readability + IDE tooling impact.
- **Evidence:** `controlled-documents/_artifacts/01-surface.md`   2
- **Linked backlog row:** [`backlog/controlled-documents-refactor.md#R-012`](../backlog/controlled-documents-refactor.md)
- **Linked ADR:** missing-ADR

---

## Coverage stats (computed at compose time)

- Public symbols undocumented: 79 / 94
- Operations missing C4 placement: 0 / 8
- Cross-deps missing in   5/  8: 0 / 19
- State transitions missing in   6: 0 / 2
- Decisions without ADR link: 9 / 12


